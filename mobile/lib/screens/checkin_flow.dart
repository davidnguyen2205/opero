import 'dart:async';
import 'dart:io';

import 'package:flutter/material.dart';

import '../attendance/attendance_service.dart';
import '../attendance/capture.dart';
import '../theme.dart';
import 'shift_view.dart';

/// The check-in/out flow as a full-screen step machine, matching the prototype:
/// locating → photo → confirm → active → done. Drives the REAL
/// [AttendanceService] (offline-tolerant, single client_id per check-in;
/// check-out reuses it). Geolocation is captured via [captureContext]; the photo
/// is captured and uploaded via `POST /media`, and the returned URL is sent as
/// the check-in/out photo_url.
///
/// The "within 25 m" copy is cosmetic — the server does no geofencing. "Break"
/// is REAL: it toggles checked_in ⇄ on_break via `POST /attendance/break`
/// against the active attendance's client_id.
enum CheckInStep { locating, photo, confirm, active, done }

class CheckInFlow extends StatefulWidget {
  final ShiftView shift;
  final AttendanceService attendance;

  /// Whether the user is already checked in (resuming an open attendance) — if
  /// so we jump straight to the active screen.
  final bool alreadyActive;

  const CheckInFlow({
    super.key,
    required this.shift,
    required this.attendance,
    this.alreadyActive = false,
  });

  @override
  State<CheckInFlow> createState() => _CheckInFlowState();
}

class _CheckInFlowState extends State<CheckInFlow> {
  late CheckInStep _step = widget.alreadyActive ? CheckInStep.active : CheckInStep.locating;

  Capture? _capture;
  bool _capturing = false;
  bool _photoCaptured = false;
  bool _uploadingPhoto = false;
  String? _photoUrl; // URL from POST /media, sent as photo_url
  bool _onBreak = false;
  bool _breakBusy = false;
  DateTime? _checkedInAt;
  Timer? _ticker;
  Duration _elapsed = Duration.zero;

  @override
  void initState() {
    super.initState();
    if (widget.alreadyActive) {
      _checkedInAt = DateTime.now();
      _startTicker();
    } else {
      _beginLocating();
    }
  }

  @override
  void dispose() {
    _ticker?.cancel();
    super.dispose();
  }

  ShiftView get shift => widget.shift;

  Future<void> _beginLocating() async {
    setState(() => _capturing = true);
    final cap = await captureContext();
    if (!mounted) return;
    setState(() {
      _capture = cap;
      _capturing = false;
    });
  }

  Future<void> _takePhoto() async {
    setState(() => _capturing = true);
    final cap = await captureContext(withPhoto: true);
    if (!mounted) return;
    setState(() {
      // Keep any location from the locating step; merge in the photo.
      _capture = Capture(
        lat: cap.lat ?? _capture?.lat,
        lng: cap.lng ?? _capture?.lng,
        photoPath: cap.photoPath,
      );
      _photoCaptured = cap.photoPath != null;
      _capturing = false;
      _step = CheckInStep.confirm;
    });
    // Upload the captured photo to object storage (POST /media). Best-effort:
    // if it fails (offline / storage hiccup) we still let the check-in proceed
    // without a photo_url rather than block attendance.
    final path = cap.photoPath;
    if (path != null) {
      setState(() => _uploadingPhoto = true);
      try {
        final media = await widget.attendance.api.uploadMedia(file: File(path));
        if (mounted) setState(() => _photoUrl = media.url);
      } catch (_) {
        // leave _photoUrl null; check-in continues without it
      } finally {
        if (mounted) setState(() => _uploadingPhoto = false);
      }
    }
  }

  Future<void> _confirmCheckIn() async {
    // Enqueue the REAL check-in (durable, offline-tolerant). photo_url is the
    // URL returned by POST /media for the captured photo (null if none/upload
    // failed — attendance is never blocked on the photo).
    await widget.attendance.checkIn(
      shiftId: shift.id,
      lat: _capture?.lat,
      lng: _capture?.lng,
      photoUrl: _photoUrl,
    );
    if (!mounted) return;
    setState(() {
      _checkedInAt = DateTime.now();
      _elapsed = Duration.zero;
      _step = CheckInStep.active;
    });
    _startTicker();
  }

  Future<void> _toggleBreak() async {
    final next = !_onBreak;
    setState(() => _breakBusy = true);
    try {
      await widget.attendance.setBreak(shiftId: shift.id, onBreak: next);
      if (!mounted) return;
      setState(() => _onBreak = next);
    } catch (_) {
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Couldn\'t update break — check your connection.')),
      );
    } finally {
      if (mounted) setState(() => _breakBusy = false);
    }
  }

  Future<void> _checkOut() async {
    final cap = await captureContext();
    await widget.attendance.checkOut(shiftId: shift.id, lat: cap.lat, lng: cap.lng);
    _ticker?.cancel();
    if (!mounted) return;
    setState(() => _step = CheckInStep.done);
  }

  void _startTicker() {
    _ticker?.cancel();
    _ticker = Timer.periodic(const Duration(seconds: 1), (_) {
      if (_checkedInAt == null) return;
      setState(() => _elapsed = DateTime.now().difference(_checkedInAt!));
    });
  }

  String _fmtElapsed(Duration d) {
    final m = d.inMinutes.remainder(60).toString().padLeft(2, '0');
    final s = d.inSeconds.remainder(60).toString().padLeft(2, '0');
    final h = d.inHours;
    return h > 0 ? '$h:$m:$s' : '$m:$s';
  }

  String _nowHm() {
    final t = _checkedInAt ?? DateTime.now();
    return '${t.hour.toString().padLeft(2, '0')}:${t.minute.toString().padLeft(2, '0')}';
  }

  @override
  Widget build(BuildContext context) {
    final title = switch (_step) {
      CheckInStep.locating => 'Confirm location',
      CheckInStep.photo => 'Check-in photo',
      CheckInStep.confirm => 'Confirm check-in',
      CheckInStep.active => 'On shift',
      CheckInStep.done => 'Shift complete',
    };
    return Scaffold(
      appBar: AppBar(
        title: Text(title, style: kCardTitle),
        leading: _step == CheckInStep.done
            ? null
            : IconButton(
                icon: const Icon(Icons.chevron_left, color: AppColors.grey700),
                onPressed: () => Navigator.of(context).maybePop(),
              ),
        automaticallyImplyLeading: false,
      ),
      body: SafeArea(
        child: Padding(
          padding: const EdgeInsets.fromLTRB(16, 8, 16, 24),
          child: switch (_step) {
            CheckInStep.locating => _locating(),
            CheckInStep.photo => _photo(),
            CheckInStep.confirm => _confirm(),
            CheckInStep.active => _active(),
            CheckInStep.done => _done(),
          },
        ),
      ),
    );
  }

  // ── LOCATING ───────────────────────────────────────────────────────────
  Widget _locating() {
    final located = !_capturing;
    final ok = _capture?.lat != null;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        _miniMap(ok: ok),
        const SizedBox(height: 14),
        Container(
          padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 12),
          decoration: BoxDecoration(
            color: ok ? AppColors.green50 : AppColors.orange50,
            borderRadius: BorderRadius.circular(12),
            border: Border.all(color: ok ? AppColors.green200 : AppColors.orange200),
          ),
          child: Row(
            children: [
              Icon(_capturing ? Icons.location_searching : Icons.check,
                  size: 18, color: ok ? AppColors.green : AppColors.orange),
              const SizedBox(width: 9),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      _capturing
                          ? 'Finding your location…'
                          : (ok ? 'You\'re at ${shift.title}' : 'Location unavailable'),
                      style: TextStyle(
                          fontSize: 14,
                          fontWeight: FontWeight.w600,
                          color: ok ? AppColors.green700 : AppColors.orange700),
                    ),
                    if (located)
                      Text(
                        ok ? 'Within 25 m of the tour start point' : 'You can still check in.',
                        style: const TextStyle(fontSize: 12, color: AppColors.grey500),
                      ),
                  ],
                ),
              ),
            ],
          ),
        ),
        const Spacer(),
        PhoneButton(
          label: 'Continue to Photo',
          icon: Icons.photo_camera,
          onPressed: _capturing ? null : () => setState(() => _step = CheckInStep.photo),
        ),
      ],
    );
  }

  // ── PHOTO ──────────────────────────────────────────────────────────────
  Widget _photo() {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        const Text('Take a check-in photo', style: kH2),
        const SizedBox(height: 6),
        const Text('A quick shot of the meeting point confirms you\'re on site.',
            style: TextStyle(fontSize: 13, color: AppColors.grey500)),
        const SizedBox(height: 14),
        Container(
          height: 250,
          decoration: BoxDecoration(
            color: AppColors.grey100,
            borderRadius: BorderRadius.circular(16),
            border: Border.all(color: AppColors.grey200),
          ),
          alignment: Alignment.center,
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Icon(_photoCaptured ? Icons.check_circle : Icons.photo_camera,
                  size: 34, color: _photoCaptured ? AppColors.green : AppColors.grey400),
              const SizedBox(height: 8),
              Text(_photoCaptured ? 'Photo captured' : 'Camera viewfinder',
                  style: const TextStyle(fontSize: 12, color: AppColors.grey400)),
            ],
          ),
        ),
        const Spacer(),
        PhoneButton(
          label: _capturing ? 'Opening camera…' : 'Capture',
          icon: Icons.photo_camera,
          onPressed: _capturing ? null : _takePhoto,
        ),
      ],
    );
  }

  // ── CONFIRM ────────────────────────────────────────────────────────────
  Widget _confirm() {
    final ok = _capture?.lat != null;
    return AnimatedBuilder(
      animation: widget.attendance,
      builder: (context, _) {
        return Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Container(
              decoration: cardDecoration(),
              child: Column(
                children: [
                  _confirmRow(Icons.place, 'Location',
                      ok ? '${shift.title} ✓' : 'Not captured', divider: true),
                  _confirmRow(
                      Icons.photo_camera,
                      'Photo',
                      !_photoCaptured
                          ? 'Skipped'
                          : _uploadingPhoto
                              ? 'Uploading…'
                              : _photoUrl != null
                                  ? 'Attached'
                                  : 'Captured',
                      divider: true),
                  _confirmRow(Icons.schedule, 'Time', _nowHm(), divider: false),
                ],
              ),
            ),
            const SizedBox(height: 14),
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
              decoration: BoxDecoration(
                color: AppColors.amber50,
                borderRadius: BorderRadius.circular(10),
                border: Border.all(color: AppColors.amber200),
              ),
              child: const Row(
                children: [
                  Icon(Icons.wifi_off, size: 15, color: AppColors.amber700),
                  SizedBox(width: 8),
                  Expanded(
                    child: Text(
                      'If you\'re offline this is saved on your phone and syncs automatically.',
                      style: TextStyle(fontSize: 12.5, color: AppColors.amber700),
                    ),
                  ),
                ],
              ),
            ),
            const Spacer(),
            PhoneButton(label: 'Confirm Check-in', icon: Icons.check, onPressed: _confirmCheckIn),
          ],
        );
      },
    );
  }

  Widget _confirmRow(IconData icon, String k, String v, {required bool divider}) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 12),
      decoration: BoxDecoration(
        border: divider
            ? const Border(bottom: BorderSide(color: AppColors.grey100))
            : null,
      ),
      child: Row(
        children: [
          Icon(icon, size: 17, color: AppColors.grey400),
          const SizedBox(width: 10),
          Text(k, style: const TextStyle(fontSize: 14, color: AppColors.grey500)),
          const Spacer(),
          Text(v, style: const TextStyle(fontSize: 14, fontWeight: FontWeight.w600, color: AppColors.grey900)),
        ],
      ),
    );
  }

  // ── ACTIVE ─────────────────────────────────────────────────────────────
  Widget _active() {
    return AnimatedBuilder(
      animation: widget.attendance,
      builder: (context, _) {
        final queued = widget.attendance.pendingCount;
        return Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 18, vertical: 20),
              decoration: BoxDecoration(
                gradient: const LinearGradient(
                  begin: Alignment.topCenter,
                  end: Alignment.bottomCenter,
                  colors: [Color(0xFF16A34A), Color(0xFF15803D)],
                ),
                borderRadius: BorderRadius.circular(16),
              ),
              child: Column(
                children: [
                  Row(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      Container(
                        width: 8,
                        height: 8,
                        decoration: const BoxDecoration(color: Colors.white, shape: BoxShape.circle),
                      ),
                      const SizedBox(width: 7),
                      Text(_onBreak ? 'On break' : 'On shift',
                          style: const TextStyle(fontSize: 13, fontWeight: FontWeight.w600, color: Colors.white)),
                    ],
                  ),
                  const SizedBox(height: 6),
                  Text(_fmtElapsed(_elapsed),
                      style: const TextStyle(
                          fontSize: 44,
                          fontWeight: FontWeight.w700,
                          color: Colors.white,
                          fontFeatures: [FontFeature.tabularFigures()])),
                  Text('Checked in ${_nowHm()} · ${shift.title}',
                      style: const TextStyle(fontSize: 13, color: Colors.white70)),
                  if (queued > 0) ...[
                    const SizedBox(height: 10),
                    Container(
                      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 3),
                      decoration: BoxDecoration(
                        color: Colors.white24,
                        borderRadius: BorderRadius.circular(9999),
                      ),
                      child: const Row(
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          Icon(Icons.wifi_off, size: 13, color: Colors.white),
                          SizedBox(width: 6),
                          Text('Saved · will sync',
                              style: TextStyle(fontSize: 12, fontWeight: FontWeight.w600, color: Colors.white)),
                        ],
                      ),
                    ),
                  ],
                ],
              ),
            ),
            const SizedBox(height: 14),
            _shiftInfoCard(),
            const Spacer(),
            Row(
              children: [
                Expanded(
                  child: PhoneButton(
                    label: _breakBusy ? '…' : (_onBreak ? 'Resume' : 'Break'),
                    icon: Icons.free_breakfast,
                    tone: PhoneButtonTone.ghost,
                    // REAL: toggles checked_in ⇄ on_break via POST /attendance/break
                    // against the active attendance's client_id.
                    onPressed: _breakBusy ? null : _toggleBreak,
                  ),
                ),
                const SizedBox(width: 10),
                Expanded(
                  child: PhoneButton(
                    label: 'Check Out',
                    icon: Icons.check,
                    tone: PhoneButtonTone.danger,
                    onPressed: _checkOut,
                  ),
                ),
              ],
            ),
          ],
        );
      },
    );
  }

  // ── DONE ───────────────────────────────────────────────────────────────
  Widget _done() {
    return Column(
      children: [
        const SizedBox(height: 32),
        Container(
          width: 76,
          height: 76,
          decoration: const BoxDecoration(color: AppColors.blue50, shape: BoxShape.circle),
          child: const Icon(Icons.check, size: 38, color: AppColors.blue),
        ),
        const SizedBox(height: 18),
        const Text('Shift complete', style: kH1),
        const SizedBox(height: 8),
        Text('You checked out at ${_nowHm()}. Nice work on the ${shift.title}.',
            textAlign: TextAlign.center,
            style: const TextStyle(fontSize: 14, color: AppColors.grey500)),
        const SizedBox(height: 18),
        Container(
          width: double.infinity,
          padding: const EdgeInsets.all(16),
          decoration: cardDecoration(),
          child: Column(
            children: [
              _doneRow('Worked', _fmtElapsed(_elapsed)),
              _doneRow('Location', shift.title),
            ],
          ),
        ),
        const Spacer(),
        PhoneButton(
          label: 'Back to Today',
          icon: Icons.chevron_left,
          tone: PhoneButtonTone.light,
          onPressed: () => Navigator.of(context).pop(true),
        ),
      ],
    );
  }

  Widget _doneRow(String k, String v) => Padding(
        padding: const EdgeInsets.symmetric(vertical: 4),
        child: Row(
          children: [
            Text(k, style: const TextStyle(fontSize: 14, color: AppColors.grey500)),
            const Spacer(),
            Text(v, style: const TextStyle(fontSize: 14, fontWeight: FontWeight.w600, color: AppColors.grey900)),
          ],
        ),
      );

  // ── shared bits ──────────────────────────────────────────────────────────
  Widget _shiftInfoCard() {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: cardDecoration(),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Container(
                width: 10,
                height: 10,
                decoration: BoxDecoration(color: shift.color, borderRadius: BorderRadius.circular(3)),
              ),
              const SizedBox(width: 8),
              Expanded(child: Text(shift.title, style: kCardTitle)),
            ],
          ),
          const SizedBox(height: 10),
          _infoRow(Icons.schedule, shift.timeRange),
          if (shift.address != null) _infoRow(Icons.place, shift.address!),
        ],
      ),
    );
  }

  Widget _infoRow(IconData icon, String text) => Padding(
        padding: const EdgeInsets.symmetric(vertical: 6),
        child: Row(
          children: [
            Icon(icon, size: 17, color: AppColors.grey400),
            const SizedBox(width: 10),
            Expanded(child: Text(text, style: const TextStyle(fontSize: 14, color: AppColors.grey700))),
          ],
        ),
      );

  Widget _miniMap({required bool ok}) {
    return Container(
      height: 150,
      decoration: BoxDecoration(
        color: AppColors.grey100,
        borderRadius: BorderRadius.circular(13),
        border: Border.all(color: AppColors.grey200),
      ),
      child: Center(
        child: _capturing
            ? const CircularProgressIndicator(color: AppColors.orange)
            : Container(
                width: 54,
                height: 54,
                decoration: BoxDecoration(
                  color: (ok ? AppColors.green : AppColors.orange).withValues(alpha: 0.16),
                  shape: BoxShape.circle,
                ),
                child: Center(
                  child: Container(
                    width: 16,
                    height: 16,
                    decoration: BoxDecoration(
                      color: ok ? AppColors.green : AppColors.orange,
                      shape: BoxShape.circle,
                      border: Border.all(color: Colors.white, width: 3),
                    ),
                  ),
                ),
              ),
      ),
    );
  }
}
