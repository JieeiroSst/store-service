// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

import '../gateway_executor.dart';
import '../models.g.dart';

class WebrtcApi {
  const WebrtcApi(this._executor);

  final GatewayExecutor _executor;

  Future<WebrtcIceServers?> webrtcIceServers({String? userId, String? selectionSet}) async {
    final data = await _executor.query('query webrtcIceServers(\$user_id: String) { webrtcIceServers(user_id: \$user_id) ${selectionSet ?? '{ ice_servers { urls username credential } }'} }', variables: {
      'user_id': ?userId,
    }
    );
    return data['webrtcIceServers'] == null ? null : WebrtcIceServers.fromJson((data['webrtcIceServers'] as Map).cast<String, dynamic>());
  }

  Future<WebrtcRoomInfo?> webrtcRoom({required String roomId, String? selectionSet}) async {
    final data = await _executor.query('query webrtcRoom(\$room_id: String!) { webrtcRoom(room_id: \$room_id) ${selectionSet ?? '{ room_id capacity participants { user_id media { audio video screen } } }'} }', variables: {
      'room_id': roomId,
    }
    );
    return data['webrtcRoom'] == null ? null : WebrtcRoomInfo.fromJson((data['webrtcRoom'] as Map).cast<String, dynamic>());
  }

}
