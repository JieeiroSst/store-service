// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

import '../gateway_executor.dart';
import '../models.g.dart';

class BasketApi {
  const BasketApi(this._executor);

  final GatewayExecutor _executor;

  Future<List<BasketBasket>?> basketBaskets({String? selectionSet}) async {
    final data = await _executor.query('query basketBaskets { basketBaskets ${selectionSet ?? '{ id status date_placed date_created date_merged date_submitted billing_address_id shipping_address_id user_id currency total_incl_tax payment_reference fulfillment_id }'} }'
    );
    return (data['basketBaskets'] as List?)?.map((e) => (e == null ? null : BasketBasket.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<BasketBasket?> basketBasket({required int id, String? selectionSet}) async {
    final data = await _executor.query('query basketBasket(\$id: Int!) { basketBasket(id: \$id) ${selectionSet ?? '{ id status date_placed date_created date_merged date_submitted billing_address_id shipping_address_id user_id currency total_incl_tax payment_reference fulfillment_id }'} }', variables: {
      'id': id,
    }
    );
    return data['basketBasket'] == null ? null : BasketBasket.fromJson((data['basketBasket'] as Map).cast<String, dynamic>());
  }

  Future<List<BasketLine>?> basketLines({String? selectionSet}) async {
    final data = await _executor.query('query basketLines { basketLines ${selectionSet ?? '{ id line_reference quantity price_currency price_excl_tax price_incl_tax basket_id product_id stockrecord_id }'} }'
    );
    return (data['basketLines'] as List?)?.map((e) => (e == null ? null : BasketLine.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<BasketLine?> basketLine({required int id, String? selectionSet}) async {
    final data = await _executor.query('query basketLine(\$id: Int!) { basketLine(id: \$id) ${selectionSet ?? '{ id line_reference quantity price_currency price_excl_tax price_incl_tax basket_id product_id stockrecord_id }'} }', variables: {
      'id': id,
    }
    );
    return data['basketLine'] == null ? null : BasketLine.fromJson((data['basketLine'] as Map).cast<String, dynamic>());
  }

  Future<List<BasketLineAttribute>?> basketLineAttributes({String? selectionSet}) async {
    final data = await _executor.query('query basketLineAttributes { basketLineAttributes ${selectionSet ?? '{ id basket_id line_id option_id value }'} }'
    );
    return (data['basketLineAttributes'] as List?)?.map((e) => (e == null ? null : BasketLineAttribute.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<BasketLineAttribute?> basketLineAttribute({required int id, String? selectionSet}) async {
    final data = await _executor.query('query basketLineAttribute(\$id: Int!) { basketLineAttribute(id: \$id) ${selectionSet ?? '{ id basket_id line_id option_id value }'} }', variables: {
      'id': id,
    }
    );
    return data['basketLineAttribute'] == null ? null : BasketLineAttribute.fromJson((data['basketLineAttribute'] as Map).cast<String, dynamic>());
  }

  Future<List<BasketOrder>?> basketOrders({String? selectionSet}) async {
    final data = await _executor.query('query basketOrders { basketOrders ${selectionSet ?? '{ id number currency total_incl_tax total_excl_tax shipping_incl_tax shipping_excl_tax shipping_method shipping_code status guest_email site_id user_id owner_id }'} }'
    );
    return (data['basketOrders'] as List?)?.map((e) => (e == null ? null : BasketOrder.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<BasketOrder?> basketOrder({required int id, String? selectionSet}) async {
    final data = await _executor.query('query basketOrder(\$id: Int!) { basketOrder(id: \$id) ${selectionSet ?? '{ id number currency total_incl_tax total_excl_tax shipping_incl_tax shipping_excl_tax shipping_method shipping_code status guest_email site_id user_id owner_id }'} }', variables: {
      'id': id,
    }
    );
    return data['basketOrder'] == null ? null : BasketOrder.fromJson((data['basketOrder'] as Map).cast<String, dynamic>());
  }

  Future<BasketUser?> basketUser({required int id, String? selectionSet}) async {
    final data = await _executor.query('query basketUser(\$id: Int!) { basketUser(id: \$id) ${selectionSet ?? '{ id username email name active }'} }', variables: {
      'id': id,
    }
    );
    return data['basketUser'] == null ? null : BasketUser.fromJson((data['basketUser'] as Map).cast<String, dynamic>());
  }

  Future<List<BasketLine>?> basketBasketLines({required int id, String? selectionSet}) async {
    final data = await _executor.query('query basketBasketLines(\$id: Int!) { basketBasketLines(id: \$id) ${selectionSet ?? '{ id line_reference quantity price_currency price_excl_tax price_incl_tax basket_id product_id stockrecord_id }'} }', variables: {
      'id': id,
    }
    );
    return (data['basketBasketLines'] as List?)?.map((e) => (e == null ? null : BasketLine.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<BasketBasket?> basketCreateBasket({required BasketBasketInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation basketCreateBasket(\$input: BasketBasketInput!) { basketCreateBasket(input: \$input) ${selectionSet ?? '{ id status date_placed date_created date_merged date_submitted billing_address_id shipping_address_id user_id currency total_incl_tax payment_reference fulfillment_id }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['basketCreateBasket'] == null ? null : BasketBasket.fromJson((data['basketCreateBasket'] as Map).cast<String, dynamic>());
  }

  Future<BasketBasket?> basketUpdateBasket({required int id, required BasketBasketInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation basketUpdateBasket(\$id: Int!, \$input: BasketBasketInput!) { basketUpdateBasket(id: \$id, input: \$input) ${selectionSet ?? '{ id status date_placed date_created date_merged date_submitted billing_address_id shipping_address_id user_id currency total_incl_tax payment_reference fulfillment_id }'} }', variables: {
      'id': id,
      'input': input.toJson(),
    }
    );
    return data['basketUpdateBasket'] == null ? null : BasketBasket.fromJson((data['basketUpdateBasket'] as Map).cast<String, dynamic>());
  }

  Future<bool?> basketDeleteBasket({required int id}) async {
    final data = await _executor.mutation('mutation basketDeleteBasket(\$id: Int!) { basketDeleteBasket(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['basketDeleteBasket'] as bool?;
  }

  Future<BasketLine?> basketCreateLine({required BasketLineInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation basketCreateLine(\$input: BasketLineInput!) { basketCreateLine(input: \$input) ${selectionSet ?? '{ id line_reference quantity price_currency price_excl_tax price_incl_tax basket_id product_id stockrecord_id }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['basketCreateLine'] == null ? null : BasketLine.fromJson((data['basketCreateLine'] as Map).cast<String, dynamic>());
  }

  Future<BasketLine?> basketUpdateLine({required int id, required BasketLineInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation basketUpdateLine(\$id: Int!, \$input: BasketLineInput!) { basketUpdateLine(id: \$id, input: \$input) ${selectionSet ?? '{ id line_reference quantity price_currency price_excl_tax price_incl_tax basket_id product_id stockrecord_id }'} }', variables: {
      'id': id,
      'input': input.toJson(),
    }
    );
    return data['basketUpdateLine'] == null ? null : BasketLine.fromJson((data['basketUpdateLine'] as Map).cast<String, dynamic>());
  }

  Future<bool?> basketDeleteLine({required int id}) async {
    final data = await _executor.mutation('mutation basketDeleteLine(\$id: Int!) { basketDeleteLine(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['basketDeleteLine'] as bool?;
  }

  Future<BasketLineAttribute?> basketCreateLineAttribute({required BasketLineAttributeInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation basketCreateLineAttribute(\$input: BasketLineAttributeInput!) { basketCreateLineAttribute(input: \$input) ${selectionSet ?? '{ id basket_id line_id option_id value }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['basketCreateLineAttribute'] == null ? null : BasketLineAttribute.fromJson((data['basketCreateLineAttribute'] as Map).cast<String, dynamic>());
  }

  Future<BasketLineAttribute?> basketUpdateLineAttribute({required int id, required BasketLineAttributeInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation basketUpdateLineAttribute(\$id: Int!, \$input: BasketLineAttributeInput!) { basketUpdateLineAttribute(id: \$id, input: \$input) ${selectionSet ?? '{ id basket_id line_id option_id value }'} }', variables: {
      'id': id,
      'input': input.toJson(),
    }
    );
    return data['basketUpdateLineAttribute'] == null ? null : BasketLineAttribute.fromJson((data['basketUpdateLineAttribute'] as Map).cast<String, dynamic>());
  }

  Future<bool?> basketDeleteLineAttribute({required int id}) async {
    final data = await _executor.mutation('mutation basketDeleteLineAttribute(\$id: Int!) { basketDeleteLineAttribute(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['basketDeleteLineAttribute'] as bool?;
  }

  Future<BasketBasket?> basketCheckout({required int id, required String walletId, required String pin, int? shippingAddressId, String? selectionSet}) async {
    final data = await _executor.mutation('mutation basketCheckout(\$id: Int!, \$wallet_id: String!, \$pin: String!, \$shipping_address_id: Int) { basketCheckout(id: \$id, wallet_id: \$wallet_id, pin: \$pin, shipping_address_id: \$shipping_address_id) ${selectionSet ?? '{ id status date_placed date_created date_merged date_submitted billing_address_id shipping_address_id user_id currency total_incl_tax payment_reference fulfillment_id }'} }', variables: {
      'id': id,
      'wallet_id': walletId,
      'pin': pin,
      'shipping_address_id': ?shippingAddressId,
    }
    );
    return data['basketCheckout'] == null ? null : BasketBasket.fromJson((data['basketCheckout'] as Map).cast<String, dynamic>());
  }

}
