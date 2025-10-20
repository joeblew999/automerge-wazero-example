package automerge

import "context"

// Rich Text Operations - For M4 Milestone
//
// Marks allow you to add formatting (bold, italic, links, etc.) to text
// ranges. Marks are CRDT-aware and merge correctly when users concurrently
// format the same text.

// Mark adds formatting to a range of text.
//
// Example:
//   // Make characters 0-5 bold
//   doc.Mark(ctx, path, Mark{
//       Name: "bold",
//       Value: NewBool(true),
//       Start: 0,
//       End: 5,
//   }, ExpandBoth)
//
// Status: ❌ Not implemented (requires M4)
func (d *Document) Mark(ctx context.Context, path Path, mark Mark, expand ExpandMark) error {
	return &NotImplementedError{
		Feature:   "Mark",
		Milestone: "M4",
		Message:   "Rich text requires am_mark WASI export",
	}
}

// Unmark removes formatting from a range of text.
//
// Status: ❌ Not implemented (requires M4)
func (d *Document) Unmark(ctx context.Context, path Path, name string, start, end uint, expand ExpandMark) error {
	return &NotImplementedError{
		Feature:   "Unmark",
		Milestone: "M4",
		Message:   "Requires am_unmark WASI export",
	}
}

// GetMarks retrieves all marks at a specific position.
//
// Status: ❌ Not implemented (requires M4)
func (d *Document) GetMarks(ctx context.Context, path Path, index uint) (MarkSet, error) {
	return nil, &NotImplementedError{
		Feature:   "GetMarks",
		Milestone: "M4",
		Message:   "Requires am_get_marks WASI export",
	}
}

// Marks retrieves all marks in the text object.
//
// Status: ❌ Not implemented (requires M4)
func (d *Document) Marks(ctx context.Context, path Path) ([]Mark, error) {
	return nil, &NotImplementedError{
		Feature:   "Marks",
		Milestone: "M4",
		Message:   "Requires am_marks WASI export",
	}
}

// SplitBlock inserts a block marker (e.g., paragraph break) at an index.
//
// Returns the path to the newly created block marker object.
//
// Status: ❌ Not implemented (requires M4)
func (d *Document) SplitBlock(ctx context.Context, path Path, index uint) (Path, error) {
	return Path{}, &NotImplementedError{
		Feature:   "SplitBlock",
		Milestone: "M4",
		Message:   "Requires am_split_block WASI export",
	}
}

// JoinBlock removes a block marker at an index.
//
// Status: ❌ Not implemented (requires M4)
func (d *Document) JoinBlock(ctx context.Context, path Path, index uint) error {
	return &NotImplementedError{
		Feature:   "JoinBlock",
		Milestone: "M4",
		Message:   "Requires am_join_block WASI export",
	}
}

// ExpandMark controls how marks expand when text is inserted at boundaries
type ExpandMark int

const (
	ExpandNone   ExpandMark = 0 // Don't expand
	ExpandBefore ExpandMark = 1 // Expand when inserting before
	ExpandAfter  ExpandMark = 2 // Expand when inserting after
	ExpandBoth   ExpandMark = 3 // Expand in both directions
)
