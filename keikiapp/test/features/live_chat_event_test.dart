import 'package:flutter_test/flutter_test.dart';
import 'package:keikiapp/features/chat/data/chat_repository.dart';

void main() {
  test('parses the ollama-service websocket frame', () {
    final e = LiveChatEvent.fromJson({
      'type': 'user',
      'messageType': 'text',
      'sender_id': '7',
      'recipient_id': '9',
      'data': {'url': '', 'content': 'chào'},
      'timestamp': '2026-10-22T03:00:00Z',
    });
    expect(e, const LiveChatEvent(senderId: 7, recipientId: 9, content: 'chào', createdAt: '2026-10-22T03:00:00Z'));
    expect(e!.groupId, isNull);
    expect(e.url, isNull);
  });

  test('keeps group ids and skips bot or empty frames', () {
    final g = LiveChatEvent.fromJson({'type': 'group', 'sender_id': '7', 'group_id': 3, 'data': {'content': 'hi'}});
    expect(g!.groupId, 3);
    expect(LiveChatEvent.fromJson({'type': 'bot', 'sender_id': 'system', 'data': {'content': 'x'}}), isNull);
    expect(LiveChatEvent.fromJson({'type': 'user', 'sender_id': '7', 'data': {'content': ''}}), isNull);
  });
}
