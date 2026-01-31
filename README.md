# LinkShortener

Full-stack URL shortener with a Go backend, SQLite storage, and a React (Vite) frontend.

## Prerequisites

- Go 1.20+ (or compatible with your local toolchain)
- Node.js 18+ and npm

## Install

```bash
npm install
npm --prefix frontend install
```

## Run (both backend and frontend)

This uses the root `npm run dev` script to start both servers at once:

```bash
npm run dev
```

- Go API: `http://localhost:8080`
- Vite dev server: `http://localhost:5173`
- Vite proxies `/api` to the Go server (see `frontend/vite.config.js`)

## Run separately

Backend only:

```bash
go run ./cmd/server
```

Frontend only:

```bash
npm --prefix frontend run dev
```

## Configuration

Environment variables for the Go server:

- `LISTEN_ADDR` (default `:8080`)
- `BASE_URL` (default `http://localhost:8080`)
- `SQLITE_PATH` (default `data.db`)
- `GEOIP_ENDPOINT` (default `https://ipapi.co/%s/country/`)

Example:

```bash
$env:LISTEN_ADDR=":8080"
$env:BASE_URL="http://localhost:8080"
$env:SQLITE_PATH="data.db"
$env:GEOIP_ENDPOINT="https://ipapi.co/%s/country/"
go run ./cmd/server
```

## Functionality

- Shorten any `http`/`https` URL.
- Optional custom alias (3-30 chars, letters/numbers/`_`/`-`).
- Optional expiration date (defaults to 30 days).
- QR code generation for each short link.
- Redirect endpoint at `/{code}`.
- Analytics:
  - List all links with total/unique counts.
  - Lookup a specific code for click history summary, country breakdown, last access.
- Rate limiting: 10 requests per minute per IP on API routes.

## API Endpoints

- `POST /api/shorten`
  - Body: `{ "url": "...", "customAlias": "...", "expiresAt": "RFC3339" }`
- `GET /api/links`
- `GET /api/links/{code}`
- `GET /{code}` (redirect)

## Architecture

- `cmd/server/main.go`: server entry point, config/env wiring.
- `internal/api`: HTTP routes, handlers, DTOs, validation, rate limiting, QR, geo lookup.
- `internal/storage`: storage interface.
- `internal/storage/sqlite`: SQLite implementation and schema management.
- `internal/model`: Link and Click domain models.
- `internal/shortcode`: random short code generator.
- `frontend`: React UI with Vite dev server and API proxy.

## Data Storage

SQLite database (default `data.db`) with tables for:

- `links` (short code, original URL, created/expiry timestamps)
- `clicks` (timestamp, IP, country, user agent)
- `unique_ips` (per-link unique visitor tracking)
