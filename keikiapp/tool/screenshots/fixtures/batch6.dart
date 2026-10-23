import 'package:flutter/widgets.dart';
import 'package:keikiapp/features/admin/admin_home_page.dart';
import 'package:keikiapp/features/admin/console/admin_console_page.dart';
import 'package:keikiapp/features/admin/console/gen/admin_consoles.g.dart';
import 'package:keikiapp/features/apps/commerce/commerce_more.dart';
import 'package:keikiapp/features/apps/entertainment/polymarket_more.dart';
import 'package:keikiapp/features/apps/entertainment/ticket_organizer.dart';

final batch6 = <String, Object?>{
  'myPermissions': [
    {'object': '/graphql/*', 'action': '.*'},
  ],
  'crmCampaigns': [
    {'id': 1, 'campaign_name': 'Ra mắt Keikibook', 'campaign_sponsor': 'Keiki Group', 'campaign_objectives': 'Thu hút 100.000 người dùng đầu tiên', 'campaign_start_date': '2026-10-01T00:00:00Z'},
    {'id': 2, 'campaign_name': 'Tết Đinh Mùi 2027', 'campaign_sponsor': 'Keiki Pay', 'campaign_objectives': 'Hoàn tiền 10% khi thanh toán bằng ví', 'campaign_start_date': '2027-01-15T00:00:00Z'},
    {'id': 3, 'campaign_name': 'Đối tác khách sạn miền Trung', 'campaign_sponsor': 'Keiki Travel', 'campaign_objectives': 'Ký mới 50 khách sạn', 'campaign_start_date': '2026-11-01T00:00:00Z'},
  ],
  'ticketEventReviewQueue': {
    'items': [
      {'id': 9, 'title': 'Lễ hội ánh sáng Hội An', 'status': 'pending_review'},
      {'id': 10, 'title': 'Giải chạy Keiki Run 2026', 'status': 'pending_review'},
    ],
  },
  'resvHotelReviewQueue': {
    'items': [
      {'id': 4, 'name': 'Sapa Cloud Hotel', 'statusName': 'pending'},
    ],
  },
  'rentHomestayReviewQueue': [],
  'ticketMyEvents': {
    'items': [
      {'id': 1, 'title': 'Đêm nhạc Hà Anh Tuấn – Chân trời rực rỡ', 'status': 'published', 'starts_at': '2026-11-15T12:30:00Z', 'venue': 'Nhà hát Hòa Bình'},
      {'id': 5, 'title': 'Workshop Flutter nâng cao', 'status': 'draft', 'starts_at': '2026-12-02T02:00:00Z', 'venue': 'Keiki Hub'},
    ],
  },
  'ticketEventReport': {
    'orders': 312,
    'tickets_sold': 540,
    'revenue': 612000000,
    'checked_in': 0,
    'by_type': [
      {'name': 'Standard', 'total': 400, 'sold': 380, 'revenue': 304000000},
      {'name': 'VIP', 'total': 120, 'sold': 116, 'revenue': 208800000},
    ],
  },
  'ticketPromotions': [
    {'id': 1, 'code': 'EARLYBIRD', 'kind': 'percent', 'value': 15, 'used': 120, 'max_uses': 200, 'valid_to': '2026-11-01T00:00:00Z', 'active': true},
  ],
  'ticketStaff': [
    {'user_id': 42, 'created_at': '2026-10-10T00:00:00Z'},
  ],
  'polyMarket': {'id': 1, 'question': 'Việt Nam vô địch AFF Cup 2026?', 'yes_price': 62, 'no_price': 38, 'share_value': 100, 'status': 'open', 'volume': 182000000, 'end_time': '2026-12-30T00:00:00Z', 'description': 'Phân xử theo kết quả chính thức của AFF.'},
  'polyOrderBook': {
    'bids': [{'price': 61, 'size': 1200}, {'price': 60, 'size': 800}, {'price': 58, 'size': 2500}],
    'asks': [{'price': 63, 'size': 900}, {'price': 64, 'size': 1500}, {'price': 66, 'size': 400}],
    'spread': 2,
    'midpoint': 62,
    'last_trade_price': 62,
  },
  'polyPriceHistory': [
    for (final (i, c) in [41, 44, 43, 48, 52, 50, 55, 58, 57, 60, 61, 62].indexed) {'time': '2026-10-${10 + i}T00:00:00Z', 'close': c},
  ],
  'polyMarketTrades': [
    {'taker_side': 'buy', 'taker_outcome': 'yes', 'size': 120, 'yes_price': 62, 'created_at': '2026-10-22T02:10:00Z'},
    {'taker_side': 'sell', 'taker_outcome': 'yes', 'size': 40, 'yes_price': 61, 'created_at': '2026-10-22T01:55:00Z'},
  ],
  'polyTopHolders': [
    {'username': 'minhanh', 'shares': 5400},
    {'username': 'tuan.dev', 'shares': 3100},
  ],
  'walletSvcMyWallets': [
    {'id': 'w-1', 'currency': 'VND', 'balance': '2450000', 'status': 'ACTIVE', 'tier': 'VERIFIED'},
  ],
  'walletSvcMyWallet': {'id': 'w-1', 'currency': 'VND', 'balance': '2450000', 'frozen_amount': '125000', 'status': 'ACTIVE', 'tier': 'VERIFIED'},
  'walletSvcMyLimits': {'currency': 'VND', 'limits': {'per_transaction': '20000000', 'daily': '50000000'}, 'used_today': '1200000', 'remaining_today': '48800000'},
  'walletSvcMyStatement': {'currency': 'VND', 'opening_balance': '1800000', 'closing_balance': '2450000', 'total_credit': '2500000', 'total_debit': '1850000', 'credit_count': 4, 'debit_count': 11},
  'walletSvcMyBankAccounts': [
    {'id': 'b1', 'account_name': 'LUU MANH QUAN', 'account_number': '0123456789', 'is_default': true, 'status': 'ACTIVE'},
  ],
  'pointMyTransactions': {
    'items': [
      {'title': 'Đơn Keiki Food', 'points': 120, 'balance_after': 1280, 'occurred_at': '2026-10-20T00:00:00Z'},
      {'title': 'Đổi voucher', 'points': -500, 'balance_after': 1160, 'occurred_at': '2026-10-18T00:00:00Z'},
    ],
  },
};

final pageShots = <String, Widget Function()>{
  'ticket_organizer': () => const TicketOrganizerPage(),
  'ticket_manage_event': () => const ManageEventPage(id: 1),
  'poly_market': () => const PolyMarketPage(id: '1'),
  'wallet_more': () => const WalletMorePage(),
  'admin_home': () => const AdminHomePage(),
  'admin_console_crm': () => const AdminConsolePage(serviceId: 'crm'),
  'admin_op_crm': () => AdminOpPage(serviceId: 'crm', op: adminConsoles['crm']!.firstWhere((o) => o.name == 'crmCampaigns'), title: 'Danh sách chiến dịch'),
};
