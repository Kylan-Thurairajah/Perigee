# Screening benchmarks

Run from `apps/api`:

```powershell
go test ./internal/orbit -run '^$' -bench '^BenchmarkScreenOMMCatalog$' -benchmem -benchtime=3x -count=3
```

The benchmark derives a deterministic, synthetic catalog from the Vallado OMM verification record already used by the SGP4 tests. It assigns unique catalog IDs and spreads the records evenly in mean anomaly. Each case uses a 60-second coarse step and a 10,000 m threshold, and checks the actual object, sample, and pair counts. Timing starts after OMM parsing and covers propagation, pair screening, and refinement. The API job queue, serialization, and browser rendering are outside the measurement.

## Local baseline

Measured on 2026-11-18 with Go 1.27.4, Windows amd64, and an AMD Ryzen 7 5700X. Each range spans three benchmark runs of three iterations. Allocation volume is per operation.

| Objects | Samples per object | Pairs | Time per operation | Allocation volume |
| ---: | ---: | ---: | ---: | ---: |
| 2 | 61 | 1 | 0.088–0.115 ms | 50.1 kB |
| 25 | 61 | 300 | 1.21–1.47 ms | 0.60 MB |
| 100 | 61 | 4,950 | 6.55–6.66 ms | 2.40 MB |
| 100 | 1,441 | 4,950 | 155–160 ms | 52.35 MB |

These measurements describe the stated synthetic workloads on one machine. Catalog geometry and the number of candidate refinements can change screening cost.

## Synthetic reference comparison

From `apps/api`:

```powershell
go test ./internal/orbit -run '^TestScreeningPairRecall$' -count=1 -v
```

The test creates 12 Vallado-derived synthetic OMM records, arranged in six crossing pairs with varied initial elements, and checks all 66 unordered pairs over one hour. The reference propagates all objects at 0.5-second intervals and tracks the smallest sampled separation for each pair. A second pass at 1-second intervals must agree on which pairs cross each threshold. The production screen is evaluated by pair presence because it reports one minimum per pair.

| Coarse step | 5 km threshold | 10 km threshold | 20 km threshold |
| ---: | ---: | ---: | ---: |
| 5, 30, 60, 300, or 600 s | 2/2 | 6/6 | 6/6 |
| 1,800 or 3,600 s | 0/2 | 6/6 | 6/6 |

The denominators count reference pairs. There were zero additional pairs at these settings. The 5- and 60-second cases assert full recovery in the fixture. The denser reference uses the same SGP4 implementation and finite sampling. The result measures agreement on this synthetic catalog, whose pairs have limited orbital diversity. It does not establish operational recall or validate the underlying propagator against observed truth.
