import 'package:flutter/material.dart';

import '../attendance/attendance_service.dart';
import '../theme.dart';
import 'checkin_flow.dart';
import 'shift_view.dart';

/// REAL — detail for one [ShiftView]. The API has only location, time window,
/// notes and status, so "Meeting point" = the location name/address and any
/// shift notes. The Check In button shows only for today's shift and pushes the
/// real [CheckInFlow].
class ShiftDetailScreen extends StatelessWidget {
  final ShiftView shift;
  final AttendanceService attendance;

  /// Whether there is an open (checked-in) attendance for this shift.
  final bool active;

  const ShiftDetailScreen({
    super.key,
    required this.shift,
    required this.attendance,
    this.active = false,
  });

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        leading: IconButton(
          icon: const Icon(Icons.chevron_left, color: AppColors.grey700),
          onPressed: () => Navigator.of(context).maybePop(),
        ),
        title: Text(shift.dayLong, style: kCardTitle),
      ),
      body: SafeArea(
        child: Column(
          children: [
            Expanded(
              child: ListView(
                padding: const EdgeInsets.fromLTRB(16, 14, 16, 24),
                children: [
                  _summaryCard(),
                  const SizedBox(height: 14),
                  _meetingCard(),
                  if (shift.notes != null) ...[
                    const SizedBox(height: 14),
                    _notesCard(shift.notes!),
                  ],
                ],
              ),
            ),
            if (shift.isToday)
              Padding(
                padding: const EdgeInsets.fromLTRB(16, 12, 16, 24),
                child: PhoneButton(
                  label: active ? 'Resume shift' : 'Check In',
                  icon: Icons.place,
                  onPressed: () async {
                    await Navigator.of(context).push(
                      MaterialPageRoute(
                        builder: (_) => CheckInFlow(
                          shift: shift,
                          attendance: attendance,
                          alreadyActive: active,
                        ),
                      ),
                    );
                  },
                ),
              ),
          ],
        ),
      ),
    );
  }

  Widget _summaryCard() {
    return Container(
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
                decoration: BoxDecoration(color: shift.color, borderRadius: BorderRadius.circular(3)),
              ),
              const SizedBox(width: 9),
              Expanded(child: Text(shift.title, style: kCardTitle)),
              if (shift.isDraft)
                const Pill(text: 'DRAFT', fg: AppColors.grey400, bg: Colors.white, border: AppColors.grey200),
            ],
          ),
          const SizedBox(height: 10),
          _row(Icons.schedule, shift.timeRange),
          if (shift.address != null) _row(Icons.place, shift.address!),
          _row(Icons.calendar_today, shift.dayLong),
        ],
      ),
    );
  }

  Widget _meetingCard() {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: cardDecoration(),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const Text('MEETING POINT',
              style: TextStyle(
                  fontSize: 12, fontWeight: FontWeight.w700, color: AppColors.grey400, letterSpacing: 0.6)),
          const SizedBox(height: 8),
          Text(shift.address ?? shift.title,
              style: const TextStyle(fontSize: 14, color: AppColors.grey700, height: 1.5)),
          const SizedBox(height: 12),
          Container(
            height: 110,
            decoration: BoxDecoration(
              color: AppColors.grey100,
              borderRadius: BorderRadius.circular(10),
              border: Border.all(color: AppColors.grey200),
            ),
            child: Center(child: Icon(Icons.place, size: 26, color: shift.color)),
          ),
        ],
      ),
    );
  }

  Widget _notesCard(String notes) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 12),
      decoration: BoxDecoration(
        color: AppColors.orange50,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: AppColors.orange200),
      ),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const Icon(Icons.warning_amber, size: 17, color: AppColors.orange),
          const SizedBox(width: 10),
          Expanded(
            child: Text(notes,
                style: const TextStyle(fontSize: 13, color: AppColors.orange700, height: 1.5)),
          ),
        ],
      ),
    );
  }

  Widget _row(IconData icon, String text) => Padding(
        padding: const EdgeInsets.symmetric(vertical: 6),
        child: Row(
          children: [
            Icon(icon, size: 17, color: AppColors.grey400),
            const SizedBox(width: 10),
            Expanded(child: Text(text, style: const TextStyle(fontSize: 14, color: AppColors.grey700))),
          ],
        ),
      );
}
