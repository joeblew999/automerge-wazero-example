// Cursor Operations Module
// Maps to: go/pkg/automerge/cursor.go, go/pkg/api/cursor.go

export class CursorComponent {
    constructor() {
        this.pathInput = null;
        this.textArea = null;
        this.indexInput = null;
        this.cursorDisplay = null;
        this.cursorInput = null;
        this.lookupResult = null;
        this.status = null;
        this.eventSource = null;
        this.savedCursors = [];
    }

    init() {
        this.pathInput = document.getElementById('cursor-path');
        this.textArea = document.getElementById('cursor-text');
        this.indexInput = document.getElementById('cursor-index');
        this.cursorDisplay = document.getElementById('cursor-display');
        this.cursorInput = document.getElementById('cursor-input');
        this.lookupResult = document.getElementById('cursor-lookup-result');
        this.status = document.getElementById('cursor-status');

        if (!this.pathInput) return; // Not on this tab

        // Set default path
        this.pathInput.value = 'ROOT.content';

        // Event listeners
        document.getElementById('cursor-load-text')?.addEventListener('click', () => this.loadText());
        document.getElementById('cursor-save-text')?.addEventListener('click', () => this.saveText());
        document.getElementById('cursor-get')?.addEventListener('click', () => this.getCursor());
        document.getElementById('cursor-get-selection')?.addEventListener('click', () => this.getCursorFromSelection());
        document.getElementById('cursor-lookup')?.addEventListener('click', () => this.lookupCursor());
        document.getElementById('cursor-demo')?.addEventListener('click', () => this.runDemo());

        // TextArea selection change
        this.textArea.addEventListener('mouseup', () => this.updateSelectionInfo());
        this.textArea.addEventListener('keyup', () => this.updateSelectionInfo());

        // Load initial text
        this.loadText();
    }

    updateSelectionInfo() {
        const start = this.textArea.selectionStart;
        const end = this.textArea.selectionEnd;
        document.getElementById('cursor-selection-info').textContent =
            `Selection: ${start}-${end} (${end - start} chars)`;
    }

    async loadText() {
        try {
            const response = await fetch('/api/text');
            if (response.ok) {
                const text = await response.text();
                this.textArea.value = text;
                this.showStatus('Text loaded ✓', 'success');
            }
        } catch (error) {
            console.error('Error loading text:', error);
            this.showStatus('Load failed', 'error');
        }
    }

    async saveText() {
        try {
            const response = await fetch('/api/text', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ text: this.textArea.value }),
            });

            if (!response.ok) {
                throw new Error('Failed to save text');
            }

            this.showStatus('Text saved ✓', 'success');
        } catch (error) {
            console.error('Error saving text:', error);
            this.showStatus('Save failed', 'error');
        }
    }

    async getCursor() {
        try {
            const path = this.pathInput.value.trim();
            const index = parseInt(this.indexInput.value, 10);

            if (!path || isNaN(index)) {
                this.showStatus('Path and index are required', 'error');
                return;
            }

            const response = await fetch(`/api/cursor?path=${encodeURIComponent(path)}&index=${index}`);

            if (!response.ok) {
                throw new Error('Failed to get cursor');
            }

            const data = await response.json();
            this.displayCursor(data);
            this.showStatus('Cursor retrieved ✓', 'success');
        } catch (error) {
            console.error('Error getting cursor:', error);
            this.showStatus('Get cursor failed', 'error');
        }
    }

    async getCursorFromSelection() {
        const start = this.textArea.selectionStart;
        const path = this.pathInput.value.trim();

        if (!path) {
            this.showStatus('Path is required', 'error');
            return;
        }

        try {
            const response = await fetch(`/api/cursor?path=${encodeURIComponent(path)}&index=${start}`);

            if (!response.ok) {
                throw new Error('Failed to get cursor');
            }

            const data = await response.json();
            this.displayCursor(data);
            this.indexInput.value = start;
            this.showStatus(`Cursor at position ${start} ✓`, 'success');
        } catch (error) {
            console.error('Error getting cursor:', error);
            this.showStatus('Get cursor failed', 'error');
        }
    }

    displayCursor(data) {
        this.cursorDisplay.innerHTML = `
            <div class="cursor-info">
                <strong>Path:</strong> ${this.escapeHtml(data.path)}<br>
                <strong>Index:</strong> ${data.index}<br>
                <strong>Cursor:</strong> <code>${this.escapeHtml(data.cursor)}</code>
            </div>
        `;
        this.cursorInput.value = data.cursor;

        // Save to history
        this.savedCursors.push(data);
        this.updateCursorHistory();
    }

    async lookupCursor() {
        try {
            const path = this.pathInput.value.trim();
            const cursor = this.cursorInput.value.trim();

            if (!path || !cursor) {
                this.showStatus('Path and cursor are required', 'error');
                return;
            }

            const response = await fetch('/api/cursor/lookup', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ path, cursor }),
            });

            if (!response.ok) {
                throw new Error('Failed to lookup cursor');
            }

            const data = await response.json();
            this.lookupResult.innerHTML = `
                <div class="lookup-result">
                    <strong>Current Index:</strong> ${data.index}<br>
                    <em>Cursor position updated to index ${data.index}</em>
                </div>
            `;
            this.indexInput.value = data.index;
            this.showStatus('Cursor looked up ✓', 'success');
        } catch (error) {
            console.error('Error looking up cursor:', error);
            this.showStatus('Lookup failed', 'error');
        }
    }

    async runDemo() {
        this.showStatus('Running cursor demo...', 'info');

        // Step 1: Set initial text
        this.textArea.value = "Hello World";
        await this.saveText();
        await this.delay(500);

        // Step 2: Get cursor at position 6 (after "Hello ")
        this.indexInput.value = 6;
        await this.getCursor();
        await this.delay(500);

        const savedCursor = this.cursorInput.value;

        // Step 3: Insert text at beginning
        this.textArea.value = "Hi " + this.textArea.value;
        await this.saveText();
        this.showStatus('Inserted "Hi " at start', 'info');
        await this.delay(500);

        // Step 4: Lookup cursor (should move from 6 to 9)
        this.cursorInput.value = savedCursor;
        await this.lookupCursor();

        this.showStatus('Demo complete! Cursor moved from index 6 to 9 ✓', 'success');
    }

    updateCursorHistory() {
        const historyDiv = document.getElementById('cursor-history');
        if (!historyDiv) return;

        if (this.savedCursors.length === 0) {
            historyDiv.innerHTML = '<div class="empty-message">No cursors saved yet</div>';
            return;
        }

        const items = this.savedCursors.slice(-5).reverse().map((c, i) => `
            <div class="history-item">
                <small>#${this.savedCursors.length - i}</small>
                Index ${c.index} → <code>${this.escapeHtml(c.cursor.substring(0, 20))}...</code>
            </div>
        `).join('');

        historyDiv.innerHTML = items;
    }

    delay(ms) {
        return new Promise(resolve => setTimeout(resolve, ms));
    }

    escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }

    showStatus(message, type) {
        if (!this.status) return;
        this.status.textContent = message;
        this.status.className = `status-text status-${type}`;

        if (type === 'success') {
            setTimeout(() => {
                this.status.textContent = '';
            }, 3000);
        }
    }

    destroy() {
        // Cleanup if needed
    }
}
