# AI Agent Development Guidelines

Welcome! This document provides guidelines for AI agents contributing to the **DevUtils** repository.

## Formatting Rules
- **Go Code**: Always run `go fmt ./...` to format all Go source files before committing.
- **JavaScript & HTML**: Do not use automatic formatters (like Prettier). Keep the formatting clean but unformatted.

## Tech Stack Overview
- **Backend**: Go (using the **Gin** Web Framework).
- **Frontend**: HTML5, Vanilla CSS, and **HTMX** for dynamic content swaps without full page reloads.
- **Fonts**: Loaded via **Bunny Fonts** (GDPR compliant). 
  - Sans-serif: `Plus Jakarta Sans`
  - Monospace: `Lilex` (developer-focused font with coding ligatures)

## Architectural Guidelines
1. **Dynamic Port Binding**: The application must bind to the port defined by the `PORT` environment variable (injected by Render/Heroku) and fall back to `8080` in local environments.
2. **Rate Limiting**: All `/api/*` endpoints are protected by the custom Token-Bucket rate limiting middleware in `middleware/ratelimit.go`. Avoid introducing external packages for this.
3. **SPA Tab Navigation**: Tab switching is handled client-side using `switchTab()`. It updates the URL hash and pushes states to the browser history (`pushState`).
4. **HTMX Error Handling**: Always check for non-2xx status codes (like `429 Too Many Requests`) using the global HTMX error handler in `static/js/app.js`.
5. **Route Separation**: Maintain a clear separation of routes between HTMX layout fragments (`/api/htmx/*`) and REST API endpoints (`/api/v1/*`).
6. **API Documentation**: The interactive REST API specification is stored in `static/openapi.json` and rendered at `/docs` using custom-themed Swagger UI assets. Any updates to REST endpoints must be reflected in the schema file.
