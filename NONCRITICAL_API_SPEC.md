# Charon Noncritical API Specification

## Document Status

- Status: Draft backlog v1
- Date: 2026-03-08
- Purpose: Track the noncritical and deferred API surface so implementation does not forget important product and operations endpoints after the critical-path APIs are built.

## 1. How To Use This Document

This document is not a fully frozen wire contract like [API_SPEC.md](e:\Projects\Charon\API_SPEC.md).

Admin and cashier endpoints now have a dedicated full contract in [ADMIN_CASHIER_API_SPEC.md](e:\Projects\Charon\ADMIN_CASHIER_API_SPEC.md).
Student self-service endpoints now have a dedicated full contract in [STUDENT_SELF_SERVICE_API_SPEC.md](e:\Projects\Charon\STUDENT_SELF_SERVICE_API_SPEC.md).
Driver and service-attachment endpoints now have a dedicated full contract in [DRIVER_SERVICE_API_SPEC.md](e:\Projects\Charon\DRIVER_SERVICE_API_SPEC.md).
Technical-admin system-ops endpoints now have a dedicated full contract in [SYSTEM_OPS_API_SPEC.md](e:\Projects\Charon\SYSTEM_OPS_API_SPEC.md).

Where this file still overlaps with a dedicated API contract, the dedicated contract wins.

It is an implementation inventory for endpoints that:

- are important to the finished product
- are already implied by the specs
- are not the first coding priority
- need to remain visible so they are not skipped

Each section uses these priority labels:

- `Soon`: needed for the first full usable product, but not required before critical boarding or telemetry work starts
- `Later`: needed after the core demo path is working
- `Optional`: useful future surface, not required for the first showcase milestone

## 2. Shared Rules

Unless a section says otherwise, these endpoints inherit the conventions in [API_SPEC.md](e:\Projects\Charon\API_SPEC.md):

- unversioned paths
- bearer-token auth for protected endpoints
- JSON request and response bodies
- rich error envelope
- limit-and-offset pagination
- `Idempotency-Key` on money-moving or other important mutating `POST` requests

## 3. Student App Deferred APIs

These endpoints support quality-of-life features already present in the student product spec.

### 3.1 Profile and Settings

- `Soon` `GET /me/profile`
  Purpose: fetch student profile summary, account status, and app-facing flags.
- `Soon` `POST /me/password/change`
  Purpose: allow logged-in password change even though password reset is admin-assisted.
- `Later` `GET /me/notification-settings`
  Purpose: read simple rider notification preferences.
- `Later` `POST /me/notification-settings`
  Purpose: update simple rider notification preferences.

### 3.2 Favorites

- `Soon` `GET /me/favorite-routes`
  Purpose: read favorite routes for home and map shortcuts.
- `Soon` `POST /me/favorite-routes`
  Purpose: add a route to favorites.
- `Soon` `DELETE /me/favorite-routes/{route_id}`
  Purpose: remove a route from favorites.
- `Soon` `GET /me/favorite-stops`
  Purpose: read favorite stops for ETA and boarding shortcuts.
- `Soon` `POST /me/favorite-stops`
  Purpose: add a stop to favorites.
- `Soon` `DELETE /me/favorite-stops/{stop_id}`
  Purpose: remove a stop from favorites.

### 3.3 Alerts and Read State

- `Soon` `GET /alerts?status=active&route_code=A&limit=20&offset=0`
  Purpose: populate the student alerts screen with filtering.
- `Soon` `POST /alerts/{alert_id}/read`
  Purpose: mark one alert as read.
- `Later` `POST /alerts/read-all`
  Purpose: mark a batch of alerts as read.

## 4. Driver App Deferred APIs

These endpoints support the driver workflow outside the telemetry socket itself.

### 4.1 Bus Attach and Service Context

- `Soon` `GET /driver/active-service`
  Purpose: restore current service state after app restart.
- `Soon` `GET /driver/bus-attachment/preview?qr_payload=...` or `?bus_code=1042`
  Purpose: show eligible service instances before driver confirms attachment.
- `Soon` `POST /driver/bus-attachments`
  Purpose: attach the authenticated driver to the selected bus and eligible service instance.
- `Soon` `POST /route-sessions/start`
  Purpose: record the operational start marker.
- `Soon` `POST /route-sessions/end`
  Purpose: record the operational end marker.

### 4.2 Driver Notices

- `Later` `GET /driver/notices?limit=20&offset=0`
  Purpose: read recent route cancellations and service advisories outside live socket delivery.

## 5. Admin Dashboard and Search APIs

These endpoints support the web app landing experience and cross-module navigation.

### 5.1 Dashboard

- `Soon` `GET /admin/dashboard`
  Purpose: return read models for active buses, stale services, blocked scans, alert count, queue backlog, old QR usage, and failed wallet operations.

### 5.2 Global Search

- `Soon` `GET /admin/students/search?q=22004`
  Purpose: search by institutional ID or name for both admin and cashier use.
- `Later` `GET /admin/buses/search?q=1042`
  Purpose: search buses quickly from admin workflows.

## 6. Admin Student and Wallet Ops APIs

### 6.1 Student Policy Management

- `Soon` `GET /admin/students/{student_id}`
  Purpose: fetch operational student details for admin review.
- `Soon` `PATCH /admin/students/{student_id}`
  Purpose: update fare exemption, overdraft limit, account status, route eligibility, and internal notes.

### 6.2 Wallet Read APIs

- `Soon` `GET /admin/wallets/{student_id}`
  Purpose: fetch wallet summary for finance support.
- `Soon` `GET /admin/wallets/{student_id}/transactions?limit=20&offset=0`
  Purpose: read wallet transaction history in the admin app.

### 6.3 Credits and Refunds

- `Soon` `POST /admin/wallets/{student_id}/credits`
  Purpose: issue cashier or admin credits.
- `Soon` `POST /admin/wallets/{student_id}/refunds`
  Purpose: create direct refund or approval-required refund request.
- `Soon` `GET /admin/refund-requests?status=pending&limit=20&offset=0`
  Purpose: show approval queue for large refunds.
- `Soon` `POST /admin/refund-requests/{request_id}/approve`
  Purpose: approve a pending refund request.
- `Soon` `POST /admin/refund-requests/{request_id}/reject`
  Purpose: reject a pending refund request.

## 7. Bus Registry and QR APIs

### 7.1 Bus CRUD

- `Soon` `GET /admin/buses?status=active&limit=20&offset=0`
  Purpose: list buses with filters.
- `Soon` `POST /admin/buses`
  Purpose: create a bus with bus code, plate, route, status, seat capacity, and notes.
- `Soon` `GET /admin/buses/{bus_id}`
  Purpose: fetch one bus record.
- `Soon` `PATCH /admin/buses/{bus_id}`
  Purpose: update bus metadata and operational status.

### 7.2 QR Operations

- `Soon` `POST /admin/buses/{bus_id}/qr/generate`
  Purpose: generate or regenerate the current bus QR PNG asset.
- `Soon` `POST /admin/buses/{bus_id}/qr/rotate`
  Purpose: bump `qr_version` and issue a replacement asset.
- `Later` `GET /admin/buses/{bus_id}/qr/history`
  Purpose: inspect QR version history and grace-period usage.

## 8. Routes, Stops, and Fare Rules APIs

### 8.1 Routes

- `Soon` `GET /admin/routes?limit=20&offset=0`
  Purpose: list routes.
- `Soon` `POST /admin/routes`
  Purpose: create route metadata and primary fare policy.
- `Soon` `GET /admin/routes/{route_id}`
  Purpose: fetch a route with current stop and fare configuration.
- `Soon` `PATCH /admin/routes/{route_id}`
  Purpose: update route metadata and route status.

### 8.2 Stops

- `Soon` `GET /admin/stops?limit=50&offset=0`
  Purpose: list stops for admin selection.
- `Soon` `POST /admin/stops`
  Purpose: create a stop.
- `Soon` `GET /admin/stops/{stop_id}`
  Purpose: fetch a stop.
- `Soon` `PATCH /admin/stops/{stop_id}`
  Purpose: update stop metadata and position.

### 8.3 Stop Ordering and Fare Rules

- `Soon` `PUT /admin/routes/{route_id}/stops`
  Purpose: replace the ordered stop list for form-based route editing.
- `Later` `GET /admin/routes/{route_id}/fare-rules`
  Purpose: read stop-based or service-based fare rules.
- `Later` `POST /admin/routes/{route_id}/fare-rules`
  Purpose: add a fare rule.
- `Later` `PATCH /admin/routes/{route_id}/fare-rules/{fare_rule_id}`
  Purpose: update a fare rule.

## 9. Schedule and Service Instance APIs

### 9.1 Weekly Templates and Calendars

- `Soon` `GET /admin/trip-templates?route_id=route_a`
  Purpose: list route trip templates.
- `Soon` `POST /admin/trip-templates`
  Purpose: create a trip template.
- `Soon` `PATCH /admin/trip-templates/{trip_template_id}`
  Purpose: update trip template metadata.
- `Soon` `GET /admin/trip-templates/{trip_template_id}/stop-times`
  Purpose: read scheduled stop times.
- `Soon` `PUT /admin/trip-templates/{trip_template_id}/stop-times`
  Purpose: replace scheduled stop times.
- `Soon` `GET /admin/service-calendars`
  Purpose: list weekly service calendars.
- `Soon` `POST /admin/service-calendars`
  Purpose: create a weekly service calendar.
- `Soon` `PATCH /admin/service-calendars/{calendar_id}`
  Purpose: update a weekly service calendar.

### 9.2 Exceptions and Ad Hoc Service Instances

- `Soon` `GET /admin/service-exceptions`
  Purpose: list holiday and special-event exceptions.
- `Soon` `POST /admin/service-exceptions`
  Purpose: create a holiday closure or schedule exception.
- `Soon` `PATCH /admin/service-exceptions/{exception_id}`
  Purpose: update a service exception.
- `Soon` `GET /admin/service-instances?status=running&limit=20&offset=0`
  Purpose: list current and upcoming service instances.
- `Soon` `POST /admin/service-instances`
  Purpose: create an ad hoc service instance.
- `Soon` `GET /admin/service-instances/{service_instance_id}`
  Purpose: fetch one service instance.
- `Later` `PATCH /admin/service-instances/{service_instance_id}`
  Purpose: edit notes or operational metadata where allowed.
- `Later` `POST /admin/service-instances/{service_instance_id}/cancel`
  Purpose: cancel a service instance and drive advisory logic.

## 10. Alerts and Public Service Feed APIs

### 10.1 Alerts

- `Soon` `GET /admin/alerts?status=open&severity=high&limit=20&offset=0`
  Purpose: populate the admin alerts module with filtering.
- `Soon` `GET /admin/alerts/{alert_id}`
  Purpose: fetch one alert with context.
- `Soon` `POST /admin/alerts/{alert_id}/acknowledge`
  Purpose: acknowledge an alert.
- `Soon` `POST /admin/alerts/{alert_id}/resolve`
  Purpose: resolve an alert.
- `Later` `POST /admin/alerts/{alert_id}/mute`
  Purpose: mute an alert.
- `Later` `POST /admin/alerts/{alert_id}/notes`
  Purpose: append alert notes.

### 10.2 Advisories and Public Feed

- `Soon` `GET /admin/advisories?limit=20&offset=0`
  Purpose: list current and recent public advisories.
- `Soon` `POST /admin/advisories`
  Purpose: create campus-wide or route-specific advisories and cancellations.
- `Soon` `PATCH /admin/advisories/{advisory_id}`
  Purpose: update advisory text or time window.
- `Later` `GET /admin/public-feed/preview`
  Purpose: preview the public live feed as the guardian surface will see it.

## 11. Audit Log APIs

Audit rows are immutable. Notes are append-only.

- `Soon` `GET /admin/audit-logs?actor_id=...&student_id=...&bus_id=...&route_id=...&action_type=...&result=...&date_from=...&date_to=...&limit=20&offset=0`
  Purpose: filtered audit-log search.
- `Soon` `GET /admin/audit-logs/{audit_log_id}`
  Purpose: fetch one audit event.
- `Soon` `GET /admin/audit-logs/{audit_log_id}/notes`
  Purpose: read investigation notes.
- `Soon` `POST /admin/audit-logs/{audit_log_id}/notes`
  Purpose: append investigation notes without mutating the original audit row.

## 12. Technical Admin and System Ops APIs

These endpoints are restricted to `technical_admin`.

### 12.1 Queue and Worker Visibility

- `Later` `GET /admin/system/queues`
  Purpose: show queue depth summary.
- `Later` `GET /admin/system/queues/{queue_name}`
  Purpose: inspect one queue.
- `Later` `GET /admin/system/queues/{queue_name}/dlq?limit=20&offset=0`
  Purpose: inspect dead-letter messages.
- `Later` `POST /admin/system/queues/{queue_name}/dlq/{message_id}/requeue`
  Purpose: requeue one dead-lettered message.
- `Later` `GET /admin/system/workers/health`
  Purpose: show worker heartbeat and health summary.

## 13. Imports and Exports APIs

These endpoints support bootstrap and correction workflows, not ongoing enterprise sync.

- `Later` `POST /admin/imports/buses`
  Purpose: upload CSV for buses.
- `Later` `POST /admin/imports/routes`
  Purpose: upload CSV for routes.
- `Later` `POST /admin/imports/stops`
  Purpose: upload CSV for stops.
- `Later` `POST /admin/imports/schedules`
  Purpose: upload CSV for schedules.
- `Later` `GET /admin/imports/{job_id}`
  Purpose: inspect import job result.
- `Later` `GET /admin/exports/buses`
  Purpose: export buses as CSV.
- `Later` `GET /admin/exports/routes`
  Purpose: export routes as CSV.
- `Later` `GET /admin/exports/stops`
  Purpose: export stops as CSV.
- `Later` `GET /admin/exports/schedules`
  Purpose: export schedules as CSV.

## 14. Still Optional or Future APIs

These are intentionally not part of the first implementation push:

- `Optional` session-management endpoints for users to inspect and revoke individual devices
- `Optional` public WebSocket or SSE feed for guardian live view
- `Optional` richer analytics and reporting export APIs
- `Optional` bulk bus QR operations
- `Optional` bulk route-edit operations beyond the current ordered-stop replacement model

## 15. Relationship To The Main API Spec

This document complements:

- [API_SPEC.md](e:\Projects\Charon\API_SPEC.md)
- [ADMIN_CASHIER_API_SPEC.md](e:\Projects\Charon\ADMIN_CASHIER_API_SPEC.md)
- [STUDENT_SELF_SERVICE_API_SPEC.md](e:\Projects\Charon\STUDENT_SELF_SERVICE_API_SPEC.md)
- [DRIVER_SERVICE_API_SPEC.md](e:\Projects\Charon\DRIVER_SERVICE_API_SPEC.md)
- [SYSTEM_OPS_API_SPEC.md](e:\Projects\Charon\SYSTEM_OPS_API_SPEC.md)
- [ADMIN_SPEC.md](e:\Projects\Charon\ADMIN_SPEC.md)
- [COMPREHENSIVE_SPEC.md](e:\Projects\Charon\COMPREHENSIVE_SPEC.md)

Use [API_SPEC.md](e:\Projects\Charon\API_SPEC.md) for endpoints that must be implemented first.

Use [ADMIN_CASHIER_API_SPEC.md](e:\Projects\Charon\ADMIN_CASHIER_API_SPEC.md) for admin and cashier wire-level details.
Use [STUDENT_SELF_SERVICE_API_SPEC.md](e:\Projects\Charon\STUDENT_SELF_SERVICE_API_SPEC.md) for student self-service wire-level details.
Use [DRIVER_SERVICE_API_SPEC.md](e:\Projects\Charon\DRIVER_SERVICE_API_SPEC.md) for driver and service-attachment wire-level details.
Use [SYSTEM_OPS_API_SPEC.md](e:\Projects\Charon\SYSTEM_OPS_API_SPEC.md) for technical-admin system-ops wire-level details.

Use this document as the backlog contract for the remaining deferred API surface so those features stay visible during implementation sequencing.
