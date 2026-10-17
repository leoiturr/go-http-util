// Webhook tester: preserve open <details> across the HTMX polling swaps and
// stop polling once the session has expired.
//
// HTMX 4 note: the swap events carry { ctx } in their detail, and the element
// being swapped is ctx.target (the event is dispatched on ctx.sourceElement).
function swappedTarget(evt) {
    const detail = evt.detail || {};
    return (detail.ctx && detail.ctx.target) || null;
}

export function initWebhook() {
    if (window.webhookStatePreserverBound) return;
    window.webhookStatePreserverBound = true;

    document.body.addEventListener('htmx:before:swap', (evt) => {
        const target = swappedTarget(evt);
        if (!target || target.id !== 'webhook-requests-container') return;

        const openIds = Array.from(target.querySelectorAll('details[open]'))
            .map((d) => d.id)
            .filter(Boolean);
        target.dataset.openIds = JSON.stringify(openIds);
    });

    document.body.addEventListener('htmx:after:swap', (evt) => {
        const target = swappedTarget(evt);
        if (!target || target.id !== 'webhook-requests-container') return;

        const openIdsStr = target.dataset.openIds;
        if (openIdsStr) {
            try {
                JSON.parse(openIdsStr).forEach((id) => {
                    const el = document.getElementById(id);
                    if (el) el.open = true;
                });
            } catch (e) {
                /* ignore malformed state */
            }
        }

        if (target.querySelector('.webhook-expired')) {
            target.removeAttribute('hx-trigger');
            target.removeAttribute('hx-get');
        }
    });
}
