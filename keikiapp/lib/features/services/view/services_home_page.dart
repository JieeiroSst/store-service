import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:go_router/go_router.dart';

import '../../../core/utils/load_status.dart';
import '../../../core/widgets/status_view.dart';
import '../../../app/theme.dart';
import '../../../core/widgets/ui.dart';
import '../bloc/services_cubit.dart';
import '../data/humanize.dart';
import '../data/schema_catalog.dart';
import '../data/service_meta.dart';
import '../../../app/role.dart';

class ServicesHomePage extends StatefulWidget {
  const ServicesHomePage({super.key});

  @override
  State<ServicesHomePage> createState() => _ServicesHomePageState();
}

class _ServicesHomePageState extends State<ServicesHomePage> {
  String _query = '';
  ServiceCategory? _category;

  @override
  void initState() {
    super.initState();
    context.read<ServicesCubit>().load();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: BlocBuilder<ServicesCubit, ServicesState>(
        builder: (context, state) {
          final cubit = context.read<ServicesCubit>();
          return CustomScrollView(
            slivers: [
              SliverAppBar.medium(title: const Text('Tất cả dịch vụ')),
              SliverPadding(
                padding: const EdgeInsets.fromLTRB(16, 0, 16, 8),
                sliver: SliverToBoxAdapter(
                  child: SearchBar(
                    hintText: 'Tìm dịch vụ hoặc chức năng…',
                    leading: const Icon(Icons.search),
                    elevation: const WidgetStatePropertyAll(0),
                    onChanged: (q) => setState(() => _query = q.trim().toLowerCase()),
                  ),
                ),
              ),
              if (state.catalog == null)
                SliverFillRemaining(
                  child: state.status.isFailure
                      ? ErrorView(message: state.error ?? '', onRetry: cubit.load)
                      : const SkeletonList(),
                )
              else if (_query.isNotEmpty)
                ..._searchResults(context, state)
              else
                ..._browse(context, state),
              const SliverToBoxAdapter(child: SizedBox(height: 32)),
            ],
          );
        },
      ),
    );
  }

  List<Widget> _browse(BuildContext context, ServicesState state) {
    final admin = context.isAdmin;
    final services = state.visibleServices.where((s) => admin || !adminOnlyApps.contains(s.id)).toList();
    final userServices = services.where((s) => state.visibleOperations(s).any((o) => o.access != OperationAccess.admin)).toList();
    final adminServices = admin ? services.where((s) => !userServices.contains(s)).toList() : const <GqlService>[];
    final categories = ServiceCategory.values.where((c) => userServices.any((s) => metaFor(s.id).category == c)).toList();
    final filtered = _category == null ? userServices : userServices.where((s) => metaFor(s.id).category == _category).toList();
    final recent = [for (final id in state.recent) ...services.where((s) => s.id == id)];
    final total = services.fold<int>(0, (n, s) => n + state.visibleOperations(s).where((o) => admin || o.access != OperationAccess.admin).length);
    return [
      SliverToBoxAdapter(
        child: Padding(
          padding: const EdgeInsets.fromLTRB(20, 4, 20, 0),
          child: Text(
            '${userServices.length} dịch vụ • $total chức năng bạn dùng được',
            style: Theme.of(context).textTheme.bodySmall?.copyWith(color: Theme.of(context).colorScheme.onSurfaceVariant),
          ),
        ),
      ),
      if (recent.isNotEmpty) ...[
        const SliverToBoxAdapter(child: SectionHeader('Dùng gần đây')),
        SliverToBoxAdapter(
          child: SizedBox(
            height: 104,
            child: ListView(
              scrollDirection: Axis.horizontal,
              padding: const EdgeInsets.symmetric(horizontal: 12),
              children: [
                for (final s in recent)
                  SizedBox(
                    width: 84,
                    child: QuickAction(
                      icon: metaFor(s.id).icon,
                      label: metaFor(s.id).title,
                      color: metaFor(s.id).category.color,
                      onTap: () => context.push('/services/${s.id}'),
                    ),
                  ),
              ],
            ),
          ),
        ),
      ],
      SliverToBoxAdapter(
        child: SizedBox(
          height: 56,
          child: ListView(
            scrollDirection: Axis.horizontal,
            padding: const EdgeInsets.fromLTRB(12, 12, 12, 0),
            children: [
              Padding(
                padding: const EdgeInsets.symmetric(horizontal: 4),
                child: ChoiceChip(label: const Text('Tất cả'), selected: _category == null, onSelected: (_) => setState(() => _category = null)),
              ),
              for (final c in categories)
                Padding(
                  padding: const EdgeInsets.symmetric(horizontal: 4),
                  child: ChoiceChip(
                    avatar: Icon(c.icon, size: 18),
                    label: Text(c.label),
                    selected: _category == c,
                    onSelected: (_) => setState(() => _category = c),
                  ),
                ),
            ],
          ),
        ),
      ),
      for (final c in (_category == null ? categories : [_category!])) ...[
        SliverToBoxAdapter(child: SectionHeader(c.label)),
        SliverPadding(
          padding: const EdgeInsets.symmetric(horizontal: 16),
          sliver: SliverGrid(
            gridDelegate: const SliverGridDelegateWithMaxCrossAxisExtent(maxCrossAxisExtent: 220, mainAxisExtent: 128, mainAxisSpacing: 12, crossAxisSpacing: 12),
            delegate: SliverChildListDelegate([
              for (final s in filtered.where((s) => metaFor(s.id).category == c)) ServiceCard(service: s, operations: state.visibleOperations(s).length),
            ]),
          ),
        ),
      ],
      if (adminServices.isNotEmpty && _category == null) ...[
        const SliverToBoxAdapter(child: SectionHeader('Quản trị')),
        SliverPadding(
          padding: const EdgeInsets.symmetric(horizontal: 16),
          sliver: SliverGrid(
            gridDelegate: const SliverGridDelegateWithMaxCrossAxisExtent(maxCrossAxisExtent: 220, mainAxisExtent: 128, mainAxisSpacing: 12, crossAxisSpacing: 12),
            delegate: SliverChildListDelegate([
              for (final s in adminServices) ServiceCard(service: s, operations: state.visibleOperations(s).length, admin: true),
            ]),
          ),
        ),
      ],
    ];
  }

  List<Widget> _searchResults(BuildContext context, ServicesState state) {
    bool hit(String s) => s.toLowerCase().contains(_query);
    final services = state.visibleServices.where((s) {
      final m = metaFor(s.id);
      return hit(m.title) || hit(m.subtitle) || hit(s.id) || hit(m.category.label);
    }).toList();
    final ops = [
      for (final s in state.visibleServices)
        for (final o in state.visibleOperations(s))
          if (hit(operationTitle(o)) || hit(o.name)) o,
    ].take(60).toList();
    if (services.isEmpty && ops.isEmpty) {
      return [const SliverFillRemaining(hasScrollBody: false, child: EmptyView(message: 'Không tìm thấy kết quả', icon: Icons.search_off))];
    }
    return [
      if (services.isNotEmpty) ...[
        SliverToBoxAdapter(child: SectionHeader('Dịch vụ (${services.length})')),
        SliverList.list(children: [for (final s in services) _ServiceTile(service: s, count: state.visibleOperations(s).length)]),
      ],
      if (ops.isNotEmpty) ...[
        SliverToBoxAdapter(child: SectionHeader('Chức năng (${ops.length})')),
        SliverList.list(children: [for (final o in ops) OperationTile(operation: o, showService: true)]),
      ],
    ];
  }
}

class ServiceCard extends StatelessWidget {
  const ServiceCard({super.key, required this.service, required this.operations, this.admin = false});

  final GqlService service;
  final int operations;
  final bool admin;

  @override
  Widget build(BuildContext context) {
    final meta = metaFor(service.id);
    final theme = Theme.of(context);
    return Card(
      clipBehavior: Clip.antiAlias,
      child: InkWell(
        onTap: () => context.push('/services/${service.id}'),
        child: Padding(
          padding: const EdgeInsets.all(14),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                children: [
                  IconBadge(icon: meta.icon, color: meta.category.color, size: 40),
                  const Spacer(),
                  if (admin) Pill('Quản trị', color: theme.colorScheme.tertiary),
                ],
              ),
              const Spacer(),
              Text(meta.title, maxLines: 1, overflow: TextOverflow.ellipsis, style: theme.textTheme.titleSmall),
              Text(
                meta.subtitle.isNotEmpty ? meta.subtitle : '$operations chức năng',
                maxLines: 1,
                overflow: TextOverflow.ellipsis,
                style: theme.textTheme.bodySmall?.copyWith(color: theme.colorScheme.onSurfaceVariant),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class _ServiceTile extends StatelessWidget {
  const _ServiceTile({required this.service, required this.count});

  final GqlService service;
  final int count;

  @override
  Widget build(BuildContext context) {
    final meta = metaFor(service.id);
    return ListTile(
      contentPadding: const EdgeInsets.symmetric(horizontal: 20),
      leading: IconBadge(icon: meta.icon, color: meta.category.color, size: 40),
      title: Text(meta.title),
      subtitle: Text('${meta.category.label} • $count chức năng'),
      trailing: const Icon(Icons.chevron_right),
      onTap: () => context.push('/services/${service.id}'),
    );
  }
}

class OperationTile extends StatelessWidget {
  const OperationTile({super.key, required this.operation, this.showService = false});

  final GqlOperation operation;
  final bool showService;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final meta = metaFor(operation.service);
    final args = operation.args.length;
    return ListTile(
      contentPadding: const EdgeInsets.symmetric(horizontal: 20, vertical: 2),
      leading: IconBadge(icon: operationIcon(operation), color: operation.isMutation ? KeikiColors.warm : meta.category.color, size: 40),
      title: Text(operationTitle(operation)),
      subtitle: Text(
        [
          if (showService) meta.title,
          operation.isMutation ? 'Thực hiện' : 'Xem',
          if (args > 0) '$args tham số',
        ].join(' • '),
      ),
      trailing: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          if (operation.access == OperationAccess.admin) Pill('Quản trị', color: theme.colorScheme.tertiary),
          if (operation.access == OperationAccess.public) Pill('Công khai', color: Colors.green),
          const Icon(Icons.chevron_right),
        ],
      ),
      onTap: () => context.push('/services/${operation.service}/${operation.name}'),
    );
  }
}
