package server

import (
	"context"
	"log"

	"github.com/joeblew999/automerge-wazero-example/pkg/automerge"
)

// Counter operations - maps to automerge/counter.go

// IncrementCounter increments a counter at the given path and key (thread-safe)
func (s *Server) IncrementCounter(ctx context.Context, path automerge.Path, key string, delta int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.doc.Increment(ctx, path, key, delta); err != nil {
		return err
	}

	if err := s.saveDocument(ctx); err != nil {
		log.Printf("Warning: failed to save snapshot: %v", err)
	}

	return nil
}

// GetCounter retrieves the current value of a counter (thread-safe)
func (s *Server) GetCounter(ctx context.Context, path automerge.Path, key string) (int64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.doc.GetCounter(ctx, path, key)
}
