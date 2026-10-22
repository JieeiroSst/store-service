import 'package:equatable/equatable.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

import '../../../core/network/graphql_client.dart';
import '../../../core/utils/load_status.dart';
import '../data/permissions.dart';
import '../data/schema_catalog.dart';

class ServicesState extends Equatable {
  const ServicesState({
    this.status = LoadStatus.initial,
    this.catalog,
    this.permissions = const UserPermissions(),
    this.recent = const [],
    this.error,
  });

  final LoadStatus status;
  final SchemaCatalog? catalog;
  final UserPermissions permissions;
  final List<String> recent;
  final String? error;

  List<GqlOperation> visibleOperations(GqlService service) =>
      service.operations.where((op) => op.name != 'health' && op.name != 'ping' && permissions.canRun(op)).toList();

  List<GqlService> get visibleServices =>
      [for (final s in catalog?.services ?? const <GqlService>[]) if (visibleOperations(s).isNotEmpty) s];

  ServicesState copyWith({
    LoadStatus? status,
    SchemaCatalog? catalog,
    UserPermissions? permissions,
    List<String>? recent,
    String? error,
  }) =>
      ServicesState(
        status: status ?? this.status,
        catalog: catalog ?? this.catalog,
        permissions: permissions ?? this.permissions,
        recent: recent ?? this.recent,
        error: error,
      );

  @override
  List<Object?> get props => [status, catalog, permissions, recent, error];
}

class ServicesCubit extends Cubit<ServicesState> {
  ServicesCubit(this._client, {Future<SchemaCatalog> Function()? loader})
      : _loader = loader ?? SchemaCatalog.load,
        super(const ServicesState());

  final GraphQLClient _client;
  final Future<SchemaCatalog> Function() _loader;

  GraphQLClient get client => _client;

  Future<void> load() async {
    if (state.catalog != null) return;
    emit(state.copyWith(status: LoadStatus.loading));
    try {
      emit(state.copyWith(status: LoadStatus.success, catalog: await _loader()));
    } catch (e) {
      emit(state.copyWith(status: LoadStatus.failure, error: e.toString()));
    }
  }

  Future<void> loadPermissions() async {
    try {
      final data = await _client.query('query { myPermissions { object action } myRoles { effectiveRoles } }');
      final grants = [
        for (final p in (data['myPermissions'] as List? ?? const []).cast<Map<String, dynamic>>())
          Grant(p['object'] as String, p['action'] as String),
      ];
      final roles = [...((data['myRoles'] as Map?)?['effectiveRoles'] as List? ?? const []).cast<String>()];
      emit(state.copyWith(permissions: UserPermissions(grants: grants, roles: roles)));
    } catch (_) {
      emit(state.copyWith(permissions: const UserPermissions()));
    }
  }

  void clearPermissions() => emit(state.copyWith(permissions: const UserPermissions(), recent: const []));

  void visited(String serviceId) {
    final next = [serviceId, ...state.recent.where((r) => r != serviceId)].take(8).toList();
    emit(state.copyWith(recent: next));
  }
}
