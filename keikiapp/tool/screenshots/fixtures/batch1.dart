import 'fixtures.dart';

Map<String, Object?> _rec(String id, String title, {String reason = 'trending', double? progress}) => {
      'video': {'id': id, 'title': title, 'description': 'Một bộ phim đầy cảm xúc về tình bạn, gia đình và những chuyến đi.', 'tags': ['Tâm lý', 'Gia đình'], 'duration': 6240, 'views': 128400, 'created_at': '2026-09-01T00:00:00Z'},
      'score': 0.9,
      'reason': reason,
      'progress': progress,
    };

final _movies = {
  'items': [
    _rec('m1', 'Mắt Biếc'),
    _rec('m2', 'Hai Phượng', reason: 'new'),
    _rec('m3', 'Bố Già'),
    _rec('m4', 'Lật Mặt 7'),
  ],
  'total': 4,
};

final _event = {
  'id': 1,
  'title': 'Đêm nhạc Hà Anh Tuấn – Chân trời rực rỡ',
  'description': 'Đêm nhạc đặc biệt với dàn nhạc giao hưởng 40 người và những bản tình ca quen thuộc.',
  'category': 'Âm nhạc',
  'city': 'TP. Hồ Chí Minh',
  'venue': 'Nhà hát Hòa Bình',
  'address': '240 Đường 3/2, Quận 10',
  'starts_at': '2026-11-15T12:30:00Z',
  'ends_at': '2026-11-15T15:30:00Z',
  'featured': true,
  'min_price': 800000,
  'sold_out': false,
  'refund_cutoff_hours': 48,
  'transferable': true,
  'ticket_types': [
    {'id': 11, 'name': 'Standard', 'price': 800000, 'available': 120, 'on_sale': true, 'sold_out': false, 'max_per_order': 6},
    {'id': 12, 'name': 'VIP', 'price': 1800000, 'available': 24, 'on_sale': true, 'sold_out': false, 'max_per_order': 4, 'description': 'Hàng ghế đầu, quà lưu niệm'},
    {'id': 13, 'name': 'SVIP', 'price': 3500000, 'available': 0, 'on_sale': true, 'sold_out': true},
  ],
};

final batch1 = <String, Object?>{
  'movieTrending': _movies,
  'movieNewReleases': _movies,
  'myMovieForUser': _movies,
  'myMovieContinueWatching': {'items': [_rec('m1', 'Mắt Biếc', progress: 0.42), _rec('m3', 'Bố Già', progress: 0.8)]},
  'ticketCategories': ['Âm nhạc', 'Hội thảo', 'Thể thao', 'Sân khấu'],
  'ticketEvents': [
    _event,
    {..._event, 'id': 2, 'title': 'Vietnam Tech Summit 2026', 'category': 'Hội thảo', 'min_price': 0, 'city': 'Hà Nội'},
    {..._event, 'id': 3, 'title': 'V-League: Hà Nội FC vs HAGL', 'category': 'Thể thao', 'min_price': 150000},
    {..._event, 'id': 4, 'title': 'Kịch: Ngôi nhà hạnh phúc', 'category': 'Sân khấu', 'sold_out': true},
  ],
  'ticketEvent': _event,
  'luckyGames': [
    {'id': 1, 'game_type': 'wheel', 'name': 'Vòng quay Keiki', 'description': 'Quay mỗi ngày, trúng tới 10.000 điểm', 'daily_spins': 3, 'prizes': List.filled(8, {'name': 'x'})},
    {'id': 2, 'game_type': 'box', 'name': 'Mở hộp quà', 'description': 'Chọn 1 trong 6 hộp quà bí ẩn', 'daily_spins': 1, 'prizes': List.filled(5, {'name': 'x'})},
    {'id': 3, 'game_type': 'scratch', 'name': 'Cào thẻ cuối tuần', 'description': 'Chỉ mở vào thứ 7 & Chủ nhật', 'daily_spins': 2, 'prizes': List.filled(4, {'name': 'x'})},
  ],
  'luckyGame': {
    'id': 1,
    'game_type': 'wheel',
    'name': 'Vòng quay Keiki',
    'description': 'Quay mỗi ngày, trúng tới 10.000 điểm',
    'daily_spins': 3,
    'prizes': [
      {'name': '10.000 điểm', 'type': 'points', 'points': 10000, 'quantity': 5, 'awarded': 1},
      {'name': 'Voucher Food 50K', 'type': 'voucher', 'quantity': 100, 'awarded': 37},
      {'name': '100 điểm', 'type': 'points', 'points': 100},
    ],
  },
  'luckyMyStatus': {'remaining': 2, 'daily_spins': 3, 'usage': {'credit_balance': 1}},
  'luckyWinners': [
    {'user': 'Minh A***', 'prize_name': 'Voucher Food 50K', 'at': '2026-10-21T10:00:00Z'},
    {'user': 'Thu H***', 'prize_name': '10.000 điểm', 'at': '2026-10-20T08:00:00Z'},
  ],
  'luckyCheckinStatus': {'checked_in_today': false, 'streak': 4, 'cycle_day': 4, 'next_reward': 50, 'rewards': [10, 20, 30, 40, 50, 80, 200]},
  'lotteryResults': {
    'date': '2026-10-21',
    'province': 'Miền Bắc',
    'prizes': {
      'DB': ['58214'],
      'G1': ['20931'],
      'G2': ['77120', '04512'],
      'G3': ['33021', '91822', '11873', '54210', '39188', '70014'],
      'G4': ['3412', '9821', '0071', '5520'],
      'G5': ['2289', '8812', '0412', '7371', '4428', '1190'],
      'G6': ['214', '880', '561'],
      'G7': ['12', '87', '45', '09'],
    },
    'source': 'xoso.com.vn',
  },
  'polyCategories': [{'category': 'Thể thao', 'events': 12}, {'category': 'Kinh tế', 'events': 8}, {'category': 'Công nghệ', 'events': 5}],
  'polyEvents': {
    'items': [
      {
        'id': 1,
        'title': 'Việt Nam vô địch AFF Cup 2026?',
        'category': 'Thể thao',
        'volume': 182000000,
        'end_date': '2026-12-30T00:00:00Z',
        'markets': [{'id': 1, 'question': 'Có', 'yes_price': 62, 'share_value': 100, 'status': 'open'}],
      },
      {
        'id': 2,
        'title': 'Giá vàng SJC vượt 100 triệu/lượng trước 2027?',
        'category': 'Kinh tế',
        'volume': 54000000,
        'end_date': '2026-12-31T00:00:00Z',
        'markets': [
          {'id': 2, 'group_item_title': 'Trước 1/12', 'yes_price': 23, 'share_value': 100},
          {'id': 3, 'group_item_title': 'Trước 31/12', 'yes_price': 41, 'share_value': 100},
        ],
      },
    ],
  },
  'bookStoreBooks': [
    {'id': 1, 'name': 'Đắc Nhân Tâm', 'price': 86000, 'quantity': 40, 'author': {'name': 'Dale Carnegie'}, 'category': {'name': 'Kỹ năng'}},
    {'id': 2, 'name': 'Nhà Giả Kim', 'price': 79000, 'quantity': 12, 'author': {'name': 'Paulo Coelho'}, 'category': {'name': 'Tiểu thuyết'}},
    {'id': 3, 'name': 'Tuổi Trẻ Đáng Giá Bao Nhiêu', 'price': 90000, 'quantity': 0, 'author': {'name': 'Rosie Nguyễn'}, 'category': {'name': 'Kỹ năng'}},
  ],
  'bookStoreMostPurchasedBook': {'id': 1, 'name': 'Đắc Nhân Tâm', 'purchase_count': 1280},
  'bookStoreMostReadAuthor': {'author_name': 'Nguyễn Nhật Ánh'},
  'videos': {
    'items': [
      for (final (i, t) in ['Review iPhone 18 Pro', 'Phở Hà Nội ngon nhất', 'Học Flutter trong 10 phút', 'Đà Lạt mùa hoa dã quỳ'].indexed)
        {'id': 'v$i', 'title': t, 'views': 12000 * (i + 1), 'duration': 420 + i * 60, 'created_at': '2026-10-1${i}T00:00:00Z'},
    ],
  },
};

const batch1Shots = <Shot>[
  Shot('movie', '/services/movie'),
  Shot('movie_continue', '/services/movie', tab: 'Đang xem'),
  Shot('ticket', '/services/ticket'),
  Shot('ticket_event', '/services/ticket', tap: 'Đêm nhạc Hà Anh Tuấn – Chân trời rực rỡ'),
  Shot('lucky', '/services/lucky'),
  Shot('lucky_game', '/services/lucky', tap: 'Vòng quay Keiki'),
  Shot('lucky_checkin', '/services/lucky', tab: 'Điểm danh'),
  Shot('lottery', '/services/lottery'),
  Shot('polymarket', '/services/polymarket'),
  Shot('book_store', '/services/book_store'),
  Shot('video', '/services/video'),
];
