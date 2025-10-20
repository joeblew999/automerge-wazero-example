package automerge

import "context"

// List Operations
//
// Currently supports string values in a global list at ROOT["list_items"].
// Future: Support multiple lists via object IDs and other value types.

// ListPush appends a value to the end of a list.
//
// Status: ✅ Implemented for global list with string values
func (d *Document) ListPush(ctx context.Context, path Path, value Value) error {
	// For now, only support global list (empty path means ROOT["list_items"])
	if path.Len() != 0 {
		return &NotImplementedError{
			Feature:   "ListPush (custom paths)",
			Milestone: "M2",
			Message:   "Only global list at ROOT[\"list_items\"] is supported currently",
		}
	}

	str, ok := value.AsString()
	if !ok {
		return &NotImplementedError{
			Feature:   "ListPush (non-string values)",
			Milestone: "M2",
			Message:   "Only string values are supported currently",
		}
	}

	return d.runtime.AmListPush(ctx, str)
}

// ListInsert inserts a value at a specific index in a list.
//
// Status: ✅ Implemented for global list with string values
func (d *Document) ListInsert(ctx context.Context, path Path, index uint, value Value) error {
	if path.Len() != 0 {
		return &NotImplementedError{
			Feature:   "ListInsert (custom paths)",
			Milestone: "M2",
			Message:   "Only global list is supported currently",
		}
	}

	str, ok := value.AsString()
	if !ok {
		return &NotImplementedError{
			Feature:   "ListInsert (non-string values)",
			Milestone: "M2",
			Message:   "Only string values are supported currently",
		}
	}

	return d.runtime.AmListInsert(ctx, index, str)
}

// ListGet retrieves a value at a specific index in a list.
//
// Status: ✅ Implemented for global list
func (d *Document) ListGet(ctx context.Context, path Path, index uint) (Value, error) {
	if path.Len() != 0 {
		return Value{}, &NotImplementedError{
			Feature:   "ListGet (custom paths)",
			Milestone: "M2",
			Message:   "Only global list is supported currently",
		}
	}

	valueStr, err := d.runtime.AmListGet(ctx, index)
	if err != nil {
		return Value{}, err
	}

	return NewString(valueStr), nil
}

// ListDelete removes a value at a specific index from a list.
//
// Status: ✅ Implemented for global list
func (d *Document) ListDelete(ctx context.Context, path Path, index uint) error {
	if path.Len() != 0 {
		return &NotImplementedError{
			Feature:   "ListDelete (custom paths)",
			Milestone: "M2",
			Message:   "Only global list is supported currently",
		}
	}

	return d.runtime.AmListDelete(ctx, index)
}

// ListLength returns the number of elements in a list.
//
// Status: ✅ Implemented for global list
func (d *Document) ListLength(ctx context.Context, path Path) (uint, error) {
	if path.Len() != 0 {
		return 0, &NotImplementedError{
			Feature:   "ListLength (custom paths)",
			Milestone: "M2",
			Message:   "Only global list is supported currently",
		}
	}

	len, err := d.runtime.AmListLen(ctx)
	return uint(len), err
}
