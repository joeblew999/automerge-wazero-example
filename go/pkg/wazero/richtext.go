package wazero

import (
	"context"
	"fmt"
)

// Rich Text (Marks) - maps to rust/automerge_wasi/src/richtext.rs

// AmMark adds a mark (formatting) to a range of text
func (r *Runtime) AmMark(ctx context.Context, name, value string, start, end uint, expand uint8) error {
	nameBytes := []byte(name)
	valueBytes := []byte(value)

	namePtr := uint32(1)
	valuePtr := uint32(1)
	mem := r.Memory()

	// Allocate and write name
	if len(nameBytes) > 0 {
		var err error
		namePtr, err = r.AmAlloc(ctx, uint32(len(nameBytes)))
		if err != nil {
			return fmt.Errorf("failed to allocate name: %w", err)
		}
		defer r.AmFree(ctx, namePtr, uint32(len(nameBytes)))

		if !mem.Write(namePtr, nameBytes) {
			return fmt.Errorf("failed to write name to WASM memory")
		}
	}

	// Allocate and write value
	if len(valueBytes) > 0 {
		var err error
		valuePtr, err = r.AmAlloc(ctx, uint32(len(valueBytes)))
		if err != nil {
			return fmt.Errorf("failed to allocate value: %w", err)
		}
		defer r.AmFree(ctx, valuePtr, uint32(len(valueBytes)))

		if !mem.Write(valuePtr, valueBytes) {
			return fmt.Errorf("failed to write value to WASM memory")
		}
	}

	// Call am_mark
	results, err := r.callExport(ctx, "am_mark",
		uint64(namePtr), uint64(len(nameBytes)),
		uint64(valuePtr), uint64(len(valueBytes)),
		uint64(start), uint64(end), uint64(expand))
	if err != nil {
		return err
	}
	return checkErrorCode("am_mark", results)
}

// AmUnmark removes a mark (formatting) from a range of text
func (r *Runtime) AmUnmark(ctx context.Context, name string, start, end uint, expand uint8) error {
	nameBytes := []byte(name)
	namePtr := uint32(1)
	mem := r.Memory()

	if len(nameBytes) > 0 {
		var err error
		namePtr, err = r.AmAlloc(ctx, uint32(len(nameBytes)))
		if err != nil {
			return fmt.Errorf("failed to allocate name: %w", err)
		}
		defer r.AmFree(ctx, namePtr, uint32(len(nameBytes)))

		if !mem.Write(namePtr, nameBytes) {
			return fmt.Errorf("failed to write name to WASM memory")
		}
	}

	// Call am_unmark
	results, err := r.callExport(ctx, "am_unmark",
		uint64(namePtr), uint64(len(nameBytes)),
		uint64(start), uint64(end), uint64(expand))
	if err != nil {
		return err
	}
	return checkErrorCode("am_unmark", results)
}

// AmGetMarksCount returns the number of marks at a specific index
func (r *Runtime) AmGetMarksCount(ctx context.Context, index uint) (uint32, error) {
	results, err := r.callExport(ctx, "am_get_marks_count", uint64(index))
	if err != nil {
		return 0, err
	}
	return uint32(results[0]), nil
}

// AmMarksLen returns the length of the marks JSON string
func (r *Runtime) AmMarksLen(ctx context.Context) (uint32, error) {
	results, err := r.callExport(ctx, "am_marks_len")
	if err != nil {
		return 0, err
	}
	return uint32(results[0]), nil
}

// AmMarks retrieves all marks in the text object as JSON
func (r *Runtime) AmMarks(ctx context.Context) (string, error) {
	// Get marks length
	marksLen, err := r.AmMarksLen(ctx)
	if err != nil {
		return "", err
	}
	if marksLen == 0 {
		return "[]", nil
	}

	// Allocate buffer
	marksPtr, err := r.AmAlloc(ctx, marksLen)
	if err != nil {
		return "", fmt.Errorf("failed to allocate marks buffer: %w", err)
	}
	defer r.AmFree(ctx, marksPtr, marksLen)

	// Get marks
	results, err := r.callExport(ctx, "am_marks", uint64(marksPtr))
	if err != nil {
		return "", err
	}
	if err := checkErrorCode("am_marks", results); err != nil {
		return "", err
	}

	// Read marks from memory
	mem := r.Memory()
	marksBytes, ok := mem.Read(marksPtr, marksLen)
	if !ok {
		return "", fmt.Errorf("failed to read marks from WASM memory")
	}

	// Trim null bytes (Rust may over-allocate buffer)
	// Find the first null byte and truncate there
	for i, b := range marksBytes {
		if b == 0 {
			marksBytes = marksBytes[:i]
			break
		}
	}

	return string(marksBytes), nil
}
