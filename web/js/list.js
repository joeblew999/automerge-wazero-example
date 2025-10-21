// M0: List CRDT Module
// Maps to: go/pkg/automerge/list.go, go/pkg/api/list.go

export class ListComponent {
    constructor() {
        this.pathInput = null;
        this.valueInput = null;
        this.indexInput = null;
        this.listItems = null;
        this.listLength = null;
        this.status = null;
        this.eventSource = null;
        this.currentPath = 'ROOT.items';
    }

    init() {
        this.pathInput = document.getElementById('list-path');
        this.valueInput = document.getElementById('list-value');
        this.indexInput = document.getElementById('list-index');
        this.listItems = document.getElementById('list-items');
        this.listLength = document.getElementById('list-length');
        this.status = document.getElementById('list-status');

        if (!this.pathInput) return; // Not on this tab

        // Set default path
        this.pathInput.value = this.currentPath;

        // Event listeners
        document.getElementById('list-push')?.addEventListener('click', () => this.push());
        document.getElementById('list-insert')?.addEventListener('click', () => this.insert());
        document.getElementById('list-refresh')?.addEventListener('click', () => this.loadList());
        document.getElementById('list-clear-all')?.addEventListener('click', () => this.clearAll());

        // Path change listener
        this.pathInput.addEventListener('change', () => {
            this.currentPath = this.pathInput.value;
            this.loadList();
        });

        // Load initial list
        this.loadList();
        this.connectSSE();
    }

    async push() {
        try {
            const path = this.pathInput.value.trim();
            const value = this.valueInput.value;

            if (!path) {
                this.showStatus('Path is required', 'error');
                return;
            }

            this.showStatus('Pushing...', 'info');

            const response = await fetch('/api/list/push', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ path, value }),
            });

            if (!response.ok) {
                throw new Error('Failed to push value');
            }

            this.showStatus('Pushed ✓', 'success');
            this.valueInput.value = '';

            // Refresh list
            setTimeout(() => this.loadList(), 100);
        } catch (error) {
            console.error('Error pushing value:', error);
            this.showStatus('Push failed', 'error');
        }
    }

    async insert() {
        try {
            const path = this.pathInput.value.trim();
            const value = this.valueInput.value;
            const index = parseInt(this.indexInput.value, 10);

            if (!path || isNaN(index)) {
                this.showStatus('Path and valid index are required', 'error');
                return;
            }

            this.showStatus('Inserting...', 'info');

            const response = await fetch('/api/list/insert', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ path, index, value }),
            });

            if (!response.ok) {
                throw new Error('Failed to insert value');
            }

            this.showStatus('Inserted ✓', 'success');
            this.valueInput.value = '';
            this.indexInput.value = '';

            // Refresh list
            setTimeout(() => this.loadList(), 100);
        } catch (error) {
            console.error('Error inserting value:', error);
            this.showStatus('Insert failed', 'error');
        }
    }

    async deleteAt(index) {
        try {
            const path = this.pathInput.value.trim();

            if (!confirm(`Delete item at index ${index}?`)) {
                return;
            }

            const response = await fetch(`/api/list?path=${encodeURIComponent(path)}&index=${index}`, {
                method: 'DELETE',
            });

            if (!response.ok) {
                throw new Error('Failed to delete item');
            }

            this.showStatus('Deleted ✓', 'success');

            // Refresh list
            setTimeout(() => this.loadList(), 100);
        } catch (error) {
            console.error('Error deleting item:', error);
            this.showStatus('Delete failed', 'error');
        }
    }

    async loadList() {
        try {
            const path = this.pathInput.value.trim();
            if (!path) return;

            // Get list length
            const lengthResponse = await fetch(`/api/list/len?path=${encodeURIComponent(path)}`);
            if (!lengthResponse.ok) {
                throw new Error('Failed to get list length');
            }
            const lengthData = await lengthResponse.json();
            const length = lengthData.length || 0;

            this.listLength.textContent = length;

            if (length === 0) {
                this.listItems.innerHTML = '<div class="empty-message">Empty list</div>';
                return;
            }

            // Get all items
            const items = [];
            for (let i = 0; i < length; i++) {
                const response = await fetch(`/api/list?path=${encodeURIComponent(path)}&index=${i}`);
                if (response.ok) {
                    const data = await response.json();
                    items.push({ index: i, value: data.value || '' });
                }
            }

            this.displayItems(items);
        } catch (error) {
            console.error('Error loading list:', error);
            this.listItems.innerHTML = '<div class="error-message">Failed to load list</div>';
        }
    }

    displayItems(items) {
        if (items.length === 0) {
            this.listItems.innerHTML = '<div class="empty-message">Empty list</div>';
            return;
        }

        const itemElements = items.map(item => {
            return `
                <div class="list-item">
                    <span class="item-index">[${item.index}]</span>
                    <span class="item-value">${this.escapeHtml(item.value)}</span>
                    <button class="btn-small btn-danger" onclick="window.listComponent.deleteAt(${item.index})">Delete</button>
                </div>
            `;
        }).join('');

        this.listItems.innerHTML = itemElements;

        // Store reference for inline onclick handlers
        window.listComponent = this;
    }

    async clearAll() {
        if (!confirm('Delete all items in this list?')) {
            return;
        }

        try {
            const path = this.pathInput.value.trim();

            // Get length
            const lengthResponse = await fetch(`/api/list/len?path=${encodeURIComponent(path)}`);
            const lengthData = await lengthResponse.json();
            const length = lengthData.length || 0;

            // Delete from end to beginning (to maintain indices)
            for (let i = length - 1; i >= 0; i--) {
                await fetch(`/api/list?path=${encodeURIComponent(path)}&index=${i}`, {
                    method: 'DELETE',
                });
            }

            this.showStatus('Cleared all items ✓', 'success');
            this.loadList();
        } catch (error) {
            console.error('Error clearing list:', error);
            this.showStatus('Clear failed', 'error');
        }
    }

    connectSSE() {
        this.eventSource = new EventSource('/api/stream');

        this.eventSource.addEventListener('snapshot', (event) => {
            console.log('SSE snapshot:', event.data);
            this.loadList();
        });

        this.eventSource.addEventListener('update', (event) => {
            console.log('SSE update:', event.data);
            this.loadList();
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
