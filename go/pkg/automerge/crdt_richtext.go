package automerge

import (
	"context"
	"encoding/json"
	"fmt"
)

// Rich Text Operations - M4 Milestone
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
// Status: ✅ Implemented (current implementation uses ROOT["content"] text object)
func (d *Document) Mark(ctx context.Context, path Path, mark Mark, expand ExpandMark) error {
	if d.runtime == nil {
		return fmt.Errorf("document not initialized")
	}

	// Convert Value to string
	var valueStr string
	if s, ok := mark.Value.AsString(); ok {
		valueStr = s
	} else if b, ok := mark.Value.AsBool(); ok {
		if b {
			valueStr = "true"
		} else {
			valueStr = "false"
		}
	} else if i, ok := mark.Value.AsInt(); ok {
		valueStr = fmt.Sprintf("%d", i)
	} else if f, ok := mark.Value.AsFloat(); ok {
		valueStr = fmt.Sprintf("%f", f)
	} else {
		valueStr = fmt.Sprintf("%v", mark.Value)
	}

	return d.runtime.AmMark(ctx, mark.Name, valueStr, mark.Start, mark.End, uint8(expand))
}

// Unmark removes formatting from a range of text.
//
// Status: ✅ Implemented
func (d *Document) Unmark(ctx context.Context, path Path, name string, start, end uint, expand ExpandMark) error {
	if d.runtime == nil {
		return fmt.Errorf("document not initialized")
	}

	return d.runtime.AmUnmark(ctx, name, start, end, uint8(expand))
}

// GetMarks retrieves all marks at a specific position.
//
// Status: ✅ Implemented
func (d *Document) GetMarks(ctx context.Context, path Path, index uint) ([]Mark, error) {
	if d.runtime == nil {
		return nil, fmt.Errorf("document not initialized")
	}

	count, err := d.runtime.AmGetMarksCount(ctx, index)
	if err != nil {
		return nil, fmt.Errorf("failed to get marks count: %w", err)
	}

	if count == 0 {
		return []Mark{}, nil
	}

	// Get all marks and filter for this index
	allMarks, err := d.Marks(ctx, path)
	if err != nil {
		return nil, err
	}

	// Filter marks that apply at this index
	var marks []Mark
	for _, mark := range allMarks {
		if mark.Start <= index && index < mark.End {
			marks = append(marks, mark)
		}
	}

	return marks, nil
}

// Marks retrieves all marks in the text object.
//
// Status: ✅ Implemented
func (d *Document) Marks(ctx context.Context, path Path) ([]Mark, error) {
	if d.runtime == nil {
		return nil, fmt.Errorf("document not initialized")
	}

	marksJSON, err := d.runtime.AmMarks(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get marks: %w", err)
	}

	if marksJSON == "[]" {
		return []Mark{}, nil
	}

	// Parse JSON
	var rawMarks []struct {
		Name  string `json:"name"`
		Value string `json:"value"`
		Start uint   `json:"start"`
		End   uint   `json:"end"`
	}

	if err := json.Unmarshal([]byte(marksJSON), &rawMarks); err != nil {
		return nil, fmt.Errorf("failed to parse marks JSON: %w", err)
	}

	// Convert to Mark structs
	marks := make([]Mark, len(rawMarks))
	for i, raw := range rawMarks {
		marks[i] = Mark{
			Name:  raw.Name,
			Value: NewString(raw.Value), // Convert string to Value
			Start: raw.Start,
			End:   raw.End,
		}
	}

	return marks, nil
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
