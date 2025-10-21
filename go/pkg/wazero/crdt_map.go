package wazero

import (
	"context"
	"fmt"
)

// Map Operations - maps to rust/automerge_wasi/src/map.rs

// AmMapSet sets a string value in the ROOT map
func (r *Runtime) AmMapSet(ctx context.Context, key, value string) error {
	keyBytes := []byte(key)
	valueBytes := []byte(value)

	// Handle empty strings - use pointer 1 (not null) but length 0
	keyPtr := uint32(1)
	valuePtr := uint32(1)

	mem := r.Memory()

	// Allocate and write key if non-empty
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

	// Allocate and write value if non-empty
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

	// Call am_map_set
	results, err := r.callExport(ctx, "am_map_set",
		uint64(keyPtr), uint64(len(keyBytes)),
		uint64(valuePtr), uint64(len(valueBytes)))
	if err != nil {
		return err
	}

	return checkErrorCode("am_map_set", results)
}

// AmMapGet retrieves a string value from the ROOT map
func (r *Runtime) AmMapGet(ctx context.Context, key string) (string, error) {
	keyBytes := []byte(key)
	mem := r.Memory()

	// Handle empty key
	keyPtr := uint32(1)
	if len(keyBytes) > 0 {
		var err error
		keyPtr, err = r.AmAlloc(ctx, uint32(len(keyBytes)))
		if err != nil {
			return "", fmt.Errorf("failed to allocate key: %w", err)
		}
		defer r.AmFree(ctx, keyPtr, uint32(len(keyBytes)))

		if !mem.Write(keyPtr, keyBytes) {
			return "", fmt.Errorf("failed to write key to WASM memory")
		}
	}

	// Get value length
	results, err := r.callExport(ctx, "am_map_get_len", uint64(keyPtr), uint64(len(keyBytes)))
	if err != nil {
		return "", err
	}
	valueLen := uint32(results[0])

	// Handle empty value (which is valid)
	if valueLen == 0 {
		// Could be empty string or key not found - need to check by trying to get it
		valuePtr := uint32(1)
		results, err = r.callExport(ctx, "am_map_get", uint64(keyPtr), uint64(len(keyBytes)), uint64(valuePtr))
		if err != nil {
			return "", err
		}
		if err := checkErrorCode("am_map_get", results); err != nil {
			return "", err
		}
		return "", nil // Empty string value
	}

	// Allocate memory for value
	valuePtr, err := r.AmAlloc(ctx, valueLen)
	if err != nil {
		return "", fmt.Errorf("failed to allocate value buffer: %w", err)
	}
	defer r.AmFree(ctx, valuePtr, valueLen)

	// Get value
	results, err = r.callExport(ctx, "am_map_get", uint64(keyPtr), uint64(len(keyBytes)), uint64(valuePtr))
	if err != nil {
		return "", err
	}
	if err := checkErrorCode("am_map_get", results); err != nil {
		return "", err
	}

	// Read value from memory
	valueBytes, ok := mem.Read(valuePtr, valueLen)
	if !ok {
		return "", fmt.Errorf("failed to read value from WASM memory")
	}

	return string(valueBytes), nil
}

// AmMapDelete deletes a key from the ROOT map
func (r *Runtime) AmMapDelete(ctx context.Context, key string) error {
	keyBytes := []byte(key)

	// Allocate memory for key
	keyPtr, err := r.AmAlloc(ctx, uint32(len(keyBytes)))
	if err != nil {
		return fmt.Errorf("failed to allocate key: %w", err)
	}
	defer r.AmFree(ctx, keyPtr, uint32(len(keyBytes)))

	// Write key to memory
	mem := r.Memory()
	if !mem.Write(keyPtr, keyBytes) {
		return fmt.Errorf("failed to write key to WASM memory")
	}

	// Call am_map_delete
	results, err := r.callExport(ctx, "am_map_delete", uint64(keyPtr), uint64(len(keyBytes)))
	if err != nil {
		return err
	}

	return checkErrorCode("am_map_delete", results)
}

// AmMapLen returns the number of keys in the ROOT map
func (r *Runtime) AmMapLen(ctx context.Context) (uint32, error) {
	results, err := r.callExport(ctx, "am_map_len")
	if err != nil {
		return 0, err
	}
	return uint32(results[0]), nil
}

// AmMapKeys returns all keys in the ROOT map
func (r *Runtime) AmMapKeys(ctx context.Context) ([]string, error) {
	// Get total size needed
	results, err := r.callExport(ctx, "am_map_keys_total_size")
	if err != nil {
		return nil, err
	}
	totalSize := uint32(results[0])
	if totalSize == 0 {
		return []string{}, nil
	}

	// Allocate memory for keys
	ptr, err := r.AmAlloc(ctx, totalSize)
	if err != nil {
		return nil, fmt.Errorf("failed to allocate keys buffer: %w", err)
	}
	defer r.AmFree(ctx, ptr, totalSize)

	// Get keys
	results, err = r.callExport(ctx, "am_map_keys", uint64(ptr))
	if err != nil {
		return nil, err
	}
	if err := checkErrorCode("am_map_keys", results); err != nil {
		return nil, err
	}

	// Read keys from memory
	mem := r.Memory()
	keysBytes, ok := mem.Read(ptr, totalSize)
	if !ok {
		return nil, fmt.Errorf("failed to read keys from WASM memory")
	}

	// Parse null-terminated strings
	var keys []string
	start := 0
	for i := 0; i < len(keysBytes); i++ {
		if keysBytes[i] == 0 {
			if i > start {
				keys = append(keys, string(keysBytes[start:i]))
			}
			start = i + 1
		}
	}

	return keys, nil
}
