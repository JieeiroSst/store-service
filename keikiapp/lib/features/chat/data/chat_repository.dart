import 'dart:async';
import 'dart:convert';

import 'package:equatable/equatable.dart';
import 'package:web_socket_channel/io.dart';

import '../../../core/network/graphql_client.dart';

class ChatContact extends Equatable {
  const ChatContact({required this.id, required this.username, this.email = ''});

  final int id;
  final String username;
  final String email;

  factory ChatContact.fromJson(Map<String, dynamic> json) => ChatContact(
        id: (json['id'] as num?)?.toInt() ?? 0,
        username: json['username'] as String? ?? '',
        email: json['email'] as String? ?? '',
      );

  @override
  List<Object?> get props => [id, username];
}

class ChatGroup extends Equatable {
  const ChatGroup({required this.id, required this.name, this.description = '', this.role = ''});

  final int id;
  final String name;
  final String description;
  final String role;

  factory ChatGroup.fromJson(Map<String, dynamic> json) => ChatGroup(
        id: (json['id'] as num?)?.toInt() ?? 0,
        name: json['name'] as String? ?? '',
        description: json['description'] as String? ?? '',
        role: json['role'] as String? ?? '',
      );

  @override
  List<Object?> get props => [id, name, role];
}

class ChatMessage extends Equatable {
  const ChatMessage({
    required this.content,
    required this.fromMe,
    this.id,
    this.senderName = '',
    this.type = 'text',
    this.url,
    this.createdAt,
  });

  final int? id;
  final String content;
  final bool fromMe;
  final String senderName;
  final String type;
  final String? url;
  final String? createdAt;

  @override
  List<Object?> get props => [id, content, fromMe, createdAt];
}

class LiveChatEvent extends Equatable {
  const LiveChatEvent({required this.content, this.senderId, this.recipientId, this.groupId, this.type = 'text', this.url, this.createdAt});

  final int? senderId;
  final int? recipientId;
  final int? groupId;
  final String content;
  final String type;
  final String? url;
  final String? createdAt;

  static LiveChatEvent? fromJson(Map<String, dynamic> json) {
    if (json['type'] == 'bot') return null;
    final data = json['data'];
    final content = data is Map ? data['content'] as String? ?? '' : '';
    final url = data is Map ? data['url'] as String? : null;
    if (content.isEmpty && (url == null || url.isEmpty)) return null;
    int? id(Object? v) => v is num ? v.toInt() : int.tryParse('${v ?? ''}');
    final group = id(json['group_id']);
    return LiveChatEvent(
      senderId: id(json['sender_id']),
      recipientId: id(json['recipient_id']),
      groupId: group == 0 ? null : group,
      content: content,
      type: json['messageType'] as String? ?? 'text',
      url: url == null || url.isEmpty ? null : url,
      createdAt: json['timestamp'] as String?,
    );
  }

  @override
  List<Object?> get props => [senderId, recipientId, groupId, content, createdAt];
}

class ChatRepository {
  ChatRepository(this._client);

  final GraphQLClient _client;

  Stream<LiveChatEvent> live() {
    late final StreamController<LiveChatEvent> out;
    IOWebSocketChannel? channel;
    StreamSubscription<dynamic>? sub;
    out = StreamController<LiveChatEvent>(
      onListen: () async {
        try {
          final token = await _client.accessToken();
          if (token == null) throw StateError('not signed in');
          channel = IOWebSocketChannel.connect(
            _client.socketUri('chat/ws'),
            headers: {'Authorization': 'Bearer $token'},
            pingInterval: const Duration(seconds: 25),
          );
          await channel!.ready;
          sub = channel!.stream.listen(
            (raw) {
              try {
                final event = LiveChatEvent.fromJson(jsonDecode(raw as String) as Map<String, dynamic>);
                if (event != null) out.add(event);
              } catch (_) {}
            },
            onError: out.addError,
            onDone: out.close,
          );
        } catch (e) {
          out.addError(e);
          await out.close();
        }
      },
      onCancel: () async {
        await sub?.cancel();
        await channel?.sink.close();
      },
    );
    return out.stream;
  }

  Future<List<ChatContact>> contacts() async {
    final data = await _client.query('query { ollamaChatContacts { contacts { id username email } } }');
    final list = (data['ollamaChatContacts'] as Map<String, dynamic>?)?['contacts'] as List? ?? const [];
    return [for (final c in list) ChatContact.fromJson(c as Map<String, dynamic>)];
  }

  Future<List<ChatGroup>> groups() async {
    final data = await _client.query('query { ollamaChatGroups { groups { id name description role } } }');
    final list = (data['ollamaChatGroups'] as Map<String, dynamic>?)?['groups'] as List? ?? const [];
    return [for (final g in list) ChatGroup.fromJson(g as Map<String, dynamic>)];
  }

  Future<List<ChatMessage>> history({required int myId, int? groupId, int? recipientId, int limit = 50}) async {
    final data = await _client.query(r'''
      query History($limit: Int, $group: Int, $recipient: Int) {
        ollamaChatChatHistory(limit: $limit, offset: 0, group_id: $group, recipient_id: $recipient) {
          messages { id message_type content created_at sender_id sender_name url }
        }
      }
    ''', variables: {'limit': limit, 'group': groupId, 'recipient': recipientId});
    final list = (data['ollamaChatChatHistory'] as Map<String, dynamic>?)?['messages'] as List? ?? const [];
    final messages = [
      for (final m in list.cast<Map<String, dynamic>>())
        ChatMessage(
          id: (m['id'] as num?)?.toInt(),
          content: m['content'] as String? ?? '',
          fromMe: (m['sender_id'] as num?)?.toInt() == myId,
          senderName: m['sender_name'] as String? ?? '',
          type: m['message_type'] as String? ?? 'text',
          url: m['url'] as String?,
          createdAt: m['created_at'] as String?,
        ),
    ];
    messages.sort((a, b) => (a.createdAt ?? '').compareTo(b.createdAt ?? ''));
    return messages;
  }

  Future<void> sendMessage({required String text, int? groupId, int? recipientId}) async {
    await _client.mutation(r'''
      mutation Send($recipient: Int, $group: Int, $content: String) {
        ollamaChatSendMessage(recipient_id: $recipient, group_id: $group, content: $content, message_type: "text") { created_at }
      }
    ''', variables: {'recipient': recipientId, 'group': groupId, 'content': text});
  }

  Future<void> createGroup(String name, {String? description}) async {
    await _client.mutation(r'''
      mutation Create($name: String!, $desc: String) { ollamaChatCreateGroup(name: $name, description: $desc) }
    ''', variables: {'name': name, 'desc': description});
  }

  Future<void> leaveGroup(int groupId) async {
    await _client.mutation(r'mutation Leave($id: Int!) { ollamaChatLeaveGroup(groupId: $id) }', variables: {'id': groupId});
  }

  Future<List<ChatMessage>> assistantHistory({int limit = 50}) async {
    final data = await _client.query(r'''
      query AiHistory($limit: Int) { aiAgentChatHistory(limit: $limit) { messages { role content created_at } } }
    ''', variables: {'limit': limit});
    final list = (data['aiAgentChatHistory'] as Map<String, dynamic>?)?['messages'] as List? ?? const [];
    return [
      for (final m in list.cast<Map<String, dynamic>>())
        ChatMessage(
          content: m['content'] as String? ?? '',
          fromMe: m['role'] == 'user',
          senderName: m['role'] == 'user' ? '' : 'Keiki AI',
          createdAt: m['created_at'] as String?,
        ),
    ];
  }

  Future<ChatMessage> askAssistant(List<ChatMessage> conversation) async {
    final data = await _client.mutation(r'''
      mutation Ask($input: AiAgentChatInput!) { aiAgentChat(input: $input) { message { role content } created_at } }
    ''', variables: {
      'input': {
        'messages': [
          for (final m in conversation) {'role': m.fromMe ? 'user' : 'assistant', 'content': m.content},
        ],
      },
    });
    final res = data['aiAgentChat'] as Map<String, dynamic>?;
    final message = res?['message'] as Map<String, dynamic>?;
    return ChatMessage(
      content: message?['content'] as String? ?? '',
      fromMe: false,
      senderName: 'Keiki AI',
      createdAt: res?['created_at'] as String?,
    );
  }
}
