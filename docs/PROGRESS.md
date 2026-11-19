# Perigee Progress

This is the lead/integrator handoff for current repository state. Update it when work is completed, validation changes, or a blocker is discovered. Agents should read this file and `AGENTS.md` before picking up work.

## Current checkpoint

- Updated: 2026-11-18
- Phase: M8 screening. The Screening tab imports local OMM JSON, submits an asynchronous job, polls status, and pages through computed encounters with a propagated replay and an educational what-if. Encounter also offers a fictional collision clip. The screening benchmark baseline is in `docs/BENCHMARKS.md`.
- The README opens with the dark-mode fictional Peregrine A × Lyra B replay at 2×. The dark-mode stills show the impact, computed what-if, OMM setup, and screening results.
- This worktree had no commit history before this initial baseline was staged.

## Verified and completed

- M0 repository foundations and fixture-backed API/web scaffold are present.
- The frontend and Go API have direct local startup instructions in `README.md`.
- On 2026-11-18, `docker version` reported Docker Engine 29.9.2, Linux; `docker compose version` reported v5.6.1.
- Docker Desktop initially had no Linux engine. Starting `C:\Program Files\Docker\Docker\Docker Desktop.exe` brought up `desktop-linux`; `docker info --format '{{.ServerVersion}} {{.OSType}}'` returned `29.9.2 linux` once.
- On the latest retry, `docker info --format '{{.ServerVersion}} {{.OSType}}'` returned `29.9.2 linux` promptly.
- `docker compose up --build -d` now builds and starts `perigee-api-1` and `perigee-web-1`. The first real build exposed that the web image did not include root-level fixture JSON; `apps/web/Dockerfile` now copies `data/fixtures/` to `/data/fixtures/` in its build stage.
- Production browser validation exposed the canvas retaining the HTML default 300x150 size; the scene CSS now sizes the nested canvas to its container. The rebuilt container renders the voxel Earth correctly.
- `docker compose ps` confirms both services are Up. `curl.exe -fsS http://localhost:8080/api/health` returns `status: ok`, `fixtureMode: true`; `curl.exe -fsS http://localhost:5173/` returns the Perigee page title.
- Added root `.dockerignore` to exclude `.git`, Node dependencies, build output, caches, coverage, logs, and TypeScript build metadata from Docker build contexts.
- Integrated `OrbitalScene` into `SceneSlot`; the visible app now renders the voxel Earth and interactive Three.js canvas instead of the prior CSS-only preview. Theme palette is passed through to the scene.
- M1 frontend validation passed: `npm run typecheck`, `npm run lint`, `npm run test -- --reporter=dot` (3 tests), and `npm run build`.
- Go API validation `go test ./...` passed earlier in this checkpoint; API source has not changed since.
- `make test` could not run because GNU Make is not installed; its constituent frontend and Go commands were run directly and passed.
- Docker build's `npm ci` reports 6 dependency audit findings (3 moderate, 1 high, 2 critical); investigate separately before treating dependencies as security-cleared.
- Direct startup commands remain available: `Set-Location apps\api; go run .` and, in a second terminal, `Set-Location apps\web; npm run dev`.
- The Compose-backed page is open in the integrated browser. Browser smoke checks verified scene rendering, object-list selection, Encounter navigation, theme toggle, and camera drag/wheel input; no console/page errors occurred on a clean reload. At 1280x800, 1440x900, and 1920x1080, the page has no horizontal overflow and the scene canvas remains visible.
- Vite initially re-optimized Three.js dependencies during integration; the old dev page instance reported a transient hook error. A clean reload resolved it; the production container path has no browser errors in the verified interaction flow.
- M1 visual acceptance is complete: light Explorer and dark Encounter views, fixture objects and encounter marker, orbit/zoom controls, selection/context panel, desktop composition, and clean browser path were visually checked.
- M2 adds deterministic 31-sample circular fixture paths for the four fictional objects over T−30 to T+30 minutes. The featured pair's generated separation reaches the fixture's 742 m minimum at TCA; the scrubber, scene positions, and chart now share the same path samples. This is a visualization-only circular model, not SGP4 or operational trajectory data.
- The TCA is centered at the scrubber midpoint and chart minimum. Focused camera framing was widened so Encounter view shows the Earth and paths instead of cropping the globe.
- On the rebuilt Compose app, browser checks confirmed path count, slider/time movement, Encounter navigation, dark-theme toggle, and no console/page errors. At 1280x800, 1440x900, and 1920x1080, document width matches viewport width with no horizontal overflow.
- M3 replaces the previous arbitrary delta-v formula with a linearized along-track offset applied to the sampled path for Lyra B after the stated impulse time. The displayed TCA comparison and chart overlay are derived from that changed path; the dashed scene path appears only after Preview is clicked.
- M3 is explicitly a local fixture visualization, not orbital propagation, collision probability, or operational guidance.
- Browser verification confirmed Preview adds a fifth path, Reset removes it, slider changes affect the next preview, the chart shows baseline/what-if lines, and no browser errors or horizontal overflow occur at 1280x800, 1440x900, or 1920x1080.
- Current frontend checks pass: `npm run typecheck`, `npm run lint`, `npm run test -- --reporter=dot` (10 tests), and `npm run build`. The production build still reports the existing ~1.04 MB main-chunk advisory. `docker compose up --build -d` succeeds with the updated web app; API health returns `ok`.
- The current production build reports a main JavaScript chunk of about 1.04 MB, above Vite's 500 kB advisory threshold.
- M5 adds a TLE adapter at `apps/api/internal/orbit` using `github.com/akhenakh/sgp4` v1.0.5. Reference tests compare TEME position and velocity against Vallado SGP4-VER near-Earth vectors, including an absolute-time call with a non-UTC input zone.
- `go test ./...`, `go vet ./...` (from `apps/api`), and `docker compose build api` pass with the Go 1.23.4 minimum. The adapter is not yet connected to fixture API trajectories or maneuver results.
- M6 adds `ReadTLECatalog` to parse named or unnamed local TLE catalogs. It requires source/snapshot/retrieval metadata and an explicit max-age policy, rejects malformed/checksum-invalid, duplicate, and future-epoch records, marks old epochs stale, and reports unsupported propagation models without treating their positions as valid.
- M7 adds `ReadOMMCatalog` for single-object or array CCSDS-style JSON OMM input and `POST /api/catalog/omm` for bounded, provenance-aware validation. It preserves raw OMM fields, rejects malformed required elements, duplicate NORAD IDs, and future epochs, and flags stale records plus incomplete/unsupported propagation fields. The SGP4 OMM adapter matches the existing Vallado epoch reference vector. Imported data are returned for inspection only, not persisted or used by fixture catalog, screening, or encounter routes.
- Backend checks pass after M7: `go test ./... -count=1` and `go vet ./...` from `apps/api`; OpenAPI YAML parses successfully. `docker compose up --build -d api` passes, and a live POST to `/api/catalog/omm` returns a supported Vallado OMM record with an unknown source field preserved while `/api/health` remains `ok`.
- Added `docs/DEMO_SCRIPT.md` with the fixture and computed screening walkthroughs.
- Browser smoke check traversed Explorer → featured encounter → playback/pause → Maneuver Lab preview at 1920×1080; it verified the changed fixture path, what-if chart/metric, visible caveat, and no console/page errors.
- Current UI polish makes the globe surface itself voxel-forward (land, sea, cloud, and polar ice blocks), adds a softly breathing atmosphere, and gives spacecraft attitude and the encounter marker restrained motion. Scene motion pauses during interaction and honors `prefers-reduced-motion`.
- The latest spacecraft pass aligns each model's radial axis away from Earth, makes its central bus, antenna, and segmented blue solar arrays more legible, and adds 0.13 scene units of render-only radial clearance to orbit paths and their displayed positions. Fixture positions, chart measurements, and analysis values remain unchanged. The updated Compose web app loads at `http://localhost:5173`, and API health remains `ok`.
- Node emitted a `DEP0174` deprecation warning during npm commands; all commands still passed.
- The worktree was greenfield at task start; no preexisting commits or tracked project files needed preservation.

## Immediate next steps

1. Broaden the synthetic reference comparison with more varied orbit geometries and an independent verification source.

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

### M3: Maneuver Lab what-if

Status: complete for the bounded fixture-backed visual preview; no real maneuver propagation is implemented.

- The along-track control offsets Lyra B's sampled path after the displayed impulse time.
- Preview adds a dashed orange path and overlays the recomputed fixture separation profile; Reset restores the baseline.
- UI states that the preview is linearized, does not propagate orbital dynamics, and is not operational guidance.

### M4: Video-demo finish

Status: the README now includes the computed workflow recording. A separate fixture walkthrough and GIF remain optional.

- Run the opening, encounter replay, and what-if flow end-to-end.
- The inline README cut runs for four seconds from the loaded approach through impact and aftermath. It shows the fictional Peregrine A × Lyra B replay at 2×. The computed workflow is shown in separate stills.
- The collision camera centers on the actual event and positions itself outward from Earth at clip entry and impact. The frontend camera geometry regression covers far-side and polar positions. All 16 web tests, lint, and the Compose web build passed for that change.

### M5: SGP4 adapter and verification

Status: complete for a validated near-Earth TLE propagation adapter; initial local ingestion and OMM API support followed in M6-M7.

- The adapter uses `github.com/akhenakh/sgp4` v1.0.5 (Apache-2.0), which requires Go 1.23.4 or newer.
- Adapter results are labeled TEME with position in km and velocity in km/s.
- The dependency supports near-Earth SGP4 but not deep-space SDP4; unsupported inputs must remain explicit.
- The adapter is not yet connected to fixture API endpoints.
- `go test ./...`, `go vet ./...`, and `docker compose build api` pass.

### M6: Orbital record ingestion

Status: complete for local TLE parsing and validation. OMM ingestion and a validation endpoint were added in M7; durable catalog activation remains unimplemented.

- `ReadTLECatalog` parses local two-line or named three-line TLE records and requires explicit source metadata and an epoch-age policy.
- Malformed checksums, duplicate NORAD IDs, incomplete records, and future epochs return errors; stale records are labeled.
- Records outside the selected SGP4 package's supported model remain available with a limitation rather than being silently propagated.
- This helper is not yet reachable from an HTTP endpoint; no live catalog download is performed.

### M7: OMM ingestion and validation API

Status: complete for JSON OMM validation; the bounded in-memory snapshot cache was added with M8 screening.

- `ReadOMMCatalog` accepts a single OMM object or JSON array, requires source/snapshot/retrieval metadata and a maximum epoch age, and marks stale records.
- `POST /api/catalog/omm` accepts at most 4 MiB, validates records, and returns OMM elements, provenance, freshness, and propagation limitations.
- Only supported near-Earth SGP4-compatible records are marked propagation-supported; an OMM Vallado reference vector matches the TLE adapter.
- The endpoint does not modify fixture catalog state. Its cached snapshots can feed M8 screening and are cleared on API restart or LRU eviction.

### M8: Screening and backend-connected analysis

Status: in progress. The Screening tab connects the bounded job lifecycle to a paginated results table, synchronized SGP4 replay, and a backend-computed educational what-if. The recorded walkthrough and bounded performance baseline are available. Broader independent reference validation remains.

- `ScreenOMMCatalog` evaluates all pairs for up to 100 supported, fresh OMM records and at most 1441 time samples.
- Coarse relative-position segments select one candidate interval per pair; a bounded golden-section search refines that minimum and reports miss distance and relative speed.
- Results explicitly report the coarse time step, sample/pair counts, and limitations; the method may miss close approaches when the step is too large and does not compute collision probability.
- OMM imports are retained in a four-entry LRU memory cache, not persisted or merged into the fixture catalog. Screening accepts only fresh, supported snapshots of 2–100 records.
- The API exposes asynchronous create/status/results routes with two workers, an eight-job queue, and bounded run retention. The results endpoint returns deterministic pages and screening caveats.
- Focused API tests cover snapshot import/idempotency/conflict, bounded request/config rejection, asynchronous completion, and result mapping. The web replay/Maneuver Lab still use fixtures.
- Verification: `go test ./... -count=5`, `go vet ./...`, OpenAPI YAML parsing, and `docker compose up --build -d api` pass. A live Compose smoke test imported an OMM snapshot, completed a screening job, and retrieved its encounter results; `/api/health` remains `ok`.
- Go race testing was attempted but is unavailable in this environment because cgo is disabled and no GCC toolchain is installed.
- Web entry points: `apps/web/src/app/App.tsx` mounts `ScreeningPanel.tsx` and retains its state across workspace navigation. The panel validates local JSON, shows snapshot eligibility and caveats, polls queued/running jobs, retries failed status or results requests on demand, and appends cursor pages. Vite and Compose proxy same-origin `/api` requests. Reloading the page resets the tab state.
- Web validation: `npm run typecheck`, `npm run lint`, `npm test` (11 tests), and `npm run build` passed from `apps/web`. The application integration test uses mocked HTTP responses and covers import, polling, a failed results fetch and retry, pagination, and tab navigation. Vitest and the build needed execution outside the sandbox after esbuild hit `spawn EPERM`. The build still reports the main-chunk size advisory.
- Live wiring: `docker compose up --build -d` completed. Through the web proxy at `http://127.0.0.1:5173/api`, a 1,127,661-byte import based on the Vallado fixture in `apps/api/api_test.go` succeeded, a job completed, and the results endpoint returned one encounter with caveats. The same proxy returned `status: ok` and `fixtureMode: true` for health. The Compose proxy accepts request bodies up to the API's 4 MiB limit. The live check exercised HTTP directly, while browser interaction was exercised in the automated UI test.
- Reproduce the backend fixture check with `go test ./... -run TestScreeningJobRunsAgainstCachedOMMSnapshot` from `apps/api`. The fixture uses a historical retrieval date and screening window, so use the dates in that test when recreating the run.
- Replay: `GET /api/screenings/{runId}/encounters/{encounterId}` returns the screened summary, pair names, provenance, and 61 paired SGP4 samples at 10-second intervals. Positions are TEME km and velocities are m/s on the wire; separation is in meters. The sample window spans five minutes either side of the reported TCA and can extend beyond the original screening interval. The retained run holds the orbital elements needed for replay after snapshot cache eviction.
- UI replay: `ScreeningReplay.tsx` opens from a Screening table row and starts on the reported TCA sample. The timeline, two-object scene, sampled separation readout, and chart share the same sample index. The built-in Encounter and Maneuver Lab tabs remain fixture-based.
- Replay verification: `go test ./... -count=3`, `go vet ./...`, OpenAPI YAML parsing, `npm run typecheck`, `npm run lint`, `npm test` (11 tests), and `npm run build` passed. Go's default cache returned access denied, so validation used a temporary `GOCACHE`. The API test confirms the cache eviction and returns an error for an unknown encounter ID. The web test verifies that moving the scrubber advances both the scene time and sample separation. The bundle-size advisory persists.
- Live replay verification: `docker compose up --build -d` succeeded. A headless Edge run using the two-record Vallado-derived fixture and the live Compose API imported the catalog, completed screening, opened the WebGL replay, and advanced playback without uncaught page exceptions. A direct check through the web proxy verified the TEME frame, 61 samples, matching center time and miss distance, and healthy fixture mode.
- Screened what-if: `POST /api/screenings/{runId}/encounters/{encounterId}/what-if` accepts a signed impulse from -0.5 to 0.5 m/s. The time and target are fixed to the replay start and object A. The backend adds the position and velocity difference between two central-Earth-gravity integrations to object A's SGP4 baseline. The chart compares separations on the original 61-sample grid. The run and standalone fixture Maneuver Lab remain independent of the preview.
- What-if verification: `go test ./... -count=3`, `go vet ./...`, OpenAPI YAML parsing, web typecheck and lint, all 11 Vitest tests, and the production build passed. The first API integration test surfaced a 405 from the top-level method guard. Screening run routes now dispatch to their own method checks, and the focused tests and full Go suite passed after that fix.
- Live what-if verification: `docker compose up --build -d` passed. Headless Edge imported the test fixture against the live Compose API, submitted a screen, opened replay, applied the preview, checked the WebGL and chart overlays plus sampled minimums, and cleared the preview without uncaught page exceptions.

## Latest development

- **Fictional collision:** Encounter offers a scripted zero-separation pair at the midpoint of its own four-minute simulation window. At 1× the presentation runs in an eight-second loop, with impact at four seconds and a two-second burst. The effect and satellite visibility depend on replay position. Pausing freezes the effect, seeking backward restores the pair, and the existing featured close approach remains available. Reduced-motion mode uses a static burst indicator during the impact interval.
- **Replay validation:** all 15 Vitest tests, typecheck, lint, and the Compose production build passed. Fake-clock tests cover impact at 0.5×, 1×, and 2×, each loop, pausing, scrubbing, changing speed mid-approach, and restarting. A live Compose check in headless Edge observed impact at 8.049 s, 4.048 s, and 2.048 s respectively in one run. A paused WebGL screenshot confirmed the burst at the encounter position. This is presentation timing for the scripted clip.
- **Screening benchmarks:** added `BenchmarkScreenOMMCatalog` and executed the three-repeat synthetic workload baseline in `docs/BENCHMARKS.md`. The benchmark reports timing and allocation volume for the bounded engine without measuring recall. The Go command emitted a telemetry cache permission warning and completed successfully.
- **Synthetic reference comparison:** `TestScreeningPairRecall` compares 66 pairs against a 0.5-second dense SGP4 reference, checks threshold classification at 1-second resolution, and locks recovery at 5 and 60 seconds. With the 5 km threshold, 30- and 60-minute steps missed both reference pairs. The screening form now explains the coarse-step tradeoff before submission. The comparison and limits are in `docs/BENCHMARKS.md`.
- **Offline demo fixture:** `data/reference/screening-demo.omm.json` is ready to upload using the dates in `docs/DEMO_SCRIPT.md`. The API integration test imports the same file and checks screening, detail, and what-if behavior.
- **Verification:** `go test ./... -count=3`, `go vet ./...`, web typecheck and lint, all 15 Vitest tests, and the production build passed. A live HTTP smoke check through the Compose web proxy imported the packaged fixture, returned one encounter, and retrieved both 61-sample trajectories for baseline and what-if. The existing bundle-size advisory remains.

## Prioritized remaining work

1. **Finish independent reference validation:** the synthetic reference comparison shares the SGP4 adapter with the engine.

## Active blockers and constraints

- The current SGP4 dependency supports selected near-Earth TLE/OMM elements; deep-space SDP4 remains unsupported.
- The Maneuver Lab is still a linearized fixture preview, not propagation; keep that limitation visible and do not call it a maneuver recommendation.
- No current Docker blocker: Compose build and smoke checks pass. The npm build reports dependency audit findings and a bundle-size advisory; neither prevented the build.
- Do not claim aspirational performance/recall numbers without reproducible measurements.
- A minimum separation is miss distance, not collision probability.
- Do not present maneuvers as operational guidance or guaranteed collision avoidance.
- Do not add authentication, accounts, cloud infrastructure, Kubernetes, Kafka, notifications, billing, or distributed services.

## Agent handoff

- The master plan is a local planning prompt. Its historical version is readable with `git show 5773da1:PERIGEE_CODEX_MASTER_PLAN.md` when a milestone gate needs reference.
- Lead/integrator: own `docs/PROGRESS.md`, `docs/ARCHITECTURE.md`, shared interfaces, contracts, build/CI, and integration gates.
- Product/UI: own `apps/web/src/app/`, `apps/web/src/components/`, `apps/web/src/features/`, `apps/web/src/styles/`, and `docs/DESIGN.md`.
- 3D visualization: own `apps/web/src/scene/` and scene-focused tests.
- Go engine/API: own `apps/api/`, backend tests/benchmarks, and backend validation/data-source documentation.
- QA/docs/demo: own test/e2e materials, `docs/DEMO_SCRIPT.md`, `docs/BENCHMARKS.md`, and README media plans.
- Respect the ownership boundaries in `AGENTS.md`; coordinate shared contract or interface changes before editing them.
