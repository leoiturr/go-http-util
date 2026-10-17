// SPA tab navigation: hash routing, history, accessibility and the
// View Transitions API (progressively enhanced).
import { refreshEditors } from './editors.js';
import { startEpochClock } from '../tools.js';

const TITLES = {
    'base64': 'Base64 Converter',
    'guid': 'GUID Generator',
    'qrcode': 'QR Code Generator',
    'json': 'JSON Formatter & Validator',
    'yaml': 'YAML Formatter & JSON to YAML',
    'sql': 'SQL Formatter & Minifier',
    'regex': 'Regex Playground',
    'url': 'URL Encoder / Decoder & Parser',
    'html-entity': 'HTML Entity Encoder / Decoder',
    'jwt': 'JWT Debugger',
    'epoch': 'Epoch Timestamp Converter',
    'cron': 'Cron Expression Parser',
    'diff': 'Text Diff / Compare',
    'webhook': 'Webhook Tester',
    'docs': 'REST API Reference'
};

export function switchTab(tabId, updateHistory = true) {
    const apply = () => {
        document.querySelectorAll('.tab-module').forEach((module) => {
            const active = module.id === `module-${tabId}`;
            module.classList.toggle('active', active);
            // inert removes hidden tabs from tab order and the a11y tree.
            module.toggleAttribute('inert', !active);
        });

        document.querySelectorAll('.nav-item').forEach((item) => {
            const active = item.dataset.tab === tabId;
            item.classList.toggle('active', active);
            if (active) {
                item.setAttribute('aria-current', 'page');
            } else {
                item.removeAttribute('aria-current');
            }
        });

        const titleEl = document.getElementById('current-tab-title');
        if (titleEl) titleEl.textContent = TITLES[tabId] || 'DevUtils';

        if (tabId === 'epoch') startEpochClock();

        // CodeMirror needs a measure once its container is visible again.
        requestAnimationFrame(() => refreshEditors());
    };

    if (updateHistory) {
        history.pushState({ tabId }, '', '#' + tabId);
    }

    if (updateHistory && document.startViewTransition) {
        document.startViewTransition(apply);
    } else {
        apply();
    }
}

export function initNavigation() {
    // Back / forward buttons.
    window.addEventListener('popstate', (event) => {
        const tabId = (event.state && event.state.tabId) || window.location.hash.substring(1) || 'base64';
        switchTab(tabId, false);
    });

    const initialTab = window.location.hash.substring(1) || 'base64';
    switchTab(initialTab, false);
    history.replaceState({ tabId: initialTab }, '', '#' + initialTab);
}
