// ═══════════════════════════════════════════════════════════════
// LAYER 7: Web Frontend (JavaScript Client)
//
// Responsibilities:
// - Handle user interactions for Text CRDT
// - Call HTTP API endpoints (Layer 6)
// - Update DOM based on SSE events
// - Manage client-side state (editor content, char count, status)
//
// Dependencies:
// ⬇️  Calls: /api/text (Layer 6 - HTTP API)
//           SSE: /api/stream (server events)
// ⬆️  Called by: web/js/app.js (orchestrator)
//
// Related Files:
// 🔁 Component: web/components/text.html (UI template)
// 🔁 Backend: go/pkg/api/handlers.go (Layer 6)
// 📝 Tests: tests/playwright/M0_TEXT_TEST_PLAN.md
// 🔗 Docs: docs/explanation/architecture.md#layer-7-web
//
// Design Note:
// This layer provides CRDT-specific UI logic. Infrastructure
// concerns (tab switching, SSE setup, routing) live in app.js.
// ═══════════════════════════════════════════════════════════════

export class TextComponent {
    constructor() {
        this.editor = null;
        this.charCount = null;
        this.status = null;
        this.eventSource = null;
        this.isLocalChange = false;
    }

    init() {
        this.editor = document.getElementById('text-editor');
        this.charCount = document.getElementById('text-char-count');
        this.status = document.getElementById('text-status');

        if (!this.editor) return; // Not on this tab

        // Event listeners
        document.getElementById('text-save')?.addEventListener('click', () => this.save());
        document.getElementById('text-clear')?.addEventListener('click', () => this.clear());
        document.getElementById('text-load')?.addEventListener('click', () => this.load());

        this.editor.addEventListener('input', () => this.updateCharCount());

        // Keyboard shortcut: Cmd/Ctrl+S
        this.editor.addEventListener('keydown', (e) => {
            if ((e.metaKey || e.ctrlKey) && e.key === 's') {
                e.preventDefault();
                this.save();
            }
        });

        // Load initial text and start SSE
        this.load();
        this.connectSSE();
    }

    updateCharCount() {
        if (this.charCount) {
            this.charCount.textContent = this.editor.value.length;
        }
    }

    async load() {
        try {
            const response = await fetch('/api/text');
            if (response.ok) {
                const text = await response.text();
                this.editor.value = text;
                this.updateCharCount();
                this.showStatus('Loaded', 'success');
            }
        } catch (error) {
            console.error('Error loading text:', error);
            this.showStatus('Load failed', 'error');
        }
    }

    async save() {
        try {
            this.isLocalChange = true;
            this.showStatus('Saving...', 'info');

            const response = await fetch('/api/text', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ text: this.editor.value }),
            });

            if (!response.ok) {
                throw new Error('Failed to save');
            }

            this.showStatus('Saved ✓', 'success');
            setTimeout(() => { this.isLocalChange = false; }, 100);
        } catch (error) {
            console.error('Error saving text:', error);
            this.showStatus('Save failed', 'error');
        }
    }

    async clear() {
        if (confirm('Clear all text?')) {
            this.editor.value = '';
            this.updateCharCount();
            await this.save();
        }
    }

    connectSSE() {
        this.eventSource = new EventSource('/api/stream');

        this.eventSource.addEventListener('snapshot', (event) => {
            console.log('SSE snapshot:', event.data);
            const data = JSON.parse(event.data);
            if (!this.isLocalChange) {
                this.editor.value = data.text;
                this.updateCharCount();
            }
        });

        this.eventSource.addEventListener('update', (event) => {
            console.log('SSE update:', event.data);
            if (!this.isLocalChange) {
                const data = JSON.parse(event.data);
                this.editor.value = data.text;
                this.updateCharCount();
            }
        });

        this.eventSource.onopen = () => {
            console.log('SSE connection opened');
            this.showStatus('Connected', 'success');
        };

        this.eventSource.onerror = (error) => {
            console.error('SSE error:', error);
            this.showStatus('Disconnected', 'error');
            this.eventSource.close();

            // Reconnect after 3 seconds
            setTimeout(() => this.connectSSE(), 3000);
        };
    }

    showStatus(message, type) {
        if (!this.status) return;
        this.status.textContent = message;
        this.status.className = `status-text status-${type}`;

        if (type === 'success') {
            setTimeout(() => {
                this.status.textContent = '';
            }, 2000);
        }
    }

    destroy() {
        if (this.eventSource) {
            this.eventSource.close();
        }
    }
}
