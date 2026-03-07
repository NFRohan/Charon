# Charon Driver and Service Attachment API Specification

## Document Status

- Status: Draft v1
- Date: 2026-03-08
- Purpose: Define the wire-level contract for driver attachment, driver service control, driver notices, device-health reporting, and driver-specific telemetry behavior.

## 1. Scope

This document covers:

- active-service recovery
- bus and service attachment preview
- bus and service attachment
- driver start and end service actions
- driver notices
- driver device-health reporting
- driver-specific WebSocket telemetry rules

This document does not redefine:

- shared auth and error-envelope conventions
- student critical-path APIs
- public live-view APIs
- admin technical queue operations

Those remain in:

- [API_SPEC.md](e:\Projects\Charon\API_SPEC.md)
- [SYSTEM_OPS_API_SPEC.md](e:\Projects\Charon\SYSTEM_OPS_API_SPEC.md)

## 2. Shared Rules

- Inherit auth, error envelope, and general HTTP conventions from [API_SPEC.md](e:\Projects\Charon\API_SPEC.md).
- Use unversioned paths.
- Protected endpoints require driver authentication.
- `Idempotency-Key` is required on mutating `POST` endpoints in this document.

Driver-specific business rules:

- a driver may have at most one active attached service at a time
- a bus may have at most one active attached driver session at a time
- driver attachment does not expose the identity of a conflicting driver to another driver
- explicit driver detach is not supported in v1
- ending the currently attached route session is the only supported detach path

## 3. GET /driver/active-service

Purpose:

- restore current driver context after app restart
- bootstrap the minimal driver home screen

Access:

- `driver`

Success:

- `200 OK` with the active-service summary

No active service:

- `404 NO_ACTIVE_SERVICE`

Response fields:

- `bus_code`
- `service_label`
- `route_name`
- `telemetry_status`
- `unsynced_buffer_count`
- `status`
- `started_at`

Notes:

- internal IDs are intentionally not exposed here
- this endpoint returns only user-facing operational fields

## 4. POST /driver/bus-attachment/preview

Purpose:

- preview eligible service-instance candidates before attachment

Access:

- `driver`

Reason for `POST`:

- raw QR payloads may be too long or awkward for query-string transport

Headers:

- `Idempotency-Key`

Request fields:

- exactly one of `qr_payload` or `bus_code` is required
- `qr_payload` optional
- `bus_code` optional

Validation:

- provide exactly one bus identity input

No eligible service:

- `400 NO_ELIGIBLE_SERVICE_INSTANCE`

Success:

- `200 OK`

Response fields:

- `exact_match` boolean
- `candidates[]`
- `candidates[].service_instance_id`
- `candidates[].service_label`
- `candidates[].route_name`
- `candidates[].start_time`
- `candidates[].expected_end_time`
- `candidates[].status`
- `candidates[].selection_reason`

Rules:

- response always uses `candidates[]` for consistency
- when there is one exact match, `candidates[]` contains one item and `exact_match=true`
- when multiple candidates exist at a boundary, `candidates[]` contains multiple items and the driver must choose explicitly

## 5. POST /driver/bus-attachments

Purpose:

- attach the authenticated driver to the chosen service instance

Access:

- `driver`

Headers:

- `Idempotency-Key`

Request fields:

- exactly one of `qr_payload` or `bus_code` is required
- `service_instance_id` conditionally required
- `device_id` optional
- `app_version` optional

Rules:

- `service_instance_id` is required when preview returned multiple candidates
- `service_instance_id` may be omitted when there was exactly one eligible service

Conflict cases:

- `409 DRIVER_ALREADY_HAS_ACTIVE_SERVICE`
- `409 DRIVER_ALREADY_ATTACHED`

Error behavior:

- if the driver is already attached elsewhere, return a generic message
- if the bus is already attached to another driver, return a generic message such as `Bus already attached to an active driver session.`
- conflicting driver identity stays hidden from the driver and is visible only through admin audit or ops tools

Success:

- `201 Created`

Response fields:

- `attachment_status`
- `bus_code`
- `service_label`
- `route_name`
- `telemetry_status`
- `unsynced_buffer_count`
- `status`
- `started_at`

Notes:

- the success payload should be enough to bootstrap the driver home screen immediately

## 6. POST /route-sessions/start

Purpose:

- record the operational start marker for the driver's currently attached service

Access:

- `driver`

Headers:

- `Idempotency-Key`

Request fields:

- `device_id` required
- `app_version` required
- `location_permission_state` required

Rules:

- the driver must already be attached to the relevant service
- this endpoint does not implicitly attach
- this endpoint acts only on the driver's current attached service

Success:

- `200 OK`

Response fields:

- `status`
- `bus_code`
- `service_label`
- `route_name`
- `started_at`
- `telemetry_status`

Failure examples:

- `400 DRIVER_NOT_ATTACHED`
- `403 LOCATION_PERMISSION_REQUIRED`

## 7. POST /route-sessions/end

Purpose:

- record the operational end marker for the driver's current attached service

Access:

- `driver`

Headers:

- `Idempotency-Key`

Request fields:

- `device_id` required
- `app_version` required

Rules:

- the driver may end only their own currently attached service
- the request does not accept arbitrary `service_instance_id` in v1

Graceful terminal behavior:

- if the service already auto-expired or was otherwise terminal, return `200 OK`
- the response should indicate that the service was already terminal rather than failing the request

Success response fields:

- `status`
- `ended_at`
- `already_terminal` boolean

## 8. GET /driver/notices

Purpose:

- fetch the latest persisted operational notices for the driver app

Access:

- `driver`

Query:

- no filtering in v1

Rules:

- only the most recent 5 notices are stored or returned
- only persisted advisories and route-level notices belong here
- backend-generated transient system warnings such as battery-optimization risk do not belong here

Success:

- `200 OK`

Response fields:

- `items[]`
- `items[].notice_id`
- `items[].notice_type`
- `items[].message`
- `items[].starts_at`
- `items[].ends_at` nullable

## 9. POST /driver/device-health

Purpose:

- let the driver app report device capability and risk state to the backend for monitoring and future ops visibility

Access:

- `driver`

Headers:

- `Idempotency-Key`

Request fields:

- `device_id` required
- `app_version` required
- `platform` required
- `location_permission_state` required
- `battery_optimization_enabled` required
- `foreground_service_active` required

Success:

- `200 OK`

Response fields:

- `received`
- `reported_at`

## 10. Driver WebSocket Rules

This document refines the driver side of the shared socket described in [API_SPEC.md](e:\Projects\Charon\API_SPEC.md).

### 10.1 Endpoint

- `GET /ws`

### 10.2 Telemetry Message Shape

Each `driver.telemetry` message must include:

- `message_id`
- `route_session_id`
- `bus_code`
- `lat`
- `lng`
- `speed_kph`
- `heading`
- `accuracy_m`
- `recorded_at`
- `is_replayed`

Notes:

- `bus_code` is intentionally included redundantly even when `route_session_id` is present because it improves debugging and operational traceability

### 10.3 Acknowledgement Rules

- every telemetry message must receive an individual ack or negative ack
- the client deletes buffered telemetry only after a positive ack for that exact `message_id`
- batch acks are not supported in v1

Positive ack shape:

- `type`: `telemetry_ack`
- `payload.message_id`
- `payload.accepted_at`
- `payload.status=accepted`

Negative ack shape:

- `type`: `telemetry_nack`
- `payload.message_id`
- `payload.rejected_at`
- `payload.error_code`
- `payload.message`

### 10.4 Replay Age Rules

- the server may accept replayed telemetry up to the archival retention window
- replay older than the accepted maximum must receive `telemetry_nack` with `error_code=TELEMETRY_TOO_OLD`
- the driver app should prune old local telemetry automatically so replay buffers do not grow without bound on the device

### 10.5 Operational Use of Replayed Telemetry

- replayed telemetry that is accepted may still be archived
- replayed telemetry older than the live-display freshness window must not be treated as current live movement

### 10.6 Driver Notices Over Socket

The backend may push persisted service notices over the shared socket using the existing notice message type.

Socket notices follow the same scope rule as `GET /driver/notices`:

- persisted advisories only
- no transient device warnings

## 11. Relationship To Other Specs

This document complements:

- [API_SPEC.md](e:\Projects\Charon\API_SPEC.md)
- [DRIVER_APP_SPEC.md](e:\Projects\Charon\DRIVER_APP_SPEC.md)
- [COMPREHENSIVE_SPEC.md](e:\Projects\Charon\COMPREHENSIVE_SPEC.md)

If there is a conflict between this document and the backlog inventory in [NONCRITICAL_API_SPEC.md](e:\Projects\Charon\NONCRITICAL_API_SPEC.md), this document wins for driver and service-attachment endpoints.
