// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

// ignore_for_file: unused_import

import '../../../apps/kit/kit.dart';
import '../../../../gen/models.g.dart';
import '../admin_op.dart';

final messageAdminOps = <AdminOp>[
  AdminOp(
    'queuePublish',
    resource: 'general',
    verb: 'publish',
    mutation: true,
    fields: [
      KField('topic', 'Topic', required: true),
      KField('payload', 'Payload', kind: FieldKind.multiline, required: true, hint: 'JSON, ví dụ {"key": "value"}'),
    ],
    run: (api, v) async => adminPlain(await api.message.queuePublish(topic: v.need(v.str('topic'), 'topic'), payload: v.need(v.json('payload'), 'payload'))),
  ),
  AdminOp(
    'queueConsume',
    resource: 'consume',
    verb: '',
    mutation: true,
    fields: [
      KField('topic', 'Topic', required: true),
    ],
    run: (api, v) async => adminPlain(await api.message.queueConsume(topic: v.need(v.str('topic'), 'topic'))),
  ),
];
