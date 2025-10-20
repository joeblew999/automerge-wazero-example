#!/bin/bash
# Generate test data snapshots for Automerge Go package tests
#
# Prerequisites:
# - Rust WASM module must be built: make build-wasi (from repo root)
# - Server must NOT be running (this script will start/stop it)
#
# Usage: ./generate_test_data.sh

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TESTDATA_DIR="$(dirname "$SCRIPT_DIR")"
SNAPSHOTS_DIR="$TESTDATA_DIR/snapshots"
GO_DIR="$(dirname "$TESTDATA_DIR")"
SERVER_DIR="$GO_DIR/cmd/server"

echo "🔧 Test Data Generator for Automerge Go"
echo "========================================"
echo ""

# Check if WASM module exists
WASM_PATH="$GO_DIR/../rust/automerge_wasi/target/wasm32-wasip1/release/automerge_wasi.wasm"
if [ ! -f "$WASM_PATH" ]; then
    echo "❌ Error: WASM module not found at $WASM_PATH"
    echo "   Run 'make build-wasi' from the repository root first."
    exit 1
fi

echo "✅ WASM module found"

# Check if server is already running
if lsof -Pi :8080 -sTCP:LISTEN -t >/dev/null 2>&1 ; then
    echo "❌ Error: Port 8080 is already in use. Please stop the server first."
    exit 1
fi

# Build the server
echo "🔨 Building server..."
cd "$SERVER_DIR"
go build -o /tmp/automerge-test-server . || {
    echo "❌ Failed to build server"
    exit 1
}

# Function to start server and wait for it to be ready
start_server() {
    local storage_dir="$1"

    # Clean up any existing doc.am
    rm -f "$storage_dir/doc.am"

    # Start server in background
    STORAGE_DIR="$storage_dir" PORT=8080 /tmp/automerge-test-server > /tmp/automerge-test-server.log 2>&1 &
    SERVER_PID=$!

    # Wait for server to be ready
    echo "⏳ Waiting for server to start..."
    for i in {1..10}; do
        if curl -s http://localhost:8080/api/text > /dev/null 2>&1; then
            echo "✅ Server ready (PID $SERVER_PID)"
            return 0
        fi
        sleep 0.5
    done

    echo "❌ Server failed to start"
    cat /tmp/automerge-test-server.log
    kill $SERVER_PID 2>/dev/null || true
    exit 1
}

# Function to stop server
stop_server() {
    if [ -n "$SERVER_PID" ]; then
        echo "🛑 Stopping server (PID $SERVER_PID)..."
        kill $SERVER_PID 2>/dev/null || true
        wait $SERVER_PID 2>/dev/null || true
    fi
}

# Ensure server is stopped on exit
trap stop_server EXIT

mkdir -p "$SNAPSHOTS_DIR"

echo ""
echo "📝 Generating test snapshots..."
echo ""

# 1. Empty document
echo "1️⃣  Generating empty.am..."
start_server "$SNAPSHOTS_DIR"
sleep 0.2
curl -s http://localhost:8080/api/doc > "$SNAPSHOTS_DIR/empty.am"
echo "   ✅ Created empty.am ($(wc -c < "$SNAPSHOTS_DIR/empty.am") bytes)"
stop_server
sleep 0.5

# 2. Hello World
echo "2️⃣  Generating hello-world.am..."
start_server "$SNAPSHOTS_DIR"
curl -s -X POST http://localhost:8080/api/text \
    -H 'Content-Type: application/json' \
    -d '{"text":"Hello, World!"}' > /dev/null
sleep 0.2
curl -s http://localhost:8080/api/doc > "$SNAPSHOTS_DIR/hello-world.am"
echo "   ✅ Created hello-world.am ($(wc -c < "$SNAPSHOTS_DIR/hello-world.am") bytes)"
stop_server
sleep 0.5

# 3. Simple text
echo "3️⃣  Generating simple-text.am..."
start_server "$SNAPSHOTS_DIR"
curl -s -X POST http://localhost:8080/api/text \
    -H 'Content-Type: application/json' \
    -d '{"text":"The quick brown fox jumps over the lazy dog."}' > /dev/null
sleep 0.2
curl -s http://localhost:8080/api/doc > "$SNAPSHOTS_DIR/simple-text.am"
echo "   ✅ Created simple-text.am ($(wc -c < "$SNAPSHOTS_DIR/simple-text.am") bytes)"
stop_server
sleep 0.5

# 4. Unicode text
echo "4️⃣  Generating unicode-text.am..."
start_server "$SNAPSHOTS_DIR"
curl -s -X POST http://localhost:8080/api/text \
    -H 'Content-Type: application/json' \
    -d '{"text":"Hello 世界! 🌍🚀 Emoji test: ✅❌🎉"}' > /dev/null
sleep 0.2
curl -s http://localhost:8080/api/doc > "$SNAPSHOTS_DIR/unicode-text.am"
echo "   ✅ Created unicode-text.am ($(wc -c < "$SNAPSHOTS_DIR/unicode-text.am") bytes)"
stop_server
sleep 0.5

# 5. Large text
echo "5️⃣  Generating large-text.am..."
start_server "$SNAPSHOTS_DIR"
# Generate ~10KB of text
LARGE_TEXT=$(for i in {1..200}; do echo -n "Line $i: The quick brown fox jumps over the lazy dog. "; done)
curl -s -X POST http://localhost:8080/api/text \
    -H 'Content-Type: application/json' \
    -d "{\"text\":\"$LARGE_TEXT\"}" > /dev/null
sleep 0.2
curl -s http://localhost:8080/api/doc > "$SNAPSHOTS_DIR/large-text.am"
echo "   ✅ Created large-text.am ($(wc -c < "$SNAPSHOTS_DIR/large-text.am") bytes)"
stop_server

echo ""
echo "✨ Test data generation complete!"
echo ""
echo "Generated snapshots in: $SNAPSHOTS_DIR"
ls -lh "$SNAPSHOTS_DIR"/*.am 2>/dev/null || echo "  (no files generated)"

# Clean up temp files
rm -f /tmp/automerge-test-server /tmp/automerge-test-server.log
