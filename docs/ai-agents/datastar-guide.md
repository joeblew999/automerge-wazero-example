# Datastar Integration Guide (M3)

**For AI Agents**: How to integrate Datastar with the Automerge WASI project

**Status**: Planning phase - M3 milestone  
**Reference**: https://data-star.dev/

---

## Summary: Datastar as Layer 7 (Separate UI Layer)

**Architecture Principle**: Datastar is a **separate UI layer**, not a refactoring.

**Recommended: Hybrid Co-existence**

Keep current vanilla JS UI (Layer 6), add Datastar UI (Layer 7) as parallel frontend.

### Perfect 7-Layer Architecture

```
Layer 1: Rust WASI Exports    (rust/automerge_wasi/src/counter.rs)
Layer 2: Go FFI Wrappers      (go/pkg/wazero/crdt_counter.go)
Layer 3: Go CRDT API          (go/pkg/automerge/crdt_counter.go)
Layer 4: Go Server            (go/pkg/server/crdt_counter.go)
Layer 5: HTTP API             (go/pkg/api/crdt_counter.go)
Layer 6: Vanilla JS UI        (web/js/crdt_counter.js) ← Current
Layer 7: Datastar UI          (datastar/components/counter.html) ← NEW
```

**Key Point**: Layers 1-5 are **unchanged**. Same HTTP endpoints serve both UIs!

**Why This is Best**:
✅ Zero risk - both UIs work independently
✅ Perfect layer separation maintained
✅ Easy A/B comparison
✅ Learn Datastar incrementally
✅ Can remove Layer 7 without breaking anything

**Key Insight**: Datastar reduces ~1500 lines of JS to ~400 lines of declarative HTML!

---

## Code Comparison: Before vs After

### Before (Current): `web/js/crdt_counter.js` - 192 lines

```javascript
class CounterComponent {
    constructor() {
        this.value = 0;
    }

    async getValue() {
        const res = await fetch('/api/counter?path=ROOT&key=counter');
        const data = await res.json();
        // Manual DOM update
        document.getElementById('counter-value').textContent = data.value;
    }

    async increment() {
        await fetch('/api/counter', {
            method: 'POST',
            body: JSON.stringify({path: 'ROOT', key: 'counter', delta: 1})
        });
        await this.getValue(); // Manual refresh
    }
    
    // ... 150 more lines of boilerplate
}
```

### After (Datastar): `datastar/components/counter.html` - ~50 lines

```html
<div data-store='{"value": 0}' 
     data-on-load="$$get('/ds/counter/init')">

  <!-- Reactive - auto-updates! -->
  <div data-text="$value"></div>

  <!-- One-line action -->
  <button data-on-click="$$post('/ds/counter/increment', {delta: 1})">
    ➕ Increment
  </button>

  <!-- SSE auto-merges state -->
  <div data-on-load="$$get('/ds/counter/stream')"></div>
</div>
```

**73% code reduction** + reactive + SSE built-in!

---

## Recommended File Structure

```
Current (keep):
web/
├── index.html
├── js/crdt_*.js          # 8 files, ~1500 lines total
└── components/*.html

NEW (add for M3):
datastar/
├── index.html            # Datastar entry
└── components/
    ├── counter.html      # Datastar version
    ├── text.html
    └── ...
```

**Routes**:
```
http://localhost:8080/           → web/index.html (vanilla JS)
http://localhost:8080/datastar   → datastar/index.html (Datastar)
```

---

## Implementation Steps

### 1. Add Dependency (5 min)
```bash
go get github.com/delaneyj/datastar
```

### 2. Create Structure (5 min)
```bash
mkdir -p datastar/components
touch datastar/index.html
touch datastar/components/counter.html
```

### 3. Add Route (10 min)
```go
// go/cmd/server/main.go
http.HandleFunc("/datastar", serveDatastarIndex)
http.HandleFunc("/ds/counter/init", DatastarCounterInit(srv))
http.HandleFunc("/ds/counter/increment", DatastarCounterIncrement(srv))
```

### 4. Convert One Component (2-3 hours)
Pick Counter (simplest) as proof-of-concept.

### 5. Test Side-by-Side (30 min)
Compare vanilla JS vs Datastar versions.

**Total**: 3-4 hours for working POC

---

## Key Datastar Concepts

| Concept | Purpose | Example |
|---------|---------|---------|
| `data-store` | Reactive state (signals) | `data-store='{"count": 0}'` |
| `data-text` | Bind signal to text | `data-text="$count"` |
| `$$get()` | Fetch via SSE | `$$get('/ds/init')` |
| `$$post()` | POST + merge response | `$$post('/ds/incr')` |
| `datastar.NewSSE()` | Go SSE helper | `sse.MergeStore({...})` |

---

## Risks & Mitigation

| Risk | Mitigation |
|------|------------|
| Learning curve | Start with 1 component |
| Datastar bugs | Keep vanilla JS (easy rollback) |
| SSE incompatibility | Test both formats |

**Risk Level**: ⚠️ Low (hybrid = zero-risk deployment)

---

## Decision: Do We Need Automerge.js?

**Answer**: NO (for M3)

- Server handles all CRDT operations (Rust WASM)
- Datastar just improves UI reactivity
- Automerge.js only needed for offline-first (M4+)

---

## Next Steps

1. Read this guide ✅
2. Decide: Proof-of-concept or skip M3?
3. If yes: Start with Counter component
4. Test with Playwright
5. Document findings
6. Decide: Migrate all or stay hybrid?

---

**Recommended**: Start with Counter POC (3-4 hours) to validate before committing to full migration.

**References**:
- https://data-star.dev/
- https://github.com/delaneyj/datastar

