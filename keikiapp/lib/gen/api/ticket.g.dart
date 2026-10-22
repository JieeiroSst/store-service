// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

import '../gateway_executor.dart';
import '../models.g.dart';

class TicketApi {
  const TicketApi(this._executor);

  final GatewayExecutor _executor;

  Future<List<TicketEvent>?> ticketEvents({String? q, String? city, String? category, String? sort, bool? featured, int? limit, int? offset, String? from, String? to, String? minPrice, String? maxPrice, String? selectionSet}) async {
    final data = await _executor.public('query ticketEvents(\$q: String, \$city: String, \$category: String, \$sort: String, \$featured: Boolean, \$limit: Int, \$offset: Int, \$from: String, \$to: String, \$min_price: String, \$max_price: String) { ticketEvents(q: \$q, city: \$city, category: \$category, sort: \$sort, featured: \$featured, limit: \$limit, offset: \$offset, from: \$from, to: \$to, min_price: \$min_price, max_price: \$max_price) ${selectionSet ?? '{ id organizer_id title description category city venue address banner_url starts_at ends_at currency status review_note featured wallet_id refund_cutoff_hours transferable resale_cap_percent series_id venue_id min_price sold_out next_session_at sessions { id event_id starts_at ends_at label status cancel_reason } ticket_types { id event_id name description price total available sold held min_per_order max_per_order max_per_user sale_starts_at sale_ends_at seated active sort_order session_id on_sale sold_out } }'} }', variables: {
      'q': ?q,
      'city': ?city,
      'category': ?category,
      'sort': ?sort,
      'featured': ?featured,
      'limit': ?limit,
      'offset': ?offset,
      'from': ?from,
      'to': ?to,
      'min_price': ?minPrice,
      'max_price': ?maxPrice,
    }
    );
    return (data['ticketEvents'] as List?)?.map((e) => (e == null ? null : TicketEvent.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<List<String>?> ticketCategories() async {
    final data = await _executor.public('query ticketCategories { ticketCategories }'
    );
    return (data['ticketCategories'] as List?)?.map((e) => (e?.toString())!).toList();
  }

  Future<TicketEvent?> ticketEvent({required int id, String? selectionSet}) async {
    final data = await _executor.public('query ticketEvent(\$id: Int!) { ticketEvent(id: \$id) ${selectionSet ?? '{ id organizer_id title description category city venue address banner_url starts_at ends_at currency status review_note featured wallet_id refund_cutoff_hours transferable resale_cap_percent series_id venue_id min_price sold_out next_session_at sessions { id event_id starts_at ends_at label status cancel_reason } ticket_types { id event_id name description price total available sold held min_per_order max_per_order max_per_user sale_starts_at sale_ends_at seated active sort_order session_id on_sale sold_out } }'} }', variables: {
      'id': id,
    }
    );
    return data['ticketEvent'] == null ? null : TicketEvent.fromJson((data['ticketEvent'] as Map).cast<String, dynamic>());
  }

  Future<TicketSeatList?> ticketSeats({required int id, required int type_, String? selectionSet}) async {
    final data = await _executor.public('query ticketSeats(\$id: Int!, \$type: Int!) { ticketSeats(id: \$id, type: \$type) ${selectionSet ?? '{ layout { width height stage { label x y w h } stages { label shape x y w h points } sections { key name label color points ticket_type_id } } items { id ticket_type_id accessible section row number label status x y } }'} }', variables: {
      'id': id,
      'type': type_,
    }
    );
    return data['ticketSeats'] == null ? null : TicketSeatList.fromJson((data['ticketSeats'] as Map).cast<String, dynamic>());
  }

  Future<List<TicketEvent>?> ticketEventSeries({required int id, String? selectionSet}) async {
    final data = await _executor.public('query ticketEventSeries(\$id: Int!) { ticketEventSeries(id: \$id) ${selectionSet ?? '{ id organizer_id title description category city venue address banner_url starts_at ends_at currency status review_note featured wallet_id refund_cutoff_hours transferable resale_cap_percent series_id venue_id min_price sold_out next_session_at sessions { id event_id starts_at ends_at label status cancel_reason } ticket_types { id event_id name description price total available sold held min_per_order max_per_order max_per_user sale_starts_at sale_ends_at seated active sort_order session_id on_sale sold_out } }'} }', variables: {
      'id': id,
    }
    );
    return (data['ticketEventSeries'] as List?)?.map((e) => (e == null ? null : TicketEvent.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<List<TicketSession>?> ticketSessions({required int id, String? selectionSet}) async {
    final data = await _executor.public('query ticketSessions(\$id: Int!) { ticketSessions(id: \$id) ${selectionSet ?? '{ id event_id starts_at ends_at label status cancel_reason }'} }', variables: {
      'id': id,
    }
    );
    return (data['ticketSessions'] as List?)?.map((e) => (e == null ? null : TicketSession.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<TicketSeatMap?> ticketSeatMap({required int id, String? selectionSet}) async {
    final data = await _executor.public('query ticketSeatMap(\$id: Int!) { ticketSeatMap(id: \$id) ${selectionSet ?? '{ layout { width height stage { label x y w h } stages { label shape x y w h points } sections { key name label color points ticket_type_id } } ticket_types { id event_id name description price total available sold held min_per_order max_per_order max_per_user sale_starts_at sale_ends_at seated active sort_order session_id on_sale sold_out } seats { id ticket_type_id accessible section row number label status x y } }'} }', variables: {
      'id': id,
    }
    );
    return data['ticketSeatMap'] == null ? null : TicketSeatMap.fromJson((data['ticketSeatMap'] as Map).cast<String, dynamic>());
  }

  Future<List<TicketTemplate>?> ticketVenueTemplates({String? selectionSet}) async {
    final data = await _executor.public('query ticketVenueTemplates { ticketVenueTemplates ${selectionSet ?? '{ name description params { name description default min max } }'} }'
    );
    return (data['ticketVenueTemplates'] as List?)?.map((e) => (e == null ? null : TicketTemplate.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<TicketResalePage?> ticketResale({required int id, int? type_, String? sort, int? limit, String? cursor, String? selectionSet}) async {
    final data = await _executor.public('query ticketResale(\$id: Int!, \$type: Int, \$sort: String, \$limit: Int, \$cursor: String) { ticketResale(id: \$id, type: \$type, sort: \$sort, limit: \$limit, cursor: \$cursor) ${selectionSet ?? '{ items { id ticket_id event_id type_name seat price face_price currency status created_at seller_id fee } next_cursor }'} }', variables: {
      'id': id,
      'type': ?type_,
      'sort': ?sort,
      'limit': ?limit,
      'cursor': ?cursor,
    }
    );
    return data['ticketResale'] == null ? null : TicketResalePage.fromJson((data['ticketResale'] as Map).cast<String, dynamic>());
  }

  Future<TicketEventPage?> ticketMyEvents({String? cursor, int? limit, String? selectionSet}) async {
    final data = await _executor.query('query ticketMyEvents(\$cursor: String, \$limit: Int) { ticketMyEvents(cursor: \$cursor, limit: \$limit) ${selectionSet ?? '{ items { id organizer_id title description category city venue address banner_url starts_at ends_at currency status review_note featured wallet_id refund_cutoff_hours transferable resale_cap_percent series_id venue_id min_price sold_out next_session_at sessions { id event_id starts_at ends_at label status cancel_reason } ticket_types { id event_id name description price total available sold held min_per_order max_per_order max_per_user sale_starts_at sale_ends_at seated active sort_order session_id on_sale sold_out } } next_cursor }'} }', variables: {
      'cursor': ?cursor,
      'limit': ?limit,
    }
    );
    return data['ticketMyEvents'] == null ? null : TicketEventPage.fromJson((data['ticketMyEvents'] as Map).cast<String, dynamic>());
  }

  Future<TicketVenuePage?> ticketVenues({String? scope, String? cursor, int? limit, String? selectionSet}) async {
    final data = await _executor.query('query ticketVenues(\$scope: String, \$cursor: String, \$limit: Int) { ticketVenues(scope: \$scope, cursor: \$cursor, limit: \$limit) ${selectionSet ?? '{ items { id owner_id name city address description shared seats sections created_at map { width height } } next_cursor }'} }', variables: {
      'scope': ?scope,
      'cursor': ?cursor,
      'limit': ?limit,
    }
    );
    return data['ticketVenues'] == null ? null : TicketVenuePage.fromJson((data['ticketVenues'] as Map).cast<String, dynamic>());
  }

  Future<TicketVenueDetail?> ticketVenue({required int id, bool? seats, String? selectionSet}) async {
    final data = await _executor.query('query ticketVenue(\$id: Int!, \$seats: Boolean) { ticketVenue(id: \$id, seats: \$seats) ${selectionSet ?? '{ venue { id owner_id name city address description shared seats sections created_at map { width height } } preview { total sections { key name seats accessible } layout { width height } seats { section row number x y accessible } } }'} }', variables: {
      'id': id,
      'seats': ?seats,
    }
    );
    return data['ticketVenue'] == null ? null : TicketVenueDetail.fromJson((data['ticketVenue'] as Map).cast<String, dynamic>());
  }

  Future<List<TicketPromotion>?> ticketPromotions({required int id, String? selectionSet}) async {
    final data = await _executor.public('query ticketPromotions(\$id: Int!) { ticketPromotions(id: \$id) ${selectionSet ?? '{ id code kind value max_uses used min_tickets valid_from valid_to active }'} }', variables: {
      'id': id,
    }
    );
    return (data['ticketPromotions'] as List?)?.map((e) => (e == null ? null : TicketPromotion.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<TicketReport?> ticketEventReport({required int id, String? selectionSet}) async {
    final data = await _executor.query('query ticketEventReport(\$id: Int!) { ticketEventReport(id: \$id) ${selectionSet ?? '{ event_id orders tickets_sold revenue refunded checked_in invited by_type { ticket_type_id name total sold held revenue } daily { day orders tickets revenue } }'} }', variables: {
      'id': id,
    }
    );
    return data['ticketEventReport'] == null ? null : TicketReport.fromJson((data['ticketEventReport'] as Map).cast<String, dynamic>());
  }

  Future<TicketOrderPage?> ticketEventOrders({required int id, String? status, String? cursor, int? limit, String? selectionSet}) async {
    final data = await _executor.query('query ticketEventOrders(\$id: Int!, \$status: String, \$cursor: String, \$limit: Int) { ticketEventOrders(id: \$id, status: \$status, cursor: \$cursor, limit: \$limit) ${selectionSet ?? '{ items { id event_id user_id status currency subtotal discount total promo_code buyer_name buyer_email buyer_phone session_id expires_at payment_method paid_at refund_amount status_note created_at items { ticket_type_id name quantity unit_price seat_ids } tickets { id order_id event_id event_title event_starts_at event_ends_at session_id session_label venue ticket_type_id type_name seat_id seat code status holder_id holder_name holder_email transfer_count checked_in_at } } next_cursor }'} }', variables: {
      'id': id,
      'status': ?status,
      'cursor': ?cursor,
      'limit': ?limit,
    }
    );
    return data['ticketEventOrders'] == null ? null : TicketOrderPage.fromJson((data['ticketEventOrders'] as Map).cast<String, dynamic>());
  }

  Future<TicketAttendeePage?> ticketAttendees({required int id, String? cursor, int? limit, String? selectionSet}) async {
    final data = await _executor.query('query ticketAttendees(\$id: Int!, \$cursor: String, \$limit: Int) { ticketAttendees(id: \$id, cursor: \$cursor, limit: \$limit) ${selectionSet ?? '{ items { buyer_name buyer_email id order_id event_id event_title event_starts_at event_ends_at session_id session_label venue ticket_type_id type_name seat_id seat code status holder_id holder_name holder_email transfer_count checked_in_at } next_cursor }'} }', variables: {
      'id': id,
      'cursor': ?cursor,
      'limit': ?limit,
    }
    );
    return data['ticketAttendees'] == null ? null : TicketAttendeePage.fromJson((data['ticketAttendees'] as Map).cast<String, dynamic>());
  }

  Future<TicketTicket?> ticketLookupTicket({required int id, required String code, String? selectionSet}) async {
    final data = await _executor.query('query ticketLookupTicket(\$id: Int!, \$code: String!) { ticketLookupTicket(id: \$id, code: \$code) ${selectionSet ?? '{ id order_id event_id event_title event_starts_at event_ends_at session_id session_label venue ticket_type_id type_name seat_id seat code status holder_id holder_name holder_email transfer_count checked_in_at }'} }', variables: {
      'id': id,
      'code': code,
    }
    );
    return data['ticketLookupTicket'] == null ? null : TicketTicket.fromJson((data['ticketLookupTicket'] as Map).cast<String, dynamic>());
  }

  Future<List<TicketStaff>?> ticketStaff({required int id, String? selectionSet}) async {
    final data = await _executor.query('query ticketStaff(\$id: Int!) { ticketStaff(id: \$id) ${selectionSet ?? '{ user_id added_by created_at }'} }', variables: {
      'id': id,
    }
    );
    return (data['ticketStaff'] as List?)?.map((e) => (e == null ? null : TicketStaff.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<List<TicketEvent>?> ticketStaffEvents({String? selectionSet}) async {
    final data = await _executor.query('query ticketStaffEvents { ticketStaffEvents ${selectionSet ?? '{ id organizer_id title description category city venue address banner_url starts_at ends_at currency status review_note featured wallet_id refund_cutoff_hours transferable resale_cap_percent series_id venue_id min_price sold_out next_session_at sessions { id event_id starts_at ends_at label status cancel_reason } ticket_types { id event_id name description price total available sold held min_per_order max_per_order max_per_user sale_starts_at sale_ends_at seated active sort_order session_id on_sale sold_out } }'} }'
    );
    return (data['ticketStaffEvents'] as List?)?.map((e) => (e == null ? null : TicketEvent.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<TicketEventPage?> ticketEventReviewQueue({String? cursor, int? limit, String? selectionSet}) async {
    final data = await _executor.query('query ticketEventReviewQueue(\$cursor: String, \$limit: Int) { ticketEventReviewQueue(cursor: \$cursor, limit: \$limit) ${selectionSet ?? '{ items { id organizer_id title description category city venue address banner_url starts_at ends_at currency status review_note featured wallet_id refund_cutoff_hours transferable resale_cap_percent series_id venue_id min_price sold_out next_session_at sessions { id event_id starts_at ends_at label status cancel_reason } ticket_types { id event_id name description price total available sold held min_per_order max_per_order max_per_user sale_starts_at sale_ends_at seated active sort_order session_id on_sale sold_out } } next_cursor }'} }', variables: {
      'cursor': ?cursor,
      'limit': ?limit,
    }
    );
    return data['ticketEventReviewQueue'] == null ? null : TicketEventPage.fromJson((data['ticketEventReviewQueue'] as Map).cast<String, dynamic>());
  }

  Future<TicketOrderPage?> ticketOrders({String? status, int? eventId, String? cursor, int? limit, String? selectionSet}) async {
    final data = await _executor.query('query ticketOrders(\$status: String, \$event_id: Int, \$cursor: String, \$limit: Int) { ticketOrders(status: \$status, event_id: \$event_id, cursor: \$cursor, limit: \$limit) ${selectionSet ?? '{ items { id event_id user_id status currency subtotal discount total promo_code buyer_name buyer_email buyer_phone session_id expires_at payment_method paid_at refund_amount status_note created_at items { ticket_type_id name quantity unit_price seat_ids } tickets { id order_id event_id event_title event_starts_at event_ends_at session_id session_label venue ticket_type_id type_name seat_id seat code status holder_id holder_name holder_email transfer_count checked_in_at } } next_cursor }'} }', variables: {
      'status': ?status,
      'event_id': ?eventId,
      'cursor': ?cursor,
      'limit': ?limit,
    }
    );
    return data['ticketOrders'] == null ? null : TicketOrderPage.fromJson((data['ticketOrders'] as Map).cast<String, dynamic>());
  }

  Future<TicketOrder?> ticketOrder({required int id, String? selectionSet}) async {
    final data = await _executor.query('query ticketOrder(\$id: Int!) { ticketOrder(id: \$id) ${selectionSet ?? '{ id event_id user_id status currency subtotal discount total promo_code buyer_name buyer_email buyer_phone session_id expires_at payment_method paid_at refund_amount status_note created_at items { ticket_type_id name quantity unit_price seat_ids } tickets { id order_id event_id event_title event_starts_at event_ends_at session_id session_label venue ticket_type_id type_name seat_id seat code status holder_id holder_name holder_email transfer_count checked_in_at } }'} }', variables: {
      'id': id,
    }
    );
    return data['ticketOrder'] == null ? null : TicketOrder.fromJson((data['ticketOrder'] as Map).cast<String, dynamic>());
  }

  Future<TicketInvoice?> ticketInvoice({required int id, String? selectionSet}) async {
    final data = await _executor.query('query ticketInvoice(\$id: Int!) { ticketInvoice(id: \$id) ${selectionSet ?? '{ number issued_at order_id seller_name seller_tax_id seller_address buyer_name buyer_email event_title event_date venue currency lines { description quantity unit_price amount } subtotal discount promo_code total vat_percent vat payment_method status refund_amount }'} }', variables: {
      'id': id,
    }
    );
    return data['ticketInvoice'] == null ? null : TicketInvoice.fromJson((data['ticketInvoice'] as Map).cast<String, dynamic>());
  }

  Future<List<TicketDocument>?> ticketOrderDocuments({required int id, String? selectionSet}) async {
    final data = await _executor.query('query ticketOrderDocuments(\$id: Int!) { ticketOrderDocuments(id: \$id) ${selectionSet ?? '{ kind file_name size created_at url }'} }', variables: {
      'id': id,
    }
    );
    return (data['ticketOrderDocuments'] as List?)?.map((e) => (e == null ? null : TicketDocument.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<TicketTicketPage?> tickets({String? status, int? eventId, String? cursor, int? limit, String? selectionSet}) async {
    final data = await _executor.query('query tickets(\$status: String, \$event_id: Int, \$cursor: String, \$limit: Int) { tickets(status: \$status, event_id: \$event_id, cursor: \$cursor, limit: \$limit) ${selectionSet ?? '{ items { id order_id event_id event_title event_starts_at event_ends_at session_id session_label venue ticket_type_id type_name seat_id seat code status holder_id holder_name holder_email transfer_count checked_in_at } next_cursor }'} }', variables: {
      'status': ?status,
      'event_id': ?eventId,
      'cursor': ?cursor,
      'limit': ?limit,
    }
    );
    return data['tickets'] == null ? null : TicketTicketPage.fromJson((data['tickets'] as Map).cast<String, dynamic>());
  }

  Future<TicketTicketPage?> ticketMyTickets({String? status, int? eventId, String? cursor, int? limit, String? selectionSet}) async {
    final data = await _executor.query('query ticketMyTickets(\$status: String, \$event_id: Int, \$cursor: String, \$limit: Int) { ticketMyTickets(status: \$status, event_id: \$event_id, cursor: \$cursor, limit: \$limit) ${selectionSet ?? '{ items { id order_id event_id event_title event_starts_at event_ends_at session_id session_label venue ticket_type_id type_name seat_id seat code status holder_id holder_name holder_email transfer_count checked_in_at } next_cursor }'} }', variables: {
      'status': ?status,
      'event_id': ?eventId,
      'cursor': ?cursor,
      'limit': ?limit,
    }
    );
    return data['ticketMyTickets'] == null ? null : TicketTicketPage.fromJson((data['ticketMyTickets'] as Map).cast<String, dynamic>());
  }

  Future<TicketTicket?> ticket({required int id, String? selectionSet}) async {
    final data = await _executor.query('query ticket(\$id: Int!) { ticket(id: \$id) ${selectionSet ?? '{ id order_id event_id event_title event_starts_at event_ends_at session_id session_label venue ticket_type_id type_name seat_id seat code status holder_id holder_name holder_email transfer_count checked_in_at }'} }', variables: {
      'id': id,
    }
    );
    return data['ticket'] == null ? null : TicketTicket.fromJson((data['ticket'] as Map).cast<String, dynamic>());
  }

  Future<List<TicketTicketEvent>?> ticketHistory({required int id, String? selectionSet}) async {
    final data = await _executor.query('query ticketHistory(\$id: Int!) { ticketHistory(id: \$id) ${selectionSet ?? '{ id event from_status to_status actor note created_at }'} }', variables: {
      'id': id,
    }
    );
    return (data['ticketHistory'] as List?)?.map((e) => (e == null ? null : TicketTicketEvent.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<TicketResalePage?> ticketMyResale({String? status, String? cursor, int? limit, String? selectionSet}) async {
    final data = await _executor.public('query ticketMyResale(\$status: String, \$cursor: String, \$limit: Int) { ticketMyResale(status: \$status, cursor: \$cursor, limit: \$limit) ${selectionSet ?? '{ items { id ticket_id event_id type_name seat price face_price currency status created_at seller_id fee } next_cursor }'} }', variables: {
      'status': ?status,
      'cursor': ?cursor,
      'limit': ?limit,
    }
    );
    return data['ticketMyResale'] == null ? null : TicketResalePage.fromJson((data['ticketMyResale'] as Map).cast<String, dynamic>());
  }

  Future<TicketTransferPage?> ticketTransfers({String? direction, String? cursor, int? limit, String? selectionSet}) async {
    final data = await _executor.query('query ticketTransfers(\$direction: String, \$cursor: String, \$limit: Int) { ticketTransfers(direction: \$direction, cursor: \$cursor, limit: \$limit) ${selectionSet ?? '{ items { id ticket_id from_user to_user to_email status message created_at expires_at resolved_at ticket { id order_id event_id event_title event_starts_at event_ends_at session_id session_label venue ticket_type_id type_name seat_id seat code status holder_id holder_name holder_email transfer_count checked_in_at } } next_cursor }'} }', variables: {
      'direction': ?direction,
      'cursor': ?cursor,
      'limit': ?limit,
    }
    );
    return data['ticketTransfers'] == null ? null : TicketTransferPage.fromJson((data['ticketTransfers'] as Map).cast<String, dynamic>());
  }

  Future<List<TicketWaitlist>?> ticketWaitlist({String? selectionSet}) async {
    final data = await _executor.query('query ticketWaitlist { ticketWaitlist ${selectionSet ?? '{ ticket_type_id event_id event_title type_name created_at }'} }'
    );
    return (data['ticketWaitlist'] as List?)?.map((e) => (e == null ? null : TicketWaitlist.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<List<TicketEvent>?> ticketWishlist({int? limit, int? offset, String? selectionSet}) async {
    final data = await _executor.query('query ticketWishlist(\$limit: Int, \$offset: Int) { ticketWishlist(limit: \$limit, offset: \$offset) ${selectionSet ?? '{ id organizer_id title description category city venue address banner_url starts_at ends_at currency status review_note featured wallet_id refund_cutoff_hours transferable resale_cap_percent series_id venue_id min_price sold_out next_session_at sessions { id event_id starts_at ends_at label status cancel_reason } ticket_types { id event_id name description price total available sold held min_per_order max_per_order max_per_user sale_starts_at sale_ends_at seated active sort_order session_id on_sale sold_out } }'} }', variables: {
      'limit': ?limit,
      'offset': ?offset,
    }
    );
    return (data['ticketWishlist'] as List?)?.map((e) => (e == null ? null : TicketEvent.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<TicketNotificationPage?> ticketNotifications({bool? unread, String? cursor, int? limit, String? selectionSet}) async {
    final data = await _executor.query('query ticketNotifications(\$unread: Boolean, \$cursor: String, \$limit: Int) { ticketNotifications(unread: \$unread, cursor: \$cursor, limit: \$limit) ${selectionSet ?? '{ items { id kind title body order_id event_id created_at read_at } next_cursor }'} }', variables: {
      'unread': ?unread,
      'cursor': ?cursor,
      'limit': ?limit,
    }
    );
    return data['ticketNotifications'] == null ? null : TicketNotificationPage.fromJson((data['ticketNotifications'] as Map).cast<String, dynamic>());
  }

  Future<TicketUnreadCount?> ticketUnreadNotifications({String? selectionSet}) async {
    final data = await _executor.query('query ticketUnreadNotifications { ticketUnreadNotifications ${selectionSet ?? '{ unread }'} }'
    );
    return data['ticketUnreadNotifications'] == null ? null : TicketUnreadCount.fromJson((data['ticketUnreadNotifications'] as Map).cast<String, dynamic>());
  }

  Future<Object?> ticketCreateEvent({required TicketEventInput input}) async {
    final data = await _executor.mutation('mutation ticketCreateEvent(\$input: TicketEventInput!) { ticketCreateEvent(input: \$input) }', variables: {
      'input': input.toJson(),
    }
    );
    return data['ticketCreateEvent'];
  }

  Future<Object?> ticketUpdateEvent({required int id, required TicketEventInput input}) async {
    final data = await _executor.mutation('mutation ticketUpdateEvent(\$id: Int64!, \$input: TicketEventInput!) { ticketUpdateEvent(id: \$id, input: \$input) }', variables: {
      'id': id,
      'input': input.toJson(),
    }
    );
    return data['ticketUpdateEvent'];
  }

  Future<Object?> ticketSubmitEvent({required int id}) async {
    final data = await _executor.mutation('mutation ticketSubmitEvent(\$id: Int64!) { ticketSubmitEvent(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['ticketSubmitEvent'];
  }

  Future<Object?> ticketApproveEvent({required int id}) async {
    final data = await _executor.mutation('mutation ticketApproveEvent(\$id: Int64!) { ticketApproveEvent(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['ticketApproveEvent'];
  }

  Future<Object?> ticketRejectEvent({required int id, required String reason}) async {
    final data = await _executor.mutation('mutation ticketRejectEvent(\$id: Int64!, \$reason: String!) { ticketRejectEvent(id: \$id, reason: \$reason) }', variables: {
      'id': id,
      'reason': reason,
    }
    );
    return data['ticketRejectEvent'];
  }

  Future<Object?> ticketSetFeatured({required int id, required Map<String, dynamic> input}) async {
    final data = await _executor.mutation('mutation ticketSetFeatured(\$id: Int64!, \$input: Map!) { ticketSetFeatured(id: \$id, input: \$input) }', variables: {
      'id': id,
      'input': input,
    }
    );
    return data['ticketSetFeatured'];
  }

  Future<Object?> ticketDuplicateEvent({required int id, required TicketDuplicateInput input}) async {
    final data = await _executor.mutation('mutation ticketDuplicateEvent(\$id: Int64!, \$input: TicketDuplicateInput!) { ticketDuplicateEvent(id: \$id, input: \$input) }', variables: {
      'id': id,
      'input': input.toJson(),
    }
    );
    return data['ticketDuplicateEvent'];
  }

  Future<Object?> ticketCancelEvent({required int id, String? reason}) async {
    final data = await _executor.mutation('mutation ticketCancelEvent(\$id: Int64!, \$reason: String) { ticketCancelEvent(id: \$id, reason: \$reason) }', variables: {
      'id': id,
      'reason': ?reason,
    }
    );
    return data['ticketCancelEvent'];
  }

  Future<Object?> ticketPreviewVenue({required Map<String, dynamic> input}) async {
    final data = await _executor.mutation('mutation ticketPreviewVenue(\$input: Map!) { ticketPreviewVenue(input: \$input) }', variables: {
      'input': input,
    }
    );
    return data['ticketPreviewVenue'];
  }

  Future<Object?> ticketCreateVenue({required Map<String, dynamic> input}) async {
    final data = await _executor.mutation('mutation ticketCreateVenue(\$input: Map!) { ticketCreateVenue(input: \$input) }', variables: {
      'input': input,
    }
    );
    return data['ticketCreateVenue'];
  }

  Future<Object?> ticketUpdateVenue({required int id, required Map<String, dynamic> input}) async {
    final data = await _executor.mutation('mutation ticketUpdateVenue(\$id: Int64!, \$input: Map!) { ticketUpdateVenue(id: \$id, input: \$input) }', variables: {
      'id': id,
      'input': input,
    }
    );
    return data['ticketUpdateVenue'];
  }

  Future<Object?> ticketDeleteVenue({required int id}) async {
    final data = await _executor.mutation('mutation ticketDeleteVenue(\$id: Int64!) { ticketDeleteVenue(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['ticketDeleteVenue'];
  }

  Future<Object?> ticketCopyVenue({required int id}) async {
    final data = await _executor.mutation('mutation ticketCopyVenue(\$id: Int64!) { ticketCopyVenue(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['ticketCopyVenue'];
  }

  Future<Object?> ticketShareVenue({required int id, required Map<String, dynamic> input}) async {
    final data = await _executor.mutation('mutation ticketShareVenue(\$id: Int64!, \$input: Map!) { ticketShareVenue(id: \$id, input: \$input) }', variables: {
      'id': id,
      'input': input,
    }
    );
    return data['ticketShareVenue'];
  }

  Future<Object?> ticketApplyVenue({required int id, required Map<String, dynamic> input}) async {
    final data = await _executor.mutation('mutation ticketApplyVenue(\$id: Int64!, \$input: Map!) { ticketApplyVenue(id: \$id, input: \$input) }', variables: {
      'id': id,
      'input': input,
    }
    );
    return data['ticketApplyVenue'];
  }

  Future<Object?> ticketCreateSession({required int id, required TicketSessionInput input}) async {
    final data = await _executor.mutation('mutation ticketCreateSession(\$id: Int64!, \$input: TicketSessionInput!) { ticketCreateSession(id: \$id, input: \$input) }', variables: {
      'id': id,
      'input': input.toJson(),
    }
    );
    return data['ticketCreateSession'];
  }

  Future<Object?> ticketUpdateSession({required int id, required int session, required TicketSessionInput input}) async {
    final data = await _executor.mutation('mutation ticketUpdateSession(\$id: Int64!, \$session: Int64!, \$input: TicketSessionInput!) { ticketUpdateSession(id: \$id, session: \$session, input: \$input) }', variables: {
      'id': id,
      'session': session,
      'input': input.toJson(),
    }
    );
    return data['ticketUpdateSession'];
  }

  Future<Object?> ticketDeleteSession({required int id, required int session}) async {
    final data = await _executor.mutation('mutation ticketDeleteSession(\$id: Int64!, \$session: Int64!) { ticketDeleteSession(id: \$id, session: \$session) }', variables: {
      'id': id,
      'session': session,
    }
    );
    return data['ticketDeleteSession'];
  }

  Future<Object?> ticketCancelSession({required int id, required int session, String? reason}) async {
    final data = await _executor.mutation('mutation ticketCancelSession(\$id: Int64!, \$session: Int64!, \$reason: String) { ticketCancelSession(id: \$id, session: \$session, reason: \$reason) }', variables: {
      'id': id,
      'session': session,
      'reason': ?reason,
    }
    );
    return data['ticketCancelSession'];
  }

  Future<Object?> ticketCreateTicketType({required int id, required TicketTypeInput input}) async {
    final data = await _executor.mutation('mutation ticketCreateTicketType(\$id: Int64!, \$input: TicketTypeInput!) { ticketCreateTicketType(id: \$id, input: \$input) }', variables: {
      'id': id,
      'input': input.toJson(),
    }
    );
    return data['ticketCreateTicketType'];
  }

  Future<Object?> ticketUpdateTicketType({required int id, required int type_, required TicketTypeInput input}) async {
    final data = await _executor.mutation('mutation ticketUpdateTicketType(\$id: Int64!, \$type: Int64!, \$input: TicketTypeInput!) { ticketUpdateTicketType(id: \$id, type: \$type, input: \$input) }', variables: {
      'id': id,
      'type': type_,
      'input': input.toJson(),
    }
    );
    return data['ticketUpdateTicketType'];
  }

  Future<Object?> ticketGenerateSeats({required int id, required int type_, required Map<String, dynamic> input}) async {
    final data = await _executor.mutation('mutation ticketGenerateSeats(\$id: Int64!, \$type: Int64!, \$input: Map!) { ticketGenerateSeats(id: \$id, type: \$type, input: \$input) }', variables: {
      'id': id,
      'type': type_,
      'input': input,
    }
    );
    return data['ticketGenerateSeats'];
  }

  Future<Object?> ticketSetSeatMap({required int id, required int type_, required Map<String, dynamic> input}) async {
    final data = await _executor.mutation('mutation ticketSetSeatMap(\$id: Int64!, \$type: Int64!, \$input: Map!) { ticketSetSeatMap(id: \$id, type: \$type, input: \$input) }', variables: {
      'id': id,
      'type': type_,
      'input': input,
    }
    );
    return data['ticketSetSeatMap'];
  }

  Future<Object?> ticketJoinWaitlist({required int id, required int type_}) async {
    final data = await _executor.mutation('mutation ticketJoinWaitlist(\$id: Int64!, \$type: Int64!) { ticketJoinWaitlist(id: \$id, type: \$type) }', variables: {
      'id': id,
      'type': type_,
    }
    );
    return data['ticketJoinWaitlist'];
  }

  Future<Object?> ticketLeaveWaitlist({required int id, required int type_}) async {
    final data = await _executor.mutation('mutation ticketLeaveWaitlist(\$id: Int64!, \$type: Int64!) { ticketLeaveWaitlist(id: \$id, type: \$type) }', variables: {
      'id': id,
      'type': type_,
    }
    );
    return data['ticketLeaveWaitlist'];
  }

  Future<Object?> ticketCreatePromotion({required int id, required TicketPromotionInput input}) async {
    final data = await _executor.mutation('mutation ticketCreatePromotion(\$id: Int64!, \$input: TicketPromotionInput!) { ticketCreatePromotion(id: \$id, input: \$input) }', variables: {
      'id': id,
      'input': input.toJson(),
    }
    );
    return data['ticketCreatePromotion'];
  }

  Future<Object?> ticketUpdatePromotion({required int id, required int promo, required TicketPromotionInput input}) async {
    final data = await _executor.mutation('mutation ticketUpdatePromotion(\$id: Int64!, \$promo: Int64!, \$input: TicketPromotionInput!) { ticketUpdatePromotion(id: \$id, promo: \$promo, input: \$input) }', variables: {
      'id': id,
      'promo': promo,
      'input': input.toJson(),
    }
    );
    return data['ticketUpdatePromotion'];
  }

  Future<Object?> ticketCheckIn({required int id, required String code, String? scannedAt}) async {
    final data = await _executor.mutation('mutation ticketCheckIn(\$id: Int64!, \$code: String!, \$scanned_at: String) { ticketCheckIn(id: \$id, code: \$code, scanned_at: \$scanned_at) }', variables: {
      'id': id,
      'code': code,
      'scanned_at': ?scannedAt,
    }
    );
    return data['ticketCheckIn'];
  }

  Future<Object?> ticketBatchCheckIn({required int id, required Map<String, dynamic> input}) async {
    final data = await _executor.mutation('mutation ticketBatchCheckIn(\$id: Int64!, \$input: Map!) { ticketBatchCheckIn(id: \$id, input: \$input) }', variables: {
      'id': id,
      'input': input,
    }
    );
    return data['ticketBatchCheckIn'];
  }

  Future<Object?> ticketInvite({required int id, required TicketInviteInput input}) async {
    final data = await _executor.mutation('mutation ticketInvite(\$id: Int64!, \$input: TicketInviteInput!) { ticketInvite(id: \$id, input: \$input) }', variables: {
      'id': id,
      'input': input.toJson(),
    }
    );
    return data['ticketInvite'];
  }

  Future<Object?> ticketAddStaff({required int id, required int userId}) async {
    final data = await _executor.mutation('mutation ticketAddStaff(\$id: Int64!, \$user_id: Int64!) { ticketAddStaff(id: \$id, user_id: \$user_id) }', variables: {
      'id': id,
      'user_id': userId,
    }
    );
    return data['ticketAddStaff'];
  }

  Future<Object?> ticketRemoveStaff({required int id, required int user}) async {
    final data = await _executor.mutation('mutation ticketRemoveStaff(\$id: Int64!, \$user: Int64!) { ticketRemoveStaff(id: \$id, user: \$user) }', variables: {
      'id': id,
      'user': user,
    }
    );
    return data['ticketRemoveStaff'];
  }

  Future<Object?> ticketReserve({required TicketReserveInput input}) async {
    final data = await _executor.mutation('mutation ticketReserve(\$input: TicketReserveInput!) { ticketReserve(input: \$input) }', variables: {
      'input': input.toJson(),
    }
    );
    return data['ticketReserve'];
  }

  Future<Object?> ticketPayOrder({required int id, required String method, String? provider}) async {
    final data = await _executor.mutation('mutation ticketPayOrder(\$id: Int64!, \$method: String!, \$provider: String) { ticketPayOrder(id: \$id, method: \$method, provider: \$provider) }', variables: {
      'id': id,
      'method': method,
      'provider': ?provider,
    }
    );
    return data['ticketPayOrder'];
  }

  Future<Object?> ticketCancelOrder({required int id, String? reason}) async {
    final data = await _executor.mutation('mutation ticketCancelOrder(\$id: Int64!, \$reason: String) { ticketCancelOrder(id: \$id, reason: \$reason) }', variables: {
      'id': id,
      'reason': ?reason,
    }
    );
    return data['ticketCancelOrder'];
  }

  Future<Object?> ticketSetHolder({required int id, required String name_, String? email}) async {
    final data = await _executor.mutation('mutation ticketSetHolder(\$id: Int64!, \$name: String!, \$email: String) { ticketSetHolder(id: \$id, name: \$name, email: \$email) }', variables: {
      'id': id,
      'name': name_,
      'email': ?email,
    }
    );
    return data['ticketSetHolder'];
  }

  Future<Object?> ticketReissue({required int id}) async {
    final data = await _executor.mutation('mutation ticketReissue(\$id: Int64!) { ticketReissue(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['ticketReissue'];
  }

  Future<Object?> ticketListForResale({required int id, required int price}) async {
    final data = await _executor.mutation('mutation ticketListForResale(\$id: Int64!, \$price: Int64!) { ticketListForResale(id: \$id, price: \$price) }', variables: {
      'id': id,
      'price': price,
    }
    );
    return data['ticketListForResale'];
  }

  Future<Object?> ticketCancelResale({required int id}) async {
    final data = await _executor.mutation('mutation ticketCancelResale(\$id: Int64!) { ticketCancelResale(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['ticketCancelResale'];
  }

  Future<Object?> ticketBuyResale({required int id}) async {
    final data = await _executor.mutation('mutation ticketBuyResale(\$id: Int64!) { ticketBuyResale(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['ticketBuyResale'];
  }

  Future<Object?> ticketOfferTransfer({required int id, int? toUserId, String? toEmail, String? message}) async {
    final data = await _executor.mutation('mutation ticketOfferTransfer(\$id: Int64!, \$to_user_id: Int64, \$to_email: String, \$message: String) { ticketOfferTransfer(id: \$id, to_user_id: \$to_user_id, to_email: \$to_email, message: \$message) }', variables: {
      'id': id,
      'to_user_id': ?toUserId,
      'to_email': ?toEmail,
      'message': ?message,
    }
    );
    return data['ticketOfferTransfer'];
  }

  Future<Object?> ticketAcceptTransfer({required int id}) async {
    final data = await _executor.mutation('mutation ticketAcceptTransfer(\$id: Int64!) { ticketAcceptTransfer(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['ticketAcceptTransfer'];
  }

  Future<Object?> ticketDeclineTransfer({required int id}) async {
    final data = await _executor.mutation('mutation ticketDeclineTransfer(\$id: Int64!) { ticketDeclineTransfer(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['ticketDeclineTransfer'];
  }

  Future<Object?> ticketCancelTransfer({required int id}) async {
    final data = await _executor.mutation('mutation ticketCancelTransfer(\$id: Int64!) { ticketCancelTransfer(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['ticketCancelTransfer'];
  }

  Future<Object?> ticketVoid({required int id, String? reason}) async {
    final data = await _executor.mutation('mutation ticketVoid(\$id: Int64!, \$reason: String) { ticketVoid(id: \$id, reason: \$reason) }', variables: {
      'id': id,
      'reason': ?reason,
    }
    );
    return data['ticketVoid'];
  }

  Future<Object?> ticketRevertCheckIn({required int id}) async {
    final data = await _executor.mutation('mutation ticketRevertCheckIn(\$id: Int64!) { ticketRevertCheckIn(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['ticketRevertCheckIn'];
  }

  Future<Object?> ticketAddWishlist({required int id}) async {
    final data = await _executor.mutation('mutation ticketAddWishlist(\$id: Int64!) { ticketAddWishlist(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['ticketAddWishlist'];
  }

  Future<Object?> ticketRemoveWishlist({required int id}) async {
    final data = await _executor.mutation('mutation ticketRemoveWishlist(\$id: Int64!) { ticketRemoveWishlist(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['ticketRemoveWishlist'];
  }

  Future<Object?> ticketMarkAllNotificationsRead() async {
    final data = await _executor.mutation('mutation ticketMarkAllNotificationsRead { ticketMarkAllNotificationsRead }'
    );
    return data['ticketMarkAllNotificationsRead'];
  }

  Future<Object?> ticketMarkNotificationRead({required int id}) async {
    final data = await _executor.mutation('mutation ticketMarkNotificationRead(\$id: Int64!) { ticketMarkNotificationRead(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['ticketMarkNotificationRead'];
  }

}
