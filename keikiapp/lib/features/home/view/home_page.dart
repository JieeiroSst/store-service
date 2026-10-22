import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:go_router/go_router.dart';

import '../../../app/mini_app.dart';
import '../../../app/mini_apps.dart';
import '../../../app/theme.dart';
import '../../../core/utils/formatters.dart';
import '../../../core/utils/load_status.dart';
import '../../../core/widgets/ui.dart';
import '../../auth/bloc/auth_bloc.dart';
import '../../services/bloc/services_cubit.dart';
import '../../services/data/service_meta.dart';
import '../../wallet/bloc/wallet_cubit.dart';

class HomePage extends StatelessWidget {
  const HomePage({super.key});

  @override
  Widget build(BuildContext context) {
    final scheme = Theme.of(context).colorScheme;
    return Scaffold(
      body: RefreshIndicator(
        onRefresh: context.read<WalletCubit>().load,
        child: CustomScrollView(
          slivers: [
            const SliverToBoxAdapter(child: _Header()),
            SliverToBoxAdapter(
              child: Padding(
                padding: const EdgeInsets.fromLTRB(16, 20, 16, 0),
                child: Material(
                  color: scheme.surface,
                  shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(18),
                    side: BorderSide(color: scheme.outlineVariant.withValues(alpha: 0.5)),
                  ),
                  child: InkWell(
                    borderRadius: BorderRadius.circular(18),
                    onTap: () => context.go('/services'),
                    child: Padding(
                      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
                      child: Row(
                        children: [
                          Icon(Icons.search_rounded, color: scheme.primary),
                          const SizedBox(width: 12),
                          Expanded(
                            child: Text(
                              'Tìm phim, vé, khách sạn, món ăn…',
                              maxLines: 1,
                              overflow: TextOverflow.ellipsis,
                              style: TextStyle(color: scheme.onSurfaceVariant),
                            ),
                          ),
                          Container(
                            padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                            decoration: BoxDecoration(gradient: KeikiColors.heroGradient, borderRadius: BorderRadius.circular(8)),
                            child: const Text(
                              'AI',
                              style: TextStyle(color: Colors.white, fontWeight: FontWeight.w800, fontSize: 11),
                            ),
                          ),
                        ],
                      ),
                    ),
                  ),
                ),
              ),
            ),
            SliverToBoxAdapter(
              child: SectionHeader('Dịch vụ nổi bật', action: 'Tất cả', onAction: () => context.go('/services')),
            ),
            SliverPadding(
              padding: const EdgeInsets.symmetric(horizontal: 12),
              sliver: SliverToBoxAdapter(
                child: Card(
                  child: Padding(
                    padding: const EdgeInsets.symmetric(vertical: 8, horizontal: 4),
                    child: GridView(
                      gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(crossAxisCount: 4, mainAxisExtent: 108),
                      shrinkWrap: true,
                      physics: const NeverScrollableScrollPhysics(),
                      padding: EdgeInsets.zero,
                      children: [
                        for (final app in featuredServices) _AppTile(app: app),
                        QuickAction(
                          icon: Icons.grid_view_rounded,
                          label: 'Tất cả',
                          color: Colors.blueGrey,
                          onTap: () => context.go('/services'),
                        ),
                      ],
                    ),
                  ),
                ),
              ),
            ),
            const SliverToBoxAdapter(child: _Promotions()),
            const SliverToBoxAdapter(child: _RecentServices()),
            const SliverToBoxAdapter(child: SizedBox(height: 32)),
          ],
        ),
      ),
    );
  }
}

class _AppTile extends StatelessWidget {
  const _AppTile({required this.app});

  final MiniApp app;

  @override
  Widget build(BuildContext context) =>
      QuickAction(icon: app.icon, label: app.title, color: app.color, onTap: () => context.push(app.path));
}

class _GlassButton extends StatelessWidget {
  const _GlassButton({required this.child, required this.onTap});

  final Widget child;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: Colors.white.withValues(alpha: 0.16),
      shape: CircleBorder(side: BorderSide(color: Colors.white.withValues(alpha: 0.25))),
      child: InkWell(
        customBorder: const CircleBorder(),
        onTap: onTap,
        child: SizedBox(width: 42, height: 42, child: Center(child: child)),
      ),
    );
  }
}

class _Header extends StatefulWidget {
  const _Header();

  @override
  State<_Header> createState() => _HeaderState();
}

class _HeaderState extends State<_Header> {
  bool _hidden = false;

  String _greeting() {
    final h = DateTime.now().hour;
    if (h < 11) return 'Chào buổi sáng';
    if (h < 14) return 'Chào buổi trưa';
    if (h < 18) return 'Chào buổi chiều';
    return 'Chào buổi tối';
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final name = context.select<AuthBloc, String>((b) => b.state.user?.displayName ?? '');
    final first = name.trim().isEmpty ? '' : name.trim().split(RegExp(r'\s+')).last;
    final top = MediaQuery.paddingOf(context).top;
    return Padding(
      padding: const EdgeInsets.only(bottom: 52),
      child: Stack(
        clipBehavior: Clip.none,
        children: [
          HeroBackdrop(
            borderRadius: const BorderRadius.vertical(bottom: Radius.circular(36)),
            child: Padding(
              padding: EdgeInsets.fromLTRB(20, top + 12, 20, 76),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [
                  Row(
                    children: [
                      const Flexible(
                        child: FittedBox(
                          fit: BoxFit.scaleDown,
                          alignment: Alignment.centerLeft,
                          child: KeikiLogo(size: 34, onDark: true, horizontal: true),
                        ),
                      ),
                      const Spacer(),
                      _GlassButton(
                        onTap: () => context.push('/notifications'),
                        child: const Icon(Icons.notifications_none_rounded, color: Colors.white),
                      ),
                      const SizedBox(width: 10),
                      _GlassButton(
                        onTap: () => context.go('/profile'),
                        child: Text(
                          name.isEmpty ? 'K' : name.characters.first.toUpperCase(),
                          style: const TextStyle(color: Colors.white, fontWeight: FontWeight.w800, fontSize: 16),
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 22),
                  Text(
                    first.isEmpty ? '${_greeting()} 👋' : '${_greeting()}, $first 👋',
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                    style: theme.textTheme.titleMedium?.copyWith(color: Colors.white.withValues(alpha: 0.92)),
                  ),
                  const SizedBox(height: 14),
                  BlocBuilder<WalletCubit, WalletState>(
                    builder: (context, state) =>
                        _BalanceCard(state: state, hidden: _hidden, onToggle: () => setState(() => _hidden = !_hidden)),
                  ),
                ],
              ),
            ),
          ),
          Positioned(
            left: 16,
            right: 16,
            bottom: -44,
            child: Material(
              color: theme.colorScheme.surface,
              elevation: 10,
              shadowColor: KeikiColors.brand.withValues(alpha: 0.25),
              borderRadius: BorderRadius.circular(24),
              child: Padding(
                padding: const EdgeInsets.symmetric(vertical: 6, horizontal: 4),
                child: Row(
                  children: [
                    Expanded(
                      child: QuickAction(icon: Icons.add_rounded, label: 'Nạp tiền', onTap: () => context.push('/wallet/topup')),
                    ),
                    Expanded(
                      child: QuickAction(
                        icon: Icons.north_east_rounded,
                        label: 'Chuyển',
                        color: KeikiColors.pink,
                        onTap: () => context.push('/wallet/transfer'),
                      ),
                    ),
                    Expanded(
                      child: QuickAction(
                        icon: Icons.qr_code_scanner_rounded,
                        label: 'Nhận tiền',
                        color: KeikiColors.accent,
                        onTap: () => context.push('/wallet/receive'),
                      ),
                    ),
                    Expanded(
                      child: QuickAction(
                        icon: Icons.receipt_long_rounded,
                        label: 'Lịch sử',
                        color: KeikiColors.warm,
                        onTap: () => context.push('/wallet/transactions'),
                      ),
                    ),
                  ],
                ),
              ),
            ),
          ),
        ],
      ),
    );
  }
}

class _BalanceCard extends StatelessWidget {
  const _BalanceCard({required this.state, required this.hidden, required this.onToggle});

  final WalletState state;
  final bool hidden;
  final VoidCallback onToggle;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final String balance;
    if (state.status.isLoading || state.status == LoadStatus.initial) {
      balance = '…';
    } else if (state.status.isFailure) {
      balance = 'Không tải được';
    } else if (!state.hasWallet) {
      balance = 'Chưa kích hoạt';
    } else {
      balance = hidden ? '•••••••' : formatMoneyString(state.balance?.available ?? state.wallet!.balance, state.wallet!.currency);
    }
    return Material(
      color: Colors.white.withValues(alpha: 0.14),
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(24),
        side: BorderSide(color: Colors.white.withValues(alpha: 0.28)),
      ),
      child: InkWell(
        borderRadius: BorderRadius.circular(24),
        onTap: () => context.push('/wallet'),
        child: Padding(
          padding: const EdgeInsets.fromLTRB(18, 14, 10, 16),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                children: [
                  const Icon(Icons.account_balance_wallet_rounded, color: Colors.white, size: 18),
                  const SizedBox(width: 8),
                  Expanded(
                    child: Text('Số dư Ví Keiki', style: theme.textTheme.bodyMedium?.copyWith(color: Colors.white.withValues(alpha: 0.85))),
                  ),
                  IconButton(
                    visualDensity: VisualDensity.compact,
                    tooltip: hidden ? 'Hiện số dư' : 'Ẩn số dư',
                    icon: Icon(hidden ? Icons.visibility_off_rounded : Icons.visibility_rounded, color: Colors.white, size: 20),
                    onPressed: onToggle,
                  ),
                ],
              ),
              FittedBox(
                fit: BoxFit.scaleDown,
                alignment: Alignment.centerLeft,
                child: Text(balance, style: theme.textTheme.headlineMedium?.copyWith(color: Colors.white, fontSize: 30)),
              ),
              if (state.points != null) ...[
                const SizedBox(height: 10),
                Container(
                  padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 5),
                  decoration: BoxDecoration(color: Colors.white.withValues(alpha: 0.18), borderRadius: BorderRadius.circular(999)),
                  child: Row(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      const Icon(Icons.stars_rounded, color: KeikiColors.sparkle, size: 16),
                      const SizedBox(width: 6),
                      Text(
                        '${state.points!.balance} điểm thưởng',
                        style: const TextStyle(color: Colors.white, fontWeight: FontWeight.w600, fontSize: 12),
                      ),
                    ],
                  ),
                ),
              ],
            ],
          ),
        ),
      ),
    );
  }
}

class _Promotions extends StatelessWidget {
  const _Promotions();

  static const _items = [
    (
      title: 'Điểm danh mỗi ngày',
      subtitle: 'Nhận lượt quay may mắn miễn phí',
      icon: Icons.casino_rounded,
      path: '/services/lucky/luckyCheckIn',
      colors: [Color(0xFFFF7043), Color(0xFFFFB74D)],
    ),
    (
      title: 'Mời bạn bè',
      subtitle: 'Cùng nhận thưởng khi bạn bè tham gia',
      icon: Icons.group_add_rounded,
      path: '/services/referral',
      colors: [Color(0xFF00BFA5), Color(0xFF26C6DA)],
    ),
    (
      title: 'Phim dành cho bạn',
      subtitle: 'Gợi ý theo sở thích của bạn',
      icon: Icons.movie_filter_rounded,
      path: '/services/movie/myMovieHome',
      colors: [Color(0xFF7C4DFF), Color(0xFFE040FB)],
    ),
  ];

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        const SectionHeader('Ưu đãi cho bạn'),
        SizedBox(
          height: 140,
          child: ListView.separated(
            scrollDirection: Axis.horizontal,
            padding: const EdgeInsets.symmetric(horizontal: 16),
            itemCount: _items.length,
            separatorBuilder: (_, _) => const SizedBox(width: 12),
            itemBuilder: (context, i) {
              final p = _items[i];
              return SizedBox(
                width: 280,
                child: GradientCard(
                  gradient: LinearGradient(colors: p.colors),
                  padding: const EdgeInsets.all(16),
                  onTap: () => context.push(p.path),
                  child: Row(
                    children: [
                      Expanded(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          mainAxisAlignment: MainAxisAlignment.center,
                          children: [
                            Text(
                              p.title,
                              maxLines: 1,
                              overflow: TextOverflow.ellipsis,
                              style: Theme.of(context).textTheme.titleMedium?.copyWith(color: Colors.white),
                            ),
                            const SizedBox(height: 4),
                            Text(
                              p.subtitle,
                              maxLines: 2,
                              overflow: TextOverflow.ellipsis,
                              style: const TextStyle(color: Colors.white70),
                            ),
                          ],
                        ),
                      ),
                      Icon(p.icon, color: Colors.white, size: 44),
                    ],
                  ),
                ),
              );
            },
          ),
        ),
      ],
    );
  }
}

class _RecentServices extends StatelessWidget {
  const _RecentServices();

  @override
  Widget build(BuildContext context) {
    final recent = context.select<ServicesCubit, List<String>>((c) => c.state.recent);
    if (recent.isEmpty) return const SizedBox.shrink();
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        const SectionHeader('Dùng gần đây'),
        Padding(
          padding: const EdgeInsets.symmetric(horizontal: 16),
          child: Card(
            child: Column(
              children: [
                for (final id in recent.take(4))
                  ListTile(
                    leading: IconBadge(icon: metaFor(id).icon, color: metaFor(id).category.color, size: 40),
                    title: Text(metaFor(id).title),
                    subtitle: Text(metaFor(id).subtitle.isEmpty ? metaFor(id).category.label : metaFor(id).subtitle),
                    trailing: const Icon(Icons.chevron_right),
                    onTap: () => context.push('/services/$id'),
                  ),
              ],
            ),
          ),
        ),
      ],
    );
  }
}
