package wazero

import (
	"context"
	"fmt"
)

// Memory Management - maps to rust/automerge_wasi/src/memory.rs

// AmAlloc allocates memory in WASM linear memory
func (r *Runtime) AmAlloc(ctx context.Context, size uint32) (uint32, error) {
	results, err := r.callExport(ctx, "am_alloc", uint64(size))
	if err != nil {
		return 0, err
	}

	ptr := uint32(results[0])
	if ptr == 0 {
		return 0, fmt.Errorf("am_alloc returned null pointer")
	}

	return ptr, nil
}

// AmFree frees memory allocated by AmAlloc
func (r *Runtime) AmFree(ctx context.Context, ptr uint32, size uint32) error {
	_, err := r.callExport(ctx, "am_free", uint64(ptr), uint64(size))
	return err
}
