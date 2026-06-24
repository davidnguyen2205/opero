/// DEMO-ONLY mock data for the field app.
///
/// IMPORTANT: Everything in this file is static, local, hand-written demo data.
/// There is NO backend API for messaging, notifications, leave/time-off, or
/// profile analytics in Opero v1. Screens that use this data are clearly labeled
/// MOCK in the UI (a small "Demo" badge) and in SHIPPING.md. Do NOT treat any of
/// this as real, persisted, or synced. When the corresponding APIs exist, delete
/// this file and wire the screens to the real client.
library;

import 'package:flutter/material.dart';

import '../theme.dart';

class MockMessage {
  final bool fromMe;
  final String text;
  final String time;
  const MockMessage({required this.fromMe, required this.text, required this.time});
}

class MockThread {
  final String id;
  final String name;
  final String initials;
  final Color color;
  final String last;
  final String time;
  final int unread;
  final List<MockMessage> messages;
  const MockThread({
    required this.id,
    required this.name,
    required this.initials,
    required this.color,
    required this.last,
    required this.time,
    required this.unread,
    required this.messages,
  });
}

class MockNotification {
  final IconData icon;
  final Color tone;
  final String title;
  final String body;
  final String time;
  final bool unread;
  const MockNotification({
    required this.icon,
    required this.tone,
    required this.title,
    required this.body,
    required this.time,
    required this.unread,
  });
}

class MockTimeOffRequest {
  final String range;
  final String days;
  final String status; // Approved | Pending
  const MockTimeOffRequest({required this.range, required this.days, required this.status});
}

class MockProfileStats {
  final int onTimePct;
  final int hoursThisWeek;
  final int toursThisMonth;
  final String tenure;
  final int daysUsed;
  final int daysTotal;
  final List<String> languages;
  final List<MockTimeOffRequest> requests;
  final String phone;
  final String employeeId;
  final String managerName;
  const MockProfileStats({
    required this.onTimePct,
    required this.hoursThisWeek,
    required this.toursThisMonth,
    required this.tenure,
    required this.daysUsed,
    required this.daysTotal,
    required this.languages,
    required this.requests,
    required this.phone,
    required this.employeeId,
    required this.managerName,
  });
}

/// All demo data lives behind this single accessor.
class FieldMock {
  FieldMock._();

  static const List<MockThread> threads = [
    MockThread(
      id: 'dispatch',
      name: 'Dispatch',
      initials: 'DP',
      color: AppColors.orange,
      last: 'Your Alfama group is confirmed — 8 guests.',
      time: '11:20',
      unread: 1,
      messages: [
        MockMessage(fromMe: false, text: 'Morning! Alfama tour at 12:00 is confirmed.', time: '11:18'),
        MockMessage(
            fromMe: false,
            text: 'Your Alfama group is confirmed — 8 guests. 2 kids in the party.',
            time: '11:20'),
      ],
    ),
    MockThread(
      id: 'helena',
      name: 'Helena Bastos',
      initials: 'HB',
      color: Color(0xFF4B5563),
      last: 'Thanks for covering Friday',
      time: 'Yesterday',
      unread: 0,
      messages: [
        MockMessage(fromMe: false, text: 'Can you take the Belém slot on Friday?', time: 'Yesterday 16:02'),
        MockMessage(fromMe: true, text: 'Yes, happy to.', time: 'Yesterday 16:10'),
        MockMessage(fromMe: false, text: 'Thanks for covering Friday', time: 'Yesterday 16:11'),
      ],
    ),
    MockThread(
      id: 'team',
      name: 'Guides — Lisbon',
      initials: 'GL',
      color: Color(0xFF7C3AED),
      last: 'Diogo: anyone have a spare audio set?',
      time: 'Yesterday',
      unread: 0,
      messages: [
        MockMessage(fromMe: false, text: 'Diogo: anyone have a spare audio set?', time: 'Yesterday 09:31'),
      ],
    ),
  ];

  static int get unreadThreads => threads.fold(0, (n, t) => n + (t.unread > 0 ? 1 : 0));

  static const List<MockNotification> notifications = [
    MockNotification(
        icon: Icons.calendar_today,
        tone: AppColors.orange,
        title: 'New roster published',
        body: 'Your shifts for next week are ready.',
        time: '08:02',
        unread: true),
    MockNotification(
        icon: Icons.check,
        tone: AppColors.green,
        title: 'Time off approved',
        body: '25 Jun · 1 day — approved by Helena.',
        time: 'Yesterday',
        unread: false),
    MockNotification(
        icon: Icons.place,
        tone: AppColors.blue,
        title: 'Meeting point updated',
        body: 'Alfama tour now starts at Miradouro de Santa Luzia.',
        time: 'Yesterday',
        unread: false),
    MockNotification(
        icon: Icons.notifications,
        tone: AppColors.amber,
        title: 'Shift reminder',
        body: 'Alfama Walking Tour starts at 12:00.',
        time: '2 days ago',
        unread: false),
  ];

  static int get unreadNotifications => notifications.where((n) => n.unread).length;

  static const MockProfileStats profile = MockProfileStats(
    onTimePct: 96,
    hoursThisWeek: 29,
    toursThisMonth: 19,
    tenure: '6 yrs',
    daysUsed: 8,
    daysTotal: 22,
    languages: ['PT', 'EN', 'ES'],
    requests: [
      MockTimeOffRequest(range: '25 Jun', days: '1 day', status: 'Approved'),
      MockTimeOffRequest(range: '14–18 Jul', days: '5 days', status: 'Pending'),
    ],
    phone: '+351 912 004 111',
    employeeId: 'TT-1002',
    managerName: 'Helena Bastos',
  );
}
