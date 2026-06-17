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
- Dynamic actions (generating GUIDs, encoding/decoding Base64, creating QR codes) are triggered via `hx-post`.
- Non-2xx status codes (like the rate limiter's `429`) are caught globally in `app.js` using the `htmx:responseError` listener and shown as clean toast notifications.

### Visual Styling & Theme
- **Theme**: Premium glowing dark mode theme, featuring neon accent gradients and card drop shadows.
- **Typography**: 
  - Sans-serif: `Plus Jakarta Sans` (sleek, geometric, modern look).
  - Monospace: `Lilex` (developer font with coding ligatures).
  - Loaded via **Bunny Fonts** (GDPR compliant).
