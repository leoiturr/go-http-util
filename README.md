# Go HTTP Developer Utility Suite

A lightweight, high-performance web application built with **Go** and the **Gin Web Framework**, featuring essential tools for developers:
- **Base64 Converter**: Encode and decode strings to/from Base64.
- **GUID/UUID Generator**: Generate multiple unique identifiers (GUIDs) with custom options.
- **QR Code Generator**: Generate QR codes from text/URLs as Base64-encoded images.
- **Regex Playground**: Test RE2 patterns with flags, highlighted matches, capture groups, and common presets.

## Features

- **Gin Web Framework**: Fast routing and middleware support.
- **HTMX Integration**: Dynamic UI updates without full page reloads.
- **API Endpoints**: Easily accessible endpoints for programmatic utility calls.

## Prerequisites

- [Go](https://go.dev/) 1.18 or higher

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
   The application will start on `http://localhost:8080`.

## API Documentation

### Base64 Converter
- **Encode**: `POST /api/base64/encode`
- **Decode**: `POST /api/base64/decode`

### GUID Generator
- **Generate**: `POST /api/guid/generate`

### QR Code Generator
- **Generate**: `POST /api/qrcode/generate`

### Regex Playground
- **Test Pattern**: `POST /api/v1/regex/test`

## Deployment

Deployments are **git-push driven**: push to the `main` branch and the connected host rebuilds the container from `Dockerfile` and redeploys.

- The app binds the port from the `PORT` environment variable and falls back to `8080`.
- The container listens on port `8080` by default; set the platform's target port to `8080` (or inject `PORT`).
- A health check endpoint is available at `/health`.

## Webhook Tester

- Generated webhook URLs capture incoming requests **in memory only**.
- Captured requests are **reset on every deploy/restart** and expire after an idle TTL (default 24h, configurable via `WEBHOOK_TTL`).
- Global retention caps (`WEBHOOK_MAX_SESSIONS`, `WEBHOOK_MAX_BODY_MB`) keep the store bounded. See `AGENTS.md` for the full retention rules.

## License

This project is open-source and available under the [MIT License](LICENSE).
