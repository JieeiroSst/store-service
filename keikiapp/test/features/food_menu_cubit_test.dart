import 'package:flutter_test/flutter_test.dart';
import 'package:keikiapp/core/utils/load_status.dart';
import 'package:keikiapp/features/food/bloc/food_cubit.dart';
import 'package:keikiapp/features/food/data/food_repository.dart';
import 'package:mocktail/mocktail.dart';

class _MockFoodRepository extends Mock implements FoodRepository {}

const _thin = Restaurant(id: 1, name: 'Phở Thìn');
const _other = Restaurant(id: 2, name: 'Bún chả');
const _pho = Food(id: 1, name: 'Phở bò', price: 50000, category: 'Món nước', restaurantId: 1);
const _tra = Food(id: 2, name: 'Trà đá', price: 5000, category: 'Đồ uống', restaurantId: 1);

void main() {
  late _MockFoodRepository repo;

  setUp(() {
    repo = _MockFoodRepository();
    when(() => repo.menu(1)).thenAnswer((_) async => [_pho, _tra]);
    when(() => repo.menu(2)).thenAnswer((_) async => []);
    when(() => repo.restaurants(query: any(named: 'query'))).thenAnswer((_) async => [_thin, _other]);
  });

  test('cart totals follow add and remove', () {
    final cubit = FoodMenuCubit(repo)
      ..add(_pho)
      ..add(_pho)
      ..add(_tra)
      ..remove(_pho);

    expect(cubit.state.cart, {_pho: 1, _tra: 1});
    expect(cubit.state.cartCount, 2);
    expect(cubit.state.cartTotal, 55000);

    cubit
      ..remove(_tra)
      ..remove(_tra);
    expect(cubit.state.cart, {_pho: 1});
  });

  test('opening a restaurant loads its menu, filters by category and search', () async {
    final cubit = FoodMenuCubit(repo);
    await cubit.openRestaurant(_thin);

    expect(cubit.state.status, LoadStatus.success);
    expect(cubit.state.categories, ['Món nước', 'Đồ uống']);
    cubit.selectCategory('Đồ uống');
    expect(cubit.state.visibleFoods, [_tra]);
    cubit
      ..selectCategory(null)
      ..search('phở');
    expect(cubit.state.visibleFoods, [_pho]);
  });

  test('switching restaurant drops the cart, reopening keeps it', () async {
    final cubit = FoodMenuCubit(repo);
    await cubit.openRestaurant(_thin);
    cubit.add(_pho);

    expect(cubit.wouldDropCart(_thin), isFalse);
    await cubit.openRestaurant(_thin);
    expect(cubit.state.cart, {_pho: 1});

    expect(cubit.wouldDropCart(_other), isTrue);
    await cubit.openRestaurant(_other);
    expect(cubit.state.cart, isEmpty);
    expect(cubit.state.restaurant, _other);
  });

  test('restaurant search passes the query through', () async {
    final cubit = RestaurantsCubit(repo);
    await cubit.load(query: '  phở ');
    verify(() => repo.restaurants(query: 'phở')).called(1);
    await cubit.load(query: '');
    verify(() => repo.restaurants(query: null)).called(1);
    expect(cubit.state.restaurants, [_thin, _other]);
  });
}
