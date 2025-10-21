# M2 RichText Marks - Playwright Test Plan

**Module**: M2 Rich Text Marks
**Mapped Files**:
- `rust/automerge_wasi/src/richtext.rs`
- `go/pkg/wazero/richtext.go`
- `go/pkg/automerge/richtext.go`
- `go/pkg/server/richtext.go`
- `go/pkg/api/richtext.go`
- `web/js/richtext.js`

## Test Scenario

Test the complete rich text marks flow from browser → HTTP → WASM → CRDT.

## Test Steps

1. **Navigate to Server**
   - URL: `http://localhost:8080/ui`
   - Verify page loads

2. **Add Text to Document**
   - POST `/api/text` with `{"text":"Hello World"}`
   - Verify 204 No Content response

3. **Apply Bold Mark**
   - POST `/api/richtext/mark`
   - Payload: `{"path":"ROOT.content","name":"bold","value":"true","start":0,"end":5,"expand":"none"}`
   - Verify 204 No Content response

4. **Get Marks at Position**
   - GET `/api/richtext/marks?path=ROOT.content&pos=2`
   - Verify response contains mark

5. **Apply Multiple Marks**
   - Apply italic mark [6, 11)
   - Get marks at pos=7
   - Verify only italic mark returned (not bold)

6. **Remove Mark**
   - POST `/api/richtext/unmark`
   - Payload: `{"path":"ROOT.content","name":"bold","start":0,"end":5,"expand":"none"}`
   - Verify 204 No Content response

7. **Verify Mark Removed**
   - GET marks at pos=2
   - Verify no marks returned

## Expected Results

- ✅ All HTTP endpoints respond correctly
- ✅ Marks applied at correct positions
- ✅ Mark queries return accurate results
- ✅ Marks can be removed
- ✅ No JSON parsing errors

## Test Data

**Mark Payload**:
```json
{
  "path": "ROOT.content",
  "name": "bold",
  "value": "true",
  "start": 0,
  "end": 5,
  "expand": "none"
}
```

**Unmark Payload**:
```json
{
  "path": "ROOT.content",
  "name": "bold",
  "start": 0,
  "end": 5,
  "expand": "none"
}
```

## Success Criteria

- HTTP 204 for mark/unmark
- HTTP 200 with valid JSON for get marks
- Marks accurately reflect character positions
- No garbage bytes in JSON (M2 bug was fixed!)
- Multiple marks don't interfere with each other
