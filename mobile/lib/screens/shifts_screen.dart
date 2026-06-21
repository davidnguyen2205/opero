import 'package:flutter/material.dart';

import '../api/api_client.dart';
import '../api/auth_store.dart';
import '../api/models.dart';
import '../attendance/attendance_service.dart';
import '../attendance/capture.dart';

class ShiftsScreen extends StatefulWidget {
  final ApiClient api;
  final AuthStore auth;
  final AttendanceService attendance;
  final VoidCallback onSignedOut;

  const ShiftsScreen({
    super.key,
    required this.api,
    required this.auth,
    required this.attendance,
    required this.onSignedOut,
  });

  @override
  State<ShiftsScreen> createState() => _ShiftsScreenState();
}

class _ShiftsScreenState extends State<ShiftsScreen> {
  List<Shift>? _shifts;
  Map<String, String?> _activeByShift = {}; // shiftId -> active clientId (or null)
  String? _error;
  bool _loading = false;

  @override
  void initState() {
    super.initState();
    widget.attendance.sync(); // flush any queued actions on open
    _load();
  }

  Future<void> _load() async {
    setState(() {
      _loading = true;
      _error = null;
    });
    try {
      final shifts = await widget.api.myShifts(status: 'published');
      final active = <String, String?>{};
      for (final s in shifts) {
        active[s.id] = await widget.attendance.activeClientId(s.id);
      }
      if (!mounted) return;
      setState(() {
        _shifts = shifts;
        _activeByShift = active;
      });
    } on ApiException catch (e) {
      if (e.statusCode == 401) {
        await widget.auth.clear();
        if (mounted) widget.onSignedOut();
        return;
      }
      setState(() => _error = e.message);
    } catch (_) {
      setState(() => _error = 'Offline — showing what we have. Pull to refresh.');
    } finally {
      if (mounted) setState(() => _loading = false);
    }
  }

  Future<void> _checkIn(Shift s) async {
    final cap = await captureContext(withPhoto: true);
    await widget.attendance.checkIn(shiftId: s.id, lat: cap.lat, lng: cap.lng);
    setState(() => _activeByShift[s.id] = 'pending');
    _toast('Checked in${cap.lat == null ? " (no location)" : ""} — will sync when online');
  }

  Future<void> _checkOut(Shift s) async {
    final cap = await captureContext();
    await widget.attendance.checkOut(shiftId: s.id, lat: cap.lat, lng: cap.lng);
    setState(() => _activeByShift[s.id] = null);
    _toast('Checked out — will sync when online');
  }

  void _toast(String msg) {
    if (!mounted) return;
    ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(msg)));
  }

  Future<void> _signOut() async {
    await widget.auth.clear();
    if (mounted) widget.onSignedOut();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('My shifts'),
        actions: [
          // Pending-sync badge driven by the attendance service.
          AnimatedBuilder(
            animation: widget.attendance,
            builder: (_, __) {
              final n = widget.attendance.pendingCount;
              if (n == 0) return const SizedBox.shrink();
              return Padding(
                padding: const EdgeInsets.symmetric(horizontal: 12),
                child: Center(child: Text('$n queued')),
              );
            },
          ),
          IconButton(onPressed: _signOut, icon: const Icon(Icons.logout)),
        ],
      ),
      body: RefreshIndicator(
        onRefresh: () async {
          await widget.attendance.sync();
          await _load();
        },
        child: _buildBody(),
      ),
    );
  }

  Widget _buildBody() {
    if (_shifts == null && _loading) {
      return const Center(child: CircularProgressIndicator());
    }
    final shifts = _shifts ?? [];
    return ListView(
      children: [
        if (_error != null)
          Padding(
            padding: const EdgeInsets.all(16),
            child: Text(_error!, style: const TextStyle(color: Colors.orange)),
          ),
        if (shifts.isEmpty)
          const Padding(
            padding: EdgeInsets.all(24),
            child: Center(child: Text('No published shifts.')),
          ),
        for (final s in shifts) _shiftTile(s),
      ],
    );
  }

  Widget _shiftTile(Shift s) {
    final active = _activeByShift[s.id];
    final isOpen = active != null;
    return Card(
      margin: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
      child: ListTile(
        title: Text('${_fmt(s.startsAt)} → ${_fmt(s.endsAt)}'),
        subtitle: Text(s.notes ?? (isOpen ? 'Checked in' : 'Not checked in')),
        trailing: isOpen
            ? OutlinedButton(onPressed: () => _checkOut(s), child: const Text('Check out'))
            : FilledButton(onPressed: () => _checkIn(s), child: const Text('Check in')),
      ),
    );
  }

  String _fmt(DateTime t) {
    final l = t.toLocal();
    String two(int n) => n.toString().padLeft(2, '0');
    return '${two(l.month)}/${two(l.day)} ${two(l.hour)}:${two(l.minute)}';
  }
}
