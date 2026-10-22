import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';
import 'package:keikiapp/app/app.dart';
import 'package:keikiapp/core/network/graphql_client.dart';
import 'package:keikiapp/features/apps/commerce/commerce_more.dart';
import 'package:keikiapp/features/apps/entertainment/polymarket_more.dart';
import 'package:keikiapp/features/apps/entertainment/ticket_organizer.dart';
import 'package:keikiapp/features/apps/extras/extras.dart';
import 'package:keikiapp/features/apps/travel/stays_host.dart';

import 'helpers.dart';

GraphQLClient _client() => GraphQLClient(
      endpoint: 'http://gw/query',
      storage: MemoryTokenStorage(validTokens()),
      httpClient: MockClient((req) async {
        final query = (jsonDecode(req.body) as Map<String, dynamic>)['query'] as String;
        final data = query.contains('me {')
            ? {
                'me': {'id': 7, 'username': 'quan', 'email': 'q@keiki.vn', 'name': 'Quan', 'roles': []},
              }
            : <String, Object?>{};
        return http.Response(jsonEncode({'data': data}), 200, headers: {'content-type': 'application/json'});
      }),
    );

final pages = <String, Widget Function()>{
  'ticket organizer': () => const TicketOrganizerPage(),
  'hotel host': () => const HotelHostPage(),
  'homestay host': () => const HomestayHostPage(),
  'poly market': () => const PolyMarketPage(id: '1'),
  'poly creator': () => const PolyCreatorPage(),
  'wallet more': () => const WalletMorePage(),
  'shop more': () => const ShopMorePage(),
  'food search': () => const FoodSearchPage(),
  'driver jobs': () => const DriverJobsPage(),
  'qr codes': () => const QrCodesPage(),
  'security': () => const SecurityPage(),
  'voucher business': () => const VoucherBusinessPage(),
  'referral tools': () => const ReferralToolsPage(),
  'serp tools': () => const SerpToolsPage(),
  'shortlink tools': () => const ShortlinkToolsPage(),
  'vend catalog': () => const VendCatalogPage(),
  'geo extras': () => const GeoExtrasPage(),
  'ai studio': () => const AiStudioPage(),
  'bonus partner': () => const BonusPartnerPage(),
  'car staff': () => const CarStaffPage(),
  'post categories': () => const PostCategoryManager(),
  'chat group search': () => const ChatGroupSearchPage(),
  'chat group members': () => const ChatGroupMembersPage(groupId: 1, title: 'Nhóm'),
};

void main() {
  for (final entry in pages.entries) {
    testWidgets('${entry.key} page opens on a small phone', (tester) async {
      tester.view.physicalSize = const Size(720, 1280);
      tester.view.devicePixelRatio = 2;
      addTearDown(tester.view.reset);
      await tester.pumpWidget(KeikiApp(client: _client(), catalogLoader: loadTestCatalog));
      await tester.pumpAndSettle();
      Navigator.of(tester.element(find.byType(Scaffold).first)).push(MaterialPageRoute<void>(builder: (_) => entry.value()));
      await tester.pumpAndSettle();
      expect(tester.takeException(), isNull);
      final tabs = find.byType(Tab);
      for (var i = 1; i < tabs.evaluate().length; i++) {
        await tester.tap(tabs.at(i));
        await tester.pumpAndSettle();
        expect(tester.takeException(), isNull);
      }
      await tester.pump(const Duration(seconds: 3));
    });
  }
}
