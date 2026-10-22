// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

import '../gateway_executor.dart';
import '../models.g.dart';

class CallCenterApi {
  const CallCenterApi(this._executor);

  final GatewayExecutor _executor;

  Future<List<CallCenterCall>?> callCenterCalls({int? skip, int? limit, String? status, String? fromNumber, String? outcome, String? selectionSet}) async {
    final data = await _executor.query('query callCenterCalls(\$skip: Int, \$limit: Int, \$status: String, \$from_number: String, \$outcome: String) { callCenterCalls(skip: \$skip, limit: \$limit, status: \$status, from_number: \$from_number, outcome: \$outcome) ${selectionSet ?? '{ id call_sid from_number to_number notes status duration start_time end_time recording_url scenario outcome created_at updated_at }'} }', variables: {
      'skip': ?skip,
      'limit': ?limit,
      'status': ?status,
      'from_number': ?fromNumber,
      'outcome': ?outcome,
    }
    );
    return (data['callCenterCalls'] as List?)?.map((e) => (e == null ? null : CallCenterCall.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<CallCenterCallDetail?> callCenterCall({required int callId, String? selectionSet}) async {
    final data = await _executor.query('query callCenterCall(\$call_id: Int!) { callCenterCall(call_id: \$call_id) ${selectionSet ?? '{ id call_sid from_number to_number notes status duration start_time end_time recording_url scenario outcome created_at updated_at messages { id call_id role content audio_url timestamp processing_time } }'} }', variables: {
      'call_id': callId,
    }
    );
    return data['callCenterCall'] == null ? null : CallCenterCallDetail.fromJson((data['callCenterCall'] as Map).cast<String, dynamic>());
  }

  Future<List<CallCenterScenario>?> callCenterScenarios({int? skip, int? limit, String? selectionSet}) async {
    final data = await _executor.query('query callCenterScenarios(\$skip: Int, \$limit: Int) { callCenterScenarios(skip: \$skip, limit: \$limit) ${selectionSet ?? '{ id name description prompt greeting phone_number is_default is_active created_at updated_at }'} }', variables: {
      'skip': ?skip,
      'limit': ?limit,
    }
    );
    return (data['callCenterScenarios'] as List?)?.map((e) => (e == null ? null : CallCenterScenario.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<CallCenterScenario?> callCenterScenario({required int scenarioId, String? selectionSet}) async {
    final data = await _executor.query('query callCenterScenario(\$scenario_id: Int!) { callCenterScenario(scenario_id: \$scenario_id) ${selectionSet ?? '{ id name description prompt greeting phone_number is_default is_active created_at updated_at }'} }', variables: {
      'scenario_id': scenarioId,
    }
    );
    return data['callCenterScenario'] == null ? null : CallCenterScenario.fromJson((data['callCenterScenario'] as Map).cast<String, dynamic>());
  }

  Future<List<CallCenterKnowledgeEntry>?> callCenterKnowledgeEntries({int? skip, int? limit, String? category, String? scenario, String? selectionSet}) async {
    final data = await _executor.query('query callCenterKnowledgeEntries(\$skip: Int, \$limit: Int, \$category: String, \$scenario: String) { callCenterKnowledgeEntries(skip: \$skip, limit: \$limit, category: \$category, scenario: \$scenario) ${selectionSet ?? '{ id question answer keywords category scenario is_active created_at updated_at }'} }', variables: {
      'skip': ?skip,
      'limit': ?limit,
      'category': ?category,
      'scenario': ?scenario,
    }
    );
    return (data['callCenterKnowledgeEntries'] as List?)?.map((e) => (e == null ? null : CallCenterKnowledgeEntry.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<CallCenterKnowledgeEntry?> callCenterKnowledgeEntry({required int entryId, String? selectionSet}) async {
    final data = await _executor.query('query callCenterKnowledgeEntry(\$entry_id: Int!) { callCenterKnowledgeEntry(entry_id: \$entry_id) ${selectionSet ?? '{ id question answer keywords category scenario is_active created_at updated_at }'} }', variables: {
      'entry_id': entryId,
    }
    );
    return data['callCenterKnowledgeEntry'] == null ? null : CallCenterKnowledgeEntry.fromJson((data['callCenterKnowledgeEntry'] as Map).cast<String, dynamic>());
  }

  Future<List<CallCenterKnowledgeEntry>?> callCenterSearchKnowledge({required String q, String? scenario, int? limit, String? selectionSet}) async {
    final data = await _executor.query('query callCenterSearchKnowledge(\$q: String!, \$scenario: String, \$limit: Int) { callCenterSearchKnowledge(q: \$q, scenario: \$scenario, limit: \$limit) ${selectionSet ?? '{ id question answer keywords category scenario is_active created_at updated_at }'} }', variables: {
      'q': q,
      'scenario': ?scenario,
      'limit': ?limit,
    }
    );
    return (data['callCenterSearchKnowledge'] as List?)?.map((e) => (e == null ? null : CallCenterKnowledgeEntry.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<CallCenterCustomer?> callCenterCustomer({required String phoneNumber, String? selectionSet}) async {
    final data = await _executor.query('query callCenterCustomer(\$phone_number: String!) { callCenterCustomer(phone_number: \$phone_number) ${selectionSet ?? '{ id phone_number name email notes total_calls last_call_date created_at updated_at }'} }', variables: {
      'phone_number': phoneNumber,
    }
    );
    return data['callCenterCustomer'] == null ? null : CallCenterCustomer.fromJson((data['callCenterCustomer'] as Map).cast<String, dynamic>());
  }

  Future<CallCenterStatistics?> callCenterStatistics({int? days, String? selectionSet}) async {
    final data = await _executor.query('query callCenterStatistics(\$days: Int) { callCenterStatistics(days: \$days) ${selectionSet ?? '{ total_calls completed_calls failed_calls average_duration total_messages resolved_calls transferred_calls callback_requests abandoned_calls }'} }', variables: {
      'days': ?days,
    }
    );
    return data['callCenterStatistics'] == null ? null : CallCenterStatistics.fromJson((data['callCenterStatistics'] as Map).cast<String, dynamic>());
  }

  Future<bool?> callCenterDeleteCall({required int callId}) async {
    final data = await _executor.mutation('mutation callCenterDeleteCall(\$call_id: Int!) { callCenterDeleteCall(call_id: \$call_id) }', variables: {
      'call_id': callId,
    }
    );
    return data['callCenterDeleteCall'] as bool?;
  }

  Future<CallCenterScenario?> callCenterCreateScenario({required CallCenterScenarioInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation callCenterCreateScenario(\$input: CallCenterScenarioInput!) { callCenterCreateScenario(input: \$input) ${selectionSet ?? '{ id name description prompt greeting phone_number is_default is_active created_at updated_at }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['callCenterCreateScenario'] == null ? null : CallCenterScenario.fromJson((data['callCenterCreateScenario'] as Map).cast<String, dynamic>());
  }

  Future<CallCenterScenario?> callCenterUpdateScenario({required int scenarioId, required CallCenterScenarioInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation callCenterUpdateScenario(\$scenario_id: Int!, \$input: CallCenterScenarioInput!) { callCenterUpdateScenario(scenario_id: \$scenario_id, input: \$input) ${selectionSet ?? '{ id name description prompt greeting phone_number is_default is_active created_at updated_at }'} }', variables: {
      'scenario_id': scenarioId,
      'input': input.toJson(),
    }
    );
    return data['callCenterUpdateScenario'] == null ? null : CallCenterScenario.fromJson((data['callCenterUpdateScenario'] as Map).cast<String, dynamic>());
  }

  Future<bool?> callCenterDeleteScenario({required int scenarioId}) async {
    final data = await _executor.mutation('mutation callCenterDeleteScenario(\$scenario_id: Int!) { callCenterDeleteScenario(scenario_id: \$scenario_id) }', variables: {
      'scenario_id': scenarioId,
    }
    );
    return data['callCenterDeleteScenario'] as bool?;
  }

  Future<CallCenterSeedResult?> callCenterSeedDefaults({String? selectionSet}) async {
    final data = await _executor.mutation('mutation callCenterSeedDefaults { callCenterSeedDefaults ${selectionSet ?? '{ scenarios knowledge_entries }'} }'
    );
    return data['callCenterSeedDefaults'] == null ? null : CallCenterSeedResult.fromJson((data['callCenterSeedDefaults'] as Map).cast<String, dynamic>());
  }

  Future<CallCenterKnowledgeEntry?> callCenterCreateKnowledgeEntry({required CallCenterKnowledgeInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation callCenterCreateKnowledgeEntry(\$input: CallCenterKnowledgeInput!) { callCenterCreateKnowledgeEntry(input: \$input) ${selectionSet ?? '{ id question answer keywords category scenario is_active created_at updated_at }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['callCenterCreateKnowledgeEntry'] == null ? null : CallCenterKnowledgeEntry.fromJson((data['callCenterCreateKnowledgeEntry'] as Map).cast<String, dynamic>());
  }

  Future<CallCenterKnowledgeEntry?> callCenterUpdateKnowledgeEntry({required int entryId, required CallCenterKnowledgeInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation callCenterUpdateKnowledgeEntry(\$entry_id: Int!, \$input: CallCenterKnowledgeInput!) { callCenterUpdateKnowledgeEntry(entry_id: \$entry_id, input: \$input) ${selectionSet ?? '{ id question answer keywords category scenario is_active created_at updated_at }'} }', variables: {
      'entry_id': entryId,
      'input': input.toJson(),
    }
    );
    return data['callCenterUpdateKnowledgeEntry'] == null ? null : CallCenterKnowledgeEntry.fromJson((data['callCenterUpdateKnowledgeEntry'] as Map).cast<String, dynamic>());
  }

  Future<bool?> callCenterDeleteKnowledgeEntry({required int entryId}) async {
    final data = await _executor.mutation('mutation callCenterDeleteKnowledgeEntry(\$entry_id: Int!) { callCenterDeleteKnowledgeEntry(entry_id: \$entry_id) }', variables: {
      'entry_id': entryId,
    }
    );
    return data['callCenterDeleteKnowledgeEntry'] as bool?;
  }

  Future<CallCenterCustomer?> callCenterUpdateCustomer({required String phoneNumber, required CallCenterCustomerInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation callCenterUpdateCustomer(\$phone_number: String!, \$input: CallCenterCustomerInput!) { callCenterUpdateCustomer(phone_number: \$phone_number, input: \$input) ${selectionSet ?? '{ id phone_number name email notes total_calls last_call_date created_at updated_at }'} }', variables: {
      'phone_number': phoneNumber,
      'input': input.toJson(),
    }
    );
    return data['callCenterUpdateCustomer'] == null ? null : CallCenterCustomer.fromJson((data['callCenterUpdateCustomer'] as Map).cast<String, dynamic>());
  }

  Future<CallCenterSimulation?> callCenterSimulate({required CallCenterSimulateInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation callCenterSimulate(\$input: CallCenterSimulateInput!) { callCenterSimulate(input: \$input) ${selectionSet ?? '{ call_sid greeting reply listening transfer_to ended knowledge }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['callCenterSimulate'] == null ? null : CallCenterSimulation.fromJson((data['callCenterSimulate'] as Map).cast<String, dynamic>());
  }

}
