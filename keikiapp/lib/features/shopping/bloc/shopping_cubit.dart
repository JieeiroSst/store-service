import 'package:equatable/equatable.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

import '../../../core/utils/load_status.dart';
import '../data/shopping_repository.dart';

class CatalogState extends Equatable {
  const CatalogState({
    this.status = LoadStatus.initial,
    this.categories = const [],
    this.products = const [],
    this.categoryId,
    this.query = '',
    this.hasMore = true,
    this.loadingMore = false,
    this.error,
  });

  final LoadStatus status;
  final List<Category> categories;
  final List<Product> products;
  final int? categoryId;
  final String query;
  final bool hasMore;
  final bool loadingMore;
  final String? error;

  CatalogState copyWith({
    LoadStatus? status,
    List<Category>? categories,
    List<Product>? products,
    int? Function()? categoryId,
    String? query,
    bool? hasMore,
    bool? loadingMore,
    String? error,
  }) => CatalogState(
    status: status ?? this.status,
    categories: categories ?? this.categories,
    products: products ?? this.products,
    categoryId: categoryId != null ? categoryId() : this.categoryId,
    query: query ?? this.query,
    hasMore: hasMore ?? this.hasMore,
    loadingMore: loadingMore ?? this.loadingMore,
    error: error,
  );

  @override
  List<Object?> get props => [status, categories, products, categoryId, query, hasMore, loadingMore, error];
}

class CatalogCubit extends Cubit<CatalogState> {
  CatalogCubit(this._repository) : super(const CatalogState());

  final ShoppingRepository _repository;
  static const _pageSize = 20;

  Future<void> load() async {
    emit(state.copyWith(status: LoadStatus.loading));
    try {
      final results = await Future.wait([
        state.categories.isEmpty ? _repository.categories() : Future.value(state.categories),
        _repository.products(query: _q, categoryId: state.categoryId, limit: _pageSize),
      ]);
      final products = results[1] as List<Product>;
      emit(
        state.copyWith(
          status: LoadStatus.success,
          categories: results[0] as List<Category>,
          products: products,
          hasMore: products.length == _pageSize,
        ),
      );
    } catch (e) {
      emit(state.copyWith(status: LoadStatus.failure, error: e.toString()));
    }
  }

  String? get _q => state.query.trim().isEmpty ? null : state.query.trim();

  Future<void> loadMore() async {
    if (!state.hasMore || state.loadingMore || !state.status.isSuccess) return;
    emit(state.copyWith(loadingMore: true));
    try {
      final next = await _repository.products(query: _q, categoryId: state.categoryId, limit: _pageSize, offset: state.products.length);
      emit(state.copyWith(products: [...state.products, ...next], hasMore: next.length == _pageSize, loadingMore: false));
    } catch (e) {
      emit(state.copyWith(loadingMore: false, error: e.toString()));
    }
  }

  Future<void> search(String query) {
    emit(state.copyWith(query: query));
    return load();
  }

  Future<void> selectCategory(int? id) {
    emit(state.copyWith(categoryId: () => id));
    return load();
  }
}

class WishlistState extends Equatable {
  const WishlistState({this.status = LoadStatus.initial, this.wishlist, this.error});

  final LoadStatus status;
  final MyWishlist? wishlist;
  final String? error;

  bool contains(int productId) => wishlist?.contains(productId) ?? false;

  @override
  List<Object?> get props => [status, wishlist, error];
}

class WishlistCubit extends Cubit<WishlistState> {
  WishlistCubit(this._repository) : super(const WishlistState());

  final ShoppingRepository _repository;

  Future<void> load() async {
    emit(WishlistState(status: LoadStatus.loading, wishlist: state.wishlist));
    try {
      emit(WishlistState(status: LoadStatus.success, wishlist: await _repository.wishlist()));
    } catch (e) {
      emit(WishlistState(status: LoadStatus.failure, wishlist: state.wishlist, error: e.toString()));
    }
  }

  Future<void> toggle(Product product) async {
    final wishlist = state.wishlist;
    final line = wishlist?.items.where((i) => i.productId == product.id).firstOrNull;
    if (line != null) {
      await _repository.removeFromWishlist(wishlist!.id, line.lineId);
    } else {
      await _repository.addToWishlist(product);
    }
    await load();
  }

  Future<void> remove(WishlistItem item) async {
    await _repository.removeFromWishlist(state.wishlist!.id, item.lineId);
    await load();
  }
}
