import 'package:flutter/material.dart';

import '../api/models.dart';
import '../theme.dart';

/// A display wrapper over a real [Shift] joined with its resolved [Location]
/// (from GET /locations). The API has NO tour name / party size / colour, so we
/// derive what we can: location name as the title, a stable colour from the
/// location id, and the published/draft status as a badge. We never fabricate
/// party sizes or tour names.
class ShiftView {
  final Shift shift;
  final Location? location;

  ShiftView(this.shift, this.location);

  String get id => shift.id;

  /// Title = location name when known; otherwise a neutral fallback. (No tour
  /// name exists in the API.)
  String get title => location?.name ?? 'Shift';

  String? get address => location?.address;

  Color get color => tourColor(shift.locationId ?? shift.id);

  bool get isPublished => shift.status == 'published';
  bool get isDraft => shift.status == 'draft';

  DateTime get start => shift.startsAt.toLocal();
  DateTime get end => shift.endsAt.toLocal();

  String? get notes => (shift.notes != null && shift.notes!.trim().isNotEmpty) ? shift.notes : null;

  bool get isToday {
    final now = DateTime.now();
    return start.year == now.year && start.month == now.month && start.day == now.day;
  }

  String get timeRange => '${_hm(start)} – ${_hm(end)}';

  /// Short weekday label, e.g. "SAT".
  String get weekdayShort => _weekdays[start.weekday - 1];

  /// Day-of-month, e.g. "21".
  String get dayOfMonth => start.day.toString();

  /// Long day label, e.g. "Sat 21 Jun".
  String get dayLong => '${_weekdaysTitle[start.weekday - 1]} ${start.day} ${_months[start.month - 1]}';

  static String _hm(DateTime t) =>
      '${t.hour.toString().padLeft(2, '0')}:${t.minute.toString().padLeft(2, '0')}';

  static const _weekdays = ['MON', 'TUE', 'WED', 'THU', 'FRI', 'SAT', 'SUN'];
  static const _weekdaysTitle = ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'];
  static const _months = [
    'Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'
  ];
}
