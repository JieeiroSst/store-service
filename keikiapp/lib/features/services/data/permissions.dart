import 'package:equatable/equatable.dart';

import 'schema_catalog.dart';

class Grant extends Equatable {
  const Grant(this.object, this.action);

  final String object;
  final String action;

  static bool keyMatch(String value, String pattern) {
    final i = pattern.indexOf('*');
    if (i == -1) return value == pattern;
    final prefix = pattern.substring(0, i);
    return value.length >= prefix.length && value.startsWith(prefix);
  }

  static bool regexMatch(String value, String pattern) {
    try {
      return RegExp(pattern).hasMatch(value);
    } on FormatException {
      return false;
    }
  }

  bool allows(String object, String action) => keyMatch(object, this.object) && regexMatch(action, this.action);

  @override
  List<Object?> get props => [object, action];
}

class UserPermissions extends Equatable {
  const UserPermissions({this.grants = const [], this.roles = const []});

  final List<Grant> grants;
  final List<String> roles;

  bool get isAdmin => roles.any((r) => r == 'admin' || r == 'super_admin');

  bool canRun(GqlOperation op) {
    if (op.access != OperationAccess.admin) return true;
    final object = op.permissionObject;
    final action = op.permissionAction;
    if (object == null || action == null) return isAdmin;
    return grants.any((g) => g.allows(object, action));
  }

  @override
  List<Object?> get props => [grants, roles];
}
