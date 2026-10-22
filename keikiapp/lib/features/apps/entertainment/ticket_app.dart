import 'package:flutter/material.dart';

import '../../../gen/models.g.dart';
import '../kit/kit.dart';
import 'ticket_more.dart';
import 'ticket_organizer.dart';

class TicketApp extends StatelessWidget {
  const TicketApp({super.key});

  @override
  Widget build(BuildContext context) {
    return ServiceApp(
      serviceId: 'ticket',
      adminActions: [
        Builder(
          builder: (context) => IconButton(
            tooltip: 'Ban tổ chức',
            icon: const Icon(Icons.storefront_outlined),
            onPressed: () => openPage(context, const TicketOrganizerPage()),
          ),
        ),
      ],
      tabs: [
        AppTab('Sự kiện', Icons.local_activity_outlined, (_) => const _Events()),
        AppTab('Vé của tôi', Icons.confirmation_number_outlined, (_) => const _MyTickets()),
        AppTab('Đơn hàng', Icons.receipt_long_outlined, (_) => const _Orders()),
        AppTab('Yêu thích', Icons.favorite_border, (context) => AsyncList<TicketEvent>(
              load: () => context.api.ticket.ticketWishlist(limit: 50),
              emptyMessage: 'Chưa có sự kiện yêu thích',
              emptyIcon: Icons.favorite_border,
              itemBuilder: (context, e, reload) => Dismissible(
                key: ValueKey(e.id),
                direction: DismissDirection.endToStart,
                background: Container(
                  alignment: Alignment.centerRight,
                  padding: const EdgeInsets.only(right: 24),
                  decoration: BoxDecoration(color: const Color(0xFFDC2626), borderRadius: BorderRadius.circular(20)),
                  child: const Icon(Icons.heart_broken, color: Colors.white),
                ),
                confirmDismiss: (_) => runAction(context, () => context.api.ticket.ticketRemoveWishlist(id: e.id!), success: 'Đã bỏ yêu thích'),
                onDismissed: (_) => reload(),
                child: _EventTile(event: e, onBack: reload),
              ),
            )),
        AppTab('Hộp thư', Icons.inbox_outlined, (_) => const TicketInbox()),
      ],
    );
  }
}

String _eventWhen(TicketEvent e) => [formatDate(e.nextSessionAt ?? e.startsAt), e.city].whereType<String>().where((s) => s.isNotEmpty).join(' • ');

class _Events extends StatelessWidget {
  const _Events();

  @override
  Widget build(BuildContext context) {
    return SearchableList<TicketEvent>(
      hint: 'Tìm concert, hội thảo, thể thao…',
      filters: () => context.api.ticket.ticketCategories(),
      load: (q, category) => context.api.ticket.ticketEvents(q: q.isEmpty ? null : q, category: category, limit: 40, sort: 'starts_at'),
      gridExtent: 220,
      gridAspect: 0.72,
      emptyMessage: 'Chưa có sự kiện phù hợp',
      emptyIcon: Icons.event_busy_outlined,
      itemBuilder: (context, e, reload) => PosterCard(
        title: e.title ?? 'Sự kiện',
        subtitle: _eventWhen(e),
        imageUrl: e.bannerUrl,
        icon: Icons.local_activity_rounded,
        badge: e.soldOut == true ? 'Hết vé' : (e.minPrice == null ? null : (e.minPrice == 0 ? 'Miễn phí' : 'Từ ${vnd(e.minPrice)}')),
        onTap: () => openPage(context, TicketEventPage(id: e.id!)),
      ),
    );
  }
}

class _EventTile extends StatelessWidget {
  const _EventTile({required this.event, this.onBack});

  final TicketEvent event;
  final Future<void> Function()? onBack;

  @override
  Widget build(BuildContext context) {
    return EntityTile(
      title: event.title ?? 'Sự kiện',
      subtitle: [event.venue, _eventWhen(event)].whereType<String>().where((s) => s.isNotEmpty).join('\n'),
      imageUrl: event.bannerUrl,
      icon: Icons.local_activity_rounded,
      caption: event.minPrice == null ? null : 'Từ ${vnd(event.minPrice)}',
      onTap: () async {
        await openPage(context, TicketEventPage(id: event.id!));
        await onBack?.call();
      },
    );
  }
}

class TicketEventPage extends StatefulWidget {
  const TicketEventPage({super.key, required this.id});

  final int id;

  @override
  State<TicketEventPage> createState() => _TicketEventPageState();
}

class _TicketEventPageState extends State<TicketEventPage> {
  final _qty = <int, int>{};
  final _seats = <int, List<int>>{};
  int? _session;
  bool _wished = false;
  bool _busy = false;

  int _total(TicketEvent e) => [
        for (final t in e.ticketTypes ?? const <TicketTicketType>[]) (t.price ?? 0) * (_qty[t.id] ?? 0),
      ].fold(0, (a, b) => a + b);

  int get _count => _qty.values.fold(0, (a, b) => a + b);

  Future<void> _checkout(TicketEvent e) async {
    final api = context.api.ticket;
    setState(() => _busy = true);
    try {
      final res = await api.ticketReserve(
        input: TicketReserveInput(
          eventId: e.id,
          requestId: newIdempotencyKey(),
          buyerName: context.userName,
          items: [
            for (final entry in _qty.entries)
              if (entry.value > 0) TicketReserveItemInput(ticketTypeId: entry.key, quantity: entry.value, seatIds: _seats[entry.key]),
          ],
        ),
      );
      final orderId = pickInt(res, 'id');
      if (orderId == null) throw Exception('Không tạo được đơn giữ chỗ');
      if (!mounted) return;
      final total = pickInt(res, 'total') ?? _total(e);
      if (total > 0) {
        final pay = await showDialog<bool>(
          context: context,
          builder: (context) => AlertDialog(
            title: const Text('Thanh toán vé'),
            content: Text('Đơn #$orderId đã được giữ chỗ.\nThanh toán ${vnd(total)} bằng Ví Keiki?'),
            actions: [
              TextButton(onPressed: () => Navigator.pop(context, false), child: const Text('Để sau')),
              FilledButton(onPressed: () => Navigator.pop(context, true), child: const Text('Thanh toán')),
            ],
          ),
        );
        if (pay != true) {
          if (mounted) showSnack(context, 'Đơn đang được giữ chỗ, hãy thanh toán trong mục Đơn hàng');
          return;
        }
        await api.ticketPayOrder(id: orderId, method: 'wallet');
      }
      if (!mounted) return;
      setState(() {
        _qty.clear();
        _seats.clear();
      });
      await showResult(context, title: 'Đặt vé thành công', message: 'Vé đã nằm trong mục “Vé của tôi”.');
    } catch (err) {
      if (mounted) showSnack(context, errorMessage(err));
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final api = context.api.ticket;
    return AsyncBuilder<(TicketEvent?, List<TicketSession>)>(
      load: () async {
        final r = await Future.wait<Object?>([api.ticketEvent(id: widget.id), api.ticketSessions(id: widget.id).catchError((_) => null)]);
        return (r[0] as TicketEvent?, (r[1] as List<TicketSession>?) ?? const <TicketSession>[]);
      },
      loading: const Scaffold(body: SkeletonList()),
      builder: (context, d, reload) {
        final e = d.$1;
        if (e == null) return const Scaffold(body: EmptyView(message: 'Không tìm thấy sự kiện'));
        final sessions = d.$2.isNotEmpty ? d.$2 : (e.sessions ?? const <TicketSession>[]);
        final session = _session ?? (sessions.length > 1 ? sessions.first.id : null);
        final types = (e.ticketTypes ?? const <TicketTicketType>[]).where((t) => t.sessionId == null || session == null || t.sessionId == session).toList();
        return DetailPage(
          title: e.title ?? 'Sự kiện',
          actions: [
            IconButton(
              tooltip: 'Sơ đồ chỗ ngồi',
              icon: const Icon(Icons.event_seat_outlined),
              onPressed: () => openPage(context, SeatMapPage(eventId: e.id!, title: e.title ?? '')),
            ),
            IconButton(
              tooltip: _wished ? 'Bỏ yêu thích' : 'Yêu thích',
              icon: Icon(_wished ? Icons.favorite : Icons.favorite_border),
              onPressed: () async {
                final add = !_wished;
                if (await runAction(
                  context,
                  () => add ? api.ticketAddWishlist(id: e.id!) : api.ticketRemoveWishlist(id: e.id!),
                  success: add ? 'Đã thêm vào yêu thích' : 'Đã bỏ yêu thích',
                )) {
                  setState(() => _wished = add);
                }
              },
            ),
          ],
          header: HeroHeader(
            title: e.title ?? 'Sự kiện',
            imageUrl: e.bannerUrl,
            icon: Icons.local_activity_rounded,
            badges: [
              if (e.category != null) Pill(e.category!),
              if (e.featured == true) const Pill('Nổi bật', icon: Icons.star_rounded),
              if (e.soldOut == true) const StatusPill('sold_out'),
            ],
          ),
          bottom: Row(
            children: [
              Expanded(
                child: Column(
                  mainAxisSize: MainAxisSize.min,
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(_count == 0 ? 'Chọn loại vé' : '$_count vé', style: Theme.of(context).textTheme.bodySmall),
                    Text(vnd(_total(e)), style: Theme.of(context).textTheme.titleLarge?.copyWith(fontWeight: FontWeight.w800)),
                  ],
                ),
              ),
              SizedBox(
                width: 160,
                child: LoadingButton(label: 'Đặt vé', loading: _busy, onPressed: _count == 0 ? null : () => _checkout(e)),
              ),
            ],
          ),
          children: [
            InfoSection(rows: [
              ('Thời gian', formatDate(e.startsAt)),
              ('Kết thúc', formatDate(e.endsAt)),
              ('Địa điểm', e.venue),
              ('Địa chỉ', [e.address, e.city].whereType<String>().join(', ')),
              ('Hoàn vé', e.refundCutoffHours == null ? null : 'Trước ${e.refundCutoffHours} giờ'),
              ('Chuyển nhượng', e.transferable == null ? null : (e.transferable! ? 'Cho phép' : 'Không')),
            ]),
            if (sessions.length > 1) ...[
              const SectionHeader('Chọn suất diễn'),
              SizedBox(
                height: 48,
                child: ListView(
                  scrollDirection: Axis.horizontal,
                  padding: const EdgeInsets.symmetric(horizontal: 16),
                  children: [
                    for (final s in sessions)
                      Padding(
                        padding: const EdgeInsets.only(right: 8),
                        child: ChoiceChip(
                          label: Text(s.label ?? formatDate(s.startsAt)),
                          selected: session == s.id,
                          onSelected: s.status == 'cancelled'
                              ? null
                              : (_) => setState(() {
                                    _session = s.id;
                                    _qty.clear();
                                    _seats.clear();
                                  }),
                        ),
                      ),
                  ],
                ),
              ),
            ],
            const SectionHeader('Loại vé'),
            if (types.isEmpty) const Padding(padding: EdgeInsets.all(20), child: Text('Sự kiện chưa mở bán vé')),
            for (final t in types)
              Padding(
                padding: const EdgeInsets.fromLTRB(16, 0, 16, 10),
                child: KCard(
                  child: Row(
                    children: [
                      Expanded(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(t.name_ ?? 'Vé', style: const TextStyle(fontWeight: FontWeight.w700, fontSize: 15)),
                            Text(vnd(t.price), style: TextStyle(color: context.appColor, fontWeight: FontWeight.w800)),
                            if (t.description != null && t.description!.isNotEmpty) Text(t.description!, style: Theme.of(context).textTheme.bodySmall),
                            if (t.available != null) Text('Còn ${t.available} vé', style: Theme.of(context).textTheme.labelSmall),
                          ],
                        ),
                      ),
                      if (t.soldOut == true)
                        OutlinedButton.icon(
                          icon: const Icon(Icons.notifications_active_outlined, size: 18),
                          label: const Text('Chờ vé'),
                          onPressed: () => runAction(context, () => api.ticketJoinWaitlist(id: e.id!, type_: t.id!), success: 'Bạn sẽ được báo khi có vé trả lại'),
                        )
                      else if (t.onSale == false)
                        const StatusPill('closed')
                      else if (t.seated == true)
                        FilledButton.tonalIcon(
                          icon: const Icon(Icons.event_seat, size: 18),
                          label: Text((_seats[t.id] ?? const []).isEmpty ? 'Chọn ghế' : '${_seats[t.id]!.length} ghế'),
                          onPressed: () async {
                            final picked = await openPage<List<int>>(
                              context,
                              SeatPickerPage(eventId: e.id!, type: t, initial: _seats[t.id] ?? const [], max: t.maxPerOrder ?? 10),
                            );
                            if (picked != null) {
                              setState(() {
                                _seats[t.id!] = picked;
                                _qty[t.id!] = picked.length;
                              });
                            }
                          },
                        )
                      else
                        _Stepper(
                          value: _qty[t.id] ?? 0,
                          max: t.maxPerOrder ?? t.available ?? 10,
                          onChanged: (v) => setState(() => _qty[t.id!] = v),
                        ),
                    ],
                  ),
                ),
              ),
            Paragraph(e.description, title: 'Giới thiệu'),
            ResaleSection(eventId: e.id!, onBought: reload),
            SeriesStrip(eventId: e.id!, open: (id) => openPage(context, TicketEventPage(id: id))),
          ],
        );
      },
    );
  }
}

class _Stepper extends StatelessWidget {
  const _Stepper({required this.value, required this.max, required this.onChanged});

  final int value;
  final int max;
  final ValueChanged<int> onChanged;

  @override
  Widget build(BuildContext context) {
    return Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        IconButton.outlined(visualDensity: VisualDensity.compact, onPressed: value > 0 ? () => onChanged(value - 1) : null, icon: const Icon(Icons.remove, size: 18)),
        SizedBox(width: 28, child: Text('$value', textAlign: TextAlign.center, style: const TextStyle(fontWeight: FontWeight.w800, fontSize: 16))),
        IconButton.filled(visualDensity: VisualDensity.compact, onPressed: value < max ? () => onChanged(value + 1) : null, icon: const Icon(Icons.add, size: 18)),
      ],
    );
  }
}

class _MyTickets extends StatelessWidget {
  const _MyTickets();

  @override
  Widget build(BuildContext context) {
    return AsyncList<TicketTicket>(
      load: () async => (await context.api.ticket.ticketMyTickets(limit: 50))?.items,
      emptyMessage: 'Bạn chưa có vé nào',
      emptyIcon: Icons.confirmation_number_outlined,
      itemBuilder: (context, t, reload) => EntityTile(
        title: t.eventTitle ?? 'Vé #${t.id}',
        subtitle: [t.typeName, t.seat, formatDate(t.eventStartsAt)].whereType<String>().where((s) => s.isNotEmpty).join(' • '),
        icon: Icons.confirmation_number_rounded,
        badges: [StatusPill(t.status)],
        onTap: () async {
          await openPage(context, TicketDetailPage(ticket: t));
          await reload();
        },
      ),
    );
  }
}

class TicketDetailPage extends StatelessWidget {
  const TicketDetailPage({super.key, required this.ticket});

  final TicketTicket ticket;

  @override
  Widget build(BuildContext context) {
    final api = context.api.ticket;
    final t = ticket;
    final theme = Theme.of(context);
    return DetailPage(
      title: 'Vé điện tử',
      header: AppBanner(
        icon: Icons.confirmation_number_rounded,
        title: t.eventTitle ?? 'Vé',
        subtitle: [t.venue, formatDate(t.eventStartsAt)].whereType<String>().join(' • '),
        children: [
          Container(
            width: double.infinity,
            padding: const EdgeInsets.symmetric(vertical: 18),
            decoration: BoxDecoration(color: Colors.white, borderRadius: BorderRadius.circular(16)),
            child: Column(
              children: [
                Icon(Icons.qr_code_2_rounded, size: 120, color: theme.colorScheme.onSurface),
                const SizedBox(height: 6),
                SelectableText(t.code ?? '—', style: const TextStyle(fontSize: 20, fontWeight: FontWeight.w800, letterSpacing: 2, color: Colors.black87)),
                const Text('Đưa mã này cho nhân viên soát vé', style: TextStyle(color: Colors.black54, fontSize: 12)),
              ],
            ),
          ),
        ],
      ),
      children: [
        InfoSection(rows: [
          ('Trạng thái', t.status),
          ('Loại vé', t.typeName),
          ('Ghế', t.seat),
          ('Suất', t.sessionLabel),
          ('Người giữ vé', t.holderName),
          ('Email', t.holderEmail),
          ('Check-in lúc', t.checkedInAt == null ? null : formatDate(t.checkedInAt)),
        ]),
        ActionButtons([
          OutlinedButton.icon(
            icon: const Icon(Icons.badge_outlined),
            label: const Text('Đổi người giữ'),
            onPressed: () => showFormSheet(
              context,
              title: 'Người giữ vé',
              fields: [
                KField('name', 'Họ tên', required: true, initial: t.holderName),
                KField('email', 'Email', kind: FieldKind.email, initial: t.holderEmail),
              ],
              success: 'Đã cập nhật người giữ vé',
              onSubmit: (v) => api.ticketSetHolder(id: t.id!, name_: v.str('name')!, email: v.str('email')),
            ),
          ),
          OutlinedButton.icon(
            icon: const Icon(Icons.send_outlined),
            label: const Text('Tặng vé'),
            onPressed: () => showFormSheet(
              context,
              title: 'Chuyển vé cho bạn bè',
              fields: const [
                KField('email', 'Email người nhận', kind: FieldKind.email, required: true),
                KField('message', 'Lời nhắn', kind: FieldKind.multiline),
              ],
              success: 'Đã gửi lời mời nhận vé',
              onSubmit: (v) => api.ticketOfferTransfer(id: t.id!, toEmail: v.str('email'), message: v.str('message')),
            ),
          ),
          OutlinedButton.icon(
            icon: const Icon(Icons.sell_outlined),
            label: const Text('Bán lại'),
            onPressed: () => showFormSheet(
              context,
              title: 'Đăng bán lại vé',
              fields: const [KField('price', 'Giá bán (₫)', kind: FieldKind.number, required: true)],
              success: 'Đã đăng bán lại',
              onSubmit: (v) => api.ticketListForResale(id: t.id!, price: v.integer('price')!),
            ),
          ),
          OutlinedButton.icon(
            icon: const Icon(Icons.autorenew),
            label: const Text('Cấp lại mã'),
            onPressed: () => runAction(
              context,
              () => api.ticketReissue(id: t.id!),
              confirm: 'Mã vé cũ sẽ hết hiệu lực. Cấp mã mới?',
              success: 'Đã cấp mã mới, mở lại vé để xem',
            ),
          ),
        ]),
        TicketHistorySection(ticketId: t.id!),
      ],
    );
  }
}

class _Orders extends StatelessWidget {
  const _Orders();

  @override
  Widget build(BuildContext context) {
    final api = context.api.ticket;
    return AsyncList<TicketOrder>(
      load: () async => (await api.ticketOrders(limit: 50))?.items,
      emptyMessage: 'Chưa có đơn đặt vé',
      emptyIcon: Icons.receipt_long_outlined,
      itemBuilder: (context, o, reload) => EntityTile(
        title: 'Đơn #${o.id}',
        subtitle: '${formatDate(o.createdAt)}${o.tickets == null ? '' : ' • ${o.tickets!.length} vé'}',
        caption: vnd(o.total),
        icon: Icons.receipt_long_rounded,
        badges: [StatusPill(o.status)],
        trailing: o.status == 'pending' || o.status == 'reserved'
            ? FilledButton(
                onPressed: () async {
                  if (await runAction(context, () => api.ticketPayOrder(id: o.id!, method: 'wallet'), success: 'Thanh toán thành công')) await reload();
                },
                child: const Text('Trả'),
              )
            : null,
        onTap: () async {
          await openPage(context, TicketOrderDetailPage(id: o.id!));
          await reload();
        },
      ),
    );
  }
}
