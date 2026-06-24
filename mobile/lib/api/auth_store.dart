import 'package:shared_preferences/shared_preferences.dart';

import 'models.dart';

/// Holds the bearer token (and the signed-in user's identity) in memory and
/// persists it so the session survives app restarts.
///
/// SECURITY CAVEAT: shared_preferences is NOT secure storage. For production,
/// store the token with `flutter_secure_storage` (Keychain / Keystore). This
/// is documented in SHIPPING.md as a known gap.
class AuthStore {
  static const _tokenKey = 'opero.auth.token';
  static const _emailKey = 'opero.auth.email';
  static const _roleKey = 'opero.auth.role';
  static const _tenantNameKey = 'opero.auth.tenant_name';

  String? _token;
  String? _email;
  String? _role;
  String? _tenantName;

  String? get token => _token;

  /// Signed-in user's email (from the auth response). Used for Profile identity.
  String? get email => _email;

  /// Signed-in user's role (admin | manager | employee).
  String? get role => _role;

  /// The tenant's display name (e.g. "Saigon Tours Co.").
  String? get tenantName => _tenantName;

  bool get isAuthenticated => _token != null && _token!.isNotEmpty;

  Future<void> load() async {
    final prefs = await SharedPreferences.getInstance();
    _token = prefs.getString(_tokenKey);
    _email = prefs.getString(_emailKey);
    _role = prefs.getString(_roleKey);
    _tenantName = prefs.getString(_tenantNameKey);
  }

  /// Persist the full auth result: token + identity fields used by Profile.
  Future<void> setSession(AuthResponse res) async {
    _token = res.token;
    _email = res.user.email;
    _role = res.user.role;
    _tenantName = res.tenant.name;
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString(_tokenKey, res.token);
    await prefs.setString(_emailKey, res.user.email);
    await prefs.setString(_roleKey, res.user.role);
    await prefs.setString(_tenantNameKey, res.tenant.name);
  }

  Future<void> clear() async {
    _token = null;
    _email = null;
    _role = null;
    _tenantName = null;
    final prefs = await SharedPreferences.getInstance();
    await prefs.remove(_tokenKey);
    await prefs.remove(_emailKey);
    await prefs.remove(_roleKey);
    await prefs.remove(_tenantNameKey);
  }
}
