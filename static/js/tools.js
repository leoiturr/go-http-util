// Client-side tools: Base64, JSON, URL, HTML entities, Text Diff, JWT, Epoch.
// Repeated card shapes come from <template> elements cloned with cloneNode.
import { escapeHTML, truncateString, pad, fromTemplate } from './core/utils.js';
import { showToast } from './core/ui.js';
import { saveToHistory } from './core/history.js';
import { getEditor, initResultEditors } from './core/editors.js';

/* ------------------------------------------------------------------ */
/* Template helpers                                                     */
/* ------------------------------------------------------------------ */

function placeholderCard(icon, message) {
    const node = fromTemplate('tpl-placeholder');
    node.querySelector('.placeholder-icon').classList.add(icon);
    node.querySelector('p').textContent = message;
    return node;
}

function errorCard(message) {
    const node = fromTemplate('tpl-error-card');
    node.querySelector('.error-message').textContent = message;
    return node;
}

function copyButton(target, label = 'Copy Output') {
    const btn = document.createElement('copy-button');
    btn.className = 'btn btn-sm btn-primary';
    btn.setAttribute('target', target);
    btn.textContent = label;
    return btn;
}

function resultCard(title, { id, copyLabel } = {}) {
    const node = fromTemplate('tpl-result-card');
    node.querySelector('.result-title').textContent = title;
    if (id) node.querySelector('.card-actions').appendChild(copyButton(id, copyLabel));
    return node;
}

function showPreResult(container, title, text, id, copyLabel) {
    const card = resultCard(title, { id, copyLabel });
    const box = document.createElement('div');
    box.className = 'result-box';
    const pre = document.createElement('pre');
    pre.id = id;
    pre.className = 'result-text';
    pre.textContent = text;
    box.appendChild(pre);
    card.querySelector('.card-body').appendChild(box);
    container.replaceChildren(card);
}

function showEditorResult(container, title, text, id, mode) {
    const card = resultCard(title, { id });
    const box = document.createElement('div');
    box.className = 'result-box';
    const ta = document.createElement('textarea');
    ta.id = id;
    ta.className = 'result-text';
    ta.setAttribute('data-mode', mode);
    ta.readOnly = true;
    ta.value = text;
    box.appendChild(ta);
    card.querySelector('.card-body').appendChild(box);
    container.replaceChildren(card);
    initResultEditors(container);
}

/* ------------------------------------------------------------------ */
/* Base64                                                               */
/* ------------------------------------------------------------------ */

export function runBase64(operation) {
    const input = document.getElementById('base64-input').value;
    const outputDiv = document.getElementById('base64-output');
    if (!outputDiv) return;

    if (!input || input.trim() === '') {
        outputDiv.replaceChildren(errorCard('Input data is empty.'));
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
            while (cleaned.length % 4) cleaned += '=';
            const binary = atob(cleaned);
            const bytes = new Uint8Array(binary.length);
            for (let i = 0; i < binary.length; i++) bytes[i] = binary.charCodeAt(i);
            result = new TextDecoder().decode(bytes);
        }
    } catch (e) {
        error = operation === 'encode'
            ? 'Encoding failed.'
            : 'Invalid Base64 string. Please verify the input.';
    }

    if (error) {
        outputDiv.replaceChildren(errorCard(error));
        return;
    }
    showPreResult(
        outputDiv,
        `Conversion Result (${operation === 'encode' ? 'Encode' : 'Decode'})`,
        result,
        'base64-result-text'
    );
}

export function clearBase64() {
    document.getElementById('base64-input').value = '';
    document.getElementById('base64-output').replaceChildren();
    showToast('Base64 inputs cleared', 'success');
}

/* ------------------------------------------------------------------ */
/* JSON                                                                 */
/* ------------------------------------------------------------------ */

export function cleanLooseJSON(input) {
    let cleaned = input;
    cleaned = cleaned.replace(/\/\*[\s\S]*?\*\//g, '');
    cleaned = cleaned.replace(/(?:^|[^:])\/\/.*$/gm, (match) =>
        match.trim().startsWith('://') ? match : '');
    cleaned = cleaned.replace(/'([^'\\]*(?:\\.[^'\\]*)*)'/g, '"$1"');
    cleaned = cleaned.replace(/([{,]\s*)([a-zA-Z_$][a-zA-Z0-9_$-]*)\s*:/g, '$1"$2":');
    cleaned = cleaned.replace(/,\s*([}\]])/g, '$1');
    return cleaned;
}

export function runJSON(operation) {
    const input = document.getElementById('json-input').value.trim();
    const outputDiv = document.getElementById('json-output');
    const indentVal = document.getElementById('json-indent').value;
    if (!outputDiv) return;

    if (input === '') {
        outputDiv.replaceChildren(errorCard('JSON input is empty.'));
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
            parsed = JSON.parse(cleanLooseJSON(input));
            isLoose = true;
        } catch (looseErr) {
            error = strictErr.message;
        }
    }

    if (error) {
        outputDiv.replaceChildren(errorCard(`Invalid JSON: ${error}`));
        return;
    }

    if (operation === 'validate') {
        const card = resultCard('JSON Result (Validate)');
        const alert = document.createElement('div');
        alert.className = 'alert alert-success';
        alert.innerHTML = '<i class="bi bi-check-circle-fill"></i><div></div>';
        alert.querySelector('div').textContent = isLoose
            ? 'Valid JavaScript Object / Loose JSON (cleaned successfully)!'
            : 'JSON is valid and well-formed!';
        card.querySelector('.card-body').appendChild(alert);
        outputDiv.replaceChildren(card);
        return;
    }

    let result = '';
    if (operation === 'prettify') {
        let space = 2;
        if (indentVal === '4') space = 4;
        else if (indentVal === 'tab') space = '\t';
        result = JSON.stringify(parsed, null, space);
    } else {
        result = JSON.stringify(parsed);
    }

    showEditorResult(
        outputDiv,
        `JSON Result (${operation.charAt(0).toUpperCase() + operation.slice(1)})`,
        result,
        'json-result-text',
        'application/json'
    );
}

export function clearJSON() {
    document.getElementById('json-input').value = '';
    const editor = getEditor('json');
    if (editor) editor.setValue('');
    document.getElementById('json-output').replaceChildren();
    showToast('JSON inputs cleared', 'success');
}

/* ------------------------------------------------------------------ */
/* URL                                                                  */
/* ------------------------------------------------------------------ */

export function runURL(operation) {
    const input = document.getElementById('url-input').value.trim();
    const outputDiv = document.getElementById('url-output');
    if (!outputDiv) return;

    if (input === '') {
        outputDiv.replaceChildren(errorCard('URL input is empty.'));
        return;
    }

    saveToHistory(`URL ${operation.charAt(0).toUpperCase() + operation.slice(1)}`, truncateString(input, 30));

    if (operation === 'encode') {
        showPreResult(outputDiv, 'URL Result (Encode)', encodeURIComponent(input), 'url-result-text');
        return;
    }

    if (operation === 'decode') {
        try {
            showPreResult(outputDiv, 'URL Result (Decode)', decodeURIComponent(input), 'url-result-text');
        } catch (e) {
            outputDiv.replaceChildren(errorCard(`Decoding failed: ${e.message}`));
        }
        return;
    }

    if (operation === 'parse') {
        runURLParse(input, outputDiv);
    }
}

function runURLParse(input, outputDiv) {
    let u = null;
    let isFullURL = true;
    const params = [];

    try {
        u = new URL(input);
    } catch (e) {
        isFullURL = false;
    }

    if (isFullURL) {
        u.searchParams.forEach((value, key) => params.push({ key, value }));
    } else {
        const searchParams = new URLSearchParams(input);
        let count = 0;
        searchParams.forEach((value, key) => { params.push({ key, value }); count++; });
        if (count === 0 || (count === 1 && params[0].key === input)) {
            outputDiv.replaceChildren(errorCard('Failed to parse URL/Query string. Make sure it is a valid absolute URL or a key=value query string.'));
            return;
        }
    }

    const scheme = isFullURL ? u.protocol.replace(':', '') : '';
    const host = isFullURL ? u.host : '';
    const path = isFullURL ? u.pathname : '';
    const opName = isFullURL ? 'Parse' : 'ParseQueryString';

    const card = resultCard(`URL Result (${isFullURL ? 'Parse' : 'ParseQueryString'})`);
    const body = card.querySelector('.card-body');
    const info = document.createElement('div');
    info.className = 'url-parts-info';

    if (isFullURL) {
        const grid = document.createElement('div');
        grid.className = 'grid grid-cols-3 gap-2 mb-4';
        grid.style.cssText = 'display:grid;grid-template-columns:repeat(3,1fr);gap:12px;margin-bottom:16px;';
        [['Scheme', scheme], ['Host', host], ['Path', path]].forEach(([label, value]) => {
            const block = document.createElement('div');
            block.className = 'info-block';
            const l = document.createElement('span'); l.className = 'label'; l.textContent = label;
            const v = document.createElement('span'); v.className = 'value'; v.textContent = value || 'N/A';
            block.append(l, v);
            grid.appendChild(block);
        });
        info.appendChild(grid);
    }

    const heading = document.createElement('h5');
    heading.style.cssText = 'margin-bottom:8px;font-size:14px;';
    heading.textContent = `Query Parameters (${params.length})`;
    info.appendChild(heading);

    if (params.length === 0) {
        const p = document.createElement('p');
        p.className = 'text-muted mt-2';
        p.textContent = 'No query parameters found.';
        info.appendChild(p);
    } else {
        const tableWrap = document.createElement('div');
        tableWrap.className = 'table-container mt-2';
        tableWrap.dataset.scheme = scheme;
        tableWrap.dataset.host = host;
        tableWrap.dataset.path = path;
        tableWrap.dataset.op = opName;

        const table = document.createElement('table');
        table.className = 'params-table';
        table.innerHTML = `<thead><tr><th style="width:35%;">Parameter Key</th><th style="width:50%;">Value</th><th style="width:15%;">Actions</th></tr></thead>`;
        const tbody = document.createElement('tbody');
        params.forEach((p) => {
            const tr = document.createElement('tr');
            const keyCell = document.createElement('td');
            const keyInput = document.createElement('input');
            keyInput.type = 'text';
            keyInput.className = 'table-input param-key';
            keyInput.value = p.key;
            keyInput.dataset.action = 'url-rebuild';
            keyCell.appendChild(keyInput);

            const valCell = document.createElement('td');
            const valInput = document.createElement('input');
            valInput.type = 'text';
            valInput.className = 'table-input param-value';
            valInput.value = p.value;
            valInput.dataset.action = 'url-rebuild';
            valCell.appendChild(valInput);

            const actionCell = document.createElement('td');
            const copyBtn = document.createElement('copy-button');
            copyBtn.className = 'btn-copy-inline';
            copyBtn.setAttribute('within', '.param-value');
            copyBtn.setAttribute('title', 'Copy Value');
            actionCell.appendChild(copyBtn);

            tr.append(keyCell, valCell, actionCell);
            tbody.appendChild(tr);
        });
        table.appendChild(tbody);
        tableWrap.appendChild(table);
        info.appendChild(tableWrap);

        const group = document.createElement('div');
        group.className = 'form-group mt-4';
        const label = document.createElement('label');
        label.style.cssText = 'display:block;margin-bottom:6px;';
        label.textContent = 'Reconstructed URL / Query String';
        const wrapper = document.createElement('div');
        wrapper.className = 'reconstructed-wrapper';
        const pre = document.createElement('pre');
        pre.id = 'reconstructed-url';
        pre.className = 'result-text';
        const copy = document.createElement('copy-button');
        copy.className = 'btn btn-sm btn-secondary';
        copy.setAttribute('target', 'reconstructed-url');
        copy.textContent = 'Copy';
        wrapper.append(pre, copy);
        group.append(label, wrapper);
        info.appendChild(group);
    }

    body.appendChild(info);
    outputDiv.replaceChildren(card);
    if (params.length > 0) rebuildURL();
}

export function rebuildURL() {
    const container = document.querySelector('.table-container');
    if (!container) return;

    const op = container.getAttribute('data-op');
    const scheme = container.getAttribute('data-scheme') || '';
    const host = container.getAttribute('data-host') || '';
    const path = container.getAttribute('data-path') || '';

    const params = [];
    document.querySelectorAll('.params-table tbody tr').forEach((row) => {
        const keyInput = row.querySelector('.param-key');
        const valInput = row.querySelector('.param-value');
        if (keyInput && valInput) {
            const key = keyInput.value.trim();
            if (key !== '') params.push(encodeURIComponent(key) + '=' + encodeURIComponent(valInput.value));
        }
    });

    const queryString = params.join('&');
    let finalURL = '';
    if (op === 'Parse' && (scheme || host || path)) {
        finalURL = (scheme ? scheme + '://' : '') + host + path;
        if (queryString) finalURL += '?' + queryString;
    } else {
        finalURL = queryString;
    }

    const outputEl = document.getElementById('reconstructed-url');
    if (outputEl) outputEl.textContent = finalURL;
}

export function clearURL() {
    document.getElementById('url-input').value = '';
    document.getElementById('url-output').replaceChildren();
    showToast('URL inputs cleared', 'success');
}

/* ------------------------------------------------------------------ */
/* HTML entities                                                        */
/* ------------------------------------------------------------------ */

export function runHTMLEntity(operation) {
    const input = document.getElementById('html-entity-input').value;
    const outputDiv = document.getElementById('html-entity-output');
    if (!outputDiv) return;

    if (!input || input.trim() === '') {
        outputDiv.replaceChildren(errorCard('Input data is empty.'));
        return;
    }

    saveToHistory(`HTML Entity ${operation.charAt(0).toUpperCase() + operation.slice(1)}`, input);

    const textarea = document.createElement('textarea');
    let result = '';
    if (operation === 'encode') {
        textarea.innerHTML = input;
        result = textarea.value
            .replace(/&/g, '&amp;')
            .replace(/</g, '&lt;')
            .replace(/>/g, '&gt;')
            .replace(/"/g, '&quot;')
            .replace(/'/g, '&#39;');
    } else {
        textarea.innerHTML = input;
        result = textarea.value;
    }

    showPreResult(
        outputDiv,
        `HTML Entity Result (${operation.charAt(0).toUpperCase() + operation.slice(1)})`,
        result,
        'html-entity-result-text'
    );
}

export function clearHTMLEntity() {
    document.getElementById('html-entity-input').value = '';
    document.getElementById('html-entity-output').replaceChildren(
        placeholderCard('bi-code-square', 'Enter text and select an operation to view output.')
    );
    showToast('HTML Entity inputs cleared', 'success');
}

/* ------------------------------------------------------------------ */
/* Text Diff                                                            */
/* ------------------------------------------------------------------ */

function renderDiffLine(content) {
    return escapeHTML(content).replace(/ /g, '·').replace(/\t/g, '→');
}

export function computeDiff(textA, textB) {
    const linesA = textA.split('\n');
    const linesB = textB.split('\n');
    const m = linesA.length;
    const n = linesB.length;
    const dp = Array.from({ length: m + 1 }, () => new Array(n + 1).fill(0));

    for (let i = 1; i <= m; i++) {
        for (let j = 1; j <= n; j++) {
            dp[i][j] = linesA[i - 1] === linesB[j - 1]
                ? dp[i - 1][j - 1] + 1
                : Math.max(dp[i - 1][j], dp[i][j - 1]);
        }
    }

    const diffLines = [];
    let i = m, j = n;
    while (i > 0 || j > 0) {
        if (i > 0 && j > 0 && linesA[i - 1] === linesB[j - 1]) {
            diffLines.unshift({ type: 'same', content: linesA[i - 1] });
            i--; j--;
        } else if (j > 0 && (i === 0 || dp[i][j - 1] >= dp[i - 1][j])) {
            diffLines.unshift({ type: 'add', content: linesB[j - 1] });
            j--;
        } else {
            diffLines.unshift({ type: 'del', content: linesA[i - 1] });
            i--;
        }
    }
    return diffLines;
}

export function runTextDiff() {
    const textA = document.getElementById('diff-text-a').value;
    const textB = document.getElementById('diff-text-b').value;
    const outputDiv = document.getElementById('diff-output');
    if (!outputDiv) return;

    if (textA.trim() === '' && textB.trim() === '') {
        outputDiv.replaceChildren(errorCard('Both text fields are empty.'));
        return;
    }

    saveToHistory('Text Diff', truncateString(textA, 20) + ' vs ' + truncateString(textB, 20));

    const diffLines = computeDiff(textA, textB);
    let added = 0, removed = 0;
    diffLines.forEach((l) => {
        if (l.type === 'add') added++;
        else if (l.type === 'del') removed++;
    });

    const card = resultCard('Text Diff Result');
    const actions = card.querySelector('.card-actions');
    const badge = document.createElement('span');
    badge.className = 'badge badge-accent';
    badge.textContent = `+${added} / -${removed}`;
    actions.appendChild(badge);

    const body = card.querySelector('.card-body');
    body.style.display = 'flex';
    body.style.flexDirection = 'column';

    const stats = document.createElement('div');
    stats.className = 'diff-stats';
    stats.style.cssText = 'display:flex;gap:12px;margin-bottom:12px;';
    const addBadge = document.createElement('span');
    addBadge.className = 'badge badge-success';
    addBadge.textContent = `+${added} added`;
    const delBadge = document.createElement('span');
    delBadge.className = 'badge badge-error';
    delBadge.textContent = `-${removed} removed`;
    stats.append(addBadge, delBadge);

    const box = document.createElement('div');
    box.className = 'result-box';
    box.style.flex = '1';
    const out = document.createElement('div');
    out.className = 'diff-output';
    out.id = 'diff-result-text';
    diffLines.forEach((line) => {
        const prefix = line.type === 'add' ? '+' : line.type === 'del' ? '-' : ' ';
        const row = document.createElement('div');
        row.className = `diff-line ${line.type}`;
        // Preserve the exact characters while swapping in visible marks.
        row.innerHTML = prefix + renderDiffLine(line.content);
        out.appendChild(row);
    });
    box.appendChild(out);
    body.append(stats, box);

    outputDiv.replaceChildren(card);
}

export function clearDiff() {
    document.getElementById('diff-text-a').value = '';
    document.getElementById('diff-text-b').value = '';
    document.getElementById('diff-output').replaceChildren(
        placeholderCard('bi-file-diff', 'Enter two text blocks and click "Compare Text" to view differences.')
    );
    showToast('Text Diff inputs cleared', 'success');
}

/* ------------------------------------------------------------------ */
/* JWT                                                                  */
/* ------------------------------------------------------------------ */

export function syncJWTScroll() {
    const textarea = document.getElementById('jwt-input');
    const highlight = document.getElementById('jwt-input-highlight');
    if (!textarea || !highlight) return;
    highlight.scrollTop = textarea.scrollTop;
    highlight.scrollLeft = textarea.scrollLeft;
}

export function highlightJWTInput() {
    const textarea = document.getElementById('jwt-input');
    const highlight = document.getElementById('jwt-input-highlight');
    if (!textarea || !highlight) return;

    const parts = textarea.value.split('.');
    const classes = ['jwt-seg-header', 'jwt-seg-payload', 'jwt-seg-signature'];
    let html = '';
    parts.forEach((part, i) => {
        if (i > 0) html += '<span class="jwt-seg-dot">.</span>';
        if (part === '') return;
        html += `<span class="${classes[i] || 'jwt-seg-signature'}">${escapeHTML(part)}</span>`;
    });
    highlight.innerHTML = html;
    syncJWTScroll();
}

export function runJWT() {
    const input = document.getElementById('jwt-input').value.trim();
    const outputDiv = document.getElementById('jwt-output');
    if (!outputDiv) return;

    if (input === '') {
        outputDiv.replaceChildren(errorCard('JWT token input is empty.'));
        return;
    }

    saveToHistory('JWT Decode', truncateString(input, 20));

    const parts = input.split('.');
    if (parts.length !== 3) {
        outputDiv.replaceChildren(errorCard('Invalid JWT format. A valid token must have exactly three segments separated by dots (header.payload.signature).'));
        return;
    }

    const decodeSegment = (seg) => {
        let base64 = seg.replace(/-/g, '+').replace(/_/g, '/');
        while (base64.length % 4) base64 += '=';
        const raw = atob(base64);
        const bytes = new Uint8Array(raw.length);
        for (let i = 0; i < raw.length; i++) bytes[i] = raw.charCodeAt(i);
        return new TextDecoder().decode(bytes);
    };

    let headerPretty = '';
    let payloadPretty = '';
    let payloadObj = null;
    try {
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
        outputDiv.replaceChildren(errorCard(`Failed to decode token segments: ${e.message}`));
        return;
    }

    let hasExp = false, isExpired = false, expTimeStr = '', expInStr = '', issuedAtStr = '';
    if (payloadObj) {
        const nowSec = Math.floor(Date.now() / 1000);
        if (payloadObj.exp !== undefined) {
            const exp = parseInt(payloadObj.exp);
            if (!isNaN(exp)) {
                hasExp = true;
                expTimeStr = new Date(exp * 1000).toISOString().replace('T', ' ').substring(0, 19) + ' UTC';
                if (nowSec > exp) {
                    isExpired = true;
                    expInStr = `Expired ${formatRelativeSeconds(nowSec - exp)} ago`;
                } else {
                    expInStr = `Expires in ${formatRelativeSeconds(exp - nowSec)}`;
                }
            }
        }
        if (payloadObj.iat !== undefined) {
            const iat = parseInt(payloadObj.iat);
            if (!isNaN(iat)) {
                issuedAtStr = new Date(iat * 1000).toISOString().replace('T', ' ').substring(0, 19) + ' UTC';
            }
        }
    }

    // Intl.RelativeTimeFormat for human-friendly durations.
    let statusBadge = '';
    if (hasExp) {
        statusBadge = isExpired
            ? `<span class="badge badge-error"><i class="bi bi-x-circle-fill"></i> Token Expired (${escapeHTML(expInStr)})</span>`
            : `<span class="badge badge-success"><i class="bi bi-check-circle-fill"></i> Token Active (${escapeHTML(expInStr)})</span>`;
    } else {
        statusBadge = '<span class="badge badge-warning"><i class="bi bi-exclamation-triangle-fill"></i> No Expiration Claim (exp)</span>';
    }
    const iatBadge = issuedAtStr ? `<span class="badge badge-info"><i class="bi bi-calendar3"></i> Issued: ${escapeHTML(issuedAtStr)}</span>` : '';
    const expBadge = expTimeStr ? `<span class="badge badge-secondary"><i class="bi bi-clock"></i> Exp: ${escapeHTML(expTimeStr)}</span>` : '';

    const card = resultCard('Decoded JSON Web Token');
    const body = card.querySelector('.card-body');
    body.innerHTML = `
        <div class="jwt-status-badges mb-4" style="display:flex;gap:8px;flex-wrap:wrap;margin-bottom:16px;">
            ${statusBadge}
            ${iatBadge}
            ${expBadge}
        </div>
        <div class="jwt-segments-grid">
            <div class="jwt-segment">
                <div class="segment-header text-red">
                    <h5 style="font-size:13px;font-weight:700;">HEADER <span class="text-xs text-muted" style="font-size:11px;font-weight:400;">(ALGORITHM &amp; TOKEN TYPE)</span></h5>
                    <copy-button class="btn-copy-inline" target="jwt-header-text" title="Copy Header"></copy-button>
                </div>
                <div class="result-box result-box-red"><pre id="jwt-header-text" class="result-text">${escapeHTML(headerPretty)}</pre></div>
            </div>
            <div class="jwt-segment mt-4">
                <div class="segment-header text-cyan">
                    <h5 style="font-size:13px;font-weight:700;">PAYLOAD <span class="text-xs text-muted" style="font-size:11px;font-weight:400;">(DATA / CLAIMS)</span></h5>
                    <copy-button class="btn-copy-inline" target="jwt-payload-text" title="Copy Payload"></copy-button>
                </div>
                <div class="result-box result-box-cyan"><pre id="jwt-payload-text" class="result-text">${escapeHTML(payloadPretty)}</pre></div>
            </div>
            <div class="jwt-segment mt-4">
                <div class="segment-header text-green">
                    <h5 style="font-size:13px;font-weight:700;">SIGNATURE <span class="text-xs text-muted" style="font-size:11px;font-weight:400;">(HMAC/RSA SHA-256)</span></h5>
                    <copy-button class="btn-copy-inline" target="jwt-signature-text" title="Copy Signature"></copy-button>
                </div>
                <div class="result-box result-box-green"><pre id="jwt-signature-text" class="result-text" style="word-break:break-all;white-space:pre-wrap;">${escapeHTML(parts[2])}</pre></div>
            </div>
        </div>
    `;
    outputDiv.replaceChildren(card);
}

export function clearJWT() {
    document.getElementById('jwt-input').value = '';
    highlightJWTInput();
    document.getElementById('jwt-output').replaceChildren();
    showToast('JWT inputs cleared', 'success');
}

/* ------------------------------------------------------------------ */
/* Epoch                                                                */
/* ------------------------------------------------------------------ */

let liveEpochInterval = null;

export function startEpochClock() {
    const valEl = document.getElementById('live-epoch-value');
    if (!valEl) return;
    if (liveEpochInterval) clearInterval(liveEpochInterval);
    const update = () => { valEl.textContent = Math.floor(Date.now() / 1000); };
    update();
    liveEpochInterval = setInterval(update, 1000);
}

export function useCurrentEpoch() {
    const valEl = document.getElementById('live-epoch-value');
    const inputEl = document.getElementById('epoch-input');
    if (valEl && inputEl) {
        inputEl.value = valEl.textContent;
        syncDateTimePicker(new Date(parseInt(valEl.textContent) * 1000));
        showToast('Inserted current epoch into input', 'success');
    }
}

export function runEpoch() {
    let input = document.getElementById('epoch-input').value.trim();
    const datetimeInput = document.getElementById('epoch-datetime').value;
    const outputDiv = document.getElementById('epoch-output');
    if (!outputDiv) return;

    if (input === '' && datetimeInput !== '') input = datetimeInput + ':00';
    if (input === '') input = Math.floor(Date.now() / 1000).toString();

    saveToHistory('Epoch Convert', input);

    let targetTime = null;
    if (/^\d+$/.test(input)) {
        const num = parseInt(input);
        targetTime = input.length > 11 ? new Date(num) : new Date(num * 1000);
    } else {
        const parsed = Date.parse(input);
        if (!isNaN(parsed)) targetTime = new Date(parsed);
    }

    if (!targetTime || isNaN(targetTime.getTime())) {
        outputDiv.replaceChildren(errorCard('Could not parse date/time value. Support Unix epoch integers (seconds or milliseconds) and standard ISO8601 date strings.'));
        return;
    }

    const seconds = Math.floor(targetTime.getTime() / 1000);
    const millis = targetTime.getTime();
    const utcStr = targetTime.toUTCString();

    const localOffset = -targetTime.getTimezoneOffset();
    const diffSign = localOffset >= 0 ? '+' : '-';
    const absOffset = Math.abs(localOffset);
    const offsetStr = `${diffSign}${pad(Math.floor(absOffset / 60))}${pad(absOffset % 60)}`;
    const localStr = `${targetTime.getFullYear()}-${pad(targetTime.getMonth() + 1)}-${pad(targetTime.getDate())} `
        + `${pad(targetTime.getHours())}:${pad(targetTime.getMinutes())}:${pad(targetTime.getSeconds())} ${offsetStr}`;

    const diffMs = Date.now() - targetTime.getTime();
    const absDiffSec = Math.floor(Math.abs(diffMs) / 1000);
    const rel = formatRelativeSeconds(absDiffSec);
    const relative = diffMs >= 0 ? `${rel} ago` : `in ${rel}`;

    const card = resultCard('Conversion Results');
    card.querySelector('.card-body').innerHTML = `
        <div class="grid grid-cols-2 gap-4" style="display:grid;grid-template-columns:repeat(2,1fr);gap:16px;">
            <div class="form-group">
                <label>Epoch Seconds (10-digit)</label>
                <div class="flex gap-2" style="display:flex;gap:8px;">
                    <output id="epoch-sec-val" class="result-text-inline font-mono">${seconds}</output>
                    <copy-button class="btn btn-sm btn-secondary" target="epoch-sec-val" title="Copy seconds"></copy-button>
                </div>
            </div>
            <div class="form-group">
                <label>Epoch Milliseconds (13-digit)</label>
                <div class="flex gap-2" style="display:flex;gap:8px;">
                    <output id="epoch-ms-val" class="result-text-inline font-mono">${millis}</output>
                    <copy-button class="btn btn-sm btn-secondary" target="epoch-ms-val" title="Copy milliseconds"></copy-button>
                </div>
            </div>
        </div>
        <hr class="divider">
        <div class="form-group">
            <label>GMT / UTC Time</label>
            <div class="flex gap-2" style="display:flex;gap:8px;">
                <output id="epoch-utc-val" class="result-text-inline font-mono">${escapeHTML(utcStr)}</output>
                <copy-button class="btn btn-sm btn-secondary" target="epoch-utc-val" title="Copy UTC time"></copy-button>
            </div>
        </div>
        <div class="form-group mt-3" style="margin-top:12px;">
            <label>Local Time (Your system timezone)</label>
            <div class="flex gap-2" style="display:flex;gap:8px;">
                <output id="epoch-local-val" class="result-text-inline font-mono">${escapeHTML(localStr)}</output>
                <copy-button class="btn btn-sm btn-secondary" target="epoch-local-val" title="Copy local time"></copy-button>
            </div>
        </div>
        <div class="form-group mt-3" style="margin-top:12px;">
            <label style="margin-bottom:4px;">Relative Time</label>
            <div><span class="badge badge-accent font-semibold" style="font-weight:600;">${escapeHTML(relative)}</span></div>
        </div>
    `;
    outputDiv.replaceChildren(card);
    syncDateTimePicker(targetTime);
}

export function clearEpoch() {
    document.getElementById('epoch-input').value = '';
    document.getElementById('epoch-datetime').value = '';
    document.getElementById('epoch-output').replaceChildren();
    showToast('Epoch inputs cleared', 'success');
}

export function syncDateTimePicker(date) {
    const datetimeInput = document.getElementById('epoch-datetime');
    if (!datetimeInput) return;
    const p = (n) => String(n).padStart(2, '0');
    datetimeInput.value = `${date.getFullYear()}-${p(date.getMonth() + 1)}-${p(date.getDate())}T${p(date.getHours())}:${p(date.getMinutes())}`;
}

export function syncDateTimePickerFromText() {
    const textInput = document.getElementById('epoch-input');
    if (!textInput) return;
    const val = textInput.value.trim();
    if (!val) return;

    let targetTime = null;
    if (/^\d+$/.test(val)) {
        const num = parseInt(val);
        targetTime = val.length > 11 ? new Date(num) : new Date(num * 1000);
    } else {
        const parsed = Date.parse(val);
        if (!isNaN(parsed)) targetTime = new Date(parsed);
    }
    if (targetTime && !isNaN(targetTime.getTime())) syncDateTimePicker(targetTime);
}

export function syncTextInputFromDateTime() {
    const datetimeInput = document.getElementById('epoch-datetime');
    const textInput = document.getElementById('epoch-input');
    if (!datetimeInput || !textInput) return;
    if (datetimeInput.value !== '') textInput.value = datetimeInput.value + ':00';
}

// Human-readable durations via Intl.RelativeTimeFormat.
function formatRelativeSeconds(sec) {
    const rtf = new Intl.RelativeTimeFormat(undefined, { numeric: 'always' });
    const units = [
        ['day', 86400],
        ['hour', 3600],
        ['minute', 60],
        ['second', 1]
    ];
    for (const [unit, size] of units) {
        if (sec >= size || unit === 'second') {
            const value = Math.floor(sec / size);
            // rtf.format returns "in 5 minutes" / "5 minutes ago"; keep the magnitude only.
            return rtf.format(value, unit).replace(/^in\s+/, '').replace(/\s+ago$/, '');
        }
    }
    return `${sec} seconds`;
}

/* ------------------------------------------------------------------ */
/* GUID + Cron + Regex helpers                                          */
/* ------------------------------------------------------------------ */

export function updateRangeVal(val) {
    const el = document.getElementById('range-val-display');
    if (el) el.textContent = val;
}

export function togglePrefixInput(typeValue) {
    const prefixGroup = document.getElementById('prefix-group');
    const helpText = document.getElementById('prefix-format-help');
    if (!prefixGroup || !helpText) return;

    if (typeValue === 'prefixed' || typeValue === 'prefixed_guid') {
        prefixGroup.hidden = false;
        helpText.textContent = typeValue === 'prefixed'
            ? 'The generated format will be: {prefix}-{12 random hex characters}'
            : 'The generated format will be: {prefix} (padded/truncated to 8 chars) followed by standard UUID blocks';
    } else {
        prefixGroup.hidden = true;
    }
}

export function clearCron() {
    document.getElementById('cron-expression').value = '';
    document.getElementById('cron-output').replaceChildren(
        placeholderCard('bi-calendar3', 'Enter a cron expression and click "Parse & Preview" to see the schedule.')
    );
    showToast('Cron inputs cleared', 'success');
}

export function clearSQL() {
    document.getElementById('sql-input').value = '';
    const editor = getEditor('sql');
    if (editor) editor.setValue('');
    document.getElementById('sql-output').replaceChildren(
        placeholderCard('bi-database', 'Enter SQL query and select an operation to view output.')
    );
    showToast('SQL inputs cleared', 'success');
}

export function clearYAML() {
    document.getElementById('yaml-input').value = '';
    const editor = getEditor('yaml');
    if (editor) editor.setValue('');
    document.getElementById('yaml-output').replaceChildren(
        placeholderCard('bi-file-earmark-text', 'Enter payload and select an operation to view output.')
    );
    showToast('YAML inputs cleared', 'success');
}
