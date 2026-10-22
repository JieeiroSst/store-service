// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

import '../gateway_executor.dart';
import '../models.g.dart';

class RentHouseApi {
  const RentHouseApi(this._executor);

  final GatewayExecutor _executor;

  Future<RentHomestay?> rentHomestay({required int id, String? selectionSet}) async {
    final data = await _executor.public('query rentHomestay(\$id: Int!) { rentHomestay(id: \$id) ${selectionSet ?? '{ id name description type type_name status status_name review_note phone_number address ward_id district_id province_id images guests bedrooms bathrooms amenity_ids version created_at updated_at host_id accepts_wallet_payment rating review_count rates { model price currency min_periods active } }'} }', variables: {
      'id': id,
    }
    );
    return data['rentHomestay'] == null ? null : RentHomestay.fromJson((data['rentHomestay'] as Map).cast<String, dynamic>());
  }

  Future<RentHomestayPage?> rentHomestays({int? provinceId, int? districtId, int? wardId, int? type_, int? guests, String? q, String? modelArg, String? minPrice, String? maxPrice, String? sort, int? limit, String? cursor, String? amenityIds, String? checkin, String? checkout, String? selectionSet}) async {
    final data = await _executor.public('query rentHomestays(\$province_id: Int, \$district_id: Int, \$ward_id: Int, \$type: Int, \$guests: Int, \$q: String, \$modelArg: String, \$min_price: String, \$max_price: String, \$sort: String, \$limit: Int, \$cursor: String, \$amenity_ids: String, \$checkin: String, \$checkout: String) { rentHomestays(province_id: \$province_id, district_id: \$district_id, ward_id: \$ward_id, type: \$type, guests: \$guests, q: \$q, modelArg: \$modelArg, min_price: \$min_price, max_price: \$max_price, sort: \$sort, limit: \$limit, cursor: \$cursor, amenity_ids: \$amenity_ids, checkin: \$checkin, checkout: \$checkout) ${selectionSet ?? '{ items { id name description type type_name status status_name review_note phone_number address ward_id district_id province_id images guests bedrooms bathrooms amenity_ids version created_at updated_at host_id accepts_wallet_payment rating review_count rates { model price currency min_periods active } } has_more next_cursor }'} }', variables: {
      'province_id': ?provinceId,
      'district_id': ?districtId,
      'ward_id': ?wardId,
      'type': ?type_,
      'guests': ?guests,
      'q': ?q,
      'modelArg': ?modelArg,
      'min_price': ?minPrice,
      'max_price': ?maxPrice,
      'sort': ?sort,
      'limit': ?limit,
      'cursor': ?cursor,
      'amenity_ids': ?amenityIds,
      'checkin': ?checkin,
      'checkout': ?checkout,
    }
    );
    return data['rentHomestays'] == null ? null : RentHomestayPage.fromJson((data['rentHomestays'] as Map).cast<String, dynamic>());
  }

  Future<List<RentSlot>?> rentAvailability({required int id, String? from, String? to, String? selectionSet}) async {
    final data = await _executor.public('query rentAvailability(\$id: Int!, \$from: String, \$to: String) { rentAvailability(id: \$id, from: \$from, to: \$to) ${selectionSet ?? '{ date price status }'} }', variables: {
      'id': id,
      'from': ?from,
      'to': ?to,
    }
    );
    return (data['rentAvailability'] as List?)?.map((e) => (e == null ? null : RentSlot.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<List<RentAmenity>?> rentAmenities({String? selectionSet}) async {
    final data = await _executor.public('query rentAmenities { rentAmenities ${selectionSet ?? '{ id name icon }'} }'
    );
    return (data['rentAmenities'] as List?)?.map((e) => (e == null ? null : RentAmenity.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<List<RentHomestay>?> rentMyHomestays({int? limit, int? offset, String? selectionSet}) async {
    final data = await _executor.query('query rentMyHomestays(\$limit: Int, \$offset: Int) { rentMyHomestays(limit: \$limit, offset: \$offset) ${selectionSet ?? '{ id name description type type_name status status_name review_note phone_number address ward_id district_id province_id images guests bedrooms bathrooms amenity_ids version created_at updated_at host_id accepts_wallet_payment rating review_count rates { model price currency min_periods active } }'} }', variables: {
      'limit': ?limit,
      'offset': ?offset,
    }
    );
    return (data['rentMyHomestays'] as List?)?.map((e) => (e == null ? null : RentHomestay.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<List<RentHomestay>?> rentHomestayReviewQueue({String? status, int? limit, int? offset, String? selectionSet}) async {
    final data = await _executor.query('query rentHomestayReviewQueue(\$status: String, \$limit: Int, \$offset: Int) { rentHomestayReviewQueue(status: \$status, limit: \$limit, offset: \$offset) ${selectionSet ?? '{ id name description type type_name status status_name review_note phone_number address ward_id district_id province_id images guests bedrooms bathrooms amenity_ids version created_at updated_at host_id accepts_wallet_payment rating review_count rates { model price currency min_periods active } }'} }', variables: {
      'status': ?status,
      'limit': ?limit,
      'offset': ?offset,
    }
    );
    return (data['rentHomestayReviewQueue'] as List?)?.map((e) => (e == null ? null : RentHomestay.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<List<RentBooking>?> rentBookings({int? userId, String? as_, int? limit, int? offset, String? selectionSet}) async {
    final data = await _executor.query('query rentBookings(\$user_id: Int, \$as: String, \$limit: Int, \$offset: Int) { rentBookings(user_id: \$user_id, as: \$as, limit: \$limit, offset: \$offset) ${selectionSet ?? '{ id user_id homestay_id checkin_date checkout_date nights guests status currency subtotal discount total_amount note request_id created_at payment_method payment_ref paid_at expires_at }'} }', variables: {
      'user_id': ?userId,
      'as': ?as_,
      'limit': ?limit,
      'offset': ?offset,
    }
    );
    return (data['rentBookings'] as List?)?.map((e) => (e == null ? null : RentBooking.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<RentBooking?> rentBooking({required int id, String? selectionSet}) async {
    final data = await _executor.query('query rentBooking(\$id: Int!) { rentBooking(id: \$id) ${selectionSet ?? '{ id user_id homestay_id checkin_date checkout_date nights guests status currency subtotal discount total_amount note request_id created_at payment_method payment_ref paid_at expires_at }'} }', variables: {
      'id': id,
    }
    );
    return data['rentBooking'] == null ? null : RentBooking.fromJson((data['rentBooking'] as Map).cast<String, dynamic>());
  }

  Future<List<RentRate>?> rentRates({required int id, String? selectionSet}) async {
    final data = await _executor.public('query rentRates(\$id: Int!) { rentRates(id: \$id) ${selectionSet ?? '{ model price currency min_periods active }'} }', variables: {
      'id': id,
    }
    );
    return (data['rentRates'] as List?)?.map((e) => (e == null ? null : RentRate.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<List<RentLease>?> rentLeases({int? userId, String? as_, int? limit, int? offset, String? selectionSet}) async {
    final data = await _executor.query('query rentLeases(\$user_id: Int, \$as: String, \$limit: Int, \$offset: Int) { rentLeases(user_id: \$user_id, as: \$as, limit: \$limit, offset: \$offset) ${selectionSet ?? '{ id user_id homestay_id model start_date end_date periods billing_day currency rent status note expires_at created_at invoices { id lease_id user_id homestay_id period period_start period_end due_date amount status overdue payment_method payment_ref paid_at } }'} }', variables: {
      'user_id': ?userId,
      'as': ?as_,
      'limit': ?limit,
      'offset': ?offset,
    }
    );
    return (data['rentLeases'] as List?)?.map((e) => (e == null ? null : RentLease.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<RentLease?> rentLease({required int id, String? selectionSet}) async {
    final data = await _executor.query('query rentLease(\$id: Int!) { rentLease(id: \$id) ${selectionSet ?? '{ id user_id homestay_id model start_date end_date periods billing_day currency rent status note expires_at created_at invoices { id lease_id user_id homestay_id period period_start period_end due_date amount status overdue payment_method payment_ref paid_at } }'} }', variables: {
      'id': id,
    }
    );
    return data['rentLease'] == null ? null : RentLease.fromJson((data['rentLease'] as Map).cast<String, dynamic>());
  }

  Future<List<RentInvoice>?> rentInvoices({int? limit, int? offset, String? status, String? dueBefore, String? selectionSet}) async {
    final data = await _executor.query('query rentInvoices(\$limit: Int, \$offset: Int, \$status: String, \$due_before: String) { rentInvoices(limit: \$limit, offset: \$offset, status: \$status, due_before: \$due_before) ${selectionSet ?? '{ id lease_id user_id homestay_id period period_start period_end due_date amount status overdue payment_method payment_ref paid_at }'} }', variables: {
      'limit': ?limit,
      'offset': ?offset,
      'status': ?status,
      'due_before': ?dueBefore,
    }
    );
    return (data['rentInvoices'] as List?)?.map((e) => (e == null ? null : RentInvoice.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<List<RentReview>?> rentReviews({required int id, int? limit, int? offset, String? selectionSet}) async {
    final data = await _executor.public('query rentReviews(\$id: Int!, \$limit: Int, \$offset: Int) { rentReviews(id: \$id, limit: \$limit, offset: \$offset) ${selectionSet ?? '{ id homestay_id user_id booking_id lease_id rating comment created_at }'} }', variables: {
      'id': id,
      'limit': ?limit,
      'offset': ?offset,
    }
    );
    return (data['rentReviews'] as List?)?.map((e) => (e == null ? null : RentReview.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<List<RentHomestay>?> rentWishlist({int? limit, int? offset, String? selectionSet}) async {
    final data = await _executor.query('query rentWishlist(\$limit: Int, \$offset: Int) { rentWishlist(limit: \$limit, offset: \$offset) ${selectionSet ?? '{ id name description type type_name status status_name review_note phone_number address ward_id district_id province_id images guests bedrooms bathrooms amenity_ids version created_at updated_at host_id accepts_wallet_payment rating review_count rates { model price currency min_periods active } }'} }', variables: {
      'limit': ?limit,
      'offset': ?offset,
    }
    );
    return (data['rentWishlist'] as List?)?.map((e) => (e == null ? null : RentHomestay.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<RentLoyalty?> rentLoyalty({String? selectionSet}) async {
    final data = await _executor.query('query rentLoyalty { rentLoyalty ${selectionSet ?? '{ points tier tier_name }'} }'
    );
    return data['rentLoyalty'] == null ? null : RentLoyalty.fromJson((data['rentLoyalty'] as Map).cast<String, dynamic>());
  }

  Future<RentWallet?> rentWallet({String? selectionSet}) async {
    final data = await _executor.query('query rentWallet { rentWallet ${selectionSet ?? '{ wallet_id balance currency status }'} }'
    );
    return data['rentWallet'] == null ? null : RentWallet.fromJson((data['rentWallet'] as Map).cast<String, dynamic>());
  }

  Future<List<RentWalletTxn>?> rentWalletTransactions({int? limit, int? offset, String? selectionSet}) async {
    final data = await _executor.query('query rentWalletTransactions(\$limit: Int, \$offset: Int) { rentWalletTransactions(limit: \$limit, offset: \$offset) ${selectionSet ?? '{ transaction_id type amount currency status reference_id description created_at }'} }', variables: {
      'limit': ?limit,
      'offset': ?offset,
    }
    );
    return (data['rentWalletTransactions'] as List?)?.map((e) => (e == null ? null : RentWalletTxn.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<RentHomestay?> rentApproveHomestay({required int id, String? selectionSet}) async {
    final data = await _executor.mutation('mutation rentApproveHomestay(\$id: Int!) { rentApproveHomestay(id: \$id) ${selectionSet ?? '{ id name description type type_name status status_name review_note phone_number address ward_id district_id province_id images guests bedrooms bathrooms amenity_ids version created_at updated_at host_id accepts_wallet_payment rating review_count rates { model price currency min_periods active } }'} }', variables: {
      'id': id,
    }
    );
    return data['rentApproveHomestay'] == null ? null : RentHomestay.fromJson((data['rentApproveHomestay'] as Map).cast<String, dynamic>());
  }

  Future<RentHomestay?> rentRejectHomestay({required int id, required String reason, String? selectionSet}) async {
    final data = await _executor.mutation('mutation rentRejectHomestay(\$id: Int!, \$reason: String!) { rentRejectHomestay(id: \$id, reason: \$reason) ${selectionSet ?? '{ id name description type type_name status status_name review_note phone_number address ward_id district_id province_id images guests bedrooms bathrooms amenity_ids version created_at updated_at host_id accepts_wallet_payment rating review_count rates { model price currency min_periods active } }'} }', variables: {
      'id': id,
      'reason': reason,
    }
    );
    return data['rentRejectHomestay'] == null ? null : RentHomestay.fromJson((data['rentRejectHomestay'] as Map).cast<String, dynamic>());
  }

  Future<RentHomestay?> rentCreateHomestay({required RentHomestayInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation rentCreateHomestay(\$input: RentHomestayInput!) { rentCreateHomestay(input: \$input) ${selectionSet ?? '{ id name description type type_name status status_name review_note phone_number address ward_id district_id province_id images guests bedrooms bathrooms amenity_ids version created_at updated_at host_id accepts_wallet_payment rating review_count rates { model price currency min_periods active } }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['rentCreateHomestay'] == null ? null : RentHomestay.fromJson((data['rentCreateHomestay'] as Map).cast<String, dynamic>());
  }

  Future<RentHomestay?> rentUpdateHomestay({required int id, required RentHomestayInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation rentUpdateHomestay(\$id: Int!, \$input: RentHomestayInput!) { rentUpdateHomestay(id: \$id, input: \$input) ${selectionSet ?? '{ id name description type type_name status status_name review_note phone_number address ward_id district_id province_id images guests bedrooms bathrooms amenity_ids version created_at updated_at host_id accepts_wallet_payment rating review_count rates { model price currency min_periods active } }'} }', variables: {
      'id': id,
      'input': input.toJson(),
    }
    );
    return data['rentUpdateHomestay'] == null ? null : RentHomestay.fromJson((data['rentUpdateHomestay'] as Map).cast<String, dynamic>());
  }

  Future<bool?> rentDeactivateHomestay({required int id}) async {
    final data = await _executor.mutation('mutation rentDeactivateHomestay(\$id: Int!) { rentDeactivateHomestay(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['rentDeactivateHomestay'] as bool?;
  }

  Future<Object?> rentSetAvailability({required int id, required RentAvailabilityInput input}) async {
    final data = await _executor.mutation('mutation rentSetAvailability(\$id: Int!, \$input: RentAvailabilityInput!) { rentSetAvailability(id: \$id, input: \$input) }', variables: {
      'id': id,
      'input': input.toJson(),
    }
    );
    return data['rentSetAvailability'];
  }

  Future<RentAmenity?> rentCreateAmenity({required String name_, String? icon, String? selectionSet}) async {
    final data = await _executor.mutation('mutation rentCreateAmenity(\$name: String!, \$icon: String) { rentCreateAmenity(name: \$name, icon: \$icon) ${selectionSet ?? '{ id name icon }'} }', variables: {
      'name': name_,
      'icon': ?icon,
    }
    );
    return data['rentCreateAmenity'] == null ? null : RentAmenity.fromJson((data['rentCreateAmenity'] as Map).cast<String, dynamic>());
  }

  Future<RentBooking?> rentBook({required RentBookingInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation rentBook(\$input: RentBookingInput!) { rentBook(input: \$input) ${selectionSet ?? '{ id user_id homestay_id checkin_date checkout_date nights guests status currency subtotal discount total_amount note request_id created_at payment_method payment_ref paid_at expires_at }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['rentBook'] == null ? null : RentBooking.fromJson((data['rentBook'] as Map).cast<String, dynamic>());
  }

  Future<RentBooking?> rentCancelBooking({required int id, String? selectionSet}) async {
    final data = await _executor.mutation('mutation rentCancelBooking(\$id: Int!) { rentCancelBooking(id: \$id) ${selectionSet ?? '{ id user_id homestay_id checkin_date checkout_date nights guests status currency subtotal discount total_amount note request_id created_at payment_method payment_ref paid_at expires_at }'} }', variables: {
      'id': id,
    }
    );
    return data['rentCancelBooking'] == null ? null : RentBooking.fromJson((data['rentCancelBooking'] as Map).cast<String, dynamic>());
  }

  Future<RentBooking?> rentPayBooking({required int id, required String method, String? provider, String? selectionSet}) async {
    final data = await _executor.mutation('mutation rentPayBooking(\$id: Int!, \$method: String!, \$provider: String) { rentPayBooking(id: \$id, method: \$method, provider: \$provider) ${selectionSet ?? '{ id user_id homestay_id checkin_date checkout_date nights guests status currency subtotal discount total_amount note request_id created_at payment_method payment_ref paid_at expires_at }'} }', variables: {
      'id': id,
      'method': method,
      'provider': ?provider,
    }
    );
    return data['rentPayBooking'] == null ? null : RentBooking.fromJson((data['rentPayBooking'] as Map).cast<String, dynamic>());
  }

  Future<RentRate?> rentSetRate({required int id, required String rateModel, required RentRateInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation rentSetRate(\$id: Int!, \$rate_model: String!, \$input: RentRateInput!) { rentSetRate(id: \$id, rate_model: \$rate_model, input: \$input) ${selectionSet ?? '{ model price currency min_periods active }'} }', variables: {
      'id': id,
      'rate_model': rateModel,
      'input': input.toJson(),
    }
    );
    return data['rentSetRate'] == null ? null : RentRate.fromJson((data['rentSetRate'] as Map).cast<String, dynamic>());
  }

  Future<bool?> rentDeleteRate({required int id, required String rateModel}) async {
    final data = await _executor.mutation('mutation rentDeleteRate(\$id: Int!, \$rate_model: String!) { rentDeleteRate(id: \$id, rate_model: \$rate_model) }', variables: {
      'id': id,
      'rate_model': rateModel,
    }
    );
    return data['rentDeleteRate'] as bool?;
  }

  Future<RentLease?> rentStartLease({required RentLeaseInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation rentStartLease(\$input: RentLeaseInput!) { rentStartLease(input: \$input) ${selectionSet ?? '{ id user_id homestay_id model start_date end_date periods billing_day currency rent status note expires_at created_at invoices { id lease_id user_id homestay_id period period_start period_end due_date amount status overdue payment_method payment_ref paid_at } }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['rentStartLease'] == null ? null : RentLease.fromJson((data['rentStartLease'] as Map).cast<String, dynamic>());
  }

  Future<RentLease?> rentCancelLease({required int id, String? selectionSet}) async {
    final data = await _executor.mutation('mutation rentCancelLease(\$id: Int!) { rentCancelLease(id: \$id) ${selectionSet ?? '{ id user_id homestay_id model start_date end_date periods billing_day currency rent status note expires_at created_at invoices { id lease_id user_id homestay_id period period_start period_end due_date amount status overdue payment_method payment_ref paid_at } }'} }', variables: {
      'id': id,
    }
    );
    return data['rentCancelLease'] == null ? null : RentLease.fromJson((data['rentCancelLease'] as Map).cast<String, dynamic>());
  }

  Future<RentLease?> rentTerminateLease({required int id, String? selectionSet}) async {
    final data = await _executor.mutation('mutation rentTerminateLease(\$id: Int!) { rentTerminateLease(id: \$id) ${selectionSet ?? '{ id user_id homestay_id model start_date end_date periods billing_day currency rent status note expires_at created_at invoices { id lease_id user_id homestay_id period period_start period_end due_date amount status overdue payment_method payment_ref paid_at } }'} }', variables: {
      'id': id,
    }
    );
    return data['rentTerminateLease'] == null ? null : RentLease.fromJson((data['rentTerminateLease'] as Map).cast<String, dynamic>());
  }

  Future<Object?> rentPayLeaseInvoice({required int id, required String period, required String method, String? provider}) async {
    final data = await _executor.mutation('mutation rentPayLeaseInvoice(\$id: Int!, \$period: String!, \$method: String!, \$provider: String) { rentPayLeaseInvoice(id: \$id, period: \$period, method: \$method, provider: \$provider) }', variables: {
      'id': id,
      'period': period,
      'method': method,
      'provider': ?provider,
    }
    );
    return data['rentPayLeaseInvoice'];
  }

  Future<RentReview?> rentCreateReview({required int id, required RentReviewInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation rentCreateReview(\$id: Int!, \$input: RentReviewInput!) { rentCreateReview(id: \$id, input: \$input) ${selectionSet ?? '{ id homestay_id user_id booking_id lease_id rating comment created_at }'} }', variables: {
      'id': id,
      'input': input.toJson(),
    }
    );
    return data['rentCreateReview'] == null ? null : RentReview.fromJson((data['rentCreateReview'] as Map).cast<String, dynamic>());
  }

  Future<bool?> rentAddWishlist({required int id}) async {
    final data = await _executor.mutation('mutation rentAddWishlist(\$id: Int!) { rentAddWishlist(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['rentAddWishlist'] as bool?;
  }

  Future<bool?> rentRemoveWishlist({required int id}) async {
    final data = await _executor.mutation('mutation rentRemoveWishlist(\$id: Int!) { rentRemoveWishlist(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['rentRemoveWishlist'] as bool?;
  }

  Future<RentWallet?> rentCreateWallet({required String currency, String? selectionSet}) async {
    final data = await _executor.mutation('mutation rentCreateWallet(\$currency: String!) { rentCreateWallet(currency: \$currency) ${selectionSet ?? '{ wallet_id balance currency status }'} }', variables: {
      'currency': currency,
    }
    );
    return data['rentCreateWallet'] == null ? null : RentWallet.fromJson((data['rentCreateWallet'] as Map).cast<String, dynamic>());
  }

}
