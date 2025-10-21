package server

import (
	"context"
	"fmt"
	"log"

	"github.com/joeblew999/automerge-wazero-example/pkg/automerge"
)

// List operations - maps to automerge/list.go

// ListPush appends a value to the end of a list (thread-safe)
func (s *Server) ListPush(ctx context.Context, path automerge.Path, value string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.doc.ListPush(ctx, path, automerge.NewString(value)); err != nil {
		return err
	}

	if err := s.saveDocument(ctx); err != nil {
		log.Printf("Warning: failed to save snapshot: %v", err)
	}

	return nil
}

// ListInsert inserts a value at a specific index (thread-safe)
func (s *Server) ListInsert(ctx context.Context, path automerge.Path, index uint, value string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.doc.ListInsert(ctx, path, index, automerge.NewString(value)); err != nil {
		return err
	}

	if err := s.saveDocument(ctx); err != nil {
		log.Printf("Warning: failed to save snapshot: %v", err)
	}

	return nil
}

// ListGet retrieves a value at a specific index (thread-safe)
func (s *Server) ListGet(ctx context.Context, path automerge.Path, index uint) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	val, err := s.doc.ListGet(ctx, path, index)
	if err != nil {
		return "", err
	}

	str, ok := val.AsString()
	if !ok {
		return "", fmt.Errorf("value is not a string")
	}

	return str, nil
}

// ListDelete removes a value at a specific index (thread-safe)
func (s *Server) ListDelete(ctx context.Context, path automerge.Path, index uint) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.doc.ListDelete(ctx, path, index); err != nil {
		return err
	}

	if err := s.saveDocument(ctx); err != nil {
		log.Printf("Warning: failed to save snapshot: %v", err)
	}

	return nil
}

// ListLen returns the number of elements in a list (thread-safe)
func (s *Server) ListLen(ctx context.Context, path automerge.Path) (uint32, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	len, err := s.doc.ListLength(ctx, path)
	return uint32(len), err
}
