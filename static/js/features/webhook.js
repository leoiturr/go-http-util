// Webhook tester: preserve open <details> across the HTMX polling swaps and
// stop polling once the session has expired.
export function initWebhook() {
    if (window.webhookStatePreserverBound) return;
    window.webhookStatePreserverBound = true;

    document.body.addEventListener('htmx:before:swap', (evt) => {
        if (evt.detail.target.id === 'webhook-requests-container') {
            const details = evt.detail.target.querySelectorAll('details[open]');
            const openIds = Array.from(details).map((d) => d.id).filter((id) => id);
            evt.detail.target.dataset.openIds = JSON.stringify(openIds);
        }
    });

    document.body.addEventListener('htmx:after:swap', (evt) => {
        if (evt.detail.target.id !== 'webhook-requests-container') return;

        const openIdsStr = evt.detail.target.dataset.openIds;
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

        if (evt.detail.target.querySelector('.webhook-expired')) {
            evt.detail.target.removeAttribute('hx-trigger');
            evt.detail.target.removeAttribute('hx-get');
        }
    });
}
