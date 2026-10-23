// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

// ignore_for_file: unused_import

import '../../../apps/kit/kit.dart';
import '../../../../gen/models.g.dart';
import '../admin_op.dart';

final restaurantConsumerAdminOps = <AdminOp>[
  AdminOp(
    'rConsumerConsumers',
    resource: 'consumer',
    verb: '',
    fields: [
      KField('Limit', 'Limit', kind: FieldKind.number),
      KField('Page', 'Page', kind: FieldKind.number),
      KField('Sort', 'Sort'),
    ],
    run: (api, v) async => adminPlain(await api.restaurantConsumer.rConsumerConsumers(limit: v.integer('Limit'), page: v.integer('Page'), sort: v.str('Sort'))),
  ),
];
