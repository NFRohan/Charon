# Charon Admin and Cashier API Specification

## Document Status

- Status: Draft v1
- Date: 2026-03-08
- Purpose: Define the wire-level contract for the shared admin and cashier web application.

## 1. Scope

This document covers:

- admin dashboard widgets
- cashier and admin student lookup
- student policy updates
- wallet lookup, credits, refunds, and refund approval flow
- bus registry and QR operations
- routes, stops, fare policy, and stop ordering
- schedules, calendars, exceptions, and service instances
- alerts, advisories, and audit logs
- CSV import and export jobs

This document does not cover:

- auth and shared API conventions in full detail
- student critical-path endpoints
- driver critical-path endpoints
- public live-view endpoints
- technical-admin queue and DLQ APIs

Those remain in:

- [API_SPEC.md](e:\Projects\Charon\API_SPEC.md)
- [NONCRITICAL_API_SPEC.md](e:\Projects\Charon\NONCRITICAL_API_SPEC.md)

## 2. Shared Rules

- Inherit auth, error envelope, pagination, and basic conventions from [API_SPEC.md](e:\Projects\Charon\API_SPEC.md).
- `Idempotency-Key` is required on every mutating `POST` endpoint in this document.
- `cashier` can search students, view wallet finance data, issue credits, and initiate refunds.
- `admin` can access the full surface in this document.
- `technical_admin` inherits admin access, but technical system-ops endpoints stay out of scope here.

Enums used here:

- `account_status`: `ACTIVE`, `SUSPENDED`, `RESTRICTED_DEBT`
- `bus_status`: `active`, `maintenance`, `retired`, `out_of_service`
- `route_status`: `active`, `inactive`
- `fare_policy_type`: `FLAT_ROUTE`, `STOP_MATRIX`, `ZERO_FARE`
- `refund_request_status`: `PENDING_APPROVAL`, `APPROVED`, `REJECTED`, `COMPLETED`
- `service_instance_status`: `scheduled`, `boarding_open`, `running`, `completed`, `expired`, `cancelled`, `conflicted`, `force_closed`
- `service_exception_type`: `CANCELLATION`, `TIME_OVERRIDE`
- `import_job_status`: `QUEUED`, `VALIDATING`, `IMPORTING`, `COMPLETED`, `FAILED`
- `export_job_status`: `QUEUED`, `EXPORTING`, `COMPLETED`, `FAILED`

## 3. Dashboard APIs

Dashboard data is split per widget so one slow query does not block the page.

- `GET /admin/dashboard/active-buses`
  Access: `admin`, `technical_admin`
  Query: `limit`, `offset`
  Returns: `bus_id`, `bus_code`, `route_id`, `route_name`, `service_instance_id`, `service_label`, `status`, `last_telemetry_at`, `is_stale`

- `GET /admin/dashboard/stale-services`
  Access: `admin`, `technical_admin`
  Query: `limit`, `offset`
  Returns: `service_instance_id`, `bus_code`, `route_name`, `service_label`, `last_telemetry_at`, `minutes_stale`, `status`

- `GET /admin/dashboard/blocked-scans`
  Access: `admin`
  Query: `limit`, `offset`, `window_hours`
  Returns: `scan_attempted_at`, `bus_code`, `student_id_masked`, `validation_result`, `service_window_result`, `boarding_mode`

- `GET /admin/dashboard/alert-counts`
  Access: `admin`
  Returns: `open_total`, `high_severity_total`, `medium_severity_total`, `low_severity_total`, `unresolved_total`

- `GET /admin/dashboard/queue-health`
  Access: `admin`, `technical_admin`
  Returns: `queues[]` with `queue_name`, `ready_count`, `inflight_count`, `dlq_count`
  Note: summary only, not detailed queue inspection

- `GET /admin/dashboard/old-qr-usage`
  Access: `admin`
  Query: `limit`, `offset`
  Returns: `bus_id`, `bus_code`, `current_qr_version`, `old_qr_version`, `grace_period_expires_at`, `usage_count`

- `GET /admin/dashboard/failed-wallet-operations`
  Access: `admin`
  Query: `limit`, `offset`
  Returns: `attempted_at`, `student_id`, `operation_type`, `amount_minor`, `error_code`, `actor_id`

- `GET /admin/dashboard/cashier-summary`
  Access: `cashier`, `admin`
  Returns: `today_credits_count`, `today_credits_amount_minor`, `today_refunds_count`, `today_refunds_amount_minor`, `pending_refund_approvals_count`

## 4. Student Lookup and Policy APIs

- `GET /admin/students/search`
  Access: `cashier`, `admin`
  Query: `q` required, `limit`, `offset`
  Returns: `id`, `name`, `institutional_id`, `status`, `available_balance_minor`, `fare_exempt`

- `GET /admin/students/{student_id}`
  Access: `admin`
  Returns: `id`, `name`, `institutional_id`, `phone_number`, `status`, `fare_exempt`, `overdraft_limit_minor`, `route_eligibility`, `internal_notes`, `available_balance_minor`, `account_restrictions[]`, `updated_at`

- `PATCH /admin/students/{student_id}`
  Access: `admin`
  Patch fields: `status`, `fare_exempt`, `overdraft_limit_minor`, `route_eligibility`, `internal_notes`
  Rule: `reason_code` is mandatory when `status` changes
  Returns: `student_id`, `updated_fields[]`, `status`, `fare_exempt`, `overdraft_limit_minor`, `route_eligibility`, `internal_notes`, `updated_at`

## 5. Wallet Ops APIs

- `GET /admin/wallets/{student_id}`
  Access: `cashier`, `admin`
  Returns: `student_id`, `account_status`, `available_balance_minor`, `overdraft_limit_minor`, `fare_exempt`, `account_restrictions[]`, `recent_adjustment_actors[]`, `updated_at`

- `GET /admin/wallets/{student_id}/transactions`
  Access: `cashier`, `admin`
  Query: `limit`, `offset`, `type`, `status`
  Returns: `transaction_id`, `type`, `amount_minor`, `status`, `actor_id`, `approval_status`, `route_code`, `bus_code`, `created_at`

- `POST /admin/wallets/{student_id}/credits`
  Access: `cashier`, `admin`
  Body: `amount_minor`, `reason_code`, `note`
  Success: `201 Created`
  Returns: `transaction_id`, `student_id`, `adjustment_type`, `amount_minor`, `before_balance_minor`, `after_balance_minor`, `actor_id`, `approval_status`, `created_at`

- `POST /admin/wallets/{student_id}/refunds`
  Access: `cashier`, `admin`
  Body: `amount_minor`, `reason_code`, `note`
  Outcomes:
  `201 Created` when within authority
  `202 Accepted` with `status=APPROVAL_REQUIRED` and `request_id` when above cashier limit

- `GET /admin/refund-requests`
  Access: `admin`
  Query: `status`, `limit`, `offset`
  Returns: `request_id`, `student_id`, `amount_minor`, `reason_code`, `requested_by`, `requested_at`, `status`

- `GET /admin/refund-requests/{request_id}`
  Access: `admin`
  Returns: `request_id`, `student_id`, `amount_minor`, `reason_code`, `note`, `requested_by`, `requested_at`, `status`, `approval_actor_id`, `approved_or_rejected_at`

- `POST /admin/refund-requests/{request_id}/approve`
  Access: `admin`
  Body: `reason_code` required, `note` optional
  Success: `201 Created`
  Returns: `request_id`, `status`, `approval_actor_id`, `transaction_id`, `approved_at`

- `POST /admin/refund-requests/{request_id}/reject`
  Access: `admin`
  Body: `reason_code` required, `note` optional
  Success: `200 OK`
  Returns: `request_id`, `status`, `rejection_actor_id`, `rejected_at`

## 6. Bus Registry and QR APIs

- `GET /admin/buses`
  Access: `admin`
  Query: `status`, `route_id`, `limit`, `offset`
  Returns: `bus_id`, `bus_code`, `plate`, `route_id`, `status`, `seat_capacity`, `qr_version`

- `POST /admin/buses`
  Access: `admin`
  Body: `bus_code`, `plate`, `route_id`, `status`, `seat_capacity`, `notes`
  Rules: `bus_code` unique, `seat_capacity > 0`
  Success: `201 Created`
  Returns: `bus_id`, `bus_code`, `plate`, `route_id`, `status`, `seat_capacity`, `qr_version`, `notes`

- `GET /admin/buses/{bus_id}`
  Access: `admin`
  Returns: `bus_id`, `bus_code`, `plate`, `route_id`, `status`, `seat_capacity`, `qr_version`, `notes`, `created_at`, `updated_at`

- `PATCH /admin/buses/{bus_id}`
  Access: `admin`
  Mutable: `plate`, `route_id`, `status`, `seat_capacity`, `notes`
  Immutable: `bus_code`
  Returns: `bus_id`, `bus_code`, `plate`, `route_id`, `status`, `seat_capacity`, `notes`, `updated_at`

- `POST /admin/buses/{bus_id}/qr/generate`
  Access: `admin`
  Body: `reason_code` optional
  Success: `201 Created`
  Returns: `bus_id`, `bus_code`, `qr_version`, `asset_url`, `generated_at`

- `POST /admin/buses/{bus_id}/qr/rotate`
  Access: `admin`
  Body: `reason_code` required, `note` optional
  Success: `201 Created`
  Returns: `bus_id`, `bus_code`, `new_qr_version`, `asset_url`, `grace_period_expires_at`, `rotated_at`

## 7. Route, Stop, and Fare Policy APIs

- `GET /admin/routes`
  Access: `admin`
  Query: `status`, `limit`, `offset`
  Returns: `route_id`, `code`, `name`, `status`, `fare_policy_type`, `default_fare_minor`

- `POST /admin/routes`
  Access: `admin`
  Body: `code`, `name`, `status`, `fare_policy_type`, `default_fare_minor`
  Rules:
  `FLAT_ROUTE` requires `default_fare_minor > 0`
  `ZERO_FARE` requires `default_fare_minor = 0`
  `STOP_MATRIX` may use `default_fare_minor = null`
  Success: `201 Created`
  Returns: `route_id`, `code`, `name`, `status`, `fare_policy_type`, `default_fare_minor`

- `GET /admin/routes/{route_id}`
  Access: `admin`
  Returns: `route_id`, `code`, `name`, `status`, `fare_policy_type`, `default_fare_minor`, `assigned_bus_ids[]`, `created_at`, `updated_at`

- `PATCH /admin/routes/{route_id}`
  Access: `admin`
  Patch fields: `name`, `status`, `fare_policy_type`, `default_fare_minor`
  Same fare-policy validation rules as create
  Returns: `route_id`, `code`, `name`, `status`, `fare_policy_type`, `default_fare_minor`, `updated_at`

- `GET /admin/stops`
  Access: `admin`
  Query: `limit`, `offset`
  Returns: `stop_id`, `name`, `lat`, `lng`, `public_label`

- `POST /admin/stops`
  Access: `admin`
  Body: `name`, `lat`, `lng`, `public_label`
  Success: `201 Created`
  Returns: `stop_id`, `name`, `lat`, `lng`, `public_label`

- `GET /admin/stops/{stop_id}`
  Access: `admin`
  Returns: `stop_id`, `name`, `lat`, `lng`, `public_label`, `created_at`, `updated_at`

- `PATCH /admin/stops/{stop_id}`
  Access: `admin`
  Patch fields: `name`, `lat`, `lng`, `public_label`
  Returns: `stop_id`, `name`, `lat`, `lng`, `public_label`, `updated_at`

- `PUT /admin/routes/{route_id}/stops`
  Access: `admin`
  Body: `stops[]` with `{stop_id, order}`
  Rules: orders unique and contiguous, all stops must exist
  Success: `200 OK`
  Returns: `route_id`, `stops[]`, `updated_at`

- `GET /admin/routes/{route_id}/fare-rules`
  Access: `admin`
  Returns: `route_id`, `fare_policy_type`, `default_fare_minor`, `rules[]`
  `rules[]` fields: `fare_rule_id`, `stop_id`, `stop_name`, `fare_minor`, `service_label`

- `PUT /admin/routes/{route_id}/fare-rules`
  Access: `admin`
  Body: `fare_policy_type`, `default_fare_minor`, `rules[]`
  `rules[]` fields: `stop_id`, `fare_minor`, `service_label`
  Rules:
  `FLAT_ROUTE` requires empty `rules[]`
  `ZERO_FARE` requires empty `rules[]`
  `STOP_MATRIX` requires non-empty `rules[]`
  no duplicate `(stop_id, service_label)` rows
  Success: `200 OK`
  Returns: `route_id`, `fare_policy_type`, `default_fare_minor`, `rules[]`, `updated_at`

## 8. Schedule and Service Instance APIs

- `GET /admin/trip-templates`
  Access: `admin`
  Query: `route_id`, `service_label`, `status`, `limit`, `offset`
  Returns: `trip_template_id`, `route_id`, `service_label`, `name`, `status`, `service_calendar_id`

- `POST /admin/trip-templates`
  Access: `admin`
  Body: `route_id`, `service_label`, `name`, `status`, `service_calendar_id`
  Success: `201 Created`
  Returns: `trip_template_id`, `route_id`, `service_label`, `name`, `status`, `service_calendar_id`

- `GET /admin/trip-templates/{trip_template_id}`
  Access: `admin`
  Returns: `trip_template_id`, `route_id`, `service_label`, `name`, `status`, `service_calendar_id`, `created_at`, `updated_at`

- `PATCH /admin/trip-templates/{trip_template_id}`
  Access: `admin`
  Mutable: `name`, `status`, `service_calendar_id`
  Returns: `trip_template_id`, `name`, `status`, `service_calendar_id`, `updated_at`

- `GET /admin/trip-templates/{trip_template_id}/stop-times`
  Access: `admin`
  Returns: `trip_template_id`, `stop_times[]`
  `stop_times[]` fields: `stop_id`, `offset_minutes`

- `PUT /admin/trip-templates/{trip_template_id}/stop-times`
  Access: `admin`
  Body: `stop_times[]` with `{stop_id, offset_minutes}`
  Rules: offsets are relative minutes, non-negative, and monotonically increasing
  Success: `200 OK`
  Returns: `trip_template_id`, `stop_times[]`, `updated_at`

- `GET /admin/service-calendars`
  Access: `admin`
  Query: `route_id`, `limit`, `offset`
  Returns: `calendar_id`, `route_id`, `weekday_mask`, `effective_from`, `effective_to`

- `POST /admin/service-calendars`
  Access: `admin`
  Body: `route_id`, `weekday_mask`, `effective_from`, `effective_to`
  Success: `201 Created`
  Returns: `calendar_id`, `route_id`, `weekday_mask`, `effective_from`, `effective_to`

- `PATCH /admin/service-calendars/{calendar_id}`
  Access: `admin`
  Mutable: `weekday_mask`, `effective_from`, `effective_to`
  Returns: `calendar_id`, `weekday_mask`, `effective_from`, `effective_to`, `updated_at`

- `GET /admin/service-exceptions`
  Access: `admin`
  Query: `service_calendar_id`, `service_date`, `limit`, `offset`
  Returns: `exception_id`, `service_calendar_id`, `service_date`, `exception_type`, `reason_code`, `override_start_time`, `override_end_time`

- `POST /admin/service-exceptions`
  Access: `admin`
  Body: `service_calendar_id`, `service_date`, `exception_type`, `reason_code`, `override_start_time`, `override_end_time`
  Rules:
  `CANCELLATION` must not include override times
  `TIME_OVERRIDE` requires both override times
  Success: `201 Created`
  Returns: `exception_id`, `service_calendar_id`, `service_date`, `exception_type`, `reason_code`, `override_start_time`, `override_end_time`

- `PATCH /admin/service-exceptions/{exception_id}`
  Access: `admin`
  Mutable: `reason_code`, `override_start_time`, `override_end_time`
  Returns: `exception_id`, `exception_type`, `reason_code`, `override_start_time`, `override_end_time`, `updated_at`

- `GET /admin/service-instances`
  Access: `admin`
  Query: `status`, `route_id`, `bus_id`, `limit`, `offset`
  Returns: `service_instance_id`, `bus_id`, `bus_code`, `route_id`, `route_name`, `service_label`, `start_time`, `expected_end_time`, `status`

- `POST /admin/service-instances`
  Access: `admin`
  Body: `bus_id`, `route_id`, `service_label`, `start_time`, `expected_end_time`, `notes`
  Success: `201 Created`
  Returns: `service_instance_id`, `bus_id`, `route_id`, `service_label`, `start_time`, `expected_end_time`, `status`, `notes`

- `GET /admin/service-instances/{service_instance_id}`
  Access: `admin`
  Returns: `service_instance_id`, `bus_id`, `bus_code`, `route_id`, `route_name`, `service_label`, `start_time`, `expected_end_time`, `status`, `notes`, `driver_id`, `created_at`, `updated_at`

- `PATCH /admin/service-instances/{service_instance_id}`
  Access: `admin`
  Mutable: `expected_end_time`, `status`
  Returns: `service_instance_id`, `expected_end_time`, `status`, `updated_at`

- `POST /admin/service-instances/{service_instance_id}/cancel`
  Access: `admin`
  Body: `reason_code` required, `note` optional
  Success: `200 OK`
  Returns: `service_instance_id`, `status`, `cancelled_at`, `advisory_draft_suggested=true`

- `POST /admin/service-instances/{service_instance_id}/force-close`
  Access: `admin`
  Body: `reason_code` required, `note` optional
  Success: `200 OK`
  Returns: `service_instance_id`, `status`, `force_closed_at`

## 9. Alerts and Advisories APIs

- `GET /admin/alerts`
  Access: `admin`
  Query: `status`, `severity`, `route_id`, `limit`, `offset`
  Returns: `alert_id`, `type`, `severity`, `status`, `target_type`, `target_id`, `opened_at`, `resolved_at`, `investigation_notes_count`, `last_actor_id`

- `GET /admin/alerts/{alert_id}`
  Access: `admin`
  Returns: `alert_id`, `type`, `severity`, `status`, `target_type`, `target_id`, `opened_at`, `resolved_at`, `last_actor_id`, `investigation_notes_count`, `context`

- `POST /admin/alerts/{alert_id}/acknowledge`
  Access: `admin`
  Body: `note` optional
  Success: `200 OK`
  Returns: `alert_id`, `status`, `last_actor_id`, `updated_at`

- `POST /admin/alerts/{alert_id}/resolve`
  Access: `admin`
  Body: `note` required
  Success: `200 OK`
  Returns: `alert_id`, `status`, `resolved_at`, `last_actor_id`

- `POST /admin/alerts/{alert_id}/mute`
  Access: `admin`
  Body: `note` required, `mute_until` required
  Success: `200 OK`
  Returns: `alert_id`, `status`, `mute_until`, `last_actor_id`

- `GET /admin/advisories`
  Access: `admin`
  Query: `route_id`, `active_only`, `limit`, `offset`
  Returns: `advisory_id`, `route_id`, `advisory_type`, `message`, `starts_at`, `ends_at`

- `POST /admin/advisories`
  Access: `admin`
  Body: `route_id` nullable, `advisory_type`, `message`, `starts_at`, `ends_at`
  Success: `201 Created`
  Returns: `advisory_id`, `route_id`, `advisory_type`, `message`, `starts_at`, `ends_at`

- `PATCH /admin/advisories/{advisory_id}`
  Access: `admin`
  Patch fields: `route_id`, `advisory_type`, `message`, `starts_at`, `ends_at`
  Returns: `advisory_id`, `route_id`, `advisory_type`, `message`, `starts_at`, `ends_at`, `updated_at`

## 10. Audit Log APIs

Audit rows are immutable. Investigation notes are append-only.

- `GET /admin/audit-logs`
  Access: `admin`
  Query: `actor_id`, `student_id`, `bus_id`, `route_id`, `action_type`, `result`, `date_from`, `date_to`, `limit`, `offset`
  Returns: `audit_log_id`, `timestamp`, `actor`, `action_type`, `target_id`, `status`

- `GET /admin/audit-logs/{audit_log_id}`
  Access: `admin`
  Returns: `audit_log_id`, `timestamp`, `actor`, `action_type`, `target_id`, `status`, `payload_json`

- `GET /admin/audit-logs/{audit_log_id}/notes`
  Access: `admin`
  Returns: `note_id`, `author_id`, `note_body`, `created_at`

- `POST /admin/audit-logs/{audit_log_id}/notes`
  Access: `admin`
  Body: `note_body`
  Success: `201 Created`
  Returns: `note_id`, `audit_log_id`, `author_id`, `note_body`, `created_at`

## 11. Import and Export APIs

Imports use sync validation and async processing. Exports always use async jobs.

- `POST /admin/imports/{resource}`
  Access: `admin`
  Path: `resource` in `buses`, `routes`, `stops`, `schedules`
  Content type: `multipart/form-data`
  Body: `file`
  Success: `202 Accepted`
  Returns: `job_id`, `resource`, `status`
  Validation failure: `400 INVALID_IMPORT_FILE`

- `GET /admin/imports/{job_id}`
  Access: `admin`
  Returns: `job_id`, `resource`, `status`, `created_at`, `started_at`, `completed_at`, `error_summary`, `processed_rows`, `failed_rows`

- `POST /admin/exports/{resource}`
  Access: `admin`
  Path: `resource` in `buses`, `routes`, `stops`, `schedules`, `wallet-transactions`, `audit-logs`
  Body: `filters` optional
  Success: `202 Accepted`
  Returns: `job_id`, `resource`, `status`

- `GET /admin/exports/{job_id}`
  Access: `admin`
  Returns: `job_id`, `resource`, `status`, `created_at`, `completed_at`, `download_url`, `expires_at`

## 12. Deferred From This Contract

Still deferred after this admin and cashier contract:

- technical-admin queue inspection and DLQ requeue APIs
- public WebSocket or SSE feeds
- device-management APIs
- bulk QR operations

Those remain tracked in [NONCRITICAL_API_SPEC.md](e:\Projects\Charon\NONCRITICAL_API_SPEC.md).

## 13. Relationship To Other Specs

This document complements:

- [API_SPEC.md](e:\Projects\Charon\API_SPEC.md)
- [ADMIN_SPEC.md](e:\Projects\Charon\ADMIN_SPEC.md)
- [COMPREHENSIVE_SPEC.md](e:\Projects\Charon\COMPREHENSIVE_SPEC.md)

If there is a conflict between this document and the backlog inventory in [NONCRITICAL_API_SPEC.md](e:\Projects\Charon\NONCRITICAL_API_SPEC.md), this document wins for admin and cashier endpoints.
