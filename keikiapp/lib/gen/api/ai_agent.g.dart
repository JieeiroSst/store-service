// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

import '../gateway_executor.dart';
import '../models.g.dart';

class AiAgentApi {
  const AiAgentApi(this._executor);

  final GatewayExecutor _executor;

  Future<AiAgentHistory?> aiAgentChatHistory({int? limit, String? selectionSet}) async {
    final data = await _executor.query('query aiAgentChatHistory(\$limit: Int) { aiAgentChatHistory(limit: \$limit) ${selectionSet ?? '{ messages { role content source created_at } }'} }', variables: {
      'limit': ?limit,
    }
    );
    return data['aiAgentChatHistory'] == null ? null : AiAgentHistory.fromJson((data['aiAgentChatHistory'] as Map).cast<String, dynamic>());
  }

  Future<AiAgentTags?> aiAgentModels({String? selectionSet}) async {
    final data = await _executor.query('query aiAgentModels { aiAgentModels ${selectionSet ?? '{ models { name model modified_at size digest details { parent_model format family families parameter_size quantization_level } } }'} }'
    );
    return data['aiAgentModels'] == null ? null : AiAgentTags.fromJson((data['aiAgentModels'] as Map).cast<String, dynamic>());
  }

  Future<AiAgentShowResponse?> aiAgentShowModel({required AiAgentShowInput input, String? selectionSet}) async {
    final data = await _executor.query('query aiAgentShowModel(\$input: AiAgentShowInput!) { aiAgentShowModel(input: \$input) ${selectionSet ?? '{ modelfile parameters template details { parent_model format family families parameter_size quantization_level } model_info capabilities }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['aiAgentShowModel'] == null ? null : AiAgentShowResponse.fromJson((data['aiAgentShowModel'] as Map).cast<String, dynamic>());
  }

  Future<AiAgentGenerateResponse?> aiAgentGenerate({required AiAgentGenerateInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation aiAgentGenerate(\$input: AiAgentGenerateInput!) { aiAgentGenerate(input: \$input) ${selectionSet ?? '{ model created_at response done total_duration load_duration prompt_eval_count prompt_eval_duration eval_count eval_duration }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['aiAgentGenerate'] == null ? null : AiAgentGenerateResponse.fromJson((data['aiAgentGenerate'] as Map).cast<String, dynamic>());
  }

  Future<AiAgentChatResponse?> aiAgentChat({required AiAgentChatInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation aiAgentChat(\$input: AiAgentChatInput!) { aiAgentChat(input: \$input) ${selectionSet ?? '{ model created_at message { role content } done total_duration load_duration prompt_eval_count prompt_eval_duration eval_count eval_duration }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['aiAgentChat'] == null ? null : AiAgentChatResponse.fromJson((data['aiAgentChat'] as Map).cast<String, dynamic>());
  }

}
