package wazero

import (
	"context"
	"fmt"
)

// Generic Object Operations - maps to rust/automerge_wasi/src/generic.rs

// AmPutRoot puts a scalar value at ROOT level
func (r *Runtime) AmPutRoot(ctx context.Context, key string, value string) error {
	keyBytes := []byte(key)
	keyLen := uint32(len(keyBytes))
	valueBytes := []byte(value)
	valueLen := uint32(len(valueBytes))

	// Allocate memory for key
	keyPtr, err := r.AmAlloc(ctx, keyLen)
	if err != nil {
		return fmt.Errorf("failed to allocate key: %w", err)
	}
	defer r.AmFree(ctx, keyPtr, keyLen)

	// Allocate memory for value
	valuePtr, err := r.AmAlloc(ctx, valueLen)
	if err != nil {
		return fmt.Errorf("failed to allocate value: %w", err)
	}
	defer r.AmFree(ctx, valuePtr, valueLen)

	// Write to memory
	mem := r.Memory()
	if !mem.Write(keyPtr, keyBytes) {
		return fmt.Errorf("failed to write key to memory")
	}
	if !mem.Write(valuePtr, valueBytes) {
		return fmt.Errorf("failed to write value to memory")
	}

	// Call am_put_root
	results, err := r.callExport(ctx, "am_put_root",
		uint64(keyPtr), uint64(keyLen),
		uint64(valuePtr), uint64(valueLen),
	)
	if err != nil {
		return err
	}

	return checkErrorCode("am_put_root", results)
}

// AmGetRoot gets a value from ROOT level
func (r *Runtime) AmGetRoot(ctx context.Context, key string) (string, error) {
	keyBytes := []byte(key)
	keyLen := uint32(len(keyBytes))

	// Allocate memory for key
	keyPtr, err := r.AmAlloc(ctx, keyLen)
	if err != nil {
		return "", fmt.Errorf("failed to allocate key: %w", err)
	}
	defer r.AmFree(ctx, keyPtr, keyLen)

	// Write key to memory
	mem := r.Memory()
	if !mem.Write(keyPtr, keyBytes) {
		return "", fmt.Errorf("failed to write key to memory")
	}

	// Call am_get_root to get value length
	results, err := r.callExport(ctx, "am_get_root", uint64(keyPtr), uint64(keyLen))
	if err != nil {
		return "", err
	}

	valueLen := int32(results[0])
	if valueLen < 0 {
		return "", fmt.Errorf("am_get_root failed with code %d", valueLen)
	}

	// Allocate buffer for value
	valuePtr, err := r.AmAlloc(ctx, uint32(valueLen))
	if err != nil {
		return "", fmt.Errorf("failed to allocate value buffer: %w", err)
	}
	defer r.AmFree(ctx, valuePtr, uint32(valueLen))

	// Call am_get_root_value to retrieve value
	results, err = r.callExport(ctx, "am_get_root_value", uint64(valuePtr))
	if err != nil {
		return "", err
	}

	if err := checkErrorCode("am_get_root_value", results); err != nil {
		return "", err
	}

	// Read value from memory
	valueBytes, ok := mem.Read(valuePtr, uint32(valueLen))
	if !ok {
		return "", fmt.Errorf("failed to read value from memory")
	}

	return string(valueBytes), nil
}

// AmDeleteRoot deletes a key from ROOT
func (r *Runtime) AmDeleteRoot(ctx context.Context, key string) error {
	keyBytes := []byte(key)
	keyLen := uint32(len(keyBytes))

	// Allocate memory for key
	keyPtr, err := r.AmAlloc(ctx, keyLen)
	if err != nil {
		return fmt.Errorf("failed to allocate key: %w", err)
	}
	defer r.AmFree(ctx, keyPtr, keyLen)

	// Write key to memory
	mem := r.Memory()
	if !mem.Write(keyPtr, keyBytes) {
		return fmt.Errorf("failed to write key to memory")
	}

	// Call am_delete_root
	results, err := r.callExport(ctx, "am_delete_root", uint64(keyPtr), uint64(keyLen))
	if err != nil {
		return err
	}

	return checkErrorCode("am_delete_root", results)
}

// AmPutObjectRoot creates a nested object (map, list, or text) at ROOT level
func (r *Runtime) AmPutObjectRoot(ctx context.Context, key string, objType string) error {
	keyBytes := []byte(key)
	keyLen := uint32(len(keyBytes))
	typeBytes := []byte(objType)
	typeLen := uint32(len(typeBytes))

	// Allocate memory for key
	keyPtr, err := r.AmAlloc(ctx, keyLen)
	if err != nil {
		return fmt.Errorf("failed to allocate key: %w", err)
	}
	defer r.AmFree(ctx, keyPtr, keyLen)

	// Allocate memory for type
	typePtr, err := r.AmAlloc(ctx, typeLen)
	if err != nil {
		return fmt.Errorf("failed to allocate type: %w", err)
	}
	defer r.AmFree(ctx, typePtr, typeLen)

	// Write to memory
	mem := r.Memory()
	if !mem.Write(keyPtr, keyBytes) {
		return fmt.Errorf("failed to write key to memory")
	}
	if !mem.Write(typePtr, typeBytes) {
		return fmt.Errorf("failed to write type to memory")
	}

	// Call am_put_object_root
	results, err := r.callExport(ctx, "am_put_object_root",
		uint64(keyPtr), uint64(keyLen),
		uint64(typePtr), uint64(typeLen),
	)
	if err != nil {
		return err
	}

	return checkErrorCode("am_put_object_root", results)
}
