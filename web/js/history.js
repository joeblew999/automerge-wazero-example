// M0: History Module
// Maps to: go/pkg/automerge/history.go, go/pkg/api/history.go

export class HistoryComponent {
    constructor() {
        this.headsDisplay = null;
        this.changesDisplay = null;
        this.status = null;
        this.eventSource = null;
    }

    init() {
        this.headsDisplay = document.getElementById('history-heads');
        this.changesDisplay = document.getElementById('history-changes');
        this.status = document.getElementById('history-status');

        if (!this.headsDisplay) return; // Not on this tab

        // Event listeners
        document.getElementById('history-refresh')?.addEventListener('click', () => this.refresh());
        document.getElementById('history-get-heads')?.addEventListener('click', () => this.getHeads());
        document.getElementById('history-get-changes')?.addEventListener('click', () => this.getChanges());
        document.getElementById('history-download-snapshot')?.addEventListener('click', () => this.downloadSnapshot());

        // Load initial data
        this.refresh();
        this.connectSSE();
    }

    async refresh() {
        await Promise.all([
            this.getHeads(),
            this.getChanges()
        ]);
    }

    async getHeads() {
        try {
            this.showStatus('Loading heads...', 'info');

            const response = await fetch('/api/heads');

            if (!response.ok) {
                throw new Error('Failed to get heads');
            }

            const data = await response.json();
            this.displayHeads(data.heads || []);
            this.showStatus('Heads loaded ✓', 'success');
        } catch (error) {
            console.error('Error getting heads:', error);
            this.showStatus('Failed to load heads', 'error');
            this.headsDisplay.innerHTML = '<div class="error-message">Failed to load heads</div>';
        }
    }

    displayHeads(heads) {
        if (heads.length === 0) {
            this.headsDisplay.innerHTML = '<div class="empty-message">No heads (empty document)</div>';
            return;
        }

        const headElements = heads.map((head, i) => {
            return `
                <div class="head-item">
                    <span class="head-index">#${i + 1}</span>
                    <code class="head-hash">${this.escapeHtml(head)}</code>
                </div>
            `;
        }).join('');

        this.headsDisplay.innerHTML = `
            <div class="heads-count">Current heads: <strong>${heads.length}</strong></div>
            ${headElements}
        `;
    }

    async getChanges() {
        try {
            this.showStatus('Loading changes...', 'info');

            const response = await fetch('/api/changes');

            if (!response.ok) {
                throw new Error('Failed to get changes');
            }

            const data = await response.json();
            this.displayChanges(data);
            this.showStatus('Changes loaded ✓', 'success');
        } catch (error) {
            console.error('Error getting changes:', error);
            this.showStatus('Failed to load changes', 'error');
            this.changesDisplay.innerHTML = '<div class="error-message">Failed to load changes</div>';
        }
    }

    displayChanges(data) {
        if (!data.changes || data.size === 0) {
            this.changesDisplay.innerHTML = '<div class="empty-message">No changes yet</div>';
            return;
        }

        const sizeKB = (data.size / 1024).toFixed(2);
        const preview = data.changes.substring(0, 100);

        this.changesDisplay.innerHTML = `
            <div class="changes-info">
                <div class="info-row">
                    <strong>Size:</strong> ${data.size} bytes (${sizeKB} KB)
                </div>
                <div class="info-row">
                    <strong>Format:</strong> Base64-encoded binary changes
                </div>
                <div class="info-row">
                    <strong>Preview:</strong>
                    <div class="changes-preview">
                        <code>${this.escapeHtml(preview)}...</code>
                    </div>
                </div>
            </div>
        `;
    }

    async downloadSnapshot() {
        try {
            this.showStatus('Downloading snapshot...', 'info');

            const response = await fetch('/api/doc');

            if (!response.ok) {
                throw new Error('Failed to download snapshot');
            }

            const blob = await response.blob();
            const url = window.URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = url;
            a.download = `automerge-doc-${Date.now()}.am`;
            document.body.appendChild(a);
            a.click();
            document.body.removeChild(a);
            window.URL.revokeObjectURL(url);

            this.showStatus('Snapshot downloaded ✓', 'success');
        } catch (error) {
            console.error('Error downloading snapshot:', error);
            this.showStatus('Download failed', 'error');
        }
    }

    connectSSE() {
        this.eventSource = new EventSource('/api/stream');

        this.eventSource.addEventListener('snapshot', (event) => {
            console.log('SSE snapshot:', event.data);
            this.refresh();
        });

        this.eventSource.addEventListener('update', (event) => {
            console.log('SSE update:', event.data);
            this.refresh();
        });

        this.eventSource.onopen = () => {
            console.log('SSE connection opened');
        };

        this.eventSource.onerror = (error) => {
            console.error('SSE error:', error);
            this.eventSource.close();
            // Reconnect after 3 seconds
            setTimeout(() => this.connectSSE(), 3000);
        };
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
            }, 2000);
        }
    }

    destroy() {
        if (this.eventSource) {
            this.eventSource.close();
        }
    }
}
