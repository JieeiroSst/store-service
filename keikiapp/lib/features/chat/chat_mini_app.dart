import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:go_router/go_router.dart';

import '../../app/deep_links.dart';

import '../../app/mini_app.dart';
import '../auth/bloc/auth_bloc.dart';
import 'bloc/chat_cubit.dart';
import 'data/chat_repository.dart';
import 'view/chat_home_page.dart';
import 'view/conversation_page.dart';

int _myId(BuildContext context) => context.read<AuthBloc>().state.user?.id ?? 0;

final chatMiniApp = MiniApp(
  id: 'chat',
  title: 'Chat',
  icon: Icons.chat_bubble,
  color: Colors.teal,
  path: '/chat',
  routes: [
    ShellRoute(
      builder: (context, state, child) => BlocProvider(create: (context) => ChatListCubit(context.read<ChatRepository>())..load(), child: child),
      routes: [
        GoRoute(
          path: '/chat',
          builder: (_, _) => const ChatHomePage(),
          routes: [
            GoRoute(
              path: 'assistant',
              builder: (context, _) => BlocProvider(
                create: (context) => ConversationCubit.assistant(context.read<ChatRepository>())..load(),
                child: const ConversationPage(title: 'Keiki AI'),
              ),
            ),
            GoRoute(
              path: 'direct/:id',
              builder: (context, state) => intParam(state, 'id') == null
                  ? LinkNotFoundPage(location: state.uri.toString())
                  : BlocProvider(
                      create: (context) =>
                          ConversationCubit.direct(context.read<ChatRepository>(), myId: _myId(context), recipientId: intParam(state, 'id')!)..load(),
                      child: ConversationPage(title: state.extra as String? ?? 'Trò chuyện'),
                    ),
            ),
            GoRoute(
              path: 'group/:id',
              builder: (context, state) => intParam(state, 'id') == null
                  ? LinkNotFoundPage(location: state.uri.toString())
                  : BlocProvider(
                      create: (context) =>
                          ConversationCubit.group(context.read<ChatRepository>(), myId: _myId(context), groupId: intParam(state, 'id')!)..load(),
                      child: ConversationPage(title: state.extra as String? ?? 'Nhóm'),
                    ),
            ),
          ],
        ),
      ],
    ),
  ],
);
