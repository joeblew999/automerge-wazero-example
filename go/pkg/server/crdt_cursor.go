package server

import (
	"context"
	"fmt"

	"github.com/joeblew999/automerge-wazero-example/pkg/automerge"
)

// GetCursor creates a cursor at the given position in a text or list object.
// Cursors provide stable position tracking across concurrent edits.
//
// This method is read-only (uses RLock) since it doesn't modify the document.
func (s *Server) GetCursor(ctx context.Context, path string, index int) (*automerge.Cursor, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	cursor, err := s.doc.GetCursor(ctx, path, index)
	if err != nil {
		return nil, fmt.Errorf("failed to get cursor: %w", err)
	}

	return cursor, nil
}

// LookupCursor finds the current position of a cursor.
// Returns the current index where the cursor points.
//
// This method is read-only (uses RLock) since it doesn't modify the document.
func (s *Server) LookupCursor(ctx context.Context, cursor *automerge.Cursor) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	index, err := s.doc.LookupCursor(ctx, cursor)
	if err != nil {
		return 0, fmt.Errorf("failed to lookup cursor: %w", err)
	}

	return index, nil
}
