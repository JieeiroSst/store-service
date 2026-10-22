import 'dart:convert';

import 'package:flutter/foundation.dart';
import 'package:flutter/services.dart';

class GqlTypeRef {
  const GqlTypeRef(this.name, {this.nonNull = false, this.ofType});

  final String? name;
  final bool nonNull;
  final GqlTypeRef? ofType;

  bool get isList => ofType != null;

  String get base => isList ? ofType!.base : name!;

  static GqlTypeRef parse(String raw) {
    var s = raw.trim();
    final nonNull = s.endsWith('!');
    if (nonNull) s = s.substring(0, s.length - 1);
    if (s.startsWith('[')) return GqlTypeRef(null, nonNull: nonNull, ofType: parse(s.substring(1, s.length - 1)));
    return GqlTypeRef(s, nonNull: nonNull);
  }

  @override
  String toString() => '${isList ? '[$ofType]' : name}${nonNull ? '!' : ''}';
}

class GqlField {
  const GqlField(this.name, this.type);

  final String name;
  final GqlTypeRef type;
}

enum OperationAccess { public, user, admin }

class GqlOperation {
  const GqlOperation({
    required this.service,
    required this.name,
    required this.isMutation,
    required this.access,
    required this.type,
    required this.selection,
    required this.args,
    this.permissionObject,
    this.permissionAction,
  });

  final String service;
  final String name;
  final bool isMutation;
  final OperationAccess access;
  final GqlTypeRef type;
  final String selection;
  final List<GqlField> args;
  final String? permissionObject;
  final String? permissionAction;

  bool get hasRequiredArgs => args.any((a) => a.type.nonNull);

  String document({String? selection}) {
    final decl = args.isEmpty ? '' : '(${args.map((a) => '\$${a.name}: ${a.type}').join(', ')})';
    final call = args.isEmpty ? '' : '(${args.map((a) => '${a.name}: \$${a.name}').join(', ')})';
    return '${isMutation ? 'mutation' : 'query'} $name$decl { $name$call ${selection ?? this.selection} }';
  }
}

class GqlTypeDef {
  const GqlTypeDef({required this.name, required this.kind, this.fields = const [], this.enumValues = const []});

  final String name;
  final String kind;
  final List<GqlField> fields;
  final List<String> enumValues;

  bool get isEnum => kind == 'enum';
  bool get isInput => kind == 'input';
}

class GqlService {
  const GqlService(this.id, this.operations);

  final String id;
  final List<GqlOperation> operations;
}

class SchemaCatalog {
  SchemaCatalog(this.services, this.types) : _byService = {for (final s in services) s.id: s};

  final List<GqlService> services;
  final Map<String, GqlTypeDef> types;
  final Map<String, GqlService> _byService;

  GqlService? service(String id) => _byService[id];

  GqlOperation? operation(String service, String name) =>
      _byService[service]?.operations.where((o) => o.name == name).firstOrNull;

  static const scalars = {'String', 'ID', 'UUID', 'Int', 'Int64', 'Int32', 'Float', 'Boolean', 'Map', 'Any', 'Upload'};

  bool isScalar(String name) => scalars.contains(name);

  static Future<SchemaCatalog> load({AssetBundle? bundle, String asset = 'assets/gateway_schema.json'}) async {
    final raw = await (bundle ?? rootBundle).loadString(asset);
    return compute(parse, raw);
  }

  static SchemaCatalog parse(String raw) {
    final json = jsonDecode(raw) as Map<String, dynamic>;
    final services = [
      for (final s in (json['services'] as List).cast<Map<String, dynamic>>())
        GqlService(s['id'] as String, [
          for (final o in (s['operations'] as List).cast<Map<String, dynamic>>())
            GqlOperation(
              service: s['id'] as String,
              name: o['name'] as String,
              isMutation: o['kind'] == 'mutation',
              access: OperationAccess.values.byName(o['access'] as String),
              type: GqlTypeRef.parse(o['type'] as String),
              selection: o['selection'] as String? ?? '',
              args: [
                for (final a in (o['args'] as List).cast<Map<String, dynamic>>())
                  GqlField(a['name'] as String, GqlTypeRef.parse(a['type'] as String)),
              ],
              permissionObject: (o['permission'] as Map?)?['object'] as String?,
              permissionAction: (o['permission'] as Map?)?['action'] as String?,
            ),
        ]),
    ];
    final types = {
      for (final e in (json['types'] as Map<String, dynamic>).entries)
        e.key: GqlTypeDef(
          name: e.key,
          kind: (e.value as Map)['kind'] as String,
          enumValues: [...((e.value as Map)['values'] as List? ?? const []).cast<String>()],
          fields: [
            for (final f in ((e.value as Map)['fields'] as List? ?? const []).cast<Map<String, dynamic>>())
              GqlField(f['name'] as String, GqlTypeRef.parse(f['type'] as String)),
          ],
        ),
    };
    return SchemaCatalog(services, types);
  }
}
