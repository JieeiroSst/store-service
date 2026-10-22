import 'package:equatable/equatable.dart';

import '../../../core/network/graphql_client.dart';

class Restaurant extends Equatable {
  const Restaurant({required this.id, required this.name, this.description = '', this.address = '', this.imageUrl});

  final int id;
  final String name;
  final String description;
  final String address;
  final String? imageUrl;

  static const fields = 'id name description address image_url is_active';

  factory Restaurant.fromJson(Map<String, dynamic> json) => Restaurant(
        id: (json['id'] as num?)?.toInt() ?? 0,
        name: json['name'] as String? ?? '',
        description: json['description'] as String? ?? '',
        address: json['address'] as String? ?? '',
        imageUrl: json['image_url'] as String?,
      );

  @override
  List<Object?> get props => [id, name];
}

class Food extends Equatable {
  const Food({required this.id, required this.name, this.description = '', this.price = 0, this.category = '', this.restaurantId});

  final int id;
  final String name;
  final String description;
  final int price;
  final String category;
  final int? restaurantId;

  factory Food.fromJson(Map<String, dynamic> json) => Food(
        id: (json['id'] as num?)?.toInt() ?? 0,
        name: json['name'] as String? ?? '',
        description: json['description'] as String? ?? '',
        price: (json['price'] as num?)?.toInt() ?? 0,
        category: json['category'] as String? ?? '',
        restaurantId: (json['restaurant_id'] as num?)?.toInt(),
      );

  @override
  List<Object?> get props => [id, name, price];
}

class FoodOrderItem extends Equatable {
  const FoodOrderItem({required this.name, required this.quantity, required this.unitPrice});

  final String name;
  final int quantity;
  final double unitPrice;

  factory FoodOrderItem.fromJson(Map<String, dynamic> json) => FoodOrderItem(
        name: json['name'] as String? ?? '',
        quantity: (json['quantity'] as num?)?.toInt() ?? 0,
        unitPrice: (json['unit_price'] as num?)?.toDouble() ?? 0,
      );

  @override
  List<Object?> get props => [name, quantity, unitPrice];
}

class FoodOrder extends Equatable {
  const FoodOrder({
    required this.id,
    required this.status,
    required this.paymentStatus,
    required this.total,
    required this.deliveryFee,
    this.placedAt,
    this.estimatedDelivery,
    this.items = const [],
  });

  final int id;
  final String status;
  final String paymentStatus;
  final double total;
  final double deliveryFee;
  final String? placedAt;
  final String? estimatedDelivery;
  final List<FoodOrderItem> items;

  bool get isCancellable => const {'created', 'confirmed'}.contains(status.toLowerCase());

  bool get isActive => !const {'delivered', 'cancelled'}.contains(status.toLowerCase());

  static const fields = '''
    id status payment_status total_amount delivery_fee placed_at estimated_delivery_time
    items { name quantity unit_price }
  ''';

  factory FoodOrder.fromJson(Map<String, dynamic> json) => FoodOrder(
        id: (json['id'] as num).toInt(),
        status: json['status'] as String? ?? '',
        paymentStatus: json['payment_status'] as String? ?? '',
        total: (json['total_amount'] as num?)?.toDouble() ?? 0,
        deliveryFee: (json['delivery_fee'] as num?)?.toDouble() ?? 0,
        placedAt: json['placed_at'] as String?,
        estimatedDelivery: json['estimated_delivery_time'] as String?,
        items: [for (final i in (json['items'] as List? ?? const [])) FoodOrderItem.fromJson(i as Map<String, dynamic>)],
      );

  @override
  List<Object?> get props => [id, status, paymentStatus, total];
}

class TrackingEvent extends Equatable {
  const TrackingEvent({required this.status, this.notes = '', this.timestamp, this.lat, this.lng});

  final String status;
  final String notes;
  final String? timestamp;
  final double? lat;
  final double? lng;

  factory TrackingEvent.fromJson(Map<String, dynamic> json) => TrackingEvent(
        status: json['status'] as String? ?? '',
        notes: json['notes'] as String? ?? '',
        timestamp: json['timestamp'] as String?,
        lat: (json['latitude'] as num?)?.toDouble(),
        lng: (json['longitude'] as num?)?.toDouble(),
      );

  @override
  List<Object?> get props => [status, timestamp];
}

class FoodRepository {
  FoodRepository(this._client);

  final GraphQLClient _client;

  static const _foodFields = 'id name description price category restaurant_id';

  Future<List<Restaurant>> restaurants({String? query}) async {
    final data = await _client.public('''
      query Restaurants(\$q: String) { foodRestaurants(q: \$q) { ${Restaurant.fields} } }
    ''', variables: {'q': ?query});
    return [for (final r in (data['foodRestaurants'] as List? ?? const [])) Restaurant.fromJson(r as Map<String, dynamic>)];
  }

  Future<List<Food>> menu(int restaurantId) async {
    final data = await _client.public('''
      query Menu(\$id: Int!) { foodRestaurantMenu(id: \$id) { $_foodFields } }
    ''', variables: {'id': restaurantId});
    return [for (final f in (data['foodRestaurantMenu'] as List? ?? const [])) Food.fromJson(f as Map<String, dynamic>)];
  }

  Future<FoodOrder> placeOrder({
    required String restaurantId,
    required String deliveryAddressId,
    required String paymentMethodId,
    required Map<Food, int> items,
    double deliveryFee = 0,
    double tip = 0,
    String? note,
  }) async {
    final data = await _client.mutation('''
      mutation Place(\$input: DoordashOrderInput!) { doordashPlaceOrder(input: \$input) { ${FoodOrder.fields} } }
    ''', variables: {
      'input': {
        'restaurant_id': restaurantId,
        'delivery_address_id': deliveryAddressId,
        'payment_method_id': paymentMethodId,
        'delivery_fee': deliveryFee,
        'service_fee': 0,
        'tax': 0,
        'tip': tip,
        'special_instructions': note ?? '',
        'items': [
          for (final e in items.entries) {'menu_item_id': '${e.key.id}', 'quantity': e.value},
        ],
      },
    });
    return FoodOrder.fromJson(data['doordashPlaceOrder'] as Map<String, dynamic>);
  }

  Future<List<FoodOrder>> myOrders() async {
    final data = await _client.query('query { myDoordashOrders { ${FoodOrder.fields} } }');
    final orders = [for (final o in data['myDoordashOrders'] as List) FoodOrder.fromJson(o as Map<String, dynamic>)];
    return orders..sort((a, b) => b.id.compareTo(a.id));
  }

  Future<FoodOrder?> order(int id) async {
    final data = await _client.query('''
      query Order(\$id: Int!) { doordashMyOrder(id: \$id) { ${FoodOrder.fields} } }
    ''', variables: {'id': id});
    final o = data['doordashMyOrder'];
    return o == null ? null : FoodOrder.fromJson(o as Map<String, dynamic>);
  }

  Future<List<TrackingEvent>> tracking(int orderId) async {
    final data = await _client.query(r'''
      query Tracking($id: Int!) { doordashMyOrderTracking(id: $id) { status notes timestamp latitude longitude } }
    ''', variables: {'id': orderId});
    return [for (final t in data['doordashMyOrderTracking'] as List) TrackingEvent.fromJson(t as Map<String, dynamic>)];
  }

  Future<void> cancel(int orderId, String reason) async {
    await _client.mutation(r'''
      mutation Cancel($id: Int!, $reason: String) { doordashCancelMyOrder(id: $id, reason: $reason) { id status } }
    ''', variables: {'id': orderId, 'reason': reason});
  }
}
