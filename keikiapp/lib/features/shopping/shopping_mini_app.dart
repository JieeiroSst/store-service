import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:go_router/go_router.dart';

import '../../app/mini_app.dart';
import 'bloc/cart_cubit.dart';
import 'bloc/shopping_cubit.dart';
import 'data/shopping_repository.dart';
import 'view/cart_page.dart';
import 'view/coupons_page.dart';
import 'view/product_page.dart';
import 'view/purchases_page.dart';
import 'view/shop_home_page.dart';
import 'view/wishlist_page.dart';

final shoppingMiniApp = MiniApp(
  id: 'shop',
  title: 'Mua sắm',
  icon: Icons.shopping_bag,
  color: Colors.pink,
  path: '/shop',
  routes: [
    ShellRoute(
      builder: (context, state, child) => MultiBlocProvider(
        providers: [
          BlocProvider(create: (context) => CatalogCubit(context.read<ShoppingRepository>())..load()),
          BlocProvider(create: (context) => WishlistCubit(context.read<ShoppingRepository>())..load()),
          BlocProvider(create: (context) => CartCubit(context.read<ShoppingRepository>())..load()),
        ],
        child: child,
      ),
      routes: [
        GoRoute(
          path: '/shop',
          builder: (_, _) => const ShopHomePage(),
          routes: [
            GoRoute(
              path: 'products/:id',
              builder: (_, state) => ProductPage(productId: int.parse(state.pathParameters['id']!)),
            ),
            GoRoute(path: 'wishlist', builder: (_, _) => const WishlistPage()),
            GoRoute(path: 'coupons', builder: (_, _) => const CouponsPage()),
            GoRoute(path: 'cart', builder: (_, _) => const CartPage()),
            GoRoute(path: 'purchases', builder: (_, _) => const PurchasesPage()),
          ],
        ),
      ],
    ),
  ],
);
