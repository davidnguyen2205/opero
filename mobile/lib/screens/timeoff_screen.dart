import 'package:flutter/material.dart';

import '../api/api_client.dart';
import '../theme.dart';

/// REAL — submits a time-off request via `POST /me/leave`. The manager reviews
/// it through the web app's /leave endpoints. Pops `true` on success so the
/// caller can refresh the request list + balance.
class TimeOffScreen extends StatefulWidget {
  final ApiClient api;
  const TimeOffScreen({super.key, required this.api});

  @override
  State<TimeOffScreen> createState() => _TimeOffScreenState();
}

class _TimeOffScreenState extends State<TimeOffScreen> {
  // UI label -> API enum (holiday|sick|personal).
  static const _types = {'Holiday': 'holiday', 'Sick': 'sick', 'Personal': 'personal'};
  String _type = 'Holiday';
  DateTime? _from;
  DateTime? _to;
  final _note = TextEditingController();
  bool _submitting = false;
  String? _error;

  @override
  void dispose() {
    _note.dispose();
    super.dispose();
  }

  Future<void> _pick(bool isFrom) async {
    final now = DateTime.now();
    final picked = await showDatePicker(
      context: context,
      initialDate: (isFrom ? _from : _to) ?? now,
      firstDate: now.subtract(const Duration(days: 1)),
      lastDate: now.add(const Duration(days: 365)),
    );
    if (picked != null) setState(() => isFrom ? _from = picked : _to = picked);
  }

  String _fmt(DateTime? d) => d == null ? 'Select' : '${d.day}/${d.month}/${d.year}';

  String _apiDate(DateTime d) =>
      '${d.year.toString().padLeft(4, '0')}-${d.month.toString().padLeft(2, '0')}-${d.day.toString().padLeft(2, '0')}';

  Future<void> _submit() async {
    final from = _from, to = _to;
    if (from == null || to == null) {
      setState(() => _error = 'Please choose a start and end date.');
      return;
    }
    if (to.isBefore(from)) {
      setState(() => _error = 'The end date must be on or after the start date.');
      return;
    }
    setState(() {
      _submitting = true;
      _error = null;
    });
    try {
      await widget.api.createLeave(
        type: _types[_type]!,
        startDate: _apiDate(from),
        endDate: _apiDate(to),
        note: _note.text.trim(),
      );
      if (!mounted) return;
      Navigator.of(context).pop(true);
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Time-off request submitted.')),
      );
    } on ApiException catch (e) {
      if (!mounted) return;
      setState(() {
        _submitting = false;
        _error = e.message;
      });
    } catch (_) {
      if (!mounted) return;
      setState(() {
        _submitting = false;
        _error = 'Couldn\'t submit your request. Check your connection and try again.';
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        leading: IconButton(
          icon: const Icon(Icons.chevron_left, color: AppColors.grey700),
          onPressed: () => Navigator.of(context).maybePop(),
        ),
        title: const Text('Request Time Off', style: kCardTitle),
      ),
      body: SafeArea(
        child: Column(
          children: [
            Expanded(
              child: ListView(
                padding: const EdgeInsets.fromLTRB(16, 16, 16, 24),
                children: [
                  _label('Type'),
                  Row(
                    children: [
                      for (final t in const ['Holiday', 'Sick', 'Personal']) ...[
                        Expanded(child: _typeChip(t)),
                        if (t != 'Personal') const SizedBox(width: 8),
                      ],
                    ],
                  ),
                  const SizedBox(height: 16),
                  Row(
                    children: [
                      Expanded(child: _dateField('From', _from, () => _pick(true))),
                      const SizedBox(width: 12),
                      Expanded(child: _dateField('To', _to, () => _pick(false))),
                    ],
                  ),
                  const SizedBox(height: 16),
                  _label('Note (optional)'),
                  TextField(
                    controller: _note,
                    maxLines: 3,
                    style: const TextStyle(fontSize: 14, color: AppColors.grey900),
                    decoration: _inputDecoration('Anything your manager should know…'),
                  ),
                  const SizedBox(height: 16),
                  if (_error != null)
                    Container(
                      padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 12),
                      decoration: BoxDecoration(
                        color: AppColors.red50,
                        borderRadius: BorderRadius.circular(12),
                        border: Border.all(color: AppColors.red200),
                      ),
                      child: Row(
                        children: [
                          const Icon(Icons.error_outline, size: 17, color: AppColors.red),
                          const SizedBox(width: 10),
                          Expanded(
                            child: Text(_error!, style: const TextStyle(fontSize: 12.5, color: AppColors.red)),
                          ),
                        ],
                      ),
                    )
                  else
                    Container(
                      padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 12),
                      decoration: BoxDecoration(
                        color: AppColors.grey50,
                        borderRadius: BorderRadius.circular(12),
                        border: Border.all(color: AppColors.grey200),
                      ),
                      child: const Row(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Icon(Icons.group, size: 17, color: AppColors.grey400),
                          SizedBox(width: 10),
                          Expanded(
                            child: Text(
                              'This goes to your manager for approval.',
                              style: TextStyle(fontSize: 12.5, color: AppColors.grey500, height: 1.5),
                            ),
                          ),
                        ],
                      ),
                    ),
                ],
              ),
            ),
            Padding(
              padding: const EdgeInsets.fromLTRB(16, 12, 16, 24),
              child: PhoneButton(
                label: _submitting ? 'Submitting…' : 'Submit Request',
                icon: Icons.send,
                onPressed: _submitting ? null : _submit,
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _typeChip(String t) {
    final on = _type == t;
    return InkWell(
      borderRadius: BorderRadius.circular(10),
      onTap: () => setState(() => _type = t),
      child: Container(
        height: 40,
        alignment: Alignment.center,
        decoration: BoxDecoration(
          color: on ? AppColors.orange50 : Colors.white,
          borderRadius: BorderRadius.circular(10),
          border: Border.all(color: on ? AppColors.orange200 : AppColors.grey200),
        ),
        child: Text(t,
            style: TextStyle(
                fontSize: 13,
                fontWeight: FontWeight.w600,
                color: on ? AppColors.orange : AppColors.grey500)),
      ),
    );
  }

  Widget _dateField(String label, DateTime? value, VoidCallback onTap) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        _label(label),
        InkWell(
          borderRadius: BorderRadius.circular(11),
          onTap: onTap,
          child: Container(
            height: 44,
            padding: const EdgeInsets.symmetric(horizontal: 12),
            decoration: BoxDecoration(
              color: Colors.white,
              borderRadius: BorderRadius.circular(11),
              border: Border.all(color: AppColors.grey200),
            ),
            child: Row(
              children: [
                Expanded(
                  child: Text(_fmt(value),
                      style: TextStyle(
                          fontSize: 14,
                          color: value == null ? AppColors.grey400 : AppColors.grey900)),
                ),
                const Icon(Icons.calendar_today, size: 15, color: AppColors.grey400),
              ],
            ),
          ),
        ),
      ],
    );
  }

  Widget _label(String t) => Padding(
        padding: const EdgeInsets.only(bottom: 7),
        child: Text(t,
            style: const TextStyle(fontSize: 12.5, fontWeight: FontWeight.w600, color: AppColors.grey700)),
      );

  InputDecoration _inputDecoration(String hint) => InputDecoration(
        hintText: hint,
        isDense: true,
        filled: true,
        fillColor: Colors.white,
        contentPadding: const EdgeInsets.all(12),
        enabledBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(11),
          borderSide: const BorderSide(color: AppColors.grey200),
        ),
        focusedBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(11),
          borderSide: const BorderSide(color: AppColors.orange),
        ),
      );
}
