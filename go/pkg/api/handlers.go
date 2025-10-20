package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/joeblew999/automerge-wazero-example/pkg/server"
)

// TextPayload represents the JSON payload for text updates
type TextPayload struct {
	Text string `json:"text"`
}

// TextHandler handles GET and POST /api/text requests
func TextHandler(srv *server.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		switch r.Method {
		case http.MethodGet:
			text, err := srv.GetText(ctx)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "text/plain")
			w.Write([]byte(text))

		case http.MethodPost:
			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "Failed to read body", http.StatusBadRequest)
				return
			}

			var payload TextPayload
			if err := json.Unmarshal(body, &payload); err != nil {
				http.Error(w, "Invalid JSON", http.StatusBadRequest)
				return
			}

			if err := srv.SetText(ctx, payload.Text); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			// Broadcast update to SSE clients
			srv.Broadcast(payload.Text)

			w.WriteHeader(http.StatusNoContent)

		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

// StreamHandler handles GET /api/stream (SSE) requests
func StreamHandler(srv *server.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set SSE headers
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
			return
		}

		// Create client channel
		clientChan := make(chan string, 10)

		srv.AddClient(clientChan)
		defer func() {
			srv.RemoveClient(clientChan)
			close(clientChan)
		}()

		// Send initial snapshot
		text, err := srv.GetText(r.Context())
		if err == nil {
			data, _ := json.Marshal(map[string]string{"text": text})
			fmt.Fprintf(w, "event: snapshot\ndata: %s\n\n", data)
			flusher.Flush()
		}

		// Listen for updates
		for {
			select {
			case text, ok := <-clientChan:
				if !ok {
					return
				}
				data, _ := json.Marshal(map[string]string{"text": text})
				fmt.Fprintf(w, "event: update\ndata: %s\n\n", data)
				flusher.Flush()
			case <-r.Context().Done():
				return
			}
		}
	}
}

// DocHandler handles GET /api/doc (download doc.am snapshot)
func DocHandler(srv *server.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		ctx := r.Context()
		data, err := srv.GetSnapshot(ctx)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to save document: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s-doc.am\"", srv.UserID()))
		w.Write(data)
		log.Printf("[%s] Sent doc.am (%d bytes)", srv.UserID(), len(data))
	}
}

// MergeHandler handles POST /api/merge (CRDT merge)
func MergeHandler(srv *server.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		ctx := r.Context()

		// Read the incoming doc.am binary
		otherDoc, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read body", http.StatusBadRequest)
			return
		}

		if len(otherDoc) == 0 {
			http.Error(w, "Empty document", http.StatusBadRequest)
			return
		}

		log.Printf("[%s] Received doc.am to merge (%d bytes)", srv.UserID(), len(otherDoc))

		// Merge the documents
		if err := srv.Merge(ctx, otherDoc); err != nil {
			http.Error(w, fmt.Sprintf("Merge failed: %v", err), http.StatusInternalServerError)
			return
		}

		// After merge, get the new text and broadcast
		text, err := srv.GetText(ctx)
		if err == nil {
			srv.Broadcast(text)
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Merged successfully! New text: %s", text)
		log.Printf("[%s] Merge complete, new text: %s", srv.UserID(), text)
	}
}
