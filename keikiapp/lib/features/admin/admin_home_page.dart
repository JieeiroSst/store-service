import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:go_router/go_router.dart';

import '../../app/role.dart';
import '../../app/theme.dart';
import '../apps/commerce/commerce_more.dart';
import '../apps/entertainment/polymarket_more.dart';
import '../apps/entertainment/ticket_organizer.dart';
import '../apps/extras/extras.dart';
import '../apps/kit/kit.dart';
import '../apps/travel/stays_host.dart';
import '../services/bloc/services_cubit.dart';
import '../services/data/schema_catalog.dart';
import '../services/data/service_meta.dart';
import 'console/admin_console_page.dart';
import '../tracking/link_tracking_page.dart';

class _Console {
  const _Console(this.title, this.subtitle, this.icon, this.color, this.open);

  final String title;
  final String subtitle;
  final IconData icon;
  final Color color;
  final void Function(BuildContext context) open;
}

void _push(BuildContext context, Color color, Widget page) =>
    Navigator.of(context).push(MaterialPageRoute<void>(builder: (context) => appScope(context, color, page)));

final _consoles = <_Console>[
  _Console('Ban tổ chức sự kiện', 'Sự kiện, vé, địa điểm, soát vé', Icons.local_activity_outlined, const Color(0xFFC026D3), (c) => _push(c, const Color(0xFFC026D3), const TicketOrganizerPage())),
  _Console('Khách sạn', 'Phòng, lịch giá, khuyến mãi, báo cáo', Icons.hotel_outlined, const Color(0xFF1D4ED8), (c) => _push(c, const Color(0xFF1D4ED8), const HotelHostPage())),
  _Console('Homestay', 'Chỗ ở, giá theo kỳ, hợp đồng', Icons.cottage_outlined, const Color(0xFF0D9488), (c) => _push(c, const Color(0xFF0D9488), const HomestayHostPage())),
  _Console('Thị trường dự đoán', 'Tạo sự kiện, cấp vốn sàn', Icons.add_chart, const Color(0xFF2563EB), (c) => _push(c, const Color(0xFF2563EB), const PolyCreatorPage())),
  _Console('Voucher doanh nghiệp', 'Phát hành, đối tác, ngân sách', Icons.business_outlined, const Color(0xFFE11D48), (c) => _push(c, const Color(0xFFE11D48), const VoucherBusinessPage())),
  _Console('Quầy giao xe', 'Giao và nhận lại xe thuê', Icons.car_rental, const Color(0xFF0369A1), (c) => _push(c, const Color(0xFF0369A1), const CarStaffPage())),
  _Console('Đối tác Bonuslink', 'Ghi nhận và tra cứu thưởng', Icons.redeem_outlined, const Color(0xFFEA580C), (c) => _push(c, const Color(0xFFEA580C), const BonusPartnerPage())),
  _Console('Chuyên mục bài viết', 'Tạo, sửa, xoá chuyên mục', Icons.folder_special_outlined, const Color(0xFF9333EA), (c) => _push(c, const Color(0xFF9333EA), const PostCategoryManager())),
  _Console('Chuyến giao tài xế', 'Đơn được giao cho tài xế', Icons.delivery_dining_outlined, const Color(0xFFEA580C), (c) => _push(c, const Color(0xFFEA580C), const DriverJobsPage())),
  _Console('Deep link & đo lường', 'Ghi nhận cài đặt, AASA, asset links', Icons.analytics_outlined, const Color(0xFF0891B2), (c) => _push(c, const Color(0xFF0891B2), const ShortlinkToolsPage())),
  _Console('Kiểm tra link & tracking', 'Branch, mã mời, nguồn cài đặt', Icons.alt_route, const Color(0xFF0E7490), (c) => _push(c, const Color(0xFF0E7490), const LinkTrackingPage())),
  _Console('Đăng nhập SSO', 'Keycloak cho cổng đối tác', Icons.key_outlined, const Color(0xFF334155), (c) => c.push('/services/manage')),
];

class AdminHomePage extends StatelessWidget {
  const AdminHomePage({super.key});

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final services = context.select<ServicesCubit, List<(GqlService, int)>>((c) {
      final s = c.state;
      return [
        for (final svc in s.visibleServices)
          if (adminConsoleSize(svc.id) > 0 && s.visibleOperations(svc).any((o) => o.access == OperationAccess.admin)) (svc, s.visibleOperations(svc).where((o) => o.access == OperationAccess.admin).length),
      ];
    });
    return Scaffold(
      body: CustomScrollView(
        slivers: [
          SliverAppBar.large(
            title: const Text('Quản trị'),
            actions: [
              TextButton.icon(
                icon: const Icon(Icons.visibility_outlined),
                label: const Text('Xem như khách'),
                onPressed: () {
                  context.read<RoleCubit>().previewAsCustomer(true);
                  showSnack(context, 'Đang xem giao diện khách hàng. Vào Tài khoản để quay lại chế độ quản trị.');
                },
              ),
            ],
          ),
          const SliverToBoxAdapter(child: _ReviewQueues()),
          const SliverToBoxAdapter(child: SectionHeader('Khu quản lý')),
          SliverPadding(
            padding: const EdgeInsets.symmetric(horizontal: 16),
            sliver: SliverGrid(
              gridDelegate: const SliverGridDelegateWithMaxCrossAxisExtent(maxCrossAxisExtent: 220, mainAxisExtent: 112, mainAxisSpacing: 12, crossAxisSpacing: 12),
              delegate: SliverChildListDelegate([
                for (final c in _consoles)
                  Material(
                    color: theme.colorScheme.surfaceContainerLowest,
                    borderRadius: BorderRadius.circular(20),
                    child: InkWell(
                      borderRadius: BorderRadius.circular(20),
                      onTap: () => c.open(context),
                      child: Padding(
                        padding: const EdgeInsets.all(14),
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            IconBadge(icon: c.icon, color: c.color, size: 38),
                            const Spacer(),
                            Text(c.title, maxLines: 1, overflow: TextOverflow.ellipsis, style: const TextStyle(fontWeight: FontWeight.w700)),
                            Text(c.subtitle, maxLines: 1, overflow: TextOverflow.ellipsis, style: theme.textTheme.bodySmall),
                          ],
                        ),
                      ),
                    ),
                  ),
              ]),
            ),
          ),
          SliverToBoxAdapter(child: SectionHeader('Dịch vụ quản trị (${services.length})')),
          if (services.isEmpty)
            const SliverToBoxAdapter(child: Padding(padding: EdgeInsets.symmetric(horizontal: 20), child: Text('Tài khoản chưa được cấp quyền quản trị dịch vụ nào'))),
          SliverList.builder(
            itemCount: services.length,
            itemBuilder: (context, i) {
              final (svc, count) = services[i];
              final meta = metaFor(svc.id);
              return ListTile(
                contentPadding: const EdgeInsets.symmetric(horizontal: 20),
                leading: IconBadge(icon: meta.icon, color: meta.category.color, size: 40),
                title: Text(meta.title),
                subtitle: Text('$count chức năng quản trị'),
                trailing: const Icon(Icons.chevron_right),
                onTap: () => _push(context, meta.category.color, AdminConsolePage(serviceId: svc.id)),
              );
            },
          ),
          const SliverToBoxAdapter(child: SizedBox(height: 32)),
        ],
      ),
    );
  }
}

class _ReviewQueues extends StatelessWidget {
  const _ReviewQueues();

  @override
  Widget build(BuildContext context) {
    final api = context.api;
    return AsyncBuilder<(int, int, int)>(
      load: () async {
        final r = await Future.wait<Object?>([
          api.ticket.ticketEventReviewQueue(limit: 100).then((p) => p?.items?.length).catchError((_) => null),
          api.reservation.resvHotelReviewQueue(limit: 100).then((p) => p?.items?.length).catchError((_) => null),
          api.rentHouse.rentHomestayReviewQueue(limit: 100).then((l) => l?.length).catchError((_) => null),
        ]);
        return ((r[0] as int?) ?? 0, (r[1] as int?) ?? 0, (r[2] as int?) ?? 0);
      },
      loading: const Padding(padding: EdgeInsets.all(16), child: Skeleton(height: 120)),
      builder: (context, d, reload) {
        final total = d.$1 + d.$2 + d.$3;
        Widget tile(String label, int n, IconData icon, Color color, Widget page) => Expanded(
              child: InkWell(
                borderRadius: BorderRadius.circular(16),
                onTap: () async {
                  _push(context, color, page);
                },
                child: Container(
                  padding: const EdgeInsets.all(12),
                  decoration: BoxDecoration(color: Colors.white.withValues(alpha: 0.16), borderRadius: BorderRadius.circular(16)),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Icon(icon, color: Colors.white, size: 20),
                      const SizedBox(height: 6),
                      Text('$n', style: const TextStyle(color: Colors.white, fontSize: 22, fontWeight: FontWeight.w800)),
                      Text(label, maxLines: 1, overflow: TextOverflow.ellipsis, style: TextStyle(color: Colors.white.withValues(alpha: 0.85), fontSize: 12)),
                    ],
                  ),
                ),
              ),
            );
        return Container(
          margin: const EdgeInsets.fromLTRB(16, 0, 16, 0),
          padding: const EdgeInsets.all(18),
          decoration: BoxDecoration(gradient: KeikiColors.heroGradient, borderRadius: BorderRadius.circular(24)),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                children: [
                  const Icon(Icons.fact_check_rounded, color: Colors.white),
                  const SizedBox(width: 8),
                  Expanded(child: Text(total == 0 ? 'Không có gì chờ duyệt' : '$total mục chờ duyệt', style: const TextStyle(color: Colors.white, fontSize: 18, fontWeight: FontWeight.w800))),
                  IconButton(onPressed: reload, icon: const Icon(Icons.refresh, color: Colors.white)),
                ],
              ),
              const SizedBox(height: 12),
              Row(
                children: [
                  tile('Sự kiện', d.$1, Icons.local_activity_rounded, const Color(0xFFC026D3), const TicketOrganizerPage()),
                  const SizedBox(width: 8),
                  tile('Khách sạn', d.$2, Icons.hotel_rounded, const Color(0xFF1D4ED8), const HotelHostPage()),
                  const SizedBox(width: 8),
                  tile('Homestay', d.$3, Icons.cottage_rounded, const Color(0xFF0D9488), const HomestayHostPage()),
                ],
              ),
            ],
          ),
        );
      },
    );
  }
}

