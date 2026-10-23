import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

import '../../apps/kit/kit.dart';
import '../../services/bloc/services_cubit.dart';
import '../../services/data/humanize.dart';
import '../../services/data/service_meta.dart';
import '../../services/widgets/result_view.dart';
import 'admin_op.dart';
import 'gen/admin_consoles.g.dart';

const _verbIcons = <String, IconData>{
  'create': Icons.add_circle_outline,
  'add': Icons.add_circle_outline,
  'register': Icons.add_circle_outline,
  'upsert': Icons.add_circle_outline,
  'put': Icons.upload_outlined,
  'update': Icons.edit_outlined,
  'edit': Icons.edit_outlined,
  'set': Icons.tune,
  'change': Icons.tune,
  'rename': Icons.drive_file_rename_outline,
  'delete': Icons.delete_outline,
  'remove': Icons.remove_circle_outline,
  'revoke': Icons.remove_circle_outline,
  'cancel': Icons.cancel_outlined,
  'approve': Icons.verified_outlined,
  'reject': Icons.block_outlined,
  'block': Icons.block_outlined,
  'unblock': Icons.lock_open_outlined,
  'activate': Icons.play_circle_outline,
  'enable': Icons.toggle_on_outlined,
  'disable': Icons.toggle_off_outlined,
  'deactivate': Icons.pause_circle_outline,
  'sync': Icons.sync,
  'refresh': Icons.refresh,
  'publish': Icons.publish_outlined,
  'send': Icons.send_outlined,
  'assign': Icons.assignment_ind_outlined,
  'grant': Icons.card_giftcard_outlined,
  'rotate': Icons.autorenew,
  'export': Icons.download_outlined,
  'run': Icons.play_arrow_outlined,
  'trigger': Icons.bolt_outlined,
  'search': Icons.search,
  'find': Icons.search,
  'count': Icons.numbers,
  'check': Icons.fact_check_outlined,
  'verify': Icons.fact_check_outlined,
};

IconData _iconFor(AdminOp op) => _verbIcons[op.verb] ?? (op.mutation ? Icons.bolt_outlined : Icons.list_alt_outlined);

String adminOpTitle(BuildContext context, String serviceId, AdminOp op) {
  final o = context.read<ServicesCubit>().state.catalog?.operation(serviceId, op.name);
  return o == null ? fieldLabel(op.name) : operationTitle(o);
}

int adminConsoleSize(String serviceId) => adminConsoles[serviceId]?.length ?? 0;

class AdminConsolePage extends StatefulWidget {
  const AdminConsolePage({super.key, required this.serviceId});

  final String serviceId;

  @override
  State<AdminConsolePage> createState() => _AdminConsolePageState();
}

class _AdminConsolePageState extends State<AdminConsolePage> {
  String _filter = '';
  bool? _writes;

  @override
  Widget build(BuildContext context) {
    final meta = metaFor(widget.serviceId);
    final color = meta.category.color;
    final state = context.watch<ServicesCubit>().state;
    final service = state.catalog?.service(widget.serviceId);
    final allowed = service == null ? null : state.visibleOperations(service).map((o) => o.name).toSet();
    final generated = adminConsoles[widget.serviceId] ?? const <AdminOp>[];
    final all = [for (final op in generated) if (allowed == null || allowed.contains(op.name)) op];
    final titles = {for (final op in all) op.name: adminOpTitle(context, widget.serviceId, op)};
    final q = _filter.toLowerCase();
    final shown = all
        .where((op) => _writes == null || op.mutation == _writes)
        .where((op) => q.isEmpty || titles[op.name]!.toLowerCase().contains(q) || op.name.toLowerCase().contains(q) || resourceTitle(op.resource).toLowerCase().contains(q))
        .toList();
    final groups = <String, List<AdminOp>>{};
    for (final op in shown) {
      groups.putIfAbsent(op.resource, () => []).add(op);
    }
    final reads = all.where((o) => !o.mutation).length;
    return appScope(
      context,
      color,
      Scaffold(
        body: CustomScrollView(
          slivers: [
            SliverAppBar.large(title: Text('Quản trị ${meta.title}')),
            SliverToBoxAdapter(
              child: Padding(
                padding: const EdgeInsets.fromLTRB(16, 0, 16, 8),
                child: StatRow([
                  StatTile(label: 'Chức năng', value: '${all.length}', icon: Icons.apps_outlined),
                  StatTile(label: 'Tra cứu', value: '$reads', icon: Icons.manage_search),
                  StatTile(label: 'Thao tác', value: '${all.length - reads}', icon: Icons.bolt_outlined),
                ]),
              ),
            ),
            SliverToBoxAdapter(
              child: Padding(
                padding: const EdgeInsets.fromLTRB(16, 8, 16, 0),
                child: TextField(
                  decoration: const InputDecoration(prefixIcon: Icon(Icons.search), hintText: 'Tìm chức năng quản trị'),
                  onChanged: (v) => setState(() => _filter = v.trim()),
                ),
              ),
            ),
            SliverToBoxAdapter(
              child: Padding(
                padding: const EdgeInsets.fromLTRB(16, 12, 16, 0),
                child: Wrap(
                  spacing: 8,
                  children: [
                    for (final (v, l) in const [(null, 'Tất cả'), (false, 'Tra cứu'), (true, 'Thao tác')])
                      ChoiceChip(label: Text(l), selected: _writes == v, onSelected: (_) => setState(() => _writes = v)),
                  ],
                ),
              ),
            ),
            if (all.isEmpty && generated.isNotEmpty)
              const SliverFillRemaining(
                hasScrollBody: false,
                child: Center(
                  child: Padding(
                    padding: EdgeInsets.all(32),
                    child: Text('Tài khoản chưa được cấp quyền cho chức năng quản trị nào của dịch vụ này. Liên hệ quản trị viên cấp quyền trong authorize-service.', textAlign: TextAlign.center),
                  ),
                ),
              )
            else if (groups.isEmpty)
              const SliverFillRemaining(
                hasScrollBody: false,
                child: Center(child: Padding(padding: EdgeInsets.all(32), child: Text('Không có chức năng phù hợp'))),
              ),
            for (final entry in groups.entries) ...[
              SliverToBoxAdapter(child: SectionHeader('${resourceTitle(entry.key)} · ${entry.value.length}')),
              SliverPadding(
                padding: const EdgeInsets.symmetric(horizontal: 16),
                sliver: SliverToBoxAdapter(
                  child: KCard(
                    padding: EdgeInsets.zero,
                    child: Column(
                      children: [
                        for (final (i, op) in entry.value.indexed) ...[
                          if (i > 0) const Divider(height: 1, indent: 68),
                          _OpTile(serviceId: widget.serviceId, op: op, title: titles[op.name]!, color: color),
                        ],
                      ],
                    ),
                  ),
                ),
              ),
            ],
            const SliverToBoxAdapter(child: SizedBox(height: 32)),
          ],
        ),
      ),
    );
  }
}

class _OpTile extends StatelessWidget {
  const _OpTile({required this.serviceId, required this.op, required this.title, required this.color});

  final String serviceId;
  final AdminOp op;
  final String title;
  final Color color;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final tone = op.destructive ? theme.colorScheme.error : (op.mutation ? color : theme.colorScheme.secondary);
    return ListTile(
      leading: IconBadge(icon: _iconFor(op), color: tone, size: 40),
      title: Text(title, maxLines: 2, overflow: TextOverflow.ellipsis),
      subtitle: Text(op.name, maxLines: 1, overflow: TextOverflow.ellipsis, style: theme.textTheme.bodySmall?.copyWith(fontFamily: 'monospace')),
      trailing: Pill(op.mutation ? 'Thao tác' : 'Tra cứu', color: tone),
      onTap: () => openPage(context, AdminOpPage(serviceId: serviceId, op: op, title: title)),
    );
  }
}

class AdminOpPage extends StatefulWidget {
  const AdminOpPage({super.key, required this.serviceId, required this.op, required this.title});

  final String serviceId;
  final AdminOp op;
  final String title;

  @override
  State<AdminOpPage> createState() => _AdminOpPageState();
}

class _AdminOpPageState extends State<AdminOpPage> {
  Map<String, String> _last = const {};
  Object? _result;
  Object? _error;
  bool _ran = false;
  bool _busy = false;

  AdminOp get op => widget.op;

  @override
  void initState() {
    super.initState();
    if (!op.mutation && !op.hasRequired) WidgetsBinding.instance.addPostFrameCallback((_) => _execute(const FormData({})));
  }

  Future<void> _execute(FormData v) async {
    final api = context.api;
    setState(() {
      _busy = true;
      _error = null;
    });
    try {
      final r = await op.run(api, v);
      if (!mounted) return;
      setState(() {
        _result = r;
        _ran = true;
        _last = v.values;
      });
    } catch (e) {
      if (mounted) setState(() => _error = e);
      if (op.mutation) rethrow;
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  Future<void> _start() async {
    if (op.fields.isEmpty) {
      if (op.mutation) {
        final ok = await runAction(
          context,
          () => _execute(const FormData({})),
          confirm: op.destructive ? 'Thực hiện "${widget.title}"? Thao tác này có thể không hoàn tác được.' : null,
          destructive: op.destructive,
          success: 'Đã thực hiện',
        );
        if (!ok) return;
      } else {
        await _execute(const FormData({}));
      }
      return;
    }
    await showFormSheet(
      context,
      title: widget.title,
      description: op.destructive ? 'Thao tác này có thể không hoàn tác được.' : null,
      submitLabel: op.mutation ? 'Thực hiện' : 'Tra cứu',
      success: op.mutation ? 'Đã thực hiện' : null,
      fields: [
        for (final f in op.fields)
          KField(f.key, f.label, kind: f.kind, required: f.required, initial: _last[f.key] ?? f.initial, hint: f.hint, options: f.options, icon: f.icon),
      ],
      onSubmit: (v) async {
        await _execute(v);
        if (!op.mutation && _error != null) throw _error!;
      },
    );
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final color = context.appColor;
    final filled = _last.entries.where((e) => e.value.trim().isNotEmpty).toList();
    final labels = {for (final f in op.fields) f.key: f.label};
    return DetailPage(
      title: widget.title,
      actions: [
        if (!op.mutation && _ran) IconButton(tooltip: 'Tải lại', icon: const Icon(Icons.refresh), onPressed: _busy ? null : () => _execute(FormData(_last))),
      ],
      header: AppBanner(
        icon: _iconFor(op),
        title: widget.title,
        subtitle: '${metaFor(widget.serviceId).title} · ${op.mutation ? 'Thao tác ghi' : 'Tra cứu'} · ${op.name}',
      ),
      children: [
        if (filled.isNotEmpty)
          Padding(
            padding: const EdgeInsets.fromLTRB(16, 16, 16, 0),
            child: Wrap(
              spacing: 8,
              runSpacing: 8,
              children: [for (final e in filled) Chip(label: Text('${labels[e.key] ?? e.key}: ${e.value}', maxLines: 1, overflow: TextOverflow.ellipsis))],
            ),
          ),
        ActionButtons([
          FilledButton.icon(
            style: op.destructive ? FilledButton.styleFrom(backgroundColor: theme.colorScheme.error) : null,
            icon: Icon(op.mutation ? _iconFor(op) : (op.fields.isEmpty ? Icons.refresh : Icons.filter_alt_outlined)),
            label: Text(op.mutation ? 'Thực hiện' : (op.fields.isEmpty ? 'Tải lại' : (_ran ? 'Đổi điều kiện' : 'Nhập điều kiện tra cứu'))),
            onPressed: _busy ? null : _start,
          ),
        ]),
        Padding(
          padding: const EdgeInsets.all(16),
          child: _busy
              ? const SkeletonList(count: 3)
              : _error != null
                  ? KCard(
                      color: theme.colorScheme.errorContainer,
                      child: Row(
                        children: [
                          Icon(Icons.error_outline, color: theme.colorScheme.onErrorContainer),
                          const SizedBox(width: 12),
                          Expanded(child: Text(errorMessage(_error!), style: TextStyle(color: theme.colorScheme.onErrorContainer))),
                        ],
                      ),
                    )
                  : !_ran
                      ? Text(
                          op.fields.isEmpty ? 'Bấm nút phía trên để thực hiện.' : 'Có ${op.fields.length} trường, ${op.fields.where((f) => f.required).length} trường bắt buộc.',
                          style: theme.textTheme.bodyMedium?.copyWith(color: theme.colorScheme.onSurfaceVariant),
                        )
                      : _result == null
                          ? Row(children: [Icon(Icons.check_circle, color: color), const SizedBox(width: 8), const Text('Hoàn tất, máy chủ không trả dữ liệu')])
                          : ResultView(value: _result),
        ),
      ],
    );
  }
}
