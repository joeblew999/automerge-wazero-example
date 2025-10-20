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

// SaveDocument saves the current document to disk (assumes lock is held)
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

// Broadcast sends a text update to all connected SSE clients
func (s *Server) Broadcast(text string) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, ch := range s.clients {
		select {
		case ch <- text:
		default:
			// Channel full, skip
		}
	}
}

// AddClient registers a new SSE client channel
func (s *Server) AddClient(ch chan string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.clients = append(s.clients, ch)
}

// RemoveClient unregisters an SSE client channel
func (s *Server) RemoveClient(ch chan string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, client := range s.clients {
		if client == ch {
			s.clients = append(s.clients[:i], s.clients[i+1:]...)
			break
		}
	}
}

// UserID returns the server's user identifier
func (s *Server) UserID() string {
	return s.userID
}
