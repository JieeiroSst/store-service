class GraphQLException implements Exception {
  const GraphQLException(this.message, {this.code, this.reason, this.details = const {}});

  final String message;
  final String? code;
  final String? reason;
  final Map<String, String> details;

  bool get isUnauthenticated => code == 'UNAUTHENTICATED';
  bool get isForbidden => code == 'FORBIDDEN';
  bool get isSessionInvalid => isUnauthenticated && (reason == null || reason == 'AUTH_TOKEN_INVALID');

  static GraphQLException fromErrors(List<dynamic> errors) {
    final first = errors.isNotEmpty ? errors.first as Map<String, dynamic> : const <String, dynamic>{};
    final extensions = first['extensions'];
    final ext = extensions is Map ? extensions : const {};
    final details = ext['details'];
    return GraphQLException(
      first['message']?.toString() ?? 'Unknown error',
      code: ext['code']?.toString(),
      reason: ext['reason']?.toString(),
      details: details is Map ? {for (final e in details.entries) e.key.toString(): e.value.toString()} : const {},
    );
  }

  @override
  String toString() => message;
}
