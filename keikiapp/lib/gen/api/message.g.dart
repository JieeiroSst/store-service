// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

import '../gateway_executor.dart';

class MessageApi {
  const MessageApi(this._executor);

  final GatewayExecutor _executor;

  Future<String?> queuePublish({required String topic, required Object payload}) async {
    final data = await _executor.mutation('mutation queuePublish(\$topic: String!, \$payload: Any!) { queuePublish(topic: \$topic, payload: \$payload) }', variables: {
      'topic': topic,
      'payload': payload,
    }
    );
    return data['queuePublish']?.toString();
  }

  Future<String?> queueConsume({required String topic}) async {
    final data = await _executor.mutation('mutation queueConsume(\$topic: String!) { queueConsume(topic: \$topic) }', variables: {
      'topic': topic,
    }
    );
    return data['queueConsume']?.toString();
  }

}
