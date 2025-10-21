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

// ListPayload represents the JSON payload for list operations
type ListPayload struct {
	Path  string `json:"path"`  // Path to the list object
	Index *uint  `json:"index,omitempty"` // Index for insert/get (optional)
	Value string `json:"value"` // Value to insert/push
}

// ListResponse represents the response for list operations
type ListResponse struct {
	Value string `json:"value,omitempty"`
	Length uint32 `json:"length,omitempty"`
}

// ListPushHandler handles POST /api/list/push {path, value} - Append to list
func ListPushHandler(srv *server.Server) http.HandlerFunc {
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

		var payload ListPayload
		if err := json.Unmarshal(body, &payload); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		if payload.Path == "" {
			http.Error(w, "Missing path", http.StatusBadRequest)
			return
		}

		if err := srv.ListPush(ctx, parsePathString(payload.Path), payload.Value); err != nil {
			http.Error(w, fmt.Sprintf("Failed to push: %v", err), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
		log.Printf("List PUSH: path=%s, value=%s", payload.Path, payload.Value)
	}
}

// ListInsertHandler handles POST /api/list/insert {path, index, value} - Insert at index
func ListInsertHandler(srv *server.Server) http.HandlerFunc {
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

		var payload ListPayload
		if err := json.Unmarshal(body, &payload); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		if payload.Path == "" || payload.Index == nil {
			http.Error(w, "Missing path or index", http.StatusBadRequest)
			return
		}

		if err := srv.ListInsert(ctx, parsePathString(payload.Path), *payload.Index, payload.Value); err != nil {
			http.Error(w, fmt.Sprintf("Failed to insert: %v", err), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
		log.Printf("List INSERT: path=%s, index=%d, value=%s", payload.Path, *payload.Index, payload.Value)
	}
}

// ListGetHandler handles GET /api/list?path=ROOT.items&index=0 - Get element at index
func ListGetHandler(srv *server.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		ctx := r.Context()
		path := r.URL.Query().Get("path")
		indexStr := r.URL.Query().Get("index")

		if path == "" || indexStr == "" {
			http.Error(w, "Missing path or index parameter", http.StatusBadRequest)
			return
		}

		index, err := strconv.ParseUint(indexStr, 10, 32)
		if err != nil {
			http.Error(w, "Invalid index parameter", http.StatusBadRequest)
			return
		}

		value, err := srv.ListGet(ctx, parsePathString(path), uint(index))
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to get: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ListResponse{Value: value})
	}
}

// ListDeleteHandler handles DELETE /api/list?path=ROOT.items&index=0 - Delete at index
func ListDeleteHandler(srv *server.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		ctx := r.Context()
		path := r.URL.Query().Get("path")
		indexStr := r.URL.Query().Get("index")

		if path == "" || indexStr == "" {
			http.Error(w, "Missing path or index parameter", http.StatusBadRequest)
			return
		}

		index, err := strconv.ParseUint(indexStr, 10, 32)
		if err != nil {
			http.Error(w, "Invalid index parameter", http.StatusBadRequest)
			return
		}

		if err := srv.ListDelete(ctx, parsePathString(path), uint(index)); err != nil {
			http.Error(w, fmt.Sprintf("Failed to delete: %v", err), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
		log.Printf("List DELETE: path=%s, index=%d", path, index)
	}
}

// ListLenHandler handles GET /api/list/len?path=ROOT.items - Get list length
func ListLenHandler(srv *server.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		ctx := r.Context()
		path := r.URL.Query().Get("path")

		if path == "" {
			http.Error(w, "Missing path parameter", http.StatusBadRequest)
			return
		}

		length, err := srv.ListLen(ctx, parsePathString(path))
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to get length: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ListResponse{Length: length})
	}
}
