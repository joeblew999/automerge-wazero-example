// Main App Orchestrator
// Handles tab switching and component lifecycle
// 1:1 mapping: coordinates all web/js/*.js modules

import { TextComponent } from './text.js';
import { SyncComponent } from './sync.js';
import { RichTextComponent } from './richtext.js';

class App {
    constructor() {
        this.components = {
            text: new TextComponent(),
            sync: new SyncComponent(),
            richtext: new RichTextComponent(),
        };
        this.currentTab = 'text';
        this.sseConnection = null;
    }

    async init() {
        console.log('🚀 Automerge WASI Demo - App initializing...');

        // Setup tab switching
        this.setupTabs();

        // Initialize connection status
        this.setupConnectionStatus();

        // Load component HTML
        await this.loadComponents();

        // Initialize first tab
        this.switchTab('text');

        console.log('✅ App initialized');
    }

    setupTabs() {
        const tabs = document.querySelectorAll('.tab-btn');
        tabs.forEach(tab => {
            tab.addEventListener('click', () => {
                const tabName = tab.dataset.tab;
                this.switchTab(tabName);
            });
        });
    }

    async switchTab(tabName) {
        console.log(`Switching to tab: ${tabName}`);

        // Update tab buttons
        document.querySelectorAll('.tab-btn').forEach(btn => {
            btn.classList.toggle('active', btn.dataset.tab === tabName);
        });

        // Update tab panes
        document.querySelectorAll('.tab-pane').forEach(pane => {
            pane.classList.toggle('active', pane.id === `tab-${tabName}`);
        });

        // Destroy old component
        if (this.components[this.currentTab]?.destroy) {
            this.components[this.currentTab].destroy();
        }

        // Initialize new component
        this.currentTab = tabName;
        if (this.components[tabName]?.init) {
            this.components[tabName].init();
        }
    }

    async loadComponents() {
        // Components are already in HTML (from index.html)
        // Just need to initialize them when their tab is active
        console.log('Components loaded from index.html');
    }

    setupConnectionStatus() {
        // Setup SSE for global connection status
        this.sseConnection = new EventSource('/api/stream');

        this.sseConnection.onopen = () => {
            this.updateConnectionStatus(true);
        };

        this.sseConnection.onerror = () => {
            this.updateConnectionStatus(false);
            this.sseConnection.close();

            // Reconnect after 3 seconds
            setTimeout(() => this.setupConnectionStatus(), 3000);
        };

        // Get server info
        this.fetchServerInfo();
    }

    updateConnectionStatus(connected) {
        const statusEl = document.getElementById('connection-status');
        if (statusEl) {
            statusEl.textContent = connected ? 'Connected' : 'Disconnected';
            statusEl.className = `status-badge ${connected ? 'connected' : 'disconnected'}`;
        }
    }

    async fetchServerInfo() {
        try {
            const response = await fetch('/api/text');
            if (response.ok) {
                const serverInfoEl = document.getElementById('server-info');
                if (serverInfoEl) {
                    serverInfoEl.textContent = 'Server: Go + wazero + Rust WASI';
                }
            }
        } catch (error) {
            console.error('Failed to fetch server info:', error);
        }
    }
}

// Initialize app when DOM is ready
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', () => {
        const app = new App();
        app.init();
    });
} else {
    const app = new App();
    app.init();
}
