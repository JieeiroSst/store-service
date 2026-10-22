// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

import '../gateway_executor.dart';
import '../models.g.dart';

class OllamaChatApi {
  const OllamaChatApi(this._executor);

  final GatewayExecutor _executor;

  Future<OllamaChatContacts?> ollamaChatContacts({String? selectionSet}) async {
    final data = await _executor.query('query ollamaChatContacts { ollamaChatContacts ${selectionSet ?? '{ contacts { id username email } }'} }'
    );
    return data['ollamaChatContacts'] == null ? null : OllamaChatContacts.fromJson((data['ollamaChatContacts'] as Map).cast<String, dynamic>());
  }

  Future<OllamaChatGroupSearch?> ollamaChatSearchGroups({required String q, String? selectionSet}) async {
    final data = await _executor.query('query ollamaChatSearchGroups(\$q: String!) { ollamaChatSearchGroups(q: \$q) ${selectionSet ?? '{ groups { id name description creator_id type is_member } }'} }', variables: {
      'q': q,
    }
    );
    return data['ollamaChatSearchGroups'] == null ? null : OllamaChatGroupSearch.fromJson((data['ollamaChatSearchGroups'] as Map).cast<String, dynamic>());
  }

  Future<OllamaChatGroups?> ollamaChatGroups({String? selectionSet}) async {
    final data = await _executor.query('query ollamaChatGroups { ollamaChatGroups ${selectionSet ?? '{ groups { id name description creator_id created_at role } }'} }'
    );
    return data['ollamaChatGroups'] == null ? null : OllamaChatGroups.fromJson((data['ollamaChatGroups'] as Map).cast<String, dynamic>());
  }

  Future<OllamaChatGroupMembers?> ollamaChatGroupMembers({required int groupId, String? selectionSet}) async {
    final data = await _executor.query('query ollamaChatGroupMembers(\$groupId: Int!) { ollamaChatGroupMembers(groupId: \$groupId) ${selectionSet ?? '{ group_id members { id username email role joined_at } }'} }', variables: {
      'groupId': groupId,
    }
    );
    return data['ollamaChatGroupMembers'] == null ? null : OllamaChatGroupMembers.fromJson((data['ollamaChatGroupMembers'] as Map).cast<String, dynamic>());
  }

  Future<OllamaChatHistory?> ollamaChatChatHistory({int? limit, int? offset, int? groupId, int? recipientId, String? selectionSet}) async {
    final data = await _executor.query('query ollamaChatChatHistory(\$limit: Int, \$offset: Int, \$group_id: Int, \$recipient_id: Int) { ollamaChatChatHistory(limit: \$limit, offset: \$offset, group_id: \$group_id, recipient_id: \$recipient_id) ${selectionSet ?? '{ messages { id message_type content created_at sender_id recipient_id group_id url sender_name } count offset limit }'} }', variables: {
      'limit': ?limit,
      'offset': ?offset,
      'group_id': ?groupId,
      'recipient_id': ?recipientId,
    }
    );
    return data['ollamaChatChatHistory'] == null ? null : OllamaChatHistory.fromJson((data['ollamaChatChatHistory'] as Map).cast<String, dynamic>());
  }

  Future<OllamaChatSentMessage?> ollamaChatSendMessage({int? recipientId, int? groupId, String? content, String? url, String? messageType, String? selectionSet}) async {
    final data = await _executor.mutation('mutation ollamaChatSendMessage(\$recipient_id: Int, \$group_id: Int, \$content: String, \$url: String, \$message_type: String) { ollamaChatSendMessage(recipient_id: \$recipient_id, group_id: \$group_id, content: \$content, url: \$url, message_type: \$message_type) ${selectionSet ?? '{ message_type sender_id recipient_id group_id content url created_at }'} }', variables: {
      'recipient_id': ?recipientId,
      'group_id': ?groupId,
      'content': ?content,
      'url': ?url,
      'message_type': ?messageType,
    }
    );
    return data['ollamaChatSendMessage'] == null ? null : OllamaChatSentMessage.fromJson((data['ollamaChatSendMessage'] as Map).cast<String, dynamic>());
  }

  Future<Object?> ollamaChatCreateGroup({required String name_, String? description}) async {
    final data = await _executor.mutation('mutation ollamaChatCreateGroup(\$name: String!, \$description: String) { ollamaChatCreateGroup(name: \$name, description: \$description) }', variables: {
      'name': name_,
      'description': ?description,
    }
    );
    return data['ollamaChatCreateGroup'];
  }

  Future<Object?> ollamaChatInviteToGroup({required int groupId, required int userId, required String role}) async {
    final data = await _executor.mutation('mutation ollamaChatInviteToGroup(\$group_id: Int!, \$user_id: Int!, \$role: String!) { ollamaChatInviteToGroup(group_id: \$group_id, user_id: \$user_id, role: \$role) }', variables: {
      'group_id': groupId,
      'user_id': userId,
      'role': role,
    }
    );
    return data['ollamaChatInviteToGroup'];
  }

  Future<Object?> ollamaChatLeaveGroup({required int groupId}) async {
    final data = await _executor.mutation('mutation ollamaChatLeaveGroup(\$groupId: Int!) { ollamaChatLeaveGroup(groupId: \$groupId) }', variables: {
      'groupId': groupId,
    }
    );
    return data['ollamaChatLeaveGroup'];
  }

  Future<Object?> ollamaChatPromoteMember({required int groupId, required int userId}) async {
    final data = await _executor.mutation('mutation ollamaChatPromoteMember(\$groupId: Int!, \$userId: Int!) { ollamaChatPromoteMember(groupId: \$groupId, userId: \$userId) }', variables: {
      'groupId': groupId,
      'userId': userId,
    }
    );
    return data['ollamaChatPromoteMember'];
  }

  Future<Object?> ollamaChatRemoveMember({required int groupId, required int userId}) async {
    final data = await _executor.mutation('mutation ollamaChatRemoveMember(\$groupId: Int!, \$userId: Int!) { ollamaChatRemoveMember(groupId: \$groupId, userId: \$userId) }', variables: {
      'groupId': groupId,
      'userId': userId,
    }
    );
    return data['ollamaChatRemoveMember'];
  }

}
