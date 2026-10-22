// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

import '../gateway_executor.dart';

class CommonApi {
  const CommonApi(this._executor);

  final GatewayExecutor _executor;

  Future<String?> health() async {
    final data = await _executor.public('query health { health }'
    );
    return data['health']?.toString();
  }

  Future<String?> ping() async {
    final data = await _executor.public('mutation ping { ping }'
    );
    return data['ping']?.toString();
  }

}
