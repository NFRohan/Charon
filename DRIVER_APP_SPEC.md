# Charon Driver App Specification

## Document Status

- Status: Draft v1
- Date: 2026-03-07
- Purpose: Define the product and system behavior of the Charon driver mobile application.

## 1. Design Intent

The driver app is a minimal operational tool, not a feature-rich navigation product.

It must be:

- simple enough for drivers with limited technical familiarity
- reliable on personal Android devices
- optimized for large buttons and low cognitive load
- able to continue telemetry in the background
- able to survive weak network conditions and aggressive Android battery behavior

The driver app exists to:

- attach a driver to a bus and service instance
- mark service start and end operationally
- publish bus telemetry
- receive simple operational notices

## 2. Platform and Device Assumptions

- Flutter remains the implementation framework.
- Android is the only supported runtime in v1.
- The codebase should still preserve future iOS compatibility where practical, but iOS is not an operational target in v1.
- The app runs on personal phones, not institute-issued hardware.

Because the app runs on personal phones:

- battery optimization behavior matters
- background execution reliability matters
- recovery after app kill matters

## 3. Authentication

- Drivers authenticate with `employee ID + password`.
- Driver login is separate from student credential assumptions.
- A logged-in driver session should be restorable where appropriate so mid-service app restarts are recoverable.

## 4. UX Shape

The app should stay extremely minimal.

Primary first screen after login:

- `Attach Bus`
- current service status
- GPS status
- primary action button for `Start Journey`

The product should avoid:

- dense text
- complex route-management controls
- embedded driver map
- stop-list-heavy flows

## 5. Core Screens

### 5.1 Attach Bus Screen

Required behavior:

- scan bus QR is the primary path
- numeric bus-code entry is the fallback
- app shows confirmation before attaching

The driver uses the same durable bus QR as students, but only as a bus selector after driver authentication.

### 5.2 Active Service Screen

After successful attachment, the main screen should show:

- bus attached
- service label
- GPS ok
- network ok
- telemetry sending
- buffer backlog
- battery saver risk
- boarding count

This should remain the main operational surface while the service is running.

### 5.3 Notices Surface

Drivers should receive simple inbound notices for:

- route cancellation
- service advisory

No richer messaging system is needed in v1.

## 6. Bus Attachment Flow

### 6.1 Attachment Inputs

The driver can attach to a bus by:

- scanning bus QR
- entering the numeric bus code

### 6.2 Confirmation

When exactly one eligible service instance is found, the app should still show a confirmation screen before attachment.

### 6.3 Multiple Eligible Service Instances

If more than one eligible service instance is available for the bus around a boundary time:

- the driver chooses manually
- the app keeps the choice simple and explicit
- the system must not silently guess

### 6.4 Ad Hoc Trips

Drivers cannot create ad hoc trips in v1.

Ad hoc service instances are admin-created only.

### 6.5 Attachment Restrictions

- A driver must end the current service before attaching to another bus.
- The system should allow only one active driver attachment for the same bus or service instance.
- If another driver is already attached, the second driver must be blocked clearly.

## 7. Service Control

### 7.1 Start Journey

`Start Journey` is an operational marker only.

It does not create boarding eligibility by itself. Boarding truth still comes from the schedule-backed service window.

### 7.2 End Service

`End Service` marks the trip completed operationally.

It does not replace schedule and grace-period rules as the source of boarding validity.

### 7.3 Offline Start Support

Offline start must be supported.

Recommended behavior:

- if the relevant service instance and bus data are already cached locally, the driver may attach and start offline
- the app records the operational actions locally
- once network returns, the app syncs the attachment and operational markers

If the app lacks enough cached data to identify the service instance safely, offline attach should fail.

## 8. Telemetry

### 8.1 Start Condition

Telemetry begins immediately after bus attachment.

It does not wait for `Start Journey`.

### 8.2 Interval

- Telemetry uses a fixed `10 second` interval in v1.
- The goal is to reduce battery drain while preserving acceptable fleet visibility for a small fixed-route deployment.

### 8.3 Background Behavior

- Telemetry must continue while the app is backgrounded.
- Telemetry must continue while the screen is locked.
- Android foreground-service behavior is required.
- The app should use a persistent notification while location streaming is active.

### 8.4 Offline Buffering

- The app must buffer at least `30 minutes` of telemetry locally.
- Replay should happen silently in the background when connectivity returns.
- The driver does not need to manage replay manually.

### 8.5 Replay Expectations

- replay preserves original timestamps
- replay should not interrupt the active UI flow
- replay backlog should be visible through the `buffer backlog` indicator

## 9. Permissions and Device Health

### 9.1 Location

Location permission is mandatory for driver operation.

If location permission is missing or disabled:

- service attachment must be hard blocked
- the app must show a clear telemetry-disabled state

### 9.2 Camera

If camera permission is denied:

- numeric bus-code entry is the fallback
- no extra recovery path is required in v1

### 9.3 Battery Optimization

The app should detect likely battery-optimization risk and warn the driver.

The warning should be surfaced through the `battery saver risk` indicator and a one-time explanatory prompt when needed.

## 10. Recovery and State Restoration

If the app is killed and reopened mid-service:

- it should restore the active bus and service state from local storage
- it should restore telemetry state
- it should continue buffering or replay as needed

The app should feel resilient rather than forcing the driver to reconstruct context manually.

## 11. Operational Data Shown to Drivers

Drivers should see:

- boarding count

Drivers should not need to see:

- route map
- stop list
- detailed passenger identities
- complex planning screens

Boarding count exists only as a practical cross-check for the driver. It is not the authoritative source of fare correctness.

## 12. Error Handling

The app must clearly handle:

- invalid bus QR
- unknown bus code
- no eligible service instance
- multiple eligible service instances
- another driver already attached
- missing location permission
- offline mode with insufficient cached service data

Error text should remain simple and operationally useful.

## 13. Relationship to Other Specs

This document complements:

- [COMPREHENSIVE_SPEC.md](e:\Projects\Charon\COMPREHENSIVE_SPEC.md)
- [BUS_QR_SPEC.md](e:\Projects\Charon\BUS_QR_SPEC.md)
- [ADMIN_SPEC.md](e:\Projects\Charon\ADMIN_SPEC.md)
- [DRIVER_SERVICE_API_SPEC.md](e:\Projects\Charon\DRIVER_SERVICE_API_SPEC.md)
- [SPRINT_20_WEEKS.md](e:\Projects\Charon\SPRINT_20_WEEKS.md)

If there is a conflict for driver-app behavior, this document should be treated as the more specific source.
