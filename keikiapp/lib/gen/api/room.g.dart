// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

import '../gateway_executor.dart';
import '../models.g.dart';

class RoomApi {
  const RoomApi(this._executor);

  final GatewayExecutor _executor;

  Future<RoomUser?> roomMe({String? selectionSet}) async {
    final data = await _executor.query('query roomMe { roomMe ${selectionSet ?? '{ id username }'} }'
    );
    return data['roomMe'] == null ? null : RoomUser.fromJson((data['roomMe'] as Map).cast<String, dynamic>());
  }

  Future<List<RoomRoom>?> rooms({String? selectionSet}) async {
    final data = await _executor.query('query rooms { rooms ${selectionSet ?? '{ id name owner member_count created_at }'} }'
    );
    return (data['rooms'] as List?)?.map((e) => (e == null ? null : RoomRoom.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<RoomRoom?> room({required int id, String? selectionSet}) async {
    final data = await _executor.query('query room(\$id: Int!) { room(id: \$id) ${selectionSet ?? '{ id name owner member_count created_at }'} }', variables: {
      'id': id,
    }
    );
    return data['room'] == null ? null : RoomRoom.fromJson((data['room'] as Map).cast<String, dynamic>());
  }

  Future<List<RoomMember>?> roomMembers({required int id, String? selectionSet}) async {
    final data = await _executor.query('query roomMembers(\$id: Int!) { roomMembers(id: \$id) ${selectionSet ?? '{ room_id username role joined_at }'} }', variables: {
      'id': id,
    }
    );
    return (data['roomMembers'] as List?)?.map((e) => (e == null ? null : RoomMember.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<List<RoomMessage>?> roomMessages({required int id, int? before, int? limit, String? selectionSet}) async {
    final data = await _executor.query('query roomMessages(\$id: Int!, \$before: Int, \$limit: Int) { roomMessages(id: \$id, before: \$before, limit: \$limit) ${selectionSet ?? '{ id room_id username content created_at }'} }', variables: {
      'id': id,
      'before': ?before,
      'limit': ?limit,
    }
    );
    return (data['roomMessages'] as List?)?.map((e) => (e == null ? null : RoomMessage.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<RoomRoom?> roomCreateRoom({required String name_, String? selectionSet}) async {
    final data = await _executor.mutation('mutation roomCreateRoom(\$name: String!) { roomCreateRoom(name: \$name) ${selectionSet ?? '{ id name owner member_count created_at }'} }', variables: {
      'name': name_,
    }
    );
    return data['roomCreateRoom'] == null ? null : RoomRoom.fromJson((data['roomCreateRoom'] as Map).cast<String, dynamic>());
  }

  Future<Object?> roomAddMember({required String id, required String username}) async {
    final data = await _executor.mutation('mutation roomAddMember(\$id: String!, \$username: String!) { roomAddMember(id: \$id, username: \$username) }', variables: {
      'id': id,
      'username': username,
    }
    );
    return data['roomAddMember'];
  }

  Future<Object?> roomRemoveMember({required String id, required String username}) async {
    final data = await _executor.mutation('mutation roomRemoveMember(\$id: String!, \$username: String!) { roomRemoveMember(id: \$id, username: \$username) }', variables: {
      'id': id,
      'username': username,
    }
    );
    return data['roomRemoveMember'];
  }

}
