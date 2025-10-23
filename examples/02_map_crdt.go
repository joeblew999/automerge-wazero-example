package main

import (
	"context"
	"fmt"
	"log"

	"github.com/joeblew999/automerge-wazero-example/pkg/automerge"
	"github.com/joeblew999/automerge-wazero-example/pkg/wazero"
)

// Example 2: Map CRDT Operations
//
// This example demonstrates:
// - Creating key-value pairs
// - Getting values
// - Deleting keys
// - Listing all keys

func main() {
	ctx := context.Background()

	runtime, err := wazero.NewRuntime(ctx, "automerge.wasm")
	if err != nil {
		log.Fatalf("Failed to create runtime: %v", err)
	}
	defer runtime.Close()

	doc := automerge.NewDocument(runtime)
	doc.Init(ctx)

	fmt.Println("🗺️  Map CRDT Example")
	fmt.Println("===================")

	// Step 1: Put values into the map
	path := automerge.NewPath() // Root path

	// Put different types of values
	values := map[string]automerge.Value{
		"name":    automerge.NewStringValue("Alice"),
		"age":     automerge.NewIntValue(30),
		"active":  automerge.NewBoolValue(true),
		"balance": automerge.NewFloatValue(123.45),
	}

	for key, val := range values {
		if err := doc.Put(ctx, path, key, val); err != nil {
			log.Fatalf("Failed to put %s: %v", key, err)
		}
		fmt.Printf("✅ Set %s = %v\n", key, val)
	}

	// Step 2: Get values back
	fmt.Println("\n📖 Reading values:")

	name, err := doc.Get(ctx, path, "name")
	if err != nil {
		log.Fatalf("Failed to get name: %v", err)
	}
	fmt.Printf("  name: %v\n", name)

	age, _ := doc.Get(ctx, path, "age")
	fmt.Printf("  age: %v\n", age)

	// Step 3: List all keys
	fmt.Println("\n🔑 All keys:")

	keys, err := doc.Keys(ctx, path)
	if err != nil {
		log.Fatalf("Failed to get keys: %v", err)
	}
	for _, key := range keys {
		val, _ := doc.Get(ctx, path, key)
		fmt.Printf("  %s: %v\n", key, val)
	}

	// Step 4: Get map length
	length, err := doc.MapLength(ctx, path)
	if err != nil {
		log.Fatalf("Failed to get length: %v", err)
	}
	fmt.Printf("\n📏 Map size: %d entries\n", length)

	// Step 5: Delete a key
	fmt.Println("\n🗑️  Deleting 'balance':")

	if err := doc.Delete(ctx, path, "balance"); err != nil {
		log.Fatalf("Failed to delete: %v", err)
	}

	keys, _ = doc.Keys(ctx, path)
	fmt.Printf("✅ Remaining keys: %v\n", keys)

	// Step 6: Overwrite a value
	fmt.Println("\n✏️  Updating 'age':")

	doc.Put(ctx, path, "age", automerge.NewIntValue(31))
	newAge, _ := doc.Get(ctx, path, "age")
	fmt.Printf("✅ Updated age: %v\n", newAge)

	// Step 7: Save and load
	fmt.Println("\n💾 Persistence:")

	snapshot, err := doc.Save(ctx)
	if err != nil {
		log.Fatalf("Failed to save: %v", err)
	}
	fmt.Printf("✅ Saved %d bytes\n", len(snapshot))

	// Load into new document
	doc2 := automerge.NewDocument(runtime)
	doc2.Load(ctx, snapshot)

	keys2, _ := doc2.Keys(ctx, path)
	fmt.Printf("📂 Loaded keys: %v\n", keys2)

	fmt.Println("\n🎉 Example complete!")
}
