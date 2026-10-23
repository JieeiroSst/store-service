// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

// ignore_for_file: unused_import

import '../../../apps/kit/kit.dart';
import '../../../../gen/models.g.dart';
import '../admin_op.dart';

final searchAdminOps = <AdminOp>[
  AdminOp(
    'searchSchemaStatus',
    resource: 'schema',
    verb: '',
    run: (api, v) async => adminPlain(await api.search.searchSchemaStatus()),
  ),
  AdminOp(
    'searchApplySchema',
    resource: 'schema',
    verb: 'apply',
    mutation: true,
    run: (api, v) async => adminPlain(await api.search.searchApplySchema()),
  ),
  AdminOp(
    'searchReindex',
    resource: 'reindex',
    verb: '',
    mutation: true,
    fields: [
      KField('index', 'Index', hint: 'Nhiều giá trị, cách nhau bởi dấu phẩy'),
      KField('unmanaged', 'Unmanaged', kind: FieldKind.toggle),
    ],
    run: (api, v) async => adminPlain(await api.search.searchReindex(index_: v.strs('index'), unmanaged: v.optFlag('unmanaged'))),
  ),
];
