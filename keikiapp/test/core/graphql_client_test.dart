import 'dart:convert';

import 'package:flutter_test/flutter_test.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';
import 'package:keikiapp/core/network/graphql_client.dart';
import 'package:keikiapp/core/network/graphql_exception.dart';
import 'package:keikiapp/core/storage/token_storage.dart';

import '../helpers.dart';

http.Response _json(Object body, [int status = 200]) =>
    http.Response(jsonEncode(body), status, headers: {'content-type': 'application/json'});

void main() {
  test('sends bearer token and device id', () async {
    late http.Request captured;
    final client = GraphQLClient(
      endpoint: 'http://gw/query',
      storage: MemoryTokenStorage(validTokens()),
      httpClient: MockClient((req) async {
        captured = req;
        return _json({'data': {'me': {'id': 1}}});
      }),
    );

    final data = await client.query('query { me { id } }');

    expect(data['me'], {'id': 1});
    expect(captured.headers['Authorization'], 'Bearer access-1');
    expect(captured.headers['X-Device-Id'], 'test-device');
  });

  test('public calls never send Authorization', () async {
    late http.Request captured;
    final client = GraphQLClient(
      endpoint: 'http://gw/query',
      storage: MemoryTokenStorage(validTokens()),
      httpClient: MockClient((req) async {
        captured = req;
        return _json({'data': {'foods': []}});
      }),
    );

    await client.public('query { foods { id } }');

    expect(captured.headers.containsKey('Authorization'), isFalse);
  });

  test('refreshes once on 401 and retries with the new token', () async {
    final storage = MemoryTokenStorage(validTokens());
    final auths = <String?>[];
    final client = GraphQLClient(
      endpoint: 'http://gw/query',
      storage: storage,
      httpClient: MockClient((req) async {
        final body = jsonDecode(req.body) as Map<String, dynamic>;
        auths.add(req.headers['Authorization']);
        if ((body['query'] as String).contains('refreshToken(')) {
          expect(body['variables'], {'refreshToken': 'refresh-1'});
          return _json({
            'data': {
              'refreshToken': {'accessToken': 'access-2', 'refreshToken': 'refresh-2', 'expiresIn': 900},
            },
          });
        }
        if (req.headers['Authorization'] == 'Bearer access-1') {
          return _json({
            'errors': [
              {'message': 'invalid or expired token', 'extensions': {'code': 'UNAUTHENTICATED'}},
            ],
          }, 401);
        }
        return _json({'data': {'ok': true}});
      }),
    );

    final data = await client.query('query { ok }');

    expect(data['ok'], isTrue);
    expect(auths, ['Bearer access-1', null, 'Bearer access-2']);
    expect(storage.tokens!.accessToken, 'access-2');
    expect(storage.tokens!.refreshToken, 'refresh-2');
  });

  test('clears session and notifies when refresh fails', () async {
    final storage = MemoryTokenStorage(validTokens());
    final client = GraphQLClient(
      endpoint: 'http://gw/query',
      storage: storage,
      httpClient: MockClient((req) async => _json({
            'errors': [
              {'message': 'expired', 'extensions': {'code': 'UNAUTHENTICATED'}},
            ],
          }, 401)),
    );
    final expired = expectLater(client.sessionExpired, emits(null));

    await expectLater(client.query('query { ok }'), throwsA(isA<GraphQLException>()));
    await expired;
    expect(storage.tokens, isNull);
    expect(storage.clears, 1);
  });

  test('refreshes proactively when the access token is expired', () async {
    final storage = MemoryTokenStorage(
      AuthTokens(accessToken: 'old', refreshToken: 'refresh-1', expiresAt: DateTime.now().subtract(const Duration(minutes: 1))),
    );
    final auths = <String?>[];
    final client = GraphQLClient(
      endpoint: 'http://gw/query',
      storage: storage,
      httpClient: MockClient((req) async {
        auths.add(req.headers['Authorization']);
        if (req.body.contains('refreshToken(')) {
          return _json({
            'data': {
              'refreshToken': {'accessToken': 'fresh', 'refreshToken': 'refresh-2', 'expiresIn': 900},
            },
          });
        }
        return _json({'data': {'ok': true}});
      }),
    );

    await client.query('query { ok }');

    expect(auths, [null, 'Bearer fresh']);
  });

  test('maps GraphQL error extensions to GraphQLException', () async {
    final client = GraphQLClient(
      endpoint: 'http://gw/query',
      storage: MemoryTokenStorage(validTokens()),
      httpClient: MockClient((_) async => _json({
            'errors': [
              {'message': 'permission denied', 'extensions': {'code': 'FORBIDDEN'}},
            ],
          })),
    );

    await expectLater(
      client.query('query { users { total } }'),
      throwsA(isA<GraphQLException>().having((e) => e.isForbidden, 'isForbidden', isTrue)),
    );
  });

  test('builds media and signaling URLs next to the GraphQL endpoint', () {
    for (final (endpoint, media, ws) in [
      ('http://gw:8080/query', 'http://gw:8080/media/video-service/api/videos/v1/stream', 'ws://gw:8080/rtc/ws?room_id=team+a'),
      ('https://api.keiki.vn/gateway/query', 'https://api.keiki.vn/gateway/media/video-service/api/videos/v1/stream', 'wss://api.keiki.vn/gateway/rtc/ws?room_id=team+a'),
    ]) {
      final client = GraphQLClient(endpoint: endpoint, storage: MemoryTokenStorage(), httpClient: MockClient((_) async => _json({})));
      expect(client.mediaUri('video-service', '/api/videos/v1/stream').toString(), media);
      expect(client.signalingUri('team a').toString(), ws);
    }
  });

  test('accessToken refreshes an expired session', () async {
    final storage = MemoryTokenStorage(AuthTokens(accessToken: 'old', refreshToken: 'r1', expiresAt: DateTime.now().subtract(const Duration(minutes: 1))));
    final client = GraphQLClient(
      endpoint: 'http://gw/query',
      storage: storage,
      httpClient: MockClient((_) async => _json({
            'data': {
              'refreshToken': {'accessToken': 'fresh', 'refreshToken': 'r2', 'expiresIn': 3600},
            },
          })),
    );
    expect(await client.accessToken(), 'fresh');
  });
}
