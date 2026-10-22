class GraphQLException implements Exception {
  const GraphQLException(this.message, {this.code});

  final String message;
  final String? code;

  bool get isUnauthenticated => code == 'UNAUTHENTICATED';
  bool get isForbidden => code == 'FORBIDDEN';

  static GraphQLException fromErrors(List<dynamic> errors) {
    final first = errors.isNotEmpty ? errors.first as Map<String, dynamic> : const <String, dynamic>{};
    final extensions = first['extensions'];
    return GraphQLException(
      first['message']?.toString() ?? 'Unknown error',
      code: extensions is Map ? extensions['code']?.toString() : null,
    );
  }

  @override
  String toString() => message;
}
