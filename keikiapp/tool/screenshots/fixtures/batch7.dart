import 'fixtures.dart';

final batch7 = <String, Object?>{
  'mediaSearchVideos': {
    'total': 6,
    'page': 1,
    'size': 40,
    'pages': 1,
    'videos': [
      {'video_id': 101, 'description': 'Hướng dẫn nấu phở bò tại nhà', 'uploaded_on': '2026-10-20T08:00:00Z', 'tag': {'tag_id': 1, 'value': 'nauan'}},
      {'video_id': 102, 'description': 'Review Đà Lạt 3 ngày 2 đêm', 'uploaded_on': '2026-10-18T08:00:00Z', 'tag': {'tag_id': 2, 'value': 'dulich'}},
      {'video_id': 103, 'description': 'Top 10 bàn thắng V-League', 'uploaded_on': '2026-10-15T08:00:00Z', 'tag': {'tag_id': 3, 'value': 'bongda'}},
      {'video_id': 104, 'description': 'Học Flutter trong 30 phút', 'uploaded_on': '2026-10-12T08:00:00Z', 'tag': {'tag_id': 4, 'value': 'congnghe'}},
      {'video_id': 105, 'description': 'Acoustic chill cuối tuần', 'uploaded_on': '2026-10-10T08:00:00Z', 'tag': {'tag_id': 5, 'value': 'amnhac'}},
      {'video_id': 106, 'description': 'Yoga 15 phút buổi sáng', 'uploaded_on': '2026-10-08T08:00:00Z', 'tag': {'tag_id': 6, 'value': 'suckhoe'}},
    ],
  },
  'liveRooms': [
    {'id': 'r1', 'title': 'Săn deal 10.10 cùng Keiki', 'description': 'Giảm tới 50% cho đơn đầu tiên', 'status': 'live'},
    {'id': 'r2', 'title': 'Đêm nhạc acoustic', 'description': 'Hát live theo yêu cầu', 'status': 'live'},
    {'id': 'r3', 'title': 'Bình luận bóng đá', 'description': 'Trước trận Việt Nam – Thái Lan', 'status': 'live'},
    {'id': 'r4', 'title': 'Nấu ăn cùng Bếp Việt', 'description': 'Bún bò Huế chuẩn vị', 'status': 'live'},
  ],
  'bookCategories': [
    {'slug': 'van-hoc', 'name': 'Văn học'},
    {'slug': 'kinh-te', 'name': 'Kinh tế'},
    {'slug': 'ky-nang', 'name': 'Kỹ năng'},
  ],
  'bookBooks': {
    'data': [
      {'id': 1, 'title': 'Dế Mèn phiêu lưu ký', 'author': 'Tô Hoài'},
      {'id': 2, 'title': 'Số đỏ', 'author': 'Vũ Trọng Phụng'},
      {'id': 3, 'title': 'Tắt đèn', 'author': 'Ngô Tất Tố'},
      {'id': 4, 'title': 'Nhà giả kim', 'author': 'Paulo Coelho'},
      {'id': 5, 'title': 'Đắc nhân tâm', 'author': 'Dale Carnegie'},
      {'id': 6, 'title': 'Tuổi trẻ đáng giá bao nhiêu', 'author': 'Rosie Nguyễn'},
    ],
    'pagination': {'limit': 40, 'hasMore': false},
  },
  'myBonusBalances': [
    {'user_id': 7, 'reward_type': 'points', 'total_value': 12500, 'reward_count': 18},
    {'user_id': 7, 'reward_type': 'cashback', 'total_value': 85000, 'reward_count': 4},
  ],
  'myBonusRewards': [
    {'reward_id': 1, 'source': 'Keiki Food', 'reward_type': 'points', 'reward_value': 120, 'created_at': '2026-10-22T12:30:00Z'},
    {'reward_id': 2, 'source': 'Keiki Shop', 'reward_type': 'cashback', 'reward_value': 25000, 'created_at': '2026-10-20T09:10:00Z'},
    {'reward_id': 3, 'source': 'Giới thiệu bạn bè', 'reward_type': 'referral', 'reward_value': 500, 'created_at': '2026-10-18T15:00:00Z'},
    {'reward_id': 4, 'source': 'Vé sự kiện', 'reward_type': 'points', 'reward_value': 300, 'created_at': '2026-10-15T20:00:00Z'},
  ],
  'myIntPayPayments': [
    {'id': 'p1', 'amount': 125000, 'currency': 'VND', 'payment_method': 'momo', 'status': 'completed', 'description': 'Đơn Keiki Food #1024', 'created_at': '2026-10-22T12:30:00Z'},
    {'id': 'p2', 'amount': 450000, 'currency': 'VND', 'payment_method': 'bank_transfer', 'status': 'pending', 'description': 'Vé concert Chân trời rực rỡ', 'created_at': '2026-10-21T10:00:00Z'},
    {'id': 'p3', 'amount': 1290000, 'currency': 'VND', 'payment_method': 'card', 'status': 'completed', 'description': 'Khách sạn Keiki Riverside', 'created_at': '2026-10-19T08:00:00Z'},
    {'id': 'p4', 'amount': 89000, 'currency': 'VND', 'payment_method': 'wallet', 'status': 'refunded', 'description': 'Gói Keiki Premium tháng', 'created_at': '2026-10-15T08:00:00Z'},
  ],
  'rewards': {
    'records': [
      {'id': 1, 'name': 'Voucher Food 50K', 'description': 'Áp dụng mọi nhà hàng', 'points': 500},
      {'id': 2, 'name': 'Vé xem phim 2D', 'description': 'Tại rạp đối tác', 'points': 900},
      {'id': 3, 'name': 'Bình giữ nhiệt Keiki', 'description': 'Phiên bản giới hạn', 'points': 1500},
      {'id': 4, 'name': 'Thẻ quà 200K', 'description': 'Dùng cho Keiki Shop', 'points': 2000},
    ],
  },
  'chatbotConversations': {
    'count': 3,
    'conversations': [
      {'id': 11, 'is_ai_chat': true, 'active_ai_model': 'Llama 3.1', 'created_at': '2026-10-22T09:00:00Z', 'updated_at': '2026-10-23T21:15:00Z'},
      {'id': 12, 'is_ai_chat': true, 'active_ai_model': 'Qwen 2.5', 'created_at': '2026-10-20T09:00:00Z', 'updated_at': '2026-10-21T08:40:00Z'},
      {'id': 13, 'is_ai_chat': false, 'created_at': '2026-10-18T09:00:00Z', 'updated_at': '2026-10-18T10:00:00Z'},
    ],
  },
  'cdnFiles': {
    'files': [
      {'id': 'f1', 'file_name': 'banner-10-10.png', 'content_type': 'image/png', 'size_bytes': 482133, 'status': 'ready', 'created_at': '2026-10-22T08:00:00Z'},
      {'id': 'f2', 'file_name': 'menu-bep-viet.pdf', 'content_type': 'application/pdf', 'size_bytes': 1288490, 'status': 'ready', 'created_at': '2026-10-21T08:00:00Z'},
      {'id': 'f3', 'file_name': 'intro.mp4', 'content_type': 'video/mp4', 'size_bytes': 18874368, 'status': 'pending', 'created_at': '2026-10-20T08:00:00Z'},
    ],
  },
  'myROrderReservations': [
    {'id': 1, 'table_name': 'Ban công', 'customer_name': 'Lưu Mạnh Quân', 'party_size': 4, 'reserved_at': '2026-10-26T12:00:00Z', 'status': 'booked'},
    {'id': 2, 'table_name': 'Phòng riêng', 'customer_name': 'Lưu Mạnh Quân', 'party_size': 8, 'reserved_at': '2026-10-18T12:00:00Z', 'status': 'completed'},
  ],
};

const batch7Shots = <Shot>[
  Shot('media', '/services/media'),
  Shot('livestream', '/services/livestream'),
  Shot('book', '/services/book'),
  Shot('bonuslink', '/services/bonuslink'),
  Shot('integrated_payment', '/services/integrated_payment'),
  Shot('payment', '/services/payment'),
  Shot('reward', '/services/reward'),
  Shot('share', '/services/share'),
  Shot('photo', '/services/photo'),
  Shot('recruitment', '/services/recruitment'),
  Shot('chatbot', '/services/chatbot'),
  Shot('draw_image', '/services/draw_image'),
  Shot('face', '/services/face'),
  Shot('search', '/services/search'),
  Shot('serpapi', '/services/serpapi'),
  Shot('shortlink', '/services/shortlink'),
  Shot('cdn', '/services/cdn'),
  Shot('shipping', '/services/shipping'),
  Shot('restaurant_kitchen', '/services/restaurant_kitchen'),
  Shot('restaurant_order', '/services/restaurant_order'),
];
