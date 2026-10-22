import 'package:flutter/material.dart';

import '../../../gen/models.g.dart';
import '../kit/kit.dart';

String rateUnit(String? model) => switch (model) {
      'month' || 'monthly' => 'tháng',
      'week' || 'weekly' => 'tuần',
      'year' || 'yearly' => 'năm',
      _ => 'đêm',
    };

class LoyaltyWalletCard extends StatelessWidget {
  const LoyaltyWalletCard({super.key, required this.loyalty, required this.wallet, required this.createWallet});

  final ({int? points, String? tierName})? loyalty;
  final ({String? walletId, int? balance, String? currency, String? status})? wallet;
  final Future<Object?> Function() createWallet;

  @override
  Widget build(BuildContext context) {
    return AppBanner(
      icon: Icons.workspace_premium_rounded,
      title: '${loyalty?.points ?? 0} điểm',
      subtitle: 'Hạng ${loyalty?.tierName ?? 'Thành viên'}',
      children: [
        if (wallet?.walletId != null)
          StatRow([
            StatTile(label: 'Ví hoàn tiền', value: formatMoney(wallet!.balance, wallet!.currency), onDark: true),
            StatTile(label: 'Trạng thái', value: wallet!.status ?? '-', onDark: true),
          ])
        else
          SizedBox(
            width: double.infinity,
            child: FilledButton.icon(
              style: FilledButton.styleFrom(backgroundColor: Colors.white, foregroundColor: context.appColor),
              icon: const Icon(Icons.account_balance_wallet_outlined),
              label: const Text('Mở ví nhận hoàn tiền'),
              onPressed: () => runAction(context, createWallet, success: 'Đã mở ví, kéo xuống để làm mới'),
            ),
          ),
      ],
    );
  }
}

class _TxnTile extends StatelessWidget {
  const _TxnTile({required this.type, required this.amount, required this.currency, required this.description, required this.at});

  final String? type;
  final int? amount;
  final String? currency;
  final String? description;
  final String? at;

  @override
  Widget build(BuildContext context) {
    final inbound = (type ?? '').toLowerCase().contains('refund') || (type ?? '').toLowerCase().contains('credit') || (type ?? '').toLowerCase().contains('deposit');
    return ListTile(
      contentPadding: const EdgeInsets.symmetric(horizontal: 20),
      leading: Icon(inbound ? Icons.south_west_rounded : Icons.north_east_rounded, color: inbound ? const Color(0xFF16A34A) : const Color(0xFFDC2626)),
      title: Text(description ?? type ?? 'Giao dịch'),
      subtitle: Text(formatDate(at)),
      trailing: Text('${inbound ? '+' : '-'}${formatMoney(amount, currency)}', style: const TextStyle(fontWeight: FontWeight.w800)),
    );
  }
}

class ResvAccountTab extends StatefulWidget {
  const ResvAccountTab({super.key});

  @override
  State<ResvAccountTab> createState() => _ResvAccountTabState();
}

class _ResvAccountTabState extends State<ResvAccountTab> {
  int _v = 0;

  @override
  Widget build(BuildContext context) {
    final api = context.api.reservation;
    return AsyncBuilder<(ResvLoyalty?, ResvWallet?, List<ResvWalletTxn>, ResvUnreadCount?, List<ResvWait>)>(
      key: ValueKey(_v),
      load: () async {
        final r = await Future.wait<Object?>([
          api.resvLoyalty().catchError((_) => null),
          api.resvWallet().catchError((_) => null),
          api.resvWalletTransactions(limit: 20).catchError((_) => null),
          api.resvUnreadNotifications().catchError((_) => null),
          api.resvWaitlist(limit: 20).then((p) => p?.items).catchError((_) => null),
        ]);
        return (
          r[0] as ResvLoyalty?,
          r[1] as ResvWallet?,
          (r[2] as List<ResvWalletTxn>?) ?? const <ResvWalletTxn>[],
          r[3] as ResvUnreadCount?,
          (r[4] as List<ResvWait>?) ?? const <ResvWait>[],
        );
      },
      builder: (context, d, reload) => RefreshIndicator(
        onRefresh: reload,
        child: ListView(
          padding: const EdgeInsets.only(bottom: 96),
          children: [
            LoyaltyWalletCard(
              loyalty: d.$1 == null ? null : (points: d.$1!.points, tierName: d.$1!.tierName),
              wallet: d.$2 == null ? null : (walletId: d.$2!.walletId, balance: d.$2!.balance, currency: d.$2!.currency, status: d.$2!.status),
              createWallet: () async {
                final r = await api.resvCreateWallet(currency: 'VND');
                setState(() => _v++);
                return r;
              },
            ),
            Padding(
              padding: const EdgeInsets.fromLTRB(16, 12, 16, 0),
              child: EntityTile(
                title: 'Thông báo',
                subtitle: (d.$4?.unread ?? 0) == 0 ? 'Không có thông báo mới' : '${d.$4!.unread} thông báo chưa đọc',
                icon: Icons.notifications_outlined,
                onTap: () async {
                  await openPage(context, const _ResvNotifications());
                  setState(() => _v++);
                },
              ),
            ),
            const SectionHeader('Danh sách chờ phòng'),
            if (d.$5.isEmpty) const Padding(padding: EdgeInsets.symmetric(horizontal: 20), child: Text('Bạn không chờ phòng nào')),
            for (final w in d.$5)
              Padding(
                padding: const EdgeInsets.fromLTRB(16, 0, 16, 8),
                child: EntityTile(
                  title: 'Khách sạn #${w.hotelId} • loại phòng #${w.roomTypeId}',
                  subtitle: '${shortDate(w.checkIn)} – ${shortDate(w.checkOut)} • ${w.rooms ?? 1} phòng',
                  icon: Icons.hourglass_top_rounded,
                  badges: [StatusPill(w.state)],
                  trailing: TextButton(
                    onPressed: () async {
                      if (await runAction(context, () => api.resvLeaveWaitlist(id: w.id!), success: 'Đã rời danh sách chờ')) setState(() => _v++);
                    },
                    child: const Text('Rời'),
                  ),
                ),
              ),
            if (d.$3.isNotEmpty) ...[
              const SectionHeader('Giao dịch ví'),
              for (final t in d.$3) _TxnTile(type: t.type_, amount: t.amount, currency: t.currency, description: t.description, at: t.createdAt),
            ],
          ],
        ),
      ),
    );
  }
}

class _ResvNotifications extends StatelessWidget {
  const _ResvNotifications();

  @override
  Widget build(BuildContext context) {
    final api = context.api.reservation;
    return Scaffold(
      appBar: AppBar(
        title: const Text('Thông báo'),
        actions: [
          Builder(
            builder: (context) => TextButton(
              onPressed: () => runAction(context, () => api.resvMarkAllNotificationsRead(), success: 'Đã đánh dấu tất cả là đã đọc'),
              child: const Text('Đọc hết'),
            ),
          ),
        ],
      ),
      body: AsyncList<ResvNotification>(
        load: () async => (await api.resvNotifications(limit: 50))?.items,
        emptyMessage: 'Chưa có thông báo',
        emptyIcon: Icons.notifications_none,
        itemBuilder: (context, n, reload) => EntityTile(
          title: n.title ?? 'Thông báo',
          subtitle: n.body,
          caption: formatDate(n.createdAt),
          icon: n.read == true ? Icons.notifications_none_rounded : Icons.notifications_active_rounded,
          onTap: () async {
            if (n.read != true) await api.resvMarkNotificationRead(id: n.id!).catchError((_) => null);
            if (n.reservationId != null && context.mounted) await openPage(context, ResvDetailPage(id: n.reservationId!));
            await reload();
          },
        ),
      ),
    );
  }
}

class ResvDetailPage extends StatelessWidget {
  const ResvDetailPage({super.key, required this.id, this.manager = false});

  final int id;
  final bool manager;

  @override
  Widget build(BuildContext context) {
    final api = context.api.reservation;
    return AsyncBuilder<(ResvReservation?, List<ResvHistoryEntry>)>(
      load: () async {
        final r = await Future.wait<Object?>([api.resvReservation(id: id), api.resvReservationHistory(id: id).catchError((_) => null)]);
        return (r[0] as ResvReservation?, (r[1] as List<ResvHistoryEntry>?) ?? const <ResvHistoryEntry>[]);
      },
      loading: const Scaffold(body: SkeletonList()),
      builder: (context, d, reload) {
        final r = d.$1;
        Future<void> step(String label, Future<Object?> Function(String? reason) call, {bool ask = false}) async {
          var ok = false;
          if (ask) {
            ok = await showFormSheet(context, title: label, fields: const [KField('reason', 'Ghi chú / lý do')], success: 'Đã cập nhật', onSubmit: (v) => call(v.str('reason')));
          } else {
            ok = await runAction(context, () => call(null), confirm: '$label đặt phòng #$id?', success: 'Đã cập nhật');
          }
          if (ok) await reload();
        }

        return DetailPage(
          title: 'Đặt phòng #$id',
          header: AppBanner(
            icon: Icons.hotel_rounded,
            title: formatMoneyString(r?.totalAmount, r?.currency),
            subtitle: '${shortDate(r?.checkIn)} – ${shortDate(r?.checkOut)} • ${r?.nights ?? '-'} đêm',
            trailing: StatusPill(r?.statusName),
          ),
          children: [
            InfoSection(rows: [
              ('Khách sạn', '#${r?.hotelId ?? ''}'),
              ('Phòng', '${r?.rooms ?? 1} × loại #${r?.roomTypeId ?? ''}'),
              ('Khách', '${r?.adults ?? 0} người lớn, ${r?.children ?? 0} trẻ em'),
              ('Tiền phòng', formatMoneyString(r?.roomTotal, r?.currency)),
              ('Dịch vụ thêm', (r?.servicesTotal ?? '0') == '0' ? null : formatMoneyString(r?.servicesTotal, r?.currency)),
              ('Giảm giá', (r?.discount ?? '0') == '0' ? null : formatMoneyString(r?.discount, r?.currency)),
              ('Mã khuyến mãi', r?.promoCode),
              ('Yêu cầu', r?.specialRequests),
              ('Lưu trú', r?.stayStatus),
              ('Phí huỷ', (r?.cancellationFee ?? '0') == '0' ? null : formatMoneyString(r?.cancellationFee, r?.currency)),
              ('Hoàn tiền', (r?.refundAmount ?? '0') == '0' ? null : formatMoneyString(r?.refundAmount, r?.currency)),
            ]),
            InfoSection(title: 'Dịch vụ đi kèm', rows: [
              for (final s in r?.services ?? const <ResvAddon>[]) ('${s.name_ ?? ''} × ${s.quantity ?? 1}', formatMoneyString(s.amount, r?.currency)),
            ]),
            if (manager)
              ActionButtons([
                FilledButton(onPressed: () => step('Nhận phòng', (x) => api.resvCheckInReservation(id: id, reason: x)), child: const Text('Nhận phòng')),
                OutlinedButton(onPressed: () => step('Trả phòng', (x) => api.resvCheckOutReservation(id: id, reason: x)), child: const Text('Trả phòng')),
                OutlinedButton(onPressed: () => step('Khách không đến', (x) => api.resvNoShowReservation(id: id, reason: x), ask: true), child: const Text('Không đến')),
                TextButton(onPressed: () => step('Từ chối', (x) => api.resvRejectReservation(id: id, reason: x), ask: true), child: const Text('Từ chối')),
              ]),
            if (d.$2.isNotEmpty) ...[
              const SectionHeader('Lịch sử'),
              for (final h in d.$2)
                ListTile(
                  dense: true,
                  leading: Icon(Icons.history, color: context.appColor),
                  title: Text(h.event ?? '${h.fromStatus} → ${h.toStatus}'),
                  subtitle: Text([formatDate(h.at), h.by, h.note].whereType<String>().where((s) => s.isNotEmpty).join(' • ')),
                ),
            ],
          ],
        );
      },
    );
  }
}

class HotelExtras extends StatelessWidget {
  const HotelExtras({super.key, required this.hotelId});

  final int hotelId;

  @override
  Widget build(BuildContext context) {
    final api = context.api.reservation;
    return AsyncBuilder<(List<ResvRoomType>, List<ResvService>, List<ResvReview>)>(
      load: () async {
        final r = await Future.wait<Object?>([
          api.resvRoomTypes(id: hotelId).catchError((_) => null),
          api.resvServices(id: hotelId).catchError((_) => null),
          api.resvReviews(id: hotelId, limit: 10).then((p) => p?.items).catchError((_) => null),
        ]);
        return ((r[0] as List<ResvRoomType>?) ?? const <ResvRoomType>[], (r[1] as List<ResvService>?) ?? const <ResvService>[], (r[2] as List<ResvReview>?) ?? const <ResvReview>[]);
      },
      loading: const SizedBox.shrink(),
      builder: (context, d, _) => Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          if (d.$1.isNotEmpty) ...[
            const SectionHeader('Các loại phòng'),
            Strip<ResvRoomType>(
              items: d.$1,
              height: 118,
              itemWidth: 220,
              itemBuilder: (context, t) => KCard(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(t.name_ ?? '', maxLines: 1, overflow: TextOverflow.ellipsis, style: const TextStyle(fontWeight: FontWeight.w800)),
                    const SizedBox(height: 4),
                    Text([if (t.capacity != null) '${t.capacity} khách', t.bed, if (t.sizeM2 != null) '${t.sizeM2} m²'].whereType<String>().join(' • '), style: const TextStyle(fontSize: 12)),
                    const Spacer(),
                    Text((t.amenities ?? const []).take(3).join(', '), maxLines: 1, overflow: TextOverflow.ellipsis, style: TextStyle(fontSize: 12, color: context.appColor)),
                  ],
                ),
              ),
            ),
          ],
          if (d.$2.isNotEmpty)
            InfoSection(title: 'Dịch vụ thêm', rows: [
              for (final s in d.$2.where((s) => s.active != false)) (s.name_ ?? '', '${formatMoneyString(s.price)}${s.unit == null ? '' : '/${s.unit}'}'),
            ]),
          const SectionHeader('Đánh giá'),
          if (d.$3.isEmpty) const Padding(padding: EdgeInsets.symmetric(horizontal: 20), child: Text('Chưa có đánh giá')),
          for (final r in d.$3)
            ListTile(
              leading: CircleAvatar(child: Text('${r.rating ?? '-'}')),
              title: Text(r.comment ?? ''),
              subtitle: Text(shortDate(r.createdAt)),
            ),
        ],
      ),
    );
  }
}

Future<List<ResvAddonInput>> pickHotelServices(BuildContext context, int hotelId) async {
  final services = (await context.api.reservation.resvServices(id: hotelId).catchError((_) => null) ?? const <ResvService>[]).where((s) => s.active != false).toList();
  if (services.isEmpty || !context.mounted) return const [];
  final picked = <ResvAddonInput>[];
  await showFormSheet(
    context,
    title: 'Thêm dịch vụ',
    description: 'Nhập số lượng cho dịch vụ muốn đặt kèm, bỏ trống nếu không cần.',
    fields: [for (final s in services) KField('s${s.id}', '${s.name_} • ${formatMoneyString(s.price)}${s.unit == null ? '' : '/${s.unit}'}', kind: FieldKind.number)],
    submitLabel: 'Tiếp tục',
    onSubmit: (v) async {
      for (final s in services) {
        final q = v.integer('s${s.id}') ?? 0;
        if (q > 0) picked.add(ResvAddonInput(serviceId: s.id, quantity: q));
      }
    },
  );
  return picked;
}

Future<bool> joinHotelWaitlist(BuildContext context, {required int hotelId, required int roomTypeId, required String checkIn, required String checkOut, required int guests}) =>
    runAction(
      context,
      () => context.api.reservation.resvJoinWaitlist(input: {
        'hotel_id': hotelId,
        'room_type_id': roomTypeId,
        'check_in': checkIn,
        'check_out': checkOut,
        'rooms': 1,
        'adults': guests,
        'children': 0,
      }),
      success: 'Đã vào danh sách chờ, bạn sẽ được báo khi có phòng',
    );

class RentAccountTab extends StatefulWidget {
  const RentAccountTab({super.key});

  @override
  State<RentAccountTab> createState() => _RentAccountTabState();
}

class _RentAccountTabState extends State<RentAccountTab> {
  int _v = 0;

  @override
  Widget build(BuildContext context) {
    final api = context.api.rentHouse;
    return AsyncBuilder<(RentLoyalty?, RentWallet?, List<RentWalletTxn>, List<RentLease>, List<RentInvoice>)>(
      key: ValueKey(_v),
      load: () async {
        final r = await Future.wait<Object?>([
          api.rentLoyalty().catchError((_) => null),
          api.rentWallet().catchError((_) => null),
          api.rentWalletTransactions(limit: 20).catchError((_) => null),
          api.rentLeases(as_: 'guest', limit: 20).catchError((_) => null),
          api.rentInvoices(status: 'unpaid', limit: 20).catchError((_) => null),
        ]);
        return (
          r[0] as RentLoyalty?,
          r[1] as RentWallet?,
          (r[2] as List<RentWalletTxn>?) ?? const <RentWalletTxn>[],
          (r[3] as List<RentLease>?) ?? const <RentLease>[],
          (r[4] as List<RentInvoice>?) ?? const <RentInvoice>[],
        );
      },
      builder: (context, d, reload) => RefreshIndicator(
        onRefresh: reload,
        child: ListView(
          padding: const EdgeInsets.only(bottom: 96),
          children: [
            LoyaltyWalletCard(
              loyalty: d.$1 == null ? null : (points: d.$1!.points, tierName: d.$1!.tierName),
              wallet: d.$2 == null ? null : (walletId: d.$2!.walletId, balance: d.$2!.balance, currency: d.$2!.currency, status: d.$2!.status),
              createWallet: () async {
                final r = await api.rentCreateWallet(currency: 'VND');
                setState(() => _v++);
                return r;
              },
            ),
            const SectionHeader('Hoá đơn cần thanh toán'),
            if (d.$5.isEmpty) const Padding(padding: EdgeInsets.symmetric(horizontal: 20), child: Text('Không có hoá đơn tới hạn')),
            for (final i in d.$5) Padding(padding: const EdgeInsets.fromLTRB(16, 0, 16, 8), child: RentInvoiceTile(invoice: i, onPaid: () => setState(() => _v++))),
            const SectionHeader('Hợp đồng thuê'),
            if (d.$4.isEmpty) const Padding(padding: EdgeInsets.symmetric(horizontal: 20), child: Text('Bạn chưa thuê dài hạn nơi nào')),
            for (final l in d.$4)
              Padding(
                padding: const EdgeInsets.fromLTRB(16, 0, 16, 8),
                child: EntityTile(
                  title: 'Hợp đồng #${l.id} • nhà #${l.homestayId}',
                  subtitle: '${shortDate(l.startDate)} – ${shortDate(l.endDate)} • ${l.periods ?? '-'} ${rateUnit(l.model)}',
                  caption: '${formatMoneyString(l.rent, l.currency)}/${rateUnit(l.model)}',
                  icon: Icons.description_rounded,
                  badges: [StatusPill(l.status)],
                  onTap: () async {
                    await openPage(context, RentLeasePage(id: l.id!));
                    setState(() => _v++);
                  },
                ),
              ),
            if (d.$3.isNotEmpty) ...[
              const SectionHeader('Giao dịch ví'),
              for (final t in d.$3) _TxnTile(type: t.type_, amount: t.amount, currency: t.currency, description: t.description, at: t.createdAt),
            ],
          ],
        ),
      ),
    );
  }
}

class RentInvoiceTile extends StatelessWidget {
  const RentInvoiceTile({super.key, required this.invoice, required this.onPaid});

  final RentInvoice invoice;
  final VoidCallback onPaid;

  @override
  Widget build(BuildContext context) {
    final i = invoice;
    final unpaid = i.status == 'unpaid' || i.status == '1';
    return EntityTile(
      title: 'Kỳ ${i.period ?? '-'} • hợp đồng #${i.leaseId}',
      subtitle: '${shortDate(i.periodStart)} – ${shortDate(i.periodEnd)} • hạn ${shortDate(i.dueDate)}',
      caption: formatMoneyString(i.amount),
      icon: Icons.receipt_rounded,
      badges: [StatusPill(i.overdue == true ? 'overdue' : i.status), if (i.paidAt != null) Pill('Trả ${shortDate(i.paidAt)}')],
      trailing: unpaid
          ? FilledButton(
              onPressed: () async {
                if (await runAction(
                  context,
                  () => context.api.rentHouse.rentPayLeaseInvoice(id: i.leaseId!, period: '${i.period}', method: 'wallet'),
                  confirm: 'Thanh toán ${formatMoneyString(i.amount)} bằng Ví Keiki?',
                  success: 'Đã thanh toán hoá đơn',
                )) {
                  onPaid();
                }
              },
              child: const Text('Trả'),
            )
          : null,
    );
  }
}

class RentLeasePage extends StatelessWidget {
  const RentLeasePage({super.key, required this.id, this.host = false});

  final int id;
  final bool host;

  @override
  Widget build(BuildContext context) {
    final api = context.api.rentHouse;
    return AsyncBuilder<RentLease?>(
      load: () => api.rentLease(id: id),
      loading: const Scaffold(body: SkeletonList()),
      builder: (context, l, reload) => DetailPage(
        title: 'Hợp đồng #$id',
        header: AppBanner(
          icon: Icons.description_rounded,
          title: '${formatMoneyString(l?.rent, l?.currency)}/${rateUnit(l?.model)}',
          subtitle: '${shortDate(l?.startDate)} – ${shortDate(l?.endDate)} • thu ngày ${l?.billingDay ?? '-'} hằng kỳ',
          trailing: StatusPill(l?.status),
        ),
        children: [
          InfoSection(rows: [('Nhà', '#${l?.homestayId ?? ''}'), ('Số kỳ', '${l?.periods ?? '-'}'), ('Ghi chú', l?.note), ('Giữ chỗ tới', l?.expiresAt == null ? null : formatDate(l!.expiresAt))]),
          const SectionHeader('Hoá đơn'),
          for (final i in l?.invoices ?? const <RentInvoice>[]) Padding(padding: const EdgeInsets.fromLTRB(16, 0, 16, 8), child: RentInvoiceTile(invoice: i, onPaid: reload)),
          ActionButtons([
            if (!host && (l?.status == 'pending' || l?.status == '1'))
              OutlinedButton(
                onPressed: () async {
                  if (await runAction(context, () => api.rentCancelLease(id: id), confirm: 'Huỷ hợp đồng chưa bắt đầu này?', destructive: true, success: 'Đã huỷ')) await reload();
                },
                child: const Text('Huỷ hợp đồng'),
              ),
            if (l?.status == 'active' || l?.status == '2')
              OutlinedButton(
                onPressed: () async {
                  if (await runAction(context, () => api.rentTerminateLease(id: id), confirm: 'Chấm dứt hợp đồng trước hạn? Các kỳ chưa tới sẽ bị huỷ.', destructive: true, success: 'Đã chấm dứt')) await reload();
                },
                child: const Text('Chấm dứt trước hạn'),
              ),
          ]),
        ],
      ),
    );
  }
}

class HomestayExtras extends StatelessWidget {
  const HomestayExtras({super.key, required this.home, required this.onLease});

  final RentHomestay home;
  final VoidCallback onLease;

  @override
  Widget build(BuildContext context) {
    final api = context.api.rentHouse;
    final id = home.id!;
    final now = DateTime.now();
    String iso(DateTime d) => d.toIso8601String().substring(0, 10);
    return AsyncBuilder<(List<RentSlot>, List<RentRate>, List<RentAmenity>)>(
      load: () async {
        final r = await Future.wait<Object?>([
          api.rentAvailability(id: id, from: iso(now), to: iso(now.add(const Duration(days: 28)))).catchError((_) => null),
          api.rentRates(id: id).catchError((_) => null),
          api.rentAmenities().catchError((_) => null),
        ]);
        return ((r[0] as List<RentSlot>?) ?? const <RentSlot>[], (r[1] as List<RentRate>?) ?? const <RentRate>[], (r[2] as List<RentAmenity>?) ?? const <RentAmenity>[]);
      },
      loading: const SizedBox.shrink(),
      builder: (context, d, _) {
        final amenities = d.$3.where((a) => (home.amenityIds ?? const []).contains(a.id)).toList();
        final longTerm = d.$2.where((r) => r.active != false && r.model != 'day').toList();
        return Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            if (amenities.isNotEmpty)
              Padding(
                padding: const EdgeInsets.fromLTRB(16, 12, 16, 0),
                child: Wrap(spacing: 6, runSpacing: 6, children: [for (final a in amenities) Pill(a.name_ ?? '')]),
              ),
            if (d.$1.isNotEmpty) ...[
              const SectionHeader('Lịch trống 4 tuần tới'),
              Padding(
                padding: const EdgeInsets.symmetric(horizontal: 16),
                child: GridView.count(
                  crossAxisCount: 7,
                  shrinkWrap: true,
                  physics: const NeverScrollableScrollPhysics(),
                  mainAxisSpacing: 4,
                  crossAxisSpacing: 4,
                  children: [
                    for (final s in d.$1)
                      Container(
                        alignment: Alignment.center,
                        decoration: BoxDecoration(
                          color: s.status == 1 ? context.appColor.withValues(alpha: 0.12) : Colors.grey.withValues(alpha: 0.18),
                          borderRadius: BorderRadius.circular(8),
                        ),
                        child: Text(
                          (s.date ?? '').length >= 10 ? s.date!.substring(8, 10) : '',
                          style: TextStyle(fontWeight: FontWeight.w700, color: s.status == 1 ? context.appColor : Colors.grey, decoration: s.status == 1 ? null : TextDecoration.lineThrough),
                        ),
                      ),
                  ],
                ),
              ),
            ],
            if (longTerm.isNotEmpty)
              ActionButtons([
                OutlinedButton.icon(
                  icon: const Icon(Icons.description_outlined),
                  label: Text('Thuê theo ${rateUnit(longTerm.first.model)}'),
                  onPressed: () => showFormSheet(
                    context,
                    title: 'Thuê dài hạn',
                    description: 'Giá ${formatMoneyString(longTerm.first.price, longTerm.first.currency)}/${rateUnit(longTerm.first.model)}, tối thiểu ${longTerm.first.minPeriods ?? 1} kỳ.',
                    fields: [
                      KField('model', 'Hình thức', kind: FieldKind.choice, initial: longTerm.first.model, options: [for (final r in longTerm) (r.model ?? '', 'Theo ${rateUnit(r.model)}')]),
                      KField('start', 'Ngày bắt đầu', kind: FieldKind.date, required: true, initial: iso(now.add(const Duration(days: 7)))),
                      KField('periods', 'Số kỳ', kind: FieldKind.number, required: true, initial: '${longTerm.first.minPeriods ?? 1}'),
                      const KField('day', 'Ngày thu tiền hằng kỳ', kind: FieldKind.number, initial: '5'),
                      const KField('note', 'Ghi chú', kind: FieldKind.multiline),
                    ],
                    submitLabel: 'Gửi yêu cầu thuê',
                    success: 'Đã tạo hợp đồng, thanh toán kỳ đầu trong tab Tài khoản',
                    onSubmit: (v) async {
                      await api.rentStartLease(
                        input: RentLeaseInput(
                          homestayId: id,
                          model: v.str('model'),
                          startDate: v.str('start'),
                          periods: v.integer('periods'),
                          billingDay: v.integer('day'),
                          note: v.str('note'),
                          requestId: newIdempotencyKey(),
                        ),
                      );
                      onLease();
                    },
                  ),
                ),
              ]),
          ],
        );
      },
    );
  }
}

class RentBookingPage extends StatelessWidget {
  const RentBookingPage({super.key, required this.id});

  final int id;

  @override
  Widget build(BuildContext context) {
    return AsyncBuilder<RentBooking?>(
      load: () => context.api.rentHouse.rentBooking(id: id),
      loading: const Scaffold(body: SkeletonList()),
      builder: (context, b, _) => DetailPage(
        title: 'Đặt chỗ #$id',
        header: AppBanner(
          icon: Icons.cottage_rounded,
          title: formatMoneyString(b?.totalAmount, b?.currency),
          subtitle: '${shortDate(b?.checkinDate)} – ${shortDate(b?.checkoutDate)} • ${b?.nights ?? '-'} đêm',
          trailing: StatusPill(switch (b?.status) { 1 => 'confirmed', 2 => 'cancelled', 3 => 'pending', _ => null }),
        ),
        children: [
          InfoSection(rows: [
            ('Homestay', '#${b?.homestayId ?? ''}'),
            ('Số khách', '${b?.guests ?? '-'}'),
            ('Tạm tính', formatMoneyString(b?.subtotal, b?.currency)),
            ('Giảm giá', (b?.discount ?? '0') == '0' ? null : formatMoneyString(b?.discount, b?.currency)),
            ('Thanh toán', b?.paymentMethod),
            ('Mã giao dịch', b?.paymentRef),
            ('Trả lúc', b?.paidAt == null ? null : formatDate(b!.paidAt)),
            ('Giữ chỗ tới', b?.status == 3 ? formatDate(b?.expiresAt) : null),
            ('Ghi chú', b?.note),
          ]),
        ],
      ),
    );
  }
}
