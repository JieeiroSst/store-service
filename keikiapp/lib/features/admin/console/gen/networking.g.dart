// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

// ignore_for_file: unused_import

import '../../../apps/kit/kit.dart';
import '../../../../gen/models.g.dart';
import '../admin_op.dart';

final networkingAdminOps = <AdminOp>[
  AdminOp(
    'netAgentSelf',
    resource: 'agent',
    verb: '',
    fields: [
      KField('query', 'Query', kind: FieldKind.multiline, hint: 'JSON, ví dụ {"key": "value"}'),
    ],
    run: (api, v) async => adminPlain(await api.networking.netAgentSelf(query: v.jsonMap('query'))),
  ),
  AdminOp(
    'netAgentMembers',
    resource: 'agent',
    verb: '',
    fields: [
      KField('query', 'Query', kind: FieldKind.multiline, hint: 'JSON, ví dụ {"key": "value"}'),
    ],
    run: (api, v) async => adminPlain(await api.networking.netAgentMembers(query: v.jsonMap('query'))),
  ),
  AdminOp(
    'netAgentServices',
    resource: 'agent',
    verb: '',
    fields: [
      KField('query', 'Query', kind: FieldKind.multiline, hint: 'JSON, ví dụ {"key": "value"}'),
    ],
    run: (api, v) async => adminPlain(await api.networking.netAgentServices(query: v.jsonMap('query'))),
  ),
  AdminOp(
    'netAgentService',
    resource: 'agent',
    verb: '',
    fields: [
      KField('id', 'ID', required: true),
      KField('query', 'Query', kind: FieldKind.multiline, hint: 'JSON, ví dụ {"key": "value"}'),
    ],
    run: (api, v) async => adminPlain(await api.networking.netAgentService(id: v.need(v.str('id'), 'id'), query: v.jsonMap('query'))),
  ),
  AdminOp(
    'netAgentChecks',
    resource: 'agent',
    verb: '',
    fields: [
      KField('query', 'Query', kind: FieldKind.multiline, hint: 'JSON, ví dụ {"key": "value"}'),
    ],
    run: (api, v) async => adminPlain(await api.networking.netAgentChecks(query: v.jsonMap('query'))),
  ),
  AdminOp(
    'netAgentHealthServiceName',
    resource: 'agent',
    verb: '',
    fields: [
      KField('name', 'Name', required: true),
      KField('query', 'Query', kind: FieldKind.multiline, hint: 'JSON, ví dụ {"key": "value"}'),
    ],
    run: (api, v) async => adminPlain(await api.networking.netAgentHealthServiceName(name_: v.need(v.str('name'), 'name'), query: v.jsonMap('query'))),
  ),
  AdminOp(
    'netCatalogDatacenters',
    resource: 'catalog',
    verb: '',
    fields: [
      KField('query', 'Query', kind: FieldKind.multiline, hint: 'JSON, ví dụ {"key": "value"}'),
    ],
    run: (api, v) async => adminPlain(await api.networking.netCatalogDatacenters(query: v.jsonMap('query'))),
  ),
  AdminOp(
    'netCatalogNodes',
    resource: 'catalog',
    verb: '',
    fields: [
      KField('query', 'Query', kind: FieldKind.multiline, hint: 'JSON, ví dụ {"key": "value"}'),
    ],
    run: (api, v) async => adminPlain(await api.networking.netCatalogNodes(query: v.jsonMap('query'))),
  ),
  AdminOp(
    'netCatalogNode',
    resource: 'catalog',
    verb: '',
    fields: [
      KField('node', 'Node', required: true),
      KField('query', 'Query', kind: FieldKind.multiline, hint: 'JSON, ví dụ {"key": "value"}'),
    ],
    run: (api, v) async => adminPlain(await api.networking.netCatalogNode(node: v.need(v.str('node'), 'node'), query: v.jsonMap('query'))),
  ),
  AdminOp(
    'netCatalogServices',
    resource: 'catalog',
    verb: '',
    fields: [
      KField('query', 'Query', kind: FieldKind.multiline, hint: 'JSON, ví dụ {"key": "value"}'),
    ],
    run: (api, v) async => adminPlain(await api.networking.netCatalogServices(query: v.jsonMap('query'))),
  ),
  AdminOp(
    'netCatalogService',
    resource: 'catalog',
    verb: '',
    fields: [
      KField('service', 'Service', required: true),
      KField('query', 'Query', kind: FieldKind.multiline, hint: 'JSON, ví dụ {"key": "value"}'),
    ],
    run: (api, v) async => adminPlain(await api.networking.netCatalogService(service: v.need(v.str('service'), 'service'), query: v.jsonMap('query'))),
  ),
  AdminOp(
    'netHealthService',
    resource: 'health',
    verb: '',
    fields: [
      KField('service', 'Service', required: true),
      KField('query', 'Query', kind: FieldKind.multiline, hint: 'JSON, ví dụ {"key": "value"}'),
    ],
    run: (api, v) async => adminPlain(await api.networking.netHealthService(service: v.need(v.str('service'), 'service'), query: v.jsonMap('query'))),
  ),
  AdminOp(
    'netHealthChecks',
    resource: 'health',
    verb: '',
    fields: [
      KField('service', 'Service', required: true),
      KField('query', 'Query', kind: FieldKind.multiline, hint: 'JSON, ví dụ {"key": "value"}'),
    ],
    run: (api, v) async => adminPlain(await api.networking.netHealthChecks(service: v.need(v.str('service'), 'service'), query: v.jsonMap('query'))),
  ),
  AdminOp(
    'netHealthNode',
    resource: 'health',
    verb: '',
    fields: [
      KField('node', 'Node', required: true),
      KField('query', 'Query', kind: FieldKind.multiline, hint: 'JSON, ví dụ {"key": "value"}'),
    ],
    run: (api, v) async => adminPlain(await api.networking.netHealthNode(node: v.need(v.str('node'), 'node'), query: v.jsonMap('query'))),
  ),
  AdminOp(
    'netHealthState',
    resource: 'health',
    verb: '',
    fields: [
      KField('state', 'State', required: true),
      KField('query', 'Query', kind: FieldKind.multiline, hint: 'JSON, ví dụ {"key": "value"}'),
    ],
    run: (api, v) async => adminPlain(await api.networking.netHealthState(state: v.need(v.str('state'), 'state'), query: v.jsonMap('query'))),
  ),
  AdminOp(
    'netKv',
    resource: 'kv',
    verb: '',
    fields: [
      KField('key', 'Key', required: true),
      KField('query', 'Query', kind: FieldKind.multiline, hint: 'JSON, ví dụ {"key": "value"}'),
    ],
    run: (api, v) async => adminPlain(await api.networking.netKv(key: v.need(v.str('key'), 'key'), query: v.jsonMap('query'))),
  ),
  AdminOp(
    'netSessionInfo',
    resource: 'session',
    verb: '',
    fields: [
      KField('id', 'ID', required: true),
      KField('query', 'Query', kind: FieldKind.multiline, hint: 'JSON, ví dụ {"key": "value"}'),
    ],
    run: (api, v) async => adminPlain(await api.networking.netSessionInfo(id: v.need(v.str('id'), 'id'), query: v.jsonMap('query'))),
  ),
  AdminOp(
    'netSessionList',
    resource: 'session',
    verb: '',
    fields: [
      KField('query', 'Query', kind: FieldKind.multiline, hint: 'JSON, ví dụ {"key": "value"}'),
    ],
    run: (api, v) async => adminPlain(await api.networking.netSessionList(query: v.jsonMap('query'))),
  ),
  AdminOp(
    'netSessionNode',
    resource: 'session',
    verb: '',
    fields: [
      KField('node', 'Node', required: true),
      KField('query', 'Query', kind: FieldKind.multiline, hint: 'JSON, ví dụ {"key": "value"}'),
    ],
    run: (api, v) async => adminPlain(await api.networking.netSessionNode(node: v.need(v.str('node'), 'node'), query: v.jsonMap('query'))),
  ),
  AdminOp(
    'netConnectIntentions',
    resource: 'connect',
    verb: '',
    fields: [
      KField('query', 'Query', kind: FieldKind.multiline, hint: 'JSON, ví dụ {"key": "value"}'),
    ],
    run: (api, v) async => adminPlain(await api.networking.netConnectIntentions(query: v.jsonMap('query'))),
  ),
  AdminOp(
    'netConnectIntentionsExact',
    resource: 'connect',
    verb: '',
    fields: [
      KField('query', 'Query', kind: FieldKind.multiline, hint: 'JSON, ví dụ {"key": "value"}'),
    ],
    run: (api, v) async => adminPlain(await api.networking.netConnectIntentionsExact(query: v.jsonMap('query'))),
  ),
  AdminOp(
    'netConnectIntentionsCheck',
    resource: 'connect',
    verb: '',
    fields: [
      KField('query', 'Query', kind: FieldKind.multiline, hint: 'JSON, ví dụ {"key": "value"}'),
    ],
    run: (api, v) async => adminPlain(await api.networking.netConnectIntentionsCheck(query: v.jsonMap('query'))),
  ),
  AdminOp(
    'netConnectIntentionsMatch',
    resource: 'connect',
    verb: '',
    fields: [
      KField('query', 'Query', kind: FieldKind.multiline, hint: 'JSON, ví dụ {"key": "value"}'),
    ],
    run: (api, v) async => adminPlain(await api.networking.netConnectIntentionsMatch(query: v.jsonMap('query'))),
  ),
  AdminOp(
    'netConnectIntentions2',
    resource: 'connect',
    verb: '',
    fields: [
      KField('id', 'ID', required: true),
      KField('query', 'Query', kind: FieldKind.multiline, hint: 'JSON, ví dụ {"key": "value"}'),
    ],
    run: (api, v) async => adminPlain(await api.networking.netConnectIntentions2(id: v.need(v.str('id'), 'id'), query: v.jsonMap('query'))),
  ),
  AdminOp(
    'netStatusLeader',
    resource: 'status',
    verb: '',
    fields: [
      KField('query', 'Query', kind: FieldKind.multiline, hint: 'JSON, ví dụ {"key": "value"}'),
    ],
    run: (api, v) async => adminPlain(await api.networking.netStatusLeader(query: v.jsonMap('query'))),
  ),
  AdminOp(
    'netStatusPeers',
    resource: 'status',
    verb: '',
    fields: [
      KField('query', 'Query', kind: FieldKind.multiline, hint: 'JSON, ví dụ {"key": "value"}'),
    ],
    run: (api, v) async => adminPlain(await api.networking.netStatusPeers(query: v.jsonMap('query'))),
  ),
  AdminOp(
    'netInternalUiServices',
    resource: 'general',
    verb: '',
    fields: [
      KField('query', 'Query', kind: FieldKind.multiline, hint: 'JSON, ví dụ {"key": "value"}'),
    ],
    run: (api, v) async => adminPlain(await api.networking.netInternalUiServices(query: v.jsonMap('query'))),
  ),
  AdminOp(
    'netPutAgentServiceRegister',
    resource: 'agent',
    verb: 'update',
    mutation: true,
    fields: [
      KField('query', 'Query', kind: FieldKind.multiline, hint: 'JSON, ví dụ {"key": "value"}'),
      KField('input', 'Input', kind: FieldKind.multiline, hint: 'JSON, ví dụ {"key": "value"}'),
    ],
    run: (api, v) async => adminPlain(await api.networking.netPutAgentServiceRegister(query: v.jsonMap('query'), input: v.jsonMap('input'))),
  ),
  AdminOp(
    'netPutAgentServiceDeregister',
    resource: 'agent',
    verb: 'update',
    mutation: true,
    fields: [
      KField('id', 'ID', required: true),
      KField('query', 'Query', kind: FieldKind.multiline, hint: 'JSON, ví dụ {"key": "value"}'),
    ],
    run: (api, v) async => adminPlain(await api.networking.netPutAgentServiceDeregister(id: v.need(v.str('id'), 'id'), query: v.jsonMap('query'))),
  ),
  AdminOp(
    'netPutAgentServiceMaintenance',
    resource: 'agent',
    verb: 'update',
    mutation: true,
    fields: [
      KField('id', 'ID', required: true),
      KField('query', 'Query', kind: FieldKind.multiline, hint: 'JSON, ví dụ {"key": "value"}'),
    ],
    run: (api, v) async => adminPlain(await api.networking.netPutAgentServiceMaintenance(id: v.need(v.str('id'), 'id'), query: v.jsonMap('query'))),
  ),
  AdminOp(
    'netPutAgentMaintenance',
    resource: 'agent',
    verb: 'update',
    mutation: true,
    fields: [
      KField('query', 'Query', kind: FieldKind.multiline, hint: 'JSON, ví dụ {"key": "value"}'),
    ],
    run: (api, v) async => adminPlain(await api.networking.netPutAgentMaintenance(query: v.jsonMap('query'))),
  ),
  AdminOp(
    'netPutAgentCheckRegister',
    resource: 'agent',
    verb: 'update',
    mutation: true,
    fields: [
      KField('query', 'Query', kind: FieldKind.multiline, hint: 'JSON, ví dụ {"key": "value"}'),
      KField('input', 'Input', kind: FieldKind.multiline, hint: 'JSON, ví dụ {"key": "value"}'),
    ],
    run: (api, v) async => adminPlain(await api.networking.netPutAgentCheckRegister(query: v.jsonMap('query'), input: v.jsonMap('input'))),
  ),
  AdminOp(
    'netPutAgentCheckDeregister',
    resource: 'agent',
    verb: 'update',
    mutation: true,
    fields: [
      KField('id', 'ID', required: true),
      KField('query', 'Query', kind: FieldKind.multiline, hint: 'JSON, ví dụ {"key": "value"}'),
    ],
    run: (api, v) async => adminPlain(await api.networking.netPutAgentCheckDeregister(id: v.need(v.str('id'), 'id'), query: v.jsonMap('query'))),
  ),
  AdminOp(
    'netPutAgentCheckPass',
    resource: 'agent',
    verb: 'update',
    mutation: true,
    fields: [
      KField('id', 'ID', required: true),
      KField('query', 'Query', kind: FieldKind.multiline, hint: 'JSON, ví dụ {"key": "value"}'),
    ],
    run: (api, v) async => adminPlain(await api.networking.netPutAgentCheckPass(id: v.need(v.str('id'), 'id'), query: v.jsonMap('query'))),
  ),
  AdminOp(
    'netPutAgentCheckWarn',
    resource: 'agent',
    verb: 'update',
    mutation: true,
    fields: [
      KField('id', 'ID', required: true),
      KField('query', 'Query', kind: FieldKind.multiline, hint: 'JSON, ví dụ {"key": "value"}'),
    ],
    run: (api, v) async => adminPlain(await api.networking.netPutAgentCheckWarn(id: v.need(v.str('id'), 'id'), query: v.jsonMap('query'))),
  ),
  AdminOp(
    'netPutAgentCheckFail',
    resource: 'agent',
    verb: 'update',
    mutation: true,
    fields: [
      KField('id', 'ID', required: true),
      KField('query', 'Query', kind: FieldKind.multiline, hint: 'JSON, ví dụ {"key": "value"}'),
    ],
    run: (api, v) async => adminPlain(await api.networking.netPutAgentCheckFail(id: v.need(v.str('id'), 'id'), query: v.jsonMap('query'))),
  ),
  AdminOp(
    'netPutAgentCheckUpdate',
    resource: 'agent',
    verb: 'update',
    mutation: true,
    fields: [
      KField('id', 'ID', required: true),
      KField('query', 'Query', kind: FieldKind.multiline, hint: 'JSON, ví dụ {"key": "value"}'),
      KField('input', 'Input', kind: FieldKind.multiline, hint: 'JSON, ví dụ {"key": "value"}'),
    ],
    run: (api, v) async => adminPlain(await api.networking.netPutAgentCheckUpdate(id: v.need(v.str('id'), 'id'), query: v.jsonMap('query'), input: v.jsonMap('input'))),
  ),
  AdminOp(
    'netPutCatalogRegister',
    resource: 'catalog',
    verb: 'update',
    mutation: true,
    fields: [
      KField('query', 'Query', kind: FieldKind.multiline, hint: 'JSON, ví dụ {"key": "value"}'),
      KField('input', 'Input', kind: FieldKind.multiline, hint: 'JSON, ví dụ {"key": "value"}'),
    ],
    run: (api, v) async => adminPlain(await api.networking.netPutCatalogRegister(query: v.jsonMap('query'), input: v.jsonMap('input'))),
  ),
  AdminOp(
    'netPutCatalogDeregister',
    resource: 'catalog',
    verb: 'update',
    mutation: true,
    fields: [
      KField('query', 'Query', kind: FieldKind.multiline, hint: 'JSON, ví dụ {"key": "value"}'),
    ],
    run: (api, v) async => adminPlain(await api.networking.netPutCatalogDeregister(query: v.jsonMap('query'))),
  ),
  AdminOp(
    'netPutKv',
    resource: 'kv',
    verb: 'update',
    mutation: true,
    fields: [
      KField('key', 'Key', required: true),
      KField('value', 'Value', required: true),
      KField('query', 'Query', kind: FieldKind.multiline, hint: 'JSON, ví dụ {"key": "value"}'),
    ],
    run: (api, v) async => adminPlain(await api.networking.netPutKv(key: v.need(v.str('key'), 'key'), value: v.need(v.str('value'), 'value'), query: v.jsonMap('query'))),
  ),
  AdminOp(
    'netDeleteKv',
    resource: 'kv',
    verb: 'delete',
    mutation: true,
    destructive: true,
    fields: [
      KField('key', 'Key', required: true),
      KField('query', 'Query', kind: FieldKind.multiline, hint: 'JSON, ví dụ {"key": "value"}'),
    ],
    run: (api, v) async => adminPlain(await api.networking.netDeleteKv(key: v.need(v.str('key'), 'key'), query: v.jsonMap('query'))),
  ),
  AdminOp(
    'netPutSessionCreate',
    resource: 'session',
    verb: 'update',
    mutation: true,
    fields: [
      KField('query', 'Query', kind: FieldKind.multiline, hint: 'JSON, ví dụ {"key": "value"}'),
      KField('input', 'Input', kind: FieldKind.multiline, hint: 'JSON, ví dụ {"key": "value"}'),
    ],
    run: (api, v) async => adminPlain(await api.networking.netPutSessionCreate(query: v.jsonMap('query'), input: v.jsonMap('input'))),
  ),
  AdminOp(
    'netPutSessionDestroy',
    resource: 'session',
    verb: 'update',
    mutation: true,
    fields: [
      KField('id', 'ID', required: true),
      KField('query', 'Query', kind: FieldKind.multiline, hint: 'JSON, ví dụ {"key": "value"}'),
    ],
    run: (api, v) async => adminPlain(await api.networking.netPutSessionDestroy(id: v.need(v.str('id'), 'id'), query: v.jsonMap('query'))),
  ),
  AdminOp(
    'netPutSessionRenew',
    resource: 'session',
    verb: 'update',
    mutation: true,
    fields: [
      KField('id', 'ID', required: true),
      KField('query', 'Query', kind: FieldKind.multiline, hint: 'JSON, ví dụ {"key": "value"}'),
    ],
    run: (api, v) async => adminPlain(await api.networking.netPutSessionRenew(id: v.need(v.str('id'), 'id'), query: v.jsonMap('query'))),
  ),
  AdminOp(
    'netPostConnectIntentions',
    resource: 'connect',
    verb: 'create',
    mutation: true,
    fields: [
      KField('query', 'Query', kind: FieldKind.multiline, hint: 'JSON, ví dụ {"key": "value"}'),
      KField('input', 'Input', kind: FieldKind.multiline, hint: 'JSON, ví dụ {"key": "value"}'),
    ],
    run: (api, v) async => adminPlain(await api.networking.netPostConnectIntentions(query: v.jsonMap('query'), input: v.jsonMap('input'))),
  ),
  AdminOp(
    'netPutConnectIntentionsExact',
    resource: 'connect',
    verb: 'update',
    mutation: true,
    fields: [
      KField('query', 'Query', kind: FieldKind.multiline, hint: 'JSON, ví dụ {"key": "value"}'),
      KField('input', 'Input', kind: FieldKind.multiline, hint: 'JSON, ví dụ {"key": "value"}'),
    ],
    run: (api, v) async => adminPlain(await api.networking.netPutConnectIntentionsExact(query: v.jsonMap('query'), input: v.jsonMap('input'))),
  ),
  AdminOp(
    'netDeleteConnectIntentionsExact',
    resource: 'connect',
    verb: 'delete',
    mutation: true,
    destructive: true,
    fields: [
      KField('query', 'Query', kind: FieldKind.multiline, hint: 'JSON, ví dụ {"key": "value"}'),
    ],
    run: (api, v) async => adminPlain(await api.networking.netDeleteConnectIntentionsExact(query: v.jsonMap('query'))),
  ),
  AdminOp(
    'netPutConnectIntentions',
    resource: 'connect',
    verb: 'update',
    mutation: true,
    fields: [
      KField('id', 'ID', required: true),
      KField('query', 'Query', kind: FieldKind.multiline, hint: 'JSON, ví dụ {"key": "value"}'),
      KField('input', 'Input', kind: FieldKind.multiline, hint: 'JSON, ví dụ {"key": "value"}'),
    ],
    run: (api, v) async => adminPlain(await api.networking.netPutConnectIntentions(id: v.need(v.str('id'), 'id'), query: v.jsonMap('query'), input: v.jsonMap('input'))),
  ),
  AdminOp(
    'netDeleteConnectIntentions',
    resource: 'connect',
    verb: 'delete',
    mutation: true,
    destructive: true,
    fields: [
      KField('id', 'ID', required: true),
      KField('query', 'Query', kind: FieldKind.multiline, hint: 'JSON, ví dụ {"key": "value"}'),
    ],
    run: (api, v) async => adminPlain(await api.networking.netDeleteConnectIntentions(id: v.need(v.str('id'), 'id'), query: v.jsonMap('query'))),
  ),
];
