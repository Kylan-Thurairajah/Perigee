# Perigee Agent Guide

## Purpose

Perigee is a local-first orbital traffic visualization and conjunction-screening portfolio application. It combines a Go screening engine with a React and TypeScript desktop-first visualization.

## Repository Status

The repository is actively implemented. Documentation describes intended or verified behavior; do not treat plans as implemented features without checking the source and tests.

Do not report a feature, test, benchmark, or numerical claim as complete unless it has been verified in the repository.

## Non-Negotiable Boundaries

- Do not implement authentication, accounts, cloud infrastructure, Kubernetes, Kafka, notifications, billing, operational alerting, or distributed microservices.
- Keep all demos usable offline after installation with deterministic, attributed fixtures.
- Do not fake or advertise 25K objects, 14 seconds, 120x, 96 percent recall, or 2,000 reference events. Such numbers require reproducible measurements.
- Distinguish conjunction screening and miss distance from collision probability.
- Do not present a maneuver as operational guidance or guaranteed collision avoidance.
- Do not modify application source files or create `contracts/openapi.yaml` unless explicitly assigned as the lead/integrator or as part of a coordinated cross-cutting change.

## Work Ownership

### Lead / integrator

Owns coordination, `contracts/openapi.yaml`, `docs/ARCHITECTURE.md`, `docs/PROGRESS.md`, shared types, build and CI integration, and final integration review. Cross-cutting API changes require lead coordination.

The lead owns this repository audit and phase gates. The lead and affected specialists agree before shared interfaces change.

### Product design + UI engineering

May modify `apps/web/src/app/`, `apps/web/src/components/`, `apps/web/src/features/`, `apps/web/src/styles/`, and `docs/DESIGN.md`.

May not modify the Go backend, data or propagation logic, scene files, or `contracts/openapi.yaml` without lead coordination.

### 3D visualization

Owns `apps/web/src/scene/` and scene-focused tests. Scene work must use typed props and callbacks agreed with UI engineering and the lead.

May not modify shared UI layout, HTTP API, or Go backend.

### Go engine + API

Owns `apps/api/`, backend tests and benchmarks, `docs/VALIDATION.md`, and the backend portion of data-source documentation.

May not modify UI assets or design. Endpoint changes must be coordinated through the lead.

### QA + docs + demo

Owns `apps/web/src/test/` or the e2e test directory, `docs/DEMO_SCRIPT.md`, `docs/BENCHMARKS.md`, README media plans, media capture directions, and demo test plans.

May make only small documented fixes in owner code. Capture media only from a functioning implementation.

## Change Rules

- Make small, reviewable, testable changes.
- Preserve existing useful code instead of rewriting it to match the plan.
- Use the existing package manager and build system once they exist.
- Run the smallest relevant validation immediately; run full relevant checks at integration.
- Record commands, failures, unresolved assumptions, and remaining work in `docs/PROGRESS.md`.
- Update directly related documentation when behavior or interfaces change.

## Scientific Rules

- Use a single documented canonical frame and unit system for calculations.
- Timestamps use RFC3339 UTC with a terminal Z.
- Numerical API fields use explicit units in names or an explicit schema unit.
- A minimum separation is a miss distance, not a collision probability.
- TLE/OMM uncertainty and stale epochs are material; fixture and public data must be labeled.
- SGP4 output requires library provenance and reference-vector validation.
- Applying a maneuver to an SGP4 state is not assumed to be physically direct. Simplified post-maneuver models must be stated and limited.

## Definition of Completion

A milestone is complete only after all acceptance criteria in `docs/PROGRESS.md` and the master plan are verified. Documentation-only completion must be separated from implementation completion.
