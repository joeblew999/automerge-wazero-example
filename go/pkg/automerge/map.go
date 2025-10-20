package automerge

import "context"

// Map Operations - For M2 Milestone (Multi-Document Support)
//
// These methods will be implemented when we add support for multiple objects
// beyond the single root "content" text object.

// Get retrieves the value at a key in a map object.
//
// Status: ❌ Not implemented (requires M2 - multi-object support)
func (d *Document) Get(ctx context.Context, path Path, key string) (Value, error) {
	return Value{}, &NotImplementedError{
		Feature:   "Get",
		Milestone: "M2",
		Message:   "Map operations require WASI exports: am_get, am_put, am_delete",
	}
}

// GetAll retrieves all conflicting values at a key (for conflict resolution).
//
// Status: ❌ Not implemented (requires M2)
func (d *Document) GetAll(ctx context.Context, path Path, key string) ([]Value, error) {
	return nil, &NotImplementedError{
		Feature:   "GetAll",
		Milestone: "M2",
		Message:   "Requires am_get_all WASI export",
	}
}

// Put sets a value at a key in a map object.
//
// Status: ❌ Not implemented (requires M2)
func (d *Document) Put(ctx context.Context, path Path, key string, value Value) error {
	return &NotImplementedError{
		Feature:   "Put",
		Milestone: "M2",
		Message:   "Requires am_put WASI export",
	}
}

// PutObject creates a new object (Map, List, or Text) at a key.
//
// Returns the path to the newly created object.
//
// Status: ❌ Not implemented (requires M2)
func (d *Document) PutObject(ctx context.Context, path Path, key string, objType ObjType) (Path, error) {
	return Path{}, &NotImplementedError{
		Feature:   "PutObject",
		Milestone: "M2",
		Message:   "Requires am_put_object WASI export",
	}
}

// Delete removes a key from a map object.
//
// Status: ❌ Not implemented (requires M2)
func (d *Document) Delete(ctx context.Context, path Path, key string) error {
	return &NotImplementedError{
		Feature:   "Delete",
		Milestone: "M2",
		Message:   "Requires am_delete WASI export",
	}
}

// Keys returns all keys in a map object.
//
// Status: ❌ Not implemented (requires M2)
func (d *Document) Keys(ctx context.Context, path Path) ([]string, error) {
	return nil, &NotImplementedError{
		Feature:   "Keys",
		Milestone: "M2",
		Message:   "Requires am_keys WASI export",
	}
}

// Length returns the number of keys in a map (or elements in a list/text).
//
// Status: ⚠️ Partially implemented (only for text via TextLength)
func (d *Document) Length(ctx context.Context, path Path) (uint, error) {
	// Check if it's the content text path
	if d.isContentPath(path) {
		len, err := d.TextLength(ctx, path)
		return uint(len), err
	}

	return 0, &NotImplementedError{
		Feature:   "Length",
		Milestone: "M2",
		Message:   "Only text length is supported currently",
	}
}
