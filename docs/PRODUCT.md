# Perigee Product

## Positioning

Perigee is a local-first orbital traffic visualization and conjunction-screening application. Its purpose is to make orbital close approaches understandable and explorable for software engineering recruiters, hiring managers, and engineers.

The product is visually distinctive on the surface and technically explicit underneath: a restrained voxel universe presents satellite paths and encounters, while a Go engine performs catalog ingestion, propagation, screening, refinement, and hypothetical maneuver recomputation.

Tagline: "A clearer view of what's happening in orbit."

## Core Experiences

### Explorer

The default desktop view contains a warm, light-mode voxel Earth, a small curated set of satellites, subtle orbit paths, and a featured encounter marker. Users can rotate and zoom the scene, select objects, and open a concise contextual panel.

### Encounter Analysis

Encounter Analysis focuses on one object pair. It presents the TCA, minimum separation, relative speed, object metadata, source epoch, and data-quality caveat. A precision timeline and separation chart replay the same sampled trajectory series used by the scene.

### Maneuver Lab

Maneuver Lab provides one educational what-if: an instantaneous along-track delta-v applied at a chosen time before an encounter. It compares original and simulated trajectories and metrics while explicitly labeling the approximation and its limitations.

## Primary User Flow

1. Open the app and view the fixture snapshot and curated orbital scene.
2. Select a featured encounter or an object.
3. Replay the close approach with a synchronized timeline and separation graph.
4. Optionally apply a bounded along-track delta-v in Maneuver Lab.
5. Recalculate and inspect original-versus-simulated results and caveats.

## MVP Scope

- Desktop-first React and TypeScript interface with persistent light/dark theme.
- Explorer, Encounter Analysis, and Maneuver Lab views.
- Procedural voxel art and controlled Three.js scene.
- Camera orbit, zoom, object selection, encounter selection, and hover/click markers.
- At least three coherent, reproducible synthetic or public-snapshot fixture scenarios.
- Go ingestion and SGP4 propagation for supported records.
- Coarse screening, candidate refinement, trajectory samples, and in-process asynchronous jobs.
- Result identity, TCA, minimum separation, relative speed, source epoch, and data-quality label.
- One documented tangential maneuver approximation.
- Local one-command setup and reproducible sample data.
- Backend, frontend, accessibility, and integration tests.
- README, screenshots, GIF, and optional short video based only on verified behavior.

## Scientific Boundaries

Perigee performs conjunction screening and reports geometric miss distance. It does not estimate collision probability unless covariance and an accepted Pc model are available. Public elements and stale epochs can materially affect close predictions.

The maneuvers are simplified educational experiments, not spacecraft planning or guaranteed avoidance actions.

## Explicit Non-Goals

- Operational collision avoidance or flight dynamics.
- Collision-probability estimation.
- Covariance propagation or uncertainty-aware operational products.
- Real-time tracking guarantees.
- Accounts, billing, notifications, messaging, or team features.
- Public cloud infrastructure, distributed queues, and production-scale deployment.
- Large-catalog or dense-scene optimization before the MVP vertical slice is polished.

## Data and Reproducibility

The default demo uses a versioned, deterministic fixture catalog and a pinned fixture snapshot date. Synthetic scenarios are labeled as synthetic. Public snapshots, if added, must include retrieval time, source epoch, attribution, and license or usage notes.

Live CelesTrak import is optional and later. It must not be an automatic demo dependency.

## Success Criteria

The MVP succeeds when an unfamiliar engineer can clone, start, explore, replay, and run a bounded what-if scenario locally; inspect verified tests and documentation; and understand that all performance and accuracy claims come from reproducible evidence.
