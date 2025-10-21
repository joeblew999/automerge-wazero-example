package automerge

import (
	"context"
	"testing"
)

func TestList_PushGet(t *testing.T) {
	ctx := context.Background()
	doc, err := NewWithWASM(ctx, TestWASMPath)
	if err != nil {
		t.Fatalf("Failed to create document: %v", err)
	}

	tests := []string{"first", "second", "third", "fourth"}

	for _, value := range tests {
		err := doc.ListPush(ctx, Root(), NewString(value))
		if err != nil {
			t.Fatalf("ListPush failed: %v", err)
		}
	}

	length, err := doc.ListLength(ctx, Root())
	if err != nil {
		t.Fatalf("ListLength failed: %v", err)
	}
	if length != uint(len(tests)) {
		t.Errorf("Length = %d, want %d", length, len(tests))
	}

	for i, expected := range tests {
		val, err := doc.ListGet(ctx, Root(), uint(i))
		if err != nil {
			t.Fatalf("ListGet(%d) failed: %v", i, err)
		}
		str, ok := val.AsString()
		if !ok || str != expected {
			t.Errorf("ListGet(%d) = %q, want %q", i, str, expected)
		}
	}
}

func TestList_Insert(t *testing.T) {
	ctx := context.Background()
	doc, err := NewWithWASM(ctx, TestWASMPath)
	if err != nil {
		t.Fatalf("Failed to create document: %v", err)
	}

	doc.ListPush(ctx, Root(), NewString("a"))
	doc.ListPush(ctx, Root(), NewString("c"))

	err = doc.ListInsert(ctx, Root(), 1, NewString("b"))
	if err != nil {
		t.Fatalf("ListInsert failed: %v", err)
	}

	expected := []string{"a", "b", "c"}
	for i, exp := range expected {
		val, err := doc.ListGet(ctx, Root(), uint(i))
		if err != nil {
			t.Fatalf("ListGet(%d) failed: %v", i, err)
		}
		str, _ := val.AsString()
		if str != exp {
			t.Errorf("Index %d: got %q, want %q", i, str, exp)
		}
	}
}

func TestList_Delete(t *testing.T) {
	ctx := context.Background()
	doc, err := NewWithWASM(ctx, TestWASMPath)
	if err != nil {
		t.Fatalf("Failed to create document: %v", err)
	}

	doc.ListPush(ctx, Root(), NewString("a"))
	doc.ListPush(ctx, Root(), NewString("b"))
	doc.ListPush(ctx, Root(), NewString("c"))

	err = doc.ListDelete(ctx, Root(), 1)
	if err != nil {
		t.Fatalf("ListDelete failed: %v", err)
	}

	length, _ := doc.ListLength(ctx, Root())
	if length != 2 {
		t.Errorf("Length after delete = %d, want 2", length)
	}

	val, _ := doc.ListGet(ctx, Root(), 0)
	if str, _ := val.AsString(); str != "a" {
		t.Errorf("Index 0 = %q, want \"a\"", str)
	}

	val, _ = doc.ListGet(ctx, Root(), 1)
	if str, _ := val.AsString(); str != "c" {
		t.Errorf("Index 1 = %q, want \"c\"", str)
	}
}

func TestList_SaveLoad(t *testing.T) {
	ctx := context.Background()
	doc1, err := NewWithWASM(ctx, TestWASMPath)
	if err != nil {
		t.Fatalf("Failed to create document: %v", err)
	}

	testData := []string{"item1", "item2", "item3"}
	for _, item := range testData {
		doc1.ListPush(ctx, Root(), NewString(item))
	}

	data, err := doc1.Save(ctx)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	doc2, err := LoadWithWASM(ctx, data, TestWASMPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	length, _ := doc2.ListLength(ctx, Root())
	if length != uint(len(testData)) {
		t.Errorf("Loaded length = %d, want %d", length, len(testData))
	}

	for i, expected := range testData {
		val, err := doc2.ListGet(ctx, Root(), uint(i))
		if err != nil {
			t.Fatalf("ListGet(%d) failed: %v", i, err)
		}
		str, _ := val.AsString()
		if str != expected {
			t.Errorf("Index %d: got %q, want %q", i, str, expected)
		}
	}
}
