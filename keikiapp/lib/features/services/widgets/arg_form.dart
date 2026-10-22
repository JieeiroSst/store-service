import 'dart:convert';

import 'package:file_picker/file_picker.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:intl/intl.dart';
import 'package:uuid/uuid.dart';

import '../../../gen/gateway_executor.dart';
import '../data/humanize.dart';
import '../data/schema_catalog.dart';

const _intTypes = {'Int', 'Int64', 'Int32'};
const _userIdNames = {'user_id', 'userId', 'userID', 'owner_id', 'ownerId', 'customer_id', 'customerId', 'buyer_id', 'requested_by'};

Object? defaultValueFor(GqlField field, {String? userId}) {
  final n = field.name;
  final base = field.type.base;
  if (field.type.isList) return null;
  if (n == 'idempotency_key' || n == 'idempotencyKey') return const Uuid().v4();
  if (userId != null && _userIdNames.contains(n)) {
    if (_intTypes.contains(base)) return int.tryParse(userId);
    if (base == 'String' || base == 'ID') return userId;
  }
  if (_intTypes.contains(base)) {
    if (n == 'page') return 1;
    if (const {'limit', 'page_size', 'pageSize', 'per_page', 'perPage', 'size', 'first'}.contains(n)) return 20;
  }
  return null;
}

class ArgEditor extends StatelessWidget {
  const ArgEditor({
    super.key,
    required this.catalog,
    required this.field,
    required this.value,
    required this.onChanged,
    this.depth = 0,
  });

  final SchemaCatalog catalog;
  final GqlField field;
  final Object? value;
  final ValueChanged<Object?> onChanged;
  final int depth;

  @override
  Widget build(BuildContext context) {
    final type = field.type;
    if (type.isList) {
      return _ListEditor(catalog: catalog, field: field, value: value as List?, onChanged: onChanged, depth: depth);
    }
    final def = catalog.types[type.base];
    if (def != null && def.isEnum) {
      return _EnumEditor(field: field, values: def.enumValues, value: value as String?, onChanged: onChanged);
    }
    if (def != null && def.isInput) {
      return _InputObjectEditor(catalog: catalog, field: field, def: def, value: value as Map<String, dynamic>?, onChanged: onChanged, depth: depth);
    }
    switch (type.base) {
      case 'Boolean':
        return _BoolEditor(field: field, value: value as bool?, onChanged: onChanged);
      case 'Upload':
        return _UploadEditor(field: field, value: value as GqlUpload?, onChanged: onChanged);
    }
    return _ScalarEditor(key: ValueKey('${field.name}-$depth'), field: field, value: value, onChanged: onChanged);
  }
}

String _label(GqlField f) => '${fieldLabel(f.name)}${f.type.nonNull ? ' *' : ''}';

class _ScalarEditor extends StatefulWidget {
  const _ScalarEditor({super.key, required this.field, required this.value, required this.onChanged});

  final GqlField field;
  final Object? value;
  final ValueChanged<Object?> onChanged;

  @override
  State<_ScalarEditor> createState() => _ScalarEditorState();
}

class _ScalarEditorState extends State<_ScalarEditor> {
  late final _controller = TextEditingController(text: _initialText());
  bool _obscure = true;

  String _initialText() {
    final v = widget.value;
    if (v == null) return '';
    if (v is Map || v is List) return const JsonEncoder.withIndent('  ').convert(v);
    return '$v';
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  String get _base => widget.field.type.base;
  String get _name => widget.field.name.toLowerCase();
  bool get _isJson => _base == 'Map' || _base == 'Any';
  bool get _isSecret => _name.contains('password') || _name == 'pin' || _name.endsWith('_pin') || _name.contains('secret');
  bool get _isDate => _base == 'String' && (_name.endsWith('_at') || _name.contains('date') || _name == 'from' || _name == 'to' || _name.endsWith('time'));
  bool get _isLong => _isJson || const ['description', 'content', 'note', 'body', 'message', 'prompt', 'text', 'comment'].any(_name.contains);

  Object? _parse(String raw) {
    final t = raw.trim();
    if (t.isEmpty) return null;
    if (_intTypes.contains(_base)) return int.tryParse(t);
    if (_base == 'Float') return double.tryParse(t.replaceAll(',', '.'));
    if (_isJson) {
      try {
        return jsonDecode(t);
      } on FormatException {
        return t;
      }
    }
    return raw;
  }

  String? _validate(String? raw) {
    final t = (raw ?? '').trim();
    if (t.isEmpty) return widget.field.type.nonNull ? 'Bắt buộc' : null;
    if (_intTypes.contains(_base) && int.tryParse(t) == null) return 'Cần một số nguyên';
    if (_base == 'Float' && double.tryParse(t.replaceAll(',', '.')) == null) return 'Cần một số';
    if (_base == 'Map') {
      try {
        if (jsonDecode(t) is! Map) return 'Cần một object JSON, ví dụ {"key": "value"}';
      } on FormatException {
        return 'JSON không hợp lệ';
      }
    }
    return null;
  }

  Future<void> _pickDate() async {
    final now = DateTime.now();
    final date = await showDatePicker(context: context, firstDate: DateTime(now.year - 50), lastDate: DateTime(now.year + 10), initialDate: now);
    if (date == null || !mounted) return;
    final time = await showTimePicker(context: context, initialTime: TimeOfDay.now());
    final value = DateTime(date.year, date.month, date.day, time?.hour ?? 0, time?.minute ?? 0);
    final text = _name.contains('date') && time == null ? DateFormat('yyyy-MM-dd').format(value) : value.toUtc().toIso8601String();
    _controller.text = text;
    widget.onChanged(text);
  }

  @override
  Widget build(BuildContext context) {
    final numeric = _intTypes.contains(_base) || _base == 'Float';
    return TextFormField(
      controller: _controller,
      obscureText: _isSecret && _obscure,
      minLines: _isLong ? 2 : 1,
      maxLines: _isSecret ? 1 : (_isLong ? 6 : 1),
      keyboardType: numeric
          ? TextInputType.numberWithOptions(decimal: _base == 'Float', signed: true)
          : (_name.contains('email') ? TextInputType.emailAddress : (_name.contains('phone') ? TextInputType.phone : null)),
      inputFormatters: _intTypes.contains(_base) ? [FilteringTextInputFormatter.allow(RegExp(r'^-?\d*'))] : null,
      style: _isJson ? const TextStyle(fontFamily: 'monospace', fontSize: 13) : null,
      decoration: InputDecoration(
        labelText: _label(widget.field),
        helperText: _isJson ? 'JSON' : null,
        suffixIcon: _isSecret
            ? IconButton(icon: Icon(_obscure ? Icons.visibility : Icons.visibility_off), onPressed: () => setState(() => _obscure = !_obscure))
            : _isDate
                ? IconButton(icon: const Icon(Icons.event), onPressed: _pickDate)
                : (widget.field.name == 'idempotency_key'
                    ? IconButton(
                        tooltip: 'Tạo mã mới',
                        icon: const Icon(Icons.autorenew),
                        onPressed: () {
                          _controller.text = const Uuid().v4();
                          widget.onChanged(_controller.text);
                        },
                      )
                    : null),
      ),
      validator: _validate,
      onChanged: (raw) => widget.onChanged(_parse(raw)),
    );
  }
}

class _BoolEditor extends StatelessWidget {
  const _BoolEditor({required this.field, required this.value, required this.onChanged});

  final GqlField field;
  final bool? value;
  final ValueChanged<Object?> onChanged;

  @override
  Widget build(BuildContext context) {
    if (field.type.nonNull) {
      return FormField<bool>(
        initialValue: value ?? false,
        builder: (state) => SwitchListTile(
          contentPadding: const EdgeInsets.symmetric(horizontal: 4),
          title: Text(_label(field)),
          value: value ?? false,
          onChanged: onChanged,
        ),
      );
    }
    return Row(
      children: [
        Expanded(child: Text(_label(field), style: Theme.of(context).textTheme.bodyLarge)),
        SegmentedButton<int>(
          showSelectedIcon: false,
          segments: const [
            ButtonSegment(value: 0, label: Text('—')),
            ButtonSegment(value: 1, label: Text('Có')),
            ButtonSegment(value: 2, label: Text('Không')),
          ],
          selected: {value == null ? 0 : (value! ? 1 : 2)},
          onSelectionChanged: (s) => onChanged(switch (s.first) { 1 => true, 2 => false, _ => null }),
        ),
      ],
    );
  }
}

class _EnumEditor extends StatelessWidget {
  const _EnumEditor({required this.field, required this.values, required this.value, required this.onChanged});

  final GqlField field;
  final List<String> values;
  final String? value;
  final ValueChanged<Object?> onChanged;

  @override
  Widget build(BuildContext context) {
    return DropdownButtonFormField<String>(
      initialValue: value,
      decoration: InputDecoration(labelText: _label(field)),
      items: [
        if (!field.type.nonNull) const DropdownMenuItem(value: null, child: Text('—')),
        for (final v in values) DropdownMenuItem(value: v, child: Text(fieldLabel(v.toLowerCase()))),
      ],
      validator: (v) => field.type.nonNull && v == null ? 'Bắt buộc' : null,
      onChanged: onChanged,
    );
  }
}

class _UploadEditor extends StatelessWidget {
  const _UploadEditor({required this.field, required this.value, required this.onChanged});

  final GqlField field;
  final GqlUpload? value;
  final ValueChanged<Object?> onChanged;

  static const _mime = {
    'jpg': 'image/jpeg',
    'jpeg': 'image/jpeg',
    'png': 'image/png',
    'gif': 'image/gif',
    'webp': 'image/webp',
    'pdf': 'application/pdf',
    'mp4': 'video/mp4',
    'mov': 'video/quicktime',
    'csv': 'text/csv',
    'json': 'application/json',
    'txt': 'text/plain',
  };

  Future<void> _pick() async {
    final file = await FilePicker.pickFile();
    if (file == null) return;
    final ext = (file.extension ?? '').replaceFirst('.', '').toLowerCase();
    onChanged(GqlUpload(bytes: await file.readAsBytes(), filename: file.name, contentType: _mime[ext]));
  }

  @override
  Widget build(BuildContext context) {
    return FormField<GqlUpload>(
      initialValue: value,
      validator: (_) => field.type.nonNull && value == null ? 'Hãy chọn một tệp' : null,
      builder: (state) => InputDecorator(
        decoration: InputDecoration(labelText: _label(field), errorText: state.errorText),
        child: Row(
          children: [
            Icon(value == null ? Icons.attach_file : Icons.insert_drive_file, size: 20),
            const SizedBox(width: 8),
            Expanded(
              child: Text(
                value == null ? 'Chưa chọn tệp' : '${value!.filename} • ${(value!.bytes.length / 1024).toStringAsFixed(1)} KB',
                overflow: TextOverflow.ellipsis,
              ),
            ),
            TextButton(onPressed: _pick, child: Text(value == null ? 'Chọn tệp' : 'Đổi')),
          ],
        ),
      ),
    );
  }
}

class _InputObjectEditor extends StatelessWidget {
  const _InputObjectEditor({
    required this.catalog,
    required this.field,
    required this.def,
    required this.value,
    required this.onChanged,
    required this.depth,
  });

  final SchemaCatalog catalog;
  final GqlField field;
  final GqlTypeDef def;
  final Map<String, dynamic>? value;
  final ValueChanged<Object?> onChanged;
  final int depth;

  @override
  Widget build(BuildContext context) {
    final scheme = Theme.of(context).colorScheme;
    final enabled = field.type.nonNull || value != null;
    final current = value ?? const <String, dynamic>{};
    final required = def.fields.where((f) => f.type.nonNull).toList();
    final optional = def.fields.where((f) => !f.type.nonNull).toList();
    Widget editor(GqlField f) => Padding(
          padding: const EdgeInsets.only(bottom: 12),
          child: ArgEditor(
            catalog: catalog,
            field: f,
            depth: depth + 1,
            value: current[f.name],
            onChanged: (v) => onChanged({...current, f.name: v}),
          ),
        );
    return Container(
      decoration: BoxDecoration(
        border: Border.all(color: scheme.outlineVariant),
        borderRadius: BorderRadius.circular(16),
        color: depth.isEven ? scheme.surfaceContainerLowest : scheme.surfaceContainerLow,
      ),
      child: Theme(
        data: Theme.of(context).copyWith(dividerColor: Colors.transparent),
        child: ExpansionTile(
          initiallyExpanded: field.type.nonNull || depth == 0,
          tilePadding: const EdgeInsets.symmetric(horizontal: 12),
          childrenPadding: const EdgeInsets.fromLTRB(12, 0, 12, 4),
          leading: const Icon(Icons.data_object),
          title: Text(_label(field)),
          subtitle: Text('${def.fields.length} trường', style: Theme.of(context).textTheme.bodySmall),
          trailing: field.type.nonNull
              ? null
              : Switch(value: enabled, onChanged: (on) => onChanged(on ? <String, dynamic>{} : null)),
          children: !enabled
              ? const []
              : [
                  for (final f in required) editor(f),
                  if (optional.isNotEmpty && required.isNotEmpty)
                    Padding(
                      padding: const EdgeInsets.only(bottom: 8),
                      child: Text('Tuỳ chọn', style: Theme.of(context).textTheme.labelMedium?.copyWith(color: scheme.onSurfaceVariant)),
                    ),
                  for (final f in optional) editor(f),
                ],
        ),
      ),
    );
  }
}

class _ListEditor extends StatelessWidget {
  const _ListEditor({required this.catalog, required this.field, required this.value, required this.onChanged, required this.depth});

  final SchemaCatalog catalog;
  final GqlField field;
  final List? value;
  final ValueChanged<Object?> onChanged;
  final int depth;

  bool get _scalarItems {
    final base = field.type.base;
    return catalog.isScalar(base) && base != 'Upload' && base != 'Map' && base != 'Any' && base != 'Boolean';
  }

  @override
  Widget build(BuildContext context) {
    final items = value ?? const [];
    final element = GqlField('${field.name} #', field.type.ofType!);
    if (_scalarItems) return _ChipListEditor(field: field, items: items, onChanged: onChanged);
    return FormField<List>(
      initialValue: items,
      validator: (_) => field.type.nonNull && items.isEmpty ? 'Cần ít nhất một mục' : null,
      builder: (state) => InputDecorator(
        decoration: InputDecoration(labelText: _label(field), errorText: state.errorText, contentPadding: const EdgeInsets.fromLTRB(12, 20, 12, 8)),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            for (var i = 0; i < items.length; i++)
              Padding(
                padding: const EdgeInsets.only(bottom: 8),
                child: Row(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Expanded(
                      child: ArgEditor(
                        catalog: catalog,
                        field: GqlField('${field.name} ${i + 1}', GqlTypeRef(element.type.name, nonNull: true, ofType: element.type.ofType)),
                        value: items[i],
                        depth: depth + 1,
                        onChanged: (v) => onChanged([...items]..[i] = v),
                      ),
                    ),
                    IconButton(icon: const Icon(Icons.remove_circle_outline), onPressed: () => onChanged([...items]..removeAt(i))),
                  ],
                ),
              ),
            Align(
              alignment: Alignment.centerLeft,
              child: TextButton.icon(
                icon: const Icon(Icons.add),
                label: const Text('Thêm mục'),
                onPressed: () => onChanged([...items, catalog.types[field.type.base]?.isInput == true ? <String, dynamic>{} : null]),
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class _ChipListEditor extends StatefulWidget {
  const _ChipListEditor({required this.field, required this.items, required this.onChanged});

  final GqlField field;
  final List items;
  final ValueChanged<Object?> onChanged;

  @override
  State<_ChipListEditor> createState() => _ChipListEditorState();
}

class _ChipListEditorState extends State<_ChipListEditor> {
  final _controller = TextEditingController();

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  void _add() {
    final raw = _controller.text.trim();
    if (raw.isEmpty) return;
    final base = widget.field.type.base;
    final parts = raw.split(',').map((p) => p.trim()).where((p) => p.isNotEmpty);
    final parsed = [
      for (final p in parts)
        if (_intTypes.contains(base)) int.tryParse(p) else if (base == 'Float') double.tryParse(p) else p,
    ].where((v) => v != null);
    widget.onChanged([...widget.items, ...parsed]);
    _controller.clear();
  }

  @override
  Widget build(BuildContext context) {
    return FormField<List>(
      initialValue: widget.items,
      validator: (_) => widget.field.type.nonNull && widget.items.isEmpty ? 'Cần ít nhất một giá trị' : null,
      builder: (state) => Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          TextField(
            controller: _controller,
            onSubmitted: (_) => _add(),
            decoration: InputDecoration(
              labelText: _label(widget.field),
              helperText: 'Nhập rồi bấm +, nhiều giá trị cách nhau bởi dấu phẩy',
              errorText: state.errorText,
              suffixIcon: IconButton(icon: const Icon(Icons.add), onPressed: _add),
            ),
          ),
          if (widget.items.isNotEmpty)
            Padding(
              padding: const EdgeInsets.only(top: 8),
              child: Wrap(
                spacing: 6,
                runSpacing: 6,
                children: [
                  for (var i = 0; i < widget.items.length; i++)
                    InputChip(label: Text('${widget.items[i]}'), onDeleted: () => widget.onChanged([...widget.items]..removeAt(i))),
                ],
              ),
            ),
        ],
      ),
    );
  }
}
