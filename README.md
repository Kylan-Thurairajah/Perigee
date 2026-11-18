# Perigee

### A clearer view of what's happening in orbit.

Perigee is a local-first orbital traffic visualization and conjunction-screening portfolio app. It combines a voxel Earth interface with deterministic synthetic fixtures and a Go API.

## Run locally

Requirements:

- Node.js 18 or newer
- Go 1.22 or newer

Start the API:

```powershell
Set-Location apps/api
go run main.go api.go
```

Start the web app in a second terminal:

```powershell
Set-Location apps/web
npm install
npm run dev
```

Open `http://localhost:5173`. The API health endpoint is available at `http://localhost:8080/api/health`.

Optional Docker Desktop setup:

```text
docker compose up --build
```

Open `http://localhost:5173` after the containers report healthy or become reachable.

## Verify

```powershell
Set-Location apps/web
npm run typecheck
npm run lint
npm run test
npm run build

Set-Location ../api
go test ./...
```

## Current data and limits

The M0/M1 demo uses clearly labeled synthetic fixtures. It does not represent operational ephemerides, real conjunction predictions, or collision probability. The Maneuver Lab is an educational what-if preview, not physical orbit propagation or operational guidance.
