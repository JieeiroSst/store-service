import 'dart:async';
import 'dart:convert';

import 'package:http/http.dart' as http;
import 'package:http_parser/http_parser.dart';

import '../../gen/gateway_executor.dart';
import '../storage/token_storage.dart';
import 'graphql_exception.dart';

class GraphQLClient implements GatewayExecutor {
  GraphQLClient({required this.endpoint, required this.storage, http.Client? httpClient})
      : _http = httpClient ?? http.Client();

  final String endpoint;
  final TokenStorage storage;
  final http.Client _http;
  final _sessionExpired = StreamController<void>.broadcast();
  Future<AuthTokens?>? _refreshing;

  Stream<void> get sessionExpired => _sessionExpired.stream;

  static const _refreshMutation = r'''
    mutation Refresh($refreshToken: String!) {
      refreshToken(refreshToken: $refreshToken) { accessToken refreshToken expiresIn }
    }
  ''';

  Uri get baseUri {
    final uri = Uri.parse(endpoint);
    final segments = [...uri.pathSegments];
    if (segments.isNotEmpty && segments.last == 'query') segments.removeLast();
    return uri.replace(pathSegments: segments, query: null);
  }

  Uri mediaUri(String service, String path) => baseUri.replace(path: '${baseUri.path.replaceAll(RegExp(r'/$'), '')}/media/$service/${path.replaceFirst(RegExp('^/'), '')}');

  Uri socketUri(String path, [Map<String, String>? query]) {
    final base = baseUri;
    return base.replace(
      scheme: base.scheme == 'https' ? 'wss' : 'ws',
      path: '${base.path.replaceAll(RegExp(r'/$'), '')}/${path.replaceFirst(RegExp('^/'), '')}',
      queryParameters: query,
    );
  }

  Uri signalingUri(String roomId) => socketUri('rtc/ws', {'room_id': roomId});

  Future<String?> accessToken() async {
    var tokens = await storage.read();
    if (tokens != null && tokens.isExpired) tokens = await _refresh(tokens);
    return tokens?.accessToken;
  }

  @override
  Future<Map<String, dynamic>> query(String document, {Map<String, dynamic>? variables}) =>
      _execute(document, variables, authenticated: true);

  @override
  Future<Map<String, dynamic>> mutation(String document, {Map<String, dynamic>? variables}) =>
      _execute(document, variables, authenticated: true);

  @override
  Future<Map<String, dynamic>> public(String document, {Map<String, dynamic>? variables}) =>
      _execute(document, variables, authenticated: false);

  Future<Map<String, dynamic>> _execute(
    String document,
    Map<String, dynamic>? variables, {
    required bool authenticated,
  }) async {
    var tokens = authenticated ? await storage.read() : null;
    if (tokens != null && tokens.isExpired) tokens = await _refresh(tokens);
    try {
      return await _send(document, variables, tokens?.accessToken);
    } on GraphQLException catch (e) {
      if (!e.isSessionInvalid || tokens == null) rethrow;
      final refreshed = await _refresh(tokens);
      if (refreshed == null) rethrow;
      return _send(document, variables, refreshed.accessToken);
    }
  }

  Future<Map<String, dynamic>> _send(String document, Map<String, dynamic>? variables, String? accessToken) async {
    final headers = {
      'X-Device-Id': await storage.deviceId(),
      if (accessToken != null) 'Authorization': 'Bearer $accessToken',
    };
    final uploads = <String, GqlUpload>{};
    final encodedVariables = variables == null ? null : _extractUploads(variables, 'variables', uploads);
    final operations = jsonEncode({'query': document, 'variables': ?encodedVariables});

    final http.Response response;
    if (uploads.isEmpty) {
      response = await _http.post(Uri.parse(endpoint), headers: {...headers, 'Content-Type': 'application/json'}, body: operations);
    } else {
      final paths = uploads.keys.toList();
      final request = http.MultipartRequest('POST', Uri.parse(endpoint))
        ..headers.addAll(headers)
        ..fields['operations'] = operations
        ..fields['map'] = jsonEncode({
          for (var i = 0; i < paths.length; i++) '$i': [paths[i]],
        });
      for (var i = 0; i < paths.length; i++) {
        final file = uploads[paths[i]]!;
        request.files.add(http.MultipartFile.fromBytes(
          '$i',
          file.bytes,
          filename: file.filename,
          contentType: file.contentType == null ? null : MediaType.parse(file.contentType!),
        ));
      }
      response = await http.Response.fromStream(await _http.send(request));
    }

    final Map<String, dynamic> body;
    try {
      body = jsonDecode(utf8.decode(response.bodyBytes)) as Map<String, dynamic>;
    } on FormatException {
      throw GraphQLException('Unexpected response (${response.statusCode})', code: 'BAD_RESPONSE');
    }

    final errors = body['errors'];
    if (errors is List && errors.isNotEmpty) throw GraphQLException.fromErrors(errors);
    if (response.statusCode == 401) throw const GraphQLException('Unauthenticated', code: 'UNAUTHENTICATED');
    if (response.statusCode >= 400) throw GraphQLException('HTTP ${response.statusCode}');
    return (body['data'] as Map<String, dynamic>?) ?? const {};
  }

  static Object? _extractUploads(Object? value, String path, Map<String, GqlUpload> uploads) {
    if (value is GqlUpload) {
      uploads[path] = value;
      return null;
    }
    if (value is Map) return {for (final e in value.entries) e.key as String: _extractUploads(e.value, '$path.${e.key}', uploads)};
    if (value is List) return [for (var i = 0; i < value.length; i++) _extractUploads(value[i], '$path.$i', uploads)];
    return value;
  }

  Future<AuthTokens?> _refresh(AuthTokens current) {
    return _refreshing ??= _doRefresh(current).whenComplete(() => _refreshing = null);
  }

  Future<AuthTokens?> _doRefresh(AuthTokens current) async {
    try {
      final data = await _send(_refreshMutation, {'refreshToken': current.refreshToken}, null);
      final tokens = tokensFromJson(data['refreshToken'] as Map<String, dynamic>);
      await storage.save(tokens);
      return tokens;
    } catch (_) {
      await storage.clear();
      _sessionExpired.add(null);
      return null;
    }
  }

  static AuthTokens tokensFromJson(Map<String, dynamic> json) {
    final expiresIn = (json['expiresIn'] as num?)?.toInt() ?? 0;
    return AuthTokens(
      accessToken: json['accessToken'] as String,
      refreshToken: json['refreshToken'] as String,
      expiresAt: DateTime.now().add(Duration(seconds: expiresIn > 0 ? expiresIn : 3600)),
    );
  }

  void dispose() {
    _sessionExpired.close();
    _http.close();
  }
}
