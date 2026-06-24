import 'package:flutter/material.dart';
import 'package:geolocator/geolocator.dart';

import '../theme.dart';

/// Client-only 4-slide onboarding carousel with cosmetic-ish permission toggles
/// (location actually prompts via geolocator; notifications is cosmetic since we
/// have no push wiring in v1). Shown once after first login — the caller
/// persists the "seen onboarding" flag.
class OnboardingScreen extends StatefulWidget {
  final VoidCallback onDone;
  const OnboardingScreen({super.key, required this.onDone});

  @override
  State<OnboardingScreen> createState() => _OnboardingScreenState();
}

class _Slide {
  final IconData icon;
  final String title;
  final String body;
  final String? permKey;
  final String? permLabel;
  const _Slide(this.icon, this.title, this.body, {this.permKey, this.permLabel});
}

class _OnboardingScreenState extends State<OnboardingScreen> {
  int _i = 0;
  final _perms = <String, bool>{};

  static const _slides = [
    _Slide(Icons.calendar_today, 'Your shifts, always in your pocket',
        'See your weekly roster the moment your manager publishes it — with times and meeting points.'),
    _Slide(Icons.place, 'Check in from the field',
        'Confirm you\'re on site with one tap. Location and a quick photo prove you\'re at the meeting point.'),
    _Slide(Icons.wifi_off, 'Works without signal',
        'No reception at the castle? Check-ins are saved on your phone and sync automatically when you\'re back online.',
        permKey: 'location', permLabel: 'Allow location while using the app'),
    _Slide(Icons.notifications, 'Stay in the loop',
        'Get notified about new rosters, shift changes and messages from dispatch.',
        permKey: 'notifs', permLabel: 'Allow notifications'),
  ];

  bool get _last => _i == _slides.length - 1;

  Future<void> _togglePerm(_Slide s) async {
    final key = s.permKey!;
    final next = !(_perms[key] ?? false);
    if (next && key == 'location') {
      // Real prompt for location; notifications stays cosmetic (no push in v1).
      try {
        var perm = await Geolocator.checkPermission();
        if (perm == LocationPermission.denied) {
          perm = await Geolocator.requestPermission();
        }
      } catch (_) {
        // ignore — toggle remains a visual hint
      }
    }
    setState(() => _perms[key] = next);
  }

  @override
  Widget build(BuildContext context) {
    final s = _slides[_i];
    return Scaffold(
      backgroundColor: Colors.white,
      body: SafeArea(
        child: Padding(
          padding: const EdgeInsets.symmetric(horizontal: 24),
          child: Column(
            children: [
              Align(
                alignment: Alignment.centerRight,
                child: TextButton(
                  onPressed: widget.onDone,
                  child: const Text('Skip',
                      style: TextStyle(color: AppColors.grey400, fontWeight: FontWeight.w600)),
                ),
              ),
              Expanded(
                child: Column(
                  mainAxisAlignment: MainAxisAlignment.center,
                  children: [
                    Container(
                      width: 96,
                      height: 96,
                      decoration: BoxDecoration(
                        color: AppColors.orange50,
                        borderRadius: BorderRadius.circular(24),
                      ),
                      child: Icon(s.icon, size: 44, color: AppColors.orange),
                    ),
                    const SizedBox(height: 22),
                    Text(s.title,
                        textAlign: TextAlign.center,
                        style: const TextStyle(
                            fontSize: 23, fontWeight: FontWeight.w700, color: AppColors.ink, height: 1.2)),
                    const SizedBox(height: 14),
                    SizedBox(
                      width: 290,
                      child: Text(s.body,
                          textAlign: TextAlign.center,
                          style: const TextStyle(fontSize: 15, color: AppColors.grey500, height: 1.5)),
                    ),
                    if (s.permKey != null) ...[
                      const SizedBox(height: 22),
                      _permToggle(s),
                    ],
                  ],
                ),
              ),
              _dots(),
              const SizedBox(height: 20),
              PhoneButton(
                label: _last ? 'Get Started' : 'Next',
                onPressed: () => _last ? widget.onDone() : setState(() => _i++),
              ),
              const SizedBox(height: 24),
            ],
          ),
        ),
      ),
    );
  }

  Widget _permToggle(_Slide s) {
    final on = _perms[s.permKey] ?? false;
    return InkWell(
      borderRadius: BorderRadius.circular(12),
      onTap: () => _togglePerm(s),
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
        decoration: BoxDecoration(
          color: on ? AppColors.green50 : Colors.white,
          borderRadius: BorderRadius.circular(12),
          border: Border.all(color: on ? AppColors.green200 : AppColors.grey200),
        ),
        child: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            Container(
              width: 22,
              height: 22,
              decoration: BoxDecoration(
                color: on ? AppColors.green : Colors.white,
                borderRadius: BorderRadius.circular(6),
                border: on ? null : Border.all(color: AppColors.grey300, width: 1.5),
              ),
              child: on ? const Icon(Icons.check, size: 14, color: Colors.white) : null,
            ),
            const SizedBox(width: 10),
            Text(s.permLabel!,
                style: TextStyle(
                    fontSize: 13.5,
                    fontWeight: FontWeight.w600,
                    color: on ? AppColors.green700 : AppColors.grey700)),
          ],
        ),
      ),
    );
  }

  Widget _dots() {
    return Row(
      mainAxisAlignment: MainAxisAlignment.center,
      children: [
        for (var j = 0; j < _slides.length; j++)
          AnimatedContainer(
            duration: const Duration(milliseconds: 200),
            margin: const EdgeInsets.symmetric(horizontal: 3.5),
            width: j == _i ? 22 : 7,
            height: 7,
            decoration: BoxDecoration(
              color: j == _i ? AppColors.orange : AppColors.grey200,
              borderRadius: BorderRadius.circular(9999),
            ),
          ),
      ],
    );
  }
}
