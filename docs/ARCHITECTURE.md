# Architecture

This document explains the high-level architecture of the LinkShortener project, how requests flow through the system, and where core responsibilities live in the codebase.

## System Overview

LinkShortener is a full-stack URL shortener with:
- Go HTTP API server
- SQLite persistence
- React (Vite) frontend UI

The backend exposes JSON API endpoints for shortening URLs and retrieving analytics, plus a public redirect endpoint. The frontend calls the API and renders the management UI.

## High-Level Components

1) Backend (Go)
- Entry point: `cmd/server/main.go`
- HTTP server + routes: `internal/api`
- Storage interface: `internal/storage`
- SQLite implementation: `internal/storage/sqlite`
- Domain models: `internal/model`
- Short-code generator: `internal/shortcode`

2) Database (SQLite)
- Physical file: `data.db` (configurable)
- Schema managed in `internal/storage/sqlite/sqlite.go`

3) Frontend (React + Vite)
- App entry: `frontend/src/main.jsx`
- UI + API calls: `frontend/src/App.jsx` and `frontend/src/components/*`
- Dev proxy: `frontend/vite.config.js` (proxies `/api` to the Go server)

## Request Flow

### 1) Create Short Link (POST /api/shorten)
1. `internal/api` validates JSON payload and URL.
2. A short code is resolved:
   - If a custom alias is provided, it is validated and checked for uniqueness.
   - Otherwise a random code is generated (`internal/shortcode`).
3. Expiration is parsed (defaults to now + 30 days).
4. Link is stored in the `storage.Store` implementation.
5. A short URL is built using `BASE_URL`, and a QR code is generated.
6. JSON response includes code, short URL, original URL, expiration, and QR.

### 2) Redirect (GET /{code})
1. `internal/api` looks up the link by code.
2. Expiration is checked; expired links return 410.
3. A click record is created (timestamp, IP, country, user agent).
   - Country is fetched via `GEOIP_ENDPOINT` and cached in-memory.
4. Click is recorded via `storage.Store.RecordClick`.
5. Server returns 302 redirect to the original URL.

### 3) Analytics
- `GET /api/links`: returns overview list with total/unique counts.
- `GET /api/links/{code}`: returns link details with per-country counts,
  last access time, and QR code.

## Data Model (SQLite)

Tables are created on startup if missing:
- `links`: code, original URL, created time, expires time
- `clicks`: per-click data (timestamp, IP, country, user agent)
- `unique_ips`: link-to-IP pairs for unique visitor counts

Foreign keys enforce cascading deletes from `links` to `clicks` and `unique_ips`.

## Storage Abstraction

`internal/storage/Store` is the primary boundary between API logic and persistence. It supports:
- `Save`, `Upsert`, `Get`, `List`, `RecordClick`

The SQLite implementation (`internal/storage/sqlite`) handles:
- Schema creation
- Queries and transactions
- Unique constraint translation to domain errors

## Configuration

Environment variables (with defaults):
- `LISTEN_ADDR` (default `:8080`)
- `BASE_URL` (default `http://localhost:8080`)
- `SQLITE_PATH` (default `data.db`)
- `GEOIP_ENDPOINT` (default `https://ipapi.co/%s/country/`)

## Key Design Decisions

- Storage is abstracted behind `internal/storage/Store` to allow future implementations.
- Short-code generation uses crypto-random selection for unpredictability.
- Rate limiting is enforced in memory (10 req/min per IP) at the API layer.
- Country detection is cached in memory to reduce external calls.
- QR codes are generated on demand as a data URL.

## Module Map

- `cmd/server/main.go`: configuration + bootstrapping
- `internal/api/server.go`: routes + middleware
- `internal/api/handlers.go`: request handlers
- `internal/api/helpers.go`: validation, QR, geo lookup, rate limiting
- `internal/model/link.go`: domain models
- `internal/storage/storage.go`: store interface + errors
- `internal/storage/sqlite/sqlite.go`: SQLite store + schema
- `internal/shortcode/generator.go`: short code generation
- `frontend/src/App.jsx`: UI, form handling, API calls
- `frontend/src/components/Analytics.jsx`: analytics UI
