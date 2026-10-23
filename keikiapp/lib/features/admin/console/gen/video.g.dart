// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

// ignore_for_file: unused_import

import '../../../apps/kit/kit.dart';
import '../../../../gen/models.g.dart';
import '../admin_op.dart';

final videoAdminOps = <AdminOp>[
  AdminOp(
    'videoUpload',
    resource: 'upload',
    verb: '',
    mutation: true,
    fields: [
      KField('file', 'File', kind: FieldKind.file, required: true),
      KField('title', 'Title'),
      KField('description', 'Description', kind: FieldKind.multiline),
    ],
    run: (api, v) async => adminPlain(await api.video.videoUpload(file: v.need(v.file('file'), 'file'), title: v.str('title'), description: v.str('description'))),
  ),
  AdminOp(
    'videoDelete',
    resource: 'general',
    verb: 'delete',
    mutation: true,
    destructive: true,
    fields: [
      KField('id', 'ID', required: true),
    ],
    run: (api, v) async => adminPlain(await api.video.videoDelete(id: v.need(v.str('id'), 'id'))),
  ),
];
