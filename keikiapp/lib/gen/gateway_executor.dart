import 'dart:typed_data';

class GqlUpload {
  const GqlUpload({required this.bytes, required this.filename, this.contentType});

  final Uint8List bytes;
  final String filename;
  final String? contentType;
}

abstract interface class GatewayExecutor {
  Future<Map<String, dynamic>> query(String document, {Map<String, dynamic>? variables});

  Future<Map<String, dynamic>> mutation(String document, {Map<String, dynamic>? variables});

  Future<Map<String, dynamic>> public(String document, {Map<String, dynamic>? variables});
}
