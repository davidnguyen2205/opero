import 'package:flutter/material.dart';

import '../api/api_client.dart';
import '../api/auth_store.dart';
import '../api/models.dart';
import '../theme.dart';
import 'timeoff_screen.dart';

/// Identity (email, role, tenant) is REAL — from the auth token / [AuthStore].
/// The stats block (shifts/mo, hours this week, on-time %, tenure) is now REAL
/// from `GET /me/stats`, and the time-off balance + request list is REAL from
/// `GET /me/leave/balance` and `GET /me/leave`. Sign out is real.
class ProfileScreen extends StatefulWidget {
  final ApiClient api;
  final AuthStore auth;
  final VoidCallback onSignOut;

  const ProfileScreen({
    super.key,
    required this.api,
    required this.auth,
    required this.onSignOut,
  });

  @override
  State<ProfileScreen> createState() => _ProfileScreenState();
}

class _ProfileScreenState extends State<ProfileScreen> {
  MyStats? _stats;
  LeaveBalance? _balance;
  List<LeaveRequest>? _leave;
  bool _loading = true;
  String? _error;

  @override
  void initState() {
    super.initState();
    _load();
  }

  Future<void> _load() async {
    setState(() {
      _loading = true;
      _error = null;
    });
    try {
      final results = await Future.wait([
        widget.api.myStats(),
        widget.api.myLeaveBalance(),
        widget.api.myLeave(),
      ]);
      if (!mounted) return;
      setState(() {
        _stats = results[0] as MyStats;
        _balance = results[1] as LeaveBalance;
        _leave = results[2] as List<LeaveRequest>;
        _loading = false;
      });
    } catch (e) {
      if (!mounted) return;
      setState(() {
        _error = 'Couldn\'t load your activity. Pull to retry.';
        _loading = false;
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    final email = widget.auth.email ?? 'field@opero.test';
    final role = _titleCase(widget.auth.role ?? 'employee');
    final tenant = widget.auth.tenantName ?? 'Opero';

    final balance = _balance;
    final pct = (balance == null || balance.entitledDays == 0)
        ? 0.0
        : balance.usedDays / balance.entitledDays;

    return RefreshIndicator(
      onRefresh: _load,
      color: AppColors.orange,
      child: ListView(
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

          if (_error != null) ...[
            _errorBanner(_error!),
            const SizedBox(height: 14),
          ],

          // Stats (REAL — GET /me/stats)
          const Text('Activity', style: kCardTitle),
          const SizedBox(height: 10),
          Row(
            children: [
              Expanded(child: _statCard(_loading ? '—' : '${_stats?.onTimePct ?? 0}%', 'On-time')),
              const SizedBox(width: 10),
              Expanded(child: _statCard(_loading ? '—' : '${_fmtHours(_stats?.hoursThisWeek)}h', 'This week')),
              const SizedBox(width: 10),
              Expanded(child: _statCard(_loading ? '—' : '${_stats?.shiftsThisMonth ?? 0}', 'Shifts/mo')),
            ],
          ),
          if (!_loading && _stats?.tenureDays != null) ...[
            const SizedBox(height: 10),
            _tenureCard(_stats!.tenureDays!),
          ],
          const SizedBox(height: 14),

          // Time off (REAL — GET /me/leave + /me/leave/balance)
          Container(
            padding: const EdgeInsets.all(16),
            decoration: cardDecoration(),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  children: [
                    const Text('Time off',
                        style: TextStyle(fontSize: 14, fontWeight: FontWeight.w700, color: AppColors.ink)),
                    const Spacer(),
                    if (balance != null)
                      Text('${balance.remainingDays} of ${balance.entitledDays} days left',
                          style: const TextStyle(fontSize: 12.5, color: AppColors.grey500)),
                  ],
                ),
                const SizedBox(height: 10),
                ClipRRect(
                  borderRadius: BorderRadius.circular(9999),
                  child: LinearProgressIndicator(
                    value: pct.clamp(0.0, 1.0),
                    minHeight: 8,
                    backgroundColor: AppColors.grey100,
                    valueColor: const AlwaysStoppedAnimation(AppColors.orange),
                  ),
                ),
                const SizedBox(height: 14),
                if (_loading)
                  const Padding(
                    padding: EdgeInsets.symmetric(vertical: 8),
                    child: Text('Loading…', style: TextStyle(fontSize: 13, color: AppColors.grey400)),
                  )
                else if ((_leave ?? const []).isEmpty)
                  const Padding(
                    padding: EdgeInsets.only(bottom: 6),
                    child: Text('No time-off requests yet.',
                        style: TextStyle(fontSize: 13.5, color: AppColors.grey500)),
                  )
                else
                  for (final r in _leave!) _requestRow(r),
                const SizedBox(height: 14),
                PhoneButton(
                  label: 'Request Time Off',
                  icon: Icons.add,
                  tone: PhoneButtonTone.light,
                  onPressed: () async {
                    final created = await Navigator.of(context).push<bool>(
                      MaterialPageRoute(builder: (_) => TimeOffScreen(api: widget.api)),
                    );
                    if (created == true) _load();
                  },
                ),
              ],
            ),
          ),
          const SizedBox(height: 14),

          // Contact (Email is REAL from the token)
          Container(
            decoration: cardDecoration(),
            child: Column(
              children: [
                _contactRow(Icons.mail_outline, 'Email', email, divider: false),
              ],
            ),
          ),
          const SizedBox(height: 16),

          // Sign out (REAL)
          SizedBox(
            width: double.infinity,
            height: 46,
            child: OutlinedButton(
              onPressed: widget.onSignOut,
              style: OutlinedButton.styleFrom(
                foregroundColor: AppColors.red,
                side: const BorderSide(color: AppColors.red200),
                shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
              ),
              child: const Text('Sign Out', style: TextStyle(fontSize: 14, fontWeight: FontWeight.w600)),
            ),
          ),
        ],
      ),
    );
  }

  Widget _errorBanner(String msg) => Container(
        padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 12),
        decoration: BoxDecoration(
          color: AppColors.amber50,
          borderRadius: BorderRadius.circular(12),
          border: Border.all(color: AppColors.amber200),
        ),
        child: Row(
          children: [
            const Icon(Icons.error_outline, size: 17, color: AppColors.amber700),
            const SizedBox(width: 10),
            Expanded(
              child: Text(msg, style: const TextStyle(fontSize: 12.5, color: AppColors.amber700)),
            ),
          ],
        ),
      );

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

  Widget _tenureCard(int days) {
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 12),
      decoration: cardDecoration(),
      child: Row(
        children: [
          const Icon(Icons.workspace_premium_outlined, size: 17, color: AppColors.grey400),
          const SizedBox(width: 10),
          const Text('Tenure', style: TextStyle(fontSize: 13.5, color: AppColors.grey500)),
          const Spacer(),
          Text(_fmtTenure(days),
              style: const TextStyle(fontSize: 13.5, fontWeight: FontWeight.w600, color: AppColors.grey900)),
        ],
      ),
    );
  }

  Widget _requestRow(LeaveRequest r) {
    final approved = r.status == 'approved';
    final rejected = r.status == 'rejected';
    final Color fg = approved ? AppColors.green : (rejected ? AppColors.red : AppColors.amber);
    final Color bg = approved ? AppColors.green50 : (rejected ? AppColors.red50 : AppColors.amber50);
    final Color border = approved ? AppColors.green200 : (rejected ? AppColors.red200 : AppColors.amber200);
    return Padding(
      padding: const EdgeInsets.only(bottom: 8),
      child: Row(
        children: [
          const Icon(Icons.calendar_today, size: 16, color: AppColors.grey400),
          const SizedBox(width: 10),
          Expanded(
            child: Text('${_fmtRange(r.startDate, r.endDate)} · ${_titleCase(r.type)}',
                style: const TextStyle(fontSize: 13.5, color: AppColors.grey700)),
          ),
          Pill(text: _titleCase(r.status), fg: fg, bg: bg, border: border),
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

  static String _fmtHours(double? h) {
    if (h == null) return '0';
    if (h == h.roundToDouble()) return h.toInt().toString();
    return h.toStringAsFixed(1);
  }

  static String _fmtTenure(int days) {
    if (days < 60) return '$days days';
    final months = days ~/ 30;
    if (months < 24) return '$months mo';
    final years = days ~/ 365;
    return '$years yr${years == 1 ? '' : 's'}';
  }

  static const _months = [
    'Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'
  ];

  static String _fmtDate(String iso) {
    final d = DateTime.tryParse(iso);
    if (d == null) return iso;
    return '${d.day} ${_months[d.month - 1]}';
  }

  static String _fmtRange(String start, String end) {
    if (start == end) return _fmtDate(start);
    return '${_fmtDate(start)}–${_fmtDate(end)}';
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
