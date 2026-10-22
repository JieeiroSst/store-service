import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:go_router/go_router.dart';

import '../../app/role.dart';

import '../apps/apps.dart';
import 'view/operation_page.dart';
import 'view/service_page.dart';

final servicesRoutes = <RouteBase>[
  GoRoute(
    path: '/services/:service',
    builder: (context, state) {
      final id = state.pathParameters['service']!;
      final app = serviceApps[id];
      final allowed = !adminOnlyApps.contains(id) || context.read<RoleCubit>().state.isAdmin;
      if (app != null && allowed && !state.uri.queryParameters.containsKey('all')) return app();
      return ServicePage(serviceId: id);
    },
    routes: [
      GoRoute(
        path: ':operation',
        builder: (_, state) => OperationPage(
          serviceId: state.pathParameters['service']!,
          operationName: state.pathParameters['operation']!,
        ),
      ),
    ],
  ),
];
