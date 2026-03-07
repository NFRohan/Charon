# Charon Architecture Plan

## Summary

Charon is a closed-loop university transit platform with two critical workloads:

- High-concurrency financial transactions for ride boarding.
- High-frequency real-time telemetry for bus tracking.

The system is designed to prove strong system design judgment without unnecessary microservice sprawl. The core shape is a Go modular monolith with workers, backed by PostgreSQL, Redis, and RabbitMQ, deployed as a small single-host Docker Compose stack.

## Goals

- Prevent double-spends during flaky mobile network conditions.
- Preserve strict financial correctness with a double-entry ledger.
- Support live bus movement on student maps without overwhelming PostgreSQL.
- Provide a guardian-facing live route view so parents can see where the bus is without exposing student-specific data.
- Keep map rendering costs predictable with MapTiler-backed tiles and local caching in Flutter.
- Archive telemetry and evaluate alerts asynchronously.
- Showcase reliability patterns such as idempotency, transactional outbox, and dead-letter queues.

## Product Surfaces

### Student App (Flutter)

- Wallet balance and recent ledger activity.
- QR scan to pay boarding fare.
- Live route map with moving buses, rendered with MapTiler-backed cached tiles.
- ETA visibility and rider alerts.

### Driver App (Flutter)

- Start and end route session.
- Attach to bus by QR or numeric code after login.
- Send telemetry every 10 seconds over WebSocket.
- Use Android foreground-service behavior for background telemetry in v1.
- Receive driver-facing operational notices if needed.

### Admin App (Next.js + TypeScript)

- Finance operations for wallet credits and refunds.
- Bus registry and durable QR generation.
- Route, stop, and timetable management.
- Live fleet operations view.
- Alerts dashboard.
- Dead-letter queue inspection and requeue tools.

### Public Live View

- Read-only guardian-facing route map.
- Current bus position and route progress.
- Route-level ETA and service advisories.
- No student-level or finance data.

## Deployment Model

- Single university deployment.
- Single-host or small-node Docker Compose environment.
- Core services:
  - `api` (Go)
  - `worker` (Go)
  - `postgres`
  - `redis`
  - `rabbitmq`
  - `reverse-proxy`

This keeps the platform operationally realistic while staying compact enough for a portfolio-grade implementation.

## Core Architecture

### 1. Identity and Access

- Students use student ID plus password; drivers use employee ID plus password with JWT-based access and refresh tokens.
- Roles: `student`, `driver`, `cashier`, `admin`.
- No multi-tenant logic in the first version.

### 2. Fintech Boarding Engine

#### Idempotent API Contract

- Student app generates a UUID and sends it as `Idempotency-Key`.
- Redis stores request state scoped by `actor_id + Idempotency-Key`.
- Request states:
  - `PROCESSING`
  - `COMPLETED`
  - `FAILED`
- Duplicate retries do not re-enter the transactional write path.

#### Double-Entry Ledger

- Store money in integer minor units.
- Use immutable `transactions` and `ledger_entries`.
- Maintain `wallet_accounts.available_balance_minor` as a snapshot for fast reads.
- Every transaction must balance to zero across entries.

#### Concurrency Control

- Lock the student account row with `SELECT ... FOR UPDATE`.
- Validate sufficient balance inside the same database transaction.
- Insert the transaction.
- Insert debit and credit ledger entries.
- Update wallet account snapshots.
- Insert the boarding event.
- Insert an outbox event.
- Commit atomically.

#### Boarding Safety Rules

- Durable signed bus QR contains `bus_id`, `qr_version`, and signature.
- Backend resolves the current boardable service instance from the scanned bus.
- Scheduled boarding opens `30 minutes` before service start and closes `15 minutes` after scheduled end.
- Boarding is rejected if the bus has no boardable service instance or has conflicting active sessions.
- Telemetry freshness is not a hard boarding gate.
- Manual numeric bus-code fallback is always available.
- Unique constraint on `student_id + route_session_id` prevents accidental repeat charge during the same bus run.
- Initial fare policy supports route-level flat fares and zero-fare routes or deployments.
- Small overdraft is deployer-configurable.
- Eligible students may be marked as fare-exempt.

### 3. Transactional Outbox

The system uses the outbox pattern to avoid dual-write failures between PostgreSQL and RabbitMQ.

- Add `outbox_events` table with fields such as:
  - `event_id`
  - `aggregate_type`
  - `aggregate_id`
  - `event_type`
  - `payload_json`
  - `attempt_count`
  - `available_at`
  - `published_at`
  - `last_error`
- Insert outbox rows inside the same Postgres transaction as the business change.
- A publisher worker polls the outbox using `FOR UPDATE SKIP LOCKED`.
- Worker publishes to RabbitMQ with publisher confirms.
- Consumers de-duplicate using `event_id`.

Apply outbox publishing to all events that originate from committed database state, including:

- wallet transactions
- credits and refunds
- route-session lifecycle events
- schedule updates
- alert state changes

### 4. Real-Time Logistics Engine

#### Telemetry Ingestion

- Driver app sends WebSocket telemetry every 10 seconds.
- Payload includes:
  - `route_session_id`
  - `bus_id`
  - `lat`
  - `lng`
  - `speed_kph`
  - `heading`
  - `accuracy_m`
  - `recorded_at`
- Driver app may attach and start from locally cached service data when offline.
- If connectivity drops, the driver app buffers telemetry locally for at least 30 minutes and replays it in order on reconnect.

#### Live Fanout

- API validates telemetry.
- API publishes live data through Redis Pub/Sub.
- API stores last-known bus position in Redis.
- Student clients subscribe to route streams over WebSocket.
- Public live view subscribes to route-safe updates only.
- API pushes `telemetry.update`, `eta.update`, `alert.created`, and `alert.cleared`.

#### Durable Async Processing

- API publishes raw telemetry events to RabbitMQ.
- Archiver worker batches messages in memory.
- Worker bulk-inserts to PostgreSQL every 60 seconds or 1000 points, whichever comes first.

### 5. Mobile Map Rendering and Tile Caching

Map rendering is treated as a cost-sensitive client concern rather than a place to spend on premium per-request APIs.

- Do not use Google Maps as the base map provider.
- Use MapTiler as the map tile and style provider in Flutter.
- Cache map tiles on-device to reduce repeated network fetches and improve resilience under weak campus connectivity.
- Prewarm the cache for campus bounds and all configured route corridors.
- Enforce a bounded cache with eviction so mobile storage usage stays predictable.
- Keep the telemetry overlay separate from the map provider so the base-map vendor can be swapped later without changing the live-tracking protocol.

### 6. RabbitMQ Reliability and DLQs

Each durable queue gets an explicit retry and dead-letter strategy.

- Primary queues:
  - `telemetry.archiver`
  - `alerts.evaluate`
  - `notifications.dispatch`
- Each queue has:
  - bounded retry count
  - dead-letter exchange
  - dedicated dead-letter queue
- After 3 failed attempts, the message moves to a `*.dlq`.
- Poison messages never block healthy traffic.

### 7. Schedules, ETA, and Spatial Logic

Phase 1.1 introduces PostGIS as a first-class dependency.

- Enable PostGIS in PostgreSQL.
- Store stop and telemetry positions as `GEOGRAPHY(Point, 4326)`.
- Store route corridors as line geometry or geography.
- Add GiST indexes for spatial lookups.

Admin-managed scheduling entities:

- `routes`
- `stops`
- `route_stop_sequences`
- `trip_templates`
- `trip_stop_times`
- `route_sessions`

ETA behavior:

- Route session links a live bus to a scheduled trip.
- ETA is rider-specific to the selected stop in the student app.
- ETA is based on schedule plus current live delay.
- If telemetry is stale for 30 seconds, ETA falls back to schedule and is marked stale.

Route deviation behavior:

- Use PostGIS spatial queries rather than ad hoc in-memory geospatial math.
- Detect deviation using distance from route corridor.

### 8. Alerts

Admin alerts:

- route deviation
- late departure
- major delay
- service disruption
- service cancellation
- driver offline

Rider alerts:

- bus approaching selected stop
- major delay
- service disruption
- service cancellation

Delivery:

- in-app real-time notifications by default
- push notifications only for cancellation and major service disruption in the initial release

## Suggested Data Model

Core financial tables:

- `users`
- `wallet_accounts`
- `transactions`
- `ledger_entries`
- `boarding_events`
- `outbox_events`

Core logistics tables:

- `routes`
- `stops`
- `route_stop_sequences`
- `trip_templates`
- `trip_stop_times`
- `route_sessions`
- `telemetry_points`
- `alerts`
- `finance_adjustments`
- `service_calendars`
- `service_exceptions`
- `service_advisories`
- `device_tokens`

## Public Interfaces

REST endpoints:

- `POST /auth/login`
- `POST /auth/refresh`
- `POST /auth/logout`
- `GET /wallet/balance`
- `GET /wallet/transactions`
- `POST /boardings`
- `POST /admin/wallets/{id}/credits`
- `POST /admin/wallets/{id}/refunds`
- `GET /routes`
- `GET /routes/{id}/stops`
- `GET /routes/{id}/eta`
- `POST /admin/routes`
- `POST /admin/stops`
- `POST /admin/trips`
- `POST /route-sessions/start`
- `POST /route-sessions/end`

WebSocket message types:

- `driver.telemetry`
- `student.subscribe_route`
- `student.unsubscribe_route`
- `telemetry.update`
- `eta.update`
- `alert.created`
- `alert.cleared`

## Delivery Phases

### Phase 1

- auth and RBAC
- wallet ledger and account snapshots
- route-level fare configuration, overdraft support, and fare exemptions
- bus registry and durable QR issuance
- idempotent QR boarding
- transactional outbox
- MapTiler provider choice and Flutter tile-cache setup
- live telemetry fanout
- telemetry archival pipeline
- cashier-issued credits and refunds with admin approval above configured thresholds

### Phase 1.1

- PostGIS enablement
- route, stop, and timetable management
- holiday and special-event schedule exceptions
- route sessions linked to schedules
- ETA engine
- route deviation detection

### Phase 1.2

- admin alerts
- rider alerts
- limited push notifications for service cancellation and major disruption
- guardian-facing public live route view
- dead-letter queue tooling and alert operations

## Validation and Testing

- Concurrent boarding tests with dozens of simultaneous riders.
- Duplicate request tests using the same idempotency key.
- Low-balance contention tests with many distinct request keys.
- Ledger invariant tests ensuring balanced entries.
- Outbox crash-recovery tests for dual-write safety.
- Telemetry load tests with multiple buses and many subscribers.
- PostGIS route deviation and ETA tests.
- DLQ tests with malformed messages and retry exhaustion.
- Alert creation, deduplication, and clearing tests.

## Defaults and Assumptions

- Single university only.
- Institutional ID plus password login.
- Route-level flat fare or zero-fare policy initially.
- Admin-managed schedules are the source of truth.
- Wallet funding starts with cashier-issued credits and refunds, with configurable approval limits and room for external campus credit integration later.
- Base map rendering uses MapTiler with on-device caching instead of Google Maps.
- Driver telemetry buffers and replays after network loss.
- Weekly timetable supports holiday and special-event exceptions.
- Durable boarding QR is generated from the admin side and attached to the physical bus.
- Schedule-backed service windows, not live telemetry freshness, control whether boarding is valid.
- Guardian live view is route-level only and never exposes student-tracking data.
- 30-day telemetry retention is enough for the demo environment.
- Redis is ephemeral and used for idempotency, presence, subscriptions, and current position data.
- PostgreSQL is the source of truth for money, schedules, alerts, and historical telemetry.
- RabbitMQ carries durable async workflows and outbox-delivered domain events.
