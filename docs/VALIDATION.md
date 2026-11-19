# Go API fixture mode

The API is a Go service under `apps/api`. It reads the
versioned JSON fixtures in `data/fixtures`; set `PERIGEE_FIXTURES_DIR` to use
another directory containing `catalog.json` and `encounters.json`. The service
walks parent directories from its working directory to find the repository
fixtures, so it can be started from the repository root or `apps/api`.

From the repository root, run:

```powershell
cd apps/api
go test ./...
go run .
```

The service listens on `:8080` by default; `ADDR` overrides the listen address.
Fixture API routes include health, catalog search/details, featured encounter,
encounter details, synchronized trajectory samples, and the maneuver endpoint.
Pagination cursors are integer offsets and are intended only for the small,
deterministically sorted fixture catalog.

Trajectory API states remain generated synthetic linear samples, not output from
an orbital propagator. They use the contract's `ECI` enum only as a
client-compatible frame marker; the response caveat explicitly disclaims a
physically propagated inertial frame. The maneuver endpoint applies a simple
post-impulse linear position offset to this fixture model and samples the
resulting separation. Neither it nor the fixture encounter metrics are
operational predictions, collision probabilities, or maneuver guidance.

`apps/api/internal/orbit` now contains a separate TLE SGP4 adapter. It uses
[`github.com/akhenakh/sgp4`](https://github.com/akhenakh/sgp4) v1.0.5
(Apache-2.0), which documents near-Earth SGP4 support and explicitly does not
implement deep-space SDP4 corrections. The adapter returns TEME position in km
and velocity in km/s. It remains separate from fixture trajectory and maneuver
endpoints, which continue to return synthetic states.

The adapter test checks a near-Earth case against the published TEME values in
Vallado's [SGP4-VER suite (AIAA 2006-6753)](https://celestrak.org/publications/AIAA/2006-6753/).
`go test ./...` and `go vet ./...` pass from `apps/api`;
`docker compose build api` also passes. The API container uses Go 1.23 or
newer to satisfy the dependency's Go 1.23.4 minimum.

The same package provides `ReadTLECatalog` for two-line and named three-line TLE
catalogs and `ReadOMMCatalog` for a single CCSDS-style OMM JSON object or array.
Both require source name, snapshot ID, retrieval time, and an explicit maximum
epoch age. They reject invalid or duplicate catalog numbers and future epochs,
mark old elements `stale`, and retain valid-but-unsupported propagation models
with an explicit limitation.

`POST /api/catalog/omm` accepts a bounded OMM JSON import with provenance and
freshness policy. It returns validated records and their propagation
limitations while preserving the original OMM JSON record. Validated snapshots
are held in a four-entry LRU memory cache for screening; they are not persisted,
are lost on API restart, and are not merged into `GET /api/catalog` or fixture
encounter routes. The API contract is in `contracts/openapi.yaml`. The OMM
adapter maps through the selected SGP4 dependency's OMM conversion and
initialization; its Vallado epoch state matches
the same reference vector as the TLE adapter. Mean-motion derivative
interpretation follows that dependency's mapping and must be checked against
the chosen source convention. Only near-Earth SGP4-compatible records are
supported; deep-space SDP4, uncertainty, and collision probability remain out
of scope.

`ScreenOMMCatalog` is the first M8 screening core. It requires fresh,
propagation-supported OMM elements, evaluates all object pairs for at most 100
objects and 1441 samples, then refines one candidate interval per pair. Its
coarse segments are time-step dependent and may miss close approaches when the
step is too large. It reports miss distance and relative speed, not collision
probability.

The [dense-grid comparison](BENCHMARKS.md#synthetic-reference-comparison)
checks whether screening recovers pairs present below the threshold on a
fine SGP4 time grid. In this fixture, 30- and 60-minute steps missed both
5 km reference pairs. The comparison shares the propagator with screening
and evaluates candidate selection on synthetic OMM data.

The API exposes `POST /api/screenings`, status polling, and paginated encounter
results. It runs two in-process workers with an eight-job queue and retains at
most 100 job records; queued and completed state is volatile and is lost on API
restart. Results include the core's caveats. The web Screening tab connects
import, screening, and a run-scoped SGP4 detail view using paired TEME samples.

### Screened encounter what-if

The fixed impulse applies to object A along its initial velocity at the
replay start, with a signed magnitude bounded by 0.5 m/s. Two integrations
start at the same SGP4 position and velocity, with the impulse added to one
initial velocity. They advance for 600 seconds using classical fourth-order
Runge?Kutta at 10-second steps and central Earth gravity
(mu = 398600.4418 km^3/s^2). Their difference is added to object A's SGP4
position and velocity. Object B follows its existing trajectory. The
separation in meters is recomputed from the resulting positions at each
sample.

This estimates the incremental response over a short educational replay.
The perturbation model omits effects such as nonspherical gravity, drag,
and maneuver execution uncertainty. The replay minimum is a minimum over
sampled timestamps. It does not estimate collision probability or provide
operational guidance.

The integrator test compares a circular orbit to its analytic solution.
Impulse tests check both signs, initial velocity and position, zero-delta
identity, and range rejection. The API test checks matching timestamps,
unchanged object B, zero-delta baseline identity, and that a nonzero preview
changes separation without altering the stored run. These tests verify
implementation properties within the approximation, not physical accuracy.
