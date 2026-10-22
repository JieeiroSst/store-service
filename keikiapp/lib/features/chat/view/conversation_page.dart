import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

import '../../../core/utils/formatters.dart';
import '../../../core/utils/load_status.dart';
import '../../../core/widgets/status_view.dart';
import '../bloc/chat_cubit.dart';
import '../data/chat_repository.dart';
import '../../../gen/gateway_api.g.dart' show GatewayApi;
import '../../apps/extras/extras.dart';

class ConversationPage extends StatefulWidget {
  const ConversationPage({super.key, required this.title});

  final String title;

  @override
  State<ConversationPage> createState() => _ConversationPageState();
}

class _ConversationPageState extends State<ConversationPage> {
  final _input = TextEditingController();

  @override
  void dispose() {
    _input.dispose();
    super.dispose();
  }

  void _send() {
    context.read<ConversationCubit>().send(_input.text);
    _input.clear();
  }

  @override
  Widget build(BuildContext context) {
    final cubit = context.read<ConversationCubit>();
    return Scaffold(
      appBar: AppBar(
        title: Text(widget.title),
        actions: [
          if (cubit.groupId != null)
            IconButton(
              tooltip: 'Thành viên',
              icon: const Icon(Icons.group_outlined),
              onPressed: () => Navigator.of(context).push(MaterialPageRoute<void>(builder: (_) => ChatGroupMembersPage(groupId: cubit.groupId!, title: widget.title))),
            ),
          if (cubit.isAssistant)
            IconButton(
              tooltip: 'AI Studio',
              icon: const Icon(Icons.auto_awesome_outlined),
              onPressed: () => Navigator.of(context).push(MaterialPageRoute<void>(builder: (_) => const AiStudioPage())),
            ),
        ],
      ),
      body: BlocConsumer<ConversationCubit, ConversationState>(
        listenWhen: (p, c) => c.error != null && c.status.isSuccess,
        listener: (context, state) => showSnack(context, state.error!),
        builder: (context, state) {
          final typing = state.sending && cubit.isAssistant;
          return Column(
            children: [
              if (!cubit.isAssistant && state.status.isSuccess && !state.live)
                Container(
                  width: double.infinity,
                  color: Theme.of(context).colorScheme.tertiaryContainer,
                  padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 6),
                  child: Row(
                    children: [
                      const SizedBox(width: 12, height: 12, child: CircularProgressIndicator(strokeWidth: 2)),
                      const SizedBox(width: 8),
                      Text('Đang kết nối lại…', style: TextStyle(color: Theme.of(context).colorScheme.onTertiaryContainer, fontSize: 12)),
                    ],
                  ),
                ),
              Expanded(
                child: StatusView(
                  status: state.status,
                  error: state.error,
                  onRetry: cubit.load,
                  isEmpty: state.messages.isEmpty && !state.sending,
                  emptyMessage: cubit.isAssistant ? 'Bắt đầu trò chuyện với Keiki AI' : 'Chưa có tin nhắn',
                  builder: (context) => ListView.builder(
                    reverse: true,
                    padding: const EdgeInsets.all(12),
                    itemCount: state.messages.length + (typing ? 1 : 0),
                    itemBuilder: (context, i) {
                      if (typing && i == 0) return const _Typing();
                      final index = state.messages.length - 1 - (i - (typing ? 1 : 0));
                      final m = state.messages[index];
                      return GestureDetector(
                        onLongPress: !m.fromMe || m.id == null || cubit.isAssistant
                            ? null
                            : () async {
                                final ok = await showDialog<bool>(
                                  context: context,
                                  builder: (context) => AlertDialog(
                                    content: const Text('Thu hồi tin nhắn này?'),
                                    actions: [
                                      TextButton(onPressed: () => Navigator.pop(context, false), child: const Text('Huỷ')),
                                      FilledButton(onPressed: () => Navigator.pop(context, true), child: const Text('Thu hồi')),
                                    ],
                                  ),
                                );
                                if (ok != true || !context.mounted) return;
                                try {
                                  await context.read<GatewayApi>().chat.deleteMyChatMessage(messageID: m.id!);
                                  await cubit.refresh();
                                } catch (e) {
                                  if (context.mounted) showSnack(context, e.toString());
                                }
                              },
                        child: _Bubble(message: m, showSender: !cubit.isAssistant),
                      );
                    },
                  ),
                ),
              ),
              SafeArea(
                child: Padding(
                  padding: const EdgeInsets.fromLTRB(12, 4, 12, 8),
                  child: Row(
                    children: [
                      Expanded(
                        child: TextField(
                          controller: _input,
                          minLines: 1,
                          maxLines: 4,
                          textInputAction: TextInputAction.send,
                          onSubmitted: (_) => _send(),
                          decoration: const InputDecoration(hintText: 'Nhập tin nhắn…', border: OutlineInputBorder()),
                        ),
                      ),
                      const SizedBox(width: 8),
                      IconButton.filled(onPressed: state.sending ? null : _send, icon: const Icon(Icons.send)),
                    ],
                  ),
                ),
              ),
            ],
          );
        },
      ),
    );
  }
}

class _Bubble extends StatelessWidget {
  const _Bubble({required this.message, required this.showSender});

  final ChatMessage message;
  final bool showSender;

  @override
  Widget build(BuildContext context) {
    final scheme = Theme.of(context).colorScheme;
    final mine = message.fromMe;
    return Align(
      alignment: mine ? Alignment.centerRight : Alignment.centerLeft,
      child: ConstrainedBox(
        constraints: BoxConstraints(maxWidth: MediaQuery.sizeOf(context).width * 0.78),
        child: Container(
          margin: const EdgeInsets.symmetric(vertical: 4),
          padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
          decoration: BoxDecoration(color: mine ? scheme.primary : scheme.surfaceContainerHighest, borderRadius: BorderRadius.circular(16)),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              if (showSender && !mine && message.senderName.isNotEmpty)
                Text(
                  message.senderName,
                  style: TextStyle(fontWeight: FontWeight.bold, color: scheme.primary, fontSize: 12),
                ),
              SelectableText(message.content, style: TextStyle(color: mine ? scheme.onPrimary : scheme.onSurfaceVariant)),
              if (message.createdAt != null)
                Text(
                  formatDate(message.createdAt),
                  style: TextStyle(fontSize: 10, color: (mine ? scheme.onPrimary : scheme.onSurfaceVariant).withValues(alpha: 0.7)),
                ),
            ],
          ),
        ),
      ),
    );
  }
}

class _Typing extends StatelessWidget {
  const _Typing();

  @override
  Widget build(BuildContext context) {
    return const Align(
      alignment: Alignment.centerLeft,
      child: Padding(
        padding: EdgeInsets.all(8),
        child: SizedBox(width: 24, height: 24, child: CircularProgressIndicator(strokeWidth: 2)),
      ),
    );
  }
}
