# Opero mobile — shipping guide

Field-staff app (Flutter): sign in → see my shifts → check in/out, offline-tolerant.

> **VERIFICATION STATUS (updated 2026-06-24).** The app now **compiles and
> `flutter analyze` is clean** (Flutter 3.44 / Dart 3.12). It has been built for
> **web** and the login → my-shifts → check-in API path was verified against the
> running backend. It has **NOT** been built or run on a real iOS/Android device
> or emulator (none available in this environment) — geolocator/image_picker and
> the offline airplane-mode flow behave differently on web and are unverified on
> device. Treat device behaviour as untested until run on a simulator/emulator.

## Screen inventory — REAL vs MOCK

The field app screens (ported from the Claude Design prototype) are:

| Screen | Backed by | Notes |
|---|---|---|
| Login (`login_screen.dart`) | **REAL** `POST /auth/login` | Adds a tenant-slug field (multi-tenant); prototype showed only email/pw. |
| Onboarding (`onboarding_screen.dart`) | client-only | 4-slide carousel, shown once (shared_preferences flag). Location toggle calls the real permission prompt; notifications toggle is cosmetic (no push in v1). |
| Today/home (`today_screen.dart`) | **REAL** `GET /me/shifts` | Today's published shift + Check In. |
| Check-in flow (`checkin_flow.dart`) | **REAL** attendance | locate→photo→confirm→active→done via offline-tolerant `AttendanceService`. Photo captured but not uploaded (no blob storage). "Within 25 m" is cosmetic; "Break" is a local-only toggle (no break API). |
| Schedule (`schedule_screen.dart`) | **REAL** `GET /me/shifts` (+ `GET /locations`) | Location name, time window, status. No tour name/party size — the API has none. |
| Shift detail (`shift_detail_screen.dart`) | **REAL** | Same data; meeting point = location address; notes = shift.notes. |
| Inbox + thread (`inbox_screen.dart`, `thread_screen.dart`) | **MOCK** | No messaging API in v1. Badged "Demo"; sent messages are local-only. |
| Notifications (`notifications_screen.dart`) | **MOCK** | No notifications API. Badged "Demo". |
| Profile (`profile_screen.dart`) | identity **REAL** (token), stats **MOCK** | Email/role/tenant from `AuthStore`; on-time/hours/tours/tenure/phone/empId/time-off are demo. Sign out is real. |
| Time-off (`timeoff_screen.dart`) | **MOCK** | No leave API (v1.1). Form does not persist/notify; badged "Demo". |

All mock data lives in `lib/mock/field_mock.dart` (clearly marked demo-only) and
mock screens show a "Demo" badge in the UI. When the messaging / notifications /
leave / analytics APIs exist, delete that file and wire the screens to the client.

> Historical note: this app was originally authored without a Flutter toolchain,
> so the logic (offline queue, client_id lifecycle, request shaping) was written
> most carefully and the plugin-specific code (geolocator, image_picker) flagged
> as most likely to need reconciliation. It has since been analyzed clean.

## 1. Prerequisites

Install Flutter (stable) + platform tooling (Xcode for iOS, Android Studio/SDK
for Android). Verify with `flutter doctor`.

## 2. Generate the platform scaffolding

`flutter create` was not run here (only `lib/` + `pubspec.yaml` exist). Generate
the `android/`, `ios/`, etc. folders **without clobbering** the hand-written
files:

```bash
cd mobile
flutter create --platforms=android,ios --project-name opero_mobile .
```

`flutter create .` on an existing directory adds platform folders and leaves
`lib/` alone, but it may rewrite `pubspec.yaml`. If it does, restore the
dependency list from this repo's `pubspec.yaml` (git diff) and re-run
`flutter pub get`.

## 3. Resolve dependencies and analyze (expect to fix things here)

```bash
flutter pub get          # if a version constraint fails: flutter pub upgrade --major-versions
flutter analyze          # FIX whatever this reports — see "Likely fix points" below
```

## 4. Platform permissions (required for geolocation + camera)

These are NOT in the generated defaults — add them or location/photo capture
will crash or silently no-op.

**iOS** — `ios/Runner/Info.plist`:
```xml
<key>NSLocationWhenInUseUsageDescription</key>
<string>Opero records your location when you check in or out of a shift.</string>
<key>NSCameraUsageDescription</key>
<string>Opero attaches a photo when you check in or out of a shift.</string>
```

**Android** — `android/app/src/main/AndroidManifest.xml` (inside `<manifest>`):
```xml
<uses-permission android:name="android.permission.ACCESS_FINE_LOCATION" />
<uses-permission android:name="android.permission.ACCESS_COARSE_LOCATION" />
<uses-permission android:name="android.permission.INTERNET" />
<uses-permission android:name="android.permission.CAMERA" />
```
(`image_picker`/`geolocator` may require a `minSdkVersion` bump in
`android/app/build.gradle` — follow the analyzer/build errors.)

## 5. Point at the backend

The base URL defaults to `http://localhost:8080`, but device networking differs:

```bash
# Android emulator (host machine is 10.0.2.2):
flutter run --dart-define=API_BASE_URL=http://10.0.2.2:8080
# iOS simulator (localhost works):
flutter run --dart-define=API_BASE_URL=http://localhost:8080
# physical device: use your machine's LAN IP, e.g. http://192.168.1.20:8080
```

Run the backend first: `cd ../backend && make up && make run` (control-plane
must be migrated; see backend README). Native apps are not subject to browser
CORS, so the CORS allowlist does not gate the device — but it does matter if you
also test from a browser.

## 6. Validate the core flow

1. Sign in with a tenant slug + an employee login. To get one: in the web app
   (or via the API) create an employee, then `POST /employees/{id}/login` to mint
   their credentials.
2. The shifts list calls `GET /me/shifts?status=published`. Publish a shift for
   that employee (web app or `POST /shifts` + `POST /shifts/{id}/publish`).
3. **Offline test (the point of this app):** enable airplane mode, tap Check in
   — it should queue (the app bar shows "N queued"); re-enable network and pull
   to refresh — the queue should flush and the manager live view (`GET /live`)
   should show the employee `checked_in`. Because the server is idempotent on
   `client_id`, replays never duplicate.

## 7. The generated-client deviation (guardrail note)

`CLAUDE.md` §2/§6 say clients consume a **generated** client from
`api/openapi.yaml`. This app instead uses a **hand-written thin client**
(`lib/api/`) — a deliberate, documented exception because the mobile surface is
tiny (login, my shifts, check-in/out) and a Dart generator could not be run/
verified here. If you prefer a generated client, `openapi-generator` (`dart-dio`)
or `swagger_dart_code_generator` can target `../api/openapi.yaml`; swap
`lib/api/api_client.dart` + `models.dart` for the generated package and keep the
offline/attendance layer as-is.

## 8. Known gaps / deferred (by design or by constraint)

- **Photo upload is not wired to storage.** v1 has no blob storage and the API
  takes a `photo_url` string. The app captures a photo (proving the capability)
  but does **not** send it as `photo_url` (it would be a local path, not a URL).
  When storage exists: upload the file from `Capture.photoPath`, send the
  returned URL. See `lib/attendance/capture.dart`.
- **Token storage uses shared_preferences, not secure storage.** Use
  `flutter_secure_storage` (Keychain/Keystore) for production. See
  `lib/api/auth_store.dart`.
- **Offline queue uses shared_preferences**, not a transactional DB. Fine for a
  handful of queued actions; consider sqflite/drift at scale. See
  `lib/offline/queue.dart`.
- **No automatic reconnect trigger.** Sync runs on app open, on each
  check-in/out, and on pull-to-refresh. Add `connectivity_plus` to auto-sync on
  network regain if desired.
- **No tests.** Add widget/unit tests once the app compiles; the offline
  queue + `AttendanceService` sync logic is the highest-value thing to test
  (mirrors how the backend tests its idempotency).

## 9. Likely fix points (where unverified code most often breaks)

- `geolocator` API: `getCurrentPosition()` and the `LocationPermission` enum
  have shifted across majors; reconcile in `lib/attendance/capture.dart`.
- `image_picker`: `pickImage(...)` return type/params — verify against installed
  version.
- `uuid`: `const Uuid().v4()` — verify the v4 call style for the installed major.
- `http`: response handling is standard, but confirm `jsonDecode` shapes match
  the live API (they mirror `api/openapi.yaml`).
