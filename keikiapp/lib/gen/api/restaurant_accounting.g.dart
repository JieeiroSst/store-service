// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

import '../gateway_executor.dart';
import '../models.g.dart';

class RestaurantAccountingApi {
  const RestaurantAccountingApi(this._executor);

  final GatewayExecutor _executor;

  Future<RAccountingPayment?> rAccountingPayment({required int orderID, String? selectionSet}) async {
    final data = await _executor.query('query rAccountingPayment(\$orderID: Int!) { rAccountingPayment(orderID: \$orderID) ${selectionSet ?? '{ id order_id amount method status }'} }', variables: {
      'orderID': orderID,
    }
    );
    return data['rAccountingPayment'] == null ? null : RAccountingPayment.fromJson((data['rAccountingPayment'] as Map).cast<String, dynamic>());
  }

  Future<Object?> rAccountingAuthCart({required Map<String, dynamic> input}) async {
    final data = await _executor.mutation('mutation rAccountingAuthCart(\$input: Map!) { rAccountingAuthCart(input: \$input) }', variables: {
      'input': input,
    }
    );
    return data['rAccountingAuthCart'];
  }

  Future<Object?> rAccountingMarkPaid({required String orderID}) async {
    final data = await _executor.mutation('mutation rAccountingMarkPaid(\$orderID: String!) { rAccountingMarkPaid(orderID: \$orderID) }', variables: {
      'orderID': orderID,
    }
    );
    return data['rAccountingMarkPaid'];
  }

}
