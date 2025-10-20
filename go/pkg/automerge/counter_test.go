package automerge

import (
	"context"
	"testing"
)

func TestCounter_Increment(t *testing.T) {
	ctx := context.Background()
	doc, err := New(ctx)
	if err != nil {
		t.Fatalf("Failed to create document: %v", err)
	}

	key := "score"

	// First increment creates counter
	err = doc.Increment(ctx, Root(), key, 10)
	if err != nil {
		t.Fatalf("Increment failed: %v", err)
	}

	val, err := doc.GetCounter(ctx, Root(), key)
	if err != nil {
		t.Fatalf("GetCounter failed: %v", err)
	}
	if val != 10 {
		t.Errorf("Counter value = %d, want 10", val)
	}

	// Increment again
	doc.Increment(ctx, Root(), key, 5)
	val, _ = doc.GetCounter(ctx, Root(), key)
	if val != 15 {
		t.Errorf("Counter value = %d, want 15", val)
	}
}

func TestCounter_Decrement(t *testing.T) {
	ctx := context.Background()
	doc, err := New(ctx)
	if err != nil {
		t.Fatalf("Failed to create document: %v", err)
	}

	key := "balance"

	doc.Increment(ctx, Root(), key, 100)
	doc.Increment(ctx, Root(), key, -30)

	val, err := doc.GetCounter(ctx, Root(), key)
	if err != nil {
		t.Fatalf("GetCounter failed: %v", err)
	}
	if val != 70 {
		t.Errorf("Counter value = %d, want 70", val)
	}
}

func TestCounter_SaveLoad(t *testing.T) {
	ctx := context.Background()
	doc1, err := New(ctx)
	if err != nil {
		t.Fatalf("Failed to create document: %v", err)
	}

	doc1.Increment(ctx, Root(), "clicks", 42)

	data, err := doc1.Save(ctx)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	doc2, err := Load(ctx, data)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	val, err := doc2.GetCounter(ctx, Root(), "clicks")
	if err != nil {
		t.Fatalf("GetCounter failed: %v", err)
	}
	if val != 42 {
		t.Errorf("Counter value after load = %d, want 42", val)
	}
}
