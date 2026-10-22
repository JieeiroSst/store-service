// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

import '../gateway_executor.dart';
import '../models.g.dart';

class BookingMiniApi {
  const BookingMiniApi(this._executor);

  final GatewayExecutor _executor;

  Future<List<BookingMiniRoom>?> bookingMiniAvailableRooms({required String fromDate, required String toDate, String? selectionSet}) async {
    final data = await _executor.public('query bookingMiniAvailableRooms(\$from_date: String!, \$to_date: String!) { bookingMiniAvailableRooms(from_date: \$from_date, to_date: \$to_date) ${selectionSet ?? '{ room_no room_type room_type_details { id description capacity } bed_type bed_count amenities }'} }', variables: {
      'from_date': fromDate,
      'to_date': toDate,
    }
    );
    return (data['bookingMiniAvailableRooms'] as List?)?.map((e) => (e == null ? null : BookingMiniRoom.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<BookingMiniBooking?> bookingMiniBooking({required int id, String? selectionSet}) async {
    final data = await _executor.query('query bookingMiniBooking(\$id: Int!) { bookingMiniBooking(id: \$id) ${selectionSet ?? '{ id from_date to_date room_no passenger_id passenger { id name } room { room_no room_type room_type_details { id description capacity } bed_type bed_count amenities } }'} }', variables: {
      'id': id,
    }
    );
    return data['bookingMiniBooking'] == null ? null : BookingMiniBooking.fromJson((data['bookingMiniBooking'] as Map).cast<String, dynamic>());
  }

  Future<List<BookingMiniBooking>?> myBookingMiniBookings({String? selectionSet}) async {
    final data = await _executor.query('query myBookingMiniBookings { myBookingMiniBookings ${selectionSet ?? '{ id from_date to_date room_no passenger_id passenger { id name } room { room_no room_type room_type_details { id description capacity } bed_type bed_count amenities } }'} }'
    );
    return (data['myBookingMiniBookings'] as List?)?.map((e) => (e == null ? null : BookingMiniBooking.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<List<BookingMiniPassenger>?> myBookingMiniPassengers({String? selectionSet}) async {
    final data = await _executor.query('query myBookingMiniPassengers { myBookingMiniPassengers ${selectionSet ?? '{ id name }'} }'
    );
    return (data['myBookingMiniPassengers'] as List?)?.map((e) => (e == null ? null : BookingMiniPassenger.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<BookingMiniPassenger?> bookingMiniPassenger({required int id, String? selectionSet}) async {
    final data = await _executor.query('query bookingMiniPassenger(\$id: Int!) { bookingMiniPassenger(id: \$id) ${selectionSet ?? '{ id name }'} }', variables: {
      'id': id,
    }
    );
    return data['bookingMiniPassenger'] == null ? null : BookingMiniPassenger.fromJson((data['bookingMiniPassenger'] as Map).cast<String, dynamic>());
  }

  Future<BookingMiniBooking?> bookingMiniCreateBooking({required BookingMiniBookingInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation bookingMiniCreateBooking(\$input: BookingMiniBookingInput!) { bookingMiniCreateBooking(input: \$input) ${selectionSet ?? '{ id from_date to_date room_no passenger_id passenger { id name } room { room_no room_type room_type_details { id description capacity } bed_type bed_count amenities } }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['bookingMiniCreateBooking'] == null ? null : BookingMiniBooking.fromJson((data['bookingMiniCreateBooking'] as Map).cast<String, dynamic>());
  }

  Future<bool?> bookingMiniCancelBooking({required int id}) async {
    final data = await _executor.mutation('mutation bookingMiniCancelBooking(\$id: Int!) { bookingMiniCancelBooking(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['bookingMiniCancelBooking'] as bool?;
  }

  Future<BookingMiniPassenger?> bookingMiniRegisterPassenger({required String name_, String? selectionSet}) async {
    final data = await _executor.mutation('mutation bookingMiniRegisterPassenger(\$name: String!) { bookingMiniRegisterPassenger(name: \$name) ${selectionSet ?? '{ id name }'} }', variables: {
      'name': name_,
    }
    );
    return data['bookingMiniRegisterPassenger'] == null ? null : BookingMiniPassenger.fromJson((data['bookingMiniRegisterPassenger'] as Map).cast<String, dynamic>());
  }

}
