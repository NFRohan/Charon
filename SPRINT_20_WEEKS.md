# Charon 20-Week Sprint Plan

## Summary

This sprint plan turns the architecture into a 20-week build roadmap optimized for a portfolio-grade implementation. The sequence is backend-first, but not frontend-last. The high-risk correctness work lands early, and the mobile apps begin once their core APIs are stable enough to test on real devices through GitHub Actions APK builds.

- financial correctness under concurrency
- reliable event propagation
- low-cost map rendering with client-side tile caching
- real-time telemetry without database thrash
- schedules, ETA, and alerts on top of a stable core
- mobile clients brought in mid-plan to validate UX and API shape before final hardening

## Sprint Assumptions

- One primary engineer or small team.
- Single-host Docker Compose deployment target.
- Backend in Go, mobile in Flutter, admin in Next.js.
- Focus on a polished demo and strong system-design evidence, not campus-wide operational completeness.
- GitHub Actions can build Flutter APK artifacts for phone testing even when the local Android toolchain is not the primary development path.
- The backend still leads the schedule because concurrency, ledger safety, and telemetry reliability are the hardest proof points.

## Week 1: Repository Foundation and Local Stack

### Goals

- Establish repo structure and local development stack.
- Stand up infrastructure dependencies.
- Freeze top-level module boundaries.

### Work

- Create monorepo layout for `api`, `worker`, `student_app`, `driver_app`, and `admin_app`.
- Add Docker Compose for Postgres, Redis, RabbitMQ, API, and worker services.
- Initialize Go service structure and configuration loading.
- Lock domain seams for `auth`, `wallet`, `boarding`, `routes`, `telemetry`, `alerts`, and `outbox`.
- Draft initial schema migration strategy and local seed-data approach.

### Exit Criteria

- Full stack boots locally with one command.
- Compose file parses and infrastructure services start cleanly.
- The repo has one agreed code layout and one agreed runtime model.

## Week 2: Auth and Session Foundation

### Goals

- Lock the identity model before business endpoints appear.

### Work

- Implement unified `POST /auth/login`, refresh, and logout flows.
- Model student, driver, admin, cashier, and technical-admin roles.
- Add student-ID and employee-ID credential paths.
- Add refresh-token persistence and current-device logout handling.
- Add auth middleware and role enforcement skeleton.

### Exit Criteria

- Protected endpoints can read authenticated identity and role context.
- Login, refresh, and logout flows are testable locally.
- Session rules match the API spec.

## Week 3: Core Schema and Migration Baseline

### Goals

- Establish the durable data model before service logic sprawls.

### Work

- Create migrations for users, wallet accounts, transactions, ledger entries, buses, routes, stops, service templates, service instances, boarding events, audit logs, outbox events, and alert records.
- Add PostGIS extension bootstrap.
- Create seed fixtures for students, drivers, routes, buses, and service templates.
- Define migration and seed execution flow for local development.

### Exit Criteria

- The base schema can be created from zero on a fresh database.
- Seed data supports development of wallet, boarding, and route flows.
- Data model matches the current specs closely enough to start service implementation.

## Week 4: Wallet Ledger and Financial Core

### Goals

- Implement the financial source of truth correctly before UI polish.

### Work

- Implement `wallet_accounts`, `transactions`, and `ledger_entries`.
- Use integer minor units only.
- Add balance snapshot updates inside the same transaction.
- Add overdraft-limit and fare-exemption support.
- Add full before-or-after adjustment audit payloads and approval chain fields.
- Build wallet read APIs for balance and transaction history.

### Exit Criteria

- Ledger entries stay balanced.
- Wallet reads come from account snapshots.
- Partial money writes cannot commit.

## Week 5: Fare Engine, Bus Registry, and Durable QR

### Goals

- Lock the boarding prerequisites before implementing payment capture.

### Work

- Build bus registry and durable QR generation metadata flow.
- Add QR rotation and one-day grace-period behavior.
- Implement route-flat, zero-fare, and stop-matrix fare policies.
- Add service-label fare overrides where allowed.
- Build schedule-backed service windows with `30 minute` early boarding and `15 minute` late grace.
- Add manual numeric bus-code fallback support.

### Exit Criteria

- Admin-defined buses, fare policies, and service windows are resolvable by backend services.
- Durable QR validation works from bus identity plus active service lookup.
- Fare preview inputs are fully server-owned.

## Week 6: Boarding Preview, Idempotency, and Concurrency Safety

### Goals

- Prove the boarding flow is safe under retries and concurrent requests.

### Work

- Implement `GET /boardings/preview`.
- Implement unified `POST /boardings` with `standard`, `sponsored`, and `emergency_sync` modes.
- Add Redis-backed idempotency handling for money-moving operations.
- Add `SELECT ... FOR UPDATE` locking for wallet debits.
- Add boarding-event creation, rider-level duplicate prevention, and fare-decision persistence.
- Add scan audit records and route-session resolution from QR or bus code.
- Write concurrency and retry tests.

### Exit Criteria

- Same idempotency key never charges twice.
- Concurrent boarding requests cannot create negative balances.
- Sponsored boarding and emergency sync follow the same ledger rules as standard boarding.

## Week 7: Transactional Outbox and Reliable Domain Events

### Goals

- Eliminate dual-write risk between Postgres and RabbitMQ.

### Work

- Finalize `outbox_events` schema and event envelope.
- Insert outbox rows in the same transaction as ledger, service-instance, and advisory changes.
- Build outbox publisher worker with polling and `FOR UPDATE SKIP LOCKED`.
- Add publisher confirms, retry backoff, and consumer deduplication rules.
- Add baseline GitHub Actions for backend tests and admin builds.

### Exit Criteria

- DB-originated events leave the system only through the outbox.
- Crash after commit does not lose downstream events.
- Outbox backlog drains safely after restart.
- CI runs cleanly for the backend and web app.

## Week 8: Admin and Cashier Finance Operations Plus Mobile CI Bootstrap

### Goals

- Expose finance workflows to the operations surface.
- Make phone-based mobile validation possible before full local Android setup.

### Work

- Implement student search, wallet summary, transaction history, credits, refunds, and approval-required paths.
- Implement refund approval and rejection flows.
- Add immutable audit-log writes plus append-only investigation notes.
- Build admin dashboard widgets for finance and system health.
- Add GitHub Actions workflows to build the student and driver APKs as downloadable artifacts.

### Exit Criteria

- Cashiers can do finance lookup, credits, and in-limit refunds.
- Admin approvals work for over-limit refunds.
- Audit behavior matches the compliance model.
- Student and driver APKs can be built in the cloud for real-device testing.

## Week 9: Student App Foundation and Wallet Flows

### Goals

- Start real student-app implementation once auth and wallet APIs are stable.

### Work

- Implement student auth and session persistence.
- Build `Home`, `Wallet`, and `Profile` flows against live APIs.
- Show balance, overdraft, exemption state, and recent transactions.
- Add basic alerts list and local route-or-stop favorites persistence.
- Add CI artifact distribution notes for phone install testing.

### Exit Criteria

- Student can log in on a real phone build.
- Wallet and profile flows work against the backend.
- Mobile work starts validating API shape before boarding UX is layered in.

## Week 10: Routes, Stops, Fare Matrices, and Weekly Schedules

### Goals

- Introduce schedule-aware transit administration.

### Work

- Implement admin CRUD for routes, stops, stop order, and stop-based fare matrices.
- Add weekly service templates with bitmask weekday selection.
- Add route-level default fare and stop-based matrix validation.
- Add exception handling for cancellations and time overrides.
- Expose route and stop read APIs.

### Exit Criteria

- Admin can publish a route with ordered stops and either flat or stop-matrix fare policy.
- Weekly schedules and exceptions are queryable by backend services.
- Route and stop data are stable enough for service-instance work.

## Week 11: Student Boarding Flows

### Goals

- Put the most important student interaction on a real device while the backend is still malleable.

### Work

- Implement `Scan to Pay` preview and confirmation.
- Add stop selection, location-warning UX, and manual bus-code fallback.
- Implement sponsored boarding UI for self plus one rider.
- Implement emergency-ride permit sync UX.
- Validate boarding receipts and retry messaging on phone builds.

### Exit Criteria

- Student can complete the full boarding flow from a real device.
- Sponsored and emergency paths can be exercised against the backend.
- UX problems surface before telemetry and maps are added.

## Week 12: Service Instances, Ad Hoc Trips, and Driver Attachment APIs

### Goals

- Lock operational trip control before telemetry is added.

### Work

- Implement service-instance creation, cancellation, and force-close flows.
- Implement driver bus-attachment preview and attach APIs.
- Implement driver start and end journey endpoints.
- Add driver device-health reporting endpoints.
- Add notice delivery for drivers.

### Exit Criteria

- Drivers can attach to eligible services with QR or bus code.
- Service instances can be scheduled, ad hoc, cancelled, or force-closed.
- Driver conflict handling is explicit and auditable.

## Week 13: Telemetry Ingestion and Driver App Foundation

### Goals

- Start real driver-device work as soon as attach and start semantics are stable.

### Work

- Implement the shared WebSocket backbone.
- Implement driver telemetry upload and per-message ack or nack.
- Validate telemetry ownership against the attached driver service.
- Build driver login, attach-bus, and start-journey flows in the app.
- Add key status indicators for GPS, network, service state, and backlog count.

### Exit Criteria

- Driver telemetry is accepted over one socket with explicit ack behavior.
- Driver can authenticate and attach to a service from a phone build.
- Current-position updates do not hit PostgreSQL synchronously.

## Week 14: RabbitMQ Archival, DLQs, and Driver Replay

### Goals

- Make telemetry durable under failure and verify the driver-side offline path.

### Work

- Publish raw telemetry events to RabbitMQ.
- Build telemetry archival worker with batch inserts.
- Configure retry limits, DLQs, and poison-message handling.
- Implement foreground telemetry buffering, replay, and battery-risk messaging in the driver app.
- Add queue health metrics and basic operational visibility.

### Exit Criteria

- Telemetry is archived in Postgres in batches.
- Poison messages land in DLQs after bounded retries.
- Buffered driver telemetry replays correctly after reconnect.

## Week 15: PostGIS, ETA, Student Map, and Route Deviation

### Goals

- Add spatial intelligence and expose it through the student experience.

### Work

- Finalize geography fields and spatial indexes.
- Add route corridor geometry.
- Build ETA calculation from schedule plus live delay.
- Detect route deviation with spatial queries.
- Implement student live map integration with MapTiler tiles and caching.
- Add stale-telemetry fallback rules.

### Exit Criteria

- ETA updates are computed from live service plus timetable.
- Route deviation detection works via PostGIS queries.
- Student map shows live buses with cached tile behavior.

## Week 16: Alerts, Advisories, Public Live View, and Driver Polish

### Goals

- Turn schedule and telemetry state into actionable system outputs.

### Work

- Build alert evaluation for route deviation, late departure, major delay, disruption, cancellation, and admin-only driver-offline events.
- Build advisory creation and public-notice APIs.
- Implement the guardian-facing public live route view backend.
- Add alert deduplication and clear logic.
- Finish driver app replay polish, restart recovery, and notices UX.

### Exit Criteria

- Student and admin alert feeds are functional.
- Public live-view endpoints expose only route-safe bus data.
- The driver app stays narrow and operationally reliable.

## Week 17: Technical Ops, Audit, and Import or Export Workflows

### Goals

- Finish the operational tooling that keeps a self-hosted deployment maintainable.

### Work

- Implement technical-admin system-ops endpoints.
- Implement import job validation and async execution flow.
- Implement export job creation and retrieval flow.
- Finalize audit browsing endpoints and investigation-note APIs.
- Add system health widgets for outbox, queues, and failed jobs.

### Exit Criteria

- Technical admins can inspect queues, DLQs, and import or export jobs.
- Audit browsing is usable for finance and ops investigation.
- CSV-based operational setup is feasible without direct DB edits.

## Week 18: Admin Web App Core Shell

### Goals

- Convert the web scaffold into a usable operator surface before final integration.

### Work

- Implement admin authentication and role-aware navigation.
- Build layout, dashboard shell, and shared data-access patterns.
- Add finance, fleet, route, schedule, and alert module shells.
- Wire key read APIs into the dashboard.

### Exit Criteria

- Admin, cashier, and technical-admin users can log in and land on a relevant dashboard.
- The admin app reflects real module boundaries, not just a static page.
- Core read flows work without manual API inspection.

## Week 19: Admin Web App Operations Workflows and End-to-End Integration

### Goals

- Finish the operator control plane and connect all surfaces together.

### Work

- Implement finance adjustment flows and approval UI.
- Implement bus and QR management UI.
- Implement route, stop, fare-matrix, and schedule editing UI.
- Implement alert, advisory, audit, and DLQ views.
- Run end-to-end flows across admin, backend, student, and driver clients.

### Exit Criteria

- Admin and cashier workflows can be exercised end to end from the web app.
- Core transit setup can be done without direct database access.
- The platform can be demonstrated across all four surfaces together.

## Week 20: Hardening, Load Testing, and Showcase Readiness

### Goals

- Prove the system works under pressure and package it as a strong showcase.

### Work

- Run concurrency tests for 100-boarding burst scenarios and ledger invariants.
- Run telemetry load tests with multiple buses and many subscribers.
- Test outbox recovery, worker crashes, stale ETA fallback, poison messages, and emergency-ride edge cases.
- Clean up UX in student, driver, and admin surfaces.
- Add architecture diagrams, seed scripts, demo walkthrough docs, and CI usage notes for mobile artifact testing.

### Exit Criteria

- Core failure scenarios are tested and reproducible.
- Demo environment can be started and explained quickly.
- Project has clear documentation for architecture, setup, and test evidence.

## Milestones

- End of Week 6: financial concurrency proof is working.
- End of Week 11: student boarding works on a real phone build.
- End of Week 14: driver telemetry, archival, and replay are working.
- End of Week 16: live tracking, ETA, alerts, and public live view are working.
- End of Week 19: admin web app can operate the system end to end.
- End of Week 20: complete showcase demo is ready.

## Acceptance Checklist for the Full 20 Weeks

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
- Mobile apps are validated on real devices during the build, not only at the end.
