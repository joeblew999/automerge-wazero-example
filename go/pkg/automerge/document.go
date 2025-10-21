package automerge

import (
	"context"

	"github.com/joeblew999/automerge-wazero-example/pkg/wazero"
)

// Document represents an Automerge CRDT document
type Document struct {
	runtime *wazero.Runtime
}

// New creates a new empty Automerge document
func New(ctx context.Context) (*Document, error) {
	// Create wazero runtime
	runtime, err := wazero.New(ctx, wazero.Config{})
	if err != nil {
		return nil, err
	}

	// Initialize document
	if err := runtime.AmInit(ctx); err != nil {
		runtime.Close(ctx)
		return nil, err
	}

	return &Document{runtime: runtime}, nil
}

// Load creates a document from a binary snapshot
func Load(ctx context.Context, data []byte) (*Document, error) {
	// Create wazero runtime
	runtime, err := wazero.New(ctx, wazero.Config{})
	if err != nil {
		return nil, err
	}

	// Load document
	if err := runtime.AmLoad(ctx, data); err != nil {
		runtime.Close(ctx)
		return nil, err
	}

	return &Document{runtime: runtime}, nil
}

// Close closes the document and frees resources
func (d *Document) Close(ctx context.Context) error {
	return d.runtime.Close(ctx)
}

// Save serializes the document to binary format
func (d *Document) Save(ctx context.Context) ([]byte, error) {
	return d.runtime.AmSave(ctx)
}

// Merge merges another document into this one (CRDT magic!)
//
// This is conflict-free and deterministic - both documents will end up
// with the same state regardless of merge order.
//
// Status: ✅ Implemented
func (d *Document) Merge(ctx context.Context, other *Document) error {
	// Save the other document
	otherData, err := other.Save(ctx)
	if err != nil {
		return err
	}

	// Merge into this document
	return d.runtime.AmMerge(ctx, otherData)
}

// GetActor returns the actor ID for this document.
//
// The actor ID uniquely identifies this peer in the distributed system.
// It's used to track which changes came from which peer.
//
// Status: ✅ Implemented
func (d *Document) GetActor(ctx context.Context) (string, error) {
	return d.runtime.AmGetActor(ctx)
}

// SetActor sets the actor ID for this document.
//
// This should be set before making any changes to the document.
// The actor ID uniquely identifies this peer.
//
// Status: ✅ Implemented
func (d *Document) SetActor(ctx context.Context, actorID string) error {
	return d.runtime.AmSetActor(ctx, actorID)
}
