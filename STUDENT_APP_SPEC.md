# Charon Student App Specification

## Document Status

- Status: Draft v1
- Date: 2026-03-08
- Purpose: Define the product and system behavior of the Charon student mobile application.

## 1. Design Intent

The student app is the main daily user surface of Charon.

It must be:

- fast for repeated daily use
- simple enough to keep boarding friction low
- reliable under weak mobile network conditions
- clear about money movement and wallet state
- useful even when parts of the system are briefly offline

The student app exists to:

- let students pay boarding fares safely
- give students a bounded fallback when they temporarily lose mobile internet
- show wallet state and transaction history
- show live buses and route progress
- provide stop-specific ETA
- surface rider alerts and service advisories

## 2. Platform and Device Assumptions

- Flutter remains the implementation framework.
- Android and iOS are both operational targets in v1 for the student app.
- English is the initial application language.
- Bangla support is deferred to a later version.

Because students may use multiple devices:

- one student account may stay signed in on more than one phone
- the backend must tolerate multiple active device sessions for the same student
- mobile UX should assume the user may switch devices without losing account access

## 3. Authentication and Session Model

- Students authenticate with `student ID + password`.
- Forgotten-password reset is admin-assisted in v1.
- Logged-in sessions should persist until the student logs out or the session is revoked.
- A logged-in student may use the app on multiple phones at the same time.

## 4. Product Shape

The app should be home-first rather than wallet-first.

Top-level navigation in v1:

- `Home`
- `Wallet`
- `Map`
- `Alerts`
- `Profile`

Important UX rules:

- there is no separate `Pay` tab
- `Scan to Pay` should be available as a prominent action from `Home`
- `Scan to Pay` should also remain available from `Wallet`
- the guardian live view remains a separate public surface and is not embedded in the student app

## 5. Core Screens

### 5.1 Home Screen

The home screen is the default first screen after login.

It should provide a compact daily summary with:

- current wallet balance
- primary `Scan to Pay` action
- quick access to favorite routes and stops
- current service advisories or major rider alerts
- shortcut into the live map

The home screen should help the student reach the next likely action quickly rather than acting as a dense dashboard.

### 5.2 Wallet Screen

The wallet screen must show:

- current balance
- overdraft usage if applicable
- fare-exempt badge if applicable
- recent transactions
- large `Scan to Pay` action
- simple `Top up at cashier` guidance

If the student has a negative balance because of overdraft, the balance should be shown plainly rather than softened or hidden.

The app should not present a fake online top-up flow in v1.

### 5.3 Map Screen

The map screen must show:

- all active buses
- route-level movement on cached MapTiler-backed tiles
- manual stop selection for ETA lookup
- favorite routes and stops for faster repeat access

The map does not need to embed the guardian/public live view because that surface is intentionally separate.

### 5.4 Alerts Screen

The alerts screen should provide:

- active alerts
- recent alert history
- route filter
- read and unread state

Notification controls can remain simple in v1.

### 5.5 Profile Screen

The profile or settings screen should provide:

- student account information
- password change
- notification settings
- privacy note
- app version
- logout

## 6. Boarding and Fare Flow

### 6.1 Entry Points

Students can enter the boarding flow from:

- the primary `Scan to Pay` action on `Home`
- the `Scan to Pay` action on `Wallet`

Manual numeric bus-code fallback should remain available as part of the same flow.

The full boarding fallback stack in v1 is:

- direct self-pay
- sponsored boarding
- emergency ride permit

### 6.2 Stop Selection

Each boarding attempt should include student stop selection.

Why this is required:

- it supports stop-specific fare calculation where a deployment needs it
- it still works for flat-fare routes because the fare engine can ignore the stop and use the route default
- it gives the system the right context for rider-specific ETA

Stop-selection rules:

- the student selects the relevant stop before final fare confirmation
- favorite stops should be surfaced first when possible
- manual selection remains the authoritative input in v1

### 6.3 Confirmation

After QR scan or manual bus-code entry, the app must show a confirmation screen before charging.

Required confirmation fields:

- bus code
- route name
- service label
- fare
- boarding stop
- service window
- expected wallet balance after charge
- location warning state if any

If the student is sponsoring another rider, the confirmation screen must additionally show:

- rider count
- total fare
- masked additional-rider identity

### 6.4 Sponsored Boarding

Sponsored boarding exists for the case where another student has mobile data and is willing to pay.

Rules:

- sponsored boarding is initiated by the connected payer
- the payer may add at most one additional rider in v1
- the payer enters the additional rider's student ID
- the app shows masked confirmation of the additional rider before payment
- the same selected stop is used for all riders in the sponsored request in v1
- the whole sponsored request is all-or-nothing

The student experience must make it explicit that the payer is covering another student's ride.

### 6.5 Emergency Ride Permit

Emergency Ride is the last fallback when the student has no internet and no sponsor available.

Rules:

- emergency ride permits are pre-issued while the app is online
- permits are tied to the student and device
- permits are one-time use and bounded to a single ride or max single fare
- permits are stored securely on device
- permit use must be clearly labeled as an emergency fallback, not normal payment
- after local use, the app must redeem the permit with the backend when connectivity returns

The app should not behave like a fully offline wallet. Emergency Ride is a tightly bounded exception path.

### 6.6 Location Safety UX

The app should perform the local campus-geofence safety check defined in the bus QR spec.

Required behavior:

- location-check calculation stays on device
- the privacy explanation can stay mostly hidden unless the student opens help or privacy details
- if the student appears outside the campus geofence, show a bold warning and require extra confirmation
- if location permission is denied, ask again on each boarding attempt
- permission denial or weak location should not hard block boarding in v1

### 6.7 Network and Retry Behavior

If the boarding request fails because of weak connectivity:

- the app should not retry silently
- the app should show a clear failure state
- the app should offer `Try Again`
- if the connected student is paying for someone else, the app may offer `Sponsored Boarding`
- if the student has a valid local emergency ride permit, the app may offer `Emergency Ride`

If the backend returns a successful idempotent replay of an earlier attempt, the UI should present it as a normal success.

### 6.8 Success State

After a successful boarding charge, the app should show:

- success confirmation
- amount paid
- route and bus
- timestamp
- updated balance
- receipt or transaction identifier

If the bus is valid but the service is not open yet, the app should show `Trip Not Active Yet` with a friendlier hint that boarding opens 30 minutes before departure.

## 7. Wallet and Transaction History

The student transaction list should show:

- amount
- type
- route
- bus code
- time
- status
- resulting balance

If the student believes a charge is wrong, the app does not need a full dispute workflow in v1.

It should instead show a small support note directing the student to contact admin or finance support.

## 8. Map, ETA, and Favorites

### 8.1 Live Map Scope

- all active buses should be visible to students
- bus movement should update in real time when telemetry is fresh
- cached tiles should keep the map usable through short network loss

### 8.2 ETA Model

ETA in the student app is stop-specific.

Rules:

- the student manually selects the stop
- the selected stop can come from favorites for speed
- ETA is personalized only to that stop selection, not to any guardian-facing public view

### 8.3 Favorites

Students should be able to favorite:

- routes
- stops

Favorites exist to speed up repeat boarding and ETA lookup rather than to introduce heavy personalization logic.

## 9. Offline and Cached Behavior

The student app should retain useful read-only behavior when briefly offline.

Expected offline support:

- cached wallet balance
- cached recent transactions
- cached MapTiler tiles
- cached last-known routes and stops

Out of scope for offline mode:

- generic offline fare payment or offline wallet sync
- new wallet adjustments
- live telemetry freshness guarantees

Bounded exception:

- locally stored emergency ride permits may be used and redeemed later

## 10. Privacy and Support Boundaries

- raw student GPS coordinates should not be uploaded for boarding validation in v1
- the public guardian live view must remain separate from the student account experience
- the app should not imply child-level tracking or sharing features
- the app should direct finance disputes to human admin or cashier support rather than faking an automated dispute system

## 11. Relationship to Other Specs

This document complements:

- [COMPREHENSIVE_SPEC.md](e:\Projects\Charon\COMPREHENSIVE_SPEC.md)
- [BUS_QR_SPEC.md](e:\Projects\Charon\BUS_QR_SPEC.md)
- [STUDENT_SELF_SERVICE_API_SPEC.md](e:\Projects\Charon\STUDENT_SELF_SERVICE_API_SPEC.md)
- [SPRINT_20_WEEKS.md](e:\Projects\Charon\SPRINT_20_WEEKS.md)

If there is a conflict for student-app UX or behavior, this document should be treated as the more specific source.
