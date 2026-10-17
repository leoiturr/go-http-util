// Toast + clipboard services.
// Toasts are delivered as a document-level CustomEvent so any module (or the
// <copy-button> custom element) can raise one without importing the UI.

export function showToast(message, type = 'success') {
    document.dispatchEvent(new CustomEvent('devutils:toast', {
        detail: { message, type }
    }));
}

export async function copyText(text) {
    await navigator.clipboard.writeText(text);
}

// Read the copyable text out of an element: the value for form controls,
// otherwise the visible text.
export function readCopyValue(el) {
    if (!el) return '';
    if (typeof el.value === 'string') return el.value;
    return el.innerText || el.textContent || '';
}
