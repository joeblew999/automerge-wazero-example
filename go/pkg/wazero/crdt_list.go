// ==============================================================================
// Layer 3: Go FFI Wrappers - List CRDT
// ==============================================================================
// ARCHITECTURE: This is the FFI wrapper layer (Layer 3/7).
//
// RESPONSIBILITIES:
// - 1:1 wrapping of WASI exports
// - Go → WASM memory marshaling
// - Error code handling
// - Memory allocation/deallocation via am_alloc/am_free
//
// DEPENDENCIES:
// - Layer 2: rust/automerge_wasi/src/list.rs (WASI exports)
// - wazero runtime (WASM execution)
//
// DEPENDENTS:
// - Layer 4: pkg/automerge/crdt_list.go (high-level API)
//
// RELATED FILES (1:1 mapping):
// - Layer 2: rust/automerge_wasi/src/list.rs (WASI exports)
// - Layer 4: pkg/automerge/crdt_list.go (Go high-level API)
// - Layer 5: pkg/server/crdt_list.go (stateful server)
// - Layer 6: pkg/api/crdt_list.go (HTTP handlers)
// - Layer 7: web/js/crdt_list.js + web/components/crdt_list.html (TODO)
//
// NOTES:
// - Each method corresponds exactly to one WASI export
// - No business logic here - just FFI bridging
// - Uses r.Memory() to write/read WASM linear memory
// ==============================================================================

package wazero

import (
	"context"
	"fmt"
)

// List Operations - maps to rust/automerge_wasi/src/list.rs

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

// AmListObjIdLen returns the buffer size needed for list object IDs
func (r *Runtime) AmListObjIdLen(ctx context.Context) (uint32, error) {
	results, err := r.callExport(ctx, "am_list_obj_id_len")
	if err != nil {
		return 0, err
	}
	return uint32(results[0]), nil
}

// AmListCreate creates a new list object at the given key and returns its object ID
func (r *Runtime) AmListCreate(ctx context.Context, key string) (string, error) {
	keyBytes := []byte(key)
	mem := r.Memory()

	// Allocate memory for key
	keyPtr, err := r.AmAlloc(ctx, uint32(len(keyBytes)))
	if err != nil {
		return "", fmt.Errorf("failed to allocate key: %w", err)
	}
	defer r.AmFree(ctx, keyPtr, uint32(len(keyBytes)))

	if !mem.Write(keyPtr, keyBytes) {
		return "", fmt.Errorf("failed to write key to WASM memory")
	}

	// Get buffer size for object ID
	objIdLen, err := r.AmListObjIdLen(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get obj_id length: %w", err)
	}

	// Allocate memory for object ID output
	objIdPtr, err := r.AmAlloc(ctx, objIdLen)
	if err != nil {
		return "", fmt.Errorf("failed to allocate obj_id buffer: %w", err)
	}
	defer r.AmFree(ctx, objIdPtr, objIdLen)

	// Call am_list_create
	results, err := r.callExport(ctx, "am_list_create", uint64(keyPtr), uint64(len(keyBytes)), uint64(objIdPtr))
	if err != nil {
		return "", err
	}
	if err := checkErrorCode("am_list_create", results); err != nil {
		return "", err
	}

	// Read object ID from memory
	objIdBytes, ok := mem.Read(objIdPtr, objIdLen)
	if !ok {
		return "", fmt.Errorf("failed to read obj_id from WASM memory")
	}

	// Trim null bytes
	objId := string(objIdBytes)
	if idx := len(objId); idx > 0 {
		for i := 0; i < len(objId); i++ {
			if objId[i] == 0 {
				idx = i
				break
			}
		}
		objId = objId[:idx]
	}

	return objId, nil
}
