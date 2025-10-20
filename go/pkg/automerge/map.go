package automerge

import "context"

// Map Operations
//
// Currently supports string values in the ROOT map.
// Future: Support nested maps, lists, and other value types.

// Get retrieves a string value at a key in the ROOT map.
//
// Status: ✅ Implemented for ROOT map with string values
//
// Note: Currently only supports ROOT map. For nested maps, use PutObject
// to create nested maps first (requires M2).
func (d *Document) Get(ctx context.Context, path Path, key string) (Value, error) {
	// For now, only support ROOT map
	if !d.isRootPath(path) {
		return Value{}, &NotImplementedError{
			Feature:   "Get (nested maps)",
			Milestone: "M2",
			Message:   "Only ROOT map is supported currently",
		}
	}

	valueStr, err := d.runtime.AmMapGet(ctx, key)
	if err != nil {
		return Value{}, err
	}

	return NewString(valueStr), nil
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

// Put sets a string value at a key in the ROOT map.
//
// Status: ✅ Implemented for ROOT map with string values
//
// Note: Currently only supports string values. For other types, use specialized
// methods like PutObject, Increment (for counters), etc.
func (d *Document) Put(ctx context.Context, path Path, key string, value Value) error {
	// For now, only support ROOT map
	if !d.isRootPath(path) {
		return &NotImplementedError{
			Feature:   "Put (nested maps)",
			Milestone: "M2",
			Message:   "Only ROOT map is supported currently",
		}
	}

	// Extract string value
	str, ok := value.AsString()
	if !ok {
		return &NotImplementedError{
			Feature:   "Put (non-string values)",
			Milestone: "M2",
			Message:   "Only string values are supported currently",
		}
	}

	return d.runtime.AmMapSet(ctx, key, str)
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

// Delete removes a key from the ROOT map.
//
// Status: ✅ Implemented for ROOT map
func (d *Document) Delete(ctx context.Context, path Path, key string) error {
	// For now, only support ROOT map
	if !d.isRootPath(path) {
		return &NotImplementedError{
			Feature:   "Delete (nested maps)",
			Milestone: "M2",
			Message:   "Only ROOT map is supported currently",
		}
	}

	return d.runtime.AmMapDelete(ctx, key)
}

// Keys returns all keys in the ROOT map.
//
// Status: ✅ Implemented for ROOT map
func (d *Document) Keys(ctx context.Context, path Path) ([]string, error) {
	// For now, only support ROOT map
	if !d.isRootPath(path) {
		return nil, &NotImplementedError{
			Feature:   "Keys (nested maps)",
			Milestone: "M2",
			Message:   "Only ROOT map is supported currently",
		}
	}

	return d.runtime.AmMapKeys(ctx)
}

// Length returns the number of keys in a map (or elements in a list/text).
//
// Status: ✅ Implemented for ROOT map and text objects
func (d *Document) Length(ctx context.Context, path Path) (uint, error) {
	// Check if it's the content text path
	if d.isContentPath(path) {
		len, err := d.TextLength(ctx, path)
		return uint(len), err
	}

	// Check if it's ROOT map
	if d.isRootPath(path) {
		len, err := d.runtime.AmMapLen(ctx)
		return uint(len), err
	}

	return 0, &NotImplementedError{
		Feature:   "Length (nested objects)",
		Milestone: "M2",
		Message:   "Only ROOT map and text length are supported currently",
	}
}

// Helper: Check if path refers to ROOT map
func (d *Document) isRootPath(path Path) bool {
	return path.Len() == 0
}
