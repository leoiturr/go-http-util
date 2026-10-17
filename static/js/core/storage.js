// Safe localStorage wrapper.
// Private Browsing / Firefox can throw a SecurityError on access, so every
// call is guarded and degrades to a no-op instead of breaking the page.
export const safeStorage = {
    getItem(key) {
        try {
            return localStorage.getItem(key);
        } catch (e) {
            console.warn('Storage disabled or blocked:', e);
            return null;
        }
    },
    setItem(key, value) {
        try {
            localStorage.setItem(key, value);
        } catch (e) {
            console.warn('Storage disabled or blocked:', e);
        }
    },
    removeItem(key) {
        try {
            localStorage.removeItem(key);
        } catch (e) {
            console.warn('Storage disabled or blocked:', e);
        }
    }
};
