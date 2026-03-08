# Charon Phase Test Plan

## Purpose

This document defines the testing strategy for each delivery phase of Charon.

It is not a generic QA wishlist. It is a phase gate. Each phase should end with:

- clear automated coverage for the highest-risk logic added in that phase
- targeted manual checks where automation is weak or too expensive early on
- saved evidence that the phase is actually stable enough to build on

## Testing Principles

- Prefer the lowest test level that can prove the behavior clearly.
- Use Go's standard `testing` package, table-driven tests, and `httptest` by default.
- Use real infrastructure for correctness-sensitive integration tests instead of mocks when persistence, transactions, Redis, or RabbitMQ behavior matters.
- Add regression tests for every bug that reaches manual testing or integration testing.
- Treat money, audit, idempotency, and privacy as invariant-heavy domains that need repeated verification at more than one layer.
- Keep end-to-end tests narrow and high-signal. Most behavior should be proven below full-stack UI automation.

## Repository Test Layout

Recommended test layout as implementation grows:

- `backend/internal/.../*_test.go`
  Domain, handler, config, and small integration tests close to the code.
- `backend/tests/integration/`
  Cross-module tests that need real Postgres, Redis, or RabbitMQ.
- `backend/tests/load/`
  Boarding-burst, telemetry-throughput, and outbox-drain benchmarks or harnesses.
- `apps/admin_app/e2e/`
  Browser-driven operator workflows once the admin UI becomes interactive.
- `apps/student_app/test/`
  Widget and lightweight app-layer tests.
- `apps/driver_app/test/`
  Widget and state tests, especially around attach flow and telemetry status.

## Test Environments

### 1. Fast Local

Use for unit tests and handler tests.

- `go test ./...`
- `flutter test`
- `npm run lint`
- `npm run build`

### 2. Compose Integration

Use for backend integration tests that require real infrastructure.

- PostgreSQL with PostGIS
- Redis
- RabbitMQ
- API and worker binaries

This environment should be the default for migration tests, ledger tests, outbox tests, and queue behavior.

### 3. CI

Use GitHub Actions for repeatable verification.

- backend unit tests
- backend integration suite
- admin build and lint
- Flutter test suites
- APK build artifacts for student and driver apps

### 4. Real Device Validation

Use cloud-built APK artifacts for phone testing.

- student boarding UX
- driver attach and telemetry behavior
- background behavior, reconnects, and permission prompts

### 5. Load and Failure Lab

Use this only for targeted proof points.

- boarding concurrency bursts
- telemetry fanout
- outbox backlog drain
- DLQ behavior
- stale dependency and restart scenarios

## Core Invariants

These are cross-phase invariants and must keep passing once introduced:

- every committed financial transaction balances to zero across ledger entries
- no wallet debit can push an account below the allowed overdraft threshold
- repeated idempotent requests never create duplicate money movement
- audit events are immutable after insert
- public live-view payloads never expose student-specific data
- live telemetry fanout does not require synchronous PostgreSQL writes
- downstream event publication from database-owned facts always goes through the outbox

## Phase Gates

### Phase 1: Foundation and Bootstrap

Weeks covered:
- Weeks 1-3

### Automated

- config validation tests for required environment variables and supported environments
- API boot test that confirms the router starts and `/healthz` returns `200`
- migration command smoke test against a fresh Postgres instance
- migration status test after applying the initial migration
- seed runner test for empty or missing seed directories
- Docker Compose config parse check

### Manual

- run `./scripts/dev-up.ps1`
- confirm Postgres, Redis, RabbitMQ, API, and worker start
- confirm RabbitMQ management UI is reachable
- confirm `./scripts/dev-down.ps1` tears the stack down cleanly

### Exit Evidence

- saved command output for `go test ./...`
- saved command output for `docker compose -f deploy/docker-compose.yml config`
- screenshot or text proof of successful `/healthz` response

### Phase 2: Auth, Wallet, Boarding, and Outbox Core

Weeks covered:
- Weeks 2-8

### Automated

- unit tests for login, refresh, logout, and role parsing
- middleware tests for role enforcement and unauthorized access
- ledger tests for balanced debit or credit creation
- overdraft-limit tests
- fare-exempt and zero-fare tests
- cashier credit and refund authority tests
- idempotency tests for duplicate `Idempotency-Key` handling
- boarding preview tests for route-flat, zero-fare, and stop-matrix pricing
- boarding submit tests for standard, sponsored, and emergency sync modes
- duplicate-boarding prevention tests per rider and service instance
- transactional outbox insertion tests inside the same database transaction as money movement
- outbox publisher retry and publish-marking tests

### Integration

- real Postgres transaction tests proving row locking under concurrent debits
- Redis-backed idempotency tests with repeated boarding retries
- outbox worker test where DB commit succeeds and publish occurs later
- crash-recovery scenario where unpublished outbox rows are drained after restart

### Load

- 100 concurrent boarding attempts against one rider with different keys and low balance
- 100 retries with the same idempotency key
- mixed sponsored and standard boarding burst against the same service window

### Manual

- inspect ledger rows and account snapshots after several boarding scenarios
- confirm audit data is written for credits, refunds, and boarding events

### Exit Evidence

- concurrency test results showing no duplicate charge and no invalid negative balance
- outbox recovery report showing eventual publish after simulated interruption
- saved SQL or logs proving balanced ledger writes

### Phase 3: Admin Finance Ops and Transit Setup

Weeks covered:
- Weeks 8-10

### Automated

- student search query tests
- refund approval workflow tests
- audit immutability tests
- investigation-note append-only tests
- route CRUD validation tests
- stop ordering validation tests
- stop-matrix fare validation tests
- weekly schedule bitmask tests
- schedule exception tests for cancellation and time override
- durable QR generation and rotation-grace tests
- service-window calculation tests including early boarding and late grace

### Integration

- route publication flow from admin input through persisted route or stop structure
- service-instance creation and cancellation tests
- driver-attach eligibility resolution tests

### Manual

- verify finance flows from the admin app or API client with realistic sample data
- verify QR metadata and rotation behavior from bus management flow

### Exit Evidence

- API-level proof that a route can be created with ordered stops and fare rules
- test output for refund approval and audit-note behavior

### Phase 4: Student App Core

Weeks covered:
- Weeks 9-11

### Automated

- widget tests for login, wallet summary, boarding confirmation, and receipt views
- state tests for session persistence, manual retry behavior, and favorite-route or stop handling
- client-side validation tests for permission-denied and location-warning branches

### Integration

- contract tests against the live auth, wallet, boarding preview, and boarding submit APIs
- test flow for sponsored boarding confirmation and failure states
- emergency permit sync flow test against a seeded account

### Manual on Real Devices

- login, logout, and session restore
- wallet refresh after top-up or refund
- QR scan and manual bus-code fallback
- boarding confirmation with stop selection
- permission denied flow for location check
- manual retry after forced network drop
- install and smoke test from GitHub Actions APK artifact

### Exit Evidence

- phone-recorded or screenshot evidence of login, wallet, and boarding success
- issue log for UX problems found during phone testing and their follow-up status

### Phase 5: Driver App and Telemetry Pipeline

Weeks covered:
- Weeks 12-14

### Automated

- driver attachment preview tests
- driver attach conflict tests
- start and end journey authorization tests
- WebSocket message validation tests for telemetry payload shape
- telemetry ack or nack tests
- replay ordering tests
- replay age limit tests
- archival batch tests
- DLQ routing tests after bounded worker failure

### Integration

- end-to-end driver attach, start, telemetry send, archive, and replay flow
- Redis live-position fanout tests
- RabbitMQ archival and retry tests

### Manual on Real Devices

- driver login and bus attach
- attach conflict behavior when another driver is already attached
- background telemetry with screen locked
- reconnect and replay after temporary network loss
- app restart while attached to an active service
- battery-optimization warning path

### Exit Evidence

- archived telemetry rows for a real device session
- logs showing telemetry ack or nack and successful replay after reconnect

### Phase 6: Spatial Intelligence, Alerts, and Public Live View

Weeks covered:
- Weeks 15-16

### Automated

- ETA calculation tests for on-time, delayed, and stale-telemetry fallback cases
- route-deviation tests using PostGIS distance checks
- alert creation tests for route deviation, late departure, major delay, disruption, and cancellation
- alert deduplication and clear-condition tests
- public payload privacy tests proving student identifiers never leak
- advisory publication tests

### Integration

- live telemetry to ETA update flow
- telemetry to alert evaluation flow
- alert to public advisory visibility flow

### Manual

- verify student map updates and ETA display on a phone
- verify guardian live view only exposes route-safe bus data
- verify public advisory text reflects route or campus scope correctly

### Exit Evidence

- saved ETA and route-deviation test results
- privacy check output for public live-view responses

### Phase 7: Technical Ops, Admin Web Completion, and End-to-End Integration

Weeks covered:
- Weeks 17-19

### Automated

- RBAC tests for admin, cashier, and technical-admin actions
- import validation tests
- export job lifecycle tests
- audit browsing and investigation-note retrieval tests
- admin web component and page tests for key workflows
- browser-driven tests for finance approval, bus QR management, route setup, and alert viewing

### Integration

- end-to-end flows covering:
- cashier top-up to student wallet refresh
- admin route setup to student boarding availability
- admin service cancellation to student alert visibility
- driver telemetry to public live-view update

### Manual

- smoke test all four surfaces together: admin, API, student, and driver
- confirm role-scoped navigation and hidden actions in the admin UI
- verify DLQ inspection flow from the technical-admin surface

### Exit Evidence

- end-to-end checklist with pass or fail status for each cross-surface workflow
- screenshots or recordings of critical admin workflows

### Phase 8: Hardening, Performance, and Showcase Readiness

Weeks covered:
- Week 20

### Automated

- 100-user boarding burst benchmark
- telemetry throughput benchmark for the expected fleet size plus headroom
- outbox backlog drain benchmark
- stale dependency and restart regression tests
- negative tests for malformed QR, malformed telemetry, poisoned queue messages, and invalid auth flows

### Failure Drills

- API restart during active use
- worker restart during queue backlog
- Redis restart with live map subscribers connected
- RabbitMQ poison message entering DLQ
- Postgres restart during non-interactive workloads

### Manual

- full demo walkthrough from clean start
- docs check: setup, architecture, and testing evidence all current
- final real-device smoke on student and driver apps

### Exit Evidence

- saved load-test results
- failure-drill report with observed behavior and recovery time
- demo checklist marked ready

## Minimum Tooling by Layer

- Go backend:
  `testing`, `httptest`, table-driven tests, real Postgres integration where transaction behavior matters
- Admin web:
  unit or component tests plus Playwright-style browser flows once the UI becomes interactive
- Flutter:
  widget tests for local confidence, manual real-device validation for permission, background, and connectivity behaviors
- Load:
  `k6`, `vegeta`, or a focused Go harness depending on whether the target is HTTP or WebSocket-heavy

## Phase Exit Rule

Do not call a phase complete only because the feature exists.

A phase is complete when:

- its automated suite is green
- the required manual checks are done
- evidence is saved in the repo, CI artifacts, or engineering notes
- known failures are either fixed or explicitly accepted with written follow-up
