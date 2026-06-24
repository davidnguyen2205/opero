import 'package:flutter/material.dart';
import 'package:shared_preferences/shared_preferences.dart';

import 'api/api_client.dart';
import 'api/auth_store.dart';
import 'attendance/attendance_service.dart';
import 'offline/queue.dart';
import 'screens/login_screen.dart';
import 'screens/onboarding_screen.dart';
import 'screens/shell.dart';
import 'theme.dart';

Future<void> main() async {
  WidgetsFlutterBinding.ensureInitialized();

  final auth = AuthStore();
  await auth.load();

  final api = ApiClient(auth);
  final attendance = AttendanceService(api: api, queue: OfflineQueue());
  await attendance.init();

  runApp(OperoApp(auth: auth, api: api, attendance: attendance));
}

class OperoApp extends StatefulWidget {
  final AuthStore auth;
  final ApiClient api;
  final AttendanceService attendance;

  const OperoApp({super.key, required this.auth, required this.api, required this.attendance});

  @override
  State<OperoApp> createState() => _OperoAppState();
}

class _OperoAppState extends State<OperoApp> {
  static const _seenOnboardingKey = 'opero.onboarding.seen';

  late bool _authed = widget.auth.isAuthenticated;
  bool _seenOnboarding = true;

  @override
  void initState() {
    super.initState();
    _loadOnboardingFlag();
  }

  Future<void> _loadOnboardingFlag() async {
    final prefs = await SharedPreferences.getInstance();
    if (mounted) {
      setState(() => _seenOnboarding = prefs.getBool(_seenOnboardingKey) ?? false);
    }
  }

  void _onAuthenticated() => setState(() => _authed = widget.auth.isAuthenticated);

  void _onSignedOut() => setState(() => _authed = widget.auth.isAuthenticated);

  Future<void> _finishOnboarding() async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.setBool(_seenOnboardingKey, true);
    if (mounted) setState(() => _seenOnboarding = true);
  }

  @override
  Widget build(BuildContext context) {
    final Widget home;
    if (!_authed) {
      home = LoginScreen(api: widget.api, auth: widget.auth, onAuthenticated: _onAuthenticated);
    } else if (!_seenOnboarding) {
      home = OnboardingScreen(onDone: _finishOnboarding);
    } else {
      home = Shell(
        api: widget.api,
        auth: widget.auth,
        attendance: widget.attendance,
        onSignedOut: _onSignedOut,
      );
    }

    return MaterialApp(
      title: 'Opero',
      debugShowCheckedModeBanner: false,
      theme: buildOperoTheme(),
      home: home,
    );
  }
}
