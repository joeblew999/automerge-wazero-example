// ==============================================================================
// Layer 6: HTTP API - Health Check Endpoints
// ==============================================================================
// ARCHITECTURE: This is the HTTP protocol layer (Layer 6/7).
//
// RESPONSIBILITIES:
// - HTTP request parsing (query params, headers)
// - HTTP response formatting (JSON, status codes, headers)
// - Input validation (HTTP-level)
// - Protocol translation (HTTP ↔ Go function calls)
//
// DEPENDENCIES:
// - Layer 5: pkg/server (business logic, state management)
//
// RELATED FILES:
// - pkg/server/server.go (Layer 5 - stateful operations)
// - No corresponding Rust/WASM layers (HTTP-only concern)
//
// NOTES:
// - Health checks are Layer 6-only (no CRDT operations involved)
// - Follows Kubernetes health check conventions
// - /healthz = liveness (is process alive?)
// - /readyz = readiness (can it accept traffic?)
// - /healthz/live = alias for liveness
// - /healthz/ready = alias for readiness
// ==============================================================================

package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/joeblew999/automerge-wazero-example/pkg/server"
)

// HealthResponse is the JSON response for health check endpoints
type HealthResponse struct {
	Status    string                 `json:"status"`              // "ok" or "error"
	Timestamp time.Time              `json:"timestamp"`           // Current server time
	Service   string                 `json:"service"`             // Service identifier
	Version   string                 `json:"version,omitempty"`   // Optional version info
	Details   map[string]interface{} `json:"details,omitempty"`   // Optional detailed status
}

// LivenessHandler implements Kubernetes liveness probe endpoint.
//
// Endpoint: GET /healthz or GET /healthz/live
//
// This checks if the process is alive and running. If this fails, the
// container should be restarted.
//
// Returns:
//   - 200 OK: Process is alive
//   - 500 Internal Server Error: Process is unhealthy (rare)
//
// Example response:
//
//	{
//	    "status": "ok",
//	    "timestamp": "2025-10-21T14:55:00Z",
//	    "service": "automerge-wazero"
//	}
//
// Status: ✅ Implemented
func LivenessHandler(srv *server.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		resp := HealthResponse{
			Status:    "ok",
			Timestamp: time.Now().UTC(),
			Service:   "automerge-wazero",
			Details: map[string]interface{}{
				"check": "liveness",
				"user_id": srv.UserID(),
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}
}

// ReadinessHandler implements Kubernetes readiness probe endpoint.
//
// Endpoint: GET /readyz or GET /healthz/ready
//
// This checks if the service is ready to accept traffic. If this fails,
// the service should be removed from load balancer rotation but NOT restarted.
//
// Returns:
//   - 200 OK: Service is ready to accept traffic
//   - 503 Service Unavailable: Service is not ready (still initializing, or dependencies unavailable)
//
// Example response:
//
//	{
//	    "status": "ok",
//	    "timestamp": "2025-10-21T14:55:00Z",
//	    "service": "automerge-wazero",
//	    "details": {
//	        "check": "readiness",
//	        "document_initialized": true,
//	        "wasm_runtime": "loaded"
//	    }
//	}
//
// Status: ✅ Implemented
func ReadinessHandler(srv *server.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Check if server is ready (document initialized, WASM loaded, etc.)
		ready, details := srv.IsReady()

		status := "ok"
		statusCode := http.StatusOK
		if !ready {
			status = "not_ready"
			statusCode = http.StatusServiceUnavailable
		}

		resp := HealthResponse{
			Status:    status,
			Timestamp: time.Now().UTC(),
			Service:   "automerge-wazero",
			Details:   details,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(resp)
	}
}

// HealthHandler is a simple combined health check (liveness + readiness).
//
// Endpoint: GET /health
//
// This is a convenience endpoint that returns both liveness and readiness status.
// Useful for simple health monitoring systems that don't distinguish between the two.
//
// Returns:
//   - 200 OK: Service is alive and ready
//   - 503 Service Unavailable: Service is alive but not ready
//   - 500 Internal Server Error: Service is unhealthy
//
// Example response:
//
//	{
//	    "status": "ok",
//	    "timestamp": "2025-10-21T14:55:00Z",
//	    "service": "automerge-wazero",
//	    "details": {
//	        "liveness": "ok",
//	        "readiness": "ok",
//	        "document_initialized": true
//	    }
//	}
//
// Status: ✅ Implemented
func HealthHandler(srv *server.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Check readiness
		ready, readyDetails := srv.IsReady()

		status := "ok"
		statusCode := http.StatusOK
		if !ready {
			status = "degraded"
			statusCode = http.StatusServiceUnavailable
		}

		// Combine details
		details := map[string]interface{}{
			"liveness":  "ok",
			"readiness": status,
		}
		for k, v := range readyDetails {
			details[k] = v
		}

		resp := HealthResponse{
			Status:    status,
			Timestamp: time.Now().UTC(),
			Service:   "automerge-wazero",
			Details:   details,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(resp)
	}
}
