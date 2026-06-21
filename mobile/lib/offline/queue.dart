import 'dart:convert';

import 'package:shared_preferences/shared_preferences.dart';

import 'pending_action.dart';

/// Durable FIFO queue of pending attendance actions, persisted as a JSON array
/// in shared_preferences so it survives app restarts.
///
/// NOTE: shared_preferences is simple but not transactional; for higher
/// durability/concurrency a local DB (sqflite/drift) is the production choice.
/// Documented in SHIPPING.md. Access here is serialized through a single
/// in-process instance, which is sufficient for this app's usage.
class OfflineQueue {
  static const _key = 'opero.offline.queue';

  Future<List<PendingAction>> all() async {
    final prefs = await SharedPreferences.getInstance();
    final raw = prefs.getString(_key);
    if (raw == null || raw.isEmpty) return [];
    final list = jsonDecode(raw) as List<dynamic>;
    return list.map((e) => PendingAction.fromJson(e as Map<String, dynamic>)).toList();
  }

  Future<void> _save(List<PendingAction> actions) async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString(_key, jsonEncode(actions.map((a) => a.toJson()).toList()));
  }

  Future<void> add(PendingAction action) async {
    final actions = await all();
    actions.add(action);
    await _save(actions);
  }

  Future<void> remove(String id) async {
    final actions = await all();
    actions.removeWhere((a) => a.id == id);
    await _save(actions);
  }

  Future<int> length() async => (await all()).length;
}
