part of 'gen_gateway.dart';

const _adminVerbs = {
  'create', 'add', 'update', 'edit', 'delete', 'remove', 'cancel', 'get', 'list', 'search', 'find', 'set', 'assign',
  'revoke', 'approve', 'reject', 'activate', 'deactivate', 'enable', 'disable', 'sync', 'publish', 'unpublish', 'archive',
  'restore', 'reset', 'rotate', 'register', 'unregister', 'submit', 'send', 'process', 'run', 'trigger', 'start', 'stop',
  'pause', 'resume', 'block', 'unblock', 'mark', 'record', 'upsert', 'put', 'rename', 'move', 'import', 'export', 'generate',
  'refresh', 'close', 'open', 'complete', 'verify', 'validate', 'grant', 'issue', 'release', 'change', 'apply', 'count',
  'check', 'fetch', 'dispatch', 'renew', 'freeze', 'unfreeze', 'credit', 'debit', 'refund', 'capture', 'authorize', 'void',
};

const _readVerbs = {'get', 'list', 'search', 'find', 'count', 'check', 'fetch', 'export', 'verify', 'validate'};

const _httpVerbs = {'post': 'create', 'put': 'update', 'patch': 'update', 'delete': 'delete'};

const _pathNoise = {'admin', 'realms'};

const _adminSplitters = {'by', 'for', 'of', 'from', 'to', 'with', 'in', 'on'};

const _destructive = {'delete', 'remove', 'revoke', 'cancel', 'block', 'deactivate', 'disable', 'reset', 'archive', 'freeze', 'void', 'unregister', 'rotate'};

const _multilineNames = ['description', 'content', 'body', 'message', 'note', 'notes', 'text', 'bio', 'html', 'markdown', 'prompt', 'reason', 'summary', 'comment', 'script', 'query', 'sql', 'template'];

List<String> _words(String s) => s
    .replaceAllMapped(RegExp(r'([a-z0-9])([A-Z])'), (m) => '${m[1]} ${m[2]}')
    .replaceAllMapped(RegExp(r'([A-Z]+)([A-Z][a-z])'), (m) => '${m[1]} ${m[2]}')
    .split(RegExp(r'[\s_\-\.]+'))
    .where((w) => w.isNotEmpty)
    .map((w) => w.toLowerCase())
    .toList();

int _prefixLength(List<Field> ops) {
  final split = [for (final o in ops) _words(o.name)];
  String? common(int n) {
    final counts = <String, int>{};
    for (final w in split) {
      if (w.length > n) counts.update(w.sublist(0, n).join(' '), (v) => v + 1, ifAbsent: () => 1);
    }
    if (counts.isEmpty) return null;
    final best = counts.entries.reduce((a, b) => a.value >= b.value ? a : b);
    return best.value * 10 >= split.length * 6 ? best.key : null;
  }

  if (common(1) == null) return 0;
  if (common(2) != null && !_adminVerbs.contains(common(2)!.split(' ').last)) return 2;
  return 1;
}

String _singular(String w) {
  if (w.endsWith('ies') && w.length > 4) return '${w.substring(0, w.length - 3)}y';
  if (w.endsWith('sses') || w.endsWith('ches') || w.endsWith('xes')) return w.substring(0, w.length - 2);
  if (w.endsWith('s') && !w.endsWith('ss') && !w.endsWith('us') && !w.endsWith('is') && w.length > 3) return w.substring(0, w.length - 1);
  return w;
}

(String verb, String resource) _classify(Field op, int prefix) {
  var w = _words(op.name);
  final known = w.length > prefix ? w.sublist(prefix) : w;
  w = known;
  var verb = '';
  if (w.length > 1 && _httpVerbs.containsKey(w.first)) {
    verb = _httpVerbs[w.first]!;
    w = w.sublist(1);
  } else if (w.isNotEmpty && (op.kind == 'mutation' ? _adminVerbs.contains(w.first) : w.length > 1 && _readVerbs.contains(w.first))) {
    verb = w.first;
    w = w.sublist(1);
  }
  while (w.length > 1 && _pathNoise.contains(w.first)) {
    w = w.sublist(1);
  }
  final cut = w.indexWhere(_adminSplitters.contains);
  if (cut >= 0) w = w.sublist(0, cut);
  if (w.isNotEmpty) w = [...w.sublist(0, w.length - 1), _singular(w.last)];
  return (verb, w.isEmpty ? 'general' : w.join(' '));
}

String _label(String name) {
  const known = {'id': 'ID', 'url': 'URL', 'uuid': 'UUID', 'qr': 'QR', 'otp': 'OTP', 'pin': 'PIN', 'utm': 'UTM', 'sku': 'SKU', 'ip': 'IP', 'kyc': 'KYC'};
  final w = _words(name).map((e) => known[e] ?? e).toList();
  if (w.isEmpty) return name;
  final s = w.join(' ');
  return s[0].toUpperCase() + s.substring(1);
}

Object? _sample(TypeRef t, [int depth = 0]) {
  if (t.isList) return [_sample(t.ofType!, depth + 1)];
  final base = t.name!;
  switch (base) {
    case 'Int':
    case 'Int64':
    case 'Int32':
      return 0;
    case 'Float':
      return 0.0;
    case 'Boolean':
      return false;
    case 'Map':
    case 'Any':
      return <String, Object?>{};
  }
  final def = types[base];
  if (def?.kind == 'enum') return def!.enumValues.first;
  if (def != null && depth < 4) {
    final picked = [...def.fields.where((f) => f.type.nonNull), ...def.fields.where((f) => !f.type.nonNull).take(6)];
    return {for (final f in def.fields) if (picked.contains(f)) f.name: _sample(f.type, depth + 1)};
  }
  return '';
}

class _AdminField {
  _AdminField(this.key, this.label, this.kind, this.required, {this.hint, this.options = const []});

  final String key;
  final String label;
  final String kind;
  final bool required;
  final String? hint;
  final List<String> options;

  String code() {
    final args = <String>[
      if (kind != 'text') 'kind: FieldKind.$kind',
      if (required) 'required: true',
      if (hint != null) 'hint: ${_dartString(hint!)}',
      if (options.isNotEmpty) 'options: [${[if (!required) "('', 'Không chọn')", for (final o in options) '(${_dartString(o)}, ${_dartString(o)})'].join(', ')}]',
    ];
    return 'KField(${_dartString(key)}, ${_dartString(label)}${args.isEmpty ? '' : ', ${args.join(', ')}'})';
  }
}

String _scalarKind(String base, String name) {
  final n = name.toLowerCase();
  switch (base) {
    case 'Int':
    case 'Int64':
    case 'Int32':
    case 'Float':
      return 'number';
    case 'Boolean':
      return 'toggle';
    case 'Upload':
      return 'file';
    case 'Map':
    case 'Any':
      return 'multiline';
  }
  if (n.contains('password') || n.contains('secret')) return 'password';
  if (n == 'email' || n.endsWith('email')) return 'email';
  if (n.contains('phone')) return 'phone';
  if (_multilineNames.any((m) => n == m || n.endsWith(m))) return 'multiline';
  return 'text';
}

String _scalarExpr(String base, String key) {
  final k = _dartString(key);
  switch (base) {
    case 'Int':
    case 'Int64':
    case 'Int32':
      return 'v.integer($k)';
    case 'Float':
      return 'v.number($k)';
    case 'Boolean':
      return 'v.optFlag($k)';
    case 'Map':
      return 'v.jsonMap($k)';
    case 'Any':
      return 'v.json($k)';
    case 'Upload':
      return 'v.file($k)';
  }
  return 'v.str($k)';
}

String _need(String expr, TypeRef t, String key, bool nonNull) {
  if (!nonNull) return expr;
  if (!t.isList && t.name == 'Boolean') return 'v.flag(${_dartString(key)})';
  return 'v.need($expr, ${_dartString(key)})';
}

String _valueExpr(TypeRef t, String key, String name, List<_AdminField> out, {required bool flatten}) {
  final base = t.base;
  final def = types[base];
  final k = _dartString(key);
  if (!t.isList && _scalars.contains(base)) {
    final kind = _scalarKind(base, name);
    final lower = name.toLowerCase();
    final temporal = base == 'String' && (lower.endsWith('at') && name.endsWith('At') || lower.endsWith('_at') || lower.endsWith('time') || lower.endsWith('date') || lower == 'from' || lower == 'to' || lower.endsWith('since') || lower.endsWith('until'));
    out.add(_AdminField(key, _label(name), kind, t.nonNull && base != 'Boolean',
        hint: base == 'Map' || base == 'Any'
            ? 'JSON, ví dụ {"key": "value"}'
            : temporal
                ? 'ISO 8601, ví dụ 2026-10-22T09:00:00Z'
                : null));
    return _need(_scalarExpr(base, key), t, key, t.nonNull);
  }
  if (!t.isList && def?.kind == 'enum') {
    out.add(_AdminField(key, _label(name), 'choice', t.nonNull, options: def!.enumValues));
    return _need(_decode(t, 'v.str($k)'), t, key, t.nonNull);
  }
  if (t.isList && !t.ofType!.isList && _scalars.contains(t.ofType!.name)) {
    final el = t.ofType!.name!;
    final numeric = const {'Int', 'Int64', 'Int32'}.contains(el);
    final simple = el == 'String' || el == 'ID' || el == 'UUID' || numeric;
    if (simple) {
      out.add(_AdminField(key, _label(name), 'text', t.nonNull, hint: 'Nhiều giá trị, cách nhau bởi dấu phẩy'));
      return _need(numeric ? 'v.intList($k)' : 'v.strs($k)', t, key, t.nonNull);
    }
  }
  if (t.isList && !t.ofType!.isList && types[t.ofType!.name]?.kind == 'enum') {
    final e = className(t.ofType!.name!);
    out.add(_AdminField(key, _label(name), 'text', t.nonNull, hint: types[t.ofType!.name]!.enumValues.join(', ')));
    return _need('v.strs($k)?.map($e.fromJson).whereType<$e>().toList()', t, key, t.nonNull);
  }
  if (!t.isList && def?.kind == 'input' && flatten && def!.fields.isNotEmpty) {
    final names = _dartNames(def.fields.map((f) => f.name));
    final parts = <String>[];
    final inner = <_AdminField>[];
    for (final f in def.fields) {
      final expr = _valueExpr(f.type, '$key.${f.name}', f.name, inner, flatten: false);
      parts.add('${names[f.name]}: $expr');
    }
    out.addAll(inner.map((f) => t.nonNull ? f : _AdminField(f.key, f.label, f.kind, false, hint: f.hint, options: f.options)));
    final ctor = '${className(base)}(${parts.join(', ')})';
    return t.nonNull ? ctor : 'v.any(${_dartString('$key.')}) ? $ctor : null';
  }
  out.add(_AdminField(key, _label(name), 'multiline', t.nonNull, hint: 'JSON, ví dụ ${jsonEncode(_sample(t))}'));
  return _need(_decode(t, 'v.json($k)'), t, key, t.nonNull);
}

Map<Field, String> _groupResources(List<Field> ops, int prefix) {
  var res = {for (final o in ops) o: _classify(o, prefix).$2};
  Map<String, int> counts() {
    final c = <String, int>{};
    for (final r in res.values) {
      c.update(r, (v) => v + 1, ifAbsent: () => 1);
    }
    return c;
  }

  var c = counts();
  res = {
    for (final e in res.entries)
      e.key: c[e.value] == 1 && e.value.contains(' ') && c.containsKey(_singular(e.value.split(' ').first)) ? _singular(e.value.split(' ').first) : e.value,
  };
  c = counts();
  if (c.length > 16) {
    res = {for (final e in res.entries) e.key: _singular(e.value.split(' ').first)};
    c = counts();
  }
  if (c.length > 6) {
    res = {for (final e in res.entries) e.key: c[e.value] == 1 ? 'general' : e.value};
  }
  return res;
}

String _adminOp(Field op, int prefix, String resource) {
  final (verb, _) = _classify(op, prefix);
  final fields = <_AdminField>[];
  final argNames = _dartNames(op.args.map((a) => a.name), taken: {'selectionSet'});
  final callArgs = <String>[];
  for (final a in op.args) {
    final expr = _valueExpr(a.type, a.name, a.name, fields, flatten: true);
    callArgs.add('${argNames[a.name]}: $expr');
  }
  final call = 'api.${serviceGetter(op.service)}.${methodName(op)}(${callArgs.join(', ')})';
  final b = StringBuffer()
    ..writeln('  AdminOp(')
    ..writeln("    '${op.name}',")
    ..writeln("    resource: ${_dartString(resource)},")
    ..writeln("    verb: ${_dartString(verb)},");
  if (op.kind == 'mutation') b.writeln('    mutation: true,');
  if (op.kind == 'mutation' && _destructive.contains(verb)) b.writeln('    destructive: true,');
  if (fields.isNotEmpty) {
    b.writeln('    fields: [');
    for (final f in fields) {
      b.writeln('      ${f.code()},');
    }
    b.writeln('    ],');
  }
  b
    ..writeln('    run: (api, v) async => adminPlain(await $call),')
    ..writeln('  ),');
  return b.toString();
}

String _adminVar(String service) => '${_camel(service)}AdminOps';

void _writeAdminConsoles(List<String> services, Directory dir) {
  dir.createSync(recursive: true);
  for (final old in dir.listSync().whereType<File>()) {
    old.deleteSync();
  }
  final withAdmin = <String>[];
  for (final s in services) {
    final all = operations.where((o) => o.service == s).toList();
    final ops = all.where((o) => o.access == 'admin').toList();
    if (ops.isEmpty) continue;
    withAdmin.add(s);
    final prefix = _prefixLength(all);
    final b = StringBuffer(_header)
      ..writeln('// ignore_for_file: unused_import\n')
      ..writeln("import '../../../apps/kit/kit.dart';")
      ..writeln("import '../../../../gen/models.g.dart';")
      ..writeln("import '../admin_op.dart';\n")
      ..writeln('final ${_adminVar(s)} = <AdminOp>[');
    final groups = _groupResources(ops, prefix);
    for (final op in ops) {
      b.write(_adminOp(op, prefix, groups[op]!));
    }
    b.writeln('];');
    File('${dir.path}/$s.g.dart').writeAsStringSync(b.toString());
  }
  final idx = StringBuffer(_header)..writeln("import '../admin_op.dart';");
  for (final s in withAdmin) {
    idx.writeln("import '$s.g.dart';");
  }
  idx
    ..writeln()
    ..writeln('final adminConsoles = <String, List<AdminOp>>{');
  for (final s in withAdmin) {
    idx.writeln("  '$s': ${_adminVar(s)},");
  }
  idx.writeln('};');
  File('${dir.path}/admin_consoles.g.dart').writeAsStringSync(idx.toString());
  stdout.writeln('gen_admin: ${operations.where((o) => o.access == 'admin').length} admin operations in ${withAdmin.length} consoles');
}
