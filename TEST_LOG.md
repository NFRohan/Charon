# Charon Test Log

## Purpose

This file is the running record of real verification work performed during the build.

It is not the test plan. The plan lives in `PHASE_TEST_PLAN.md`.
This log records what was actually run, when it was run, what passed, and any notable limitations.

## Entry Format

- `Date`: when the verification was performed
- `Phase`: sprint or milestone context
- `Scope`: what was being validated
- `Environment`: local, CI, or device context
- `Checks`: commands or actions performed
- `Result`: pass or fail summary
- `Notes`: anything worth preserving for future debugging or auditability

## 2026-03-08

### Sprint 1 Bootstrap Verification

- `Phase`: Sprint 1
- `Scope`: repository scaffold, local infrastructure, migration tooling, and base service startup
- `Environment`: local Windows workstation with Docker Compose
- `Checks`:
  - `go test ./...` in `backend`
  - `flutter test` in `apps/student_app`
  - `flutter test` in `apps/driver_app`
  - `npm run build` in `apps/admin_app`
  - `npm run lint` in `apps/admin_app`
  - `docker compose -f deploy/docker-compose.yml config`
  - `powershell -ExecutionPolicy Bypass -File .\scripts\dev-up.ps1`
  - manual health check against `http://localhost:8080/healthz`
- `Result`: PASS
- `Notes`:
  - API health endpoint returned `{"status":"ok"}`.
  - Flutter template tests passed, but Android build tooling was not yet part of the local workflow.
  - The generated Flutter Android projects warned about very new local Java versions, but test execution still completed.

### Sprint 2 Auth and Session Foundation Verification

- `Phase`: Sprint 2
- `Scope`: secure authentication, session persistence, route protection, migration wiring, and development seed usability
- `Environment`: local Windows workstation with Docker Compose and live API verification against `http://localhost:8080`
- `Checks`:
  - `go test ./...` in `backend`
  - `docker compose -f deploy/docker-compose.yml config`
  - `powershell -ExecutionPolicy Bypass -File .\scripts\dev-up.ps1`
  - `powershell -ExecutionPolicy Bypass -File .\scripts\seed.ps1`
  - live API verification of:
    - `POST /auth/login`
    - `POST /auth/refresh`
    - `POST /auth/logout`
    - protected `GET /wallet/balance`
    - suspended-account login behavior
- `Result`: PASS
- `Notes`:
  - Login returned `200 OK` for the seeded student account.
  - Refresh returned `200 OK` and preserved the same refresh token, matching the stable-token mobile-network requirement.
  - `GET /wallet/balance` returned `401` without auth and `501` with a valid student token because the route is now protected but still intentionally stubbed.
  - After logout, the previously issued access token returned `401`, confirming DB-backed session revocation is enforced on authenticated requests.
  - Suspended login returned `403`, confirming disabled-account enforcement.
  - Docker helper scripts were updated to rebuild service images so local verification cannot silently run stale binaries.

### Sprint 3 Core Schema and Migration Baseline Verification

- `Phase`: Sprint 3
- `Scope`: core schema bootstrap, fresh-database migration, and development route or wallet seed coverage
- `Environment`: local Windows workstation with Docker Compose and a fresh Postgres volume
- `Checks`:
  - `go test ./...` in `backend`
  - `docker compose -f deploy/docker-compose.yml down -v`
  - `powershell -ExecutionPolicy Bypass -File .\scripts\dev-up.ps1`
  - `powershell -ExecutionPolicy Bypass -File .\scripts\seed.ps1`
  - `powershell -ExecutionPolicy Bypass -File .\scripts\migrate.ps1 status`
  - manual health check against `http://localhost:8080/healthz`
  - Postgres sanity queries for:
    - `wallet_accounts`
    - `finance_adjustments`
    - `routes`
    - `stops`
    - `buses`
    - `service_calendars`
    - `trip_templates`
    - `trip_stop_times`
    - `route_fare_rules`
- `Result`: PASS
- `Notes`:
  - Migration version `3` applied cleanly from zero on a fresh database.
  - Development seed created `4` wallet accounts: `2` student accounts and `2` system accounts.
  - Development seed created `2` routes, `5` stops, `2` buses, `2` service calendars, `4` trip templates, `12` trip stop times, and `4` stop-based fare rules.
  - Route A is seeded as `FLAT_ROUTE` and Route B is seeded as `STOP_MATRIX` so both fare-model branches have local fixtures before boarding work begins.

### Sprint 2 Auth Hardening Follow-up Verification

- `Phase`: Sprint 2 follow-up
- `Scope`: production auth hardening for config secrets and password hashing recommendations
- `Environment`: local Windows workstation
- `Checks`:
  - `gofmt -w backend/internal/config/config.go backend/internal/config/config_test.go backend/internal/domain/auth/password.go backend/internal/domain/auth/password_test.go backend/internal/app/api.go`
  - `go test ./...` in `backend`
- `Result`: PASS
- `Notes`:
  - Production startup now rejects the exact Docker Compose development placeholder secrets for access-token signing and refresh-token pepper values.
  - Auth now selects stronger Argon2id parameters in production while keeping development and test defaults cheap enough for local iteration.
  - Existing DB-backed session validation still means logout revocation is enforced on authenticated requests immediately, rather than waiting for access-token expiry.
