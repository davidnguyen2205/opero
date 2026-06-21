import 'package:geolocator/geolocator.dart';
import 'package:image_picker/image_picker.dart';

/// Result of capturing field context at check-in/out: geolocation and an
/// optional local photo path.
class Capture {
  final double? lat;
  final double? lng;

  /// Local file path of a captured photo, if any. NOT uploaded — v1 has no blob
  /// storage, so this is for on-device preview only and is not sent as
  /// photo_url. When storage exists, upload this file and send the returned URL.
  final String? photoPath;

  Capture({this.lat, this.lng, this.photoPath});
}

/// Acquires the current position, requesting permission as needed. Returns
/// null lat/lng (rather than throwing) if location is unavailable/denied, so
/// attendance still records — geolocation is best-effort context, not a gate.
///
/// API CAVEAT (unverified): geolocator's permission + getCurrentPosition API
/// shifts across major versions (LocationSettings, etc.). Verify against the
/// installed version; see SHIPPING.md.
Future<Capture> captureContext({bool withPhoto = false}) async {
  double? lat, lng;
  try {
    final serviceOn = await Geolocator.isLocationServiceEnabled();
    if (serviceOn) {
      var perm = await Geolocator.checkPermission();
      if (perm == LocationPermission.denied) {
        perm = await Geolocator.requestPermission();
      }
      if (perm == LocationPermission.always || perm == LocationPermission.whileInUse) {
        final pos = await Geolocator.getCurrentPosition();
        lat = pos.latitude;
        lng = pos.longitude;
      }
    }
  } catch (_) {
    // best-effort: leave lat/lng null
  }

  String? photoPath;
  if (withPhoto) {
    try {
      final x = await ImagePicker().pickImage(source: ImageSource.camera, maxWidth: 1280);
      photoPath = x?.path;
    } catch (_) {
      // best-effort: no photo
    }
  }

  return Capture(lat: lat, lng: lng, photoPath: photoPath);
}
