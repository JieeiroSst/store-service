// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

import '../gateway_executor.dart';
import '../models.g.dart';

class ChatbotApi {
  const ChatbotApi(this._executor);

  final GatewayExecutor _executor;

  Future<ChatbotHistory?> chatbotConversationHistory({required int conversationId, int? limit, int? offset, String? selectionSet}) async {
    final data = await _executor.query('query chatbotConversationHistory(\$conversation_id: Int!, \$limit: Int, \$offset: Int) { chatbotConversationHistory(conversation_id: \$conversation_id, limit: \$limit, offset: \$offset) ${selectionSet ?? '{ messages { id conversation_id sender_id content message_type ai_model created_at } count }'} }', variables: {
      'conversation_id': conversationId,
      'limit': ?limit,
      'offset': ?offset,
    }
    );
    return data['chatbotConversationHistory'] == null ? null : ChatbotHistory.fromJson((data['chatbotConversationHistory'] as Map).cast<String, dynamic>());
  }

  Future<ChatbotConversationList?> chatbotConversations({String? selectionSet}) async {
    final data = await _executor.query('query chatbotConversations { chatbotConversations ${selectionSet ?? '{ conversations { id user1_id user2_id is_ai_chat active_ai_model created_at updated_at } count }'} }'
    );
    return data['chatbotConversations'] == null ? null : ChatbotConversationList.fromJson((data['chatbotConversations'] as Map).cast<String, dynamic>());
  }

}
