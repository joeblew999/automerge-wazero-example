# Test Data for Automerge Go Packages

This directory contains test data for the Automerge Go implementation.

## Directory Structure

- **snapshots/** - Binary Automerge document snapshots (.am files) for testing
- **expected/** - Expected output files for test comparisons
- **scripts/** - Helper scripts for generating and managing test data

## Generating Test Data

Test snapshots must be generated using a working Automerge implementation. Since we're building the implementation, we'll generate these once the WASM module is built.

### Manual Test Data Generation

Once the server is running, you can generate test snapshots like this:

```bash
# Start the server
cd go/cmd/server
go run main.go

# In another terminal, create test data
curl -X POST http://localhost:8080/api/text \
  -H 'Content-Type: application/json' \
  -d '{"text":"Hello, World!"}'

# Download the snapshot
curl http://localhost:8080/api/doc > ../../testdata/snapshots/hello-world.am

# Create an empty document
# (restart server to get fresh document)
curl http://localhost:8080/api/doc > ../../testdata/snapshots/empty.am
```

### Using the Generation Script

```bash
cd testdata/scripts
./generate_test_data.sh
```

## Test Snapshot Descriptions

### snapshots/empty.am
An empty Automerge document with just the root content text object initialized.

### snapshots/hello-world.am
Document containing "Hello, World!" text.

### snapshots/simple-text.am
Document with simple ASCII text for basic testing.

### snapshots/unicode-text.am
Document with Unicode characters (emoji, multi-byte chars) for UTF-8 testing.

### snapshots/large-text.am
Document with ~10KB of text for performance testing.

## Adding New Test Data

1. Create the snapshot using the server or WASM module directly
2. Save it to `snapshots/` with a descriptive name
3. If there's expected output, save it to `expected/`
4. Update this README with a description
5. Add corresponding test cases in `pkg/automerge/*_test.go`

## Note on Binary Files

The `.am` files are binary Automerge snapshots. They should be committed to git and are relatively small (<100KB typically).
