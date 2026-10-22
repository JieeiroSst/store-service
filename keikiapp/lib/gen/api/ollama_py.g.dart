// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

import '../gateway_executor.dart';
import '../models.g.dart';

class OllamaPyApi {
  const OllamaPyApi(this._executor);

  final GatewayExecutor _executor;

  Future<OllamaPyTags?> ollamaPyTags({String? selectionSet}) async {
    final data = await _executor.query('query ollamaPyTags { ollamaPyTags ${selectionSet ?? '{ models { name model modified_at size digest details { parent_model format family families parameter_size quantization_level } } }'} }'
    );
    return data['ollamaPyTags'] == null ? null : OllamaPyTags.fromJson((data['ollamaPyTags'] as Map).cast<String, dynamic>());
  }

  Future<Object?> ollamaPyShow({required OllamaPyShowInput input}) async {
    final data = await _executor.query('query ollamaPyShow(\$input: OllamaPyShowInput!) { ollamaPyShow(input: \$input) }', variables: {
      'input': input.toJson(),
    }
    );
    return data['ollamaPyShow'];
  }

  Future<OllamaPyGenerate?> ollamaPyGenerate({required OllamaPyGenerateInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation ollamaPyGenerate(\$input: OllamaPyGenerateInput!) { ollamaPyGenerate(input: \$input) ${selectionSet ?? '{ model created_at response done total_duration eval_count }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['ollamaPyGenerate'] == null ? null : OllamaPyGenerate.fromJson((data['ollamaPyGenerate'] as Map).cast<String, dynamic>());
  }

  Future<OllamaPyChat?> ollamaPyChat({required OllamaPyChatInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation ollamaPyChat(\$input: OllamaPyChatInput!) { ollamaPyChat(input: \$input) ${selectionSet ?? '{ model created_at message { role content } done total_duration eval_count }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['ollamaPyChat'] == null ? null : OllamaPyChat.fromJson((data['ollamaPyChat'] as Map).cast<String, dynamic>());
  }

}
