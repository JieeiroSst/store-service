// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

// ignore_for_file: unused_import

import '../../../apps/kit/kit.dart';
import '../../../../gen/models.g.dart';
import '../admin_op.dart';

final chatAdminOps = <AdminOp>[
  AdminOp(
    'chatMessage',
    resource: 'chat message',
    verb: '',
    fields: [
      KField('id', 'ID', kind: FieldKind.number, required: true),
    ],
    run: (api, v) async => adminPlain(await api.chat.chatMessage(id: v.need(v.integer('id'), 'id'))),
  ),
  AdminOp(
    'chatUserReports',
    resource: 'chat user report',
    verb: '',
    fields: [
      KField('userID', 'User ID', kind: FieldKind.number, required: true),
    ],
    run: (api, v) async => adminPlain(await api.chat.chatUserReports(userID: v.need(v.integer('userID'), 'userID'))),
  ),
];
