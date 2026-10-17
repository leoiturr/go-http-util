// Small shared helpers.

export function escapeHTML(str) {
    return String(str == null ? '' : str)
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;')
        .replace(/"/g, '&quot;')
        .replace(/'/g, '&#039;');
}

export function truncateString(str, len) {
    if (str.length <= len) return str;
    return str.substring(0, len) + '...';
}

export function formatDuration(sec) {
    if (sec < 60) return `${sec}s`;
    const min = Math.floor(sec / 60);
    if (min < 60) return `${min}m`;
    const hr = Math.floor(min / 60);
    if (hr < 24) return `${hr}h`;
    const day = Math.floor(hr / 24);
    return `${day}d`;
}

export function pad(n) {
    return n < 10 ? '0' + n : n;
}

// Instantiate a <template> by id. Returns a cloned DocumentFragment.
export function fromTemplate(id) {
    const tpl = document.getElementById(id);
    if (!(tpl instanceof HTMLTemplateElement)) {
        throw new Error(`Missing <template id="${id}">`);
    }
    return tpl.content.cloneNode(true);
}
