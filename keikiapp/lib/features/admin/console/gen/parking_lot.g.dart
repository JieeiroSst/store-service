// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

// ignore_for_file: unused_import

import '../../../apps/kit/kit.dart';
import '../../../../gen/models.g.dart';
import '../admin_op.dart';

final parkingLotAdminOps = <AdminOp>[
  AdminOp(
    'parkingHistory',
    resource: 'history',
    verb: '',
    fields: [
      KField('plate', 'Plate', required: true),
      KField('limit', 'Limit', kind: FieldKind.number),
      KField('offset', 'Offset', kind: FieldKind.number),
    ],
    run: (api, v) async => adminPlain(await api.parkingLot.parkingHistory(plate: v.need(v.str('plate'), 'plate'), limit: v.integer('limit'), offset: v.integer('offset'))),
  ),
  AdminOp(
    'parkingCheckIn',
    resource: 'general',
    verb: 'check',
    mutation: true,
    fields: [
      KField('license_plate', 'License plate', required: true),
      KField('vehicle_type', 'Vehicle type', required: true),
      KField('entry_gate_id', 'Entry gate ID'),
    ],
    run: (api, v) async => adminPlain(await api.parkingLot.parkingCheckIn(licensePlate: v.need(v.str('license_plate'), 'license_plate'), vehicleType: v.need(v.str('vehicle_type'), 'vehicle_type'), entryGateId: v.str('entry_gate_id'))),
  ),
  AdminOp(
    'parkingCheckOut',
    resource: 'out',
    verb: 'check',
    mutation: true,
    fields: [
      KField('ticket_id', 'Ticket ID', required: true),
      KField('method', 'Method', required: true),
      KField('exit_gate_id', 'Exit gate ID'),
    ],
    run: (api, v) async => adminPlain(await api.parkingLot.parkingCheckOut(ticketId: v.need(v.str('ticket_id'), 'ticket_id'), method: v.need(v.str('method'), 'method'), exitGateId: v.str('exit_gate_id'))),
  ),
];
