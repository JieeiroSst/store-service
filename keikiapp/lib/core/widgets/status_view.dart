import 'package:flutter/material.dart';

import '../utils/load_status.dart';
import 'ui.dart';

class StatusView extends StatelessWidget {
  const StatusView({
    super.key,
    required this.status,
    required this.builder,
    this.error,
    this.onRetry,
    this.isEmpty = false,
    this.emptyMessage = 'Chưa có dữ liệu',
  });

  final LoadStatus status;
  final WidgetBuilder builder;
  final String? error;
  final VoidCallback? onRetry;
  final bool isEmpty;
  final String emptyMessage;

  @override
  Widget build(BuildContext context) {
    switch (status) {
      case LoadStatus.initial:
      case LoadStatus.loading:
        return const SkeletonList();
      case LoadStatus.failure:
        return ErrorView(message: error ?? 'Đã có lỗi xảy ra', onRetry: onRetry);
      case LoadStatus.success:
        if (isEmpty) return EmptyView(message: emptyMessage, onRefresh: onRetry);
        return builder(context);
    }
  }
}

class ErrorView extends StatelessWidget {
  const ErrorView({super.key, required this.message, this.onRetry});

  final String message;
  final VoidCallback? onRetry;

  @override
  Widget build(BuildContext context) {
    return _Illustrated(
      icon: Icons.cloud_off_outlined,
      color: Theme.of(context).colorScheme.error,
      title: 'Không tải được',
      message: message,
      action: onRetry == null ? null : FilledButton.tonalIcon(onPressed: onRetry, icon: const Icon(Icons.refresh), label: const Text('Thử lại')),
    );
  }
}

class EmptyView extends StatelessWidget {
  const EmptyView({super.key, required this.message, this.onRefresh, this.icon = Icons.inbox_outlined});

  final String message;
  final VoidCallback? onRefresh;
  final IconData icon;

  @override
  Widget build(BuildContext context) {
    return _Illustrated(
      icon: icon,
      color: Theme.of(context).colorScheme.primary,
      title: message,
      action: onRefresh == null ? null : TextButton.icon(onPressed: onRefresh, icon: const Icon(Icons.refresh), label: const Text('Tải lại')),
    );
  }
}

class _Illustrated extends StatelessWidget {
  const _Illustrated({required this.icon, required this.color, required this.title, this.message, this.action});

  final IconData icon;
  final Color color;
  final String title;
  final String? message;
  final Widget? action;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Center(
      child: SingleChildScrollView(
        padding: const EdgeInsets.all(32),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Container(
              width: 96,
              height: 96,
              decoration: BoxDecoration(color: color.withValues(alpha: 0.1), shape: BoxShape.circle),
              child: Icon(icon, size: 44, color: color),
            ),
            const SizedBox(height: 20),
            Text(title, textAlign: TextAlign.center, style: theme.textTheme.titleMedium),
            if (message != null) ...[
              const SizedBox(height: 8),
              Text(message!, textAlign: TextAlign.center, style: theme.textTheme.bodyMedium?.copyWith(color: theme.colorScheme.onSurfaceVariant)),
            ],
            if (action != null) ...[const SizedBox(height: 20), action!],
          ],
        ),
      ),
    );
  }
}

void showSnack(BuildContext context, String message) {
  ScaffoldMessenger.of(context)
    ..hideCurrentSnackBar()
    ..showSnackBar(SnackBar(content: Text(message)));
}

Future<String?> askPin(BuildContext context) {
  final controller = TextEditingController();
  return showDialog<String>(
    context: context,
    builder: (context) => AlertDialog(
      title: const Text('Nhập mã PIN'),
      content: TextField(
        controller: controller,
        autofocus: true,
        obscureText: true,
        keyboardType: TextInputType.number,
        maxLength: 6,
        decoration: const InputDecoration(hintText: '••••••'),
      ),
      actions: [
        TextButton(onPressed: () => Navigator.pop(context), child: const Text('Huỷ')),
        FilledButton(onPressed: () => Navigator.pop(context, controller.text), child: const Text('Xác nhận')),
      ],
    ),
  );
}
