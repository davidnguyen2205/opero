import 'package:flutter/material.dart';

import '../mock/field_mock.dart';
import '../theme.dart';
import 'notifications_screen.dart';
import 'thread_screen.dart';

/// MOCK — there is no messaging API in Opero v1. This thread list is static
/// demo data (clearly badged "Demo"). The bell opens the (also-mock)
/// Notifications screen.
class InboxScreen extends StatelessWidget {
  const InboxScreen({super.key});

  @override
  Widget build(BuildContext context) {
    const threads = FieldMock.threads;
    return ListView(
      padding: const EdgeInsets.fromLTRB(16, 4, 16, 96),
      children: [
        Row(
          children: [
            const Text('Inbox', style: kH1),
            const SizedBox(width: 10),
            const DemoBadge(),
            const Spacer(),
            _bellButton(context),
          ],
        ),
        const SizedBox(height: 14),
        for (final t in threads) ...[
          _ThreadRow(thread: t),
          const SizedBox(height: 8),
        ],
      ],
    );
  }

  Widget _bellButton(BuildContext context) {
    return InkWell(
      borderRadius: BorderRadius.circular(10),
      onTap: () => Navigator.of(context).push(
        MaterialPageRoute(builder: (_) => const NotificationsScreen()),
      ),
      child: Container(
        width: 38,
        height: 38,
        decoration: cardDecoration(radius: 10),
        child: Stack(
          alignment: Alignment.center,
          children: [
            const Icon(Icons.notifications_none, size: 18, color: AppColors.grey700),
            if (FieldMock.unreadNotifications > 0)
              Positioned(
                top: 8,
                right: 8,
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
      ),
    );
  }
}

class _ThreadRow extends StatelessWidget {
  final MockThread thread;
  const _ThreadRow({required this.thread});

  @override
  Widget build(BuildContext context) {
    return InkWell(
      borderRadius: BorderRadius.circular(14),
      onTap: () => Navigator.of(context).push(
        MaterialPageRoute(builder: (_) => ThreadScreen(thread: thread)),
      ),
      child: Container(
        padding: const EdgeInsets.all(13),
        decoration: cardDecoration(),
        child: Row(
          children: [
            InitialsAvatar(initials: thread.initials, color: thread.color, size: 44),
            const SizedBox(width: 12),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(
                    children: [
                      Expanded(
                        child: Text(thread.name,
                            maxLines: 1,
                            overflow: TextOverflow.ellipsis,
                            style: const TextStyle(
                                fontSize: 14.5, fontWeight: FontWeight.w700, color: AppColors.ink)),
                      ),
                      Text(thread.time, style: const TextStyle(fontSize: 11.5, color: AppColors.grey400)),
                    ],
                  ),
                  const SizedBox(height: 2),
                  Row(
                    children: [
                      Expanded(
                        child: Text(thread.last,
                            maxLines: 1,
                            overflow: TextOverflow.ellipsis,
                            style: TextStyle(
                                fontSize: 13,
                                fontWeight: thread.unread > 0 ? FontWeight.w600 : FontWeight.w400,
                                color: thread.unread > 0 ? AppColors.grey700 : AppColors.grey400)),
                      ),
                      if (thread.unread > 0) ...[
                        const SizedBox(width: 8),
                        Container(
                          constraints: const BoxConstraints(minWidth: 18),
                          height: 18,
                          alignment: Alignment.center,
                          decoration: BoxDecoration(
                            color: AppColors.orange,
                            borderRadius: BorderRadius.circular(9999),
                          ),
                          child: Text('${thread.unread}',
                              style: const TextStyle(
                                  fontSize: 10, fontWeight: FontWeight.w700, color: Colors.white)),
                        ),
                      ],
                    ],
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }
}
