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

This is an actively implemented local-first monorepo. The API and web application run locally through Docker Compose or their direct development commands.

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

- A Go `net/http` service with fixture-backed catalog, encounter, trajectory, and maneuver routes.
- Local TLE/OMM parsing, provenance and epoch validation, and near-Earth SGP4 adapters.
- A bounded in-process OMM screening core that evaluates all object pairs over coarse segments and refines a selected minimum per pair.
- OMM imports are validated and retained in a four-entry in-memory snapshot cache. Screening is exposed through bounded asynchronous API jobs, but is not yet connected to the web application.

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

The Screening tab submits a local OMM file and provenance to the Go API, polls the resulting screening job, and retrieves cursor-paginated encounter summaries. The web app uses same-origin API requests via the Vite development proxy or the Compose nginx proxy. The Encounter replay and Maneuver Lab use synchronized synthetic fixtures. From a Screening result, the web app loads the run-scoped detail and uses the same paired propagated samples for the scene, scrubber, and separation chart.

## API Contract

The lead owns `contracts/openapi.yaml` as the authoritative API contract. Health, fixture catalog/encounter/trajectory/maneuver routes, OMM validation/cache, and asynchronous screening creation/status/results, run-scoped encounter detail, and screened what-if previews are implemented.

The HTTP error format is `{code, message, details}`. Timestamps are RFC3339 UTC with Z. Units are explicit in field names or schema units. IDs, sorting, and pagination are deterministic. The API must support fixture mode without an external database.

`POST /api/catalog/omm` validates a caller-supplied OMM JSON snapshot and provenance, then retains it in a four-entry LRU in-memory cache for screening. The cache is not persisted, is cleared on API restart, and does not change fixture catalog or encounter routes.

A retained successful run serves `GET /api/screenings/{runId}/encounters/{encounterId}`. Each run retains the normalized OMM elements needed to regenerate replay after snapshot eviction. The handler propagates the two objects around the reported TCA on demand, returning 61 synchronized TEME samples spaced 10 seconds apart, with position in km, velocity in m/s, and separation in m. The fixed window can extend beyond the screening interval. The reported miss distance describes the screen and is evaluated at its TCA.

The run-scoped `POST /api/screenings/{runId}/encounters/{encounterId}/what-if` computes an educational preview for object A. A signed impulse parallel to its velocity is fixed to the first replay sample. Two central-gravity RK4 integrations estimate the incremental change in position and velocity at each 10-second step. The API adds that difference to the original SGP4 samples and recomputes paired separation in TEME. The request is stateless and leaves the screening result available for comparison.

## State and Jobs

Screening uses two in-process workers and a queue of up to eight pending jobs. At most 100 run records are retained; when full, the oldest completed run is evicted, and requests are rejected if all retained runs are active. The four-entry OMM cache evicts its least-recently-used snapshot. Job, queue, and snapshot state are volatile and are lost on API restart. No cancellation route or external queue is implemented.

## Persistence

No database is currently used. Fixture mode works without a database dependency. Imported OMM snapshots and screening results exist only in bounded process memory; they are not durable and are not merged into fixture data.

## Numerical Model

- Supported ingestion formats are TLE or OMM records.
- Original object ID, epoch, source, and repeatable source fields are preserved.
- Propagation uses an explicitly documented and maintained SGP4 adapter or a vetted port.
- SGP4 output is TEME; do not label it ECI without an explicit frame conversion. The selected adapter accepts supported near-Earth TLEs and compatible OMM elements; deep-space SDP4 remains unsupported.
- The current SGP4 adapter returns TEME positions in km and velocities in km/s; no frame conversion is implemented.
- The screening core evaluates all pairs for catalogs up to 100 supported objects and at most 1441 time samples. It uses piecewise-linear coarse segments and refines one selected candidate interval per pair; a coarse step that is too large may miss close approaches.
- Maneuver what-if behavior uses a clearly documented approximation; if SGP4 is inappropriate for post-maneuver propagation, the system must state and implement the alternative model.
- No collision probability is calculated.

## Deployment

Docker Compose and direct local development commands run one Go API and one web application. No hosted deployment or service is required for MVP.

## Data

`data/fixtures/` contains deterministic demo catalog and encounter data. `data/reference/` contains provenance and verification assets. Generated artifacts stay out of source control unless intentionally versioned.

Fixture scenarios must distinguish synthetic, public snapshot, or other source type in UI and metadata.

## Testing and Integration

- Go tests cover TLE/OMM ingestion, SGP4 reference vectors, bounded screening/refinement, asynchronous screening API jobs/results, API behavior, and fixture maneuvers. The web integration test exercises import, status polling, recovery from a failed results request, cursor pagination, and synchronized replay with mocked HTTP responses.
- Frontend tests cover formatting, timeline behavior, themes, and interaction.
- E2E tests cover the happy path from Explorer to Encounter Analysis to Maneuver Lab.
- Docker Compose smoke testing starts the system cleanly.
- Integration gates run typecheck, lint, tests, and build where available.
