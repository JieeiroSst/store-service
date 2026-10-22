import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:go_router/go_router.dart';

import '../../../core/widgets/status_view.dart';
import '../../../app/theme.dart';
import '../../../core/utils/formatters.dart';
import '../bloc/cart_cubit.dart';
import '../bloc/shopping_cubit.dart';
import '../data/shopping_repository.dart';
import '../../apps/commerce/commerce_more.dart';

class ShopHomePage extends StatelessWidget {
  const ShopHomePage({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Keiki Shop'),
        actions: [
          IconButton(
            tooltip: 'Mua sắm nâng cao',
            icon: const Icon(Icons.dashboard_customize_outlined),
            onPressed: () => Navigator.of(context).push(MaterialPageRoute<void>(builder: (_) => const ShopMorePage())),
          ),
          IconButton(tooltip: 'Coupon', icon: const Icon(Icons.local_offer_outlined), onPressed: () => context.push('/shop/coupons')),
          IconButton(tooltip: 'Đơn mua', icon: const Icon(Icons.receipt_long_outlined), onPressed: () => context.push('/shop/purchases')),
          BlocBuilder<CartCubit, CartState>(
            builder: (context, state) => IconButton(
              tooltip: 'Giỏ hàng',
              icon: Badge(isLabelVisible: state.count > 0, label: Text('${state.count}'), child: const Icon(Icons.shopping_cart_outlined)),
              onPressed: () => context.push('/shop/cart'),
            ),
          ),
          BlocBuilder<WishlistCubit, WishlistState>(
            builder: (context, state) => IconButton(
              tooltip: 'Yêu thích',
              icon: Badge(
                isLabelVisible: (state.wishlist?.items.length ?? 0) > 0,
                label: Text('${state.wishlist?.items.length ?? 0}'),
                child: const Icon(Icons.favorite_border),
              ),
              onPressed: () => context.push('/shop/wishlist'),
            ),
          ),
        ],
      ),
      body: BlocBuilder<CatalogCubit, CatalogState>(
        builder: (context, state) {
          final cubit = context.read<CatalogCubit>();
          return Column(
            children: [
              Padding(
                padding: const EdgeInsets.fromLTRB(16, 8, 16, 4),
                child: SearchBar(hintText: 'Tìm sản phẩm', leading: const Icon(Icons.search), onSubmitted: cubit.search),
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
                        child: ChoiceChip(
                          label: const Text('Tất cả'),
                          selected: state.categoryId == null,
                          onSelected: (_) => cubit.selectCategory(null),
                        ),
                      ),
                      for (final c in state.categories.where((c) => c.parentId == null))
                        Padding(
                          padding: const EdgeInsets.all(4),
                          child: ChoiceChip(
                            label: Text(c.name),
                            selected: state.categoryId == c.id,
                            onSelected: (_) => cubit.selectCategory(c.id),
                          ),
                        ),
                    ],
                  ),
                ),
              Expanded(
                child: StatusView(
                  status: state.status,
                  error: state.error,
                  onRetry: cubit.load,
                  isEmpty: state.products.isEmpty,
                  emptyMessage: 'Không có sản phẩm',
                  builder: (context) => NotificationListener<ScrollNotification>(
                    onNotification: (n) {
                      if (n.metrics.extentAfter < 300) cubit.loadMore();
                      return false;
                    },
                    child: RefreshIndicator(
                      onRefresh: cubit.load,
                      child: GridView.builder(
                        padding: const EdgeInsets.all(12),
                        gridDelegate: const SliverGridDelegateWithMaxCrossAxisExtent(
                          maxCrossAxisExtent: 220,
                          mainAxisSpacing: 12,
                          crossAxisSpacing: 12,
                          childAspectRatio: 0.72,
                        ),
                        itemCount: state.products.length,
                        itemBuilder: (context, i) => ProductCard(product: state.products[i]),
                      ),
                    ),
                  ),
                ),
              ),
            ],
          );
        },
      ),
    );
  }
}

class ProductCard extends StatelessWidget {
  const ProductCard({super.key, required this.product});

  final Product product;

  @override
  Widget build(BuildContext context) {
    final liked = context.select<WishlistCubit, bool>((c) => c.state.contains(product.id));
    return Card(
      clipBehavior: Clip.antiAlias,
      child: InkWell(
        onTap: () => context.push('/shop/products/${product.id}'),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Expanded(
              child: Stack(
                fit: StackFit.expand,
                children: [
                  ProductImage(url: product.thumbnail),
                  Positioned(
                    top: 4,
                    right: 4,
                    child: IconButton.filledTonal(
                      iconSize: 18,
                      icon: Icon(liked ? Icons.favorite : Icons.favorite_border, color: liked ? Colors.red : null),
                      onPressed: () => context.read<WishlistCubit>().toggle(product).catchError((Object e) {
                        if (context.mounted) showSnack(context, e.toString());
                      }),
                    ),
                  ),
                ],
              ),
            ),
            Padding(
              padding: const EdgeInsets.all(8),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(product.title, maxLines: 2, overflow: TextOverflow.ellipsis),
                  const SizedBox(height: 2),
                  product.price == null
                      ? Text(
                          'Hết hàng',
                          style: Theme.of(context).textTheme.labelMedium?.copyWith(color: Theme.of(context).colorScheme.error),
                        )
                      : Text(
                          formatMoney(product.price, product.currency),
                          style: Theme.of(context).textTheme.titleSmall?.copyWith(color: Theme.of(context).colorScheme.primary),
                        ),
                  if (product.rating != null)
                    Row(
                      children: [
                        const Icon(Icons.star, size: 14, color: Colors.amber),
                        Text(' ${product.rating!.toStringAsFixed(1)}'),
                      ],
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

class ProductImage extends StatelessWidget {
  const ProductImage({super.key, required this.url});

  final String? url;

  @override
  Widget build(BuildContext context) {
    final placeholder = DecoratedBox(
      decoration: const BoxDecoration(gradient: KeikiColors.softGradient),
      child: Center(child: Icon(Icons.shopping_bag_outlined, size: 40, color: KeikiColors.violet.withValues(alpha: 0.45))),
    );
    if (url == null || !url!.startsWith('http')) return placeholder;
    return Image.network(url!, fit: BoxFit.cover, errorBuilder: (_, _, _) => placeholder);
  }
}
