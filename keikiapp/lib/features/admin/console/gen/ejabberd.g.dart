// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

// ignore_for_file: unused_import

import '../../../apps/kit/kit.dart';
import '../../../../gen/models.g.dart';
import '../admin_op.dart';

final ejabberdAdminOps = <AdminOp>[
  AdminOp(
    'xmppStatus',
    resource: 'general',
    verb: '',
    run: (api, v) async => adminPlain(await api.ejabberd.xmppStatus()),
  ),
  AdminOp(
    'xmppCheckAccount',
    resource: 'general',
    verb: 'check',
    fields: [
      KField('user', 'User', required: true),
      KField('host', 'Host', required: true),
    ],
    run: (api, v) async => adminPlain(await api.ejabberd.xmppCheckAccount(user: v.need(v.str('user'), 'user'), host: v.need(v.str('host'), 'host'))),
  ),
  AdminOp(
    'xmppRegisteredUsers',
    resource: 'registered',
    verb: '',
    fields: [
      KField('host', 'Host', required: true),
    ],
    run: (api, v) async => adminPlain(await api.ejabberd.xmppRegisteredUsers(host: v.need(v.str('host'), 'host'))),
  ),
  AdminOp(
    'xmppRegisteredVhosts',
    resource: 'registered',
    verb: '',
    run: (api, v) async => adminPlain(await api.ejabberd.xmppRegisteredVhosts()),
  ),
  AdminOp(
    'xmppConnectedUsers',
    resource: 'connected',
    verb: '',
    run: (api, v) async => adminPlain(await api.ejabberd.xmppConnectedUsers()),
  ),
  AdminOp(
    'xmppConnectedUsersNumber',
    resource: 'connected',
    verb: '',
    run: (api, v) async => adminPlain(await api.ejabberd.xmppConnectedUsersNumber()),
  ),
  AdminOp(
    'xmppConnectedUsersVhost',
    resource: 'connected',
    verb: '',
    fields: [
      KField('host', 'Host', required: true),
    ],
    run: (api, v) async => adminPlain(await api.ejabberd.xmppConnectedUsersVhost(host: v.need(v.str('host'), 'host'))),
  ),
  AdminOp(
    'xmppUserResources',
    resource: 'general',
    verb: '',
    fields: [
      KField('user', 'User', required: true),
      KField('host', 'Host', required: true),
    ],
    run: (api, v) async => adminPlain(await api.ejabberd.xmppUserResources(user: v.need(v.str('user'), 'user'), host: v.need(v.str('host'), 'host'))),
  ),
  AdminOp(
    'xmppNumResources',
    resource: 'general',
    verb: '',
    fields: [
      KField('user', 'User', required: true),
      KField('host', 'Host', required: true),
    ],
    run: (api, v) async => adminPlain(await api.ejabberd.xmppNumResources(user: v.need(v.str('user'), 'user'), host: v.need(v.str('host'), 'host'))),
  ),
  AdminOp(
    'xmppRoster',
    resource: 'roster',
    verb: '',
    fields: [
      KField('user', 'User', required: true),
      KField('host', 'Host', required: true),
    ],
    run: (api, v) async => adminPlain(await api.ejabberd.xmppRoster(user: v.need(v.str('user'), 'user'), host: v.need(v.str('host'), 'host'))),
  ),
  AdminOp(
    'xmppOfflineCount',
    resource: 'general',
    verb: '',
    fields: [
      KField('user', 'User', required: true),
      KField('host', 'Host', required: true),
    ],
    run: (api, v) async => adminPlain(await api.ejabberd.xmppOfflineCount(user: v.need(v.str('user'), 'user'), host: v.need(v.str('host'), 'host'))),
  ),
  AdminOp(
    'xmppStats',
    resource: 'general',
    verb: '',
    fields: [
      KField('name', 'Name', required: true),
    ],
    run: (api, v) async => adminPlain(await api.ejabberd.xmppStats(name_: v.need(v.str('name'), 'name'))),
  ),
  AdminOp(
    'xmppCommand',
    resource: 'general',
    verb: '',
    fields: [
      KField('command', 'Command', required: true),
      KField('args', 'Args', kind: FieldKind.multiline, hint: 'JSON, ví dụ {"key": "value"}'),
    ],
    run: (api, v) async => adminPlain(await api.ejabberd.xmppCommand(command: v.need(v.str('command'), 'command'), args: v.jsonMap('args'))),
  ),
  AdminOp(
    'xmppRegister',
    resource: 'general',
    verb: 'register',
    mutation: true,
    fields: [
      KField('user', 'User', required: true),
      KField('host', 'Host', required: true),
      KField('password', 'Password', kind: FieldKind.password, required: true),
    ],
    run: (api, v) async => adminPlain(await api.ejabberd.xmppRegister(user: v.need(v.str('user'), 'user'), host: v.need(v.str('host'), 'host'), password: v.need(v.str('password'), 'password'))),
  ),
  AdminOp(
    'xmppUnregister',
    resource: 'general',
    verb: 'unregister',
    mutation: true,
    destructive: true,
    fields: [
      KField('user', 'User', required: true),
      KField('host', 'Host', required: true),
    ],
    run: (api, v) async => adminPlain(await api.ejabberd.xmppUnregister(user: v.need(v.str('user'), 'user'), host: v.need(v.str('host'), 'host'))),
  ),
  AdminOp(
    'xmppChangePassword',
    resource: 'general',
    verb: 'change',
    mutation: true,
    fields: [
      KField('user', 'User', required: true),
      KField('host', 'Host', required: true),
      KField('newpass', 'Newpass', required: true),
    ],
    run: (api, v) async => adminPlain(await api.ejabberd.xmppChangePassword(user: v.need(v.str('user'), 'user'), host: v.need(v.str('host'), 'host'), newpass: v.need(v.str('newpass'), 'newpass'))),
  ),
  AdminOp(
    'xmppKickUser',
    resource: 'general',
    verb: '',
    mutation: true,
    fields: [
      KField('user', 'User', required: true),
      KField('host', 'Host', required: true),
    ],
    run: (api, v) async => adminPlain(await api.ejabberd.xmppKickUser(user: v.need(v.str('user'), 'user'), host: v.need(v.str('host'), 'host'))),
  ),
  AdminOp(
    'xmppSendMessage',
    resource: 'general',
    verb: 'send',
    mutation: true,
    fields: [
      KField('type', 'Type', required: true),
      KField('from', 'From', required: true, hint: 'ISO 8601, ví dụ 2026-10-22T09:00:00Z'),
      KField('to', 'To', required: true, hint: 'ISO 8601, ví dụ 2026-10-22T09:00:00Z'),
      KField('subject', 'Subject'),
      KField('body', 'Body', kind: FieldKind.multiline, required: true),
    ],
    run: (api, v) async => adminPlain(await api.ejabberd.xmppSendMessage(type_: v.need(v.str('type'), 'type'), from: v.need(v.str('from'), 'from'), to: v.need(v.str('to'), 'to'), subject: v.str('subject'), body: v.need(v.str('body'), 'body'))),
  ),
  AdminOp(
    'xmppSendStanza',
    resource: 'general',
    verb: 'send',
    mutation: true,
    fields: [
      KField('from', 'From', required: true, hint: 'ISO 8601, ví dụ 2026-10-22T09:00:00Z'),
      KField('to', 'To', required: true, hint: 'ISO 8601, ví dụ 2026-10-22T09:00:00Z'),
      KField('stanza', 'Stanza', required: true),
    ],
    run: (api, v) async => adminPlain(await api.ejabberd.xmppSendStanza(from: v.need(v.str('from'), 'from'), to: v.need(v.str('to'), 'to'), stanza: v.need(v.str('stanza'), 'stanza'))),
  ),
  AdminOp(
    'xmppAddRosterItem',
    resource: 'roster',
    verb: 'add',
    mutation: true,
    fields: [
      KField('localuser', 'Localuser', required: true),
      KField('localhost', 'Localhost', required: true),
      KField('user', 'User', required: true),
      KField('host', 'Host', required: true),
      KField('nick', 'Nick'),
      KField('group', 'Group'),
      KField('subs', 'Subs'),
    ],
    run: (api, v) async => adminPlain(await api.ejabberd.xmppAddRosterItem(localuser: v.need(v.str('localuser'), 'localuser'), localhost: v.need(v.str('localhost'), 'localhost'), user: v.need(v.str('user'), 'user'), host: v.need(v.str('host'), 'host'), nick: v.str('nick'), group: v.str('group'), subs: v.str('subs'))),
  ),
  AdminOp(
    'xmppDeleteRosterItem',
    resource: 'roster',
    verb: 'delete',
    mutation: true,
    destructive: true,
    fields: [
      KField('localuser', 'Localuser', required: true),
      KField('localhost', 'Localhost', required: true),
      KField('user', 'User', required: true),
      KField('host', 'Host', required: true),
    ],
    run: (api, v) async => adminPlain(await api.ejabberd.xmppDeleteRosterItem(localuser: v.need(v.str('localuser'), 'localuser'), localhost: v.need(v.str('localhost'), 'localhost'), user: v.need(v.str('user'), 'user'), host: v.need(v.str('host'), 'host'))),
  ),
  AdminOp(
    'xmppDeleteExpiredMessages',
    resource: 'general',
    verb: 'delete',
    mutation: true,
    destructive: true,
    run: (api, v) async => adminPlain(await api.ejabberd.xmppDeleteExpiredMessages()),
  ),
];
