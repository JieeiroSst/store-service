import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:go_router/go_router.dart';

import '../core/widgets/ui.dart';
import '../features/auth/bloc/auth_bloc.dart';
import '../features/auth/view/login_page.dart';
import '../features/auth/view/signup_page.dart';
import '../features/home/view/home_page.dart';
import '../features/notifications/bloc/notification_settings_cubit.dart';
import '../features/notifications/data/notification_repository.dart';
import '../features/notifications/view/notification_settings_page.dart';
import '../features/profile/view/profile_page.dart';
import '../features/services/view/services_home_page.dart';
import '../features/admin/admin_home_page.dart';
import 'mini_apps.dart';
import 'role.dart';

final _rootKey = GlobalKey<NavigatorState>();

GoRouter buildRouter(AuthBloc auth, RoleCubit role) {
  return GoRouter(
    navigatorKey: _rootKey,
    initialLocation: '/',
    refreshListenable: Listenable.merge([
      _StreamListenable(auth.stream.map((s) => s.status).distinct()),
      _StreamListenable(role.stream.map((s) => s.isAdmin).distinct()),
    ]),
    redirect: (context, state) {
      final status = auth.state.status;
      final loc = state.matchedLocation;
      final atAuth = loc == '/login' || loc == '/signup';
      if (status == AuthStatus.unknown) return loc == '/splash' ? null : '/splash';
      if (status == AuthStatus.unauthenticated) return atAuth ? null : '/login';
      if (atAuth || loc == '/splash') return '/';
      if (loc.startsWith('/admin') && !role.state.isAdmin) return '/';
      return null;
    },
    routes: [
      GoRoute(path: '/splash', builder: (_, _) => const _SplashPage()),
      GoRoute(path: '/login', builder: (_, _) => const LoginPage()),
      GoRoute(path: '/signup', builder: (_, _) => const SignUpPage()),
      StatefulShellRoute.indexedStack(
        builder: (context, state, shell) => _AppShell(shell: shell),
        branches: [
          StatefulShellBranch(routes: [GoRoute(path: '/', builder: (_, _) => const HomePage())]),
          StatefulShellBranch(routes: [GoRoute(path: '/services', builder: (_, _) => const ServicesHomePage())]),
          StatefulShellBranch(routes: [GoRoute(path: '/profile', builder: (_, _) => const ProfilePage())]),
          StatefulShellBranch(routes: [GoRoute(path: '/admin', builder: (_, _) => const AdminHomePage())]),
        ],
      ),
      GoRoute(path: '/profile/edit', builder: (_, _) => const EditProfilePage()),
      GoRoute(
        path: '/notifications',
        builder: (context, _) => BlocProvider(
          create: (context) => NotificationSettingsCubit(context.read<NotificationRepository>())..load(),
          child: const NotificationSettingsPage(),
        ),
      ),
      ...miniAppRoutes,
    ],
  );
}

class _AppShell extends StatelessWidget {
  const _AppShell({required this.shell});

  final StatefulNavigationShell shell;

  @override
  Widget build(BuildContext context) {
    final admin = context.isAdmin;
    final index = !admin && shell.currentIndex > 2 ? 0 : shell.currentIndex;
    return Scaffold(
      body: shell,
      bottomNavigationBar: NavigationBar(
        selectedIndex: index,
        onDestinationSelected: (i) => shell.goBranch(i, initialLocation: i == shell.currentIndex),
        destinations: [
          const NavigationDestination(icon: Icon(Icons.home_outlined), selectedIcon: Icon(Icons.home_rounded), label: 'Trang chủ'),
          const NavigationDestination(icon: Icon(Icons.apps_outlined), selectedIcon: Icon(Icons.apps_rounded), label: 'Dịch vụ'),
          const NavigationDestination(icon: Icon(Icons.person_outline), selectedIcon: Icon(Icons.person_rounded), label: 'Tài khoản'),
          if (admin) const NavigationDestination(icon: Icon(Icons.admin_panel_settings_outlined), selectedIcon: Icon(Icons.admin_panel_settings), label: 'Quản trị'),
        ],
      ),
    );
  }
}

class _SplashPage extends StatelessWidget {
  const _SplashPage();

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: HeroBackdrop(
        child: SizedBox.expand(
          child: Center(
            child: TweenAnimationBuilder<double>(
              tween: Tween(begin: 0.85, end: 1),
              duration: const Duration(milliseconds: 600),
              curve: Curves.easeOutBack,
              builder: (context, v, child) => Opacity(opacity: ((v - 0.85) / 0.15).clamp(0, 1), child: Transform.scale(scale: v, child: child)),
              child: const KeikiLogo(size: 96, onDark: true),
            ),
          ),
        ),
      ),
    );
  }
}

class _StreamListenable extends ChangeNotifier {
  _StreamListenable(Stream<dynamic> stream) {
    _sub = stream.listen((_) => notifyListeners());
  }

  late final StreamSubscription<dynamic> _sub;

  @override
  void dispose() {
    _sub.cancel();
    super.dispose();
  }
}
