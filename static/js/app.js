// Dashboard Client Logic

// Safe localStorage helper to prevent SecurityError in Private Browsing / Firefox
const safeStorage = {
    getItem(key) {
        try {
            return localStorage.getItem(key);
        } catch (e) {
            console.warn('Storage disabled or blocked:', e);
            return null;
        }
    },
    setItem(key, value) {
        try {
            localStorage.setItem(key, value);
        } catch (e) {
            console.warn('Storage disabled or blocked:', e);
        }
    },
    removeItem(key) {
        try {
            localStorage.removeItem(key);
        } catch (e) {
            console.warn('Storage disabled or blocked:', e);
        }
    }
};

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

    // Trigger live clock initialization if switching to epoch tab
    if (tabId === 'epoch') {
        startEpochClock();
    }

    // Update Topbar Title
    const titleMap = {
        'base64': 'Base64 Converter',
        'guid': 'GUID Generator',
        'qrcode': 'QR Code Generator',
        'json': 'JSON Formatter & Validator',
        'url': 'URL Encoder / Decoder & Parser',
        'jwt': 'JWT Debugger',
        'epoch': 'Epoch Timestamp Converter',
        'docs': 'REST API Reference'
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

// Clear helper for JSON Formatter
function clearJSON() {
    document.getElementById('json-input').value = '';
    const outputDiv = document.getElementById('json-output');
    if (outputDiv) {
        outputDiv.innerHTML = '';
    }
    showToast('JSON inputs cleared', 'success');
}

// Clear helper for URL Encoder
function clearURL() {
    document.getElementById('url-input').value = '';
    const outputDiv = document.getElementById('url-output');
    if (outputDiv) {
        outputDiv.innerHTML = '';
    }
    showToast('URL inputs cleared', 'success');
}

// Clear helper for JWT Debugger
function clearJWT() {
    document.getElementById('jwt-input').value = '';
    const outputDiv = document.getElementById('jwt-output');
    if (outputDiv) {
        outputDiv.innerHTML = '';
    }
    showToast('JWT inputs cleared', 'success');
}

// Clear helper for Epoch Converter
function clearEpoch() {
    document.getElementById('epoch-input').value = '';
    const outputDiv = document.getElementById('epoch-output');
    if (outputDiv) {
        outputDiv.innerHTML = '';
    }
    showToast('Epoch inputs cleared', 'success');
}

// URL query parameters builder helpers
function copyTableValue(btn) {
    const input = btn.closest('tr').querySelector('.param-value');
    if (!input) return;
    navigator.clipboard.writeText(input.value)
        .then(() => showToast('Parameter value copied!', 'success'))
        .catch(() => showToast('Failed to copy value', 'error'));
}

function rebuildURL() {
    const container = document.querySelector('.table-container');
    if (!container) return;

    const op = container.getAttribute('data-op');
    const scheme = container.getAttribute('data-scheme') || '';
    const host = container.getAttribute('data-host') || '';
    const path = container.getAttribute('data-path') || '';

    const params = [];
    const rows = document.querySelectorAll('.params-table tbody tr');
    rows.forEach(row => {
        const keyInput = row.querySelector('.param-key');
        const valInput = row.querySelector('.param-value');
        if (keyInput && valInput) {
            const key = keyInput.value.trim();
            const val = valInput.value;
            if (key !== '') {
                params.push(encodeURIComponent(key) + '=' + encodeURIComponent(val));
            }
        }
    });

    const queryString = params.join('&');
    let finalURL = '';
    if (op === 'Parse' && (scheme || host || path)) {
        finalURL = (scheme ? scheme + '://' : '') + host + path;
        if (queryString) {
            finalURL += '?' + queryString;
        }
    } else {
        finalURL = queryString;
    }

    const outputEl = document.getElementById('reconstructed-url');
    if (outputEl) {
        outputEl.textContent = finalURL;
    }
}

// Live epoch clock helpers
let liveEpochInterval = null;
function startEpochClock() {
    const valEl = document.getElementById('live-epoch-value');
    if (!valEl) return;
    
    if (liveEpochInterval) clearInterval(liveEpochInterval);
    
    const updateTime = () => {
        const nowSec = Math.floor(Date.now() / 1000);
        valEl.textContent = nowSec;
    };
    updateTime();
    liveEpochInterval = setInterval(updateTime, 1000);
}

function copyLiveEpoch() {
    const valEl = document.getElementById('live-epoch-value');
    if (!valEl) return;
    navigator.clipboard.writeText(valEl.textContent)
        .then(() => showToast('Live epoch timestamp copied!', 'success'))
        .catch(() => showToast('Failed to copy timestamp', 'error'));
}
function useCurrentEpoch() {
    const valEl = document.getElementById('live-epoch-value');
    const inputEl = document.getElementById('epoch-input');
    if (valEl && inputEl) {
        inputEl.value = valEl.textContent;
        showToast('Inserted current epoch into input', 'success');
    }
}

// Client-Side Execution Logic for Security (Zero Data Exits Browser)

function runBase64(operation) {
    const input = document.getElementById('base64-input').value;
    const outputDiv = document.getElementById('base64-output');
    if (!outputDiv) return;

    if (!input || input.trim() === '') {
        outputDiv.innerHTML = `
            <div class="card glass mt-4 animate-fade-in">
                <div class="card-body">
                    <div class="alert alert-error">
                        <i class="fa-solid fa-circle-exclamation"></i>
                        <div>Input data is empty.</div>
                    </div>
                </div>
            </div>
        `;
        return;
    }

    saveToHistory(`Base64 ${operation.charAt(0).toUpperCase() + operation.slice(1)}`, input);

    let result = '';
    let error = '';

    try {
        if (operation === 'encode') {
            const utf8Bytes = new TextEncoder().encode(input);
            let binary = '';
            for (let i = 0; i < utf8Bytes.length; i++) {
                binary += String.fromCharCode(utf8Bytes[i]);
            }
            result = btoa(binary);
        } else {
            let cleaned = input.trim().replace(/-/g, '+').replace(/_/g, '/');
            while (cleaned.length % 4) {
                cleaned += '=';
            }
            const binary = atob(cleaned);
            const bytes = new Uint8Array(binary.length);
            for (let i = 0; i < binary.length; i++) {
                bytes[i] = binary.charCodeAt(i);
            }
            result = new TextDecoder().decode(bytes);
        }
    } catch (e) {
        error = operation === 'encode' ? 'Encoding failed.' : 'Invalid Base64 string. Please verify the input.';
    }

    if (error) {
        outputDiv.innerHTML = `
            <div class="card glass mt-4 animate-fade-in">
                <div class="card-header flex justify-between items-center">
                    <h4>Conversion Result (${operation === 'encode' ? 'Encode' : 'Decode'})</h4>
                </div>
                <div class="card-body">
                    <div class="alert alert-error">
                        <i class="fa-solid fa-circle-exclamation"></i>
                        <div>${error}</div>
                    </div>
                </div>
            </div>
        `;
    } else {
        outputDiv.innerHTML = `
            <div class="card glass mt-4 animate-fade-in">
                <div class="card-header flex justify-between items-center">
                    <h4>Conversion Result (${operation === 'encode' ? 'Encode' : 'Decode'})</h4>
                    <div class="card-actions">
                        <button class="btn btn-sm btn-primary" onclick="copyToClipboard('base64-result-text')">
                            <i class="fa-solid fa-copy"></i> Copy Output
                        </button>
                    </div>
                </div>
                <div class="card-body">
                    <div class="result-box">
                        <pre id="base64-result-text" class="result-text">${escapeHTML(result)}</pre>
                    </div>
                </div>
            </div>
        `;
    }
}

function cleanLooseJSON(input) {
    let cleaned = input;
    
    // 1. Strip block comments
    cleaned = cleaned.replace(/\/\*[\s\S]*?\*\//g, '');
    
    // 2. Strip single line comments (ignore :// in URLs)
    cleaned = cleaned.replace(/(?:^|[^:])\/\/.*$/gm, function(match) {
        if (match.trim().startsWith('://')) {
            return match;
        }
        return '';
    });

    // 3. Convert single quotes to double quotes
    cleaned = cleaned.replace(/'([^'\\]*(?:\\.[^'\\]*)*)'/g, '"$1"');

    // 4. Wrap unquoted keys
    cleaned = cleaned.replace(/([{,]\s*)([a-zA-Z_$][a-zA-Z0-9_$-]*)\s*:/g, '$1"$2":');

    // 5. Strip trailing commas
    cleaned = cleaned.replace(/,\s*([}\]])/g, '$1');

    return cleaned;
}

function runJSON(operation) {
    const input = document.getElementById('json-input').value.trim();
    const outputDiv = document.getElementById('json-output');
    const indentVal = document.getElementById('json-indent').value;
    if (!outputDiv) return;

    if (input === '') {
        outputDiv.innerHTML = `
            <div class="card glass mt-4 animate-fade-in">
                <div class="card-body">
                    <div class="alert alert-error">
                        <i class="fa-solid fa-circle-exclamation"></i>
                        <div>JSON input is empty.</div>
                    </div>
                </div>
            </div>
        `;
        return;
    }

    saveToHistory(`JSON ${operation.charAt(0).toUpperCase() + operation.slice(1)}`, truncateString(input, 30));

    let parsed = null;
    let error = '';
    let isLoose = false;

    try {
        parsed = JSON.parse(input);
    } catch (strictErr) {
        try {
            const cleaned = cleanLooseJSON(input);
            parsed = JSON.parse(cleaned);
            isLoose = true;
        } catch (looseErr) {
            error = strictErr.message;
        }
    }

    if (error) {
        outputDiv.innerHTML = `
            <div class="card glass mt-4 animate-fade-in">
                <div class="card-header flex justify-between items-center">
                    <h4>JSON Result (${operation.charAt(0).toUpperCase() + operation.slice(1)})</h4>
                </div>
                <div class="card-body">
                    <div class="alert alert-error">
                        <i class="fa-solid fa-circle-exclamation"></i>
                        <div>Invalid JSON: ${escapeHTML(error)}</div>
                    </div>
                </div>
            </div>
        `;
        return;
    }

    if (operation === 'validate') {
        const msg = isLoose 
            ? "Valid JavaScript Object / Loose JSON (cleaned successfully)!" 
            : "JSON is valid and well-formed!";
        outputDiv.innerHTML = `
            <div class="card glass mt-4 animate-fade-in">
                <div class="card-header flex justify-between items-center">
                    <h4>JSON Result (Validate)</h4>
                </div>
                <div class="card-body">
                    <div class="alert alert-success" style="background: rgba(16, 185, 129, 0.1); border: 1px solid rgba(16, 185, 129, 0.2); padding: 12px; border-radius: var(--radius-md); display: flex; align-items: center; gap: 10px; color: var(--success);">
                        <i class="fa-solid fa-circle-check"></i>
                        <div>${escapeHTML(msg)}</div>
                    </div>
                </div>
            </div>
        `;
    } else {
        let result = '';
        if (operation === 'prettify') {
            let space = 2;
            if (indentVal === '4') space = 4;
            else if (indentVal === 'tab') space = '\t';
            result = JSON.stringify(parsed, null, space);
        } else {
            result = JSON.stringify(parsed);
        }

        outputDiv.innerHTML = `
            <div class="card glass mt-4 animate-fade-in">
                <div class="card-header flex justify-between items-center">
                    <h4>JSON Result (${operation.charAt(0).toUpperCase() + operation.slice(1)})</h4>
                    <div class="card-actions">
                        <button class="btn btn-sm btn-primary" onclick="copyToClipboard('json-result-text')">
                            <i class="fa-solid fa-copy"></i> Copy Output
                        </button>
                    </div>
                </div>
                <div class="card-body">
                    <div class="result-box">
                        <pre id="json-result-text" class="result-text">${escapeHTML(result)}</pre>
                    </div>
                </div>
            </div>
        `;
    }
}

function runURL(operation) {
    const input = document.getElementById('url-input').value.trim();
    const outputDiv = document.getElementById('url-output');
    if (!outputDiv) return;

    if (input === '') {
        outputDiv.innerHTML = `
            <div class="card glass mt-4 animate-fade-in">
                <div class="card-body">
                    <div class="alert alert-error">
                        <i class="fa-solid fa-circle-exclamation"></i>
                        <div>URL input is empty.</div>
                    </div>
                </div>
            </div>
        `;
        return;
    }

    saveToHistory(`URL ${operation.charAt(0).toUpperCase() + operation.slice(1)}`, truncateString(input, 30));

    if (operation === 'encode') {
        const result = encodeURIComponent(input);
        outputDiv.innerHTML = `
            <div class="card glass mt-4 animate-fade-in">
                <div class="card-header flex justify-between items-center">
                    <h4>URL Result (Encode)</h4>
                    <div class="card-actions">
                        <button class="btn btn-sm btn-primary" onclick="copyToClipboard('url-result-text')">
                            <i class="fa-solid fa-copy"></i> Copy Output
                        </button>
                    </div>
                </div>
                <div class="card-body">
                    <div class="result-box">
                        <pre id="url-result-text" class="result-text">${escapeHTML(result)}</pre>
                    </div>
                </div>
            </div>
        `;
    } else if (operation === 'decode') {
        try {
            const result = decodeURIComponent(input);
            outputDiv.innerHTML = `
                <div class="card glass mt-4 animate-fade-in">
                    <div class="card-header flex justify-between items-center">
                        <h4>URL Result (Decode)</h4>
                        <div class="card-actions">
                            <button class="btn btn-sm btn-primary" onclick="copyToClipboard('url-result-text')">
                                <i class="fa-solid fa-copy"></i> Copy Output
                            </button>
                        </div>
                    </div>
                    <div class="card-body">
                        <div class="result-box">
                            <pre id="url-result-text" class="result-text">${escapeHTML(result)}</pre>
                        </div>
                    </div>
                </div>
            `;
        } catch (e) {
            outputDiv.innerHTML = `
                <div class="card glass mt-4 animate-fade-in">
                    <div class="card-body">
                        <div class="alert alert-error">
                            <i class="fa-solid fa-circle-exclamation"></i>
                            <div>Decoding failed: ${escapeHTML(e.message)}</div>
                        </div>
                    </div>
                </div>
            `;
        }
    } else if (operation === 'parse') {
        let u = null;
        let isFullURL = true;
        let params = [];

        try {
            u = new URL(input);
        } catch (e) {
            isFullURL = false;
        }

        if (isFullURL) {
            u.searchParams.forEach((value, key) => {
                params.push({ key, value });
            });
        } else {
            try {
                const searchParams = new URLSearchParams(input);
                let count = 0;
                searchParams.forEach((value, key) => {
                    params.push({ key, value });
                    count++;
                });
                if (count === 0 || (count === 1 && params[0].key === input)) {
                    throw new Error("Not a query string");
                }
            } catch (errQ) {
                outputDiv.innerHTML = `
                    <div class="card glass mt-4 animate-fade-in">
                        <div class="card-body">
                            <div class="alert alert-error">
                                <i class="fa-solid fa-circle-exclamation"></i>
                                <div>Failed to parse URL/Query string. Make sure it is a valid absolute URL or a key=value query string.</div>
                            </div>
                        </div>
                    </div>
                `;
                return;
            }
        }

        const scheme = isFullURL ? u.protocol.replace(':', '') : '';
        const host = isFullURL ? u.host : '';
        const path = isFullURL ? u.pathname : '';
        const opName = isFullURL ? 'Parse' : 'ParseQueryString';

        let partsHTML = '';
        if (isFullURL) {
            partsHTML = `
                <div class="grid grid-cols-3 gap-2 mb-4" style="display: grid; grid-template-columns: repeat(3, 1fr); gap: 12px; margin-bottom: 16px;">
                    <div class="info-block" style="background: rgba(0, 0, 0, 0.2); border: 1px solid var(--card-border); padding: 8px 12px; border-radius: var(--radius-sm);">
                        <span class="label" style="display: block; font-size: 10px; color: var(--text-muted); text-transform: uppercase;">Scheme</span>
                        <span class="value" style="font-family: var(--font-mono); font-size: 13px;">${escapeHTML(scheme || 'N/A')}</span>
                    </div>
                    <div class="info-block" style="background: rgba(0, 0, 0, 0.2); border: 1px solid var(--card-border); padding: 8px 12px; border-radius: var(--radius-sm);">
                        <span class="label" style="display: block; font-size: 10px; color: var(--text-muted); text-transform: uppercase;">Host</span>
                        <span class="value" style="font-family: var(--font-mono); font-size: 13px;">${escapeHTML(host || 'N/A')}</span>
                    </div>
                    <div class="info-block" style="background: rgba(0, 0, 0, 0.2); border: 1px solid var(--card-border); padding: 8px 12px; border-radius: var(--radius-sm);">
                        <span class="label" style="display: block; font-size: 10px; color: var(--text-muted); text-transform: uppercase;">Path</span>
                        <span class="value" style="font-family: var(--font-mono); font-size: 13px;">${escapeHTML(path || 'N/A')}</span>
                    </div>
                </div>
            `;
        }

        let tableRowsHTML = params.map((p, idx) => `
            <tr>
                <td><input type="text" class="table-input param-key" value="${escapeHTML(p.key)}" oninput="rebuildURL()"></td>
                <td><input type="text" class="table-input param-value" value="${escapeHTML(p.value)}" oninput="rebuildURL()"></td>
                <td>
                    <button class="btn-copy-inline" onclick="copyTableValue(this)" title="Copy Value">
                        <i class="fa-regular fa-copy"></i>
                    </button>
                </td>
            </tr>
        `).join('');

        let tableHTML = '';
        if (params.length === 0) {
            tableHTML = `<p class="text-muted mt-2" style="color: var(--text-muted); font-size: 13px;">No query parameters found.</p>`;
        } else {
            tableHTML = `
                <div class="table-container mt-2" data-scheme="${escapeHTML(scheme)}" data-host="${escapeHTML(host)}" data-path="${escapeHTML(path)}" data-op="${opName}">
                    <table class="params-table">
                        <thead>
                            <tr>
                                <th style="width: 35%;">Parameter Key</th>
                                <th style="width: 50%;">Value</th>
                                <th style="width: 15%;">Actions</th>
                            </tr>
                        </thead>
                        <tbody>
                            ${tableRowsHTML}
                        </tbody>
                    </table>
                </div>
                
                <div class="form-group mt-4">
                    <label style="display: block; margin-bottom: 6px;">Reconstructed URL / Query String</label>
                    <div class="reconstructed-wrapper">
                        <pre id="reconstructed-url" class="result-text"></pre>
                        <button class="btn btn-sm btn-secondary" onclick="copyToClipboard('reconstructed-url')">
                            <i class="fa-solid fa-copy"></i> Copy
                        </button>
                    </div>
                </div>
            `;
        }

        outputDiv.innerHTML = `
            <div class="card glass mt-4 animate-fade-in">
                <div class="card-header flex justify-between items-center">
                    <h4>URL Result (${isFullURL ? 'Parse' : 'ParseQueryString'})</h4>
                </div>
                <div class="card-body">
                    <div class="url-parts-info">
                        ${partsHTML}
                        <h5 style="margin-bottom: 8px; font-size: 14px;">Query Parameters (${params.length})</h5>
                        ${tableHTML}
                    </div>
                </div>
            </div>
        `;

        if (params.length > 0) {
            setTimeout(rebuildURL, 50);
        }
    }
}

function runJWT() {
    const input = document.getElementById('jwt-input').value.trim();
    const outputDiv = document.getElementById('jwt-output');
    if (!outputDiv) return;

    if (input === '') {
        outputDiv.innerHTML = `
            <div class="card glass mt-4 animate-fade-in">
                <div class="card-body">
                    <div class="alert alert-error">
                        <i class="fa-solid fa-circle-exclamation"></i>
                        <div>JWT token input is empty.</div>
                    </div>
                </div>
            </div>
        `;
        return;
    }

    saveToHistory('JWT Decode', truncateString(input, 20));

    const parts = input.split('.');
    if (parts.length !== 3) {
        outputDiv.innerHTML = `
            <div class="card glass mt-4 animate-fade-in">
                <div class="card-body">
                    <div class="alert alert-error">
                        <i class="fa-solid fa-circle-exclamation"></i>
                        <div>Invalid JWT format. A valid token must have exactly three segments separated by dots (header.payload.signature).</div>
                    </div>
                </div>
            </div>
        `;
        return;
    }

    let headerPretty = '';
    let payloadPretty = '';
    let payloadObj = null;

    try {
        const decodeSegment = (seg) => {
            let base64 = seg.replace(/-/g, '+').replace(/_/g, '/');
            while (base64.length % 4) {
                base64 += '=';
            }
            const raw = atob(base64);
            const bytes = new Uint8Array(raw.length);
            for (let i = 0; i < raw.length; i++) {
                bytes[i] = raw.charCodeAt(i);
            }
            return new TextDecoder().decode(bytes);
        };

        const headerDec = decodeSegment(parts[0]);
        const payloadDec = decodeSegment(parts[1]);

        try {
            headerPretty = JSON.stringify(JSON.parse(headerDec), null, 2);
        } catch (e) {
            headerPretty = headerDec;
        }

        try {
            payloadObj = JSON.parse(payloadDec);
            payloadPretty = JSON.stringify(payloadObj, null, 2);
        } catch (e) {
            payloadPretty = payloadDec;
        }

    } catch (e) {
        outputDiv.innerHTML = `
            <div class="card glass mt-4 animate-fade-in">
                <div class="card-body">
                    <div class="alert alert-error">
                        <i class="fa-solid fa-circle-exclamation"></i>
                        <div>Failed to decode token segments: ${escapeHTML(e.message)}</div>
                    </div>
                </div>
            </div>
        `;
        return;
    }

    let hasExp = false;
    let isExpired = false;
    let expTimeStr = '';
    let expInStr = '';
    let issuedAtStr = '';

    if (payloadObj) {
        const nowSec = Math.floor(Date.now() / 1000);

        if (payloadObj.exp !== undefined) {
            const exp = parseInt(payloadObj.exp);
            if (!isNaN(exp)) {
                hasExp = true;
                const expDate = new Date(exp * 1000);
                expTimeStr = expDate.toISOString().replace('T', ' ').substring(0, 19) + ' UTC';

                if (nowSec > exp) {
                    isExpired = true;
                    const diffSec = nowSec - exp;
                    expInStr = `Expired ${formatDuration(diffSec)} ago`;
                } else {
                    isExpired = false;
                    const diffSec = exp - nowSec;
                    expInStr = `Expires in ${formatDuration(diffSec)}`;
                }
            }
        }

        if (payloadObj.iat !== undefined) {
            const iat = parseInt(payloadObj.iat);
            if (!isNaN(iat)) {
                const iatDate = new Date(iat * 1000);
                issuedAtStr = iatDate.toISOString().replace('T', ' ').substring(0, 19) + ' UTC';
            }
        }
    }

    let statusBadgeHTML = '';
    if (hasExp) {
        if (isExpired) {
            statusBadgeHTML = `<span class="badge badge-error"><i class="fa-solid fa-circle-xmark"></i> Token Expired (${escapeHTML(expInStr)})</span>`;
        } else {
            statusBadgeHTML = `<span class="badge badge-success"><i class="fa-solid fa-circle-check"></i> Token Active (${escapeHTML(expInStr)})</span>`;
        }
    } else {
        statusBadgeHTML = `<span class="badge badge-warning"><i class="fa-solid fa-triangle-exclamation"></i> No Expiration Claim (exp)</span>`;
    }

    let iatBadgeHTML = issuedAtStr ? `<span class="badge badge-info"><i class="fa-solid fa-calendar-days"></i> Issued: ${escapeHTML(issuedAtStr)}</span>` : '';
    let expBadgeHTML = expTimeStr ? `<span class="badge badge-secondary"><i class="fa-solid fa-clock"></i> Exp: ${escapeHTML(expTimeStr)}</span>` : '';

    outputDiv.innerHTML = `
        <div class="card glass mt-4 animate-fade-in">
            <div class="card-header">
                <h4>Decoded JSON Web Token</h4>
            </div>
            <div class="card-body">
                <div class="jwt-status-badges mb-4 flex gap-2 flex-wrap" style="display: flex; gap: 8px; flex-wrap: wrap; margin-bottom: 16px;">
                    ${statusBadgeHTML}
                    ${iatBadgeHTML}
                    ${expBadgeHTML}
                </div>

                <div class="jwt-segments-grid">
                    <div class="jwt-segment">
                        <div class="segment-header text-red">
                            <h5 style="font-size: 13px; font-weight: 700;">HEADER <span class="text-xs text-muted" style="font-size: 11px; color: var(--text-muted); font-weight: 400;">(ALGORITHM & TOKEN TYPE)</span></h5>
                            <button class="btn-copy-inline" onclick="copyToClipboard('jwt-header-text')" title="Copy Header">
                                <i class="fa-regular fa-copy"></i>
                            </button>
                        </div>
                        <div class="result-box result-box-red">
                            <pre id="jwt-header-text" class="result-text">${escapeHTML(headerPretty)}</pre>
                        </div>
                    </div>

                    <div class="jwt-segment mt-4">
                        <div class="segment-header text-cyan">
                            <h5 style="font-size: 13px; font-weight: 700;">PAYLOAD <span class="text-xs text-muted" style="font-size: 11px; color: var(--text-muted); font-weight: 400;">(DATA / CLAIMS)</span></h5>
                            <button class="btn-copy-inline" onclick="copyToClipboard('jwt-payload-text')" title="Copy Payload">
                                <i class="fa-regular fa-copy"></i>
                            </button>
                        </div>
                        <div class="result-box result-box-cyan">
                            <pre id="jwt-payload-text" class="result-text">${escapeHTML(payloadPretty)}</pre>
                        </div>
                    </div>

                    <div class="jwt-segment mt-4">
                        <div class="segment-header text-green">
                            <h5 style="font-size: 13px; font-weight: 700;">SIGNATURE <span class="text-xs text-muted" style="font-size: 11px; color: var(--text-muted); font-weight: 400;">(HMAC/RSA SHA-256)</span></h5>
                            <button class="btn-copy-inline" onclick="copyToClipboard('jwt-signature-text')" title="Copy Signature">
                                <i class="fa-regular fa-copy"></i>
                            </button>
                        </div>
                        <div class="result-box result-box-green">
                            <pre id="jwt-signature-text" class="result-text" style="word-break: break-all; white-space: pre-wrap;">${escapeHTML(parts[2])}</pre>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    `;
}

function formatDuration(sec) {
    if (sec < 60) return `${sec}s`;
    const min = Math.floor(sec / 60);
    if (min < 60) return `${min}m`;
    const hr = Math.floor(min / 60);
    if (hr < 24) return `${hr}h`;
    const day = Math.floor(hr / 24);
    return `${day}d`;
}

function runEpoch() {
    let input = document.getElementById('epoch-input').value.trim();
    const outputDiv = document.getElementById('epoch-output');
    if (!outputDiv) return;

    if (input === '') {
        input = Math.floor(Date.now() / 1000).toString();
    }

    saveToHistory('Epoch Convert', input);

    let targetTime = null;

    if (/^\d+$/.test(input)) {
        const num = parseInt(input);
        if (input.length > 11) {
            targetTime = new Date(num);
        } else {
            targetTime = new Date(num * 1000);
        }
    } else {
        const parsed = Date.parse(input);
        if (!isNaN(parsed)) {
            targetTime = new Date(parsed);
        }
    }

    if (!targetTime || isNaN(targetTime.getTime())) {
        outputDiv.innerHTML = `
            <div class="card glass mt-4 animate-fade-in">
                <div class="card-body">
                    <div class="alert alert-error">
                        <i class="fa-solid fa-circle-exclamation"></i>
                        <div>Could not parse date/time value. Support Unix epoch integers (seconds or milliseconds) and standard ISO8601 date strings.</div>
                    </div>
                </div>
            </div>
        `;
        return;
    }

    const seconds = Math.floor(targetTime.getTime() / 1000);
    const millis = targetTime.getTime();
    const utcStr = targetTime.toUTCString();
    
    const localOffset = -targetTime.getTimezoneOffset();
    const diffSign = localOffset >= 0 ? '+' : '-';
    const pad = (n) => (n < 10 ? '0' + n : n);
    const absOffset = Math.abs(localOffset);
    const offsetHours = pad(Math.floor(absOffset / 60));
    const offsetMins = pad(absOffset % 60);
    const offsetStr = `${diffSign}${offsetHours}${offsetMins}`;

    const localStr = `${targetTime.getFullYear()}-${pad(targetTime.getMonth()+1)}-${pad(targetTime.getDate())} ` +
                     `${pad(targetTime.getHours())}:${pad(targetTime.getMinutes())}:${pad(targetTime.getSeconds())} ` +
                     `${offsetStr}`;

    let relative = '';
    const now = Date.now();
    const diffMs = now - targetTime.getTime();
    const absDiffSec = Math.floor(Math.abs(diffMs) / 1000);

    const getRelativeStr = (sec) => {
        if (sec < 60) return `${sec} seconds`;
        const min = Math.floor(sec / 60);
        if (min < 60) return `${min} minutes`;
        const hr = Math.floor(min / 60);
        if (hr < 24) return `${hr} hours`;
        const day = Math.floor(hr / 24);
        return `${day} days`;
    };

    if (diffMs >= 0) {
        relative = `${getRelativeStr(absDiffSec)} ago`;
    } else {
        relative = `in ${getRelativeStr(absDiffSec)}`;
    }

    outputDiv.innerHTML = `
        <div class="card glass mt-4 animate-fade-in">
            <div class="card-header">
                <h4>Conversion Results</h4>
            </div>
            <div class="card-body">
                <div class="grid grid-cols-2 gap-4" style="display: grid; grid-template-columns: repeat(2, 1fr); gap: 16px;">
                    <div class="form-group">
                        <label>Epoch Seconds (10-digit)</label>
                        <div class="flex gap-2" style="display: flex; gap: 8px;">
                            <span id="epoch-sec-val" class="result-text-inline font-mono">${seconds}</span>
                            <button class="btn btn-sm btn-secondary" onclick="copyToClipboard('epoch-sec-val')">
                                <i class="fa-solid fa-copy"></i>
                            </button>
                        </div>
                    </div>

                    <div class="form-group">
                        <label>Epoch Milliseconds (13-digit)</label>
                        <div class="flex gap-2" style="display: flex; gap: 8px;">
                            <span id="epoch-ms-val" class="result-text-inline font-mono">${millis}</span>
                            <button class="btn btn-sm btn-secondary" onclick="copyToClipboard('epoch-ms-val')">
                                <i class="fa-solid fa-copy"></i>
                            </button>
                        </div>
                    </div>
                </div>

                <hr class="divider">

                <div class="form-group">
                    <label>GMT / UTC Time</label>
                    <div class="flex gap-2" style="display: flex; gap: 8px;">
                        <span id="epoch-utc-val" class="result-text-inline font-mono">${escapeHTML(utcStr)}</span>
                        <button class="btn btn-sm btn-secondary" onclick="copyToClipboard('epoch-utc-val')">
                            <i class="fa-solid fa-copy"></i>
                        </button>
                    </div>
                </div>

                <div class="form-group mt-3" style="margin-top: 12px;">
                    <label>Local Time (Your system timezone)</label>
                    <div class="flex gap-2" style="display: flex; gap: 8px;">
                        <span id="epoch-local-val" class="result-text-inline font-mono">${escapeHTML(localStr)}</span>
                        <button class="btn btn-sm btn-secondary" onclick="copyToClipboard('epoch-local-val')">
                            <i class="fa-solid fa-copy"></i>
                        </button>
                    </div>
                </div>

                <div class="form-group mt-3" style="margin-top: 12px;">
                    <label style="margin-bottom: 4px;">Relative Time</label>
                    <div>
                        <span class="badge badge-accent font-semibold" style="font-weight: 600;">${escapeHTML(relative)}</span>
                    </div>
                </div>
            </div>
        </div>
    `;
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

    const history = JSON.parse(safeStorage.getItem('devutils_history') || '[]');
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

    safeStorage.setItem('devutils_history', JSON.stringify(history));
}

// Render History Panel items
function renderHistory() {
    const list = document.getElementById('history-list');
    if (!list) return;

    const history = JSON.parse(safeStorage.getItem('devutils_history') || '[]');

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
    safeStorage.removeItem('devutils_history');
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
    } else if (operation.includes('JSON')) {
        switchTab('json');
        document.getElementById('json-input').value = details;
        showToast('Restored payload to JSON panel', 'success');
    } else if (operation.includes('URL')) {
        switchTab('url');
        document.getElementById('url-input').value = details;
        showToast('Restored payload to URL panel', 'success');
    } else if (operation.includes('JWT')) {
        switchTab('jwt');
        document.getElementById('jwt-input').value = details;
        showToast('Restored token to JWT panel', 'success');
    } else if (operation.includes('Epoch')) {
        switchTab('epoch');
        document.getElementById('epoch-input').value = details;
        showToast('Restored timestamp to Epoch panel', 'success');
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
