// Operation history log + slide-out popover panel.
import { safeStorage } from './storage.js';
import { fromTemplate } from './utils.js';
import { showToast } from './ui.js';
import { switchTab } from './nav.js';
import { setEditorValue } from './editors.js';
import { togglePrefixInput, updateRangeVal, highlightJWTInput } from '../tools.js';

const HISTORY_KEY = 'devutils_history';
const HISTORY_LIMIT = 30;

export function loadHistory() {
    try {
        const parsed = JSON.parse(safeStorage.getItem(HISTORY_KEY) || '[]');
        return Array.isArray(parsed) ? parsed : [];
    } catch (e) {
        console.warn('Ignoring invalid history data:', e);
        return [];
    }
}

export function saveToHistory(operation, details) {
    if (typeof details !== 'string' || details.trim() === '') return;

    const history = loadHistory();
    const timestamp = new Date().toLocaleTimeString([], {
        hour: '2-digit', minute: '2-digit', second: '2-digit'
    });

    history.unshift({ id: Date.now(), operation, details, time: timestamp });
    if (history.length > HISTORY_LIMIT) history.pop();

    safeStorage.setItem(HISTORY_KEY, JSON.stringify(history));
}

export function renderHistory() {
    const list = document.getElementById('history-list');
    if (!list) return;

    const history = loadHistory();
    list.replaceChildren();

    if (history.length === 0) {
        const empty = fromTemplate('tpl-history-empty');
        list.appendChild(empty);
        return;
    }

    const fragment = document.createDocumentFragment();
    history.forEach((entry, index) => {
        const node = fromTemplate('tpl-history-item');
        node.querySelector('.history-type').textContent = entry.operation;
        node.querySelector('.history-time').textContent = entry.time;
        node.querySelector('.history-content').textContent = entry.details;
        const button = node.querySelector('.history-item');
        button.dataset.index = String(index);
        fragment.appendChild(node);
    });
    list.appendChild(fragment);
}

export function clearHistoryLog() {
    safeStorage.removeItem(HISTORY_KEY);
    renderHistory();
    showToast('History log cleared', 'success');
}

export function loadHistoryItem(operation, details) {
    if (operation.includes('Base64')) {
        switchTab('base64');
        document.getElementById('base64-input').value = details;
        showToast('Restored text to Base64 panel', 'success');
    } else if (operation.includes('QR Code')) {
        switchTab('qrcode');
        document.getElementById('qrcode-text').value = details;
        showToast('Restored text to QR Code panel', 'success');
    } else if (operation.includes('GUID')) {
        switchTab('guid');
        const parts = details.split(' x');
        if (parts.length === 2) {
            document.getElementById('guid-type').value = parts[0];
            togglePrefixInput(parts[0]);
            document.getElementById('guid-count').value = parts[1];
            updateRangeVal(parts[1]);
            showToast('Restored settings to GUID panel', 'success');
        }
    } else if (operation.includes('JSON to YAML')) {
        switchTab('json');
        document.getElementById('json-input').value = details;
        setEditorValue('json', details);
        showToast('Restored payload to JSON panel', 'success');
    } else if (operation.includes('YAML to JSON')) {
        switchTab('yaml');
        document.getElementById('yaml-input').value = details;
        setEditorValue('yaml', details);
        showToast('Restored payload to YAML panel', 'success');
    } else if (operation.includes('JSON')) {
        switchTab('json');
        document.getElementById('json-input').value = details;
        setEditorValue('json', details);
        showToast('Restored payload to JSON panel', 'success');
    } else if (operation.includes('URL')) {
        switchTab('url');
        document.getElementById('url-input').value = details;
        showToast('Restored payload to URL panel', 'success');
    } else if (operation.includes('Regex')) {
        switchTab('regex');
        document.getElementById('regex-preset').value = 'custom';
        document.getElementById('regex-pattern').value = details;
        showToast('Restored pattern to Regex Playground', 'success');
    } else if (operation.includes('JWT')) {
        switchTab('jwt');
        document.getElementById('jwt-input').value = details;
        highlightJWTInput();
        showToast('Restored token to JWT panel', 'success');
    } else if (operation.includes('Epoch')) {
        switchTab('epoch');
        document.getElementById('epoch-input').value = details;
        showToast('Restored timestamp to Epoch panel', 'success');
    } else if (operation.includes('Cron')) {
        switchTab('cron');
        document.getElementById('cron-expression').value = details;
        showToast('Restored expression to Cron panel', 'success');
    } else if (operation.includes('YAML')) {
        switchTab('yaml');
        document.getElementById('yaml-input').value = details;
        setEditorValue('yaml', details);
        showToast('Restored payload to YAML panel', 'success');
    } else if (operation.includes('SQL')) {
        switchTab('sql');
        document.getElementById('sql-input').value = details;
        setEditorValue('sql', details);
        showToast('Restored query to SQL panel', 'success');
    }
}

export function initHistoryPanel() {
    const panel = document.getElementById('history-panel');
    if (!panel) return;
    // Render fresh each time the popover opens.
    panel.addEventListener('toggle', (event) => {
        if (event.newState === 'open') renderHistory();
    });
}
