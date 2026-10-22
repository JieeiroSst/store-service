import 'fixtures.dart';

final batch3 = <String, Object?>{
  'myWallet': {'wallet_id': 'pw_7f3a91', 'balance': 3250000, 'currency': 'VND', 'status': 'active', 'daily_limit': 50000000, 'per_transaction_limit': 20000000},
  'myWalletPockets': [
    {'pocket_id': 'p1', 'name': 'Du lịch Đà Lạt', 'balance': 4500000, 'currency': 'VND'},
    {'pocket_id': 'p2', 'name': 'Quỹ khẩn cấp', 'balance': 12000000, 'currency': 'VND'},
  ],
  'myWalletTransactions': [
    {'transaction_id': 't1', 'type': 'transfer_in', 'amount': 500000, 'currency': 'VND', 'description': 'Minh Anh chuyển tiền ăn', 'created_at': '2026-10-21T09:00:00Z'},
    {'transaction_id': 't2', 'type': 'transfer_out', 'amount': 120000, 'currency': 'VND', 'description': 'Trả tiền cà phê', 'created_at': '2026-10-20T15:00:00Z'},
    {'transaction_id': 't3', 'type': 'deposit', 'amount': 2000000, 'currency': 'VND', 'description': 'Nạp từ Techcombank', 'created_at': '2026-10-19T08:00:00Z'},
  ],
  'vouchers': [
    {'id': 'vc1', 'code': 'GRAB-50K-8F2A', 'merchant_id': 'Grab', 'status': 'active'},
    {'id': 'vc2', 'code': 'CGV-2D-91KQ', 'merchant_id': 'CGV', 'status': 'pending'},
    {'id': 'vc3', 'code': 'HL-COFFEE-7Q1', 'merchant_id': 'Highlands', 'status': 'used'},
  ],
  'voucherWalletBalance': {'amount': 350000, 'currency': 'VND'},
  'myReferralStats': {'total_invited': 18, 'total_installed': 11, 'total_rewarded': 9, 'total_reward_amt': 450000},
  'referralActiveRewardProgram': {'name': 'Mời 3 bạn nhận 100K', 'status': 'active'},
  'myReferralLinks': {
    'data': [
      {'ref_code': 'QUAN2026', 'share_url': 'https://keiki.vn/r/QUAN2026', 'channel': 'zalo', 'status': 'active'},
      {'ref_code': 'QUANFB', 'share_url': 'https://keiki.vn/r/QUANFB', 'channel': 'facebook', 'status': 'active'},
    ],
  },
  'myRecompensePoints': {'Points': 2480, 'TierName': 'Vàng', 'PointsToNextTier': 520},
  'recompenses': [
    {'ID': '1', 'ReductionDescription': 'Giảm 10% đơn từ 200K', 'ReductionType': 'percentage', 'ReductionValue': 10, 'MinimumPoints': 500, 'CodeDeReduction': 'GOLD10'},
    {'ID': '2', 'ReductionDescription': 'Giảm 50.000 ₫ phí giao hàng', 'ReductionType': 'fixed', 'ReductionValue': 50000, 'MinimumPoints': 300, 'CodeDeReduction': 'SHIP50'},
  ],
  'calcLocations': [
    {'name': 'Vũng Tàu', 'coordinate': {'lat': 10.34, 'lon': 107.08}},
    {'name': 'Nha Trang', 'coordinate': {'lat': 12.24, 'lon': 109.19}},
  ],
  'calcCurrentWeather': {
    'current': {'temperature_c': 29.4, 'precipitation_mm': 0, 'wind_speed_kmh': 14},
    'next_tide': {'time': '2026-10-22T16:20:00Z', 'type': 'high', 'height_m': 3.4},
  },
  'billingPlans': [
    {'plan_id': 1, 'name': 'Keiki Plus', 'description': 'Miễn phí giao hàng, hoàn tiền 2% mọi đơn', 'price': 49000, 'billing_cycle': 'monthly'},
    {'plan_id': 2, 'name': 'Keiki Business', 'description': 'Hoá đơn VAT, quản lý nhân viên, báo cáo chi tiêu', 'price': 990000, 'billing_cycle': 'yearly'},
  ],
  'mySardPaymentAuthentications': [
    {'id': 'c1', 'status': 'pending', 'merchant': 'Apple.com/bill', 'amount': 59000, 'currency': 'VND', 'masked_pan': '•••• 4242', 'attempts_left': 3, 'otp_expires_at': '2026-10-22T10:05:00Z'},
  ],
};

const batch3Shots = <Shot>[
  Shot('payment_wallet', '/services/payment_wallet'),
  Shot('voucher', '/services/voucher'),
  Shot('referral', '/services/referral'),
  Shot('recompense', '/services/recompense'),
  Shot('calculate', '/services/calculate'),
  Shot('billing', '/services/billing'),
  Shot('sard_auth', '/services/sard_auth'),
];
