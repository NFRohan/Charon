# Charon Engineering Story

## Purpose

This is the living engineering story for Charon. It records what has been done, which decisions were taken, why those decisions were made, and how each decision changes the build path.

The point is not just to preserve outcomes. It is to preserve reasoning so future changes stay coherent instead of turning into disconnected one-off choices.

## Current Status

Implementation is underway. The project currently has:

- a system architecture plan
- a 20-week sprint plan
- a comprehensive specification
- a locked first major platform decision for map rendering: MapTiler with client-side caching in Flutter
- a bootstrapped monorepo with local infrastructure and service startup tooling
- a completed Sprint 2 auth and session foundation
- a completed Sprint 3 core schema and development seed baseline

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
- Replaced the simple flat-fare assumption with route-level flat fare, selected-stop fare, or zero-fare policy support.
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

### 14. Student app operating model defined

Work completed:

- Defined the student app as a real cross-platform target for both Android and iOS.
- Made the app home-first with `Home`, `Wallet`, `Map`, `Alerts`, and `Profile` instead of a dedicated pay tab.
- Added stop selection to every boarding attempt so the same interaction supports stop-specific ETA and configurable fare rules.
- Kept boarding retry behavior manual so the student stays in control when the network is weak.
- Kept top-up guidance cashier-based instead of pretending online funding exists before it does.

Why this was done:

- The student app is the most visible user-facing part of the project and needs to feel like a finished product, not just a wallet screen.
- Student device mix is broader than the driver-device mix, so iOS matters here even though it does not matter for the driver app yet.
- Stop selection creates one consistent boarding input that works for flat fares now and more flexible campus fare models later.
- Manual retry keeps the payment experience understandable instead of hiding request uncertainty behind client-side automation.

### 15. Student boarding fallback model defined

Work completed:

- Added a three-layer boarding fallback model of direct self-pay, sponsored boarding, and emergency ride permits.
- Defined sponsored boarding as one connected student paying for self plus one additional rider in one atomic transaction.
- Added a bounded emergency ride permit path instead of full offline wallet sync.
- Kept rider-level boarding events and duplicate protection intact even when one payer covers multiple riders.

Why this was done:

- A student losing internet at boarding is a real field case, and "find a friend" is not enough as the only backup.
- Full offline wallet synchronization would make the financial model much riskier and more complex than the project needs.
- Sponsored boarding fits real campus behavior, while emergency permits cover the no-friend solo case without breaking the architecture.
- The fallback needed to stay consistent with the ledger, idempotency, and schedule-backed boarding rules already chosen.

### 16. First wire-level API contract defined

Work completed:

- Added the first dedicated API spec document for the critical flows only.
- Locked the API to unversioned paths in v1 with one unified login endpoint.
- Chose server-calculated `GET /boardings/preview` plus one unified `POST /boardings` with mode switching.
- Chose raw QR payload submission, rich error envelopes, limit-and-offset pagination, and one shared WebSocket per app.
- Kept the public live API anonymous but limited it to pre-shaped public models only.

Why this was done:

- The project had enough product detail that implementation without a wire contract would start creating avoidable drift between Flutter, Go, and the public view.
- The API needed to reflect the actual constraints already chosen: bad mobile networks, server-owned fare logic, rider-safe boarding flows, and privacy-safe public route data.
- A first-draft API spec keeps focus on the risky flows first instead of pretending the whole admin surface needs to be frozen before coding begins.

### 17. Deferred API surface captured explicitly

Work completed:

- Added a separate noncritical API document for the deferred endpoint inventory.
- Captured student quality-of-life APIs, driver attachment and recovery APIs, admin CRUD, alerts, audit logs, imports or exports, and technical-admin ops endpoints.
- Split the API planning into critical-path wire contract versus backlog contract instead of leaving the rest as vague future work.

Why this was done:

- Once implementation starts, undocumented secondary APIs are easy to postpone until they quietly disappear from scope.
- The project already has enough surface area that "we will remember later" is not a reliable planning strategy.
- A backlog-style API inventory keeps the implementation roadmap honest without forcing every low-risk endpoint to be frozen in full detail up front.

### 18. Admin and cashier API contract locked next

Work completed:

- Added a dedicated admin and cashier API contract document instead of leaving that domain in backlog form.
- Split the shared web-app API into dashboard, student policy, wallet ops, buses, routes, fare policy, schedules, service instances, alerts, advisories, audit, and import or export sections.
- Kept deployer-controlled fare policy flexibility so a route can use flat pricing, zero-fare, or stop-based matrices.
- Added the admin-only `force-close` service-instance endpoint so intervention remains auditable and separate from driver start and end actions.

Why this was done:

- The shared admin app is too large and too operationally important to remain only a backlog list once implementation planning starts.
- Cashier and admin behavior affects finance correctness, bus onboarding, route setup, and service reliability, so it benefits from a tighter contract before coding.
- Fare configuration needed to stay flexible because different institutes may want route-flat or stop-based pricing without changing the boarding API shape.

### 19. Remaining API domains were converted from backlog to contracts

Work completed:

- Added a dedicated driver and service-attachment API contract.
- Added a dedicated student self-service API contract.
- Added a dedicated technical-admin system-ops API contract.
- Reduced the role of the noncritical API document from "missing major domains" to a true backlog and future-surface tracker.

Why this was done:

- At this point the project has enough wire-level definition to move into implementation without large blind spots.
- Driver attachment, device-health reporting, student favorites and settings, and DLQ tooling all affect real implementation structure even if they are not the first demo clicks.
- Finishing the contract set now is cheaper than discovering missing edge cases halfway through coding.

### 20. Delivery sequencing expanded to 20 weeks with mobile implementation deferred

Work completed:

- Replaced the original 10-week sprint plan with a 20-week plan.
- Moved real student and driver feature implementation to the end of the schedule.
- Rebalanced the roadmap so backend, admin, and system-ops work land before heavy Flutter implementation.

Why this was done:

- The active development environment is ready for Go and Next.js work, but not for full Android or Flutter feature execution.
- The backend and admin surfaces carry more architectural risk than the mobile UI shells, so they should stabilize first anyway.
- Deferring real mobile implementation reduces churn because the Flutter apps can build against stable contracts instead of moving targets.

### 21. Delivery sequencing revised again for cloud-built mobile validation

Work completed:

- Revised the 20-week plan so student and driver implementation move into the middle of the schedule instead of the very end.
- Added GitHub Actions as a practical way to produce APK artifacts for real-device testing.
- Reframed the schedule as backend-first but mobile-parallel once the relevant APIs stabilize.

Why this was done:

- The main blocker was not Flutter itself. It was the lack of a comfortable local Android build loop.
- Cloud APK builds are enough to validate the student and driver apps on real phones while backend work continues locally.
- Bringing mobile in earlier will expose boarding UX, telemetry behavior, and API mistakes sooner.
- The backend still deserves to lead the plan because ledger safety, outbox reliability, and telemetry architecture remain the highest-risk work.

### 22. Repository scaffold and local infrastructure bootstrap landed

Work completed:

- Created the monorepo scaffold for backend, mobile apps, admin app, scripts, and deployment files.
- Added Docker Compose-backed local infrastructure for PostgreSQL, Redis, RabbitMQ, API, worker, migration, and seed execution.
- Added migration and seed commands and a basic health endpoint.
- Added a phase-based test plan to define verification gates before feature work ramps up.

Why this was done:

- The project needed a reproducible local environment before domain features could be implemented safely.
- Starting with infrastructure and workflow tooling reduces friction for every later sprint.
- The build needed a credible operational baseline, not just a folder structure.

### 23. Secure auth and session foundation implemented

Work completed:

- Added the `users` and `auth_sessions` schema.
- Implemented unified login, refresh, and logout with DB-backed session state.
- Added Argon2id password hashing and opaque HMAC-hashed refresh-token storage.
- Added JWT access-token issuance and middleware that re-checks session and account state on protected requests.
- Added development auth seeds, route protection, and automated tests for config, token, service, and HTTP auth paths.

Why this was done:

- Financially sensitive systems need authentication and session revocation to be first-class, not bolted on later.
- The project’s next major vertical slices depend on role-aware protected routes.
- Verifying session state from the database on authenticated requests gives stronger revocation behavior than trusting JWTs alone.
- Stable refresh tokens match the earlier mobile-network decision while still keeping refresh tokens off the database in raw form.

### 24. Core schema and development fixtures implemented

Work completed:

- Added the first full domain schema migration covering wallet accounts, transactions, ledger entries, routes, stops, buses, schedules, route sessions, boarding records, audit records, outbox records, alerts, telemetry archival, and finance adjustments.
- Added reusable `updated_at` trigger wiring for mutable tables.
- Added development seed fixtures for wallet accounts, system ledger accounts, routes, stop sequences, fare rules, buses, calendars, and trip templates.
- Verified the schema from zero on a fresh Postgres volume instead of only migrating forward on an already-used database.

Why this was done:

- The project needed a durable, testable data model before wallet logic, fare calculation, and boarding flows start writing money-adjacent code.
- A fresh-db migration check is the real bar for schema work, because drift often hides inside already-migrated development databases.
- Rich seed data reduces friction for the next sprints by giving wallet, route, and QR work real fixtures to build against.

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

### Decision 023: Build backend and admin first, then implement mobile features late in the schedule

Decision:

- Expand the delivery plan from 10 weeks to 20 weeks.
- Treat the backend and admin web app as the primary implementation focus for the first 17 weeks.
- Push real Flutter feature work into the final implementation window after the service contracts and ops workflows are stable.

Reasoning:

- The current development environment does not yet support comfortable end-to-end Android work.
- The higher-risk engineering work is in concurrency, ledger correctness, telemetry reliability, schedules, and operator tooling, not in drawing mobile screens.
- Stable APIs and admin workflows make the later mobile work faster and less error-prone.

Tradeoff accepted:

- The visible mobile experience will arrive later in the timeline than it would in a UI-first build.
- That is acceptable because the system's hardest guarantees and operational model need to be correct before mobile polish matters.

### Decision 024: Move mobile work into the middle of the plan once cloud APK builds are available

Decision:

- Keep the schedule backend-first.
- Start the student app after auth, wallet, and boarding APIs are stable.
- Start the driver app after service-instance and telemetry APIs are stable.
- Use GitHub Actions to build mobile artifacts for real-device validation during development.

Reasoning:

- Cloud builds remove the strongest reason to delay mobile implementation all the way to the end.
- Student boarding and driver telemetry are too central to leave until the final sprint window if they can be tested earlier.
- Real-device feedback in the middle of the project is more valuable than discovering UX and integration issues during final hardening.
- This keeps the plan pragmatic: the backend still leads, but mobile work no longer waits for everything else to be complete.

Tradeoff accepted:

- The project now has more parallelism and therefore more context-switching than the previous backend-then-mobile sequence.
- That is acceptable because the earlier integration feedback is worth the scheduling complexity.

### Decision 025: Access tokens are short-lived JWTs, but session truth stays in PostgreSQL

Decision:

- Use short-lived JWT access tokens for authenticated requests.
- Use stable opaque refresh tokens whose hashed values are stored in PostgreSQL.
- On every protected request, validate the JWT and then load the backing session from PostgreSQL before trusting the identity.

Reasoning:

- JWTs keep the request path simple for clients, but money-adjacent systems need reliable revocation and account-state enforcement.
- A purely stateless token model would allow logged-out or revoked sessions to remain usable until token expiry.
- Stable refresh tokens are still the right fit for weak mobile networks, but storing only hashed token material reduces blast radius if the session table is exposed.
- This model keeps revocation strong without forcing every client into refresh-token rotation race conditions.

Tradeoff accepted:

- Every protected request now performs a database lookup instead of trusting the access token alone.
- That is acceptable because the expected deployment scale is small and the security benefit is worth the cost.

Impact on the build:

- Protected routes now depend on both JWT validation and session-table state.
- Logout immediately invalidates the session for future authenticated requests.
- The auth schema and middleware became Sprint 2’s main foundation instead of a lightweight placeholder.

### Decision 026: Ledger accounts must cover both students and system counterparties

Decision:

- Model `wallet_accounts` so they can belong either to a user or to a named system account.
- Seed explicit system accounts for fare collection and cashier settlement.
- Add a separate `finance_adjustments` table so manual credits and refunds have their own approval workflow record instead of being implied only through transactions.

Reasoning:

- A real double-entry ledger cannot stop at student wallets; every debit needs a credible counterparty account.
- Bus fares, cashier credits, and refunds all need traceable system-side accounts if the ledger is going to remain balanced and auditable.
- Manual adjustments are operational events as well as financial events, so they deserve their own workflow record instead of being hidden inside transaction metadata.

Tradeoff accepted:

- The account model is more flexible and therefore more complex than a simple `user_id -> balance` table.
- That is acceptable because the simpler model would collapse as soon as the first system-side credit leg or approval flow appears.

Impact on the build:

- Future wallet logic can stay truly double-entry without inventing fake or nullable counterparties.
- Cashier credit and refund workflows now have a clean table to target in Sprint 4.
- Development seeds now include both student and system accounts, which makes ledger testing practical much earlier.

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
- Support route-level flat fare, selected-stop fare, and zero-fare policies.
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

### Decision 017: Student app is home-first, cross-platform, and stop-aware

Decision:

- Support Android and iOS for the student app in v1.
- Use a home-first app shape with `Home`, `Wallet`, `Map`, `Alerts`, and `Profile`.
- Keep `Scan to Pay` as a prominent action instead of dedicating a full tab to payment.
- Require stop selection during boarding so fare and ETA both operate on the same rider input.

Reasoning:

- Students are more likely than drivers to use a broad device mix, including iPhones.
- A separate pay tab would take permanent navigation space for a single action that can be reached cleanly from home and wallet.
- Stop selection gives the product a realistic path to support both flat and stop-based campus fare systems without jumping to GPS-derived pricing.
- The same selected stop also improves ETA usefulness immediately, so the extra tap carries real product value.

Tradeoff accepted:

- Boarding has one more explicit user input than a bare scan-and-pay flow.
- That is acceptable because the stop choice improves both fare accuracy and ETA relevance while still keeping the flow simple enough for daily use.

### Decision 018: Boarding fallback stays bounded instead of turning the wallet fully offline

Decision:

- Use a three-layer fallback model: normal online self-pay, sponsored boarding, and emergency ride permits.
- Let sponsored boarding cover self plus one additional rider in v1.
- Use pre-issued, device-bound, one-time emergency permits for the solo no-internet case.

Reasoning:

- Student connectivity failure at boarding is common enough to deserve a first-class product answer.
- A fully offline-sync wallet would create a much larger correctness and abuse surface than the project needs.
- Sponsored boarding mirrors real human behavior, and emergency permits cover the cases where a helpful friend is not available.
- This keeps the financial model mostly online while still giving students a humane fallback.

Tradeoff accepted:

- Emergency ride permits introduce a small bounded offline-trust surface and additional issuance or redemption logic.
- That is acceptable because the risk is tightly capped and much lower than building a general offline payment system.

### Decision 019: API v1 is unversioned, REST-first, and server-trusted

Decision:

- Use unversioned REST paths in v1.
- Use one unified auth login endpoint and one unified boarding endpoint with mode-based branching.
- Require server-side preview for fare confirmation and raw QR payload submission for security-sensitive scans.
- Use one shared WebSocket per mobile app with explicit message types and telemetry acknowledgements.

Reasoning:

- The project is still early enough that path versioning would add ceremony without solving a real compatibility problem yet.
- The most important contract boundary is not resource purity. It is keeping money and boarding decisions centralized and auditable.
- Preview plus submit matches the product UX while preventing client-side fare logic from drifting.
- A single socket per app is the simplest mobile-operational model that still supports telemetry, ETA, and alerts cleanly.

Tradeoff accepted:

- Some endpoints are more workflow-oriented than perfectly REST-pure.
- That is acceptable because the system is optimizing for correctness, mobile resilience, and implementation clarity rather than textbook API aesthetics.

### Decision 020: Split API planning into critical contract and deferred inventory

Decision:

- Keep one detailed API spec for the risky first-build flows.
- Keep a second backlog-style API spec for the noncritical and deferred surface.

Reasoning:

- Not every endpoint deserves the same planning depth at the same time.
- Critical flows such as auth, boarding, and telemetry need example payloads and tighter contracts before coding starts.
- Secondary surfaces such as admin CRUD, imports, audit browsing, and quality-of-life endpoints still need to stay visible so they are not forgotten.

Tradeoff accepted:

- The project now has two API planning documents instead of one.
- That is acceptable because the split reduces overload while still preserving implementation visibility across the full product surface.

### Decision 021: Admin and cashier APIs are locked before the rest of the deferred surface

Decision:

- Turn the admin and cashier domain into a full contract before fully specifying the remaining deferred APIs.
- Keep fare-policy control in admin hands with support for flat route pricing, zero-fare, and stop-based matrices.

Reasoning:

- The shared web app sets up and governs most of the platform's behavior, so it has outsized influence on whether the rest of the system remains coherent.
- Locking this domain first gives the project a strong operational backbone without forcing every lower-priority endpoint to be overdesigned up front.
- Fare-policy flexibility is part of the product's deployer story and should not be lost just because the first admin contract is being written.

Tradeoff accepted:

- Some lower-priority driver and technical-admin APIs remain at backlog level for now.
- That is acceptable because the most operationally central deferred domain is now specified in detail.

### Decision 022: Finish the remaining API contracts before coding the platform

Decision:

- Convert the remaining high-value deferred API domains into dedicated contracts before implementation begins.
- Keep only truly optional or future surfaces in the backlog inventory.

Reasoning:

- The project has crossed the point where undocumented domains create more risk than the extra planning time costs.
- Driver recovery, student self-service, and technical-admin ops all influence service boundaries, state handling, and UI work in ways that are expensive to improvise later.
- A near-complete contract set lowers coordination cost once coding starts across Go, Flutter, and the admin app.

Tradeoff accepted:

- The spec surface is now broader and more document-heavy than a minimal prototype would require.
- That is acceptable because Charon is intentionally being built as a portfolio-grade, implementation-ready system rather than a loose concept sketch.

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
- `device_tokens` schema and push-notification storage once notification work begins
- how route corridors are authored and stored
- auth UX decisions for student, driver, and admin roles
- first concurrency benchmark results for boarding
- first telemetry throughput benchmark results for live tracking
