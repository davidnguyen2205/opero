import 'package:flutter/material.dart';

import '../mock/field_mock.dart';
import '../theme.dart';

/// MOCK — no notifications API in v1. Static demo list, clearly badged.
class NotificationsScreen extends StatelessWidget {
  const NotificationsScreen({super.key});

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
            Text('Notifications', style: kCardTitle),
            SizedBox(width: 8),
            DemoBadge(),
          ],
        ),
      ),
      body: ListView(
        padding: const EdgeInsets.fromLTRB(16, 12, 16, 24),
        children: [
          for (final n in FieldMock.notifications) ...[
            _row(n),
            const SizedBox(height: 10),
          ],
        ],
      ),
    );
  }

  Widget _row(MockNotification n) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 13, vertical: 12),
      decoration: BoxDecoration(
        color: n.unread ? AppColors.orange50 : Colors.white,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: n.unread ? AppColors.orange200 : AppColors.grey200),
      ),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Container(
            width: 38,
            height: 38,
            decoration: BoxDecoration(
              color: n.tone.withValues(alpha: 0.1),
              borderRadius: BorderRadius.circular(10),
            ),
            child: Icon(n.icon, size: 18, color: n.tone),
          ),
          const SizedBox(width: 12),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  children: [
                    Expanded(
                      child: Text(n.title,
                          style: const TextStyle(fontSize: 14, fontWeight: FontWeight.w700, color: AppColors.ink)),
                    ),
                    Text(n.time, style: const TextStyle(fontSize: 11, color: AppColors.grey400)),
                  ],
                ),
                const SizedBox(height: 2),
                Text(n.body,
                    style: const TextStyle(fontSize: 13, color: AppColors.grey500, height: 1.4)),
              ],
            ),
          ),
        ],
      ),
    );
  }
}
