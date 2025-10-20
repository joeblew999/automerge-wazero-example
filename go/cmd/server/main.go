package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/joeblew999/automerge-wazero-example/pkg/automerge"
)

const (
	wasmPath = "../../../rust/automerge_wasi/target/wasm32-wasip1/release/automerge_wasi.wasm"
)

// getEnv returns environment variable or default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

var (
	// Storage directory for doc.am (configurable for multi-laptop testing)
	storageDir = getEnv("STORAGE_DIR", "../../../")
	// Server port (configurable for running multiple instances)
	port = getEnv("PORT", "8080")
	// User ID for logging (optional, helps distinguish laptop A vs B)
	userID = getEnv("USER_ID", "default")
)

type Server struct {
	doc     *automerge.Document
	mu      sync.RWMutex
	clients []chan string
}

type TextPayload struct {
	Text string `json:"text"`
}

func main() {
	ctx := context.Background()

	server := &Server{
		clients: make([]chan string, 0),
	}

	// Initialize or load document
	if err := server.initializeDocument(ctx); err != nil {
		log.Fatalf("Failed to initialize document: %v", err)
	}

	// Setup HTTP handlers
	http.HandleFunc("/api/text", server.handleText)
	http.HandleFunc("/api/stream", server.handleStream)
	http.HandleFunc("/api/merge", server.handleMerge)
	http.HandleFunc("/api/doc", server.handleDoc)
	http.HandleFunc("/", server.handleUI)

	log.Printf("Server starting on http://localhost:%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func (s *Server) initializeDocument(ctx context.Context) error {
	// Construct snapshot path from storage directory
	snapshotPath := filepath.Join(storageDir, "doc.am")

	// Try to load existing snapshot
	if data, err := os.ReadFile(snapshotPath); err == nil {
		log.Printf("[%s] Loading existing snapshot from %s...", userID, snapshotPath)
		doc, err := automerge.Load(ctx, data)
		if err != nil {
			return fmt.Errorf("failed to load document: %w", err)
		}
		s.doc = doc
		return nil
	}

	// Initialize new document
	log.Printf("[%s] Initializing new document...", userID)
	doc, err := automerge.New(ctx)
	if err != nil {
		return fmt.Errorf("failed to create document: %w", err)
	}

	s.doc = doc
	return nil
}

func (s *Server) getText(ctx context.Context) (string, error) {
	// Get text from root["content"] path
	path := automerge.Root().Get("content")
	return s.doc.GetText(ctx, path)
}

func (s *Server) setText(ctx context.Context, text string) error {
	// This is a simple text replacement using SpliceText
	// First get current length, then replace entire content
	path := automerge.Root().Get("content")

	currentLen, err := s.doc.TextLength(ctx, path)
	if err != nil {
		return err
	}

	// Delete all current text and insert new text (position 0, delete all, insert new)
	if err := s.doc.SpliceText(ctx, path, 0, int(currentLen), text); err != nil {
		return err
	}

	// Save snapshot
	if err := s.saveDocument(ctx); err != nil {
		log.Printf("Warning: failed to save snapshot: %v", err)
	}

	return nil
}

func (s *Server) saveDocument(ctx context.Context) error {
	// Save to bytes
	data, err := s.doc.Save(ctx)
	if err != nil {
		return err
	}

	// Construct snapshot path from storage directory
	snapshotPath := filepath.Join(storageDir, "doc.am")

	// Write to file
	if err := os.MkdirAll(filepath.Dir(snapshotPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(snapshotPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write snapshot: %w", err)
	}

	log.Printf("[%s] Saved document to %s (%d bytes)", userID, snapshotPath, len(data))
	return nil
}

func (s *Server) handleText(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	switch r.Method {
	case http.MethodGet:
		s.mu.RLock()
		text, err := s.getText(ctx)
		s.mu.RUnlock()

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(text))

	case http.MethodPost:
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read body", http.StatusBadRequest)
			return
		}

		var payload TextPayload
		if err := json.Unmarshal(body, &payload); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		s.mu.Lock()
		err = s.setText(ctx, payload.Text)
		s.mu.Unlock()

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Broadcast update to SSE clients
		s.broadcast(payload.Text)

		w.WriteHeader(http.StatusNoContent)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleStream(w http.ResponseWriter, r *http.Request) {
	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	// Create client channel
	clientChan := make(chan string, 10)

	s.mu.Lock()
	s.clients = append(s.clients, clientChan)
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		for i, ch := range s.clients {
			if ch == clientChan {
				s.clients = append(s.clients[:i], s.clients[i+1:]...)
				break
			}
		}
		s.mu.Unlock()
		close(clientChan)
	}()

	// Send initial snapshot
	s.mu.RLock()
	text, err := s.getText(r.Context())
	s.mu.RUnlock()

	if err == nil {
		data, _ := json.Marshal(map[string]string{"text": text})
		fmt.Fprintf(w, "event: snapshot\ndata: %s\n\n", data)
		flusher.Flush()
	}

	// Listen for updates
	for {
		select {
		case text, ok := <-clientChan:
			if !ok {
				return
			}
			data, _ := json.Marshal(map[string]string{"text": text})
			fmt.Fprintf(w, "event: update\ndata: %s\n\n", data)
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}

func (s *Server) broadcast(text string) {
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

func (s *Server) handleUI(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, "../../../ui/ui.html")
}

// GET /api/doc - Download the current doc.am file
func (s *Server) handleDoc(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	s.mu.RLock()
	data, err := s.doc.Save(ctx)
	s.mu.RUnlock()

	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to save document: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s-doc.am\"", userID))
	w.Write(data)
	log.Printf("[%s] Sent doc.am (%d bytes)", userID, len(data))
}

// POST /api/merge - Merge another doc.am into this one (CRDT magic!)
func (s *Server) handleMerge(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()

	// Read the incoming doc.am binary
	otherDoc, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}

	if len(otherDoc) == 0 {
		http.Error(w, "Empty document", http.StatusBadRequest)
		return
	}

	log.Printf("[%s] Received doc.am to merge (%d bytes)", userID, len(otherDoc))

	// Load the other document
	other, err := automerge.Load(ctx, otherDoc)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to load document to merge: %v", err), http.StatusInternalServerError)
		return
	}
	defer other.Close(ctx)

	// Merge it into our document
	s.mu.Lock()
	err = s.doc.Merge(ctx, other)
	s.mu.Unlock()

	if err != nil {
		http.Error(w, fmt.Sprintf("Merge failed: %v", err), http.StatusInternalServerError)
		return
	}

	// After merge, get the new text and broadcast
	s.mu.RLock()
	text, err := s.getText(ctx)
	s.mu.RUnlock()

	if err == nil {
		s.broadcast(text)
	}

	// Save the merged document
	s.mu.Lock()
	s.saveDocument(ctx)
	s.mu.Unlock()

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Merged successfully! New text: %s", text)
	log.Printf("[%s] Merge complete, new text: %s", userID, text)
}
