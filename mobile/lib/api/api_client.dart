import 'dart:convert';

import 'package:http/http.dart' as http;

import '../config.dart';
import 'auth_store.dart';
import 'models.dart';

/// Thrown for non-2xx responses. `code`/`message` come from the backend's
/// uniform {code, message} Error shape when present.
class ApiException implements Exception {
  final int statusCode;
  final String code;
  final String message;
  ApiException(this.statusCode, this.code, this.message);
  @override
  String toString() => 'ApiException($statusCode, $code): $message';
}

/// Thin typed HTTP client over the Opero REST API. Only the endpoints the field
/// app needs are implemented (login, my shifts, check-in/out). Attaches the
/// bearer token from AuthStore.
class ApiClient {
  final AuthStore auth;
  final http.Client _http;

  ApiClient(this.auth, {http.Client? client}) : _http = client ?? http.Client();

  Map<String, String> _headers({bool jsonBody = false}) {
    final h = <String, String>{'Accept': 'application/json'};
    if (jsonBody) h['Content-Type'] = 'application/json';
    final t = auth.token;
    if (t != null && t.isNotEmpty) h['Authorization'] = 'Bearer $t';
    return h;
  }

  Never _raise(http.Response r) {
    String code = 'error', message = r.reasonPhrase ?? 'request failed';
    try {
      final body = jsonDecode(r.body) as Map<String, dynamic>;
      code = (body['code'] as String?) ?? code;
      message = (body['message'] as String?) ?? message;
    } catch (_) {
      // non-JSON body; keep defaults
    }
    throw ApiException(r.statusCode, code, message);
  }

  Future<AuthResponse> login({
    required String tenantSlug,
    required String email,
    required String password,
  }) async {
    final r = await _http.post(
      Uri.parse('$apiBaseUrl/auth/login'),
      headers: _headers(jsonBody: true),
      body: jsonEncode({'tenant_slug': tenantSlug, 'email': email, 'password': password}),
    );
    if (r.statusCode != 200) _raise(r);
    return AuthResponse.fromJson(jsonDecode(r.body) as Map<String, dynamic>);
  }

  Future<List<Shift>> myShifts({String? status, DateTime? from, DateTime? to}) async {
    final q = <String, String>{};
    if (status != null) q['status'] = status;
    if (from != null) q['from'] = from.toUtc().toIso8601String();
    if (to != null) q['to'] = to.toUtc().toIso8601String();
    final uri = Uri.parse('$apiBaseUrl/me/shifts').replace(queryParameters: q.isEmpty ? null : q);
    final r = await _http.get(uri, headers: _headers());
    if (r.statusCode != 200) _raise(r);
    final list = jsonDecode(r.body) as List<dynamic>;
    return list.map((e) => Shift.fromJson(e as Map<String, dynamic>)).toList();
  }

  /// POST /attendance/check-in. Idempotent on clientId server-side (200 on
  /// replay, 201 on first insert) — both are success here.
  Future<AttendanceRecord> checkIn({
    required String clientId,
    String? shiftId,
    double? lat,
    double? lng,
    String? photoUrl,
  }) async {
    final r = await _http.post(
      Uri.parse('$apiBaseUrl/attendance/check-in'),
      headers: _headers(jsonBody: true),
      body: jsonEncode({
        'client_id': clientId,
        if (shiftId != null) 'shift_id': shiftId,
        if (lat != null) 'lat': lat,
        if (lng != null) 'lng': lng,
        if (photoUrl != null) 'photo_url': photoUrl,
      }),
    );
    if (r.statusCode != 200 && r.statusCode != 201) _raise(r);
    return AttendanceRecord.fromJson(jsonDecode(r.body) as Map<String, dynamic>);
  }

  /// POST /attendance/check-out against the same clientId used at check-in.
  Future<AttendanceRecord> checkOut({
    required String clientId,
    double? lat,
    double? lng,
    String? photoUrl,
  }) async {
    final r = await _http.post(
      Uri.parse('$apiBaseUrl/attendance/check-out'),
      headers: _headers(jsonBody: true),
      body: jsonEncode({
        'client_id': clientId,
        if (lat != null) 'lat': lat,
        if (lng != null) 'lng': lng,
        if (photoUrl != null) 'photo_url': photoUrl,
      }),
    );
    if (r.statusCode != 200) _raise(r);
    return AttendanceRecord.fromJson(jsonDecode(r.body) as Map<String, dynamic>);
  }
}
