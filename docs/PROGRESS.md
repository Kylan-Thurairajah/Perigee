# Perigee Progress

This is the lead/integrator handoff for current repository state. Update it when work is completed, validation changes, or a blocker is discovered. Agents should read this file and `AGENTS.md` before picking up work.

## Current checkpoint

- Updated: 2026-11-18
- Phase: M1 visual hook; implementation is present, acceptance verification remains open.
- Docker Desktop started and answered once, then its CLI/API requests began hanging. Compose startup remains blocked pending Docker recovery.
- `docker compose config --quiet` passes, but neither service is currently running.
- This worktree had no commit history before this initial baseline was staged.

## Verified and completed

- M0 repository foundations and fixture-backed API/web scaffold are present.
- The frontend and Go API have direct local startup instructions in `README.md`.
- Current-checkpoint frontend typecheck, lint, Vitest, production build, and Go API tests all pass; see the exact commands below.
- On 2026-11-18, `docker version` reported Docker Engine 29.9.2, Linux; `docker compose version` reported v5.6.1.
- Docker Desktop initially had no Linux engine. Starting `C:\Program Files\Docker\Docker\Docker Desktop.exe` brought up `desktop-linux`; `docker info --format '{{.ServerVersion}} {{.OSType}}'` returned `29.9.2 linux` once.
- `docker compose config --quiet` passes. Two `docker compose up --build -d` attempts hung without output; the first ran for several minutes. Later Docker CLI requests (`docker compose ps`, `docker version`, and `docker info`) also hung. A Docker Desktop shutdown/restart attempt did not restore the CLI during the validation window. The attempts were stopped; no Compose service was verified as started.
- Added root `.dockerignore` to exclude `.git`, Node dependencies, build output, caches, coverage, logs, and TypeScript build metadata from Docker build contexts.
- Direct validation passed: `npm run typecheck`, `npm run lint`, `npm run test` (3 tests), `npm run build`, and `go test ./...`.
- `make test` could not run because GNU Make is not installed; its constituent frontend and Go commands were run directly and passed.
- At the time of the endpoint check, neither `http://localhost:8080/api/health` nor `http://localhost:5173` was listening. Node emitted a `DEP0174` deprecation warning during npm commands; all commands still passed.
- The worktree was greenfield at task start; no preexisting commits or tracked project files needed preservation.

## Immediate next steps

1. Recover Docker Desktop and confirm `docker info` returns promptly. If it still hangs after restarting the PC, inspect Docker Desktop diagnostics before retrying Compose.
2. Run `docker compose up --build -d`; resolve any image/build/startup errors.
3. Verify `http://localhost:8080/api/health` and `http://localhost:5173`, then record results.
4. If Compose passes, verify container logs and `docker compose down` after smoke testing.
5. Continue M1 browser QA: interaction, target viewport layouts, theme behavior, and browser console.
6. Update this checkpoint with exact commands, outcomes, and any remaining blockers.

## Milestones

### M0: Repository audit and foundations

Status: complete for the startup baseline.

- React/Vite/TypeScript frontend and Go API scaffold exist.
- Docker Compose and direct local startup paths exist.
- Fixture-backed demo data is synthetic and should remain labeled as such.
- API contract and documentation are present.

### M1: Visual hook

Status: implementation present; browser/acceptance verification incomplete.

- The desktop app shell, voxel Earth scene, satellite/encounter presentation, and fixture UI are present in the repository.
- Do not mark M1 complete until its visual and interaction acceptance criteria below have been checked in a browser.
- Record any actual limitations rather than claiming unsupported performance.

Acceptance still to verify:

- Light-mode desktop app shell and intentional dark theme.
- Procedural voxel Earth; curated synthetic satellites; subtle featured encounter marker.
- Camera orbit/zoom and object selection with a contextual panel.
- Layouts at 1440x900, 1280x800, and 1920x1080.
- No console errors on the verified browser path.

### M2-M5

Status: not started. See the prioritized issue plan below and the master plan for scope.

## Prioritized remaining work

1. **Finish setup verification:** recover Docker Desktop, build and start Compose, then run health/UI smoke checks. Direct frontend/API type, lint, test, build, and Go test checks currently pass.
2. **Finish M1 QA:** browser acceptance checks and truthful QA record.
3. **Implement verified Go SGP4 and catalog ingestion:** choose a maintained library, document provenance, validate reference vectors, and address stale epochs/uncertainty.
4. **Implement computed screening/jobs:** broad-phase candidate search, refinement, trajectory sampling, bounded workers, and tests.
5. **Build synchronized encounter analysis:** ensure plotted minimum and time of closest approach agree with backend samples within documented tolerance.
6. **Implement bounded Maneuver Lab:** educational what-if only; document simplified model limitations and do not imply operational guidance.
7. **Add QA/demo assets:** startup/e2e smoke tests and truthful media after a functioning implementation.

## Active blockers and constraints

- No verified SGP4 implementation or reference-vector evidence exists.
- M1 browser visual QA is incomplete.
- Docker Desktop's CLI/API became unresponsive after initially starting. Compose config parses, but image build, service startup, and smoke checks remain unverified. Direct app endpoints were not listening during the last check.
- Do not claim aspirational performance/recall numbers without reproducible measurements.
- A minimum separation is miss distance, not collision probability.
- Do not present maneuvers as operational guidance or guaranteed collision avoidance.
- Do not add authentication, accounts, cloud infrastructure, Kubernetes, Kafka, notifications, billing, or distributed services.

## Agent handoff

- Lead/integrator: own `docs/PROGRESS.md`, `docs/ARCHITECTURE.md`, shared interfaces, contracts, build/CI, and integration gates.
- Product/UI: own `apps/web/src/app/`, `apps/web/src/components/`, `apps/web/src/features/`, `apps/web/src/styles/`, and `docs/DESIGN.md`.
- 3D visualization: own `apps/web/src/scene/` and scene-focused tests.
- Go engine/API: own `apps/api/`, backend tests/benchmarks, and backend validation/data-source documentation.
- QA/docs/demo: own test/e2e materials, `docs/DEMO_SCRIPT.md`, `docs/BENCHMARKS.md`, and README media plans.
- Respect the ownership boundaries in `AGENTS.md`; coordinate shared contract or interface changes before editing them.
