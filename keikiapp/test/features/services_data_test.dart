import 'dart:convert';
import 'dart:io';
import 'dart:typed_data';

import 'package:flutter_test/flutter_test.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';
import 'package:keikiapp/core/network/graphql_client.dart';
import 'package:keikiapp/features/services/data/humanize.dart';
import 'package:keikiapp/features/services/data/permissions.dart';
import 'package:keikiapp/features/services/data/schema_catalog.dart';
import 'package:keikiapp/gen/gateway_api.g.dart';

import '../helpers.dart';

void main() {
  late SchemaCatalog catalog;

  setUpAll(() {
    catalog = SchemaCatalog.parse(File('assets/gateway_schema.json').readAsStringSync());
  });

  test('catalog covers every gateway root field', () {
    final ops = catalog.services.expand((s) => s.operations).toList();
    expect(ops.length, greaterThan(1800));
    final topUp = catalog.operation('wallet_service', 'walletSvcMyTopUp')!;
    expect(topUp.isMutation, isTrue);
    expect(topUp.access, OperationAccess.user);
    expect(topUp.args.map((a) => '${a.name}:${a.type}'), contains('idempotency_key:String!'));
    expect(topUp.document(), startsWith(r'mutation walletSvcMyTopUp($wallet_id: String!'));
    final admin = catalog.operation('wallet_service', 'walletSvcCredit')!;
    expect(admin.access, OperationAccess.admin);
    expect(admin.permissionObject, '/graphql/wallet-service');
    expect(catalog.operation('food', 'foods')!.access, OperationAccess.public);
  });

  test('operation titles read naturally in Vietnamese', () {
    expect(operationTitle(catalog.operation('wallet_service', 'walletSvcMyWallets')!), 'Danh sách ví của tôi');
    expect(operationTitle(catalog.operation('reservation', 'resvHotels')!), 'Danh sách khách sạn');
    expect(operationTitle(catalog.operation('reservation', 'resvHotel')!), 'Chi tiết khách sạn');
    expect(operationTitle(catalog.operation('car_rental', 'carRentalSearchAvailableVehicles')!), 'Tìm xe còn trống');
    expect(operationTitle(catalog.operation('payment_wallet', 'myPaymentMethods')!), 'Danh sách phương thức thanh toán của tôi');
    expect(operationTitle(catalog.operation('lucky', 'luckyCheckIn')!), 'Điểm danh');
    expect(operationTitle(catalog.operation('wallet_service', 'walletSvcMyTopUp')!), 'Nạp tiền của tôi');
    expect(operationTitle(catalog.operation('movie', 'movieNewReleases')!), 'Danh sách phim mới');
    expect(fieldLabel('page_size'), 'Page size');
    expect(fieldLabel('wallet_id'), 'Wallet ID');
  });

  test('permissions follow casbin keyMatch and regexMatch', () {
    const perms = UserPermissions(grants: [Grant('/graphql/wallet-*', 'QUERY|MUTATION')]);
    final credit = catalog.operation('wallet_service', 'walletSvcCredit')!;
    final crm = catalog.operation('crm', 'crmConvertLead')!;
    expect(perms.canRun(credit), isTrue);
    expect(perms.canRun(crm), isFalse);
    expect(perms.canRun(catalog.operation('wallet_service', 'walletSvcMyWallets')!), isTrue);
    expect(Grant.keyMatch('/graphql/users', '/graphql/*'), isTrue);
    expect(Grant.keyMatch('/graphql', '/graphql/*'), isFalse);
    expect(Grant.regexMatch('QUERY', 'QUERY'), isTrue);
    expect(Grant.regexMatch('MUTATION', 'QUERY'), isFalse);
  });

  test('typed SDK sends the generated document and decodes models', () async {
    late Map<String, dynamic> body;
    final client = GraphQLClient(
      endpoint: 'http://gw/query',
      storage: MemoryTokenStorage(validTokens()),
      httpClient: MockClient((req) async {
        body = jsonDecode(req.body) as Map<String, dynamic>;
        return http.Response(
          jsonEncode({
            'data': {
              'walletSvcMyWallets': [
                {'id': 'w1', 'currency': 'VND', 'balance': '150000', 'has_pin': true, 'version': '3'},
              ],
            },
          }),
          200,
        );
      }),
    );
    final api = GatewayApi(client);

    final wallets = await api.walletService.walletSvcMyWallets();

    expect(body['query'], contains('walletSvcMyWallets { id user_id balance'));
    expect(wallets!.single.balance, '150000');
    expect(wallets.single.hasPin, isTrue);
    expect(wallets.single.version, 3);
  });

  test('uploads use the GraphQL multipart request spec', () async {
    late http.BaseRequest captured;
    late String operations;
    late String map;
    final client = GraphQLClient(
      endpoint: 'http://gw/query',
      storage: MemoryTokenStorage(validTokens()),
      httpClient: MockClient.streaming((req, stream) async {
        captured = req;
        final raw = utf8.decode(await stream.toBytes(), allowMalformed: true);
        operations = RegExp(r'name="operations"\r\n\r\n(.*?)\r\n--', dotAll: true).firstMatch(raw)!.group(1)!;
        map = RegExp(r'name="map"\r\n\r\n(.*?)\r\n--', dotAll: true).firstMatch(raw)!.group(1)!;
        return http.StreamedResponse(Stream.value(utf8.encode(jsonEncode({'data': {'uploadFile': {'id': 'f1'}}}))), 200);
      }),
    );

    await client.mutation(
      r'mutation ($file: Upload!) { uploadFile(file: $file) { id } }',
      variables: {'file': GqlUpload(bytes: Uint8List.fromList([1, 2, 3]), filename: 'a.png', contentType: 'image/png')},
    );

    expect(captured.headers['content-type'], startsWith('multipart/form-data'));
    expect(captured.headers['Authorization'], 'Bearer access-1');
    expect(jsonDecode(operations)['variables'], {'file': null});
    expect(jsonDecode(map), {
      '0': ['variables.file'],
    });
  });
}
