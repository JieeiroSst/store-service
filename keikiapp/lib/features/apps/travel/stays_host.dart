import 'package:flutter/material.dart';

import '../../../gen/models.g.dart';
import '../kit/kit.dart';
import 'stays_more.dart';

List<String> _list(String? s) => (s ?? '').split(',').map((e) => e.trim()).where((e) => e.isNotEmpty).toList();

String _iso(DateTime d) => d.toIso8601String().substring(0, 10);

Scaffold _hostScaffold(BuildContext context, {required String title, required List<String> tabs, required List<Widget> views, Widget? fab}) {
  final color = context.appColor;
  return Scaffold(
    appBar: AppBar(
      foregroundColor: Colors.white,
      backgroundColor: color,
      flexibleSpace: DecoratedBox(decoration: BoxDecoration(gradient: appGradient(color))),
      title: Text(title, style: const TextStyle(color: Colors.white, fontWeight: FontWeight.w700)),
      bottom: TabBar(
        labelColor: Colors.white,
        unselectedLabelColor: Colors.white70,
        indicatorColor: Colors.white,
        dividerColor: Colors.transparent,
        tabs: [for (final t in tabs) Tab(text: t)],
      ),
    ),
    floatingActionButton: fab,
    body: TabBarView(children: views),
  );
}

List<KField> _hotelFields([ResvHotel? h]) => [
      KField('name', 'Tên khách sạn', required: true, initial: h?.name_),
      KField('stars', 'Hạng sao', kind: FieldKind.choice, initial: '${h?.stars ?? 3}', options: [for (var i = 1; i <= 5; i++) ('$i', '$i sao')]),
      KField('city', 'Thành phố', required: true, initial: h?.city),
      KField('address', 'Địa chỉ', required: true, initial: h?.address),
      KField('phone', 'Điện thoại', kind: FieldKind.phone, initial: h?.phoneNumber),
      KField('email', 'Email', kind: FieldKind.email, initial: h?.email),
      KField('checkin', 'Giờ nhận phòng', initial: h?.checkInTime ?? '14:00'),
      KField('checkout', 'Giờ trả phòng', initial: h?.checkOutTime ?? '12:00'),
      KField('free', 'Huỷ miễn phí trước (giờ)', kind: FieldKind.number, initial: '${h?.freeCancelHours ?? 24}'),
      KField('amenities', 'Tiện nghi (phân cách bởi dấu phẩy)', initial: (h?.amenities ?? const []).join(', ')),
      KField('images', 'Ảnh (URL, phân cách bởi dấu phẩy)', initial: (h?.images ?? const []).join(', ')),
      KField('description', 'Giới thiệu', kind: FieldKind.multiline, initial: h?.description),
    ];

ResvHotelInput _hotelInput(FormData v) => ResvHotelInput(
      name_: v.str('name'),
      stars: v.integer('stars'),
      city: v.str('city'),
      address: v.str('address'),
      phoneNumber: v.str('phone'),
      email: v.str('email'),
      checkInTime: v.str('checkin'),
      checkOutTime: v.str('checkout'),
      freeCancelHours: v.integer('free'),
      amenities: _list(v.str('amenities')),
      images: _list(v.str('images')),
      description: v.str('description'),
      currency: 'VND',
    );

class HotelHostPage extends StatelessWidget {
  const HotelHostPage({super.key});

  @override
  Widget build(BuildContext context) {
    final api = context.api.reservation;
    return DefaultTabController(
      length: 2,
      child: _hostScaffold(
        context,
        title: 'Quản lý khách sạn',
        tabs: const ['Khách sạn của tôi', 'Kiểm duyệt'],
        fab: Builder(
          builder: (context) => FloatingActionButton.extended(
            backgroundColor: context.appColor,
            foregroundColor: Colors.white,
            icon: const Icon(Icons.add_business_outlined),
            label: const Text('Đăng ký khách sạn'),
            onPressed: () => showFormSheet(
              context,
              title: 'Đăng ký khách sạn',
              description: 'Khách sạn sẽ hiển thị sau khi quản trị duyệt.',
              fields: _hotelFields(),
              submitLabel: 'Gửi đăng ký',
              success: 'Đã gửi đăng ký, kéo xuống để làm mới',
              onSubmit: (v) => api.resvRegisterHotel(input: _hotelInput(v)),
            ),
          ),
        ),
        views: [
          AsyncList<ResvHotel>(
            load: () async => (await api.resvMyHotels(limit: 50))?.items,
            emptyMessage: 'Bạn chưa quản lý khách sạn nào',
            emptyIcon: Icons.hotel_outlined,
            itemBuilder: (context, h, reload) => EntityTile(
              title: h.name_ ?? '',
              subtitle: [h.address, h.city].whereType<String>().join(', '),
              imageUrl: (h.images ?? const []).firstOrNull,
              icon: Icons.hotel_rounded,
              badges: [StatusPill(h.statusName)],
              onTap: () async {
                await openPage(context, _ManageHotel(id: h.id!));
                await reload();
              },
            ),
          ),
          AsyncList<ResvHotel>(
            load: () async => (await api.resvHotelReviewQueue(limit: 50))?.items,
            header: const Padding(padding: EdgeInsets.fromLTRB(20, 16, 20, 0), child: Text('Dành cho quản trị nền tảng: duyệt khách sạn mới đăng ký.')),
            emptyMessage: 'Không có khách sạn chờ duyệt',
            emptyIcon: Icons.fact_check_outlined,
            itemBuilder: (context, h, reload) => EntityTile(
              title: h.name_ ?? '',
              subtitle: '${h.city ?? ''} • quản lý #${h.managerId ?? ''}',
              icon: Icons.fact_check_rounded,
              badges: [StatusPill(h.statusName)],
              trailing: PopupMenuButton<String>(
                onSelected: (a) async {
                  final ok = a == 'approve'
                      ? await runAction(context, () => api.resvApproveHotel(id: h.id!), success: 'Đã duyệt')
                      : await showFormSheet(
                          context,
                          title: 'Từ chối ${h.name_}',
                          fields: const [KField('reason', 'Lý do', kind: FieldKind.multiline, required: true)],
                          success: 'Đã từ chối',
                          onSubmit: (v) => api.resvRejectHotel(id: h.id!, reason: v.str('reason')!),
                        );
                  if (ok) await reload();
                },
                itemBuilder: (_) => const [PopupMenuItem(value: 'approve', child: Text('Duyệt')), PopupMenuItem(value: 'reject', child: Text('Từ chối'))],
              ),
            ),
          ),
        ],
      ),
    );
  }
}

class _ManageHotel extends StatefulWidget {
  const _ManageHotel({required this.id});

  final int id;

  @override
  State<_ManageHotel> createState() => _ManageHotelState();
}

class _ManageHotelState extends State<_ManageHotel> {
  int _v = 0;

  void _bump() => setState(() => _v++);

  Future<void> _roomType(ResvRoomType? t) async {
    final api = context.api.reservation;
    if (await showFormSheet(
      context,
      title: t == null ? 'Loại phòng mới' : 'Sửa ${t.name_}',
      fields: [
        KField('name', 'Tên', required: true, initial: t?.name_),
        KField('capacity', 'Sức chứa (khách)', kind: FieldKind.number, required: true, initial: t?.capacity?.toString()),
        KField('bed', 'Giường', initial: t?.bed),
        KField('size', 'Diện tích (m²)', kind: FieldKind.number, initial: t?.sizeM2?.toString()),
        KField('view', 'Hướng nhìn', initial: t?.view),
        KField('amenities', 'Tiện nghi (phân cách dấu phẩy)', initial: (t?.amenities ?? const []).join(', ')),
        KField('active', 'Đang mở bán', kind: FieldKind.toggle, initial: '${t?.active ?? true}'),
        KField('description', 'Mô tả', kind: FieldKind.multiline, initial: t?.description),
      ],
      success: 'Đã lưu loại phòng',
      onSubmit: (v) {
        final input = ResvRoomTypeInput(
          name_: v.str('name'),
          capacity: v.integer('capacity'),
          bed: v.str('bed'),
          sizeM2: v.integer('size'),
          view: v.str('view'),
          amenities: _list(v.str('amenities')),
          active: v.flag('active'),
          description: v.str('description'),
        );
        return t == null ? api.resvCreateRoomType(id: widget.id, input: input) : api.resvUpdateRoomType(id: widget.id, type_: '${t.id}', input: input);
      },
    )) {
      _bump();
    }
  }

  Future<void> _inventory(ResvRoomType t) async {
    final api = context.api.reservation;
    final now = DateTime.now();
    final inv = await api.resvInventory(id: widget.id, type_: t.id!, from: _iso(now), to: _iso(now.add(const Duration(days: 14)))).catchError((_) => null) ?? const <ResvInventory>[];
    if (!mounted) return;
    await showModalBottomSheet<void>(
      context: context,
      showDragHandle: true,
      isScrollControlled: true,
      builder: (_) => SafeArea(
        child: SizedBox(
          height: MediaQuery.sizeOf(context).height * 0.7,
          child: Column(
            children: [
              ListTile(
                title: Text('Phòng trống • ${t.name_}', style: const TextStyle(fontWeight: FontWeight.w800)),
                trailing: FilledButton(
                  onPressed: () async {
                    Navigator.pop(context);
                    if (await showFormSheet(
                      context,
                      title: 'Đặt số phòng & giá',
                      fields: [
                        KField('from', 'Từ ngày', kind: FieldKind.date, required: true, initial: _iso(now)),
                        KField('to', 'Đến ngày', kind: FieldKind.date, required: true, initial: _iso(now.add(const Duration(days: 30)))),
                        const KField('total', 'Số phòng mở bán', kind: FieldKind.number, required: true),
                        const KField('rate', 'Giá mỗi đêm (₫)', kind: FieldKind.number),
                        const KField('off', 'Tạm ngừng bán', kind: FieldKind.toggle, initial: 'false'),
                      ],
                      success: 'Đã cập nhật lịch phòng',
                      onSubmit: (v) => api.resvSetInventory(
                        id: widget.id,
                        type_: '${t.id}',
                        input: ResvInventoryInput(from: v.str('from'), to: v.str('to'), totalInventory: v.integer('total'), rate: v.number('rate'), rateOff: v.flag('off')),
                      ),
                    )) {
                      _bump();
                    }
                  },
                  child: const Text('Cập nhật'),
                ),
              ),
              Expanded(
                child: ListView(
                  children: [
                    if (inv.isEmpty) const Padding(padding: EdgeInsets.all(24), child: Text('Chưa đặt lịch phòng cho 2 tuần tới')),
                    for (final i in inv)
                      ListTile(
                        dense: true,
                        title: Text(shortDate(i.date)),
                        subtitle: Text('Đã đặt ${i.totalReserved ?? 0}/${i.totalInventory ?? 0}'),
                        trailing: Text('${formatMoneyString(i.rate)} • còn ${i.left ?? 0}', style: const TextStyle(fontWeight: FontWeight.w700)),
                      ),
                  ],
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final api = context.api.reservation;
    final id = widget.id;
    final now = DateTime.now();
    return AsyncBuilder<(ResvHotel?, List<ResvRoomType>, List<ResvRoom>, List<ResvPromotion>, List<ResvService>, ResvReport?)>(
      key: ValueKey(_v),
      load: () async {
        final r = await Future.wait<Object?>([
          api.resvHotel(id: id),
          api.resvRoomTypes(id: id).catchError((_) => null),
          api.resvRooms(id: id).catchError((_) => null),
          api.resvPromotions(id: id).catchError((_) => null),
          api.resvServices(id: id).catchError((_) => null),
          api.resvHotelReport(id: id, from: _iso(now.subtract(const Duration(days: 30))), to: _iso(now)).catchError((_) => null),
        ]);
        return (
          r[0] as ResvHotel?,
          (r[1] as List<ResvRoomType>?) ?? const <ResvRoomType>[],
          (r[2] as List<ResvRoom>?) ?? const <ResvRoom>[],
          (r[3] as List<ResvPromotion>?) ?? const <ResvPromotion>[],
          (r[4] as List<ResvService>?) ?? const <ResvService>[],
          r[5] as ResvReport?,
        );
      },
      loading: const Scaffold(body: SkeletonList()),
      builder: (context, d, _) {
        final h = d.$1;
        final rep = d.$6;
        return DetailPage(
          title: h?.name_ ?? 'Khách sạn',
          actions: [
            PopupMenuButton<String>(
              onSelected: (a) async {
                if (a == 'edit') {
                  if (await showFormSheet(context, title: 'Sửa khách sạn', fields: _hotelFields(h), success: 'Đã lưu', onSubmit: (v) => api.resvUpdateHotel(id: id, input: _hotelInput(v)))) _bump();
                } else if (await runAction(context, () => api.resvDeactivateHotel(id: id), confirm: 'Tạm ngừng nhận khách?', destructive: true, success: 'Đã tạm ngừng')) {
                  _bump();
                }
              },
              itemBuilder: (_) => const [PopupMenuItem(value: 'edit', child: Text('Sửa thông tin')), PopupMenuItem(value: 'off', child: Text('Tạm ngừng'))],
            ),
          ],
          header: AppBanner(
            icon: Icons.insights_rounded,
            title: formatMoneyString(rep?.revenue?.netRevenue, rep?.revenue?.currency),
            subtitle: 'Doanh thu ròng 30 ngày • ${h?.statusName ?? ''}',
            children: [
              StatRow([
                StatTile(label: 'Công suất', value: '${(rep?.occupancy?.occupancyPercent ?? 0).toStringAsFixed(0)}%', onDark: true),
                StatTile(label: 'Đêm bán', value: '${rep?.occupancy?.roomNightsSold ?? 0}', onDark: true),
                StatTile(label: 'Không đến', value: '${rep?.noShows ?? 0}', onDark: true),
              ]),
            ],
          ),
          children: [
            ActionButtons([
              FilledButton.icon(icon: const Icon(Icons.event_note), label: const Text('Đặt phòng của khách'), onPressed: () => openPage(context, _HotelReservations(hotelId: id))),
            ]),
            SectionHeader('Loại phòng', action: 'Thêm', onAction: () => _roomType(null)),
            for (final t in d.$2)
              Padding(
                padding: const EdgeInsets.fromLTRB(16, 0, 16, 8),
                child: EntityTile(
                  title: t.name_ ?? '',
                  subtitle: '${t.capacity ?? '-'} khách • ${d.$3.where((r) => r.roomTypeId == t.id).length} phòng',
                  icon: Icons.bed_rounded,
                  badges: [if (t.active == false) const StatusPill('inactive')],
                  onTap: () => _roomType(t),
                  trailing: IconButton(tooltip: 'Lịch phòng & giá', icon: const Icon(Icons.calendar_month), onPressed: () => _inventory(t)),
                ),
              ),
            SectionHeader('Phòng', action: 'Thêm', onAction: () async {
              if (d.$2.isEmpty) {
                showSnack(context, 'Hãy tạo loại phòng trước');
                return;
              }
              if (await _roomForm(context, id, null, d.$2)) _bump();
            }),
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: 16),
              child: Wrap(
                spacing: 8,
                runSpacing: 8,
                children: [
                  for (final r in d.$3)
                    ActionChip(
                      avatar: Icon(r.available == false ? Icons.block : Icons.meeting_room_outlined, size: 18),
                      label: Text('${r.name_} • T${r.floor ?? '-'}'),
                      onPressed: () async {
                        if (await _roomForm(context, id, r, d.$2)) _bump();
                      },
                    ),
                ],
              ),
            ),
            SectionHeader('Khuyến mãi', action: 'Thêm', onAction: () async {
              if (await _promoForm(context, id, null)) _bump();
            }),
            for (final p in d.$4)
              Padding(
                padding: const EdgeInsets.fromLTRB(16, 0, 16, 8),
                child: EntityTile(
                  title: p.code ?? '',
                  subtitle: p.description ?? ((p.percentOff ?? 0) > 0 ? 'Giảm ${p.percentOff}%' : 'Giảm ${formatMoneyString(p.amountOff)}'),
                  caption: 'Đã dùng ${p.usedCount ?? 0}/${p.maxUses ?? '∞'} • tối thiểu ${p.minNights ?? 1} đêm',
                  icon: Icons.local_offer_rounded,
                  badges: [if (p.active == false) const StatusPill('inactive')],
                  onTap: () async {
                    if (await _promoForm(context, id, p)) _bump();
                  },
                ),
              ),
            SectionHeader('Dịch vụ thêm', action: 'Thêm', onAction: () async {
              if (await _serviceForm(context, id, null)) _bump();
            }),
            for (final s in d.$5)
              Padding(
                padding: const EdgeInsets.fromLTRB(16, 0, 16, 8),
                child: EntityTile(
                  title: s.name_ ?? '',
                  subtitle: s.description,
                  caption: '${formatMoneyString(s.price)}${s.unit == null ? '' : '/${s.unit}'}',
                  icon: Icons.room_service_rounded,
                  badges: [if (s.active == false) const StatusPill('inactive')],
                  onTap: () async {
                    if (await _serviceForm(context, id, s)) _bump();
                  },
                ),
              ),
          ],
        );
      },
    );
  }
}

Future<bool> _roomForm(BuildContext context, int hotelId, ResvRoom? r, List<ResvRoomType> types) {
  final api = context.api.reservation;
  return showFormSheet(
    context,
    title: r == null ? 'Phòng mới' : 'Phòng ${r.name_}',
    fields: [
      KField('name', 'Số phòng', required: true, initial: r?.name_),
      KField('type', 'Loại phòng', kind: FieldKind.choice, initial: '${r?.roomTypeId ?? types.first.id}', options: [for (final t in types) ('${t.id}', t.name_ ?? '')]),
      KField('floor', 'Tầng', kind: FieldKind.number, initial: r?.floor?.toString()),
      KField('available', 'Sẵn sàng đón khách', kind: FieldKind.toggle, initial: '${r?.available ?? true}'),
    ],
    success: 'Đã lưu phòng',
    onSubmit: (v) {
      final input = ResvRoomInput(name_: v.str('name'), roomTypeId: v.integer('type'), floor: v.integer('floor'), available: v.flag('available'));
      return r == null ? api.resvCreateRoom(id: hotelId, input: input) : api.resvUpdateRoom(id: hotelId, room: '${r.id}', input: input);
    },
  );
}

Future<bool> _promoForm(BuildContext context, int hotelId, ResvPromotion? p) {
  final api = context.api.reservation;
  return showFormSheet(
    context,
    title: p == null ? 'Khuyến mãi mới' : 'Sửa ${p.code}',
    fields: [
      KField('code', 'Mã', required: true, initial: p?.code),
      KField('description', 'Mô tả', initial: p?.description),
      KField('percent', 'Giảm (%)', kind: FieldKind.number, initial: (p?.percentOff ?? 0) == 0 ? null : '${p!.percentOff}'),
      KField('amount', 'Hoặc giảm số tiền (₫)', kind: FieldKind.number, initial: p?.amountOff),
      KField('min', 'Tối thiểu số đêm', kind: FieldKind.number, initial: p?.minNights?.toString()),
      KField('max', 'Số lượt tối đa (0 = không giới hạn)', kind: FieldKind.number, initial: p?.maxUses?.toString()),
      KField('from', 'Áp dụng từ ngày nhận phòng', kind: FieldKind.date, initial: p?.validFrom),
      KField('to', 'Đến ngày', kind: FieldKind.date, initial: p?.validTo),
      KField('active', 'Đang hoạt động', kind: FieldKind.toggle, initial: '${p?.active ?? true}'),
    ],
    success: 'Đã lưu khuyến mãi',
    onSubmit: (v) {
      final input = {
        'code': v.str('code')?.toUpperCase(),
        'description': v.str('description'),
        'percent_off': ?v.integer('percent'),
        'amount_off': ?v.str('amount'),
        'min_nights': ?v.integer('min'),
        'max_uses': ?v.integer('max'),
        'valid_from': ?v.str('from'),
        'valid_to': ?v.str('to'),
        'active': v.flag('active'),
      };
      return p == null ? api.resvCreatePromotion(id: hotelId, input: input) : api.resvUpdatePromotion(id: hotelId, promo: '${p.id}', input: input);
    },
  );
}

Future<bool> _serviceForm(BuildContext context, int hotelId, ResvService? s) {
  final api = context.api.reservation;
  return showFormSheet(
    context,
    title: s == null ? 'Dịch vụ mới' : 'Sửa ${s.name_}',
    fields: [
      KField('name', 'Tên dịch vụ', required: true, initial: s?.name_),
      KField('price', 'Giá (₫)', kind: FieldKind.number, required: true, initial: s?.price),
      KField('unit', 'Đơn vị', initial: s?.unit, hint: 'lượt, người, ngày…'),
      KField('active', 'Đang cung cấp', kind: FieldKind.toggle, initial: '${s?.active ?? true}'),
      KField('description', 'Mô tả', kind: FieldKind.multiline, initial: s?.description),
    ],
    success: 'Đã lưu dịch vụ',
    onSubmit: (v) {
      final input = ResvServiceInput(name_: v.str('name'), price: v.number('price'), unit: v.str('unit'), active: v.flag('active'), description: v.str('description'));
      return s == null ? api.resvCreateService(id: hotelId, input: input) : api.resvUpdateService(id: hotelId, service: '${s.id}', input: input);
    },
  );
}

class _HotelReservations extends StatelessWidget {
  const _HotelReservations({required this.hotelId});

  final int hotelId;

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Đặt phòng của khách')),
      body: AsyncList<ResvReservation>(
        load: () async => (await context.api.reservation.resvReservations(as_: 'manager', hotelId: hotelId, limit: 100))?.items,
        emptyMessage: 'Chưa có đặt phòng',
        itemBuilder: (context, r, reload) => EntityTile(
          title: 'Đặt phòng #${r.id} • khách #${r.guestId ?? ''}',
          subtitle: '${shortDate(r.checkIn)} – ${shortDate(r.checkOut)} • ${r.rooms ?? 1} phòng',
          caption: formatMoneyString(r.totalAmount, r.currency),
          icon: Icons.hotel_rounded,
          badges: [StatusPill(r.statusName), if (r.stayStatus != null) Pill(r.stayStatus!)],
          onTap: () async {
            await openPage(context, ResvDetailPage(id: r.id!, manager: true));
            await reload();
          },
        ),
      ),
    );
  }
}

const _propertyTypes = [('1', 'Homestay'), ('2', 'Khách sạn nhỏ'), ('3', 'Nhà / căn hộ cho thuê')];

List<KField> _homeFields([RentHomestay? h]) => [
      KField('name', 'Tên chỗ ở', required: true, initial: h?.name_),
      KField('type', 'Loại hình', kind: FieldKind.choice, initial: '${h?.type_ ?? 1}', options: _propertyTypes),
      KField('address', 'Địa chỉ', required: true, initial: h?.address),
      KField('phone', 'Điện thoại', kind: FieldKind.phone, initial: h?.phoneNumber),
      KField('guests', 'Số khách tối đa', kind: FieldKind.number, required: true, initial: h?.guests?.toString()),
      KField('bedrooms', 'Phòng ngủ', kind: FieldKind.number, initial: h?.bedrooms?.toString()),
      KField('bathrooms', 'Phòng tắm', kind: FieldKind.number, initial: h?.bathrooms?.toString()),
      KField('amenities', 'Mã tiện nghi (phân cách dấu phẩy)', initial: (h?.amenityIds ?? const []).join(', ')),
      KField('images', 'Ảnh (URL, phân cách dấu phẩy)', initial: (h?.images ?? const []).join(', ')),
      KField('description', 'Giới thiệu', kind: FieldKind.multiline, initial: h?.description),
    ];

RentHomestayInput _homeInput(FormData v) => RentHomestayInput(
      name_: v.str('name'),
      type_: v.integer('type'),
      address: v.str('address'),
      phoneNumber: v.str('phone'),
      guests: v.integer('guests'),
      bedrooms: v.integer('bedrooms'),
      bathrooms: v.integer('bathrooms'),
      amenityIds: v.ints('amenities'),
      images: _list(v.str('images')),
      description: v.str('description'),
    );

class HomestayHostPage extends StatelessWidget {
  const HomestayHostPage({super.key});

  @override
  Widget build(BuildContext context) {
    final api = context.api.rentHouse;
    return DefaultTabController(
      length: 4,
      child: _hostScaffold(
        context,
        title: 'Chủ nhà',
        tabs: const ['Chỗ ở', 'Đặt chỗ', 'Hợp đồng', 'Kiểm duyệt'],
        fab: Builder(
          builder: (context) => FloatingActionButton.extended(
            backgroundColor: context.appColor,
            foregroundColor: Colors.white,
            icon: const Icon(Icons.add_home_outlined),
            label: const Text('Đăng chỗ ở'),
            onPressed: () async {
              final amenities = await api.rentAmenities().catchError((_) => null) ?? const <RentAmenity>[];
              if (!context.mounted) return;
              await showFormSheet(
                context,
                title: 'Đăng chỗ ở mới',
                description: amenities.isEmpty ? null : 'Tiện nghi: ${amenities.map((a) => '${a.id}=${a.name_}').join(', ')}',
                fields: _homeFields(),
                submitLabel: 'Gửi duyệt',
                success: 'Đã gửi, chỗ ở hiển thị sau khi được duyệt',
                onSubmit: (v) => api.rentCreateHomestay(input: _homeInput(v)),
              );
            },
          ),
        ),
        views: [
          AsyncList<RentHomestay>(
            load: () => api.rentMyHomestays(limit: 50),
            header: Padding(
              padding: const EdgeInsets.fromLTRB(16, 12, 16, 0),
              child: Align(
                alignment: Alignment.centerLeft,
                child: TextButton.icon(
                  icon: const Icon(Icons.add),
                  label: const Text('Thêm loại tiện nghi'),
                  onPressed: () => showFormSheet(
                    context,
                    title: 'Tiện nghi mới',
                    fields: const [KField('name', 'Tên', required: true), KField('icon', 'Biểu tượng', hint: 'wifi, pool…')],
                    success: 'Đã thêm tiện nghi',
                    onSubmit: (v) => api.rentCreateAmenity(name_: v.str('name')!, icon: v.str('icon')),
                  ),
                ),
              ),
            ),
            emptyMessage: 'Bạn chưa đăng chỗ ở nào',
            emptyIcon: Icons.cottage_outlined,
            itemBuilder: (context, h, reload) => EntityTile(
              title: h.name_ ?? '',
              subtitle: h.address,
              imageUrl: (h.images ?? const []).firstOrNull,
              icon: Icons.cottage_rounded,
              badges: [StatusPill(h.statusName)],
              onTap: () async {
                await openPage(context, _ManageHome(home: h));
                await reload();
              },
            ),
          ),
          AsyncList<RentBooking>(
            load: () => api.rentBookings(as_: 'host', limit: 100),
            emptyMessage: 'Chưa có khách đặt',
            itemBuilder: (context, b, _) => EntityTile(
              title: 'Đặt chỗ #${b.id} • nhà #${b.homestayId}',
              subtitle: '${shortDate(b.checkinDate)} – ${shortDate(b.checkoutDate)} • ${b.guests ?? '-'} khách',
              caption: formatMoneyString(b.totalAmount, b.currency),
              icon: Icons.luggage_rounded,
              badges: [StatusPill(switch (b.status) { 1 => 'confirmed', 2 => 'cancelled', 3 => 'pending', _ => null })],
              onTap: () => openPage(context, RentBookingPage(id: b.id!)),
            ),
          ),
          AsyncList<RentLease>(
            load: () => api.rentLeases(as_: 'host', limit: 100),
            emptyMessage: 'Chưa có hợp đồng thuê',
            itemBuilder: (context, l, reload) => EntityTile(
              title: 'Hợp đồng #${l.id} • khách #${l.userId}',
              subtitle: '${shortDate(l.startDate)} – ${shortDate(l.endDate)}',
              caption: '${formatMoneyString(l.rent, l.currency)}/${rateUnit(l.model)}',
              icon: Icons.description_rounded,
              badges: [StatusPill(l.status)],
              onTap: () async {
                await openPage(context, RentLeasePage(id: l.id!, host: true));
                await reload();
              },
            ),
          ),
          AsyncList<RentHomestay>(
            load: () => api.rentHomestayReviewQueue(limit: 50),
            header: const Padding(padding: EdgeInsets.fromLTRB(20, 16, 20, 0), child: Text('Dành cho quản trị nền tảng: duyệt chỗ ở mới.')),
            emptyMessage: 'Không có chỗ ở chờ duyệt',
            emptyIcon: Icons.fact_check_outlined,
            itemBuilder: (context, h, reload) => EntityTile(
              title: h.name_ ?? '',
              subtitle: '${h.address ?? ''} • chủ #${h.hostId ?? ''}',
              icon: Icons.fact_check_rounded,
              trailing: PopupMenuButton<String>(
                onSelected: (a) async {
                  final ok = a == 'approve'
                      ? await runAction(context, () => api.rentApproveHomestay(id: h.id!), success: 'Đã duyệt')
                      : await showFormSheet(
                          context,
                          title: 'Từ chối ${h.name_}',
                          fields: const [KField('reason', 'Lý do', kind: FieldKind.multiline, required: true)],
                          success: 'Đã từ chối',
                          onSubmit: (v) => api.rentRejectHomestay(id: h.id!, reason: v.str('reason')!),
                        );
                  if (ok) await reload();
                },
                itemBuilder: (_) => const [PopupMenuItem(value: 'approve', child: Text('Duyệt')), PopupMenuItem(value: 'reject', child: Text('Từ chối'))],
              ),
            ),
          ),
        ],
      ),
    );
  }
}

class _ManageHome extends StatefulWidget {
  const _ManageHome({required this.home});

  final RentHomestay home;

  @override
  State<_ManageHome> createState() => _ManageHomeState();
}

class _ManageHomeState extends State<_ManageHome> {
  int _v = 0;

  @override
  Widget build(BuildContext context) {
    final api = context.api.rentHouse;
    final h = widget.home;
    final id = h.id!;
    return AsyncBuilder<List<RentRate>>(
      key: ValueKey(_v),
      load: () async => await api.rentRates(id: id) ?? const <RentRate>[],
      loading: const Scaffold(body: SkeletonList()),
      builder: (context, rates, _) => DetailPage(
        title: h.name_ ?? 'Chỗ ở',
        actions: [
          PopupMenuButton<String>(
            onSelected: (a) async {
              final ok = a == 'edit'
                  ? await showFormSheet(context, title: 'Sửa chỗ ở', fields: _homeFields(h), success: 'Đã lưu', onSubmit: (v) => api.rentUpdateHomestay(id: id, input: _homeInput(v)))
                  : await runAction(context, () => api.rentDeactivateHomestay(id: id), confirm: 'Tạm ẩn chỗ ở này?', destructive: true, success: 'Đã tạm ẩn');
              if (ok) setState(() => _v++);
            },
            itemBuilder: (_) => const [PopupMenuItem(value: 'edit', child: Text('Sửa thông tin')), PopupMenuItem(value: 'off', child: Text('Tạm ẩn'))],
          ),
        ],
        header: HeroHeader(title: h.name_ ?? '', subtitle: h.address, imageUrl: (h.images ?? const []).firstOrNull, icon: Icons.cottage_rounded, badges: [StatusPill(h.statusName)]),
        children: [
          SectionHeader('Bảng giá', action: 'Đặt giá', onAction: () async {
            if (await showFormSheet(
              context,
              title: 'Giá thuê',
              fields: const [
                KField('model', 'Theo', kind: FieldKind.choice, options: [('day', 'Đêm'), ('week', 'Tuần'), ('month', 'Tháng'), ('year', 'Năm')]),
                KField('price', 'Giá (₫)', kind: FieldKind.number, required: true),
                KField('min', 'Tối thiểu số kỳ', kind: FieldKind.number, initial: '1'),
                KField('active', 'Đang áp dụng', kind: FieldKind.toggle, initial: 'true'),
              ],
              success: 'Đã lưu giá',
              onSubmit: (v) => api.rentSetRate(id: id, rateModel: v.str('model')!, input: RentRateInput(price: v.number('price'), minPeriods: v.integer('min'), active: v.flag('active'), currency: 'VND')),
            )) {
              setState(() => _v++);
            }
          }),
          for (final r in rates)
            Padding(
              padding: const EdgeInsets.fromLTRB(16, 0, 16, 8),
              child: EntityTile(
                title: 'Theo ${rateUnit(r.model)}',
                subtitle: 'Tối thiểu ${r.minPeriods ?? 1} kỳ',
                caption: formatMoneyString(r.price, r.currency),
                icon: Icons.sell_rounded,
                badges: [if (r.active == false) const StatusPill('inactive')],
                trailing: IconButton(
                  tooltip: 'Xoá giá',
                  icon: const Icon(Icons.delete_outline),
                  onPressed: () async {
                    if (await runAction(context, () => api.rentDeleteRate(id: id, rateModel: r.model!), confirm: 'Xoá giá theo ${rateUnit(r.model)}?', destructive: true, success: 'Đã xoá')) {
                      setState(() => _v++);
                    }
                  },
                ),
              ),
            ),
          ActionButtons([
            OutlinedButton.icon(
              icon: const Icon(Icons.calendar_month),
              label: const Text('Chặn / mở lịch'),
              onPressed: () => showFormSheet(
                context,
                title: 'Lịch trống',
                description: 'Đặt giá riêng hoặc chặn các ngày không nhận khách.',
                fields: [
                  KField('from', 'Từ ngày', kind: FieldKind.date, required: true, initial: _iso(DateTime.now())),
                  KField('to', 'Đến ngày', kind: FieldKind.date, required: true, initial: _iso(DateTime.now().add(const Duration(days: 7)))),
                  const KField('status', 'Trạng thái', kind: FieldKind.choice, options: [('1', 'Mở nhận khách'), ('3', 'Chặn')]),
                  const KField('price', 'Giá riêng mỗi đêm (₫)', kind: FieldKind.number),
                ],
                success: 'Đã cập nhật lịch',
                onSubmit: (v) => api.rentSetAvailability(id: id, input: RentAvailabilityInput(from: v.str('from'), to: v.str('to'), status: v.integer('status'), price: v.number('price'))),
              ),
            ),
          ]),
          HomestayExtras(home: h, onLease: () {}),
        ],
      ),
    );
  }
}
