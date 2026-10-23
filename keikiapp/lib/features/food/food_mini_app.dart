import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:go_router/go_router.dart';

import '../../app/deep_links.dart';

import '../../app/mini_app.dart';
import 'bloc/food_cubit.dart';
import 'data/food_repository.dart';
import 'view/food_checkout_page.dart';
import 'view/food_home_page.dart';
import 'view/food_order_page.dart';

final foodMiniApp = MiniApp(
  id: 'food',
  title: 'Food',
  icon: Icons.restaurant,
  color: Colors.deepOrange,
  path: '/food',
  routes: [
    ShellRoute(
      builder: (context, state, child) => MultiBlocProvider(
        providers: [
          BlocProvider(create: (context) => RestaurantsCubit(context.read<FoodRepository>())..load()),
          BlocProvider(create: (context) => FoodMenuCubit(context.read<FoodRepository>())),
          BlocProvider(create: (context) => FoodOrdersCubit(context.read<FoodRepository>())..load()),
        ],
        child: child,
      ),
      routes: [
        GoRoute(
          path: '/food',
          builder: (_, _) => const FoodHomePage(),
          routes: [
            GoRoute(path: 'checkout', builder: (_, _) => const FoodCheckoutPage()),
            GoRoute(path: 'restaurants/:id', builder: (_, state) => RestaurantPage(key: ValueKey(state.pathParameters['id']), id: int.tryParse(state.pathParameters['id'] ?? ''))),
            GoRoute(
              path: 'orders/:id',
              builder: (_, state) => intParam(state, 'id') == null ? LinkNotFoundPage(location: state.uri.toString()) : FoodOrderPage(orderId: intParam(state, 'id')!),
            ),
          ],
        ),
      ],
    ),
  ],
);
