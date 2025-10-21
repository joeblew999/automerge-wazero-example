package api_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/joeblew999/automerge-wazero-example/pkg/api"
)

// TestHistoryOperations tests History operations via HTTP
func TestHistoryOperations(t *testing.T) {
	srv := newTestServer(t)

	// Make some changes to create history
	mapHandler := api.MapHandler(srv)
	for i := 0; i < 3; i++ {
		payload := map[string]interface{}{
			"path":  "ROOT",
			"key":   "counter",
			"value": string(rune('A' + i)),
		}
		doRequest(t, mapHandler, "POST", "/api/map", payload)
	}

	t.Run("GET heads", func(t *testing.T) {
		handler := api.HeadsHandler(srv)

		rr := doRequest(t, handler, "GET", "/api/heads", nil)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("GET heads returned wrong status: got %v want %v", status, http.StatusOK)
		}

		var resp map[string]interface{}
		if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		heads, ok := resp["heads"].([]interface{})
		if !ok {
			t.Fatalf("Response does not contain heads array")
		}

		if len(heads) < 1 {
			t.Errorf("Expected at least 1 head, got %d", len(heads))
		}

		t.Logf("Heads: %v", heads)
	})

	t.Run("GET changes", func(t *testing.T) {
		handler := api.ChangesHandler(srv)

		rr := doRequest(t, handler, "GET", "/api/changes", nil)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("GET changes returned wrong status: got %v want %v", status, http.StatusOK)
		}

		var resp map[string]interface{}
		if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if _, ok := resp["changes"]; !ok {
			t.Error("Response does not contain changes")
		}

		if size, ok := resp["size"].(float64); !ok || size <= 0 {
			t.Errorf("Expected size > 0, got %v", resp["size"])
		}

		t.Logf("Changes size: %v bytes", resp["size"])
	})
}
