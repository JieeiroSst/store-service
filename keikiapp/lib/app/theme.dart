import 'package:flutter/material.dart';

class KeikiColors {
  static const brand = Color(0xFF5B3BE0);
  static const brandDeep = Color(0xFF3B2AA8);
  static const indigo = Color(0xFF4338CA);
  static const violet = Color(0xFF7C3AED);
  static const pink = Color(0xFFDB2777);
  static const sparkle = Color(0xFFFDE68A);
  static const accent = Color(0xFF00BFA5);
  static const warm = Color(0xFFFF7043);
  static const success = Color(0xFF16A34A);
  static const canvas = Color(0xFFF6F5FB);

  static const heroGradient = LinearGradient(
    begin: Alignment.topLeft,
    end: Alignment.bottomRight,
    colors: [indigo, violet, pink],
    stops: [0, 0.55, 1],
  );

  static const softGradient = LinearGradient(
    begin: Alignment.topLeft,
    end: Alignment.bottomRight,
    colors: [Color(0xFFEEF0FF), Color(0xFFF5EEFF), Color(0xFFFFF0F6)],
  );
}

const _font = 'BeVietnamPro';

ThemeData buildTheme(Brightness brightness) {
  final light = brightness == Brightness.light;
  final scheme = ColorScheme.fromSeed(
    seedColor: KeikiColors.brand,
    brightness: brightness,
  ).copyWith(
    primary: light ? KeikiColors.brand : const Color(0xFFB9A8FF),
    secondary: light ? KeikiColors.pink : const Color(0xFFF9A8D4),
    tertiary: light ? const Color(0xFFB45309) : KeikiColors.sparkle,
    surface: light ? Colors.white : const Color(0xFF15131F),
  );
  final radius = BorderRadius.circular(16);
  final base = ThemeData(useMaterial3: true, colorScheme: scheme, brightness: brightness, fontFamily: _font);
  final t = base.textTheme;
  final text = t.copyWith(
    displaySmall: t.displaySmall?.copyWith(fontWeight: FontWeight.w800, letterSpacing: -1),
    headlineMedium: t.headlineMedium?.copyWith(fontWeight: FontWeight.w800, letterSpacing: -0.8),
    headlineSmall: t.headlineSmall?.copyWith(fontWeight: FontWeight.w800, letterSpacing: -0.5),
    titleLarge: t.titleLarge?.copyWith(fontWeight: FontWeight.w700, letterSpacing: -0.3),
    titleMedium: t.titleMedium?.copyWith(fontWeight: FontWeight.w700, letterSpacing: -0.2),
    titleSmall: t.titleSmall?.copyWith(fontWeight: FontWeight.w600),
    bodyLarge: t.bodyLarge?.copyWith(height: 1.45),
    bodyMedium: t.bodyMedium?.copyWith(height: 1.45),
    labelLarge: t.labelLarge?.copyWith(fontWeight: FontWeight.w600),
    labelMedium: t.labelMedium?.copyWith(fontWeight: FontWeight.w600),
  );
  return base.copyWith(
    textTheme: text,
    scaffoldBackgroundColor: light ? KeikiColors.canvas : const Color(0xFF0E0C16),
    appBarTheme: AppBarTheme(
      centerTitle: false,
      backgroundColor: Colors.transparent,
      surfaceTintColor: Colors.transparent,
      scrolledUnderElevation: 0,
      titleTextStyle: TextStyle(fontFamily: _font, fontSize: 20, fontWeight: FontWeight.w700, letterSpacing: -0.3, color: scheme.onSurface),
    ),
    inputDecorationTheme: InputDecorationTheme(
      filled: true,
      fillColor: light ? Colors.white : scheme.surfaceContainerHigh,
      hintStyle: TextStyle(color: scheme.onSurfaceVariant.withValues(alpha: 0.7)),
      border: OutlineInputBorder(borderRadius: radius, borderSide: BorderSide(color: scheme.outlineVariant)),
      enabledBorder: OutlineInputBorder(borderRadius: radius, borderSide: BorderSide(color: scheme.outlineVariant.withValues(alpha: 0.7))),
      focusedBorder: OutlineInputBorder(borderRadius: radius, borderSide: BorderSide(color: scheme.primary, width: 2)),
      contentPadding: const EdgeInsets.symmetric(horizontal: 16, vertical: 16),
    ),
    filledButtonTheme: FilledButtonThemeData(
      style: FilledButton.styleFrom(
        minimumSize: const Size.fromHeight(54),
        shape: RoundedRectangleBorder(borderRadius: radius),
        textStyle: text.labelLarge?.copyWith(fontSize: 16, fontFamily: _font),
      ),
    ),
    outlinedButtonTheme: OutlinedButtonThemeData(
      style: OutlinedButton.styleFrom(minimumSize: const Size.fromHeight(50), shape: RoundedRectangleBorder(borderRadius: radius)),
    ),
    cardTheme: CardThemeData(
      elevation: 0,
      margin: EdgeInsets.zero,
      color: light ? Colors.white : scheme.surfaceContainer,
      shadowColor: KeikiColors.brand.withValues(alpha: 0.08),
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(22),
        side: BorderSide(color: light ? const Color(0xFFECEAF5) : scheme.outlineVariant.withValues(alpha: 0.4)),
      ),
    ),
    chipTheme: base.chipTheme.copyWith(
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(999)),
      side: BorderSide(color: scheme.outlineVariant.withValues(alpha: 0.6)),
      labelStyle: text.labelMedium,
    ),
    listTileTheme: ListTileThemeData(shape: RoundedRectangleBorder(borderRadius: radius)),
    searchBarTheme: SearchBarThemeData(
      elevation: const WidgetStatePropertyAll(0),
      backgroundColor: WidgetStatePropertyAll(light ? Colors.white : scheme.surfaceContainerHigh),
      side: WidgetStatePropertyAll(BorderSide(color: light ? const Color(0xFFE6E3F3) : scheme.outlineVariant)),
      shape: WidgetStatePropertyAll(RoundedRectangleBorder(borderRadius: BorderRadius.circular(18))),
      padding: const WidgetStatePropertyAll(EdgeInsets.symmetric(horizontal: 16)),
    ),
    navigationBarTheme: NavigationBarThemeData(
      height: 70,
      elevation: 0,
      backgroundColor: light ? Colors.white : scheme.surfaceContainer,
      indicatorColor: scheme.primary.withValues(alpha: 0.12),
      iconTheme: WidgetStateProperty.resolveWith(
        (s) => IconThemeData(color: s.contains(WidgetState.selected) ? scheme.primary : scheme.onSurfaceVariant),
      ),
      labelTextStyle: WidgetStateProperty.resolveWith(
        (s) => text.labelSmall?.copyWith(
          fontWeight: s.contains(WidgetState.selected) ? FontWeight.w700 : FontWeight.w500,
          color: s.contains(WidgetState.selected) ? scheme.primary : scheme.onSurfaceVariant,
        ),
      ),
    ),
    bottomSheetTheme: BottomSheetThemeData(
      showDragHandle: true,
      backgroundColor: light ? Colors.white : scheme.surfaceContainer,
      shape: const RoundedRectangleBorder(borderRadius: BorderRadius.vertical(top: Radius.circular(28))),
    ),
    dialogTheme: DialogThemeData(shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(28))),
    snackBarTheme: SnackBarThemeData(behavior: SnackBarBehavior.floating, shape: RoundedRectangleBorder(borderRadius: radius)),
    dividerTheme: DividerThemeData(color: scheme.outlineVariant.withValues(alpha: 0.4), space: 1),
  );
}
