// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

import '../gateway_executor.dart';
import '../models.g.dart';

class LivestreamApi {
  const LivestreamApi(this._executor);

  final GatewayExecutor _executor;

  Future<List<LiveRoom>?> liveRooms({bool? live, String? selectionSet}) async {
    final data = await _executor.public('query liveRooms(\$live: Boolean) { liveRooms(live: \$live) ${selectionSet ?? '{ id ownerUserId slug title description streamKey status createdAt updatedAt }'} }', variables: {
      'live': ?live,
    }
    );
    return (data['liveRooms'] as List?)?.map((e) => (e == null ? null : LiveRoom.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<LiveRoom?> liveRoom({required String id, String? selectionSet}) async {
    final data = await _executor.public('query liveRoom(\$id: String!) { liveRoom(id: \$id) ${selectionSet ?? '{ id ownerUserId slug title description streamKey status createdAt updatedAt }'} }', variables: {
      'id': id,
    }
    );
    return data['liveRoom'] == null ? null : LiveRoom.fromJson((data['liveRoom'] as Map).cast<String, dynamic>());
  }

  Future<LiveStream?> liveActiveStream({required String id, String? selectionSet}) async {
    final data = await _executor.public('query liveActiveStream(\$id: String!) { liveActiveStream(id: \$id) ${selectionSet ?? '{ id roomId nodeId status startedAt endedAt peakViewer createdAt }'} }', variables: {
      'id': id,
    }
    );
    return data['liveActiveStream'] == null ? null : LiveStream.fromJson((data['liveActiveStream'] as Map).cast<String, dynamic>());
  }

  Future<List<LiveRecording>?> liveRecordings({required String id, String? selectionSet}) async {
    final data = await _executor.public('query liveRecordings(\$id: String!) { liveRecordings(id: \$id) ${selectionSet ?? '{ id streamId roomId objectKey durationSeconds sizeBytes createdAt }'} }', variables: {
      'id': id,
    }
    );
    return (data['liveRecordings'] as List?)?.map((e) => (e == null ? null : LiveRecording.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<LiveViewerCount?> liveViewerCount({required String id, String? selectionSet}) async {
    final data = await _executor.public('query liveViewerCount(\$id: String!) { liveViewerCount(id: \$id) ${selectionSet ?? '{ viewers }'} }', variables: {
      'id': id,
    }
    );
    return data['liveViewerCount'] == null ? null : LiveViewerCount.fromJson((data['liveViewerCount'] as Map).cast<String, dynamic>());
  }

  Future<LivePlaybackInfo?> livePlayback({required String id, String? selectionSet}) async {
    final data = await _executor.public('query livePlayback(\$id: String!) { livePlayback(id: \$id) ${selectionSet ?? '{ URL ExpiresAt IsLive }'} }', variables: {
      'id': id,
    }
    );
    return data['livePlayback'] == null ? null : LivePlaybackInfo.fromJson((data['livePlayback'] as Map).cast<String, dynamic>());
  }

  Future<LiveRoom?> liveCreateRoom({required LiveCreateRoomInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation liveCreateRoom(\$input: LiveCreateRoomInput!) { liveCreateRoom(input: \$input) ${selectionSet ?? '{ id ownerUserId slug title description streamKey status createdAt updatedAt }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['liveCreateRoom'] == null ? null : LiveRoom.fromJson((data['liveCreateRoom'] as Map).cast<String, dynamic>());
  }

  Future<String?> liveRegenerateStreamKey({required String id}) async {
    final data = await _executor.mutation('mutation liveRegenerateStreamKey(\$id: String!) { liveRegenerateStreamKey(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['liveRegenerateStreamKey']?.toString();
  }

  Future<Object?> liveRequestIngest({required String id}) async {
    final data = await _executor.mutation('mutation liveRequestIngest(\$id: String!) { liveRequestIngest(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['liveRequestIngest'];
  }

  Future<bool?> liveEndStream({required String id}) async {
    final data = await _executor.mutation('mutation liveEndStream(\$id: String!) { liveEndStream(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['liveEndStream'] as bool?;
  }

  Future<bool?> liveBanChat({required String id, required String targetUserId, int? durationSeconds}) async {
    final data = await _executor.mutation('mutation liveBanChat(\$id: String!, \$targetUserId: String!, \$durationSeconds: Int) { liveBanChat(id: \$id, targetUserId: \$targetUserId, durationSeconds: \$durationSeconds) }', variables: {
      'id': id,
      'targetUserId': targetUserId,
      'durationSeconds': ?durationSeconds,
    }
    );
    return data['liveBanChat'] as bool?;
  }

  Future<bool?> liveUnbanChat({required String id, required String targetUserId}) async {
    final data = await _executor.mutation('mutation liveUnbanChat(\$id: String!, \$targetUserId: String!) { liveUnbanChat(id: \$id, targetUserId: \$targetUserId) }', variables: {
      'id': id,
      'targetUserId': targetUserId,
    }
    );
    return data['liveUnbanChat'] as bool?;
  }

  Future<bool?> liveViewerHeartbeat({required String id, required String sessionId}) async {
    final data = await _executor.public('mutation liveViewerHeartbeat(\$id: String!, \$sessionId: String!) { liveViewerHeartbeat(id: \$id, sessionId: \$sessionId) }', variables: {
      'id': id,
      'sessionId': sessionId,
    }
    );
    return data['liveViewerHeartbeat'] as bool?;
  }

  Future<bool?> liveReportQoE({required String id, double? bitrateKbps, int? bufferingEvents}) async {
    final data = await _executor.public('mutation liveReportQoE(\$id: String!, \$bitrateKbps: Float, \$bufferingEvents: Int) { liveReportQoE(id: \$id, bitrateKbps: \$bitrateKbps, bufferingEvents: \$bufferingEvents) }', variables: {
      'id': id,
      'bitrateKbps': ?bitrateKbps,
      'bufferingEvents': ?bufferingEvents,
    }
    );
    return data['liveReportQoE'] as bool?;
  }

  Future<bool?> liveAdminDeleteRoom({required String id}) async {
    final data = await _executor.mutation('mutation liveAdminDeleteRoom(\$id: String!) { liveAdminDeleteRoom(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['liveAdminDeleteRoom'] as bool?;
  }

}
