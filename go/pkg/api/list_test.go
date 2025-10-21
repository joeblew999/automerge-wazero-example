package api_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/joeblew999/automerge-wazero-example/pkg/api"
)

// TestListOperations tests List operations via HTTP
func TestListOperations(t *testing.T) {
	srv := newTestServer(t)

	t.Run("PUSH to list", func(t *testing.T) {
		handler := api.ListPushHandler(srv)
		payload := map[string]interface{}{
			"path":  "ROOT.items",
			"value": "item1",
		}

		rr := doRequest(t, handler, "POST", "/api/list/push", payload)

		if status := rr.Code; status != http.StatusNoContent {
			t.Errorf("PUSH returned wrong status: got %v want %v, body: %s", status, http.StatusNoContent, rr.Body.String())
		}
	})

	t.Run("INSERT into list", func(t *testing.T) {
		handler := api.ListInsertHandler(srv)
		payload := map[string]interface{}{
			"path":  "ROOT.items",
			"index": 0,
			"value": "item0",
		}

		rr := doRequest(t, handler, "POST", "/api/list/insert", payload)

		if status := rr.Code; status != http.StatusNoContent {
			t.Errorf("INSERT returned wrong status: got %v want %v, body: %s", status, http.StatusNoContent, rr.Body.String())
		}
	})

	t.Run("GET list length", func(t *testing.T) {
		handler := api.ListLenHandler(srv)

		rr := doRequest(t, handler, "GET", "/api/list/len?path=ROOT.items", nil)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("GET length returned wrong status: got %v want %v", status, http.StatusOK)
		}

		var resp map[string]interface{}
		if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		// Should have at least 2 items (from PUSH and INSERT)
		if length, ok := resp["length"].(float64); !ok || length < 2 {
			t.Errorf("Expected length >= 2, got %v", resp["length"])
		}
	})

	t.Run("GET list item", func(t *testing.T) {
		handler := api.ListGetHandler(srv)

		rr := doRequest(t, handler, "GET", "/api/list?path=ROOT.items&index=0", nil)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("GET item returned wrong status: got %v want %v", status, http.StatusOK)
		}

		var resp map[string]interface{}
		if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if _, ok := resp["value"]; !ok {
			t.Error("Response does not contain value")
		}
	})

	t.Run("DELETE list item", func(t *testing.T) {
		handler := api.ListDeleteHandler(srv)

		rr := doRequest(t, handler, "DELETE", "/api/list?path=ROOT.items&index=0", nil)

		if status := rr.Code; status != http.StatusNoContent {
			t.Errorf("DELETE returned wrong status: got %v want %v", status, http.StatusNoContent)
		}
	})
}
