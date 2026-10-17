// Reusable custom elements.
// - <copy-button> standardises every copy affordance in the app and in
//   server-rendered HTMX fragments.
// - <toast-stack> hosts transient <output> notifications.
import { showToast, copyText, readCopyValue } from './core/ui.js';

class CopyButton extends HTMLElement {
    connectedCallback() {
        if (this._ready) return;
        this._ready = true;

        const label = this.getAttribute('label') ?? this.textContent.trim();
        const icon = this.getAttribute('icon') || 'bi-clipboard';
        const title = this.getAttribute('title') || label || 'Copy';

        const button = document.createElement('button');
        button.type = 'button';
        button.className = this.getAttribute('class') || 'btn btn-sm btn-secondary';
        button.title = title;
        button.innerHTML = `<i class="bi ${icon}" aria-hidden="true"></i>`;
        if (label) {
            const span = document.createElement('span');
            span.textContent = label;
            button.appendChild(span);
        }
        button.addEventListener('click', () => this.copy());

        this.removeAttribute('class');
        this.replaceChildren(button);
    }

    async copy() {
        if (this.getAttribute('mode') === 'image') {
            return this.copyImage();
        }

        let text = '';
        const targetId = this.getAttribute('target');
        const select = this.getAttribute('select');
        const within = this.getAttribute('within');

        if (targetId) {
            text = readCopyValue(document.getElementById(targetId));
        } else if (select) {
            text = Array.from(document.querySelectorAll(select))
                .map((el) => el.textContent.trim())
                .filter(Boolean)
                .join('\n');
        } else if (within) {
            const scope = this.closest('tr') || this.parentElement;
            text = readCopyValue(scope ? scope.querySelector(within) : null);
        }

        if (!text) {
            showToast('Nothing to copy', 'error');
            return;
        }

        try {
            await copyText(text);
            showToast('Copied to clipboard!', 'success');
        } catch (err) {
            showToast('Failed to copy text', 'error');
            console.error('Copy failed:', err);
        }
    }

    async copyImage() {
        const img = document.getElementById(this.getAttribute('target'));
        if (!img || !img.getAttribute('src')) {
            showToast('Image not available', 'error');
            return;
        }
        const src = img.getAttribute('src');
        try {
            const blob = await (await fetch(src)).blob();
            await navigator.clipboard.write([new ClipboardItem({ 'image/png': blob })]);
            showToast('QR Code image copied to clipboard!', 'success');
        } catch (err) {
            try {
                await copyText(src);
                showToast('Copied QR Code Data URL (image copy blocked by browser)', 'warning');
            } catch (fallbackErr) {
                showToast('Failed to copy QR Code image', 'error');
                console.error(fallbackErr);
            }
        }
    }
}

class ToastStack extends HTMLElement {
    connectedCallback() {
        this._onToast = (event) => this.push(event.detail);
        document.addEventListener('devutils:toast', this._onToast);
    }

    disconnectedCallback() {
        document.removeEventListener('devutils:toast', this._onToast);
    }

    push({ message, type = 'success' }) {
        const icon = type === 'error'
            ? 'bi-x-circle-fill'
            : type === 'warning'
                ? 'bi-exclamation-circle-fill'
                : 'bi-check-circle-fill';

        const output = document.createElement('output');
        output.className = `toast toast-${type}`;
        const iconEl = document.createElement('i');
        iconEl.className = `bi ${icon}`;
        const span = document.createElement('span');
        span.textContent = message;
        output.append(iconEl, span);
        this.appendChild(output);

        setTimeout(() => {
            output.classList.add('fade-out');
            output.addEventListener('animationend', () => output.remove());
        }, 3000);
    }
}

customElements.define('copy-button', CopyButton);
customElements.define('toast-stack', ToastStack);

export { CopyButton, ToastStack };
