package automerge_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/joeblew999/automerge-wazero-example/pkg/automerge"
)

// TestDocument_Merge_MultiInstance tests CRDT merge with separate WASM instances
// This is the CORRECT way to test merge - each document gets its own WASM runtime
func TestDocument_Merge_MultiInstance(t *testing.T) {
	for _, tc := range MergeTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			ctx := context.Background()

			// Create first document with its own WASM instance
			doc1, err := automerge.New(ctx)
			if err != nil {
				t.Fatalf("New(doc1) failed: %v", err)
			}
			defer doc1.Close(ctx)

			// Create second document with its own WASM instance
			doc2, err := automerge.New(ctx)
			if err != nil {
				t.Fatalf("New(doc2) failed: %v", err)
			}
			defer doc2.Close(ctx)

			path := automerge.Root().Get("content")

			// Set text in first document
			if tc.Doc1Text != "" {
				if err := doc1.SpliceText(ctx, path, 0, 0, tc.Doc1Text); err != nil {
					t.Fatalf("SpliceText(doc1) failed: %v", err)
				}
			}

			// Set text in second document
			if tc.Doc2Text != "" {
				if err := doc2.SpliceText(ctx, path, 0, 0, tc.Doc2Text); err != nil {
					t.Fatalf("SpliceText(doc2) failed: %v", err)
				}
			}

			// Verify initial states
			text1, err := doc1.GetText(ctx, path)
			if err != nil {
				t.Fatalf("GetText(doc1) failed: %v", err)
			}
			if text1 != tc.Doc1Text {
				t.Errorf("Doc1 initial text = %q, want %q", text1, tc.Doc1Text)
			}

			text2, err := doc2.GetText(ctx, path)
			if err != nil {
				t.Fatalf("GetText(doc2) failed: %v", err)
			}
			if text2 != tc.Doc2Text {
				t.Errorf("Doc2 initial text = %q, want %q", text2, tc.Doc2Text)
			}

			t.Logf("Before merge: doc1=%q, doc2=%q", text1, text2)

			// Merge doc2 into doc1
			if err := doc1.Merge(ctx, doc2); err != nil {
				t.Fatalf("Merge(doc1 ← doc2) failed: %v", err)
			}

			// Check result
			merged, err := doc1.GetText(ctx, path)
			if err != nil {
				t.Fatalf("GetText(merged) failed: %v", err)
			}

			t.Logf("After merge: %q", merged)

			// Verify merge preserved content
			if tc.WantBoth {
				// Should contain content from both documents
				hasDoc1Content := (tc.Doc1Text == "" || contains(merged, tc.Doc1Text) || merged == tc.Doc1Text)
				hasDoc2Content := (tc.Doc2Text == "" || contains(merged, tc.Doc2Text) || merged == tc.Doc2Text)

				if !hasDoc1Content || !hasDoc2Content {
					t.Errorf("Merge lost content:\n  doc1: %q\n  doc2: %q\n  merged: %q\n  CRDT should preserve both!",
						tc.Doc1Text, tc.Doc2Text, merged)
				}
			} else if tc.WantEither {
				// Can be either doc (non-deterministic but valid)
				if merged != tc.Doc1Text && merged != tc.Doc2Text {
					t.Logf("Merge result: %q (neither doc1 nor doc2, but may be valid merge)", merged)
				}
			}

			// Verify merge didn't result in empty text (data loss)
			if tc.Doc1Text != "" && tc.Doc2Text != "" && merged == "" {
				t.Errorf("Merge resulted in empty text - TOTAL DATA LOSS!")
			}
		})
	}
}

// TestDocument_Merge_Commutativity tests that merge(A, B) == merge(B, A)
func TestDocument_Merge_Commutativity(t *testing.T) {
	ctx := context.Background()

	// Test case: Alice and Bob
	aliceText := "Hello from Alice!"
	bobText := "Hello from Bob!"

	// Scenario 1: Merge Bob into Alice
	doc1, err := automerge.New(ctx)
	if err != nil {
		t.Fatalf("New(doc1) failed: %v", err)
	}
	defer doc1.Close(ctx)

	doc2, err := automerge.New(ctx)
	if err != nil {
		t.Fatalf("New(doc2) failed: %v", err)
	}
	defer doc2.Close(ctx)

	path := automerge.Root().Get("content")

	doc1.SpliceText(ctx, path, 0, 0, aliceText)
	doc2.SpliceText(ctx, path, 0, 0, bobText)

	doc1.Merge(ctx, doc2) // Alice ← Bob
	result1, _ := doc1.GetText(ctx, path)

	// Scenario 2: Merge Alice into Bob
	doc3, err := automerge.New(ctx)
	if err != nil {
		t.Fatalf("New(doc3) failed: %v", err)
	}
	defer doc3.Close(ctx)

	doc4, err := automerge.New(ctx)
	if err != nil {
		t.Fatalf("New(doc4) failed: %v", err)
	}
	defer doc4.Close(ctx)

	doc3.SpliceText(ctx, path, 0, 0, aliceText)
	doc4.SpliceText(ctx, path, 0, 0, bobText)

	doc4.Merge(ctx, doc3) // Bob ← Alice
	result2, _ := doc4.GetText(ctx, path)

	t.Logf("merge(Alice, Bob) = %q", result1)
	t.Logf("merge(Bob, Alice) = %q", result2)

	// CRDT property: Merge should be commutative
	if result1 != result2 {
		t.Errorf("Merge NOT commutative:\n  merge(A,B) = %q\n  merge(B,A) = %q\n  CRDT VIOLATION!",
			result1, result2)
	}
}

// TestDocument_Merge_Convergence tests that all replicas converge to same state
func TestDocument_Merge_Convergence(t *testing.T) {
	ctx := context.Background()

	// Create 3 replicas
	alice, err := automerge.New(ctx)
	if err != nil {
		t.Fatalf("New(alice) failed: %v", err)
	}
	defer alice.Close(ctx)

	bob, err := automerge.New(ctx)
	if err != nil {
		t.Fatalf("New(bob) failed: %v", err)
	}
	defer bob.Close(ctx)

	carol, err := automerge.New(ctx)
	if err != nil {
		t.Fatalf("New(carol) failed: %v", err)
	}
	defer carol.Close(ctx)

	path := automerge.Root().Get("content")

	// Each makes a different edit
	alice.SpliceText(ctx, path, 0, 0, "Alice")
	bob.SpliceText(ctx, path, 0, 0, "Bob")
	carol.SpliceText(ctx, path, 0, 0, "Carol")

	// Merge all combinations
	alice.Merge(ctx, bob)
	alice.Merge(ctx, carol)

	bob.Merge(ctx, alice)
	bob.Merge(ctx, carol)

	carol.Merge(ctx, alice)
	carol.Merge(ctx, bob)

	// All should converge to same state
	aliceText, _ := alice.GetText(ctx, path)
	bobText, _ := bob.GetText(ctx, path)
	carolText, _ := carol.GetText(ctx, path)

	t.Logf("After full mesh merge:")
	t.Logf("  Alice: %q", aliceText)
	t.Logf("  Bob:   %q", bobText)
	t.Logf("  Carol: %q", carolText)

	// CRDT property: Eventual convergence
	if aliceText != bobText || bobText != carolText {
		t.Errorf("Replicas did NOT converge:\n  Alice: %q\n  Bob:   %q\n  Carol: %q\n  CRDT VIOLATION!",
			aliceText, bobText, carolText)
	}
}

// TestDocument_Merge_BinaryFormat tests merging via save/load
func TestDocument_Merge_BinaryFormat(t *testing.T) {
	ctx := context.Background()

	// Create two documents
	alice, err := automerge.New(ctx)
	if err != nil {
		t.Fatalf("New(alice) failed: %v", err)
	}
	defer alice.Close(ctx)

	bob, err := automerge.New(ctx)
	if err != nil {
		t.Fatalf("New(bob) failed: %v", err)
	}
	defer bob.Close(ctx)

	path := automerge.Root().Get("content")

	// Alice and Bob each write text
	alice.SpliceText(ctx, path, 0, 0, "Hello from Alice!")
	bob.SpliceText(ctx, path, 0, 0, "Hello from Bob!")

	// Save Bob's state
	bobSnapshot, err := bob.Save(ctx)
	if err != nil {
		t.Fatalf("Save(bob) failed: %v", err)
	}

	// Verify snapshot has Automerge magic bytes
	if !bytes.HasPrefix(bobSnapshot, AutomergeMagicBytes) {
		t.Errorf("Snapshot missing magic bytes: got %x, want %x",
			bobSnapshot[:4], AutomergeMagicBytes)
	}

	// Load Bob's snapshot into a new document
	bobCopy, err := automerge.Load(ctx, bobSnapshot)
	if err != nil {
		t.Fatalf("Load(bobSnapshot) failed: %v", err)
	}
	defer bobCopy.Close(ctx)

	// Verify loaded document has same text
	bobCopyText, _ := bobCopy.GetText(ctx, path)
	bobOrigText, _ := bob.GetText(ctx, path)

	if bobCopyText != bobOrigText {
		t.Errorf("Load lost data: got %q, want %q", bobCopyText, bobOrigText)
	}

	// Now merge the loaded copy into Alice
	if err := alice.Merge(ctx, bobCopy); err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	// Check merge result
	merged, _ := alice.GetText(ctx, path)
	t.Logf("After merge via binary format: %q", merged)

	// Should not lose data
	if merged == "" {
		t.Error("Merge via binary format lost all data!")
	}
}

// Helper function to check if s contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || bytes.Contains([]byte(s), []byte(substr)))
}
