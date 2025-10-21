// M0: Map CRDT Module
// Maps to: go/pkg/automerge/map.go, go/pkg/api/map.go

export class MapComponent {
    constructor() {
        this.pathInput = null;
        this.keyInput = null;
        this.valueInput = null;
        this.mapKeys = null;
        this.status = null;
        this.eventSource = null;
        this.currentPath = 'ROOT';
    }

    init() {
        this.pathInput = document.getElementById('map-path');
        this.keyInput = document.getElementById('map-key');
        this.valueInput = document.getElementById('map-value');
        this.mapKeys = document.getElementById('map-keys');
        this.status = document.getElementById('map-status');

        if (!this.pathInput) return; // Not on this tab

        // Set default path
        this.pathInput.value = this.currentPath;

        // Event listeners
        document.getElementById('map-put')?.addEventListener('click', () => this.put());
        document.getElementById('map-get')?.addEventListener('click', () => this.get());
        document.getElementById('map-delete')?.addEventListener('click', () => this.deleteKey());
        document.getElementById('map-list-keys')?.addEventListener('click', () => this.listKeys());
        document.getElementById('map-clear-all')?.addEventListener('click', () => this.clearAll());

        // Path change listener
        this.pathInput.addEventListener('change', () => {
            this.currentPath = this.pathInput.value;
            this.listKeys();
        });

        // Load initial keys
        this.listKeys();
        this.connectSSE();
    }

    async put() {
        try {
            const path = this.pathInput.value.trim();
            const key = this.keyInput.value.trim();
            const value = this.valueInput.value;

            if (!path || !key) {
                this.showStatus('Path and key are required', 'error');
                return;
            }

            this.showStatus('Saving...', 'info');

            const response = await fetch('/api/map', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ path, key, value }),
            });

            if (!response.ok) {
                throw new Error('Failed to put value');
            }

            this.showStatus('Saved ✓', 'success');
            this.keyInput.value = '';
            this.valueInput.value = '';

            // Refresh keys list
            setTimeout(() => this.listKeys(), 100);
        } catch (error) {
            console.error('Error putting value:', error);
            this.showStatus('Put failed', 'error');
        }
    }

    async get() {
        try {
            const path = this.pathInput.value.trim();
            const key = this.keyInput.value.trim();

            if (!path || !key) {
                this.showStatus('Path and key are required', 'error');
                return;
            }

            const response = await fetch(`/api/map?path=${encodeURIComponent(path)}&key=${encodeURIComponent(key)}`);

            if (!response.ok) {
                throw new Error('Failed to get value');
            }

            const data = await response.json();
            this.valueInput.value = data.value || '';
            this.showStatus('Loaded ✓', 'success');
        } catch (error) {
            console.error('Error getting value:', error);
            this.showStatus('Get failed', 'error');
        }
    }

    async deleteKey() {
        try {
            const path = this.pathInput.value.trim();
            const key = this.keyInput.value.trim();

            if (!path || !key) {
                this.showStatus('Path and key are required', 'error');
                return;
            }

            if (!confirm(`Delete key "${key}" from ${path}?`)) {
                return;
            }

            const response = await fetch(`/api/map?path=${encodeURIComponent(path)}&key=${encodeURIComponent(key)}`, {
                method: 'DELETE',
            });

            if (!response.ok) {
                throw new Error('Failed to delete key');
            }

            this.showStatus('Deleted ✓', 'success');
            this.keyInput.value = '';
            this.valueInput.value = '';

            // Refresh keys list
            setTimeout(() => this.listKeys(), 100);
        } catch (error) {
            console.error('Error deleting key:', error);
            this.showStatus('Delete failed', 'error');
        }
    }

    async listKeys() {
        try {
            const path = this.pathInput.value.trim();
            if (!path) return;

            const response = await fetch(`/api/map/keys?path=${encodeURIComponent(path)}`);

            if (!response.ok) {
                throw new Error('Failed to list keys');
            }

            const data = await response.json();
            this.displayKeys(data.keys || []);
        } catch (error) {
            console.error('Error listing keys:', error);
            this.mapKeys.innerHTML = '<div class="error-message">Failed to load keys</div>';
        }
    }

    displayKeys(keys) {
        if (keys.length === 0) {
            this.mapKeys.innerHTML = '<div class="empty-message">No keys in this map</div>';
            return;
        }

        const keyElements = keys.map(key => {
            return `
                <div class="key-item">
                    <span class="key-name">${this.escapeHtml(key)}</span>
                    <button class="btn-small" onclick="document.getElementById('map-key').value='${this.escapeHtml(key)}'; document.getElementById('map-get').click();">Load</button>
                </div>
            `;
        }).join('');

        this.mapKeys.innerHTML = keyElements;
    }

    async clearAll() {
        if (!confirm('Delete all keys in this map?')) {
            return;
        }

        try {
            const path = this.pathInput.value.trim();
            const response = await fetch(`/api/map/keys?path=${encodeURIComponent(path)}`);
            const data = await response.json();
            const keys = data.keys || [];

            for (const key of keys) {
                await fetch(`/api/map?path=${encodeURIComponent(path)}&key=${encodeURIComponent(key)}`, {
                    method: 'DELETE',
                });
            }

            this.showStatus('Cleared all keys ✓', 'success');
            this.listKeys();
        } catch (error) {
            console.error('Error clearing map:', error);
            this.showStatus('Clear failed', 'error');
        }
    }

    connectSSE() {
        this.eventSource = new EventSource('/api/stream');

        this.eventSource.addEventListener('snapshot', (event) => {
            console.log('SSE snapshot:', event.data);
            this.listKeys();
        });

        this.eventSource.addEventListener('update', (event) => {
            console.log('SSE update:', event.data);
            this.listKeys();
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
