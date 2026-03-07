# Charon Student Self-Service API Specification

## Document Status

- Status: Draft v1
- Date: 2026-03-08
- Purpose: Define the wire-level contract for student self-service APIs outside the boarding and wallet critical path.

## 1. Scope

This document covers:

- student profile
- password change
- simple notification settings
- favorite routes and stops
- student alerts list and read state

This document does not redefine:

- login or token handling
- wallet balance and transaction history
- boarding, sponsorship, or emergency voucher flows
- public live-view endpoints

Those remain in:

- [API_SPEC.md](e:\Projects\Charon\API_SPEC.md)

## 2. Shared Rules

- Inherit auth, error envelope, and general HTTP conventions from [API_SPEC.md](e:\Projects\Charon\API_SPEC.md).
- Use unversioned paths.
- Protected endpoints require student authentication.
- `Idempotency-Key` is required on mutating `POST` endpoints in this document.

Student-specific limits:

- maximum 5 favorite routes
- maximum 10 favorite stops

## 3. GET /me/profile

Purpose:

- fetch the student's app-facing profile and policy summary

Access:

- `student`

Success:

- `200 OK`

Response fields:

- `name`
- `student_id`
- `status`
- `fare_exempt`
- `phone_number` nullable
- `overdraft_limit_minor`

## 4. POST /me/password/change

Purpose:

- allow an authenticated student to change their password

Access:

- `student`

Headers:

- `Idempotency-Key`

Request fields:

- `old_password` required
- `new_password` required

Rules:

- the current password must be supplied
- a valid session alone is not enough

Success:

- `200 OK`

Response fields:

- `password_changed`
- `changed_at`

Failure examples:

- `400 INCORRECT_OLD_PASSWORD`
- `400 WEAK_PASSWORD`

## 5. Notification Settings APIs

### 5.1 GET /me/notification-settings

Access:

- `student`

Success:

- `200 OK`

Response fields:

- `alerts_enabled`

### 5.2 POST /me/notification-settings

Access:

- `student`

Headers:

- `Idempotency-Key`

Request fields:

- `alerts_enabled` required

Success:

- `200 OK`

Response fields:

- `alerts_enabled`
- `updated_at`

## 6. Favorite Routes APIs

### 6.1 GET /me/favorite-routes

Access:

- `student`

Success:

- `200 OK`

Response fields:

- `route_ids[]`

Notes:

- return IDs only
- the client hydrates route detail from cached route data or normal route-read endpoints

### 6.2 POST /me/favorite-routes

Access:

- `student`

Headers:

- `Idempotency-Key`

Request fields:

- `route_id` required

Rules:

- maximum 5 favorite routes
- duplicates should be idempotent or rejected cleanly

Success:

- `201 Created`

Response fields:

- `route_ids[]`
- `count`

Failure examples:

- `400 MAX_FAVORITE_ROUTES_REACHED`

### 6.3 DELETE /me/favorite-routes/{route_id}

Access:

- `student`

Success:

- `200 OK`

Response fields:

- `route_ids[]`
- `count`

## 7. Favorite Stops APIs

### 7.1 GET /me/favorite-stops

Access:

- `student`

Success:

- `200 OK`

Response fields:

- `stop_ids[]`

Notes:

- return IDs only

### 7.2 POST /me/favorite-stops

Access:

- `student`

Headers:

- `Idempotency-Key`

Request fields:

- `stop_id` required

Rules:

- maximum 10 favorite stops

Success:

- `201 Created`

Response fields:

- `stop_ids[]`
- `count`

Failure examples:

- `400 MAX_FAVORITE_STOPS_REACHED`

### 7.3 DELETE /me/favorite-stops/{stop_id}

Access:

- `student`

Success:

- `200 OK`

Response fields:

- `stop_ids[]`
- `count`

## 8. Student Alerts APIs

### 8.1 GET /alerts

Purpose:

- power the student alerts screen with lightweight filtering

Access:

- `student`

Query:

- `status` optional
- `route_id` optional
- `limit`
- `offset`

Filter values:

- `status=active`
- `status=expired`

Success:

- `200 OK`

Response fields:

- `items[]`
- `items[].alert_id`
- `items[].type`
- `items[].severity`
- `items[].status`
- `items[].route_id` nullable
- `items[].message`
- `items[].opened_at`
- `items[].closed_at` nullable
- `items[].read_at` nullable
- `limit`
- `offset`
- `total`

### 8.2 POST /alerts/{alert_id}/read

Purpose:

- mark an alert as read for the current student

Access:

- `student`

Headers:

- `Idempotency-Key`

Rules:

- read state is a simple timestamp
- there is no separate dismissed state in v1

Success:

- `200 OK`

Response fields:

- `alert_id`
- `read_at`

## 9. Relationship To Other Specs

This document complements:

- [API_SPEC.md](e:\Projects\Charon\API_SPEC.md)
- [STUDENT_APP_SPEC.md](e:\Projects\Charon\STUDENT_APP_SPEC.md)
- [COMPREHENSIVE_SPEC.md](e:\Projects\Charon\COMPREHENSIVE_SPEC.md)

If there is a conflict between this document and the backlog inventory in [NONCRITICAL_API_SPEC.md](e:\Projects\Charon\NONCRITICAL_API_SPEC.md), this document wins for student self-service endpoints.
