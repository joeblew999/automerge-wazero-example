# Repository Cleanup Analysis

## 🗑️ Files That Should Be Removed

### Build Artifacts (Not in Git)
```bash
# These are generated files that should NOT be committed:
go/server                  # 11MB compiled binary (ignored by .gitignore)
doc.am                     # 483B runtime state (ignored by .gitignore)
package-lock.json          # 503B npm lock file (ignored by .gitignore)
```

**Action**: Already ignored by `.gitignore`, but currently tracked. Should remove from Git:
```bash
git rm --cached go/server doc.am package-lock.json
```

### Potentially Redundant Documentation

#### 1. `COMPLETE.md` (12K) - **CONSIDER REMOVING**
**Content**: Milestone completion summary (Phases 0-10)
**Referenced by**: README.md, TODO.md
**Status**: Historical record of M0 completion
**Recommendation**:
- ⚠️ **ARCHIVE** - Move to `docs/archive/M0_COMPLETE.md`
- It's a snapshot of project state at M0 milestone
- Useful for history but not current workflow

#### 2. `DEMO.md` (11K) - **KEEP (FOR NOW)**
**Content**: Complete demo guide with instructions
**Referenced by**: README.md (line 148: "See DEMO.md for complete documentation")
**Status**: Actively referenced
**Recommendation**:
- ✅ **KEEP** - README.md links to it
- Consider merging into README.md if they're too similar
- Or rename to `USAGE.md` for clarity

#### 3. `AUTOMERGE_JS_VS_RUST_COMPARISON.md` (10K) - **EVALUATE**
**Content**: Version comparison and upgrade strategy
**Referenced by**: None
**Status**: Reference material for version decisions
**Recommendation**:
- 🤔 **EVALUATE** - Check if still relevant
- If not actively used, move to `docs/reference/`
- Or remove if upgrade decisions are documented elsewhere

### Test Files in Root (MOVE TO testdata/)

#### 1. `test_merge.sh` (5.2K) - **MOVE**
**Content**: CRDT merge test script (Alice & Bob concurrent edit scenario)
**Status**: Executable test script with color output
**Recommendation**:
- 🗂️ **MOVE** to `testdata/integration/test_merge.sh`
- This is an integration test, not a root-level script
- Aligns with Go test data organization

#### 2. `test_text_crdt.html` (4.6K) - **EVALUATE/REMOVE**
**Content**: Browser-based Automerge.js Text CRDT test suite
**Status**: Loads Automerge.js 3.1.2 via CDN, runs 6 tests
**Issue**: ⚠️ **Automerge.js was removed from ui/ui.html due to browser WASM errors**
**Recommendation**:
- ❌ **REMOVE** or archive to `testdata/archive/test_text_crdt.html`
- This test relies on Automerge.js browser loading which currently fails
- M0 uses server-side CRDT only (no client-side Automerge.js)
- Can revisit for M2 when client-side Automerge is re-added

#### 3. `test_text_crdt.mjs` (5.1K) - **EVALUATE/REMOVE**
**Content**: Test script using `@automerge/automerge` import
**Status**: ⚠️ **NOT USABLE** - Project doesn't use Node.js, no npm install
**Issue**: Requires `npm install @automerge/automerge` but project is Go+Rust only
**Recommendation**:
- ❌ **REMOVE** or archive to `testdata/archive/test_text_crdt.mjs`
- This project intentionally avoids Node.js/npm for server-side
- UI uses CDN imports, not npm packages
- Server uses WASI (Rust compiled to WASM), not Node.js

### Package Files

#### `package.json` (65B) - **KEEP**
**Content**:
```json
{
  "devDependencies": {
    "@automerge/automerge": "^3.2.0-alpha.0"
  }
}
```
**Status**: Defines Automerge.js dependency for UI
**Recommendation**: ✅ **KEEP** - Needed for `npm install`

#### `package-lock.json` (503B) - **REMOVE FROM GIT**
**Status**: Already in `.gitignore` but was committed before
**Recommendation**:
```bash
git rm --cached package-lock.json
```

---

## 📁 Current Documentation Structure

### Active Documentation (KEEP ALL)
```
README.md (4.6K)              # User-facing quick start
CLAUDE.md (35K)               # AI agent instructions (PRIMARY)
TODO.md (16K)                 # Task tracking
AGENT_AUTOMERGE.md (24K)      # AI: Automerge concepts
API_MAPPING.md (37K)          # API coverage matrix
MCP_PLAYWRIGHT_GUIDE.md (8K)  # Playwright MCP testing
DEMO.md (11K)                 # Complete demo guide (linked from README)
```

### Historical Documentation (ARCHIVE)
```
COMPLETE.md (12K)             # M0 milestone completion summary
```

### Reference Documentation (EVALUATE)
```
AUTOMERGE_JS_VS_RUST_COMPARISON.md (10K)  # Version comparison
```

---

## 🎯 Recommended Actions

### Immediate Cleanup

```bash
# 1. Remove build artifacts from Git tracking
git rm --cached go/server
git rm --cached doc.am
git rm --cached package-lock.json

# 2. Create directory structure
mkdir -p docs/archive
mkdir -p docs/reference
mkdir -p testdata/integration

# 3. Move historical docs
git mv COMPLETE.md docs/archive/M0_COMPLETE.md

# 4. Move reference docs
git mv AUTOMERGE_JS_VS_RUST_COMPARISON.md docs/reference/

# 5. Move integration test to testdata
git mv test_merge.sh testdata/integration/

# 6. Remove obsolete test files (not usable in current stack)
git rm test_text_crdt.html   # Obsolete: Automerge.js browser errors
git rm test_text_crdt.mjs     # Not usable: requires Node.js/npm

# 7. Commit cleanup
git commit -m "chore: Clean up repository - remove build artifacts, organize docs, remove obsolete tests"
```

### Future Considerations

**DEMO.md vs README.md**:
- **Option 1**: Keep both (DEMO has detailed instructions, README is brief)
- **Option 2**: Merge DEMO.md into README.md (consolidate)
- **Option 3**: Rename DEMO.md to USAGE.md (clearer purpose)

**Recommendation**: Keep both for now, they serve different purposes:
- README.md = Quick overview + getting started
- DEMO.md = Complete guide with architecture details

---

## 📊 File Size Analysis

### Total Documentation: ~148K
```
API_MAPPING.md               37K  (25%)  # Detailed API coverage
CLAUDE.md                    35K  (24%)  # AI agent instructions
AGENT_AUTOMERGE.md           24K  (16%)  # AI reference material
TODO.md                      16K  (11%)  # Task tracking
COMPLETE.md                  12K  (8%)   # Historical (archive?)
DEMO.md                      11K  (7%)   # Demo guide
AUTOMERGE_JS_VS_RUST_*.md    10K  (7%)   # Reference (evaluate?)
MCP_PLAYWRIGHT_GUIDE.md      8K   (5%)   # Testing guide
README.md                    5K   (3%)   # Quick start
```

**Observation**:
- Top 3 files (API_MAPPING, CLAUDE, AGENT_AUTOMERGE) = 65% of docs
- These are AI agent reference materials (essential)
- Historical/reference docs (COMPLETE, AUTOMERGE_JS_VS_RUST) = 15%

---

## ✅ Final Recommendations

### Must Do (Critical)
1. ✅ Remove `go/server` from Git (11MB build artifact)
2. ✅ Remove `doc.am` from Git (runtime state)
3. ✅ Remove `package-lock.json` from Git (npm artifact)

### Should Do (Organizational)
4. 🗂️ Archive `COMPLETE.md` → `docs/archive/M0_COMPLETE.md`
5. 🗂️ Move `AUTOMERGE_JS_VS_RUST_COMPARISON.md` → `docs/reference/`
6. 🗂️ Move `test_merge.sh` → `testdata/integration/` (valid integration test)
7. ❌ Remove `test_text_crdt.html` (obsolete - Automerge.js browser errors)
8. ❌ Remove `test_text_crdt.mjs` (not usable - requires Node.js/npm)

### Consider
9. 🤔 Evaluate if DEMO.md should be merged into README.md
10. 🤔 Create organized folder structure:
   ```
   docs/
   ├── archive/          # Historical docs
   │   └── M0_COMPLETE.md
   └── reference/        # Reference materials
       └── AUTOMERGE_JS_VS_RUST_COMPARISON.md

   testdata/
   ├── integration/      # Integration test scripts
   │   └── test_merge.sh  (MOVED from root)
   ├── snapshots/       # Automerge .am files (existing)
   └── screenshots/     # Test screenshots (existing)

   go/pkg/automerge/
   └── document_test.go  (existing - 11/12 tests passing)
   ```

---

## 🧪 Testing Strategy Clarification

**Current Testing Stack** (Go + Rust + Bash):

| Test Type | Tool | Location | Status |
|-----------|------|----------|--------|
| **Unit Tests** | Go `testing` | `go/pkg/automerge/document_test.go` | ✅ 11/12 passing |
| **Integration Tests** | Bash script | `test_merge.sh` → `testdata/integration/` | ✅ Working |
| **E2E Tests** | Playwright MCP | Manual via Claude Code | ✅ Working |
| **Rust Tests** | `cargo test` | `rust/automerge_wasi/src/*.rs` | ✅ Working |

**Obsolete Test Files** (Node.js/Automerge.js):

| File | Why Obsolete | Action |
|------|--------------|--------|
| `test_text_crdt.html` | Loads Automerge.js 3.1.2 via CDN - removed from ui.html due to WASM errors | ❌ Remove |
| `test_text_crdt.mjs` | Requires `npm install @automerge/automerge` - project doesn't use Node.js | ❌ Remove |

**Why Not Using Node.js**:
- ✅ Server-side CRDT: Rust (via WASI) + Go (wazero runtime)
- ✅ Client-side UI: Plain HTML + vanilla JS + SSE
- ✅ Testing: Go tests + Bash scripts + Playwright MCP
- ❌ NO npm dependencies for server (only `package.json` for Automerge.js reference in UI - currently unused)

**M2 Strategy** (when re-adding client-side Automerge.js):
- Use CDN imports (not npm packages)
- Test via Playwright MCP browser automation
- Keep server-side Rust WASI (no Node.js)

---

## 🔍 Files NOT to Touch

### Build Configuration (ESSENTIAL)
- `Makefile` - Build automation
- `go.mod`, `go.sum` - Go dependencies
- `Cargo.toml` - Rust configuration
- `package.json` - Node.js dependencies
- `.gitignore` - Git exclusions
- `.claude/` - Claude Code configuration

### Source Code (ESSENTIAL)
- `go/` - Go server code
- `rust/` - Rust WASI code
- `ui/` - Browser UI
- `testdata/` - Test data
- `screenshots/` - README images

### Documentation (KEEP)
- `README.md` - User-facing
- `CLAUDE.md` - AI instructions (PRIMARY)
- `TODO.md` - Task tracking
- `AGENT_AUTOMERGE.md` - AI reference
- `API_MAPPING.md` - API coverage
- `MCP_PLAYWRIGHT_GUIDE.md` - Testing guide
- `DEMO.md` - Demo guide (linked from README)
