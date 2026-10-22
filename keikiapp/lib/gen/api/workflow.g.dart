// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

import '../gateway_executor.dart';

class WorkflowApi {
  const WorkflowApi(this._executor);

  final GatewayExecutor _executor;

  Future<String?> workflowProcessOrder({required String id}) async {
    final data = await _executor.mutation('mutation workflowProcessOrder(\$id: String!) { workflowProcessOrder(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['workflowProcessOrder']?.toString();
  }

  Future<Object?> workflowStartOrderWorkflow({required String id}) async {
    final data = await _executor.mutation('mutation workflowStartOrderWorkflow(\$id: String!) { workflowStartOrderWorkflow(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['workflowStartOrderWorkflow'];
  }

}
