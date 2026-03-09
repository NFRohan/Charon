# Charon

Charon is a self-hosted university transit platform with three product surfaces:

- `apps/student_app`: Flutter app for riders.
- `apps/driver_app`: Flutter app for drivers.
- `apps/admin_app`: Next.js web app for admin, cashier, and technical admin workflows.
- `backend`: Go API and worker processes backing wallet, boarding, telemetry, alerts, and outbox delivery.
- `scripts`: local developer helpers for booting infrastructure, migrations, and seeds.

## Repository layout

```text
.
|-- apps/
|   |-- admin_app/
|   |-- driver_app/
|   `-- student_app/
|-- backend/
|   |-- cmd/
|   |   |-- api/
|   |   |-- migrate/
|   |   |-- seed/
|   |   `-- worker/
|   |-- migrations/
|   |-- seeds/
|   `-- internal/
|       |-- app/
|       |-- config/
|       |-- domain/
|       |-- httpapi/
|       `-- platform/
|-- deploy/
|   |-- docker-compose.yml
|   `-- postgres/init/
|-- scripts/
`-- *.md
```

## Quick start

1. Copy `.env.example` to `.env` and adjust values if needed.
2. Start local services:

   ```powershell
   ./scripts/dev-up.ps1
   ```

3. Start the admin app:

   ```powershell
   cd apps/admin_app
   npm install
   npm run dev
   ```

4. Run either mobile app:

   ```powershell
   cd apps/student_app
   flutter run
   ```

## Backend utility commands

- Apply migrations:

   ```powershell
   ./scripts/migrate.ps1 up
   ```

- Check migration status:

   ```powershell
   ./scripts/migrate.ps1 status
   ```

- Apply environment seeds:

   ```powershell
   ./scripts/seed.ps1
   ```

- Stop local services:

   ```powershell
   ./scripts/dev-down.ps1
   ```

## Local infrastructure

`deploy/docker-compose.yml` provisions:

- PostgreSQL with PostGIS enabled
- Redis
- RabbitMQ with the management UI
- the Go API container
- the Go worker container

The compose stack is intentionally minimal because the product is designed for single-institute deployments.

## Notes

- The Go service now includes validated config loading, application bootstrap wiring, auth and session flows, migration and seed commands, and the first wallet read APIs for balance and transaction history.
- Development auth seed credentials are loaded by `./scripts/seed.ps1`:
  - student: `220041234` / `ChangeMe123!`
  - driver: `DRV1001` / `ChangeMe123!`
  - cashier: `CASH1001` / `ChangeMe123!`
  - admin: `ADM1001` / `ChangeMe123!`
- The same development seed also creates baseline wallet accounts, system ledger accounts, routes, stops, buses, calendars, and trip templates for local feature work.
- The mobile apps are real Flutter shells, not placeholder folders, so Android and iOS platform directories already exist where needed.
- The generated Flutter Android projects may require either a compatible JDK or a Gradle wrapper upgrade on machines using very new Java versions.
- Migration files live in `backend/migrations`, while environment-specific seed files live in `backend/seeds/<environment>`.

## Key specifications

- `COMPREHENSIVE_SPEC.md`
- `API_SPEC.md`
- `ADMIN_CASHIER_API_SPEC.md`
- `DRIVER_SERVICE_API_SPEC.md`
- `STUDENT_SELF_SERVICE_API_SPEC.md`
- `SYSTEM_OPS_API_SPEC.md`
- `PHASE_TEST_PLAN.md`
- `TEST_LOG.md`
