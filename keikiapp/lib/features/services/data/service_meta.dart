import 'package:flutter/material.dart';

enum ServiceCategory {
  finance('Tài chính', Icons.account_balance_wallet_outlined, Color(0xFF3949AB)),
  shopping('Mua sắm', Icons.shopping_bag_outlined, Color(0xFFD81B60)),
  food('Ăn uống', Icons.restaurant_outlined, Color(0xFFF4511E)),
  travel('Du lịch & lưu trú', Icons.luggage_outlined, Color(0xFF00897B)),
  entertainment('Giải trí', Icons.celebration_outlined, Color(0xFF8E24AA)),
  social('Cộng đồng & chat', Icons.forum_outlined, Color(0xFF1E88E5)),
  ai('Trợ lý AI', Icons.auto_awesome_outlined, Color(0xFF6D4C41)),
  health('Sức khỏe', Icons.favorite_outline, Color(0xFFE53935)),
  utility('Tiện ích', Icons.widgets_outlined, Color(0xFF43A047)),
  account('Tài khoản & bảo mật', Icons.verified_user_outlined, Color(0xFF546E7A)),
  business('Doanh nghiệp', Icons.business_center_outlined, Color(0xFF5E35B1)),
  system('Hệ thống', Icons.settings_suggest_outlined, Color(0xFF757575));

  const ServiceCategory(this.label, this.icon, this.color);

  final String label;
  final IconData icon;
  final Color color;
}

class ServiceMeta {
  const ServiceMeta(this.title, this.icon, this.category, [this.subtitle = '']);

  final String title;
  final IconData icon;
  final ServiceCategory category;
  final String subtitle;
}

const _c = ServiceCategory.values;

final serviceMeta = <String, ServiceMeta>{
  'account_transaction': ServiceMeta('Giao dịch tài khoản', Icons.swap_horiz, _c[0], 'Sổ cái và đối soát'),
  'address': ServiceMeta('Sổ địa chỉ', Icons.location_on_outlined, _c[9], 'Địa chỉ giao hàng của bạn'),
  'admanagement': ServiceMeta('Quảng cáo', Icons.campaign_outlined, _c[10], 'Hiển thị và đo lường quảng cáo'),
  'ai_agent': ServiceMeta('Trợ lý Keiki AI', Icons.auto_awesome, _c[6], 'Trò chuyện và tạo nội dung'),
  'airflow': ServiceMeta('Luồng dữ liệu', Icons.account_tree_outlined, _c[11]),
  'arrange': ServiceMeta('Sắp xếp lịch', Icons.view_agenda_outlined, _c[11]),
  'authorize': ServiceMeta('Quyền truy cập', Icons.admin_panel_settings_outlined, _c[9], 'Vai trò và quyền của bạn'),
  'automatic_payment': ServiceMeta('Thanh toán tự động', Icons.autorenew, _c[0]),
  'banking': ServiceMeta('Ngân hàng', Icons.account_balance, _c[0]),
  'basket': ServiceMeta('Giỏ hàng', Icons.shopping_cart_outlined, _c[1], 'Giỏ hàng và đơn mua'),
  'billing': ServiceMeta('Gói dịch vụ', Icons.receipt_long_outlined, _c[0], 'Bảng giá và hoá đơn'),
  'bonuslink': ServiceMeta('Bonuslink', Icons.loyalty_outlined, _c[0], 'Điểm thưởng liên kết'),
  'book': ServiceMeta('Đọc sách', Icons.menu_book_outlined, _c[4], 'Sách và chương truyện'),
  'book_store': ServiceMeta('Nhà sách', Icons.local_library_outlined, _c[1], 'Mua sách, tác giả nổi bật'),
  'booking_mini': ServiceMeta('Đặt phòng nhanh', Icons.hotel_outlined, _c[3]),
  'bot': ServiceMeta('Bot', Icons.smart_toy_outlined, _c[11]),
  'calculate': ServiceMeta('Thời tiết & thị trường', Icons.wb_sunny_outlined, _c[8], 'Dự báo, thuỷ triều, radar'),
  'call_center': ServiceMeta('Tổng đài AI', Icons.support_agent, _c[10]),
  'car_rental': ServiceMeta('Thuê xe', Icons.directions_car_outlined, _c[3], 'Tìm và đặt xe'),
  'catalogues': ServiceMeta('Danh mục sản phẩm', Icons.category_outlined, _c[1]),
  'cdn': ServiceMeta('Tệp tin', Icons.cloud_outlined, _c[8], 'Tải lên và chia sẻ tệp'),
  'chat': ServiceMeta('Báo cáo chat', Icons.report_outlined, _c[5]),
  'chatbot': ServiceMeta('Chatbot', Icons.chat_outlined, _c[6]),
  'common': ServiceMeta('Trạng thái hệ thống', Icons.monitor_heart_outlined, _c[11]),
  'coupon': ServiceMeta('Mã giảm giá', Icons.local_offer_outlined, _c[1]),
  'crm': ServiceMeta('CRM', Icons.contacts_outlined, _c[10]),
  'doordash': ServiceMeta('Giao đồ ăn', Icons.delivery_dining_outlined, _c[2]),
  'draw_image': ServiceMeta('Ghép ảnh', Icons.collections_outlined, _c[8]),
  'ejabberd': ServiceMeta('XMPP', Icons.dns_outlined, _c[11]),
  'ekyc': ServiceMeta('Xác thực danh tính', Icons.badge_outlined, _c[9], 'eKYC bằng CCCD và khuôn mặt'),
  'face': ServiceMeta('Nhận diện khuôn mặt', Icons.face_retouching_natural, _c[9]),
  'food': ServiceMeta('Thực đơn', Icons.ramen_dining_outlined, _c[2]),
  'geoservice': ServiceMeta('Bản đồ & địa giới', Icons.map_outlined, _c[8], 'Tỉnh, quận, phường, đường'),
  'hospital': ServiceMeta('Bệnh viện', Icons.local_hospital_outlined, _c[7]),
  'identified': ServiceMeta('Phòng gym', Icons.fitness_center, _c[7], 'Lớp tập và thẻ tập'),
  'integrated_payment': ServiceMeta('Cổng thanh toán', Icons.payments_outlined, _c[0]),
  'kms': ServiceMeta('Quản lý khoá', Icons.key_outlined, _c[11]),
  'livestream': ServiceMeta('Livestream', Icons.live_tv_outlined, _c[4], 'Phát trực tiếp và xem lại'),
  'lottery': ServiceMeta('Xổ số', Icons.confirmation_number_outlined, _c[4], 'Kết quả và dò vé'),
  'lucky': ServiceMeta('Vòng quay may mắn', Icons.casino_outlined, _c[4], 'Chơi, điểm danh, nhận quà'),
  'manage': ServiceMeta('Keycloak', Icons.lock_person_outlined, _c[11]),
  'media': ServiceMeta('Video YouTube', Icons.smart_display_outlined, _c[4]),
  'medical': ServiceMeta('Y tế', Icons.medical_services_outlined, _c[7]),
  'message': ServiceMeta('Hàng đợi tin', Icons.queue_outlined, _c[11]),
  'movie': ServiceMeta('Phim', Icons.movie_outlined, _c[4], 'Phim mới, đề xuất cho bạn'),
  'networking': ServiceMeta('Mạng', Icons.lan_outlined, _c[11]),
  'notification': ServiceMeta('Thông báo', Icons.notifications_outlined, _c[9]),
  'notifyhub': ServiceMeta('Notify Hub', Icons.hub_outlined, _c[11]),
  'ollama_chat': ServiceMeta('Keiki Chat', Icons.chat_bubble_outline, _c[5], 'Bạn bè và nhóm'),
  'ollama_py': ServiceMeta('Mô hình AI', Icons.psychology_outlined, _c[6]),
  'parking_lot': ServiceMeta('Bãi đỗ xe', Icons.local_parking, _c[8], 'Chỗ trống và giá gửi xe'),
  'partner': ServiceMeta('Đối tác', Icons.handshake_outlined, _c[10]),
  'payer': ServiceMeta('Người trả', Icons.person_pin_outlined, _c[11]),
  'payment': ServiceMeta('Thanh toán thẻ', Icons.credit_card, _c[0]),
  'payment_wallet': ServiceMeta('Ví túi tiền', Icons.savings_outlined, _c[0], 'Ví, túi tiền, phương thức'),
  'photo': ServiceMeta('Ghép ảnh nghệ thuật', Icons.photo_library_outlined, _c[8]),
  'point': ServiceMeta('Điểm thưởng', Icons.stars_outlined, _c[0]),
  'polymarket': ServiceMeta('Thị trường dự đoán', Icons.query_stats, _c[4], 'Sự kiện, kèo, bảng xếp hạng'),
  'post': ServiceMeta('Bài viết', Icons.article_outlined, _c[5]),
  'qr': ServiceMeta('Mã QR', Icons.qr_code_2, _c[8], 'Tạo và thống kê mã QR'),
  'recompense': ServiceMeta('Đổi thưởng', Icons.redeem, _c[0]),
  'recruitment': ServiceMeta('Tuyển dụng', Icons.work_outline, _c[10], 'Ứng tuyển việc làm'),
  'referral': ServiceMeta('Giới thiệu bạn bè', Icons.group_add_outlined, _c[5], 'Link mời và phần thưởng'),
  'rent_house': ServiceMeta('Homestay', Icons.cottage_outlined, _c[3], 'Thuê nhà, đặt homestay'),
  'reservation': ServiceMeta('Khách sạn', Icons.apartment_outlined, _c[3], 'Đặt phòng khách sạn'),
  'restaurant_accounting': ServiceMeta('Kế toán nhà hàng', Icons.calculate_outlined, _c[10]),
  'restaurant_consumer': ServiceMeta('Gọi món', Icons.room_service_outlined, _c[2]),
  'restaurant_delivery': ServiceMeta('Giao hàng nhà hàng', Icons.moped_outlined, _c[10]),
  'restaurant_kitchen': ServiceMeta('Bếp & món ăn', Icons.soup_kitchen_outlined, _c[2]),
  'restaurant_order': ServiceMeta('Đặt bàn', Icons.table_restaurant_outlined, _c[2]),
  'reward': ServiceMeta('Phần thưởng', Icons.emoji_events_outlined, _c[0]),
  'room': ServiceMeta('Phòng chat', Icons.meeting_room_outlined, _c[5]),
  'sard_auth': ServiceMeta('Xác thực thanh toán', Icons.verified_outlined, _c[0]),
  'sard_card': ServiceMeta('Thẻ SARD', Icons.style_outlined, _c[0]),
  'sard_customer': ServiceMeta('Khách hàng SARD', Icons.people_outline, _c[10]),
  'search': ServiceMeta('Tìm kiếm', Icons.search, _c[8]),
  'serpapi': ServiceMeta('Tìm kiếm web', Icons.travel_explore, _c[8]),
  'share': ServiceMeta('Hỏi & chia sẻ', Icons.question_answer_outlined, _c[6]),
  'shipping': ServiceMeta('Vận chuyển', Icons.local_shipping_outlined, _c[8], 'Báo giá giao hàng'),
  'shopify': ServiceMeta('Shopify', Icons.storefront_outlined, _c[1]),
  'shortlink': ServiceMeta('Link rút gọn', Icons.link, _c[8]),
  'solana': ServiceMeta('Solana', Icons.currency_bitcoin, _c[0]),
  'threads': ServiceMeta('Threads', Icons.alternate_email, _c[5], 'Bài đăng, theo dõi, bình luận'),
  'ticket': ServiceMeta('Vé sự kiện', Icons.local_activity_outlined, _c[4], 'Sự kiện, ghế, vé của tôi'),
  'toggle': ServiceMeta('Feature flag', Icons.toggle_on_outlined, _c[11]),
  'tool': ServiceMeta('Công cụ', Icons.build_outlined, _c[11]),
  'upload': ServiceMeta('Tải tệp', Icons.upload_file, _c[8]),
  'user': ServiceMeta('Tài khoản', Icons.person_outline, _c[9]),
  'vending': ServiceMeta('Máy bán hàng', Icons.local_drink_outlined, _c[1], 'Tìm máy, mua hàng tự động'),
  'video': ServiceMeta('Video', Icons.ondemand_video, _c[4]),
  'voucher': ServiceMeta('Voucher', Icons.card_giftcard, _c[1], 'Mua, kích hoạt, sử dụng'),
  'wallet_service': ServiceMeta('Ví Keiki', Icons.account_balance_wallet, _c[0], 'Số dư, chuyển tiền, thẻ'),
  'webrtc': ServiceMeta('Gọi video', Icons.video_call_outlined, _c[5]),
  'wishlists': ServiceMeta('Yêu thích', Icons.favorite_border, _c[1]),
  'workflow': ServiceMeta('Quy trình', Icons.schema_outlined, _c[11]),
};

ServiceMeta metaFor(String id) =>
    serviceMeta[id] ??
    ServiceMeta(
      id.split('_').map((p) => p.isEmpty ? p : p[0].toUpperCase() + p.substring(1)).join(' '),
      Icons.extension_outlined,
      ServiceCategory.system,
    );
