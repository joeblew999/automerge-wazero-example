package wazero

import (
	"context"
	"fmt"
)

// Counter Operations - maps to rust/automerge_wasi/src/counter.rs

// AmCounterCreate creates a new counter with an initial value
func (r *Runtime) AmCounterCreate(ctx context.Context, key string, value int64) error {
	keyBytes := []byte(key)
	keyPtr := uint32(1)
	mem := r.Memory()

	if len(keyBytes) > 0 {
		var err error
		keyPtr, err = r.AmAlloc(ctx, uint32(len(keyBytes)))
		if err != nil {
			return fmt.Errorf("failed to allocate key: %w", err)
		}
		defer r.AmFree(ctx, keyPtr, uint32(len(keyBytes)))

		if !mem.Write(keyPtr, keyBytes) {
			return fmt.Errorf("failed to write key to WASM memory")
		}
	}

	results, err := r.callExport(ctx, "am_counter_create", uint64(keyPtr), uint64(len(keyBytes)), uint64(value))
	if err != nil {
		return err
	}
	return checkErrorCode("am_counter_create", results)
}

// AmCounterIncrement increments (or decrements if negative) a counter
func (r *Runtime) AmCounterIncrement(ctx context.Context, key string, delta int64) error {
	keyBytes := []byte(key)
	keyPtr := uint32(1)
	mem := r.Memory()

	if len(keyBytes) > 0 {
		var err error
		keyPtr, err = r.AmAlloc(ctx, uint32(len(keyBytes)))
		if err != nil {
			return fmt.Errorf("failed to allocate key: %w", err)
		}
		defer r.AmFree(ctx, keyPtr, uint32(len(keyBytes)))

		if !mem.Write(keyPtr, keyBytes) {
			return fmt.Errorf("failed to write key to WASM memory")
		}
	}

	results, err := r.callExport(ctx, "am_counter_increment", uint64(keyPtr), uint64(len(keyBytes)), uint64(delta))
	if err != nil {
		return err
	}
	return checkErrorCode("am_counter_increment", results)
}

// AmCounterGet retrieves the current value of a counter
func (r *Runtime) AmCounterGet(ctx context.Context, key string) (int64, error) {
	keyBytes := []byte(key)
	keyPtr := uint32(1)
	mem := r.Memory()

	if len(keyBytes) > 0 {
		var err error
		keyPtr, err = r.AmAlloc(ctx, uint32(len(keyBytes)))
		if err != nil {
			return 0, fmt.Errorf("failed to allocate key: %w", err)
		}
		defer r.AmFree(ctx, keyPtr, uint32(len(keyBytes)))

		if !mem.Write(keyPtr, keyBytes) {
			return 0, fmt.Errorf("failed to write key to WASM memory")
		}
	}

	// Allocate space for return value
	valuePtr, err := r.AmAlloc(ctx, 8) // i64 = 8 bytes
	if err != nil {
		return 0, fmt.Errorf("failed to allocate value buffer: %w", err)
	}
	defer r.AmFree(ctx, valuePtr, 8)

	results, err := r.callExport(ctx, "am_counter_get", uint64(keyPtr), uint64(len(keyBytes)), uint64(valuePtr))
	if err != nil {
		return 0, err
	}
	if err := checkErrorCode("am_counter_get", results); err != nil {
		return 0, err
	}

	// Read i64 from memory
	valueBytes, ok := mem.Read(valuePtr, 8)
	if !ok {
		return 0, fmt.Errorf("failed to read value from WASM memory")
	}

	// Convert bytes to i64 (little-endian)
	value := int64(valueBytes[0]) | int64(valueBytes[1])<<8 | int64(valueBytes[2])<<16 | int64(valueBytes[3])<<24 |
		int64(valueBytes[4])<<32 | int64(valueBytes[5])<<40 | int64(valueBytes[6])<<48 | int64(valueBytes[7])<<56

	return value, nil
}
