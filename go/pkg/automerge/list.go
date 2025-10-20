package automerge

import "context"

// List Operations - For M2 Milestone
//
// Lists in Automerge are CRDT-aware sequences that support concurrent
// insertions and deletions without conflicts.

// Insert inserts a value at an index in a list object.
//
// Status: ❌ Not implemented (requires M2)
func (d *Document) Insert(ctx context.Context, path Path, index uint, value Value) error {
	return &NotImplementedError{
		Feature:   "Insert",
		Milestone: "M2",
		Message:   "Requires am_insert WASI export",
	}
}

// InsertObject inserts a new object (Map, List, or Text) at an index in a list.
//
// Returns the path to the newly created object.
//
// Status: ❌ Not implemented (requires M2)
func (d *Document) InsertObject(ctx context.Context, path Path, index uint, objType ObjType) (Path, error) {
	return Path{}, &NotImplementedError{
		Feature:   "InsertObject",
		Milestone: "M2",
		Message:   "Requires am_insert_object WASI export",
	}
}

// Remove removes an element at an index in a list.
//
// Note: This is equivalent to Delete for lists.
//
// Status: ❌ Not implemented (requires M2)
func (d *Document) Remove(ctx context.Context, path Path, index uint) error {
	return &NotImplementedError{
		Feature:   "Remove",
		Milestone: "M2",
		Message:   "Requires am_delete WASI export (works for lists too)",
	}
}

// Splice replaces a range of list elements with new values.
//
// Similar to SpliceText but for generic values instead of text.
//
// Status: ❌ Not implemented (requires M2)
func (d *Document) Splice(ctx context.Context, path Path, pos uint, del int, values []Value) error {
	return &NotImplementedError{
		Feature:   "Splice",
		Milestone: "M2",
		Message:   "Requires am_splice WASI export",
	}
}
