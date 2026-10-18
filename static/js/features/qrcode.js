// QR code generator: preset field visibility, live range readouts, and theme
// aware color defaults.
//
// The colors are user data rather than theme decoration, so they are read from
// the resolved CSS custom properties once per theme instead of being hardcoded.
// That keeps the swatches following the active theme without ever asking the
// author to invent a second per-theme color block.

// Which preset shows which group of sub-fields. Anything not listed renders no
// sub-fields and falls back to the free-text payload box.
const PRESET_FIELD_GROUPS = {
    url: ['qrcode-fields-url'],
    wifi: ['qrcode-fields-wifi'],
    vcard: ['qrcode-fields-contact'],
    mecard: ['qrcode-fields-contact'],
    mailto: ['qrcode-fields-contact'],
    tel: ['qrcode-fields-contact'],
    sms: ['qrcode-fields-message'],
    geo: ['qrcode-fields-geo']
};

// Placeholder + help copy per preset, so the main textarea always explains
// what it currently encodes.
const PRESET_TEXT_HELP = {
    text: 'Any text, link, or configuration payload is encoded verbatim.',
    url: 'Leave blank to use the address above, or type a path to encode.',
    wifi: 'Leave blank to encode the network details from the fields above.',
    vcard: 'Leave blank to encode the contact details from the fields above.',
    mecard: 'Leave blank to encode the contact details from the fields above.',
    mailto: 'Leave blank to encode the address from the fields above.',
    tel: 'Leave blank to encode the number from the fields above.',
    sms: 'Leave blank to encode the number from the fields above.',
    geo: 'Leave blank to encode the coordinates from the fields above.'
};

const DEFAULT_TEXT_PLACEHOLDER = 'https://example.com or any text data...';

function $(id) {
    return document.getElementById(id);
}

// resolveTokenColor reads a CSS custom property and returns it as a #rrggbb
// string the native color input can accept. getComputedStyle resolves
// light-dark() and color-mix() for us, so no parsing of the token text is
// needed. Returns null when the value cannot be represented.
function resolveTokenColor(token) {
    const raw = getComputedStyle(document.documentElement).getPropertyValue(token).trim();
    if (!raw) return null;

    // Tokens may hold hex, rgb()/rgba(), light-dark(), or color-mix(). Applying
    // the value to a throwaway element makes the engine resolve all of those
    // down to a computed rgb() triple for us.
    const probe = document.createElement('div');
    probe.style.color = raw;
    probe.style.display = 'none';
    document.body.appendChild(probe);
    const computed = getComputedStyle(probe).color;
    probe.remove();

    const match = computed.match(/rgba?\(\s*(\d+(?:\.\d+)?)[,\s]+(\d+(?:\.\d+)?)[,\s]+(\d+(?:\.\d+)?)/i);
    if (!match) return null;

    const toHex = (n) => Math.max(0, Math.min(255, Math.round(parseFloat(n))))
        .toString(16).padStart(2, '0');
    return `#${toHex(match[1])}${toHex(match[2])}${toHex(match[3])}`;
}

function themeColors() {
    return {
        fg: resolveTokenColor('--color-primary'),
        bg: resolveTokenColor('--card-bg')
    };
}

function setColorInputs() {
    const fg = $('qrcode-fg');
    const bg = $('qrcode-bg');
    if (!fg || !bg) return;

    const { fg: fgHex, bg: bgHex } = themeColors();
    // A color input cannot represent alpha, so translucent tokens such as
    // --card-bg are flattened to their resolved rgb channels. That is also the
    // right value for the generated code, which must be opaque to scan.
    if (fgHex) fg.value = fgHex;
    if (bgHex) bg.value = bgHex;
    updateContrastHelp();
}

function relativeLuminance(hex) {
    const m = /^#?([0-9a-f]{6})$/i.exec(hex || '');
    if (!m) return 0;
    const v = parseInt(m[1], 16);
    const lin = (c) => {
        const cs = c / 255;
        return cs <= 0.03928 ? cs / 12.92 : Math.pow((cs + 0.055) / 1.055, 2.4);
    };
    return 0.2126 * lin((v >> 16) & 0xff) + 0.7152 * lin((v >> 8) & 0xff) + 0.0722 * lin(v & 0xff);
}

// contrastRatio mirrors the server-side WCAG calculation so the warning shows
// up while the user drags the swatches, not only after generating.
function contrastRatio(a, b) {
    const la = relativeLuminance(a);
    const lb = relativeLuminance(b);
    const [hi, lo] = la > lb ? [la, lb] : [lb, la];
    return (hi + 0.05) / (lo + 0.05);
}

function updateContrastHelp() {
    const help = $('qrcode-contrast-help');
    const fg = $('qrcode-fg');
    const bg = $('qrcode-bg');
    if (!help || !fg || !bg) return;

    const ratio = contrastRatio(fg.value, bg.value);
    const weak = ratio < 3;
    help.textContent = weak
        ? `Contrast is ${ratio.toFixed(1)}:1. Most scanners need at least 3:1 — darken the modules or lighten the background.`
        : `Contrast is ${ratio.toFixed(1)}:1. Keep the two colors this far apart in brightness or scanners may struggle.`;
    help.style.color = weak ? 'var(--warning)' : '';
}

export function setQRCodePreset(value) {
    const groups = PRESET_FIELD_GROUPS[value] || [];

    const wrapper = $('qrcode-preset-fields');
    if (wrapper) wrapper.hidden = groups.length === 0;

    for (const group of Object.values(PRESET_FIELD_GROUPS)) {
        for (const id of group) {
            const el = $(id);
            if (el) el.hidden = !groups.includes(id);
        }
    }

    const textarea = $('qrcode-text');
    const help = $('qrcode-text-help');
    // The free-text box is only the primary input for the plain-text preset.
    // For every other preset the server builds the payload from the fields.
    if (textarea) textarea.required = value === 'text' || value === 'url';
    if (textarea && value !== 'text' && value !== 'url') {
        textarea.placeholder = 'Optional — overrides the fields above';
    } else if (textarea) {
        textarea.placeholder = DEFAULT_TEXT_PLACEHOLDER;
    }
    if (help) {
        help.textContent = PRESET_TEXT_HELP[value] || PRESET_TEXT_HELP.text;
    }
}

export function updateQRCodeRange(input) {
    const target = input.dataset.target ? $(input.dataset.target) : null;
    if (!target) return;

    const unit = input.id === 'qrcode-size' ? 'px' : '';
    target.textContent = `${input.value}${unit}`;

    if (input.id === 'qrcode-margin' && Number(input.value) < 2) {
        target.style.color = 'var(--warning)';
    } else {
        target.style.color = '';
    }
}

export function refreshQRCodeContrast() {
    updateContrastHelp();
}

export function resetQRCodeColors() {
    setColorInputs();
}

export function initQRCode() {
    setQRCodePreset('text');
    setColorInputs();

    // Keep the swatches in step with the theme. The theme toggle mutates
    // data-theme on <html>, so re-read the tokens once it flips.
    if (window.qrcodeThemeBound) return;
    window.qrcodeThemeBound = true;

    const observer = new MutationObserver(() => setColorInputs());
    observer.observe(document.documentElement, { attributes: true, attributeFilter: ['data-theme'] });
}