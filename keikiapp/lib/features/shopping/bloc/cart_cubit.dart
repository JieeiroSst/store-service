import 'package:equatable/equatable.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

import '../../../core/utils/load_status.dart';
import '../../../gen/models.g.dart';
import '../data/shopping_repository.dart';

class CartState extends Equatable {
  const CartState({this.status = LoadStatus.initial, this.basket, this.lines = const [], this.busy = false, this.error});

  final LoadStatus status;
  final BasketBasket? basket;
  final List<CartLine> lines;
  final bool busy;
  final String? error;

  int get count => lines.fold(0, (n, l) => n + l.quantity);

  double get total => lines.fold(0.0, (n, l) => n + l.total);

  String get currency => lines.isEmpty ? 'VND' : (lines.first.line.priceCurrency ?? 'VND');

  CartState copyWith({LoadStatus? status, BasketBasket? basket, List<CartLine>? lines, bool? busy, String? error}) => CartState(
    status: status ?? this.status,
    basket: basket ?? this.basket,
    lines: lines ?? this.lines,
    busy: busy ?? this.busy,
    error: error,
  );

  @override
  List<Object?> get props => [status, basket?.id, lines, busy, error];
}

class CartCubit extends Cubit<CartState> {
  CartCubit(this._repository) : super(const CartState());

  final ShoppingRepository _repository;

  Future<void> load() async {
    emit(state.copyWith(status: state.basket == null ? LoadStatus.loading : state.status));
    try {
      final basket = await _repository.openBasket();
      final lines = await _repository.cartLines(basket.id ?? 0);
      emit(CartState(status: LoadStatus.success, basket: basket, lines: lines));
    } catch (e) {
      emit(state.copyWith(status: LoadStatus.failure, error: e.toString()));
    }
  }

  Future<void> add(Offer offer, int quantity) => _mutate(() async {
    final basket = state.basket ?? await _repository.openBasket();
    await _repository.addToCart(basket.id ?? 0, offer, quantity);
  });

  Future<void> setQuantity(CartLine line, int quantity) => _mutate(() => _repository.setQuantity(line.id, quantity));

  Future<BasketShippingQuote?> shippingQuote(int addressId) async {
    final basket = state.basket;
    if (basket?.id == null) return null;
    return _repository.shippingQuote(basketId: basket!.id!, addressId: addressId);
  }

  Future<BasketBasket?> checkout({required String walletId, required String pin, int? addressId, double? shippingFee}) async {
    final basket = state.basket;
    if (basket == null) return null;
    emit(state.copyWith(busy: true));
    try {
      final done = await _repository.checkout(basketId: basket.id ?? 0, walletId: walletId, pin: pin, addressId: addressId, shippingFee: shippingFee);
      emit(const CartState(status: LoadStatus.loading));
      await load();
      return done;
    } catch (e) {
      emit(state.copyWith(busy: false, error: e.toString()));
      rethrow;
    }
  }

  Future<void> _mutate(Future<void> Function() action) async {
    emit(state.copyWith(busy: true));
    try {
      await action();
      await load();
    } catch (e) {
      emit(state.copyWith(busy: false, error: e.toString()));
      rethrow;
    }
  }
}
