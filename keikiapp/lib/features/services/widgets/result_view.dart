import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:flutter/services.dart';

import '../../../core/utils/formatters.dart';
import '../../../core/widgets/status_view.dart';
import '../../../core/widgets/ui.dart';
import '../data/humanize.dart';

const _titleKeys = ['title', 'name', 'display_name', 'displayName', 'full_name', 'username', 'label', 'code', 'question', 'slug', 'email'];
const _subtitleKeys = ['description', 'subtitle', 'summary', 'content', 'message', 'status', 'type', 'category', 'address', 'email'];
const _imageKeys = ['image', 'image_url', 'imageUrl', 'thumbnail', 'thumbnail_url', 'avatar', 'avatar_url', 'cover', 'poster', 'poster_url', 'photo', 'url', 'original'];
const _listKeys = ['items', 'data', 'results', 'nodes', 'edges', 'list', 'records', 'rows'];
const _moneyKeys = ['amount', 'price', 'balance', 'total', 'total_amount', 'subtotal', 'fee', 'available', 'cost'];

bool _isImageUrl(Object? v) =>
    v is String && v.startsWith('http') && RegExp(r'\.(png|jpe?g|gif|webp|svg)(\?.*)?$', caseSensitive: false).hasMatch(v);

String? _pick(Map m, List<String> keys) {
  for (final k in keys) {
    final v = m[k];
    if (v is String && v.trim().isNotEmpty) return v;
    if (v is num) return '$v';
  }
  return null;
}

String formatValue(String key, Object? v) {
  if (v == null) return '—';
  if (v is bool) return v ? 'Có' : 'Không';
  final k = key.toLowerCase();
  if (v is num && _moneyKeys.any((m) => k == m || k.endsWith('_$m'))) return formatMoney(v);
  if (v is String) {
    if (_moneyKeys.any((m) => k == m || k.endsWith('_$m')) && num.tryParse(v) != null) return formatMoneyString(v);
    if (RegExp(r'^\d{4}-\d{2}-\d{2}T').hasMatch(v)) return formatDate(v);
    return v;
  }
  if (v is num && (k.endsWith('_at') || k.endsWith('date')) && v > 1000000000) {
    final ms = v > 100000000000 ? v.toInt() : v.toInt() * 1000;
    return formatDate(DateTime.fromMillisecondsSinceEpoch(ms).toIso8601String());
  }
  return '$v';
}

class ResultView extends StatefulWidget {
  const ResultView({super.key, required this.value});

  final Object? value;

  @override
  State<ResultView> createState() => _ResultViewState();
}

class _ResultViewState extends State<ResultView> {
  bool _raw = false;

  @override
  Widget build(BuildContext context) {
    final json = const JsonEncoder.withIndent('  ').convert(widget.value);
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        Row(
          children: [
            SegmentedButton<bool>(
              showSelectedIcon: false,
              segments: const [
                ButtonSegment(value: false, label: Text('Kết quả'), icon: Icon(Icons.view_agenda_outlined)),
                ButtonSegment(value: true, label: Text('JSON'), icon: Icon(Icons.data_object)),
              ],
              selected: {_raw},
              onSelectionChanged: (s) => setState(() => _raw = s.first),
            ),
            const Spacer(),
            IconButton(
              tooltip: 'Sao chép JSON',
              icon: const Icon(Icons.copy_all_outlined),
              onPressed: () {
                Clipboard.setData(ClipboardData(text: json));
                showSnack(context, 'Đã sao chép kết quả');
              },
            ),
          ],
        ),
        const SizedBox(height: 12),
        if (_raw)
          Container(
            padding: const EdgeInsets.all(16),
            decoration: BoxDecoration(
              color: Theme.of(context).colorScheme.surfaceContainerHighest.withValues(alpha: 0.5),
              borderRadius: BorderRadius.circular(16),
            ),
            child: SelectableText(json, style: const TextStyle(fontFamily: 'monospace', fontSize: 12)),
          )
        else
          FriendlyValue(value: widget.value),
      ],
    );
  }
}

class FriendlyValue extends StatelessWidget {
  const FriendlyValue({super.key, required this.value, this.depth = 0});

  final Object? value;
  final int depth;

  @override
  Widget build(BuildContext context) {
    final v = value;
    if (v == null) return const Padding(padding: EdgeInsets.symmetric(vertical: 24), child: EmptyView(message: 'Không có dữ liệu'));
    if (v is List) return _ListResult(items: v, depth: depth);
    if (v is Map) {
      final listKey = _listKeys.where((k) => v[k] is List).firstOrNull;
      if (listKey != null && depth == 0) {
        final meta = {for (final e in v.entries) if (e.key != listKey && e.value is! List && e.value is! Map) e.key: e.value};
        return Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            if (meta.isNotEmpty)
              Padding(
                padding: const EdgeInsets.only(bottom: 12),
                child: Wrap(
                  spacing: 8,
                  runSpacing: 8,
                  children: [for (final e in meta.entries) Pill('${fieldLabel('${e.key}')}: ${formatValue('${e.key}', e.value)}')],
                ),
              ),
            _ListResult(items: v[listKey] as List, depth: depth),
          ],
        );
      }
      return _ObjectCard(map: v, depth: depth);
    }
    return _ScalarResult(value: v);
  }
}

class _ScalarResult extends StatelessWidget {
  const _ScalarResult({required this.value});

  final Object value;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final ok = value == true || value == 'ok' || value == 'OK';
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(20),
        child: Row(
          children: [
            IconBadge(icon: value is bool ? (value == true ? Icons.check_circle : Icons.cancel) : Icons.info_outline, color: ok || value is! bool ? theme.colorScheme.primary : theme.colorScheme.error),
            const SizedBox(width: 16),
            Expanded(child: SelectableText(value is bool ? (value == true ? 'Thành công' : 'Không') : '$value', style: theme.textTheme.titleMedium)),
          ],
        ),
      ),
    );
  }
}

class _ListResult extends StatelessWidget {
  const _ListResult({required this.items, required this.depth});

  final List items;
  final int depth;

  @override
  Widget build(BuildContext context) {
    if (items.isEmpty) return const Padding(padding: EdgeInsets.symmetric(vertical: 24), child: EmptyView(message: 'Danh sách trống'));
    if (items.every((e) => e is! Map && e is! List)) {
      return Wrap(spacing: 8, runSpacing: 8, children: [for (final e in items) Chip(label: Text('$e'))]);
    }
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        Padding(
          padding: const EdgeInsets.only(bottom: 8),
          child: Text('${items.length} mục', style: Theme.of(context).textTheme.labelLarge),
        ),
        for (final (i, e) in items.indexed)
          Padding(
            padding: const EdgeInsets.only(bottom: 10),
            child: e is Map ? _ItemCard(map: e, index: i) : FriendlyValue(value: e, depth: depth + 1),
          ),
      ],
    );
  }
}

class _ItemCard extends StatelessWidget {
  const _ItemCard({required this.map, required this.index});

  final Map map;
  final int index;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final title = _pick(map, _titleKeys) ?? (map['id'] != null ? '#${map['id']}' : 'Mục ${index + 1}');
    final subtitle = _pick(map, _subtitleKeys.where((k) => !_titleKeys.contains(k) || map[k] != title).toList());
    final image = _imageKeys.map((k) => map[k]).firstWhere(_isImageUrl, orElse: () => null) as String?;
    final status = map['status'] ?? map['state'];
    final money = _moneyKeys.where((k) => map[k] != null).firstOrNull;
    return Card(
      clipBehavior: Clip.antiAlias,
      child: InkWell(
        onTap: () => showModalBottomSheet(
          context: context,
          isScrollControlled: true,
          builder: (_) => DraggableScrollableSheet(
            expand: false,
            initialChildSize: 0.7,
            maxChildSize: 0.95,
            builder: (context, scroll) => ListView(
              controller: scroll,
              padding: const EdgeInsets.fromLTRB(20, 0, 20, 24),
              children: [
                Text(title, style: Theme.of(context).textTheme.titleLarge),
                const SizedBox(height: 12),
                _ObjectCard(map: map, depth: 1, flat: true),
              ],
            ),
          ),
        ),
        child: Padding(
          padding: const EdgeInsets.all(14),
          child: Row(
            children: [
              if (image != null)
                ClipRRect(
                  borderRadius: BorderRadius.circular(12),
                  child: Image.network(image, width: 56, height: 56, fit: BoxFit.cover, errorBuilder: (_, _, _) => const IconBadge(icon: Icons.image_not_supported_outlined, color: Colors.grey, size: 56)),
                )
              else
                IconBadge(icon: Icons.description_outlined, color: theme.colorScheme.primary),
              const SizedBox(width: 14),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(title, maxLines: 2, overflow: TextOverflow.ellipsis, style: theme.textTheme.titleSmall),
                    if (subtitle != null && subtitle != title && subtitle != '$status')
                      Text(subtitle, maxLines: 2, overflow: TextOverflow.ellipsis, style: theme.textTheme.bodySmall?.copyWith(color: theme.colorScheme.onSurfaceVariant)),
                    if (status != null) Padding(padding: const EdgeInsets.only(top: 6), child: Pill('$status')),
                  ],
                ),
              ),
              if (money != null) Text(formatValue(money, map[money]), style: theme.textTheme.titleSmall?.copyWith(color: theme.colorScheme.primary)),
              const Icon(Icons.chevron_right),
            ],
          ),
        ),
      ),
    );
  }
}

class _ObjectCard extends StatelessWidget {
  const _ObjectCard({required this.map, required this.depth, this.flat = false});

  final Map map;
  final int depth;
  final bool flat;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final scalars = [for (final e in map.entries) if (e.value is! Map && e.value is! List) e];
    final nested = [for (final e in map.entries) if (e.value is Map || e.value is List) e];
    final image = scalars.map((e) => e.value).firstWhere(_isImageUrl, orElse: () => null) as String?;
    final body = Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        if (image != null && depth <= 1)
          Padding(
            padding: const EdgeInsets.only(bottom: 12),
            child: ClipRRect(borderRadius: BorderRadius.circular(16), child: Image.network(image, height: 180, fit: BoxFit.cover, errorBuilder: (_, _, _) => const SizedBox())),
          ),
        for (final e in scalars)
          Padding(
            padding: const EdgeInsets.symmetric(vertical: 6),
            child: Row(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                SizedBox(
                  width: 130,
                  child: Text(fieldLabel('${e.key}'), style: theme.textTheme.bodySmall?.copyWith(color: theme.colorScheme.onSurfaceVariant)),
                ),
                Expanded(
                  child: e.value is bool
                      ? Align(
                          alignment: Alignment.centerLeft,
                          child: Pill(e.value == true ? 'Có' : 'Không', color: e.value == true ? Colors.green : theme.colorScheme.outline),
                        )
                      : SelectableText(formatValue('${e.key}', e.value), style: theme.textTheme.bodyMedium?.copyWith(fontWeight: FontWeight.w500)),
                ),
              ],
            ),
          ),
        for (final e in nested)
          Theme(
            data: theme.copyWith(dividerColor: Colors.transparent),
            child: ExpansionTile(
              tilePadding: EdgeInsets.zero,
              initiallyExpanded: depth == 0 && nested.length <= 2,
              title: Text(fieldLabel('${e.key}'), style: theme.textTheme.titleSmall),
              subtitle: e.value is List ? Text('${(e.value as List).length} mục') : null,
              children: [FriendlyValue(value: e.value, depth: depth + 1)],
            ),
          ),
      ],
    );
    if (flat) return body;
    return Card(child: Padding(padding: const EdgeInsets.all(16), child: body));
  }
}
