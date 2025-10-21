# M1 Sync Protocol - Playwright Test Plan

**Module**: M1 Sync Protocol
**Mapped Files**:
- `rust/automerge_wasi/src/sync.rs`
- `go/pkg/wazero/sync.go`
- `go/pkg/automerge/sync.go`
- `go/pkg/server/sync.go`
- `go/pkg/api/sync.go`
- `web/js/sync.js`

## Test Scenario

Test the complete sync protocol flow from browser → HTTP → WASM → CRDT.

## Test Steps

1. **Navigate to Server**
   - URL: `http://localhost:8080/ui`
   - Verify page loads

2. **Initialize Sync via HTTP**
   - POST `/api/sync` with peer_id
   - Verify 200 OK response
   - Verify JSON response has `has_more` field

3. **Send Sync Message**
   - POST `/api/sync` with same peer_id
   - Verify response structure
   - Check for sync message in response

4. **Verify Server State**
   - Server maintains per-peer sync state
   - Multiple peers can sync independently

## Expected Results

- ✅ HTTP endpoint responds correctly
- ✅ Per-peer sync state created
- ✅ Sync messages exchanged
- ✅ No errors in browser console

## Test Data

```json
{
  "peer_id": "playwright-test-peer",
  "message": ""
}
```

## Success Criteria

- HTTP 200 response
- Valid JSON with `has_more` boolean
- Optional `message` field (base64)
- Server logs show sync state creation
