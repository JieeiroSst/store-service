// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

// ignore_for_file: unused_import

import '../../../apps/kit/kit.dart';
import '../../../../gen/models.g.dart';
import '../admin_op.dart';

final livestreamAdminOps = <AdminOp>[
  AdminOp(
    'liveAdminDeleteRoom',
    resource: 'delete room',
    verb: '',
    mutation: true,
    fields: [
      KField('id', 'ID', required: true),
    ],
    run: (api, v) async => adminPlain(await api.livestream.liveAdminDeleteRoom(id: v.need(v.str('id'), 'id'))),
  ),
];
