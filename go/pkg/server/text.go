package server

import (
	"context"
	"log"

	"github.com/joeblew999/automerge-wazero-example/pkg/automerge"
)

// Text operations - maps to automerge/text.go

// GetText returns the current text from the document (thread-safe)
func (s *Server) GetText(ctx context.Context) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	path := automerge.Root().Get("content")
	return s.doc.GetText(ctx, path)
}

// SetText replaces the entire text in the document (thread-safe)
func (s *Server) SetText(ctx context.Context, text string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	path := automerge.Root().Get("content")

	currentLen, err := s.doc.TextLength(ctx, path)
	if err != nil {
		return err
	}

	// Delete all current text and insert new text
	if err := s.doc.SpliceText(ctx, path, 0, int(currentLen), text); err != nil {
		return err
	}

	// Save snapshot
	if err := s.saveDocument(ctx); err != nil {
		log.Printf("Warning: failed to save snapshot: %v", err)
	}

	return nil
}
