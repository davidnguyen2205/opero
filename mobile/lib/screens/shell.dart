import 'package:flutter/material.dart';

import '../api/api_client.dart';
import '../api/auth_store.dart';
import '../attendance/attendance_service.dart';
import '../mock/field_mock.dart';
import '../theme.dart';
import 'field_data.dart';
import 'inbox_screen.dart';
import 'notifications_screen.dart';
import 'profile_screen.dart';
import 'schedule_screen.dart';
import 'today_screen.dart';

/// The signed-in app shell: a header (logo, notifications bell, avatar) plus a
/// bottom tab bar (Today / Schedule / Inbox / Me). Owns the shared [FieldData]
/// (real shifts+locations) and reacts to 401 by signing out.
class Shell extends StatefulWidget {
  final ApiClient api;
  final AuthStore auth;
  final AttendanceService attendance;
  final VoidCallback onSignedOut;

  const Shell({
    super.key,
    required this.api,
    required this.auth,
    required this.attendance,
    required this.onSignedOut,
  });

  @override
  State<Shell> createState() => _ShellState();
}

class _ShellState extends State<Shell> {
  int _tab = 0;
  late final FieldData _data = FieldData(api: widget.api, auth: widget.auth);

  @override
  void initState() {
    super.initState();
    _data.addListener(_onData);
    widget.attendance.sync(); // flush any queued attendance on open
    _data.load();
  }

  @override
  void dispose() {
    _data.removeListener(_onData);
    super.dispose();
  }

  void _onData() {
    if (_data.unauthorized) {
      _signOut();
    }
  }

  Future<void> _signOut() async {
    await widget.auth.clear();
    if (mounted) widget.onSignedOut();
  }

  @override
  Widget build(BuildContext context) {
    final body = IndexedStack(
      index: _tab,
      children: [
        TodayScreen(
          data: _data,
          attendance: widget.attendance,
          emailForGreeting: widget.auth.email ?? '',
          activeClientId: widget.attendance.activeClientId,
        ),
        ScheduleScreen(
          data: _data,
          attendance: widget.attendance,
          activeClientId: widget.attendance.activeClientId,
        ),
        const InboxScreen(),
        ProfileScreen(auth: widget.auth, onSignOut: _signOut),
      ],
    );

    return Scaffold(
      body: SafeArea(
        bottom: false,
        child: Column(
          children: [
            _header(),
            Expanded(child: body),
          ],
        ),
      ),
      bottomNavigationBar: _bottomNav(),
    );
  }

  Widget _header() {
    return Container(
      padding: const EdgeInsets.fromLTRB(16, 10, 12, 10),
      decoration: const BoxDecoration(
        border: Border(bottom: BorderSide(color: AppColors.grey100)),
      ),
      child: Row(
        children: [
          const OperoMark(size: 26),
          const SizedBox(width: 10),
          const Text('Opero',
              style: TextStyle(fontSize: 17, fontWeight: FontWeight.w700, color: AppColors.ink)),
          const Spacer(),
          AnimatedBuilder(
            animation: widget.attendance,
            builder: (context, _) {
              final n = widget.attendance.pendingCount;
              if (n == 0) return const SizedBox.shrink();
              return Padding(
                padding: const EdgeInsets.only(right: 8),
                child: Container(
                  padding: const EdgeInsets.symmetric(horizontal: 9, vertical: 4),
                  decoration: BoxDecoration(
                    color: AppColors.amber50,
                    borderRadius: BorderRadius.circular(9999),
                    border: Border.all(color: AppColors.amber200),
                  ),
                  child: Row(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      const Icon(Icons.wifi_off, size: 12, color: AppColors.amber700),
                      const SizedBox(width: 4),
                      Text('$n queued',
                          style: const TextStyle(
                              fontSize: 11, fontWeight: FontWeight.w600, color: AppColors.amber700)),
                    ],
                  ),
                ),
              );
            },
          ),
          IconButton(
            icon: Stack(
              clipBehavior: Clip.none,
              children: [
                const Icon(Icons.notifications_none, size: 22, color: AppColors.grey700),
                if (FieldMock.unreadNotifications > 0)
                  Positioned(
                    top: -1,
                    right: -1,
                    child: Container(
                      width: 7,
                      height: 7,
                      decoration: BoxDecoration(
                        color: AppColors.orange,
                        shape: BoxShape.circle,
                        border: Border.all(color: Colors.white, width: 1.5),
                      ),
                    ),
                  ),
              ],
            ),
            onPressed: () => Navigator.of(context).push(
              MaterialPageRoute(builder: (_) => const NotificationsScreen()),
            ),
          ),
          InitialsAvatar(initials: _initials(widget.auth.email ?? ''), color: AppColors.blue, size: 30),
        ],
      ),
    );
  }

  Widget _bottomNav() {
    final items = [
      (Icons.place, 'Today'),
      (Icons.calendar_today, 'Schedule'),
      (Icons.send, 'Inbox'),
      (Icons.person, 'Me'),
    ];
    return Container(
      decoration: const BoxDecoration(
        color: Colors.white,
        border: Border(top: BorderSide(color: AppColors.grey200)),
      ),
      child: SafeArea(
        top: false,
        child: SizedBox(
          height: 62,
          child: Row(
            children: [
              for (var i = 0; i < items.length; i++)
                Expanded(child: _navItem(i, items[i].$1, items[i].$2)),
            ],
          ),
        ),
      ),
    );
  }

  Widget _navItem(int i, IconData icon, String label) {
    final on = _tab == i;
    final showBadge = i == 2 && FieldMock.unreadThreads > 0;
    return InkWell(
      onTap: () => setState(() => _tab = i),
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Stack(
            clipBehavior: Clip.none,
            children: [
              Icon(icon, size: 22, color: on ? AppColors.orange : AppColors.grey400),
              if (showBadge)
                Positioned(
                  top: -4,
                  right: -8,
                  child: Container(
                    constraints: const BoxConstraints(minWidth: 15),
                    height: 15,
                    alignment: Alignment.center,
                    padding: const EdgeInsets.symmetric(horizontal: 3),
                    decoration: BoxDecoration(
                      color: AppColors.orange,
                      borderRadius: BorderRadius.circular(9999),
                      border: Border.all(color: Colors.white, width: 1.5),
                    ),
                    child: Text('${FieldMock.unreadThreads}',
                        style: const TextStyle(fontSize: 9, fontWeight: FontWeight.w700, color: Colors.white)),
                  ),
                ),
            ],
          ),
          const SizedBox(height: 3),
          Text(label,
              style: TextStyle(
                  fontSize: 10.5,
                  fontWeight: FontWeight.w600,
                  color: on ? AppColors.orange : AppColors.grey400)),
        ],
      ),
    );
  }

  static String _initials(String email) {
    final name = email.split('@').first;
    final parts = name.split(RegExp(r'[._\-]+')).where((p) => p.isNotEmpty).toList();
    if (parts.isEmpty) return '?';
    if (parts.length == 1) return parts[0].substring(0, parts[0].length >= 2 ? 2 : 1).toUpperCase();
    return (parts[0][0] + parts[parts.length - 1][0]).toUpperCase();
  }
}
