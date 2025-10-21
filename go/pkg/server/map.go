package server

import (
	"context"
	"fmt"
	"log"

	"github.com/joeblew999/automerge-wazero-example/pkg/automerge"
)

// Map operations - maps to automerge/map.go

// GetMapValue gets a value from a map at the given path and key (thread-safe)
func (s *Server) GetMapValue(ctx context.Context, path automerge.Path, key string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	value, err := s.doc.Get(ctx, path, key)
	if err != nil {
		return "", err
	}

	str, ok := value.AsString()
	if !ok {
		return "", fmt.Errorf("value is not a string")
	}

	return str, nil
}

// PutMapValue sets a value in a map at the given path and key (thread-safe)
func (s *Server) PutMapValue(ctx context.Context, path automerge.Path, key string, value string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.doc.Put(ctx, path, key, automerge.NewString(value)); err != nil {
		return err
	}

	// Save snapshot
	if err := s.saveDocument(ctx); err != nil {
		log.Printf("Warning: failed to save snapshot: %v", err)
	}

	return nil
}

// DeleteMapKey deletes a key from a map (thread-safe)
func (s *Server) DeleteMapKey(ctx context.Context, path automerge.Path, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.doc.Delete(ctx, path, key); err != nil {
		return err
	}

	// Save snapshot
	if err := s.saveDocument(ctx); err != nil {
		log.Printf("Warning: failed to save snapshot: %v", err)
	}

	return nil
}

// GetMapKeys returns all keys in a map (thread-safe)
func (s *Server) GetMapKeys(ctx context.Context, path automerge.Path) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.doc.Keys(ctx, path)
}
