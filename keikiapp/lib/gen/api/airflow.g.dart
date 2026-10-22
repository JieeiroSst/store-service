// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

import '../gateway_executor.dart';
import '../models.g.dart';

class AirflowApi {
  const AirflowApi(this._executor);

  final GatewayExecutor _executor;

  Future<AirflowHealth?> airflowHealth({String? selectionSet}) async {
    final data = await _executor.query('query airflowHealth { airflowHealth ${selectionSet ?? '{ metadatabase_status scheduler_status latest_scheduler_heartbeat }'} }'
    );
    return data['airflowHealth'] == null ? null : AirflowHealth.fromJson((data['airflowHealth'] as Map).cast<String, dynamic>());
  }

  Future<AirflowDagList?> airflowDags({int? limit, int? offset, String? selectionSet}) async {
    final data = await _executor.query('query airflowDags(\$limit: Int, \$offset: Int) { airflowDags(limit: \$limit, offset: \$offset) ${selectionSet ?? '{ dags { dag_id is_paused is_active is_subdag owners description fileloc } total_entries }'} }', variables: {
      'limit': ?limit,
      'offset': ?offset,
    }
    );
    return data['airflowDags'] == null ? null : AirflowDagList.fromJson((data['airflowDags'] as Map).cast<String, dynamic>());
  }

  Future<AirflowDag?> airflowDag({required String dagId, String? selectionSet}) async {
    final data = await _executor.query('query airflowDag(\$dagId: String!) { airflowDag(dagId: \$dagId) ${selectionSet ?? '{ dag_id is_paused is_active is_subdag owners description fileloc }'} }', variables: {
      'dagId': dagId,
    }
    );
    return data['airflowDag'] == null ? null : AirflowDag.fromJson((data['airflowDag'] as Map).cast<String, dynamic>());
  }

  Future<AirflowDagRunList?> airflowDagRuns({required String dagId, int? limit, int? offset, String? selectionSet}) async {
    final data = await _executor.query('query airflowDagRuns(\$dagId: String!, \$limit: Int, \$offset: Int) { airflowDagRuns(dagId: \$dagId, limit: \$limit, offset: \$offset) ${selectionSet ?? '{ dag_runs { dag_id dag_run_id logical_date start_date end_date state external_trigger conf note } total_entries }'} }', variables: {
      'dagId': dagId,
      'limit': ?limit,
      'offset': ?offset,
    }
    );
    return data['airflowDagRuns'] == null ? null : AirflowDagRunList.fromJson((data['airflowDagRuns'] as Map).cast<String, dynamic>());
  }

  Future<AirflowDagRun?> airflowDagRun({required String dagId, required String dagRunId, String? selectionSet}) async {
    final data = await _executor.query('query airflowDagRun(\$dagId: String!, \$dagRunId: String!) { airflowDagRun(dagId: \$dagId, dagRunId: \$dagRunId) ${selectionSet ?? '{ dag_id dag_run_id logical_date start_date end_date state external_trigger conf note }'} }', variables: {
      'dagId': dagId,
      'dagRunId': dagRunId,
    }
    );
    return data['airflowDagRun'] == null ? null : AirflowDagRun.fromJson((data['airflowDagRun'] as Map).cast<String, dynamic>());
  }

  Future<AirflowTaskInstanceList?> airflowTaskInstances({required String dagId, required String dagRunId, int? limit, int? offset, String? selectionSet}) async {
    final data = await _executor.query('query airflowTaskInstances(\$dagId: String!, \$dagRunId: String!, \$limit: Int, \$offset: Int) { airflowTaskInstances(dagId: \$dagId, dagRunId: \$dagRunId, limit: \$limit, offset: \$offset) ${selectionSet ?? '{ task_instances { task_id dag_id dag_run_id execution_date start_date end_date duration state try_number map_index pool operator } total_entries }'} }', variables: {
      'dagId': dagId,
      'dagRunId': dagRunId,
      'limit': ?limit,
      'offset': ?offset,
    }
    );
    return data['airflowTaskInstances'] == null ? null : AirflowTaskInstanceList.fromJson((data['airflowTaskInstances'] as Map).cast<String, dynamic>());
  }

  Future<AirflowTaskInstance?> airflowTaskInstance({required String dagId, required String dagRunId, required String taskId, String? selectionSet}) async {
    final data = await _executor.query('query airflowTaskInstance(\$dagId: String!, \$dagRunId: String!, \$taskId: String!) { airflowTaskInstance(dagId: \$dagId, dagRunId: \$dagRunId, taskId: \$taskId) ${selectionSet ?? '{ task_id dag_id dag_run_id execution_date start_date end_date duration state try_number map_index pool operator }'} }', variables: {
      'dagId': dagId,
      'dagRunId': dagRunId,
      'taskId': taskId,
    }
    );
    return data['airflowTaskInstance'] == null ? null : AirflowTaskInstance.fromJson((data['airflowTaskInstance'] as Map).cast<String, dynamic>());
  }

  Future<AirflowVariableList?> airflowVariables({int? limit, int? offset, String? selectionSet}) async {
    final data = await _executor.query('query airflowVariables(\$limit: Int, \$offset: Int) { airflowVariables(limit: \$limit, offset: \$offset) ${selectionSet ?? '{ variables { key value description } total_entries }'} }', variables: {
      'limit': ?limit,
      'offset': ?offset,
    }
    );
    return data['airflowVariables'] == null ? null : AirflowVariableList.fromJson((data['airflowVariables'] as Map).cast<String, dynamic>());
  }

  Future<AirflowVariable?> airflowVariable({required String key, String? selectionSet}) async {
    final data = await _executor.query('query airflowVariable(\$key: String!) { airflowVariable(key: \$key) ${selectionSet ?? '{ key value description }'} }', variables: {
      'key': key,
    }
    );
    return data['airflowVariable'] == null ? null : AirflowVariable.fromJson((data['airflowVariable'] as Map).cast<String, dynamic>());
  }

  Future<AirflowPoolList?> airflowPools({int? limit, int? offset, String? selectionSet}) async {
    final data = await _executor.query('query airflowPools(\$limit: Int, \$offset: Int) { airflowPools(limit: \$limit, offset: \$offset) ${selectionSet ?? '{ pools { name slots occupied_slots used_slots queued_slots open_slots description } total_entries }'} }', variables: {
      'limit': ?limit,
      'offset': ?offset,
    }
    );
    return data['airflowPools'] == null ? null : AirflowPoolList.fromJson((data['airflowPools'] as Map).cast<String, dynamic>());
  }

  Future<AirflowPool?> airflowPool({required String name_, String? selectionSet}) async {
    final data = await _executor.query('query airflowPool(\$name: String!) { airflowPool(name: \$name) ${selectionSet ?? '{ name slots occupied_slots used_slots queued_slots open_slots description }'} }', variables: {
      'name': name_,
    }
    );
    return data['airflowPool'] == null ? null : AirflowPool.fromJson((data['airflowPool'] as Map).cast<String, dynamic>());
  }

  Future<AirflowDag?> airflowSetDagPaused({required String dagId, required bool isPaused, String? selectionSet}) async {
    final data = await _executor.mutation('mutation airflowSetDagPaused(\$dagId: String!, \$is_paused: Boolean!) { airflowSetDagPaused(dagId: \$dagId, is_paused: \$is_paused) ${selectionSet ?? '{ dag_id is_paused is_active is_subdag owners description fileloc }'} }', variables: {
      'dagId': dagId,
      'is_paused': isPaused,
    }
    );
    return data['airflowSetDagPaused'] == null ? null : AirflowDag.fromJson((data['airflowSetDagPaused'] as Map).cast<String, dynamic>());
  }

  Future<AirflowDagRun?> airflowTriggerDagRun({required String dagId, required AirflowTriggerDagRunInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation airflowTriggerDagRun(\$dagId: String!, \$input: AirflowTriggerDagRunInput!) { airflowTriggerDagRun(dagId: \$dagId, input: \$input) ${selectionSet ?? '{ dag_id dag_run_id logical_date start_date end_date state external_trigger conf note }'} }', variables: {
      'dagId': dagId,
      'input': input.toJson(),
    }
    );
    return data['airflowTriggerDagRun'] == null ? null : AirflowDagRun.fromJson((data['airflowTriggerDagRun'] as Map).cast<String, dynamic>());
  }

  Future<bool?> airflowDeleteDagRun({required String dagId, required String dagRunId}) async {
    final data = await _executor.mutation('mutation airflowDeleteDagRun(\$dagId: String!, \$dagRunId: String!) { airflowDeleteDagRun(dagId: \$dagId, dagRunId: \$dagRunId) }', variables: {
      'dagId': dagId,
      'dagRunId': dagRunId,
    }
    );
    return data['airflowDeleteDagRun'] as bool?;
  }

  Future<AirflowVariable?> airflowCreateVariable({required AirflowVariableInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation airflowCreateVariable(\$input: AirflowVariableInput!) { airflowCreateVariable(input: \$input) ${selectionSet ?? '{ key value description }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['airflowCreateVariable'] == null ? null : AirflowVariable.fromJson((data['airflowCreateVariable'] as Map).cast<String, dynamic>());
  }

  Future<AirflowVariable?> airflowUpdateVariable({required String key, required AirflowVariableInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation airflowUpdateVariable(\$key: String!, \$input: AirflowVariableInput!) { airflowUpdateVariable(key: \$key, input: \$input) ${selectionSet ?? '{ key value description }'} }', variables: {
      'key': key,
      'input': input.toJson(),
    }
    );
    return data['airflowUpdateVariable'] == null ? null : AirflowVariable.fromJson((data['airflowUpdateVariable'] as Map).cast<String, dynamic>());
  }

  Future<bool?> airflowDeleteVariable({required String key}) async {
    final data = await _executor.mutation('mutation airflowDeleteVariable(\$key: String!) { airflowDeleteVariable(key: \$key) }', variables: {
      'key': key,
    }
    );
    return data['airflowDeleteVariable'] as bool?;
  }

  Future<AirflowPool?> airflowCreatePool({required AirflowPoolInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation airflowCreatePool(\$input: AirflowPoolInput!) { airflowCreatePool(input: \$input) ${selectionSet ?? '{ name slots occupied_slots used_slots queued_slots open_slots description }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['airflowCreatePool'] == null ? null : AirflowPool.fromJson((data['airflowCreatePool'] as Map).cast<String, dynamic>());
  }

  Future<AirflowPool?> airflowUpdatePool({required String name_, required AirflowPoolInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation airflowUpdatePool(\$name: String!, \$input: AirflowPoolInput!) { airflowUpdatePool(name: \$name, input: \$input) ${selectionSet ?? '{ name slots occupied_slots used_slots queued_slots open_slots description }'} }', variables: {
      'name': name_,
      'input': input.toJson(),
    }
    );
    return data['airflowUpdatePool'] == null ? null : AirflowPool.fromJson((data['airflowUpdatePool'] as Map).cast<String, dynamic>());
  }

  Future<bool?> airflowDeletePool({required String name_}) async {
    final data = await _executor.mutation('mutation airflowDeletePool(\$name: String!) { airflowDeletePool(name: \$name) }', variables: {
      'name': name_,
    }
    );
    return data['airflowDeletePool'] as bool?;
  }

}
