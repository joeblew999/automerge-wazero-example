package wazero

import (
	"context"
	"fmt"
)

// GetCursor gets a cursor for a position in a text or list object.
// The cursor remains stable across concurrent edits.
//
// Returns the cursor string and error.
// The cursor can be used with LookupCursor to find its current position.
func (r *Runtime) GetCursor(ctx context.Context, path string, index int) (string, error) {
	pathBytes := []byte(path)
	pathLen := uint32(len(pathBytes))

	// Allocate memory for path
	pathPtr, err := r.AmAlloc(ctx, pathLen)
	if err != nil {
		return "", fmt.Errorf("failed to alloc path: %w", err)
	}
	defer r.AmFree(ctx, pathPtr, pathLen)

	// Write path to memory
	mem := r.Memory()
	if !mem.Write(pathPtr, pathBytes) {
		return "", fmt.Errorf("failed to write path to memory")
	}

	// Call am_get_cursor to get cursor length
	results, err := r.callExport(ctx, "am_get_cursor",
		uint64(pathPtr),
		uint64(pathLen),
		uint64(index),
	)
	if err != nil {
		return "", fmt.Errorf("am_get_cursor failed: %w", err)
	}

	cursorLen := int32(results[0])
	if cursorLen < 0 {
		switch cursorLen {
		case -1:
			return "", fmt.Errorf("invalid path: %s", path)
		case -2:
			return "", fmt.Errorf("invalid index: %d", index)
		case -3:
			return "", fmt.Errorf("not a text or list object")
		default:
			return "", fmt.Errorf("am_get_cursor failed with code %d", cursorLen)
		}
	}

	// Allocate buffer for cursor string
	cursorPtr, err := r.AmAlloc(ctx, uint32(cursorLen))
	if err != nil {
		return "", fmt.Errorf("failed to alloc cursor buffer: %w", err)
	}
	defer r.AmFree(ctx, cursorPtr, uint32(cursorLen))

	// Call am_get_cursor_str to retrieve cursor string
	results, err = r.callExport(ctx, "am_get_cursor_str", uint64(cursorPtr))
	if err != nil {
		return "", fmt.Errorf("am_get_cursor_str failed: %w", err)
	}

	if int32(results[0]) != 0 {
		return "", fmt.Errorf("am_get_cursor_str failed with code %d", results[0])
	}

	// Read cursor string from memory
	cursorBytes, ok := mem.Read(cursorPtr, uint32(cursorLen))
	if !ok {
		return "", fmt.Errorf("failed to read cursor from memory")
	}

	return string(cursorBytes), nil
}

// LookupCursor looks up the current index for a cursor.
// Returns the current position of the cursor in the object.
// Cursors track positions that remain stable across concurrent edits.
func (r *Runtime) LookupCursor(ctx context.Context, path string, cursor string) (int, error) {
	pathBytes := []byte(path)
	pathLen := uint32(len(pathBytes))
	cursorBytes := []byte(cursor)
	cursorLen := uint32(len(cursorBytes))

	// Allocate memory for path
	pathPtr, err := r.AmAlloc(ctx, pathLen)
	if err != nil {
		return 0, fmt.Errorf("failed to alloc path: %w", err)
	}
	defer r.AmFree(ctx, pathPtr, pathLen)

	// Write path to memory
	mem := r.Memory()
	if !mem.Write(pathPtr, pathBytes) {
		return 0, fmt.Errorf("failed to write path to memory")
	}

	// Allocate memory for cursor
	cursorPtr, err := r.AmAlloc(ctx, cursorLen)
	if err != nil {
		return 0, fmt.Errorf("failed to alloc cursor: %w", err)
	}
	defer r.AmFree(ctx, cursorPtr, cursorLen)

	// Write cursor to memory
	if !mem.Write(cursorPtr, cursorBytes) {
		return 0, fmt.Errorf("failed to write cursor to memory")
	}

	// Call am_lookup_cursor
	results, err := r.callExport(ctx, "am_lookup_cursor",
		uint64(pathPtr),
		uint64(pathLen),
		uint64(cursorPtr),
		uint64(cursorLen),
	)
	if err != nil {
		return 0, fmt.Errorf("am_lookup_cursor failed: %w", err)
	}

	index := int32(results[0])
	if index < 0 {
		switch index {
		case -1:
			return 0, fmt.Errorf("invalid path: %s", path)
		case -2:
			return 0, fmt.Errorf("invalid cursor: %s", cursor)
		case -3:
			return 0, fmt.Errorf("cursor not found in object")
		default:
			return 0, fmt.Errorf("am_lookup_cursor failed with code %d", index)
		}
	}

	return int(index), nil
}
