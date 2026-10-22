import 'package:equatable/equatable.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

import '../../../core/utils/load_status.dart';
import '../data/food_repository.dart';

class RestaurantsState extends Equatable {
  const RestaurantsState({this.status = LoadStatus.initial, this.restaurants = const [], this.error});

  final LoadStatus status;
  final List<Restaurant> restaurants;
  final String? error;

  @override
  List<Object?> get props => [status, restaurants, error];
}

class RestaurantsCubit extends Cubit<RestaurantsState> {
  RestaurantsCubit(this._repository) : super(const RestaurantsState());

  final FoodRepository _repository;

  Future<void> load({String? query}) async {
    emit(RestaurantsState(status: LoadStatus.loading, restaurants: state.restaurants));
    try {
      final q = query?.trim();
      emit(RestaurantsState(status: LoadStatus.success, restaurants: await _repository.restaurants(query: q == null || q.isEmpty ? null : q)));
    } catch (e) {
      emit(RestaurantsState(status: LoadStatus.failure, restaurants: state.restaurants, error: e.toString()));
    }
  }
}

class FoodMenuState extends Equatable {
  const FoodMenuState({
    this.status = LoadStatus.initial,
    this.restaurant,
    this.foods = const [],
    this.query = '',
    this.category,
    this.cart = const {},
    this.error,
  });

  final LoadStatus status;
  final Restaurant? restaurant;
  final List<Food> foods;
  final String query;
  final String? category;
  final Map<Food, int> cart;
  final String? error;

  List<String> get categories => {for (final f in foods) if (f.category.isNotEmpty) f.category}.toList()..sort();

  List<Food> get visibleFoods {
    final q = query.trim().toLowerCase();
    return foods
        .where((f) => category == null || f.category == category)
        .where((f) => q.isEmpty || f.name.toLowerCase().contains(q) || f.description.toLowerCase().contains(q))
        .toList();
  }

  int get cartCount => cart.values.fold(0, (a, b) => a + b);

  int get cartTotal => cart.entries.fold(0, (sum, e) => sum + e.key.price * e.value);

  FoodMenuState copyWith({
    LoadStatus? status,
    Restaurant? Function()? restaurant,
    List<Food>? foods,
    String? query,
    String? Function()? category,
    Map<Food, int>? cart,
    String? error,
  }) =>
      FoodMenuState(
        status: status ?? this.status,
        restaurant: restaurant != null ? restaurant() : this.restaurant,
        foods: foods ?? this.foods,
        query: query ?? this.query,
        category: category != null ? category() : this.category,
        cart: cart ?? this.cart,
        error: error,
      );

  @override
  List<Object?> get props => [status, restaurant, foods, query, category, cart, error];
}

class FoodMenuCubit extends Cubit<FoodMenuState> {
  FoodMenuCubit(this._repository) : super(const FoodMenuState());

  final FoodRepository _repository;

  bool wouldDropCart(Restaurant r) => state.cart.isNotEmpty && state.restaurant?.id != r.id;

  Future<void> openRestaurant(Restaurant r) async {
    final same = state.restaurant?.id == r.id;
    emit(state.copyWith(
      restaurant: () => r,
      status: LoadStatus.loading,
      cart: same ? state.cart : const {},
      category: same ? null : () => null,
      query: same ? null : '',
    ));
    await load();
  }

  Future<void> load() async {
    final r = state.restaurant;
    if (r == null) return;
    emit(state.copyWith(status: LoadStatus.loading));
    try {
      emit(state.copyWith(status: LoadStatus.success, foods: await _repository.menu(r.id)));
    } catch (e) {
      emit(state.copyWith(status: LoadStatus.failure, error: e.toString()));
    }
  }

  void search(String query) => emit(state.copyWith(query: query));

  void selectCategory(String? category) => emit(state.copyWith(category: () => category));

  void add(Food food) => emit(state.copyWith(cart: {...state.cart, food: (state.cart[food] ?? 0) + 1}));

  void remove(Food food) {
    final next = {...state.cart};
    final qty = (next[food] ?? 0) - 1;
    if (qty <= 0) {
      next.remove(food);
    } else {
      next[food] = qty;
    }
    emit(state.copyWith(cart: next));
  }

  void clearCart() => emit(state.copyWith(cart: const {}));
}

class FoodOrdersState extends Equatable {
  const FoodOrdersState({this.status = LoadStatus.initial, this.orders = const [], this.error});

  final LoadStatus status;
  final List<FoodOrder> orders;
  final String? error;

  @override
  List<Object?> get props => [status, orders, error];
}

class FoodOrdersCubit extends Cubit<FoodOrdersState> {
  FoodOrdersCubit(this._repository) : super(const FoodOrdersState());

  final FoodRepository _repository;

  Future<void> load() async {
    emit(FoodOrdersState(status: LoadStatus.loading, orders: state.orders));
    try {
      emit(FoodOrdersState(status: LoadStatus.success, orders: await _repository.myOrders()));
    } catch (e) {
      emit(FoodOrdersState(status: LoadStatus.failure, orders: state.orders, error: e.toString()));
    }
  }
}
