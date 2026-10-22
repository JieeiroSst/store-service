import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

import '../../../core/utils/load_status.dart';
import '../../../core/widgets/status_view.dart';
import '../../../app/theme.dart';
import '../../../core/widgets/ui.dart';
import '../../auth/bloc/auth_bloc.dart';
import '../bloc/operation_cubit.dart';
import '../bloc/services_cubit.dart';
import '../data/humanize.dart';
import '../data/schema_catalog.dart';
import '../data/service_meta.dart';
import '../widgets/arg_form.dart';
import '../widgets/result_view.dart';
import '../../../app/role.dart';

class OperationPage extends StatelessWidget {
  const OperationPage({super.key, required this.serviceId, required this.operationName});

  final String serviceId;
  final String operationName;

  @override
  Widget build(BuildContext context) {
    return BlocBuilder<ServicesCubit, ServicesState>(
      builder: (context, state) {
        final found = state.catalog?.operation(serviceId, operationName);
        final op = found != null && found.access == OperationAccess.admin && !context.isAdmin ? null : found;
        if (state.catalog == null) return const Scaffold(body: SkeletonList());
        if (op == null) {
          return Scaffold(
            appBar: AppBar(),
            body: const EmptyView(message: 'Không tìm thấy chức năng', icon: Icons.help_outline),
          );
        }
        return BlocProvider(
          create: (context) => OperationCubit(context.read<ServicesCubit>().client, op),
          child: _OperationView(catalog: state.catalog!, operation: op, allowed: state.permissions.canRun(op)),
        );
      },
    );
  }
}

class _OperationView extends StatefulWidget {
  const _OperationView({required this.catalog, required this.operation, required this.allowed});

  final SchemaCatalog catalog;
  final GqlOperation operation;
  final bool allowed;

  @override
  State<_OperationView> createState() => _OperationViewState();
}

class _OperationViewState extends State<_OperationView> {
  final _form = GlobalKey<FormState>();
  final _scroll = ScrollController();
  final _resultKey = GlobalKey();
  late final Map<String, Object?> _values;

  GqlOperation get op => widget.operation;

  @override
  void initState() {
    super.initState();
    final userId = context.read<AuthBloc>().state.user?.id.toString();
    _values = {for (final a in op.args) a.name: defaultValueFor(a, userId: userId)};
    if (!op.isMutation && !op.args.any((a) => a.type.nonNull && _values[a.name] == null)) {
      WidgetsBinding.instance.addPostFrameCallback((_) => _run());
    }
  }

  @override
  void dispose() {
    _scroll.dispose();
    super.dispose();
  }

  Future<void> _run() async {
    if (!(_form.currentState?.validate() ?? true)) return;
    if (op.isMutation) {
      final ok = await showDialog<bool>(
        context: context,
        builder: (context) => AlertDialog(
          icon: Icon(operationIcon(op)),
          title: Text(operationTitle(op)),
          content: const Text('Bạn có chắc muốn thực hiện thao tác này?'),
          actions: [
            TextButton(onPressed: () => Navigator.pop(context, false), child: const Text('Huỷ')),
            FilledButton(onPressed: () => Navigator.pop(context, true), child: const Text('Thực hiện')),
          ],
        ),
      );
      if (ok != true || !mounted) return;
    }
    await context.read<OperationCubit>().run(Map.of(_values));
    if (!mounted) return;
    final target = _resultKey.currentContext;
    if (target != null && target.mounted) {
      Scrollable.ensureVisible(target, duration: const Duration(milliseconds: 350), curve: Curves.easeOut);
    }
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final meta = metaFor(op.service);
    final required = op.args.where((a) => a.type.nonNull).toList();
    final optional = op.args.where((a) => !a.type.nonNull).toList();
    Widget editor(GqlField a) => Padding(
      padding: const EdgeInsets.only(bottom: 14),
      child: ArgEditor(catalog: widget.catalog, field: a, value: _values[a.name], onChanged: (v) => setState(() => _values[a.name] = v)),
    );
    return Scaffold(
      appBar: AppBar(title: Text(meta.title)),
      bottomNavigationBar: BottomActionBar(
        child: BlocBuilder<OperationCubit, OperationState>(
          builder: (context, state) => LoadingButton(
            label: op.isMutation ? 'Thực hiện' : (state.runs == 0 ? 'Xem kết quả' : 'Tải lại'),
            icon: op.isMutation ? Icons.bolt : Icons.play_arrow_rounded,
            loading: state.status.isLoading,
            onPressed: widget.allowed ? _run : null,
          ),
        ),
      ),
      body: Form(
        key: _form,
        child: SingleChildScrollView(
          controller: _scroll,
          padding: const EdgeInsets.fromLTRB(16, 0, 16, 24),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              Row(
                children: [
                  IconBadge(icon: operationIcon(op), color: meta.category.color, size: 52),
                  const SizedBox(width: 14),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(operationTitle(op), style: theme.textTheme.titleLarge),
                        const SizedBox(height: 4),
                        Wrap(
                          spacing: 6,
                          runSpacing: 4,
                          children: [
                            Pill(op.isMutation ? 'Thực hiện' : 'Xem', color: op.isMutation ? KeikiColors.warm : meta.category.color),
                            if (op.access == OperationAccess.admin)
                              Pill('Quản trị', color: theme.colorScheme.tertiary, icon: Icons.shield_outlined),
                            if (op.access == OperationAccess.public) const Pill('Không cần đăng nhập', color: Colors.green),
                          ],
                        ),
                      ],
                    ),
                  ),
                ],
              ),
              if (!widget.allowed)
                Padding(
                  padding: const EdgeInsets.only(top: 16),
                  child: Card(
                    color: theme.colorScheme.errorContainer,
                    child: const ListTile(
                      leading: Icon(Icons.lock_outline),
                      title: Text('Tài khoản của bạn chưa được cấp quyền cho chức năng này'),
                    ),
                  ),
                ),
              if (op.args.isNotEmpty) ...[
                if (required.isNotEmpty)
                  const SectionHeader('Thông tin cần nhập', padding: EdgeInsets.fromLTRB(4, 24, 4, 12))
                else
                  const SizedBox(height: 16),
                for (final a in required) editor(a),
                if (optional.isNotEmpty)
                  Theme(
                    data: theme.copyWith(dividerColor: Colors.transparent),
                    child: ExpansionTile(
                      tilePadding: const EdgeInsets.symmetric(horizontal: 4),
                      initiallyExpanded: required.isNotEmpty && optional.length <= 3,
                      leading: Icon(required.isEmpty ? Icons.tune : Icons.add_circle_outline),
                      title: Text(
                        required.isEmpty ? 'Bộ lọc (${optional.length})' : 'Tuỳ chọn thêm (${optional.length})',
                        style: theme.textTheme.titleSmall,
                      ),
                      children: [for (final a in optional) editor(a)],
                    ),
                  ),
              ],
              BlocBuilder<OperationCubit, OperationState>(
                builder: (context, state) {
                  if (state.status == LoadStatus.initial) return const SizedBox.shrink();
                  return Padding(
                    key: _resultKey,
                    padding: const EdgeInsets.only(top: 24),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.stretch,
                      children: [
                        Row(
                          children: [
                            Text('Kết quả', style: theme.textTheme.titleMedium),
                            const Spacer(),
                            if (state.elapsed != null)
                              Text(
                                '${state.elapsed!.inMilliseconds} ms',
                                style: theme.textTheme.labelSmall?.copyWith(color: theme.colorScheme.onSurfaceVariant),
                              ),
                          ],
                        ),
                        const SizedBox(height: 12),
                        if (state.status.isLoading)
                          const SizedBox(height: 200, child: SkeletonList(count: 3))
                        else if (state.status.isFailure)
                          Card(
                            color: theme.colorScheme.errorContainer.withValues(alpha: 0.6),
                            child: ListTile(
                              leading: Icon(Icons.error_outline, color: theme.colorScheme.error),
                              title: const Text('Không thực hiện được'),
                              subtitle: Text(state.error ?? ''),
                            ),
                          )
                        else
                          ResultView(value: state.result),
                      ],
                    ),
                  );
                },
              ),
            ],
          ),
        ),
      ),
    );
  }
}
