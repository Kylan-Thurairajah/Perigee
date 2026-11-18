# Perigee Architecture

## Repository Shape

Perigee is a straightforward monorepo:

```text
perigee/
|-- AGENTS.md
|-- README.md
|-- docker-compose.yml
|-- Makefile
|-- apps/
|   |-- web/
|   `-- api/
|-- contracts/
|   `-- openapi.yaml
|-- data/
|   |-- fixtures/
|   |-- reference/
|   `-- generated/
|-- docs/
|   |-- PRODUCT.md
|   |-- DESIGN.md
|   |-- ARCHITECTURE.md
|   `-- PROGRESS.md
`-- scripts/
```

This structure is the target repository layout. It does not imply that directories or files beyond this documentation task currently exist.

## Runtime Components

### Web application

- React and TypeScript in Vite.
- React Router for app views.
- CSS variables or one consistent styling approach.
- TanStack Query is optional.
- Three.js through React Three Fiber and Drei for the scene.
- Recharts or a simple custom SVG for the separation chart.
- A typed HTTP client synchronized with `contracts/openapi.yaml`.

### Go service

- A modular Go monolith using `net/http`.
- Catalog ingestion and validation.
- Propagator adapter.
- Broad-phase candidate screening.
- Close-approach refinement.
- In-process asynchronous jobs and progress polling.
- Maneuver what-if recomputation.
- HTTP DTO and error mapping.
- Fixture-backed data when PostgreSQL or SQLite is unavailable.

## Data Flow

```text
Fixture catalog or ingested record
          |
          v
Catalog + source metadata
          |
          v
SGP4/propagation adapter
          |
          v
Coarse spatial screening
          |
          v
Close-approach refinement
          |
          v
Encounter and sampled trajectory results
          |
          v
Typed REST/JSON + React scene and chart
```

An encounter replay and Maneuver Lab receive trajectory samples from the same backend series. The UI must not independently synthesize time, object position, TCA, or graph values.

## API Contract

The lead owns `contracts/openapi.yaml` and freezes endpoint behavior before independent implementation. The contract is expected to cover:

- `GET /api/health`
- catalog list and detail
- featured demo encounter
- asynchronous screening creation and status
- encounter pagination and detail
- equal-time-sampled trajectories
- educational maneuver scenarios

The HTTP error format is `{code, message, details}`. Timestamps are RFC3339 UTC with Z. Units are explicit in field names or schema units. IDs, sorting, and pagination are deterministic. The API must support fixture mode without an external database.

The lead-owned contract now exists at `contracts/openapi.yaml`. Independent frontend and API work must treat it as authoritative and report, rather than silently make, cross-cutting changes.

## State and Jobs

- Screening uses a bounded in-process worker pool.
- A POST request creates a run and immediately returns `202`.
- Status polling exposes phase, progress, timestamps, and errors.
- Results are paginated and deterministically sorted.
- Jobs are cancelled on service shutdown.
- No Redis, Kafka, or external queue is required for MVP.

## Persistence

The initial implementation should choose SQLite or PostgreSQL based on the repository and developer setup. SQLite is an acceptable MVP reduction. PostgreSQL is optional. Fixture mode must work without a database dependency.

## Numerical Model

- Supported ingestion formats are TLE or OMM records.
- Original object ID, epoch, source, and repeatable source fields are preserved.
- Propagation uses an explicitly documented and maintained SGP4 adapter or a vetted port.
- A canonical frame and unit system is chosen and conversions are documented.
- Coarse screening uses a spatial grid or an equally conservative broad phase.
- Candidate intervals are refined by bounded local minimum search.
- Maneuver what-if behavior uses a clearly documented approximation; if SGP4 is inappropriate for post-maneuver propagation, the system must state and implement the alternative model.
- No collision probability is calculated.

## Deployment

M0 and M1 target local Docker Compose and direct development commands. A later environment may contain one Go service and one static or dev frontend service. PostgreSQL is included only if chosen for the implementation.

No deployment or hosted service is required for MVP.

## Data

`data/fixtures/` contains deterministic demo catalog and encounter data. `data/reference/` contains provenance and verification assets. Generated artifacts stay out of source control unless intentionally versioned.

Fixture scenarios must distinguish synthetic, public snapshot, or other source type in UI and metadata.

## Testing and Integration

- Go tests cover ingestion, propagation reference vectors, candidate search, refinement, API behavior, and maneuvers.
- Frontend tests cover formatting, timeline behavior, themes, and interaction.
- E2E tests cover the happy path from Explorer to Encounter Analysis to Maneuver Lab.
- Docker Compose smoke testing starts the system cleanly.
- Integration gates run typecheck, lint, tests, and build where available.
