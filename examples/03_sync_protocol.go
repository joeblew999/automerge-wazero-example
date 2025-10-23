package main

import (
	"context"
	"fmt"
	"log"

	"github.com/joeblew999/automerge-wazero-example/pkg/automerge"
	"github.com/joeblew999/automerge-wazero-example/pkg/wazero"
)

// Example 3: Sync Protocol
//
// This example demonstrates:
// - Two peers making independent changes
// - Generating sync messages
// - Receiving sync messages
// - CRDT convergence (both peers end up with same state)

func main() {
	ctx := context.Background()

	runtime, err := wazero.NewRuntime(ctx, "automerge.wasm")
	if err != nil {
		log.Fatalf("Failed to create runtime: %v", err)
	}
	defer runtime.Close()

	fmt.Println("🔄 Sync Protocol Example")
	fmt.Println("========================")

	// Step 1: Create two independent documents (Alice and Bob)
	alice := automerge.NewDocument(runtime)
	alice.Init(ctx)
	alice.SetActor(ctx, "alice")

	bob := automerge.NewDocument(runtime)
	bob.Init(ctx)
	bob.SetActor(ctx, "bob")

	fmt.Println("✅ Created two independent documents")

	// Step 2: Alice makes a change
	fmt.Println("\n📝 Alice's changes:")
	alice.TextSplice(ctx, 0, 0, "Hello from Alice!")
	aliceText, _ := alice.GetText(ctx)
	fmt.Printf("  Alice's document: %q\n", aliceText)

	// Step 3: Bob makes a different change
	fmt.Println("\n📝 Bob's changes:")
	bob.TextSplice(ctx, 0, 0, "Hello from Bob!")
	bobText, _ := bob.GetText(ctx)
	fmt.Printf("  Bob's document: %q\n", bobText)

	fmt.Println("\n⚠️  Documents are now diverged!")

	// Step 4: Initialize sync state for each peer
	fmt.Println("\n🔄 Starting sync:")

	aliceSyncState, err := alice.InitSyncState(ctx)
	if err != nil {
		log.Fatalf("Failed to init Alice's sync state: %v", err)
	}
	defer alice.FreeSyncState(ctx, aliceSyncState)

	bobSyncState, err := bob.InitSyncState(ctx)
	if err != nil {
		log.Fatalf("Failed to init Bob's sync state: %v", err)
	}
	defer bob.FreeSyncState(ctx, bobSyncState)

	// Step 5: Alice generates sync message for Bob
	fmt.Println("\n→ Alice generating sync message...")
	aliceMsg, err := alice.GenerateSyncMessage(ctx, aliceSyncState)
	if err != nil {
		log.Fatalf("Failed to generate Alice's message: %v", err)
	}
	fmt.Printf("✅ Alice's message: %d bytes\n", len(aliceMsg))

	// Step 6: Bob receives Alice's message
	fmt.Println("\n← Bob receiving Alice's message...")
	if err := bob.ReceiveSyncMessage(ctx, bobSyncState, aliceMsg); err != nil {
		log.Fatalf("Failed for Bob to receive message: %v", err)
	}
	fmt.Println("✅ Bob processed message")

	// Step 7: Bob generates response message
	fmt.Println("\n→ Bob generating sync message...")
	bobMsg, err := bob.GenerateSyncMessage(ctx, bobSyncState)
	if err != nil {
		log.Fatalf("Failed to generate Bob's message: %v", err)
	}
	fmt.Printf("✅ Bob's message: %d bytes\n", len(bobMsg))

	// Step 8: Alice receives Bob's message
	fmt.Println("\n← Alice receiving Bob's message...")
	if err := alice.ReceiveSyncMessage(ctx, aliceSyncState, bobMsg); err != nil {
		log.Fatalf("Failed for Alice to receive message: %v", err)
	}
	fmt.Println("✅ Alice processed message")

	// Step 9: Check convergence
	fmt.Println("\n🎯 After sync:")

	aliceText, _ = alice.GetText(ctx)
	bobText, _ = bob.GetText(ctx)

	fmt.Printf("  Alice's document: %q\n", aliceText)
	fmt.Printf("  Bob's document:   %q\n", bobText)

	if aliceText == bobText {
		fmt.Println("\n✅ SUCCESS: Documents have converged!")
	} else {
		fmt.Println("\n❌ ERROR: Documents did not converge!")
	}

	// Step 10: Show how changes merge
	fmt.Println("\n📊 Explaining the merge:")
	fmt.Println("  Both Alice and Bob's changes are preserved")
	fmt.Println("  Automerge's CRDT algorithm ensures they see the same result")
	fmt.Println("  This is CONFLICT-FREE merge!")

	fmt.Println("\n🎉 Example complete!")
}
