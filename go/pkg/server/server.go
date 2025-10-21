package server

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/joeblew999/automerge-wazero-example/pkg/automerge"
)

// Server manages the Automerge document and SSE client connections
type Server struct {
	doc        *automerge.Document
	mu         sync.RWMutex
	clients    []chan string
	storageDir string
	userID     string
}

// Config holds server configuration
type Config struct {
	StorageDir string
	UserID     string
}

// New creates a new Server instance
func New(cfg Config) *Server {
	return &Server{
		clients:    make([]chan string, 0),
		storageDir: cfg.StorageDir,
		userID:     cfg.UserID,
	}
}

// Initialize loads or creates a new Automerge document
func (s *Server) Initialize(ctx context.Context) error {
	snapshotPath := filepath.Join(s.storageDir, "doc.am")

	// Try to load existing snapshot
	if data, err := os.ReadFile(snapshotPath); err == nil {
		log.Printf("[%s] Loading existing snapshot from %s...", s.userID, snapshotPath)
		doc, err := automerge.Load(ctx, data)
		if err != nil {
			return fmt.Errorf("failed to load document: %w", err)
		}
		s.doc = doc
		return nil
	}

	// Initialize new document
	log.Printf("[%s] Initializing new document...", s.userID)
	doc, err := automerge.New(ctx)
	if err != nil {
		return fmt.Errorf("failed to create document: %w", err)
	}

	s.doc = doc
	return nil
}

// UserID returns the server's user identifier
func (s *Server) UserID() string {
	return s.userID
}

// Close closes the document and cleans up resources
func (s *Server) Close(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.doc != nil {
		return s.doc.Close(ctx)
	}
	return nil
}
