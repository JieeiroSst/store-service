// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

import '../gateway_executor.dart';

class NetworkingApi {
  const NetworkingApi(this._executor);

  final GatewayExecutor _executor;

  Future<Object?> netAgentSelf({Map<String, dynamic>? query}) async {
    final data = await _executor.query('query netAgentSelf(\$query: Map) { netAgentSelf(query: \$query) }', variables: {
      'query': ?query,
    }
    );
    return data['netAgentSelf'];
  }

  Future<Object?> netAgentMembers({Map<String, dynamic>? query}) async {
    final data = await _executor.query('query netAgentMembers(\$query: Map) { netAgentMembers(query: \$query) }', variables: {
      'query': ?query,
    }
    );
    return data['netAgentMembers'];
  }

  Future<Object?> netAgentServices({Map<String, dynamic>? query}) async {
    final data = await _executor.query('query netAgentServices(\$query: Map) { netAgentServices(query: \$query) }', variables: {
      'query': ?query,
    }
    );
    return data['netAgentServices'];
  }

  Future<Object?> netAgentService({required String id, Map<String, dynamic>? query}) async {
    final data = await _executor.query('query netAgentService(\$id: String!, \$query: Map) { netAgentService(id: \$id, query: \$query) }', variables: {
      'id': id,
      'query': ?query,
    }
    );
    return data['netAgentService'];
  }

  Future<Object?> netAgentChecks({Map<String, dynamic>? query}) async {
    final data = await _executor.query('query netAgentChecks(\$query: Map) { netAgentChecks(query: \$query) }', variables: {
      'query': ?query,
    }
    );
    return data['netAgentChecks'];
  }

  Future<Object?> netAgentHealthServiceName({required String name_, Map<String, dynamic>? query}) async {
    final data = await _executor.query('query netAgentHealthServiceName(\$name: String!, \$query: Map) { netAgentHealthServiceName(name: \$name, query: \$query) }', variables: {
      'name': name_,
      'query': ?query,
    }
    );
    return data['netAgentHealthServiceName'];
  }

  Future<Object?> netCatalogDatacenters({Map<String, dynamic>? query}) async {
    final data = await _executor.query('query netCatalogDatacenters(\$query: Map) { netCatalogDatacenters(query: \$query) }', variables: {
      'query': ?query,
    }
    );
    return data['netCatalogDatacenters'];
  }

  Future<Object?> netCatalogNodes({Map<String, dynamic>? query}) async {
    final data = await _executor.query('query netCatalogNodes(\$query: Map) { netCatalogNodes(query: \$query) }', variables: {
      'query': ?query,
    }
    );
    return data['netCatalogNodes'];
  }

  Future<Object?> netCatalogNode({required String node, Map<String, dynamic>? query}) async {
    final data = await _executor.query('query netCatalogNode(\$node: String!, \$query: Map) { netCatalogNode(node: \$node, query: \$query) }', variables: {
      'node': node,
      'query': ?query,
    }
    );
    return data['netCatalogNode'];
  }

  Future<Object?> netCatalogServices({Map<String, dynamic>? query}) async {
    final data = await _executor.query('query netCatalogServices(\$query: Map) { netCatalogServices(query: \$query) }', variables: {
      'query': ?query,
    }
    );
    return data['netCatalogServices'];
  }

  Future<Object?> netCatalogService({required String service, Map<String, dynamic>? query}) async {
    final data = await _executor.query('query netCatalogService(\$service: String!, \$query: Map) { netCatalogService(service: \$service, query: \$query) }', variables: {
      'service': service,
      'query': ?query,
    }
    );
    return data['netCatalogService'];
  }

  Future<Object?> netHealthService({required String service, Map<String, dynamic>? query}) async {
    final data = await _executor.query('query netHealthService(\$service: String!, \$query: Map) { netHealthService(service: \$service, query: \$query) }', variables: {
      'service': service,
      'query': ?query,
    }
    );
    return data['netHealthService'];
  }

  Future<Object?> netHealthChecks({required String service, Map<String, dynamic>? query}) async {
    final data = await _executor.query('query netHealthChecks(\$service: String!, \$query: Map) { netHealthChecks(service: \$service, query: \$query) }', variables: {
      'service': service,
      'query': ?query,
    }
    );
    return data['netHealthChecks'];
  }

  Future<Object?> netHealthNode({required String node, Map<String, dynamic>? query}) async {
    final data = await _executor.query('query netHealthNode(\$node: String!, \$query: Map) { netHealthNode(node: \$node, query: \$query) }', variables: {
      'node': node,
      'query': ?query,
    }
    );
    return data['netHealthNode'];
  }

  Future<Object?> netHealthState({required String state, Map<String, dynamic>? query}) async {
    final data = await _executor.query('query netHealthState(\$state: String!, \$query: Map) { netHealthState(state: \$state, query: \$query) }', variables: {
      'state': state,
      'query': ?query,
    }
    );
    return data['netHealthState'];
  }

  Future<Object?> netKv({required String key, Map<String, dynamic>? query}) async {
    final data = await _executor.query('query netKv(\$key: String!, \$query: Map) { netKv(key: \$key, query: \$query) }', variables: {
      'key': key,
      'query': ?query,
    }
    );
    return data['netKv'];
  }

  Future<Object?> netSessionInfo({required String id, Map<String, dynamic>? query}) async {
    final data = await _executor.query('query netSessionInfo(\$id: String!, \$query: Map) { netSessionInfo(id: \$id, query: \$query) }', variables: {
      'id': id,
      'query': ?query,
    }
    );
    return data['netSessionInfo'];
  }

  Future<Object?> netSessionList({Map<String, dynamic>? query}) async {
    final data = await _executor.query('query netSessionList(\$query: Map) { netSessionList(query: \$query) }', variables: {
      'query': ?query,
    }
    );
    return data['netSessionList'];
  }

  Future<Object?> netSessionNode({required String node, Map<String, dynamic>? query}) async {
    final data = await _executor.query('query netSessionNode(\$node: String!, \$query: Map) { netSessionNode(node: \$node, query: \$query) }', variables: {
      'node': node,
      'query': ?query,
    }
    );
    return data['netSessionNode'];
  }

  Future<Object?> netConnectIntentions({Map<String, dynamic>? query}) async {
    final data = await _executor.query('query netConnectIntentions(\$query: Map) { netConnectIntentions(query: \$query) }', variables: {
      'query': ?query,
    }
    );
    return data['netConnectIntentions'];
  }

  Future<Object?> netConnectIntentionsExact({Map<String, dynamic>? query}) async {
    final data = await _executor.query('query netConnectIntentionsExact(\$query: Map) { netConnectIntentionsExact(query: \$query) }', variables: {
      'query': ?query,
    }
    );
    return data['netConnectIntentionsExact'];
  }

  Future<Object?> netConnectIntentionsCheck({Map<String, dynamic>? query}) async {
    final data = await _executor.query('query netConnectIntentionsCheck(\$query: Map) { netConnectIntentionsCheck(query: \$query) }', variables: {
      'query': ?query,
    }
    );
    return data['netConnectIntentionsCheck'];
  }

  Future<Object?> netConnectIntentionsMatch({Map<String, dynamic>? query}) async {
    final data = await _executor.query('query netConnectIntentionsMatch(\$query: Map) { netConnectIntentionsMatch(query: \$query) }', variables: {
      'query': ?query,
    }
    );
    return data['netConnectIntentionsMatch'];
  }

  Future<Object?> netConnectIntentions2({required String id, Map<String, dynamic>? query}) async {
    final data = await _executor.query('query netConnectIntentions2(\$id: String!, \$query: Map) { netConnectIntentions2(id: \$id, query: \$query) }', variables: {
      'id': id,
      'query': ?query,
    }
    );
    return data['netConnectIntentions2'];
  }

  Future<Object?> netStatusLeader({Map<String, dynamic>? query}) async {
    final data = await _executor.query('query netStatusLeader(\$query: Map) { netStatusLeader(query: \$query) }', variables: {
      'query': ?query,
    }
    );
    return data['netStatusLeader'];
  }

  Future<Object?> netStatusPeers({Map<String, dynamic>? query}) async {
    final data = await _executor.query('query netStatusPeers(\$query: Map) { netStatusPeers(query: \$query) }', variables: {
      'query': ?query,
    }
    );
    return data['netStatusPeers'];
  }

  Future<Object?> netInternalUiServices({Map<String, dynamic>? query}) async {
    final data = await _executor.query('query netInternalUiServices(\$query: Map) { netInternalUiServices(query: \$query) }', variables: {
      'query': ?query,
    }
    );
    return data['netInternalUiServices'];
  }

  Future<Object?> netPutAgentServiceRegister({Map<String, dynamic>? query, Map<String, dynamic>? input}) async {
    final data = await _executor.mutation('mutation netPutAgentServiceRegister(\$query: Map, \$input: Map) { netPutAgentServiceRegister(query: \$query, input: \$input) }', variables: {
      'query': ?query,
      'input': ?input,
    }
    );
    return data['netPutAgentServiceRegister'];
  }

  Future<Object?> netPutAgentServiceDeregister({required String id, Map<String, dynamic>? query}) async {
    final data = await _executor.mutation('mutation netPutAgentServiceDeregister(\$id: String!, \$query: Map) { netPutAgentServiceDeregister(id: \$id, query: \$query) }', variables: {
      'id': id,
      'query': ?query,
    }
    );
    return data['netPutAgentServiceDeregister'];
  }

  Future<Object?> netPutAgentServiceMaintenance({required String id, Map<String, dynamic>? query}) async {
    final data = await _executor.mutation('mutation netPutAgentServiceMaintenance(\$id: String!, \$query: Map) { netPutAgentServiceMaintenance(id: \$id, query: \$query) }', variables: {
      'id': id,
      'query': ?query,
    }
    );
    return data['netPutAgentServiceMaintenance'];
  }

  Future<Object?> netPutAgentMaintenance({Map<String, dynamic>? query}) async {
    final data = await _executor.mutation('mutation netPutAgentMaintenance(\$query: Map) { netPutAgentMaintenance(query: \$query) }', variables: {
      'query': ?query,
    }
    );
    return data['netPutAgentMaintenance'];
  }

  Future<Object?> netPutAgentCheckRegister({Map<String, dynamic>? query, Map<String, dynamic>? input}) async {
    final data = await _executor.mutation('mutation netPutAgentCheckRegister(\$query: Map, \$input: Map) { netPutAgentCheckRegister(query: \$query, input: \$input) }', variables: {
      'query': ?query,
      'input': ?input,
    }
    );
    return data['netPutAgentCheckRegister'];
  }

  Future<Object?> netPutAgentCheckDeregister({required String id, Map<String, dynamic>? query}) async {
    final data = await _executor.mutation('mutation netPutAgentCheckDeregister(\$id: String!, \$query: Map) { netPutAgentCheckDeregister(id: \$id, query: \$query) }', variables: {
      'id': id,
      'query': ?query,
    }
    );
    return data['netPutAgentCheckDeregister'];
  }

  Future<Object?> netPutAgentCheckPass({required String id, Map<String, dynamic>? query}) async {
    final data = await _executor.mutation('mutation netPutAgentCheckPass(\$id: String!, \$query: Map) { netPutAgentCheckPass(id: \$id, query: \$query) }', variables: {
      'id': id,
      'query': ?query,
    }
    );
    return data['netPutAgentCheckPass'];
  }

  Future<Object?> netPutAgentCheckWarn({required String id, Map<String, dynamic>? query}) async {
    final data = await _executor.mutation('mutation netPutAgentCheckWarn(\$id: String!, \$query: Map) { netPutAgentCheckWarn(id: \$id, query: \$query) }', variables: {
      'id': id,
      'query': ?query,
    }
    );
    return data['netPutAgentCheckWarn'];
  }

  Future<Object?> netPutAgentCheckFail({required String id, Map<String, dynamic>? query}) async {
    final data = await _executor.mutation('mutation netPutAgentCheckFail(\$id: String!, \$query: Map) { netPutAgentCheckFail(id: \$id, query: \$query) }', variables: {
      'id': id,
      'query': ?query,
    }
    );
    return data['netPutAgentCheckFail'];
  }

  Future<Object?> netPutAgentCheckUpdate({required String id, Map<String, dynamic>? query, Map<String, dynamic>? input}) async {
    final data = await _executor.mutation('mutation netPutAgentCheckUpdate(\$id: String!, \$query: Map, \$input: Map) { netPutAgentCheckUpdate(id: \$id, query: \$query, input: \$input) }', variables: {
      'id': id,
      'query': ?query,
      'input': ?input,
    }
    );
    return data['netPutAgentCheckUpdate'];
  }

  Future<Object?> netPutCatalogRegister({Map<String, dynamic>? query, Map<String, dynamic>? input}) async {
    final data = await _executor.mutation('mutation netPutCatalogRegister(\$query: Map, \$input: Map) { netPutCatalogRegister(query: \$query, input: \$input) }', variables: {
      'query': ?query,
      'input': ?input,
    }
    );
    return data['netPutCatalogRegister'];
  }

  Future<Object?> netPutCatalogDeregister({Map<String, dynamic>? query}) async {
    final data = await _executor.mutation('mutation netPutCatalogDeregister(\$query: Map) { netPutCatalogDeregister(query: \$query) }', variables: {
      'query': ?query,
    }
    );
    return data['netPutCatalogDeregister'];
  }

  Future<Object?> netPutKv({required String key, required String value, Map<String, dynamic>? query}) async {
    final data = await _executor.mutation('mutation netPutKv(\$key: String!, \$value: String!, \$query: Map) { netPutKv(key: \$key, value: \$value, query: \$query) }', variables: {
      'key': key,
      'value': value,
      'query': ?query,
    }
    );
    return data['netPutKv'];
  }

  Future<Object?> netDeleteKv({required String key, Map<String, dynamic>? query}) async {
    final data = await _executor.mutation('mutation netDeleteKv(\$key: String!, \$query: Map) { netDeleteKv(key: \$key, query: \$query) }', variables: {
      'key': key,
      'query': ?query,
    }
    );
    return data['netDeleteKv'];
  }

  Future<Object?> netPutSessionCreate({Map<String, dynamic>? query, Map<String, dynamic>? input}) async {
    final data = await _executor.mutation('mutation netPutSessionCreate(\$query: Map, \$input: Map) { netPutSessionCreate(query: \$query, input: \$input) }', variables: {
      'query': ?query,
      'input': ?input,
    }
    );
    return data['netPutSessionCreate'];
  }

  Future<Object?> netPutSessionDestroy({required String id, Map<String, dynamic>? query}) async {
    final data = await _executor.mutation('mutation netPutSessionDestroy(\$id: String!, \$query: Map) { netPutSessionDestroy(id: \$id, query: \$query) }', variables: {
      'id': id,
      'query': ?query,
    }
    );
    return data['netPutSessionDestroy'];
  }

  Future<Object?> netPutSessionRenew({required String id, Map<String, dynamic>? query}) async {
    final data = await _executor.mutation('mutation netPutSessionRenew(\$id: String!, \$query: Map) { netPutSessionRenew(id: \$id, query: \$query) }', variables: {
      'id': id,
      'query': ?query,
    }
    );
    return data['netPutSessionRenew'];
  }

  Future<Object?> netPostConnectIntentions({Map<String, dynamic>? query, Map<String, dynamic>? input}) async {
    final data = await _executor.mutation('mutation netPostConnectIntentions(\$query: Map, \$input: Map) { netPostConnectIntentions(query: \$query, input: \$input) }', variables: {
      'query': ?query,
      'input': ?input,
    }
    );
    return data['netPostConnectIntentions'];
  }

  Future<Object?> netPutConnectIntentionsExact({Map<String, dynamic>? query, Map<String, dynamic>? input}) async {
    final data = await _executor.mutation('mutation netPutConnectIntentionsExact(\$query: Map, \$input: Map) { netPutConnectIntentionsExact(query: \$query, input: \$input) }', variables: {
      'query': ?query,
      'input': ?input,
    }
    );
    return data['netPutConnectIntentionsExact'];
  }

  Future<Object?> netDeleteConnectIntentionsExact({Map<String, dynamic>? query}) async {
    final data = await _executor.mutation('mutation netDeleteConnectIntentionsExact(\$query: Map) { netDeleteConnectIntentionsExact(query: \$query) }', variables: {
      'query': ?query,
    }
    );
    return data['netDeleteConnectIntentionsExact'];
  }

  Future<Object?> netPutConnectIntentions({required String id, Map<String, dynamic>? query, Map<String, dynamic>? input}) async {
    final data = await _executor.mutation('mutation netPutConnectIntentions(\$id: String!, \$query: Map, \$input: Map) { netPutConnectIntentions(id: \$id, query: \$query, input: \$input) }', variables: {
      'id': id,
      'query': ?query,
      'input': ?input,
    }
    );
    return data['netPutConnectIntentions'];
  }

  Future<Object?> netDeleteConnectIntentions({required String id, Map<String, dynamic>? query}) async {
    final data = await _executor.mutation('mutation netDeleteConnectIntentions(\$id: String!, \$query: Map) { netDeleteConnectIntentions(id: \$id, query: \$query) }', variables: {
      'id': id,
      'query': ?query,
    }
    );
    return data['netDeleteConnectIntentions'];
  }

}
