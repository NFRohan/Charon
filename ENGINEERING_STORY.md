# Charon Engineering Story

## Purpose

This is the living engineering story for Charon. It records what has been done, which decisions were taken, why those decisions were made, and how each decision changes the build path.

The point is not just to preserve outcomes. It is to preserve reasoning so future changes stay coherent instead of turning into disconnected one-off choices.

## Current Status

Planning is complete enough to start implementation. The project currently has:

- a system architecture plan
- a 10-week sprint plan
- a comprehensive specification
- a locked first major platform decision for map rendering: MapTiler with client-side caching in Flutter

## Build Principles

- Prefer correctness before convenience in all money flows.
- Prefer bounded operational complexity over premature service splitting.
- Prefer asynchronous durability for non-interactive workflows.
- Prefer cost control when a feature can silently become expensive at scale.
- Prefer decisions that strengthen the portfolio story, not just the codebase.

## Work Done So Far

### 1. Architecture baselined

Work completed:

- Defined the system as a Go modular monolith with workers.
- Chose PostgreSQL, Redis, and RabbitMQ as the core backend stack.
- Split the product into student, driver, and admin surfaces.

Why this was done:

- The project needs to prove system-design maturity without hiding behind infrastructure sprawl.
- A modular monolith is enough to show domain boundaries, concurrency control, and async workflows.
- Separate workers allow durability and throughput without committing to early microservices.

### 2. Financial correctness model locked

Work completed:

- Chose a double-entry ledger.
- Kept a wallet balance snapshot for fast reads.
- Added Redis idempotency for flaky-network retries.
- Added row-level locking for overdraft prevention.

Why this was done:

- The core claim of the project is that it can survive boarding surges and poor mobile networks.
- A balance-only model is simpler but weak as a showcase and weak for auditability.
- The snapshot plus immutable ledger approach keeps both correctness and performance.

### 3. Reliability model upgraded

Work completed:

- Added a transactional outbox to close the Postgres-to-RabbitMQ dual-write gap.
- Added explicit per-queue DLQ strategy.
- Added recovery and poison-message testing to the roadmap.

Why this was done:

- A financially committed event must not disappear just because the API crashes after commit.
- Async workers need bounded failure handling or a poison message can stall the platform.
- These patterns raise the architectural bar and make the system interview-ready.

### 4. Spatial model upgraded

Work completed:

- Added PostGIS to the roadmap for ETA and route-deviation logic.
- Planned spatial indexing instead of only raw latitude and longitude math in application code.

Why this was done:

- Route deviation and proximity queries are spatial problems, not just generic CRUD problems.
- PostGIS removes a lot of fragile custom math and makes the design more credible.

### 5. Comprehensive spec consolidated

Work completed:

- Created a single comprehensive specification that combines product scope, workflows, interfaces, data expectations, and reliability rules.
- Kept the shorter architecture, sprint, and engineering-story documents as companion views rather than competing sources of truth.

Why this was done:

- The project now has enough architectural detail that fragmented notes would start drifting.
- A comprehensive spec makes implementation, review, and future changes much easier to keep aligned.

### 6. Product rules clarified through spec review

Work completed:

- Locked student ID plus password for students and employee ID plus password for drivers as the launch login model.
- Replaced the simple flat-fare assumption with route-level flat fare or zero-fare policy support.
- Added small overdraft support and optional fare exemptions.
- Simplified the driver flow so route assignment is admin-managed and the driver mainly presses start.
- Added offline telemetry buffering and replay.
- Added weekly schedules with holiday and special-event exceptions.
- Narrowed alerts to the routes and audiences that actually matter.
- Limited push notifications to service cancellation and major disruption cases.
- Added 30-day telemetry retention and a 100-concurrent-boarding showcase target.

Why this was done:

- These choices move the spec away from generic transit software and toward the real university operating model.
- The system now reflects the constraints of the intended user base more honestly, especially around driver simplicity, finance control, and predictable map cost.
- It also sharpens the portfolio story by making the performance target and product rules concrete.

### 7. Guardian live view promoted to a first-class feature

Work completed:

- Promoted the public live route view from an optional note to a proper product surface.
- Added product, workflow, interface, and acceptance-criteria coverage for the guardian-facing live view.
- Added explicit privacy boundaries so the live view exposes route and bus status, not student tracking.

Why this was done:

- A guardian-facing live route page is a strong product addition for a university bus system.
- It reuses the existing telemetry pipeline efficiently, so the feature adds visible value without changing the core system shape.
- The privacy boundary needed to be explicit, because "parents can see where the bus is" is acceptable while "parents can track a child" is a very different and riskier product claim.

### 8. Monolith-first stance explicitly defended

Work completed:

- Confirmed that the project should remain a modular monolith with separate worker processes.
- Added explicit extraction seams to the specification.
- Recorded that the project is open source and intended for institute self-hosting, not SaaS expansion.

Why this was done:

- The likely operators are campus engineers with limited time and infrastructure appetite.
- The real problem is correctness, reliability, and usability, not extreme scale.
- A more distributed system would make deployment, debugging, and support worse for the target audience.

### 9. Durable bus QR replaces driver phone display

Work completed:

- Replaced the driver-presented session QR model with durable admin-issued QR assets tied to physical buses.
- Moved QR generation responsibility to the admin side.
- Updated the boarding flow so scan validation resolves the active route session from the bus instead of the driver's phone screen.

Why this was done:

- Expecting a driver to keep a phone visible for the entire boarding window is poor physical UX and unrealistic operationally.
- A printed or mounted QR is simpler, more durable, and less dependent on driver behavior, battery, or device quality.
- The change keeps the backend guarantees intact while removing friction from the physical boarding process.

### 10. Bus QR behavior detailed and decoupled from telemetry

Work completed:

- Defined a dedicated bus QR specification.
- Shifted boarding authorization to schedule-backed service windows instead of telemetry freshness.
- Added on-device campus-geofence safety checks with warning and override behavior.
- Added manual numeric bus-code fallback.
- Added QR rotation grace-period and scan-audit requirements.

Why this was done:

- The boarding path must survive weak driver devices, Android background-kill behavior, and imperfect human operations.
- The goal is to prevent accidental or casual abuse without turning the product into a privacy-heavy or admin-heavy system.
- The bus QR flow needed a tighter operational model than the earlier high-level spec provided.

### 11. Admin operating model defined

Work completed:

- Defined a single shared web app for admins, cashiers, and technical admins.
- Locked cashier scope to finance lookup, student search, credits, and refunds.
- Added refund approval rules, technical-admin system-ops boundaries, and CSV import or export requirements.
- Defined dynamic driver attachment as authenticated bus selection through bus QR or bus code, rather than rigid pre-assignment.

Why this was done:

- Campus operators need one understandable control surface, not multiple fragmented portals.
- Cashier scope needed to stay narrow so finance workflows remain simple and safer.
- Driver attachment needed to stay flexible enough for changing real-world assignments without making the public QR itself privileged.

### 12. Audit log immutability tightened

Work completed:

- Corrected the admin model so audit events remain immutable.
- Split investigation commentary from source audit facts by introducing linked investigation notes instead of editable audit rows.

Why this was done:

- Financial and operational audit trails should not be editable in place.
- Admins still need space for investigation context, but that must be append-only around the audit event rather than a mutation of it.

### 13. Driver app operating model defined

Work completed:

- Defined the driver app as an Android-first, minimal operational tool for personal phones.
- Switched driver authentication to employee ID plus password.
- Reduced telemetry frequency to 10 seconds for better battery behavior.
- Added foreground-service, battery-optimization warning, offline attach or start, and app-restart recovery requirements.
- Kept route maps and complex route detail out of the driver experience.

Why this was done:

- The driver workflow must stay simple enough for low-friction daily use.
- Personal Android phones and aggressive background-kill behavior are the real environmental constraints.
- The app should optimize for reliable telemetry and low cognitive load, not feature richness.

## Decision Log

### Decision 001: Go modular monolith plus workers

Decision:

- Use one main Go codebase for API and domain logic.
- Use separate worker processes for outbox publishing, telemetry archival, alerts, and notifications.

Reasoning:

- This keeps deployment simple while still showing clean separation of synchronous and asynchronous concerns.
- It avoids service explosion before the system actually needs it.

Tradeoff accepted:

- Some modules will share one deployment artifact early on.
- Strong internal boundaries will matter more than process boundaries at first.

### Decision 002: PostgreSQL is the financial and historical source of truth

Decision:

- Keep money, schedules, alerts, and historical telemetry in PostgreSQL.

Reasoning:

- The ledger needs ACID guarantees and a relational model.
- Historical telemetry and schedules are operationally important and need durable querying.

Tradeoff accepted:

- PostgreSQL must be protected from write amplification.
- Live telemetry therefore cannot go straight to disk on every ping.

### Decision 003: Redis handles idempotency, fanout, and hot position state

Decision:

- Use Redis for request idempotency, current bus position cache, and live pub/sub fanout.

Reasoning:

- These are high-frequency, low-latency, short-lived concerns.
- Redis keeps the hot path off PostgreSQL.

Tradeoff accepted:

- Redis state is disposable and must never become the financial source of truth.

### Decision 004: RabbitMQ carries durable async workflows

Decision:

- Use RabbitMQ for telemetry archival, alerts, notifications, and outbox-delivered domain events.

Reasoning:

- These flows need persistence, retry semantics, and controlled back-pressure.
- Pub/sub alone is not enough for reliable worker processing.

Tradeoff accepted:

- Message topology is more complex than direct in-process callbacks.
- That complexity is justified because reliability is part of the product story.

### Decision 005: MapTiler with Flutter caching replaces Google Maps

Decision:

- Use MapTiler as the base map provider in Flutter.
- Add local tile caching and cache prewarm for campus and route views.
- Avoid Google Maps as the default map layer.

Reasoning:

- Google Maps can become an uncontrolled recurring cost very quickly for a map-heavy student app.
- Charon needs predictable operating cost if it is to look realistic for a university deployment.
- Cached MapTiler tiles also improve usability on weak campus mobile data and reduce repeated map fetches.
- This decision keeps map spending aligned with the rest of the project philosophy: pay for correctness where it matters, not for preventable per-view costs.

Tradeoff accepted:

- Some Flutter map integration work becomes more manual than a plug-and-play Google Maps setup.
- Tile caching, key handling, and style management must be owned explicitly.

Impact on the build:

- Week 1 now includes locking the map provider and cache design.
- Week 5 now includes MapTiler integration and viewport prewarming.
- The mobile map layer should stay provider-agnostic above the tile source so future swaps remain possible.

### Decision 006: Outbox for all DB-originated events

Decision:

- Any event whose truth begins in PostgreSQL must leave the system through the outbox.

Reasoning:

- This closes the dual-write gap completely instead of only for the wallet domain.
- It gives one consistent reliability rule for the entire platform.

Tradeoff accepted:

- Event delivery becomes slightly more delayed than direct publish in the request handler.
- That latency cost is acceptable because consistency matters more than immediate best-effort fanout for those events.

### Decision 007: Product rules follow real campus operations, not generic transit defaults

Decision:

- Use student ID plus password for students and employee ID plus password for drivers at launch.
- Support route-level flat fare and zero-fare policies.
- Allow small overdraft and optional fare exemptions.
- Keep the driver start flow extremely simple through admin-managed route setup.
- Support holiday closures, special-event schedule overrides, and public-facing service disruption notices.

Reasoning:

- The target users and deployers are universities, not generic city transit operators.
- The product must reflect how campuses actually run buses, fares, and staffing constraints.

Tradeoff accepted:

- The spec becomes more opinionated and less plug-and-play for every possible transit model.
- That is acceptable because realism and clarity matter more than pretending the first version solves every deployment shape.

### Decision 008: Guardian live view is public-safe bus visibility, not student tracking

Decision:

- Add a guardian-facing live route page as part of the project.
- Expose only route-level bus position, route progress, ETA, and service advisories.
- Do not expose student identity, boarding state, or child-specific location claims.

Reasoning:

- This is a meaningful product feature for parents and families.
- It leverages the existing live telemetry path with relatively low extra system complexity.
- The feature is only acceptable if its privacy boundary is explicit and enforced.

Tradeoff accepted:

- The public feature must be narrower than the student experience.
- That limitation is intentional because privacy and clarity matter more than feature symmetry.

### Decision 009: Stay modular-monolith-first, but preserve clean extraction seams

Decision:

- Keep Charon as a modular monolith with separate API and worker processes.
- Do not pivot toward microservices for the planned deployment model.
- Preserve clear internal module ownership and typed event seams so future extraction remains possible if reality ever demands it.

Reasoning:

- The project is open source and meant for self-hosting by single institutes with small fleets and limited engineering bandwidth.
- The likely failure modes are operational confusion, broken money flows, and weak user experience, not horizontal scale exhaustion.
- A simple architecture is more likely to produce a good user experience and a maintainable deployment than a prematurely distributed one.

Tradeoff accepted:

- Independent runtime scaling is limited compared with a more distributed architecture.
- That is acceptable because the actual workload is bounded and the operational simplicity is more valuable than theoretical scale headroom.

### Decision 010: Durable admin-issued bus QR is better than driver-held session QR

Decision:

- Generate durable signed QR assets from the admin side and attach them to physical buses.
- Resolve the active route session at scan time using the scanned bus identifier.
- Remove any requirement for drivers to hold out a phone during boarding.

Reasoning:

- This is much better physical UX for the real boarding environment.
- It reduces reliance on driver devices during the busiest and most chaotic part of the flow.
- It still preserves strong server-side control because the bus must be in an active route session and the QR remains signed.

Tradeoff accepted:

- The QR is less dynamic than a route-session-specific on-screen code.
- That is acceptable because active-session validation and QR rotation controls provide enough safety for the bounded campus use case.

### Decision 011: Boarding authorization is schedule-authoritative, not telemetry-authoritative

Decision:

- Use schedule-backed service windows as the source of boarding truth.
- Open boarding `30 minutes` before service start and keep it valid until `15 minutes` after scheduled end.
- Do not require fresh telemetry for payment authorization.

Reasoning:

- Target Android devices often kill background apps aggressively.
- Students can arrive before the driver is operationally ready.
- Financial correctness and boarding UX are more important than forcing payment to depend on live vehicle signals.

Tradeoff accepted:

- A valid QR can still be scanned even when live telemetry is stale.
- That is acceptable because the service window, confirmation screen, scan audits, and QR rotation controls are enough for the bounded campus use case.

### Decision 012: Student location stays on device and acts as a warning layer, not a hard gate

Decision:

- Run a campus-geofence safety check on device.
- Do not send raw student GPS coordinates to the backend.
- If the device appears outside campus or permission is denied, warn and require extra confirmation, but still allow override.

Reasoning:

- This reduces privacy risk and admin complexity.
- The purpose is to catch careless remote scans, not to enforce precise geofencing everywhere.
- Stop-level geofence setup is too heavy for the intended campus operator.

Tradeoff accepted:

- Remote-scan protection is advisory rather than absolute.
- That is acceptable because the user is primarily protected from accidentally burning their own money, while suspicious patterns are still audit-visible.

### Decision 013: One role-based operations app is better than separate admin and cashier portals

Decision:

- Use one shared web app for admin, cashier, and technical-admin users.
- Restrict features through role-based screens and actions.

Reasoning:

- This is easier to deploy, learn, and maintain for self-hosted institutes.
- Campus operators should not need to juggle multiple portals for closely related work.

Tradeoff accepted:

- The application needs careful role-based navigation and permission handling.
- That is acceptable because the operational simplicity outweighs the added internal access-control work.

### Decision 014: Driver attachment is dynamic and authenticated, not pre-assigned and not QR-authorized

Decision:

- Allow drivers to attach themselves to the current eligible service instance after login by scanning bus QR or entering bus code.
- Treat the bus QR only as bus selection, never as driver authorization.

Reasoning:

- Drivers can change between trips in real operations.
- Rigid pre-assignment would create unnecessary friction.
- Using the same bus QR as a selector keeps the field workflow simple without making the public QR itself privileged.

Tradeoff accepted:

- Driver assignment becomes a little more dynamic and needs clear conflict handling.
- That is acceptable because it better matches real operations and still keeps authorization anchored to authenticated driver identity.

### Decision 015: Audit logs stay immutable; investigation notes are separate records

Decision:

- Keep `audit_logs` immutable after insert.
- Store operator commentary in a separate `audit_investigation_notes` table linked by `audit_log_id`.

Reasoning:

- The source audit event should remain a trustworthy record of what happened.
- Investigation context is useful, but it should be appended alongside the audit event rather than changing the event itself.

Tradeoff accepted:

- The data model and admin UI are slightly more complex than editable audit rows.
- That is acceptable because the integrity of the audit trail matters more than the convenience of mutating a single record.

### Decision 016: Driver app is Android-first, battery-aware, and deliberately minimal

Decision:

- Support Android only in v1, while keeping Flutter implementation choices compatible with future iOS builds where practical.
- Use employee ID plus password for driver login.
- Emit telemetry every 10 seconds with foreground-service support and 30-minute local buffering.
- Keep the UI focused on bus attach, service state, health indicators, and simple notices.

Reasoning:

- The real deployment environment is personal Android phones, not idealized managed devices.
- Reliability and usability matter more than maximizing telemetry granularity or shipping a dense feature set.
- A simpler driver app is more likely to survive real-world operational use.

Tradeoff accepted:

- Fleet movement is slightly less granular than the earlier 3-second concept.
- That is acceptable because the small campus fleet and bounded route network do not need higher-frequency telemetry badly enough to justify the extra battery cost.

## Ongoing Story Format

Add new entries using this format as the build progresses:

### Decision XXX: Short title

Decision:

- What changed.

Reasoning:

- Why this path was chosen.

Tradeoff accepted:

- What cost or limitation was accepted.

Impact on the build:

- What changed in architecture, sprinting, or implementation order.

## Next Expected Story Entries

- final Flutter map package choice around MapTiler integration
- caching policy details and storage limits
- how route corridors are authored and stored
- auth UX decisions for student, driver, and admin roles
- first concurrency benchmark results for boarding
- first telemetry throughput benchmark results for live tracking
