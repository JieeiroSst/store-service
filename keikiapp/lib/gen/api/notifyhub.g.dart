// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

import '../gateway_executor.dart';
import '../models.g.dart';

class NotifyhubApi {
  const NotifyhubApi(this._executor);

  final GatewayExecutor _executor;

  Future<List<HubChannel>?> hubChannels({String? type_, bool? active, String? selectionSet}) async {
    final data = await _executor.query('query hubChannels(\$type: String, \$active: Boolean) { hubChannels(type: \$type, active: \$active) ${selectionSet ?? '{ id name type is_active created_at updated_at }'} }', variables: {
      'type': ?type_,
      'active': ?active,
    }
    );
    return (data['hubChannels'] as List?)?.map((e) => (e == null ? null : HubChannel.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<HubChannel?> hubChannel({required String id, String? selectionSet}) async {
    final data = await _executor.query('query hubChannel(\$id: String!) { hubChannel(id: \$id) ${selectionSet ?? '{ id name type is_active created_at updated_at }'} }', variables: {
      'id': id,
    }
    );
    return data['hubChannel'] == null ? null : HubChannel.fromJson((data['hubChannel'] as Map).cast<String, dynamic>());
  }

  Future<List<HubDataSource>?> hubDataSources({String? selectionSet}) async {
    final data = await _executor.query('query hubDataSources { hubDataSources ${selectionSet ?? '{ id name url method headers body auth_type json_path is_active created_at updated_at }'} }'
    );
    return (data['hubDataSources'] as List?)?.map((e) => (e == null ? null : HubDataSource.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<HubDataSource?> hubDataSource({required String id, String? selectionSet}) async {
    final data = await _executor.query('query hubDataSource(\$id: String!) { hubDataSource(id: \$id) ${selectionSet ?? '{ id name url method headers body auth_type json_path is_active created_at updated_at }'} }', variables: {
      'id': id,
    }
    );
    return data['hubDataSource'] == null ? null : HubDataSource.fromJson((data['hubDataSource'] as Map).cast<String, dynamic>());
  }

  Future<List<HubTemplate>?> hubTemplates({String? channel, String? selectionSet}) async {
    final data = await _executor.query('query hubTemplates(\$channel: String) { hubTemplates(channel: \$channel) ${selectionSet ?? '{ id name channel subject body variables is_active created_at updated_at }'} }', variables: {
      'channel': ?channel,
    }
    );
    return (data['hubTemplates'] as List?)?.map((e) => (e == null ? null : HubTemplate.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<HubTemplate?> hubTemplate({required String id, String? selectionSet}) async {
    final data = await _executor.query('query hubTemplate(\$id: String!) { hubTemplate(id: \$id) ${selectionSet ?? '{ id name channel subject body variables is_active created_at updated_at }'} }', variables: {
      'id': id,
    }
    );
    return data['hubTemplate'] == null ? null : HubTemplate.fromJson((data['hubTemplate'] as Map).cast<String, dynamic>());
  }

  Future<HubJobPage?> hubJobs({String? status, int? page, int? pageSize, String? selectionSet}) async {
    final data = await _executor.query('query hubJobs(\$status: String, \$page: Int, \$page_size: Int) { hubJobs(status: \$status, page: \$page, page_size: \$page_size) ${selectionSet ?? '{ data { id name description channel_id template_id data_source_id schedule_type cron_expr run_at interval_sec recipients static_payload status last_run_at next_run_at run_count fail_count max_runs created_at updated_at channel { id name type is_active created_at updated_at } template { id name channel subject body variables is_active created_at updated_at } data_source { id name url method headers body auth_type json_path is_active created_at updated_at } } total page page_size ts request_id }'} }', variables: {
      'status': ?status,
      'page': ?page,
      'page_size': ?pageSize,
    }
    );
    return data['hubJobs'] == null ? null : HubJobPage.fromJson((data['hubJobs'] as Map).cast<String, dynamic>());
  }

  Future<HubNotifyJob?> hubJob({required String id, String? selectionSet}) async {
    final data = await _executor.query('query hubJob(\$id: String!) { hubJob(id: \$id) ${selectionSet ?? '{ id name description channel_id template_id data_source_id schedule_type cron_expr run_at interval_sec recipients static_payload status last_run_at next_run_at run_count fail_count max_runs created_at updated_at channel { id name type is_active created_at updated_at } template { id name channel subject body variables is_active created_at updated_at } data_source { id name url method headers body auth_type json_path is_active created_at updated_at } }'} }', variables: {
      'id': id,
    }
    );
    return data['hubJob'] == null ? null : HubNotifyJob.fromJson((data['hubJob'] as Map).cast<String, dynamic>());
  }

  Future<HubHistoryPage?> hubHistory({String? jobId, String? status, int? page, int? pageSize, String? selectionSet}) async {
    final data = await _executor.query('query hubHistory(\$job_id: String, \$status: String, \$page: Int, \$page_size: Int) { hubHistory(job_id: \$job_id, status: \$status, page: \$page, page_size: \$page_size) ${selectionSet ?? '{ data { id job_id channel_type recipient subject body status retry_count error sent_at created_at updated_at } total page page_size ts request_id }'} }', variables: {
      'job_id': ?jobId,
      'status': ?status,
      'page': ?page,
      'page_size': ?pageSize,
    }
    );
    return data['hubHistory'] == null ? null : HubHistoryPage.fromJson((data['hubHistory'] as Map).cast<String, dynamic>());
  }

  Future<HubSchedulerStatus?> hubSchedulerStatus({String? selectionSet}) async {
    final data = await _executor.query('query hubSchedulerStatus { hubSchedulerStatus ${selectionSet ?? '{ scheduled_jobs count ts }'} }'
    );
    return data['hubSchedulerStatus'] == null ? null : HubSchedulerStatus.fromJson((data['hubSchedulerStatus'] as Map).cast<String, dynamic>());
  }

  Future<HubChannel?> hubCreateChannel({required HubChannelInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation hubCreateChannel(\$input: HubChannelInput!) { hubCreateChannel(input: \$input) ${selectionSet ?? '{ id name type is_active created_at updated_at }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['hubCreateChannel'] == null ? null : HubChannel.fromJson((data['hubCreateChannel'] as Map).cast<String, dynamic>());
  }

  Future<HubChannel?> hubUpdateChannel({required String id, required HubChannelInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation hubUpdateChannel(\$id: String!, \$input: HubChannelInput!) { hubUpdateChannel(id: \$id, input: \$input) ${selectionSet ?? '{ id name type is_active created_at updated_at }'} }', variables: {
      'id': id,
      'input': input.toJson(),
    }
    );
    return data['hubUpdateChannel'] == null ? null : HubChannel.fromJson((data['hubUpdateChannel'] as Map).cast<String, dynamic>());
  }

  Future<bool?> hubDeleteChannel({required String id}) async {
    final data = await _executor.mutation('mutation hubDeleteChannel(\$id: String!) { hubDeleteChannel(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['hubDeleteChannel'] as bool?;
  }

  Future<HubDataSource?> hubCreateDataSource({required HubDataSourceInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation hubCreateDataSource(\$input: HubDataSourceInput!) { hubCreateDataSource(input: \$input) ${selectionSet ?? '{ id name url method headers body auth_type json_path is_active created_at updated_at }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['hubCreateDataSource'] == null ? null : HubDataSource.fromJson((data['hubCreateDataSource'] as Map).cast<String, dynamic>());
  }

  Future<HubDataSource?> hubUpdateDataSource({required String id, required HubDataSourceInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation hubUpdateDataSource(\$id: String!, \$input: HubDataSourceInput!) { hubUpdateDataSource(id: \$id, input: \$input) ${selectionSet ?? '{ id name url method headers body auth_type json_path is_active created_at updated_at }'} }', variables: {
      'id': id,
      'input': input.toJson(),
    }
    );
    return data['hubUpdateDataSource'] == null ? null : HubDataSource.fromJson((data['hubUpdateDataSource'] as Map).cast<String, dynamic>());
  }

  Future<bool?> hubDeleteDataSource({required String id}) async {
    final data = await _executor.mutation('mutation hubDeleteDataSource(\$id: String!) { hubDeleteDataSource(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['hubDeleteDataSource'] as bool?;
  }

  Future<HubTemplate?> hubCreateTemplate({required HubTemplateInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation hubCreateTemplate(\$input: HubTemplateInput!) { hubCreateTemplate(input: \$input) ${selectionSet ?? '{ id name channel subject body variables is_active created_at updated_at }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['hubCreateTemplate'] == null ? null : HubTemplate.fromJson((data['hubCreateTemplate'] as Map).cast<String, dynamic>());
  }

  Future<HubTemplate?> hubUpdateTemplate({required String id, required HubTemplateInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation hubUpdateTemplate(\$id: String!, \$input: HubTemplateInput!) { hubUpdateTemplate(id: \$id, input: \$input) ${selectionSet ?? '{ id name channel subject body variables is_active created_at updated_at }'} }', variables: {
      'id': id,
      'input': input.toJson(),
    }
    );
    return data['hubUpdateTemplate'] == null ? null : HubTemplate.fromJson((data['hubUpdateTemplate'] as Map).cast<String, dynamic>());
  }

  Future<bool?> hubDeleteTemplate({required String id}) async {
    final data = await _executor.mutation('mutation hubDeleteTemplate(\$id: String!) { hubDeleteTemplate(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['hubDeleteTemplate'] as bool?;
  }

  Future<HubNotifyJob?> hubCreateJob({required HubNotifyJobInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation hubCreateJob(\$input: HubNotifyJobInput!) { hubCreateJob(input: \$input) ${selectionSet ?? '{ id name description channel_id template_id data_source_id schedule_type cron_expr run_at interval_sec recipients static_payload status last_run_at next_run_at run_count fail_count max_runs created_at updated_at channel { id name type is_active created_at updated_at } template { id name channel subject body variables is_active created_at updated_at } data_source { id name url method headers body auth_type json_path is_active created_at updated_at } }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['hubCreateJob'] == null ? null : HubNotifyJob.fromJson((data['hubCreateJob'] as Map).cast<String, dynamic>());
  }

  Future<HubNotifyJob?> hubUpdateJob({required String id, required HubNotifyJobInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation hubUpdateJob(\$id: String!, \$input: HubNotifyJobInput!) { hubUpdateJob(id: \$id, input: \$input) ${selectionSet ?? '{ id name description channel_id template_id data_source_id schedule_type cron_expr run_at interval_sec recipients static_payload status last_run_at next_run_at run_count fail_count max_runs created_at updated_at channel { id name type is_active created_at updated_at } template { id name channel subject body variables is_active created_at updated_at } data_source { id name url method headers body auth_type json_path is_active created_at updated_at } }'} }', variables: {
      'id': id,
      'input': input.toJson(),
    }
    );
    return data['hubUpdateJob'] == null ? null : HubNotifyJob.fromJson((data['hubUpdateJob'] as Map).cast<String, dynamic>());
  }

  Future<bool?> hubDeleteJob({required String id}) async {
    final data = await _executor.mutation('mutation hubDeleteJob(\$id: String!) { hubDeleteJob(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['hubDeleteJob'] as bool?;
  }

  Future<Object?> hubPauseJob({required String id}) async {
    final data = await _executor.mutation('mutation hubPauseJob(\$id: String!) { hubPauseJob(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['hubPauseJob'];
  }

  Future<Object?> hubResumeJob({required String id}) async {
    final data = await _executor.mutation('mutation hubResumeJob(\$id: String!) { hubResumeJob(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['hubResumeJob'];
  }

  Future<Object?> hubTriggerJob({required String id}) async {
    final data = await _executor.mutation('mutation hubTriggerJob(\$id: String!) { hubTriggerJob(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['hubTriggerJob'];
  }

}
