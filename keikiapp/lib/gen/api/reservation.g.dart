// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

import '../gateway_executor.dart';
import '../models.g.dart';

class ReservationApi {
  const ReservationApi(this._executor);

  final GatewayExecutor _executor;

  Future<ResvHotelPage?> resvHotels({String? city, String? q, int? rooms, int? guests, String? sort, int? limit, String? minRate, String? maxRate, String? amenities, String? cursor, String? checkIn, String? checkOut, String? selectionSet}) async {
    final data = await _executor.public('query resvHotels(\$city: String, \$q: String, \$rooms: Int, \$guests: Int, \$sort: String, \$limit: Int, \$min_rate: String, \$max_rate: String, \$amenities: String, \$cursor: String, \$check_in: String, \$check_out: String) { resvHotels(city: \$city, q: \$q, rooms: \$rooms, guests: \$guests, sort: \$sort, limit: \$limit, min_rate: \$min_rate, max_rate: \$max_rate, amenities: \$amenities, cursor: \$cursor, check_in: \$check_in, check_out: \$check_out) ${selectionSet ?? '{ items { id name description stars city address phone_number email images amenities check_in_time check_out_time currency free_cancel_hours cancellation_policy { hours_before fee_percent } status status_name review_note manager_id accepts_wallet_payment rating review_count room_types { id name description capacity bed size_m2 view amenities active } created_at } has_more next_cursor }'} }', variables: {
      'city': ?city,
      'q': ?q,
      'rooms': ?rooms,
      'guests': ?guests,
      'sort': ?sort,
      'limit': ?limit,
      'min_rate': ?minRate,
      'max_rate': ?maxRate,
      'amenities': ?amenities,
      'cursor': ?cursor,
      'check_in': ?checkIn,
      'check_out': ?checkOut,
    }
    );
    return data['resvHotels'] == null ? null : ResvHotelPage.fromJson((data['resvHotels'] as Map).cast<String, dynamic>());
  }

  Future<ResvHotel?> resvHotel({required int id, String? selectionSet}) async {
    final data = await _executor.public('query resvHotel(\$id: Int!) { resvHotel(id: \$id) ${selectionSet ?? '{ id name description stars city address phone_number email images amenities check_in_time check_out_time currency free_cancel_hours cancellation_policy { hours_before fee_percent } status status_name review_note manager_id accepts_wallet_payment rating review_count room_types { id name description capacity bed size_m2 view amenities active } created_at }'} }', variables: {
      'id': id,
    }
    );
    return data['resvHotel'] == null ? null : ResvHotel.fromJson((data['resvHotel'] as Map).cast<String, dynamic>());
  }

  Future<List<ResvRoomType>?> resvRoomTypes({required int id, String? selectionSet}) async {
    final data = await _executor.public('query resvRoomTypes(\$id: Int!) { resvRoomTypes(id: \$id) ${selectionSet ?? '{ id name description capacity bed size_m2 view amenities active }'} }', variables: {
      'id': id,
    }
    );
    return (data['resvRoomTypes'] as List?)?.map((e) => (e == null ? null : ResvRoomType.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<List<ResvInventory>?> resvInventory({required int id, required int type_, String? from, String? to, String? selectionSet}) async {
    final data = await _executor.public('query resvInventory(\$id: Int!, \$type: Int!, \$from: String, \$to: String) { resvInventory(id: \$id, type: \$type, from: \$from, to: \$to) ${selectionSet ?? '{ date total_inventory total_reserved left rate }'} }', variables: {
      'id': id,
      'type': type_,
      'from': ?from,
      'to': ?to,
    }
    );
    return (data['resvInventory'] as List?)?.map((e) => (e == null ? null : ResvInventory.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<List<ResvQuote>?> resvQuotes({required int id, required String checkIn, required String checkOut, int? rooms, String? selectionSet}) async {
    final data = await _executor.public('query resvQuotes(\$id: Int!, \$check_in: String!, \$check_out: String!, \$rooms: Int) { resvQuotes(id: \$id, check_in: \$check_in, check_out: \$check_out, rooms: \$rooms) ${selectionSet ?? '{ room_type { id name description capacity bed size_m2 view amenities active } nights rooms_available total bookable }'} }', variables: {
      'id': id,
      'check_in': checkIn,
      'check_out': checkOut,
      'rooms': ?rooms,
    }
    );
    return (data['resvQuotes'] as List?)?.map((e) => (e == null ? null : ResvQuote.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<List<ResvService>?> resvServices({required int id, String? selectionSet}) async {
    final data = await _executor.public('query resvServices(\$id: Int!) { resvServices(id: \$id) ${selectionSet ?? '{ id name description price unit active }'} }', variables: {
      'id': id,
    }
    );
    return (data['resvServices'] as List?)?.map((e) => (e == null ? null : ResvService.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<ResvReviewPage?> resvReviews({required int id, String? cursor, int? limit, String? selectionSet}) async {
    final data = await _executor.public('query resvReviews(\$id: Int!, \$cursor: String, \$limit: Int) { resvReviews(id: \$id, cursor: \$cursor, limit: \$limit) ${selectionSet ?? '{ items { id hotel_id user_id reservation_id rating comment created_at } has_more next_cursor }'} }', variables: {
      'id': id,
      'cursor': ?cursor,
      'limit': ?limit,
    }
    );
    return data['resvReviews'] == null ? null : ResvReviewPage.fromJson((data['resvReviews'] as Map).cast<String, dynamic>());
  }

  Future<ResvHotelPage?> resvMyHotels({int? limit, String? selectionSet}) async {
    final data = await _executor.query('query resvMyHotels(\$limit: Int) { resvMyHotels(limit: \$limit) ${selectionSet ?? '{ items { id name description stars city address phone_number email images amenities check_in_time check_out_time currency free_cancel_hours cancellation_policy { hours_before fee_percent } status status_name review_note manager_id accepts_wallet_payment rating review_count room_types { id name description capacity bed size_m2 view amenities active } created_at } has_more next_cursor }'} }', variables: {
      'limit': ?limit,
    }
    );
    return data['resvMyHotels'] == null ? null : ResvHotelPage.fromJson((data['resvMyHotels'] as Map).cast<String, dynamic>());
  }

  Future<ResvHotelPage?> resvHotelReviewQueue({String? status, int? limit, String? selectionSet}) async {
    final data = await _executor.query('query resvHotelReviewQueue(\$status: String, \$limit: Int) { resvHotelReviewQueue(status: \$status, limit: \$limit) ${selectionSet ?? '{ items { id name description stars city address phone_number email images amenities check_in_time check_out_time currency free_cancel_hours cancellation_policy { hours_before fee_percent } status status_name review_note manager_id accepts_wallet_payment rating review_count room_types { id name description capacity bed size_m2 view amenities active } created_at } has_more next_cursor }'} }', variables: {
      'status': ?status,
      'limit': ?limit,
    }
    );
    return data['resvHotelReviewQueue'] == null ? null : ResvHotelPage.fromJson((data['resvHotelReviewQueue'] as Map).cast<String, dynamic>());
  }

  Future<List<ResvRoom>?> resvRooms({required int id, int? roomTypeId, String? selectionSet}) async {
    final data = await _executor.query('query resvRooms(\$id: Int!, \$room_type_id: Int) { resvRooms(id: \$id, room_type_id: \$room_type_id) ${selectionSet ?? '{ id room_type_id name floor available }'} }', variables: {
      'id': id,
      'room_type_id': ?roomTypeId,
    }
    );
    return (data['resvRooms'] as List?)?.map((e) => (e == null ? null : ResvRoom.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<List<ResvPromotion>?> resvPromotions({required int id, String? selectionSet}) async {
    final data = await _executor.query('query resvPromotions(\$id: Int!) { resvPromotions(id: \$id) ${selectionSet ?? '{ id code description percent_off amount_off min_nights max_uses used_count valid_from valid_to active }'} }', variables: {
      'id': id,
    }
    );
    return (data['resvPromotions'] as List?)?.map((e) => (e == null ? null : ResvPromotion.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<ResvReport?> resvHotelReport({required int id, String? from, String? to, String? selectionSet}) async {
    final data = await _executor.query('query resvHotelReport(\$id: Int!, \$from: String, \$to: String) { resvHotelReport(id: \$id, from: \$from, to: \$to) ${selectionSet ?? '{ from to occupancy { room_nights_available room_nights_sold occupancy_percent days { date rooms reserved occupancy_percent } } revenue { currency paid_reservations paid_revenue refunded_reservations refunded_amount cancellation_fees_kept net_revenue } reservations_by_status no_shows }'} }', variables: {
      'id': id,
      'from': ?from,
      'to': ?to,
    }
    );
    return data['resvHotelReport'] == null ? null : ResvReport.fromJson((data['resvHotelReport'] as Map).cast<String, dynamic>());
  }

  Future<ResvReservationPage?> resvReservations({String? status, String? cursor, String? as_, int? hotelId, int? limit, String? selectionSet}) async {
    final data = await _executor.query('query resvReservations(\$status: String, \$cursor: String, \$as: String, \$hotel_id: Int, \$limit: Int) { resvReservations(status: \$status, cursor: \$cursor, as: \$as, hotel_id: \$hotel_id, limit: \$limit) ${selectionSet ?? '{ items { id hotel_id room_type_id guest_id check_in check_out nights rooms adults children status status_name status_note currency room_total promo_code discount services_total total_amount cancellation_fee refund_amount stay_status checked_in_at checked_out_at special_requests services { service_id name quantity unit_price amount } request_id payment_method payment_ref paid_at expires_at created_at } has_more next_cursor }'} }', variables: {
      'status': ?status,
      'cursor': ?cursor,
      'as': ?as_,
      'hotel_id': ?hotelId,
      'limit': ?limit,
    }
    );
    return data['resvReservations'] == null ? null : ResvReservationPage.fromJson((data['resvReservations'] as Map).cast<String, dynamic>());
  }

  Future<ResvReservation?> resvReservation({required int id, String? selectionSet}) async {
    final data = await _executor.query('query resvReservation(\$id: Int!) { resvReservation(id: \$id) ${selectionSet ?? '{ id hotel_id room_type_id guest_id check_in check_out nights rooms adults children status status_name status_note currency room_total promo_code discount services_total total_amount cancellation_fee refund_amount stay_status checked_in_at checked_out_at special_requests services { service_id name quantity unit_price amount } request_id payment_method payment_ref paid_at expires_at created_at }'} }', variables: {
      'id': id,
    }
    );
    return data['resvReservation'] == null ? null : ResvReservation.fromJson((data['resvReservation'] as Map).cast<String, dynamic>());
  }

  Future<ResvCancellationQuote?> resvCancellationQuote({required int id, String? selectionSet}) async {
    final data = await _executor.query('query resvCancellationQuote(\$id: Int!) { resvCancellationQuote(id: \$id) ${selectionSet ?? '{ status_name hours_before_check_in fee_percent fee refund cancellable reason policy { hours_before fee_percent } }'} }', variables: {
      'id': id,
    }
    );
    return data['resvCancellationQuote'] == null ? null : ResvCancellationQuote.fromJson((data['resvCancellationQuote'] as Map).cast<String, dynamic>());
  }

  Future<List<ResvHistoryEntry>?> resvReservationHistory({required int id, String? selectionSet}) async {
    final data = await _executor.query('query resvReservationHistory(\$id: Int!) { resvReservationHistory(id: \$id) ${selectionSet ?? '{ at by event from_status to_status note }'} }', variables: {
      'id': id,
    }
    );
    return (data['resvReservationHistory'] as List?)?.map((e) => (e == null ? null : ResvHistoryEntry.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<ResvWaitPage?> resvWaitlist({String? cursor, int? limit, String? selectionSet}) async {
    final data = await _executor.query('query resvWaitlist(\$cursor: String, \$limit: Int) { resvWaitlist(cursor: \$cursor, limit: \$limit) ${selectionSet ?? '{ items { id hotel_id room_type_id check_in check_out rooms adults children state reservation_id created_at offered_at } has_more next_cursor }'} }', variables: {
      'cursor': ?cursor,
      'limit': ?limit,
    }
    );
    return data['resvWaitlist'] == null ? null : ResvWaitPage.fromJson((data['resvWaitlist'] as Map).cast<String, dynamic>());
  }

  Future<ResvHotelPage?> resvWishlist({String? cursor, int? limit, String? selectionSet}) async {
    final data = await _executor.query('query resvWishlist(\$cursor: String, \$limit: Int) { resvWishlist(cursor: \$cursor, limit: \$limit) ${selectionSet ?? '{ items { id name description stars city address phone_number email images amenities check_in_time check_out_time currency free_cancel_hours cancellation_policy { hours_before fee_percent } status status_name review_note manager_id accepts_wallet_payment rating review_count room_types { id name description capacity bed size_m2 view amenities active } created_at } has_more next_cursor }'} }', variables: {
      'cursor': ?cursor,
      'limit': ?limit,
    }
    );
    return data['resvWishlist'] == null ? null : ResvHotelPage.fromJson((data['resvWishlist'] as Map).cast<String, dynamic>());
  }

  Future<ResvNotificationPage?> resvNotifications({String? cursor, bool? unread, int? limit, String? selectionSet}) async {
    final data = await _executor.query('query resvNotifications(\$cursor: String, \$unread: Boolean, \$limit: Int) { resvNotifications(cursor: \$cursor, unread: \$unread, limit: \$limit) ${selectionSet ?? '{ items { id kind title body reservation_id hotel_id created_at read } has_more next_cursor }'} }', variables: {
      'cursor': ?cursor,
      'unread': ?unread,
      'limit': ?limit,
    }
    );
    return data['resvNotifications'] == null ? null : ResvNotificationPage.fromJson((data['resvNotifications'] as Map).cast<String, dynamic>());
  }

  Future<ResvUnreadCount?> resvUnreadNotifications({String? selectionSet}) async {
    final data = await _executor.query('query resvUnreadNotifications { resvUnreadNotifications ${selectionSet ?? '{ unread }'} }'
    );
    return data['resvUnreadNotifications'] == null ? null : ResvUnreadCount.fromJson((data['resvUnreadNotifications'] as Map).cast<String, dynamic>());
  }

  Future<ResvLoyalty?> resvLoyalty({String? selectionSet}) async {
    final data = await _executor.query('query resvLoyalty { resvLoyalty ${selectionSet ?? '{ points tier tier_name }'} }'
    );
    return data['resvLoyalty'] == null ? null : ResvLoyalty.fromJson((data['resvLoyalty'] as Map).cast<String, dynamic>());
  }

  Future<ResvWallet?> resvWallet({String? selectionSet}) async {
    final data = await _executor.query('query resvWallet { resvWallet ${selectionSet ?? '{ wallet_id balance currency status }'} }'
    );
    return data['resvWallet'] == null ? null : ResvWallet.fromJson((data['resvWallet'] as Map).cast<String, dynamic>());
  }

  Future<List<ResvWalletTxn>?> resvWalletTransactions({int? limit, int? offset, String? selectionSet}) async {
    final data = await _executor.query('query resvWalletTransactions(\$limit: Int, \$offset: Int) { resvWalletTransactions(limit: \$limit, offset: \$offset) ${selectionSet ?? '{ transaction_id type amount currency status reference_id description created_at }'} }', variables: {
      'limit': ?limit,
      'offset': ?offset,
    }
    );
    return (data['resvWalletTransactions'] as List?)?.map((e) => (e == null ? null : ResvWalletTxn.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<Object?> resvRegisterHotel({required ResvHotelInput input}) async {
    final data = await _executor.mutation('mutation resvRegisterHotel(\$input: ResvHotelInput!) { resvRegisterHotel(input: \$input) }', variables: {
      'input': input.toJson(),
    }
    );
    return data['resvRegisterHotel'];
  }

  Future<Object?> resvApproveHotel({required int id}) async {
    final data = await _executor.mutation('mutation resvApproveHotel(\$id: Int!) { resvApproveHotel(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['resvApproveHotel'];
  }

  Future<Object?> resvRejectHotel({required int id, required String reason}) async {
    final data = await _executor.mutation('mutation resvRejectHotel(\$id: Int!, \$reason: String!) { resvRejectHotel(id: \$id, reason: \$reason) }', variables: {
      'id': id,
      'reason': reason,
    }
    );
    return data['resvRejectHotel'];
  }

  Future<Object?> resvUpdateHotel({required int id, required ResvHotelInput input}) async {
    final data = await _executor.mutation('mutation resvUpdateHotel(\$id: Int!, \$input: ResvHotelInput!) { resvUpdateHotel(id: \$id, input: \$input) }', variables: {
      'id': id,
      'input': input.toJson(),
    }
    );
    return data['resvUpdateHotel'];
  }

  Future<bool?> resvDeactivateHotel({required int id}) async {
    final data = await _executor.mutation('mutation resvDeactivateHotel(\$id: Int!) { resvDeactivateHotel(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['resvDeactivateHotel'] as bool?;
  }

  Future<Object?> resvCreateRoomType({required int id, required ResvRoomTypeInput input}) async {
    final data = await _executor.mutation('mutation resvCreateRoomType(\$id: Int!, \$input: ResvRoomTypeInput!) { resvCreateRoomType(id: \$id, input: \$input) }', variables: {
      'id': id,
      'input': input.toJson(),
    }
    );
    return data['resvCreateRoomType'];
  }

  Future<Object?> resvUpdateRoomType({required int id, required String type_, required ResvRoomTypeInput input}) async {
    final data = await _executor.mutation('mutation resvUpdateRoomType(\$id: Int!, \$type: String!, \$input: ResvRoomTypeInput!) { resvUpdateRoomType(id: \$id, type: \$type, input: \$input) }', variables: {
      'id': id,
      'type': type_,
      'input': input.toJson(),
    }
    );
    return data['resvUpdateRoomType'];
  }

  Future<Object?> resvSetInventory({required int id, required String type_, required ResvInventoryInput input}) async {
    final data = await _executor.mutation('mutation resvSetInventory(\$id: Int!, \$type: String!, \$input: ResvInventoryInput!) { resvSetInventory(id: \$id, type: \$type, input: \$input) }', variables: {
      'id': id,
      'type': type_,
      'input': input.toJson(),
    }
    );
    return data['resvSetInventory'];
  }

  Future<Object?> resvCreateRoom({required int id, required ResvRoomInput input}) async {
    final data = await _executor.mutation('mutation resvCreateRoom(\$id: Int!, \$input: ResvRoomInput!) { resvCreateRoom(id: \$id, input: \$input) }', variables: {
      'id': id,
      'input': input.toJson(),
    }
    );
    return data['resvCreateRoom'];
  }

  Future<Object?> resvUpdateRoom({required int id, required String room, required ResvRoomInput input}) async {
    final data = await _executor.mutation('mutation resvUpdateRoom(\$id: Int!, \$room: String!, \$input: ResvRoomInput!) { resvUpdateRoom(id: \$id, room: \$room, input: \$input) }', variables: {
      'id': id,
      'room': room,
      'input': input.toJson(),
    }
    );
    return data['resvUpdateRoom'];
  }

  Future<Object?> resvCreatePromotion({required int id, required Map<String, dynamic> input}) async {
    final data = await _executor.mutation('mutation resvCreatePromotion(\$id: Int!, \$input: Map!) { resvCreatePromotion(id: \$id, input: \$input) }', variables: {
      'id': id,
      'input': input,
    }
    );
    return data['resvCreatePromotion'];
  }

  Future<Object?> resvUpdatePromotion({required int id, required String promo, required Map<String, dynamic> input}) async {
    final data = await _executor.mutation('mutation resvUpdatePromotion(\$id: Int!, \$promo: String!, \$input: Map!) { resvUpdatePromotion(id: \$id, promo: \$promo, input: \$input) }', variables: {
      'id': id,
      'promo': promo,
      'input': input,
    }
    );
    return data['resvUpdatePromotion'];
  }

  Future<Object?> resvCreateService({required int id, required ResvServiceInput input}) async {
    final data = await _executor.mutation('mutation resvCreateService(\$id: Int!, \$input: ResvServiceInput!) { resvCreateService(id: \$id, input: \$input) }', variables: {
      'id': id,
      'input': input.toJson(),
    }
    );
    return data['resvCreateService'];
  }

  Future<Object?> resvUpdateService({required int id, required String service, required ResvServiceInput input}) async {
    final data = await _executor.mutation('mutation resvUpdateService(\$id: Int!, \$service: String!, \$input: ResvServiceInput!) { resvUpdateService(id: \$id, service: \$service, input: \$input) }', variables: {
      'id': id,
      'service': service,
      'input': input.toJson(),
    }
    );
    return data['resvUpdateService'];
  }

  Future<Object?> resvReserve({required ResvReservationInput input}) async {
    final data = await _executor.mutation('mutation resvReserve(\$input: ResvReservationInput!) { resvReserve(input: \$input) }', variables: {
      'input': input.toJson(),
    }
    );
    return data['resvReserve'];
  }

  Future<Object?> resvPayReservation({required int id, required String method, String? provider}) async {
    final data = await _executor.mutation('mutation resvPayReservation(\$id: Int!, \$method: String!, \$provider: String) { resvPayReservation(id: \$id, method: \$method, provider: \$provider) }', variables: {
      'id': id,
      'method': method,
      'provider': ?provider,
    }
    );
    return data['resvPayReservation'];
  }

  Future<Object?> resvCancelReservation({required int id, String? reason}) async {
    final data = await _executor.mutation('mutation resvCancelReservation(\$id: Int!, \$reason: String) { resvCancelReservation(id: \$id, reason: \$reason) }', variables: {
      'id': id,
      'reason': ?reason,
    }
    );
    return data['resvCancelReservation'];
  }

  Future<Object?> resvRejectReservation({required int id, String? reason}) async {
    final data = await _executor.mutation('mutation resvRejectReservation(\$id: Int!, \$reason: String) { resvRejectReservation(id: \$id, reason: \$reason) }', variables: {
      'id': id,
      'reason': ?reason,
    }
    );
    return data['resvRejectReservation'];
  }

  Future<Object?> resvCheckInReservation({required int id, String? reason}) async {
    final data = await _executor.mutation('mutation resvCheckInReservation(\$id: Int!, \$reason: String) { resvCheckInReservation(id: \$id, reason: \$reason) }', variables: {
      'id': id,
      'reason': ?reason,
    }
    );
    return data['resvCheckInReservation'];
  }

  Future<Object?> resvCheckOutReservation({required int id, String? reason}) async {
    final data = await _executor.mutation('mutation resvCheckOutReservation(\$id: Int!, \$reason: String) { resvCheckOutReservation(id: \$id, reason: \$reason) }', variables: {
      'id': id,
      'reason': ?reason,
    }
    );
    return data['resvCheckOutReservation'];
  }

  Future<Object?> resvNoShowReservation({required int id, String? reason}) async {
    final data = await _executor.mutation('mutation resvNoShowReservation(\$id: Int!, \$reason: String) { resvNoShowReservation(id: \$id, reason: \$reason) }', variables: {
      'id': id,
      'reason': ?reason,
    }
    );
    return data['resvNoShowReservation'];
  }

  Future<Object?> resvCreateReview({required int id, required Map<String, dynamic> input}) async {
    final data = await _executor.mutation('mutation resvCreateReview(\$id: Int!, \$input: Map!) { resvCreateReview(id: \$id, input: \$input) }', variables: {
      'id': id,
      'input': input,
    }
    );
    return data['resvCreateReview'];
  }

  Future<Object?> resvJoinWaitlist({required Map<String, dynamic> input}) async {
    final data = await _executor.mutation('mutation resvJoinWaitlist(\$input: Map!) { resvJoinWaitlist(input: \$input) }', variables: {
      'input': input,
    }
    );
    return data['resvJoinWaitlist'];
  }

  Future<bool?> resvLeaveWaitlist({required int id}) async {
    final data = await _executor.mutation('mutation resvLeaveWaitlist(\$id: Int!) { resvLeaveWaitlist(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['resvLeaveWaitlist'] as bool?;
  }

  Future<bool?> resvAddFavourite({required int id}) async {
    final data = await _executor.mutation('mutation resvAddFavourite(\$id: Int!) { resvAddFavourite(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['resvAddFavourite'] as bool?;
  }

  Future<bool?> resvRemoveFavourite({required int id}) async {
    final data = await _executor.mutation('mutation resvRemoveFavourite(\$id: Int!) { resvRemoveFavourite(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['resvRemoveFavourite'] as bool?;
  }

  Future<Object?> resvMarkAllNotificationsRead() async {
    final data = await _executor.mutation('mutation resvMarkAllNotificationsRead { resvMarkAllNotificationsRead }'
    );
    return data['resvMarkAllNotificationsRead'];
  }

  Future<Object?> resvMarkNotificationRead({required int id}) async {
    final data = await _executor.mutation('mutation resvMarkNotificationRead(\$id: Int!) { resvMarkNotificationRead(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['resvMarkNotificationRead'];
  }

  Future<Object?> resvCreateWallet({required String currency}) async {
    final data = await _executor.mutation('mutation resvCreateWallet(\$currency: String!) { resvCreateWallet(currency: \$currency) }', variables: {
      'currency': currency,
    }
    );
    return data['resvCreateWallet'];
  }

}
