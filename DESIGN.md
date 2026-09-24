# DevUtils - Architectural and UI Design Document

This document outlines the core architecture, layout design, and implementation details of the **DevUtils** suite.

## 1. Backend Architecture

### Web Server & Routing
- Built on the **Gin Web Framework** for Go.
- Serves static assets from `/static` and compiles HTML templates from `/templates`.
- Dynamic port detection:
  ```go
  port := os.Getenv("PORT")
  if port == "" {
      port = "8080"
  }
  ```
  This is optimized for hosting providers like **Render.com** that assign random listening ports.
- **Route Separation**: Routes are split into separate groups:
  - `/api/htmx/*`: Endpoint group returning HTML template fragments for HTMX-driven views (e.g., GUID and QR Code generation columns).
  - `/api/v1/*`: Versioned JSON REST API endpoint group supporting standard REST clients.
- **API Documentation & Specs**:
  - The OpenAPI 3.0.3 specification is defined statically in `static/openapi.json`.
  - An interactive REST API docs site is hosted at `/docs` rendering Swagger UI customized to match the dashboard's dark-theme aesthetic.

### API Rate Limiting
- **Implementation**: Located in `middleware/ratelimit.go`. Utilizes a custom, in-memory **Token Bucket** algorithm.
- **Thread Safety**: Protected with a `sync.Mutex` lock to handle concurrent requests safely.
- **Memory Leak Protection**: Runs a background goroutine that cleans up client states that haven't made requests in over 3 minutes.
- **Configuration**:
  - `RATE_LIMIT_RPS` (Requests Per Second, default `2.0`)
  - `RATE_LIMIT_BURST` (Burst capacity, default `4.0`)
- **HTTP Response**: Returns `429 Too Many Requests`.

---

## 2. Frontend & UI Design

### Single Page Application (SPA) Routing
- Avoids full page refreshes.
- Navigation utilizes HTML5 History API (`history.pushState` and `history.replaceState`) along with URL hashes (e.g. `#guid`, `#qrcode`).
- Supports the browser's **Back** and **Forward** buttons using the `popstate` event listener.
- Supports **Deep Linking** by reading the URL hash on initial page load.

### HTMX Integration & Error Handling
- Dynamic actions (generating GUIDs, encoding/decoding Base64, creating QR codes, and testing regular expressions) are triggered via `hx-post`.
- Non-2xx status codes (like the rate limiter's `429`) are caught globally in `app.js` using the `htmx:responseError` listener and shown as clean toast notifications.

### Visual Styling & Theme
- **Design System**: **Warm Paper Workbench** — warm muted surfaces, flat paper-card components with hairline borders, a subtle SVG grain texture layered over soft ambient washes, and a single bold **vermilion** accent used sparingly for actions, active navigation, and highlights.
- **Dual Theme**: Dark mode (warm charcoal-walnut) is the default; light mode (`[data-theme="light"]`) is a cream paper variant. A toggle button in the top bar switches themes. The preference is persisted in `localStorage` under the key `devutils_theme`. On first visit, the user's system preference (`prefers-color-scheme`) is respected, defaulting to dark if unavailable.
- **CSS Variables**: All colors are defined as CSS custom properties on `:root` (dark defaults) and overridden via `[data-theme="light"]`. This ensures all components automatically adapt to the active theme. Derived tints use `color-mix(in srgb, var(--token) X%, transparent)` rather than hardcoded channel colors.
- **Palette**:
  | Token | Dark (walnut) | Light (paper) |
  |---|---|---|
  | `--bg-primary` | `#16120f` | `#f5efe3` |
  | `--bg-secondary` | `#1d1813` | `#fcf8ef` |
  | `--color-primary` (vermilion) | `#e4482a` | `#c53819` |
  | `--color-accent` | `#ef5b34` | `#d1441f` |
  | `--color-pink` (ochre) | `#e0a03a` | `#a86f16` |
  | `--text-primary` | `#f2e9db` | `#2b2218` |
  | `--text-muted` | `#a59685` | `#7d6f5c` |
- **Background Texture**: The decorative background uses soft radial washes (warm tints of the accent family) plus a `paper-grain` overlay (SVG `feTurbulence` noise blended at low opacity). In light mode the grain switches to a gentle `multiply` blend for a cleaner paper feel.
- **Typography**:
  - Display/headings: **Fraunces** (warm editorial serif) — used for the brand wordmark, top-bar titles, and card headings. Loaded via **Bunny Fonts** (GDPR compliant).
  - Sans-serif: **Plus Jakarta Sans** (UI/body text) — loaded via **Bunny Fonts**.
  - Monospace: **Comic Shanns** (Comic Sans-inspired monospace) — used for code, results, and editors; served via `@font-face` from jsDelivr (MIT licensed, not available on Bunny Fonts).
- **Glow & Lift**: Cards use soft drop shadows with restrained depth; the vermilion accent glows only on primary buttons, the active nav item, focus rings, and the brand logo.
- **Inset Code Surfaces**: Long-form inputs such as the JWT editor use a nested paper frame. The outer frame carries the border, focus ring, rounded `--radius-md` corners, and a small padding gutter; the inner editor uses `--radius-sm` and clips overflow so long tokens cannot create square corners. JWT settings cards keep responsive inset padding and clip their child surfaces to `--radius-lg`.

### Utility Curation
- New utilities should be focused, local-first where privacy matters, and small enough to complete in one workbench view.
- The **Regex Playground** lives in the **Inspect & Format** group. It accepts pattern presets, supports `i`, `m`, `s`, and `U` flags, highlights matches in escaped result segments, and exposes capture groups as structured cards. An Engine selector (styled like the flag tiles) switches between **Go RE2 (server)** — the linear-time default — and **JavaScript (browser)**, which tests locally with the tab's own engine for instant feedback and lookahead support.
- In JavaScript mode the `U` flag tile is dimmed and disabled (RE2-only modifier), the pattern help text explains the trade-off, and execution runs in a sandboxed Web Worker with a 2-second timeout so catastrophic backtracking only freezes the worker, never the page. Timeouts and compile errors render through the same Regex Match Report error card. Both engines produce identical report markup and UTF-8 byte offsets; `(?P<name>` patterns port automatically to JS `(?<name>`.
- Regex requests and responses are bounded to protect the workbench: patterns are limited to 4 KB, test text to 200 KB, and returned match details to the first 100 matches. Match positions are documented as UTF-8 byte offsets.
- Regex pattern and test-text editors reuse nested rounded paper frames so code-like content remains visually separate from the settings card. On narrow screens, flag and capture grids collapse to one column.

### Mobile-First Responsive Design
- **Mobile-First CSS**: Designed base CSS rules to fit mobile screens (vertical stacked layouts, full-width sidebars serving as top menus, scrollable nav menus, and compact card/grid spacing).
- **Desktop Enhancements**: Uses `min-width: 1025px` media queries to expand the layout into side-by-side splits. The sidebar and main panel are inset from the viewport, separated by a 12px gutter, and each uses `--radius-lg` so their outer corners stay rounded. The sticky top bar bleeds through the main panel's 32px side padding (`margin: 0 -32px` with matching padding) so its background reaches the panel's rounded corners while content stays aligned with the cards below.
- **Highlighted Test Text**: In the Regex Match Report, the `white-space: pre-wrap` highlight preview must contain only the result segments. Template markup inside it stays on one line, because any newline or indentation would render as literal text and shift the first line out of sync with the editor's test text.
- **Native Select Menus**: Closed selects follow the active theme. Their option popups are OS-rendered and remain light, so option text uses `--select-menu-text` on `--select-menu-bg` rather than the theme's light foreground. This keeps every option readable.
