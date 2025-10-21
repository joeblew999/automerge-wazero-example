package server

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/joeblew999/automerge-wazero-example/pkg/automerge"
)

// Document lifecycle operations - maps to automerge/document.go

// saveDocument saves the current document to disk (assumes lock is held)
func (s *Server) saveDocument(ctx context.Context) error {
	data, err := s.doc.Save(ctx)
	if err != nil {
		return err
	}

	snapshotPath := filepath.Join(s.storageDir, "doc.am")

	if err := os.MkdirAll(filepath.Dir(snapshotPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(snapshotPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write snapshot: %w", err)
	}

	log.Printf("[%s] Saved document to %s (%d bytes)", s.userID, snapshotPath, len(data))
	return nil
}

// GetSnapshot returns the current document as bytes (thread-safe)
func (s *Server) GetSnapshot(ctx context.Context) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.doc.Save(ctx)
}

// Merge merges another document into this one (thread-safe)
func (s *Server) Merge(ctx context.Context, otherData []byte) error {
	// Load the other document
	other, err := automerge.Load(ctx, otherData)
	if err != nil {
		return fmt.Errorf("failed to load document to merge: %w", err)
	}
	defer other.Close(ctx)

	// Merge it into our document
	s.mu.Lock()
	err = s.doc.Merge(ctx, other)
	if err != nil {
		s.mu.Unlock()
		return fmt.Errorf("merge failed: %w", err)
	}

	// Save the merged document
	if err := s.saveDocument(ctx); err != nil {
		log.Printf("Warning: failed to save after merge: %v", err)
	}
	s.mu.Unlock()

	return nil
}
