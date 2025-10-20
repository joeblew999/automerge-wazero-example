// Package integration provides integration tests for the Automerge WASI server
package integration

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/joeblew999/automerge-wazero-example/pkg/client"
)

// TestServer represents a running test server instance
type TestServer struct {
	Port       int
	StorageDir string
	UserID     string
	cmd        *exec.Cmd
	Client     *client.Client
}

// StartTestServer starts a server instance for testing
func StartTestServer(t *testing.T, port int, userID string) *TestServer {
	t.Helper()

	// Create temporary storage directory
	storageDir := filepath.Join(t.TempDir(), userID)
	if err := os.MkdirAll(storageDir, 0755); err != nil {
		t.Fatalf("Failed to create storage dir: %v", err)
	}

	// Build the server if not already built
	buildCmd := exec.Command("make", "build-wasi")
	buildCmd.Dir = filepath.Join("..", "..", "..") // Go to repo root
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build WASI module: %v\n%s", err, output)
	}

	// Start the server
	cmd := exec.Command("go", "run", "main.go")
	cmd.Dir = filepath.Join("..", "..", "cmd", "server")
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("PORT=%d", port),
		fmt.Sprintf("STORAGE_DIR=%s", storageDir),
		fmt.Sprintf("USER_ID=%s", userID),
	)

	// Capture output for debugging
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	ts := &TestServer{
		Port:       port,
		StorageDir: storageDir,
		UserID:     userID,
		cmd:        cmd,
		Client:     client.New(fmt.Sprintf("http://localhost:%d", port)),
	}

	// Wait for server to be ready
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := ts.Client.WaitForReady(ctx); err != nil {
		ts.Stop(t)
		t.Fatalf("Server failed to start: %v", err)
	}

	t.Logf("Started server %s on port %d (storage: %s)", userID, port, storageDir)

	return ts
}

// Stop stops the test server
func (ts *TestServer) Stop(t *testing.T) {
	t.Helper()

	if ts.cmd != nil && ts.cmd.Process != nil {
		t.Logf("Stopping server %s (port %d)", ts.UserID, ts.Port)
		if err := ts.cmd.Process.Kill(); err != nil {
			t.Logf("Warning: failed to kill server: %v", err)
		}
		ts.cmd.Wait() // Clean up zombie process
	}
}

// AssertText verifies the server's text content matches expected
func (ts *TestServer) AssertText(t *testing.T, expected string) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	actual, err := ts.Client.GetText(ctx)
	if err != nil {
		t.Fatalf("GetText() failed: %v", err)
	}

	if actual != expected {
		t.Errorf("Text mismatch:\n  got:  %q\n  want: %q", actual, expected)
	}
}

// SetText sets the server's text content
func (ts *TestServer) SetText(t *testing.T, text string) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := ts.Client.SetText(ctx, text); err != nil {
		t.Fatalf("SetText() failed: %v", err)
	}
}

// GetSnapshot downloads the server's document snapshot
func (ts *TestServer) GetSnapshot(t *testing.T) []byte {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	snapshot, err := ts.Client.GetDocument(ctx)
	if err != nil {
		t.Fatalf("GetDocument() failed: %v", err)
	}

	return snapshot
}

// MergeSnapshot merges another server's snapshot into this one
func (ts *TestServer) MergeSnapshot(t *testing.T, snapshot []byte) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := ts.Client.MergeDocument(ctx, snapshot); err != nil {
		t.Fatalf("MergeDocument() failed: %v", err)
	}
}

// VerifyAutomergeMagicBytes checks if a snapshot has valid Automerge header
func VerifyAutomergeMagicBytes(t *testing.T, snapshot []byte) {
	t.Helper()

	if len(snapshot) < 4 {
		t.Fatalf("Snapshot too small: %d bytes", len(snapshot))
	}

	// Automerge magic bytes: 0x85 0x6f 0x4a 0x83
	expected := []byte{0x85, 0x6f, 0x4a, 0x83}
	actual := snapshot[0:4]

	for i := 0; i < 4; i++ {
		if actual[i] != expected[i] {
			t.Errorf("Magic byte[%d] = 0x%02x, want 0x%02x", i, actual[i], expected[i])
		}
	}

	if t.Failed() {
		t.Fatalf("Invalid Automerge magic bytes: got %x, want %x", actual, expected)
	}
}
