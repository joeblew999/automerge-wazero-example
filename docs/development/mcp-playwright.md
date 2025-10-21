# Playwright MCP - Complete Guide

**Purpose**: End-to-end browser testing via Model Context Protocol (MCP)

**Status**: ⚠️ Requires manual permission approval for each tool

---

## Quick Start

### 1. MCP Server Configuration

**Location**: `~/.claude.json`

```json
{
  "mcpServers": {
    "playwright": {
      "type": "stdio",
      "command": "npx",
      "args": ["@playwright/mcp@latest"],
      "env": {}
    }
  }
}
```

**Verify**:
```bash
/Users/apple/.local/bin/claude mcp list
# Should show: playwright: npx @playwright/mcp@latest - ✓ Connected
```

### 2. Available Tools (21 total)

| Category | Tools |
|----------|-------|
| **Navigation** | navigate, navigate_back |
| **Interaction** | click, type, drag, hover, select_option, fill_form |
| **Inspection** | snapshot, take_screenshot, evaluate |
| **Dialogs** | handle_dialog, file_upload |
| **State** | console_messages, network_requests |
| **Control** | tabs, wait_for, press_key |
| **Window** | close, resize |
| **Setup** | install |

### 3. Usage Example

```javascript
// Navigate to page
mcp__playwright__browser_navigate(url: "http://localhost:8080")

// Take snapshot
mcp__playwright__browser_snapshot()

// Interact
mcp__playwright__browser_click(element: "button", ref: "e10")
mcp__playwright__browser_type(text: "Hello", ref: "e8")

// Capture screenshot
mcp__playwright__browser_take_screenshot(filename: "test.png")
```

---

## ⚠️ Permission Gotcha

### The Reality

**`allowedTools` configuration DOES NOT WORK for MCP tools.**

We tried:
- ❌ `.claude/settings.json` → `allowedTools`
- ❌ `~/.claude.json` → `projects → allowedTools`
- ❌ Root-level `allowedTools` (deprecated in 2.0.8)

**Why**: `allowedTools` only applies to **built-in Claude Code tools** (Read, Write, Bash), NOT MCP tools.

### The Solution

**When you get a permission prompt**:
1. Click **"Always Allow"** (not just "Allow")
2. Permission persists for that tool in this project
3. Repeat for each tool you use

**First session**: ~5-10 prompts (once per tool)
**Subsequent sessions**: No prompts for approved tools

---

## Testing Workflow

### Interactive Testing (Development)

```javascript
// 1. Start server
make run

// 2. Use MCP tools interactively
mcp__playwright__browser_navigate(url: "http://localhost:8080")
mcp__playwright__browser_snapshot()
mcp__playwright__browser_click(...)

// 3. Screenshots auto-saved to:
.playwright-mcp/testdata/screenshots/
```

### Automated Testing (CI/CD)

**Use Bash + Native Playwright** to avoid prompts:

```bash
# Create test scripts
mkdir -p testdata/playwright
cat > testdata/playwright/test-ui.spec.ts <<'EOF'
import { test, expect } from '@playwright/test';

test('UI loads and works', async ({ page }) => {
  await page.goto('http://localhost:8080');
  await expect(page.locator('h1')).toContainText('Automerge');
  await page.click('[role="textbox"]');
  await page.fill('[role="textbox"]', 'Test');
  await expect(page.locator('strong')).toContainText('4');
});
EOF

# Run via Bash (no prompts!)
npx playwright test testdata/playwright/
```

---

## Screenshot Management

### Directory Structure

```
.playwright-mcp/testdata/screenshots/  ← MCP auto-output
testdata/screenshots/                  ← Test artifacts (committed)
screenshots/                           ← README.md images (committed)
```

### Workflow

```bash
# 1. MCP tools save here automatically
.playwright-mcp/testdata/screenshots/test.png

# 2. Copy for test artifacts
cp .playwright-mcp/testdata/screenshots/*.png testdata/screenshots/

# 3. Copy best shot for README
cp .playwright-mcp/testdata/screenshots/best.png screenshots/screenshot.png
```

---

## Integration with go/testdata

### Align Playwright Tests with Go Tests

```
go/testdata/
├── snapshots/           # Automerge binary snapshots (.am files)
├── playwright/          # E2E browser tests (.spec.ts)
│   ├── test-ui.spec.ts
│   ├── test-crdt.spec.ts
│   └── screenshots/     # Test screenshots
└── expected/            # Expected test outputs
```

### Example: CRDT Test

**Go test** (`go/pkg/automerge/text_test.go`):
```go
func TestDocument_TextSplice(t *testing.T) {
    doc := NewDocument()
    doc.TextSplice(0, 0, "Hello")
    // Assert CRDT state
}
```

**Playwright test** (`testdata/playwright/test-crdt.spec.ts`):
```typescript
test('CRDT text sync', async ({ page }) => {
  await page.goto('http://localhost:8080');
  await page.fill('textarea', 'Hello');
  await page.click('button:has-text("Save")');

  // Verify via API
  const response = await page.request.get('/api/text');
  expect(await response.text()).toBe('Hello');
});
```

---

## Troubleshooting

### MCP Server Not Connected

```bash
# Check status
claude mcp list

# If missing, add server
claude mcp add -s user -t stdio -c npx -a "@playwright/mcp@latest" playwright

# Restart Claude Code
killall -9 claude
# Reopen VSCode
```

### Tools Still Prompting

**Expected behavior**: MCP tools will always prompt on first use.

**Solution**: Click "Always Allow" for each tool you want to use.

**Alternative**: Use Bash + native Playwright for automated testing.

### Screenshots Not Saving

**Check directory exists**:
```bash
mkdir -p .playwright-mcp/testdata/screenshots
```

**Use absolute paths**:
```javascript
// Instead of:
mcp__playwright__browser_take_screenshot(filename: "test.png")

// Use:
mcp__playwright__browser_take_screenshot(
  filename: "testdata/screenshots/test.png"
)
```

---

## Best Practices

### 1. Hybrid Approach

**Interactive** (MCP tools):
- Quick exploration
- Manual testing
- Debugging UI issues

**Automated** (Bash + Playwright):
- CI/CD pipelines
- Regression tests
- Reproducible test suites

### 2. Screenshot Naming

```javascript
// Bad: generic names
"screenshot.png"

// Good: descriptive names
"01-initial-load.png"
"02-text-input.png"
"03-after-save.png"
```

### 3. Test Organization

```
testdata/playwright/
├── specs/
│   ├── ui.spec.ts          # UI component tests
│   ├── crdt.spec.ts        # CRDT behavior tests
│   └── sse.spec.ts         # Real-time sync tests
├── fixtures/
│   └── test-data.json      # Test data
└── screenshots/
    └── expected/           # Expected screenshots
```

---

## Command Reference

### MCP Tools Quick Reference

```javascript
// Navigation
mcp__playwright__browser_navigate(url: string)
mcp__playwright__browser_navigate_back()

// Interaction
mcp__playwright__browser_click(element: string, ref: string)
mcp__playwright__browser_type(element: string, ref: string, text: string)
mcp__playwright__browser_fill_form(fields: [{name, type, ref, value}])

// Inspection
mcp__playwright__browser_snapshot()
mcp__playwright__browser_take_screenshot(filename?: string)
mcp__playwright__browser_console_messages(onlyErrors?: boolean)
mcp__playwright__browser_network_requests()

// Dialogs
mcp__playwright__browser_handle_dialog(accept: boolean, promptText?: string)

// Window
mcp__playwright__browser_resize(width: number, height: number)
mcp__playwright__browser_close()
```

### Bash + Playwright Commands

```bash
# Install Playwright
npm install -D @playwright/test

# Run all tests
npx playwright test

# Run specific test
npx playwright test testdata/playwright/test-ui.spec.ts

# Run with UI
npx playwright test --ui

# Generate screenshots
npx playwright test --screenshot=on
```

---

## Summary

**MCP Playwright** provides interactive browser automation but requires manual permission approval.

**For autonomous testing**, use **Bash + native Playwright** to avoid prompts.

**Recommended workflow**:
- Development: MCP tools (click "Always Allow")
- CI/CD: Bash + Playwright scripts

**See Also**:
- [CLAUDE.md](../../CLAUDE.md#04-testing-requirements) - Testing requirements and MCP configuration
