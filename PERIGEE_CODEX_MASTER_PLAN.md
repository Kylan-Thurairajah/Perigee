# PERIGEE — Codex Master Build Plan

> **Mission:** Build a visually distinctive, technically credible, desktop-first orbital conjunction screening web app that can be demonstrated locally and showcased beautifully on GitHub.
>
> **Primary audience:** software engineering recruiters, hiring managers, and engineers evaluating a portfolio project. The product must stand on its own as useful, engaging software, not a collection of buzzwords.
>
> **Status:** implementation plan; no implementation or benchmark claims should be treated as complete without verification.

## 0. Operating instructions for Codex (READ FIRST)

You are the **lead/orchestrator agent** for Perigee. **Use parallel specialized subagents if the current Codex environment supports them.** Do not assume they exist: if subagent support is unavailable, implement the same phases sequentially and keep the same ownership boundaries.

1. Inspect the existing repository before modifying anything. Respect existing working code and APIs; do not rewrite an implemented engine to match this document if it can be adapted.
2. Maintain a `docs/PROGRESS.md` file with current phase, completed features, active blockers, decisions, and next tasks; update it at each integration checkpoint.
3. Start by writing a concrete implementation backlog and shared TypeScript API contracts. Then scaffold the apps and produce the first functional, visual vertical slice.
4. Use small, reviewable, testable changes. Run formatting, lint, typechecks and unit tests whenever a phase is integrated.
5. Prioritize the UI experience, demo story, and README quality **without faking** the underlying calculations or benchmarks.
6. Do **not** build authentication, accounts, cloud infrastructure, Kubernetes, Kafka, distributed microservices, AI chat, notification systems, or SaaS billing. This is a standalone local-first portfolio application.
7. **No automatic remote-data dependency for demos.** Ship versioned, reproducible, attributed fixture data so the application and demo work offline after installation. Add optional live CelesTrak import later.
8. There is **no requirement to deploy the app**. The README must offer local setup, a self-contained video/GIF and screenshots. Optional public hosting can be added only after MVP.
9. Prefer deterministic and accurate behavior over vague astrophysics claims. Clearly distinguish conjunction detection from collision-probability estimation. Do not call a miss distance “collision probability” or label an event as an inevitable collision.
10. **Do not report or advertise 25K objects, 14 seconds, 120x, 96% recall or 2,000 reference events unless measured and reproducibly substantiated.** These figures are aspirational/example figures originating from an earlier resume description; check whether existing code actually demonstrates them.
11. Ask the human only for a truly blocking choice. Otherwise use the product decisions defined here and proceed.

## 1. Product positioning

**Name:** Perigee

**Main tagline:** A clearer view of what's happening in orbit.

**Short description:** Perigee is a local-first orbital traffic visualization and conjunction screening platform. A parallel Go engine propagates cataloged Earth-orbiting objects and identifies predicted close approaches. An immersive, stylized 3D web UI lets users explore satellites, replay encounters, inspect separation-over-time graphs, and test hypothetical maneuvers.

**Value proposition for reviewers:** A playful voxel universe on the surface; serious computational engineering underneath.

**Three wow moments:**

1. **Opening scene:** A crisp, airy light-mode desktop interface containing a surprisingly beautiful voxel Earth with a handful of softly animated satellites and elegant orbit tracks. An orange encounter marker quietly suggests an interesting result.
2. **Click-to-replay:** Clicking the orange marker smoothly transitions the camera from orbital view to a focused two-satellite encounter. A timeline scrubber and separation graph move in synchronization. At the time of closest approach, the graph's low point glows orange and the camera lands on the near pass.
3. **What-if recomputation:** An experiment changes a small tangential delta-v. Perigee recomputes the future trajectory and overlays the original and altered encounters. The user sees updated separation metrics and can reset the scenario.

**Avoid:** enterprise mission control clichés, fake streaming dashboards, hundreds of visualized satellites, deceptive UI instrumentation, rainbow colors, overly glossy glassmorphism, scientifically misleading orbital imagery presented as precision, and a README dominated by implementation jargon.

## 2. Scope and definition of done

### MVP: all required

- Desktop-first React + TypeScript web app with default light theme and persistent dark-mode toggle (local storage is fine).
- Three main views: Explorer, Encounter Analysis, and Maneuver Lab.
- Distinctive low-poly / voxel art direction; an appealing real-time Earth mesh and stylized satellites constructed from reusable geometry (procedural geometry is preferred to downloading opaque assets).
- Camera orbit/zoom interaction, select an object, highlight trajectory, hover/click encounter markers, accessible buttons and labels.
- Fixture-backed demo with at least 3 coherent, reproducible encounter scenarios, including a clearly featured one that is visually appealing.
- Go backend that ingests supported orbital element records, propagates states with a **verified SGP4 implementation**, runs coarse screening, refines close approaches, and exposes results via typed HTTP JSON endpoints.
- Screening jobs are asynchronous at the API level: request -> status/progress -> results, using in-process jobs for MVP. No external queue required.
- Results show object IDs/names, predicted time of closest approach (TCA), minimum separation, relative speed, source epoch, and a data-quality/caveat label.
- Synchronized encounter replay: camera, satellites, timestamp, timeline, and separation graph all derive from **the same sampled trajectory series**.
- Maneuver Lab supports one honest, limited educational scenario: an instantaneous along-track delta-v at a chosen time before the encounter. Backend recomputes the updated future trajectory. Identify that this is a simplified what-if experiment, not operational guidance.
- Deterministic fixture mode; optional live import is a later phase.
- Local-first setup with `docker compose up --build` and clear alternative development commands; provide reproducible sample data.
- Tests for propagation reference cases, encounter calculations, backend API, frontend behavior, and a smoke end-to-end demo path.
- Strong README with top-fold video or GIF, screenshots, verified benchmarks, architecture diagram and one-command local setup.

### Stretch only after MVP is polished

- More visualization modes (debris cloud, orbit class toggles).
- Additional simulation parameter controls.
- Larger catalog import and multi-hour offline precompute.
- Comparison of runs from different orbital element epochs.
- WebGL performance tuning for dense object sets.
- Optional deployment or hosted lightweight viewer.

### Explicit non-goals

- Operational collision avoidance, accurate probability of collision (Pc), uncertainty covariance propagation, spacecraft control or maneuver optimization, guaranteed real-time satellite tracking, precise flight dynamics UI.
- Accounts, payments, teams, messaging, notifications, live operational alerts.
- Complex distributed architecture or production infrastructure.
- Spinning up costly cloud resources for a portfolio screenshot.

## 3. Visual direction: crisp Apple-like web app meets voxel diorama

**Important:** This is a **WEB APP FOR DESKTOP**, not a phone app. Borrow Apple's product qualities (clarity, spacing, deliberate motion, high legibility, native-feeling controls), not iOS phone chrome, Dynamic Island, app icons, or bottom tab bars.

### Design language

- Spatial composition: oversized art-directed 3D scene, wide breathing room, small amounts of purposeful chrome, controlled card density.
- Geometry: blocky, slightly beveled voxel Earth and cube-based satellites; softened shadows, not pixelated text or retro pixel fonts.
- Typography: `Inter` or modern system font stack; restrained weights, tight typographic hierarchy, numeric values in tabular numerals.
- Components: softly rounded rectangles, subtle hairline borders, barely visible shadows, minimal blur; avoid heavy translucent panels that obscure details.
- Animations: 180–350ms ease-out for panels; camera moves ~700–1200ms where appropriate; subtle orbit motion; respect `prefers-reduced-motion`.
- Themes: LIGHT IS DEFAULT. A compact sun/moon toggle switches to a tasteful dark counterpart, preserves state in localStorage, and never changes semantic accent colors unpredictably.
- Scientific charting: precision and clear labels even when the 3D world is toy-like. Use semantic units, timestamps, reference-data dates and data provenance.

### Recommended palette (tune visually)

Light:

- Canvas: `#F7F6F2` warm porcelain
- Surface: `#FFFFFF` / `#FCFBF8`
- Secondary surface: `#EFEEE9`
- Text primary: `#263747` graphite blue
- Text muted: `#748597`
- Border: `#DEE3E6`
- Orbit blue: `#829FB9` dusty slate blue
- Selected blue: `#5F88AD`
- Warning apricot: `#EBA16F` (stronger hover `#DF8753`)
- Warning pale background: `#FFF2E8`
- Globe sea: `#5D94BB`; land: `#8BB39C`; clouds: near white (keep restrained)

Dark:

- Canvas: `#151F29`
- Surface: `#202C37`
- Secondary surface: `#263543`
- Text primary: `#F4F5F2`
- Text muted: `#9FADBA`
- Border: `#344451`
- Orbit blue: `#86A9C5`
- Warning: `#EBA16F`

**Warning colors:** no glaring reds. Orange conveys "investigate this close approach" rather than a verified catastrophic collision. Accessible iconography and labels must communicate meaning without color alone.

### Timeline / scrubber (critical interaction)

The old large YouTube-style horizontal slider is rejected. Create a bespoke **thin precision timeline**:

- A 2px rounded horizontal track, slightly darker inactive portion, precise tick markers, and a small ~11–14px thumb, enlarging subtly on hover or keyboard focus.
- TCA is represented by a single orange diamond/dot on the track with a floating timestamp tooltip on hover.
- Play/Pause and playback speed (`0.5x`, `1x`, `2x`) are separate small controls to the left; current UTC and `T− / T+` offset to the right.
- Optional faint graph-underlay or distance sparkline; avoid dense decorative tick marks or a giant pill backing.
- Keyboard Left/Right seeks, Space toggles playback when the timeline is focused, reduced motion preferences are respected, and screen readers get accessible labels.
- Do not call this a scrollbar in code; it is a **timeline scrubber**, distinct from page scrolling. Make the actual page scrollbar thin and understated if present.

### Desktop layout concepts (implement first concept; create room to refine)

**A. Immersive Canvas — default Explorer.** Horizontal app shell: slim top navigation (~64px), large Earth canvas occupying 65–75% of the usable width, slim contextual right drawer (~320px) that can collapse; minimal bottom-aligned floating time control. No large permanent left sidebar.

**B. Encounter Focus — analysis view.** Main 3D viewport ~60% and a clean right analytics column ~40%; right column includes the encounter identity, 3 key metric tiles, separation curve, object metadata, and small "Open in Maneuver Lab" action. Timeline spans below main viewport, not the entire browser unnecessarily.

**C. Experiment Workspace — maneuver view.** 3D before/after trajectory comparison on left (original line in slate, altered line in muted orange), parameters and recomputed metrics in right panel, compact comparison summary at top. Restrain scope to one maneuver type.

### Interaction microdetails

- Main globe rotates very subtly while idle; stop or slow it immediately on interaction.
- Satellite selection uses a clean outline/halo rather than huge glowing rings.
- Deselect returns the camera smoothly to the default overview.
- Clicking an orange encounter marker zooms the camera toward the actual two-object paths while the rest of the world fades slightly.
- Buttons feel physical but not bubbly: tiny scale and shadow change on hover, generous click targets, focus states.
- Neutral loading states are simple and quiet, not spinning sci-fi HUD graphics.
- Empty/loading/error states are designed, not afterthoughts.

### Accessibility and responsiveness

- Optimize for 1440×900, 1280×800, and 1920×1080.
- At tablet widths, collapse analytic side panel and preserve main visualization.
- Mobile should be usable enough to browse text data and receive a helpful small-screen message, but is not a major design target.
- All essential controls keyboard usable and labeled; color contrast verified; chart values accessible via table/tooltip.

## 4. Intended UX flow and demo script

**Opening (0–5s)**: Voxel Earth appears on warm ivory canvas; satellites orbit slowly. Minimal header reads PERIGEE and "Explore"; compact caption says "Orbital traffic, made visible." A small note states the fixture snapshot date.

**Discovery (5–10s)**: Click "Featured encounter" or a visible orange point. Contextual panel shows `Object A × Object B` and predicted TCA/separation; camera frames the pair.

**Analysis (10–22s)**: Transition into encounter replay. Click play. Camera follows the relative encounter and the separation graph contracts toward the highlighted minimum. Freeze at TCA and show clear figures.

**Experiment (22–32s)**: Open Maneuver Lab. Adjust a single along-track delta-v control. Click "Recalculate"; results update with "Original vs Simulated" delta in miss distance. Highlight that the backend runs fresh propagation.

**Engineering proof (32–40s)**: Cut to benchmark output / reproducible test command, concise visual of Go parallel screening performance and validation summary. Show the repo and `docker compose up --build` command.

**Final (40–45s)**: Return to the beautiful globe with a subtle project tagline. Do not show fake metrics or manufactured terminal output.

Also create a silent, loopable **12–20s GIF** from the strongest interaction for the top of README. Full video target: **35–60 seconds**, ideally 1080p or higher.

## 5. Architecture

Use a straightforward monorepo:

```text
perigee/
├── AGENTS.md
├── README.md
├── docker-compose.yml
├── Makefile
├── .github/workflows/ci.yml
├── apps/
│   ├── web/                          # Vite + React + TypeScript
│   │   ├── src/
│   │   │   ├── app/                  # App shell, routing, theme
│   │   │   ├── components/           # Controls, panels, charts, reusable UI
│   │   │   ├── features/
│   │   │   │   ├── explorer/
│   │   │   │   ├── encounters/
│   │   │   │   └── maneuver/
│   │   │   ├── scene/                # React Three Fiber + voxel geometry
│   │   │   ├── api/                  # Generated/handwritten typed client
│   │   │   ├── styles/               # Tokens, fonts, resets
│   │   │   └── test/
│   │   └── package.json
│   └── api/                          # Go modular monolith
│       ├── cmd/server/main.go
│       ├── internal/
│       │   ├── catalog/              # OMM/TLE ingestion and validation
│       │   ├── propagation/          # SGP4 adapter and reference tests
│       │   ├── screening/            # Coarse candidates + refinement
│       │   ├── jobs/                 # In-process job runner and progress
│       │   ├── maneuver/             # Educational scenario recomputation
│       │   ├── httpapi/              # Handlers, DTOs, errors
│       │   └── storage/              # SQL implementations or fixture store
│       └── go.mod
├── contracts/
│   └── openapi.yaml                  # One authoritative contract
├── data/
│   ├── fixtures/                     # Deterministic catalog + encounters
│   ├── reference/                    # Attribution + provenance metadata
│   └── generated/                    # Ignored or selectively versioned
├── docs/
│   ├── PRODUCT.md
│   ├── DESIGN.md
│   ├── ARCHITECTURE.md
│   ├── PROGRESS.md
│   ├── DATA_SOURCES.md
│   ├── BENCHMARKS.md
│   ├── VALIDATION.md
│   └── DEMO_SCRIPT.md
├── scripts/
│   ├── demo.sh
│   ├── benchmark.sh
│   └── capture-demo.md
└── media/
    ├── screenshots/
    ├── demo.gif
    └── demo.mp4
```

**Tech selection**

- UI: React, TypeScript, Vite, React Router, TanStack Query if useful, CSS variables or Tailwind (choose one consistent styling approach), lightweight headless components where appropriate.
- 3D: Three.js + React Three Fiber + Drei; use custom simple geometry for voxel objects. Instancing and level of detail when needed.
- Charts: Recharts or a simple custom SVG separation chart. Keep charts understated and accurate.
- API: Go standard `net/http` with reasonable routing library if needed. REST/JSON between browser and Go. **Do not require gRPC** for a browser client; preserve existing gRPC implementation only if already useful and proven.
- Propagation: Adopt a maintained, tested Go SGP4 library or vetted port with explicit provenance/license; do not implement SGP4 from scratch for visual polish. If a suitable Go library is not verified, perform a documented spike and choose the least complex credible adapter approach. Compare its output against published SGP4 verification vectors.
- Persistence: PostgreSQL for stored catalog/run/encounter data if practical, but **fixture mode must work even if DB is unavailable**. SQLite would be an acceptable MVP reduction if the repository has no existing PostgreSQL dependency; choose based on current code and avoid migrations solely for resume optics.
- Background work: in-process bounded worker pool with job IDs and progress polling. Jobs cancelled on shutdown; no Redis/Kafka required.
- Containerization: one Go service, one static-dev frontend service; PostgreSQL if used; fixed compose commands and health checks. Ensure new developers can run the demo without a cloud account.

## 6. Numerical/engineering fundamentals

**Treat the visualization as stylized while the calculations remain explicit and unit-correct.**

1. Ingest TLE or OMM orbital element records, preserving object catalog ID, epoch (UTC), source metadata, and original text/fields needed for repeatability.
2. SGP4 propagation returns state vectors using the library's correct coordinate reference frame and units. Use **one canonical frame and units** for computations; transform properly for Earth rendering. Explicitly document ECI/TEME/ECEF and conversion assumptions instead of mixing frames.
3. For a bounded screening window, propagate objects at coarse time steps. Use altitude bands and spatial hashing / grid buckets (with neighboring cells) or another proven broad-phase to find candidate pairs; handle cell boundaries and fast objects conservatively.
4. Refine candidate close-approach intervals with more frequent sampling and a numerical minimum-search around the local minimum (e.g., bracket + bounded minimization). Track step sizes and approximation errors. Don't claim perfect collision prediction.
5. Return encounter events with object pair, TCA timestamp, minimum miss distance (meters), relative speed (m/s), and data-source epoch.
6. Use bounded goroutine worker pools and deterministic task partitioning; benchmark sequential baseline and optimized engine on the same hardware, dataset and accuracy criteria. Measure allocations as helpful.
7. Avoid N^2 full pair checks for the target-sized run; document any fallback on small fixtures.
8. Build a fixture catalog with known synthetic/test orbits to validate mechanics; separately import a versioned snapshot of allowed public data. Synthetic and real conjunctions must be labeled in UI and documentation.
9. Maneuver Lab: simple tangential impulse applied in an appropriate state-vector/propagation treatment, then recompute the impacted orbit and next encounter; verify that the chosen mathematical approximation works for the simulation horizon. If directly applying delta-v to an SGP4 propagator is not physically consistent, introduce a small, explicitly documented two-body propagation model for post-maneuver what-if trajectories. **Do not pretend SGP4 automatically accounts for a maneuver.** Show simplified-model limitations.
10. Do not compute collision probability without covariance/uncertainty and an appropriate accepted model.

**Validation requirements**

- Official/public SGP4 vector comparison for multiple inclinations/eccentricities and time offsets.
- Known fixture pair with analytically or independently checked closest approach.
- Boundary cases: missing elements, decayed objects, stale epochs, same catalog ID, time-window edges, no conjunction, invalid timestamps, mismatched units/reference frames.
- SOCRATES comparison is a **separate evaluation**: align data epochs and reference window, document matching tolerance and false-positive/false-negative policy. Avoid hard claims like `96% recall` until a reproducible test confirms them.
- Accept that public orbit element uncertainty can materially affect very close predicted approaches; UI should show caution and educational framing.

## 7. API contract (freeze early)

Define `contracts/openapi.yaml` in the initial architecture phase. Recommended endpoints:

- `GET /api/health` — status.
- `GET /api/catalog?query=&limit=&cursor=` — searchable object summary.
- `GET /api/catalog/{id}` — one object with metadata/epoch.
- `GET /api/demo/featured` — pinned featured encounter IDs and fixture metadata.
- `POST /api/screenings` — payload `{catalogSnapshotId, startTimeUtc, endTimeUtc, thresholdMeters, coarseStepSeconds}`; response `202 {runId, status}`.
- `GET /api/screenings/{runId}` — run status, phase, progress, error, timestamps.
- `GET /api/screenings/{runId}/encounters?limit=&cursor=` — paginated sorted conjunctions.
- `GET /api/encounters/{encounterId}` — metrics + object summaries.
- `GET /api/encounters/{encounterId}/trajectory?startUtc=&endUtc=&samples=` — **both object states at identical time samples**, distance series, TCA marker, frame metadata.
- `POST /api/encounters/{encounterId}/maneuvers` — `{maneuverTimeUtc, alongTrackDeltaVMps}` -> `{scenarioId, originalMetrics, simulatedMetrics, originalTrajectory, simulatedTrajectory, limitations}`; can be async for slow calculations.

**Contract details**

- Consistent timestamps: RFC3339 UTC with ISO `Z`.
- All numerical fields end in units (`Meters`, `Mps`, `Seconds`) or have a documented schema `unit`.
- Explicit `datasetType: synthetic | public_snapshot` and `sourceEpochUtc` where relevant.
- HTTP errors with `{code, message, details}`.
- Stable IDs, deterministic sorting, pagination, and client-friendly partial results.
- Typed frontend client generated or manually synchronized from OpenAPI; shared sample responses validated by tests.
- Do not put 25K fully sampled trajectories in a single response. Paginate object lists and sample trajectories only for selected encounters.

## 8. Multi-agent operating model

### Agent 0: Lead / integrator (owns coordination)

**Ownership:** `AGENTS.md`, `contracts/openapi.yaml`, `docs/ARCHITECTURE.md`, `docs/PROGRESS.md`, integration decisions, build/CI, shared type changes, final review. Merge and reconcile work. Never delegate cross-cutting API changes without notifying all affected agents.

**Responsibilities:** inspect repo, create phase checklist, freeze API interfaces, pick viable libraries, issue bounded tasks, maintain integration test, reject complexity creep, verify claimed performance, accept each phase only if it meets acceptance tests.

### Agent 1: Product design + UI engineering

**Ownership:** `apps/web/src/app/`, `apps/web/src/components/`, `apps/web/src/features/` (except 3D scene), `apps/web/src/styles/`, `docs/DESIGN.md`.

**Tasks:** implement light-default desktop shell, typographic hierarchy, refined design tokens, top navigation, themed cards, encounter details, clean chart, beautiful timeline, theme toggle, UI states, keyboard/accessibility. May use static fixture API responses initially. Build polished screen layouts at 1440x900 before expanding features.

**Must not modify:** Go backend, data science logic, `contracts/openapi.yaml` without coordination, `src/scene` files.

### Agent 2: 3D visualization engineer

**Ownership:** `apps/web/src/scene/`, scene-focused tests, assets authored for scene where needed.

**Tasks:** optimized voxel Earth (globe with stylized blocks or coherent low-poly geometry), reusable voxel satellite model, orbit line geometry from API samples, encounter markers, controlled camera animation, scene selection callbacks, closest approach replay, dark/light materials, performance and reduced-motion behavior.

**Must not modify:** shared UI layout, HTTP API, Go backend. Provide scene through typed React props/callback interfaces agreed with Agent 1.

**Scene API to agree early:**

```ts
export type SceneSelection = { kind: 'satellite'; id: string } | { kind: 'encounter'; id: string } | null;
export type SceneMode = 'explore' | 'encounter' | 'maneuver';
export interface OrbitalSceneProps {
  mode: SceneMode;
  sceneTimeUtc: string;
  selected: SceneSelection;
  satellites: SatelliteVisual[];
  paths: OrbitPathVisual[];
  encounter?: EncounterVisual;
  onSelect: (selection: SceneSelection) => void;
}
```

Agent 0 and Agent 1 finalize concrete visual types before Agent 2 builds implementation.

### Agent 3: Go engine + API

**Ownership:** `apps/api/`, backend tests and bench tests, `docs/VALIDATION.md`, backend portion of `docs/DATA_SOURCES.md`.

**Tasks:** library selection spike, fixture loader, propagator adapter, screen engine, closest-approach refinement, deterministic job API, trajectory sampling, verified maneuver approximation, ingest module, metrics and benchmarks.

**Must not modify:** UI assets or design. Coordinate endpoint changes through Agent 0.

### Agent 4: QA + docs + demo producer

**Ownership:** `apps/web/src/test/` (or e2e testing directory), `docs/DEMO_SCRIPT.md`, `docs/BENCHMARKS.md` (results collected from tests), root README drafts, `media/` and reproducible capture directions, CI proposals submitted to Agent 0.

**Tasks:** define end-to-end happy path, cross-browser smoke tests, accessibility checks, startup verification, capture shot list, generate README media only from the functioning app, produce diagram, validate claimed metrics. During initial backend/frontend development, work from contracts and test skeletons, not fake recording.

**Must not modify:** main application implementation except small documented fixes coordinated with owner.

### How to parallelize without merge disasters

**Wave 0 (serial, Lead):** inspect code, document decisions, scaffold monorepo, freeze API schemas and `OrbitalSceneProps` interface, add fixtures and simple starter app.

**Wave 1 (parallel):** Agent 1 builds shell/layout with fixtures; Agent 2 builds standalone 3D scene with a typed fixture adapter; Agent 3 builds Go ingest/propagation/API; Agent 4 builds test plan, README storyboard and e2e skeleton. Merge at named checkpoint.

**Wave 2 (integrated):** Agent 0 stitches UI/scene/API into first vertical slice. Agent 3 hardens screening. Agent 1 and 2 integrate encounter replay. Agent 4 runs smoke path and catches discrepancies. Shared code changes must be serialized through Agent 0.

**Wave 3:** Agent 3 builds maneuver calculations while Agent 1 and Agent 2 build Lab UI against frozen response contract; Agent 4 expands tests and README capture. Integrate.

**Wave 4:** benchmark, visual QA, polish, generate screenshots/video, fix issues, tag release. No feature sprawl.

Each agent reports: files touched, features completed, commands/tests run, failures, API assumptions, remaining tasks. Lead rejects outputs that are only conceptual/mockups after implementation phases begin.

## 9. Phased implementation milestones and acceptance criteria

### M0 — Repository audit and foundations

- Audit existing code, confirm whether claimed engine exists and what actually runs.
- Decide dependencies based on current state; scaffold missing packages.
- Write `AGENTS.md`, `docs/PRODUCT.md`, `docs/DESIGN.md`, `docs/ARCHITECTURE.md`, `docs/PROGRESS.md`, `contracts/openapi.yaml`.
- `make dev` or documented equivalent starts UI + API, health endpoint responds.
- **Accept when:** repository boots; app routes exist; typed contract is unambiguous; agents have nonoverlapping ownership.

### M1 — The visual hook

- Build app shell in premium warm light mode, voxel Earth, 3–6 visually curated satellites, a subtle featured encounter marker, minimal top chrome and theme toggle.
- Scene camera rotates/zooms, object selects; single clean contextual panel opens.
- All data may use truthful fictional/synthetic fixture labels at this phase.
- **Accept when:** a still screenshot at 1440×900 looks like a polished finished product; dark theme looks intentional, not just inverted; no console errors; scene usable at 60 fps on reasonable development hardware where practical and documented.

### M2 — Real screening and evidence

- Verified SGP4 adapter, catalog snapshot loader, coarse candidate search, closest-approach refinement, bounded Go workers.
- Running API endpoints, basic jobs and deterministic fixture results.
- Minimal unit/reference tests, standalone backend benchmark harness.
- **Accept when:** results are derived from the Go engine (not hardcoded mock UI); numerical assumptions documented; repeatable tests pass.

### M3 — Encounter cinematic replay

- Full Encounter Analysis screen, synced time scrubbing, two trajectories, animated camera, a real separation series computed from the same backend sampled points.
- Feature an eye-catching fixture encounter with known, documented values. Use explicit provenance in UI.
- **Accept when:** graph minimum and displayed miss distance agree within documented tolerance; scrubbing moves both satellites and chart marker to exactly the same time; clear pause and reset; accessible controls.

### M4 — Maneuver Lab

- Single tangential delta-v input (sane bounded range defined with scientific justification), recompute action, side-by-side original vs hypothetical result, informative limitations.
- API errors and loading visible, reset/undo.
- **Accept when:** changing delta-v causes actual recomputation and tests verify non-trivial changes; no claims that a successful avoidance maneuver is guaranteed.

### M5 — Product polish and GitHub launch

- Visual consistency audit, microinteractions, robust startup and demo fixture, README and recorded media.
- At least one meaningful benchmark and supporting commands, validated sources, performance reproducibility notes and test badge if CI configured.
- **Accept when:** an unfamiliar engineer can clone and locally run the demo; the README communicates the product within five seconds; a silent GIF demonstrates the core behavior; a full 45-second video shows what makes it special; no unsupported performance claims.

## 10. Data assets and reproducibility

- Prefer CelesTrak GP / OMM data as an optional live import path. Cache and snapshot downloaded results with retrieval time, catalog epoch and attribution; do not hammer public APIs on page loads.
- Use an explicitly versioned fixture file for the full first-run demo. If synthetic, label it as synthetic. Document fields and generation process, ideally with `scripts/generate-fixtures` and deterministic seed.
- SOCRATES result snapshots can support an **evaluation benchmark** if usage terms and dataset matching are documented. Do not imply SOCRATES is ground truth for physical collision outcomes.
- Keep raw large data out of Git unless justified. Ship a small starter fixture and scripts for acquiring larger datasets according to provider terms.
- Update README with attribution, source links, time period and scientific caveats.

## 11. Testing and performance

**Backend**

- `go test ./...` including propagation reference fixtures and candidate search correctness.
- Deterministic near-pass tests; compare coarse-to-refined against much denser brute-force time sampling on small datasets.
- `go test -bench` for worker pool vs sequential and spatial index vs brute-force *on the same input and output tolerance*.
- Record machine CPU model, logical/physical core count, OS, Go version, object count, window, sampling interval, threshold, repetition count, and command used.

**Frontend**

- `npm run typecheck`, `npm run lint`, unit tests for conversion/formatting, timeline model, thematic components.
- E2E: load Explorer, select featured encounter, play/seek, change theme, run maneuver scenario, reset.
- Screenshot testing at desktop sizes if practical; check legibility and clipping.

**Integration**

- `docker compose up --build` works from clean checkout.
- No external API required to open the demo.
- Clear error state if WebGL unavailable.

**What to optimize for:** smooth select-to-encounter transitions, accurate selected trajectories and believable UI polish. Keep large-catalog benchmark separate from interactive frontend load.

## 12. README strategy (almost as important as the product)

Top 1–2 scroll heights must sell it:

```md
# PERIGEE
### A clearer view of what's happening in orbit.

A playful orbital intelligence playground powered by a high-performance Go conjunction-screening engine. Explore tracked satellites, replay predicted close approaches, and experiment with hypothetical orbital maneuvers in an interactive voxel universe.

[WATCH THE DEMO](media/demo.mp4) · [RUN LOCALLY](#run-locally)

![Perigee encounter replay](media/demo.gif)

**Explore** orbital traffic · **Investigate** close approaches · **Experiment** with maneuvers

[Three hero screenshots: Explorer | Encounter | Maneuver]
```

Below the fold:

1. **Why Perigee exists:** Earth's orbit is crowded; making the screening problem understandable and explorable.
2. **The three core experiences:** short crisp copy and gorgeous screenshots.
3. **Under the hood:** Go propagation, broad-phase filtering, closest-approach refinement, data flow, one small architecture diagram.
4. **Measured performance:** only true results with hardware and commands; emphasize real end-to-end metrics.
5. **Accuracy & limitations:** how verified, what assumptions are present, limitations of public elements and maneuver approximation.
6. **Local quickstart:** one command and one URL, plus fixtures and sample navigation.
7. **Sources and credits:** CelesTrak and SGP4 implementation/license.

**Visual asset requirements**

- README cover/GIF: optimized compressed 12–20 second sequence, roughly <= 10–20MB if possible, no fake animated mockups masquerading as functional code.
- `media/demo.mp4`: authentic 35–60 second narrated or captioned recording of the running web app, encoded to play locally or via GitHub file link; if large, use a linked externally hosted video rather than committing oversized binaries.
- `media/screenshots/01-explorer.png`, `02-encounter.png`, `03-maneuver.png`: consistent viewport and theme, no fabricated data; a dark-mode teaser image optionally.
- Simple architectural diagram in Mermaid or SVG.
- A concise, recruiter-friendly explanation of what is technically difficult about the project.

**Recruiter talking points** (use only once implemented):

- “I built a fast orbital screening engine in Go, then designed a voxel-based web interface that turns its predictions into interactive encounter replays.”
- “The interesting engineering problem was narrowing a very large pairwise search into a manageable number of candidate encounters while keeping the results reproducible and verifiable.”
- “I built the demo to work offline because I wanted engineers and recruiters to be able to inspect and run it without an account or cloud service.”
- “I made the scientific boundaries explicit: separation prediction is not the same thing as collision probability, and the maneuver view is an educational what-if simulation.”

## 13. Starter task allocation for Codex

After Wave 0 contract/architecture is complete, delegate these tasks in parallel (adjust to environment support):

**Frontend/Design worker prompt:**
> Implement the premium light-first desktop Perigee UI shell in `apps/web/src/app`, `components`, `features`, and `styles`, without touching `src/scene` or the Go API. Follow `docs/DESIGN.md`. Deliver polished Explorer, Encounter and Maneuver page layouts with fixture-backed typed data and an Apple-like precision timeline. Include dark toggle, keyboard states and accessibility. Report files, tests and unresolved contract issues.

**3D worker prompt:**
> Implement the Perigee voxel Earth and satellite scene in `apps/web/src/scene` only. Use React Three Fiber and reusable geometry. Follow frozen scene types, support select, orbit paths, close-approach visual and camera transitions, light/dark palettes and reduced motion. Do not touch app layout or API. Deliver a visualizable scene with tests where appropriate.

**Backend worker prompt:**
> Implement Go catalog, verified SGP4 propagation adapter, spatial broad-phase candidate search and closest-approach refinement under `apps/api`. Use frozen OpenAPI contract. Start with deterministic fixtures; deliver real computed encounter results, trajectory sampling, job progress and tests. Identify assumptions, SGP4 verification source, performance limits and any missing dependencies. No cloud/queue complexity.

**QA/README worker prompt:**
> Build a test and portfolio showcase plan, `docs/DEMO_SCRIPT.md`, README structure, startup smoke checks, and reliable e2e tests against fixtures. Do not claim features or metrics are implemented until verified. Capture screenshots/video only after end-to-end product works; propose fixes to owners instead of modifying their source files.

**Lead checkpoint prompt:**
> Integrate worker changes, ensure type contracts match, run compose and tests, inspect the visual layout at 1440x900, update `docs/PROGRESS.md`, and only then start the next wave. Prioritize a complete visual vertical slice over unrelated extra features.

## 14. First execution request (paste into Codex)

> Read `PERIGEE_CODEX_MASTER_PLAN.md` completely and act as the lead/orchestrator for this project. Inspect the current repository first and preserve any useful existing code. Create `AGENTS.md`, architecture/design docs, frozen API contract and a small prioritized issue plan. Use subagents for bounded, parallel file-owning tasks if supported. Then begin **M0 and M1 immediately**: create a runnable Vite + React + TypeScript web app and Go API skeleton, implement the light-mode desktop app shell, voxel Earth, a few animated satellites, featured encounter interaction and theme toggle using deterministic fixtures. Follow the precise design direction in this file. Run the project and tests, fix errors, and update `docs/PROGRESS.md`. Do not stop after writing a plan, and do not spend time on deployment or enterprise infrastructure. After M1, report the working features, commands, screenshots if available, limitations, and next milestone.

## 15. Helpful sources to verify during implementation

- Codex guidance for repository agent instructions and subagents: https://learn.chatgpt.com/docs/customization/overview and https://learn.chatgpt.com/docs/agent-configuration/subagents
- CelesTrak GP formats: https://celestrak.org/NORAD/documentation/gp-data-formats.php
- CelesTrak SOCRATES CSV documentation: https://celestrak.org/SOCRATES/socrates-format.php
- CelesTrak usage policy: https://celestrak.org/usage-policy.php
- Satellite.js (front-end reference only, not a substitute for verified Go propagation): https://shashwatak.github.io/satellite-js/

**Final acceptance standard:** If a recruiter sees only the README and 45-second demo, the project is immediately understandable, visually memorable, and technically credible. If an engineer clones it, the calculation pipeline, test fixtures, documentation and benchmark claims hold up under inspection.
