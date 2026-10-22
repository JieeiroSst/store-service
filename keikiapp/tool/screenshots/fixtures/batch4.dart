import 'fixtures.dart';

Map<String, Object?> _thread(String id, String user, String content, {List<String> tags = const [], int likes = 0, int comments = 0}) => {
      'id': id,
      'user_id': '9$id',
      'content': content,
      'like_count': likes,
      'comment_count': comments,
      'repost_count': likes ~/ 7,
      'created_at': DateTime.now().subtract(Duration(hours: int.parse(id) * 3)).toUtc().toIso8601String(),
      'author': {'id': '9$id', 'username': user},
      'tags': tags,
    };

final batch4 = <String, Object?>{
  'threadsHomeFeed': {
    'rows': [
      _thread('1', 'minhanh', 'Sáng nay thử quán phở mới ở Lò Đúc, nước dùng trong mà đậm, 10 điểm không có nhưng! 🍜', tags: ['amthuc', 'hanoi'], likes: 128, comments: 24),
      _thread('2', 'tuan.dev', 'Vừa ship xong tính năng thanh toán bằng Ví Keiki cho app của team. Ai cần review code không?', tags: ['flutter'], likes: 56, comments: 9),
      _thread('3', 'linh.travel', 'Đà Lạt mùa này mây đẹp quá, homestay view đồi thông chỉ 1tr2/đêm.', tags: ['dulich', 'dalat'], likes: 342, comments: 51),
    ],
  },
  'threadsTrendingTags': [
    {'name': 'dulich', 'count': 1820},
    {'name': 'amthuc', 'count': 1204},
    {'name': 'flutter', 'count': 380},
  ],
  'posts': {
    'rows': [
      {'id': 'p1', 'name': '5 cách tiết kiệm khi đặt đồ ăn online', 'description': 'Mẹo dùng voucher, điểm thưởng và combo để bữa trưa rẻ hơn 30%.', 'created_at': '2026-10-20T00:00:00Z', 'categories': [{'name': 'Mẹo hay'}]},
      {'id': 'p2', 'name': 'Hướng dẫn kích hoạt Ví Keiki trong 2 phút', 'description': 'Từ eKYC đến đặt mã PIN, tất cả ngay trong app.', 'created_at': '2026-10-18T00:00:00Z', 'categories': [{'name': 'Hướng dẫn'}]},
    ],
  },
  'rooms': [
    {'id': 1, 'name': 'Team Keiki Mobile', 'owner': 'quan', 'member_count': 8},
    {'id': 2, 'name': 'Hội đi phượt cuối tuần', 'owner': 'linh', 'member_count': 23},
  ],
  'myChatReports': [
    {'id': 1, 'report_type': 'scam', 'notes': 'Tài khoản lạ nhắn tin yêu cầu chuyển tiền đặt cọc', 'status': 'resolved', 'created_at': '2026-10-12T00:00:00Z'},
    {'id': 2, 'report_type': 'spam', 'notes': 'Gửi link quảng cáo liên tục trong nhóm', 'status': 'pending', 'created_at': '2026-10-21T00:00:00Z'},
  ],
  'webrtcIceServers': {
    'ice_servers': [
      {'urls': ['stun:stun.keiki.vn:3478']},
      {'urls': ['turn:turn.keiki.vn:3478?transport=udp', 'turns:turn.keiki.vn:5349']},
    ],
  },
  'adServe': {
    'impression_id': 77,
    'ad': {'id': 5, 'campaign_id': 12, 'title': 'Giảm 50% đơn Keiki Food đầu tiên', 'description': 'Áp dụng cho khách hàng mới, tối đa 50.000 ₫. Hạn dùng đến 31/12.', 'ad_type': 'banner', 'target_url': 'keiki://food', 'start_date': '2026-10-01T00:00:00Z', 'end_date': '2026-12-31T00:00:00Z'},
  },
};

const batch4Shots = <Shot>[
  Shot('threads', '/services/threads'),
  Shot('threads_trending', '/services/threads', tab: 'Xu hướng'),
  Shot('post', '/services/post'),
  Shot('room', '/services/room'),
  Shot('chat_report', '/services/chat'),
  Shot('webrtc', '/services/webrtc'),
  Shot('admanagement', '/services/admanagement'),
  Shot('manage', '/services/manage'),
];
