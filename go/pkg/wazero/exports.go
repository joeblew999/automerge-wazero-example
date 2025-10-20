package wazero

import (
	"context"
	"fmt"
)

// Memory Management Exports

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

// Document Lifecycle Exports

// AmInit initializes a new Automerge document with a Text object at ROOT["content"]
func (r *Runtime) AmInit(ctx context.Context) error {
	results, err := r.callExport(ctx, "am_init")
	if err != nil {
		return err
	}
	return checkErrorCode("am_init", results)
}

// Text Operations Exports

// AmTextSplice performs a proper Text CRDT splice operation
func (r *Runtime) AmTextSplice(ctx context.Context, pos uint, del int64, text string) error {
	textBytes := []byte(text)
	textLen := uint32(len(textBytes))

	var textPtr uint32
	var err error

	if textLen > 0 {
		// Allocate memory for text
		textPtr, err = r.AmAlloc(ctx, textLen)
		if err != nil {
			return fmt.Errorf("failed to allocate memory for text: %w", err)
		}
		defer r.AmFree(ctx, textPtr, textLen)

		// Write text to WASM memory
		mem := r.Memory()
		if !mem.Write(textPtr, textBytes) {
			return fmt.Errorf("failed to write text to WASM memory")
		}
	}

	// Call am_text_splice
	results, err := r.callExport(ctx, "am_text_splice",
		uint64(pos),
		uint64(del),
		uint64(textPtr),
		uint64(textLen),
	)
	if err != nil {
		return err
	}

	return checkErrorCode("am_text_splice", results)
}

// AmSetText replaces all text content (DEPRECATED - use AmTextSplice)
func (r *Runtime) AmSetText(ctx context.Context, text string) error {
	textBytes := []byte(text)
	textLen := uint32(len(textBytes))

	// Allocate memory
	ptr, err := r.AmAlloc(ctx, textLen)
	if err != nil {
		return fmt.Errorf("failed to allocate memory: %w", err)
	}
	defer r.AmFree(ctx, ptr, textLen)

	// Write to memory
	mem := r.Memory()
	if !mem.Write(ptr, textBytes) {
		return fmt.Errorf("failed to write to WASM memory")
	}

	// Call am_set_text
	results, err := r.callExport(ctx, "am_set_text", uint64(ptr), uint64(textLen))
	if err != nil {
		return err
	}

	return checkErrorCode("am_set_text", results)
}

// AmGetTextLen returns the byte length of the current text content
func (r *Runtime) AmGetTextLen(ctx context.Context) (uint32, error) {
	results, err := r.callExport(ctx, "am_get_text_len")
	if err != nil {
		return 0, err
	}

	return uint32(results[0]), nil
}

// AmGetText retrieves the current text content
func (r *Runtime) AmGetText(ctx context.Context) (string, error) {
	// Get text length
	textLen, err := r.AmGetTextLen(ctx)
	if err != nil {
		return "", err
	}

	if textLen == 0 {
		return "", nil
	}

	// Allocate buffer
	ptr, err := r.AmAlloc(ctx, textLen)
	if err != nil {
		return "", fmt.Errorf("failed to allocate memory: %w", err)
	}
	defer r.AmFree(ctx, ptr, textLen)

	// Get text
	results, err := r.callExport(ctx, "am_get_text", uint64(ptr))
	if err != nil {
		return "", err
	}

	if err := checkErrorCode("am_get_text", results); err != nil {
		return "", err
	}

	// Read from memory
	mem := r.Memory()
	data, ok := mem.Read(ptr, textLen)
	if !ok {
		return "", fmt.Errorf("failed to read text from WASM memory")
	}

	return string(data), nil
}

// Persistence Exports

// AmSaveLen returns the byte size of the serialized document
func (r *Runtime) AmSaveLen(ctx context.Context) (uint32, error) {
	results, err := r.callExport(ctx, "am_save_len")
	if err != nil {
		return 0, err
	}

	return uint32(results[0]), nil
}

// AmSave serializes the document to binary format
func (r *Runtime) AmSave(ctx context.Context) ([]byte, error) {
	// Get save length
	saveLen, err := r.AmSaveLen(ctx)
	if err != nil {
		return nil, err
	}

	if saveLen == 0 {
		return []byte{}, nil
	}

	// Allocate buffer
	ptr, err := r.AmAlloc(ctx, saveLen)
	if err != nil {
		return nil, fmt.Errorf("failed to allocate memory: %w", err)
	}
	defer r.AmFree(ctx, ptr, saveLen)

	// Save
	results, err := r.callExport(ctx, "am_save", uint64(ptr))
	if err != nil {
		return nil, err
	}

	if err := checkErrorCode("am_save", results); err != nil {
		return nil, err
	}

	// Read from memory
	mem := r.Memory()
	data, ok := mem.Read(ptr, saveLen)
	if !ok {
		return nil, fmt.Errorf("failed to read save data from WASM memory")
	}

	// Make a copy since data is backed by WASM memory
	result := make([]byte, len(data))
	copy(result, data)

	return result, nil
}

// AmLoad loads a document from binary format
func (r *Runtime) AmLoad(ctx context.Context, data []byte) error {
	dataLen := uint32(len(data))

	// Allocate buffer
	ptr, err := r.AmAlloc(ctx, dataLen)
	if err != nil {
		return fmt.Errorf("failed to allocate memory: %w", err)
	}
	defer r.AmFree(ctx, ptr, dataLen)

	// Write to memory
	mem := r.Memory()
	if !mem.Write(ptr, data) {
		return fmt.Errorf("failed to write data to WASM memory")
	}

	// Load
	results, err := r.callExport(ctx, "am_load", uint64(ptr), uint64(dataLen))
	if err != nil {
		return err
	}

	return checkErrorCode("am_load", results)
}

// Merging Exports

// AmMerge merges another document into the current document (CRDT magic!)
func (r *Runtime) AmMerge(ctx context.Context, otherDoc []byte) error {
	dataLen := uint32(len(otherDoc))

	// Allocate buffer
	ptr, err := r.AmAlloc(ctx, dataLen)
	if err != nil {
		return fmt.Errorf("failed to allocate memory: %w", err)
	}
	defer r.AmFree(ctx, ptr, dataLen)

	// Write to memory
	mem := r.Memory()
	if !mem.Write(ptr, otherDoc) {
		return fmt.Errorf("failed to write data to WASM memory")
	}

	// Merge
	results, err := r.callExport(ctx, "am_merge", uint64(ptr), uint64(dataLen))
	if err != nil {
		return err
	}

	return checkErrorCode("am_merge", results)
}

// Map Operations Exports

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

// List Operations Exports

// AmListPush appends a string value to the end of the list
func (r *Runtime) AmListPush(ctx context.Context, value string) error {
	valueBytes := []byte(value)
	valuePtr := uint32(1)
	mem := r.Memory()

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

	results, err := r.callExport(ctx, "am_list_push", uint64(valuePtr), uint64(len(valueBytes)))
	if err != nil {
		return err
	}
	return checkErrorCode("am_list_push", results)
}

// AmListInsert inserts a string value at a specific index
func (r *Runtime) AmListInsert(ctx context.Context, index uint, value string) error {
	valueBytes := []byte(value)
	valuePtr := uint32(1)
	mem := r.Memory()

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

	results, err := r.callExport(ctx, "am_list_insert", uint64(index), uint64(valuePtr), uint64(len(valueBytes)))
	if err != nil {
		return err
	}
	return checkErrorCode("am_list_insert", results)
}

// AmListGet retrieves a string value at a specific index
func (r *Runtime) AmListGet(ctx context.Context, index uint) (string, error) {
	mem := r.Memory()

	// Get value length
	results, err := r.callExport(ctx, "am_list_get_len", uint64(index))
	if err != nil {
		return "", err
	}
	valueLen := uint32(results[0])

	if valueLen == 0 {
		valuePtr := uint32(1)
		results, err = r.callExport(ctx, "am_list_get", uint64(index), uint64(valuePtr))
		if err != nil {
			return "", err
		}
		if err := checkErrorCode("am_list_get", results); err != nil {
			return "", err
		}
		return "", nil
	}

	valuePtr, err := r.AmAlloc(ctx, valueLen)
	if err != nil {
		return "", fmt.Errorf("failed to allocate value buffer: %w", err)
	}
	defer r.AmFree(ctx, valuePtr, valueLen)

	results, err = r.callExport(ctx, "am_list_get", uint64(index), uint64(valuePtr))
	if err != nil {
		return "", err
	}
	if err := checkErrorCode("am_list_get", results); err != nil {
		return "", err
	}

	valueBytes, ok := mem.Read(valuePtr, valueLen)
	if !ok {
		return "", fmt.Errorf("failed to read value from WASM memory")
	}

	return string(valueBytes), nil
}

// AmListDelete removes a value at a specific index
func (r *Runtime) AmListDelete(ctx context.Context, index uint) error {
	results, err := r.callExport(ctx, "am_list_delete", uint64(index))
	if err != nil {
		return err
	}
	return checkErrorCode("am_list_delete", results)
}

// AmListLen returns the number of elements in the list
func (r *Runtime) AmListLen(ctx context.Context) (uint32, error) {
	results, err := r.callExport(ctx, "am_list_len")
	if err != nil {
		return 0, err
	}
	return uint32(results[0]), nil
}

// Counter Operations Exports

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
