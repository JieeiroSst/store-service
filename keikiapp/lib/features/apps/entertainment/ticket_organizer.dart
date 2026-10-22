import 'package:flutter/material.dart';

import '../../../gen/models.g.dart';
import '../kit/kit.dart';

const _statusNames = {
  'draft': 'Bản nháp',
  'pending_review': 'Chờ duyệt',
  'published': 'Đang bán',
  'rejected': 'Bị từ chối',
  'cancelled': 'Đã huỷ',
};

Widget _eventStatus(String? s) => s == null ? const SizedBox.shrink() : StatusPill(_statusNames[s] == null ? s : _eventStatusKey(s));

String _eventStatusKey(String s) => switch (s) {
      'pending_review' => 'review',
      _ => s,
    };

List<KField> _eventFields([TicketEvent? e]) => [
      KField('title', 'Tên sự kiện', required: true, initial: e?.title),
      KField('category', 'Thể loại', initial: e?.category, hint: 'Âm nhạc, Hội thảo, Thể thao…'),
      KField('starts', 'Bắt đầu', kind: FieldKind.datetime, required: true, initial: e?.startsAt),
      KField('ends', 'Kết thúc', kind: FieldKind.datetime, required: true, initial: e?.endsAt),
      KField('venue', 'Địa điểm', required: true, initial: e?.venue),
      KField('address', 'Địa chỉ', initial: e?.address),
      KField('city', 'Thành phố', initial: e?.city),
      KField('banner', 'Ảnh bìa (URL)', initial: e?.bannerUrl),
      KField('refund', 'Hoàn vé trước (giờ)', kind: FieldKind.number, initial: '${e?.refundCutoffHours ?? 48}'),
      KField('resale', 'Trần giá bán lại (% giá gốc)', kind: FieldKind.number, initial: '${e?.resaleCapPercent ?? 110}'),
      KField('transferable', 'Cho phép chuyển nhượng', kind: FieldKind.toggle, initial: '${e?.transferable ?? true}'),
      KField('description', 'Giới thiệu', kind: FieldKind.multiline, initial: e?.description),
    ];

TicketEventInput _eventInput(FormData v) => TicketEventInput(
      title: v.str('title'),
      category: v.str('category'),
      startsAt: v.str('starts'),
      endsAt: v.str('ends'),
      venue: v.str('venue'),
      address: v.str('address'),
      city: v.str('city'),
      bannerUrl: v.str('banner'),
      refundCutoffHours: v.integer('refund'),
      resaleCapPercent: v.integer('resale'),
      transferable: v.flag('transferable'),
      description: v.str('description'),
      currency: 'VND',
    );

class TicketOrganizerPage extends StatelessWidget {
  const TicketOrganizerPage({super.key});

  @override
  Widget build(BuildContext context) {
    final color = context.appColor;
    return DefaultTabController(
      length: 4,
      child: Scaffold(
        appBar: AppBar(
          foregroundColor: Colors.white,
          backgroundColor: color,
          flexibleSpace: DecoratedBox(decoration: BoxDecoration(gradient: appGradient(color))),
          title: const Text('Ban tổ chức', style: TextStyle(color: Colors.white, fontWeight: FontWeight.w700)),
          bottom: const TabBar(
            isScrollable: true,
            tabAlignment: TabAlignment.start,
            labelColor: Colors.white,
            unselectedLabelColor: Colors.white70,
            indicatorColor: Colors.white,
            dividerColor: Colors.transparent,
            tabs: [Tab(text: 'Sự kiện của tôi'), Tab(text: 'Địa điểm'), Tab(text: 'Soát vé'), Tab(text: 'Kiểm duyệt')],
          ),
        ),
        floatingActionButton: Builder(
          builder: (context) => FloatingActionButton.extended(
            backgroundColor: color,
            foregroundColor: Colors.white,
            icon: const Icon(Icons.add),
            label: const Text('Tạo sự kiện'),
            onPressed: () => showFormSheet(
              context,
              title: 'Sự kiện mới',
              description: 'Sự kiện được lưu dưới dạng nháp. Thêm loại vé rồi gửi duyệt để mở bán.',
              fields: _eventFields(),
              submitLabel: 'Tạo nháp',
              success: 'Đã tạo sự kiện, kéo xuống để làm mới',
              onSubmit: (v) => context.api.ticket.ticketCreateEvent(input: _eventInput(v)),
            ),
          ),
        ),
        body: const TabBarView(children: [_MyEvents(), _Venues(), _Scanner(), _Moderation()]),
      ),
    );
  }
}

class _MyEvents extends StatelessWidget {
  const _MyEvents();

  @override
  Widget build(BuildContext context) {
    return AsyncList<TicketEvent>(
      load: () async => (await context.api.ticket.ticketMyEvents(limit: 50))?.items,
      emptyMessage: 'Bạn chưa tạo sự kiện nào',
      emptyIcon: Icons.event_outlined,
      itemBuilder: (context, e, reload) => EntityTile(
        title: e.title ?? 'Sự kiện',
        subtitle: '${formatDate(e.startsAt)} • ${e.venue ?? ''}',
        imageUrl: e.bannerUrl,
        icon: Icons.local_activity_rounded,
        badges: [_eventStatus(e.status)],
        onTap: () async {
          await openPage(context, ManageEventPage(id: e.id!));
          await reload();
        },
      ),
    );
  }
}

class ManageEventPage extends StatefulWidget {
  const ManageEventPage({super.key, required this.id});

  final int id;

  @override
  State<ManageEventPage> createState() => _ManageEventPageState();
}

class _ManageEventPageState extends State<ManageEventPage> {
  int _v = 0;

  void _bump() => setState(() => _v++);

  Future<void> _act(Future<Object?> Function() call, String success, {String? confirm, bool destructive = false}) async {
    if (await runAction(context, call, success: success, confirm: confirm, destructive: destructive)) _bump();
  }

  @override
  Widget build(BuildContext context) {
    final api = context.api.ticket;
    final id = widget.id;
    return AsyncBuilder<(TicketEvent?, TicketReport?, List<TicketSession>, List<TicketPromotion>, List<TicketStaff>)>(
      key: ValueKey(_v),
      load: () async {
        final r = await Future.wait<Object?>([
          api.ticketEvent(id: id),
          api.ticketEventReport(id: id).catchError((_) => null),
          api.ticketSessions(id: id).catchError((_) => null),
          api.ticketPromotions(id: id).catchError((_) => null),
          api.ticketStaff(id: id).catchError((_) => null),
        ]);
        return (
          r[0] as TicketEvent?,
          r[1] as TicketReport?,
          (r[2] as List<TicketSession>?) ?? const <TicketSession>[],
          (r[3] as List<TicketPromotion>?) ?? const <TicketPromotion>[],
          (r[4] as List<TicketStaff>?) ?? const <TicketStaff>[],
        );
      },
      loading: const Scaffold(body: SkeletonList()),
      builder: (context, d, _) {
        final e = d.$1;
        if (e == null) return const Scaffold(body: EmptyView(message: 'Không tìm thấy sự kiện'));
        final report = d.$2;
        final types = e.ticketTypes ?? const <TicketTicketType>[];
        final draft = e.status == 'draft' || e.status == 'rejected';
        return DetailPage(
          title: e.title ?? 'Sự kiện',
          actions: [
            PopupMenuButton<String>(
              onSelected: (a) async {
                switch (a) {
                  case 'edit':
                    if (await showFormSheet(context, title: 'Sửa sự kiện', fields: _eventFields(e), success: 'Đã lưu', onSubmit: (v) => api.ticketUpdateEvent(id: id, input: _eventInput(v)))) _bump();
                  case 'dup':
                    await showFormSheet(
                      context,
                      title: 'Nhân bản sự kiện',
                      fields: [
                        KField('title', 'Tên', required: true, initial: '${e.title ?? ''} (bản sao)'),
                        const KField('starts', 'Bắt đầu', kind: FieldKind.datetime, required: true),
                        const KField('ends', 'Kết thúc', kind: FieldKind.datetime, required: true),
                        const KField('promos', 'Chép mã khuyến mãi', kind: FieldKind.toggle, initial: 'true'),
                      ],
                      success: 'Đã tạo bản sao trong Sự kiện của tôi',
                      onSubmit: (v) => api.ticketDuplicateEvent(id: id, input: TicketDuplicateInput(title: v.str('title'), startsAt: v.str('starts'), endsAt: v.str('ends'), copyPromotions: v.flag('promos'))),
                    );
                  case 'cancel':
                    if (await showFormSheet(
                      context,
                      title: 'Huỷ sự kiện',
                      description: 'Mọi đơn đã thanh toán sẽ được hoàn tiền.',
                      fields: const [KField('reason', 'Lý do', kind: FieldKind.multiline, required: true)],
                      submitLabel: 'Huỷ sự kiện',
                      success: 'Đã huỷ sự kiện',
                      onSubmit: (v) => api.ticketCancelEvent(id: id, reason: v.str('reason')),
                    )) {
                      _bump();
                    }
                }
              },
              itemBuilder: (_) => [
                const PopupMenuItem(value: 'edit', child: Text('Sửa thông tin')),
                const PopupMenuItem(value: 'dup', child: Text('Nhân bản')),
                if (e.status != 'cancelled') const PopupMenuItem(value: 'cancel', child: Text('Huỷ sự kiện')),
              ],
            ),
          ],
          header: AppBanner(
            icon: Icons.insights_rounded,
            title: formatMoney(report?.revenue, e.currency),
            subtitle: 'Doanh thu • ${_statusNames[e.status] ?? e.status ?? ''}',
            children: [
              StatRow([
                StatTile(label: 'Đơn', value: '${report?.orders ?? 0}', onDark: true),
                StatTile(label: 'Vé bán', value: '${report?.ticketsSold ?? 0}', onDark: true),
                StatTile(label: 'Check-in', value: '${report?.checkedIn ?? 0}', onDark: true),
              ]),
              if (draft) ...[
                const SizedBox(height: 12),
                SizedBox(
                  width: double.infinity,
                  child: FilledButton.icon(
                    style: FilledButton.styleFrom(backgroundColor: Colors.white, foregroundColor: context.appColor),
                    icon: const Icon(Icons.send),
                    label: const Text('Gửi duyệt để mở bán'),
                    onPressed: types.isEmpty ? null : () => _act(() => api.ticketSubmitEvent(id: id), 'Đã gửi duyệt'),
                  ),
                ),
              ],
            ],
          ),
          children: [
            if (report != null && (report.byType ?? const []).isNotEmpty)
              InfoSection(title: 'Bán theo loại vé', rows: [
                for (final t in report.byType!) (t.name_ ?? 'Vé', '${t.sold ?? 0}/${t.total ?? 0} • ${formatMoney(t.revenue, e.currency)}'),
              ]),
            SectionHeader('Suất diễn', action: 'Thêm', onAction: () async {
              if (await showFormSheet(
                context,
                title: 'Suất diễn mới',
                fields: const [
                  KField('label', 'Tên suất', hint: 'Tối thứ Bảy'),
                  KField('starts', 'Bắt đầu', kind: FieldKind.datetime, required: true),
                  KField('ends', 'Kết thúc', kind: FieldKind.datetime, required: true),
                ],
                success: 'Đã thêm suất diễn',
                onSubmit: (v) => api.ticketCreateSession(id: id, input: TicketSessionInput(label: v.str('label'), startsAt: v.str('starts'), endsAt: v.str('ends'))),
              )) {
                _bump();
              }
            }),
            if (d.$3.isEmpty) const Padding(padding: EdgeInsets.symmetric(horizontal: 20), child: Text('Sự kiện chỉ có một suất')),
            for (final s in d.$3)
              Padding(
                padding: const EdgeInsets.fromLTRB(16, 0, 16, 8),
                child: EntityTile(
                  title: s.label ?? formatDate(s.startsAt),
                  subtitle: '${formatDate(s.startsAt)} – ${formatDate(s.endsAt)}',
                  icon: Icons.schedule,
                  badges: [StatusPill(s.status)],
                  trailing: PopupMenuButton<String>(
                    onSelected: (a) async {
                      if (a == 'edit') {
                        if (await showFormSheet(
                          context,
                          title: 'Sửa suất diễn',
                          fields: [
                            KField('label', 'Tên suất', initial: s.label),
                            KField('starts', 'Bắt đầu', kind: FieldKind.datetime, required: true, initial: s.startsAt),
                            KField('ends', 'Kết thúc', kind: FieldKind.datetime, required: true, initial: s.endsAt),
                          ],
                          success: 'Đã lưu',
                          onSubmit: (v) => api.ticketUpdateSession(id: id, session: s.id!, input: TicketSessionInput(label: v.str('label'), startsAt: v.str('starts'), endsAt: v.str('ends'))),
                        )) {
                          _bump();
                        }
                      } else if (a == 'cancel') {
                        if (!context.mounted) return;
                        if (await showFormSheet(
                          context,
                          title: 'Huỷ suất diễn',
                          fields: const [KField('reason', 'Lý do', required: true)],
                          success: 'Đã huỷ suất, vé của suất này được hoàn tiền',
                          onSubmit: (v) => api.ticketCancelSession(id: id, session: s.id!, reason: v.str('reason')),
                        )) {
                          _bump();
                        }
                      } else {
                        await _act(() => api.ticketDeleteSession(id: id, session: s.id!), 'Đã xoá suất', confirm: 'Xoá suất chưa bán vé này?', destructive: true);
                      }
                    },
                    itemBuilder: (_) => const [
                      PopupMenuItem(value: 'edit', child: Text('Sửa')),
                      PopupMenuItem(value: 'cancel', child: Text('Huỷ suất')),
                      PopupMenuItem(value: 'delete', child: Text('Xoá')),
                    ],
                  ),
                ),
              ),
            SectionHeader('Loại vé', action: 'Thêm', onAction: () async {
              if (await _typeForm(context, null, d.$3, (input) => api.ticketCreateTicketType(id: id, input: input))) _bump();
            }),
            for (final t in types)
              Padding(
                padding: const EdgeInsets.fromLTRB(16, 0, 16, 8),
                child: EntityTile(
                  title: t.name_ ?? 'Vé',
                  subtitle: 'Đã bán ${t.sold ?? 0}/${t.total ?? 0} • giữ ${t.held ?? 0}',
                  caption: vnd(t.price),
                  icon: t.seated == true ? Icons.event_seat_rounded : Icons.confirmation_number_rounded,
                  badges: [if (t.active == false) const StatusPill('inactive'), if (t.seated == true) const Pill('Ghế đánh số')],
                  trailing: PopupMenuButton<String>(
                    onSelected: (a) async {
                      switch (a) {
                        case 'edit':
                          if (await _typeForm(context, t, d.$3, (input) => api.ticketUpdateTicketType(id: id, type_: t.id!, input: input))) _bump();
                        case 'seats':
                          if (await showFormSheet(
                            context,
                            title: 'Sinh ghế cho ${t.name_}',
                            fields: const [
                              KField('section', 'Khu', required: true, initial: 'A'),
                              KField('rows', 'Số hàng', kind: FieldKind.number, required: true, initial: '5'),
                              KField('per', 'Ghế mỗi hàng', kind: FieldKind.number, required: true, initial: '10'),
                              KField('first', 'Hàng đầu tiên', initial: 'A'),
                            ],
                            success: 'Đã sinh ghế',
                            onSubmit: (v) => api.ticketGenerateSeats(id: id, type_: t.id!, input: {
                              'section': v.str('section'),
                              'rows': v.integer('rows'),
                              'seats_per_row': v.integer('per'),
                              'first_row': v.str('first') ?? 'A',
                            }),
                          )) {
                            _bump();
                          }
                        case 'map':
                          if (await showFormSheet(
                            context,
                            title: 'Sơ đồ ghế thủ công',
                            description: 'Mỗi hàng một dòng: tên hàng rồi số ghế cuối, ví dụ "A 12".',
                            fields: const [KField('rows', 'Các hàng', kind: FieldKind.multiline, required: true, initial: 'A 10\nB 10\nC 12')],
                            success: 'Đã lưu sơ đồ',
                            onSubmit: (v) {
                              final seats = <Map<String, Object>>[];
                              for (final (y, line) in (v.str('rows') ?? '').split('\n').where((l) => l.trim().isNotEmpty).indexed) {
                                final parts = line.trim().split(RegExp(r'\s+'));
                                final count = int.tryParse(parts.length > 1 ? parts[1] : '') ?? 0;
                                for (var n = 1; n <= count; n++) {
                                  seats.add({'row': parts[0], 'number': n, 'x': n * 20.0, 'y': y * 24.0});
                                }
                              }
                              return api.ticketSetSeatMap(id: id, type_: t.id!, input: {'width': 20.0 * 30, 'height': 24.0 * 12, 'seats': seats});
                            },
                          )) {
                            _bump();
                          }
                        case 'invite':
                          await showFormSheet(
                            context,
                            title: 'Mời khách (vé mời)',
                            fields: const [
                              KField('name', 'Tên khách', required: true),
                              KField('email', 'Email', kind: FieldKind.email, required: true),
                              KField('qty', 'Số vé', kind: FieldKind.number, initial: '1'),
                              KField('note', 'Ghi chú'),
                            ],
                            success: 'Đã gửi vé mời',
                            onSubmit: (v) => api.ticketInvite(
                              id: id,
                              input: TicketInviteInput(ticketTypeId: t.id, quantity: v.integer('qty') ?? 1, name_: v.str('name'), email: v.str('email'), note: v.str('note')),
                            ),
                          );
                      }
                    },
                    itemBuilder: (_) => [
                      const PopupMenuItem(value: 'edit', child: Text('Sửa')),
                      if (t.seated == true) ...[
                        const PopupMenuItem(value: 'seats', child: Text('Sinh ghế theo hàng')),
                        const PopupMenuItem(value: 'map', child: Text('Nhập sơ đồ ghế')),
                      ],
                      const PopupMenuItem(value: 'invite', child: Text('Gửi vé mời')),
                    ],
                  ),
                ),
              ),
            if (types.any((t) => t.seated == true))
              ActionButtons([
                OutlinedButton.icon(
                  icon: const Icon(Icons.map_outlined),
                  label: const Text('Áp sơ đồ từ địa điểm'),
                  onPressed: () => _applyVenue(context, id, types),
                ),
              ]),
            SectionHeader('Mã khuyến mãi', action: 'Thêm', onAction: () async {
              if (await _promoForm(context, null, (input) => api.ticketCreatePromotion(id: id, input: input))) _bump();
            }),
            for (final p in d.$4)
              Padding(
                padding: const EdgeInsets.fromLTRB(16, 0, 16, 8),
                child: EntityTile(
                  title: p.code ?? '',
                  subtitle: '${p.kind == 'fixed' ? 'Giảm ${vnd(p.value)}' : 'Giảm ${p.value}%'} • đã dùng ${p.used ?? 0}/${p.maxUses ?? '∞'}',
                  caption: p.validTo == null ? null : 'Hết hạn ${formatDate(p.validTo)}',
                  icon: Icons.local_offer_rounded,
                  badges: [if (p.active == false) const StatusPill('inactive')],
                  onTap: () async {
                    if (await _promoForm(context, p, (input) => api.ticketUpdatePromotion(id: id, promo: p.id!, input: input))) _bump();
                  },
                ),
              ),
            SectionHeader('Nhân viên soát vé', action: 'Thêm', onAction: () async {
              if (await showFormSheet(
                context,
                title: 'Thêm nhân viên',
                fields: const [KField('user', 'Mã người dùng', kind: FieldKind.number, required: true)],
                success: 'Đã thêm nhân viên',
                onSubmit: (v) => api.ticketAddStaff(id: id, userId: v.integer('user')!),
              )) {
                _bump();
              }
            }),
            for (final s in d.$5)
              ListTile(
                contentPadding: const EdgeInsets.symmetric(horizontal: 20),
                leading: const CircleAvatar(child: Icon(Icons.badge_outlined)),
                title: Text('Người dùng #${s.userId}'),
                subtitle: Text('Thêm lúc ${formatDate(s.createdAt)}'),
                trailing: IconButton(
                  tooltip: 'Gỡ',
                  icon: const Icon(Icons.person_remove_outlined),
                  onPressed: () => _act(() => api.ticketRemoveStaff(id: id, user: s.userId!), 'Đã gỡ nhân viên', confirm: 'Gỡ nhân viên #${s.userId}?', destructive: true),
                ),
              ),
            ActionButtons([
              OutlinedButton.icon(icon: const Icon(Icons.receipt_long_outlined), label: const Text('Đơn hàng'), onPressed: () => openPage(context, _EventOrders(id: id))),
              OutlinedButton.icon(icon: const Icon(Icons.people_outline), label: const Text('Khách tham dự'), onPressed: () => openPage(context, _Attendees(id: id))),
            ]),
          ],
        );
      },
    );
  }
}

Future<bool> _typeForm(BuildContext context, TicketTicketType? t, List<TicketSession> sessions, Future<Object?> Function(TicketTypeInput) save) => showFormSheet(
      context,
      title: t == null ? 'Loại vé mới' : 'Sửa ${t.name_}',
      fields: [
        KField('name', 'Tên loại vé', required: true, initial: t?.name_),
        KField('price', 'Giá (₫)', kind: FieldKind.number, required: true, initial: t?.price?.toString()),
        KField('total', 'Số lượng', kind: FieldKind.number, required: true, initial: t?.total?.toString()),
        KField('max', 'Tối đa mỗi đơn', kind: FieldKind.number, initial: '${t?.maxPerOrder ?? 6}'),
        KField('maxUser', 'Tối đa mỗi người', kind: FieldKind.number, initial: t?.maxPerUser?.toString()),
        KField('saleStart', 'Mở bán', kind: FieldKind.datetime, initial: t?.saleStartsAt),
        KField('saleEnd', 'Đóng bán', kind: FieldKind.datetime, initial: t?.saleEndsAt),
        if (sessions.isNotEmpty)
          KField('session', 'Suất diễn', kind: FieldKind.choice, initial: t?.sessionId?.toString() ?? '', options: [
            ('', 'Mọi suất'),
            for (final s in sessions) ('${s.id}', s.label ?? formatDate(s.startsAt)),
          ]),
        KField('seated', 'Ghế đánh số', kind: FieldKind.toggle, initial: '${t?.seated ?? false}'),
        KField('active', 'Đang mở bán', kind: FieldKind.toggle, initial: '${t?.active ?? true}'),
        KField('description', 'Mô tả', kind: FieldKind.multiline, initial: t?.description),
      ],
      success: 'Đã lưu loại vé',
      onSubmit: (v) => save(TicketTypeInput(
        name_: v.str('name'),
        price: v.integer('price'),
        total: v.integer('total'),
        maxPerOrder: v.integer('max'),
        maxPerUser: v.integer('maxUser'),
        saleStartsAt: v.str('saleStart'),
        saleEndsAt: v.str('saleEnd'),
        sessionId: v.integer('session'),
        seated: v.flag('seated'),
        active: v.flag('active'),
        description: v.str('description'),
      )),
    );

Future<bool> _promoForm(BuildContext context, TicketPromotion? p, Future<Object?> Function(TicketPromotionInput) save) => showFormSheet(
      context,
      title: p == null ? 'Mã khuyến mãi mới' : 'Sửa ${p.code}',
      fields: [
        KField('code', 'Mã', required: true, initial: p?.code),
        KField('kind', 'Kiểu giảm', kind: FieldKind.choice, initial: p?.kind ?? 'percent', options: const [('percent', 'Theo %'), ('fixed', 'Số tiền cố định')]),
        KField('value', 'Giá trị', kind: FieldKind.number, required: true, initial: p?.value?.toString()),
        KField('max', 'Số lượt tối đa', kind: FieldKind.number, initial: p?.maxUses?.toString()),
        KField('min', 'Tối thiểu số vé', kind: FieldKind.number, initial: p?.minTickets?.toString()),
        KField('from', 'Bắt đầu', kind: FieldKind.datetime, initial: p?.validFrom),
        KField('to', 'Kết thúc', kind: FieldKind.datetime, initial: p?.validTo),
        KField('active', 'Đang hoạt động', kind: FieldKind.toggle, initial: '${p?.active ?? true}'),
      ],
      success: 'Đã lưu mã khuyến mãi',
      onSubmit: (v) => save(TicketPromotionInput(
        code: v.str('code')?.toUpperCase(),
        kind: v.str('kind'),
        value: v.integer('value'),
        maxUses: v.integer('max'),
        minTickets: v.integer('min'),
        validFrom: v.str('from'),
        validTo: v.str('to'),
        active: v.flag('active'),
      )),
    );

Future<void> _applyVenue(BuildContext context, int eventId, List<TicketTicketType> types) async {
  final api = context.api.ticket;
  final venues = (await api.ticketVenues(limit: 50).catchError((_) => null))?.items ?? const <TicketVenue>[];
  final shared = (await api.ticketVenues(scope: 'shared', limit: 50).catchError((_) => null))?.items ?? const <TicketVenue>[];
  final all = [...venues, ...shared.where((s) => !venues.any((v) => v.id == s.id))];
  if (!context.mounted) return;
  if (all.isEmpty) {
    showSnack(context, 'Chưa có địa điểm nào, tạo ở tab Địa điểm trước');
    return;
  }
  final seated = types.where((t) => t.seated == true).toList();
  await showFormSheet(
    context,
    title: 'Áp sơ đồ địa điểm',
    description: 'Mỗi dòng: tên khu = mã loại vé. Loại vé có ghế: ${seated.map((t) => '${t.name_} (#${t.id})').join(', ')}',
    fields: [
      KField('venue', 'Địa điểm', kind: FieldKind.choice, options: [for (final v in all) ('${v.id}', '${v.name_ ?? ''} • ${v.seats ?? 0} ghế')]),
      KField('map', 'Gán khu', kind: FieldKind.multiline, required: true, initial: seated.isEmpty ? '' : 'A = ${seated.first.id}'),
      const KField('replace', 'Thay ghế hiện có', kind: FieldKind.toggle, initial: 'false'),
    ],
    success: 'Đã áp sơ đồ',
    onSubmit: (v) => api.ticketApplyVenue(id: eventId, input: {
      'venue_id': v.integer('venue'),
      'replace': v.flag('replace'),
      'assignments': [
        for (final line in (v.str('map') ?? '').split('\n'))
          if (line.contains('='))
            {'section': line.split('=').first.trim(), 'ticket_type_id': int.tryParse(line.split('=').last.trim())},
      ],
    }),
  );
}

class _EventOrders extends StatelessWidget {
  const _EventOrders({required this.id});

  final int id;

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Đơn của sự kiện')),
      body: AsyncList<TicketOrder>(
        load: () async => (await context.api.ticket.ticketEventOrders(id: id, limit: 100))?.items,
        emptyMessage: 'Chưa có đơn',
        itemBuilder: (context, o, _) => EntityTile(
          title: '${o.buyerName ?? 'Khách'} • đơn #${o.id}',
          subtitle: '${o.buyerEmail ?? ''} • ${formatDate(o.createdAt)}',
          caption: formatMoney(o.total, o.currency),
          icon: Icons.receipt_long_rounded,
          badges: [StatusPill(o.status)],
        ),
      ),
    );
  }
}

class _Attendees extends StatelessWidget {
  const _Attendees({required this.id});

  final int id;

  @override
  Widget build(BuildContext context) {
    final api = context.api.ticket;
    return Scaffold(
      appBar: AppBar(title: const Text('Khách tham dự')),
      body: AsyncList<TicketAttendee>(
        load: () async => (await api.ticketAttendees(id: id, limit: 200))?.items,
        emptyMessage: 'Chưa có khách',
        spacing: 6,
        itemBuilder: (context, a, reload) => EntityTile(
          title: a.holderName ?? a.buyerName ?? 'Khách',
          subtitle: '${a.typeName ?? ''}${a.seat == null ? '' : ' • ghế ${a.seat}'} • ${a.code ?? ''}',
          icon: a.checkedInAt == null ? Icons.person_outline : Icons.how_to_reg_rounded,
          badges: [StatusPill(a.status)],
          trailing: a.status == 'void' || a.status == 'voided'
              ? null
              : IconButton(
                  tooltip: 'Vô hiệu vé',
                  icon: const Icon(Icons.block),
                  onPressed: () async {
                    if (await showFormSheet(
                      context,
                      title: 'Vô hiệu vé ${a.code}',
                      fields: const [KField('reason', 'Lý do', required: true)],
                      success: 'Đã vô hiệu vé',
                      onSubmit: (v) => api.ticketVoid(id: a.id!, reason: v.str('reason')),
                    )) {
                      await reload();
                    }
                  },
                ),
        ),
      ),
    );
  }
}

class _Venues extends StatefulWidget {
  const _Venues();

  @override
  State<_Venues> createState() => _VenuesState();
}

class _VenuesState extends State<_Venues> {
  bool _shared = false;
  int _v = 0;

  Future<void> _create(BuildContext context, [TicketVenue? existing]) async {
    final api = context.api.ticket;
    final templates = await api.ticketVenueTemplates().catchError((_) => null) ?? const <TicketTemplate>[];
    if (!context.mounted) return;
    String? template = templates.isEmpty ? null : templates.first.name_;
    final params = templates.isEmpty ? const <TicketTemplateParam>[] : templates.first.params ?? const <TicketTemplateParam>[];
    final ok = await showFormSheet(
      context,
      title: existing == null ? 'Địa điểm mới' : 'Sửa ${existing.name_}',
      description: templates.isEmpty ? null : 'Mẫu sơ đồ: ${templates.map((t) => '${t.name_} (${t.description ?? ''})').join('; ')}',
      fields: [
        KField('name', 'Tên địa điểm', required: true, initial: existing?.name_),
        KField('city', 'Thành phố', initial: existing?.city),
        KField('address', 'Địa chỉ', initial: existing?.address),
        KField('description', 'Mô tả', kind: FieldKind.multiline, initial: existing?.description),
        if (templates.isNotEmpty) KField('template', 'Mẫu sơ đồ', kind: FieldKind.choice, initial: template, options: [for (final t in templates) (t.name_ ?? '', t.name_ ?? '')]),
        for (final p in params) KField('p_${p.name_}', '${p.description ?? p.name_} (${p.min}–${p.max})', kind: FieldKind.number, initial: '${p.default_ ?? ''}'),
      ],
      submitLabel: existing == null ? 'Xem trước & tạo' : 'Lưu',
      onSubmit: (v) async {
        template = v.str('template');
        final input = {
          'name': v.str('name'),
          'city': v.str('city'),
          'address': v.str('address'),
          'description': v.str('description'),
          'template': ?template,
          'params': {for (final p in params) if (v.integer('p_${p.name_}') != null) p.name_: v.integer('p_${p.name_}')},
        };
        if (existing == null) {
          final preview = await api.ticketPreviewVenue(input: input);
          final total = pickInt(preview, 'total');
          if (total == 0) throw Exception('Sơ đồ không có ghế nào, hãy chỉnh lại thông số');
          await api.ticketCreateVenue(input: input);
        } else {
          await api.ticketUpdateVenue(id: existing.id!, input: input);
        }
      },
    );
    if (ok && mounted) {
      setState(() => _v++);
      if (context.mounted) showSnack(context, 'Đã lưu địa điểm');
    }
  }

  @override
  Widget build(BuildContext context) {
    final api = context.api.ticket;
    return Column(
      children: [
        Padding(
          padding: const EdgeInsets.fromLTRB(16, 12, 16, 0),
          child: Row(
            children: [
              Expanded(
                child: SegmentedButton<bool>(
                  showSelectedIcon: false,
                  segments: const [ButtonSegment(value: false, label: Text('Của tôi')), ButtonSegment(value: true, label: Text('Được chia sẻ'))],
                  selected: {_shared},
                  onSelectionChanged: (s) => setState(() => _shared = s.first),
                ),
              ),
              const SizedBox(width: 8),
              IconButton.filled(tooltip: 'Thêm địa điểm', onPressed: () => _create(context), icon: const Icon(Icons.add_location_alt_outlined)),
            ],
          ),
        ),
        Expanded(
          child: AsyncList<TicketVenue>(
            key: ValueKey('$_shared$_v'),
            load: () async => (await api.ticketVenues(scope: _shared ? 'shared' : 'mine', limit: 50))?.items,
            emptyMessage: 'Chưa có địa điểm',
            emptyIcon: Icons.location_city_outlined,
            itemBuilder: (context, v, reload) => EntityTile(
              title: v.name_ ?? 'Địa điểm',
              subtitle: [v.address, v.city].whereType<String>().join(', '),
              caption: '${v.seats ?? 0} ghế • ${v.sections ?? 0} khu',
              icon: Icons.stadium_rounded,
              badges: [if (v.shared == true) const Pill('Đang chia sẻ', icon: Icons.share)],
              onTap: () => openPage(context, _VenueDetail(id: v.id!)),
              trailing: PopupMenuButton<String>(
                onSelected: (a) async {
                  var ok = false;
                  if (a == 'edit') {
                    await _create(context, v);
                  } else if (a == 'copy') {
                    ok = await runAction(context, () => api.ticketCopyVenue(id: v.id!), success: 'Đã sao chép về danh sách của bạn');
                  } else if (a == 'share') {
                    ok = await runAction(context, () => api.ticketShareVenue(id: v.id!, input: {'shared': v.shared != true}), success: v.shared == true ? 'Đã ngừng chia sẻ' : 'Đã chia sẻ cho các ban tổ chức khác');
                  } else {
                    ok = await runAction(context, () => api.ticketDeleteVenue(id: v.id!), confirm: 'Xoá địa điểm ${v.name_}?', destructive: true, success: 'Đã xoá');
                  }
                  if (ok) await reload();
                },
                itemBuilder: (_) => [
                  if (!_shared) const PopupMenuItem(value: 'edit', child: Text('Sửa')),
                  const PopupMenuItem(value: 'copy', child: Text('Sao chép')),
                  if (!_shared) PopupMenuItem(value: 'share', child: Text(v.shared == true ? 'Ngừng chia sẻ' : 'Chia sẻ')),
                  if (!_shared) const PopupMenuItem(value: 'delete', child: Text('Xoá')),
                ],
              ),
            ),
          ),
        ),
      ],
    );
  }
}

class _VenueDetail extends StatelessWidget {
  const _VenueDetail({required this.id});

  final int id;

  @override
  Widget build(BuildContext context) {
    return AsyncBuilder<TicketVenueDetail?>(
      load: () => context.api.ticket.ticketVenue(id: id, seats: true),
      loading: const Scaffold(body: SkeletonList()),
      builder: (context, d, _) {
        final v = d?.venue;
        return DetailPage(
          title: v?.name_ ?? 'Địa điểm',
          children: [
            InfoSection(rows: [
              ('Địa chỉ', [v?.address, v?.city].whereType<String>().join(', ')),
              ('Tổng ghế', '${d?.preview?.total ?? v?.seats ?? 0}'),
              ('Tạo lúc', formatDate(v?.createdAt)),
            ]),
            InfoSection(title: 'Các khu', rows: [
              for (final s in d?.preview?.sections ?? const <TicketSectionCount>[]) (s.name_ ?? s.key ?? '', '${s.seats ?? 0} ghế'),
            ]),
            Paragraph(v?.description),
          ],
        );
      },
    );
  }
}

class _Scanner extends StatefulWidget {
  const _Scanner();

  @override
  State<_Scanner> createState() => _ScannerState();
}

class _ScannerState extends State<_Scanner> {
  TicketEvent? _event;
  final _code = TextEditingController();
  TicketTicket? _last;
  bool _busy = false;

  @override
  void dispose() {
    _code.dispose();
    super.dispose();
  }

  Future<void> _lookup({bool checkIn = false}) async {
    final code = _code.text.trim().toUpperCase();
    if (code.isEmpty || _event == null) return;
    final api = context.api.ticket;
    setState(() => _busy = true);
    try {
      if (checkIn) await api.ticketCheckIn(id: _event!.id!, code: code, scannedAt: DateTime.now().toUtc().toIso8601String());
      final t = await api.ticketLookupTicket(id: _event!.id!, code: code);
      if (!mounted) return;
      setState(() => _last = t);
      if (checkIn) {
        showSnack(context, 'Check-in thành công: ${t?.holderName ?? code}');
        _code.clear();
      }
    } catch (e) {
      if (mounted) showSnack(context, errorMessage(e));
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final api = context.api.ticket;
    if (_event == null) {
      return AsyncList<TicketEvent>(
        load: () => api.ticketStaffEvents(),
        header: const Padding(padding: EdgeInsets.fromLTRB(20, 16, 20, 0), child: Text('Chọn sự kiện bạn được phân công soát vé')),
        emptyMessage: 'Bạn chưa được phân công sự kiện nào',
        emptyIcon: Icons.qr_code_scanner,
        itemBuilder: (context, e, _) => EntityTile(
          title: e.title ?? '',
          subtitle: formatDate(e.startsAt),
          icon: Icons.qr_code_scanner,
          onTap: () => setState(() => _event = e),
        ),
      );
    }
    final t = _last;
    return ListView(
      padding: const EdgeInsets.only(bottom: 96),
      children: [
        ListTile(
          leading: IconButton(icon: const Icon(Icons.arrow_back), onPressed: () => setState(() => _event = null)),
          title: Text(_event!.title ?? '', style: const TextStyle(fontWeight: FontWeight.w800)),
          subtitle: Text(formatDate(_event!.startsAt)),
        ),
        Padding(
          padding: const EdgeInsets.symmetric(horizontal: 16),
          child: TextField(
            controller: _code,
            textCapitalization: TextCapitalization.characters,
            textAlign: TextAlign.center,
            style: const TextStyle(fontSize: 22, fontWeight: FontWeight.w800, letterSpacing: 2),
            decoration: const InputDecoration(hintText: 'Nhập mã vé'),
            onSubmitted: (_) => _lookup(checkIn: true),
          ),
        ),
        ActionButtons([
          OutlinedButton.icon(onPressed: _busy ? null : _lookup, icon: const Icon(Icons.search), label: const Text('Tra cứu')),
          FilledButton.icon(onPressed: _busy ? null : () => _lookup(checkIn: true), icon: const Icon(Icons.how_to_reg), label: const Text('Check-in')),
          OutlinedButton.icon(
            icon: const Icon(Icons.playlist_add_check),
            label: const Text('Check-in hàng loạt'),
            onPressed: () => showFormSheet(
              context,
              title: 'Check-in hàng loạt',
              description: 'Dán danh sách mã, mỗi dòng một mã (dùng khi máy quét ngoại tuyến đồng bộ lại).',
              fields: const [KField('codes', 'Các mã vé', kind: FieldKind.multiline, required: true)],
              success: 'Đã gửi danh sách check-in',
              onSubmit: (v) => api.ticketBatchCheckIn(id: _event!.id!, input: {
                'scans': [
                  for (final c in (v.str('codes') ?? '').split(RegExp(r'[\s,]+')).where((c) => c.isNotEmpty)) {'code': c.toUpperCase()},
                ],
              }),
            ),
          ),
        ]),
        if (t != null) ...[
          InfoSection(title: 'Vé ${t.code}', rows: [
            ('Người giữ', t.holderName),
            ('Loại vé', t.typeName),
            ('Ghế', t.seat),
            ('Trạng thái', t.status),
            ('Check-in', t.checkedInAt == null ? 'Chưa' : formatDate(t.checkedInAt)),
          ]),
          if (t.checkedInAt != null)
            ActionButtons([
              TextButton.icon(
                icon: const Icon(Icons.undo),
                label: const Text('Huỷ check-in'),
                onPressed: () async {
                  if (await runAction(context, () => api.ticketRevertCheckIn(id: t.id!), confirm: 'Huỷ check-in vé ${t.code}?', success: 'Đã huỷ check-in')) {
                    _code.text = t.code ?? '';
                    await _lookup();
                  }
                },
              ),
            ]),
        ],
      ],
    );
  }
}

class _Moderation extends StatelessWidget {
  const _Moderation();

  @override
  Widget build(BuildContext context) {
    final api = context.api.ticket;
    return AsyncList<TicketEvent>(
      load: () async => (await api.ticketEventReviewQueue(limit: 50))?.items,
      header: const Padding(
        padding: EdgeInsets.fromLTRB(20, 16, 20, 0),
        child: Text('Dành cho quản trị nền tảng: duyệt sự kiện trước khi mở bán.'),
      ),
      emptyMessage: 'Không có sự kiện chờ duyệt',
      emptyIcon: Icons.fact_check_outlined,
      itemBuilder: (context, e, reload) => EntityTile(
        title: e.title ?? '',
        subtitle: '${formatDate(e.startsAt)} • ${e.venue ?? ''} • BTC #${e.organizerId ?? ''}',
        icon: Icons.fact_check_rounded,
        badges: [_eventStatus(e.status), if (e.featured == true) const Pill('Nổi bật', icon: Icons.star_rounded)],
        trailing: PopupMenuButton<String>(
          onSelected: (a) async {
            var ok = false;
            if (a == 'approve') {
              ok = await runAction(context, () => api.ticketApproveEvent(id: e.id!), success: 'Đã duyệt, sự kiện bắt đầu mở bán');
            } else if (a == 'feature') {
              ok = await runAction(context, () => api.ticketSetFeatured(id: e.id!, input: {'featured': e.featured != true}), success: 'Đã cập nhật');
            } else {
              ok = await showFormSheet(
                context,
                title: 'Từ chối sự kiện',
                fields: const [KField('reason', 'Lý do (gửi cho ban tổ chức)', kind: FieldKind.multiline, required: true)],
                success: 'Đã từ chối',
                onSubmit: (v) => api.ticketRejectEvent(id: e.id!, reason: v.str('reason')!),
              );
            }
            if (ok) await reload();
          },
          itemBuilder: (_) => [
            const PopupMenuItem(value: 'approve', child: Text('Duyệt')),
            const PopupMenuItem(value: 'reject', child: Text('Từ chối')),
            PopupMenuItem(value: 'feature', child: Text(e.featured == true ? 'Bỏ nổi bật' : 'Đánh dấu nổi bật')),
          ],
        ),
      ),
    );
  }
}
