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

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

const (
	wasmPath     = "../../../rust/automerge_wasi/target/wasm32-wasip1/release/automerge_wasi.wasm"
	snapshotPath = "../../../doc.am"
	port         = "8080"
)

type Server struct {
	runtime wazero.Runtime
	module  wazero.CompiledModule
	modInst api.Module
	mu      sync.RWMutex
	clients []chan string
}

type TextPayload struct {
	Text string `json:"text"`
}

func main() {
	ctx := context.Background()

	// Create wazero runtime
	runtime := wazero.NewRuntime(ctx)
	defer runtime.Close(ctx)

	// Instantiate WASI
	wasi_snapshot_preview1.MustInstantiate(ctx, runtime)

	// Load WASM module
	wasmBytes, err := os.ReadFile(wasmPath)
	if err != nil {
		log.Fatalf("Failed to read WASM file: %v", err)
	}

	compiled, err := runtime.CompileModule(ctx, wasmBytes)
	if err != nil {
		log.Fatalf("Failed to compile WASM module: %v", err)
	}

	// Instantiate module
	modInst, err := runtime.InstantiateModule(ctx, compiled, wazero.NewModuleConfig())
	if err != nil {
		log.Fatalf("Failed to instantiate module: %v", err)
	}

	server := &Server{
		runtime: runtime,
		module:  compiled,
		modInst: modInst,
		clients: make([]chan string, 0),
	}

	// Initialize or load document
	if err := server.initializeDocument(ctx); err != nil {
		log.Fatalf("Failed to initialize document: %v", err)
	}

	// Setup HTTP handlers
	http.HandleFunc("/api/text", server.handleText)
	http.HandleFunc("/api/stream", server.handleStream)
	http.HandleFunc("/", server.handleUI)

	log.Printf("Server starting on http://localhost:%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func (s *Server) initializeDocument(ctx context.Context) error {
	// Try to load existing snapshot
	if data, err := os.ReadFile(snapshotPath); err == nil {
		log.Println("Loading existing snapshot...")
		return s.loadDocument(ctx, data)
	}

	// Initialize new document
	log.Println("Initializing new document...")
	initFn := s.modInst.ExportedFunction("am_init")
	if initFn == nil {
		return fmt.Errorf("am_init function not found")
	}

	results, err := initFn.Call(ctx)
	if err != nil {
		return fmt.Errorf("failed to call am_init: %w", err)
	}

	if len(results) > 0 && results[0] != 0 {
		return fmt.Errorf("am_init returned error code: %d", results[0])
	}

	return nil
}

func (s *Server) getText(ctx context.Context) (string, error) {
	// Get text length
	getLenFn := s.modInst.ExportedFunction("am_get_text_len")
	if getLenFn == nil {
		return "", fmt.Errorf("am_get_text_len function not found")
	}

	results, err := getLenFn.Call(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get text length: %w", err)
	}

	textLen := uint32(results[0])
	if textLen == 0 {
		return "", nil
	}

	// Allocate buffer
	allocFn := s.modInst.ExportedFunction("am_alloc")
	if allocFn == nil {
		return "", fmt.Errorf("am_alloc function not found")
	}

	results, err = allocFn.Call(ctx, uint64(textLen))
	if err != nil {
		return "", fmt.Errorf("failed to allocate memory: %w", err)
	}

	ptr := uint32(results[0])
	if ptr == 0 {
		return "", fmt.Errorf("allocation failed")
	}

	defer func() {
		freeFn := s.modInst.ExportedFunction("am_free")
		if freeFn != nil {
			freeFn.Call(ctx, uint64(ptr), uint64(textLen))
		}
	}()

	// Get text
	getTextFn := s.modInst.ExportedFunction("am_get_text")
	if getTextFn == nil {
		return "", fmt.Errorf("am_get_text function not found")
	}

	results, err = getTextFn.Call(ctx, uint64(ptr))
	if err != nil {
		return "", fmt.Errorf("failed to get text: %w", err)
	}

	if results[0] != 0 {
		return "", fmt.Errorf("am_get_text returned error: %d", results[0])
	}

	// Read from memory
	mem := s.modInst.Memory()
	if mem == nil {
		return "", fmt.Errorf("memory not found")
	}

	data, ok := mem.Read(ptr, textLen)
	if !ok {
		return "", fmt.Errorf("failed to read memory")
	}

	return string(data), nil
}

func (s *Server) setText(ctx context.Context, text string) error {
	textBytes := []byte(text)
	textLen := uint32(len(textBytes))

	// Allocate buffer
	allocFn := s.modInst.ExportedFunction("am_alloc")
	if allocFn == nil {
		return fmt.Errorf("am_alloc function not found")
	}

	results, err := allocFn.Call(ctx, uint64(textLen))
	if err != nil {
		return fmt.Errorf("failed to allocate memory: %w", err)
	}

	ptr := uint32(results[0])
	if ptr == 0 {
		return fmt.Errorf("allocation failed")
	}

	defer func() {
		freeFn := s.modInst.ExportedFunction("am_free")
		if freeFn != nil {
			freeFn.Call(ctx, uint64(ptr), uint64(textLen))
		}
	}()

	// Write to memory
	mem := s.modInst.Memory()
	if mem == nil {
		return fmt.Errorf("memory not found")
	}

	if !mem.Write(ptr, textBytes) {
		return fmt.Errorf("failed to write to memory")
	}

	// Set text
	setTextFn := s.modInst.ExportedFunction("am_set_text")
	if setTextFn == nil {
		return fmt.Errorf("am_set_text function not found")
	}

	results, err = setTextFn.Call(ctx, uint64(ptr), uint64(textLen))
	if err != nil {
		return fmt.Errorf("failed to set text: %w", err)
	}

	if results[0] != 0 {
		return fmt.Errorf("am_set_text returned error: %d", results[0])
	}

	// Save snapshot
	if err := s.saveDocument(ctx); err != nil {
		log.Printf("Warning: failed to save snapshot: %v", err)
	}

	return nil
}

func (s *Server) saveDocument(ctx context.Context) error {
	// Get save length
	saveLenFn := s.modInst.ExportedFunction("am_save_len")
	if saveLenFn == nil {
		return fmt.Errorf("am_save_len function not found")
	}

	results, err := saveLenFn.Call(ctx)
	if err != nil {
		return fmt.Errorf("failed to get save length: %w", err)
	}

	saveLen := uint32(results[0])
	if saveLen == 0 {
		return nil
	}

	// Allocate buffer
	allocFn := s.modInst.ExportedFunction("am_alloc")
	if allocFn == nil {
		return fmt.Errorf("am_alloc function not found")
	}

	results, err = allocFn.Call(ctx, uint64(saveLen))
	if err != nil {
		return fmt.Errorf("failed to allocate memory: %w", err)
	}

	ptr := uint32(results[0])
	if ptr == 0 {
		return fmt.Errorf("allocation failed")
	}

	defer func() {
		freeFn := s.modInst.ExportedFunction("am_free")
		if freeFn != nil {
			freeFn.Call(ctx, uint64(ptr), uint64(saveLen))
		}
	}()

	// Save
	saveFn := s.modInst.ExportedFunction("am_save")
	if saveFn == nil {
		return fmt.Errorf("am_save function not found")
	}

	results, err = saveFn.Call(ctx, uint64(ptr))
	if err != nil {
		return fmt.Errorf("failed to save: %w", err)
	}

	if results[0] != 0 {
		return fmt.Errorf("am_save returned error: %d", results[0])
	}

	// Read from memory
	mem := s.modInst.Memory()
	if mem == nil {
		return fmt.Errorf("memory not found")
	}

	data, ok := mem.Read(ptr, saveLen)
	if !ok {
		return fmt.Errorf("failed to read memory")
	}

	// Write to file
	if err := os.MkdirAll(filepath.Dir(snapshotPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(snapshotPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write snapshot: %w", err)
	}

	return nil
}

func (s *Server) loadDocument(ctx context.Context, data []byte) error {
	dataLen := uint32(len(data))

	// Allocate buffer
	allocFn := s.modInst.ExportedFunction("am_alloc")
	if allocFn == nil {
		return fmt.Errorf("am_alloc function not found")
	}

	results, err := allocFn.Call(ctx, uint64(dataLen))
	if err != nil {
		return fmt.Errorf("failed to allocate memory: %w", err)
	}

	ptr := uint32(results[0])
	if ptr == 0 {
		return fmt.Errorf("allocation failed")
	}

	defer func() {
		freeFn := s.modInst.ExportedFunction("am_free")
		if freeFn != nil {
			freeFn.Call(ctx, uint64(ptr), uint64(dataLen))
		}
	}()

	// Write to memory
	mem := s.modInst.Memory()
	if mem == nil {
		return fmt.Errorf("memory not found")
	}

	if !mem.Write(ptr, data) {
		return fmt.Errorf("failed to write to memory")
	}

	// Load
	loadFn := s.modInst.ExportedFunction("am_load")
	if loadFn == nil {
		return fmt.Errorf("am_load function not found")
	}

	results, err = loadFn.Call(ctx, uint64(ptr), uint64(dataLen))
	if err != nil {
		return fmt.Errorf("failed to load: %w", err)
	}

	if results[0] != 0 {
		return fmt.Errorf("am_load returned error: %d", results[0])
	}

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
