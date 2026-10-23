// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

// ignore_for_file: unused_import

import '../../../apps/kit/kit.dart';
import '../../../../gen/models.g.dart';
import '../admin_op.dart';

final identifiedAdminOps = <AdminOp>[
  AdminOp(
    'gymCreateClient',
    resource: 'client',
    verb: 'create',
    mutation: true,
    fields: [
      KField('input.first_name', 'First name'),
      KField('input.last_name', 'Last name'),
      KField('input.email', 'Email', kind: FieldKind.email),
    ],
    run: (api, v) async => adminPlain(await api.identified.gymCreateClient(input: GymClientInput(firstName: v.str('input.first_name'), lastName: v.str('input.last_name'), email: v.str('input.email')))),
  ),
  AdminOp(
    'gymUpdateClient',
    resource: 'client',
    verb: 'update',
    mutation: true,
    fields: [
      KField('client_id', 'Client ID', required: true),
      KField('input.first_name', 'First name'),
      KField('input.last_name', 'Last name'),
      KField('input.email', 'Email', kind: FieldKind.email),
    ],
    run: (api, v) async => adminPlain(await api.identified.gymUpdateClient(clientId: v.need(v.str('client_id'), 'client_id'), input: GymClientInput(firstName: v.str('input.first_name'), lastName: v.str('input.last_name'), email: v.str('input.email')))),
  ),
  AdminOp(
    'gymArchiveClient',
    resource: 'client',
    verb: 'archive',
    mutation: true,
    destructive: true,
    fields: [
      KField('client_id', 'Client ID', required: true),
    ],
    run: (api, v) async => adminPlain(await api.identified.gymArchiveClient(clientId: v.need(v.str('client_id'), 'client_id'))),
  ),
  AdminOp(
    'gymExportClients',
    resource: 'client',
    verb: 'export',
    mutation: true,
    fields: [
      KField('format', 'Format', required: true),
    ],
    run: (api, v) async => adminPlain(await api.identified.gymExportClients(format: v.need(v.str('format'), 'format'))),
  ),
  AdminOp(
    'gymPausePass',
    resource: 'pass',
    verb: 'pause',
    mutation: true,
    fields: [
      KField('gym_pass_id', 'Gym pass ID', required: true),
      KField('days', 'Days', kind: FieldKind.number, required: true),
    ],
    run: (api, v) async => adminPlain(await api.identified.gymPausePass(gymPassId: v.need(v.str('gym_pass_id'), 'gym_pass_id'), days: v.need(v.integer('days'), 'days'))),
  ),
  AdminOp(
    'gymRenewPass',
    resource: 'pass',
    verb: 'renew',
    mutation: true,
    fields: [
      KField('gym_pass_id', 'Gym pass ID', required: true),
    ],
    run: (api, v) async => adminPlain(await api.identified.gymRenewPass(gymPassId: v.need(v.str('gym_pass_id'), 'gym_pass_id'))),
  ),
  AdminOp(
    'gymCreateClass',
    resource: 'class',
    verb: 'create',
    mutation: true,
    fields: [
      KField('input.name', 'Name'),
      KField('input.day', 'Day'),
      KField('input.time', 'Time', hint: 'ISO 8601, ví dụ 2026-10-22T09:00:00Z'),
      KField('input.coach', 'Coach'),
      KField('input.description', 'Description', kind: FieldKind.multiline),
    ],
    run: (api, v) async => adminPlain(await api.identified.gymCreateClass(input: GymClassInput(name_: v.str('input.name'), day: v.str('input.day'), time: v.str('input.time'), coach: v.str('input.coach'), description: v.str('input.description')))),
  ),
  AdminOp(
    'gymUpdateClass',
    resource: 'class',
    verb: 'update',
    mutation: true,
    fields: [
      KField('class_id', 'Class ID', required: true),
      KField('input.name', 'Name'),
      KField('input.day', 'Day'),
      KField('input.time', 'Time', hint: 'ISO 8601, ví dụ 2026-10-22T09:00:00Z'),
      KField('input.coach', 'Coach'),
      KField('input.description', 'Description', kind: FieldKind.multiline),
    ],
    run: (api, v) async => adminPlain(await api.identified.gymUpdateClass(classId: v.need(v.str('class_id'), 'class_id'), input: GymClassInput(name_: v.str('input.name'), day: v.str('input.day'), time: v.str('input.time'), coach: v.str('input.coach'), description: v.str('input.description')))),
  ),
  AdminOp(
    'gymDeleteClass',
    resource: 'class',
    verb: 'delete',
    mutation: true,
    destructive: true,
    fields: [
      KField('class_id', 'Class ID', required: true),
    ],
    run: (api, v) async => adminPlain(await api.identified.gymDeleteClass(classId: v.need(v.str('class_id'), 'class_id'))),
  ),
];
