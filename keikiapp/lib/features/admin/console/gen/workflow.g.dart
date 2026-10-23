// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

// ignore_for_file: unused_import

import '../../../apps/kit/kit.dart';
import '../../../../gen/models.g.dart';
import '../admin_op.dart';

final workflowAdminOps = <AdminOp>[
  AdminOp(
    'workflowProcessOrder',
    resource: 'order',
    verb: 'process',
    mutation: true,
    fields: [
      KField('id', 'ID', required: true),
    ],
    run: (api, v) async => adminPlain(await api.workflow.workflowProcessOrder(id: v.need(v.str('id'), 'id'))),
  ),
  AdminOp(
    'workflowStartOrderWorkflow',
    resource: 'order',
    verb: 'start',
    mutation: true,
    fields: [
      KField('id', 'ID', required: true),
    ],
    run: (api, v) async => adminPlain(await api.workflow.workflowStartOrderWorkflow(id: v.need(v.str('id'), 'id'))),
  ),
];
