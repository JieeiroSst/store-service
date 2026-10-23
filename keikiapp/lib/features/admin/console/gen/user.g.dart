// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

// ignore_for_file: unused_import

import '../../../apps/kit/kit.dart';
import '../../../../gen/models.g.dart';
import '../admin_op.dart';

final userAdminOps = <AdminOp>[
  AdminOp(
    'users',
    resource: 'user',
    verb: '',
    fields: [
      KField('input.username', 'Username'),
      KField('input.email', 'Email', kind: FieldKind.email),
      KField('input.page', 'Page', kind: FieldKind.number),
      KField('input.limit', 'Limit', kind: FieldKind.number),
    ],
    run: (api, v) async => adminPlain(await api.user.users(input: v.any('input.') ? FindUsersInput(username: v.str('input.username'), email: v.str('input.email'), page: v.integer('input.page'), limit: v.integer('input.limit')) : null)),
  ),
  AdminOp(
    'userRole',
    resource: 'user role',
    verb: '',
    fields: [
      KField('id', 'ID', kind: FieldKind.number, required: true),
    ],
    run: (api, v) async => adminPlain(await api.user.userRole(id: v.need(v.integer('id'), 'id'))),
  ),
  AdminOp(
    'userRoles',
    resource: 'user role',
    verb: '',
    fields: [
      KField('page', 'Page', kind: FieldKind.number),
      KField('limit', 'Limit', kind: FieldKind.number),
    ],
    run: (api, v) async => adminPlain(await api.user.userRoles(page: v.integer('page'), limit: v.integer('limit'))),
  ),
  AdminOp(
    'adminUpdateProfile',
    resource: 'update profile',
    verb: '',
    mutation: true,
    fields: [
      KField('input.id', 'ID', kind: FieldKind.number, required: true),
      KField('input.name', 'Name'),
      KField('input.email', 'Email', kind: FieldKind.email),
      KField('input.phone', 'Phone', kind: FieldKind.phone),
      KField('input.address', 'Address'),
      KField('input.sex', 'Sex'),
    ],
    run: (api, v) async => adminPlain(await api.user.adminUpdateProfile(input: AdminUpdateProfileInput(id: v.need(v.integer('input.id'), 'input.id'), name_: v.str('input.name'), email: v.str('input.email'), phone: v.str('input.phone'), address: v.str('input.address'), sex: v.str('input.sex')))),
  ),
  AdminOp(
    'createUserRole',
    resource: 'user role',
    verb: 'create',
    mutation: true,
    fields: [
      KField('input.id', 'ID', kind: FieldKind.number),
      KField('input.name', 'Name', required: true),
    ],
    run: (api, v) async => adminPlain(await api.user.createUserRole(input: UserRoleInput(id: v.integer('input.id'), name_: v.need(v.str('input.name'), 'input.name')))),
  ),
  AdminOp(
    'updateUserRole',
    resource: 'user role',
    verb: 'update',
    mutation: true,
    fields: [
      KField('id', 'ID', kind: FieldKind.number, required: true),
      KField('name', 'Name', required: true),
    ],
    run: (api, v) async => adminPlain(await api.user.updateUserRole(id: v.need(v.integer('id'), 'id'), name_: v.need(v.str('name'), 'name'))),
  ),
  AdminOp(
    'deleteUserRole',
    resource: 'user role',
    verb: 'delete',
    mutation: true,
    destructive: true,
    fields: [
      KField('id', 'ID', kind: FieldKind.number, required: true),
    ],
    run: (api, v) async => adminPlain(await api.user.deleteUserRole(id: v.need(v.integer('id'), 'id'))),
  ),
  AdminOp(
    'assignUserRole',
    resource: 'user role',
    verb: 'assign',
    mutation: true,
    fields: [
      KField('input.userId', 'User ID', kind: FieldKind.number, required: true),
      KField('input.roleId', 'Role ID', kind: FieldKind.number, required: true),
    ],
    run: (api, v) async => adminPlain(await api.user.assignUserRole(input: AssignUserRoleInput(userId: v.need(v.integer('input.userId'), 'input.userId'), roleId: v.need(v.integer('input.roleId'), 'input.roleId')))),
  ),
  AdminOp(
    'addUserRole',
    resource: 'user role',
    verb: 'add',
    mutation: true,
    fields: [
      KField('input.userId', 'User ID', kind: FieldKind.number, required: true),
      KField('input.roleId', 'Role ID', kind: FieldKind.number, required: true),
    ],
    run: (api, v) async => adminPlain(await api.user.addUserRole(input: AssignUserRoleInput(userId: v.need(v.integer('input.userId'), 'input.userId'), roleId: v.need(v.integer('input.roleId'), 'input.roleId')))),
  ),
  AdminOp(
    'removeUserRoles',
    resource: 'user role',
    verb: 'remove',
    mutation: true,
    destructive: true,
    fields: [
      KField('userId', 'User ID', kind: FieldKind.number, required: true),
    ],
    run: (api, v) async => adminPlain(await api.user.removeUserRoles(userId: v.need(v.integer('userId'), 'userId'))),
  ),
];
