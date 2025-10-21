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

## How Datastar Enhances Existing JavaScript

**Key Insight**: Datastar doesn't replace your JS - it adds reactive DOM updates!

### JavaScript Layer (Stays the Same!)

`web/js/crdt_counter.js` - **Used by both demos**:

```javascript
class CounterComponent {
    constructor() {
        this.value = 0;
    }

    async getValue() {
        const res = await fetch('/api/counter?path=ROOT&key=counter');
        const data = await res.json();

        // OLD WAY (web/components/counter.html):
        // document.getElementById('counter-value').textContent = data.value;

        // NEW WAY (datastar/components/counter.html):
        // Just update the signal - Datastar auto-updates DOM!
        if (window.dsStore) {
            window.dsStore.counter = data.value;
        } else {
            document.getElementById('counter-value').textContent = data.value;
        }
    }

    // ... rest of logic unchanged
}
```

### HTML Layer (Two Versions)

**Vanilla** (`web/components/counter.html`):
```html
<div id="counter-component">
    <div id="counter-value">0</div>
    <button onclick="counterComponent.increment()">Increment</button>
</div>
```

**Datastar** (`datastar/components/counter.html`):
```html
<div id="counter-component" data-store='{"counter": 0}'>
    <!-- Datastar watches signal, auto-renders -->
    <div data-text="$counter"></div>

    <!-- Same JS method, Datastar handles the rest! -->
    <button onclick="counterComponent.increment()">Increment</button>
</div>
```

**Benefit**: Minimal JS changes, massive UX improvement!

---

## Two Parallel Demos (Both Stay Permanently)

### Purpose of Each Demo

| Demo | Purpose | Use Case |
|------|---------|----------|
| **web/** | Testing & Validation | Playwright tests, CI/CD, debugging |
| **datastar/** | Production Showcase | Demos, user-facing, better UX |

### File Structure

```
web/                      # SHARED JavaScript + Vanilla demo
├── index.html            # Demo 1: Vanilla JS entry
├── js/crdt_*.js         # 8 JS files - SHARED by both demos!
│   ├── crdt_counter.js   # Used by web/ AND datastar/
│   ├── crdt_text.js
│   └── ...
└── components/*.html     # Vanilla HTML (manual DOM updates)

datastar/                 # Demo 2: Datastar (only HTML different!)
├── index.html            # Datastar demo entry
└── components/
    ├── counter.html      # Datastar-enhanced HTML (data-* attributes)
    ├── text.html
    └── ...
```

**Smart Serving - No Symlinks!**

Go server serves `/web/js/` to BOTH demos:

```go
// go/cmd/server/main.go
http.Handle("/web/js/", http.StripPrefix("/web/js/",
    http.FileServer(http.Dir("../web/js"))))  // Shared by both!

http.HandleFunc("/", serveVanillaDemo)     // Loads web/index.html
http.HandleFunc("/datastar", serveDatastarDemo)  // Loads datastar/index.html
```

**Both HTML files load the same JS**:
```html
<!-- web/index.html AND datastar/index.html -->
<script src="/web/js/crdt_counter.js"></script>
```

**Key Insight**:
- ✅ **Same JavaScript** served to both demos
- ✅ **Different HTML** - Datastar adds reactive `data-*` attributes
- ✅ **No duplication** - single source of truth for JS
- ✅ **No symlinks** - just smart HTTP routes!

**Routes**:
```
http://localhost:8080/           → web/index.html (testing/validation)
http://localhost:8080/datastar   → datastar/index.html (production demo)
```

**Why Keep Both?**:
- ✅ Vanilla JS is valuable for **automated testing** (Playwright)
- ✅ Vanilla JS is **reference implementation** (known-good)
- ✅ Datastar is **better UX** for demos and users
- ✅ Both use **same HTTP API** (Layers 1-5 unchanged!)

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

