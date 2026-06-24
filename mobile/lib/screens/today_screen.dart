import 'package:flutter/material.dart';

import '../attendance/attendance_service.dart';
import '../theme.dart';
import 'checkin_flow.dart';
import 'field_data.dart';
import 'shift_view.dart';

/// REAL — the "Today" home. Shows the greeting and today's published shift (from
/// [FieldData] → GET /me/shifts), with a Check In button that opens the real
/// [CheckInFlow]. If an attendance is already open for the shift, the button
/// resumes the active screen instead.
class TodayScreen extends StatelessWidget {
  final FieldData data;
  final AttendanceService attendance;
  final String emailForGreeting;
  final Future<String?> Function(String shiftId) activeClientId;

  const TodayScreen({
    super.key,
    required this.data,
    required this.attendance,
    required this.emailForGreeting,
    required this.activeClientId,
  });

  @override
  Widget build(BuildContext context) {
    return AnimatedBuilder(
      animation: data,
      builder: (context, _) {
        final shift = data.todayShift;
        return RefreshIndicator(
          onRefresh: data.load,
          child: ListView(
            padding: const EdgeInsets.fromLTRB(16, 4, 16, 96),
            children: [
              Text('Good day, ${_firstName(emailForGreeting)}',
                  style: const TextStyle(fontSize: 13, color: AppColors.grey500)),
              const SizedBox(height: 2),
              Text('Today · ${_todayLong()}', style: kH1),
              const SizedBox(height: 14),
              if (data.loading && !data.loaded)
                const Padding(
                  padding: EdgeInsets.only(top: 60),
                  child: Center(child: CircularProgressIndicator(color: AppColors.orange)),
                )
              else if (shift == null)
                _noShift(data.error)
              else
                _ShiftToday(
                  shift: shift,
                  attendance: attendance,
                  activeClientId: activeClientId,
                  onReturned: data.load,
                ),
            ],
          ),
        );
      },
    );
  }

  Widget _noShift(String? error) {
    return Container(
      margin: const EdgeInsets.only(top: 24),
      padding: const EdgeInsets.all(28),
      decoration: BoxDecoration(
        borderRadius: BorderRadius.circular(14),
        border: Border.all(color: AppColors.grey200),
      ),
      child: Column(
        children: [
          const Icon(Icons.event_available, size: 32, color: AppColors.grey300),
          const SizedBox(height: 12),
          Text(error ?? 'No shift scheduled today.',
              textAlign: TextAlign.center,
              style: const TextStyle(fontSize: 14, color: AppColors.grey500)),
        ],
      ),
    );
  }

  static String _firstName(String email) {
    final raw = email.split('@').first.split(RegExp(r'[._\-]+')).first;
    if (raw.isEmpty) return 'there';
    return raw[0].toUpperCase() + raw.substring(1);
  }

  static String _todayLong() {
    const wd = ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'];
    const mo = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'];
    final n = DateTime.now();
    return '${wd[n.weekday - 1]} ${n.day} ${mo[n.month - 1]}';
  }
}

class _ShiftToday extends StatefulWidget {
  final ShiftView shift;
  final AttendanceService attendance;
  final Future<String?> Function(String shiftId) activeClientId;
  final Future<void> Function() onReturned;

  const _ShiftToday({
    required this.shift,
    required this.attendance,
    required this.activeClientId,
    required this.onReturned,
  });

  @override
  State<_ShiftToday> createState() => _ShiftTodayState();
}

class _ShiftTodayState extends State<_ShiftToday> {
  bool _active = false;

  @override
  void initState() {
    super.initState();
    _refreshActive();
  }

  Future<void> _refreshActive() async {
    final id = await widget.activeClientId(widget.shift.id);
    if (mounted) setState(() => _active = id != null);
  }

  Future<void> _openFlow() async {
    await Navigator.of(context).push(
      MaterialPageRoute(
        builder: (_) => CheckInFlow(
          shift: widget.shift,
          attendance: widget.attendance,
          alreadyActive: _active,
        ),
      ),
    );
    await widget.onReturned();
    await _refreshActive();
  }

  @override
  Widget build(BuildContext context) {
    final s = widget.shift;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        Container(
          padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
          decoration: BoxDecoration(
            color: _active ? AppColors.green50 : AppColors.orange50,
            borderRadius: BorderRadius.circular(10),
            border: Border.all(color: _active ? AppColors.green200 : AppColors.orange200),
          ),
          child: Row(
            children: [
              Icon(_active ? Icons.check_circle : Icons.schedule,
                  size: 16, color: _active ? AppColors.green : AppColors.orange),
              const SizedBox(width: 7),
              Expanded(
                child: Text(
                  _active ? 'You\'re on shift' : 'Starts at ${s.timeRange.split(' – ').first}',
                  style: TextStyle(
                      fontSize: 13,
                      fontWeight: FontWeight.w600,
                      color: _active ? AppColors.green700 : AppColors.orange700),
                ),
              ),
            ],
          ),
        ),
        const SizedBox(height: 14),
        Container(
          padding: const EdgeInsets.all(16),
          decoration: cardDecoration(),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                children: [
                  Container(
                    width: 10,
                    height: 10,
                    decoration: BoxDecoration(color: s.color, borderRadius: BorderRadius.circular(3)),
                  ),
                  const SizedBox(width: 8),
                  Expanded(child: Text(s.title, style: kCardTitle)),
                ],
              ),
              const SizedBox(height: 10),
              _row(Icons.schedule, s.timeRange),
              if (s.address != null) _row(Icons.place, s.address!),
              if (s.notes != null) _row(Icons.sticky_note_2_outlined, s.notes!),
            ],
          ),
        ),
        const SizedBox(height: 16),
        PhoneButton(
          label: _active ? 'Resume shift' : 'Check In',
          icon: Icons.place,
          onPressed: _openFlow,
        ),
        const SizedBox(height: 10),
        const Center(
          child: Text('Check-in captures your location and a photo',
              style: TextStyle(fontSize: 12, color: AppColors.grey400)),
        ),
      ],
    );
  }

  Widget _row(IconData icon, String text) => Padding(
        padding: const EdgeInsets.symmetric(vertical: 6),
        child: Row(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Icon(icon, size: 17, color: AppColors.grey400),
            const SizedBox(width: 10),
            Expanded(child: Text(text, style: const TextStyle(fontSize: 14, color: AppColors.grey700))),
          ],
        ),
      );
}
