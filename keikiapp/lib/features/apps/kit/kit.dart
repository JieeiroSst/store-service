import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:file_picker/file_picker.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:go_router/go_router.dart';

import '../../../core/network/graphql_exception.dart';
import '../../../core/utils/formatters.dart';
import '../../../core/widgets/status_view.dart';
import '../../../core/widgets/ui.dart';
import '../../../gen/gateway_api.g.dart';
import '../../auth/bloc/auth_bloc.dart';
import '../../services/bloc/services_cubit.dart';
import '../../services/data/service_meta.dart';
import '../../../app/role.dart';

export '../../../core/utils/formatters.dart';
export '../../../gen/gateway_executor.dart' show GqlUpload;
export '../../../core/widgets/status_view.dart' show showSnack, askPin, EmptyView, ErrorView;
export '../../../core/widgets/ui.dart' show IconBadge, SectionHeader, Pill, LoadingButton, BottomActionBar, Skeleton, SkeletonList;

extension KitContext on BuildContext {
  GatewayApi get api => read<GatewayApi>();
  int get userId => read<AuthBloc>().state.user?.id ?? 0;
  String get userName => read<AuthBloc>().state.user?.name ?? read<AuthBloc>().state.user?.username ?? '';
  Color get appColor => AppColor.of(this);
}

String errorMessage(Object error) {
  if (error is GraphQLException) return error.message;
  final text = error.toString();
  if (text.contains('SocketException') || text.contains('ClientException')) return 'Không kết nối được máy chủ';
  return text.replaceFirst('Exception: ', '');
}

const _accents = <String, Color>{
  'movie': Color(0xFF6D28D9),
  'ticket': Color(0xFFC026D3),
  'lucky': Color(0xFFD97706),
  'lottery': Color(0xFFDC2626),
  'polymarket': Color(0xFF2563EB),
  'video': Color(0xFFE11D48),
  'media': Color(0xFF0284C7),
  'livestream': Color(0xFFBE123C),
  'book': Color(0xFF0F766E),
  'book_store': Color(0xFFDB2777),
  'rent_house': Color(0xFF0D9488),
  'reservation': Color(0xFF1D4ED8),
  'car_rental': Color(0xFF0369A1),
  'booking_mini': Color(0xFF4F46E5),
  'parking_lot': Color(0xFF0E7490),
  'billing': Color(0xFF4338CA),
  'bonuslink': Color(0xFFEA580C),
  'calculate': Color(0xFF0284C7),
  'integrated_payment': Color(0xFF059669),
  'payment': Color(0xFF7C3AED),
  'payment_wallet': Color(0xFF0F766E),
  'recompense': Color(0xFFB45309),
  'reward': Color(0xFFDB2777),
  'voucher': Color(0xFFE11D48),
  'referral': Color(0xFF16A34A),
  'sard_auth': Color(0xFF1E40AF),
  'threads': Color(0xFF111827),
  'post': Color(0xFF9333EA),
  'room': Color(0xFF0891B2),
  'share': Color(0xFF7C3AED),
  'photo': Color(0xFFDB2777),
  'chat': Color(0xFF0D9488),
  'webrtc': Color(0xFF2563EB),
  'recruitment': Color(0xFF15803D),
  'admanagement': Color(0xFFEA580C),
  'manage': Color(0xFF334155),
  'chatbot': Color(0xFF6D28D9),
  'ollama_py': Color(0xFF4338CA),
  'draw_image': Color(0xFFC026D3),
  'ekyc': Color(0xFF0F766E),
  'face': Color(0xFF0369A1),
  'identified': Color(0xFFEA580C),
  'search': Color(0xFF1D4ED8),
  'serpapi': Color(0xFF2563EB),
  'shortlink': Color(0xFF0891B2),
  'upload': Color(0xFF475569),
  'cdn': Color(0xFF0E7490),
  'geoservice': Color(0xFF15803D),
  'shipping': Color(0xFFD97706),
  'shopify': Color(0xFF65A30D),
  'restaurant_kitchen': Color(0xFFDC2626),
  'restaurant_consumer': Color(0xFFEA580C),
  'restaurant_order': Color(0xFFB45309),
  'vending': Color(0xFF0284C7),
};

Color serviceColor(String id) => _accents[id] ?? metaFor(id).category.color;

class AppColor extends InheritedWidget {
  const AppColor({super.key, required this.color, required super.child});

  final Color color;

  static Color of(BuildContext context) =>
      context.dependOnInheritedWidgetOfExactType<AppColor>()?.color ?? Theme.of(context).colorScheme.primary;

  @override
  bool updateShouldNotify(AppColor oldWidget) => oldWidget.color != color;
}

Widget appScope(BuildContext context, Color color, Widget child) {
  final theme = Theme.of(context);
  const size = WidgetStatePropertyAll(Size(64, 46));
  return AppColor(
    color: color,
    child: Theme(
      data: theme.copyWith(
        colorScheme: theme.colorScheme.copyWith(primary: color, onPrimary: Colors.white),
        filledButtonTheme: FilledButtonThemeData(style: (theme.filledButtonTheme.style ?? const ButtonStyle()).copyWith(minimumSize: size)),
        outlinedButtonTheme: OutlinedButtonThemeData(style: (theme.outlinedButtonTheme.style ?? const ButtonStyle()).copyWith(minimumSize: size)),
      ),
      child: child,
    ),
  );
}

Color _deep(Color c) => HSLColor.fromColor(c).withLightness((HSLColor.fromColor(c).lightness * 0.62).clamp(0, 1)).toColor();

LinearGradient appGradient(Color c) => LinearGradient(
      begin: Alignment.topLeft,
      end: Alignment.bottomRight,
      colors: [c, _deep(c)],
    );

class AppTab {
  const AppTab(this.label, this.icon, this.builder);

  final String label;
  final IconData icon;
  final WidgetBuilder builder;
}

class ServiceApp extends StatefulWidget {
  const ServiceApp({super.key, required this.serviceId, required this.tabs, this.title, this.actions = const [], this.adminActions = const [], this.fab});

  final String serviceId;
  final String? title;
  final List<AppTab> tabs;
  final List<Widget> actions;
  final List<Widget> adminActions;
  final Widget? fab;

  @override
  State<ServiceApp> createState() => _ServiceAppState();
}

class _ServiceAppState extends State<ServiceApp> {
  @override
  void initState() {
    super.initState();
    try {
      context.read<ServicesCubit>().visited(widget.serviceId);
    } catch (_) {}
  }

  @override
  Widget build(BuildContext context) {
    final meta = metaFor(widget.serviceId);
    final color = serviceColor(widget.serviceId);
    final tabs = widget.tabs;
    return appScope(
      context,
      color,
      DefaultTabController(
        length: tabs.length,
        child: Scaffold(
          appBar: AppBar(
            systemOverlayStyle: SystemUiOverlayStyle.light,
            foregroundColor: Colors.white,
            backgroundColor: color,
            flexibleSpace: DecoratedBox(decoration: BoxDecoration(gradient: appGradient(color))),
            titleSpacing: 4,
            title: Row(
              children: [
                Container(
                  width: 34,
                  height: 34,
                  decoration: BoxDecoration(color: Colors.white.withValues(alpha: 0.18), borderRadius: BorderRadius.circular(11)),
                  child: Icon(meta.icon, size: 20, color: Colors.white),
                ),
                const SizedBox(width: 10),
                Flexible(
                  child: Text(
                    widget.title ?? meta.title,
                    overflow: TextOverflow.ellipsis,
                    style: const TextStyle(color: Colors.white, fontSize: 19, fontWeight: FontWeight.w700),
                  ),
                ),
              ],
            ),
            actions: [
              ...widget.actions,
              if (context.isAdmin) ...widget.adminActions,
              PopupMenuButton<String>(
                icon: const Icon(Icons.more_vert),
                onSelected: (_) => context.push('/services/${widget.serviceId}?all=1'),
                itemBuilder: (_) => const [
                  PopupMenuItem(
                    value: 'all',
                    child: ListTile(contentPadding: EdgeInsets.zero, leading: Icon(Icons.apps_rounded), title: Text('Tất cả chức năng')),
                  ),
                ],
              ),
            ],
            bottom: tabs.length < 2
                ? null
                : TabBar(
                    isScrollable: tabs.length > 3,
                    tabAlignment: tabs.length > 3 ? TabAlignment.start : null,
                    labelColor: Colors.white,
                    unselectedLabelColor: Colors.white70,
                    indicatorColor: Colors.white,
                    indicatorWeight: 3,
                    indicatorSize: TabBarIndicatorSize.label,
                    dividerColor: Colors.transparent,
                    labelStyle: const TextStyle(fontWeight: FontWeight.w700, fontSize: 14),
                    tabs: [for (final t in tabs) Tab(text: t.label, height: 44)],
                  ),
          ),
          floatingActionButton: widget.fab,
          body: tabs.length == 1 ? Builder(builder: tabs.first.builder) : TabBarView(children: [for (final t in tabs) Builder(builder: t.builder)]),
        ),
      ),
    );
  }
}

class DetailPage extends StatelessWidget {
  const DetailPage({super.key, required this.title, required this.children, this.color, this.header, this.bottom, this.actions = const []});

  final String title;
  final Color? color;
  final Widget? header;
  final List<Widget> children;
  final Widget? bottom;
  final List<Widget> actions;

  @override
  Widget build(BuildContext context) {
    final c = color ?? context.appColor;
    return appScope(
      context,
      c,
      Scaffold(
        appBar: AppBar(
          systemOverlayStyle: SystemUiOverlayStyle.light,
          foregroundColor: Colors.white,
          backgroundColor: c,
          flexibleSpace: DecoratedBox(decoration: BoxDecoration(gradient: appGradient(c))),
          title: Text(title, overflow: TextOverflow.ellipsis, style: const TextStyle(color: Colors.white, fontSize: 18, fontWeight: FontWeight.w700)),
          actions: actions,
        ),
        body: ListView(
          padding: const EdgeInsets.only(bottom: 24),
          children: [?header, ...children],
        ),
        bottomNavigationBar: bottom == null ? null : BottomActionBar(child: SizedBox(width: double.infinity, child: bottom)),
      ),
    );
  }
}

Future<T?> openPage<T>(BuildContext context, Widget page) {
  final color = context.appColor;
  return Navigator.of(context).push<T>(MaterialPageRoute(builder: (context) => appScope(context, color, page)));
}

class AsyncBuilder<T> extends StatefulWidget {
  const AsyncBuilder({super.key, required this.load, required this.builder, this.loading});

  final Future<T> Function() load;
  final Widget Function(BuildContext context, T data, Future<void> Function() reload) builder;
  final Widget? loading;

  @override
  State<AsyncBuilder<T>> createState() => _AsyncBuilderState<T>();
}

class _AsyncBuilderState<T> extends State<AsyncBuilder<T>> {
  late Future<T> _future = widget.load();

  Future<void> _reload() async {
    final next = widget.load();
    setState(() => _future = next);
    try {
      await next;
    } catch (_) {}
  }

  @override
  Widget build(BuildContext context) {
    return FutureBuilder<T>(
      future: _future,
      builder: (context, snap) {
        if (snap.connectionState != ConnectionState.done) {
          if (snap.hasData) return widget.builder(context, snap.data as T, _reload);
          return widget.loading ?? const SkeletonList();
        }
        if (snap.hasError) return ErrorView(message: errorMessage(snap.error!), onRetry: _reload);
        return widget.builder(context, snap.data as T, _reload);
      },
    );
  }
}

class AsyncList<T> extends StatelessWidget {
  const AsyncList({
    super.key,
    required this.load,
    required this.itemBuilder,
    this.header,
    this.emptyMessage = 'Chưa có dữ liệu',
    this.emptyIcon = Icons.inbox_outlined,
    this.gridExtent,
    this.gridAspect = 0.7,
    this.spacing = 10,
  });

  final Future<List<T>?> Function() load;
  final Widget Function(BuildContext context, T item, Future<void> Function() reload) itemBuilder;
  final Widget? header;
  final String emptyMessage;
  final IconData emptyIcon;
  final double? gridExtent;
  final double gridAspect;
  final double spacing;

  @override
  Widget build(BuildContext context) {
    return AsyncBuilder<List<T>?>(
      load: load,
      builder: (context, data, reload) {
        final items = data ?? const [];
        return RefreshIndicator(
          onRefresh: reload,
          child: CustomScrollView(
            physics: const AlwaysScrollableScrollPhysics(),
            slivers: [
              if (header != null) SliverToBoxAdapter(child: header),
              if (items.isEmpty)
                SliverFillRemaining(hasScrollBody: false, child: EmptyView(message: emptyMessage, icon: emptyIcon, onRefresh: reload))
              else if (gridExtent != null)
                SliverPadding(
                  padding: const EdgeInsets.fromLTRB(16, 12, 16, 96),
                  sliver: SliverGrid.builder(
                    gridDelegate: SliverGridDelegateWithMaxCrossAxisExtent(
                      maxCrossAxisExtent: gridExtent!,
                      childAspectRatio: gridAspect,
                      mainAxisSpacing: 12,
                      crossAxisSpacing: 12,
                    ),
                    itemCount: items.length,
                    itemBuilder: (context, i) => itemBuilder(context, items[i], reload),
                  ),
                )
              else
                SliverPadding(
                  padding: const EdgeInsets.fromLTRB(16, 12, 16, 96),
                  sliver: SliverList.separated(
                    itemCount: items.length,
                    separatorBuilder: (_, _) => SizedBox(height: spacing),
                    itemBuilder: (context, i) => itemBuilder(context, items[i], reload),
                  ),
                ),
            ],
          ),
        );
      },
    );
  }
}

class SearchableList<T> extends StatefulWidget {
  const SearchableList({
    super.key,
    required this.load,
    required this.itemBuilder,
    this.hint = 'Tìm kiếm',
    this.emptyMessage = 'Không tìm thấy kết quả',
    this.emptyIcon = Icons.search_off_rounded,
    this.gridExtent,
    this.gridAspect = 0.7,
    this.filters,
  });

  final Future<List<T>?> Function(String query, String? filter) load;
  final Widget Function(BuildContext context, T item, Future<void> Function() reload) itemBuilder;
  final String hint;
  final String emptyMessage;
  final IconData emptyIcon;
  final double? gridExtent;
  final double gridAspect;
  final Future<List<String>?> Function()? filters;

  @override
  State<SearchableList<T>> createState() => _SearchableListState<T>();
}

class _SearchableListState<T> extends State<SearchableList<T>> {
  String _query = '';
  String? _filter;
  late final Future<List<String>?>? _filters = widget.filters?.call().catchError((_) => <String>[]);

  @override
  Widget build(BuildContext context) {
    return Column(
      children: [
        Padding(
          padding: const EdgeInsets.fromLTRB(16, 14, 16, 0),
          child: TextField(
            textInputAction: TextInputAction.search,
            onSubmitted: (v) => setState(() => _query = v.trim()),
            decoration: InputDecoration(
              hintText: widget.hint,
              prefixIcon: const Icon(Icons.search_rounded),
              isDense: true,
            ),
          ),
        ),
        if (_filters != null)
          FutureBuilder<List<String>?>(
            future: _filters,
            builder: (context, snap) {
              final values = snap.data ?? const [];
              if (values.isEmpty) return const SizedBox(height: 4);
              return SizedBox(
                height: 52,
                child: ListView(
                  scrollDirection: Axis.horizontal,
                  padding: const EdgeInsets.fromLTRB(16, 10, 16, 2),
                  children: [
                    for (final v in [null, ...values])
                      Padding(
                        padding: const EdgeInsets.only(right: 8),
                        child: ChoiceChip(
                          label: Text(v ?? 'Tất cả'),
                          selected: _filter == v,
                          onSelected: (_) => setState(() => _filter = v),
                        ),
                      ),
                  ],
                ),
              );
            },
          ),
        Expanded(
          child: AsyncList<T>(
            key: ValueKey('$_query|$_filter'),
            load: () => widget.load(_query, _filter),
            itemBuilder: widget.itemBuilder,
            emptyMessage: widget.emptyMessage,
            emptyIcon: widget.emptyIcon,
            gridExtent: widget.gridExtent,
            gridAspect: widget.gridAspect,
          ),
        ),
      ],
    );
  }
}

class KCard extends StatelessWidget {
  const KCard({super.key, required this.child, this.onTap, this.padding = const EdgeInsets.all(14), this.color});

  final Widget child;
  final VoidCallback? onTap;
  final EdgeInsetsGeometry padding;
  final Color? color;

  @override
  Widget build(BuildContext context) {
    final scheme = Theme.of(context).colorScheme;
    return Material(
      color: color ?? scheme.surfaceContainerLowest,
      borderRadius: BorderRadius.circular(20),
      clipBehavior: Clip.antiAlias,
      child: InkWell(
        onTap: onTap,
        child: Container(
          padding: padding,
          decoration: BoxDecoration(
            borderRadius: BorderRadius.circular(20),
            border: Border.all(color: scheme.outlineVariant.withValues(alpha: 0.5)),
          ),
          child: child,
        ),
      ),
    );
  }
}

class KThumb extends StatelessWidget {
  const KThumb({super.key, this.url, this.icon = Icons.image_outlined, this.color, this.size = 56, this.radius = 16, this.width, this.height});

  final String? url;
  final IconData icon;
  final Color? color;
  final double size;
  final double radius;
  final double? width;
  final double? height;

  @override
  Widget build(BuildContext context) {
    final c = color ?? context.appColor;
    final placeholder = Container(
      width: width ?? size,
      height: height ?? size,
      decoration: BoxDecoration(
        borderRadius: BorderRadius.circular(radius),
        gradient: LinearGradient(
          begin: Alignment.topLeft,
          end: Alignment.bottomRight,
          colors: [c.withValues(alpha: 0.22), c.withValues(alpha: 0.08)],
        ),
      ),
      child: Icon(icon, color: c, size: (height ?? size) * 0.42),
    );
    final u = url;
    if (u == null || !u.startsWith('http')) return placeholder;
    return ClipRRect(
      borderRadius: BorderRadius.circular(radius),
      child: Image.network(
        u,
        width: width ?? size,
        height: height ?? size,
        fit: BoxFit.cover,
        errorBuilder: (_, _, _) => placeholder,
      ),
    );
  }
}

class EntityTile extends StatelessWidget {
  const EntityTile({
    super.key,
    required this.title,
    this.subtitle,
    this.caption,
    this.icon = Icons.circle_outlined,
    this.imageUrl,
    this.leading,
    this.trailing,
    this.badges = const [],
    this.onTap,
    this.color,
  });

  final String title;
  final String? subtitle;
  final String? caption;
  final IconData icon;
  final String? imageUrl;
  final Widget? leading;
  final Widget? trailing;
  final List<Widget> badges;
  final VoidCallback? onTap;
  final Color? color;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return KCard(
      onTap: onTap,
      child: Row(
        children: [
          leading ?? KThumb(url: imageUrl, icon: icon, color: color),
          const SizedBox(width: 14),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(title, maxLines: 2, overflow: TextOverflow.ellipsis, style: theme.textTheme.titleSmall?.copyWith(fontWeight: FontWeight.w700)),
                if (subtitle != null && subtitle!.isNotEmpty) ...[
                  const SizedBox(height: 3),
                  Text(subtitle!, maxLines: 2, overflow: TextOverflow.ellipsis, style: theme.textTheme.bodySmall?.copyWith(color: theme.colorScheme.onSurfaceVariant)),
                ],
                if (caption != null && caption!.isNotEmpty) ...[
                  const SizedBox(height: 3),
                  Text(caption!, maxLines: 1, overflow: TextOverflow.ellipsis, style: theme.textTheme.labelMedium?.copyWith(color: color ?? context.appColor, fontWeight: FontWeight.w700)),
                ],
                if (badges.isNotEmpty) ...[
                  const SizedBox(height: 6),
                  Wrap(spacing: 6, runSpacing: 4, children: badges),
                ],
              ],
            ),
          ),
          if (trailing != null) ...[const SizedBox(width: 8), trailing!] else if (onTap != null) Icon(Icons.chevron_right_rounded, color: theme.colorScheme.outline),
        ],
      ),
    );
  }
}

class PosterCard extends StatelessWidget {
  const PosterCard({super.key, required this.title, this.subtitle, this.imageUrl, this.icon = Icons.image_outlined, this.onTap, this.badge, this.width});

  final String title;
  final String? subtitle;
  final String? imageUrl;
  final IconData icon;
  final VoidCallback? onTap;
  final String? badge;
  final double? width;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final color = context.appColor;
    final card = InkWell(
      borderRadius: BorderRadius.circular(18),
      onTap: onTap,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Expanded(
            child: Stack(
              fit: StackFit.expand,
              children: [
                LayoutBuilder(
                  builder: (context, c) => KThumb(url: imageUrl, icon: icon, width: c.maxWidth, height: c.maxHeight, radius: 18),
                ),
                if (badge != null)
                  Positioned(
                    left: 8,
                    top: 8,
                    child: Container(
                      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
                      decoration: BoxDecoration(color: color, borderRadius: BorderRadius.circular(999)),
                      child: Text(badge!, style: const TextStyle(color: Colors.white, fontSize: 11, fontWeight: FontWeight.w700)),
                    ),
                  ),
              ],
            ),
          ),
          const SizedBox(height: 8),
          Text(title, maxLines: 2, overflow: TextOverflow.ellipsis, style: theme.textTheme.titleSmall?.copyWith(fontWeight: FontWeight.w700, height: 1.25)),
          if (subtitle != null && subtitle!.isNotEmpty)
            Text(subtitle!, maxLines: 1, overflow: TextOverflow.ellipsis, style: theme.textTheme.bodySmall?.copyWith(color: theme.colorScheme.onSurfaceVariant)),
        ],
      ),
    );
    return width == null ? card : SizedBox(width: width, child: card);
  }
}

class Strip<T> extends StatelessWidget {
  const Strip({super.key, required this.items, required this.itemBuilder, this.height = 230, this.itemWidth = 140});

  final List<T> items;
  final Widget Function(BuildContext context, T item) itemBuilder;
  final double height;
  final double itemWidth;

  @override
  Widget build(BuildContext context) {
    return SizedBox(
      height: height,
      child: ListView.separated(
        scrollDirection: Axis.horizontal,
        padding: const EdgeInsets.symmetric(horizontal: 16),
        itemCount: items.length,
        separatorBuilder: (_, _) => const SizedBox(width: 12),
        itemBuilder: (context, i) => SizedBox(width: itemWidth, child: itemBuilder(context, items[i])),
      ),
    );
  }
}

class AppBanner extends StatelessWidget {
  const AppBanner({super.key, required this.title, this.subtitle, this.icon, this.trailing, this.children = const [], this.margin = const EdgeInsets.fromLTRB(16, 16, 16, 4)});

  final String title;
  final String? subtitle;
  final IconData? icon;
  final Widget? trailing;
  final List<Widget> children;
  final EdgeInsetsGeometry margin;

  @override
  Widget build(BuildContext context) {
    final color = context.appColor;
    return Container(
      margin: margin,
      padding: const EdgeInsets.all(18),
      decoration: BoxDecoration(
        gradient: appGradient(color),
        borderRadius: BorderRadius.circular(24),
        boxShadow: [BoxShadow(color: color.withValues(alpha: 0.28), blurRadius: 20, offset: const Offset(0, 8))],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              if (icon != null) ...[
                Container(
                  width: 44,
                  height: 44,
                  decoration: BoxDecoration(color: Colors.white.withValues(alpha: 0.18), borderRadius: BorderRadius.circular(14)),
                  child: Icon(icon, color: Colors.white),
                ),
                const SizedBox(width: 12),
              ],
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(title, style: const TextStyle(color: Colors.white, fontSize: 18, fontWeight: FontWeight.w800)),
                    if (subtitle != null && subtitle!.isNotEmpty) ...[
                      const SizedBox(height: 2),
                      Text(subtitle!, style: TextStyle(color: Colors.white.withValues(alpha: 0.85), fontSize: 13)),
                    ],
                  ],
                ),
              ),
              ?trailing,
            ],
          ),
          if (children.isNotEmpty) ...[const SizedBox(height: 14), ...children],
        ],
      ),
    );
  }
}

class HeroHeader extends StatelessWidget {
  const HeroHeader({super.key, required this.title, this.subtitle, this.imageUrl, this.icon = Icons.image_outlined, this.badges = const []});

  final String title;
  final String? subtitle;
  final String? imageUrl;
  final IconData icon;
  final List<Widget> badges;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Padding(
          padding: const EdgeInsets.fromLTRB(16, 16, 16, 0),
          child: AspectRatio(
            aspectRatio: 16 / 9,
            child: LayoutBuilder(builder: (context, c) => KThumb(url: imageUrl, icon: icon, width: c.maxWidth, height: c.maxHeight, radius: 24)),
          ),
        ),
        Padding(
          padding: const EdgeInsets.fromLTRB(20, 16, 20, 4),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(title, style: theme.textTheme.headlineSmall?.copyWith(fontWeight: FontWeight.w800)),
              if (subtitle != null && subtitle!.isNotEmpty) ...[
                const SizedBox(height: 6),
                Text(subtitle!, style: theme.textTheme.bodyMedium?.copyWith(color: theme.colorScheme.onSurfaceVariant)),
              ],
              if (badges.isNotEmpty) ...[const SizedBox(height: 10), Wrap(spacing: 6, runSpacing: 6, children: badges)],
            ],
          ),
        ),
      ],
    );
  }
}

class InfoSection extends StatelessWidget {
  const InfoSection({super.key, this.title, required this.rows});

  final String? title;
  final List<(String, String?)> rows;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final shown = rows.where((r) => r.$2 != null && r.$2!.isNotEmpty && r.$2 != 'null').toList();
    if (shown.isEmpty) return const SizedBox.shrink();
    return Padding(
      padding: const EdgeInsets.fromLTRB(16, 12, 16, 0),
      child: KCard(
        padding: const EdgeInsets.fromLTRB(16, 12, 16, 12),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            if (title != null) ...[Text(title!, style: theme.textTheme.titleSmall?.copyWith(fontWeight: FontWeight.w800)), const SizedBox(height: 6)],
            for (final (i, r) in shown.indexed) ...[
              if (i > 0) Divider(height: 1, color: theme.colorScheme.outlineVariant.withValues(alpha: 0.5)),
              Padding(
                padding: const EdgeInsets.symmetric(vertical: 9),
                child: Row(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    SizedBox(
                      width: 116,
                      child: Text(r.$1, style: theme.textTheme.bodyMedium?.copyWith(color: theme.colorScheme.onSurfaceVariant)),
                    ),
                    Expanded(child: SelectableText(r.$2!, textAlign: TextAlign.right, style: theme.textTheme.bodyMedium?.copyWith(fontWeight: FontWeight.w600))),
                  ],
                ),
              ),
            ],
          ],
        ),
      ),
    );
  }
}

class Paragraph extends StatelessWidget {
  const Paragraph(this.text, {super.key, this.title});

  final String? text;
  final String? title;

  @override
  Widget build(BuildContext context) {
    if (text == null || text!.trim().isEmpty) return const SizedBox.shrink();
    final theme = Theme.of(context);
    return Padding(
      padding: const EdgeInsets.fromLTRB(20, 16, 20, 0),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          if (title != null) ...[Text(title!, style: theme.textTheme.titleSmall?.copyWith(fontWeight: FontWeight.w800)), const SizedBox(height: 6)],
          Text(text!.trim(), style: theme.textTheme.bodyMedium?.copyWith(height: 1.5)),
        ],
      ),
    );
  }
}

class StatTile extends StatelessWidget {
  const StatTile({super.key, required this.label, required this.value, this.icon, this.onDark = false});

  final String label;
  final String value;
  final IconData? icon;
  final bool onDark;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final fg = onDark ? Colors.white : theme.colorScheme.onSurface;
    final sub = onDark ? Colors.white.withValues(alpha: 0.8) : theme.colorScheme.onSurfaceVariant;
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
      decoration: BoxDecoration(
        color: onDark ? Colors.white.withValues(alpha: 0.14) : context.appColor.withValues(alpha: 0.08),
        borderRadius: BorderRadius.circular(16),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              if (icon != null) ...[Icon(icon, size: 14, color: sub), const SizedBox(width: 4)],
              Flexible(child: Text(label, maxLines: 1, overflow: TextOverflow.ellipsis, style: TextStyle(color: sub, fontSize: 12))),
            ],
          ),
          const SizedBox(height: 4),
          FittedBox(
            fit: BoxFit.scaleDown,
            alignment: Alignment.centerLeft,
            child: Text(value, style: TextStyle(color: fg, fontSize: 17, fontWeight: FontWeight.w800)),
          ),
        ],
      ),
    );
  }
}

class StatRow extends StatelessWidget {
  const StatRow(this.tiles, {super.key, this.padding = EdgeInsets.zero});

  final List<Widget> tiles;
  final EdgeInsetsGeometry padding;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: padding,
      child: Row(
        children: [
          for (final (i, t) in tiles.indexed) ...[if (i > 0) const SizedBox(width: 8), Expanded(child: t)],
        ],
      ),
    );
  }
}

const _statusLabels = <String, (String, Color)>{
  'active': ('Đang hoạt động', Color(0xFF16A34A)),
  'live': ('Đang live', Color(0xFFDC2626)),
  'open': ('Đang mở', Color(0xFF16A34A)),
  'published': ('Đã đăng', Color(0xFF16A34A)),
  'approved': ('Đã duyệt', Color(0xFF16A34A)),
  'confirmed': ('Đã xác nhận', Color(0xFF16A34A)),
  'completed': ('Hoàn tất', Color(0xFF16A34A)),
  'complete': ('Hoàn tất', Color(0xFF16A34A)),
  'done': ('Hoàn tất', Color(0xFF16A34A)),
  'success': ('Thành công', Color(0xFF16A34A)),
  'succeeded': ('Thành công', Color(0xFF16A34A)),
  'paid': ('Đã thanh toán', Color(0xFF16A34A)),
  'valid': ('Hợp lệ', Color(0xFF16A34A)),
  'available': ('Còn trống', Color(0xFF16A34A)),
  'delivered': ('Đã giao', Color(0xFF16A34A)),
  'ready': ('Sẵn sàng', Color(0xFF16A34A)),
  'verified': ('Đã xác thực', Color(0xFF16A34A)),
  'used': ('Đã dùng', Color(0xFF64748B)),
  'checked_in': ('Đã check-in', Color(0xFF64748B)),
  'pending': ('Đang chờ', Color(0xFFD97706)),
  'processing': ('Đang xử lý', Color(0xFFD97706)),
  'in_progress': ('Đang xử lý', Color(0xFFD97706)),
  'reserved': ('Đang giữ chỗ', Color(0xFFD97706)),
  'booked': ('Đã giữ bàn', Color(0xFF16A34A)),
  'draft': ('Bản nháp', Color(0xFF64748B)),
  'review': ('Chờ duyệt', Color(0xFFD97706)),
  'in_review': ('Chờ duyệt', Color(0xFFD97706)),
  'shipping': ('Đang giao', Color(0xFF2563EB)),
  'scheduled': ('Đã lên lịch', Color(0xFF2563EB)),
  'upcoming': ('Sắp diễn ra', Color(0xFF2563EB)),
  'idle': ('Chưa phát', Color(0xFF64748B)),
  'offline': ('Ngoại tuyến', Color(0xFF64748B)),
  'ended': ('Đã kết thúc', Color(0xFF64748B)),
  'closed': ('Đã đóng', Color(0xFF64748B)),
  'expired': ('Hết hạn', Color(0xFF64748B)),
  'inactive': ('Ngừng hoạt động', Color(0xFF64748B)),
  'paused': ('Tạm dừng', Color(0xFF64748B)),
  'resolved': ('Đã phân xử', Color(0xFF7C3AED)),
  'listed': ('Đang bán lại', Color(0xFF7C3AED)),
  'transferred': ('Đã chuyển', Color(0xFF7C3AED)),
  'cancelled': ('Đã huỷ', Color(0xFFDC2626)),
  'canceled': ('Đã huỷ', Color(0xFFDC2626)),
  'failed': ('Thất bại', Color(0xFFDC2626)),
  'rejected': ('Bị từ chối', Color(0xFFDC2626)),
  'refunded': ('Đã hoàn tiền', Color(0xFF64748B)),
  'void': ('Đã vô hiệu', Color(0xFFDC2626)),
  'voided': ('Đã vô hiệu', Color(0xFFDC2626)),
  'blocked': ('Bị khoá', Color(0xFFDC2626)),
  'disputed': ('Đang khiếu nại', Color(0xFFDC2626)),
  'sold_out': ('Hết vé', Color(0xFFDC2626)),
  'out_of_stock': ('Hết hàng', Color(0xFFDC2626)),
  'rented': ('Đang cho thuê', Color(0xFFD97706)),
  'maintenance': ('Bảo trì', Color(0xFF64748B)),
  'retired': ('Ngừng khai thác', Color(0xFF64748B)),
  'overdue': ('Quá hạn', Color(0xFFDC2626)),
  'unspecified': ('Không rõ', Color(0xFF64748B)),
  'parked': ('Đang gửi', Color(0xFF2563EB)),
  'left': ('Đã lấy xe', Color(0xFF64748B)),
};

class StatusPill extends StatelessWidget {
  const StatusPill(this.status, {super.key});

  final String? status;

  @override
  Widget build(BuildContext context) {
    final s = status?.trim() ?? '';
    if (s.isEmpty) return const SizedBox.shrink();
    var key = s.toLowerCase();
    final i = key.indexOf('status_');
    if (i >= 0) key = key.substring(i + 7);
    final known = _statusLabels[key];
    return Pill(known?.$1 ?? s, color: known?.$2 ?? Theme.of(context).colorScheme.outline);
  }
}

enum FieldKind { text, number, multiline, email, phone, password, date, datetime, choice, file, toggle }

class KField {
  const KField(this.key, this.label, {this.kind = FieldKind.text, this.required = false, this.initial, this.hint, this.options = const [], this.icon});

  final String key;
  final String label;
  final FieldKind kind;
  final bool required;
  final String? initial;
  final String? hint;
  final List<(String, String)> options;
  final IconData? icon;
}

class FormData {
  const FormData(this.values, [this.files = const {}]);

  final Map<String, String> values;
  final Map<String, GqlUpload> files;

  String? str(String key) {
    final v = values[key]?.trim();
    return (v == null || v.isEmpty) ? null : v;
  }

  int? integer(String key) => int.tryParse((str(key) ?? '').replaceAll('.', '').replaceAll(',', ''));

  double? number(String key) => double.tryParse((str(key) ?? '').replaceAll(',', '.'));

  GqlUpload? file(String key) => files[key];

  bool flag(String key) => values[key] == 'true';

  List<int> ints(String key) => (str(key) ?? '').split(RegExp(r'[,\s]+')).map(int.tryParse).whereType<int>().toList();
}

const _mimes = {
  'jpg': 'image/jpeg',
  'jpeg': 'image/jpeg',
  'png': 'image/png',
  'gif': 'image/gif',
  'webp': 'image/webp',
  'heic': 'image/heic',
  'pdf': 'application/pdf',
  'mp4': 'video/mp4',
  'mov': 'video/quicktime',
  'mp3': 'audio/mpeg',
  'wav': 'audio/wav',
  'csv': 'text/csv',
  'json': 'application/json',
  'txt': 'text/plain',
};

Future<GqlUpload?> pickUpload() async {
  final file = await FilePicker.pickFile();
  if (file == null) return null;
  final ext = (file.extension ?? '').replaceFirst('.', '').toLowerCase();
  return GqlUpload(bytes: await file.readAsBytes(), filename: file.name, contentType: _mimes[ext]);
}

Future<bool> showFormSheet(
  BuildContext context, {
  required String title,
  required List<KField> fields,
  required Future<void> Function(FormData values) onSubmit,
  String submitLabel = 'Xác nhận',
  String? success,
  String? description,
}) async {
  final color = context.appColor;
  final ok = await showModalBottomSheet<bool>(
    context: context,
    isScrollControlled: true,
    showDragHandle: true,
    useSafeArea: true,
    builder: (context) => appScope(
      context,
      color,
      _FormSheet(title: title, fields: fields, onSubmit: onSubmit, submitLabel: submitLabel, description: description),
    ),
  );
  if (ok == true && success != null && context.mounted) showSnack(context, success);
  return ok == true;
}

class _FormSheet extends StatefulWidget {
  const _FormSheet({required this.title, required this.fields, required this.onSubmit, required this.submitLabel, this.description});

  final String title;
  final List<KField> fields;
  final Future<void> Function(FormData values) onSubmit;
  final String submitLabel;
  final String? description;

  @override
  State<_FormSheet> createState() => _FormSheetState();
}

class _FormSheetState extends State<_FormSheet> {
  final _form = GlobalKey<FormState>();
  late final Map<String, TextEditingController> _controllers = {
    for (final f in widget.fields) f.key: TextEditingController(text: f.initial ?? (f.kind == FieldKind.choice && f.options.isNotEmpty ? f.options.first.$1 : '')),
  };
  final _files = <String, GqlUpload>{};
  bool _busy = false;
  String? _error;

  @override
  void dispose() {
    for (final c in _controllers.values) {
      c.dispose();
    }
    super.dispose();
  }

  Future<void> _submit() async {
    if (!_form.currentState!.validate()) return;
    setState(() {
      _busy = true;
      _error = null;
    });
    try {
      await widget.onSubmit(FormData({for (final e in _controllers.entries) e.key: e.value.text}, Map.of(_files)));
      if (mounted) Navigator.pop(context, true);
    } catch (e) {
      if (mounted) setState(() => _error = errorMessage(e));
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  Future<void> _pickDateTime(TextEditingController c) async {
    final now = DateTime.now();
    final current = DateTime.tryParse(c.text)?.toLocal() ?? now.add(const Duration(days: 7));
    final d = await showDatePicker(context: context, initialDate: current, firstDate: DateTime(now.year - 5), lastDate: DateTime(now.year + 5));
    if (d == null || !mounted) return;
    final t = await showTimePicker(context: context, initialTime: TimeOfDay.fromDateTime(current));
    if (t == null) return;
    c.text = DateTime(d.year, d.month, d.day, t.hour, t.minute).toUtc().toIso8601String();
    setState(() {});
  }

  Future<void> _pickDate(TextEditingController c) async {
    final now = DateTime.now();
    final picked = await showDatePicker(
      context: context,
      initialDate: DateTime.tryParse(c.text) ?? now,
      firstDate: DateTime(now.year - 100),
      lastDate: DateTime(now.year + 5),
    );
    if (picked != null) c.text = picked.toIso8601String().substring(0, 10);
  }

  Widget _field(KField f) {
    final c = _controllers[f.key]!;
    String? validate(String? v) => f.required && (v == null || v.trim().isEmpty) ? 'Vui lòng nhập ${f.label.toLowerCase()}' : null;
    if (f.kind == FieldKind.file) {
      final picked = _files[f.key];
      return FormField<GqlUpload>(
        validator: (_) => f.required && _files[f.key] == null ? 'Hãy chọn tệp' : null,
        builder: (state) => InkWell(
          borderRadius: BorderRadius.circular(16),
          onTap: () async {
            final u = await pickUpload();
            if (u != null) setState(() => _files[f.key] = u);
          },
          child: InputDecorator(
            decoration: InputDecoration(labelText: f.label, errorText: state.errorText, prefixIcon: Icon(f.icon ?? Icons.attach_file)),
            child: Text(
              picked == null ? 'Chạm để chọn tệp' : '${picked.filename} • ${(picked.bytes.length / 1024).toStringAsFixed(0)} KB',
              overflow: TextOverflow.ellipsis,
            ),
          ),
        ),
      );
    }
    if (f.kind == FieldKind.toggle) {
      return SwitchListTile(
        contentPadding: EdgeInsets.zero,
        title: Text(f.label),
        value: c.text == 'true',
        onChanged: (v) => setState(() => c.text = '$v'),
      );
    }
    if (f.kind == FieldKind.datetime) {
      return FormField<String>(
        validator: (_) => validate(c.text),
        builder: (state) => InkWell(
          borderRadius: BorderRadius.circular(16),
          onTap: () => _pickDateTime(c),
          child: InputDecorator(
            decoration: InputDecoration(labelText: f.label, errorText: state.errorText, prefixIcon: const Icon(Icons.schedule)),
            child: Text(c.text.isEmpty ? 'Chạm để chọn ngày giờ' : formatDate(c.text)),
          ),
        ),
      );
    }
    if (f.kind == FieldKind.choice) {
      return DropdownButtonFormField<String>(
        initialValue: c.text.isEmpty ? null : c.text,
        isExpanded: true,
        decoration: InputDecoration(labelText: f.label, prefixIcon: f.icon == null ? null : Icon(f.icon)),
        items: [for (final (v, l) in f.options) DropdownMenuItem(value: v, child: Text(l))],
        onChanged: (v) => c.text = v ?? '',
        validator: validate,
      );
    }
    return TextFormField(
      controller: c,
      readOnly: f.kind == FieldKind.date,
      onTap: f.kind == FieldKind.date ? () => _pickDate(c) : null,
      obscureText: f.kind == FieldKind.password,
      minLines: f.kind == FieldKind.multiline ? 3 : 1,
      maxLines: f.kind == FieldKind.multiline ? 6 : 1,
      keyboardType: switch (f.kind) {
        FieldKind.number => const TextInputType.numberWithOptions(decimal: true),
        FieldKind.email => TextInputType.emailAddress,
        FieldKind.phone => TextInputType.phone,
        FieldKind.multiline => TextInputType.multiline,
        _ => TextInputType.text,
      },
      decoration: InputDecoration(
        labelText: f.label,
        hintText: f.hint,
        prefixIcon: f.icon == null ? (f.kind == FieldKind.date ? const Icon(Icons.event_outlined) : null) : Icon(f.icon),
      ),
      validator: validate,
    );
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Padding(
      padding: EdgeInsets.only(bottom: MediaQuery.viewInsetsOf(context).bottom),
      child: SingleChildScrollView(
        padding: const EdgeInsets.fromLTRB(20, 0, 20, 20),
        child: Form(
          key: _form,
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              Text(widget.title, style: theme.textTheme.titleLarge?.copyWith(fontWeight: FontWeight.w800)),
              if (widget.description != null) ...[
                const SizedBox(height: 4),
                Text(widget.description!, style: theme.textTheme.bodyMedium?.copyWith(color: theme.colorScheme.onSurfaceVariant)),
              ],
              const SizedBox(height: 16),
              for (final f in widget.fields) ...[_field(f), const SizedBox(height: 12)],
              if (_error != null) ...[
                Container(
                  padding: const EdgeInsets.all(12),
                  decoration: BoxDecoration(color: theme.colorScheme.errorContainer, borderRadius: BorderRadius.circular(12)),
                  child: Text(_error!, style: TextStyle(color: theme.colorScheme.onErrorContainer)),
                ),
                const SizedBox(height: 12),
              ],
              const SizedBox(height: 4),
              LoadingButton(label: widget.submitLabel, loading: _busy, onPressed: _submit),
            ],
          ),
        ),
      ),
    );
  }
}

Future<bool> runAction(
  BuildContext context,
  Future<Object?> Function() action, {
  String? success,
  String? confirm,
  String confirmLabel = 'Đồng ý',
  bool destructive = false,
}) async {
  if (confirm != null) {
    final ok = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        content: Text(confirm),
        actions: [
          TextButton(onPressed: () => Navigator.pop(context, false), child: const Text('Huỷ')),
          FilledButton(
            style: destructive ? FilledButton.styleFrom(backgroundColor: Theme.of(context).colorScheme.error) : null,
            onPressed: () => Navigator.pop(context, true),
            child: Text(confirmLabel),
          ),
        ],
      ),
    );
    if (ok != true) return false;
  }
  try {
    await action();
    if (success != null && context.mounted) showSnack(context, success);
    return true;
  } catch (e) {
    if (context.mounted) showSnack(context, errorMessage(e));
    return false;
  }
}

Future<void> showResult(BuildContext context, {required String title, required String message, IconData icon = Icons.check_circle_rounded, Color? color}) {
  final c = color ?? context.appColor;
  return showDialog<void>(
    context: context,
    builder: (context) => AlertDialog(
      icon: Icon(icon, size: 48, color: c),
      title: Text(title, textAlign: TextAlign.center),
      content: Text(message, textAlign: TextAlign.center),
      actions: [FilledButton(onPressed: () => Navigator.pop(context), child: const Text('Đóng'))],
    ),
  );
}

class ActionButtons extends StatelessWidget {
  const ActionButtons(this.buttons, {super.key, this.padding = const EdgeInsets.fromLTRB(16, 16, 16, 0)});

  final List<Widget> buttons;
  final EdgeInsetsGeometry padding;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: padding,
      child: Wrap(spacing: 8, runSpacing: 8, children: buttons),
    );
  }
}

String shortDate(String? iso) {
  final d = DateTime.tryParse(iso ?? '');
  if (d == null) return iso ?? '';
  final l = d.toLocal();
  return '${l.day.toString().padLeft(2, '0')}/${l.month.toString().padLeft(2, '0')}/${l.year}';
}

String compact(num? n) {
  final v = n ?? 0;
  if (v >= 1000000) return '${(v / 1000000).toStringAsFixed(1).replaceAll('.0', '')}tr';
  if (v >= 1000) return '${(v / 1000).toStringAsFixed(1).replaceAll('.0', '')}k';
  return v.toStringAsFixed(v is int || v == v.roundToDouble() ? 0 : 1);
}

String duration(num? seconds) {
  final s = (seconds ?? 0).round();
  final h = s ~/ 3600;
  final m = (s % 3600) ~/ 60;
  final sec = s % 60;
  if (h > 0) return '$h:${m.toString().padLeft(2, '0')}:${sec.toString().padLeft(2, '0')}';
  return '$m:${sec.toString().padLeft(2, '0')}';
}

String vnd(num? amount) => formatMoney(amount);

Object? pick(Object? data, String key) {
  var level = <Object?>[data];
  while (level.isNotEmpty) {
    final next = <Object?>[];
    for (final node in level) {
      if (node is Map) {
        if (node[key] != null) return node[key];
        next.addAll(node.values);
      } else if (node is List) {
        next.addAll(node);
      }
    }
    level = next;
  }
  return null;
}

int? pickInt(Object? data, String key) {
  final v = pick(data, key);
  return v is num ? v.toInt() : int.tryParse('$v');
}

class LocalPrefs {
  static const _storage = FlutterSecureStorage();
  static final _memory = <String, String>{};

  static Future<String?> get(String key) async {
    try {
      return await _storage.read(key: 'app.$key').timeout(const Duration(seconds: 2)) ?? _memory[key];
    } catch (_) {
      return _memory[key];
    }
  }

  static Future<void> set(String key, String value) async {
    _memory[key] = value;
    try {
      await _storage.write(key: 'app.$key', value: value).timeout(const Duration(seconds: 2));
    } catch (_) {}
  }
}

const ratingOptions = [
  ('5', '5 sao – Tuyệt vời'),
  ('4', '4 sao – Hài lòng'),
  ('3', '3 sao – Bình thường'),
  ('2', '2 sao – Chưa tốt'),
  ('1', '1 sao – Tệ'),
];
