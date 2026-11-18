# Go API fixture mode

The initial API is a standard-library Go service under `apps/api`. It reads the
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

Trajectory states are generated synthetic linear samples, not output from an
orbital propagator. They use the contract's `ECI` enum only as a client-compatible
frame marker; the response caveat explicitly disclaims a physically propagated
inertial frame. The maneuver endpoint applies a simple post-impulse linear
position offset to this fixture model and samples the resulting separation.
Neither it nor the fixture encounter metrics are operational predictions,
collision probabilities, or maneuver guidance.

Validation command: `go test ./...` from `apps/api`. This environment did not
have `go` or `gofmt` available on `PATH`, so tests could not be run here.
