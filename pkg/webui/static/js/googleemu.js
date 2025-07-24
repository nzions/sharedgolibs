/*
 * GoogleEmu Web Components Library
 * Version: 1.0.0
 * 
 * JavaScript utilities and components for GoogleEmu frontend interfaces
 */

class GoogleEmuComponents {
    static version = "1.0.0";

    constructor() {
        this.currentTheme = localStorage.getItem('ge-theme') || 'glassmorphism';
        this.autoRefreshInterval = null;
        this.init();
    }

    init() {
        this.applyTheme(this.currentTheme);
        this.setupThemeSwitcher();
        this.registerCustomElements();
    }

    // Theme management
    applyTheme(themeName) {
        document.documentElement.setAttribute('data-theme', themeName);
        this.currentTheme = themeName;
        localStorage.setItem('ge-theme', themeName);

        // Emit theme change event
        const event = new CustomEvent('ge-theme-changed', {
            detail: { theme: themeName }
        });
        document.dispatchEvent(event);
    }

    setupThemeSwitcher() {
        const existing = document.querySelector('.ge-theme-switcher');
        if (existing) return;

        const switcher = document.createElement('div');
        switcher.className = 'ge-theme-switcher';
        switcher.innerHTML = `
            <select id="ge-theme-select">
                <option value="glassmorphism">🌌 Glassmorphism</option>
                <option value="professional">💼 Professional</option>
                <option value="hacker">🖥️ Hacker</option>
                <option value="puppies">🐕 Puppies</option>
                <option value="weyland">🛸 Weyland-Yutani</option>
            </select>
        `;

        document.body.appendChild(switcher);

        const select = document.getElementById('ge-theme-select');
        select.value = this.currentTheme;
        select.addEventListener('change', (e) => {
            this.applyTheme(e.target.value);
        });
    }

    // Auto-refresh functionality
    startAutoRefresh(intervalSeconds = 30, callback = () => location.reload()) {
        this.stopAutoRefresh(); // Clear any existing interval

        let countdown = intervalSeconds;
        const indicator = this.createAutoRefreshIndicator(countdown);

        this.autoRefreshInterval = setInterval(() => {
            countdown--;
            this.updateRefreshIndicator(countdown);

            if (countdown <= 0) {
                countdown = intervalSeconds;
                callback();
            }
        }, 1000);

        return this.autoRefreshInterval;
    }

    stopAutoRefresh() {
        if (this.autoRefreshInterval) {
            clearInterval(this.autoRefreshInterval);
            this.autoRefreshInterval = null;
        }
    }

    createAutoRefreshIndicator(seconds) {
        const existing = document.querySelector('.ge-auto-refresh');
        if (existing) existing.remove();

        const indicator = document.createElement('div');
        indicator.className = 'ge-auto-refresh';
        indicator.innerHTML = `
            🔄 Auto-refresh: <span class="ge-refresh-indicator">${seconds}s</span>
        `;

        document.body.appendChild(indicator);
        return indicator;
    }

    updateRefreshIndicator(seconds) {
        const indicator = document.querySelector('.ge-refresh-indicator');
        if (indicator) {
            indicator.textContent = `${seconds}s`;
        }
    }

    // Utility functions
    static createServiceCard(service) {
        const statusClass = service.status === 'running' ? 'running' :
            service.status === 'stopped' ? 'stopped' : 'warning';

        return `
            <div class="ge-service-card ge-fade-in">
                <h3>
                    <span class="ge-status-dot ge-status-dot-${statusClass}"></span>
                    ${service.name}
                </h3>
                <div class="service-info">
                    <span class="ge-status ge-status-${statusClass}">${service.status}</span>
                    <span class="port">Port: ${service.port}</span>
                    ${service.url ? `<a href="${service.url}" target="_blank" class="ge-btn ge-btn-primary ge-btn-sm">Open</a>` : ''}
                </div>
                ${service.description ? `<p class="description">${service.description}</p>` : ''}
                ${service.uptime ? `<div class="uptime">Uptime: ${service.uptime}</div>` : ''}
            </div>
        `;
    }

    static createHeader(title, subtitle = '', version = '') {
        return `
            <div class="ge-header ge-fade-in">
                <h1>${title}</h1>
                ${subtitle ? `<div class="subtitle">${subtitle}</div>` : ''}
                ${version ? `<div class="version">v${version}</div>` : ''}
            </div>
        `;
    }

    static createButton(text, type = 'primary', size = '', onclick = '', href = '') {
        const sizeClass = size ? `ge-btn-${size}` : '';
        const clickHandler = onclick ? `onclick="${onclick}"` : '';

        if (href) {
            return `<a href="${href}" class="ge-btn ge-btn-${type} ${sizeClass}">${text}</a>`;
        } else {
            return `<button class="ge-btn ge-btn-${type} ${sizeClass}" ${clickHandler}>${text}</button>`;
        }
    }

    static createGrid(items, columns = 'auto') {
        const gridClass = columns === 'auto' ? 'ge-grid-auto' : `ge-grid-${columns}`;
        const itemsHtml = Array.isArray(items) ? items.join('') : items;

        return `<div class="ge-grid ${gridClass}">${itemsHtml}</div>`;
    }

    // Animation utilities
    static fadeIn(element, duration = 500) {
        element.style.opacity = '0';
        element.style.transition = `opacity ${duration}ms ease-in-out`;

        requestAnimationFrame(() => {
            element.style.opacity = '1';
        });
    }

    static slideUp(element, duration = 500) {
        element.style.opacity = '0';
        element.style.transform = 'translateY(20px)';
        element.style.transition = `opacity ${duration}ms ease-out, transform ${duration}ms ease-out`;

        requestAnimationFrame(() => {
            element.style.opacity = '1';
            element.style.transform = 'translateY(0)';
        });
    }

    // HTTP utilities for AJAX
    static async fetchJSON(url, options = {}) {
        try {
            const response = await fetch(url, {
                headers: {
                    'Content-Type': 'application/json',
                    ...options.headers
                },
                ...options
            });

            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }

            return await response.json();
        } catch (error) {
            console.error('Fetch error:', error);
            throw error;
        }
    }

    static async postJSON(url, data, options = {}) {
        return this.fetchJSON(url, {
            method: 'POST',
            body: JSON.stringify(data),
            ...options
        });
    }

    // Notification system
    static showNotification(message, type = 'info', duration = 5000) {
        const notification = document.createElement('div');
        notification.className = `ge-notification ge-notification-${type} ge-fade-in`;
        notification.style.cssText = `
            position: fixed;
            top: 20px;
            right: 20px;
            z-index: 9999;
            padding: 16px 20px;
            border-radius: 8px;
            color: white;
            font-weight: 500;
            box-shadow: 0 4px 12px rgba(0,0,0,0.15);
            max-width: 400px;
            word-wrap: break-word;
        `;

        // Set background color based on type
        const colors = {
            info: '#3b82f6',
            success: '#10b981',
            warning: '#f59e0b',
            danger: '#ef4444'
        };
        notification.style.backgroundColor = colors[type] || colors.info;

        notification.textContent = message;
        document.body.appendChild(notification);

        // Auto-remove after duration
        setTimeout(() => {
            notification.style.opacity = '0';
            setTimeout(() => {
                if (notification.parentNode) {
                    notification.parentNode.removeChild(notification);
                }
            }, 300);
        }, duration);

        return notification;
    }

    // Register custom web components
    registerCustomElements() {
        // Service Card Web Component
        if (!customElements.get('ge-service-card')) {
            customElements.define('ge-service-card', class extends HTMLElement {
                connectedCallback() {
                    const name = this.getAttribute('name') || '';
                    const status = this.getAttribute('status') || 'unknown';
                    const port = this.getAttribute('port') || '';
                    const url = this.getAttribute('url') || '';
                    const description = this.getAttribute('description') || '';
                    const uptime = this.getAttribute('uptime') || '';

                    this.innerHTML = GoogleEmuComponents.createServiceCard({
                        name, status, port, url, description, uptime
                    });
                }
            });
        }

        // Header Web Component
        if (!customElements.get('ge-header')) {
            customElements.define('ge-header', class extends HTMLElement {
                connectedCallback() {
                    const title = this.getAttribute('title') || 'GoogleEmu';
                    const subtitle = this.getAttribute('subtitle') || '';
                    const version = this.getAttribute('version') || '';

                    this.innerHTML = GoogleEmuComponents.createHeader(title, subtitle, version);
                }
            });
        }

        // Button Web Component
        if (!customElements.get('ge-button')) {
            customElements.define('ge-button', class extends HTMLElement {
                connectedCallback() {
                    const text = this.getAttribute('text') || this.textContent || 'Button';
                    const type = this.getAttribute('type') || 'primary';
                    const size = this.getAttribute('size') || '';
                    const onclick = this.getAttribute('onclick') || '';
                    const href = this.getAttribute('href') || '';

                    this.innerHTML = GoogleEmuComponents.createButton(text, type, size, onclick, href);
                }
            });
        }
    }

    // Page builder utilities
    static buildFullPage(title, content, options = {}) {
        const {
            subtitle = '',
            version = '',
            autoRefresh = false,
            refreshInterval = 30,
            theme = 'weyland'
        } = options;

        const autoRefreshScript = autoRefresh ? `
            <script>
                const ge = new GoogleEmuComponents();
                ge.startAutoRefresh(${refreshInterval});
            </script>
        ` : '';

        return `
<!DOCTYPE html>
<html lang="en" data-theme="${theme}">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>${title}</title>
    <link rel="stylesheet" href="/static/css/googleemu.css">
</head>
<body>
    <div class="ge-container">
        ${GoogleEmuComponents.createHeader(title, subtitle, version)}
        ${content}
    </div>
    
    <script src="/static/js/googleemu.js"></script>
    ${autoRefreshScript}
</body>
</html>
        `;
    }

    // Theme-specific helpers
    static getThemeEmoji() {
        const emojis = {
            glassmorphism: '🌌',
            professional: '💼',
            hacker: '🖥️',
            puppies: '🐕'
        };
        return emojis[this.currentTheme] || '🌌';
    }

    static getThemeDescription() {
        const descriptions = {
            glassmorphism: 'Futuristic glassmorphism design',
            professional: 'Clean corporate interface',
            hacker: 'Terminal-inspired dark theme',
            puppies: 'Fun and colorful design'
        };
        return descriptions[this.currentTheme] || descriptions.glassmorphism;
    }
}

// Initialize components when DOM is ready
document.addEventListener('DOMContentLoaded', () => {
    window.GoogleEmu = new GoogleEmuComponents();
});

// Expose utilities globally
window.GoogleEmuComponents = GoogleEmuComponents;
