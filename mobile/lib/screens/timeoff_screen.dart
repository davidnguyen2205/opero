import 'package:flutter/material.dart';

import '../mock/field_mock.dart';
import '../theme.dart';

/// MOCK — no leave/time-off API in v1 (it's a v1.1 area). This form is a
/// front-end demo: submitting does not persist or notify anyone. Clearly badged.
class TimeOffScreen extends StatefulWidget {
  const TimeOffScreen({super.key});

  @override
  State<TimeOffScreen> createState() => _TimeOffScreenState();
}

class _TimeOffScreenState extends State<TimeOffScreen> {
  String _type = 'Holiday';
  DateTime? _from;
  DateTime? _to;
  final _note = TextEditingController();

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

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        leading: IconButton(
          icon: const Icon(Icons.chevron_left, color: AppColors.grey700),
          onPressed: () => Navigator.of(context).maybePop(),
        ),
        title: const Row(
          children: [
            Text('Request Time Off', style: kCardTitle),
            SizedBox(width: 8),
            DemoBadge(),
          ],
        ),
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
                  Container(
                    padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 12),
                    decoration: BoxDecoration(
                      color: AppColors.grey50,
                      borderRadius: BorderRadius.circular(12),
                      border: Border.all(color: AppColors.grey200),
                    ),
                    child: Row(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        const Icon(Icons.group, size: 17, color: AppColors.grey400),
                        const SizedBox(width: 10),
                        Expanded(
                          child: Text.rich(
                            TextSpan(
                              style: const TextStyle(fontSize: 12.5, color: AppColors.grey500, height: 1.5),
                              children: [
                                const TextSpan(text: 'In the full product this goes to '),
                                TextSpan(
                                    text: FieldMock.profile.managerName,
                                    style: const TextStyle(fontWeight: FontWeight.w600, color: AppColors.grey700)),
                                const TextSpan(text: ' for approval. (Demo — not submitted anywhere.)'),
                              ],
                            ),
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
                label: 'Submit Request',
                icon: Icons.send,
                onPressed: () {
                  Navigator.of(context).pop();
                  ScaffoldMessenger.of(context).showSnackBar(
                    const SnackBar(content: Text('Demo only — request was not submitted.')),
                  );
                },
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
