import 'dart:async';
import 'dart:convert';

import 'package:flutter/foundation.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:uuid/uuid.dart';

import '../api/api_client.dart';
import '../offline/pending_action.dart';
import '../offline/queue.dart';

/// Orchestrates offline-tolerant attendance.
///
/// Lifecycle: checking in generates a single `client_id` (the idempotency key),
/// persisted as the "active" attendance for the shift so the matching check-out
/// reuses it. Actions are enqueued durably first, then a sync loop replays them
/// FIFO. Because the server is idempotent on `client_id`, replays are safe.
class AttendanceService extends ChangeNotifier {
  static const _activeKey = 'opero.attendance.active'; // shiftKey -> clientId

  final ApiClient api;
  final OfflineQueue queue;
  final _uuid = const Uuid();

  int _pendingCount = 0;
  int get pendingCount => _pendingCount;

  bool _syncing = false;

  AttendanceService({required this.api, required this.queue});

  /// Sentinel key for an unscheduled check-in (no shift).
  static const _unscheduled = '__unscheduled__';
  String _shiftKey(String? shiftId) => shiftId ?? _unscheduled;

  Future<Map<String, String>> _activeMap() async {
    final prefs = await SharedPreferences.getInstance();
    final raw = prefs.getString(_activeKey);
    if (raw == null || raw.isEmpty) return {};
    return (jsonDecode(raw) as Map<String, dynamic>).map((k, v) => MapEntry(k, v as String));
  }

  Future<void> _saveActive(Map<String, String> m) async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString(_activeKey, jsonEncode(m));
  }

  /// The clientId of the open (checked-in, not yet checked-out) attendance for a
  /// shift, if any. Lets the UI show "check out" vs "check in".
  Future<String?> activeClientId(String? shiftId) async => (await _activeMap())[_shiftKey(shiftId)];

  /// Begin a check-in. Generates the client_id, records it as active, enqueues
  /// the action, and kicks a sync. Returns the client_id.
  Future<String> checkIn({String? shiftId, double? lat, double? lng, String? photoUrl}) async {
    final clientId = _uuid.v4();
    final active = await _activeMap();
    active[_shiftKey(shiftId)] = clientId;
    await _saveActive(active);

    await queue.add(PendingAction(
      id: _uuid.v4(),
      type: ActionType.checkIn,
      clientId: clientId,
      shiftId: shiftId,
      lat: lat,
      lng: lng,
      photoUrl: photoUrl,
      createdAt: DateTime.now().toUtc().toIso8601String(),
    ));
    await _refreshPending();
    unawaited(sync());
    return clientId;
  }

  /// End an open attendance for a shift, reusing its client_id.
  Future<void> checkOut({String? shiftId, double? lat, double? lng, String? photoUrl}) async {
    final active = await _activeMap();
    final key = _shiftKey(shiftId);
    final clientId = active[key];
    if (clientId == null) {
      throw StateError('no active check-in to check out for this shift');
    }
    await queue.add(PendingAction(
      id: _uuid.v4(),
      type: ActionType.checkOut,
      clientId: clientId,
      shiftId: shiftId,
      lat: lat,
      lng: lng,
      photoUrl: photoUrl,
      createdAt: DateTime.now().toUtc().toIso8601String(),
    ));
    // Clear the active marker; the queued check-out is durable from here.
    active.remove(key);
    await _saveActive(active);
    await _refreshPending();
    unawaited(sync());
  }

  Future<void> _refreshPending() async {
    _pendingCount = await queue.length();
    notifyListeners();
  }

  /// Replay queued actions FIFO. Stops on the first retryable failure (offline /
  /// 5xx) to preserve ordering; drops poison messages (terminal 4xx) so the
  /// queue can make progress. Safe to call repeatedly (idempotent server).
  Future<void> sync() async {
    if (_syncing) return;
    _syncing = true;
    try {
      final actions = await queue.all(); // FIFO
      for (final a in actions) {
        try {
          if (a.type == ActionType.checkIn) {
            await api.checkIn(
                clientId: a.clientId, shiftId: a.shiftId, lat: a.lat, lng: a.lng, photoUrl: a.photoUrl);
          } else {
            await api.checkOut(clientId: a.clientId, lat: a.lat, lng: a.lng, photoUrl: a.photoUrl);
          }
          await queue.remove(a.id);
        } on ApiException catch (e) {
          if (e.statusCode >= 400 && e.statusCode < 500) {
            // Terminal (validation/conflict/not-found). Drop the poison message
            // and keep going — it will never succeed on retry.
            await queue.remove(a.id);
            continue;
          }
          // Retryable (5xx): preserve order, retry the whole queue later.
          break;
        } catch (_) {
          // Network/offline: preserve order, retry later.
          break;
        }
      }
    } finally {
      _syncing = false;
      await _refreshPending();
    }
  }

  Future<void> init() async => _refreshPending();
}
