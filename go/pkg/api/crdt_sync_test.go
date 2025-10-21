package api_test

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/joeblew999/automerge-wazero-example/pkg/api"
)

// TestSyncOperations tests Sync protocol via HTTP (M1)
func TestSyncOperations(t *testing.T) {
	srv1 := newTestServer(t)
	_ = newTestServer(t) // srv2 for future two-way sync test

	// Make changes on server 1
	mapHandler1 := api.MapHandler(srv1)
	payload := map[string]interface{}{
		"path":  "ROOT",
		"key":   "sync_test",
		"value": "value_from_srv1",
	}
	doRequest(t, mapHandler1, "POST", "/api/map", payload)

	t.Run("Initialize sync", func(t *testing.T) {
		handler := api.SyncHandler(srv1)

		// Peer initiates sync (no message, just announces presence)
		syncPayload := map[string]interface{}{
			"peer_id": "peer-1",
		}

		rr := doRequest(t, handler, "POST", "/api/sync", syncPayload)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("Sync init returned wrong status: got %v want %v, body: %s", status, http.StatusOK, rr.Body.String())
		}

		var resp map[string]interface{}
		if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		// Server may respond with sync message (or empty if nothing to sync)
		if message, ok := resp["message"].(string); ok && message != "" {
			t.Logf("Received sync message: %d bytes (base64)", len(message))

			// Verify it's valid base64
			if _, err := base64.StdEncoding.DecodeString(message); err != nil {
				t.Errorf("Sync message is not valid base64: %v", err)
			}
		} else {
			t.Logf("No sync message (peer already in sync)")
		}

		hasMore, ok := resp["has_more"].(bool)
		if !ok {
			t.Error("Response does not contain has_more field")
		}
		t.Logf("Has more: %v", hasMore)
	})

	t.Run("Sync with message exchange", func(t *testing.T) {
		handler := api.SyncHandler(srv1)

		// First request to get initial sync message
		syncPayload1 := map[string]interface{}{
			"peer_id": "peer-2",
		}
		rr1 := doRequest(t, handler, "POST", "/api/sync", syncPayload1)

		var resp1 map[string]interface{}
		json.Unmarshal(rr1.Body.Bytes(), &resp1)

		// Check if we got a message
		message1, hasMessage := resp1["message"].(string)
		if !hasMessage || message1 == "" {
			t.Skip("No sync message to exchange (documents already in sync)")
			return
		}

		// Second request sending the message back (simulating peer response)
		syncPayload2 := map[string]interface{}{
			"peer_id": "peer-2",
			"message": message1,
		}
		rr2 := doRequest(t, handler, "POST", "/api/sync", syncPayload2)

		if status := rr2.Code; status != http.StatusOK {
			t.Errorf("Sync exchange returned wrong status: got %v want %v", status, http.StatusOK)
		}

		var resp2 map[string]interface{}
		if err := json.Unmarshal(rr2.Body.Bytes(), &resp2); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		t.Logf("Sync exchange completed, has_more=%v", resp2["has_more"])
	})

	// TODO: Test actual two-way sync between srv1 and srv2
	// This requires implementing full sync protocol flow
	t.Run("Two-way sync (placeholder)", func(t *testing.T) {
		t.Skip("Full two-way sync requires implementing complete sync loop")

		// Outline of what this test would do:
		// 1. Make changes on srv1
		// 2. Make different changes on srv2
		// 3. Exchange sync messages until both sides have has_more=false
		// 4. Verify both servers have merged state
	})
}
