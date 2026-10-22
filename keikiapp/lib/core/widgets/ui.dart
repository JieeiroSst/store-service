import 'package:flutter/material.dart';
import 'package:flutter_svg/flutter_svg.dart';

import '../../app/theme.dart';

class IconBadge extends StatelessWidget {
  const IconBadge({super.key, required this.icon, required this.color, this.size = 48});

  final IconData icon;
  final Color color;
  final double size;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: size,
      height: size,
      decoration: BoxDecoration(
        color: color.withValues(alpha: 0.12),
        borderRadius: BorderRadius.circular(size * 0.32),
      ),
      child: Icon(icon, color: color, size: size * 0.52),
    );
  }
}

class SectionHeader extends StatelessWidget {
  const SectionHeader(this.title, {super.key, this.action, this.onAction, this.padding = const EdgeInsets.fromLTRB(20, 24, 12, 8)});

  final String title;
  final String? action;
  final VoidCallback? onAction;
  final EdgeInsets padding;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: padding,
      child: Row(
        children: [
          Expanded(child: Text(title, style: Theme.of(context).textTheme.titleMedium)),
          if (action != null) TextButton(onPressed: onAction, child: Text(action!)),
        ],
      ),
    );
  }
}

class GradientCard extends StatelessWidget {
  const GradientCard({super.key, required this.child, this.gradient = KeikiColors.heroGradient, this.padding = const EdgeInsets.all(20), this.onTap});

  final Widget child;
  final Gradient gradient;
  final EdgeInsets padding;
  final VoidCallback? onTap;

  @override
  Widget build(BuildContext context) {
    return DecoratedBox(
      decoration: BoxDecoration(
        gradient: gradient,
        borderRadius: BorderRadius.circular(24),
        boxShadow: [BoxShadow(color: gradient.colors.first.withValues(alpha: 0.35), blurRadius: 24, offset: const Offset(0, 10))],
      ),
      child: Material(
        type: MaterialType.transparency,
        child: InkWell(
          borderRadius: BorderRadius.circular(24),
          onTap: onTap,
          child: Padding(padding: padding, child: DefaultTextStyle.merge(style: const TextStyle(color: Colors.white), child: child)),
        ),
      ),
    );
  }
}

class QuickAction extends StatelessWidget {
  const QuickAction({super.key, required this.icon, required this.label, required this.onTap, this.color, this.onDark = false});

  final IconData icon;
  final String label;
  final VoidCallback onTap;
  final Color? color;
  final bool onDark;

  @override
  Widget build(BuildContext context) {
    final scheme = Theme.of(context).colorScheme;
    final c = color ?? scheme.primary;
    return InkWell(
      borderRadius: BorderRadius.circular(16),
      onTap: onTap,
      child: Padding(
        padding: const EdgeInsets.symmetric(vertical: 6, horizontal: 2),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Container(
              width: 50,
              height: 50,
              decoration: BoxDecoration(
                color: onDark ? Colors.white.withValues(alpha: 0.18) : null,
                gradient: onDark
                    ? null
                    : LinearGradient(
                        begin: Alignment.topLeft,
                        end: Alignment.bottomRight,
                        colors: [c.withValues(alpha: 0.16), c.withValues(alpha: 0.07)],
                      ),
                borderRadius: BorderRadius.circular(17),
              ),
              child: Icon(icon, color: onDark ? Colors.white : c, size: 24),
            ),
            const SizedBox(height: 7),
            Text(
              label,
              maxLines: 2,
              overflow: TextOverflow.ellipsis,
              textAlign: TextAlign.center,
              style: Theme.of(context).textTheme.labelMedium?.copyWith(
                    color: onDark ? Colors.white : null,
                    fontSize: 12,
                    height: 1.2,
                  ),
            ),
          ],
        ),
      ),
    );
  }
}

class Skeleton extends StatefulWidget {
  const Skeleton({super.key, this.height = 16, this.width, this.radius = 12});

  final double height;
  final double? width;
  final double radius;

  @override
  State<Skeleton> createState() => _SkeletonState();
}

class _SkeletonState extends State<Skeleton> with SingleTickerProviderStateMixin {
  late final _controller = AnimationController(vsync: this, duration: const Duration(milliseconds: 900))..repeat(reverse: true);

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final color = Theme.of(context).colorScheme.surfaceContainerHighest;
    return FadeTransition(
      opacity: Tween(begin: 0.45, end: 1.0).animate(_controller),
      child: Container(
        height: widget.height,
        width: widget.width,
        decoration: BoxDecoration(color: color, borderRadius: BorderRadius.circular(widget.radius)),
      ),
    );
  }
}

class SkeletonList extends StatelessWidget {
  const SkeletonList({super.key, this.count = 6});

  final int count;

  @override
  Widget build(BuildContext context) {
    return ListView.separated(
      physics: const NeverScrollableScrollPhysics(),
      padding: const EdgeInsets.all(16),
      itemCount: count,
      separatorBuilder: (_, _) => const SizedBox(height: 12),
      itemBuilder: (_, _) => const Row(
        children: [
          Skeleton(height: 48, width: 48, radius: 16),
          SizedBox(width: 12),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [Skeleton(height: 14), SizedBox(height: 8), Skeleton(height: 12, width: 140)],
            ),
          ),
        ],
      ),
    );
  }
}

class Pill extends StatelessWidget {
  const Pill(this.text, {super.key, this.color, this.icon});

  final String text;
  final Color? color;
  final IconData? icon;

  @override
  Widget build(BuildContext context) {
    final c = color ?? Theme.of(context).colorScheme.primary;
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
      decoration: BoxDecoration(color: c.withValues(alpha: 0.12), borderRadius: BorderRadius.circular(999)),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          if (icon != null) ...[Icon(icon, size: 14, color: c), const SizedBox(width: 4)],
          Text(text, style: Theme.of(context).textTheme.labelSmall?.copyWith(color: c, fontWeight: FontWeight.w700)),
        ],
      ),
    );
  }
}

class LoadingButton extends StatelessWidget {
  const LoadingButton({super.key, required this.label, required this.onPressed, this.loading = false, this.icon});

  final String label;
  final VoidCallback? onPressed;
  final bool loading;
  final IconData? icon;

  @override
  Widget build(BuildContext context) {
    final child = loading
        ? const SizedBox(height: 22, width: 22, child: CircularProgressIndicator(strokeWidth: 2.5, color: Colors.white))
        : Text(label);
    if (icon != null && !loading) return FilledButton.icon(onPressed: onPressed, icon: Icon(icon), label: child);
    return FilledButton(onPressed: loading ? null : onPressed, child: child);
  }
}

class BottomActionBar extends StatelessWidget {
  const BottomActionBar({super.key, required this.child});

  final Widget child;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: Theme.of(context).colorScheme.surface,
      elevation: 8,
      child: SafeArea(top: false, child: Padding(padding: const EdgeInsets.fromLTRB(16, 12, 16, 12), child: child)),
    );
  }
}

class BrandMark extends StatelessWidget {
  const BrandMark({super.key, this.size = 64, this.shadow = true});

  final double size;
  final bool shadow;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: size,
      height: size,
      decoration: BoxDecoration(
        borderRadius: BorderRadius.circular(size * 0.227),
        boxShadow: shadow
            ? [BoxShadow(color: KeikiColors.violet.withValues(alpha: 0.35), blurRadius: size * 0.35, offset: Offset(0, size * 0.12))]
            : null,
      ),
      child: SvgPicture.asset('assets/brand/keikibook_icon.svg', width: size, height: size),
    );
  }
}

class BrandWordmark extends StatelessWidget {
  const BrandWordmark({super.key, this.fontSize = 28, this.onDark = false});

  final double fontSize;
  final bool onDark;

  @override
  Widget build(BuildContext context) {
    final text = Text(
      'Keikibook',
      style: TextStyle(
        fontFamily: 'BeVietnamPro',
        fontSize: fontSize,
        fontWeight: FontWeight.w800,
        letterSpacing: -fontSize * 0.035,
        height: 1.1,
        color: onDark ? Colors.white : Colors.black,
      ),
    );
    if (onDark) return text;
    return ShaderMask(
      blendMode: BlendMode.srcIn,
      shaderCallback: (r) => KeikiColors.heroGradient.createShader(r),
      child: text,
    );
  }
}

class KeikiLogo extends StatelessWidget {
  const KeikiLogo({super.key, this.size = 64, this.onDark = false, this.horizontal = false});

  final double size;
  final bool onDark;
  final bool horizontal;

  @override
  Widget build(BuildContext context) {
    final mark = BrandMark(size: size, shadow: !onDark);
    if (horizontal) {
      return Row(
        mainAxisSize: MainAxisSize.min,
        children: [mark, SizedBox(width: size * 0.3), BrandWordmark(fontSize: size * 0.58, onDark: onDark)],
      );
    }
    return Column(
      mainAxisSize: MainAxisSize.min,
      children: [mark, SizedBox(height: size * 0.24), BrandWordmark(fontSize: size * 0.42, onDark: onDark)],
    );
  }
}

class HeroBackdrop extends StatelessWidget {
  const HeroBackdrop({super.key, required this.child, this.borderRadius = BorderRadius.zero});

  final Widget child;
  final BorderRadius borderRadius;

  @override
  Widget build(BuildContext context) {
    Widget glow(double size, Color color) => Container(
          width: size,
          height: size,
          decoration: BoxDecoration(
            shape: BoxShape.circle,
            gradient: RadialGradient(colors: [color.withValues(alpha: 0.55), color.withValues(alpha: 0)]),
          ),
        );
    return ClipRRect(
      borderRadius: borderRadius,
      child: DecoratedBox(
        decoration: const BoxDecoration(gradient: KeikiColors.heroGradient),
        child: Stack(
          children: [
            Positioned(top: -90, right: -70, child: glow(260, const Color(0xFFF0ABFC))),
            Positioned(bottom: -120, left: -80, child: glow(300, const Color(0xFF818CF8))),
            Positioned(top: 40, left: 120, child: glow(120, KeikiColors.sparkle)),
            child,
          ],
        ),
      ),
    );
  }
}
