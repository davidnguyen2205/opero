import 'package:flutter/material.dart';

import 'api/api_client.dart';
import 'api/auth_store.dart';
import 'attendance/attendance_service.dart';
import 'offline/queue.dart';
import 'screens/login_screen.dart';
import 'screens/shifts_screen.dart';

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
  late bool _authed = widget.auth.isAuthenticated;

  void _refreshAuth() => setState(() => _authed = widget.auth.isAuthenticated);

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'Opero',
      theme: ThemeData(colorSchemeSeed: Colors.indigo, useMaterial3: true),
      home: _authed
          ? ShiftsScreen(
              api: widget.api,
              auth: widget.auth,
              attendance: widget.attendance,
              onSignedOut: _refreshAuth,
            )
          : LoginScreen(
              api: widget.api,
              auth: widget.auth,
              onAuthenticated: _refreshAuth,
            ),
    );
  }
}
