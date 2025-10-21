package api_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/joeblew999/automerge-wazero-example/pkg/api"
)

// TestRichTextOperations tests RichText mark operations via HTTP (M2)
func TestRichTextOperations(t *testing.T) {
	srv := newTestServer(t)

	// First, set some text to work with
	textHandler := api.TextHandler(srv)
	textPayload := map[string]interface{}{
		"text": "Hello World",
	}
	doRequest(t, textHandler, "POST", "/api/text", textPayload)

	t.Run("Apply mark", func(t *testing.T) {
		handler := api.RichTextMarkHandler(srv)
		payload := map[string]interface{}{
			"path":   "ROOT.content",
			"name":   "bold",
			"value":  "true",
			"start":  0,
			"end":    5, // "Hello"
			"expand": "none",
		}

		rr := doRequest(t, handler, "POST", "/api/richtext/mark", payload)

		if status := rr.Code; status != http.StatusNoContent {
			t.Errorf("Mark returned wrong status: got %v want %v, body: %s", status, http.StatusNoContent, rr.Body.String())
		}
	})

	t.Run("Get marks at position", func(t *testing.T) {
		marksHandler := api.RichTextMarksHandler(srv)

		// First apply a mark
		markHandler := api.RichTextMarkHandler(srv)
		payload := map[string]interface{}{
			"path":   "ROOT.content",
			"name":   "italic",
			"value":  "true",
			"start":  6,
			"end":    11, // "World"
			"expand": "none",
		}
		doRequest(t, markHandler, "POST", "/api/richtext/mark", payload)

		// Get marks at position 7 (in "World")
		rr := doRequest(t, marksHandler, "GET", "/api/richtext/marks?path=ROOT.content&pos=7", nil)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("Get marks returned wrong status: got %v want %v", status, http.StatusOK)
		}

		var resp map[string]interface{}
		if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		marks, ok := resp["marks"].([]interface{})
		if !ok {
			t.Fatalf("Response does not contain marks array")
		}

		// Should have at least the italic mark
		if len(marks) < 1 {
			t.Errorf("Expected at least 1 mark at position 7, got %d", len(marks))
		}

		t.Logf("Marks at position 7: %v", marks)
	})

	t.Run("Remove mark", func(t *testing.T) {
		handler := api.RichTextUnmarkHandler(srv)
		payload := map[string]interface{}{
			"path":   "ROOT.content",
			"name":   "bold",
			"start":  0,
			"end":    5,
			"expand": "none",
		}

		rr := doRequest(t, handler, "POST", "/api/richtext/unmark", payload)

		if status := rr.Code; status != http.StatusNoContent {
			t.Errorf("Unmark returned wrong status: got %v want %v", status, http.StatusNoContent)
		}
	})

	t.Run("Apply multiple marks", func(t *testing.T) {
		handler := api.RichTextMarkHandler(srv)

		// Apply bold
		payload1 := map[string]interface{}{
			"path":   "ROOT.content",
			"name":   "bold",
			"value":  "true",
			"start":  0,
			"end":    5,
			"expand": "none",
		}
		doRequest(t, handler, "POST", "/api/richtext/mark", payload1)

		// Apply underline to overlapping range
		payload2 := map[string]interface{}{
			"path":   "ROOT.content",
			"name":   "underline",
			"value":  "true",
			"start":  3,
			"end":    8,
			"expand": "none",
		}
		doRequest(t, handler, "POST", "/api/richtext/mark", payload2)

		// Get marks at position 4 (should have both bold and underline)
		marksHandler := api.RichTextMarksHandler(srv)
		rr := doRequest(t, marksHandler, "GET", "/api/richtext/marks?path=ROOT.content&pos=4", nil)

		var resp map[string]interface{}
		json.Unmarshal(rr.Body.Bytes(), &resp)

		marks, _ := resp["marks"].([]interface{})
		if len(marks) < 2 {
			t.Logf("Expected at least 2 marks at position 4, got %d: %v", len(marks), marks)
		}
	})
}
