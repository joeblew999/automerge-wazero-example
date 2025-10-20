# Test Data for Automerge Go Packages

This directory contains all test data organized by test type following Go conventions.

## Directory Structure

```
go/testdata/
├── unit/           # Unit test data (Go package tests)
│   ├── snapshots/  # Binary .am files for testing
│   ├── expected/   # Expected output files
│   └── scripts/    # Test data generation scripts
├── integration/    # Integration test scripts
│   └── test_merge.sh
└── e2e/            # End-to-end test artifacts
    └── screenshots/  # Playwright test screenshots
```

## Unit Test Data (`unit/`)

Binary Automerge document snapshots (`.am` files) used by Go package tests.

### Generating Unit Test Data

Test snapshots are generated using the WASM module:

**Manual Generation**:
```bash
# Start the server
cd go/cmd/server
go run main.go

# In another terminal, create test data
curl -X POST http://localhost:8080/api/text \
  -H 'Content-Type: application/json' \
  -d '{"text":"Hello, World!"}'

# Download the snapshot
curl http://localhost:8080/api/doc > ../testdata/unit/snapshots/hello-world.am
```

**Using the Generation Script**:
```bash
cd go/testdata/unit/scripts
./generate_test_data.sh
```

### Unit Test Snapshots

| File | Description | Size |
|------|-------------|------|
| `empty.am` | Empty document (root + text object) | ~50 bytes |
| `hello-world.am` | Document with "Hello, World!" | ~200 bytes |
| `simple-text.am` | Simple ASCII text | ~150 bytes |
| `unicode-text.am` | Unicode + emoji testing | ~300 bytes |
| `large-text.am` | ~10KB text (performance) | ~10KB |

## Integration Tests (`integration/`)

Bash scripts for end-to-end CRDT testing.

### test_merge.sh

Tests the complete CRDT merge scenario:
1. Start two independent servers (Alice & Bob)
2. Each creates different content offline
3. Download Alice's `doc.am`
4. Merge into Bob's server via `/api/merge`
5. Verify CRDT properties (no data loss)

**Run**:
```bash
cd go/testdata/integration
./test_merge.sh
```

**Or use Makefile**:
```bash
make test-two-laptops  # Alternative way to run
```

## E2E Test Artifacts (`e2e/`)

Screenshots and artifacts from Playwright MCP end-to-end tests.

### Screenshots

| File | Description |
|------|-------------|
| `playwright-test-save.png` | UI after saving text |
| `playwright-test-final.png` | Complete test state |

**Note**: `.playwright-mcp/testdata/` may also contain screenshots from recent test runs.

## Adding New Test Data

### For Unit Tests

1. Generate snapshot using server or WASM
2. Save to `unit/snapshots/` with descriptive name
3. If needed, save expected output to `unit/expected/`
4. Update this README with description
5. Add test case in `../pkg/automerge/*_test.go`

### For Integration Tests

1. Create bash script in `integration/`
2. Follow pattern from `test_merge.sh`
3. Make executable: `chmod +x script.sh`
4. Add description here

### For E2E Tests

1. Run Playwright MCP test
2. Copy screenshots from `.playwright-mcp/testdata/` to `e2e/screenshots/`
3. Update this README with screenshot descriptions

## Note on Binary Files

All `.am` files are binary Automerge snapshots:
- ✅ Committed to git
- ✅ Small size (<100KB typically)
- ✅ Start with magic bytes: `85 6f 4a 83`

---

**Last Updated**: 2025-10-20 (Reorganized into unit/integration/e2e)
