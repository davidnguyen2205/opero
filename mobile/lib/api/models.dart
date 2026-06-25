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

class Location {
  final String id;
  final String name;
  final String? address;
  final double? lat;
  final double? lng;

  Location({
    required this.id,
    required this.name,
    required this.address,
    required this.lat,
    required this.lng,
  });

  factory Location.fromJson(Map<String, dynamic> j) => Location(
        id: j['id'] as String,
        name: j['name'] as String,
        address: j['address'] as String?,
        lat: (j['lat'] as num?)?.toDouble(),
        lng: (j['lng'] as num?)?.toDouble(),
      );
}

DateTime? _parseTime(dynamic v) => v == null ? null : DateTime.parse(v as String);

/// GET /me/stats — computed activity aggregate for the profile screen.
class MyStats {
  final int shiftsThisMonth;
  final double hoursThisWeek;
  final int onTimePct;
  final int? tenureDays;

  MyStats({
    required this.shiftsThisMonth,
    required this.hoursThisWeek,
    required this.onTimePct,
    required this.tenureDays,
  });

  factory MyStats.fromJson(Map<String, dynamic> j) => MyStats(
        shiftsThisMonth: j['shifts_this_month'] as int,
        hoursThisWeek: (j['hours_this_week'] as num).toDouble(),
        onTimePct: j['on_time_pct'] as int,
        tenureDays: j['tenure_days'] as int?,
      );
}

/// A leave/time-off request (LeaveRequest in the spec). `type` is holiday|sick|
/// personal; `status` is pending|approved|rejected. Dates are ISO date strings.
class LeaveRequest {
  final String id;
  final String employeeId;
  final String type;
  final String startDate; // YYYY-MM-DD
  final String endDate; // YYYY-MM-DD
  final String? note;
  final String status;
  final String? reviewedBy;
  final DateTime? reviewedAt;
  final DateTime createdAt;
  final DateTime updatedAt;

  LeaveRequest({
    required this.id,
    required this.employeeId,
    required this.type,
    required this.startDate,
    required this.endDate,
    required this.note,
    required this.status,
    required this.reviewedBy,
    required this.reviewedAt,
    required this.createdAt,
    required this.updatedAt,
  });

  factory LeaveRequest.fromJson(Map<String, dynamic> j) => LeaveRequest(
        id: j['id'] as String,
        employeeId: j['employee_id'] as String,
        type: j['type'] as String,
        startDate: j['start_date'] as String,
        endDate: j['end_date'] as String,
        note: j['note'] as String?,
        status: j['status'] as String,
        reviewedBy: j['reviewed_by'] as String?,
        reviewedAt: _parseTime(j['reviewed_at']),
        createdAt: DateTime.parse(j['created_at'] as String),
        updatedAt: DateTime.parse(j['updated_at'] as String),
      );
}

/// GET /me/leave/balance — leave entitlement for the current year.
class LeaveBalance {
  final int year;
  final int entitledDays;
  final int usedDays;
  final int remainingDays;

  LeaveBalance({
    required this.year,
    required this.entitledDays,
    required this.usedDays,
    required this.remainingDays,
  });

  factory LeaveBalance.fromJson(Map<String, dynamic> j) => LeaveBalance(
        year: j['year'] as int,
        entitledDays: j['entitled_days'] as int,
        usedDays: j['used_days'] as int,
        remainingDays: j['remaining_days'] as int,
      );
}

/// POST /media — the URL of an uploaded file (e.g. a check-in photo).
class MediaUpload {
  final String url;
  MediaUpload({required this.url});
  factory MediaUpload.fromJson(Map<String, dynamic> j) => MediaUpload(url: j['url'] as String);
}
