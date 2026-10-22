import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:go_router/go_router.dart';

import '../../../core/utils/formatters.dart';
import '../../../core/widgets/status_view.dart';
import '../bloc/food_cubit.dart';
import '../data/food_repository.dart';
import '../../apps/commerce/commerce_more.dart';

class FoodHomePage extends StatelessWidget {
  const FoodHomePage({super.key});

  @override
  Widget build(BuildContext context) {
    return DefaultTabController(
      length: 2,
      child: Scaffold(
        appBar: AppBar(
          title: const Text('Keiki Food'),
          actions: [
            IconButton(
              tooltip: 'Tìm món',
              icon: const Icon(Icons.search),
              onPressed: () => Navigator.of(context).push(MaterialPageRoute<void>(builder: (_) => const FoodSearchPage())),
            ),
          ],
          bottom: const TabBar(tabs: [Tab(text: 'Nhà hàng'), Tab(text: 'Đơn của tôi')]),
        ),
        body: const TabBarView(children: [_RestaurantsTab(), _OrdersTab()]),
        floatingActionButton: const CartButton(),
      ),
    );
  }
}

class CartButton extends StatelessWidget {
  const CartButton({super.key});

  @override
  Widget build(BuildContext context) {
    return BlocBuilder<FoodMenuCubit, FoodMenuState>(
      buildWhen: (p, c) => p.cart != c.cart,
      builder: (context, state) => state.cartCount == 0
          ? const SizedBox.shrink()
          : FloatingActionButton.extended(
              onPressed: () async {
                final orderId = await context.push<int>('/food/checkout');
                if (orderId != null && context.mounted) {
                  context.read<FoodOrdersCubit>().load();
                  context.push('/food/orders/$orderId');
                }
              },
              icon: Badge(label: Text('${state.cartCount}'), child: const Icon(Icons.shopping_bag_outlined)),
              label: Text(formatMoney(state.cartTotal)),
            ),
    );
  }
}

class _RestaurantsTab extends StatelessWidget {
  const _RestaurantsTab();

  Future<void> _open(BuildContext context, Restaurant r) async {
    final menu = context.read<FoodMenuCubit>();
    if (menu.wouldDropCart(r)) {
      final ok = await showDialog<bool>(
        context: context,
        builder: (context) => AlertDialog(
          title: const Text('Đổi nhà hàng?'),
          content: Text('Giỏ hàng từ ${menu.state.restaurant?.name ?? 'nhà hàng khác'} sẽ bị xoá.'),
          actions: [
            TextButton(onPressed: () => Navigator.pop(context, false), child: const Text('Giữ giỏ')),
            FilledButton(onPressed: () => Navigator.pop(context, true), child: const Text('Đổi')),
          ],
        ),
      );
      if (ok != true || !context.mounted) return;
    }
    menu.openRestaurant(r);
    context.push('/food/restaurants/${r.id}');
  }

  @override
  Widget build(BuildContext context) {
    return BlocBuilder<RestaurantsCubit, RestaurantsState>(
      builder: (context, state) {
        final cubit = context.read<RestaurantsCubit>();
        return Column(
          children: [
            Padding(
              padding: const EdgeInsets.fromLTRB(16, 12, 16, 4),
              child: SearchBar(
                hintText: 'Tìm nhà hàng',
                leading: const Icon(Icons.search),
                elevation: const WidgetStatePropertyAll(0),
                onSubmitted: (q) => cubit.load(query: q),
              ),
            ),
            Expanded(
              child: StatusView(
                status: state.status,
                error: state.error,
                onRetry: cubit.load,
                isEmpty: state.restaurants.isEmpty,
                emptyMessage: 'Chưa có nhà hàng nào',
                builder: (context) => RefreshIndicator(
                  onRefresh: cubit.load,
                  child: ListView.separated(
                    padding: const EdgeInsets.fromLTRB(16, 8, 16, 96),
                    itemCount: state.restaurants.length,
                    separatorBuilder: (_, _) => const SizedBox(height: 12),
                    itemBuilder: (context, i) => _RestaurantCard(restaurant: state.restaurants[i], onTap: () => _open(context, state.restaurants[i])),
                  ),
                ),
              ),
            ),
          ],
        );
      },
    );
  }
}

class _RestaurantCard extends StatelessWidget {
  const _RestaurantCard({required this.restaurant, required this.onTap});

  final Restaurant restaurant;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final image = restaurant.imageUrl;
    return Card(
      clipBehavior: Clip.antiAlias,
      child: InkWell(
        onTap: onTap,
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            SizedBox(
              height: 120,
              child: image != null && image.startsWith('http')
                  ? Image.network(image, fit: BoxFit.cover, errorBuilder: (_, _, _) => const _RestaurantPlaceholder())
                  : const _RestaurantPlaceholder(),
            ),
            Padding(
              padding: const EdgeInsets.all(14),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(restaurant.name, style: theme.textTheme.titleMedium),
                  if (restaurant.address.isNotEmpty)
                    Row(children: [
                      Icon(Icons.location_on_outlined, size: 16, color: theme.colorScheme.onSurfaceVariant),
                      const SizedBox(width: 4),
                      Expanded(child: Text(restaurant.address, maxLines: 1, overflow: TextOverflow.ellipsis, style: theme.textTheme.bodySmall)),
                    ]),
                  if (restaurant.description.isNotEmpty)
                    Padding(
                      padding: const EdgeInsets.only(top: 4),
                      child: Text(restaurant.description, maxLines: 2, overflow: TextOverflow.ellipsis, style: theme.textTheme.bodySmall?.copyWith(color: theme.colorScheme.onSurfaceVariant)),
                    ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class _RestaurantPlaceholder extends StatelessWidget {
  const _RestaurantPlaceholder();

  @override
  Widget build(BuildContext context) {
    return const DecoratedBox(
      decoration: BoxDecoration(gradient: LinearGradient(colors: [Color(0xFFFF7043), Color(0xFFFFB74D)])),
      child: Center(child: Icon(Icons.storefront_rounded, color: Colors.white, size: 48)),
    );
  }
}

class RestaurantPage extends StatelessWidget {
  const RestaurantPage({super.key});

  @override
  Widget build(BuildContext context) {
    return BlocBuilder<FoodMenuCubit, FoodMenuState>(
      builder: (context, state) {
        final cubit = context.read<FoodMenuCubit>();
        return Scaffold(
          appBar: AppBar(title: Text(state.restaurant?.name ?? 'Thực đơn')),
          floatingActionButton: const CartButton(),
          body: Column(
            children: [
              Padding(
                padding: const EdgeInsets.fromLTRB(16, 4, 16, 4),
                child: SearchBar(
                  hintText: 'Tìm món',
                  leading: const Icon(Icons.search),
                  elevation: const WidgetStatePropertyAll(0),
                  onChanged: cubit.search,
                ),
              ),
              if (state.categories.isNotEmpty)
                SizedBox(
                  height: 48,
                  child: ListView(
                    scrollDirection: Axis.horizontal,
                    padding: const EdgeInsets.symmetric(horizontal: 12),
                    children: [
                      Padding(
                        padding: const EdgeInsets.all(4),
                        child: ChoiceChip(label: const Text('Tất cả'), selected: state.category == null, onSelected: (_) => cubit.selectCategory(null)),
                      ),
                      for (final c in state.categories)
                        Padding(
                          padding: const EdgeInsets.all(4),
                          child: ChoiceChip(label: Text(c), selected: state.category == c, onSelected: (_) => cubit.selectCategory(c)),
                        ),
                    ],
                  ),
                ),
              Expanded(
                child: StatusView(
                  status: state.status,
                  error: state.error,
                  onRetry: cubit.load,
                  isEmpty: state.visibleFoods.isEmpty,
                  emptyMessage: 'Không có món nào',
                  builder: (context) => RefreshIndicator(
                    onRefresh: cubit.load,
                    child: ListView.builder(
                      padding: const EdgeInsets.only(top: 4, bottom: 96),
                      itemCount: state.visibleFoods.length,
                      itemBuilder: (context, i) => FoodTile(food: state.visibleFoods[i], quantity: state.cart[state.visibleFoods[i]] ?? 0),
                    ),
                  ),
                ),
              ),
            ],
          ),
        );
      },
    );
  }
}

class FoodTile extends StatelessWidget {
  const FoodTile({super.key, required this.food, required this.quantity});

  final Food food;
  final int quantity;

  static const _palette = [Colors.deepOrange, Colors.teal, Colors.indigo, Colors.pink, Colors.green, Colors.brown, Colors.purple];

  @override
  Widget build(BuildContext context) {
    final cubit = context.read<FoodMenuCubit>();
    final theme = Theme.of(context);
    final color = _palette[food.category.hashCode.abs() % _palette.length];
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 6),
      child: Card(
        child: Padding(
          padding: const EdgeInsets.all(12),
          child: Row(
            children: [
              Container(
                width: 64,
                height: 64,
                decoration: BoxDecoration(color: color.withValues(alpha: 0.12), borderRadius: BorderRadius.circular(16)),
                alignment: Alignment.center,
                child: Text(food.name.isEmpty ? '?' : food.name.characters.first.toUpperCase(), style: theme.textTheme.headlineSmall?.copyWith(color: color)),
              ),
              const SizedBox(width: 14),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(food.name, style: theme.textTheme.titleSmall),
                    if (food.description.isNotEmpty)
                      Text(food.description, maxLines: 2, overflow: TextOverflow.ellipsis, style: theme.textTheme.bodySmall?.copyWith(color: theme.colorScheme.onSurfaceVariant)),
                    const SizedBox(height: 6),
                    Text(formatMoney(food.price), style: theme.textTheme.titleSmall?.copyWith(color: theme.colorScheme.primary)),
                  ],
                ),
              ),
              if (quantity == 0)
                IconButton.filled(icon: const Icon(Icons.add), onPressed: () => cubit.add(food))
              else
                Container(
                  decoration: BoxDecoration(color: theme.colorScheme.primaryContainer, borderRadius: BorderRadius.circular(999)),
                  child: Row(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      IconButton(visualDensity: VisualDensity.compact, icon: const Icon(Icons.remove), onPressed: () => cubit.remove(food)),
                      Text('$quantity', style: theme.textTheme.titleSmall),
                      IconButton(visualDensity: VisualDensity.compact, icon: const Icon(Icons.add), onPressed: () => cubit.add(food)),
                    ],
                  ),
                ),
            ],
          ),
        ),
      ),
    );
  }
}

class _OrdersTab extends StatelessWidget {
  const _OrdersTab();

  @override
  Widget build(BuildContext context) {
    return BlocBuilder<FoodOrdersCubit, FoodOrdersState>(
      builder: (context, state) {
        final cubit = context.read<FoodOrdersCubit>();
        return StatusView(
          status: state.status,
          error: state.error,
          onRetry: cubit.load,
          isEmpty: state.orders.isEmpty,
          emptyMessage: 'Bạn chưa có đơn nào',
          builder: (context) => RefreshIndicator(
            onRefresh: cubit.load,
            child: ListView(
              children: [
                for (final o in state.orders)
                  ListTile(
                    leading: Icon(o.isActive ? Icons.delivery_dining : Icons.receipt_long),
                    title: Text('Đơn #${o.id} • ${formatMoney(o.total)}'),
                    subtitle: Text('${o.items.length} món • ${formatDate(o.placedAt)}'),
                    trailing: OrderStatusChip(status: o.status),
                    onTap: () => context.push('/food/orders/${o.id}'),
                  ),
              ],
            ),
          ),
        );
      },
    );
  }
}

class OrderStatusChip extends StatelessWidget {
  const OrderStatusChip({super.key, required this.status});

  final String status;

  static const _labels = {
    'created': 'Đã đặt',
    'confirmed': 'Đã xác nhận',
    'preparing': 'Đang chuẩn bị',
    'ready_for_pickup': 'Chờ lấy hàng',
    'picked_up': 'Đang giao',
    'delivered': 'Đã giao',
    'cancelled': 'Đã huỷ',
  };

  @override
  Widget build(BuildContext context) {
    final scheme = Theme.of(context).colorScheme;
    final key = status.toLowerCase();
    final color = switch (key) {
      'delivered' => Colors.green,
      'cancelled' => scheme.error,
      _ => scheme.primary,
    };
    return Chip(
      label: Text(_labels[key] ?? status, style: TextStyle(color: color)),
      side: BorderSide(color: color),
      visualDensity: VisualDensity.compact,
    );
  }
}
