import 'package:equatable/equatable.dart';

import '../../../core/network/graphql_client.dart';
import '../../../gen/gateway_api.g.dart';

class Category extends Equatable {
  const Category({required this.id, required this.name, this.parentId, this.image = ''});

  final int id;
  final String name;
  final int? parentId;
  final String image;

  factory Category.fromJson(Map<String, dynamic> json) => Category(
    id: (json['id'] as num).toInt(),
    name: json['name'] as String? ?? '',
    parentId: (json['parent_id'] as num?)?.toInt(),
    image: json['image'] as String? ?? '',
  );

  @override
  List<Object?> get props => [id, name];
}

class Offer extends Equatable {
  const Offer({required this.id, required this.price, required this.currency, this.available});

  final int id;
  final double price;
  final String currency;
  final int? available;

  bool get inStock => available == null || available! > 0;

  @override
  List<Object?> get props => [id, price, currency, available];
}

class Product extends Equatable {
  const Product({
    required this.id,
    required this.title,
    this.description = '',
    this.rating,
    this.images = const [],
    this.categoryIds = const [],
    this.recommendationIds = const [],
    this.price,
    this.currency = 'VND',
    this.available,
    this.inStock = false,
    this.offers = const [],
  });

  final int id;
  final String title;
  final String description;
  final double? rating;
  final List<String> images;
  final List<int> categoryIds;
  final List<int> recommendationIds;
  final double? price;
  final String currency;
  final int? available;
  final bool inStock;
  final List<Offer> offers;

  Offer? get bestOffer {
    final open = offers.where((o) => o.inStock).toList()..sort((a, b) => a.price.compareTo(b.price));
    return open.firstOrNull;
  }

  String? get thumbnail => images.isEmpty ? null : images.first;

  static const listFields = 'id title rating price price_currency in_stock num_available images { original display_order }';
  static const detailFields =
      'id title description rating category_ids price price_currency in_stock num_available images { original display_order } recommendations { product_id ranking } stockrecords { id price price_currency num_available }';

  factory Product.fromJson(Map<String, dynamic> json) {
    final images = [for (final i in (json['images'] as List? ?? const [])) i as Map<String, dynamic>]
      ..sort((a, b) => ((a['display_order'] as num?) ?? 0).compareTo((b['display_order'] as num?) ?? 0));
    final recs = [for (final r in (json['recommendations'] as List? ?? const [])) r as Map<String, dynamic>]
      ..sort((a, b) => ((a['ranking'] as num?) ?? 0).compareTo((b['ranking'] as num?) ?? 0));
    return Product(
      id: (json['id'] as num).toInt(),
      title: json['title'] as String? ?? '',
      description: json['description'] as String? ?? '',
      rating: (json['rating'] as num?)?.toDouble(),
      images: [for (final i in images) i['original'] as String? ?? ''].where((s) => s.isNotEmpty).toList(),
      categoryIds: [for (final c in (json['category_ids'] as List? ?? const [])) (c as num).toInt()],
      recommendationIds: [for (final r in recs) (r['product_id'] as num).toInt()],
      price: (json['price'] as num?)?.toDouble(),
      currency: json['price_currency'] as String? ?? 'VND',
      available: (json['num_available'] as num?)?.toInt(),
      inStock: json['in_stock'] as bool? ?? false,
      offers: [
        for (final o in (json['stockrecords'] as List? ?? const []).cast<Map<String, dynamic>>())
          Offer(
            id: (o['id'] as num).toInt(),
            price: (o['price'] as num?)?.toDouble() ?? 0,
            currency: o['price_currency'] as String? ?? 'VND',
            available: (o['num_available'] as num?)?.toInt(),
          ),
      ],
    );
  }

  @override
  List<Object?> get props => [id, title];
}

class WishlistItem extends Equatable {
  const WishlistItem({required this.lineId, required this.productId, required this.title, required this.quantity});

  final int lineId;
  final int productId;
  final String title;
  final int quantity;

  @override
  List<Object?> get props => [lineId, productId, quantity];
}

class MyWishlist extends Equatable {
  const MyWishlist({required this.id, required this.name, this.items = const []});

  final int id;
  final String name;
  final List<WishlistItem> items;

  bool contains(int productId) => items.any((i) => i.productId == productId);

  factory MyWishlist.fromJson(Map<String, dynamic> json) => MyWishlist(
    id: (json['id'] as num).toInt(),
    name: json['name'] as String? ?? '',
    items: [
      for (final l in (json['lines'] as List? ?? const []))
        WishlistItem(
          lineId: ((l as Map)['id'] as num).toInt(),
          productId: (l['product_id'] as num).toInt(),
          title: l['title'] as String? ?? '',
          quantity: (l['quantity'] as num?)?.toInt() ?? 1,
        ),
    ],
  );

  @override
  List<Object?> get props => [id, items];
}

class Coupon extends Equatable {
  const Coupon({
    required this.id,
    required this.code,
    required this.type,
    required this.value,
    this.minimumPurchase = 0,
    this.maxDiscount,
    this.description = '',
    this.endDate,
  });

  final int id;
  final String code;
  final String type;
  final double value;
  final double minimumPurchase;
  final double? maxDiscount;
  final String description;
  final String? endDate;

  bool get isPercentage => type.toLowerCase().contains('percent');

  static const fields = 'id code type discount_value minimum_purchase max_discount_amount description end_date';

  factory Coupon.fromJson(Map<String, dynamic> json) => Coupon(
    id: (json['id'] as num).toInt(),
    code: json['code'] as String? ?? '',
    type: json['type'] as String? ?? '',
    value: (json['discount_value'] as num?)?.toDouble() ?? 0,
    minimumPurchase: (json['minimum_purchase'] as num?)?.toDouble() ?? 0,
    maxDiscount: (json['max_discount_amount'] as num?)?.toDouble(),
    description: json['description'] as String? ?? '',
    endDate: json['end_date'] as String?,
  );

  @override
  List<Object?> get props => [id, code];
}

class UserCoupon extends Equatable {
  const UserCoupon({required this.id, required this.couponId, required this.isUsed, this.assignedAt, this.coupon});

  final int id;
  final int couponId;
  final bool isUsed;
  final String? assignedAt;
  final Coupon? coupon;

  @override
  List<Object?> get props => [id, couponId, isUsed, coupon];
}

class CartLine extends Equatable {
  const CartLine({required this.line, this.product});

  final BasketLine line;
  final Product? product;

  int get id => line.id ?? 0;
  int get quantity => line.quantity ?? 0;
  double get unitPrice => line.priceInclTax ?? 0;
  double get total => unitPrice * quantity;
  String get title => product?.title ?? 'Sản phẩm #${line.productId}';

  @override
  List<Object?> get props => [line.id, line.quantity, line.priceInclTax, product];
}

class ShoppingRepository {
  ShoppingRepository(this._client) : _api = GatewayApi(_client);

  final GraphQLClient _client;
  final GatewayApi _api;

  Future<BasketBasket> openBasket() async {
    final baskets = await _api.basket.basketBaskets() ?? const [];
    final open = baskets.where((b) => b.status == 'Open' || b.status == '').toList()..sort((a, b) => (b.id ?? 0).compareTo(a.id ?? 0));
    if (open.isNotEmpty) return open.first;
    return (await _api.basket.basketCreateBasket(input: const BasketBasketInput(status: 'Open')))!;
  }

  Future<List<BasketBasket>> purchases() async {
    final baskets = await _api.basket.basketBaskets() ?? const [];
    return baskets.where((b) => b.status == 'Submitted').toList()..sort((a, b) => (b.id ?? 0).compareTo(a.id ?? 0));
  }

  Future<List<CartLine>> cartLines(int basketId) async {
    final lines = await _api.basket.basketBasketLines(id: basketId) ?? const [];
    final products = await productsByIds({for (final l in lines) l.productId ?? 0}.where((id) => id > 0).toList());
    final byId = {for (final p in products) p.id: p};
    return [for (final l in lines) CartLine(line: l, product: byId[l.productId])];
  }

  Future<void> addToCart(int basketId, Offer offer, int quantity) async {
    await _api.basket.basketCreateLine(
      input: BasketLineInput(basketId: basketId, stockrecordId: offer.id, quantity: quantity),
    );
  }

  Future<void> setQuantity(int lineId, int quantity) async {
    if (quantity <= 0) {
      await _api.basket.basketDeleteLine(id: lineId);
    } else {
      await _api.basket.basketUpdateLine(
        id: lineId,
        input: BasketLineInput(quantity: quantity),
      );
    }
  }

  Future<BasketBasket> checkout({required int basketId, required String walletId, required String pin, int? addressId, double? shippingFee}) async {
    return (await _api.basket.basketCheckout(id: basketId, walletId: walletId, pin: pin, shippingAddressId: addressId, shippingFee: shippingFee))!;
  }

  Future<BasketShippingQuote?> shippingQuote({required int basketId, required int addressId}) =>
      _api.basket.basketShippingQuote(id: basketId, shippingAddressId: addressId);

  Future<List<Category>> categories() async {
    final data = await _client.public('query { catalogueCategories { id name parent_id image is_public } }');
    return [
      for (final c in data['catalogueCategories'] as List)
        if ((c as Map)['is_public'] != false) Category.fromJson(c as Map<String, dynamic>),
    ];
  }

  Future<List<Product>> products({String? query, int? categoryId, int limit = 20, int offset = 0}) async {
    final data = await _client.public(
      '''
      query Products(\$q: String, \$cat: Int, \$limit: Int, \$offset: Int) {
        catalogueProducts(q: \$q, category_id: \$cat, limit: \$limit, offset: \$offset) { ${Product.listFields} }
      }
    ''',
      variables: {'q': query, 'cat': categoryId, 'limit': limit, 'offset': offset},
    );
    return [for (final p in data['catalogueProducts'] as List) Product.fromJson(p as Map<String, dynamic>)];
  }

  Future<Product?> product(int id) async {
    final data = await _client.public(
      '''
      query Product(\$id: Int!) { catalogueProduct(id: \$id) { ${Product.detailFields} } }
    ''',
      variables: {'id': id},
    );
    final p = data['catalogueProduct'];
    return p == null ? null : Product.fromJson(p as Map<String, dynamic>);
  }

  Future<List<Product>> productsByIds(List<int> ids) async {
    if (ids.isEmpty) return const [];
    final fields = [for (final (i, id) in ids.indexed) 'p$i: catalogueProduct(id: $id) { ${Product.listFields} }'].join('\n');
    final data = await _client.public('query { $fields }');
    return [
      for (var i = 0; i < ids.length; i++)
        if (data['p$i'] != null) Product.fromJson(data['p$i'] as Map<String, dynamic>),
    ];
  }

  Future<MyWishlist?> wishlist() async {
    final data = await _client.query('query { wishlistDefault { id name lines { id product_id title quantity } } }');
    final w = data['wishlistDefault'];
    return w == null ? null : MyWishlist.fromJson(w as Map<String, dynamic>);
  }

  Future<void> addToWishlist(Product product) async {
    await _client.mutation(
      r'''
      mutation Add($input: WishlistLineInput!) { wishlistQuickAdd(input: $input) { id } }
    ''',
      variables: {
        'input': {'product_id': product.id, 'title': product.title, 'quantity': 1},
      },
    );
  }

  Future<void> removeFromWishlist(int wishlistId, int lineId) async {
    await _client.mutation(
      r'''
      mutation Remove($id: Int64!, $line: Int64!) { wishlistRemoveLine(id: $id, lineId: $line) }
    ''',
      variables: {'id': wishlistId, 'line': lineId},
    );
  }

  Future<List<UserCoupon>> myCoupons() async {
    final data = await _client.query('query { myCoupons { id coupon_id is_used assigned_at coupon { ${Coupon.fields} } } }');
    return [
      for (final c in data['myCoupons'] as List)
        UserCoupon(
          id: ((c as Map)['id'] as num).toInt(),
          couponId: (c['coupon_id'] as num).toInt(),
          isUsed: c['is_used'] as bool? ?? false,
          assignedAt: c['assigned_at'] as String?,
          coupon: c['coupon'] == null ? null : Coupon.fromJson(c['coupon'] as Map<String, dynamic>),
        ),
    ];
  }

  Future<Coupon?> couponByCode(String code) async {
    final data = await _client.query(
      '''
      query Coupon(\$code: String!) { couponByCode(code: \$code) { ${Coupon.fields} } }
    ''',
      variables: {'code': code},
    );
    final c = data['couponByCode'];
    return c == null ? null : Coupon.fromJson(c as Map<String, dynamic>);
  }
}
