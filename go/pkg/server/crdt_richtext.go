// ==============================================================================
// Layer 5: Go Server - Rich Text (Stateful + Thread-safe)
// ==============================================================================
// ARCHITECTURE: This is the stateful server layer (Layer 5/7).
//
// RESPONSIBILITIES:
// - Thread-safe CRDT operations (mutex protection)
// - State management (owns *automerge.Document, sync.RWMutex)
// - Persistence (saveDocument after mutations)
// - SSE broadcasting to connected clients
//
// DEPENDENCIES:
// - Layer 4: pkg/automerge (pure CRDT operations)
//
// DEPENDENTS:
// - Layer 6: pkg/api (HTTP handlers)
//
// RELATED FILES (1:1 mapping):
// - Layer 2: rust/automerge_wasi/src/richtext.rs (WASI exports)
// - Layer 3: pkg/wazero/crdt_richtext.go (FFI wrappers)
// - Layer 4: pkg/automerge/crdt_richtext.go (pure CRDT API)
// - Layer 6: pkg/api/crdt_richtext.go (HTTP handlers)
// - Layer 7: web/js/crdt_richtext.js + web/components/crdt_richtext.html
//
// NOTES:
// - All public methods are thread-safe (use s.mu.Lock/RLock)
// - This layer delegates to Layer 4 for actual CRDT operations
// - Broadcasts updates to SSE clients after mutations
// ==============================================================================

package server

import (
	"context"
	"log"

	"github.com/joeblew999/automerge-wazero-example/pkg/automerge"
)

// RichText operations - maps to automerge/richtext.go (M2 milestone)

// RichTextMark applies a mark (bold, italic, etc.) to a range of text (thread-safe)
func (s *Server) RichTextMark(ctx context.Context, path automerge.Path, mark automerge.Mark, expand automerge.ExpandMark) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.doc.Mark(ctx, path, mark, expand); err != nil {
		return err
	}

	if err := s.saveDocument(ctx); err != nil {
		log.Printf("Warning: failed to save snapshot: %v", err)
	}

	return nil
}

// RichTextUnmark removes a mark from a range of text (thread-safe)
func (s *Server) RichTextUnmark(ctx context.Context, path automerge.Path, name string, start, end uint, expand automerge.ExpandMark) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.doc.Unmark(ctx, path, name, start, end, expand); err != nil {
		return err
	}

	if err := s.saveDocument(ctx); err != nil {
		log.Printf("Warning: failed to save snapshot: %v", err)
	}

	return nil
}

// GetRichTextMarks retrieves all marks at a specific position (thread-safe)
func (s *Server) GetRichTextMarks(ctx context.Context, path automerge.Path, pos uint) ([]automerge.Mark, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.doc.GetMarks(ctx, path, pos)
}
