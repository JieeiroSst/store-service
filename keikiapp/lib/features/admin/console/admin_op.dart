import 'dart:convert';

import '../../../gen/gateway_api.g.dart';
import '../../apps/kit/kit.dart';

class AdminOp {
  const AdminOp(
    this.name, {
    required this.resource,
    required this.verb,
    required this.run,
    this.mutation = false,
    this.destructive = false,
    this.fields = const [],
  });

  final String name;
  final String resource;
  final String verb;
  final bool mutation;
  final bool destructive;
  final List<KField> fields;
  final Future<Object?> Function(GatewayApi api, FormData v) run;

  bool get hasRequired => fields.any((f) => f.required);
}

class AdminInputError implements Exception {
  const AdminInputError(this.message);

  final String message;

  @override
  String toString() => message;
}

Object? adminPlain(Object? value) {
  if (value == null || value is String || value is num || value is bool) return value;
  if (value is List) return value.map(adminPlain).toList();
  if (value is Map) return value.map((k, v) => MapEntry('$k', adminPlain(v)));
  if (value is GqlUpload) return value.filename;
  try {
    return adminPlain((value as dynamic).toJson());
  } on NoSuchMethodError {
    return '$value';
  }
}

extension AdminFormData on FormData {
  T need<T>(T? value, String key) => value ?? (throw AdminInputError('Thiếu hoặc sai giá trị: $key'));

  bool? optFlag(String key) => str(key) == null ? null : flag(key);

  bool any(String prefix) => values.entries.any((e) => e.key.startsWith(prefix) && e.value.trim().isNotEmpty) || files.keys.any((k) => k.startsWith(prefix));

  Object? json(String key) {
    final raw = str(key);
    if (raw == null) return null;
    try {
      return jsonDecode(raw);
    } on FormatException {
      throw AdminInputError('$key phải là JSON hợp lệ');
    }
  }

  Map<String, dynamic>? jsonMap(String key) {
    final v = json(key);
    if (v == null) return null;
    if (v is Map) return v.cast<String, dynamic>();
    throw AdminInputError('$key phải là một đối tượng JSON');
  }

  List<String>? strs(String key) {
    final raw = str(key);
    if (raw == null) return null;
    return raw.split(',').map((s) => s.trim()).where((s) => s.isNotEmpty).toList();
  }

  List<int>? intList(String key) => str(key) == null ? null : ints(key);
}
