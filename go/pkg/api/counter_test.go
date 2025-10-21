package api_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/joeblew999/automerge-wazero-example/pkg/api"
)

// TestCounterOperations tests Counter operations via HTTP
func TestCounterOperations(t *testing.T) {
	srv := newTestServer(t)

	t.Run("INCREMENT counter", func(t *testing.T) {
		handler := api.CounterIncrementHandler(srv)
		payload := map[string]interface{}{
			"path":  "ROOT",
			"key":   "clicks",
			"delta": 5,
		}

		rr := doRequest(t, handler, "POST", "/api/counter/increment", payload)

		if status := rr.Code; status != http.StatusNoContent {
			t.Errorf("INCREMENT returned wrong status: got %v want %v, body: %s", status, http.StatusNoContent, rr.Body.String())
		}
	})

	t.Run("GET counter value", func(t *testing.T) {
		// First increment
		incrementHandler := api.CounterIncrementHandler(srv)
		payload := map[string]interface{}{
			"path":  "ROOT",
			"key":   "score",
			"delta": 10,
		}
		doRequest(t, incrementHandler, "POST", "/api/counter/increment", payload)

		// Then GET
		getHandler := api.CounterGetHandler(srv)
		rr := doRequest(t, getHandler, "GET", "/api/counter?path=ROOT&key=score", nil)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("GET counter returned wrong status: got %v want %v", status, http.StatusOK)
		}

		var resp map[string]interface{}
		if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if val, ok := resp["value"].(float64); !ok || val != 10 {
			t.Errorf("GET counter returned wrong value: got %v want %v", resp["value"], 10)
		}
	})

	t.Run("Combined counter handler", func(t *testing.T) {
		handler := api.CounterHandler(srv)

		// POST to increment
		payload := map[string]interface{}{
			"path":  "ROOT",
			"key":   "combo",
			"delta": 3,
		}
		rr := doRequest(t, handler, "POST", "/api/counter", payload)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("POST counter returned wrong status: got %v want %v", status, http.StatusOK)
		}

		var resp map[string]interface{}
		if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if val, ok := resp["value"].(float64); !ok || val != 3 {
			t.Errorf("POST counter returned wrong value: got %v want %v", resp["value"], 3)
		}

		// GET to read
		rr = doRequest(t, handler, "GET", "/api/counter?path=ROOT&key=combo", nil)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("GET counter returned wrong status: got %v want %v", status, http.StatusOK)
		}

		if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if val, ok := resp["value"].(float64); !ok || val != 3 {
			t.Errorf("GET counter returned wrong value: got %v want %v", resp["value"], 3)
		}
	})
}
