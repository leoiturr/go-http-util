// Dashboard Client Logic

// Tab Navigation
function switchTab(tabId, updateHistory = true) {
    // Hide all modules
    document.querySelectorAll('.tab-module').forEach(module => {
        module.classList.remove('active');
    });

    // Deactivate all nav items
    document.querySelectorAll('.nav-item').forEach(item => {
        item.classList.remove('active');
    });

    // Show selected module
    const targetModule = document.getElementById(`module-${tabId}`);
    if (targetModule) {
        targetModule.classList.add('active');
    }

    // Activate selected nav item
    const clickedBtn = Array.from(document.querySelectorAll('.nav-item')).find(item => 
        item.getAttribute('onclick').includes(`'${tabId}'`)
    );
    if (clickedBtn) {
        clickedBtn.classList.add('active');
    }

    // Update Topbar Title
    const titleMap = {
        'base64': 'Base64 Converter',
        'guid': 'GUID Generator',
        'qrcode': 'QR Code Generator'
    };
    document.getElementById('current-tab-title').innerText = titleMap[tabId] || 'DevUtils';

    // Update browser history state
    if (updateHistory) {
        history.pushState({ tabId: tabId }, '', '#' + tabId);
    }
}

// GUID Slider Updater
function updateRangeVal(val) {
    document.getElementById('range-val-display').innerText = val;
}

// Toggle Custom Prefix Input
function togglePrefixInput(typeValue) {
    const prefixGroup = document.getElementById('prefix-group');
    const helpText = document.getElementById('prefix-format-help');
    
    if (typeValue === 'prefixed' || typeValue === 'prefixed_guid') {
        prefixGroup.style.display = 'block';
        if (typeValue === 'prefixed') {
            helpText.innerText = 'The generated format will be: {prefix}-{12 random hex characters}';
        } else {
            helpText.innerText = 'The generated format will be: {prefix} (padded/truncated to 8 chars) followed by standard UUID blocks';
        }
    } else {
        prefixGroup.style.display = 'none';
    }
}

// Base64 helper to clear textarea and results
function clearBase64() {
    document.getElementById('base64-input').value = '';
    const outputDiv = document.getElementById('base64-output');
    if (outputDiv) {
        outputDiv.innerHTML = '';
    }
    showToast('Base64 inputs cleared', 'success');
}

// Toast Notifications System
function showToast(message, type = 'success') {
    const container = document.getElementById('toast-container');
    if (!container) return;

    const toast = document.createElement('div');
    toast.className = `toast toast-${type}`;
    
    let iconClass = 'fa-circle-check';
    if (type === 'error') iconClass = 'fa-circle-xmark';
    if (type === 'warning') iconClass = 'fa-circle-exclamation';

    toast.innerHTML = `
        <i class="fa-solid ${iconClass}"></i>
        <span>${message}</span>
    `;

    container.appendChild(toast);

    // Auto-remove toast after 3 seconds
    setTimeout(() => {
        toast.classList.add('fade-out');
        toast.addEventListener('animationend', () => {
            toast.remove();
        });
    }, 3000);
}

// Generic Copy To Clipboard
function copyToClipboard(elementId) {
    const element = document.getElementById(elementId);
    if (!element) return;

    // Get either innerText or textContent (for pre-tags)
    const textToCopy = element.innerText || element.textContent;

    navigator.clipboard.writeText(textToCopy)
        .then(() => {
            showToast('Copied to clipboard!', 'success');
        })
        .catch(err => {
            showToast('Failed to copy text', 'error');
            console.error('Copy failed:', err);
        });
}

// Copy All GUIDs in a list
function copyAllGUIDs() {
    const guidValues = Array.from(document.querySelectorAll('.guid-val'))
        .map(el => el.textContent);
    
    if (guidValues.length === 0) return;

    const joined = guidValues.join('\n');

    navigator.clipboard.writeText(joined)
        .then(() => {
            showToast(`Copied all ${guidValues.length} GUIDs!`, 'success');
        })
        .catch(err => {
            showToast('Failed to copy GUIDs', 'error');
            console.error('Copy all failed:', err);
        });
}

// Copy QR Code Image to Clipboard
function copyQRCodeImage() {
    const img = document.getElementById('qr-code-img');
    if (!img) return;

    // Fetch the base64 PNG data, convert it to a Blob and copy it
    const src = img.getAttribute('src');
    
    fetch(src)
        .then(res => res.blob())
        .then(blob => {
            const item = new ClipboardItem({ 'image/png': blob });
            navigator.clipboard.write([item])
                .then(() => {
                    showToast('QR Code image copied to clipboard!', 'success');
                })
                .catch(err => {
                    // Fallback to copying raw base64 string
                    navigator.clipboard.writeText(src)
                        .then(() => {
                            showToast('Copied QR Code Data URL (Image copy blocked by browser)', 'warning');
                        })
                        .catch(fallbackErr => {
                            showToast('Failed to copy QR Code image', 'error');
                            console.error(fallbackErr);
                        });
                });
        })
        .catch(err => {
            showToast('Failed to generate image copy blob', 'error');
            console.error(err);
        });
}

// Slide out history log panel
function toggleHistory() {
    const panel = document.getElementById('history-panel');
    if (panel) {
        panel.classList.toggle('open');
        if (panel.classList.contains('open')) {
            renderHistory();
        }
    }
}

// Operations History Log
function saveToHistory(operation, details) {
    if (!details || details.trim() === '') return;

    const history = JSON.parse(localStorage.getItem('devutils_history') || '[]');
    const timestamp = new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' });
    
    const entry = {
        id: Date.now(),
        operation: operation,
        details: details,
        time: timestamp
    };

    // Keep only the last 30 entries
    history.unshift(entry);
    if (history.length > 30) {
        history.pop();
    }

    localStorage.setItem('devutils_history', JSON.stringify(history));
}

// Render History Panel items
function renderHistory() {
    const list = document.getElementById('history-list');
    if (!list) return;

    const history = JSON.parse(localStorage.getItem('devutils_history') || '[]');

    if (history.length === 0) {
        list.innerHTML = `
            <div class="placeholder-content mt-4" style="text-align: center; color: var(--text-dark);">
                <i class="fa-solid fa-clock-rotate-left" style="font-size: 24px; opacity: 0.15; margin-bottom: 8px;"></i>
                <p style="font-size: 12px;">No recent operations recorded.</p>
            </div>
        `;
        return;
    }

    list.innerHTML = history.map(entry => `
        <div class="history-item" onclick="loadHistoryItem('${entry.operation}', \`${entry.details.replace(/`/g, '\\`').replace(/\n/g, '\\n')}\`)">
            <div class="history-meta">
                <span class="history-type">${entry.operation}</span>
                <span class="history-time">${entry.time}</span>
            </div>
            <div class="history-content">${escapeHTML(entry.details)}</div>
        </div>
    `).join('');
}

// Clear History Log
function clearHistoryLog() {
    localStorage.removeItem('devutils_history');
    renderHistory();
    showToast('History log cleared', 'success');
}

// Utility to escape HTML strings
function escapeHTML(str) {
    return str
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;')
        .replace(/"/g, '&quot;')
        .replace(/'/g, '&#039;');
}

// Utility to truncate string
function truncateString(str, len) {
    if (str.length <= len) return str;
    return str.substring(0, len) + '...';
}

// Load a previous input from history into the relevant tab
function loadHistoryItem(operation, details) {
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
        // Parse config representation (e.g. "v7 x25")
        const parts = details.split(' x');
        if (parts.length === 2) {
            const type = parts[0];
            const count = parts[1];
            document.getElementById('guid-type').value = type;
            togglePrefixInput(type);
            document.getElementById('guid-count').value = count;
            updateRangeVal(count);
            showToast('Restored settings to GUID panel', 'success');
        }
    }
}

// Global HTMX Error Handling
document.addEventListener('htmx:responseError', function(evt) {
    if (evt.detail.xhr.status === 429) {
        showToast('Too many requests. Please wait a moment.', 'error');
    } else {
        showToast('An error occurred during request processing.', 'error');
    }
});

// Listen to popstate event (back/forward browser buttons)
window.addEventListener('popstate', function(event) {
    const tabId = (event.state && event.state.tabId) || window.location.hash.substring(1) || 'base64';
    switchTab(tabId, false);
});

// Initialize active tab on page load based on URL hash
document.addEventListener('DOMContentLoaded', () => {
    const initialTab = window.location.hash.substring(1) || 'base64';
    switchTab(initialTab, false);
    history.replaceState({ tabId: initialTab }, '', '#' + initialTab);
});
