# Charon API Specification

## Document Status

- Status: Draft v1
- Date: 2026-03-08
- Purpose: Define the first wire-level API contract for Charon's critical flows.

## 1. Scope

This first API draft covers only the critical flows:

- authentication and session management
- wallet balance and transaction history
- emergency voucher issuance
- boarding preview
- boarding submission for all three modes
- shared WebSocket protocol for student and driver apps
- public live-view read APIs

This draft does not yet fully document:

- admin CRUD for buses, routes, stops, and schedules
- full admin alerts APIs
- technical-admin system-ops APIs

Those can be added in companion contracts or a later revision.

The deferred and companion surface now lives across:

- [ADMIN_CASHIER_API_SPEC.md](e:\Projects\Charon\ADMIN_CASHIER_API_SPEC.md)
- [STUDENT_SELF_SERVICE_API_SPEC.md](e:\Projects\Charon\STUDENT_SELF_SERVICE_API_SPEC.md)
- [DRIVER_SERVICE_API_SPEC.md](e:\Projects\Charon\DRIVER_SERVICE_API_SPEC.md)
- [SYSTEM_OPS_API_SPEC.md](e:\Projects\Charon\SYSTEM_OPS_API_SPEC.md)
- [NONCRITICAL_API_SPEC.md](e:\Projects\Charon\NONCRITICAL_API_SPEC.md)

## 2. API Style

- REST-first for request or response workflows.
- WebSockets only for live updates and telemetry.
- No version segment in the path in v1.
- JSON request and response bodies for all documented REST endpoints.
- Money is always represented in integer minor units.
- The client never calculates fares.
- The client sends the raw signed bus QR payload to the backend.
- The backend is the only place allowed to verify QR signatures, resolve service instances, and calculate final fares.

## 3. Base Assumptions

- Base API path examples use root-relative paths such as `/auth/login` and `/boardings`.
- Authenticated endpoints use bearer tokens in the `Authorization` header.
- Timestamps use RFC 3339 in UTC.
- Mobile and web clients should treat all IDs as opaque strings.

## 4. Common Conventions

### 4.1 Headers

Authenticated REST requests:

- `Authorization: Bearer <access_token>`
- `Content-Type: application/json`

Idempotent state-changing requests:

- `Idempotency-Key: <uuid-or-client-generated-unique-key>`

### 4.2 Idempotency Rules

`Idempotency-Key` is required on all non-authentication `POST` endpoints in this draft that mutate state, including:

- `POST /boardings`
- `POST /wallet/emergency-voucher/issue`

Authentication endpoints are exempt in v1:

- `POST /auth/login`
- `POST /auth/refresh`
- `POST /auth/logout`

### 4.3 Error Envelope

All non-success responses should use this shape:

```json
{
  "error_code": "INSUFFICIENT_FUNDS",
  "message": "Wallet balance too low.",
  "trace_id": "req_123abc",
  "field_errors": []
}
```

Rules:

- `error_code` is required.
- `message` is required and safe to show to end users unless otherwise noted.
- `trace_id` is required so operators can correlate logs.
- `field_errors` is optional and may be empty.

Example with field errors:

```json
{
  "error_code": "VALIDATION_ERROR",
  "message": "One or more fields are invalid.",
  "trace_id": "req_456def",
  "field_errors": [
    {
      "field": "stop_id",
      "message": "Stop does not belong to this route."
    }
  ]
}
```

### 4.4 HTTP Status Mapping

- `200 OK` for successful processing.
- `400 Bad Request` for validation failures and business-rule failures such as invalid QR, insufficient funds, duplicate boarding, or trip not active.
- `401 Unauthorized` for missing or invalid access token.
- `403 Forbidden` for valid auth with insufficient role or blocked account state.
- `404 Not Found` for true resource misses on read endpoints.
- `409 Conflict` for idempotency conflicts or concurrent state conflicts.
- `500 Internal Server Error` for unexpected failures.

Important exception:

- `POST /boardings` with `mode=emergency_sync` still returns `200 OK` when the sync was processed successfully even if the outcome is debt restriction instead of a normal wallet debit.
- In that case the request succeeded, a terminal state was recorded, and the response must describe the resulting account state.

### 4.5 Pagination

List endpoints in v1 use limit or offset pagination:

- `?limit=20&offset=0`

Recommended defaults:

- default `limit=20`
- max `limit=100`

### 4.6 Shared Data Types

Common enum values used in this draft:

- `role`: `student`, `driver`, `cashier`, `admin`, `technical_admin`
- `boarding_mode`: `standard`, `sponsored`, `emergency_sync`
- `location_check_result`: `INSIDE_CAMPUS_GEOFENCE`, `OUTSIDE_CAMPUS_GEOFENCE`, `PERMISSION_DENIED`, `LOCATION_UNAVAILABLE`, `LOW_ACCURACY`
- `account_status`: `ACTIVE`, `SUSPENDED`, `RESTRICTED_DEBT`

## 5. Authentication and Sessions

### 5.1 POST /auth/login

Unified login endpoint for all roles.

Request:

```json
{
  "login_id": "220041234",
  "password": "redacted"
}
```

Response:

```json
{
  "access_token": "jwt-access-token",
  "access_token_expires_at": "2026-03-08T08:30:00Z",
  "refresh_token": "opaque-refresh-token",
  "refresh_token_expires_at": "2026-03-15T08:00:00Z",
  "role": "student",
  "user_id": "usr_123",
  "profile_summary": {
    "name": "Student A",
    "status": "ACTIVE",
    "fare_exempt": false
  }
}
```

Notes:

- `login_id` is student ID for students and employee ID for drivers.
- The backend determines role from the credentials.
- Active service information is not returned here.

Failure examples:

- `401 INVALID_CREDENTIALS`
- `403 ACCOUNT_DISABLED`

### 5.2 POST /auth/refresh

Refreshes the access token.

Stable refresh tokens are used in v1 to avoid race conditions on flaky networks.

Request:

```json
{
  "refresh_token": "opaque-refresh-token"
}
```

Response:

```json
{
  "access_token": "new-jwt-access-token",
  "access_token_expires_at": "2026-03-08T09:00:00Z",
  "refresh_token": "opaque-refresh-token",
  "refresh_token_expires_at": "2026-03-15T08:00:00Z"
}
```

Failure examples:

- `401 REFRESH_TOKEN_INVALID`
- `401 REFRESH_TOKEN_EXPIRED`

### 5.3 POST /auth/logout

Logs out the current device session only.

Request:

```json
{
  "refresh_token": "opaque-refresh-token"
}
```

Response:

```json
{
  "logged_out": true
}
```

Notes:

- The backend invalidates only the submitted refresh token.
- Other device sessions remain active.

## 6. Wallet APIs

### 6.1 GET /wallet/balance

Returns the student's current wallet and account summary.

Response:

```json
{
  "user_id": "usr_123",
  "account_status": "ACTIVE",
  "balance_minor": 1800,
  "overdraft_limit_minor": 200,
  "fare_exempt": false,
  "available_emergency_voucher_count": 1
}
```

### 6.2 GET /wallet/transactions

Returns paginated wallet history.

Query parameters:

- `limit`
- `offset`

Example:

`GET /wallet/transactions?limit=20&offset=0`

Response:

```json
{
  "items": [
    {
      "transaction_id": "tx_123",
      "type": "BOARDING_FARE",
      "amount_minor": 2000,
      "route_code": "A",
      "bus_code": "1042",
      "status": "SUCCESS",
      "resulting_balance_minor": 1800,
      "created_at": "2026-03-08T08:05:00Z"
    }
  ],
  "limit": 20,
  "offset": 0,
  "total": 1
}
```

### 6.3 POST /wallet/emergency-voucher/issue

Issues or refreshes the student's bounded emergency voucher inventory for the current device.

This endpoint is intended for hidden app behavior, not a visible button in the UI.

Headers:

- `Authorization`
- `Idempotency-Key`

Request:

```json
{
  "device_id": "dev_iphone_001",
  "platform": "ios"
}
```

Response:

```json
{
  "issued_count": 1,
  "vouchers": [
    {
      "voucher_token": "ev_opaque_token",
      "max_fare_minor": 2000,
      "expires_at": "2026-03-15T08:00:00Z"
    }
  ]
}
```

Notes:

- The server may return zero, one, or more vouchers depending on policy.
- The app stores returned vouchers securely on device.
- Voucher issuance should be blocked for restricted or suspended accounts.

Failure examples:

- `403 ACCOUNT_RESTRICTED_DEBT`
- `403 ACCOUNT_SUSPENDED`

## 7. Boarding APIs

### 7.1 GET /boardings/preview

Returns the server-calculated preview before payment.

The client must call preview before `POST /boardings`.

Supported query inputs:

- `qr_payload=<url-encoded-raw-signed-qr-payload>` or
- `bus_code=<numeric-bus-code>`

Required query inputs:

- `stop_id`

Optional query inputs:

- `mode`
- `sponsored_student_id`

Rules:

- exactly one of `qr_payload` or `bus_code` must be provided
- `mode` defaults to `standard`
- `sponsored_student_id` is only valid when `mode=sponsored`

Example standard preview:

`GET /boardings/preview?qr_payload=BUSQR%3A...&stop_id=stop_12`

Example sponsored preview:

`GET /boardings/preview?bus_code=1042&stop_id=stop_12&mode=sponsored&sponsored_student_id=220041111`

Standard preview response:

```json
{
  "mode": "standard",
  "bus_code": "1042",
  "route_code": "A",
  "route_name": "Route A",
  "destination_label": "Campus Return",
  "service_label": "morning",
  "selected_stop_id": "stop_12",
  "fare_minor_per_rider": 2000,
  "riders_covered": 1,
  "total_amount_minor": 2000,
  "wallet_balance_after_minor": 1800,
  "boarding_window": {
    "opens_at": "2026-03-08T07:30:00Z",
    "closes_at": "2026-03-08T10:15:00Z"
  }
}
```

Sponsored preview response:

```json
{
  "mode": "sponsored",
  "bus_code": "1042",
  "route_code": "A",
  "route_name": "Route A",
  "destination_label": "Campus Return",
  "service_label": "morning",
  "selected_stop_id": "stop_12",
  "fare_minor_per_rider": 2000,
  "riders_covered": 2,
  "total_amount_minor": 4000,
  "wallet_balance_after_minor": 500,
  "sponsored_rider_summary": {
    "student_id_masked": "22004***11",
    "name_masked": "Student B"
  },
  "boarding_window": {
    "opens_at": "2026-03-08T07:30:00Z",
    "closes_at": "2026-03-08T10:15:00Z"
  }
}
```

Failure examples:

- `400 INVALID_QR`
- `400 BUS_NOT_FOUND`
- `400 TRIP_NOT_ACTIVE`
- `400 INVALID_STOP`
- `400 SPONSORED_RIDER_NOT_FOUND`

Important:

- Preview is advisory only.
- `POST /boardings` must revalidate everything.

### 7.2 POST /boardings

Unified boarding submission endpoint.

Headers:

- `Authorization`
- `Idempotency-Key`

The request body always includes:

- `mode`
- exactly one of `qr_payload` or `bus_code`
- `stop_id`
- `location_check_result`
- `location_override_used`

#### 7.2.1 Standard Mode

Request:

```json
{
  "mode": "standard",
  "qr_payload": "BUSQR:raw-signed-payload",
  "stop_id": "stop_12",
  "location_check_result": "INSIDE_CAMPUS_GEOFENCE",
  "location_override_used": false
}
```

Success response:

```json
{
  "boarding_status": "SUCCESS",
  "mode": "standard",
  "transaction_id": "tx_123",
  "boarding_reference_id": "be_123",
  "bus_code": "1042",
  "route_code": "A",
  "service_label": "morning",
  "selected_stop_id": "stop_12",
  "fare_minor": 2000,
  "total_amount_minor": 2000,
  "riders_covered": 1,
  "wallet_balance_after_minor": 1800,
  "created_at": "2026-03-08T08:05:00Z"
}
```

Failure examples:

- `400 INSUFFICIENT_FUNDS`
- `400 TRIP_NOT_ACTIVE`
- `400 ALREADY_BOARDED`
- `400 INVALID_QR`
- `409 IDEMPOTENCY_CONFLICT`

#### 7.2.2 Sponsored Mode

Request:

```json
{
  "mode": "sponsored",
  "bus_code": "1042",
  "stop_id": "stop_12",
  "sponsored_student_id": "220041111",
  "location_check_result": "INSIDE_CAMPUS_GEOFENCE",
  "location_override_used": false
}
```

Rules:

- payer is the authenticated student
- sponsored rider is the `sponsored_student_id`
- v1 allows only one additional rider
- both riders use the same `stop_id`
- the operation is atomic

Success response:

```json
{
  "boarding_status": "SUCCESS",
  "mode": "sponsored",
  "transaction_id": "tx_124",
  "bus_code": "1042",
  "route_code": "A",
  "service_label": "morning",
  "selected_stop_id": "stop_12",
  "total_amount_minor": 4000,
  "riders_covered": 2,
  "wallet_balance_after_minor": 500,
  "created_at": "2026-03-08T08:06:00Z"
}
```

Failure examples:

- `400 INSUFFICIENT_FUNDS`
- `400 SPONSORED_RIDER_NOT_FOUND`
- `400 SPONSORED_RIDER_ALREADY_BOARDED`
- `400 SPONSORED_RIDER_INELIGIBLE`
- `400 TRIP_NOT_ACTIVE`

#### 7.2.3 Emergency Sync Mode

This mode is used when the student already boarded offline using a pre-issued emergency voucher and the app is now syncing that locally consumed voucher.

Request:

```json
{
  "mode": "emergency_sync",
  "qr_payload": "BUSQR:raw-signed-payload",
  "stop_id": "stop_12",
  "emergency_voucher_token": "ev_opaque_token",
  "device_id": "dev_iphone_001",
  "locally_consumed_at": "2026-03-08T08:07:00Z",
  "location_check_result": "LOCATION_UNAVAILABLE",
  "location_override_used": true
}
```

Success response when redeemed normally:

```json
{
  "boarding_status": "SYNC_REDEEMED",
  "mode": "emergency_sync",
  "voucher_status": "REDEEMED",
  "transaction_id": "tx_125",
  "boarding_reference_id": "be_125",
  "bus_code": "1042",
  "route_code": "A",
  "service_label": "morning",
  "selected_stop_id": "stop_12",
  "fare_minor": 2000,
  "wallet_balance_after_minor": 0,
  "account_status": "ACTIVE",
  "created_at": "2026-03-08T08:10:00Z"
}
```

Success response when processed into debt restriction:

```json
{
  "boarding_status": "SYNC_RESTRICTED_DEBT",
  "mode": "emergency_sync",
  "voucher_status": "CONSUMED",
  "transaction_id": "tx_126",
  "boarding_reference_id": "be_126",
  "fare_minor": 2000,
  "wallet_balance_after_minor": -2000,
  "account_status": "RESTRICTED_DEBT",
  "restriction_reason": "EMERGENCY_VOUCHER_DEBT",
  "created_at": "2026-03-08T08:12:00Z"
}
```

Failure examples:

- `400 EMERGENCY_VOUCHER_INVALID`
- `400 EMERGENCY_VOUCHER_EXPIRED`
- `400 EMERGENCY_VOUCHER_ALREADY_REDEEMED`
- `400 TRIP_NOT_ACTIVE`
- `400 ALREADY_BOARDED`

## 8. Shared WebSocket Protocol

### 8.1 Endpoint

Authenticated mobile applications use one shared socket endpoint:

- `GET /ws`

The access token is supplied during handshake using the normal bearer-token mechanism supported by the client platform.

### 8.2 Message Envelope

Client to server and server to client messages use a typed envelope:

```json
{
  "type": "telemetry_update",
  "trace_id": "msg_123",
  "payload": {}
}
```

Rules:

- `type` is required
- `trace_id` is optional but recommended for correlation
- `payload` is required

### 8.3 Student Socket Messages

Client messages:

- `subscribe_route`
- `unsubscribe_route`

Example subscribe:

```json
{
  "type": "subscribe_route",
  "trace_id": "msg_sub_1",
  "payload": {
    "route_code": "A"
  }
}
```

Server messages:

- `telemetry_update`
- `eta_update`
- `alert_broadcast`
- `socket_error`

Example telemetry update:

```json
{
  "type": "telemetry_update",
  "payload": {
    "route_code": "A",
    "bus_code": "1042",
    "lat": 23.9001,
    "lng": 90.3002,
    "recorded_at": "2026-03-08T08:15:00Z"
  }
}
```

Example ETA update:

```json
{
  "type": "eta_update",
  "payload": {
    "route_code": "A",
    "stop_id": "stop_12",
    "eta_at": "2026-03-08T08:22:00Z",
    "is_stale": false
  }
}
```

Example alert:

```json
{
  "type": "alert_broadcast",
  "payload": {
    "alert_type": "SERVICE_DISRUPTION",
    "route_code": "A",
    "message": "Morning service delayed."
  }
}
```

### 8.4 Driver Socket Messages

Driver telemetry is sent over the same socket endpoint using a driver-specific message type.

Client telemetry message:

```json
{
  "type": "driver.telemetry",
  "trace_id": "telemetry_row_001",
  "payload": {
    "message_id": "telemetry_row_001",
    "route_session_id": "rs_123",
    "bus_code": "1042",
    "lat": 23.9001,
    "lng": 90.3002,
    "speed_kph": 28.5,
    "heading": 110,
    "accuracy_m": 8,
    "recorded_at": "2026-03-08T08:15:00Z",
    "is_replayed": false
  }
}
```

Server ack:

```json
{
  "type": "telemetry_ack",
  "payload": {
    "message_id": "telemetry_row_001",
    "accepted_at": "2026-03-08T08:15:01Z",
    "status": "accepted"
  }
}
```

Rules:

- the driver app deletes locally buffered telemetry rows only after `telemetry_ack`
- missing ack means the row stays buffered for retry or replay

Driver notice example:

```json
{
  "type": "service_notice",
  "payload": {
    "notice_type": "SERVICE_ADVISORY",
    "message": "Route A return trip delayed."
  }
}
```

## 9. Public Live View APIs

Public live-view endpoints are anonymous in v1.

They must return pre-shaped public models only.

They must never expose:

- internal bus UUIDs
- driver IDs
- route-session IDs
- finance data
- student-specific data

### 9.1 GET /public/routes/active

Returns the set of currently active public route models.

Response:

```json
{
  "items": [
    {
      "route_code": "A",
      "route_name": "Route A",
      "service_label": "morning",
      "bus_code": "1042",
      "lat": 23.9001,
      "lng": 90.3002,
      "updated_at": "2026-03-08T08:15:00Z",
      "is_stale": false
    }
  ]
}
```

### 9.2 GET /public/routes/{route_code}/live

Returns the route-specific public live snapshot.

Response:

```json
{
  "route_code": "A",
  "route_name": "Route A",
  "service_label": "morning",
  "buses": [
    {
      "bus_code": "1042",
      "lat": 23.9001,
      "lng": 90.3002,
      "updated_at": "2026-03-08T08:15:00Z",
      "is_stale": false
    }
  ],
  "advisories": [
    {
      "advisory_type": "SERVICE_DISRUPTION",
      "message": "Morning service delayed.",
      "starts_at": "2026-03-08T08:00:00Z"
    }
  ]
}
```

### 9.3 GET /public/advisories

Returns active public advisories.

Optional query parameters:

- `route_code`
- `active_only=true`

Response:

```json
{
  "items": [
    {
      "route_code": "A",
      "advisory_type": "SERVICE_CANCELLATION",
      "message": "Evening service cancelled due to campus closure.",
      "starts_at": "2026-03-08T14:00:00Z",
      "ends_at": null
    }
  ]
}
```

## 10. Deferred API Areas

The following are intentionally deferred from this first draft:

- admin CRUD write payloads for buses, routes, stops, schedules, and advisories
- cashier and admin finance adjustment endpoints in full detail
- public WebSocket or SSE protocol
- device-management endpoints for multi-session visibility
- advanced filtering and export endpoints

## 11. Design Notes and Rationale

- Unified `POST /auth/login` keeps auth maintenance simple while still returning role-specific context.
- Stable refresh tokens are deliberate because mobile race conditions on bad networks are more harmful here than the security gain from immediate rotation.
- `GET /boardings/preview` keeps fare calculation server-side and makes the confirmation screen trustworthy.
- Unified `POST /boardings` keeps ledger logic, duplicate protection, and audit paths centralized.
- `mode=emergency_sync` is not a generic offline wallet. It is a bounded reconciliation path for a pre-issued emergency voucher.
- One shared WebSocket per app keeps mobile connection management simpler than multiple parallel sockets.

## 12. Relationship to Other Specs

This document complements:

- [COMPREHENSIVE_SPEC.md](e:\Projects\Charon\COMPREHENSIVE_SPEC.md)
- [BUS_QR_SPEC.md](e:\Projects\Charon\BUS_QR_SPEC.md)
- [STUDENT_APP_SPEC.md](e:\Projects\Charon\STUDENT_APP_SPEC.md)
- [DRIVER_APP_SPEC.md](e:\Projects\Charon\DRIVER_APP_SPEC.md)
- [STUDENT_SELF_SERVICE_API_SPEC.md](e:\Projects\Charon\STUDENT_SELF_SERVICE_API_SPEC.md)
- [DRIVER_SERVICE_API_SPEC.md](e:\Projects\Charon\DRIVER_SERVICE_API_SPEC.md)

If there is a conflict between this API contract and an earlier high-level summary, this document should win for the covered endpoints and message formats.
