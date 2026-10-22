import 'package:flutter/material.dart';
import 'package:flutter/services.dart';

import '../../../gen/models.g.dart';
import '../kit/kit.dart';

Color _seatColor(BuildContext context, String? status, bool picked) {
  if (picked) return context.appColor;
  return switch (status) {
    'available' || null => const Color(0xFF16A34A),
    'held' || 'reserved' => const Color(0xFFD97706),
    _ => const Color(0xFF94A3B8),
  };
}

class SeatPickerPage extends StatefulWidget {
  const SeatPickerPage({super.key, required this.eventId, required this.type, this.initial = const [], this.max = 10});

  final int eventId;
  final TicketTicketType type;
  final List<int> initial;
  final int max;

  @override
  State<SeatPickerPage> createState() => _SeatPickerPageState();
}

class _SeatPickerPageState extends State<SeatPickerPage> {
  late final _picked = {...widget.initial};

  @override
  Widget build(BuildContext context) {
    return AsyncBuilder<TicketSeatList?>(
      load: () => context.api.ticket.ticketSeats(id: widget.eventId, type_: widget.type.id!),
      loading: const Scaffold(body: SkeletonList()),
      builder: (context, list, _) {
        final seats = list?.items ?? const <TicketSeat>[];
        final sections = <String, List<TicketSeat>>{};
        for (final s in seats) {
          sections.putIfAbsent(s.section ?? 'Khu chung', () => []).add(s);
        }
        return DetailPage(
          title: 'Chọn ghế • ${widget.type.name_ ?? ''}',
          bottom: Row(
            children: [
              Expanded(child: Text('Đã chọn ${_picked.length}/${widget.max} ghế', style: const TextStyle(fontWeight: FontWeight.w700))),
              FilledButton(onPressed: () => Navigator.pop(context, _picked.toList()), child: const Text('Xong')),
            ],
          ),
          children: [
            Padding(
              padding: const EdgeInsets.fromLTRB(16, 12, 16, 0),
              child: Wrap(spacing: 12, children: [
                for (final (label, status) in const [('Còn trống', 'available'), ('Đang giữ', 'held'), ('Đã bán', 'sold')])
                  Row(mainAxisSize: MainAxisSize.min, children: [
                    Container(width: 14, height: 14, decoration: BoxDecoration(color: _seatColor(context, status, false), borderRadius: BorderRadius.circular(4))),
                    const SizedBox(width: 4),
                    Text(label, style: const TextStyle(fontSize: 12)),
                  ]),
              ]),
            ),
            if (seats.isEmpty) const Padding(padding: EdgeInsets.all(32), child: Center(child: Text('Loại vé này chưa có sơ đồ ghế'))),
            for (final entry in sections.entries) ...[
              SectionHeader(entry.key),
              Padding(
                padding: const EdgeInsets.symmetric(horizontal: 16),
                child: Wrap(
                  spacing: 6,
                  runSpacing: 6,
                  children: [
                    for (final s in entry.value)
                      Semantics(
                        button: true,
                        label: 'Ghế ${s.label ?? '${s.row}${s.number}'}',
                        child: InkWell(
                          borderRadius: BorderRadius.circular(8),
                          onTap: s.status != 'available' && !_picked.contains(s.id)
                              ? null
                              : () => setState(() {
                                    if (!_picked.remove(s.id) && _picked.length < widget.max) _picked.add(s.id!);
                                  }),
                          child: Container(
                            width: 40,
                            height: 34,
                            alignment: Alignment.center,
                            decoration: BoxDecoration(
                              color: _seatColor(context, s.status, _picked.contains(s.id)).withValues(alpha: _picked.contains(s.id) ? 1 : 0.18),
                              borderRadius: BorderRadius.circular(8),
                              border: Border.all(color: _seatColor(context, s.status, _picked.contains(s.id))),
                            ),
                            child: Text(
                              s.label ?? '${s.row ?? ''}${s.number ?? ''}',
                              style: TextStyle(fontSize: 11, fontWeight: FontWeight.w700, color: _picked.contains(s.id) ? Colors.white : null),
                            ),
                          ),
                        ),
                      ),
                  ],
                ),
              ),
            ],
          ],
        );
      },
    );
  }
}

class SeatMapPage extends StatelessWidget {
  const SeatMapPage({super.key, required this.eventId, required this.title});

  final int eventId;
  final String title;

  @override
  Widget build(BuildContext context) {
    return AsyncBuilder<TicketSeatMap?>(
      load: () => context.api.ticket.ticketSeatMap(id: eventId),
      loading: const Scaffold(body: SkeletonList()),
      builder: (context, map, _) {
        final seats = map?.seats ?? const <TicketSeat>[];
        final w = map?.layout?.width ?? seats.fold<double>(0, (a, s) => (s.x ?? 0) > a ? s.x! : a) + 20;
        final h = map?.layout?.height ?? seats.fold<double>(0, (a, s) => (s.y ?? 0) > a ? s.y! : a) + 20;
        final types = {for (final t in map?.ticketTypes ?? const <TicketTicketType>[]) t.id: t};
        final palette = [const Color(0xFF7C3AED), const Color(0xFFDB2777), const Color(0xFF2563EB), const Color(0xFF16A34A), const Color(0xFFD97706)];
        final colorOf = {for (final (i, id) in types.keys.indexed) id: palette[i % palette.length]};
        return DetailPage(
          title: 'Sơ đồ • $title',
          children: [
            if (seats.isEmpty)
              const Padding(padding: EdgeInsets.all(32), child: Center(child: Text('Sự kiện không có ghế ngồi đánh số')))
            else
              Padding(
                padding: const EdgeInsets.all(16),
                child: KCard(
                  child: AspectRatio(
                    aspectRatio: w <= 0 || h <= 0 ? 1 : w / h,
                    child: LayoutBuilder(
                      builder: (context, c) {
                        final sx = c.maxWidth / (w <= 0 ? 1 : w);
                        final sy = c.maxHeight / (h <= 0 ? 1 : h);
                        return Stack(
                          children: [
                            Positioned(
                              left: 0,
                              right: 0,
                              top: 0,
                              child: Container(
                                height: 18,
                                alignment: Alignment.center,
                                decoration: BoxDecoration(color: Colors.black87, borderRadius: BorderRadius.circular(6)),
                                child: const Text('SÂN KHẤU', style: TextStyle(color: Colors.white, fontSize: 10, fontWeight: FontWeight.w800)),
                              ),
                            ),
                            for (final s in seats)
                              Positioned(
                                left: (s.x ?? 0) * sx,
                                top: 22 + (s.y ?? 0) * sy * 0.9,
                                child: Container(
                                  width: 7,
                                  height: 7,
                                  decoration: BoxDecoration(
                                    shape: BoxShape.circle,
                                    color: s.status == 'available' ? colorOf[s.ticketTypeId] ?? context.appColor : const Color(0xFFCBD5E1),
                                  ),
                                ),
                              ),
                          ],
                        );
                      },
                    ),
                  ),
                ),
              ),
            InfoSection(title: 'Chú thích', rows: [
              for (final t in types.values) (t.name_ ?? 'Vé', '${vnd(t.price)} • còn ${t.available ?? '-'}'),
            ]),
            Padding(
              padding: const EdgeInsets.fromLTRB(16, 8, 16, 0),
              child: Wrap(spacing: 10, runSpacing: 6, children: [
                for (final t in types.values)
                  Row(mainAxisSize: MainAxisSize.min, children: [
                    CircleAvatar(radius: 5, backgroundColor: colorOf[t.id]),
                    const SizedBox(width: 4),
                    Text(t.name_ ?? '', style: const TextStyle(fontSize: 12)),
                  ]),
              ]),
            ),
          ],
        );
      },
    );
  }
}

class ResaleSection extends StatelessWidget {
  const ResaleSection({super.key, required this.eventId, required this.onBought});

  final int eventId;
  final VoidCallback onBought;

  @override
  Widget build(BuildContext context) {
    final api = context.api.ticket;
    return AsyncBuilder<TicketResalePage?>(
      load: () => api.ticketResale(id: eventId, sort: 'price', limit: 10).catchError((_) => null),
      loading: const SizedBox.shrink(),
      builder: (context, page, reload) {
        final items = page?.items ?? const <TicketResale>[];
        if (items.isEmpty) return const SizedBox.shrink();
        return Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            const SectionHeader('Vé bán lại'),
            for (final r in items)
              Padding(
                padding: const EdgeInsets.fromLTRB(16, 0, 16, 8),
                child: EntityTile(
                  title: '${r.typeName ?? 'Vé'}${r.seat == null ? '' : ' • ghế ${r.seat}'}',
                  subtitle: 'Giá gốc ${formatMoney(r.facePrice, r.currency)}${(r.fee ?? 0) > 0 ? ' • phí ${formatMoney(r.fee, r.currency)}' : ''}',
                  caption: formatMoney(r.price, r.currency),
                  icon: Icons.sell_rounded,
                  trailing: FilledButton(
                    onPressed: () async {
                      if (await runAction(
                        context,
                        () => api.ticketBuyResale(id: r.id!),
                        confirm: 'Mua lại vé này với giá ${formatMoney(r.price, r.currency)}? Tiền trả bằng Ví Keiki.',
                        success: 'Đã mua vé, xem trong Vé của tôi',
                      )) {
                        await reload();
                        onBought();
                      }
                    },
                    child: const Text('Mua'),
                  ),
                ),
              ),
          ],
        );
      },
    );
  }
}

class SeriesStrip extends StatelessWidget {
  const SeriesStrip({super.key, required this.eventId, required this.open});

  final int eventId;
  final void Function(int id) open;

  @override
  Widget build(BuildContext context) {
    return AsyncBuilder<List<TicketEvent>?>(
      load: () => context.api.ticket.ticketEventSeries(id: eventId).catchError((_) => null),
      loading: const SizedBox.shrink(),
      builder: (context, list, _) {
        final others = (list ?? const <TicketEvent>[]).where((e) => e.id != eventId).toList();
        if (others.isEmpty) return const SizedBox.shrink();
        return Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            const SectionHeader('Các ngày khác'),
            Strip<TicketEvent>(
              items: others,
              height: 92,
              itemWidth: 170,
              itemBuilder: (context, e) => KCard(
                onTap: () => open(e.id!),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(shortDate(e.startsAt), style: TextStyle(fontWeight: FontWeight.w800, color: context.appColor)),
                    const SizedBox(height: 4),
                    Text(e.city ?? e.venue ?? '', maxLines: 2, overflow: TextOverflow.ellipsis, style: const TextStyle(fontSize: 12)),
                  ],
                ),
              ),
            ),
          ],
        );
      },
    );
  }
}

class TicketOrderDetailPage extends StatelessWidget {
  const TicketOrderDetailPage({super.key, required this.id});

  final int id;

  @override
  Widget build(BuildContext context) {
    final api = context.api.ticket;
    return AsyncBuilder<(TicketOrder?, TicketInvoice?, List<TicketDocument>)>(
      load: () async {
        final r = await Future.wait<Object?>([
          api.ticketOrder(id: id),
          api.ticketInvoice(id: id).catchError((_) => null),
          api.ticketOrderDocuments(id: id).catchError((_) => null),
        ]);
        return (r[0] as TicketOrder?, r[1] as TicketInvoice?, (r[2] as List<TicketDocument>?) ?? const <TicketDocument>[]);
      },
      loading: const Scaffold(body: SkeletonList()),
      builder: (context, d, reload) {
        final o = d.$1;
        final inv = d.$2;
        return DetailPage(
          title: 'Đơn #$id',
          header: AppBanner(
            icon: Icons.receipt_long_rounded,
            title: formatMoney(o?.total, o?.currency),
            subtitle: '${o?.tickets?.length ?? 0} vé • ${formatDate(o?.createdAt)}',
            trailing: StatusPill(o?.status),
          ),
          children: [
            InfoSection(title: 'Đơn hàng', rows: [
              for (final it in o?.items ?? const <TicketOrderItem>[]) ('${it.name_ ?? 'Vé'} × ${it.quantity ?? 0}', formatMoney((it.unitPrice ?? 0) * (it.quantity ?? 0), o?.currency)),
              ('Tạm tính', formatMoney(o?.subtotal, o?.currency)),
              ('Giảm giá', (o?.discount ?? 0) == 0 ? null : '-${formatMoney(o!.discount, o.currency)}'),
              ('Mã khuyến mãi', o?.promoCode),
              ('Thanh toán', o?.paymentMethod),
              ('Hạn thanh toán', o?.status == 'pending' ? formatDate(o?.expiresAt) : null),
              ('Đã hoàn', (o?.refundAmount ?? 0) == 0 ? null : formatMoney(o!.refundAmount, o.currency)),
            ]),
            if (inv != null)
              InfoSection(title: 'Hoá đơn ${inv.number ?? ''}', rows: [
                ('Ngày xuất', formatDate(inv.issuedAt)),
                ('Bên bán', inv.sellerName),
                ('Mã số thuế', inv.sellerTaxId),
                ('Người mua', inv.buyerName),
                ('Thuế VAT', inv.vatPercent == null ? null : '${inv.vatPercent}% • ${formatMoney(inv.vat, inv.currency)}'),
                ('Tổng cộng', formatMoney(inv.total, inv.currency)),
              ]),
            if (d.$3.isNotEmpty) ...[
              const SectionHeader('Chứng từ'),
              for (final doc in d.$3)
                Padding(
                  padding: const EdgeInsets.fromLTRB(16, 0, 16, 8),
                  child: EntityTile(
                    title: doc.fileName ?? doc.kind ?? 'Tài liệu',
                    subtitle: '${doc.kind ?? ''} • ${formatDate(doc.createdAt)}',
                    icon: Icons.picture_as_pdf_rounded,
                    trailing: IconButton(
                      tooltip: 'Sao chép liên kết',
                      icon: const Icon(Icons.link),
                      onPressed: () async {
                        await Clipboard.setData(ClipboardData(text: doc.url ?? ''));
                        if (context.mounted) showSnack(context, 'Đã sao chép liên kết tải chứng từ');
                      },
                    ),
                  ),
                ),
            ],
            ActionButtons([
              if (o?.status == 'pending' || o?.status == 'reserved')
                FilledButton.icon(
                  icon: const Icon(Icons.account_balance_wallet_outlined),
                  label: const Text('Thanh toán bằng ví'),
                  onPressed: () async {
                    if (await runAction(context, () => api.ticketPayOrder(id: id, method: 'wallet'), success: 'Thanh toán thành công')) await reload();
                  },
                ),
              if (o != null && o.status != 'cancelled' && o.status != 'refunded')
                OutlinedButton.icon(
                  icon: const Icon(Icons.cancel_outlined),
                  label: const Text('Huỷ đơn'),
                  onPressed: () async {
                    if (await showFormSheet(
                      context,
                      title: 'Huỷ đơn #$id',
                      description: 'Đơn đã thanh toán được hoàn tiền theo chính sách hoàn vé của sự kiện.',
                      fields: const [KField('reason', 'Lý do')],
                      submitLabel: 'Huỷ đơn',
                      success: 'Đã huỷ đơn',
                      onSubmit: (v) => api.ticketCancelOrder(id: id, reason: v.str('reason')),
                    )) {
                      await reload();
                    }
                  },
                ),
            ]),
          ],
        );
      },
    );
  }
}

class TicketHistorySection extends StatelessWidget {
  const TicketHistorySection({super.key, required this.ticketId});

  final int ticketId;

  @override
  Widget build(BuildContext context) {
    return AsyncBuilder<List<TicketTicketEvent>?>(
      load: () => context.api.ticket.ticketHistory(id: ticketId).catchError((_) => null),
      loading: const SizedBox.shrink(),
      builder: (context, events, _) => (events ?? const []).isEmpty
          ? const SizedBox.shrink()
          : Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                const SectionHeader('Lịch sử vé'),
                for (final e in events!)
                  ListTile(
                    dense: true,
                    leading: Icon(Icons.history, color: context.appColor),
                    title: Text(_eventLabel(e.event)),
                    subtitle: Text([if (e.fromStatus != null) '${e.fromStatus} → ${e.toStatus}', e.note, formatDate(e.createdAt)].whereType<String>().where((s) => s.isNotEmpty).join(' • ')),
                  ),
              ],
            ),
    );
  }
}

String _eventLabel(String? e) => switch (e) {
      'issued' => 'Phát hành vé',
      'transferred' => 'Chuyển nhượng',
      'checked_in' => 'Check-in',
      'check_in_reverted' => 'Huỷ check-in',
      'listed' => 'Đăng bán lại',
      'resold' => 'Đã bán lại',
      'reissued' => 'Cấp lại mã',
      'voided' => 'Vô hiệu',
      'holder_changed' => 'Đổi người giữ',
      _ => e ?? 'Cập nhật',
    };

class TicketInbox extends StatefulWidget {
  const TicketInbox({super.key});

  @override
  State<TicketInbox> createState() => _TicketInboxState();
}

class _TicketInboxState extends State<TicketInbox> {
  int _tab = 0;

  @override
  Widget build(BuildContext context) {
    final api = context.api.ticket;
    return Column(
      children: [
        Padding(
          padding: const EdgeInsets.fromLTRB(16, 12, 16, 0),
          child: SegmentedButton<int>(
            showSelectedIcon: false,
            segments: const [
              ButtonSegment(value: 0, label: Text('Thông báo')),
              ButtonSegment(value: 1, label: Text('Chuyển vé')),
              ButtonSegment(value: 2, label: Text('Bán lại')),
              ButtonSegment(value: 3, label: Text('Chờ vé')),
            ],
            selected: {_tab},
            onSelectionChanged: (s) => setState(() => _tab = s.first),
          ),
        ),
        Expanded(
          child: switch (_tab) {
            0 => AsyncList<TicketNotification>(
                key: const ValueKey('n'),
                load: () async => (await api.ticketNotifications(limit: 50))?.items,
                header: AsyncBuilder<TicketUnreadCount?>(
                  load: () => api.ticketUnreadNotifications().catchError((_) => null),
                  loading: const SizedBox.shrink(),
                  builder: (context, c, _) => (c?.unread ?? 0) == 0
                      ? const SizedBox.shrink()
                      : Padding(
                          padding: const EdgeInsets.fromLTRB(16, 8, 16, 0),
                          child: Row(
                            children: [
                              Expanded(child: Text('${c!.unread} thông báo chưa đọc', style: const TextStyle(fontWeight: FontWeight.w700))),
                              TextButton(
                                onPressed: () async {
                                  if (await runAction(context, () => api.ticketMarkAllNotificationsRead())) setState(() {});
                                },
                                child: const Text('Đánh dấu đã đọc'),
                              ),
                            ],
                          ),
                        ),
                ),
                emptyMessage: 'Chưa có thông báo',
                emptyIcon: Icons.notifications_none,
                itemBuilder: (context, n, reload) => EntityTile(
                  title: n.title ?? 'Thông báo',
                  subtitle: n.body,
                  caption: formatDate(n.createdAt),
                  icon: n.readAt == null ? Icons.notifications_active_rounded : Icons.notifications_none_rounded,
                  onTap: () async {
                    if (n.readAt == null) await api.ticketMarkNotificationRead(id: n.id!).catchError((_) => null);
                    if (n.orderId != null && context.mounted) await openPage(context, TicketOrderDetailPage(id: n.orderId!));
                    await reload();
                  },
                ),
              ),
            1 => AsyncList<TicketTransfer>(
                key: const ValueKey('t'),
                load: () async => (await api.ticketTransfers(limit: 50))?.items,
                emptyMessage: 'Không có lời mời chuyển vé',
                emptyIcon: Icons.swap_horiz,
                itemBuilder: (context, t, reload) {
                  final incoming = t.toUser == context.userId;
                  final open = t.status == 'pending';
                  return EntityTile(
                    title: t.ticket?.eventTitle ?? 'Vé #${t.ticketId}',
                    subtitle: incoming ? 'Từ người dùng #${t.fromUser}${t.message == null ? '' : ': “${t.message}”'}' : 'Gửi tới ${t.toEmail ?? '#${t.toUser}'}',
                    caption: open ? 'Hết hạn ${formatDate(t.expiresAt)}' : null,
                    icon: incoming ? Icons.call_received_rounded : Icons.call_made_rounded,
                    badges: [StatusPill(t.status)],
                    trailing: !open
                        ? null
                        : incoming
                            ? Row(
                                mainAxisSize: MainAxisSize.min,
                                children: [
                                  IconButton(
                                    tooltip: 'Từ chối',
                                    icon: const Icon(Icons.close),
                                    onPressed: () async {
                                      if (await runAction(context, () => api.ticketDeclineTransfer(id: t.id!), success: 'Đã từ chối')) await reload();
                                    },
                                  ),
                                  IconButton.filled(
                                    tooltip: 'Nhận vé',
                                    icon: const Icon(Icons.check),
                                    onPressed: () async {
                                      if (await runAction(context, () => api.ticketAcceptTransfer(id: t.id!), success: 'Vé đã về tài khoản của bạn')) await reload();
                                    },
                                  ),
                                ],
                              )
                            : TextButton(
                                onPressed: () async {
                                  if (await runAction(context, () => api.ticketCancelTransfer(id: t.id!), confirm: 'Huỷ lời mời chuyển vé?', success: 'Đã huỷ')) await reload();
                                },
                                child: const Text('Huỷ'),
                              ),
                  );
                },
              ),
            2 => AsyncList<TicketResale>(
                key: const ValueKey('r'),
                load: () async => (await api.ticketMyResale(limit: 50))?.items,
                emptyMessage: 'Bạn chưa đăng bán lại vé nào',
                emptyIcon: Icons.sell_outlined,
                itemBuilder: (context, r, reload) => EntityTile(
                  title: '${r.typeName ?? 'Vé'}${r.seat == null ? '' : ' • ghế ${r.seat}'}',
                  subtitle: 'Đăng lúc ${formatDate(r.createdAt)}',
                  caption: formatMoney(r.price, r.currency),
                  icon: Icons.sell_rounded,
                  badges: [StatusPill(r.status)],
                  trailing: r.status == 'listed' || r.status == 'active'
                      ? TextButton(
                          onPressed: () async {
                            if (await runAction(context, () => api.ticketCancelResale(id: r.id!), success: 'Đã gỡ khỏi chợ bán lại')) await reload();
                          },
                          child: const Text('Gỡ'),
                        )
                      : null,
                ),
              ),
            _ => AsyncList<TicketWaitlist>(
                key: const ValueKey('w'),
                load: () => api.ticketWaitlist(),
                emptyMessage: 'Bạn không chờ vé nào',
                emptyIcon: Icons.hourglass_empty,
                itemBuilder: (context, w, reload) => EntityTile(
                  title: w.eventTitle ?? 'Sự kiện #${w.eventId}',
                  subtitle: '${w.typeName ?? ''} • đăng ký ${formatDate(w.createdAt)}',
                  icon: Icons.hourglass_top_rounded,
                  trailing: TextButton(
                    onPressed: () async {
                      if (await runAction(context, () => api.ticketLeaveWaitlist(id: w.eventId!, type_: w.ticketTypeId!), success: 'Đã rời danh sách chờ')) await reload();
                    },
                    child: const Text('Rời'),
                  ),
                ),
              ),
          },
        ),
      ],
    );
  }
}
