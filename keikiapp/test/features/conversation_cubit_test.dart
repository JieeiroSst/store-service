import 'dart:async';

import 'package:flutter_test/flutter_test.dart';
import 'package:keikiapp/core/utils/load_status.dart';
import 'package:keikiapp/features/chat/bloc/chat_cubit.dart';
import 'package:keikiapp/features/chat/data/chat_repository.dart';
import 'package:mocktail/mocktail.dart';

class _MockChatRepository extends Mock implements ChatRepository {}

const _hello = ChatMessage(id: 1, content: 'chào', fromMe: false, createdAt: '2026-10-21T10:00:00Z');

void main() {
  late _MockChatRepository repo;
  late StreamController<LiveChatEvent> socket;
  var connects = 0;

  setUp(() {
    repo = _MockChatRepository();
    connects = 0;
    socket = StreamController<LiveChatEvent>.broadcast();
    when(() => repo.history(myId: 7, recipientId: 9, groupId: null)).thenAnswer((_) async => [_hello]);
    when(() => repo.live()).thenAnswer((_) {
      connects++;
      return socket.stream;
    });
  });

  tearDown(() => socket.close());

  ConversationCubit direct() => ConversationCubit.direct(repo, myId: 7, recipientId: 9, pollInterval: const Duration(hours: 1));

  test('sends a direct message and keeps it in the thread', () async {
    when(() => repo.sendMessage(text: 'alo', recipientId: 9, groupId: null)).thenAnswer((_) async {});
    final cubit = direct();
    await cubit.load();

    await cubit.send('  alo ');

    verify(() => repo.sendMessage(text: 'alo', recipientId: 9, groupId: null)).called(1);
    expect(cubit.state.sending, isFalse);
    expect(cubit.state.messages.map((m) => m.content), ['chào', 'alo']);
    expect(cubit.state.messages.last.fromMe, isTrue);
    await cubit.close();
  });

  test('drops the pending message when sending fails', () async {
    when(() => repo.sendMessage(text: any(named: 'text'), recipientId: 9, groupId: null)).thenThrow(Exception('offline'));
    final cubit = direct();
    await cubit.load();

    await cubit.send('alo');

    expect(cubit.state.messages, [_hello]);
    expect(cubit.state.error, contains('offline'));
    await cubit.close();
  });

  test('refresh picks up new messages without a loading state', () async {
    final cubit = direct();
    await cubit.load();
    const reply = ChatMessage(id: 2, content: 'đây', fromMe: false, createdAt: '2026-10-21T10:01:00Z');
    when(() => repo.history(myId: 7, recipientId: 9, groupId: null)).thenAnswer((_) async => [_hello, reply]);

    final states = <ConversationState>[];
    final sub = cubit.stream.listen(states.add);
    await cubit.refresh();
    await Future<void>.delayed(Duration.zero);
    await sub.cancel();

    expect(states.map((s) => s.status), everyElement(isNot(LoadStatus.loading)));
    expect(cubit.state.messages, [_hello, reply]);
    await cubit.close();
  });

  test('appends live messages for this conversation only', () async {
    final cubit = direct();
    await cubit.load();
    expect(cubit.state.live, isTrue);

    socket.add(const LiveChatEvent(senderId: 9, recipientId: 7, content: 'có ai không', createdAt: '2026-10-21T10:02:00Z'));
    socket.add(const LiveChatEvent(senderId: 5, recipientId: 7, content: 'người khác', createdAt: '2026-10-21T10:02:01Z'));
    socket.add(const LiveChatEvent(senderId: 9, groupId: 3, content: 'trong nhóm', createdAt: '2026-10-21T10:02:02Z'));
    await Future<void>.delayed(Duration.zero);

    expect(cubit.state.messages.map((m) => m.content), ['chào', 'có ai không']);
    expect(cubit.state.messages.last.fromMe, isFalse);
    await cubit.close();
  });

  test('the echo of my own message replaces the pending copy', () async {
    when(() => repo.sendMessage(text: 'alo', recipientId: 9, groupId: null)).thenAnswer((_) async {});
    final cubit = direct();
    await cubit.load();
    await cubit.send('alo');

    socket.add(const LiveChatEvent(senderId: 7, recipientId: 9, content: 'alo', createdAt: '2026-10-21T10:03:00Z'));
    await Future<void>.delayed(Duration.zero);

    expect(cubit.state.messages.map((m) => m.content), ['chào', 'alo']);
    expect(cubit.state.messages.last.createdAt, '2026-10-21T10:03:00Z');
    await cubit.close();
  });

  test('reconnects after the socket drops and catches up from history', () async {
    final cubit = direct();
    await cubit.load();
    socket.addError(Exception('closed'));
    await Future<void>.delayed(Duration.zero);
    expect(cubit.state.live, isFalse);

    const missed = ChatMessage(id: 3, content: 'tin bị lỡ', fromMe: false, createdAt: '2026-10-21T10:04:00Z');
    when(() => repo.history(myId: 7, recipientId: 9, groupId: null)).thenAnswer((_) async => [_hello, missed]);
    await Future<void>.delayed(const Duration(milliseconds: 1100));

    expect(connects, 2);
    expect(cubit.state.live, isTrue);
    expect(cubit.state.messages.last.content, 'tin bị lỡ');
    await cubit.close();
  });
}
