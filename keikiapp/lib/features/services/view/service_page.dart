import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

import '../../../core/widgets/status_view.dart';
import '../../../core/widgets/ui.dart';
import '../bloc/services_cubit.dart';
import '../data/humanize.dart';
import '../data/schema_catalog.dart';
import '../data/service_meta.dart';
import 'services_home_page.dart';
import '../../../app/role.dart';

class ServicePage extends StatefulWidget {
  const ServicePage({super.key, required this.serviceId});

  final String serviceId;

  @override
  State<ServicePage> createState() => _ServicePageState();
}

class _ServicePageState extends State<ServicePage> {
  String _filter = '';

  @override
  void initState() {
    super.initState();
    final cubit = context.read<ServicesCubit>()..load();
    cubit.visited(widget.serviceId);
  }

  @override
  Widget build(BuildContext context) {
    final meta = metaFor(widget.serviceId);
    final theme = Theme.of(context);
    return Scaffold(
      body: BlocBuilder<ServicesCubit, ServicesState>(
        builder: (context, state) {
          final service = state.catalog?.service(widget.serviceId);
          final admin = context.isAdmin;
          final ops = service == null || (!admin && adminOnlyApps.contains(widget.serviceId))
              ? const <GqlOperation>[]
              : state.visibleOperations(service).where((o) => admin || o.access != OperationAccess.admin).toList();
          final shown = _filter.isEmpty
              ? ops
              : ops.where((o) => operationTitle(o).toLowerCase().contains(_filter) || o.name.toLowerCase().contains(_filter)).toList();
          final queries = shown.where((o) => !o.isMutation).toList();
          final mutations = shown.where((o) => o.isMutation).toList();
          return CustomScrollView(
            slivers: [
              SliverAppBar.large(
                title: Text(meta.title),
                flexibleSpace: FlexibleSpaceBar(
                  background: DecoratedBox(
                    decoration: BoxDecoration(
                      gradient: LinearGradient(
                        begin: Alignment.topCenter,
                        end: Alignment.bottomCenter,
                        colors: [meta.category.color.withValues(alpha: 0.18), theme.scaffoldBackgroundColor],
                      ),
                    ),
                  ),
                ),
              ),
              if (state.catalog == null)
                const SliverFillRemaining(child: SkeletonList())
              else if (service == null || ops.isEmpty)
                const SliverFillRemaining(hasScrollBody: false, child: EmptyView(message: 'Bạn chưa có quyền dùng dịch vụ này', icon: Icons.lock_outline))
              else ...[
                SliverToBoxAdapter(
                  child: Padding(
                    padding: const EdgeInsets.fromLTRB(20, 0, 20, 8),
                    child: Row(
                      children: [
                        IconBadge(icon: meta.icon, color: meta.category.color, size: 56),
                        const SizedBox(width: 16),
                        Expanded(
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              if (meta.subtitle.isNotEmpty) Text(meta.subtitle, style: theme.textTheme.bodyLarge),
                              const SizedBox(height: 6),
                              Wrap(spacing: 6, children: [
                                Pill(meta.category.label, color: meta.category.color),
                                Pill('${ops.length} chức năng'),
                              ]),
                            ],
                          ),
                        ),
                      ],
                    ),
                  ),
                ),
                if (ops.length > 8)
                  SliverPadding(
                    padding: const EdgeInsets.fromLTRB(16, 8, 16, 0),
                    sliver: SliverToBoxAdapter(
                      child: TextField(
                        decoration: const InputDecoration(hintText: 'Lọc chức năng', prefixIcon: Icon(Icons.filter_list)),
                        onChanged: (v) => setState(() => _filter = v.trim().toLowerCase()),
                      ),
                    ),
                  ),
                if (queries.isNotEmpty) ...[
                  SliverToBoxAdapter(child: SectionHeader('Xem thông tin (${queries.length})')),
                  SliverList.list(children: [for (final o in queries) OperationTile(operation: o)]),
                ],
                if (mutations.isNotEmpty) ...[
                  SliverToBoxAdapter(child: SectionHeader('Thực hiện (${mutations.length})')),
                  SliverList.list(children: [for (final o in mutations) OperationTile(operation: o)]),
                ],
                const SliverToBoxAdapter(child: SizedBox(height: 32)),
              ],
            ],
          );
        },
      ),
    );
  }
}
