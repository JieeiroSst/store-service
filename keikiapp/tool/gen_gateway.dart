import 'dart:convert';
import 'dart:io';

import 'package:gql/ast.dart';
import 'package:gql/language.dart' as lang;

part 'gen_admin.dart';

const _scalars = {'String', 'ID', 'UUID', 'Int', 'Int64', 'Int32', 'Float', 'Boolean', 'Map', 'Any', 'Upload'};
const _rootTypes = {'Query', 'Mutation'};
const _maxDepth = 2;

const _dartKeywords = {
  'abstract', 'as', 'assert', 'async', 'await', 'base', 'break', 'case', 'catch', 'class', 'const', 'continue',
  'covariant', 'default', 'deferred', 'do', 'dynamic', 'else', 'enum', 'export', 'extends', 'extension', 'external',
  'factory', 'false', 'final', 'finally', 'for', 'function', 'get', 'hide', 'if', 'implements', 'import', 'in',
  'interface', 'is', 'late', 'library', 'mixin', 'new', 'null', 'of', 'on', 'operator', 'part', 'required', 'rethrow',
  'return', 'sealed', 'set', 'show', 'static', 'super', 'switch', 'sync', 'this', 'throw', 'true', 'try', 'type',
  'typedef', 'var', 'void', 'when', 'while', 'with', 'yield', 'values', 'index', 'hashCode', 'runtimeType', 'toString',
  'noSuchMethod', 'toJson', 'fromJson', 'name',
};

const _reservedClassNames = {
  'Object', 'Map', 'List', 'Set', 'String', 'Duration', 'DateTime', 'Error', 'Exception', 'Function', 'Type', 'Symbol',
  'Iterable', 'Future', 'Stream', 'Record', 'Pattern', 'Match', 'Uri', 'Comparable', 'Enum', 'Null', 'Never', 'Sink',
};

class TypeRef {
  TypeRef(this.name, {this.nonNull = false, this.ofType});

  final String? name;
  final bool nonNull;
  final TypeRef? ofType;

  bool get isList => ofType != null;

  String get base => isList ? ofType!.base : name!;

  static TypeRef from(TypeNode node) {
    if (node is ListTypeNode) return TypeRef(null, nonNull: node.isNonNull, ofType: from(node.type));
    final named = node as NamedTypeNode;
    return TypeRef(named.name.value, nonNull: named.isNonNull);
  }

  @override
  String toString() => '${isList ? '[$ofType]' : name}${nonNull ? '!' : ''}';
}

class Field {
  Field(this.name, this.type, {this.args = const [], this.service = '', this.access = 'user', this.permission});

  final String name;
  final TypeRef type;
  final List<Field> args;
  final String service;
  final String access;
  final Map<String, String>? permission;
  late final String kind;
}

class TypeDef {
  TypeDef(this.name, this.kind, this.service);

  final String name;
  final String kind;
  final String service;
  final fields = <Field>[];
  final enumValues = <String>[];
}

final types = <String, TypeDef>{};
final operations = <Field>[];

void main(List<String> argv) {
  final schemaDir = Directory(argv.isNotEmpty ? argv[0] : '../gateway-api/graph');
  final outDir = Directory(argv.length > 1 ? argv[1] : 'lib/gen');
  final assetFile = File(argv.length > 2 ? argv[2] : 'assets/gateway_schema.json');

  final files = schemaDir.listSync().whereType<File>().where((f) => f.path.endsWith('.graphqls')).toList()
    ..sort((a, b) => a.path.compareTo(b.path));
  for (final file in files) {
    final service = file.uri.pathSegments.last.replaceAll('.graphqls', '');
    _collect(lang.parseString(file.readAsStringSync(), url: file.path), service);
  }

  final apiDir = Directory('${outDir.path}/api')..createSync(recursive: true);
  for (final old in apiDir.listSync().whereType<File>()) {
    old.deleteSync();
  }
  File('${outDir.path}/models.g.dart').writeAsStringSync(_models());
  final services = (operations.map((o) => o.service).toSet().toList()..sort());
  for (final s in services) {
    File('${apiDir.path}/$s.g.dart').writeAsStringSync(_serviceApi(s));
  }
  File('${outDir.path}/gateway_api.g.dart').writeAsStringSync(_gatewayApi(services));
  assetFile.parent.createSync(recursive: true);
  assetFile.writeAsStringSync(jsonEncode(_catalog(services)));
  _writeAdminConsoles(services, Directory(argv.length > 3 ? argv[3] : 'lib/features/admin/console/gen'));

  stdout.writeln('gen_gateway: ${operations.length} operations, ${services.length} services, '
      '${types.values.where((t) => t.kind == 'object').length} object types, '
      '${types.values.where((t) => t.kind == 'input').length} inputs, '
      '${types.values.where((t) => t.kind == 'enum').length} enums');
}

void _collect(DocumentNode doc, String service) {
  for (final def in doc.definitions) {
    if (def is ObjectTypeDefinitionNode || def is ObjectTypeExtensionNode) {
      final name = def is ObjectTypeDefinitionNode ? def.name.value : (def as ObjectTypeExtensionNode).name.value;
      final fields = def is ObjectTypeDefinitionNode ? def.fields : (def as ObjectTypeExtensionNode).fields;
      if (_rootTypes.contains(name)) {
        for (final f in fields) {
          final op = _field(f, service)..kind = name == 'Query' ? 'query' : 'mutation';
          operations.add(op);
        }
        continue;
      }
      final t = types.putIfAbsent(name, () => TypeDef(name, 'object', service));
      t.fields.addAll(fields.where((f) => f.args.isEmpty).map((f) => Field(f.name.value, TypeRef.from(f.type))));
    } else if (def is InputObjectTypeDefinitionNode) {
      final t = types.putIfAbsent(def.name.value, () => TypeDef(def.name.value, 'input', service));
      t.fields.addAll(def.fields.map((f) => Field(f.name.value, TypeRef.from(f.type))));
    } else if (def is EnumTypeDefinitionNode) {
      final t = types.putIfAbsent(def.name.value, () => TypeDef(def.name.value, 'enum', service));
      t.enumValues.addAll(def.values.map((v) => v.name.value));
    }
  }
}

Field _field(FieldDefinitionNode f, String service) {
  var access = 'user';
  Map<String, String>? permission;
  for (final d in f.directives) {
    if (d.name.value == 'public') access = 'public';
    if (d.name.value == 'hasPermission') {
      access = 'admin';
      permission = {
        for (final a in d.arguments)
          if (a.value is StringValueNode) a.name.value: (a.value as StringValueNode).value,
      };
    }
  }
  return Field(
    f.name.value,
    TypeRef.from(f.type),
    args: [for (final a in f.args) Field(a.name.value, TypeRef.from(a.type))],
    service: service,
    access: access,
    permission: permission,
  );
}

bool _isScalarOrEnum(String base) => _scalars.contains(base) || types[base]?.kind == 'enum';

String _selection(String typeName, [int depth = 0, Set<String> visiting = const {}]) {
  final t = types[typeName];
  if (t == null || t.kind != 'object') return '';
  final parts = <String>[];
  for (final f in t.fields) {
    final base = f.type.base;
    if (_isScalarOrEnum(base)) {
      parts.add(f.name);
    } else if (depth < _maxDepth && !visiting.contains(base)) {
      final sub = _selection(base, depth + 1, {...visiting, typeName});
      if (sub.isNotEmpty) parts.add('${f.name} $sub');
    }
  }
  return parts.isEmpty ? '' : '{ ${parts.join(' ')} }';
}

String className(String name) => _reservedClassNames.contains(name) ? 'G$name' : name;

String _camel(String raw) {
  final parts = raw.split(RegExp(r'[_\-\s]+')).where((p) => p.isNotEmpty).toList();
  if (parts.isEmpty) return 'value';
  final buf = StringBuffer();
  for (var i = 0; i < parts.length; i++) {
    final p = parts[i];
    if (i == 0) {
      buf.write(p == p.toUpperCase() ? p.toLowerCase() : p[0].toLowerCase() + p.substring(1));
    } else {
      buf.write(p[0].toUpperCase() + p.substring(1));
    }
  }
  var out = buf.toString().replaceAll(RegExp(r'[^A-Za-z0-9_]'), '');
  if (out.isEmpty || RegExp(r'^[0-9]').hasMatch(out)) out = 'v$out';
  if (_dartKeywords.contains(out)) out = '${out}_';
  return out;
}

Map<String, String> _dartNames(Iterable<String> names, {Set<String> taken = const {}}) {
  final used = <String>{...taken};
  final out = <String, String>{};
  for (final n in names) {
    var d = _camel(n);
    var i = 2;
    while (used.contains(d)) {
      d = '${_camel(n)}$i';
      i++;
    }
    used.add(d);
    out[n] = d;
  }
  return out;
}

String _dartBase(String base) {
  switch (base) {
    case 'String':
    case 'ID':
    case 'UUID':
      return 'String';
    case 'Int':
    case 'Int64':
    case 'Int32':
      return 'int';
    case 'Float':
      return 'double';
    case 'Boolean':
      return 'bool';
    case 'Map':
      return 'Map<String, dynamic>';
    case 'Any':
      return 'Object';
    case 'Upload':
      return 'GqlUpload';
  }
  return className(base);
}

String dartType(TypeRef t, {bool nullable = true}) {
  final inner = t.isList ? 'List<${dartType(t.ofType!, nullable: !t.ofType!.nonNull)}>' : _dartBase(t.name!);
  if (inner == 'Object' && nullable) return 'Object?';
  return nullable ? '$inner?' : inner;
}

String _decode(TypeRef t, String expr) {
  if (t.isList) {
    final el = t.ofType!;
    final conv = _decode(el, 'e');
    final elExpr = el.nonNull && conv != 'e' ? '($conv)!' : conv;
    final cast = el.nonNull && conv == 'e' ? 'e as ${dartType(el, nullable: false)}' : elExpr;
    return '($expr as List?)?.map((e) => $cast).toList()';
  }
  final base = t.name!;
  switch (base) {
    case 'String':
    case 'ID':
    case 'UUID':
      return '$expr?.toString()';
    case 'Int':
    case 'Int64':
    case 'Int32':
      return 'gqlInt($expr)';
    case 'Float':
      return '($expr as num?)?.toDouble()';
    case 'Boolean':
      return '$expr as bool?';
    case 'Map':
      return '($expr as Map?)?.cast<String, dynamic>()';
    case 'Any':
      return expr;
    case 'Upload':
      return '$expr as GqlUpload?';
  }
  final def = types[base];
  if (def?.kind == 'enum') return '$expr == null ? null : ${className(base)}.fromJson($expr.toString())';
  if (def != null) return '$expr == null ? null : ${className(base)}.fromJson(($expr as Map).cast<String, dynamic>())';
  return expr;
}

String _encode(TypeRef t, String expr, {bool nullable = true}) {
  final q = nullable ? '?' : '';
  if (t.isList) {
    final el = t.ofType!;
    final inner = _encode(el, 'e', nullable: !el.nonNull);
    if (inner == 'e') return expr;
    return '$expr$q.map((e) => $inner).toList()';
  }
  final def = types[t.name!];
  if (def?.kind == 'enum' || def?.kind == 'input' || def?.kind == 'object') return '$expr$q.toJson()';
  return expr;
}

const _header = '// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart\n'
    '// ignore_for_file: type=lint\n\n';

String _models() {
  final b = StringBuffer()
    ..writeln('int? gqlInt(Object? v) => v is num ? v.toInt() : (v is String ? int.tryParse(v) : null);')
    ..writeln();
  final sorted = types.values.toList()..sort((a, b) => a.name.compareTo(b.name));
  for (final t in sorted) {
    final cls = className(t.name);
    if (t.kind == 'enum') {
      final names = _dartNames(t.enumValues);
      b.writeln('enum $cls {');
      b.writeln(t.enumValues.map((v) => "  ${names[v]}('$v')").join(',\n'));
      b.writeln(';\n');
      b.writeln('  const $cls(this.value);\n');
      b.writeln('  final String value;\n');
      b.writeln('  static $cls? fromJson(String value) {');
      b.writeln('    for (final e in $cls.values) {');
      b.writeln('      if (e.value == value) return e;');
      b.writeln('    }');
      b.writeln('    return null;');
      b.writeln('  }\n');
      b.writeln('  String toJson() => value;');
      b.writeln('}\n');
      continue;
    }
    final names = _dartNames(t.fields.map((f) => f.name));
    final isInput = t.kind == 'input';
    b.writeln('class $cls {');
    if (t.fields.isEmpty) {
      b.writeln('  const $cls();\n');
    } else {
      b.writeln('  const $cls({');
      for (final f in t.fields) {
        final req = isInput && f.type.nonNull;
        b.writeln('    ${req ? 'required ' : ''}this.${names[f.name]},');
      }
      b.writeln('  });\n');
    }
    for (final f in t.fields) {
      final req = isInput && f.type.nonNull;
      b.writeln('  final ${dartType(f.type, nullable: !req)} ${names[f.name]};');
    }
    if (t.fields.isNotEmpty) b.writeln();
    b.writeln('  factory $cls.fromJson(Map<String, dynamic> json) => $cls(');
    for (final f in t.fields) {
      final req = isInput && f.type.nonNull;
      final decoded = _decode(f.type, "json['${f.name}']");
      b.writeln('        ${names[f.name]}: ${req ? '($decoded)${dartType(f.type, nullable: false) == 'Object' ? '' : '!'}' : decoded},');
    }
    b.writeln('      );\n');
    b.writeln('  Map<String, dynamic> toJson() => {');
    for (final f in t.fields) {
      final req = isInput && f.type.nonNull;
      final n = names[f.name]!;
      final enc = _encode(f.type, n, nullable: !req);
      if (req) {
        b.writeln("        '${f.name}': $enc,");
      } else if (enc == n) {
        b.writeln("        '${f.name}': ?$n,");
      } else {
        b.writeln("        if ($n != null) '${f.name}': ${_encode(f.type, '$n!', nullable: false)},");
      }
    }
    b.writeln('      };');
    b.writeln('}\n');
  }
  final body = b.toString();
  return '$_header${body.contains('GqlUpload') ? "import 'gateway_executor.dart';\n\n" : ''}$body';
}

String serviceClass(String service) => '${service.split('_').map((p) => p[0].toUpperCase() + p.substring(1)).join()}Api';

String serviceGetter(String service) => _camel(service);

String _serviceApi(String service) {
  final ops = operations.where((o) => o.service == service).toList();
  final b = StringBuffer()
    ..writeln('class ${serviceClass(service)} {')
    ..writeln('  const ${serviceClass(service)}(this._executor);\n')
    ..writeln('  final GatewayExecutor _executor;\n');
  for (final op in ops) {
    final argNames = _dartNames(op.args.map((a) => a.name), taken: {'selectionSet'});
    final selection = _selection(op.type.base);
    final retType = dartType(op.type);
    final params = <String>[
      for (final a in op.args)
        if (a.type.nonNull) 'required ${dartType(a.type, nullable: false)} ${argNames[a.name]}' else '${dartType(a.type)} ${argNames[a.name]}',
      if (selection.isNotEmpty) 'String? selectionSet',
    ];
    final varDecl = op.args.isEmpty ? '' : '(${op.args.map((a) => '\\\$${a.name}: ${a.type}').join(', ')})';
    final callArgs = op.args.isEmpty ? '' : '(${op.args.map((a) => '${a.name}: \\\$${a.name}').join(', ')})';
    final sel = selection.isEmpty ? '' : ' \${selectionSet ?? ${_dartString(selection)}}';
    final exec = op.access == 'public' ? 'public' : op.kind;
    b.writeln('  Future<$retType> ${methodName(op)}(${params.isEmpty ? '' : '{${params.join(', ')}}'}) async {');
    b.writeln("    final data = await _executor.$exec('${op.kind} ${op.name}$varDecl { ${op.name}$callArgs$sel }'${op.args.isEmpty ? '' : ', variables: {'}");
    if (op.args.isNotEmpty) {
      for (final a in op.args) {
        final n = argNames[a.name]!;
        final enc = _encode(a.type, n, nullable: !a.type.nonNull);
        if (a.type.nonNull) {
          b.writeln("      '${a.name}': $enc,");
        } else if (enc == n) {
          b.writeln("      '${a.name}': ?$n,");
        } else {
          b.writeln("      if ($n != null) '${a.name}': ${_encode(a.type, n, nullable: false)},");
        }
      }
      b.writeln('    }');
    }
    b.writeln('    );');
    b.writeln("    return ${_decode(op.type, "data['${op.name}']")};");
    b.writeln('  }\n');
  }
  b.writeln('}');
  final body = b.toString();
  final usesModels = ops.any((o) => !_scalars.contains(o.type.base) || o.args.any((a) => !_scalars.contains(a.type.base))) ||
      body.contains('gqlInt(');
  return '$_header'
      "import '../gateway_executor.dart';\n"
      "${usesModels ? "import '../models.g.dart';\n" : ''}\n"
      '$body';
}

String methodName(Field op) {
  final clash = operations.any((o) => o != op && o.name == op.name && o.kind != op.kind);
  return clash && op.kind == 'mutation' ? '${op.name}Mutation' : op.name;
}

String _dartString(String s) => "'${s.replaceAll(r'\', r'\\').replaceAll("'", r"\'").replaceAll(r'$', r'\$')}'";

String _gatewayApi(List<String> services) {
  final b = StringBuffer(_header)..writeln("import 'gateway_executor.dart';");
  for (final s in services) {
    b.writeln("import 'api/$s.g.dart';");
  }
  b.writeln();
  for (final s in services) {
    b.writeln("export 'api/$s.g.dart';");
  }
  b
    ..writeln("export 'gateway_executor.dart';")
    ..writeln("export 'models.g.dart';\n")
    ..writeln('class GatewayApi {')
    ..writeln('  GatewayApi(this.executor);\n')
    ..writeln('  final GatewayExecutor executor;\n');
  for (final s in services) {
    b.writeln('  late final ${serviceGetter(s)} = ${serviceClass(s)}(executor);');
  }
  b.writeln('}');
  return b.toString();
}

Map<String, Object?> _catalog(List<String> services) => {
      'services': [
        for (final s in services)
          {
            'id': s,
            'operations': [
              for (final op in operations.where((o) => o.service == s))
                {
                  'name': op.name,
                  'kind': op.kind,
                  'access': op.access,
                  if (op.permission != null) 'permission': op.permission,
                  'type': op.type.toString(),
                  'selection': _selection(op.type.base),
                  'args': [for (final a in op.args) {'name': a.name, 'type': a.type.toString()}],
                },
            ],
          },
      ],
      'types': {
        for (final t in types.values)
          t.name: {
            'kind': t.kind,
            if (t.kind == 'enum') 'values': t.enumValues,
            if (t.kind != 'enum') 'fields': [for (final f in t.fields) {'name': f.name, 'type': f.type.toString()}],
          },
      },
    };
