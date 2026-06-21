/// App configuration. The API base URL can be overridden at build/run time:
///   flutter run --dart-define=API_BASE_URL=http://10.0.2.2:8080
///
/// Defaults assume a local backend. Note the host differs per platform:
///   - Android emulator: the host machine is reachable at 10.0.2.2
///   - iOS simulator: localhost works
///   - physical device: use the machine's LAN IP (and add that origin to the
///     backend CORS allowlist — though native apps are not subject to browser
///     CORS, the value matters if you also test from a browser).
library;

const String apiBaseUrl = String.fromEnvironment(
  'API_BASE_URL',
  defaultValue: 'http://localhost:8080',
);
