import 'dart:convert';

import 'package:flutter_test/flutter_test.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';
import 'package:keikiapp/core/network/graphql_client.dart';
import 'package:keikiapp/core/utils/load_status.dart';
import 'package:keikiapp/features/shopping/bloc/cart_cubit.dart';
import 'package:keikiapp/features/shopping/data/shopping_repository.dart';

import '../helpers.dart';

class _FakeGateway {
  final requests = <Map<String, dynamic>>[];
  final baskets = <Map<String, dynamic>>[];
  final lines = <Map<String, dynamic>>[];

  Map<String, dynamic> handle(Map<String, dynamic> body) {
    requests.add(body);
    final q = body['query'] as String;
    final v = (body['variables'] as Map<String, dynamic>?) ?? const {};
    if (q.startsWith('query basketBaskets')) return {'basketBaskets': baskets};
    if (q.startsWith('mutation basketCreateBasket')) {
      final b = {'id': baskets.length + 1, 'status': 'Open', 'user_id': 7, 'total_incl_tax': 0, 'currency': ''};
      baskets.add(b);
      return {'basketCreateBasket': b};
    }
    if (q.startsWith('query basketBasketLines')) return {'basketBasketLines': lines.where((l) => l['basket_id'] == v['id']).toList()};
    if (q.startsWith('mutation basketCreateLine')) {
      final input = v['input'] as Map<String, dynamic>;
      final line = {
        'id': lines.length + 1,
        'basket_id': input['basket_id'],
        'stockrecord_id': input['stockrecord_id'],
        'product_id': 42,
        'quantity': input['quantity'],
        'price_incl_tax': 150000,
        'price_currency': 'VND',
      };
      lines.add(line);
      return {'basketCreateLine': line};
    }
    if (q.startsWith('mutation basketUpdateLine')) {
      final line = lines.firstWhere((l) => l['id'] == v['id']);
      line['quantity'] = (v['input'] as Map)['quantity'];
      return {'basketUpdateLine': line};
    }
    if (q.startsWith('mutation basketDeleteLine')) {
      lines.removeWhere((l) => l['id'] == v['id']);
      return {'basketDeleteLine': true};
    }
    if (q.startsWith('mutation basketCheckout')) {
      final b = baskets.firstWhere((b) => b['id'] == v['id']);
      b
        ..['status'] = 'Submitted'
        ..['total_incl_tax'] = 300000
        ..['currency'] = 'VND';
      return {'basketCheckout': b};
    }
    if (q.contains('catalogueProduct(id: 42)')) {
      return {
        'p0': {'id': 42, 'title': 'Áo thun Keiki', 'images': []},
      };
    }
    return {};
  }
}

void main() {
  late _FakeGateway gateway;
  late CartCubit cubit;

  setUp(() {
    gateway = _FakeGateway();
    final client = GraphQLClient(
      endpoint: 'http://gw/query',
      storage: MemoryTokenStorage(validTokens()),
      httpClient: MockClient((req) async {
        final data = gateway.handle(jsonDecode(req.body) as Map<String, dynamic>);
        return http.Response(jsonEncode({'data': data}), 200, headers: {'content-type': 'application/json'});
      }),
    );
    cubit = CartCubit(ShoppingRepository(client));
  });

  test('opens a basket, adds priced lines, changes quantity and checks out', () async {
    await cubit.load();
    expect(cubit.state.status, LoadStatus.success);
    expect(gateway.baskets, hasLength(1));

    await cubit.add(const Offer(id: 9, price: 150000, currency: 'VND', available: 5), 1);
    expect(cubit.state.lines.single.title, 'Áo thun Keiki');
    final create = gateway.requests.firstWhere((r) => (r['query'] as String).startsWith('mutation basketCreateLine'));
    expect(create['variables'], {
      'input': {'basket_id': 1, 'stockrecord_id': 9, 'quantity': 1},
    });

    await cubit.setQuantity(cubit.state.lines.single, 2);
    expect(cubit.state.count, 2);
    expect(cubit.state.total, 300000);

    final done = await cubit.checkout(walletId: 'w-1', pin: '123456', addressId: 3);
    expect(done!.status, 'Submitted');
    final checkout = gateway.requests.firstWhere((r) => (r['query'] as String).startsWith('mutation basketCheckout'));
    expect(checkout['variables'], {'id': 1, 'wallet_id': 'w-1', 'pin': '123456', 'shipping_address_id': 3});
    expect(gateway.baskets, hasLength(2), reason: 'a fresh open basket is created after checkout');
    expect(cubit.state.lines, isEmpty);
  });

  test('setting quantity to zero removes the line', () async {
    await cubit.load();
    await cubit.add(const Offer(id: 9, price: 1, currency: 'VND'), 1);
    await cubit.setQuantity(cubit.state.lines.single, 0);
    expect(cubit.state.lines, isEmpty);
    expect(gateway.requests.where((r) => (r['query'] as String).startsWith('mutation basketDeleteLine')), hasLength(1));
  });
}
