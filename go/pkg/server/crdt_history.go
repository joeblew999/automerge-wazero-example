package server

import (
	"context"

	"github.com/joeblew999/automerge-wazero-example/pkg/automerge"
)

// History operations - maps to automerge/history.go

// GetHeads returns the current heads (latest change hashes) of the document (thread-safe)
func (s *Server) GetHeads(ctx context.Context) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	heads, err := s.doc.GetHeads(ctx)
	if err != nil {
		return nil, err
	}

	// Convert []ChangeHash to []string
	result := make([]string, len(heads))
	for i, h := range heads {
		result[i] = h.String()
	}

	return result, nil
}

// GetChanges returns the raw changes bytes (optionally filtered by 'since' heads) (thread-safe)
func (s *Server) GetChanges(ctx context.Context, since []automerge.ChangeHash) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.doc.GetChanges(ctx, since)
}
