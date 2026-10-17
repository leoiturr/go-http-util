// Regex Playground.
// Go RE2 (server, default) is tested through HTMX; the optional JavaScript
// engine runs locally in a killable Web Worker so a catastrophic pattern can
// only freeze the worker, never the page.
import { fromTemplate } from '../core/utils.js';
import { showToast } from '../core/ui.js';

const PRESETS = {
    email: {
        pattern: '(?P<local>[\\w.+-]+)@(?P<domain>[\\w.-]+)',
        text: 'Email ada@example.com\nInvalid: not-an-email'
    },
    url: {
        pattern: 'https?://[^\\s]+',
        text: 'Documentation: https://devutils.local/docs\nSupport: mailto:team@devutils.local'
    },
    ipv4: {
        pattern: '((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\\.){3}(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)',
        text: 'Gateway 192.168.1.1\nPublic 8.8.8.8\nInvalid 999.1.1.1'
    },
    date: {
        pattern: '([0-9]{4})-([0-9]{2})-([0-9]{2})',
        text: 'Release 2026-09-23\nReview 2026-10-01'
    },
    hex: {
        pattern: '#[0-9A-Fa-f]{6}',
        text: 'Primary #E4482A\nSecondary #ef5b34\nInvalid #12345'
    }
};

let engineMode = 're2';
const JS_TIMEOUT_MS = 2000;
let jsWorker = null;
let jsWorkerUrl = null;
let jsRunId = 0;

export function loadRegexPreset(presetName) {
    const preset = PRESETS[presetName];
    if (!preset) return;
    document.getElementById('regex-pattern').value = preset.pattern;
    document.getElementById('regex-test-text').value = preset.text;
}

export function clearRegex() {
    document.getElementById('regex-preset').value = 'email';
    loadRegexPreset('email');
    document.querySelectorAll('#module-regex input[name="flags"]').forEach((input) => {
        input.checked = false;
    });
    document.getElementById('regex-output').replaceChildren();
    showToast('Regex Playground reset', 'success');
}

export function setRegexEngine(engine) {
    engineMode = engine === 'js' ? 'js' : 're2';

    const ungreedy = document.querySelector('#module-regex .regex-flag[data-flag="U"]');
    if (ungreedy) {
        ungreedy.querySelector('input').disabled = engineMode === 'js';
        ungreedy.classList.toggle('regex-flag-unsupported', engineMode === 'js');
        ungreedy.title = engineMode === 'js'
            ? 'Ungreedy is a Go RE2 modifier and is unavailable in JavaScript.'
            : '';
    }

    const help = document.getElementById('regex-syntax-help');
    if (help) {
        help.textContent = engineMode === 'js'
            ? 'JavaScript runs locally in this tab and supports lookaheads, but it backtracks: slow patterns are stopped after 2 seconds.'
            : 'RE2 syntax is linear-time and resistant to catastrophic backtracking.';
    }
}

function getJSWorker() {
    if (jsWorker) return jsWorker;
    if (!jsWorkerUrl) {
        jsWorkerUrl = URL.createObjectURL(new Blob([jsWorkerSource], { type: 'text/javascript' }));
    }
    jsWorker = new Worker(jsWorkerUrl);
    return jsWorker;
}

function runRegexJavaScript(form) {
    const output = document.getElementById('regex-output');
    const loader = document.getElementById('regex-loader');
    const flags = Array.from(form.elements.namedItem('flags'))
        .filter((input) => input.checked && input.value !== 'U')
        .map((input) => input.value);

    let worker;
    try {
        worker = getJSWorker();
    } catch (err) {
        renderRegexReport(output, { Error: 'This browser could not start a Web Worker for the JavaScript engine. Switch the engine back to Go RE2.' });
        return;
    }

    const runId = ++jsRunId;
    loader.classList.add('htmx-request');

    const timer = setTimeout(() => {
        if (runId !== jsRunId) return;
        worker.terminate();
        jsWorker = null;
        loader.classList.remove('htmx-request');
        renderRegexReport(output, { Error: 'JavaScript regex execution timed out after 2 seconds. This pattern likely causes catastrophic backtracking. Simplify it or switch the engine to Go RE2.' });
    }, JS_TIMEOUT_MS);

    worker.onmessage = (evt) => {
        if (runId !== jsRunId) return;
        clearTimeout(timer);
        loader.classList.remove('htmx-request');
        const data = evt.data || {};
        renderRegexReport(output, data.error ? { Error: data.error } : { Result: data.result });
    };

    worker.onerror = () => {
        if (runId !== jsRunId) return;
        clearTimeout(timer);
        loader.classList.remove('htmx-request');
        renderRegexReport(output, { Error: 'The JavaScript engine could not evaluate this pattern.' });
    };

    worker.postMessage({
        pattern: form.elements.namedItem('pattern').value,
        text: form.elements.namedItem('test_text').value,
        flags,
        matchLimit: 100
    });
}

// Mirror of the "regex-result" Go template so both engines render the same report.
export function renderRegexReport(container, payload) {
    const card = fromTemplate('tpl-regex-report');
    const actions = card.querySelector('.card-actions');
    const body = card.querySelector('.card-body');
    const result = payload.Result;

    if (!result) {
        const alert = document.createElement('div');
        alert.className = 'alert alert-error';
        alert.innerHTML = '<i class="bi bi-exclamation-circle-fill"></i><div></div>';
        alert.querySelector('div').textContent = payload.Error || 'Unknown error';
        body.replaceChildren(alert);
        container.replaceChildren(card);
        return;
    }

    const badge = document.createElement('span');
    badge.className = 'badge badge-secondary';
    badge.textContent = result.engine;
    actions.appendChild(badge);

    const summary = document.createElement('div');
    summary.className = 'regex-summary';
    summary.setAttribute('aria-label', 'Regex result summary');
    summary.append(
        summaryItem(result.matchCount, 'Matches'),
        summaryItem(result.captureCount, result.truncated ? 'Captures (shown)' : 'Captures'),
        summaryItem(result.groupCount, 'Groups')
    );
    body.appendChild(summary);

    const flags = document.createElement('div');
    flags.className = 'regex-flags-summary';
    const flagLabel = document.createElement('span');
    flagLabel.className = 'form-group-label';
    flagLabel.textContent = 'Active Flags';
    flags.appendChild(flagLabel);
    if ((result.flags || []).length) {
        result.flags.forEach((flag) => {
            const key = document.createElement('span');
            key.className = 'regex-flag-key';
            key.textContent = flag;
            flags.appendChild(key);
        });
    } else {
        const none = document.createElement('span');
        none.className = 'text-muted text-sm';
        none.textContent = 'None';
        flags.appendChild(none);
    }
    body.appendChild(flags);

    if (result.truncated) {
        const warn = document.createElement('div');
        warn.className = 'alert alert-warning mb-4';
        warn.innerHTML = '<i class="bi bi-exclamation-triangle-fill"></i><div></div>';
        warn.querySelector('div').textContent = `Showing the first ${result.matchLimit} matches. Refine the pattern to inspect additional results.`;
        body.appendChild(warn);
    }

    const highlightGroup = document.createElement('div');
    highlightGroup.className = 'form-group';
    const highlightLabel = document.createElement('span');
    highlightLabel.className = 'form-group-label';
    highlightLabel.textContent = 'Highlighted Test Text';
    const preview = document.createElement('div');
    preview.className = 'regex-highlight-preview';
    preview.setAttribute('aria-label', 'Test text with matched sections highlighted');
    // Contains only the result segments (no stray whitespace) so it lines up
    // with the test-text editor.
    (result.segments || []).forEach((segment) => {
        if (segment.matched) {
            const mark = document.createElement('mark');
            mark.className = 'regex-match-highlight';
            mark.title = `Match ${segment.matchIndex}`;
            mark.textContent = segment.value;
            preview.appendChild(mark);
        } else {
            preview.appendChild(document.createTextNode(segment.value));
        }
    });
    highlightGroup.append(highlightLabel, preview);
    body.appendChild(highlightGroup);

    const divider = document.createElement('hr');
    divider.className = 'divider';
    const list = document.createElement('div');
    list.className = 'regex-match-list';

    const matches = result.matches || [];
    if (matches.length) {
        matches.forEach((match, index) => list.appendChild(matchCard(match, index)));
    } else {
        list.appendChild(fromTemplate('tpl-regex-empty'));
    }
    body.append(divider, list);

    container.replaceChildren(card);
}

function summaryItem(value, label) {
    const node = fromTemplate('tpl-regex-summary-item');
    node.querySelector('.regex-summary-value').textContent = value;
    node.querySelector('.regex-summary-label').textContent = label;
    return node;
}

function matchCard(match, index) {
    const node = fromTemplate('tpl-regex-match-card');
    const copy = node.querySelector('copy-button');
    copy.setAttribute('target', `regex-match-${index}`);
    copy.setAttribute('title', `Copy match ${match.index}`);
    node.querySelector('.regex-match-number').textContent = `Match ${match.index}`;
    node.querySelector('.regex-match-range').textContent = `bytes ${match.start}–${match.end}`;
    const pre = node.querySelector('.regex-match-value');
    pre.id = `regex-match-${index}`;
    pre.textContent = match.value;

    const capturesWrap = node.querySelector('.regex-captures');
    if ((match.captures || []).length) {
        const grid = document.createElement('div');
        grid.className = 'regex-capture-grid';
        match.captures.forEach((capture) => {
            const item = fromTemplate('tpl-regex-capture-item');
            item.querySelector('.regex-capture-name').textContent = capture.name || `Group ${capture.index}`;
            item.querySelector('.regex-capture-value').textContent = capture.matched
                ? (capture.value || '(empty)')
                : 'Not matched';
            grid.appendChild(item);
        });
        capturesWrap.appendChild(grid);
    } else {
        const none = document.createElement('div');
        none.className = 'regex-no-captures';
        none.textContent = 'No capture groups in this pattern.';
        capturesWrap.appendChild(none);
    }
    return node;
}

// Cancels the HTMX request when the local JavaScript engine is active and runs
// the test in the browser instead.
export function initRegex() {
    document.addEventListener('htmx:before:request', (evt) => {
        const form = evt.target && evt.target.closest ? evt.target.closest('#regex-form') : null;
        if (form && engineMode === 'js') {
            evt.preventDefault();
            runRegexJavaScript(form);
        }
    });

    // Mark the preset as custom when the pattern is edited directly.
    const pattern = document.getElementById('regex-pattern');
    if (pattern) {
        pattern.addEventListener('input', () => {
            document.getElementById('regex-preset').value = 'custom';
        });
    }

    const initialEngine = document.querySelector('#regex-form input[name="engine"]:checked');
    if (initialEngine) setRegexEngine(initialEngine.value);
}

// Sandbox: the pattern only ever runs here, in a worker the main thread can
// terminate after JS_TIMEOUT_MS.
const jsWorkerSource = `
'use strict';

function parseGroupNames(source) {
    const names = {};
    let count = 0;
    let inClass = false;
    for (let i = 0; i < source.length; i++) {
        const ch = source.charAt(i);
        if (ch === '\\\\') { i += 1; continue; }
        if (inClass) { if (ch === ']') inClass = false; continue; }
        if (ch === '[') { inClass = true; continue; }
        if (ch !== '(') continue;
        if (source.startsWith('(?<', i) && !source.startsWith('(?<=', i) && !source.startsWith('(?<!', i)) {
            const close = source.indexOf('>', i + 3);
            if (close !== -1) {
                count += 1;
                names[count] = source.slice(i + 3, close);
                i = close;
                continue;
            }
        }
        if (source.charAt(i + 1) !== '?') count += 1;
    }
    return { names, count };
}

self.onmessage = function (evt) {
    const pattern = evt.data.pattern || '';
    const text = evt.data.text || '';
    const requestedFlags = evt.data.flags || [];
    const matchLimit = evt.data.matchLimit || 100;

    if (!pattern) { postMessage({ error: 'regex pattern is required' }); return; }
    if (pattern.length > 4096) { postMessage({ error: 'regex pattern exceeds the 4 KB limit' }); return; }
    if (text.length > 204800) { postMessage({ error: 'test text exceeds the 200 KB limit' }); return; }

    const flags = ['i', 'm', 's'].filter(function (flag) { return requestedFlags.indexOf(flag) !== -1; });
    const jsPattern = pattern.split('(?P<').join('(?<');

    let expression;
    try {
        expression = new RegExp(jsPattern, flags.join('') + 'g');
    } catch (err) {
        postMessage({ error: 'invalid regular expression: ' + err.message });
        return;
    }

    const groups = parseGroupNames(jsPattern);
    const encoder = new TextEncoder();
    const collected = [];
    let total = 0;
    let groupCount = groups.count;
    let match;
    while ((match = expression.exec(text)) !== null) {
        total += 1;
        if (collected.length < matchLimit) {
            const captures = [];
            for (let g = 1; g < match.length; g++) {
                const value = match[g];
                captures.push({ index: g, name: groups.names[g] || '', value: value === undefined ? '' : value, matched: value !== undefined });
            }
            collected.push({ start: match.index, end: match.index + match[0].length, value: match[0], captures: captures });
        }
        if (match.length - 1 > groupCount) groupCount = match.length - 1;
        if (match[0] === '') expression.lastIndex += 1;
    }

    const matches = collected.map(function (item, index) {
        return {
            index: index + 1,
            value: item.value,
            start: encoder.encode(text.slice(0, item.start)).length,
            end: encoder.encode(text.slice(0, item.end)).length,
            captures: item.captures
        };
    });

    const segments = [];
    let cursor = 0;
    collected.forEach(function (item, index) {
        if (item.start > cursor) segments.push({ value: text.slice(cursor, item.start), matched: false });
        segments.push({ value: item.value, matched: true, matchIndex: index + 1 });
        if (item.end > cursor) cursor = item.end;
    });
    if (cursor < text.length) segments.push({ value: text.slice(cursor), matched: false });

    let captureCount = 0;
    matches.forEach(function (item) {
        item.captures.forEach(function (capture) { if (capture.matched) captureCount += 1; });
    });

    postMessage({ result: {
        pattern: pattern,
        flags: flags,
        segments: segments,
        matches: matches,
        matchCount: total,
        groupCount: groupCount,
        captureCount: captureCount,
        matchLimit: matchLimit,
        truncated: total > matchLimit,
        engine: 'JavaScript'
    } });
};
`;
