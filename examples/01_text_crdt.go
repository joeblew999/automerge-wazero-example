package main

import (
	"context"
	"fmt"
	"log"

	"github.com/joeblew999/automerge-wazero-example/pkg/automerge"
	"github.com/joeblew999/automerge-wazero-example/pkg/wazero"
)

// Example 1: Basic Text CRDT Operations
//
// This example demonstrates:
// - Creating an Automerge document
// - Inserting text
// - Getting text
// - Saving and loading documents

func main() {
	ctx := context.Background()

	// Step 1: Initialize the wazero runtime
	// This loads the Rust WASM module
	runtime, err := wazero.NewRuntime(ctx, "automerge.wasm")
	if err != nil {
		log.Fatalf("Failed to create runtime: %v", err)
	}
	defer runtime.Close()

	// Step 2: Create a new Automerge document
	doc := automerge.NewDocument(runtime)
	if err := doc.Init(ctx); err != nil {
		log.Fatalf("Failed to initialize document: %v", err)
	}

	fmt.Println("📝 Text CRDT Example")
	fmt.Println("===================")

	// Step 3: Insert text using TextSplice
	// TextSplice(ctx, index, delete_count, insert_text)
	text := "Hello, Automerge!"
	if err := doc.TextSplice(ctx, 0, 0, text); err != nil {
		log.Fatalf("Failed to insert text: %v", err)
	}
	fmt.Printf("✅ Inserted: %q\n", text)

	// Step 4: Read the text back
	result, err := doc.GetText(ctx)
	if err != nil {
		log.Fatalf("Failed to get text: %v", err)
	}
	fmt.Printf("📖 Current text: %q\n", result)

	// Step 5: Modify text (insert in middle)
	// Insert " CRDT" after "Hello,"
	if err := doc.TextSplice(ctx, 6, 0, " CRDT"); err != nil {
		log.Fatalf("Failed to modify text: %v", err)
	}

	result, _ = doc.GetText(ctx)
	fmt.Printf("✏️  After edit: %q\n", result)

	// Step 6: Delete some text
	// Delete " CRDT" (5 characters starting at index 6)
	if err := doc.TextSplice(ctx, 6, 5, ""); err != nil {
		log.Fatalf("Failed to delete text: %v", err)
	}

	result, _ = doc.GetText(ctx)
	fmt.Printf("🗑️  After delete: %q\n", result)

	// Step 7: Save document to binary format
	fmt.Println("\n💾 Persistence")
	fmt.Println("==============")

	snapshot, err := doc.Save(ctx)
	if err != nil {
		log.Fatalf("Failed to save document: %v", err)
	}
	fmt.Printf("✅ Saved %d bytes\n", len(snapshot))

	// Step 8: Load document from binary
	doc2 := automerge.NewDocument(runtime)
	if err := doc2.Load(ctx, snapshot); err != nil {
		log.Fatalf("Failed to load document: %v", err)
	}

	loaded, _ := doc2.GetText(ctx)
	fmt.Printf("📂 Loaded text: %q\n", loaded)

	// Step 9: Unicode support
	fmt.Println("\n🌍 Unicode Support")
	fmt.Println("==================")

	doc3 := automerge.NewDocument(runtime)
	doc3.Init(ctx)

	unicodeText := "Hello 世界! 🌟"
	doc3.TextSplice(ctx, 0, 0, unicodeText)

	result, _ = doc3.GetText(ctx)
	fmt.Printf("✅ Unicode text: %q\n", result)

	fmt.Println("\n🎉 Example complete!")
}
