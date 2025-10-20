package automerge

import "context"

// History and Time-Travel Operations - For Future Milestones
//
// Automerge preserves the complete history of all changes. You can query
// the history, fork at specific points, and read historical values.

// GetHeads returns the current heads (frontier) of the document.
//
// Heads identify the current state of the document. After merging, a document
// may have multiple heads temporarily until the next change.
//
// Status: ❌ Not implemented
func (d *Document) GetHeads(ctx context.Context) ([]ChangeHash, error) {
	return nil, &NotImplementedError{
		Feature:   "GetHeads",
		Milestone: "",
		Message:   "Requires am_get_heads WASI export",
	}
}

// GetChanges returns all changes since the given dependencies.
//
// Used internally by the sync protocol. Useful for debugging or building
// custom sync mechanisms.
//
// Status: ❌ Not implemented
func (d *Document) GetChanges(ctx context.Context, have []ChangeHash) ([]Change, error) {
	return nil, &NotImplementedError{
		Feature:   "GetChanges",
		Milestone: "",
		Message:   "Requires am_get_changes WASI export",
	}
}

// GetChangeByHash retrieves a specific change by its hash.
//
// Status: ❌ Not implemented
func (d *Document) GetChangeByHash(ctx context.Context, hash ChangeHash) (*Change, error) {
	return nil, &NotImplementedError{
		Feature:   "GetChangeByHash",
		Milestone: "",
		Message:   "Requires am_get_change_by_hash WASI export",
	}
}

// Fork creates a copy of the document at the current state.
//
// Changes to the fork don't affect the original, but they can be merged later.
//
// Status: ❌ Not implemented
func (d *Document) Fork(ctx context.Context) (*Document, error) {
	return nil, &NotImplementedError{
		Feature:   "Fork",
		Milestone: "",
		Message:   "Requires am_fork WASI export",
	}
}

// ForkAt creates a copy of the document at a specific point in history.
//
// Useful for time-travel debugging or exploring alternative histories.
//
// Status: ❌ Not implemented
func (d *Document) ForkAt(ctx context.Context, heads []ChangeHash) (*Document, error) {
	return nil, &NotImplementedError{
		Feature:   "ForkAt",
		Milestone: "",
		Message:   "Requires am_fork_at WASI export",
	}
}

// GetTextAt retrieves text content at a specific point in history.
//
// Status: ❌ Not implemented
func (d *Document) GetTextAt(ctx context.Context, path Path, heads []ChangeHash) (string, error) {
	return "", &NotImplementedError{
		Feature:   "GetTextAt",
		Milestone: "",
		Message:   "Requires am_get_text_at WASI export",
	}
}

// GetAt retrieves a value at a specific point in history.
//
// Status: ❌ Not implemented
func (d *Document) GetAt(ctx context.Context, path Path, key string, heads []ChangeHash) (Value, error) {
	return Value{}, &NotImplementedError{
		Feature:   "GetAt",
		Milestone: "",
		Message:   "Requires am_get_at WASI export",
	}
}
