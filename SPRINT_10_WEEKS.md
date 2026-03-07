# Charon 10-Week Sprint Plan

## Summary

This sprint plan turns the architecture into a 10-week build roadmap optimized for a portfolio-grade implementation. The sequence prioritizes the hardest proof points first:

- financial correctness under concurrency
- reliable event propagation
- low-cost map rendering with client-side tile caching
- real-time telemetry without database thrash
- schedules, ETA, and alerts on top of a stable core

## Sprint Assumptions

- One primary engineer or small team.
- Single-host Docker Compose deployment target.
- Backend in Go, mobile in Flutter, admin in Next.js.
- Focus on a polished demo and strong system-design evidence, not campus-wide operational completeness.

## Week 1: Foundation and Project Skeleton

### Goals

- Establish repo structure and local development stack.
- Stand up infrastructure dependencies.
- Define shared contracts and domain boundaries.

### Work

- Create monorepo layout for `api`, `worker`, `student_app`, `driver_app`, and `admin_app`.
- Add Docker Compose for Postgres, Redis, RabbitMQ, API, and worker services.
- Initialize Go service structure and configuration loading.
- Enable JWT auth skeleton and role model, including student ID and driver employee ID login paths.
- Lock MapTiler as the base map provider and define Flutter tile-cache strategy, storage cap, and API-key handling.
- Draft initial schema migrations for users, wallet accounts, routes, and route sessions.

### Exit Criteria

- Full stack boots locally with one command.
- Auth skeleton is wired end to end.
- Map provider and cache strategy are documented before mobile map work starts.
- Database migrations run successfully.

## Week 2: Wallet Ledger and Financial Core

### Goals

- Implement the financial source of truth correctly before UI polish.

### Work

- Create `transactions`, `ledger_entries`, and `wallet_accounts` schema.
- Implement double-entry ledger rules using integer minor units.
- Add balance snapshot updates inside the same transaction.
- Add route-level and selected-stop fare configuration, overdraft limits, and fare-exemption support.
- Add finance-adjustment audit fields for before and after values, reason codes, and approval trail.
- Build wallet read APIs for balance and transaction history.
- Seed test users and starting balances.

### Exit Criteria

- Credits and debits create balanced ledger entries.
- Wallet balance reads come from account snapshots.
- Cashier credit and refund rules respect configured limits and audit requirements.
- No money operation can commit partial rows.

## Week 3: QR Boarding, Idempotency, and Concurrency Safety

### Goals

- Prove the boarding flow is safe under retries and concurrent requests.

### Work

- Build bus registry and durable signed QR generation from the admin side.
- Build schedule-backed service windows with `30 minute` early boarding and `15 minute` late grace.
- Build route-session and service-instance binding so the scanned bus resolves to the current boardable run rather than a driver-held QR.
- Implement `POST /boardings` with Redis-backed idempotency state.
- Add student stop selection, favorite-stop shortcuts, and fare preview before final confirmation.
- Add sponsored boarding for self plus one additional rider with atomic rider-level event creation.
- Add emergency ride permit issuance, local consumption rules, and backend redemption flow.
- Add manual numeric bus-code fallback and scan-audit logging.
- Add `SELECT ... FOR UPDATE` locking on the student wallet account.
- Insert `boarding_events`, fare-decision tracking, and unique duplicate-prevention constraints.
- Write concurrency and retry tests.

### Exit Criteria

- Same idempotency key never charges twice.
- Concurrent boarding requests cannot produce negative balances.
- Driver can start a session without needing to display a phone for boarding.
- Durable admin-issued bus QR works for boarding when the bus is in an active route session.
- Boarding still works when telemetry is stale, as long as the schedule-backed service window is valid.
- Student can select a stop and see the resolved fare before confirming payment.
- A student who loses internet can still board through a sponsor or a valid emergency ride permit.
- Manual numeric bus-code fallback works through the same authorization path.

## Week 4: Transactional Outbox and Reliable Domain Events

### Goals

- Eliminate dual-write risk between Postgres and RabbitMQ.

### Work

- Add `outbox_events` schema and event envelope format.
- Insert outbox rows in the same transaction as ledger and route-session changes.
- Build outbox publisher worker with polling and `FOR UPDATE SKIP LOCKED`.
- Add publisher confirm handling and retry backoff.
- Define downstream consumer deduplication by `event_id`.

### Exit Criteria

- Financial and route-session events are published only from the outbox.
- Crash after DB commit does not lose downstream events.
- Outbox backlog can drain safely after recovery.

## Week 5: Real-Time Telemetry and Live Map

### Goals

- Deliver a fast live-tracking path that bypasses PostgreSQL for current position updates.

### Work

- Implement driver telemetry WebSocket endpoint.
- Validate telemetry payloads and route-session ownership.
- Publish telemetry to Redis Pub/Sub.
- Track last-known bus positions in Redis.
- Implement Android foreground-service telemetry flow on the driver app.
- Buffer telemetry locally for at least 30 minutes during connectivity loss and replay it in order after reconnect.
- Add offline attach or start support when the relevant service data is already cached locally.
- Add battery-optimization risk detection and active-service state restoration after app restart.
- Build the student app shell with `Home`, `Wallet`, `Map`, `Alerts`, and `Profile`.
- Implement student route subscription WebSocket flow.
- Render moving buses on the student map using MapTiler-backed tiles.
- Implement local tile caching and prewarm the campus viewport and primary routes.

### Exit Criteria

- Driver app streams telemetry every 10 seconds.
- Student app receives live map updates with low latency.
- Replayed telemetry is archived correctly after temporary disconnection.
- Driver app restores active service state after restart and continues telemetry correctly.
- Student map remains usable from cached tiles during brief connectivity loss.
- PostgreSQL is not in the synchronous live-tracking path.

## Week 6: RabbitMQ Archival and Admin Finance Operations

### Goals

- Make telemetry durable asynchronously and expose financial operations to admins.

### Work

- Publish raw telemetry events to RabbitMQ.
- Build archiver worker with batch insert behavior.
- Create admin wallet credit and refund flows.
- Add admin finance views for balances and transaction history.
- Add basic operational observability for queue depth and worker health.

### Exit Criteria

- Telemetry is archived in Postgres in bulk batches.
- Admin can issue wallet credits and refunds.
- Archiver can recover cleanly from worker restarts.

## Week 7: Schedules, Stops, and Route Session Planning

### Goals

- Introduce schedule-aware transit operations.

### Work

- Build admin CRUD for routes, stops, stop order, trip templates, and stop times.
- Add weekly schedule exceptions for holidays and special events.
- Add service advisory management for closures and disruptions.
- Link driver route sessions to a scheduled trip.
- Expose route and stop APIs for student and driver clients.
- Add admin views for timetable management.

### Exit Criteria

- Admin can create a route with ordered stops and scheduled trip times.
- Admin can represent holiday closures and special-event schedule changes.
- A driver can start a route session tied to a published trip.
- Student app can fetch route and stop data.

## Week 8: PostGIS, ETA, and Route Deviation

### Goals

- Add spatial intelligence on top of the telemetry pipeline.

### Work

- Enable PostGIS and migrate location fields to geography types.
- Add GiST indexes and route corridor geometry.
- Build ETA calculation using schedule plus live delay.
- Detect route deviation with spatial queries.
- Surface ETA in student and admin interfaces.

### Exit Criteria

- ETA updates are computed from live session plus timetable.
- Stale telemetry triggers schedule fallback.
- Route deviation detection works using PostGIS queries.

## Week 9: Alerts, Push Notifications, and DLQs

### Goals

- Turn telemetry and schedule data into actionable notifications.

### Work

- Build alert evaluation worker for route deviation, late departure, major delay, disruption, cancellation, and admin-only driver-offline alerts.
- Add FCM device-token registration and push integration.
- Configure RabbitMQ retry limits, dead-letter exchanges, and per-queue DLQs.
- Add admin alert dashboard and DLQ inspection/requeue view.
- Build the guardian-facing public live route view with route-safe telemetry, ETA, and advisories.
- Implement alert deduplication and clear logic.

### Exit Criteria

- Operator and rider alerts are created from system events.
- Push notifications are limited to cancellation and major disruption cases.
- Guardians can open a public-safe live route page and see bus position, route progress, and service state.
- Failed async messages land in the right DLQ after retry exhaustion.
- Admin can inspect and requeue failed messages.

## Week 10: Hardening, Load Testing, and Demo Readiness

### Goals

- Prove the system works under pressure and package it as a strong showcase.

### Work

- Run concurrency tests for 100-boarding burst scenarios and ledger invariants.
- Run telemetry load tests with multiple buses and many subscribers.
- Test outbox recovery, worker crashes, stale ETA fallback, and poison messages.
- Clean up UX in mobile and admin surfaces.
- Add architecture diagrams, seed scripts, and demo walkthrough docs.

### Exit Criteria

- Core failure scenarios are tested and reproducible.
- Demo environment can be started and explained quickly.
- Project has clear documentation for architecture, setup, and test evidence.

## Milestones

- End of Week 3: financial concurrency proof is working.
- End of Week 5: live map demo is working.
- End of Week 8: schedule-aware ETA and spatial checks are working.
- End of Week 10: complete showcase demo is ready.

## Acceptance Checklist for the Full 10 Weeks

- Student can log in, view balance, scan QR, and pay exactly once.
- Student can select a stop for boarding and receive stop-specific ETA.
- Student can still complete boarding when solo and temporarily offline through a bounded emergency permit path.
- Driver can start a route session and stream telemetry continuously.
- Admin can credit wallets, manage routes and schedules, and inspect alerts.
- Guardians can use the public live route page without seeing student-specific data.
- Outbox guarantees downstream event delivery after DB commit.
- RabbitMQ queues have bounded retries and explicit dead-letter queues.
- Historical telemetry is archived in Postgres in batches.
- Student map uses cached MapTiler tiles instead of a high-cost Google Maps dependency.
- ETA and route deviation use PostGIS-backed spatial logic.
- Rider and operator alerts work in-app and via push where enabled.
