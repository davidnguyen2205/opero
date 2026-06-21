/// DTOs mirroring api/openapi.yaml (snake_case JSON, ISO-8601 timestamps, UUID
/// strings). Hand-written — see SHIPPING.md for the deliberate, documented
/// deviation from the "generated client" guardrail and how to swap to a
/// generated client if desired. Keep these in sync with the spec.
library;

class UserSummary {
  final String id;
  final String email;
  final String role;
  final String status;

  UserSummary({required this.id, required this.email, required this.role, required this.status});

  factory UserSummary.fromJson(Map<String, dynamic> j) => UserSummary(
        id: j['id'] as String,
        email: j['email'] as String,
        role: j['role'] as String,
        status: j['status'] as String,
      );
}

class TenantSummary {
  final String id;
  final String name;
  final String slug;
  final String status;
  final String plan;

  TenantSummary({
    required this.id,
    required this.name,
    required this.slug,
    required this.status,
    required this.plan,
  });

  factory TenantSummary.fromJson(Map<String, dynamic> j) => TenantSummary(
        id: j['id'] as String,
        name: j['name'] as String,
        slug: j['slug'] as String,
        status: j['status'] as String,
        plan: j['plan'] as String,
      );
}

class AuthResponse {
  final String token;
  final String tokenType;
  final DateTime expiresAt;
  final UserSummary user;
  final TenantSummary tenant;

  AuthResponse({
    required this.token,
    required this.tokenType,
    required this.expiresAt,
    required this.user,
    required this.tenant,
  });

  factory AuthResponse.fromJson(Map<String, dynamic> j) => AuthResponse(
        token: j['token'] as String,
        tokenType: j['token_type'] as String,
        expiresAt: DateTime.parse(j['expires_at'] as String),
        user: UserSummary.fromJson(j['user'] as Map<String, dynamic>),
        tenant: TenantSummary.fromJson(j['tenant'] as Map<String, dynamic>),
      );
}

class Shift {
  final String id;
  final String employeeId;
  final String? locationId;
  final DateTime startsAt;
  final DateTime endsAt;
  final String? notes;
  final String status; // draft | published

  Shift({
    required this.id,
    required this.employeeId,
    required this.locationId,
    required this.startsAt,
    required this.endsAt,
    required this.notes,
    required this.status,
  });

  factory Shift.fromJson(Map<String, dynamic> j) => Shift(
        id: j['id'] as String,
        employeeId: j['employee_id'] as String,
        locationId: j['location_id'] as String?,
        startsAt: DateTime.parse(j['starts_at'] as String),
        endsAt: DateTime.parse(j['ends_at'] as String),
        notes: j['notes'] as String?,
        status: j['status'] as String,
      );
}

class AttendanceRecord {
  final String id;
  final String employeeId;
  final String? shiftId;
  final String clientId;
  final DateTime? checkInAt;
  final DateTime? checkOutAt;
  final String status; // checked_in | checked_out | missed

  AttendanceRecord({
    required this.id,
    required this.employeeId,
    required this.shiftId,
    required this.clientId,
    required this.checkInAt,
    required this.checkOutAt,
    required this.status,
  });

  factory AttendanceRecord.fromJson(Map<String, dynamic> j) => AttendanceRecord(
        id: j['id'] as String,
        employeeId: j['employee_id'] as String,
        shiftId: j['shift_id'] as String?,
        clientId: j['client_id'] as String,
        checkInAt: _parseTime(j['check_in_at']),
        checkOutAt: _parseTime(j['check_out_at']),
        status: j['status'] as String,
      );
}

DateTime? _parseTime(dynamic v) => v == null ? null : DateTime.parse(v as String);
