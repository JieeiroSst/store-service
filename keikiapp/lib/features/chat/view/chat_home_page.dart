import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:go_router/go_router.dart';

import '../../../core/utils/load_status.dart';
import '../../../core/widgets/status_view.dart';
import '../bloc/chat_cubit.dart';
import '../data/chat_repository.dart';
import '../../apps/extras/extras.dart';

class ChatHomePage extends StatelessWidget {
  const ChatHomePage({super.key});

  Future<void> _createGroup(BuildContext context) async {
    final name = TextEditingController();
    final desc = TextEditingController();
    final ok = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Tạo nhóm'),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            TextField(controller: name, decoration: const InputDecoration(labelText: 'Tên nhóm')),
            TextField(controller: desc, decoration: const InputDecoration(labelText: 'Mô tả')),
          ],
        ),
        actions: [
          TextButton(onPressed: () => Navigator.pop(context, false), child: const Text('Huỷ')),
          FilledButton(onPressed: () => Navigator.pop(context, true), child: const Text('Tạo')),
        ],
      ),
    );
    if (ok != true || name.text.trim().isEmpty || !context.mounted) return;
    try {
      await context.read<ChatListCubit>().createGroup(name.text.trim(), desc.text.trim());
    } catch (e) {
      if (context.mounted) showSnack(context, e.toString());
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Keiki Chat'),
        actions: [
          IconButton(
            tooltip: 'Tìm nhóm',
            icon: const Icon(Icons.travel_explore),
            onPressed: () => Navigator.of(context).push(MaterialPageRoute<void>(builder: (_) => const ChatGroupSearchPage())),
          ),
        ],
      ),
      floatingActionButton: FloatingActionButton(
        tooltip: 'Tạo nhóm',
        onPressed: () => _createGroup(context),
        child: const Icon(Icons.group_add_outlined),
      ),
      body: BlocBuilder<ChatListCubit, ChatListState>(
        builder: (context, state) {
          final cubit = context.read<ChatListCubit>();
          return RefreshIndicator(
            onRefresh: cubit.load,
            child: ListView(
              children: [
                Card(
                  margin: const EdgeInsets.all(12),
                  color: Theme.of(context).colorScheme.tertiaryContainer,
                  child: ListTile(
                    leading: const CircleAvatar(child: Icon(Icons.auto_awesome)),
                    title: const Text('Trợ lý Keiki AI'),
                    subtitle: const Text('Hỏi gì cũng được'),
                    trailing: const Icon(Icons.chevron_right),
                    onTap: () => context.push('/chat/assistant'),
                  ),
                ),
                if (state.status.isFailure && state.contacts.isEmpty && state.groups.isEmpty)
                  ErrorView(message: state.error ?? '', onRetry: cubit.load),
                if (state.status.isLoading && state.contacts.isEmpty && state.groups.isEmpty)
                  const Padding(padding: EdgeInsets.all(32), child: Center(child: CircularProgressIndicator())),
                if (state.groups.isNotEmpty) const _Header('Nhóm'),
                for (final g in state.groups) _GroupTile(group: g),
                if (state.contacts.isNotEmpty) const _Header('Danh bạ'),
                for (final c in state.contacts)
                  ListTile(
                    leading: CircleAvatar(child: Text(c.username.isEmpty ? '?' : c.username[0].toUpperCase())),
                    title: Text(c.username),
                    subtitle: Text(c.email),
                    onTap: () => context.push('/chat/direct/${c.id}', extra: c.username),
                  ),
              ],
            ),
          );
        },
      ),
    );
  }
}

class _Header extends StatelessWidget {
  const _Header(this.text);

  final String text;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.fromLTRB(16, 16, 16, 4),
      child: Text(text, style: Theme.of(context).textTheme.titleSmall),
    );
  }
}

class _GroupTile extends StatelessWidget {
  const _GroupTile({required this.group});

  final ChatGroup group;

  @override
  Widget build(BuildContext context) {
    return ListTile(
      leading: const CircleAvatar(child: Icon(Icons.group)),
      title: Text(group.name),
      subtitle: Text(group.description.isEmpty ? group.role : group.description),
      onTap: () => context.push('/chat/group/${group.id}', extra: group.name),
      trailing: PopupMenuButton<String>(
        onSelected: (_) async {
          try {
            await context.read<ChatListCubit>().leaveGroup(group);
          } catch (e) {
            if (context.mounted) showSnack(context, e.toString());
          }
        },
        itemBuilder: (_) => const [PopupMenuItem(value: 'leave', child: Text('Rời nhóm'))],
      ),
    );
  }
}
