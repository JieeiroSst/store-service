import 'package:flutter/material.dart';

import 'schema_catalog.dart';

const _servicePrefixes = <String, List<String>>{
  'wallet_service': ['walletSvc'],
  'payment_wallet': ['wallet'],
  'rent_house': ['rent'],
  'reservation': ['resv'],
  'polymarket': ['poly'],
  'restaurant_kitchen': ['rKitchen'],
  'restaurant_order': ['rOrder'],
  'restaurant_consumer': ['rConsumer'],
  'restaurant_delivery': ['rDelivery'],
  'restaurant_accounting': ['rAccounting'],
  'integrated_payment': ['intPay'],
  'calculate': ['calc'],
  'vending': ['vend'],
  'parking_lot': ['parking'],
  'ollama_chat': ['ollamaChat'],
  'ollama_py': ['ollamaPy'],
  'ai_agent': ['aiAgent'],
  'car_rental': ['carRental'],
  'book_store': ['bookStore'],
  'booking_mini': ['bookingMini'],
  'draw_image': ['drawImage'],
  'notification': ['notif'],
  'recruitment': ['recruit'],
  'shipping': ['ship'],
  'serpapi': ['serp'],
  'livestream': ['live'],
  'manage': ['kc'],
  'catalogues': ['catalogue'],
  'wishlists': ['wishlist'],
  'bonuslink': ['bonus'],
  'geoservice': ['geo'],
  'identified': ['gym'],
  'sard_auth': ['sard'],
  'account_transaction': ['acctTx', 'accountTransaction'],
};

const _words = <String, String>{
  'my': 'của tôi',
  'create': 'Tạo',
  'add': 'Thêm',
  'update': 'Cập nhật',
  'edit': 'Sửa',
  'delete': 'Xoá',
  'remove': 'Gỡ',
  'cancel': 'Huỷ',
  'get': 'Xem',
  'list': 'Danh sách',
  'search': 'Tìm',
  'find': 'Tìm',
  'check': 'Kiểm tra',
  'verify': 'Xác minh',
  'validate': 'Kiểm tra',
  'submit': 'Gửi',
  'send': 'Gửi',
  'pay': 'Thanh toán',
  'buy': 'Mua',
  'book': 'Đặt',
  'reserve': 'Giữ chỗ',
  'register': 'Đăng ký',
  'join': 'Tham gia',
  'leave': 'Rời',
  'follow': 'Theo dõi',
  'unfollow': 'Bỏ theo dõi',
  'like': 'Thích',
  'unlike': 'Bỏ thích',
  'upload': 'Tải lên',
  'download': 'Tải xuống',
  'redeem': 'Đổi',
  'claim': 'Nhận',
  'apply': 'Áp dụng',
  'play': 'Chơi',
  'start': 'Bắt đầu',
  'end': 'Kết thúc',
  'stop': 'Dừng',
  'confirm': 'Xác nhận',
  'approve': 'Duyệt',
  'reject': 'Từ chối',
  'decline': 'Từ chối',
  'transfer': 'Chuyển',
  'withdraw': 'Rút',
  'top': 'Nạp',
  'checkin': 'Điểm danh',
  'login': 'Đăng nhập',
  'logout': 'Đăng xuất',
  'history': 'Lịch sử',
  'status': 'Trạng thái',
  'stats': 'Thống kê',
  'balance': 'Số dư',
  'balances': 'Số dư',
  'orders': 'Đơn hàng',
  'order': 'Đơn hàng',
  'events': 'Sự kiện',
  'event': 'Sự kiện',
  'tickets': 'Vé',
  'ticket': 'Vé',
  'bookings': 'Lượt đặt',
  'booking': 'Lượt đặt',
  'rooms': 'Phòng',
  'room': 'Phòng',
  'hotels': 'Khách sạn',
  'hotel': 'Khách sạn',
  'reviews': 'Đánh giá',
  'review': 'Đánh giá',
  'comments': 'Bình luận',
  'comment': 'Bình luận',
  'posts': 'Bài viết',
  'post': 'Bài viết',
  'users': 'Người dùng',
  'user': 'Người dùng',
  'products': 'Sản phẩm',
  'product': 'Sản phẩm',
  'categories': 'Danh mục',
  'category': 'Danh mục',
  'transactions': 'Giao dịch',
  'transaction': 'Giao dịch',
  'payments': 'Thanh toán',
  'payment': 'Thanh toán',
  'wallets': 'Ví',
  'wallet': 'Ví',
  'cards': 'Thẻ',
  'card': 'Thẻ',
  'results': 'Kết quả',
  'winners': 'Người thắng',
  'games': 'Trò chơi',
  'game': 'Trò chơi',
  'movies': 'Phim',
  'movie': 'Phim',
  'vehicles': 'Xe',
  'vehicle': 'Xe',
  'reservations': 'Đặt chỗ',
  'reservation': 'Đặt chỗ',
  'rewards': 'Phần thưởng',
  'reward': 'Phần thưởng',
  'points': 'Điểm',
  'new': 'Mới',
  'releases': 'phát hành',
  'trending': 'Thịnh hành',
  'available': 'Còn trống',
  'and': 'và',
  'by': 'theo',
  'for': 'cho',
  'seats': 'ghế',
  'seat': 'ghế',
  'sessions': 'suất',
  'session': 'phiên',
  'venue': 'địa điểm',
  'templates': 'mẫu',
  'template': 'mẫu',
  'resale': 'bán lại',
  'similar': 'tương tự',
  'home': 'trang chủ',
  'record': 'ghi nhận',
  'from': 'từ',
  'types': 'loại',
  'inventory': 'tồn kho',
  'quotes': 'báo giá',
  'services': 'dịch vụ',
  'queue': 'hàng chờ',
  'availability': 'lịch trống',
  'amenities': 'tiện nghi',
  'rates': 'giá',
  'plays': 'lượt chơi',
  'followers': 'người theo dõi',
  'following': 'đang theo dõi',
  'bookmarks': 'đã lưu',
  'tags': 'thẻ',
  'rental': 'lượt thuê',
  'rentals': 'lượt thuê',
  'pockets': 'túi tiền',
  'pocket': 'túi tiền',
  'methods': 'phương thức',
  'link': 'liên kết',
  'links': 'liên kết',
  'share': 'chia sẻ',
  'targets': 'kênh',
  'active': 'đang chạy',
  'program': 'chương trình',
  'referral': 'giới thiệu',
  'attribution': 'nguồn',
  'report': 'báo cáo',
  'generate': 'tạo',
  'machines': 'máy',
  'machine': 'máy',
  'series': 'chuỗi',
  'homestays': 'homestay',
  'to': 'tới',
  'with': 'với',
  'of': 'của',
  'details': 'chi tiết',
  'detail': 'chi tiết',
  'info': 'thông tin',
  'members': 'thành viên',
  'member': 'thành viên',
  'groups': 'nhóm',
  'group': 'nhóm',
  'messages': 'tin nhắn',
  'contacts': 'danh bạ',
  'files': 'tệp',
  'file': 'tệp',
  'image': 'ảnh',
  'images': 'ảnh',
  'videos': 'video',
  'related': 'liên quan',
  'view': 'lượt xem',
  'profile': 'hồ sơ',
  'settings': 'cài đặt',
  'notifications': 'thông báo',
  'devices': 'thiết bị',
  'device': 'thiết bị',
  'contact': 'liên hệ',
  'locations': 'địa điểm',
  'location': 'vị trí',
  'provinces': 'tỉnh thành',
  'districts': 'quận huyện',
  'wards': 'phường xã',
  'streets': 'đường',
  'markets': 'thị trường',
  'market': 'thị trường',
  'leaderboard': 'bảng xếp hạng',
  'trades': 'giao dịch',
  'summary': 'tổng quan',
  'spots': 'chỗ',
  'limits': 'hạn mức',
  'ledger': 'sổ cái',
  'statement': 'sao kê',
  'requests': 'yêu cầu',
  'request': 'yêu cầu',
  'code': 'mã',
  'codes': 'mã',
  'roles': 'vai trò',
  'permissions': 'quyền',
  'permission': 'quyền',
  'engines': 'công cụ',
  'archive': 'lưu trữ',
  'upcoming': 'sắp diễn ra',
  'popular': 'phổ biến',
  'default': 'mặc định',
  'shared': 'được chia sẻ',
  'lines': 'dòng',
  'line': 'dòng',
  'classes': 'lớp',
  'pass': 'thẻ',
  'recordings': 'bản ghi',
  'playback': 'phát lại',
  'count': 'số lượng',
  'viewer': 'người xem',
  'stream': 'luồng',
  'ingest': 'nhận luồng',
  'winner': 'người thắng',
  'plan': 'gói',
  'plans': 'gói',
  'usages': 'lượt dùng',
  'reports': 'báo cáo',
  'conversations': 'cuộc trò chuyện',
  'conversation': 'cuộc trò chuyện',
  'models': 'mô hình',
  'chat': 'trò chuyện',
  'ask': 'hỏi',
  'reply': 'trả lời',
  'feed': 'bảng tin',
  'map': 'sơ đồ',
  'regenerate': 'tạo lại',
  'resend': 'gửi lại',
  'mark': 'đánh dấu',
  'set': 'đặt',
};

const _phrases = <String, String>{
  'check in': 'Điểm danh',
  'top up': 'Nạp tiền',
  'sign up': 'Đăng ký',
  'sign in': 'Đăng nhập',
  'log in': 'Đăng nhập',
  'log out': 'Đăng xuất',
  'new releases': 'Phim mới',
  'continue watching': 'Đang xem dở',
  'payment request': 'Yêu cầu thanh toán',
  'payment requests': 'Yêu cầu thanh toán',
  'bank account': 'Tài khoản ngân hàng',
  'bank accounts': 'Tài khoản ngân hàng',
  'home feed': 'Bảng tin',
};

String _stripPrefix(String service, String name) {
  for (final p in _servicePrefixes[service] ?? const <String>[]) {
    if (name.startsWith(p) && name.length > p.length && name[p.length].toUpperCase() == name[p.length]) {
      return name.substring(p.length);
    }
  }
  final single = service.split('_').first;
  if (name.toLowerCase().startsWith(single) && name.length > single.length && name[single.length].toUpperCase() == name[single.length]) {
    return name.substring(single.length);
  }
  return name;
}

List<String> splitWords(String s) => s
    .replaceAllMapped(RegExp(r'([a-z0-9])([A-Z])'), (m) => '${m[1]} ${m[2]}')
    .replaceAllMapped(RegExp(r'([A-Z]+)([A-Z][a-z])'), (m) => '${m[1]} ${m[2]}')
    .split(RegExp(r'[\s_\-]+'))
    .where((w) => w.isNotEmpty)
    .toList();

const _verbs = {
  'create', 'add', 'update', 'edit', 'delete', 'remove', 'cancel', 'get', 'list', 'search', 'find', 'check', 'verify',
  'validate', 'submit', 'send', 'pay', 'buy', 'book', 'reserve', 'register', 'join', 'leave', 'follow', 'unfollow', 'like',
  'unlike', 'upload', 'download', 'redeem', 'claim', 'apply', 'play', 'start', 'end', 'stop', 'confirm', 'approve',
  'reject', 'decline', 'transfer', 'withdraw', 'generate', 'share', 'record', 'set', 'regenerate', 'resend', 'mark', 'ask',
};

const _prepositions = {'from', 'to', 'for', 'by', 'with', 'of', 'in', 'on'};

String _translate(String word) {
  final t = _words[word.toLowerCase()];
  if (t != null) return t.toLowerCase();
  return word == word.toUpperCase() ? word : word.toLowerCase();
}

String operationTitle(GqlOperation op) {
  final words = splitWords(_stripPrefix(op.service, op.name));
  if (words.isEmpty) return op.name;
  final mine = words.any((w) => w.toLowerCase() == 'my');
  final rest = words.where((w) => w.toLowerCase() != 'my').toList();

  final tokens = <String>[];
  final raw = <String>[];
  for (var i = 0; i < rest.length; i++) {
    final pair = i + 1 < rest.length ? '${rest[i]} ${rest[i + 1]}'.toLowerCase() : null;
    if (pair != null && _phrases.containsKey(pair)) {
      tokens.add(_phrases[pair]!.toLowerCase());
      raw.add(pair);
      i++;
    } else {
      tokens.add(_translate(rest[i]));
      raw.add(rest[i].toLowerCase());
    }
  }

  final out = <String>[];
  var start = 0;
  if (raw.isNotEmpty && _verbs.contains(raw.first)) {
    out.add(tokens.first);
    start = 1;
  } else if (!op.isMutation && (op.type.isList || RegExp(r'(Page|List|Connection|Results)$').hasMatch(op.type.base))) {
    out.add('danh sách');
  } else if (!op.isMutation && op.args.any((a) => a.type.nonNull && (a.name == 'id' || a.name.endsWith('Id') || a.name.endsWith('_id')))) {
    out.add('chi tiết');
  }
  var segment = <String>[];
  String? prep;
  void flush() {
    if (prep != null) out.add(prep!);
    out.addAll(segment.reversed);
    segment = [];
    prep = null;
  }

  for (var i = start; i < tokens.length; i++) {
    if (_prepositions.contains(raw[i])) {
      flush();
      prep = tokens[i];
    } else {
      segment.add(tokens[i]);
    }
  }
  flush();
  if (mine) out.add('của tôi');
  final title = out.join(' ');
  if (title.isEmpty) return op.name;
  return title[0].toUpperCase() + title.substring(1);
}

String fieldLabel(String name) {
  final words = splitWords(name);
  if (words.isEmpty) return name;
  const known = {
    'id': 'ID',
    'url': 'URL',
    'uuid': 'UUID',
    'qr': 'QR',
    'otp': 'OTP',
    'pin': 'PIN',
  };
  final out = [for (final w in words) known[w.toLowerCase()] ?? w.toLowerCase()];
  final s = out.join(' ');
  return s[0].toUpperCase() + s.substring(1);
}

IconData operationIcon(GqlOperation op) {
  final first = splitWords(_stripPrefix(op.service, op.name)).map((w) => w.toLowerCase()).toList();
  final verb = first.isEmpty ? '' : (first.first == 'my' && first.length > 1 ? first[1] : first.first);
  switch (verb) {
    case 'create':
    case 'add':
    case 'register':
      return Icons.add_circle_outline;
    case 'update':
    case 'edit':
    case 'set':
      return Icons.edit_outlined;
    case 'delete':
    case 'remove':
    case 'cancel':
      return Icons.delete_outline;
    case 'search':
    case 'find':
      return Icons.search;
    case 'pay':
    case 'buy':
    case 'transfer':
    case 'withdraw':
    case 'top':
      return Icons.payments_outlined;
    case 'upload':
      return Icons.upload_outlined;
    case 'check':
    case 'verify':
    case 'validate':
      return Icons.fact_check_outlined;
    case 'history':
      return Icons.history;
    case 'stats':
      return Icons.insights_outlined;
  }
  return op.isMutation ? Icons.bolt_outlined : (op.type.isList ? Icons.list_alt : Icons.visibility_outlined);
}
