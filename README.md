# Charon

Charon is a self-hosted university transit platform with three product surfaces:

- `apps/student_app`: Flutter app for riders.
- `apps/driver_app`: Flutter app for drivers.
- `apps/admin_app`: Next.js web app for admin, cashier, and technical admin workflows.
- `backend`: Go API and worker processes backing wallet, boarding, telemetry, alerts, and outbox delivery.

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
|   |   `-- worker/
|   `-- internal/
|       |-- config/
|       |-- domain/
|       |-- httpapi/
|       `-- platform/
|-- deploy/
|   |-- docker-compose.yml
|   `-- postgres/init/
`-- *.md
```

## Quick start

1. Copy `.env.example` to `.env` and adjust values if needed.
2. Start infrastructure:

   ```powershell
   docker compose -f deploy/docker-compose.yml up -d postgres redis rabbitmq
   ```

3. Start the API:

   ```powershell
   cd backend
   go run ./cmd/api
   ```

4. Start the worker in another terminal:

   ```powershell
   cd backend
   go run ./cmd/worker
   ```

5. Start the admin app:

   ```powershell
   cd apps/admin_app
   npm install
   npm run dev
   ```

6. Run either mobile app:

   ```powershell
   cd apps/student_app
   flutter run
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

- The Go service is scaffolded with domain seams matching the architecture spec, but the business logic is still to be implemented.
- The mobile apps are real Flutter shells, not placeholder folders, so Android and iOS platform directories already exist where needed.
- The generated Flutter Android projects may require either a compatible JDK or a Gradle wrapper upgrade on machines using very new Java versions.

## Key specifications

- `COMPREHENSIVE_SPEC.md`
- `API_SPEC.md`
- `ADMIN_CASHIER_API_SPEC.md`
- `DRIVER_SERVICE_API_SPEC.md`
- `STUDENT_SELF_SERVICE_API_SPEC.md`
- `SYSTEM_OPS_API_SPEC.md`
