import 'package:shared_preferences/shared_preferences.dart';

/// Holds the bearer token in memory and persists it so the session survives
/// app restarts.
///
/// SECURITY CAVEAT: shared_preferences is NOT secure storage. For production,
/// store the token with `flutter_secure_storage` (Keychain / Keystore). This
/// is documented in SHIPPING.md as a known gap.
class AuthStore {
  static const _tokenKey = 'opero.auth.token';

  String? _token;
  String? get token => _token;
  bool get isAuthenticated => _token != null && _token!.isNotEmpty;

  Future<void> load() async {
    final prefs = await SharedPreferences.getInstance();
    _token = prefs.getString(_tokenKey);
  }

  Future<void> setToken(String token) async {
    _token = token;
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString(_tokenKey, token);
  }

  Future<void> clear() async {
    _token = null;
    final prefs = await SharedPreferences.getInstance();
    await prefs.remove(_tokenKey);
  }
}
