import 'package:equatable/equatable.dart';
import 'package:flutter/widgets.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

enum AppRole { customer, admin }

const adminRoleNames = {'admin', 'super_admin'};

const adminOnlyApps = {'manage'};

class RoleState extends Equatable {
  const RoleState({this.serverAdmin = false, this.previewCustomer = false});

  final bool serverAdmin;
  final bool previewCustomer;

  AppRole get role => serverAdmin && !previewCustomer ? AppRole.admin : AppRole.customer;

  bool get isAdmin => role == AppRole.admin;

  @override
  List<Object?> get props => [serverAdmin, previewCustomer];
}

class RoleCubit extends Cubit<RoleState> {
  RoleCubit() : super(const RoleState());

  final _sources = <String, bool>{};

  void report(String source, Iterable<String> roles) {
    _sources[source] = roles.any(adminRoleNames.contains);
    final admin = _sources.values.any((v) => v);
    emit(RoleState(serverAdmin: admin, previewCustomer: admin && state.previewCustomer));
  }

  void reset() {
    _sources.clear();
    emit(const RoleState());
  }

  void previewAsCustomer(bool on) => emit(RoleState(serverAdmin: state.serverAdmin, previewCustomer: state.serverAdmin && on));
}

extension RoleContext on BuildContext {
  bool get isAdmin => select<RoleCubit, bool>((c) => c.state.isAdmin);

  bool get isAdminNow => read<RoleCubit>().state.isAdmin;
}

class AdminOnly extends StatelessWidget {
  const AdminOnly({super.key, required this.child, this.fallback = const SizedBox.shrink()});

  final Widget child;
  final Widget fallback;

  @override
  Widget build(BuildContext context) {
    final admin = context.select<RoleCubit, bool>((c) => c.state.isAdmin);
    return admin ? child : fallback;
  }
}
