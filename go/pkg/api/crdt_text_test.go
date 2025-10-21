package api_test

import (
	"net/http"
	"testing"

	"github.com/joeblew999/automerge-wazero-example/pkg/api"
)

// TestTextOperations tests existing Text operations still work
func TestTextOperations(t *testing.T) {
	srv := newTestServer(t)
	handler := api.TextHandler(srv)

	t.Run("GET text", func(t *testing.T) {
		rr := doRequest(t, handler, "GET", "/api/text", nil)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("GET text returned wrong status: got %v want %v", status, http.StatusOK)
		}

		// Should return empty text initially
		body := rr.Body.String()
		if body != "" && body != "null" {
			t.Logf("Initial text: %q", body)
		}
	})

	t.Run("POST text", func(t *testing.T) {
		payload := map[string]interface{}{
			"text": "Hello CRDT World!",
		}

		rr := doRequest(t, handler, "POST", "/api/text", payload)

		if status := rr.Code; status != http.StatusNoContent {
			t.Errorf("POST text returned wrong status: got %v want %v", status, http.StatusNoContent)
		}

		// Verify text was set
		rr = doRequest(t, handler, "GET", "/api/text", nil)
		body := rr.Body.String()
		if body != `"Hello CRDT World!"` && body != "Hello CRDT World!" {
			t.Errorf("GET text after POST returned wrong value: %q", body)
		}
	})
}
