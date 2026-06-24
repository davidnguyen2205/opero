import 'package:flutter/foundation.dart';

import '../api/api_client.dart';
import '../api/auth_store.dart';
import 'shift_view.dart';

/// Loads and holds the REAL field data: the signed-in user's shifts joined with
/// the tenant's locations. Shared by the Today, Schedule and Shift-detail
/// screens so they stay consistent and avoid duplicate fetches.
class FieldData extends ChangeNotifier {
  final ApiClient api;
  final AuthStore auth;

  FieldData({required this.api, required this.auth});

  bool _loading = false;
  bool get loading => _loading;

  String? _error;
  String? get error => _error;

  /// True when the last load failed for a reason that signals the session is
  /// gone (401). The shell listens and signs out.
  bool _unauthorized = false;
  bool get unauthorized => _unauthorized;

  List<ShiftView> _shifts = [];
  List<ShiftView> get shifts => _shifts;

  bool _loaded = false;
  bool get loaded => _loaded;

  /// Today's first shift (the one the Today/check-in flow targets), if any.
  ShiftView? get todayShift {
    for (final s in _shifts) {
      if (s.isToday && s.isPublished) return s;
    }
    return null;
  }

  Future<void> load() async {
    _loading = true;
    _error = null;
    _unauthorized = false;
    notifyListeners();
    try {
      final shifts = await api.myShifts(status: 'published');
      // Locations are best-effort: if they fail we still show shifts.
      Map<String, dynamic> locById = {};
      try {
        final locs = await api.locations();
        locById = {for (final l in locs) l.id: l};
      } catch (_) {
        // ignore — shifts can render without resolved location names
      }
      final views = shifts
          .map((s) => ShiftView(s, s.locationId == null ? null : locById[s.locationId]))
          .toList()
        ..sort((a, b) => a.shift.startsAt.compareTo(b.shift.startsAt));
      _shifts = views;
      _loaded = true;
    } on ApiException catch (e) {
      if (e.statusCode == 401) {
        _unauthorized = true;
      } else {
        _error = e.message;
      }
    } catch (_) {
      _error = 'Offline — showing what we have. Pull to refresh.';
    } finally {
      _loading = false;
      notifyListeners();
    }
  }

  ShiftView? byId(String id) {
    for (final s in _shifts) {
      if (s.id == id) return s;
    }
    return null;
  }
}
