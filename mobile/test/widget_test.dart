// Minimal placeholder test. The generated boilerplate referenced a non-existent
// `MyApp`; the real app (`OperoApp`) needs AuthStore/ApiClient/AttendanceService
// wiring and plugin mocks, so meaningful widget tests are deferred (see
// SHIPPING.md §8). This asserts the build-time config default so `flutter test`
// stays green and the file compiles.
import 'package:flutter_test/flutter_test.dart';
import 'package:opero_mobile/config.dart';

void main() {
  test('apiBaseUrl falls back to the local backend default', () {
    expect(apiBaseUrl, isNotEmpty);
    expect(apiBaseUrl.startsWith('http'), isTrue);
  });
}
