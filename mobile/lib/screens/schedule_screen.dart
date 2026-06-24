import 'package:flutter/material.dart';

import '../attendance/attendance_service.dart';
import '../theme.dart';
import 'field_data.dart';
import 'shift_detail_screen.dart';
import 'shift_view.dart';

/// REAL — the signed-in user's upcoming shifts from GET /me/shifts (via
/// [FieldData]), grouped as a simple chronological list. Tapping a shift opens
/// its detail. No party size / tour name exists in the API, so each row shows
/// location name, time window and status only.
class ScheduleScreen extends StatelessWidget {
  final FieldData data;
  final AttendanceService attendance;
  final Future<String?> Function(String shiftId) activeClientId;

  const ScheduleScreen({
    super.key,
    required this.data,
    required this.attendance,
    required this.activeClientId,
  });

  @override
  Widget build(BuildContext context) {
    return AnimatedBuilder(
      animation: data,
      builder: (context, _) {
        final shifts = data.shifts;
        return RefreshIndicator(
          onRefresh: data.load,
          child: ListView(
            padding: const EdgeInsets.fromLTRB(16, 4, 16, 96),
            children: [
              const Text('My Schedule', style: kH1),
              const SizedBox(height: 4),
              Text(
                shifts.isEmpty ? 'No upcoming shifts' : 'Your published shifts',
                style: const TextStyle(fontSize: 13, color: AppColors.grey500),
              ),
              const SizedBox(height: 14),
              if (data.loading && shifts.isEmpty)
                const Padding(
                  padding: EdgeInsets.only(top: 60),
                  child: Center(child: CircularProgressIndicator(color: AppColors.orange)),
                )
              else if (shifts.isEmpty)
                _empty(data.error)
              else
                for (final s in shifts) ...[
                  _ScheduleRow(
                    shift: s,
                    onTap: () => _open(context, s),
                  ),
                  const SizedBox(height: 10),
                ],
            ],
          ),
        );
      },
    );
  }

  Future<void> _open(BuildContext context, ShiftView s) async {
    final active = (await activeClientId(s.id)) != null;
    if (!context.mounted) return;
    await Navigator.of(context).push(
      MaterialPageRoute(
        builder: (_) => ShiftDetailScreen(shift: s, attendance: attendance, active: active),
      ),
    );
    await data.load();
  }

  Widget _empty(String? error) {
    return Container(
      margin: const EdgeInsets.only(top: 24),
      padding: const EdgeInsets.all(28),
      decoration: BoxDecoration(
        borderRadius: BorderRadius.circular(14),
        border: Border.all(color: AppColors.grey200, style: BorderStyle.solid),
      ),
      child: Center(
        child: Text(
          error ?? 'Nothing scheduled yet. Pull to refresh.',
          textAlign: TextAlign.center,
          style: const TextStyle(fontSize: 13.5, color: AppColors.grey500),
        ),
      ),
    );
  }
}

class _ScheduleRow extends StatelessWidget {
  final ShiftView shift;
  final VoidCallback onTap;
  const _ScheduleRow({required this.shift, required this.onTap});

  @override
  Widget build(BuildContext context) {
    return InkWell(
      borderRadius: BorderRadius.circular(14),
      onTap: onTap,
      child: Container(
        decoration: cardDecoration(),
        clipBehavior: Clip.antiAlias,
        child: IntrinsicHeight(
          child: Row(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              Container(width: 5, color: shift.color),
              Container(
                width: 50,
                padding: const EdgeInsets.symmetric(vertical: 12),
                decoration: const BoxDecoration(
                  border: Border(right: BorderSide(color: AppColors.grey100)),
                ),
                child: Column(
                  mainAxisAlignment: MainAxisAlignment.center,
                  children: [
                    Text(shift.weekdayShort,
                        style: TextStyle(
                            fontSize: 11,
                            fontWeight: FontWeight.w600,
                            color: shift.isToday ? AppColors.orange : AppColors.grey400)),
                    Text(shift.dayOfMonth,
                        style: TextStyle(
                            fontSize: 18,
                            fontWeight: FontWeight.w700,
                            color: shift.isToday ? AppColors.orange : AppColors.grey700)),
                  ],
                ),
              ),
              Expanded(
                child: Padding(
                  padding: const EdgeInsets.symmetric(horizontal: 13, vertical: 11),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      Row(
                        children: [
                          Flexible(
                            child: Text(shift.title,
                                maxLines: 1,
                                overflow: TextOverflow.ellipsis,
                                style: const TextStyle(
                                    fontSize: 14.5, fontWeight: FontWeight.w700, color: AppColors.ink)),
                          ),
                          const SizedBox(width: 7),
                          if (shift.isToday)
                            const Pill(text: 'TODAY', fg: AppColors.orange, bg: AppColors.orange50, border: AppColors.orange200)
                          else if (shift.isDraft)
                            const Pill(text: 'DRAFT', fg: AppColors.grey400, bg: Colors.white, border: AppColors.grey200),
                        ],
                      ),
                      const SizedBox(height: 5),
                      Row(
                        children: [
                          const Icon(Icons.schedule, size: 13, color: AppColors.grey400),
                          const SizedBox(width: 4),
                          Text(shift.timeRange,
                              style: const TextStyle(fontSize: 12.5, color: AppColors.grey500)),
                        ],
                      ),
                    ],
                  ),
                ),
              ),
              const Padding(
                padding: EdgeInsets.only(right: 10),
                child: Icon(Icons.chevron_right, size: 18, color: AppColors.grey300),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
