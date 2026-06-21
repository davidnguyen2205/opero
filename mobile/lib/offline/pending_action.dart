/// A queued attendance mutation awaiting delivery to the server. Persisted so
/// it survives app restarts; replayed by the sync loop. The server is
/// idempotent on `clientId`, so replaying any action any number of times is
/// safe — this is what makes the offline queue correct.
enum ActionType { checkIn, checkOut }

class PendingAction {
  /// Local queue-entry id (NOT the attendance client_id). Used to dedupe/remove
  /// within the local queue.
  final String id;
  final ActionType type;

  /// The attendance idempotency key. A check-in and its matching check-out share
  /// the SAME clientId (per the API contract).
  final String clientId;

  final String? shiftId;
  final double? lat;
  final double? lng;
  final String? photoUrl;
  final String createdAt; // ISO-8601, for ordering/debugging

  PendingAction({
    required this.id,
    required this.type,
    required this.clientId,
    this.shiftId,
    this.lat,
    this.lng,
    this.photoUrl,
    required this.createdAt,
  });

  Map<String, dynamic> toJson() => {
        'id': id,
        'type': type.name,
        'client_id': clientId,
        'shift_id': shiftId,
        'lat': lat,
        'lng': lng,
        'photo_url': photoUrl,
        'created_at': createdAt,
      };

  factory PendingAction.fromJson(Map<String, dynamic> j) => PendingAction(
        id: j['id'] as String,
        type: ActionType.values.firstWhere((t) => t.name == j['type']),
        clientId: j['client_id'] as String,
        shiftId: j['shift_id'] as String?,
        lat: (j['lat'] as num?)?.toDouble(),
        lng: (j['lng'] as num?)?.toDouble(),
        photoUrl: j['photo_url'] as String?,
        createdAt: j['created_at'] as String,
      );
}
