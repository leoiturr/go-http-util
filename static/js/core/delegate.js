// Event delegation engine.
// Every interactive element declares data-action="name"; handlers are looked
// up in a registry. One listener per event type on document means server-
// rendered HTMX fragments work the moment they are swapped in, with no inline
// handlers and no re-binding.

const actions = new Map();

export function registerActions(map) {
    for (const [name, handler] of Object.entries(map)) {
        actions.set(name, handler);
    }
}

function findActionElement(target) {
    if (!(target instanceof Element)) return null;
    return target.closest('[data-action]');
}

function dispatch(event) {
    const el = findActionElement(event.target);
    if (!el) return;
    const handler = actions.get(el.dataset.action);
    if (!handler) return;
    handler(el, event);
}

let started = false;

export function initDelegation() {
    if (started) return;
    started = true;
    // 'submit' bubbles, so form actions can run here too. Handlers that need
    // to suppress the browser default call event.preventDefault() themselves.
    for (const type of ['click', 'input', 'change', 'submit']) {
        document.addEventListener(type, dispatch);
    }
}
