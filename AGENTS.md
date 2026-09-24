# AI Agent Development Guidelines

Welcome! This document provides guidelines for AI agents contributing to the **DevUtils** repository.

## Formatting Rules
- **Go Code**: Always run `go fmt ./...` to format all Go source files before committing.
- **JavaScript & HTML**: Do not use automatic formatters (like Prettier). Keep the formatting clean but unformatted.
- **SQL**: Format SQL queries using standard trailing commas and new lines for keywords. For example:
  ```sql
  SELECT
    id,
    name,
    email
  FROM
    users
  WHERE
    status = 'active'
  ORDER BY
    created_at DESC
  LIMIT
    10;
  ```

## Documentation & Curation Guidelines
- **Keep Design Documentation Current**: Update `DESIGN.md` in the same change whenever architecture, layout, interaction, typography, theming, or reusable UI treatment changes. Do not defer design decisions to a later cleanup.
- **Promote Durable Rules**: Update this file when a repeated implementation rule is discovered, such as documentation synchronization, responsive behavior, accessibility, or component geometry.
- **Surface Geometry**: Cards and nested input/code surfaces must use deliberate padding, existing radius tokens, and overflow clipping. Avoid square inner corners or content touching an outer card edge.
- **Suggest Complementary Utilities**: When asked for a new tool beside DevUtils, first suggest one focused, local-first developer utility. Check the existing navigation for overlap, choose the best-fit nav group, and explain its value before implementation.
- **Do Not Install or Use Playwright**: Never install, add, or run Playwright in this repository. If browser-based visual review or a design opinion is needed, ask the user first.

## Git Commit Guidelines
- **Oneline commits**: All commit messages must be a single line (max 72 chars).
- **Conventional commits**: Use `type(scope): description` format.
  - Types: `feat`, `fix`, `refactor`, `chore`, `docs`, `style`, `test`
  - Scope: Optional, indicates affected area (e.g., `ui`, `api`, `auth`)
  - Description: Imperative mood, lowercase, no trailing period
- **Examples**: 
  - `feat(ui): add datetime-local picker to Epoch Converter`
  - `fix(api): handle 429 rate limit in webhook handler`
  - `refactor: extract token bucket logic to middleware`
- **Amend instead of new commit**: When fixing a recent commit, use `git commit --amend` to keep history clean.

## Versioning Guidelines
- **Application Version**: The application version string is defined in `main.go` inside the `getVersion()` function.
  - When bumping versions (e.g. to `v1.0.3`), update all version prefixes in `main.go`:
    1. The Render environment variable branch: `"v<VERSION>-" + renderCommit[:7]` and `"v<VERSION>-" + renderCommit`
    2. The local git commit branch: `"v<VERSION>-" + commit`
    3. The final fallback return value: `"v<VERSION>"`
  - Format Go code (`go fmt ./...`) and run tests (`go test ./...`) before committing.
  - Commit using conventional commit format: `chore: bump version to <VERSION>` (e.g., `chore: bump version to 1.0.3`).

## Tech Stack Overview
- **Backend**: Go (using the **Gin** Web Framework).
- **Frontend**: HTML5, Vanilla CSS, and **HTMX 4.x** for dynamic content swaps without full page reloads.
- **Fonts**: Loaded via **Bunny Fonts** (GDPR compliant). 
  - Display/headings: `Fraunces` (warm editorial serif; `--font-display`)
  - Sans-serif: `Plus Jakarta Sans` (UI/body; `--font-sans`)
  - Monospace: `Comic Shanns` (self-served via jsDelivr; Comic Sans-inspired monospace; `--font-mono`)

## Architectural Guidelines
1. **Dynamic Port Binding**: The application must bind to the port defined by the `PORT` environment variable (injected by Render/Heroku) and fall back to `8080` in local environments.
2. **Rate Limiting**: All `/api/*` endpoints are protected by the custom Token-Bucket rate limiting middleware in `middleware/ratelimit.go`. Avoid introducing external packages for this.
3. **SPA Tab Navigation**: Tab switching is handled client-side using `switchTab()`. It updates the URL hash and pushes states to the browser history (`pushState`).
4. **HTMX Error Handling**: Always check for non-2xx status codes (like `429 Too Many Requests`) using the global HTMX error handler in `static/js/app.js`.
   - HTMX 4.x uses `fetch()` internally; access response status via `evt.detail.ctx.response.status` (not `evt.detail.xhr`).
   - HTMX 4.x event names use colon-separated format: `htmx:response:error`, `htmx:before:swap`, `htmx:after:swap`, `htmx:after:settle`, etc. Never use the legacy camelCase names from 1.x/2.x.
5. **Route Separation**: Maintain a clear separation of routes between HTMX layout fragments (`/api/htmx/*`) and REST API endpoints (`/api/v1/*`).
6. **API Documentation**: The interactive REST API specification is stored in `static/openapi.json` and rendered at `/docs` using custom-themed Swagger UI assets. Any updates to REST endpoints must be reflected in the schema file.
7. **Safe Regex Execution**: The Regex Playground defaults to Go's linear-time RE2 engine. An explicit review (user-approved) added an opt-in browser JavaScript engine: it runs patterns in a dedicated, killable Web Worker (`regexJsWorkerSource` in `static/js/app.js`) with a 2-second timeout, and mirrors the server's pattern, test-text, and match limits. Never evaluate user patterns on the server with a backtracking engine or outside the sandboxed worker. Keep RE2 as the default selection. The JavaScript report renderer (`renderRegexReport` in `static/js/app.js`) mirrors the `regex-result` Go template; update both together when the report markup changes.

## Theming & Styling Conventions
- **Design System**: The UI follows a **Warm Paper Workbench** aesthetic (see `DESIGN.md`): warm muted surfaces, paper-grain texture over soft warm washes, and a single bold **vermilion** accent (`--color-primary`). Dark mode is the default; light mode (`[data-theme="light"]`) is a cream paper variant.
- **CSS Variables**: All colors are defined as CSS custom properties in `:root` (dark defaults) and overridden via `[data-theme="light"]` in `static/css/styles.css`. Never use hardcoded hex/rgb values for colors in components; always reference `var(--*)` tokens. Derived tints should use `color-mix(in srgb, var(--token) X%, transparent)` and never hardcode the channel colors.
- **Available Tokens**: `--bg-primary`, `--bg-secondary`, `--sidebar-bg`, `--card-bg`, `--card-border`, `--card-border-hover`, `--text-primary`, `--text-muted`, `--text-dark`, `--color-primary` (vermilion), `--color-accent`, `--color-pink`, `--success`, `--danger`, `--warning`, plus font tokens `--font-display`. `--font-sans`, `--font-mono`, and layout tokens `--radius-lg`, `--radius-md`, `--radius-sm`, `--transition-fast`, `--transition-normal`.
- **Theme Toggle**: The toggle button is in the top bar (`.theme-toggle`). Theme preference is stored in `localStorage` under `devutils_theme` and initialized via `initTheme()` in `static/js/app.js`.
- **Inline Styles**: Avoid inline color values. If inline styles are necessary, use CSS variables (e.g., `color: var(--text-muted)`) or `color-mix()` with tokens. The only exceptions are semantic per-item colors (like HTTP method badges in webhook requests) that are identical in both themes.
- **New Components**: Always use existing CSS variable tokens. If a new color is needed, add it to both `:root` and `[data-theme="light"]` in `styles.css`.
- **Icons**: Use only Font Awesome Free icons available from the loaded 6.5.2 stylesheet. Do not use Pro-only names such as `fa-brackets-curly`.
- **Native Selects**: Option popup text must remain readable on the light OS menu. Use `--select-menu-text` and `--select-menu-bg`; do not inherit the dark theme's light foreground into `option` elements.
