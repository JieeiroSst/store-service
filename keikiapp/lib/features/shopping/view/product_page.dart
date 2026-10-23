import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

import 'package:go_router/go_router.dart';

import '../../../core/links/share_link.dart';
import '../../../core/utils/formatters.dart';
import '../../../core/widgets/status_view.dart';
import '../../../core/widgets/ui.dart';
import '../bloc/cart_cubit.dart';
import '../bloc/shopping_cubit.dart';
import '../data/shopping_repository.dart';
import 'shop_home_page.dart';

class ProductPage extends StatefulWidget {
  const ProductPage({super.key, required this.productId});

  final int productId;

  @override
  State<ProductPage> createState() => _ProductPageState();
}

class _ProductPageState extends State<ProductPage> {
  late Future<(Product?, List<Product>)> _future = _load();

  Future<(Product?, List<Product>)> _load() async {
    final repo = context.read<ShoppingRepository>();
    final product = await repo.product(widget.productId);
    final related = product == null ? <Product>[] : await repo.productsByIds(product.recommendationIds.take(8).toList());
    return (product, related);
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: FutureBuilder<(Product?, List<Product>)>(
        future: _future,
        builder: (context, snap) {
          if (snap.hasError) {
            return Scaffold(
              appBar: AppBar(),
              body: ErrorView(message: snap.error.toString(), onRetry: () => setState(() => _future = _load())),
            );
          }
          if (!snap.hasData) return const Center(child: CircularProgressIndicator());
          final (product, related) = snap.data!;
          if (product == null) {
            return Scaffold(
              appBar: AppBar(),
              body: const EmptyView(message: 'Sản phẩm không tồn tại'),
            );
          }
          return Scaffold(
            body: _ProductBody(product: product, related: related),
            bottomNavigationBar: _AddToCartBar(product: product),
          );
        },
      ),
    );
  }
}

class _ProductBody extends StatefulWidget {
  const _ProductBody({required this.product, required this.related});

  final Product product;
  final List<Product> related;

  @override
  State<_ProductBody> createState() => _ProductBodyState();
}

class _ProductBodyState extends State<_ProductBody> {
  int _image = 0;

  @override
  Widget build(BuildContext context) {
    final product = widget.product;
    final theme = Theme.of(context);
    final liked = context.select<WishlistCubit, bool>((c) => c.state.contains(product.id));
    return CustomScrollView(
      slivers: [
        SliverAppBar(
          expandedHeight: 320,
          pinned: true,
          flexibleSpace: FlexibleSpaceBar(
            background: product.images.isEmpty
                ? const ProductImage(url: null)
                : PageView(
                    onPageChanged: (i) => setState(() => _image = i),
                    children: [for (final url in product.images) ProductImage(url: url)],
                  ),
          ),
          actions: [
            ShareLinkButton(path: '/shop/products/${product.id}', title: product.title, imageUrl: product.thumbnail),
            IconButton(
              icon: Icon(liked ? Icons.favorite : Icons.favorite_border, color: liked ? Colors.red : null),
              onPressed: () => context.read<WishlistCubit>().toggle(product).catchError((Object e) {
                if (context.mounted) showSnack(context, e.toString());
              }),
            ),
          ],
        ),
        SliverPadding(
          padding: const EdgeInsets.all(16),
          sliver: SliverList.list(
            children: [
              if (product.images.length > 1)
                Center(child: Text('${_image + 1}/${product.images.length}', style: theme.textTheme.labelSmall)),
              Text(product.title, style: theme.textTheme.headlineSmall),
              if (product.rating != null)
                Padding(
                  padding: const EdgeInsets.only(top: 4),
                  child: Row(
                    children: [
                      const Icon(Icons.star, color: Colors.amber, size: 18),
                      Text(' ${product.rating!.toStringAsFixed(1)}'),
                    ],
                  ),
                ),
              const SizedBox(height: 16),
              Text(product.description.isEmpty ? 'Chưa có mô tả' : product.description),
              if (widget.related.isNotEmpty) ...[
                const SizedBox(height: 24),
                Text('Có thể bạn cũng thích', style: theme.textTheme.titleMedium),
                const SizedBox(height: 8),
                SizedBox(
                  height: 240,
                  child: ListView.separated(
                    scrollDirection: Axis.horizontal,
                    itemCount: widget.related.length,
                    separatorBuilder: (_, _) => const SizedBox(width: 8),
                    itemBuilder: (context, i) => SizedBox(width: 160, child: ProductCard(product: widget.related[i])),
                  ),
                ),
              ],
            ],
          ),
        ),
      ],
    );
  }
}

class _AddToCartBar extends StatefulWidget {
  const _AddToCartBar({required this.product});

  final Product product;

  @override
  State<_AddToCartBar> createState() => _AddToCartBarState();
}

class _AddToCartBarState extends State<_AddToCartBar> {
  int _qty = 1;

  @override
  Widget build(BuildContext context) {
    final offer = widget.product.bestOffer;
    final theme = Theme.of(context);
    final busy = context.select<CartCubit, bool>((c) => c.state.busy);
    final max = offer?.available;
    return BottomActionBar(
      child: offer == null
          ? FilledButton(onPressed: null, child: const Text('Tạm hết hàng'))
          : Row(
              children: [
                Column(
                  mainAxisSize: MainAxisSize.min,
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      formatMoney(offer.price, offer.currency),
                      style: theme.textTheme.titleLarge?.copyWith(color: theme.colorScheme.primary),
                    ),
                    if (max != null) Text('Còn $max sản phẩm', style: theme.textTheme.bodySmall),
                  ],
                ),
                const Spacer(),
                IconButton(icon: const Icon(Icons.remove), onPressed: _qty > 1 ? () => setState(() => _qty--) : null),
                Text('$_qty', style: theme.textTheme.titleMedium),
                IconButton(icon: const Icon(Icons.add), onPressed: max == null || _qty < max ? () => setState(() => _qty++) : null),
                const SizedBox(width: 8),
                SizedBox(
                  width: 150,
                  child: LoadingButton(
                    label: 'Thêm vào giỏ',
                    loading: busy,
                    onPressed: () async {
                      try {
                        await context.read<CartCubit>().add(offer, _qty);
                        if (!context.mounted) return;
                        ScaffoldMessenger.of(context)
                          ..hideCurrentSnackBar()
                          ..showSnackBar(
                            SnackBar(
                              content: Text('Đã thêm $_qty sản phẩm vào giỏ'),
                              action: SnackBarAction(label: 'Xem giỏ', onPressed: () => context.push('/shop/cart')),
                            ),
                          );
                      } catch (e) {
                        if (context.mounted) showSnack(context, e.toString());
                      }
                    },
                  ),
                ),
              ],
            ),
    );
  }
}
