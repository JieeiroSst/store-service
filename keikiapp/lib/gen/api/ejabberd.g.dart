// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

import '../gateway_executor.dart';
import '../models.g.dart';

class EjabberdApi {
  const EjabberdApi(this._executor);

  final GatewayExecutor _executor;

  Future<String?> xmppStatus() async {
    final data = await _executor.query('query xmppStatus { xmppStatus }'
    );
    return data['xmppStatus']?.toString();
  }

  Future<int?> xmppCheckAccount({required String user, required String host}) async {
    final data = await _executor.query('query xmppCheckAccount(\$user: String!, \$host: String!) { xmppCheckAccount(user: \$user, host: \$host) }', variables: {
      'user': user,
      'host': host,
    }
    );
    return gqlInt(data['xmppCheckAccount']);
  }

  Future<List<String>?> xmppRegisteredUsers({required String host}) async {
    final data = await _executor.query('query xmppRegisteredUsers(\$host: String!) { xmppRegisteredUsers(host: \$host) }', variables: {
      'host': host,
    }
    );
    return (data['xmppRegisteredUsers'] as List?)?.map((e) => (e?.toString())!).toList();
  }

  Future<List<String>?> xmppRegisteredVhosts() async {
    final data = await _executor.query('query xmppRegisteredVhosts { xmppRegisteredVhosts }'
    );
    return (data['xmppRegisteredVhosts'] as List?)?.map((e) => (e?.toString())!).toList();
  }

  Future<List<String>?> xmppConnectedUsers() async {
    final data = await _executor.query('query xmppConnectedUsers { xmppConnectedUsers }'
    );
    return (data['xmppConnectedUsers'] as List?)?.map((e) => (e?.toString())!).toList();
  }

  Future<Object?> xmppConnectedUsersNumber() async {
    final data = await _executor.query('query xmppConnectedUsersNumber { xmppConnectedUsersNumber }'
    );
    return data['xmppConnectedUsersNumber'];
  }

  Future<List<String>?> xmppConnectedUsersVhost({required String host}) async {
    final data = await _executor.query('query xmppConnectedUsersVhost(\$host: String!) { xmppConnectedUsersVhost(host: \$host) }', variables: {
      'host': host,
    }
    );
    return (data['xmppConnectedUsersVhost'] as List?)?.map((e) => (e?.toString())!).toList();
  }

  Future<List<String>?> xmppUserResources({required String user, required String host}) async {
    final data = await _executor.query('query xmppUserResources(\$user: String!, \$host: String!) { xmppUserResources(user: \$user, host: \$host) }', variables: {
      'user': user,
      'host': host,
    }
    );
    return (data['xmppUserResources'] as List?)?.map((e) => (e?.toString())!).toList();
  }

  Future<Object?> xmppNumResources({required String user, required String host}) async {
    final data = await _executor.query('query xmppNumResources(\$user: String!, \$host: String!) { xmppNumResources(user: \$user, host: \$host) }', variables: {
      'user': user,
      'host': host,
    }
    );
    return data['xmppNumResources'];
  }

  Future<Object?> xmppRoster({required String user, required String host}) async {
    final data = await _executor.query('query xmppRoster(\$user: String!, \$host: String!) { xmppRoster(user: \$user, host: \$host) }', variables: {
      'user': user,
      'host': host,
    }
    );
    return data['xmppRoster'];
  }

  Future<Object?> xmppOfflineCount({required String user, required String host}) async {
    final data = await _executor.query('query xmppOfflineCount(\$user: String!, \$host: String!) { xmppOfflineCount(user: \$user, host: \$host) }', variables: {
      'user': user,
      'host': host,
    }
    );
    return data['xmppOfflineCount'];
  }

  Future<Object?> xmppStats({required String name_}) async {
    final data = await _executor.query('query xmppStats(\$name: String!) { xmppStats(name: \$name) }', variables: {
      'name': name_,
    }
    );
    return data['xmppStats'];
  }

  Future<Object?> xmppCommand({required String command, Map<String, dynamic>? args}) async {
    final data = await _executor.query('query xmppCommand(\$command: String!, \$args: Map) { xmppCommand(command: \$command, args: \$args) }', variables: {
      'command': command,
      'args': ?args,
    }
    );
    return data['xmppCommand'];
  }

  Future<String?> xmppRegister({required String user, required String host, required String password}) async {
    final data = await _executor.mutation('mutation xmppRegister(\$user: String!, \$host: String!, \$password: String!) { xmppRegister(user: \$user, host: \$host, password: \$password) }', variables: {
      'user': user,
      'host': host,
      'password': password,
    }
    );
    return data['xmppRegister']?.toString();
  }

  Future<String?> xmppUnregister({required String user, required String host}) async {
    final data = await _executor.mutation('mutation xmppUnregister(\$user: String!, \$host: String!) { xmppUnregister(user: \$user, host: \$host) }', variables: {
      'user': user,
      'host': host,
    }
    );
    return data['xmppUnregister']?.toString();
  }

  Future<int?> xmppChangePassword({required String user, required String host, required String newpass}) async {
    final data = await _executor.mutation('mutation xmppChangePassword(\$user: String!, \$host: String!, \$newpass: String!) { xmppChangePassword(user: \$user, host: \$host, newpass: \$newpass) }', variables: {
      'user': user,
      'host': host,
      'newpass': newpass,
    }
    );
    return gqlInt(data['xmppChangePassword']);
  }

  Future<Object?> xmppKickUser({required String user, required String host}) async {
    final data = await _executor.mutation('mutation xmppKickUser(\$user: String!, \$host: String!) { xmppKickUser(user: \$user, host: \$host) }', variables: {
      'user': user,
      'host': host,
    }
    );
    return data['xmppKickUser'];
  }

  Future<int?> xmppSendMessage({required String type_, required String from, required String to, String? subject, required String body}) async {
    final data = await _executor.mutation('mutation xmppSendMessage(\$type: String!, \$from: String!, \$to: String!, \$subject: String, \$body: String!) { xmppSendMessage(type: \$type, from: \$from, to: \$to, subject: \$subject, body: \$body) }', variables: {
      'type': type_,
      'from': from,
      'to': to,
      'subject': ?subject,
      'body': body,
    }
    );
    return gqlInt(data['xmppSendMessage']);
  }

  Future<int?> xmppSendStanza({required String from, required String to, required String stanza}) async {
    final data = await _executor.mutation('mutation xmppSendStanza(\$from: String!, \$to: String!, \$stanza: String!) { xmppSendStanza(from: \$from, to: \$to, stanza: \$stanza) }', variables: {
      'from': from,
      'to': to,
      'stanza': stanza,
    }
    );
    return gqlInt(data['xmppSendStanza']);
  }

  Future<int?> xmppAddRosterItem({required String localuser, required String localhost, required String user, required String host, String? nick, String? group, String? subs}) async {
    final data = await _executor.mutation('mutation xmppAddRosterItem(\$localuser: String!, \$localhost: String!, \$user: String!, \$host: String!, \$nick: String, \$group: String, \$subs: String) { xmppAddRosterItem(localuser: \$localuser, localhost: \$localhost, user: \$user, host: \$host, nick: \$nick, group: \$group, subs: \$subs) }', variables: {
      'localuser': localuser,
      'localhost': localhost,
      'user': user,
      'host': host,
      'nick': ?nick,
      'group': ?group,
      'subs': ?subs,
    }
    );
    return gqlInt(data['xmppAddRosterItem']);
  }

  Future<int?> xmppDeleteRosterItem({required String localuser, required String localhost, required String user, required String host}) async {
    final data = await _executor.mutation('mutation xmppDeleteRosterItem(\$localuser: String!, \$localhost: String!, \$user: String!, \$host: String!) { xmppDeleteRosterItem(localuser: \$localuser, localhost: \$localhost, user: \$user, host: \$host) }', variables: {
      'localuser': localuser,
      'localhost': localhost,
      'user': user,
      'host': host,
    }
    );
    return gqlInt(data['xmppDeleteRosterItem']);
  }

  Future<Object?> xmppDeleteExpiredMessages() async {
    final data = await _executor.mutation('mutation xmppDeleteExpiredMessages { xmppDeleteExpiredMessages }'
    );
    return data['xmppDeleteExpiredMessages'];
  }

}
