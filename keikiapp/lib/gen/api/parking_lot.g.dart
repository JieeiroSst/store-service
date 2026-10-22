// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

import '../gateway_executor.dart';
import '../models.g.dart';

class ParkingLotApi {
  const ParkingLotApi(this._executor);

  final GatewayExecutor _executor;

  Future<List<ParkingSpotAvailability>?> parkingAvailableSpots({String? selectionSet}) async {
    final data = await _executor.public('query parkingAvailableSpots { parkingAvailableSpots ${selectionSet ?? '{ type available total }'} }'
    );
    return (data['parkingAvailableSpots'] as List?)?.map((e) => (e == null ? null : ParkingSpotAvailability.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<List<ParkingRate>?> parkingRates({String? selectionSet}) async {
    final data = await _executor.public('query parkingRates { parkingRates ${selectionSet ?? '{ RateID VehicleType HourlyRate DailyMaxRate EffectiveFrom }'} }'
    );
    return (data['parkingRates'] as List?)?.map((e) => (e == null ? null : ParkingRate.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<ParkingTicket?> parkingTicket({required String id, String? selectionSet}) async {
    final data = await _executor.query('query parkingTicket(\$id: String!) { parkingTicket(id: \$id) ${selectionSet ?? '{ TicketID VehiclePlate SpotID EntryGateID ExitGateID ParkedTime LeaveTime Status }'} }', variables: {
      'id': id,
    }
    );
    return data['parkingTicket'] == null ? null : ParkingTicket.fromJson((data['parkingTicket'] as Map).cast<String, dynamic>());
  }

  Future<List<ParkingTicket>?> parkingHistory({required String plate, int? limit, int? offset, String? selectionSet}) async {
    final data = await _executor.query('query parkingHistory(\$plate: String!, \$limit: Int, \$offset: Int) { parkingHistory(plate: \$plate, limit: \$limit, offset: \$offset) ${selectionSet ?? '{ TicketID VehiclePlate SpotID EntryGateID ExitGateID ParkedTime LeaveTime Status }'} }', variables: {
      'plate': plate,
      'limit': ?limit,
      'offset': ?offset,
    }
    );
    return (data['parkingHistory'] as List?)?.map((e) => (e == null ? null : ParkingTicket.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<ParkingTicket?> parkingCheckIn({required String licensePlate, required String vehicleType, String? entryGateId, String? selectionSet}) async {
    final data = await _executor.mutation('mutation parkingCheckIn(\$license_plate: String!, \$vehicle_type: String!, \$entry_gate_id: String) { parkingCheckIn(license_plate: \$license_plate, vehicle_type: \$vehicle_type, entry_gate_id: \$entry_gate_id) ${selectionSet ?? '{ TicketID VehiclePlate SpotID EntryGateID ExitGateID ParkedTime LeaveTime Status }'} }', variables: {
      'license_plate': licensePlate,
      'vehicle_type': vehicleType,
      'entry_gate_id': ?entryGateId,
    }
    );
    return data['parkingCheckIn'] == null ? null : ParkingTicket.fromJson((data['parkingCheckIn'] as Map).cast<String, dynamic>());
  }

  Future<ParkingCheckOut?> parkingCheckOut({required String ticketId, required String method, String? exitGateId, String? selectionSet}) async {
    final data = await _executor.mutation('mutation parkingCheckOut(\$ticket_id: String!, \$method: String!, \$exit_gate_id: String) { parkingCheckOut(ticket_id: \$ticket_id, method: \$method, exit_gate_id: \$exit_gate_id) ${selectionSet ?? '{ ticket { TicketID VehiclePlate SpotID EntryGateID ExitGateID ParkedTime LeaveTime Status } payment }'} }', variables: {
      'ticket_id': ticketId,
      'method': method,
      'exit_gate_id': ?exitGateId,
    }
    );
    return data['parkingCheckOut'] == null ? null : ParkingCheckOut.fromJson((data['parkingCheckOut'] as Map).cast<String, dynamic>());
  }

}
