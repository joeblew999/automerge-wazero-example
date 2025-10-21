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
