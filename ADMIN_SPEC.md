# Charon Admin and Operations Specification

## Document Status

- Status: Draft v1
- Date: 2026-03-07
- Purpose: Define the admin, cashier, and technical-operations behavior of the Charon web application at the product and system level.

## 1. Design Intent

The admin application is the operational control surface for Charon.

It must be:

- understandable by average campus operators
- powerful enough for finance and service management
- simple enough to support self-hosted institutes
- separate from the public live view, which remains open to guardians

The admin product is not meant to be a generic enterprise transit ERP. It is a focused campus operations tool.

## 2. Product Shape

- One shared web app for `admin`, `cashier`, and `technical_admin` users.
- Access is controlled through role-based screens and actions, not separate portals.
- Desktop-first design is required.
- Responsive behavior is still required so the app remains usable on tablets and smaller screens.
- English is the initial UI language.
- Multi-language support is deferred to a later version.

## 3. Role Model

### 3.1 Admin

Admins can access:

- operational dashboard
- student management
- wallet oversight
- bus registry
- route management
- schedule management
- service instances
- alerts
- public advisories
- audit logs
- CSV import and export

Admins can:

- edit student operational flags
- create and rotate bus QR assets
- create and edit routes and schedules
- create ad hoc service instances
- approve large refunds
- create cancellations and service advisories
- acknowledge and resolve alerts
- attach notes to audit entries

### 3.2 Cashier

Cashiers are finance-focused and must remain intentionally narrow.

Cashiers can access:

- student search
- finance lookup
- wallet credit issuance
- wallet refund issuance

Cashiers cannot:

- edit student profile flags
- manage buses, routes, schedules, or service instances
- modify alerts or advisories
- access system-operations tooling

Cashier rules:

- credits have no per-day limit in v1
- refunds have a per-day limit
- refunds above the cashier's daily limit require one admin approval

### 3.3 Technical Admin

Technical admins are a restricted subset of admins.

They can additionally access:

- dead-letter queue visibility
- requeue tooling
- queue-health and worker-health operations
- system diagnostics pages

Normal admins must not have routine access to these lower-level operational tools.

## 4. Module Inventory

The web app should be structured around these modules:

- `Dashboard`
- `Students`
- `Wallet Ops`
- `Buses`
- `Routes`
- `Schedules`
- `Service Instances`
- `Alerts`
- `Public Service Feed`
- `Audit Logs`
- `System Ops`

The guardian live view is not an admin-only module. It remains public. The admin app may link to it or preview it, but it must not become a gated feature.

## 5. Dashboard

The dashboard is the main landing page for admins.

Recommended real-time cards:

- active buses
- stale services
- blocked scans
- alert count
- queue backlog
- old QR usage
- failed wallet operations

Admin dashboard expectations:

- prioritize exception visibility over vanity metrics
- highlight problems needing action
- provide drill-down links into the relevant module

Cashier dashboard can be simplified to:

- today's credits
- today's refunds
- pending refund approvals
- student lookup shortcut

## 6. Students Module

### 6.1 Purpose

The Students module exists for operational policy management, not for broad academic records.

### 6.2 Search

Admins and cashiers must be able to search students by:

- institutional ID
- name

### 6.3 Admin Editable Fields

Admins can edit:

- fare exemption flag
- overdraft limit
- account status
- route eligibility
- internal notes

### 6.4 Cashier Limits

Cashiers must not directly edit any student policy or profile fields.

Cashier interaction with students is limited to:

- finding the student
- viewing finance-related information
- issuing a credit or refund within role limits

## 7. Wallet Ops Module

### 7.1 Purpose

Wallet Ops is the finance domain surface inside the admin web app.

### 7.2 Supported Actions

Cashiers and admins can:

- search a student
- view wallet balance
- view transaction history
- issue a credit
- issue a refund

### 7.3 Refund Approval Model

- Cashier refund authority is limited per day.
- Credits have no cashier limit in v1.
- If a refund would exceed the cashier's daily limit, the system must create an approval-required request.
- One admin approval is enough.
- The approval chain must be audit recorded.

### 7.4 Audit Requirements

Every adjustment must record:

- actor
- approval actor if any
- reason code
- before balance
- after balance
- timestamp

### 7.5 Read Boundaries

Cashier finance lookup may include:

- current balance
- recent ledger history
- recent credits and refunds

It should not expose unrelated operational data.

## 8. Buses Module

### 8.1 Managed Fields

Required bus fields in v1:

- bus code
- plate
- route
- status
- seat capacity
- qr version
- notes

### 8.2 Statuses

Supported bus statuses:

- `active`
- `maintenance`
- `retired`
- `out_of_service`

### 8.3 QR Operations

Admins can:

- generate QR for one bus at a time
- rotate QR for one bus at a time
- download QR as PNG

The app does not need bulk QR rotation in v1.

### 8.4 Rotation Visibility

When an older QR version is still used during the one-day grace period, the admin UI must flag it clearly.

## 9. Routes Module

Admins can:

- create routes
- edit route metadata
- assign buses to routes in the v1 one-bus-one-route model
- manage stop lists

Stop ordering can remain form-based in v1. Drag-and-drop is not required.

## 10. Schedules Module

### 10.1 Schedule Model

The schedule workflow is:

- define weekly templates first
- define exceptions second

The app does not need a full calendar-style planner in v1.

### 10.2 Weekly Templates

Admins must be able to configure:

- route
- morning service definition
- evening service definition
- planned start and end times

### 10.3 Exceptions

Admins must be able to define:

- holiday closures
- special-event schedule changes
- route-specific overrides

## 11. Service Instances Module

### 11.1 Purpose

Service Instances represent the actual boardable operational runs derived from schedule or created ad hoc.

### 11.2 Visible States

Supported states:

- `scheduled`
- `boarding_open`
- `running`
- `completed`
- `expired`
- `cancelled`
- `conflicted`

### 11.3 Ad Hoc Service Instances

Admins can create ad hoc service instances with:

- bus
- route
- service label
- start time
- expected end
- notes

### 11.4 Driver Attachment

Drivers are not pre-assigned rigidly in v1.

Recommended attachment model:

- driver authenticates in the driver app
- driver scans the same bus QR used by students, or enters the numeric bus code
- backend uses that scan only as bus selection, not as driver authorization
- backend attaches the authenticated driver to the currently eligible service instance for that bus

Rules:

- the QR does not grant driver privilege by itself
- driver identity comes from login
- bus identity comes from QR or bus code
- if there is no eligible service instance, the driver cannot attach
- if there is a conflict, the driver flow must fail clearly and the service instance must be visible in admin review

### 11.5 Manual Closing

Admins do not need a force-close action in v1 because service windows auto-expire after their allowed grace period.

## 12. Alerts Module

Admins must be able to:

- acknowledge alerts
- resolve alerts
- mute alerts
- add notes
- broadcast an advisory from an alert context

The alerts module is for action and visibility, not only passive viewing.

## 13. Public Service Feed Module

The public live view itself remains openly accessible and is not gated by the admin app.

Admin-facing responsibilities here are:

- publish campus-wide advisories
- publish route-specific advisories
- publish service cancellations
- preview what the public-facing service feed is showing

There is no need for route-level or bus-level visibility toggles in v1. Guardians should always be able to see bus movement for active services.

## 14. Audit Logs Module

### 14.1 Required Filters

Audit log search must support:

- actor
- student
- bus
- route
- action type
- result
- date range

### 14.2 Notes

Admins must be able to attach investigation notes to an audit entry.

The original audit event must remain immutable.

Investigation context must be stored separately through linked notes rather than by editing the audit row itself.

Recommended model:

- immutable `audit_logs` table for original event facts
- separate `audit_investigation_notes` table with a foreign key to `audit_log_id`

This keeps the admin UX flexible while preserving strict audit integrity.

## 15. System Ops Module

This module is restricted to technical admins.

It should include:

- DLQ visibility
- DLQ message detail
- requeue actions
- queue backlog views
- worker-health status
- recent failed background jobs

The purpose is operational support, not everyday admin workflow.

## 16. Imports and Exports

CSV import and export should be included in v1 for:

- buses
- routes
- stops
- schedules

Imports should be designed for bootstrap and correction workflows, not for continuous enterprise data synchronization.

## 17. UX Principles

- desktop-first layout
- responsive enough for tablets and smaller screens
- role-based navigation so users see only relevant modules
- clear confirmation before destructive or money-moving actions
- low setup burden for campus operators
- no unnecessary workflow complexity for common daily tasks

## 18. Out of Scope

- separate cashier portal
- multilingual admin UI in v1
- bulk QR rotation
- drag-and-drop route editing
- public live-view visibility toggles
- manual force-close of service instances as a normal workflow
- rich enterprise analytics beyond core operational screens

## 19. Relationship to Other Specs

This document complements:

- [COMPREHENSIVE_SPEC.md](e:\Projects\Charon\COMPREHENSIVE_SPEC.md)
- [BUS_QR_SPEC.md](e:\Projects\Charon\BUS_QR_SPEC.md)
- [SPRINT_10_WEEKS.md](e:\Projects\Charon\SPRINT_10_WEEKS.md)

If there is a conflict for admin behavior, role boundaries, or operational workflows, this document should be treated as the more specific source.
