// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

// ignore_for_file: unused_import

import '../../../apps/kit/kit.dart';
import '../../../../gen/models.g.dart';
import '../admin_op.dart';

final cdnAdminOps = <AdminOp>[
  AdminOp(
    'cdnFiles',
    resource: 'file',
    verb: '',
    fields: [
      KField('limit', 'Limit', kind: FieldKind.number),
      KField('offset', 'Offset', kind: FieldKind.number),
    ],
    run: (api, v) async => adminPlain(await api.cdn.cdnFiles(limit: v.integer('limit'), offset: v.integer('offset'))),
  ),
  AdminOp(
    'cdnDeleteFile',
    resource: 'file',
    verb: 'delete',
    mutation: true,
    destructive: true,
    fields: [
      KField('id', 'ID', required: true),
    ],
    run: (api, v) async => adminPlain(await api.cdn.cdnDeleteFile(id: v.need(v.str('id'), 'id'))),
  ),
];
