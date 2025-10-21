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
	wasmPath   string
}

// Config holds server configuration
type Config struct {
	StorageDir string
	UserID     string
	WASMPath   string
}

// New creates a new Server instance
func New(cfg Config) *Server {
	return &Server{
		clients:    make([]chan string, 0),
		storageDir: cfg.StorageDir,
		userID:     cfg.UserID,
		wasmPath:   cfg.WASMPath,
	}
}

// Initialize loads or creates a new Automerge document
func (s *Server) Initialize(ctx context.Context) error {
	snapshotPath := filepath.Join(s.storageDir, "doc.am")

	// Try to load existing snapshot
	if data, err := os.ReadFile(snapshotPath); err == nil {
		log.Printf("[%s] Loading existing snapshot from %s...", s.userID, snapshotPath)
		doc, err := automerge.LoadWithWASM(ctx, data, s.wasmPath)
		if err != nil {
			return fmt.Errorf("failed to load document: %w", err)
		}
		s.doc = doc
		return nil
	}

	// Initialize new document
	log.Printf("[%s] Initializing new document...", s.userID)
	doc, err := automerge.NewWithWASM(ctx, s.wasmPath)
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

// IsReady checks if the server is ready to accept traffic.
//
// Returns:
//   - ready: true if the server is ready, false otherwise
//   - details: map with detailed status information
//
// This is used by readiness probes (Kubernetes, load balancers, etc.)
func (s *Server) IsReady() (bool, map[string]interface{}) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	details := map[string]interface{}{
		"check":   "readiness",
		"user_id": s.userID,
	}

	// Check if document is initialized
	if s.doc == nil {
		details["document_initialized"] = false
		details["wasm_runtime"] = "not_loaded"
		return false, details
	}

	// Document exists = WASM runtime is loaded
	details["document_initialized"] = true
	details["wasm_runtime"] = "loaded"
	details["storage_dir"] = s.storageDir

	return true, details
}
