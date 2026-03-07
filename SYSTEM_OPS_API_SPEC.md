# Charon System Ops API Specification

## Document Status

- Status: Draft v1
- Date: 2026-03-08
- Purpose: Define the wire-level contract for technical-admin queue inspection, DLQ visibility, DLQ requeue, and worker-health APIs.

## 1. Scope

This document covers:

- queue summaries
- queue detail views
- DLQ message listing
- DLQ message detail
- single-message DLQ requeue
- worker-health visibility

This document is restricted to technical operations and does not define:

- normal admin dashboard APIs
- import and export jobs
- student, driver, or public-facing endpoints

Those remain in:

- [ADMIN_CASHIER_API_SPEC.md](e:\Projects\Charon\ADMIN_CASHIER_API_SPEC.md)
- [API_SPEC.md](e:\Projects\Charon\API_SPEC.md)

## 2. Shared Rules

- Inherit auth, error envelope, and general HTTP conventions from [API_SPEC.md](e:\Projects\Charon\API_SPEC.md).
- Use unversioned paths.
- All endpoints in this document require `technical_admin`.
- `Idempotency-Key` is required on requeue actions.

## 3. GET /admin/system/queues

Purpose:

- provide queue-depth summary across the async stack

Access:

- `technical_admin`

Success:

- `200 OK`

Response fields:

- `queues[]`
- `queues[].queue_name`
- `queues[].ready_count`
- `queues[].inflight_count`
- `queues[].dlq_count`

## 4. GET /admin/system/queues/{queue_name}

Purpose:

- inspect one queue at a higher detail level than the summary dashboard

Access:

- `technical_admin`

Success:

- `200 OK`

Response fields:

- `queue_name`
- `ready_count`
- `inflight_count`
- `dlq_count`
- `consumer_count`
- `last_activity_at` nullable

## 5. GET /admin/system/queues/{queue_name}/dlq

Purpose:

- list dead-lettered messages for one queue

Access:

- `technical_admin`

Query:

- `limit`
- `offset`

Success:

- `200 OK`

Response fields:

- `items[]`
- `items[].message_id`
- `items[].attempt_count`
- `items[].last_error`
- `items[].dead_lettered_at`
- `limit`
- `offset`
- `total`

## 6. GET /admin/system/queues/{queue_name}/dlq/{message_id}

Purpose:

- inspect one dead-lettered message before deciding whether to requeue it

Access:

- `technical_admin`

Success:

- `200 OK`

Response fields:

- `message_id`
- `queue_name`
- `headers`
- `payload`
- `attempt_count`
- `last_error`
- `dead_lettered_at`

## 7. POST /admin/system/queues/{queue_name}/dlq/{message_id}/requeue

Purpose:

- requeue one dead-lettered message

Access:

- `technical_admin`

Headers:

- `Idempotency-Key`

Rules:

- single-message requeue only in v1
- bulk requeue is not supported

Request fields:

- `reason_code` required
- `note` optional

Success:

- `200 OK`

Response fields:

- `message_id`
- `queue_name`
- `requeued`
- `requeued_at`

## 8. GET /admin/system/workers/health

Purpose:

- provide a worker-health summary for technical operations

Access:

- `technical_admin`

Success:

- `200 OK`

Response fields:

- `workers[]`
- `workers[].worker_name`
- `workers[].status`
- `workers[].last_heartbeat_at`
- `workers[].active_jobs`
- `workers[].recent_failures`

## 9. Relationship To Other Specs

This document complements:

- [API_SPEC.md](e:\Projects\Charon\API_SPEC.md)
- [ADMIN_CASHIER_API_SPEC.md](e:\Projects\Charon\ADMIN_CASHIER_API_SPEC.md)
- [COMPREHENSIVE_SPEC.md](e:\Projects\Charon\COMPREHENSIVE_SPEC.md)

If there is a conflict between this document and the backlog inventory in [NONCRITICAL_API_SPEC.md](e:\Projects\Charon\NONCRITICAL_API_SPEC.md), this document wins for technical-admin system-ops endpoints.
