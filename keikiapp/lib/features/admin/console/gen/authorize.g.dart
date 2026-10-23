// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

// ignore_for_file: unused_import

import '../../../apps/kit/kit.dart';
import '../../../../gen/models.g.dart';
import '../admin_op.dart';

final authorizeAdminOps = <AdminOp>[
  AdminOp(
    'authzRoles',
    resource: 'general',
    verb: '',
    run: (api, v) async => adminPlain(await api.authorize.authzRoles()),
  ),
  AdminOp(
    'authzUserRoles',
    resource: 'general',
    verb: '',
    fields: [
      KField('userId', 'User ID', required: true),
    ],
    run: (api, v) async => adminPlain(await api.authorize.authzUserRoles(userId: v.need(v.str('userId'), 'userId'))),
  ),
  AdminOp(
    'authzUserPermissions',
    resource: 'general',
    verb: '',
    fields: [
      KField('userId', 'User ID', required: true),
    ],
    run: (api, v) async => adminPlain(await api.authorize.authzUserPermissions(userId: v.need(v.str('userId'), 'userId'))),
  ),
  AdminOp(
    'casbinRules',
    resource: 'casbin rule',
    verb: '',
    fields: [
      KField('page', 'Page', kind: FieldKind.number),
      KField('limit', 'Limit', kind: FieldKind.number),
      KField('sort', 'Sort'),
    ],
    run: (api, v) async => adminPlain(await api.authorize.casbinRules(page: v.integer('page'), limit: v.integer('limit'), sort: v.str('sort'))),
  ),
  AdminOp(
    'casbinRule',
    resource: 'casbin rule',
    verb: '',
    fields: [
      KField('id', 'ID', kind: FieldKind.number, required: true),
    ],
    run: (api, v) async => adminPlain(await api.authorize.casbinRule(id: v.need(v.integer('id'), 'id'))),
  ),
  AdminOp(
    'authzAssignRoles',
    resource: 'general',
    verb: '',
    mutation: true,
    fields: [
      KField('userId', 'User ID', required: true),
      KField('roles', 'Roles', required: true, hint: 'Nhiều giá trị, cách nhau bởi dấu phẩy'),
    ],
    run: (api, v) async => adminPlain(await api.authorize.authzAssignRoles(userId: v.need(v.str('userId'), 'userId'), roles: v.need(v.strs('roles'), 'roles'))),
  ),
  AdminOp(
    'authzRevokeRoles',
    resource: 'general',
    verb: '',
    mutation: true,
    fields: [
      KField('userId', 'User ID', required: true),
      KField('roles', 'Roles', required: true, hint: 'Nhiều giá trị, cách nhau bởi dấu phẩy'),
    ],
    run: (api, v) async => adminPlain(await api.authorize.authzRevokeRoles(userId: v.need(v.str('userId'), 'userId'), roles: v.need(v.strs('roles'), 'roles'))),
  ),
  AdminOp(
    'authzSetUserRoles',
    resource: 'general',
    verb: '',
    mutation: true,
    fields: [
      KField('userId', 'User ID', required: true),
      KField('roles', 'Roles', required: true, hint: 'Nhiều giá trị, cách nhau bởi dấu phẩy'),
    ],
    run: (api, v) async => adminPlain(await api.authorize.authzSetUserRoles(userId: v.need(v.str('userId'), 'userId'), roles: v.need(v.strs('roles'), 'roles'))),
  ),
  AdminOp(
    'authzRemoveUser',
    resource: 'general',
    verb: '',
    mutation: true,
    fields: [
      KField('userId', 'User ID', required: true),
    ],
    run: (api, v) async => adminPlain(await api.authorize.authzRemoveUser(userId: v.need(v.str('userId'), 'userId'))),
  ),
  AdminOp(
    'createCasbinRule',
    resource: 'casbin rule',
    verb: 'create',
    mutation: true,
    fields: [
      KField('input.ptype', 'Ptype', required: true),
      KField('input.v0', 'V0'),
      KField('input.v1', 'V1'),
      KField('input.v2', 'V2'),
      KField('input.v3', 'V3'),
      KField('input.v4', 'V4'),
      KField('input.v5', 'V5'),
    ],
    run: (api, v) async => adminPlain(await api.authorize.createCasbinRule(input: CasbinRuleInput(ptype: v.need(v.str('input.ptype'), 'input.ptype'), v0: v.str('input.v0'), v1: v.str('input.v1'), v2: v.str('input.v2'), v3: v.str('input.v3'), v4: v.str('input.v4'), v5: v.str('input.v5')))),
  ),
  AdminOp(
    'updateCasbinRule',
    resource: 'casbin rule',
    verb: 'update',
    mutation: true,
    fields: [
      KField('id', 'ID', kind: FieldKind.number, required: true),
      KField('field', 'Field', required: true),
      KField('value', 'Value', required: true),
    ],
    run: (api, v) async => adminPlain(await api.authorize.updateCasbinRule(id: v.need(v.integer('id'), 'id'), field: v.need(v.str('field'), 'field'), value: v.need(v.str('value'), 'value'))),
  ),
  AdminOp(
    'deleteCasbinRule',
    resource: 'casbin rule',
    verb: 'delete',
    mutation: true,
    destructive: true,
    fields: [
      KField('id', 'ID', kind: FieldKind.number, required: true),
    ],
    run: (api, v) async => adminPlain(await api.authorize.deleteCasbinRule(id: v.need(v.integer('id'), 'id'))),
  ),
  AdminOp(
    'createOtp',
    resource: 'general',
    verb: 'create',
    mutation: true,
    fields: [
      KField('userId', 'User ID', required: true),
      KField('username', 'Username', required: true),
    ],
    run: (api, v) async => adminPlain(await api.authorize.createOtp(userId: v.need(v.str('userId'), 'userId'), username: v.need(v.str('username'), 'username'))),
  ),
];
