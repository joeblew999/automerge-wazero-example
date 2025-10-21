package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/joeblew999/automerge-wazero-example/pkg/automerge"
	"github.com/joeblew999/automerge-wazero-example/pkg/server"
)

// Helper to create test server
func newTestServer(t *testing.T) *server.Server {
	t.Helper()
	ctx := context.Background()

	srv := server.New(server.Config{
		StorageDir: t.TempDir(),
		UserID:     "test-user",
		WASMPath:   automerge.TestWASMPath,
	})

	if err := srv.Initialize(ctx); err != nil {
		t.Fatalf("Failed to initialize server: %v", err)
	}

	return srv
}

// Helper to make HTTP requests
func doRequest(t *testing.T, handler http.HandlerFunc, method, url string, body interface{}) *httptest.ResponseRecorder {
	t.Helper()

	var reqBody io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("Failed to marshal request body: %v", err)
		}
		reqBody = bytes.NewReader(b)
	}

	req := httptest.NewRequest(method, url, reqBody)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	rr := httptest.NewRecorder()
	handler(rr, req)

	return rr
}
