import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:go_router/go_router.dart';

import '../../../core/widgets/status_view.dart';
import '../bloc/shopping_cubit.dart';

class WishlistPage extends StatelessWidget {
  const WishlistPage({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Yêu thích')),
      body: BlocBuilder<WishlistCubit, WishlistState>(
        builder: (context, state) {
          final cubit = context.read<WishlistCubit>();
          final items = state.wishlist?.items ?? const [];
          return StatusView(
            status: state.status,
            error: state.error,
            onRetry: cubit.load,
            isEmpty: items.isEmpty,
            emptyMessage: 'Chưa có sản phẩm yêu thích',
            builder: (context) => ListView(
              children: [
                for (final item in items)
                  Dismissible(
                    key: ValueKey(item.lineId),
                    direction: DismissDirection.endToStart,
                    background: Container(
                      color: Theme.of(context).colorScheme.errorContainer,
                      alignment: Alignment.centerRight,
                      padding: const EdgeInsets.only(right: 24),
                      child: const Icon(Icons.delete_outline),
                    ),
                    onDismissed: (_) => cubit.remove(item),
                    child: ListTile(
                      leading: const Icon(Icons.favorite, color: Colors.red),
                      title: Text(item.title),
                      subtitle: Text('Số lượng: ${item.quantity}'),
                      trailing: const Icon(Icons.chevron_right),
                      onTap: () => context.push('/shop/products/${item.productId}'),
                    ),
                  ),
              ],
            ),
          );
        },
      ),
    );
  }
}
