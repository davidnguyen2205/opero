import 'package:flutter/material.dart';

/// Opero field-app theme — ported from the Claude Design prototype.
/// Orange-led, light Material 3 with white cards and grey body text.
class AppColors {
  AppColors._();

  static const orange = Color(0xFFEA580C); // primary
  static const orange50 = Color(0xFFFFF7ED);
  static const orange200 = Color(0xFFFED7AA);
  static const orange700 = Color(0xFFC2410C);

  static const ink = Color(0xFF0C0A09); // near-black headings
  static const grey900 = Color(0xFF1F2937);
  static const grey700 = Color(0xFF374151);
  static const grey500 = Color(0xFF6B7280);
  static const grey400 = Color(0xFF9CA3AF);
  static const grey300 = Color(0xFFD1D5DB);
  static const grey200 = Color(0xFFE5E7EB);
  static const grey100 = Color(0xFFF3F4F6);
  static const grey50 = Color(0xFFF9FAFB);

  static const green = Color(0xFF16A34A);
  static const green700 = Color(0xFF166534);
  static const green50 = Color(0xFFF0FDF4);
  static const green200 = Color(0xFFBBF7D0);

  static const amber = Color(0xFFD97706);
  static const amber700 = Color(0xFFB45309);
  static const amber50 = Color(0xFFFFFBEB);
  static const amber200 = Color(0xFFFDE68A);
  static const amber100 = Color(0xFFFEF3C7);

  static const red = Color(0xFFDC2626);
  static const red50 = Color(0xFFFEF2F2);
  static const red200 = Color(0xFFFECACA);

  static const blue = Color(0xFF2563EB);
  static const blue50 = Color(0xFFDBEAFE);

  /// Stable palette for deriving a per-location accent colour (the API has no
  /// colour field — we hash location_id into one of these).
  static const tourPalette = <Color>[
    Color(0xFFEA580C), // orange
    Color(0xFF0D9488), // teal
    Color(0xFF2563EB), // blue
    Color(0xFF7C3AED), // violet
    Color(0xFFDB2777), // pink
    Color(0xFF16A34A), // green
  ];
}

/// Derive a stable accent colour for a shift from its location id (or shift id
/// when unscheduled). Cosmetic only — keeps the prototype's coloured spine.
Color tourColor(String? seed) {
  if (seed == null || seed.isEmpty) return AppColors.grey400;
  var h = 0;
  for (final c in seed.codeUnits) {
    h = (h * 31 + c) & 0x7fffffff;
  }
  return AppColors.tourPalette[h % AppColors.tourPalette.length];
}

ThemeData buildOperoTheme() {
  final scheme = ColorScheme.fromSeed(
    seedColor: AppColors.orange,
    brightness: Brightness.light,
  ).copyWith(
    primary: AppColors.orange,
    surface: Colors.white,
  );

  return ThemeData(
    useMaterial3: true,
    colorScheme: scheme,
    scaffoldBackgroundColor: Colors.white,
    fontFamily: null,
    appBarTheme: const AppBarTheme(
      backgroundColor: Colors.white,
      foregroundColor: AppColors.ink,
      elevation: 0,
      scrolledUnderElevation: 0,
      centerTitle: false,
    ),
    cardTheme: CardThemeData(
      color: Colors.white,
      elevation: 0,
      margin: EdgeInsets.zero,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(14),
        side: const BorderSide(color: AppColors.grey200),
      ),
    ),
    snackBarTheme: const SnackBarThemeData(behavior: SnackBarBehavior.floating),
  );
}

/// Standard card decoration matching the prototype (white, 1px border, r14).
BoxDecoration cardDecoration({Color? color, Color? border, double radius = 14}) {
  return BoxDecoration(
    color: color ?? Colors.white,
    borderRadius: BorderRadius.circular(radius),
    border: Border.all(color: border ?? AppColors.grey200),
  );
}

/// Full-width 52h pill button matching the prototype's PhoneBtn.
class PhoneButton extends StatelessWidget {
  final String label;
  final IconData? icon;
  final VoidCallback? onPressed;
  final PhoneButtonTone tone;

  const PhoneButton({
    super.key,
    required this.label,
    this.icon,
    this.onPressed,
    this.tone = PhoneButtonTone.primary,
  });

  @override
  Widget build(BuildContext context) {
    final disabled = onPressed == null;
    late Color bg, fg, border;
    switch (tone) {
      case PhoneButtonTone.primary:
        bg = disabled ? AppColors.grey200 : AppColors.orange;
        fg = disabled ? AppColors.grey400 : Colors.white;
        border = bg;
        break;
      case PhoneButtonTone.light:
        bg = Colors.white;
        fg = AppColors.grey900;
        border = AppColors.grey200;
        break;
      case PhoneButtonTone.danger:
        bg = Colors.white;
        fg = AppColors.red;
        border = AppColors.red200;
        break;
      case PhoneButtonTone.ghost:
        bg = AppColors.grey100;
        fg = AppColors.grey700;
        border = bg;
        break;
    }
    return SizedBox(
      width: double.infinity,
      height: 52,
      child: Material(
        color: bg,
        borderRadius: BorderRadius.circular(13),
        child: InkWell(
          borderRadius: BorderRadius.circular(13),
          onTap: onPressed,
          child: Container(
            decoration: BoxDecoration(
              borderRadius: BorderRadius.circular(13),
              border: Border.all(color: border),
            ),
            child: Row(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                if (icon != null) ...[
                  Icon(icon, size: 19, color: fg),
                  const SizedBox(width: 8),
                ],
                Text(label,
                    style: TextStyle(
                        fontSize: 16, fontWeight: FontWeight.w600, color: fg)),
              ],
            ),
          ),
        ),
      ),
    );
  }
}

enum PhoneButtonTone { primary, light, danger, ghost }

/// The little orange rounded-square Opero logo with a target glyph.
class OperoMark extends StatelessWidget {
  final double size;
  const OperoMark({super.key, this.size = 26});

  @override
  Widget build(BuildContext context) {
    return Container(
      width: size,
      height: size,
      decoration: BoxDecoration(
        color: AppColors.orange,
        borderRadius: BorderRadius.circular(size * 0.28),
      ),
      child: Center(
        child: Icon(Icons.my_location, size: size * 0.58, color: Colors.white),
      ),
    );
  }
}

/// Coloured avatar with initials (prototype style).
class InitialsAvatar extends StatelessWidget {
  final String initials;
  final Color color;
  final double size;
  const InitialsAvatar({
    super.key,
    required this.initials,
    required this.color,
    this.size = 30,
  });

  @override
  Widget build(BuildContext context) {
    return Container(
      width: size,
      height: size,
      decoration: BoxDecoration(color: color, shape: BoxShape.circle),
      alignment: Alignment.center,
      child: Text(
        initials,
        style: TextStyle(
          color: Colors.white,
          fontWeight: FontWeight.w600,
          fontSize: size * 0.38,
        ),
      ),
    );
  }
}

/// Small pill badge (NEXT / DRAFT / status chips).
class Pill extends StatelessWidget {
  final String text;
  final Color fg;
  final Color bg;
  final Color border;
  const Pill({
    super.key,
    required this.text,
    required this.fg,
    required this.bg,
    required this.border,
  });

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
      decoration: BoxDecoration(
        color: bg,
        borderRadius: BorderRadius.circular(9999),
        border: Border.all(color: border),
      ),
      child: Text(
        text,
        style: TextStyle(fontSize: 10, fontWeight: FontWeight.w700, color: fg),
      ),
    );
  }
}

/// "Demo data" hint shown on MOCK screens so it is unambiguous in the UI.
class DemoBadge extends StatelessWidget {
  const DemoBadge({super.key});

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
      decoration: BoxDecoration(
        color: AppColors.grey100,
        borderRadius: BorderRadius.circular(9999),
        border: Border.all(color: AppColors.grey200),
      ),
      child: const Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(Icons.science_outlined, size: 12, color: AppColors.grey500),
          SizedBox(width: 4),
          Text('Demo',
              style: TextStyle(
                  fontSize: 10,
                  fontWeight: FontWeight.w700,
                  color: AppColors.grey500)),
        ],
      ),
    );
  }
}

/// Heading styles used across screens.
const kH1 = TextStyle(
    fontSize: 22, fontWeight: FontWeight.w700, color: AppColors.ink, letterSpacing: -0.4);
const kH2 = TextStyle(fontSize: 20, fontWeight: FontWeight.w700, color: AppColors.ink);
const kCardTitle = TextStyle(fontSize: 16, fontWeight: FontWeight.w700, color: AppColors.ink);
