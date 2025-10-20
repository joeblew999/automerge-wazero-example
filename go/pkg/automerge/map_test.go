package automerge

import (
	"context"
	"testing"
)

// TestMap_PutGet tests basic map operations
func TestMap_PutGet(t *testing.T) {
	ctx := context.Background()
	doc, err := New(ctx)
	if err != nil {
		t.Fatalf("Failed to create document: %v", err)
	}

	tests := []struct {
		name  string
		key   string
		value string
	}{
		{"simple string", "name", "Alice"},
		{"empty value", "empty", ""},
		{"unicode key", "名前", "Bob"},
		{"unicode value", "greeting", "Hello 世界!"},
		{"long value", "bio", "This is a very long biographical string that contains many characters and words."},
		{"special chars", "email", "test@example.com"},
		{"numbers as string", "count", "12345"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Put value
			err := doc.Put(ctx, Root(), tt.key, NewString(tt.value))
			if err != nil {
				t.Fatalf("Put failed: %v", err)
			}

			// Get value
			got, err := doc.Get(ctx, Root(), tt.key)
			if err != nil {
				t.Fatalf("Get failed: %v", err)
			}

			// Verify
			str, ok := got.AsString()
			if !ok {
				t.Fatalf("Expected string value, got non-string")
			}
			if str != tt.value {
				t.Errorf("Get returned wrong value: got %q, want %q", str, tt.value)
			}
		})
	}
}

// TestMap_Delete tests deletion of map keys
func TestMap_Delete(t *testing.T) {
	ctx := context.Background()
	doc, err := New(ctx)
	if err != nil {
		t.Fatalf("Failed to create document: %v", err)
	}

	// Add some keys
	keys := []string{"a", "b", "c", "d"}
	for _, key := range keys {
		err := doc.Put(ctx, Root(), key, NewString("value_"+key))
		if err != nil {
			t.Fatalf("Put failed for key %q: %v", key, err)
		}
	}

	// Verify length before deletion (4 user keys + "content" from init = 5)
	length, err := doc.Length(ctx, Root())
	if err != nil {
		t.Fatalf("Length failed: %v", err)
	}
	if length != 5 {
		t.Errorf("Expected length 5, got %d", length)
	}

	// Delete one key
	err = doc.Delete(ctx, Root(), "b")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify length after deletion
	length, err = doc.Length(ctx, Root())
	if err != nil {
		t.Fatalf("Length failed: %v", err)
	}
	if length != 4 {
		t.Errorf("Expected length 4 after deletion, got %d", length)
	}

	// Verify key is gone
	_, err = doc.Get(ctx, Root(), "b")
	if err == nil {
		t.Error("Expected error getting deleted key, got nil")
	}

	// Verify other keys still exist
	for _, key := range []string{"a", "c", "d"} {
		val, err := doc.Get(ctx, Root(), key)
		if err != nil {
			t.Errorf("Failed to get key %q after deleting different key: %v", key, err)
		}
		str, ok := val.AsString()
		if !ok || str != "value_"+key {
			t.Errorf("Wrong value for key %q: got %q", key, str)
		}
	}
}

// TestMap_Keys tests retrieving all map keys
func TestMap_Keys(t *testing.T) {
	ctx := context.Background()
	doc, err := New(ctx)
	if err != nil {
		t.Fatalf("Failed to create document: %v", err)
	}

	// Add keys
	expectedKeys := map[string]bool{
		"content": true, // From am_init()
		"foo":     true,
		"bar":     true,
		"baz":     true,
	}

	for key := range expectedKeys {
		if key == "content" {
			continue // Skip "content" - already exists from init
		}
		err := doc.Put(ctx, Root(), key, NewString("value"))
		if err != nil {
			t.Fatalf("Put failed for key %q: %v", key, err)
		}
	}

	// Get all keys
	keys, err := doc.Keys(ctx, Root())
	if err != nil {
		t.Fatalf("Keys failed: %v", err)
	}

	// Verify count
	if len(keys) != len(expectedKeys) {
		t.Errorf("Expected %d keys, got %d", len(expectedKeys), len(keys))
	}

	// Verify all keys present
	gotKeys := make(map[string]bool)
	for _, key := range keys {
		gotKeys[key] = true
	}

	for expected := range expectedKeys {
		if !gotKeys[expected] {
			t.Errorf("Missing key: %q", expected)
		}
	}
}

// TestMap_Length tests map length calculation
func TestMap_Length(t *testing.T) {
	ctx := context.Background()
	doc, err := New(ctx)
	if err != nil {
		t.Fatalf("Failed to create document: %v", err)
	}

	// Initial length (just "content" from init)
	length, err := doc.Length(ctx, Root())
	if err != nil {
		t.Fatalf("Length failed: %v", err)
	}
	if length != 1 {
		t.Errorf("Expected initial length 1, got %d", length)
	}

	// Add keys one by one
	for i := 1; i <= 10; i++ {
		key := string(rune('a' + i - 1))
		err := doc.Put(ctx, Root(), key, NewString("value"))
		if err != nil {
			t.Fatalf("Put failed: %v", err)
		}

		length, err := doc.Length(ctx, Root())
		if err != nil {
			t.Fatalf("Length failed: %v", err)
		}
		expected := uint(i + 1) // +1 for "content"
		if length != expected {
			t.Errorf("After adding %d keys, expected length %d, got %d", i, expected, length)
		}
	}
}

// TestMap_Overwrite tests overwriting existing keys
func TestMap_Overwrite(t *testing.T) {
	ctx := context.Background()
	doc, err := New(ctx)
	if err != nil {
		t.Fatalf("Failed to create document: %v", err)
	}

	key := "status"

	// Set initial value
	err = doc.Put(ctx, Root(), key, NewString("initial"))
	if err != nil {
		t.Fatalf("Put failed: %v", err)
	}

	// Overwrite multiple times
	values := []string{"updated", "modified", "final"}
	for _, value := range values {
		err := doc.Put(ctx, Root(), key, NewString(value))
		if err != nil {
			t.Fatalf("Put failed: %v", err)
		}

		got, err := doc.Get(ctx, Root(), key)
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}

		str, ok := got.AsString()
		if !ok || str != value {
			t.Errorf("Expected %q, got %q", value, str)
		}
	}

	// Verify length didn't increase (still same key)
	length, err := doc.Length(ctx, Root())
	if err != nil {
		t.Fatalf("Length failed: %v", err)
	}
	if length != 2 { // "content" + "status"
		t.Errorf("Expected length 2, got %d", length)
	}
}

// TestMap_EmptyKey tests edge case of empty string key
func TestMap_EmptyKey(t *testing.T) {
	ctx := context.Background()
	doc, err := New(ctx)
	if err != nil {
		t.Fatalf("Failed to create document: %v", err)
	}

	// Empty key should work
	err = doc.Put(ctx, Root(), "", NewString("empty key value"))
	if err != nil {
		t.Fatalf("Put with empty key failed: %v", err)
	}

	val, err := doc.Get(ctx, Root(), "")
	if err != nil {
		t.Fatalf("Get with empty key failed: %v", err)
	}

	str, ok := val.AsString()
	if !ok || str != "empty key value" {
		t.Errorf("Wrong value for empty key: got %q", str)
	}
}

// TestMap_NonStringValue tests that non-string values return NotImplementedError
func TestMap_NonStringValue(t *testing.T) {
	ctx := context.Background()
	doc, err := New(ctx)
	if err != nil {
		t.Fatalf("Failed to create document: %v", err)
	}

	tests := []struct {
		name  string
		value Value
	}{
		{"int", NewInt(42)},
		{"bool", NewBool(true)},
		{"float", NewFloat(3.14)},
		{"null", NewNull()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := doc.Put(ctx, Root(), "test", tt.value)
			if err == nil {
				t.Error("Expected NotImplementedError for non-string value, got nil")
			}
			if _, ok := err.(*NotImplementedError); !ok {
				t.Errorf("Expected NotImplementedError, got %T: %v", err, err)
			}
		})
	}
}

// TestMap_SaveLoad tests persistence of map operations
func TestMap_SaveLoad(t *testing.T) {
	ctx := context.Background()

	// Create document and add data
	doc1, err := New(ctx)
	if err != nil {
		t.Fatalf("Failed to create document: %v", err)
	}

	testData := map[string]string{
		"name":  "Alice",
		"email": "alice@example.com",
		"role":  "admin",
	}

	for k, v := range testData {
		err := doc1.Put(ctx, Root(), k, NewString(v))
		if err != nil {
			t.Fatalf("Put failed: %v", err)
		}
	}

	// Save
	data, err := doc1.Save(ctx)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Load into new document
	doc2, err := Load(ctx, data)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Verify all data survived
	for k, expected := range testData {
		got, err := doc2.Get(ctx, Root(), k)
		if err != nil {
			t.Fatalf("Get failed for key %q: %v", k, err)
		}

		str, ok := got.AsString()
		if !ok || str != expected {
			t.Errorf("Key %q: expected %q, got %q", k, expected, str)
		}
	}

	// Verify length
	length, err := doc2.Length(ctx, Root())
	if err != nil {
		t.Fatalf("Length failed: %v", err)
	}
	expectedLen := uint(len(testData) + 1) // +1 for "content"
	if length != expectedLen {
		t.Errorf("Expected length %d, got %d", expectedLen, length)
	}
}

// TestMap_NestedNotSupported tests that nested maps return NotImplementedError
func TestMap_NestedNotSupported(t *testing.T) {
	ctx := context.Background()
	doc, err := New(ctx)
	if err != nil {
		t.Fatalf("Failed to create document: %v", err)
	}

	// Try to access nested path
	nestedPath := Root().Get("nested")

	_, err = doc.Get(ctx, nestedPath, "key")
	if err == nil {
		t.Error("Expected NotImplementedError for nested map, got nil")
	}
	if _, ok := err.(*NotImplementedError); !ok {
		t.Errorf("Expected NotImplementedError, got %T: %v", err, err)
	}

	err = doc.Put(ctx, nestedPath, "key", NewString("value"))
	if err == nil {
		t.Error("Expected NotImplementedError for nested map, got nil")
	}

	err = doc.Delete(ctx, nestedPath, "key")
	if err == nil {
		t.Error("Expected NotImplementedError for nested map, got nil")
	}

	_, err = doc.Keys(ctx, nestedPath)
	if err == nil {
		t.Error("Expected NotImplementedError for nested map, got nil")
	}
}
