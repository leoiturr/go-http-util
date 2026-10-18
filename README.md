# DevUtils — Go HTTP Developer Utility Suite

A lightweight, high-performance developer toolbox built with **Go** and the **Gin Web Framework**. Local-first, no build step, and no accounts — everything runs in the browser plus a small Go server.

## Tools

- **Base64 Converter**: Encode and decode strings to/from Base64.
- **GUID/UUID Generator**: Generate multiple unique identifiers (GUIDs) with custom options.
- **QR Code Generator**: Generate QR codes from text, URLs, or ready-made payloads (WiFi, vCard, MECARD, mailto, tel, SMS, geo). Customize error correction, module/background colors, quiet zone, and image size, then export as PNG or SVG.
- **JSON Formatter**: Prettify, minify, and validate JSON in CodeMirror 6 editors.
- **YAML Tools**: Convert JSON ⇄ YAML, prettify, and validate.
- **SQL Formatter**: Prettify and minify SQL with syntax highlighting.
- **URL Encoder**: Percent-encode/decode URLs and parse query parameters.
- **HTML Entity Encoder**: Escape and unescape HTML entities.
- **Regex Playground**: Test Go RE2 patterns with flags, highlighted matches, capture groups, and presets. Includes an opt-in JavaScript engine that runs patterns in a sandboxed, killable Web Worker with a 2-second timeout.
- **JWT Decoder**: Decode JWT headers and payloads.
- **Epoch Converter**: Convert between Unix timestamps and dates, with relative time formatting.
- **Cron Parser**: Explain cron expressions and preview upcoming runs.
- **Text Diff**: Instant client-side text comparison.
- **Webhook Tester**: Generate capture URLs and inspect incoming requests live.

## Features

- **Gin Web Framework**: Fast routing and middleware support.
- **HTMX 4**: Server-rendered fragments swap dynamically without full page reloads.
- **No Build Step**: Vanilla ES modules with an import map; CodeMirror 6 loads from CDN with a plain-textarea fallback.
- **Warm Paper Workbench UI**: Dark mode by default with a light paper variant, driven by `light-dark()` CSS tokens and a persisted theme toggle.
- **SPA Tab Navigation**: Client-side tab switching that syncs the URL hash and browser history.
- **Rate Limiting**: All `/api/*` endpoints are protected by a token-bucket middleware.
- **Interactive API Docs**: The OpenAPI spec (`static/openapi.json`) is rendered at `/docs` using the Scalar API Reference.

## Prerequisites

- [Go](https://go.dev/) 1.26 or higher

## Getting Started

1. Clone the repository:
   ```bash
   git clone https://github.com/leoiturr/go-http-util.git
   cd go-http-util
   ```

2. Download dependencies:
   ```bash
   go mod tidy
   ```

3. Run the development server:
   ```bash
   go run main.go
   ```
   The application starts on `http://localhost:8080` (or the port from `PORT`).

4. Run the test suite:
   ```bash
   go test ./...
   ```

## API Documentation

The JSON REST API lives under `/api/v1/*`; each endpoint also has an HTML-fragment twin under `/api/htmx/*` used by the UI. All are rate limited. The interactive reference is available at [`/docs`](http://localhost:8080/docs).

| Tool | Endpoints (prefix `/api/v1`) |
|---|---|
| Base64 | `POST /base64/encode`, `POST /base64/decode` |
| GUID | `POST /guid/generate` |
| QR Code | `POST /qrcode/generate` |
| JSON | `POST /json/prettify`, `POST /json/minify`, `POST /json/validate` |
| YAML | `POST /yaml/json2yaml`, `POST /yaml/yaml2json`, `POST /yaml/prettify`, `POST /yaml/validate` |
| SQL | `POST /sql/prettify`, `POST /sql/minify` |
| URL | `POST /url/encode`, `POST /url/decode`, `POST /url/parse` |
| Regex | `POST /regex/test` |
| JWT | `POST /jwt/decode` |
| Epoch | `POST /epoch/convert` |
| Cron | `POST /cron/parse` |
| HTML Entities | `POST /html-entity/encode`, `POST /html-entity/decode` |
| Text Diff | `POST /diff/text` |

A health check endpoint is available at `GET /health`.

## Deployment

Deployments are **git-push driven**: push to the `main` branch and the connected host rebuilds the container from `Dockerfile` and redeploys.

- The app binds the port from the `PORT` environment variable and falls back to `8080`.
- The container listens on port `8080` by default; set the platform's target port to `8080` (or inject `PORT`).
- A health check endpoint is available at `/health`.
- Live instances are linked from the sidebar footer on desktop: the Northflank deployment at <https://p01--devutils--p5rzzdynfjt7.code.run/> and the companion `go-http-util` service at <https://go-http-util.onrender.com>.

## Webhook Tester

- Generated webhook URLs capture incoming requests **in memory only**.
- Captured requests are **reset on every deploy/restart** and expire after an idle TTL (default 24h, configurable via `WEBHOOK_TTL`).
- Global retention caps (`WEBHOOK_MAX_SESSIONS`, `WEBHOOK_MAX_BODY_MB`) keep the store bounded. See `AGENTS.md` for the full retention rules.

## License

This project is open-source and available under the [MIT License](LICENSE).
