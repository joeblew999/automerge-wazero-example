// ==============================================================================
// Layer 7: Web Frontend - Rich Text (JavaScript Module)
// ==============================================================================
// ARCHITECTURE: This is the web frontend layer (Layer 7/7).
//
// RESPONSIBILITIES:
// - User interaction (DOM events, mark buttons, text selection)
// - HTTP API calls to Layer 6 (fetch /api/richtext/mark)
// - UI updates (rendering formatted text with marks)
// - State management (current text, applied marks)
//
// DEPENDENCIES:
// - Layer 6: pkg/api/crdt_richtext.go (HTTP handlers)
// - Browser APIs: fetch(), DOM, Selection API
//
// RELATED FILES (1:1 mapping):
// - Layer 2: rust/automerge_wasi/src/richtext.rs (WASI exports)
// - Layer 3: pkg/wazero/crdt_richtext.go (FFI wrappers)
// - Layer 4: pkg/automerge/crdt_richtext.go (pure CRDT API)
// - Layer 5: pkg/server/crdt_richtext.go (stateful server operations)
// - Layer 6: pkg/api/crdt_richtext.go (HTTP handlers)
// - Layer 7: web/components/crdt_richtext.html (HTML template)
//
// NOTES:
// - This module exports a class that's instantiated by app.js
// - Handles text selection and applies marks to selected ranges
// - Marks are CRDT-aware (concurrent formatting merges correctly)
// ==============================================================================

// M2: RichText Marks Module
// Maps to: go/pkg/automerge/richtext.go, go/pkg/api/richtext.go

export class RichTextComponent {
    constructor() {
        this.editor = null;
        this.path = 'ROOT.content';
    }

    init() {
        this.editor = document.getElementById('richtext-editor');
        if (!this.editor) return; // Not on this tab

        // Event listeners
        document.getElementById('mark-apply')?.addEventListener('click', () => this.applyMark());
        document.getElementById('unmark-apply')?.addEventListener('click', () => this.removeMark());
        document.getElementById('marks-get')?.addEventListener('click', () => this.getMarks());
        document.getElementById('richtext-use-selection')?.addEventListener('click', () => this.useSelection());
        document.getElementById('richtext-clear-all')?.addEventListener('click', () => this.clearAll());

        // Update selection display
        this.editor.addEventListener('select', () => this.updateSelection());
        this.editor.addEventListener('click', () => this.updateSelection());
        this.editor.addEventListener('keyup', () => this.updateSelection());

        this.updateSelection();
    }

    updateSelection() {
        const start = this.editor.selectionStart;
        const end = this.editor.selectionEnd;
        const cursor = start;

        const selectionEl = document.getElementById('richtext-selection');
        const cursorEl = document.getElementById('richtext-cursor');

        if (start === end) {
            selectionEl.textContent = 'None';
        } else {
            selectionEl.textContent = `${start}-${end} (${end - start} chars)`;
        }

        cursorEl.textContent = cursor;
    }

    useSelection() {
        const start = this.editor.selectionStart;
        const end = this.editor.selectionEnd;

        document.getElementById('mark-start').value = start;
        document.getElementById('mark-end').value = end;
        document.getElementById('unmark-start').value = start;
        document.getElementById('unmark-end').value = end;
        document.getElementById('marks-pos').value = start;
    }

    async applyMark() {
        const name = document.getElementById('mark-name').value;
        const start = parseInt(document.getElementById('mark-start').value);
        const end = parseInt(document.getElementById('mark-end').value);
        const expand = document.getElementById('mark-expand').value;

        if (isNaN(start) || isNaN(end) || start < 0 || end <= start) {
            alert('Invalid range. End must be greater than start.');
            return;
        }

        console.log(`Applying mark: ${name} [${start}, ${end}) expand=${expand}`);

        try {
            const response = await fetch('/api/richtext/mark', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    path: this.path,
                    name: name,
                    value: 'true',
                    start: start,
                    end: end,
                    expand: expand,
                }),
            });

            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${await response.text()}`);
            }

            console.log(`✓ Mark applied: ${name}`);
            alert(`✓ Applied ${name} mark to range [${start}, ${end})`);
        } catch (error) {
            console.error('Error applying mark:', error);
            alert(`Failed to apply mark: ${error.message}`);
        }
    }

    async removeMark() {
        const name = document.getElementById('unmark-name').value;
        const start = parseInt(document.getElementById('unmark-start').value);
        const end = parseInt(document.getElementById('unmark-end').value);

        if (isNaN(start) || isNaN(end) || start < 0 || end <= start) {
            alert('Invalid range. End must be greater than start.');
            return;
        }

        console.log(`Removing mark: ${name} [${start}, ${end})`);

        try {
            const response = await fetch('/api/richtext/unmark', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    path: this.path,
                    name: name,
                    start: start,
                    end: end,
                    expand: 'none',
                }),
            });

            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${await response.text()}`);
            }

            console.log(`✓ Mark removed: ${name}`);
            alert(`✓ Removed ${name} mark from range [${start}, ${end})`);
        } catch (error) {
            console.error('Error removing mark:', error);
            alert(`Failed to remove mark: ${error.message}`);
        }
    }

    async getMarks() {
        const pos = parseInt(document.getElementById('marks-pos').value);

        if (isNaN(pos) || pos < 0) {
            alert('Invalid position');
            return;
        }

        console.log(`Getting marks at position: ${pos}`);

        try {
            const response = await fetch(`/api/richtext/marks?path=${encodeURIComponent(this.path)}&pos=${pos}`);

            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${await response.text()}`);
            }

            const data = await response.json();
            console.log('Marks:', data);

            const displayEl = document.getElementById('marks-display');
            if (data.marks && data.marks.length > 0) {
                displayEl.textContent = JSON.stringify(data.marks, null, 2);
            } else {
                displayEl.textContent = 'No marks at this position';
            }
        } catch (error) {
            console.error('Error getting marks:', error);
            alert(`Failed to get marks: ${error.message}`);
        }
    }

    async clearAll() {
        if (!confirm('Clear all marks? This will require manual unmark calls.')) {
            return;
        }

        alert('Note: There is no "clear all marks" API. You need to unmark each range individually.');
    }
}
