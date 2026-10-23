// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

// ignore_for_file: unused_import

import '../../../apps/kit/kit.dart';
import '../../../../gen/models.g.dart';
import '../admin_op.dart';

final toolAdminOps = <AdminOp>[
  AdminOp(
    'toolJobs',
    resource: 'job',
    verb: '',
    run: (api, v) async => adminPlain(await api.tool.toolJobs()),
  ),
  AdminOp(
    'toolJob',
    resource: 'job',
    verb: '',
    fields: [
      KField('id', 'ID', required: true),
    ],
    run: (api, v) async => adminPlain(await api.tool.toolJob(id: v.need(v.str('id'), 'id'))),
  ),
  AdminOp(
    'toolMobileDevices',
    resource: 'mobile',
    verb: '',
    run: (api, v) async => adminPlain(await api.tool.toolMobileDevices()),
  ),
  AdminOp(
    'toolLearningStats',
    resource: 'learning',
    verb: '',
    run: (api, v) async => adminPlain(await api.tool.toolLearningStats()),
  ),
  AdminOp(
    'toolLearningPending',
    resource: 'learning',
    verb: '',
    run: (api, v) async => adminPlain(await api.tool.toolLearningPending()),
  ),
  AdminOp(
    'toolLearningHistory',
    resource: 'learning',
    verb: '',
    run: (api, v) async => adminPlain(await api.tool.toolLearningHistory()),
  ),
  AdminOp(
    'toolSuite',
    resource: 'suite',
    verb: '',
    fields: [
      KField('name', 'Name', required: true),
    ],
    run: (api, v) async => adminPlain(await api.tool.toolSuite(name_: v.need(v.str('name'), 'name'))),
  ),
  AdminOp(
    'toolGenerateTestCases',
    resource: 'test',
    verb: 'generate',
    mutation: true,
    fields: [
      KField('input', 'Input', kind: FieldKind.multiline, hint: 'JSON, ví dụ {"key": "value"}'),
    ],
    run: (api, v) async => adminPlain(await api.tool.toolGenerateTestCases(input: v.jsonMap('input'))),
  ),
  AdminOp(
    'toolRunTestCases',
    resource: 'test',
    verb: 'run',
    mutation: true,
    fields: [
      KField('input', 'Input', kind: FieldKind.multiline, hint: 'JSON, ví dụ {"key": "value"}'),
    ],
    run: (api, v) async => adminPlain(await api.tool.toolRunTestCases(input: v.jsonMap('input'))),
  ),
  AdminOp(
    'toolRunBenchmark',
    resource: 'general',
    verb: 'run',
    mutation: true,
    fields: [
      KField('input', 'Input', kind: FieldKind.multiline, hint: 'JSON, ví dụ {"key": "value"}'),
    ],
    run: (api, v) async => adminPlain(await api.tool.toolRunBenchmark(input: v.jsonMap('input'))),
  ),
  AdminOp(
    'toolAnalyzeCode',
    resource: 'analyze',
    verb: '',
    mutation: true,
    fields: [
      KField('input', 'Input', kind: FieldKind.multiline, hint: 'JSON, ví dụ {"key": "value"}'),
    ],
    run: (api, v) async => adminPlain(await api.tool.toolAnalyzeCode(input: v.jsonMap('input'))),
  ),
  AdminOp(
    'toolAnalyzeTask',
    resource: 'analyze',
    verb: '',
    mutation: true,
    fields: [
      KField('input', 'Input', kind: FieldKind.multiline, hint: 'JSON, ví dụ {"key": "value"}'),
    ],
    run: (api, v) async => adminPlain(await api.tool.toolAnalyzeTask(input: v.jsonMap('input'))),
  ),
  AdminOp(
    'toolQaScan',
    resource: 'qa',
    verb: '',
    mutation: true,
    fields: [
      KField('input', 'Input', kind: FieldKind.multiline, hint: 'JSON, ví dụ {"key": "value"}'),
    ],
    run: (api, v) async => adminPlain(await api.tool.toolQaScan(input: v.jsonMap('input'))),
  ),
  AdminOp(
    'toolExportXlsx',
    resource: 'general',
    verb: 'export',
    mutation: true,
    fields: [
      KField('input', 'Input', kind: FieldKind.multiline, hint: 'JSON, ví dụ {"key": "value"}'),
    ],
    run: (api, v) async => adminPlain(await api.tool.toolExportXlsx(input: v.jsonMap('input'))),
  ),
  AdminOp(
    'toolQaDatabase',
    resource: 'qa',
    verb: '',
    mutation: true,
    fields: [
      KField('input', 'Input', kind: FieldKind.multiline, hint: 'JSON, ví dụ {"key": "value"}'),
    ],
    run: (api, v) async => adminPlain(await api.tool.toolQaDatabase(input: v.jsonMap('input'))),
  ),
  AdminOp(
    'toolQaAsync',
    resource: 'qa',
    verb: '',
    mutation: true,
    fields: [
      KField('input', 'Input', kind: FieldKind.multiline, hint: 'JSON, ví dụ {"key": "value"}'),
    ],
    run: (api, v) async => adminPlain(await api.tool.toolQaAsync(input: v.jsonMap('input'))),
  ),
  AdminOp(
    'toolQaQueue',
    resource: 'qa',
    verb: '',
    mutation: true,
    fields: [
      KField('input', 'Input', kind: FieldKind.multiline, hint: 'JSON, ví dụ {"key": "value"}'),
    ],
    run: (api, v) async => adminPlain(await api.tool.toolQaQueue(input: v.jsonMap('input'))),
  ),
  AdminOp(
    'toolQaBrowser',
    resource: 'qa',
    verb: '',
    mutation: true,
    fields: [
      KField('input', 'Input', kind: FieldKind.multiline, hint: 'JSON, ví dụ {"key": "value"}'),
    ],
    run: (api, v) async => adminPlain(await api.tool.toolQaBrowser(input: v.jsonMap('input'))),
  ),
  AdminOp(
    'toolQaActive',
    resource: 'qa',
    verb: '',
    mutation: true,
    fields: [
      KField('input', 'Input', kind: FieldKind.multiline, hint: 'JSON, ví dụ {"key": "value"}'),
    ],
    run: (api, v) async => adminPlain(await api.tool.toolQaActive(input: v.jsonMap('input'))),
  ),
  AdminOp(
    'toolJobStress',
    resource: 'job',
    verb: '',
    mutation: true,
    fields: [
      KField('input', 'Input', kind: FieldKind.multiline, hint: 'JSON, ví dụ {"key": "value"}'),
    ],
    run: (api, v) async => adminPlain(await api.tool.toolJobStress(input: v.jsonMap('input'))),
  ),
  AdminOp(
    'toolJobSoak',
    resource: 'job',
    verb: '',
    mutation: true,
    fields: [
      KField('input', 'Input', kind: FieldKind.multiline, hint: 'JSON, ví dụ {"key": "value"}'),
    ],
    run: (api, v) async => adminPlain(await api.tool.toolJobSoak(input: v.jsonMap('input'))),
  ),
  AdminOp(
    'toolMobileStatic',
    resource: 'mobile',
    verb: '',
    mutation: true,
    fields: [
      KField('input', 'Input', kind: FieldKind.multiline, hint: 'JSON, ví dụ {"key": "value"}'),
    ],
    run: (api, v) async => adminPlain(await api.tool.toolMobileStatic(input: v.jsonMap('input'))),
  ),
  AdminOp(
    'toolMobileAndroid',
    resource: 'mobile',
    verb: '',
    mutation: true,
    fields: [
      KField('input', 'Input', kind: FieldKind.multiline, hint: 'JSON, ví dụ {"key": "value"}'),
    ],
    run: (api, v) async => adminPlain(await api.tool.toolMobileAndroid(input: v.jsonMap('input'))),
  ),
  AdminOp(
    'toolMobileIos',
    resource: 'mobile',
    verb: '',
    mutation: true,
    fields: [
      KField('input', 'Input', kind: FieldKind.multiline, hint: 'JSON, ví dụ {"key": "value"}'),
    ],
    run: (api, v) async => adminPlain(await api.tool.toolMobileIos(input: v.jsonMap('input'))),
  ),
  AdminOp(
    'toolQcRun',
    resource: 'general',
    verb: '',
    mutation: true,
    fields: [
      KField('input', 'Input', kind: FieldKind.multiline, hint: 'JSON, ví dụ {"key": "value"}'),
    ],
    run: (api, v) async => adminPlain(await api.tool.toolQcRun(input: v.jsonMap('input'))),
  ),
  AdminOp(
    'toolLearningFeedback',
    resource: 'learning',
    verb: '',
    mutation: true,
    fields: [
      KField('input', 'Input', kind: FieldKind.multiline, hint: 'JSON, ví dụ {"key": "value"}'),
    ],
    run: (api, v) async => adminPlain(await api.tool.toolLearningFeedback(input: v.jsonMap('input'))),
  ),
  AdminOp(
    'toolLearningReflect',
    resource: 'learning',
    verb: '',
    mutation: true,
    fields: [
      KField('input', 'Input', kind: FieldKind.multiline, hint: 'JSON, ví dụ {"key": "value"}'),
    ],
    run: (api, v) async => adminPlain(await api.tool.toolLearningReflect(input: v.jsonMap('input'))),
  ),
  AdminOp(
    'toolLearningEvolve',
    resource: 'learning',
    verb: '',
    mutation: true,
    fields: [
      KField('input', 'Input', kind: FieldKind.multiline, hint: 'JSON, ví dụ {"key": "value"}'),
    ],
    run: (api, v) async => adminPlain(await api.tool.toolLearningEvolve(input: v.jsonMap('input'))),
  ),
  AdminOp(
    'toolCancelJob',
    resource: 'job',
    verb: 'cancel',
    mutation: true,
    destructive: true,
    fields: [
      KField('id', 'ID', required: true),
    ],
    run: (api, v) async => adminPlain(await api.tool.toolCancelJob(id: v.need(v.str('id'), 'id'))),
  ),
  AdminOp(
    'toolApproveCases',
    resource: 'case',
    verb: 'approve',
    mutation: true,
    fields: [
      KField('name', 'Name', required: true),
      KField('input', 'Input', kind: FieldKind.multiline, hint: 'JSON, ví dụ {"key": "value"}'),
    ],
    run: (api, v) async => adminPlain(await api.tool.toolApproveCases(name_: v.need(v.str('name'), 'name'), input: v.jsonMap('input'))),
  ),
  AdminOp(
    'toolRemoveCase',
    resource: 'case',
    verb: 'remove',
    mutation: true,
    destructive: true,
    fields: [
      KField('name', 'Name', required: true),
      KField('hash', 'Hash', required: true),
    ],
    run: (api, v) async => adminPlain(await api.tool.toolRemoveCase(name_: v.need(v.str('name'), 'name'), hash: v.need(v.str('hash'), 'hash'))),
  ),
  AdminOp(
    'toolRunSuite',
    resource: 'suite',
    verb: 'run',
    mutation: true,
    fields: [
      KField('name', 'Name', required: true),
      KField('input', 'Input', kind: FieldKind.multiline, hint: 'JSON, ví dụ {"key": "value"}'),
    ],
    run: (api, v) async => adminPlain(await api.tool.toolRunSuite(name_: v.need(v.str('name'), 'name'), input: v.jsonMap('input'))),
  ),
];
