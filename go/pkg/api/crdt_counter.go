package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/joeblew999/automerge-wazero-example/pkg/server"
)

// CounterPayload represents the JSON payload for counter operations
type CounterPayload struct {
	Path  string `json:"path"` // Path to the object containing the counter
	Key   string `json:"key"`  // Key of the counter
	Delta int64  `json:"delta"` // Delta to increment (can be negative)
}

// CounterResponse represents the response for counter operations
type CounterResponse struct {
	Value int64 `json:"value"`
}

// CounterIncrementHandler handles POST /api/counter/increment {path, key, delta} - Increment counter
func CounterIncrementHandler(srv *server.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		ctx := r.Context()
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read body", http.StatusBadRequest)
			return
		}

		var payload CounterPayload
		if err := json.Unmarshal(body, &payload); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		if payload.Path == "" || payload.Key == "" {
			http.Error(w, "Missing path or key", http.StatusBadRequest)
			return
		}

		if err := srv.IncrementCounter(ctx, parsePathString(payload.Path), payload.Key, payload.Delta); err != nil {
			http.Error(w, fmt.Sprintf("Failed to increment: %v", err), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
		log.Printf("Counter INCREMENT: path=%s, key=%s, delta=%d", payload.Path, payload.Key, payload.Delta)
	}
}

// CounterGetHandler handles GET /api/counter?path=ROOT&key=clicks - Get counter value
func CounterGetHandler(srv *server.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		ctx := r.Context()
		path := r.URL.Query().Get("path")
		key := r.URL.Query().Get("key")

		if path == "" || key == "" {
			http.Error(w, "Missing path or key parameter", http.StatusBadRequest)
			return
		}

		value, err := srv.GetCounter(ctx, parsePathString(path), key)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to get counter: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(CounterResponse{Value: value})
	}
}

// CounterHandler handles both GET and POST for convenience
func CounterHandler(srv *server.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			// GET /api/counter?path=ROOT&key=clicks
			CounterGetHandler(srv)(w, r)

		case http.MethodPost:
			// POST /api/counter {path, key, delta}
			ctx := r.Context()
			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "Failed to read body", http.StatusBadRequest)
				return
			}

			var payload CounterPayload
			if err := json.Unmarshal(body, &payload); err != nil {
				http.Error(w, "Invalid JSON", http.StatusBadRequest)
				return
			}

			if payload.Path == "" || payload.Key == "" {
				http.Error(w, "Missing path or key", http.StatusBadRequest)
				return
			}

			// If delta not specified, default to 1
			delta := payload.Delta
			if delta == 0 {
				deltaStr := r.URL.Query().Get("delta")
				if deltaStr != "" {
					parsedDelta, err := strconv.ParseInt(deltaStr, 10, 64)
					if err == nil {
						delta = parsedDelta
					} else {
						delta = 1
					}
				} else {
					delta = 1
				}
			}

			if err := srv.IncrementCounter(ctx, parsePathString(payload.Path), payload.Key, delta); err != nil {
				http.Error(w, fmt.Sprintf("Failed to increment: %v", err), http.StatusInternalServerError)
				return
			}

			// Return current value after increment
			value, err := srv.GetCounter(ctx, parsePathString(payload.Path), payload.Key)
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to get updated counter: %v", err), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(CounterResponse{Value: value})
			log.Printf("Counter INCREMENT: path=%s, key=%s, delta=%d, new_value=%d", payload.Path, payload.Key, delta, value)

		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}
