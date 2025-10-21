package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/joeblew999/automerge-wazero-example/pkg/server"
)

func TestLivenessHandler(t *testing.T) {
	// Create mock server
	srv := server.New(server.Config{
		StorageDir: t.TempDir(),
		UserID:     "test",
		WASMPath:   "../../../rust/automerge_wasi/target/wasm32-wasip1/debug/automerge_wasi.wasm",
	})

	// Create handler
	handler := LivenessHandler(srv)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()

	// Call handler
	handler(w, req)

	// Check status code
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	// Parse response
	var resp HealthResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	// Check response fields
	if resp.Status != "ok" {
		t.Errorf("expected status 'ok', got '%s'", resp.Status)
	}

	if resp.Service != "automerge-wazero" {
		t.Errorf("expected service 'automerge-wazero', got '%s'", resp.Service)
	}

	if resp.Details["check"] != "liveness" {
		t.Errorf("expected check 'liveness', got '%v'", resp.Details["check"])
	}

	if resp.Details["user_id"] != "test" {
		t.Errorf("expected user_id 'test', got '%v'", resp.Details["user_id"])
	}
}

func TestReadinessHandler_NotReady(t *testing.T) {
	// Create server WITHOUT initializing document
	srv := server.New(server.Config{
		StorageDir: t.TempDir(),
		UserID:     "test",
		WASMPath:   "../../../rust/automerge_wasi/target/wasm32-wasip1/debug/automerge_wasi.wasm",
	})

	// Create handler
	handler := ReadinessHandler(srv)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	w := httptest.NewRecorder()

	// Call handler
	handler(w, req)

	// Check status code (should be 503 since document not initialized)
	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("expected status 503, got %d", w.Code)
	}

	// Parse response
	var resp HealthResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	// Check response fields
	if resp.Status != "not_ready" {
		t.Errorf("expected status 'not_ready', got '%s'", resp.Status)
	}

	if resp.Details["document_initialized"] != false {
		t.Errorf("expected document_initialized false, got %v", resp.Details["document_initialized"])
	}
}

func TestHealthHandler_Combined(t *testing.T) {
	// Create mock server
	srv := server.New(server.Config{
		StorageDir: t.TempDir(),
		UserID:     "test",
		WASMPath:   "../../../rust/automerge_wasi/target/wasm32-wasip1/debug/automerge_wasi.wasm",
	})

	// Create handler
	handler := HealthHandler(srv)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	// Call handler
	handler(w, req)

	// Parse response
	var resp HealthResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	// Check that both liveness and readiness are included
	if resp.Details["liveness"] != "ok" {
		t.Errorf("expected liveness 'ok', got '%v'", resp.Details["liveness"])
	}

	// Should have readiness status
	if _, ok := resp.Details["readiness"]; !ok {
		t.Error("expected readiness field in details")
	}
}

func TestHealthHandler_MethodNotAllowed(t *testing.T) {
	srv := server.New(server.Config{
		StorageDir: t.TempDir(),
		UserID:     "test",
		WASMPath:   "../../../rust/automerge_wasi/target/wasm32-wasip1/debug/automerge_wasi.wasm",
	})

	handler := HealthHandler(srv)

	// Try POST instead of GET
	req := httptest.NewRequest(http.MethodPost, "/health", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	// Should return 405 Method Not Allowed
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}
