// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

// ignore_for_file: unused_import

import '../../../apps/kit/kit.dart';
import '../../../../gen/models.g.dart';
import '../admin_op.dart';

final airflowAdminOps = <AdminOp>[
  AdminOp(
    'airflowHealth',
    resource: 'health',
    verb: '',
    run: (api, v) async => adminPlain(await api.airflow.airflowHealth()),
  ),
  AdminOp(
    'airflowDags',
    resource: 'dag',
    verb: '',
    fields: [
      KField('limit', 'Limit', kind: FieldKind.number),
      KField('offset', 'Offset', kind: FieldKind.number),
    ],
    run: (api, v) async => adminPlain(await api.airflow.airflowDags(limit: v.integer('limit'), offset: v.integer('offset'))),
  ),
  AdminOp(
    'airflowDag',
    resource: 'dag',
    verb: '',
    fields: [
      KField('dagId', 'Dag ID', required: true),
    ],
    run: (api, v) async => adminPlain(await api.airflow.airflowDag(dagId: v.need(v.str('dagId'), 'dagId'))),
  ),
  AdminOp(
    'airflowDagRuns',
    resource: 'dag run',
    verb: '',
    fields: [
      KField('dagId', 'Dag ID', required: true),
      KField('limit', 'Limit', kind: FieldKind.number),
      KField('offset', 'Offset', kind: FieldKind.number),
    ],
    run: (api, v) async => adminPlain(await api.airflow.airflowDagRuns(dagId: v.need(v.str('dagId'), 'dagId'), limit: v.integer('limit'), offset: v.integer('offset'))),
  ),
  AdminOp(
    'airflowDagRun',
    resource: 'dag run',
    verb: '',
    fields: [
      KField('dagId', 'Dag ID', required: true),
      KField('dagRunId', 'Dag run ID', required: true),
    ],
    run: (api, v) async => adminPlain(await api.airflow.airflowDagRun(dagId: v.need(v.str('dagId'), 'dagId'), dagRunId: v.need(v.str('dagRunId'), 'dagRunId'))),
  ),
  AdminOp(
    'airflowTaskInstances',
    resource: 'task instance',
    verb: '',
    fields: [
      KField('dagId', 'Dag ID', required: true),
      KField('dagRunId', 'Dag run ID', required: true),
      KField('limit', 'Limit', kind: FieldKind.number),
      KField('offset', 'Offset', kind: FieldKind.number),
    ],
    run: (api, v) async => adminPlain(await api.airflow.airflowTaskInstances(dagId: v.need(v.str('dagId'), 'dagId'), dagRunId: v.need(v.str('dagRunId'), 'dagRunId'), limit: v.integer('limit'), offset: v.integer('offset'))),
  ),
  AdminOp(
    'airflowTaskInstance',
    resource: 'task instance',
    verb: '',
    fields: [
      KField('dagId', 'Dag ID', required: true),
      KField('dagRunId', 'Dag run ID', required: true),
      KField('taskId', 'Task ID', required: true),
    ],
    run: (api, v) async => adminPlain(await api.airflow.airflowTaskInstance(dagId: v.need(v.str('dagId'), 'dagId'), dagRunId: v.need(v.str('dagRunId'), 'dagRunId'), taskId: v.need(v.str('taskId'), 'taskId'))),
  ),
  AdminOp(
    'airflowVariables',
    resource: 'variable',
    verb: '',
    fields: [
      KField('limit', 'Limit', kind: FieldKind.number),
      KField('offset', 'Offset', kind: FieldKind.number),
    ],
    run: (api, v) async => adminPlain(await api.airflow.airflowVariables(limit: v.integer('limit'), offset: v.integer('offset'))),
  ),
  AdminOp(
    'airflowVariable',
    resource: 'variable',
    verb: '',
    fields: [
      KField('key', 'Key', required: true),
    ],
    run: (api, v) async => adminPlain(await api.airflow.airflowVariable(key: v.need(v.str('key'), 'key'))),
  ),
  AdminOp(
    'airflowPools',
    resource: 'pool',
    verb: '',
    fields: [
      KField('limit', 'Limit', kind: FieldKind.number),
      KField('offset', 'Offset', kind: FieldKind.number),
    ],
    run: (api, v) async => adminPlain(await api.airflow.airflowPools(limit: v.integer('limit'), offset: v.integer('offset'))),
  ),
  AdminOp(
    'airflowPool',
    resource: 'pool',
    verb: '',
    fields: [
      KField('name', 'Name', required: true),
    ],
    run: (api, v) async => adminPlain(await api.airflow.airflowPool(name_: v.need(v.str('name'), 'name'))),
  ),
  AdminOp(
    'airflowSetDagPaused',
    resource: 'dag',
    verb: 'set',
    mutation: true,
    fields: [
      KField('dagId', 'Dag ID', required: true),
      KField('is_paused', 'Is paused', kind: FieldKind.toggle),
    ],
    run: (api, v) async => adminPlain(await api.airflow.airflowSetDagPaused(dagId: v.need(v.str('dagId'), 'dagId'), isPaused: v.flag('is_paused'))),
  ),
  AdminOp(
    'airflowTriggerDagRun',
    resource: 'dag run',
    verb: 'trigger',
    mutation: true,
    fields: [
      KField('dagId', 'Dag ID', required: true),
      KField('input.dag_run_id', 'Dag run ID'),
      KField('input.logical_date', 'Logical date', hint: 'ISO 8601, ví dụ 2026-10-22T09:00:00Z'),
      KField('input.conf', 'Conf', kind: FieldKind.multiline, hint: 'JSON, ví dụ {"key": "value"}'),
      KField('input.note', 'Note', kind: FieldKind.multiline),
    ],
    run: (api, v) async => adminPlain(await api.airflow.airflowTriggerDagRun(dagId: v.need(v.str('dagId'), 'dagId'), input: AirflowTriggerDagRunInput(dagRunId: v.str('input.dag_run_id'), logicalDate: v.str('input.logical_date'), conf: v.jsonMap('input.conf'), note: v.str('input.note')))),
  ),
  AdminOp(
    'airflowDeleteDagRun',
    resource: 'dag run',
    verb: 'delete',
    mutation: true,
    destructive: true,
    fields: [
      KField('dagId', 'Dag ID', required: true),
      KField('dagRunId', 'Dag run ID', required: true),
    ],
    run: (api, v) async => adminPlain(await api.airflow.airflowDeleteDagRun(dagId: v.need(v.str('dagId'), 'dagId'), dagRunId: v.need(v.str('dagRunId'), 'dagRunId'))),
  ),
  AdminOp(
    'airflowCreateVariable',
    resource: 'variable',
    verb: 'create',
    mutation: true,
    fields: [
      KField('input.key', 'Key'),
      KField('input.value', 'Value'),
      KField('input.description', 'Description', kind: FieldKind.multiline),
    ],
    run: (api, v) async => adminPlain(await api.airflow.airflowCreateVariable(input: AirflowVariableInput(key: v.str('input.key'), value: v.str('input.value'), description: v.str('input.description')))),
  ),
  AdminOp(
    'airflowUpdateVariable',
    resource: 'variable',
    verb: 'update',
    mutation: true,
    fields: [
      KField('key', 'Key', required: true),
      KField('input.key', 'Key'),
      KField('input.value', 'Value'),
      KField('input.description', 'Description', kind: FieldKind.multiline),
    ],
    run: (api, v) async => adminPlain(await api.airflow.airflowUpdateVariable(key: v.need(v.str('key'), 'key'), input: AirflowVariableInput(key: v.str('input.key'), value: v.str('input.value'), description: v.str('input.description')))),
  ),
  AdminOp(
    'airflowDeleteVariable',
    resource: 'variable',
    verb: 'delete',
    mutation: true,
    destructive: true,
    fields: [
      KField('key', 'Key', required: true),
    ],
    run: (api, v) async => adminPlain(await api.airflow.airflowDeleteVariable(key: v.need(v.str('key'), 'key'))),
  ),
  AdminOp(
    'airflowCreatePool',
    resource: 'pool',
    verb: 'create',
    mutation: true,
    fields: [
      KField('input.name', 'Name'),
      KField('input.slots', 'Slots', kind: FieldKind.number),
      KField('input.description', 'Description', kind: FieldKind.multiline),
    ],
    run: (api, v) async => adminPlain(await api.airflow.airflowCreatePool(input: AirflowPoolInput(name_: v.str('input.name'), slots: v.integer('input.slots'), description: v.str('input.description')))),
  ),
  AdminOp(
    'airflowUpdatePool',
    resource: 'pool',
    verb: 'update',
    mutation: true,
    fields: [
      KField('name', 'Name', required: true),
      KField('input.name', 'Name'),
      KField('input.slots', 'Slots', kind: FieldKind.number),
      KField('input.description', 'Description', kind: FieldKind.multiline),
    ],
    run: (api, v) async => adminPlain(await api.airflow.airflowUpdatePool(name_: v.need(v.str('name'), 'name'), input: AirflowPoolInput(name_: v.str('input.name'), slots: v.integer('input.slots'), description: v.str('input.description')))),
  ),
  AdminOp(
    'airflowDeletePool',
    resource: 'pool',
    verb: 'delete',
    mutation: true,
    destructive: true,
    fields: [
      KField('name', 'Name', required: true),
    ],
    run: (api, v) async => adminPlain(await api.airflow.airflowDeletePool(name_: v.need(v.str('name'), 'name'))),
  ),
];
