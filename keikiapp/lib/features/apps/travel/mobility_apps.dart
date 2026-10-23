import 'package:flutter/material.dart';

import '../../../gen/models.g.dart';
import '../extras/extras.dart';
import '../kit/kit.dart';

String _utc(String? date) => DateTime.tryParse(date ?? '')?.toUtc().toIso8601String() ?? '';

const _vehicleTypes = {
  'Ô tô': CarRentalVehicleType.vehicleTYPECAR,
  'Xe máy': CarRentalVehicleType.vehicleTYPEMOTORCYCLE,
  'Xe đạp': CarRentalVehicleType.vehicleTYPEBICYCLE,
};

IconData _vehicleIcon(CarRentalVehicleType? t) => switch (t) {
      CarRentalVehicleType.vehicleTYPEMOTORCYCLE => Icons.two_wheeler_rounded,
      CarRentalVehicleType.vehicleTYPEBICYCLE => Icons.pedal_bike_rounded,
      _ => Icons.directions_car_rounded,
    };

Future<bool> _ensureLicense(BuildContext context) async {
  final api = context.api.carRental;
  final me = await api.carRentalGetUser(input: CarRentalGetUserRequestInput()).catchError((_) => null);
  if ((me?.drivingLicense ?? '').isNotEmpty || !context.mounted) return context.mounted;
  return showFormSheet(
    context,
    title: 'Giấy phép lái xe',
    description: 'Cần một lần trước khi đặt xe, nhân viên sẽ đối chiếu khi bạn nhận xe.',
    fields: [
      const KField('license', 'Số GPLX', required: true),
      KField('phone', 'Số điện thoại', kind: FieldKind.phone, initial: me?.phoneNumber),
    ],
    submitLabel: 'Lưu',
    onSubmit: (v) => api.carRentalUpdateUser(input: CarRentalUpdateUserRequestInput(drivingLicense: v.str('license'), phoneNumber: v.str('phone'))),
  );
}

class CarRentalApp extends StatelessWidget {
  const CarRentalApp({super.key});

  @override
  Widget build(BuildContext context) {
    final api = context.api.carRental;
    return ServiceApp(
      serviceId: 'car_rental',
      actions: [
        Builder(
          builder: (context) => IconButton(
            tooltip: 'Xe trống theo ngày',
            icon: const Icon(Icons.event_available),
            onPressed: () async {
              final now = DateTime.now();
              final range = await showDateRangePicker(context: context, firstDate: now, lastDate: now.add(const Duration(days: 180)));
              if (range == null || !context.mounted) return;
              await openPage(
                context,
                Scaffold(
                  appBar: AppBar(title: Text('Xe trống ${shortDate(range.start.toIso8601String())} – ${shortDate(range.end.toIso8601String())}')),
                  body: AsyncList<CarRentalVehicleResponse>(
                    load: () => carAvailable(context, range.start, range.end),
                    emptyMessage: 'Không còn xe trống trong khoảng này',
                    itemBuilder: (context, v, _) => _VehicleTile(vehicle: v),
                  ),
                ),
              );
            },
          ),
        ),
      ],
      adminActions: [
        Builder(builder: (context) => IconButton(tooltip: 'Quầy giao xe', icon: const Icon(Icons.badge_outlined), onPressed: () => openPage(context, const CarStaffPage()))),
      ],
      tabs: [
        AppTab('Chọn xe', Icons.directions_car_outlined, (context) => SearchableList<CarRentalVehicleResponse>(
              hint: 'Hãng hoặc mẫu xe',
              filters: () async => _vehicleTypes.keys.toList(),
              load: (q, type) async {
                final r = await api.carRentalListVehicles(input: CarRentalListVehiclesRequestInput(pageSize: 50, vehicleType: _vehicleTypes[type]));
                final items = r?.vehicles ?? const <CarRentalVehicleResponse>[];
                final needle = q.toLowerCase();
                return needle.isEmpty ? items : items.where((v) => '${v.make} ${v.model}'.toLowerCase().contains(needle)).toList();
              },
              emptyMessage: 'Chưa có xe phù hợp',
              emptyIcon: Icons.no_crash_outlined,
              itemBuilder: (context, v, _) => _VehicleTile(vehicle: v),
            )),
        AppTab('Đặt xe', Icons.event_available_outlined, (_) => const _CarReservations()),
        AppTab('Chuyến thuê', Icons.route_outlined, (_) => const _CarRentals()),
      ],
    );
  }
}

class _VehicleTile extends StatelessWidget {
  const _VehicleTile({required this.vehicle});

  final CarRentalVehicleResponse vehicle;

  @override
  Widget build(BuildContext context) {
    final v = vehicle;
    return EntityTile(
      title: '${v.make ?? ''} ${v.model ?? ''} ${v.year ?? ''}'.trim(),
      subtitle: [
        v.color,
        if (v.carDetails?.seatingCapacity != null) '${v.carDetails!.seatingCapacity} chỗ',
        v.carDetails?.transmission,
        v.carDetails?.fuelType,
      ].whereType<String>().where((s) => s.isNotEmpty).join(' • '),
      icon: _vehicleIcon(v.vehicleType),
      caption: '${vnd(v.dailyRate)}/ngày • ${vnd(v.hourlyRate)}/giờ',
      badges: [StatusPill(v.status?.value)],
      onTap: () async {
        final fresh = v.vehicleId == null ? null : await carVehicle(context, v.vehicleId!);
        if (context.mounted) await openPage(context, _VehiclePage(vehicle: fresh ?? v));
      },
    );
  }
}

class _VehiclePage extends StatelessWidget {
  const _VehiclePage({required this.vehicle});

  final CarRentalVehicleResponse vehicle;

  Future<void> _book(BuildContext context) async {
    final api = context.api.carRental;
    if (!await _ensureLicense(context) || !context.mounted) return;
    final locations = await api.carRentalListLocations(input: CarRentalListLocationsRequestInput(pageSize: 50)).catchError((_) => null);
    if (!context.mounted) return;
    final options = [for (final l in locations?.locations ?? const <CarRentalLocationResponse>[]) (l.locationId ?? '', '${l.name_ ?? ''} • ${l.city ?? ''}')];
    final tomorrow = DateTime.now().add(const Duration(days: 1));
    await showFormSheet(
      context,
      title: 'Đặt ${vehicle.make} ${vehicle.model}',
      fields: [
        KField('start', 'Ngày nhận xe', kind: FieldKind.date, required: true, initial: tomorrow.toIso8601String().substring(0, 10)),
        KField('end', 'Ngày trả xe', kind: FieldKind.date, required: true, initial: tomorrow.add(const Duration(days: 2)).toIso8601String().substring(0, 10)),
        if (options.isNotEmpty) ...[
          KField('pickup', 'Nơi nhận xe', kind: FieldKind.choice, options: options, initial: vehicle.locationId ?? options.first.$1),
          KField('return', 'Nơi trả xe', kind: FieldKind.choice, options: options, initial: vehicle.locationId ?? options.first.$1),
        ],
      ],
      submitLabel: 'Đặt xe',
      success: 'Đặt xe thành công, xem trong mục Đặt xe',
      onSubmit: (v) => api.carRentalCreateReservation(
        input: CarRentalCreateReservationRequestInput(
          vehicleId: vehicle.vehicleId,
          startTime: _utc(v.str('start')),
          endTime: _utc(v.str('end')),
          pickupLocationId: v.str('pickup') ?? vehicle.locationId,
          returnLocationId: v.str('return') ?? vehicle.locationId,
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final v = vehicle;
    final api = context.api.carRental;
    return DetailPage(
      title: '${v.make ?? ''} ${v.model ?? ''}',
      actions: [ShareLinkButton(path: '/vehicle/${v.vehicleId}', title: '${v.make ?? ''} ${v.model ?? ''}'.trim())],
      header: HeroHeader(
        title: '${v.make ?? ''} ${v.model ?? ''}',
        subtitle: '${v.year ?? ''} • ${v.color ?? ''} • ${v.registrationNumber ?? ''}',
        icon: _vehicleIcon(v.vehicleType),
        badges: [StatusPill(v.status?.value), if (v.carDetails?.airConditioning == true) const Pill('Điều hoà', icon: Icons.ac_unit)],
      ),
      bottom: Row(
        children: [
          Expanded(child: Text('${vnd(v.dailyRate)}/ngày', style: const TextStyle(fontWeight: FontWeight.w800, fontSize: 18))),
          FilledButton(
            onPressed: v.status == CarRentalVehicleStatus.vehicleSTATUSAVAILABLE || v.status == null ? () => _book(context) : null,
            child: const Text('Đặt xe'),
          ),
        ],
      ),
      children: [
        InfoSection(rows: [
          ('Giá theo giờ', vnd(v.hourlyRate)),
          ('Giá theo ngày', vnd(v.dailyRate)),
          ('Số km đã đi', v.mileage == null ? null : '${compact(v.mileage)} km'),
          ('Nhiên liệu', v.carDetails?.fuelType),
          ('Hộp số', v.carDetails?.transmission),
          ('Số chỗ', v.carDetails?.seatingCapacity?.toString()),
          ('Dung tích máy', v.motorcycleDetails?.engineCapacity?.toString()),
        ]),
        const SectionHeader('Đánh giá'),
        AsyncBuilder<CarRentalListReviewsResponse?>(
          load: () => api.carRentalListVehicleReviews(input: CarRentalListVehicleReviewsRequestInput(vehicleId: v.vehicleId, pageSize: 10)),
          loading: const SizedBox(height: 60),
          builder: (context, r, _) => Column(
            children: [
              if ((r?.reviews ?? const []).isEmpty) const Padding(padding: EdgeInsets.all(16), child: Text('Chưa có đánh giá')),
              for (final rv in r?.reviews ?? const <CarRentalReviewResponse>[])
                ListTile(
                  leading: CircleAvatar(child: Text('${rv.rating ?? '-'}')),
                  title: Text(rv.comment ?? ''),
                  subtitle: Text('${rv.user?.firstName ?? ''} • ${shortDate(rv.createdAt)}'),
                ),
            ],
          ),
        ),
      ],
    );
  }
}

class _CarReservations extends StatelessWidget {
  const _CarReservations();

  @override
  Widget build(BuildContext context) {
    final api = context.api.carRental;
    return AsyncList<CarRentalReservationResponse>(
      load: () async => (await api.carRentalListUserReservations(input: CarRentalListUserReservationsRequestInput(pageSize: 50)))?.reservations,
      emptyMessage: 'Bạn chưa đặt xe nào',
      emptyIcon: Icons.event_available_outlined,
      itemBuilder: (context, r, reload) => EntityTile(
        onTap: () => carReservationDetail(context, r.reservationId!, reload),
        title: '${r.vehicle?.make ?? 'Xe'} ${r.vehicle?.model ?? ''}',
        subtitle: '${formatDate(r.startTime)} – ${formatDate(r.endTime)}',
        icon: _vehicleIcon(r.vehicle?.vehicleType),
        badges: [StatusPill(r.status?.value)],
        trailing: r.status == CarRentalReservationStatus.reservationSTATUSPENDING || r.status == CarRentalReservationStatus.reservationSTATUSCONFIRMED
            ? IconButton(
                tooltip: 'Huỷ',
                icon: const Icon(Icons.cancel_outlined),
                onPressed: () async {
                  if (await runAction(
                    context,
                    () => api.carRentalCancelReservation(input: CarRentalCancelReservationRequestInput(reservationId: r.reservationId, cancellationReason: 'Khách huỷ trên app')),
                    confirm: 'Huỷ đặt xe này?',
                    destructive: true,
                    success: 'Đã huỷ đặt xe',
                  )) {
                    await reload();
                  }
                },
              )
            : null,
      ),
    );
  }
}

class _CarRentals extends StatelessWidget {
  const _CarRentals();

  @override
  Widget build(BuildContext context) {
    final api = context.api.carRental;
    return AsyncList<CarRentalRentalResponse>(
      load: () async => (await api.carRentalListUserRentals(input: CarRentalListUserRentalsRequestInput(pageSize: 50)))?.rentals,
      emptyMessage: 'Chưa có chuyến thuê nào',
      emptyIcon: Icons.route_outlined,
      itemBuilder: (context, r, reload) {
        final total = (r.baseFee ?? 0) + (r.additionalFees ?? 0);
        final unpaid = r.paymentStatus == CarRentalPaymentStatus.paymentSTATUSPENDING;
        return EntityTile(
          onTap: () => carRentalDetail(context, r.rentalId!),
          title: '${r.vehicle?.make ?? 'Xe'} ${r.vehicle?.model ?? ''}',
          subtitle: 'Nhận ${formatDate(r.pickupTime)}${r.actualReturnTime == null ? '' : '\nTrả ${formatDate(r.actualReturnTime)}'}',
          caption: vnd(total),
          icon: _vehicleIcon(r.vehicle?.vehicleType),
          badges: [StatusPill(r.status?.value), StatusPill(r.paymentStatus?.value)],
          trailing: PopupMenuButton<String>(
            onSelected: (v) async {
              var ok = false;
              if (v == 'pay') {
                ok = await runAction(
                  context,
                  () => api.carRentalProcessPayment(
                    input: CarRentalProcessPaymentRequestInput(rentalId: r.rentalId, amount: total, paymentMethod: CarRentalPaymentMethod.paymentMETHODONLINE, transactionId: newIdempotencyKey()),
                  ),
                  success: 'Thanh toán thành công',
                );
              } else {
                ok = await showFormSheet(
                  context,
                  title: 'Đánh giá chuyến đi',
                  fields: const [
                    KField('rating', 'Số sao', kind: FieldKind.choice, options: ratingOptions),
                    KField('comment', 'Nhận xét', kind: FieldKind.multiline, required: true),
                  ],
                  success: 'Cảm ơn bạn đã đánh giá',
                  onSubmit: (v) => api.carRentalSubmitReview(
                    input: CarRentalSubmitReviewRequestInput(rentalId: r.rentalId, vehicleId: r.vehicleId, rating: v.integer('rating'), comment: v.str('comment')),
                  ),
                );
              }
              if (ok) await reload();
            },
            itemBuilder: (_) => [
              if (unpaid) const PopupMenuItem(value: 'pay', child: Text('Thanh toán')),
              const PopupMenuItem(value: 'review', child: Text('Đánh giá')),
            ],
          ),
        );
      },
    );
  }
}

class BookingMiniApp extends StatefulWidget {
  const BookingMiniApp({super.key});

  @override
  State<BookingMiniApp> createState() => _BookingMiniAppState();
}

class _BookingMiniAppState extends State<BookingMiniApp> {
  DateTimeRange _range = DateTimeRange(start: DateTime.now().add(const Duration(days: 1)), end: DateTime.now().add(const Duration(days: 2)));

  String _d(DateTime d) => d.toIso8601String().substring(0, 10);

  Future<int?> _passenger() async {
    final api = context.api.bookingMini;
    final name = context.userName.isEmpty ? 'Khách Keiki' : context.userName;
    final mine = await api.myBookingMiniPassengers();
    if ((mine ?? const []).isNotEmpty) return mine!.first.id;
    return (await api.bookingMiniRegisterPassenger(name_: name))?.id;
  }

  Future<void> _book(BookingMiniRoom room) async {
    final api = context.api.bookingMini;
    final ok = await runAction(
      context,
      () async {
        final pid = await _passenger();
        if (pid == null) throw Exception('Không đăng ký được khách lưu trú');
        return api.bookingMiniCreateBooking(input: BookingMiniBookingInput(passengerId: pid, roomNo: room.roomNo, fromDate: _d(_range.start), toDate: _d(_range.end)));
      },
      confirm: 'Đặt phòng ${room.roomNo} từ ${shortDate(_range.start.toIso8601String())} đến ${shortDate(_range.end.toIso8601String())}?',
      success: 'Đặt phòng thành công',
    );
    if (ok && mounted) setState(() {});
  }

  @override
  Widget build(BuildContext context) {
    final api = context.api.bookingMini;
    return ServiceApp(
      serviceId: 'booking_mini',
      tabs: [
        AppTab('Phòng trống', Icons.meeting_room_outlined, (context) => Column(
              children: [
                Padding(
                  padding: const EdgeInsets.fromLTRB(16, 12, 16, 0),
                  child: SizedBox(
                    width: double.infinity,
                    child: OutlinedButton.icon(
                      icon: const Icon(Icons.date_range),
                      label: Text('${shortDate(_range.start.toIso8601String())} – ${shortDate(_range.end.toIso8601String())}'),
                      onPressed: () async {
                        final now = DateTime.now();
                        final r = await showDateRangePicker(context: context, firstDate: now, lastDate: now.add(const Duration(days: 365)), initialDateRange: _range);
                        if (r != null) setState(() => _range = r);
                      },
                    ),
                  ),
                ),
                Expanded(
                  child: AsyncList<BookingMiniRoom>(
                    key: ValueKey(_range),
                    load: () => api.bookingMiniAvailableRooms(fromDate: _d(_range.start), toDate: _d(_range.end)),
                    emptyMessage: 'Hết phòng trong khoảng ngày này',
                    emptyIcon: Icons.no_meeting_room_outlined,
                    itemBuilder: (context, r, _) => EntityTile(
                      title: 'Phòng ${r.roomNo} • ${r.roomType ?? ''}',
                      subtitle: [r.roomTypeDetails?.description, '${r.bedCount ?? 1} giường ${r.bedType ?? ''}', if (r.roomTypeDetails?.capacity != null) '${r.roomTypeDetails!.capacity} khách']
                          .whereType<String>()
                          .join(' • '),
                      caption: r.amenities,
                      icon: Icons.king_bed_rounded,
                      trailing: FilledButton(onPressed: () => _book(r), child: const Text('Đặt')),
                    ),
                  ),
                ),
              ],
            )),
        AppTab('Của tôi', Icons.bookmark_border, (context) => AsyncList<BookingMiniBooking>(
              load: () => api.myBookingMiniBookings(),
              emptyMessage: 'Bạn chưa đặt phòng nào',
              emptyIcon: Icons.bookmark_border,
              itemBuilder: (context, b, reload) => EntityTile(
                onTap: () async {
                  final r = await Future.wait<Object?>([api.bookingMiniBooking(id: b.id!), api.bookingMiniPassenger(id: b.passengerId!)]).catchError((_) => <Object?>[null, null]);
                  final full = r[0] as BookingMiniBooking?;
                  final guest = r[1] as BookingMiniPassenger?;
                  if (!context.mounted || full == null) return;
                  await showResult(
                    context,
                    icon: Icons.king_bed_rounded,
                    title: 'Phòng ${full.roomNo} • ${full.room?.roomType ?? ''}',
                    message: '${shortDate(full.fromDate)} – ${shortDate(full.toDate)}\nKhách: ${guest?.name_ ?? full.passenger?.name_ ?? ''}\n${full.room?.amenities ?? ''}',
                  );
                },
                title: 'Đặt phòng #${b.id} • Phòng ${b.roomNo}',
                subtitle: '${shortDate(b.fromDate)} – ${shortDate(b.toDate)}',
                caption: b.room?.roomType,
                icon: Icons.king_bed_rounded,
                trailing: IconButton(
                  tooltip: 'Huỷ',
                  icon: const Icon(Icons.cancel_outlined),
                  onPressed: () async {
                    if (await runAction(context, () => api.bookingMiniCancelBooking(id: b.id!), confirm: 'Huỷ đặt phòng #${b.id}?', destructive: true, success: 'Đã huỷ')) {
                      await reload();
                    }
                  },
                ),
              ),
            )),
      ],
    );
  }
}

class ParkingApp extends StatelessWidget {
  const ParkingApp({super.key});

  @override
  Widget build(BuildContext context) {
    final api = context.api.parkingLot;
    return ServiceApp(
      serviceId: 'parking_lot',
      tabs: [
        AppTab('Bãi xe', Icons.local_parking, (context) => AsyncBuilder<(List<ParkingSpotAvailability>, List<ParkingRate>)>(
              load: () async {
                final r = await Future.wait<Object?>([api.parkingAvailableSpots(), api.parkingRates()]);
                return ((r[0] as List<ParkingSpotAvailability>?) ?? const <ParkingSpotAvailability>[], (r[1] as List<ParkingRate>?) ?? const <ParkingRate>[]);
              },
              builder: (context, d, reload) => RefreshIndicator(
                onRefresh: reload,
                child: ListView(
                  padding: const EdgeInsets.only(bottom: 24),
                  children: [
                    AppBanner(
                      icon: Icons.local_parking_rounded,
                      title: '${d.$1.fold<int>(0, (a, s) => a + (s.available ?? 0))} chỗ trống',
                      subtitle: 'Cập nhật theo thời gian thực',
                      children: [
                        Row(
                          children: [
                            Expanded(
                              child: FilledButton.icon(
                                style: FilledButton.styleFrom(backgroundColor: Colors.white, foregroundColor: context.appColor),
                                icon: const Icon(Icons.login),
                                label: const Text('Gửi xe'),
                                onPressed: () => _checkIn(context, d.$2),
                              ),
                            ),
                            const SizedBox(width: 8),
                            Expanded(
                              child: FilledButton.tonalIcon(
                                icon: const Icon(Icons.logout),
                                label: const Text('Lấy xe'),
                                onPressed: () => _checkOut(context),
                              ),
                            ),
                          ],
                        ),
                      ],
                    ),
                    const SectionHeader('Chỗ trống theo loại xe'),
                    for (final s in d.$1)
                      Padding(
                        padding: const EdgeInsets.fromLTRB(16, 0, 16, 10),
                        child: KCard(
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              Row(
                                children: [
                                  Expanded(child: Text(_vehicleName(s.type_), style: const TextStyle(fontWeight: FontWeight.w700))),
                                  Text('${s.available ?? 0}/${s.total ?? 0}', style: TextStyle(fontWeight: FontWeight.w800, color: context.appColor)),
                                ],
                              ),
                              const SizedBox(height: 8),
                              ClipRRect(
                                borderRadius: BorderRadius.circular(99),
                                child: LinearProgressIndicator(value: (s.total ?? 0) == 0 ? 0 : (s.available ?? 0) / s.total!, minHeight: 8, color: context.appColor),
                              ),
                            ],
                          ),
                        ),
                      ),
                    const SectionHeader('Bảng giá'),
                    InfoSection(rows: [
                      for (final r in d.$2) (_vehicleName(r.vehicleType), '${vnd(r.hourlyRate)}/giờ • tối đa ${vnd(r.dailyMaxRate)}/ngày'),
                    ]),
                  ],
                ),
              ),
            )),
        AppTab('Lịch sử', Icons.history, (_) => const _ParkingHistory()),
      ],
    );
  }

  Future<void> _checkIn(BuildContext context, List<ParkingRate> rates) async {
    final types = {for (final r in rates) r.vehicleType ?? ''}..remove('');
    final key = 'parking_plate_${context.userId}';
    ParkingTicket? ticket;
    await showFormSheet(
      context,
      title: 'Gửi xe',
      fields: [
        const KField('plate', 'Biển số xe', required: true, icon: Icons.pin_outlined),
        KField('type', 'Loại xe', kind: FieldKind.choice, options: [for (final t in types.isEmpty ? {'car', 'motorbike'} : types) (t, _vehicleName(t))]),
      ],
      submitLabel: 'Lấy vé',
      onSubmit: (v) async {
        ticket = await context.api.parkingLot.parkingCheckIn(licensePlate: v.str('plate')!.toUpperCase(), vehicleType: v.str('type') ?? 'car');
        await LocalPrefs.set(key, v.str('plate')!.toUpperCase());
      },
    );
    final t = ticket;
    if (t == null || !context.mounted) return;
    await showResult(context, title: 'Vé gửi xe', message: 'Mã vé: ${t.ticketID}\nVị trí: ${t.spotID ?? '-'}\nVào lúc: ${formatDate(t.parkedTime)}', icon: Icons.local_parking_rounded);
  }

  Future<void> _checkOut(BuildContext context) => showFormSheet(
        context,
        title: 'Lấy xe & thanh toán',
        fields: const [
          KField('ticket', 'Mã vé gửi xe', required: true, icon: Icons.confirmation_number_outlined, scannable: true),
          KField('method', 'Thanh toán', kind: FieldKind.choice, options: [('wallet', 'Ví Keiki'), ('cash', 'Tiền mặt'), ('card', 'Thẻ')]),
        ],
        submitLabel: 'Lấy xe',
        success: 'Đã lấy xe. Chúc bạn thượng lộ bình an!',
        onSubmit: (v) => context.api.parkingLot.parkingCheckOut(ticketId: v.str('ticket')!, method: v.str('method') ?? 'wallet'),
      );
}

String _vehicleName(String? t) => switch ((t ?? '').toLowerCase()) {
      'car' || 'auto' || 'oto' => 'Ô tô',
      'motorbike' || 'motorcycle' || 'moto' => 'Xe máy',
      'bicycle' || 'bike' => 'Xe đạp',
      'truck' => 'Xe tải',
      'ev' || 'electric' => 'Xe điện',
      _ => t ?? 'Khác',
    };

class _ParkingHistory extends StatefulWidget {
  const _ParkingHistory();

  @override
  State<_ParkingHistory> createState() => _ParkingHistoryState();
}

class _ParkingHistoryState extends State<_ParkingHistory> {
  String? _plate;

  @override
  void initState() {
    super.initState();
    LocalPrefs.get('parking_plate_${context.userId}').then((p) {
      if (mounted && p != null) setState(() => _plate = p);
    });
  }

  @override
  Widget build(BuildContext context) {
    return Column(
      children: [
        Padding(
          padding: const EdgeInsets.fromLTRB(16, 14, 16, 0),
          child: TextField(
            key: ValueKey(_plate),
            controller: TextEditingController(text: _plate),
            textCapitalization: TextCapitalization.characters,
            decoration: const InputDecoration(hintText: 'Nhập biển số, ví dụ 51H-123.45', prefixIcon: Icon(Icons.pin_outlined), isDense: true),
            onSubmitted: (v) => setState(() => _plate = v.trim().toUpperCase()),
          ),
        ),
        Expanded(
          child: _plate == null || _plate!.isEmpty
              ? const EmptyView(message: 'Nhập biển số để xem lịch sử gửi xe', icon: Icons.directions_car_outlined)
              : AsyncList<ParkingTicket>(
                  key: ValueKey(_plate),
                  load: () => context.api.parkingLot.parkingHistory(plate: _plate!, limit: 50),
                  emptyMessage: 'Chưa có lượt gửi xe',
                  itemBuilder: (context, t, _) => EntityTile(
                    onTap: () async {
                      final full = await context.api.parkingLot.parkingTicket(id: t.ticketID!).catchError((_) => null);
                      if (!context.mounted || full == null) return;
                      await showResult(
                        context,
                        icon: Icons.local_parking_rounded,
                        title: 'Vé ${full.ticketID}',
                        message: 'Biển số ${full.vehiclePlate}\nVị trí ${full.spotID ?? '-'} • cổng vào ${full.entryGateID ?? '-'}\nVào ${formatDate(full.parkedTime)}${full.leaveTime == null ? '' : '\nRa ${formatDate(full.leaveTime)} qua cổng ${full.exitGateID ?? '-'}'}',
                      );
                    },
                    title: 'Vé ${t.ticketID ?? ''}',
                    subtitle: 'Vào ${formatDate(t.parkedTime)}${t.leaveTime == null ? '' : '\nRa ${formatDate(t.leaveTime)}'}',
                    caption: t.spotID == null ? null : 'Vị trí ${t.spotID}',
                    icon: Icons.local_parking_rounded,
                    badges: [StatusPill(t.status)],
                  ),
                ),
        ),
      ],
    );
  }
}

class VehicleLinkPage extends StatelessWidget {
  const VehicleLinkPage({super.key, required this.id});

  final String id;

  @override
  Widget build(BuildContext context) => AsyncBuilder<CarRentalVehicleResponse?>(
        load: () => context.api.carRental.carRentalGetVehicle(input: CarRentalGetVehicleRequestInput(vehicleId: id)),
        loading: const Scaffold(body: SkeletonList()),
        builder: (context, v, _) => v == null ? const DetailPage(title: 'Thuê xe', children: [EmptyView(message: 'Không tìm thấy xe')]) : _VehiclePage(vehicle: v),
      );
}
