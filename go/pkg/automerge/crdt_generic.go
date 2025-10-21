package automerge

import (
	"context"
	"fmt"
)

// Generic object operations for working with arbitrary CRDT structures.
// These operate at the ROOT level. Full nested path support can be added later.

// PutRoot puts a scalar value at the document root.
//
// Values are automatically typed:
// - Numbers (integer or float)
// - Booleans ("true" or "false")
// - Strings (any other value)
// - Null ("null")
//
// Example:
//   doc.PutRoot(ctx, "name", "Alice")
//   doc.PutRoot(ctx, "age", "30")
//   doc.PutRoot(ctx, "active", "true")
//
// Status: ✅ Implemented
func (d *Document) PutRoot(ctx context.Context, key string, value string) error {
	if d.runtime == nil {
		return fmt.Errorf("document not initialized")
	}
	return d.runtime.AmPutRoot(ctx, key, value)
}

// GetRoot gets a value from the document root.
//
// Returns the value as a string. For objects, returns "<object>".
//
// Status: ✅ Implemented
func (d *Document) GetRoot(ctx context.Context, key string) (string, error) {
	if d.runtime == nil {
		return "", fmt.Errorf("document not initialized")
	}
	return d.runtime.AmGetRoot(ctx, key)
}

// DeleteRoot deletes a key from the document root.
//
// This is a CRDT delete - other peers will see the deletion even if they
// have concurrent modifications.
//
// Status: ✅ Implemented
func (d *Document) DeleteRoot(ctx context.Context, key string) error {
	if d.runtime == nil {
		return fmt.Errorf("document not initialized")
	}
	return d.runtime.AmDeleteRoot(ctx, key)
}

// PutObjectRoot creates a nested CRDT object at the document root.
//
// Object types:
// - "map" - A key-value CRDT map
// - "list" - An ordered CRDT list
// - "text" - A collaborative text object
//
// Example:
//   doc.PutObjectRoot(ctx, "users", "map")
//   doc.PutObjectRoot(ctx, "items", "list")
//   doc.PutObjectRoot(ctx, "notes", "text")
//
// Status: ✅ Implemented
func (d *Document) PutObjectRoot(ctx context.Context, key string, objType string) error {
	if d.runtime == nil {
		return fmt.Errorf("document not initialized")
	}

	// Validate object type
	switch objType {
	case "map", "list", "text":
		// Valid
	default:
		return fmt.Errorf("invalid object type: %s (must be map, list, or text)", objType)
	}

	return d.runtime.AmPutObjectRoot(ctx, key, objType)
}
