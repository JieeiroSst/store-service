import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

import '../../../core/utils/formatters.dart';
import '../../../core/widgets/status_view.dart';
import '../../../core/widgets/ui.dart';
import '../../../gen/models.g.dart';
import '../data/shopping_repository.dart';

class PurchasesPage extends StatefulWidget {
  const PurchasesPage({super.key});

  @override
  State<PurchasesPage> createState() => _PurchasesPageState();
}

class _PurchasesPageState extends State<PurchasesPage> {
  late Future<List<BasketBasket>> _future = context.read<ShoppingRepository>().purchases();

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Scaffold(
      appBar: AppBar(title: const Text('Đơn mua')),
      body: FutureBuilder<List<BasketBasket>>(
        future: _future,
        builder: (context, snap) {
          if (snap.hasError) {
            return ErrorView(
              message: snap.error.toString(),
              onRetry: () => setState(() => _future = context.read<ShoppingRepository>().purchases()),
            );
          }
          if (!snap.hasData) return const SkeletonList();
          if (snap.data!.isEmpty) return const EmptyView(message: 'Bạn chưa có đơn mua nào', icon: Icons.receipt_long_outlined);
          return ListView.separated(
            padding: const EdgeInsets.all(16),
            itemCount: snap.data!.length,
            separatorBuilder: (_, _) => const SizedBox(height: 10),
            itemBuilder: (context, i) {
              final b = snap.data![i];
              return Card(
                child: ListTile(
                  leading: const IconBadge(icon: Icons.shopping_bag_outlined, color: Colors.pink, size: 44),
                  title: Text('Đơn KS${(b.id ?? 0).toString().padLeft(8, '0')}'),
                  subtitle: Text(formatDate(b.dateSubmitted)),
                  trailing: Column(
                    mainAxisAlignment: MainAxisAlignment.center,
                    crossAxisAlignment: CrossAxisAlignment.end,
                    children: [
                      Text(
                        formatMoney(b.totalInclTax, b.currency),
                        style: theme.textTheme.titleSmall?.copyWith(color: theme.colorScheme.primary),
                      ),
                      (b.fulfillmentId ?? '').isEmpty
                          ? const Pill('Đã thanh toán', color: Colors.green)
                          : const Pill('Đang xử lý', color: Colors.indigo),
                    ],
                  ),
                ),
              );
            },
          );
        },
      ),
    );
  }
}
