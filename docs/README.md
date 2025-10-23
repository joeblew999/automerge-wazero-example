# Documentation Index

This directory contains all project documentation organized using the [Diátaxis framework](https://diataxis.fr/).

## 📚 Quick Links

- **[Getting Started](tutorials/getting-started.md)** - Start here if you're new
- **[API Mapping](reference/api-mapping.md)** - Complete API coverage matrix
- **[Architecture](explanation/architecture.md)** - Understand the 4-layer design
- **[STATUS](../STATUS.md)** - Project status and future milestones

---

## 📖 Documentation Structure

### 🎓 Tutorials (Learning-Oriented)

Step-by-step lessons for learning. **Best for**: First-time users, learning fundamentals.

- [Getting Started](tutorials/getting-started.md) - Your first CRDT document

### 🛠️ How-To Guides (Goal-Oriented)

Recipes for solving specific problems. **Best for**: Accomplishing specific tasks.

- [Build Automerge.js](how-to/build-automerge-js.md) - Building from source
- [Add New WASI Export](how-to/) - Coming soon
- [Debug WASM Issues](how-to/) - Coming soon
- [Test with Playwright](development/mcp-playwright.md) - Playwright MCP testing guide

### 📋 Reference (Information-Oriented)

Technical lookup and specifications. **Best for**: Looking up details, checking coverage.

- [API Mapping](reference/api-mapping.md) - Automerge API → WASI → Go mapping matrix
- [HTTP API Complete](reference/http-api-complete.md) - Full HTTP endpoint reference
- [Automerge Comparison](reference/automerge-comparison.md) - JavaScript vs Rust API differences

### 💡 Explanation (Understanding-Oriented)

Conceptual understanding and design decisions. **Best for**: Understanding why things work the way they do.

- [Architecture](explanation/architecture.md) - 7-layer architecture deep dive
- [Web Architecture](explanation/web-architecture.md) - Web folder 1:1 mapping
- [CRDT Basics](ai-agents/automerge-guide.md) - How Automerge CRDTs work
- [AI Readability](explanation/ai-readability-improvements.md) - Code patterns for AI agents

### 🔧 Development

Developer workflow, testing, and contributing.

- [Testing Guide](development/testing.md) - Unit, integration, and E2E testing
- [MCP Playwright](development/mcp-playwright.md) - Browser testing with Playwright MCP
- [Project Status](../STATUS.md) - Implementation status and future milestones

### 🤖 AI Agent Guides

Specialized documentation for AI agents (Claude Code, etc.)

- [Automerge Guide](ai-agents/automerge-guide.md) - CRDT concepts, patterns, best practices
- [Datastar Guide](ai-agents/datastar-guide.md) - Coming soon (M4)

### 📦 Archive

Historical documentation and completed milestone records.

- [M0 Complete](archive/M0_COMPLETE.md) - Milestone 0 completion summary
- [Cleanup Analysis](archive/cleanup-analysis.md) - Historical refactoring analysis
- [Implementation Status 2025-10-20](archive/IMPLEMENTATION_STATUS.md) - Point-in-time status snapshot
- [Quality Report](archive/QUALITY_REPORT.md) - Code quality analysis (2025-10-20)
- [Session Summaries](archive/sessions/) - Detailed session-by-session logs

---

## 🎯 Finding What You Need

### "I want to..."

- **...get started** → [Getting Started](tutorials/getting-started.md)
- **...understand the architecture** → [Architecture](explanation/architecture.md)
- **...add a new feature** → Check [API Mapping](reference/api-mapping.md), then see [CLAUDE.md](../CLAUDE.md) section 0.2
- **...check implementation status** → [STATUS](../STATUS.md)
- **...run tests** → [Testing Guide](development/testing.md)
- **...understand CRDTs** → [Automerge Guide](ai-agents/automerge-guide.md)

### By Technology

- **Automerge (Rust CRDT)**: [Automerge Guide](ai-agents/automerge-guide.md), [API Comparison](reference/automerge-comparison.md)
- **WASI/WASM**: [Architecture](explanation/architecture.md)
- **Go (wazero)**: [API Mapping](reference/api-mapping.md)
- **Testing**: [Testing Guide](development/testing.md), [Playwright](development/mcp-playwright.md)

---

## 📝 Documentation Methodology

We use the **Diátaxis framework** to keep documentation organized:

```
                LEARNING-ORIENTED | GOAL-ORIENTED
                ------------------+---------------
    PRACTICAL   TUTORIALS         | HOW-TO GUIDES
                ------------------+---------------
    THEORETICAL EXPLANATION       | REFERENCE
```

### Rules

1. **Single Source of Truth** - Each piece of info lives in ONE place only
2. **Name by Purpose** - Files named for what they're FOR, not what they contain
3. **Root Level = Critical** - Only `README.md` and `CLAUDE.md` in repo root
4. **Use Links** - Reference, don't duplicate

### For AI Agents

If you're an AI agent (Claude Code, etc.):

1. Start with [CLAUDE.md](../CLAUDE.md) - Master instructions
2. Read [Automerge Guide](ai-agents/automerge-guide.md) - Understand CRDTs
3. Check [API Mapping](reference/api-mapping.md) - Know what's implemented
4. Follow [Architecture](explanation/architecture.md) - Understand the layers

---

**Last Updated**: 2025-10-21
