// DevUtils frontend entry point.
// Loaded as <script type="module">; imports the custom elements (whose
// definitions upgrade any <copy-button>/<toast-stack> in the markup or in
// HTMX fragments), registers delegated actions and boots the app.
import './components.js';

import { registerActions, initDelegation } from './core/delegate.js';
import { initTheme, toggleTheme } from './core/theme.js';
import { initNavigation, switchTab } from './core/nav.js';
import { initHistoryPanel, clearHistoryLog, loadHistoryItem, loadHistory, saveToHistory } from './core/history.js';
import { mountEditor, initResultEditors, refreshEditors } from './core/editors.js';
import { showToast } from './core/ui.js';
import { truncateString } from './core/utils.js';
import { initRegex, loadRegexPreset, setRegexEngine, clearRegex } from './features/regex.js';
import { initWebhook } from './features/webhook.js';
import * as tools from './tools.js';

// HTMX-driven history buttons/forms declare data-history-op (+ optional
// data-history-from). Form submits and button clicks are disambiguated by
// the event type so a form is logged once, on submit.
function logHistory(el, evt) {
    if (el.tagName === 'FORM') {
        if (evt.type !== 'submit') return;
    } else if (evt.type !== 'click') {
        return;
    }

    const op = el.dataset.historyOp;
    if (!op) return;

    let details = '';
    if (op === 'GUID Generation') {
        details = `${document.getElementById('guid-type').value} x${document.getElementById('guid-count').value}`;
    } else {
        const from = el.dataset.historyFrom ? document.getElementById(el.dataset.historyFrom) : null;
        details = from ? from.value : '';
    }
    saveToHistory(op, truncateString(details, 30));
}

registerActions({
    // Navigation & shell
    'tab': (el) => switchTab(el.dataset.tab),
    'theme-toggle': () => toggleTheme(),
    'history-clear': () => clearHistoryLog(),
    'history-item': (el) => {
        const entry = loadHistory()[Number(el.dataset.index)];
        if (entry) loadHistoryItem(entry.operation, entry.details);
    },
    'log': logHistory,

    // Client-side tools
    'base64': (el) => tools.runBase64(el.dataset.op),
    'clear-base64': () => tools.clearBase64(),
    'json': (el) => tools.runJSON(el.dataset.op),
    'clear-json': () => tools.clearJSON(),
    'clear-sql': () => tools.clearSQL(),
    'clear-yaml': () => tools.clearYAML(),
    'url': (el) => tools.runURL(el.dataset.op),
    'url-rebuild': () => tools.rebuildURL(),
    'clear-url': () => tools.clearURL(),
    'html-entity': (el) => tools.runHTMLEntity(el.dataset.op),
    'clear-html-entity': () => tools.clearHTMLEntity(),
    'jwt': () => tools.runJWT(),
    'clear-jwt': () => tools.clearJWT(),
    'diff': (el, evt) => { evt.preventDefault(); tools.runTextDiff(); },
    'clear-diff': () => tools.clearDiff(),
    'epoch': (el, evt) => { evt.preventDefault(); tools.runEpoch(); },
    'epoch-use-current': () => tools.useCurrentEpoch(),
    'epoch-from-text': () => tools.syncDateTimePickerFromText(),
    'epoch-from-datetime': () => tools.syncTextInputFromDateTime(),
    'clear-epoch': () => tools.clearEpoch(),
    'clear-cron': () => tools.clearCron(),

    // Regex
    'regex-preset': (el) => loadRegexPreset(el.value),
    'regex-engine': (el) => setRegexEngine(el.value),
    'clear-regex': () => clearRegex(),

    // GUID
    'guid-count': (el) => tools.updateRangeVal(el.value),
    'guid-type': (el) => tools.togglePrefixInput(el.value)
});

function boot() {
    initTheme();
    initDelegation();
    initNavigation();
    initHistoryPanel();
    initRegex();
    initWebhook();

    // Editable CodeMirror editors.
    mountEditor('json', 'json-input', 'application/json');
    mountEditor('yaml', 'yaml-input', 'text/x-yaml');
    mountEditor('sql', 'sql-input', 'text/x-sql');

    // JWT input syntax highlighting overlay.
    const jwtInput = document.getElementById('jwt-input');
    if (jwtInput) {
        jwtInput.addEventListener('input', tools.highlightJWTInput);
        jwtInput.addEventListener('scroll', tools.syncJWTScroll);
        jwtInput.addEventListener('keyup', tools.syncJWTScroll);
        jwtInput.addEventListener('click', tools.syncJWTScroll);
        tools.highlightJWTInput();
    }

    // Upgrade any read-only result editors inside swapped-in fragments and
    // restore the URL parser's reconstructed value. HTMX 4 dispatches
    // htmx:after:settle on the swapped element, so evt.target is it.
    document.addEventListener('htmx:after:settle', (evt) => {
        const target = evt.target instanceof Element ? evt.target : document;
        initResultEditors(target);
        if (target.querySelector && target.querySelector('#reconstructed-url')) {
            tools.rebuildURL();
        }
    });

    // Global HTMX error handling.
    document.addEventListener('htmx:response:error', (evt) => {
        const status = evt.detail && evt.detail.ctx && evt.detail.ctx.response
            ? evt.detail.ctx.response.status
            : 0;
        if (status === 429) {
            showToast('Too many requests. Please wait a moment.', 'error');
        } else {
            showToast('An error occurred during request processing.', 'error');
        }
    });

    // Refresh editors once fonts/layout settle.
    requestAnimationFrame(() => refreshEditors());
}

if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', boot);
} else {
    boot();
}
