import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:go_router/go_router.dart';

import '../core/config/env.dart';
import '../core/network/graphql_client.dart';
import '../core/push/firebase_push_transport.dart';
import '../core/push/push_service.dart';
import '../core/storage/token_storage.dart';
import '../features/address/data/address_repository.dart';
import '../features/auth/bloc/auth_bloc.dart';
import '../features/auth/data/auth_repository.dart';
import '../features/chat/data/chat_repository.dart';
import '../features/food/data/food_repository.dart';
import '../features/notifications/data/notification_repository.dart';
import '../features/services/bloc/services_cubit.dart';
import '../features/services/data/schema_catalog.dart';
import '../features/shopping/data/shopping_repository.dart';
import '../features/wallet/bloc/wallet_cubit.dart';
import '../features/wallet/data/wallet_repository.dart';
import '../gen/gateway_api.g.dart';
import 'role.dart';
import 'router.dart';
import 'theme.dart';

class KeikiApp extends StatefulWidget {
  const KeikiApp({super.key, this.client, this.catalogLoader, this.pushTransport});

  final GraphQLClient? client;
  final Future<SchemaCatalog> Function()? catalogLoader;
  final PushTransport? pushTransport;

  @override
  State<KeikiApp> createState() => _KeikiAppState();
}

class _KeikiAppState extends State<KeikiApp> {
  late final GraphQLClient _client = widget.client ?? GraphQLClient(endpoint: Env.graphqlEndpoint, storage: TokenStorage());
  late final AuthRepository _authRepository = AuthRepository(_client);
  late final WalletRepository _walletRepository = WalletRepository(_client);
  late final AuthBloc _auth = AuthBloc(_authRepository)..add(const AuthStarted());
  late final WalletCubit _wallet = WalletCubit(_walletRepository);
  late final ServicesCubit _services = ServicesCubit(_client, loader: widget.catalogLoader)..load();
  final RoleCubit _role = RoleCubit();
  late final GoRouter _router = buildRouter(_auth, _role);
  late final NotificationRepository _notifications = NotificationRepository(_client);
  late final PushService _push = PushService(_notifications, widget.pushTransport ?? (FirebasePushConfig.configured ? FirebasePushTransport() : null));
  final _messenger = GlobalKey<ScaffoldMessengerState>();
  final _subs = <StreamSubscription<dynamic>>[];

  @override
  void initState() {
    super.initState();
    _subs
      ..add(_push.navigate.listen(_router.push))
      ..add(_push.foreground.listen(_showForeground));
    _push.start();
  }

  void _showForeground(PushMessage m) {
    final route = m.route;
    _messenger.currentState
      ?..hideCurrentSnackBar()
      ..showSnackBar(
        SnackBar(
          content: Text([m.title, m.body].whereType<String>().where((s) => s.isNotEmpty).join('\n')),
          action: route == null ? null : SnackBarAction(label: 'Xem', onPressed: () => _router.push(route)),
        ),
      );
  }

  @override
  void dispose() {
    for (final s in _subs) {
      s.cancel();
    }
    _push.dispose();
    _router.dispose();
    _wallet.close();
    _services.close();
    _role.close();
    _auth.close();
    _client.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return MultiRepositoryProvider(
      providers: [
        RepositoryProvider.value(value: _client),
        RepositoryProvider(create: (_) => GatewayApi(_client)),
        RepositoryProvider.value(value: _authRepository),
        RepositoryProvider.value(value: _walletRepository),
        RepositoryProvider(create: (_) => AddressRepository(_client)),
        RepositoryProvider(create: (_) => FoodRepository(_client)),
        RepositoryProvider(create: (_) => ShoppingRepository(_client)),
        RepositoryProvider(create: (_) => ChatRepository(_client)),
        RepositoryProvider.value(value: _notifications),
        RepositoryProvider.value(value: _push),
      ],
      child: MultiBlocProvider(
        providers: [
          BlocProvider.value(value: _auth),
          BlocProvider.value(value: _wallet),
          BlocProvider.value(value: _services),
          BlocProvider.value(value: _role),
        ],
        child: MultiBlocListener(
          listeners: [
            BlocListener<AuthBloc, AuthState>(
              listenWhen: (p, c) => p.user != c.user && c.status == AuthStatus.authenticated,
              listener: (context, state) => _role.report('me', state.user?.roles ?? const []),
            ),
            BlocListener<ServicesCubit, ServicesState>(
              listenWhen: (p, c) => p.permissions != c.permissions,
              listener: (context, state) => _role.report('authorize', state.permissions.roles),
            ),
            BlocListener<AuthBloc, AuthState>(
              listenWhen: (p, c) => p.status != c.status,
              listener: (context, state) {
                if (state.status == AuthStatus.authenticated) {
                  _role.report('me', state.user?.roles ?? const []);
                  _wallet.load();
                  _services.loadPermissions();
                  _push.onLogin();
                }
                if (state.status == AuthStatus.unauthenticated) {
                  _role.reset();
                  _wallet.reset();
                  _services.clearPermissions();
                }
              },
            ),
          ],
          child: MaterialApp.router(
            scaffoldMessengerKey: _messenger,
            title: 'Keiki',
            debugShowCheckedModeBanner: false,
            theme: buildTheme(Brightness.light),
            darkTheme: buildTheme(Brightness.dark),
            routerConfig: _router,
          ),
        ),
      ),
    );
  }
}
