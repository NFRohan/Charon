# Charon Comprehensive Specification

## Document Status

- Status: Draft v1
- Date: 2026-03-08
- Purpose: Single source of truth for product behavior, architecture, interfaces, reliability rules, and delivery expectations.

## 1. Product Overview

Charon is a closed-loop university transit platform that combines:

- a digital wallet for ride boarding
- a high-concurrency fare-collection backend
- a live fleet-tracking system
- schedule-aware ETA and alerting
- lightweight operations tooling for transit staff
- a guardian-facing live route view for parents and other approved public viewers

The project is designed to demonstrate strong backend engineering, distributed systems thinking, and practical product judgment for a single-university deployment.

## 2. Product Goals

- Prevent double-charging during flaky mobile network conditions.
- Guarantee that wallet balances never go negative because of concurrent requests.
- Preserve a bounded boarding fallback when a student temporarily loses internet.
- Show live bus positions on a low-cost mobile map stack.
- Provide a privacy-safe live route view so guardians can see where the bus is.
- Keep current-position fanout off PostgreSQL.
- Archive telemetry for historical analysis, ETA, and alerts.
- Provide clear admin tooling for finance and operations.
- Showcase reliability patterns including idempotency, transactional outbox, and dead-letter queues.

## 3. Non-Goals

- Multi-university multi-tenant support.
- Public consumer payment gateway in the first release.
- Full route optimization or dispatch automation.
- Enterprise-grade multi-region deployment.
- GPS-derived or fully dynamic distance pricing in the first release.

## 3.1 Architecture Choice

Charon is intentionally designed as a modular monolith with worker processes, not a microservice platform.

This is the correct fit for the intended deployment model:

- self-hosted by individual institutes
- small fixed-route fleets
- low infrastructure budgets
- limited on-site engineering capacity
- strong need for operational clarity and low maintenance burden

The architecture prioritizes:

- financial correctness
- live user experience
- low operational complexity
- low recurring cost
- easy debugging by campus engineers

The architecture does not optimize for SaaS-style multi-tenant growth, because that is not a project goal.

## 3.2 Internal Extraction Seams

Although Charon is a modular monolith, the codebase must maintain clean seams so parts can be extracted later if there is ever a real need.

Required seams:

- `auth` module owns login, tokens, and role checks.
- `wallet` module owns accounts, transactions, ledger entries, overdraft, exemptions, and finance adjustments.
- `boarding` module owns QR validation, fare resolution, boarding rules, and boarding events.
- `routes` module owns routes, stops, timetables, service calendars, and service advisories.
- `telemetry` module owns live ingest, replay handling, current position state, and historical archival contracts.
- `eta` module owns stop ETA calculation and route-progress projection.
- `alerts` module owns alert rules, alert state, and notification triggers.
- `public_live_view` module owns guardian-safe projection of route state.
- `outbox` module owns DB-originated event publication.

Seam rules:

- Each module owns its write paths and invariants.
- Modules must not mutate another module's tables directly outside explicit service-layer interfaces.
- Cross-module communication inside the monolith should use service interfaces and typed domain events, not ad hoc shared logic.
- Event payloads must be versioned so async consumers can evolve safely.
- Worker consumers must depend on contracts owned by the source module, not duplicate source-of-truth rules.
- The public live view must consume a route-safe projection and must never reach into student or finance data directly.
- API and worker processes are separate deployable processes even if they share the same repository and modules.

Extraction triggers in the future would only be considered if one of these becomes true:

- a workload clearly needs independent scaling
- security or compliance requires hard runtime isolation
- separate teams need independent release cycles
- the deployment model changes beyond single-institute self-hosting

## 4. Users and Roles

### Student

- logs in to the mobile app
- views wallet balance and ledger history
- scans bus QR code to pay fare
- selects a stop during boarding for ETA and fare resolution
- views live bus location and ETA
- receives rider alerts

### Driver

- logs in to the mobile app
- uses employee ID plus password
- starts and ends a route session
- can self-attach to the current eligible service instance by scanning bus QR or entering bus code after login
- streams telemetry every 10 seconds

### Cashier

- credits student wallets
- processes refunds or corrections
- views transaction history for finance support

### Admin

- manages routes, stops, schedules, and route sessions
- monitors live fleet activity
- reviews and resolves alerts
- inspects dead-letter queues and requeues failed jobs

### Public Viewer

- is not an authenticated role
- may access a deployer-enabled read-only live route view
- can see route-level bus location, route progress, route-level ETA, and service advisories
- cannot see wallet, student-specific ETA preferences, boarding history, or finance data
- must never be presented as exact child tracking

## 5. Product Surfaces

### Student App (Flutter)

- Home-first mobile experience
- Wallet view with scan action
- QR scanner for boarding
- Live map with MapTiler-backed cached tiles
- Rider-specific stop ETA view with favorites
- Rider alerts and notification center
- Profile and settings

### Driver App (Flutter)

- Route start and stop workflow
- Bus attach by QR or numeric code
- Background telemetry stream
- Android-first foreground-service behavior for telemetry
- Driver status surface for current route session
- Basic operational notices

### Admin App (Next.js + TypeScript)

- shared admin and cashier web app with role-based screens
- finance operations
- bus registry and durable QR generation
- route and stop management
- timetable management
- live fleet view
- alert operations
- DLQ inspection and requeue tooling

### Public Live View

- deployer-enabled guardian-facing route map
- current bus position and route progress
- route-level ETA for the bus, not student-specific ETA
- service disruption and cancellation notices
- privacy-safe public presentation with no student-specific personalization

## 6. Scope by Release

### Phase 1

- auth and RBAC
- wallet ledger and balance snapshots
- minimal route and stop registry for boarding and fare resolution
- bus registry and durable QR issuance
- idempotent QR boarding
- sponsored boarding
- emergency ride permit fallback
- transactional outbox
- live telemetry fanout
- telemetry archival
- MapTiler map integration and tile caching
- cashier-issued wallet credits and refunds with admin approval above configured limits

### Phase 1.1

- full route, stop, and timetable management
- weekly schedules with holiday and special-event exceptions
- route sessions linked to schedules
- PostGIS enablement
- ETA engine
- route deviation detection

### Phase 1.2

- admin alerts
- rider alerts
- limited push notifications for service cancellation and major disruption only
- DLQ operations surface
- guardian-facing public live route view

## 7. Core Functional Requirements

### 7.1 Authentication and Authorization

- The system must support role-based login for student, driver, cashier, and admin users.
- The initial credential model must use role-specific credentials.
- Student users log in with student ID plus password.
- Driver users log in with employee ID plus password.
- The backend must issue JWT access tokens and refresh tokens.
- Protected endpoints must enforce role checks.
- Student sessions should remain logged in across app restarts by default.
- Student accounts may have multiple active device sessions in v1.
- Forgotten student-password reset is admin-assisted in v1.
- Driver-only endpoints must validate that the caller is assigned to the active route session when required.
- Admin, cashier, and technical operations users share one web app with role-based module access.

### 7.2 Wallet, Ledger, and Fare Policy

- All money values must be stored in integer minor units.
- Every financial change must create an immutable `transaction`.
- Every transaction must create balanced debit and credit `ledger_entries`.
- `wallet_accounts.available_balance_minor` must be updated inside the same database transaction for fast balance reads.
- Wallet history must be queryable by user.
- The fare model must support route-level flat fares, selected-stop-based fares, and zero-fare deployments or routes.
- Each boarding attempt must include a selected stop so ETA and fare logic receive consistent rider input.
- Each route session must resolve its fare from the assigned route configuration and the selected stop when the fare policy requires it.
- The boarding domain must support three payment layers: direct self-pay, sponsored boarding, and emergency ride permit fallback.
- Students may have a deployer-configurable small overdraft limit.
- Students may also be marked as fare-exempt.
- Credits and refunds must be cashier-driven at a campus counter, with room for future third-party integration.
- Cashiers may issue credits directly in v1.
- Refunds above the cashier limit must require admin approval.
- Every manual balance adjustment must record full before and after values, reason code, actor, and approval chain if any.

### 7.3 Boarding Flow

- Admin users must register buses and generate durable signed QR assets for each physical bus.
- The durable QR payload must contain `institute_id`, `bus_id`, `qr_version`, and signature.
- Scheduled morning and evening runs must be represented as separate boardable service instances.
- Each service instance covers the full outbound and return cycle for that run.
- Scheduled boarding becomes valid `30 minutes` before service start and remains valid until `15 minutes` after scheduled end.
- Driver start and end actions are operational signals, not the sole source of boarding authorization.
- The student app must send an `Idempotency-Key` with each boarding attempt.
- The backend must use Redis to store idempotency state scoped by `actor_id + Idempotency-Key`.
- The backend must resolve the current boardable service instance from the scanned `bus_id`.
- If the scanned bus has no boardable service instance, boarding must be rejected with the user-facing state `Trip Not Active Yet`.
- If the scanned bus has multiple boardable service instances because of bad operations data, boarding must be rejected and flagged for admin cleanup.
- The student app must include the selected stop in the boarding request.
- Direct self-pay remains the default boarding mode.
- Sponsored boarding must allow one authenticated student to pay for self plus at most one additional student in v1.
- Sponsored boarding must create separate rider-level boarding events while charging the payer in one atomic transaction.
- Sponsored boarding must be all-or-nothing in v1; if any rider in the request is invalid, duplicated, blocked, or financially disallowed, the whole request fails.
- Sponsored boarding in v1 uses the same selected stop for all riders in the request.
- Emergency ride permits must provide a bounded offline fallback for a student who has no internet and no sponsor available.
- Emergency ride permits must be pre-issued while the student device is online, stored securely on device, signed or strongly verifiable server-side, and bound to `student_id + device_id`.
- Emergency ride permits must be one-time use, capped to a single ride or configured max single fare, and short-lived.
- Emergency ride permit use must be recorded locally and redeemed with the backend when connectivity returns.
- The system must limit how many unresolved emergency permit uses can exist for one student or device at the same time.
- The backend must lock the student wallet row with `SELECT ... FOR UPDATE` before validating funds.
- A boarding charge is allowed when the post-charge balance stays within the configured overdraft limit.
- Fare-exempt students must be allowed to board without a student wallet debit, while still generating an auditable boarding record and policy decision.
- The system must prevent accidental duplicate charging on the same service instance using a uniqueness rule on `student_id + route_session_id`.
- The system assumes one boarding charge per student per service instance.
- The driver flow must not depend on keeping a phone screen visible for boarding.
- Telemetry freshness must not be a hard requirement for boarding authorization.
- Manual fallback using a short numeric bus code must always be available.
- Authenticated drivers may use the same bus QR or numeric bus code to bind themselves to the current eligible service instance, but the QR does not itself grant driver privilege.
- If multiple eligible service instances are available for the same bus around a boundary time, the authenticated driver may choose manually during attachment.
- Student device location must be evaluated on-device against the campus geofence only in v1.
- Raw student GPS coordinates must not leave the device for boarding validation.
- If location indicates likely remote scan or permission is denied, the app must show warning plus extra confirmation but still allow override.
- Scans outside the valid service window must be blocked and audit-logged.
- Repeated scans across multiple buses in a suspiciously short period must be audit-flagged.

### 7.4 Telemetry and Live Map

- The driver app must send telemetry every 10 seconds over WebSocket.
- Telemetry payload must include `route_session_id`, `bus_id`, `lat`, `lng`, `speed_kph`, `heading`, `accuracy_m`, and `recorded_at`.
- The API must publish live telemetry through Redis Pub/Sub.
- The API must maintain last-known bus positions in Redis.
- Student clients must subscribe to route updates over WebSocket.
- Live telemetry must not require synchronous PostgreSQL writes.
- The driver app must use Android foreground-service behavior while telemetry is active.
- If connectivity drops, the driver app must buffer at least 30 minutes of telemetry locally and replay it in order once the connection returns.
- The driver app must be able to start from locally cached service data when offline.
- Replayed telemetry must preserve original timestamps.
- Replayed stale telemetry must still be archived, but it must not be broadcast as fresh live movement if it is too old for real-time display.

### 7.5 Map Rendering and Tile Caching

- The mobile apps must not use Google Maps as the base map provider.
- The student app must use MapTiler-backed tiles and styles.
- Map tiles must be cached on-device.
- The cache must be bounded with LRU-style eviction.
- The app must prewarm the cache for campus bounds and all configured route corridors.
- The map experience must remain usable during brief network loss by serving cached tiles.
- Telemetry overlays must remain provider-agnostic so the map provider can be changed later without breaking the live location pipeline.

### 7.6 Telemetry Archival

- Raw telemetry must be published to RabbitMQ for durable async handling.
- The archiver worker must batch writes in memory.
- The archiver worker must bulk insert every 60 seconds or 1000 points, whichever comes first.
- Historical telemetry must be queryable for analytics, alerts, and future reporting.

### 7.7 Schedules and ETA

- Admin users must create and edit routes, stops, stop order, trip templates, and scheduled stop times.
- The scheduling model must support a weekly timetable plus holiday and special-event exceptions.
- Admin users must be able to mark service as unavailable for planned closures or disruptions.
- Scheduled trips must create separate morning and evening service instances for boarding and operations.
- A route session must be linked to a scheduled trip once scheduling is enabled.
- ETA for students must be stop-specific and tied to the rider's manually selected or favorited stop.
- ETA for the public live view must be route-level only and not personalized.
- ETA must be computed from scheduled stop times plus live delay.
- If telemetry is stale for 30 seconds, ETA must fall back to scheduled time and be flagged as stale.

### 7.8 Route Deviation

- The system must identify when a vehicle is more than 200 meters away from its route corridor for 3 consecutive telemetry points.
- Spatial lookups must use PostGIS rather than only application-layer math.

### 7.9 Alerts and Notifications

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

Delivery requirements:

- only students and admins receive alerts in v1
- alerts must appear in-app in real time
- push notifications are limited in v1 to service cancellation and major service disruption alerts
- alerts must support creation, deduplication, and clear semantics

### 7.10 Guardian Live View

- The system must provide a public-facing live route view built from the same telemetry pipeline as the student experience.
- The public view must show active buses, route progress, route-level ETA, and active service advisories.
- The public view must never expose wallet data, student identities, boarding events, or student-specific ETA preferences.
- The public view must be safe for guardians and parents to understand where a bus is, but it must not claim to show a child's exact current location.
- The public view must degrade gracefully when telemetry is stale by marking the bus as stale or recently updated rather than showing misleading fresh movement.
- The public view must be deployer-controlled so the university can disable it if needed.

## 8. Key Workflows

### 8.1 Student Boarding Workflow

Supported modes:

- direct self-pay
- sponsored boarding
- emergency ride permit fallback

#### 8.1.1 Direct Self-Pay

1. Student scans the durable bus QR or enters the numeric bus code manually.
2. App prompts the student to select the relevant stop, with favorites surfaced first where available.
3. App performs the campus-geofence safety check locally and determines whether warning or override UX is needed.
4. App shows confirmation with bus, route, service label, selected stop, fare, service window, and expected balance after charge.
5. Student confirms and the app sends `POST /boardings` with `Idempotency-Key`.
6. API checks Redis idempotency state.
7. API validates the QR or manual bus lookup and resolves the current boardable service instance from `bus_id`.
8. API resolves the applicable route fare from the fare policy, selected stop, and rider exemption or overdraft policy.
9. API opens a Postgres transaction.
10. API locks the wallet account row.
11. API validates duplicate-boarding rule and financial allowance.
12. API inserts the financial transaction and ledger entries when money is being charged.
13. API updates balance snapshots when money moves.
14. API inserts the boarding event, fare decision record, selected-stop metadata, and scan audit metadata.
15. API inserts the outbox event.
16. API commits and stores the final response in Redis.

#### 8.1.2 Sponsored Boarding

1. Connected payer scans the bus QR or enters the numeric bus code.
2. Payer selects the stop and may add one additional rider by entering student ID.
3. App resolves the additional rider with masked confirmation and shows rider count plus total fare.
4. Payer confirms the sponsored boarding request.
5. API validates service window, duplicate-boarding rule, rider eligibility, and payer financial allowance for all riders together.
6. API creates one financial transaction for the payer.
7. API creates one boarding event per rider, all linked to the same route session and transaction.
8. API commits atomically or rejects the whole request.

#### 8.1.3 Emergency Ride Permit Fallback

1. Student has a valid pre-issued emergency ride permit already stored securely on device.
2. Student scans the durable bus QR or enters the numeric bus code and selects the stop.
3. App validates that a local emergency permit is available and presents a clearly labeled emergency-ride confirmation.
4. App marks the permit as locally consumed and stores a pending redemption record.
5. Student is allowed to board without full live backend completion at that moment.
6. When connectivity returns, the app redeems the permit with the backend.
7. Backend validates the permit, service window, duplicate rule, and fare cap, then creates the normal boarding event and ledger debit.
8. If redemption later fails, the account is handled through debt, overdraft, and permit-issuance policy rather than pretending the ride never happened.

### 8.2 Telemetry Workflow

1. Driver authenticates with employee ID and attaches to a bus or current service instance using QR or numeric bus code.
2. Telemetry begins immediately and is emitted every 10 seconds.
3. If offline, driver app stores telemetry locally until reconnect.
4. API validates route session and payload shape.
5. Fresh telemetry is published through Redis Pub/Sub and updates last-known position in Redis.
6. Raw telemetry payloads, including replayed points, are published to RabbitMQ.
7. Archiver worker batches and bulk writes telemetry to PostgreSQL.

### 8.3 Outbox Workflow

1. Business transaction commits in PostgreSQL.
2. Matching outbox row commits in the same transaction.
3. Outbox publisher worker polls pending rows using `FOR UPDATE SKIP LOCKED`.
4. Worker publishes to RabbitMQ with publisher confirms.
5. Worker marks the outbox row as published.
6. Consumers use `event_id` for deduplication.

### 8.4 Alert Workflow

1. Telemetry, schedule, or route-session events enter worker processing.
2. Alert evaluator applies rules and thresholds.
3. New or updated alert is stored in PostgreSQL.
4. Alert state change produces an outbox event.
5. Notification dispatcher sends in-app and optional push delivery.

### 8.5 Guardian Live View Workflow

1. Public viewer opens the live route page.
2. Frontend requests public route list, active route sessions, latest bus positions, and advisories.
3. Frontend subscribes to public-safe live route updates.
4. API pushes route-level telemetry and ETA updates only.
5. If a bus becomes stale, the view marks the update age and service state rather than pretending the bus is still moving live.

## 9. System Architecture

## 9.1 Runtime Components

- `student_app` (Flutter)
- `driver_app` (Flutter)
- `admin_app` (Next.js + TypeScript)
- `public_live_view` (optional Next.js public route or page)
- `api` (Go)
- `worker` (Go)
- `postgres`
- `redis`
- `rabbitmq`
- `reverse-proxy`
- `MapTiler`
- `FCM`

### 9.2 Responsibility Split

`api` handles:

- auth
- synchronous wallet and boarding flows
- WebSocket connections
- live telemetry fanout
- admin CRUD APIs
- public route-status and live-view APIs

`worker` handles:

- outbox publishing
- telemetry archival
- alerts evaluation
- notifications dispatch

This split is the primary operational seam in the first version:

- `api` remains optimized for user-facing latency
- `worker` remains optimized for durable background processing

`postgres` handles:

- financial source of truth
- schedules and route data
- alert history
- historical telemetry

`redis` handles:

- idempotency state
- current bus positions
- live pub/sub channels

`rabbitmq` handles:

- durable async telemetry
- durable async alerting
- durable async notifications
- outbox-delivered domain events

## 10. Data Model

### 10.1 Core Tables

`users`

- id
- role
- name
- institutional_id
- status
- fare_exempt

`wallet_accounts`

- id
- user_id
- available_balance_minor
- overdraft_limit_minor
- status
- updated_at

`transactions`

- id
- type
- amount_minor
- status
- actor_id
- route_session_id nullable
- created_at

`ledger_entries`

- id
- transaction_id
- account_id
- direction
- amount_minor
- created_at

`boarding_events`

- id
- student_id
- route_session_id
- selected_stop_id nullable
- paid_by_student_id nullable
- transaction_id
- boarding_mode
- fare_minor
- fare_policy_type
- charge_mode
- exemption_reason_code nullable
- emergency_permit_id nullable
- created_at

`outbox_events`

- event_id
- aggregate_type
- aggregate_id
- event_type
- payload_json
- attempt_count
- available_at
- published_at nullable
- last_error nullable

`routes`

- id
- code
- name
- fare_policy_type
- default_fare_minor nullable
- status

`route_fare_rules`

- id
- route_id
- service_label nullable
- stop_id nullable
- fare_minor
- effective_from nullable
- effective_to nullable

`emergency_ride_permits`

- id
- student_id
- device_id
- permit_token_hash
- max_fare_minor
- status
- expires_at
- issued_at
- used_at nullable
- redeemed_at nullable

`stops`

- id
- name
- position

`route_stop_sequences`

- id
- route_id
- stop_id
- stop_order

`buses`

- id
- code_or_plate
- status
- qr_version
- default_route_id nullable

`trip_templates`

- id
- route_id
- service_calendar_id
- name
- status

`trip_stop_times`

- id
- trip_template_id
- stop_id
- scheduled_time

`route_sessions`

- id
- trip_template_id nullable in Phase 1
- bus_id
- session_source
- service_label
- scheduled_start
- scheduled_end
- driver_id
- started_at
- ended_at nullable
- status

`telemetry_points`

- id
- route_session_id
- bus_id
- position
- is_replayed
- speed_kph
- heading
- accuracy_m
- recorded_at
- received_at

`alerts`

- id
- type
- severity
- target_type
- target_id
- status
- opened_at
- closed_at nullable

`finance_adjustments`

- id
- wallet_account_id
- transaction_id nullable
- adjustment_type
- requested_by
- approved_by nullable
- approval_status
- reason_code
- before_balance_minor
- after_balance_minor
- created_at

`service_calendars`

- id
- route_id
- weekday_mask
- effective_from
- effective_to nullable

`service_exceptions`

- id
- service_calendar_id
- service_date
- exception_type
- reason_code

`service_advisories`

- id
- route_id nullable
- advisory_type
- message
- starts_at
- ends_at nullable

`audit_logs`

- id
- actor_id nullable
- subject_type
- subject_id nullable
- action_type
- result
- payload_json
- created_at

`audit_investigation_notes`

- id
- audit_log_id
- author_id
- note_body
- created_at

`device_tokens`

- id
- user_id
- platform
- token
- push_enabled

### 10.2 Invariants

- every financial transaction balances to zero
- every boarding resolves exactly one fare decision, whether charged, exempt, or zero-fare
- sponsored boarding may charge one payer for multiple riders, but each rider still gets a separate immutable boarding event
- wallet balance snapshot and ledger entries change in the same transaction
- outbox event exists for every DB-originated event that must leave the system
- no duplicate boarding charge for the same student and route session
- manual credits and refunds record actor, before and after values, reason code, and approval chain where applicable
- audit log rows are immutable after insert
- investigation context is appended through linked notes, never by mutating original audit events
- module boundaries remain enforceable and no cross-module direct table mutation bypasses service rules
- Redis is never the financial source of truth
- telemetry unavailability alone must never block otherwise valid boarding
- student-selected stop is recorded with the boarding event whether or not the fare policy uses it

## 11. Interface Contracts

### 11.1 REST Endpoints

Auth:

- `POST /auth/login`
- `POST /auth/refresh`
- `POST /auth/logout`

Student wallet:

- `GET /wallet/balance`
- `GET /wallet/transactions`
- `POST /wallet/emergency-voucher/issue`
- `GET /boardings/preview`
- `POST /boardings`

Admin finance:

- `POST /admin/wallets/{id}/credits`
- `POST /admin/wallets/{id}/refunds`

Routes and schedules:

- `GET /routes`
- `GET /routes/{id}/stops`
- `GET /routes/{id}/eta`
- `POST /admin/routes`
- `POST /admin/stops`
- `POST /admin/trips`

Route sessions:

- `POST /route-sessions/start`
- `POST /route-sessions/end`

Public live view:

- `GET /public/routes/active`
- `GET /public/routes/{route_code}/live`
- `GET /public/advisories`

### 11.2 WebSocket Message Types

Driver to API:

- `driver.telemetry`

Student to API:

- `student.subscribe_route`
- `student.unsubscribe_route`

API to clients:

- `telemetry.update`
- `eta.update`
- `alert.created`
- `alert.cleared`
- `public.route_update`
- `public.advisory_update`

### 11.3 Event Types

Examples of DB-originated domain events:

- `wallet.transaction.created`
- `wallet.transaction.completed`
- `wallet.credit.issued`
- `wallet.refund.issued`
- `route_session.started`
- `route_session.ended`
- `schedule.updated`
- `alert.opened`
- `alert.closed`

## 12. Reliability and Failure Handling

### 12.1 Idempotency

- Redis stores `PROCESSING`, `COMPLETED`, or `FAILED`.
- Idempotency keys have a 24-hour TTL.
- Duplicate requests must return the stored result or a bounded retry response.

### 12.2 Concurrency Control

- Boarding flow uses row-level locking on the student wallet account.
- Low balance plus burst traffic must never create negative balances.

### 12.3 Transactional Outbox

- No DB-originated event is published directly from the request handler.
- All such events must leave via the outbox worker.

### 12.4 RabbitMQ Retry and DLQ Rules

- `telemetry.archiver`, `alerts.evaluate`, and `notifications.dispatch` each get dedicated retry and DLQ handling.
- After 3 failed attempts, the message must move to its queue-specific DLQ.
- Admin users must be able to inspect and requeue failed messages.

### 12.5 Map Resilience

- Cached tiles must continue to serve the visible campus map during short network loss.
- Loss of live telemetry should not break the base map.
- Loss of base map refresh should not stop telemetry overlays from updating.

### 12.6 Public View Safety and Freshness

- Public live updates must be filtered to route-safe fields only.
- Public route state must include update age so stale telemetry is visible to the viewer.
- The public view must not infer, calculate, or display any student-specific movement.

### 12.7 Telemetry Replay

- Driver apps must replay buffered telemetry in order after reconnect.
- Replayed telemetry must preserve original `recorded_at`.
- Replayed telemetry older than the live-display freshness window should be archived without being rebroadcast as current movement.

### 12.8 Retention

- Historical telemetry must be retained for 30 days in the primary demo environment.
- Retention cleanup must not interfere with current operational queries.

### 12.9 Durable QR Safety

- Durable bus QR assets must be signed so students cannot forge bus identifiers.
- Boarding depends on an active route session lookup after scan, not just on QR presence.
- Admins must be able to rotate and reissue a bus QR by bumping `qr_version` if a sticker is damaged or leaked.
- Old QR versions may continue to work for a `1 day` grace period, but their use must be flagged.
- Schedule-backed service windows, not telemetry freshness, are the primary guard against forgotten route endings.
- Manual numeric bus-code fallback must go through the same authorization and audit path as QR scan.

## 13. Security Requirements

- JWT signing keys and third-party secrets must be stored in environment-backed secret configuration.
- Durable bus QR codes must be signed and validated server-side.
- Admin and cashier endpoints must enforce strict RBAC.
- Audit trails must exist for credits, refunds, and route-session changes.
- Audit trails must also exist for durable QR issuance and rotation.
- Audit events must be immutable; operator investigation notes must be stored separately from the source audit row.
- PII must be minimized in logs and message payloads.
- The guardian live view must expose only route-level operational data and no student-identifying information.

## 14. Performance Targets

- The design target is 100 concurrent boarding attempts against a single bus route session without double-charge or negative-balance failures.
- Boarding requests under normal load should complete in sub-second time.
- Live telemetry fanout should feel real time to the student map.
- Telemetry archival must not write per ping directly to PostgreSQL.
- Map tile caching must reduce repeated mobile tile fetches for common campus views.

## 15. Operational Visibility

The system should expose at least:

- request latency
- idempotency hit rate
- boarding success and failure counts
- overdraft usage rate
- outbox backlog
- RabbitMQ queue depth
- DLQ counts
- telemetry ingest rate
- telemetry replay volume
- active WebSocket connection count
- public live-view connection count
- stale route session count
- blocked out-of-window scan count
- manual bus-code fallback count
- suspicious multi-bus scan count

Operational simplicity is a first-class requirement:

- a campus engineer should be able to diagnose queue backlog, stale telemetry, failed credits, and live-map issues without tracing through a large distributed system
- deployment and recovery should remain understandable with a small set of processes and clear logs

## 16. Testing and Validation

### 16.1 Financial Tests

- concurrent boarding load tests
- same-key retry tests
- different-key low-balance contention tests
- overdraft threshold tests
- exempt-rider boarding tests
- sponsored-boarding atomicity tests
- emergency-permit issue, consume, and redeem tests
- ledger invariant tests
- admin credit and refund tests

### 16.2 Reliability Tests

- API crash after DB commit but before event publication
- outbox replay and recovery
- worker restart during backlog processing
- poison-message routing to DLQ

### 16.3 Telemetry and Map Tests

- multiple active buses and subscribers
- offline driver buffering and ordered replay
- brief network loss with cached tile fallback
- telemetry archival batching
- stale ETA fallback
- route deviation detection
- public live view freshness and privacy filtering tests
- boarding outside service window tests
- manual bus-code fallback tests
- on-device location warning and override tests

### 16.4 Alert Tests

- speeding alert
- driver offline alert
- route deviation alert
- late departure alert
- bus approaching stop alert
- major delay alert
- alert clear behavior

## 17. Delivery Plan

- Use [ARCHITECTURE_PLAN.md](e:\Projects\Charon\ARCHITECTURE_PLAN.md) as the shorter architecture companion.
- Use [ADMIN_SPEC.md](e:\Projects\Charon\ADMIN_SPEC.md) as the detailed admin and cashier operations reference.
- Use [BUS_QR_SPEC.md](e:\Projects\Charon\BUS_QR_SPEC.md) as the detailed durable QR and boarding-behavior reference.
- Use [DRIVER_APP_SPEC.md](e:\Projects\Charon\DRIVER_APP_SPEC.md) as the detailed driver behavior reference.
- Use [STUDENT_APP_SPEC.md](e:\Projects\Charon\STUDENT_APP_SPEC.md) as the detailed student behavior reference.
- Use [API_SPEC.md](e:\Projects\Charon\API_SPEC.md) as the first wire-level contract for auth, boarding, wallet, sockets, and public live view.
- Use [ADMIN_CASHIER_API_SPEC.md](e:\Projects\Charon\ADMIN_CASHIER_API_SPEC.md) as the wire-level contract for the shared admin and cashier web application.
- Use [STUDENT_SELF_SERVICE_API_SPEC.md](e:\Projects\Charon\STUDENT_SELF_SERVICE_API_SPEC.md) as the wire-level contract for student profile, settings, favorites, and alert read-state.
- Use [DRIVER_SERVICE_API_SPEC.md](e:\Projects\Charon\DRIVER_SERVICE_API_SPEC.md) as the wire-level contract for driver attachment, service control, notices, and device-health reporting.
- Use [SYSTEM_OPS_API_SPEC.md](e:\Projects\Charon\SYSTEM_OPS_API_SPEC.md) as the wire-level contract for technical-admin queue, DLQ, and worker-health endpoints.
- Use [NONCRITICAL_API_SPEC.md](e:\Projects\Charon\NONCRITICAL_API_SPEC.md) as the backlog contract for deferred student, driver, admin, and system-ops endpoints.
- Use [SPRINT_20_WEEKS.md](e:\Projects\Charon\SPRINT_20_WEEKS.md) as the delivery timeline.
- Use [ENGINEERING_STORY.md](e:\Projects\Charon\ENGINEERING_STORY.md) as the running decision and reasoning log.
- Keep this document at system level; wire-level details now live in [API_SPEC.md](e:\Projects\Charon\API_SPEC.md).

## 18. Acceptance Criteria

The project is successful when all of the following are true:

- a student can board exactly once under retry-heavy network conditions
- a student can use the app on Android and iOS
- a student can scan a durable bus QR without requiring the driver to present a phone
- a student can also board using the manual numeric bus-code fallback when scan fails
- a student who loses internet can still board through sponsored boarding or a valid emergency ride permit without breaking financial invariants
- wallet balances remain correct under concurrent request bursts
- route-based flat fare, selected-stop fare, and zero-fare policies can be configured per deployment
- small overdraft and exempt-student rules behave as configured and remain auditable
- the live map shows buses without PostgreSQL being in the synchronous path
- the map stack uses cached MapTiler tiles instead of an expensive Google Maps dependency
- driver telemetry can buffer offline and replay without corrupting current live position
- DB-originated events survive API crashes because of the outbox
- malformed worker messages do not block the system because of DLQs
- ETA and route deviation logic work using PostGIS-backed spatial queries
- holiday closures and special-event schedule changes can be represented
- admin users can operate finance, schedules, alerts, and failed async jobs from one coherent surface
- guardians can open a public-safe live route view to see where the bus is and whether service is disrupted
- the system is polished enough to serve as a portfolio-ready showcase

## 19. Defaults and Assumptions

- single university deployment
- open-source project intended for institute self-hosting, not SaaS growth
- student ID plus password for students, employee ID plus password for drivers
- Android and iOS student app in v1, with Android-first driver app
- student sessions persist by default and may exist on multiple devices
- forgotten student-password reset is admin-assisted
- durable admin-issued QR per physical bus, with active-session lookup at boarding time
- schedule-authoritative boarding window with `30 minute` early boarding and `15 minute` late grace
- route-based flat fare, selected-stop fare, or zero-fare policy in the first release, with no GPS-derived distance pricing yet
- three-layer boarding fallback of direct self-pay, sponsored boarding, and bounded emergency ride permit
- cashier counter funding with future third-party credit-system integration left open
- small deployer-configurable overdraft support
- optional fare exemptions for eligible students
- MapTiler-backed map stack with client-side tile caching
- cached campus and all-route map coverage
- Android-first driver app in v1 with background telemetry via foreground service
- PostGIS introduced when schedule-aware spatial features begin
- weekly timetable plus exception dates and service advisories
- in-app alerts by default, with push reserved for service cancellation and major disruption
- guardian live view is route-level only and never student-tracking
- 30-day telemetry retention in the demo environment
- RabbitMQ for durable async workflows
- Redis for ephemeral hot-path state
- PostgreSQL as the source of truth
