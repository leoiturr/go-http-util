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
  - Sans-serif: `Plus Jakarta Sans`
  - Monospace: `Lilex` (developer-focused font with coding ligatures)

## Architectural Guidelines
1. **Dynamic Port Binding**: The application must bind to the port defined by the `PORT` environment variable (injected by Render/Heroku) and fall back to `8080` in local environments.
2. **Rate Limiting**: All `/api/*` endpoints are protected by the custom Token-Bucket rate limiting middleware in `middleware/ratelimit.go`. Avoid introducing external packages for this.
3. **SPA Tab Navigation**: Tab switching is handled client-side using `switchTab()`. It updates the URL hash and pushes states to the browser history (`pushState`).
4. **HTMX Error Handling**: Always check for non-2xx status codes (like `429 Too Many Requests`) using the global HTMX error handler in `static/js/app.js`.
   - HTMX 4.x uses `fetch()` internally; access response status via `evt.detail.ctx.response.status` (not `evt.detail.xhr`).
   - HTMX 4.x event names use colon-separated format: `htmx:response:error`, `htmx:before:swap`, `htmx:after:swap`, `htmx:after:settle`, etc. Never use the legacy camelCase names from 1.x/2.x.
5. **Route Separation**: Maintain a clear separation of routes between HTMX layout fragments (`/api/htmx/*`) and REST API endpoints (`/api/v1/*`).
6. **API Documentation**: The interactive REST API specification is stored in `static/openapi.json` and rendered at `/docs` using custom-themed Swagger UI assets. Any updates to REST endpoints must be reflected in the schema file.

## Theming & Styling Conventions
- **CSS Variables**: All colors are defined as CSS custom properties in `:root` (dark defaults) and overridden via `[data-theme="light"]` in `static/css/styles.css`. Never use hardcoded hex/rgb values for colors in components; always reference `var(--*)` tokens.
- **Available Tokens**: `--bg-primary`, `--bg-secondary`, `--sidebar-bg`, `--card-bg`, `--card-border`, `--text-primary`, `--text-muted`, `--text-dark`, `--color-primary`, `--color-accent`, `--color-pink`, `--success`, `--danger`, `--warning`.
- **Theme Toggle**: The toggle button is in the top bar (`.theme-toggle`). Theme preference is stored in `localStorage` under `devutils_theme` and initialized via `initTheme()` in `static/js/app.js`.
- **Inline Styles**: Avoid inline color values. If inline styles are necessary, use CSS variables (e.g., `color: var(--text-muted)`). The only exceptions are semantic per-item colors (like HTTP method badges in webhook requests) that are identical in both themes.
- **New Components**: Always use existing CSS variable tokens. If a new color is needed, add it to both `:root` and `[data-theme="light"]` in `styles.css`.
