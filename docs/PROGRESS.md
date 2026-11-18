# Perigee Progress

This is the lead/integrator handoff for current repository state. Update it when work is completed, validation changes, or a blocker is discovered. Agents should read this file and `AGENTS.md` before picking up work.

## Current checkpoint

- Updated: 2026-11-18
- Phase: M2 demo-first encounter replay; the interactive scene, fixture tracks, timeline, and separation chart now follow one deterministic synthetic sample set.
- Docker Desktop is responsive again; both Compose services build and run.
- The latest Compose-backed app is open in the integrated browser at `http://localhost:5173`, on Encounter analysis in dark theme.
- The initial setup baseline is pushed on `main`; all changes since then are uncommitted and unpushed.
- This worktree had no commit history before this initial baseline was staged.

## Verified and completed

- M0 repository foundations and fixture-backed API/web scaffold are present.
- The frontend and Go API have direct local startup instructions in `README.md`.
- Current-checkpoint frontend typecheck, lint, Vitest, production build, and Go API tests all pass; see the exact commands below.
- On 2026-11-18, `docker version` reported Docker Engine 29.9.2, Linux; `docker compose version` reported v5.6.1.
- Docker Desktop initially had no Linux engine. Starting `C:\Program Files\Docker\Docker\Docker Desktop.exe` brought up `desktop-linux`; `docker info --format '{{.ServerVersion}} {{.OSType}}'` returned `29.9.2 linux` once.
- On the latest retry, `docker info --format '{{.ServerVersion}} {{.OSType}}'` returned `29.9.2 linux` promptly.
- `docker compose up --build -d` now builds and starts `perigee-api-1` and `perigee-web-1`. The first real build exposed that the web image did not include root-level fixture JSON; `apps/web/Dockerfile` now copies `data/fixtures/` to `/data/fixtures/` in its build stage.
- Production browser validation exposed the canvas retaining the HTML default 300x150 size; the scene CSS now sizes the nested canvas to its container. The rebuilt container renders the voxel Earth correctly.
- `docker compose ps` confirms both services are Up. `curl.exe -fsS http://localhost:8080/api/health` returns `status: ok`, `fixtureMode: true`; `curl.exe -fsS http://localhost:5173/` returns the Perigee page title.
- Added root `.dockerignore` to exclude `.git`, Node dependencies, build output, caches, coverage, logs, and TypeScript build metadata from Docker build contexts.
- Integrated `OrbitalScene` into `SceneSlot`; the visible app now renders the voxel Earth and interactive Three.js canvas instead of the prior CSS-only preview. Theme palette is passed through to the scene.
- Current frontend validation passed: `npm run typecheck`, `npm run lint`, `npm run test -- --reporter=dot` (3 tests), and `npm run build`. The production build succeeds with a Vite warning that the main JavaScript chunk is about 1.04 MB, above its 500 kB advisory threshold.
- Go API validation `go test ./...` passed earlier in this checkpoint; API source has not changed since.
- `make test` could not run because GNU Make is not installed; its constituent frontend and Go commands were run directly and passed.
- Docker build's `npm ci` reports 6 dependency audit findings (3 moderate, 1 high, 2 critical); investigate separately before treating dependencies as security-cleared.
- Direct startup commands remain available: `Set-Location apps\api; go run main.go api.go` and, in a second terminal, `Set-Location apps\web; npm run dev`.
- The Compose-backed page is open in the integrated browser. Browser smoke checks verified scene rendering, object-list selection, Encounter navigation, theme toggle, and camera drag/wheel input; no console/page errors occurred on a clean reload. At 1280x800, 1440x900, and 1920x1080, the page has no horizontal overflow and the scene canvas remains visible.
- Vite initially re-optimized Three.js dependencies during integration; the old dev page instance reported a transient hook error. A clean reload resolved it; the production container path has no browser errors in the verified interaction flow.
- M1 visual acceptance is complete: light Explorer and dark Encounter views, fixture objects and encounter marker, orbit/zoom controls, selection/context panel, desktop composition, and clean browser path were visually checked.
- M2 adds deterministic 31-sample circular fixture paths for the four fictional objects over T−30 to T+30 minutes. The featured pair's generated separation reaches the fixture's 742 m minimum at TCA; the scrubber, scene positions, and chart now share the same path samples. This is a visualization-only circular model, not SGP4 or operational trajectory data.
- The TCA is centered at the scrubber midpoint and chart minimum. Focused camera framing was widened so Encounter view shows the Earth and paths instead of cropping the globe.
- On the rebuilt Compose app, browser checks confirmed path count, slider/time movement, Encounter navigation, dark-theme toggle, and no console/page errors. At 1280x800, 1440x900, and 1920x1080, document width matches viewport width with no horizontal overflow.
- Current frontend checks pass: `npm run typecheck`, `npm run lint`, `npm run test -- --reporter=dot` (4 tests), and `npm run build`. `docker compose up --build -d` also succeeds with the updated web app; API health and web title checks passed earlier in this checkpoint.
- Node emitted a `DEP0174` deprecation warning during npm commands; all commands still passed.
- The worktree was greenfield at task start; no preexisting commits or tracked project files needed preservation.

## Immediate next steps

1. M3 demo-first Maneuver Lab: replace the current arbitrary placeholder metric adjustment with a clearly labeled local what-if that visibly overlays a changed fixture path and updates the separation profile. Do not claim real propagation or operational guidance.
2. M4 video packaging: run the short demo path end-to-end, add a concise demo script and capture directions, then record the 35–60 second walkthrough/GIF from the functioning app.
3. Keep Compose services available during the demo; use `docker compose down` when finished.
4. After the video demo is captured, prioritize verified SGP4, catalog ingestion, screening jobs, and backend-connected analysis/maneuver behavior.
5. Defer bundle splitting and dependency audit remediation unless they block the demo or a safe, quick fix is identified.

## Milestones

### M0: Repository audit and foundations

Status: complete for the startup baseline.

- React/Vite/TypeScript frontend and Go API scaffold exist.
- Docker Compose and direct local startup paths exist.
- Fixture-backed demo data is synthetic and should remain labeled as such.
- API contract and documentation are present.

### M1: Visual hook

Status: complete for the fixture-backed visual slice.

- The desktop app shell, integrated voxel Earth scene, satellite/encounter presentation, and fixture UI are present.
- Browser checks cover canvas rendering, list selection, encounter navigation, theme toggle, camera gestures, and the 1280x800, 1440x900, and 1920x1080 desktop viewports.
- Fixture tracks are visibly illustrative and explicitly marked not to scale.

### M2: Encounter replay (demo-first frontend slice)

Status: complete for the featured synthetic encounter; backend propagation and screening are not implemented.

- Scrubber time, satellite positions, encounter focus, and separation chart use the same deterministic fixture samples.
- TCA and minimum separation align at the center of the replay window.
- Clearly disclose that the paths are simplified circular fixture visuals, not SGP4.

### M3-M4: Video-demo finish

Status: next; prioritize these visible steps before deeper backend work.

- Replace the Maneuver Lab's placeholder output with an honest, visibly updated illustrative comparison.
- Run the opening, encounter replay, and what-if flow; create the video/GIF and demo script.

Rough estimate: about 2–3 focused implementation/QA sessions to reach a credible fixture-backed video demo, assuming the scope stays focused and no visual blocker appears. This is not a completion estimate for the full engineering MVP.

### M5-M8: Engineering MVP

Status: not started; required before presenting the project as an implemented screening engine.

- Add a maintained SGP4 implementation and reference-vector validation.
- Ingest/validate orbital catalog records and implement screening jobs/results.
- Connect replay and maneuver experiences to backend trajectory samples and document model limitations.
- Finish integration/e2e validation, README evidence, and release/demo packaging.

## Prioritized remaining work

1. **Video demo, fast path:** finish the bounded fixture-backed Maneuver Lab, record the 35–60 second demo, and add demo script/capture instructions.
2. **Verified orbital inputs:** choose a maintained Go SGP4 library, document provenance, validate reference vectors, and handle stale epochs/uncertainty.
3. **Computed screening:** broad-phase candidate search, close-approach refinement, trajectory samples, bounded jobs, and API tests.
4. **Backend-connected analysis:** replace local visual fixture paths with API samples and synchronize TCA, minimum separation, and chart values.
5. **Backend-connected what-if:** recompute a limited educational scenario from supported trajectory data with explicit model limitations.
6. **Finish project evidence:** e2e demo path, reproducible README/media, measured performance only when actually benchmarked.

## Active blockers and constraints

- No verified SGP4 implementation or reference-vector evidence exists.
- The current what-if number is still a local placeholder and is not propagation; replace it before presenting the maneuver flow as a recomputation.
- No current Docker blocker: Compose build and smoke checks pass. The npm build reports dependency audit findings and a bundle-size advisory; neither prevented the build.
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
