// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

// ignore_for_file: unused_import

import '../../../apps/kit/kit.dart';
import '../../../../gen/models.g.dart';
import '../admin_op.dart';

final kmsAdminOps = <AdminOp>[
  AdminOp(
    'kmsKeys',
    resource: 'key',
    verb: '',
    run: (api, v) async => adminPlain(await api.kms.kmsKeys()),
  ),
  AdminOp(
    'kmsKey',
    resource: 'key',
    verb: '',
    fields: [
      KField('id', 'ID', required: true),
    ],
    run: (api, v) async => adminPlain(await api.kms.kmsKey(id: v.need(v.str('id'), 'id'))),
  ),
  AdminOp(
    'kmsKeyForUse',
    resource: 'key',
    verb: '',
    fields: [
      KField('id', 'ID', required: true),
    ],
    run: (api, v) async => adminPlain(await api.kms.kmsKeyForUse(id: v.need(v.str('id'), 'id'))),
  ),
  AdminOp(
    'kmsKeyUsageStats',
    resource: 'key',
    verb: '',
    fields: [
      KField('id', 'ID', required: true),
    ],
    run: (api, v) async => adminPlain(await api.kms.kmsKeyUsageStats(id: v.need(v.str('id'), 'id'))),
  ),
  AdminOp(
    'kmsKeyAuditLogs',
    resource: 'key',
    verb: '',
    fields: [
      KField('id', 'ID', required: true),
      KField('limit', 'Limit', kind: FieldKind.number),
      KField('offset', 'Offset', kind: FieldKind.number),
    ],
    run: (api, v) async => adminPlain(await api.kms.kmsKeyAuditLogs(id: v.need(v.str('id'), 'id'), limit: v.integer('limit'), offset: v.integer('offset'))),
  ),
  AdminOp(
    'kmsAuditLogs',
    resource: 'audit log',
    verb: '',
    fields: [
      KField('limit', 'Limit', kind: FieldKind.number),
      KField('offset', 'Offset', kind: FieldKind.number),
    ],
    run: (api, v) async => adminPlain(await api.kms.kmsAuditLogs(limit: v.integer('limit'), offset: v.integer('offset'))),
  ),
  AdminOp(
    'kmsCreateKey',
    resource: 'key',
    verb: 'create',
    mutation: true,
    fields: [
      KField('input.alias', 'Alias'),
      KField('input.algorithm', 'Algorithm'),
      KField('input.expires_in', 'Expires in', kind: FieldKind.number),
      KField('input.tags', 'Tags', kind: FieldKind.multiline, hint: 'JSON, ví dụ {"key": "value"}'),
    ],
    run: (api, v) async => adminPlain(await api.kms.kmsCreateKey(input: KmsCreateKeyInput(alias: v.str('input.alias'), algorithm: v.str('input.algorithm'), expiresIn: v.integer('input.expires_in'), tags: v.jsonMap('input.tags')))),
  ),
  AdminOp(
    'kmsRotateKey',
    resource: 'key',
    verb: 'rotate',
    mutation: true,
    destructive: true,
    fields: [
      KField('id', 'ID', required: true),
      KField('force', 'Force', kind: FieldKind.toggle),
    ],
    run: (api, v) async => adminPlain(await api.kms.kmsRotateKey(id: v.need(v.str('id'), 'id'), force: v.optFlag('force'))),
  ),
  AdminOp(
    'kmsDeleteKey',
    resource: 'key',
    verb: 'delete',
    mutation: true,
    destructive: true,
    fields: [
      KField('id', 'ID', required: true),
    ],
    run: (api, v) async => adminPlain(await api.kms.kmsDeleteKey(id: v.need(v.str('id'), 'id'))),
  ),
];
