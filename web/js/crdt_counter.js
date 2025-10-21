// M0: Counter CRDT Module
// Maps to: go/pkg/automerge/counter.go, go/pkg/api/counter.go

export class CounterComponent {
    constructor() {
        this.pathInput = null;
        this.keyInput = null;
        this.deltaInput = null;
        this.counterValue = null;
        this.status = null;
        this.eventSource = null;
        this.currentPath = 'ROOT';
        this.currentKey = 'counter';
    }

    init() {
        this.pathInput = document.getElementById('counter-path');
        this.keyInput = document.getElementById('counter-key');
        this.deltaInput = document.getElementById('counter-delta');
        this.counterValue = document.getElementById('counter-value');
        this.status = document.getElementById('counter-status');

        if (!this.pathInput) return; // Not on this tab

        // Set defaults
        this.pathInput.value = this.currentPath;
        this.keyInput.value = this.currentKey;
        this.deltaInput.value = '1';

        // Event listeners
        document.getElementById('counter-increment')?.addEventListener('click', () => this.increment());
        document.getElementById('counter-decrement')?.addEventListener('click', () => this.decrement());
        document.getElementById('counter-add-custom')?.addEventListener('click', () => this.addCustom());
        document.getElementById('counter-refresh')?.addEventListener('click', () => this.getValue());
        document.getElementById('counter-reset')?.addEventListener('click', () => this.reset());

        // Quick increment buttons
        document.getElementById('counter-plus-1')?.addEventListener('click', () => this.quickIncrement(1));
        document.getElementById('counter-plus-5')?.addEventListener('click', () => this.quickIncrement(5));
        document.getElementById('counter-plus-10')?.addEventListener('click', () => this.quickIncrement(10));
        document.getElementById('counter-minus-1')?.addEventListener('click', () => this.quickIncrement(-1));
        document.getElementById('counter-minus-5')?.addEventListener('click', () => this.quickIncrement(-5));
        document.getElementById('counter-minus-10')?.addEventListener('click', () => this.quickIncrement(-10));

        // Load initial value
        this.getValue();
        this.connectSSE();
    }

    async increment() {
        await this.addDelta(1);
    }

    async decrement() {
        await this.addDelta(-1);
    }

    async addCustom() {
        const delta = parseInt(this.deltaInput.value, 10);
        if (isNaN(delta)) {
            this.showStatus('Invalid delta value', 'error');
            return;
        }
        await this.addDelta(delta);
    }

    async quickIncrement(delta) {
        await this.addDelta(delta);
    }

    async addDelta(delta) {
        try {
            const path = this.pathInput.value.trim();
            const key = this.keyInput.value.trim();

            if (!path || !key) {
                this.showStatus('Path and key are required', 'error');
                return;
            }

            this.showStatus('Updating...', 'info');

            const response = await fetch('/api/counter', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ path, key, delta }),
            });

            if (!response.ok) {
                throw new Error('Failed to increment counter');
            }

            const data = await response.json();
            this.updateDisplay(data.value);
            this.showStatus(`${delta > 0 ? '+' : ''}${delta} ✓`, 'success');
        } catch (error) {
            console.error('Error incrementing counter:', error);
            this.showStatus('Update failed', 'error');
        }
    }

    async getValue() {
        try {
            const path = this.pathInput.value.trim();
            const key = this.keyInput.value.trim();

            if (!path || !key) {
                this.showStatus('Path and key are required', 'error');
                return;
            }

            const response = await fetch(`/api/counter?path=${encodeURIComponent(path)}&key=${encodeURIComponent(key)}`);

            if (!response.ok) {
                throw new Error('Failed to get counter value');
            }

            const data = await response.json();
            this.updateDisplay(data.value);
            this.showStatus('Loaded ✓', 'success');
        } catch (error) {
            console.error('Error getting counter:', error);
            this.showStatus('Load failed', 'error');
            this.updateDisplay(0);
        }
    }

    async reset() {
        try {
            const path = this.pathInput.value.trim();
            const key = this.keyInput.value.trim();

            if (!confirm(`Reset counter "${key}" to 0?`)) {
                return;
            }

            // Get current value
            const getResponse = await fetch(`/api/counter?path=${encodeURIComponent(path)}&key=${encodeURIComponent(key)}`);
            const currentData = await getResponse.json();
            const currentValue = currentData.value || 0;

            // Subtract current value to reset to 0
            const delta = -currentValue;

            const response = await fetch('/api/counter', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ path, key, delta }),
            });

            if (!response.ok) {
                throw new Error('Failed to reset counter');
            }

            this.updateDisplay(0);
            this.showStatus('Reset to 0 ✓', 'success');
        } catch (error) {
            console.error('Error resetting counter:', error);
            this.showStatus('Reset failed', 'error');
        }
    }

    updateDisplay(value) {
        if (this.counterValue) {
            this.counterValue.textContent = value;

            // Add animation
            this.counterValue.classList.add('counter-flash');
            setTimeout(() => {
                this.counterValue.classList.remove('counter-flash');
            }, 300);
        }
    }

    connectSSE() {
        this.eventSource = new EventSource('/api/stream');

        this.eventSource.addEventListener('snapshot', (event) => {
            console.log('SSE snapshot:', event.data);
            this.getValue();
        });

        this.eventSource.addEventListener('update', (event) => {
            console.log('SSE update:', event.data);
            this.getValue();
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
