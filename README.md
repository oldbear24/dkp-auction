[![Go Report Card](https://goreportcard.com/badge/github.com/oldbear24/dkp-auction)](https://goreportcard.com/report/github.com/oldbear24/dkp-auction)
# DKP Auction House

DKP Auction House is a PocketBase application for managing auctions and bids in a DKP (Dragon Kill Points) system. The backend is written in Go and the UI lives in the `web/` SvelteKit app.

## Features

- Create and manage auctions
- Place bids on auctions
- Automatically finish auctions and determine winners
- Notify users about auction updates
- Scheduled maintenance tasks (auctions, user sync, token health checks)

## Requirements

- Go 1.22+ (backend)
- Node.js 18+ (frontend, optional for UI changes)

## Getting Started

### Backend (PocketBase)

1. Clone the repository:
    ```sh
    git clone https://github.com/oldbear24/dkp-auction.git
    cd dkp-auction
    ```

2. Install dependencies:
    ```sh
    go mod tidy
    ```

3. Run the application:
    ```sh
    go run . serve
    ```

PocketBase will create its local data directory and serve the API/UI once running. Static assets are served from `pb_public/` when present.

### Frontend (SvelteKit)

The UI lives under `web/` and can be run separately for local development.

```sh
cd web
npm install
npm run dev
```

Use `npm run build` to generate a production build.

## Migrations

Migrations are handled via the PocketBase migrate plugin. When running with `go run`, collection changes made in the admin dashboard will auto-generate migrations under `migrations/`.

## Project Structure

- `main.go` and `routes*.go`: backend entry point and route registration
- `migrations/`: PocketBase migrations
- `web/`: SvelteKit frontend

## Testing

```sh
go test ./...
```

## Build versioning

The backend and web UI can embed version metadata at build time. The Go binary reads
`VERSION`, `COMMIT`, and `BUILD_DATE` via linker flags, and the Svelte app reads the same
values via `VITE_APP_*` env vars in the Docker build.

Example Docker build:

```sh
docker build \
  --build-arg VERSION=$(git describe --tags --always) \
  --build-arg COMMIT=$(git rev-parse --short HEAD) \
  --build-arg BUILD_DATE=$(date -u +%Y-%m-%dT%H:%M:%SZ) \
  -t dkp-auction:$(git describe --tags --always) .
```
