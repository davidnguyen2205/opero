import 'package:flutter/material.dart';

import '../api/auth_store.dart';
import '../mock/field_mock.dart';
import '../theme.dart';
import 'timeoff_screen.dart';

/// Identity (email, role, tenant) is REAL — from the auth token / [AuthStore].
/// The stats block (on-time %, hours, tours, tenure, languages) and the
/// time-off balance/requests are MOCK — no analytics or leave API in v1. The
/// stats card is clearly badged "Demo". Sign out is real.
class ProfileScreen extends StatelessWidget {
  final AuthStore auth;
  final VoidCallback onSignOut;

  const ProfileScreen({super.key, required this.auth, required this.onSignOut});

  @override
  Widget build(BuildContext context) {
    final email = auth.email ?? 'field@opero.test';
    final role = _titleCase(auth.role ?? 'employee');
    final tenant = auth.tenantName ?? 'Opero';
    const stats = FieldMock.profile;
    final pct = stats.daysTotal == 0 ? 0.0 : stats.daysUsed / stats.daysTotal;

    return ListView(
      padding: const EdgeInsets.fromLTRB(16, 8, 16, 96),
      children: [
        // Identity (REAL)
        Column(
          children: [
            InitialsAvatar(initials: _initials(email), color: AppColors.blue, size: 76),
            const SizedBox(height: 10),
            Text(email,
                textAlign: TextAlign.center,
                style: const TextStyle(fontSize: 19, fontWeight: FontWeight.w700, color: AppColors.ink)),
            const SizedBox(height: 2),
            Text('$role · $tenant',
                style: const TextStyle(fontSize: 13.5, color: AppColors.grey500)),
          ],
        ),
        const SizedBox(height: 16),

        // Stats (MOCK)
        const Row(
          children: [
            Text('Activity', style: kCardTitle),
            SizedBox(width: 8),
            DemoBadge(),
          ],
        ),
        const SizedBox(height: 10),
        Row(
          children: [
            Expanded(child: _statCard('${stats.onTimePct}%', 'On-time')),
            const SizedBox(width: 10),
            Expanded(child: _statCard('${stats.hoursThisWeek}h', 'This week')),
            const SizedBox(width: 10),
            Expanded(child: _statCard('${stats.toursThisMonth}', 'Tours/mo')),
          ],
        ),
        const SizedBox(height: 14),

        // Time off (MOCK)
        Container(
          padding: const EdgeInsets.all(16),
          decoration: cardDecoration(),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                children: [
                  const Text('Time off', style: TextStyle(fontSize: 14, fontWeight: FontWeight.w700, color: AppColors.ink)),
                  const SizedBox(width: 8),
                  const DemoBadge(),
                  const Spacer(),
                  Text('${stats.daysTotal - stats.daysUsed} of ${stats.daysTotal} days left',
                      style: const TextStyle(fontSize: 12.5, color: AppColors.grey500)),
                ],
              ),
              const SizedBox(height: 10),
              ClipRRect(
                borderRadius: BorderRadius.circular(9999),
                child: LinearProgressIndicator(
                  value: pct,
                  minHeight: 8,
                  backgroundColor: AppColors.grey100,
                  valueColor: const AlwaysStoppedAnimation(AppColors.orange),
                ),
              ),
              const SizedBox(height: 14),
              for (final r in stats.requests) _requestRow(r),
              const SizedBox(height: 14),
              PhoneButton(
                label: 'Request Time Off',
                icon: Icons.add,
                tone: PhoneButtonTone.light,
                onPressed: () => Navigator.of(context).push(
                  MaterialPageRoute(builder: (_) => const TimeOffScreen()),
                ),
              ),
            ],
          ),
        ),
        const SizedBox(height: 14),

        // Contact (identity REAL where known; phone/empId are MOCK)
        Container(
          decoration: cardDecoration(),
          child: Column(
            children: [
              _contactRow(Icons.mail_outline, 'Email', email, divider: true),
              _contactRow(Icons.phone, 'Phone', stats.phone, divider: true),
              _contactRow(Icons.badge_outlined, 'Employee ID', stats.employeeId, divider: false),
            ],
          ),
        ),
        const SizedBox(height: 16),

        // Sign out (REAL)
        SizedBox(
          width: double.infinity,
          height: 46,
          child: OutlinedButton(
            onPressed: onSignOut,
            style: OutlinedButton.styleFrom(
              foregroundColor: AppColors.red,
              side: const BorderSide(color: AppColors.red200),
              shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
            ),
            child: const Text('Sign Out', style: TextStyle(fontSize: 14, fontWeight: FontWeight.w600)),
          ),
        ),
      ],
    );
  }

  Widget _statCard(String value, String label) {
    return Container(
      padding: const EdgeInsets.symmetric(vertical: 12, horizontal: 8),
      decoration: cardDecoration(),
      child: Column(
        children: [
          Text(value, style: const TextStyle(fontSize: 20, fontWeight: FontWeight.w700, color: AppColors.ink)),
          const SizedBox(height: 2),
          Text(label, style: const TextStyle(fontSize: 11, color: AppColors.grey400)),
        ],
      ),
    );
  }

  Widget _requestRow(MockTimeOffRequest r) {
    final approved = r.status == 'Approved';
    return Padding(
      padding: const EdgeInsets.only(bottom: 8),
      child: Row(
        children: [
          const Icon(Icons.calendar_today, size: 16, color: AppColors.grey400),
          const SizedBox(width: 10),
          Expanded(
            child: Text('${r.range} · ${r.days}',
                style: const TextStyle(fontSize: 13.5, color: AppColors.grey700)),
          ),
          Pill(
            text: r.status,
            fg: approved ? AppColors.green : AppColors.amber,
            bg: approved ? AppColors.green50 : AppColors.amber50,
            border: approved ? AppColors.green200 : AppColors.amber200,
          ),
        ],
      ),
    );
  }

  Widget _contactRow(IconData icon, String k, String v, {required bool divider}) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 13, vertical: 12),
      decoration: BoxDecoration(
        border: divider ? const Border(bottom: BorderSide(color: AppColors.grey100)) : null,
      ),
      child: Row(
        children: [
          Icon(icon, size: 17, color: AppColors.grey400),
          const SizedBox(width: 12),
          Text(k, style: const TextStyle(fontSize: 13.5, color: AppColors.grey500)),
          const Spacer(),
          Flexible(
            child: Text(v,
                maxLines: 1,
                overflow: TextOverflow.ellipsis,
                style: const TextStyle(fontSize: 13.5, fontWeight: FontWeight.w600, color: AppColors.grey900)),
          ),
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

  static String _titleCase(String s) =>
      s.isEmpty ? s : s[0].toUpperCase() + s.substring(1);
}
