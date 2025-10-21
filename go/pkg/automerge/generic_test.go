package automerge_test

import (
	"testing"
)

func TestDocument_PutGetRoot(t *testing.T) {
	doc, ctx := newTestDoc(t)

	// Test string value
	err := doc.PutRoot(ctx, "name", "Alice")
	if err != nil {
		t.Fatalf("PutRoot failed: %v", err)
	}

	value, err := doc.GetRoot(ctx, "name")
	if err != nil {
		t.Fatalf("GetRoot failed: %v", err)
	}

	if value != "Alice" {
		t.Errorf("Expected 'Alice', got %q", value)
	}
}

func TestDocument_PutRoot_Numbers(t *testing.T) {
	doc, ctx := newTestDoc(t)

	// Test integer
	err := doc.PutRoot(ctx, "age", "30")
	if err != nil {
		t.Fatalf("PutRoot failed: %v", err)
	}

	value, err := doc.GetRoot(ctx, "age")
	if err != nil {
		t.Fatalf("GetRoot failed: %v", err)
	}

	if value != "30" {
		t.Errorf("Expected '30', got %q", value)
	}
}

func TestDocument_PutRoot_Boolean(t *testing.T) {
	doc, ctx := newTestDoc(t)

	// Test boolean
	err := doc.PutRoot(ctx, "active", "true")
	if err != nil {
		t.Fatalf("PutRoot failed: %v", err)
	}

	value, err := doc.GetRoot(ctx, "active")
	if err != nil {
		t.Fatalf("GetRoot failed: %v", err)
	}

	if value != "true" {
		t.Errorf("Expected 'true', got %q", value)
	}
}

func TestDocument_DeleteRoot(t *testing.T) {
	doc, ctx := newTestDoc(t)

	// Put a value
	err := doc.PutRoot(ctx, "temp", "delete-me")
	if err != nil {
		t.Fatalf("PutRoot failed: %v", err)
	}

	// Verify it exists
	value, err := doc.GetRoot(ctx, "temp")
	if err != nil {
		t.Fatalf("GetRoot failed: %v", err)
	}
	if value != "delete-me" {
		t.Errorf("Expected 'delete-me', got %q", value)
	}

	// Delete it
	err = doc.DeleteRoot(ctx, "temp")
	if err != nil {
		t.Fatalf("DeleteRoot failed: %v", err)
	}

	// Verify it's gone (should return error)
	_, err = doc.GetRoot(ctx, "temp")
	if err == nil {
		t.Error("Expected error when getting deleted key, got none")
	}
}

func TestDocument_PutObjectRoot(t *testing.T) {
	doc, ctx := newTestDoc(t)

	tests := []struct {
		name    string
		key     string
		objType string
		wantErr bool
	}{
		{"create map", "users", "map", false},
		{"create list", "items", "list", false},
		{"create text", "notes", "text", false},
		{"invalid type", "bad", "invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := doc.PutObjectRoot(ctx, tt.key, tt.objType)
			if (err != nil) != tt.wantErr {
				t.Errorf("PutObjectRoot() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				// Verify object was created (should return "<object>")
				value, err := doc.GetRoot(ctx, tt.key)
				if err != nil {
					t.Errorf("GetRoot failed: %v", err)
				}
				if value != "<object>" {
					t.Errorf("Expected '<object>', got %q", value)
				}
			}
		})
	}
}

func TestDocument_MultipleRootValues(t *testing.T) {
	doc, ctx := newTestDoc(t)

	// Put multiple values
	values := map[string]string{
		"name":    "Bob",
		"age":     "25",
		"active":  "false",
		"score":   "100",
	}

	for key, value := range values {
		err := doc.PutRoot(ctx, key, value)
		if err != nil {
			t.Fatalf("PutRoot(%s) failed: %v", key, err)
		}
	}

	// Verify all values
	for key, expected := range values {
		value, err := doc.GetRoot(ctx, key)
		if err != nil {
			t.Fatalf("GetRoot(%s) failed: %v", key, err)
		}
		if value != expected {
			t.Errorf("Key %s: expected %q, got %q", key, expected, value)
		}
	}
}

func TestDocument_OverwriteValue(t *testing.T) {
	doc, ctx := newTestDoc(t)

	// Put initial value
	err := doc.PutRoot(ctx, "counter", "0")
	if err != nil {
		t.Fatalf("PutRoot failed: %v", err)
	}

	// Overwrite with new value
	err = doc.PutRoot(ctx, "counter", "42")
	if err != nil {
		t.Fatalf("PutRoot failed: %v", err)
	}

	// Verify new value
	value, err := doc.GetRoot(ctx, "counter")
	if err != nil {
		t.Fatalf("GetRoot failed: %v", err)
	}

	if value != "42" {
		t.Errorf("Expected '42', got %q", value)
	}
}

func TestDocument_GetRoot_NonExistentKey(t *testing.T) {
	doc, ctx := newTestDoc(t)

	// Try to get non-existent key
	_, err := doc.GetRoot(ctx, "nonexistent")
	if err == nil {
		t.Error("Expected error when getting non-existent key, got none")
	}
}
