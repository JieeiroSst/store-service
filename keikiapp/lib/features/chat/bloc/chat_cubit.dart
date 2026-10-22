import 'dart:async';

import 'package:equatable/equatable.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

import '../../../core/utils/load_status.dart';
import '../data/chat_repository.dart';

class ChatListState extends Equatable {
  const ChatListState({
    this.status = LoadStatus.initial,
    this.contacts = const [],
    this.groups = const [],
    this.error,
  });

  final LoadStatus status;
  final List<ChatContact> contacts;
  final List<ChatGroup> groups;
  final String? error;

  @override
  List<Object?> get props => [status, contacts, groups, error];
}

class ChatListCubit extends Cubit<ChatListState> {
  ChatListCubit(this._repository) : super(const ChatListState());

  final ChatRepository _repository;

  Future<void> load() async {
    emit(ChatListState(status: LoadStatus.loading, contacts: state.contacts, groups: state.groups));
    try {
      final results = await Future.wait([_repository.contacts(), _repository.groups()]);
      emit(ChatListState(
        status: LoadStatus.success,
        contacts: results[0] as List<ChatContact>,
        groups: results[1] as List<ChatGroup>,
      ));
    } catch (e) {
      emit(ChatListState(status: LoadStatus.failure, contacts: state.contacts, groups: state.groups, error: e.toString()));
    }
  }

  Future<void> createGroup(String name, String description) async {
    await _repository.createGroup(name, description: description);
    await load();
  }

  Future<void> leaveGroup(ChatGroup group) async {
    await _repository.leaveGroup(group.id);
    await load();
  }
}

class ConversationState extends Equatable {
  const ConversationState({this.status = LoadStatus.initial, this.messages = const [], this.sending = false, this.live = false, this.error});

  final LoadStatus status;
  final List<ChatMessage> messages;
  final bool sending;
  final bool live;
  final String? error;

  ConversationState copyWith({LoadStatus? status, List<ChatMessage>? messages, bool? sending, bool? live, String? error}) =>
      ConversationState(
        status: status ?? this.status,
        messages: messages ?? this.messages,
        sending: sending ?? this.sending,
        live: live ?? this.live,
        error: error,
      );

  @override
  List<Object?> get props => [status, messages, sending, live, error];
}

class ConversationCubit extends Cubit<ConversationState> {
  ConversationCubit.direct(this._repository, {required this.myId, required int this._recipientId, this.pollInterval = _fallbackPoll})
      : _groupId = null,
        isAssistant = false,
        super(const ConversationState());

  ConversationCubit.group(this._repository, {required this.myId, required int this._groupId, this.pollInterval = _fallbackPoll})
      : _recipientId = null,
        isAssistant = false,
        super(const ConversationState());

  ConversationCubit.assistant(this._repository)
      : myId = 0,
        _groupId = null,
        _recipientId = null,
        isAssistant = true,
        pollInterval = null,
        super(const ConversationState());

  static const _fallbackPoll = Duration(seconds: 15);
  static const _maxBackoff = Duration(seconds: 30);

  final ChatRepository _repository;
  final int myId;
  final int? _groupId;
  final int? _recipientId;
  final bool isAssistant;
  final Duration? pollInterval;
  StreamSubscription<LiveChatEvent>? _live;

  int? get groupId => _groupId;
  Timer? _poll;
  Timer? _retry;
  int _attempt = 0;

  Future<List<ChatMessage>> _fetch() => isAssistant
      ? _repository.assistantHistory()
      : _repository.history(myId: myId, groupId: _groupId, recipientId: _recipientId);

  Future<void> load() async {
    emit(state.copyWith(status: LoadStatus.loading));
    try {
      emit(state.copyWith(status: LoadStatus.success, messages: await _fetch()));
      if (!isAssistant) _connect();
    } catch (e) {
      emit(state.copyWith(status: LoadStatus.failure, error: e.toString()));
    }
  }

  void _connect() {
    if (isClosed) return;
    _retry?.cancel();
    _live?.cancel();
    _live = _repository.live().listen(
      _onEvent,
      onError: (Object _) => _lost(),
      onDone: _lost,
      cancelOnError: true,
    );
    _poll?.cancel();
    _poll = null;
    emit(state.copyWith(live: true));
  }

  void _lost() {
    if (isClosed) return;
    _live = null;
    emit(state.copyWith(live: false));
    if (pollInterval != null) _poll ??= Timer.periodic(pollInterval!, (_) => refresh());
    final delay = Duration(seconds: 1 << _attempt.clamp(0, 5));
    _attempt++;
    _retry = Timer(delay > _maxBackoff ? _maxBackoff : delay, () async {
      _connect();
      await refresh();
    });
  }

  bool _belongs(LiveChatEvent e) {
    if (_groupId != null) return e.groupId == _groupId;
    if (e.groupId != null) return false;
    return (e.senderId == _recipientId && e.recipientId == myId) || (e.senderId == myId && e.recipientId == _recipientId);
  }

  void _onEvent(LiveChatEvent e) {
    _attempt = 0;
    if (isClosed || !_belongs(e)) return;
    final fromMe = e.senderId == myId;
    final incoming = ChatMessage(content: e.content, fromMe: fromMe, type: e.type, url: e.url, createdAt: e.createdAt);
    final messages = [...state.messages];
    if (messages.any((m) => m.content == e.content && m.createdAt == e.createdAt && m.fromMe == fromMe)) return;
    final pending = fromMe ? messages.lastIndexWhere((m) => m.id == null && m.fromMe && m.content == e.content) : -1;
    if (pending >= 0) {
      messages[pending] = incoming;
    } else {
      messages.add(incoming);
    }
    emit(state.copyWith(messages: messages));
  }

  Future<void> refresh() async {
    if (isClosed || state.sending || !state.status.isSuccess) return;
    try {
      final messages = await _fetch();
      if (!isClosed && !state.sending) emit(state.copyWith(messages: messages));
    } catch (_) {}
  }

  Future<void> send(String text) async {
    if (text.trim().isEmpty || state.sending) return;
    final mine = ChatMessage(content: text.trim(), fromMe: true, createdAt: DateTime.now().toIso8601String());
    final conversation = [...state.messages, mine];
    emit(state.copyWith(messages: conversation, sending: true));
    try {
      if (isAssistant) {
        final reply = await _repository.askAssistant(conversation.length > 20 ? conversation.sublist(conversation.length - 20) : conversation);
        emit(state.copyWith(messages: [...conversation, reply], sending: false));
      } else {
        await _repository.sendMessage(text: mine.content, groupId: _groupId, recipientId: _recipientId);
        emit(state.copyWith(sending: false));
      }
    } catch (e) {
      emit(state.copyWith(messages: isAssistant ? conversation : state.messages.where((m) => !identical(m, mine)).toList(), sending: false, error: e.toString()));
    }
  }

  @override
  Future<void> close() {
    _poll?.cancel();
    _retry?.cancel();
    _live?.cancel();
    return super.close();
  }
}
