// Theme management. Tokens resolve through CSS light-dark(), so switching the
// data-theme attribute (and therefore color-scheme) is all that is required.
import { safeStorage } from './storage.js';

const THEME_KEY = 'devutils_theme';

export function initTheme() {
    const saved = safeStorage.getItem(THEME_KEY);
    if (saved) {
        document.documentElement.setAttribute('data-theme', saved);
    } else if (window.matchMedia && window.matchMedia('(prefers-color-scheme: light)').matches) {
        document.documentElement.setAttribute('data-theme', 'light');
    } else {
        document.documentElement.setAttribute('data-theme', 'dark');
    }
}

export function toggleTheme() {
    const current = document.documentElement.getAttribute('data-theme');
    const next = current === 'dark' ? 'light' : 'dark';
    document.documentElement.setAttribute('data-theme', next);
    safeStorage.setItem(THEME_KEY, next);
}
