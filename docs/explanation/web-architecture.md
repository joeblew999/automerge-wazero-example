# Web Folder Architecture (1:1 Mapping)

The `web/` folder follows the same 1:1 file mapping principle as the rest of the codebase.

## Structure

```
web/
├── index.html          # Main entry point with tab navigation
├── css/
│   └── main.css        # Shared styles (600+ lines, gradient UI)
├── js/                 # 1:1 with go/pkg/automerge/*.go
│   ├── app.js          # Orchestrator (tab switching, SSE, init)
│   ├── text.js         # Maps to text.go (M0)
│   ├── map.js          # Maps to map.go (M0) - TODO
│   ├── list.js         # Maps to list.go (M0) - TODO
│   ├── counter.js      # Maps to counter.go (M0) - TODO
│   ├── history.js      # Maps to history.go (M0) - TODO
│   ├── sync.js         # Maps to sync.go (M1) ✅ COMPLETE
│   └── richtext.js     # Maps to richtext.go (M2) ✅ COMPLETE
└── components/         # 1:1 with go/pkg/api/*.go
    ├── text.html       # Maps to api/text.go (M0)
    ├── sync.html       # Maps to api/sync.go (M1) ✅ COMPLETE
    └── richtext.html   # Maps to api/richtext.go (M2) ✅ COMPLETE
```

## 1:1 Mapping Table

| Go API Handler | Web Component | Web JS Module | Status |
|----------------|---------------|---------------|--------|
| api/text.go | text.html | text.js | ✅ M0 |
| api/map.go | map.html | map.js | 🚧 TODO |
| api/list.go | list.html | list.js | 🚧 TODO |
| api/counter.go | counter.html | counter.js | 🚧 TODO |
| api/history.go | history.html | history.js | 🚧 TODO |
| api/sync.go | sync.html | sync.js | ✅ M1 |
| api/richtext.go | richtext.html | richtext.js | ✅ M2 |

## Adding New Web Components

When creating a new web component, maintain 1:1 mapping:

**Example: Adding Map component**

1. Create `web/components/map.html` (UI template)
2. Create `web/js/map.js` (exports `class MapComponent`)
3. Update `web/js/app.js` to import and initialize
4. Update `Makefile` variables:
   ```makefile
   WEB_JS = ... $(WEB_DIR)/js/map.js
   WEB_COMPONENTS = ... $(WEB_DIR)/components/map.html
   ```
5. Run `make verify-web` to ensure structure is correct

## Web Module Pattern

**Every `web/js/*.js` file exports a class**:

```javascript
// web/js/sync.js (M1 example)
export class SyncComponent {
    constructor() {
        this.peerID = null;
    }

    init() {
        // Setup event listeners
        // Initialize UI
    }

    async initSync() {
        // Call /api/sync endpoint
    }

    destroy() {
        // Cleanup when switching tabs
    }
}
```

**Orchestrated by app.js**:

```javascript
// web/js/app.js
import { SyncComponent } from './sync.js';

class App {
    constructor() {
        this.components = {
            sync: new SyncComponent(),
            // ...
        };
    }

    switchTab(tabName) {
        this.components[tabName].init();
    }
}
```

## Verification

```bash
make verify-web
```

**Output**:
```
🔍 Verifying web folder structure (1:1 mapping)...

Checking required files:
  ✅ web/index.html
  ✅ web/css/main.css
  ✅ web/js/app.js
  ✅ web/js/text.js
  ✅ web/js/sync.js
  ✅ web/js/richtext.js
  ✅ web/components/text.html
  ✅ web/components/sync.html
  ✅ web/components/richtext.html
  ✅ ui/vendor/automerge.js

Checking Automerge.js:
  ✅ ui/vendor/automerge.js (3.4M)
  ✅ web/index.html references /vendor/automerge.js

✅ Web folder structure valid!
```

## See Also

- [Architecture Guide](architecture.md) - Complete 7-layer architecture
- [Build Automerge.js](../how-to/build-automerge-js.md) - Building from source
- [CLAUDE.md](../../CLAUDE.md) - Section 0.2 for 1:1 mapping principles
