// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

// ignore_for_file: unused_import

import '../../../apps/kit/kit.dart';
import '../../../../gen/models.g.dart';
import '../admin_op.dart';

final callCenterAdminOps = <AdminOp>[
  AdminOp(
    'callCenterCalls',
    resource: 'call',
    verb: '',
    fields: [
      KField('skip', 'Skip', kind: FieldKind.number),
      KField('limit', 'Limit', kind: FieldKind.number),
      KField('status', 'Status'),
      KField('from_number', 'From number'),
      KField('outcome', 'Outcome'),
    ],
    run: (api, v) async => adminPlain(await api.callCenter.callCenterCalls(skip: v.integer('skip'), limit: v.integer('limit'), status: v.str('status'), fromNumber: v.str('from_number'), outcome: v.str('outcome'))),
  ),
  AdminOp(
    'callCenterCall',
    resource: 'call',
    verb: '',
    fields: [
      KField('call_id', 'Call ID', kind: FieldKind.number, required: true),
    ],
    run: (api, v) async => adminPlain(await api.callCenter.callCenterCall(callId: v.need(v.integer('call_id'), 'call_id'))),
  ),
  AdminOp(
    'callCenterScenarios',
    resource: 'scenario',
    verb: '',
    fields: [
      KField('skip', 'Skip', kind: FieldKind.number),
      KField('limit', 'Limit', kind: FieldKind.number),
    ],
    run: (api, v) async => adminPlain(await api.callCenter.callCenterScenarios(skip: v.integer('skip'), limit: v.integer('limit'))),
  ),
  AdminOp(
    'callCenterScenario',
    resource: 'scenario',
    verb: '',
    fields: [
      KField('scenario_id', 'Scenario ID', kind: FieldKind.number, required: true),
    ],
    run: (api, v) async => adminPlain(await api.callCenter.callCenterScenario(scenarioId: v.need(v.integer('scenario_id'), 'scenario_id'))),
  ),
  AdminOp(
    'callCenterKnowledgeEntries',
    resource: 'knowledge entry',
    verb: '',
    fields: [
      KField('skip', 'Skip', kind: FieldKind.number),
      KField('limit', 'Limit', kind: FieldKind.number),
      KField('category', 'Category'),
      KField('scenario', 'Scenario'),
    ],
    run: (api, v) async => adminPlain(await api.callCenter.callCenterKnowledgeEntries(skip: v.integer('skip'), limit: v.integer('limit'), category: v.str('category'), scenario: v.str('scenario'))),
  ),
  AdminOp(
    'callCenterKnowledgeEntry',
    resource: 'knowledge entry',
    verb: '',
    fields: [
      KField('entry_id', 'Entry ID', kind: FieldKind.number, required: true),
    ],
    run: (api, v) async => adminPlain(await api.callCenter.callCenterKnowledgeEntry(entryId: v.need(v.integer('entry_id'), 'entry_id'))),
  ),
  AdminOp(
    'callCenterSearchKnowledge',
    resource: 'general',
    verb: 'search',
    fields: [
      KField('q', 'Q', required: true),
      KField('scenario', 'Scenario'),
      KField('limit', 'Limit', kind: FieldKind.number),
    ],
    run: (api, v) async => adminPlain(await api.callCenter.callCenterSearchKnowledge(q: v.need(v.str('q'), 'q'), scenario: v.str('scenario'), limit: v.integer('limit'))),
  ),
  AdminOp(
    'callCenterCustomer',
    resource: 'customer',
    verb: '',
    fields: [
      KField('phone_number', 'Phone number', kind: FieldKind.phone, required: true),
    ],
    run: (api, v) async => adminPlain(await api.callCenter.callCenterCustomer(phoneNumber: v.need(v.str('phone_number'), 'phone_number'))),
  ),
  AdminOp(
    'callCenterStatistics',
    resource: 'general',
    verb: '',
    fields: [
      KField('days', 'Days', kind: FieldKind.number),
    ],
    run: (api, v) async => adminPlain(await api.callCenter.callCenterStatistics(days: v.integer('days'))),
  ),
  AdminOp(
    'callCenterDeleteCall',
    resource: 'call',
    verb: 'delete',
    mutation: true,
    destructive: true,
    fields: [
      KField('call_id', 'Call ID', kind: FieldKind.number, required: true),
    ],
    run: (api, v) async => adminPlain(await api.callCenter.callCenterDeleteCall(callId: v.need(v.integer('call_id'), 'call_id'))),
  ),
  AdminOp(
    'callCenterCreateScenario',
    resource: 'scenario',
    verb: 'create',
    mutation: true,
    fields: [
      KField('input.name', 'Name'),
      KField('input.description', 'Description', kind: FieldKind.multiline),
      KField('input.prompt', 'Prompt', kind: FieldKind.multiline),
      KField('input.greeting', 'Greeting'),
      KField('input.phone_number', 'Phone number', kind: FieldKind.phone),
      KField('input.is_default', 'Is default', kind: FieldKind.number),
      KField('input.is_active', 'Is active', kind: FieldKind.number),
    ],
    run: (api, v) async => adminPlain(await api.callCenter.callCenterCreateScenario(input: CallCenterScenarioInput(name_: v.str('input.name'), description: v.str('input.description'), prompt: v.str('input.prompt'), greeting: v.str('input.greeting'), phoneNumber: v.str('input.phone_number'), isDefault: v.integer('input.is_default'), isActive: v.integer('input.is_active')))),
  ),
  AdminOp(
    'callCenterUpdateScenario',
    resource: 'scenario',
    verb: 'update',
    mutation: true,
    fields: [
      KField('scenario_id', 'Scenario ID', kind: FieldKind.number, required: true),
      KField('input.name', 'Name'),
      KField('input.description', 'Description', kind: FieldKind.multiline),
      KField('input.prompt', 'Prompt', kind: FieldKind.multiline),
      KField('input.greeting', 'Greeting'),
      KField('input.phone_number', 'Phone number', kind: FieldKind.phone),
      KField('input.is_default', 'Is default', kind: FieldKind.number),
      KField('input.is_active', 'Is active', kind: FieldKind.number),
    ],
    run: (api, v) async => adminPlain(await api.callCenter.callCenterUpdateScenario(scenarioId: v.need(v.integer('scenario_id'), 'scenario_id'), input: CallCenterScenarioInput(name_: v.str('input.name'), description: v.str('input.description'), prompt: v.str('input.prompt'), greeting: v.str('input.greeting'), phoneNumber: v.str('input.phone_number'), isDefault: v.integer('input.is_default'), isActive: v.integer('input.is_active')))),
  ),
  AdminOp(
    'callCenterDeleteScenario',
    resource: 'scenario',
    verb: 'delete',
    mutation: true,
    destructive: true,
    fields: [
      KField('scenario_id', 'Scenario ID', kind: FieldKind.number, required: true),
    ],
    run: (api, v) async => adminPlain(await api.callCenter.callCenterDeleteScenario(scenarioId: v.need(v.integer('scenario_id'), 'scenario_id'))),
  ),
  AdminOp(
    'callCenterSeedDefaults',
    resource: 'general',
    verb: '',
    mutation: true,
    run: (api, v) async => adminPlain(await api.callCenter.callCenterSeedDefaults()),
  ),
  AdminOp(
    'callCenterCreateKnowledgeEntry',
    resource: 'knowledge entry',
    verb: 'create',
    mutation: true,
    fields: [
      KField('input.question', 'Question'),
      KField('input.answer', 'Answer'),
      KField('input.keywords', 'Keywords'),
      KField('input.category', 'Category'),
      KField('input.scenario', 'Scenario'),
      KField('input.is_active', 'Is active', kind: FieldKind.number),
    ],
    run: (api, v) async => adminPlain(await api.callCenter.callCenterCreateKnowledgeEntry(input: CallCenterKnowledgeInput(question: v.str('input.question'), answer: v.str('input.answer'), keywords: v.str('input.keywords'), category: v.str('input.category'), scenario: v.str('input.scenario'), isActive: v.integer('input.is_active')))),
  ),
  AdminOp(
    'callCenterUpdateKnowledgeEntry',
    resource: 'knowledge entry',
    verb: 'update',
    mutation: true,
    fields: [
      KField('entry_id', 'Entry ID', kind: FieldKind.number, required: true),
      KField('input.question', 'Question'),
      KField('input.answer', 'Answer'),
      KField('input.keywords', 'Keywords'),
      KField('input.category', 'Category'),
      KField('input.scenario', 'Scenario'),
      KField('input.is_active', 'Is active', kind: FieldKind.number),
    ],
    run: (api, v) async => adminPlain(await api.callCenter.callCenterUpdateKnowledgeEntry(entryId: v.need(v.integer('entry_id'), 'entry_id'), input: CallCenterKnowledgeInput(question: v.str('input.question'), answer: v.str('input.answer'), keywords: v.str('input.keywords'), category: v.str('input.category'), scenario: v.str('input.scenario'), isActive: v.integer('input.is_active')))),
  ),
  AdminOp(
    'callCenterDeleteKnowledgeEntry',
    resource: 'knowledge entry',
    verb: 'delete',
    mutation: true,
    destructive: true,
    fields: [
      KField('entry_id', 'Entry ID', kind: FieldKind.number, required: true),
    ],
    run: (api, v) async => adminPlain(await api.callCenter.callCenterDeleteKnowledgeEntry(entryId: v.need(v.integer('entry_id'), 'entry_id'))),
  ),
  AdminOp(
    'callCenterUpdateCustomer',
    resource: 'customer',
    verb: 'update',
    mutation: true,
    fields: [
      KField('phone_number', 'Phone number', kind: FieldKind.phone, required: true),
      KField('input.name', 'Name'),
      KField('input.email', 'Email', kind: FieldKind.email),
      KField('input.notes', 'Notes', kind: FieldKind.multiline),
    ],
    run: (api, v) async => adminPlain(await api.callCenter.callCenterUpdateCustomer(phoneNumber: v.need(v.str('phone_number'), 'phone_number'), input: CallCenterCustomerInput(name_: v.str('input.name'), email: v.str('input.email'), notes: v.str('input.notes')))),
  ),
  AdminOp(
    'callCenterSimulate',
    resource: 'general',
    verb: '',
    mutation: true,
    fields: [
      KField('input.message', 'Message', kind: FieldKind.multiline),
      KField('input.call_sid', 'Call sid'),
      KField('input.scenario', 'Scenario'),
      KField('input.from_number', 'From number'),
      KField('input.to_number', 'To number'),
    ],
    run: (api, v) async => adminPlain(await api.callCenter.callCenterSimulate(input: CallCenterSimulateInput(message: v.str('input.message'), callSid: v.str('input.call_sid'), scenario: v.str('input.scenario'), fromNumber: v.str('input.from_number'), toNumber: v.str('input.to_number')))),
  ),
];
