// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

import '../gateway_executor.dart';
import '../models.g.dart';

class ShareApi {
  const ShareApi(this._executor);

  final GatewayExecutor _executor;

  Future<ShareAnswer?> shareAsk({required String content, String? selectionSet}) async {
    final data = await _executor.mutation('mutation shareAsk(\$content: String!) { shareAsk(content: \$content) ${selectionSet ?? '{ result model }'} }', variables: {
      'content': content,
    }
    );
    return data['shareAsk'] == null ? null : ShareAnswer.fromJson((data['shareAsk'] as Map).cast<String, dynamic>());
  }

}
