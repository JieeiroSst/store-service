import 'package:flutter/material.dart';

import '../../../gen/models.g.dart';
import '../kit/kit.dart';
import 'stays_host.dart';
import 'stays_more.dart';

String _iso(DateTime d) => d.toIso8601String().substring(0, 10);

class _StayQuery {
  _StayQuery() : checkIn = DateTime.now().add(const Duration(days: 1)), checkOut = DateTime.now().add(const Duration(days: 3));

  factory _StayQuery.from(Map<String, String> q) {
    final r = _StayQuery();
    final inDate = DateTime.tryParse(q['check_in'] ?? '');
    final outDate = DateTime.tryParse(q['check_out'] ?? '');
    if (inDate != null) r.checkIn = inDate;
    if (outDate != null && outDate.isAfter(r.checkIn)) {
      r.checkOut = outDate;
    } else if (inDate != null) {
      r.checkOut = r.checkIn.add(const Duration(days: 1));
    }
    r.guests = (int.tryParse(q['guests'] ?? '') ?? r.guests).clamp(1, 30);
    return r;
  }

  Map<String, String> get params => {'check_in': _day(checkIn), 'check_out': _day(checkOut), 'guests': '$guests'};

  static String _day(DateTime d) => d.toIso8601String().substring(0, 10);

  DateTime checkIn;
  DateTime checkOut;
  int guests = 2;

  int get nights => checkOut.difference(checkIn).inDays.clamp(1, 365);
}

class _DatesBar extends StatelessWidget {
  const _DatesBar({required this.query, required this.onChanged});

  final _StayQuery query;
  final VoidCallback onChanged;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.fromLTRB(16, 12, 16, 0),
      child: Row(
        children: [
          Expanded(
            child: OutlinedButton.icon(
              icon: const Icon(Icons.date_range),
              label: Text('${shortDate(query.checkIn.toIso8601String()).substring(0, 5)} – ${shortDate(query.checkOut.toIso8601String()).substring(0, 5)} • ${query.nights} đêm'),
              onPressed: () async {
                final now = DateTime.now();
                final r = await showDateRangePicker(
                  context: context,
                  firstDate: now,
                  lastDate: now.add(const Duration(days: 365)),
                  initialDateRange: DateTimeRange(start: query.checkIn, end: query.checkOut),
                );
                if (r != null) {
                  query
                    ..checkIn = r.start
                    ..checkOut = r.end.isAfter(r.start) ? r.end : r.start.add(const Duration(days: 1));
                  onChanged();
                }
              },
            ),
          ),
          const SizedBox(width: 8),
          OutlinedButton.icon(
            icon: const Icon(Icons.person_outline),
            label: Text('${query.guests}'),
            onPressed: () {
              query.guests = query.guests % 8 + 1;
              onChanged();
            },
          ),
        ],
      ),
    );
  }
}

Widget _stars(int? n, double rating, int? reviews) => Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        const Icon(Icons.star_rounded, size: 16, color: Color(0xFFF59E0B)),
        Text(' ${rating.toStringAsFixed(1)}', style: const TextStyle(fontWeight: FontWeight.w700)),
        if (reviews != null) Text(' ($reviews)', style: const TextStyle(fontSize: 12)),
        if (n != null && n > 0) Text('  •  $n sao', style: const TextStyle(fontSize: 12)),
      ],
    );

Future<void> _payWithWallet(BuildContext context, {required String title, required String amount, required Future<Object?> Function() pay}) async {
  final ok = await showDialog<bool>(
    context: context,
    builder: (context) => AlertDialog(
      title: Text(title),
      content: Text('Thanh toán $amount bằng Ví Keiki?'),
      actions: [
        TextButton(onPressed: () => Navigator.pop(context, false), child: const Text('Để sau')),
        FilledButton(onPressed: () => Navigator.pop(context, true), child: const Text('Thanh toán')),
      ],
    ),
  );
  if (ok != true || !context.mounted) return;
  await runAction(context, pay, success: 'Thanh toán thành công');
}

class RentHouseApp extends StatefulWidget {
  const RentHouseApp({super.key});

  @override
  State<RentHouseApp> createState() => _RentHouseAppState();
}

class _RentHouseAppState extends State<RentHouseApp> {
  final _q = _StayQuery();

  @override
  Widget build(BuildContext context) {
    final api = context.api.rentHouse;
    return ServiceApp(
      serviceId: 'rent_house',
      adminActions: [
        Builder(
          builder: (context) => IconButton(tooltip: 'Chủ nhà', icon: const Icon(Icons.key_outlined), onPressed: () => openPage(context, const HomestayHostPage())),
        ),
      ],
      tabs: [
        AppTab('Khám phá', Icons.travel_explore, (context) => Column(
              children: [
                _DatesBar(query: _q, onChanged: () => setState(() {})),
                Expanded(
                  child: SearchableList<RentHomestay>(
                    key: ValueKey('${_q.checkIn}${_q.checkOut}${_q.guests}'),
                    hint: 'Đà Lạt, Hội An, view biển…',
                    load: (q, _) async => (await api.rentHomestays(
                      q: q.isEmpty ? null : q,
                      guests: _q.guests,
                      checkin: _iso(_q.checkIn),
                      checkout: _iso(_q.checkOut),
                      limit: 30,
                    ))
                        ?.items,
                    emptyMessage: 'Không còn chỗ trống cho ngày đã chọn',
                    emptyIcon: Icons.house_outlined,
                    itemBuilder: (context, h, _) => _HomestayCard(home: h, query: _q),
                  ),
                ),
              ],
            )),
        AppTab('Chuyến đi', Icons.luggage_outlined, (context) => AsyncList<RentBooking>(
              load: () => api.rentBookings(as_: 'guest', limit: 50),
              emptyMessage: 'Bạn chưa đặt homestay nào',
              emptyIcon: Icons.luggage_outlined,
              itemBuilder: (context, b, reload) => _RentBookingTile(booking: b, reload: reload),
            )),
        AppTab('Yêu thích', Icons.favorite_border, (context) => AsyncList<RentHomestay>(
              load: () => api.rentWishlist(limit: 50),
              emptyMessage: 'Chưa có homestay yêu thích',
              emptyIcon: Icons.favorite_border,
              itemBuilder: (context, h, reload) => Dismissible(
                key: ValueKey(h.id),
                direction: DismissDirection.endToStart,
                background: const _UnlikeBackground(),
                confirmDismiss: (_) => runAction(context, () => api.rentRemoveWishlist(id: h.id!), success: 'Đã bỏ yêu thích'),
                onDismissed: (_) => reload(),
                child: _HomestayCard(home: h, query: _q),
              ),
            )),
        AppTab('Tài khoản', Icons.person_outline, (_) => const RentAccountTab()),
      ],
    );
  }
}

String? _nightly(RentHomestay h) {
  final rate = (h.rates ?? const <RentRate>[]).where((r) => r.active != false).firstOrNull;
  if (rate == null) return null;
  return '${formatMoneyString(rate.price, rate.currency)}/${rateUnit(rate.model)}';
}

class _UnlikeBackground extends StatelessWidget {
  const _UnlikeBackground();

  @override
  Widget build(BuildContext context) => Container(
        alignment: Alignment.centerRight,
        padding: const EdgeInsets.only(right: 24),
        decoration: BoxDecoration(color: const Color(0xFFDC2626), borderRadius: BorderRadius.circular(20)),
        child: const Icon(Icons.heart_broken, color: Colors.white),
      );
}

class _HomestayCard extends StatelessWidget {
  const _HomestayCard({required this.home, required this.query});

  final RentHomestay home;
  final _StayQuery query;

  @override
  Widget build(BuildContext context) {
    final h = home;
    return KCard(
      padding: EdgeInsets.zero,
      onTap: () => openPage(context, _HomestayPage(id: h.id!, query: query)),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          SizedBox(
            height: 170,
            width: double.infinity,
            child: LayoutBuilder(
              builder: (context, c) => KThumb(url: (h.images ?? const []).firstOrNull, icon: Icons.cottage_rounded, width: c.maxWidth, height: c.maxHeight, radius: 0),
            ),
          ),
          Padding(
            padding: const EdgeInsets.all(14),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  children: [
                    Expanded(child: Text(h.name_ ?? 'Homestay', maxLines: 1, overflow: TextOverflow.ellipsis, style: const TextStyle(fontWeight: FontWeight.w800, fontSize: 16))),
                    if ((h.rating ?? 0) > 0) _stars(null, h.rating!, h.reviewCount),
                  ],
                ),
                const SizedBox(height: 2),
                Text(h.address ?? '', maxLines: 1, overflow: TextOverflow.ellipsis, style: Theme.of(context).textTheme.bodySmall),
                const SizedBox(height: 6),
                Row(
                  children: [
                    Text('${h.guests ?? '-'} khách • ${h.bedrooms ?? '-'} phòng ngủ', style: Theme.of(context).textTheme.bodySmall),
                    const Spacer(),
                    if (_nightly(h) != null) Text(_nightly(h)!, style: TextStyle(fontWeight: FontWeight.w800, color: context.appColor)),
                  ],
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}

class _HomestayPage extends StatelessWidget {
  const _HomestayPage({required this.id, required this.query});

  final int id;
  final _StayQuery query;

  @override
  Widget build(BuildContext context) {
    final api = context.api.rentHouse;
    return AsyncBuilder<RentHomestay?>(
      load: () => api.rentHomestay(id: id),
      loading: const Scaffold(body: SkeletonList()),
      builder: (context, h, _) {
        if (h == null) return const Scaffold(body: EmptyView(message: 'Không tìm thấy homestay'));
        return DetailPage(
          title: h.name_ ?? 'Homestay',
          actions: [
            ShareLinkButton(path: Uri(path: '/homestay/$id', queryParameters: query.params).toString(), title: h.name_ ?? 'Homestay', description: h.address, imageUrl: (h.images ?? const []).firstOrNull),
            IconButton(
              tooltip: 'Yêu thích',
              icon: const Icon(Icons.favorite_border),
              onPressed: () => runAction(context, () => api.rentAddWishlist(id: id), success: 'Đã lưu vào yêu thích'),
            ),
          ],
          header: HeroHeader(
            title: h.name_ ?? 'Homestay',
            subtitle: h.address,
            imageUrl: (h.images ?? const []).firstOrNull,
            icon: Icons.cottage_rounded,
            badges: [
              if (h.typeName != null) Pill(h.typeName!),
              if ((h.rating ?? 0) > 0) Pill('${h.rating!.toStringAsFixed(1)} sao (${h.reviewCount ?? 0})'),
              if (h.acceptsWalletPayment == true) const Pill('Nhận Ví Keiki', icon: Icons.account_balance_wallet_outlined),
            ],
          ),
          bottom: Row(
            children: [
              Expanded(
                child: Column(
                  mainAxisSize: MainAxisSize.min,
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(_nightly(h) ?? 'Liên hệ', style: const TextStyle(fontWeight: FontWeight.w800, fontSize: 16)),
                    Text('${shortDate(query.checkIn.toIso8601String())} • ${query.nights} đêm', style: Theme.of(context).textTheme.bodySmall),
                  ],
                ),
              ),
              FilledButton(
                onPressed: () async {
                  RentBooking? booked;
                  await showFormSheet(
                    context,
                    title: 'Đặt ${h.name_}',
                    fields: [
                      KField('checkin', 'Nhận phòng', kind: FieldKind.date, required: true, initial: _iso(query.checkIn)),
                      KField('checkout', 'Trả phòng', kind: FieldKind.date, required: true, initial: _iso(query.checkOut)),
                      KField('guests', 'Số khách', kind: FieldKind.number, required: true, initial: '${query.guests}'),
                      const KField('note', 'Ghi chú cho chủ nhà', kind: FieldKind.multiline),
                    ],
                    submitLabel: 'Giữ chỗ',
                    onSubmit: (v) async {
                      booked = await api.rentBook(
                        input: RentBookingInput(
                          homestayId: id,
                          checkinDate: v.str('checkin'),
                          checkoutDate: v.str('checkout'),
                          guests: v.integer('guests'),
                          note: v.str('note'),
                          currency: 'VND',
                          requestId: newIdempotencyKey(),
                        ),
                      );
                    },
                  );
                  final b = booked;
                  if (b?.id == null || !context.mounted) return;
                  await _payWithWallet(context, title: 'Đặt chỗ #${b!.id}', amount: formatMoneyString(b.totalAmount, b.currency), pay: () => api.rentPayBooking(id: b.id!, method: 'wallet'));
                },
                child: const Text('Đặt ngay'),
              ),
            ],
          ),
          children: [
            StatRow(
              [
                StatTile(label: 'Khách', value: '${h.guests ?? '-'}', icon: Icons.people_outline),
                StatTile(label: 'Phòng ngủ', value: '${h.bedrooms ?? '-'}', icon: Icons.bed_outlined),
                StatTile(label: 'Phòng tắm', value: '${h.bathrooms ?? '-'}', icon: Icons.bathtub_outlined),
              ],
              padding: const EdgeInsets.fromLTRB(16, 12, 16, 0),
            ),
            Paragraph(h.description, title: 'Giới thiệu'),
            HomestayExtras(home: h, onLease: () {}),
            InfoSection(title: 'Bảng giá', rows: [
              for (final r in h.rates ?? const <RentRate>[]) ('Theo ${rateUnit(r.model)}', formatMoneyString(r.price, r.currency)),
            ]),
            InfoSection(rows: [('Điện thoại', h.phoneNumber)]),
            const SectionHeader('Đánh giá'),
            AsyncBuilder<List<RentReview>?>(
              load: () => api.rentReviews(id: id, limit: 10),
              loading: const SizedBox(height: 60),
              builder: (context, reviews, _) => Column(
                children: [
                  if ((reviews ?? const []).isEmpty) const Padding(padding: EdgeInsets.all(16), child: Text('Chưa có đánh giá')),
                  for (final r in reviews ?? const <RentReview>[])
                    ListTile(
                      leading: CircleAvatar(child: Text('${r.rating ?? '-'}')),
                      title: Text(r.comment ?? ''),
                      subtitle: Text(shortDate(r.createdAt)),
                    ),
                ],
              ),
            ),
          ],
        );
      },
    );
  }
}

class _RentBookingTile extends StatelessWidget {
  const _RentBookingTile({required this.booking, required this.reload});

  final RentBooking booking;
  final Future<void> Function() reload;

  @override
  Widget build(BuildContext context) {
    final api = context.api.rentHouse;
    final b = booking;
    final status = switch (b.status) { 1 => 'confirmed', 2 => 'cancelled', 3 => 'pending', _ => null };
    return EntityTile(
      title: 'Đặt chỗ #${b.id} • Homestay #${b.homestayId}',
      subtitle: '${shortDate(b.checkinDate)} – ${shortDate(b.checkoutDate)} • ${b.nights ?? '-'} đêm • ${b.guests ?? '-'} khách',
      caption: formatMoneyString(b.totalAmount, b.currency),
      icon: Icons.cottage_rounded,
      badges: [StatusPill(status)],
      onTap: () async {
        await openPage(context, RentBookingPage(id: b.id!));
        await reload();
      },
      trailing: PopupMenuButton<String>(
        onSelected: (v) async {
          var ok = false;
          if (v == 'pay') {
            ok = await runAction(context, () => api.rentPayBooking(id: b.id!, method: 'wallet'), success: 'Thanh toán thành công');
          } else if (v == 'cancel') {
            ok = await runAction(context, () => api.rentCancelBooking(id: b.id!), confirm: 'Huỷ đặt chỗ #${b.id}?', destructive: true, success: 'Đã huỷ');
          } else {
            ok = await showFormSheet(
              context,
              title: 'Đánh giá kỳ nghỉ',
              fields: const [
                KField('rating', 'Số sao', kind: FieldKind.choice, options: ratingOptions),
                KField('comment', 'Cảm nhận', kind: FieldKind.multiline, required: true),
              ],
              success: 'Cảm ơn bạn đã đánh giá',
              onSubmit: (v) => api.rentCreateReview(id: b.homestayId!, input: RentReviewInput(bookingId: b.id, rating: v.integer('rating'), comment: v.str('comment'))),
            );
          }
          if (ok) await reload();
        },
        itemBuilder: (_) => [
          if (b.status == 3) const PopupMenuItem(value: 'pay', child: Text('Thanh toán')),
          if (b.status != 2) const PopupMenuItem(value: 'cancel', child: Text('Huỷ đặt chỗ')),
          if (b.status == 1) const PopupMenuItem(value: 'review', child: Text('Viết đánh giá')),
        ],
      ),
    );
  }
}

class ReservationApp extends StatefulWidget {
  const ReservationApp({super.key});

  @override
  State<ReservationApp> createState() => _ReservationAppState();
}

class _ReservationAppState extends State<ReservationApp> {
  final _q = _StayQuery();

  @override
  Widget build(BuildContext context) {
    final api = context.api.reservation;
    return ServiceApp(
      serviceId: 'reservation',
      adminActions: [
        Builder(
          builder: (context) => IconButton(tooltip: 'Quản lý khách sạn', icon: const Icon(Icons.business_center_outlined), onPressed: () => openPage(context, const HotelHostPage())),
        ),
      ],
      tabs: [
        AppTab('Khách sạn', Icons.hotel_outlined, (context) => Column(
              children: [
                _DatesBar(query: _q, onChanged: () => setState(() {})),
                Expanded(
                  child: SearchableList<ResvHotel>(
                    key: ValueKey('${_q.checkIn}${_q.checkOut}${_q.guests}'),
                    hint: 'Thành phố hoặc tên khách sạn',
                    load: (q, _) async => (await api.resvHotels(
                      q: q.isEmpty ? null : q,
                      guests: _q.guests,
                      checkIn: _iso(_q.checkIn),
                      checkOut: _iso(_q.checkOut),
                      limit: 30,
                    ))
                        ?.items,
                    emptyMessage: 'Không tìm thấy khách sạn phù hợp',
                    emptyIcon: Icons.hotel_outlined,
                    itemBuilder: (context, h, _) => EntityTile(
                      title: h.name_ ?? 'Khách sạn',
                      subtitle: [h.address, h.city].whereType<String>().join(', '),
                      imageUrl: (h.images ?? const []).firstOrNull,
                      icon: Icons.hotel_rounded,
                      badges: [
                        if ((h.stars ?? 0) > 0) Pill('${h.stars} sao', icon: Icons.star_rounded, color: const Color(0xFFF59E0B)),
                        if ((h.rating ?? 0) > 0) Pill('${h.rating!.toStringAsFixed(1)} điểm'),
                        if ((h.freeCancelHours ?? 0) > 0) const Pill('Huỷ miễn phí', color: Color(0xFF16A34A)),
                      ],
                      onTap: () => openPage(context, _HotelPage(id: h.id!, query: _q)),
                    ),
                  ),
                ),
              ],
            )),
        AppTab('Đặt phòng', Icons.event_note_outlined, (context) => AsyncList<ResvReservation>(
              load: () async => (await api.resvReservations(as_: 'guest', limit: 50))?.items,
              emptyMessage: 'Bạn chưa có đặt phòng nào',
              emptyIcon: Icons.event_note_outlined,
              itemBuilder: (context, r, reload) => _ResvTile(r: r, reload: reload),
            )),
        AppTab('Yêu thích', Icons.favorite_border, (context) => AsyncList<ResvHotel>(
              load: () async => (await api.resvWishlist(limit: 50))?.items,
              emptyMessage: 'Chưa có khách sạn yêu thích',
              emptyIcon: Icons.favorite_border,
              itemBuilder: (context, h, reload) => Dismissible(
                key: ValueKey(h.id),
                direction: DismissDirection.endToStart,
                background: const _UnlikeBackground(),
                confirmDismiss: (_) => runAction(context, () => api.resvRemoveFavourite(id: h.id!), success: 'Đã bỏ yêu thích'),
                onDismissed: (_) => reload(),
                child: EntityTile(
                  title: h.name_ ?? '',
                  subtitle: h.city,
                  icon: Icons.hotel_rounded,
                  imageUrl: (h.images ?? const []).firstOrNull,
                  onTap: () => openPage(context, _HotelPage(id: h.id!, query: _q)),
                ),
              ),
            )),
        AppTab('Tài khoản', Icons.person_outline, (_) => const ResvAccountTab()),
      ],
    );
  }
}

class _HotelPage extends StatelessWidget {
  const _HotelPage({required this.id, required this.query});

  final int id;
  final _StayQuery query;

  @override
  Widget build(BuildContext context) {
    final api = context.api.reservation;
    return AsyncBuilder<(ResvHotel?, List<ResvQuote>)>(
      load: () async {
        final r = await Future.wait<Object?>([
          api.resvHotel(id: id),
          api.resvQuotes(id: id, checkIn: _iso(query.checkIn), checkOut: _iso(query.checkOut), rooms: 1).catchError((_) => <ResvQuote>[]),
        ]);
        return (r[0] as ResvHotel?, (r[1] as List<ResvQuote>?) ?? const <ResvQuote>[]);
      },
      loading: const Scaffold(body: SkeletonList()),
      builder: (context, d, _) {
        final h = d.$1;
        if (h == null) return const Scaffold(body: EmptyView(message: 'Không tìm thấy khách sạn'));
        return DetailPage(
          title: h.name_ ?? 'Khách sạn',
          actions: [
            ShareLinkButton(path: Uri(path: '/hotel/$id', queryParameters: query.params).toString(), title: h.name_ ?? 'Khách sạn', description: [h.address, h.city].whereType<String>().join(', ')),
            IconButton(
              tooltip: 'Yêu thích',
              icon: const Icon(Icons.favorite_border),
              onPressed: () => runAction(context, () => api.resvAddFavourite(id: id), success: 'Đã lưu khách sạn'),
            ),
          ],
          header: HeroHeader(
            title: h.name_ ?? '',
            subtitle: [h.address, h.city].whereType<String>().join(', '),
            imageUrl: (h.images ?? const []).firstOrNull,
            icon: Icons.hotel_rounded,
            badges: [
              if ((h.stars ?? 0) > 0) Pill('${h.stars} sao', icon: Icons.star_rounded, color: const Color(0xFFF59E0B)),
              if ((h.rating ?? 0) > 0) Pill('${h.rating!.toStringAsFixed(1)} (${h.reviewCount ?? 0} đánh giá)'),
              for (final a in (h.amenities ?? const <String>[]).take(4)) Pill(a),
            ],
          ),
          children: [
            InfoSection(rows: [
              ('Nhận phòng', h.checkInTime),
              ('Trả phòng', h.checkOutTime),
              ('Huỷ miễn phí', (h.freeCancelHours ?? 0) > 0 ? 'Trước ${h.freeCancelHours} giờ' : null),
              ('Điện thoại', h.phoneNumber),
            ]),
            SectionHeader('Phòng trống ${shortDate(query.checkIn.toIso8601String())} • ${query.nights} đêm'),
            if (d.$2.isEmpty) const Padding(padding: EdgeInsets.all(16), child: Text('Không còn phòng cho ngày đã chọn')),
            for (final q in d.$2)
              Padding(
                padding: const EdgeInsets.fromLTRB(16, 0, 16, 10),
                child: KCard(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(q.roomType?.name_ ?? 'Phòng', style: const TextStyle(fontWeight: FontWeight.w800, fontSize: 16)),
                      const SizedBox(height: 4),
                      Text(
                        [
                          if (q.roomType?.capacity != null) '${q.roomType!.capacity} khách',
                          q.roomType?.bed,
                          if (q.roomType?.sizeM2 != null) '${q.roomType!.sizeM2} m²',
                          q.roomType?.view,
                        ].whereType<String>().join(' • '),
                        style: Theme.of(context).textTheme.bodySmall,
                      ),
                      const SizedBox(height: 10),
                      Row(
                        children: [
                          Expanded(
                            child: Column(
                              crossAxisAlignment: CrossAxisAlignment.start,
                              children: [
                                Text(formatMoneyString(q.total, h.currency), style: TextStyle(fontWeight: FontWeight.w800, fontSize: 18, color: context.appColor)),
                                Text('${q.nights ?? query.nights} đêm • còn ${q.roomsAvailable ?? 0} phòng', style: Theme.of(context).textTheme.bodySmall),
                              ],
                            ),
                          ),
                          if (q.bookable == false)
                            OutlinedButton(
                              onPressed: () => joinHotelWaitlist(
                                context,
                                hotelId: h.id!,
                                roomTypeId: q.roomType!.id!,
                                checkIn: _iso(query.checkIn),
                                checkOut: _iso(query.checkOut),
                                guests: query.guests,
                              ),
                              child: const Text('Chờ phòng'),
                            )
                          else
                            FilledButton(onPressed: () => _reserve(context, h, q), child: const Text('Đặt')),
                        ],
                      ),
                    ],
                  ),
                ),
              ),
            Paragraph(h.description, title: 'Giới thiệu'),
            HotelExtras(hotelId: h.id!),
          ],
        );
      },
    );
  }

  Future<void> _reserve(BuildContext context, ResvHotel h, ResvQuote q) async {
    final api = context.api.reservation;
    final addons = await pickHotelServices(context, h.id!);
    if (!context.mounted) return;
    Object? res;
    await showFormSheet(
      context,
      title: q.roomType?.name_ ?? 'Đặt phòng',
      description: '${shortDate(query.checkIn.toIso8601String())} – ${shortDate(query.checkOut.toIso8601String())}',
      fields: [
        KField('adults', 'Người lớn', kind: FieldKind.number, required: true, initial: '${query.guests}'),
        const KField('children', 'Trẻ em', kind: FieldKind.number, initial: '0'),
        const KField('promo', 'Mã khuyến mãi', scannable: true),
        const KField('requests', 'Yêu cầu đặc biệt', kind: FieldKind.multiline),
      ],
      submitLabel: 'Giữ phòng',
      onSubmit: (v) async {
        res = await api.resvReserve(
          input: ResvReservationInput(
            hotelId: h.id,
            roomTypeId: q.roomType?.id,
            checkIn: _iso(query.checkIn),
            checkOut: _iso(query.checkOut),
            rooms: 1,
            adults: v.integer('adults'),
            children: v.integer('children'),
            promoCode: v.str('promo'),
            specialRequests: v.str('requests'),
            services: addons.isEmpty ? null : addons,
            requestId: newIdempotencyKey(),
          ),
        );
      },
    );
    final rid = pickInt(res, 'id');
    if (rid == null || !context.mounted) return;
    await _payWithWallet(
      context,
      title: 'Đặt phòng #$rid',
      amount: formatMoneyString('${pick(res, 'total_amount') ?? q.total}', h.currency),
      pay: () => api.resvPayReservation(id: rid, method: 'wallet'),
    );
  }
}

class _ResvTile extends StatelessWidget {
  const _ResvTile({required this.r, required this.reload});

  final ResvReservation r;
  final Future<void> Function() reload;

  @override
  Widget build(BuildContext context) {
    final api = context.api.reservation;
    final pending = (r.statusName ?? '').toLowerCase().contains('pending');
    return EntityTile(
      title: 'Đặt phòng #${r.id}',
      subtitle: '${shortDate(r.checkIn)} – ${shortDate(r.checkOut)} • ${r.nights ?? '-'} đêm • ${r.adults ?? 0} người lớn',
      caption: formatMoneyString(r.totalAmount, r.currency),
      icon: Icons.hotel_rounded,
      badges: [StatusPill(r.statusName)],
      onTap: () async {
        await openPage(context, ResvDetailPage(id: r.id!));
        await reload();
      },
      trailing: PopupMenuButton<String>(
        onSelected: (v) async {
          var ok = false;
          if (v == 'pay') ok = await runAction(context, () => api.resvPayReservation(id: r.id!, method: 'wallet'), success: 'Thanh toán thành công');
          if (v == 'cancel') {
            final quote = await api.resvCancellationQuote(id: r.id!).catchError((_) => null);
            if (!context.mounted) return;
            ok = await runAction(
              context,
              () => api.resvCancelReservation(id: r.id!),
              confirm: quote == null ? 'Huỷ đặt phòng #${r.id}?' : 'Phí huỷ ${formatMoneyString(quote.fee, r.currency)}, hoàn lại ${formatMoneyString(quote.refund, r.currency)}. Tiếp tục huỷ?',
              destructive: true,
              success: 'Đã huỷ đặt phòng',
            );
          }
          if (v == 'review' && context.mounted) {
            ok = await showFormSheet(
              context,
              title: 'Đánh giá khách sạn',
              fields: const [
                KField('rating', 'Số sao', kind: FieldKind.choice, options: ratingOptions),
                KField('comment', 'Cảm nhận', kind: FieldKind.multiline, required: true),
              ],
              success: 'Cảm ơn bạn đã đánh giá',
              onSubmit: (v) => api.resvCreateReview(id: r.hotelId!, input: {'reservation_id': r.id, 'rating': v.integer('rating'), 'comment': v.str('comment')}),
            );
          }
          if (ok) await reload();
        },
        itemBuilder: (_) => [
          if (pending) const PopupMenuItem(value: 'pay', child: Text('Thanh toán')),
          const PopupMenuItem(value: 'cancel', child: Text('Huỷ đặt phòng')),
          const PopupMenuItem(value: 'review', child: Text('Viết đánh giá')),
        ],
      ),
    );
  }
}

class HotelLinkPage extends StatelessWidget {
  const HotelLinkPage({super.key, required this.id, this.query = const {}});

  final int id;
  final Map<String, String> query;

  @override
  Widget build(BuildContext context) => _HotelPage(id: id, query: _StayQuery.from(query));
}

class HomestayLinkPage extends StatelessWidget {
  const HomestayLinkPage({super.key, required this.id, this.query = const {}});

  final int id;
  final Map<String, String> query;

  @override
  Widget build(BuildContext context) => _HomestayPage(id: id, query: _StayQuery.from(query));
}
