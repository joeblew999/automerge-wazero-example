package api_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/joeblew999/automerge-wazero-example/pkg/api"
)

// TestMapOperations tests Map CRUD via HTTP
func TestMapOperations(t *testing.T) {
	srv := newTestServer(t)
	handler := api.MapHandler(srv)

	t.Run("PUT value", func(t *testing.T) {
		payload := map[string]interface{}{
			"path":  "ROOT",
			"key":   "name",
			"value": "Alice",
		}

		rr := doRequest(t, handler, "POST", "/api/map", payload)

		if status := rr.Code; status != http.StatusNoContent {
			t.Errorf("PUT returned wrong status: got %v want %v", status, http.StatusNoContent)
		}
	})

	t.Run("GET value", func(t *testing.T) {
		// First PUT a value
		payload := map[string]interface{}{
			"path":  "ROOT",
			"key":   "age",
			"value": "30",
		}
		doRequest(t, handler, "POST", "/api/map", payload)

		// Then GET it
		rr := doRequest(t, handler, "GET", "/api/map?path=ROOT&key=age", nil)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("GET returned wrong status: got %v want %v", status, http.StatusOK)
		}

		var resp map[string]interface{}
		if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if val, ok := resp["value"].(string); !ok || val != "30" {
			t.Errorf("GET returned wrong value: got %v want %v", resp["value"], "30")
		}
	})

	t.Run("DELETE value", func(t *testing.T) {
		// First PUT a value
		payload := map[string]interface{}{
			"path":  "ROOT",
			"key":   "temp",
			"value": "delete me",
		}
		doRequest(t, handler, "POST", "/api/map", payload)

		// Then DELETE it
		rr := doRequest(t, handler, "DELETE", "/api/map?path=ROOT&key=temp", nil)

		if status := rr.Code; status != http.StatusNoContent {
			t.Errorf("DELETE returned wrong status: got %v want %v", status, http.StatusNoContent)
		}

		// Verify it's gone (GET should fail or return empty)
		rr = doRequest(t, handler, "GET", "/api/map?path=ROOT&key=temp", nil)
		if status := rr.Code; status == http.StatusOK {
			var resp map[string]interface{}
			json.Unmarshal(rr.Body.Bytes(), &resp)
			if val, ok := resp["value"].(string); ok && val == "delete me" {
				t.Error("DELETE did not remove the value")
			}
		}
	})

	t.Run("GET keys", func(t *testing.T) {
		keysHandler := api.MapKeysHandler(srv)

		// PUT several values
		for _, key := range []string{"k1", "k2", "k3"} {
			payload := map[string]interface{}{
				"path":  "ROOT",
				"key":   key,
				"value": "test",
			}
			doRequest(t, handler, "POST", "/api/map", payload)
		}

		// GET all keys
		rr := doRequest(t, keysHandler, "GET", "/api/map/keys?path=ROOT", nil)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("GET keys returned wrong status: got %v want %v", status, http.StatusOK)
		}

		var resp map[string]interface{}
		if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		keys, ok := resp["keys"].([]interface{})
		if !ok {
			t.Fatalf("Response does not contain keys array")
		}

		// Should have at least our 3 keys (plus "content" from initialization)
		if len(keys) < 3 {
			t.Errorf("Expected at least 3 keys, got %d: %v", len(keys), keys)
		}
	})
}
