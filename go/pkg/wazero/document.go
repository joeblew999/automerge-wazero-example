package wazero

import (
	"context"
	"fmt"
)

// Document Lifecycle - maps to rust/automerge_wasi/src/document.rs

// AmInit initializes a new Automerge document with a Text object at ROOT["content"]
func (r *Runtime) AmInit(ctx context.Context) error {
	results, err := r.callExport(ctx, "am_init")
	if err != nil {
		return err
	}
	return checkErrorCode("am_init", results)
}

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

// AmGetActorLen returns the byte length of the actor ID string
func (r *Runtime) AmGetActorLen(ctx context.Context) (uint32, error) {
	results, err := r.callExport(ctx, "am_get_actor_len")
	if err != nil {
		return 0, err
	}
	return uint32(results[0]), nil
}

// AmGetActor returns the actor ID for the current document
func (r *Runtime) AmGetActor(ctx context.Context) (string, error) {
	// Get actor length
	actorLen, err := r.AmGetActorLen(ctx)
	if err != nil {
		return "", err
	}

	if actorLen == 0 {
		return "", fmt.Errorf("document not initialized")
	}

	// Allocate buffer
	ptr, err := r.AmAlloc(ctx, actorLen)
	if err != nil {
		return "", fmt.Errorf("failed to allocate memory: %w", err)
	}
	defer r.AmFree(ctx, ptr, actorLen)

	// Get actor
	results, err := r.callExport(ctx, "am_get_actor", uint64(ptr))
	if err != nil {
		return "", err
	}

	if err := checkErrorCode("am_get_actor", results); err != nil {
		return "", err
	}

	// Read from memory
	mem := r.Memory()
	data, ok := mem.Read(ptr, actorLen)
	if !ok {
		return "", fmt.Errorf("failed to read actor from WASM memory")
	}

	return string(data), nil
}

// AmSetActor sets the actor ID for the current document
func (r *Runtime) AmSetActor(ctx context.Context, actorID string) error {
	actorBytes := []byte(actorID)
	actorLen := uint32(len(actorBytes))

	// Allocate buffer
	ptr, err := r.AmAlloc(ctx, actorLen)
	if err != nil {
		return fmt.Errorf("failed to allocate memory: %w", err)
	}
	defer r.AmFree(ctx, ptr, actorLen)

	// Write to memory
	mem := r.Memory()
	if !mem.Write(ptr, actorBytes) {
		return fmt.Errorf("failed to write actor to WASM memory")
	}

	// Set actor
	results, err := r.callExport(ctx, "am_set_actor", uint64(ptr), uint64(actorLen))
	if err != nil {
		return err
	}

	return checkErrorCode("am_set_actor", results)
}

// AmFork creates an independent copy of the document with a new actor ID
func (r *Runtime) AmFork(ctx context.Context) error {
	results, err := r.callExport(ctx, "am_fork")
	if err != nil {
		return err
	}
	return checkErrorCode("am_fork", results)
}
