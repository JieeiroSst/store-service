import 'package:equatable/equatable.dart';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

import '../../../gen/gateway_api.g.dart';

class GhnArea extends Equatable {
  const GhnArea({this.provinceId, this.provinceName = '', this.districtId, this.districtName = '', this.wardCode, this.wardName = ''});

  final int? provinceId;
  final String provinceName;
  final int? districtId;
  final String districtName;
  final String? wardCode;
  final String wardName;

  bool get complete => provinceId != null && districtId != null && wardCode != null;

  String get code => complete ? 'ghn:$provinceId:$districtId:$wardCode' : '';

  static GhnArea? parse(String? code) {
    final p = (code ?? '').split(':');
    if (p.length != 4 || p[0] != 'ghn') return null;
    final province = int.tryParse(p[1]);
    final district = int.tryParse(p[2]);
    if (province == null || district == null || p[3].isEmpty) return null;
    return GhnArea(provinceId: province, districtId: district, wardCode: p[3]);
  }

  @override
  List<Object?> get props => [provinceId, districtId, wardCode];
}

class _Option {
  const _Option(this.key, this.name);

  final String key;
  final String name;
}

String foldVi(String s) {
  const from = 'àáạảãâầấậẩẫăằắặẳẵèéẹẻẽêềếệểễìíịỉĩòóọỏõôồốộổỗơờớợởỡùúụủũưừứựửữỳýỵỷỹđ';
  const to = 'aaaaaaaaaaaaaaaaaeeeeeeeeeeeiiiiiooooooooooooooooouuuuuuuuuuuyyyyyd';
  final b = StringBuffer();
  for (final ch in s.toLowerCase().split('')) {
    final i = from.indexOf(ch);
    b.write(i >= 0 ? to[i] : ch);
  }
  return b.toString();
}

const _areaPrefixes = ['thanh pho ', 'tinh ', 'tp. ', 'tp ', 'tp.', 'quan ', 'huyen ', 'thi xa ', 'thi tran ', 'phuong ', 'xa ', 'q. ', 'p. '];

String areaKey(String name) {
  var s = foldVi(name).replaceAll(RegExp(r'\s+'), ' ').trim();
  var changed = true;
  while (changed) {
    changed = false;
    for (final p in _areaPrefixes) {
      if (s.startsWith(p) && s.length > p.length) {
        s = s.substring(p.length).trim();
        changed = true;
      }
    }
  }
  return s.replaceFirst(RegExp(r'^0+(?=\d)'), '');
}

T? _bestMatch<T>(Iterable<T> items, String? wanted, String Function(T) name) {
  final key = areaKey(wanted ?? '');
  if (key.isEmpty) return null;
  for (final i in items) {
    if (areaKey(name(i)) == key) return i;
  }
  final loose = [for (final i in items) if (areaKey(name(i)).contains(key) || key.contains(areaKey(name(i)))) i];
  return loose.length == 1 ? loose.first : null;
}

Future<GhnArea> guessGhnArea(ShippingApi api, {String? province, String? district, String? ward}) async {
  final p = _bestMatch(await api.shipProvinces() ?? const <ShipProvince>[], province, (ShipProvince p) => p.name_ ?? '');
  if (p?.id == null) return const GhnArea();
  var area = GhnArea(provinceId: p!.id, provinceName: p.name_ ?? '');
  final d = _bestMatch(await api.shipDistricts(provinceId: p.id!) ?? const <ShipDistrict>[], district, (ShipDistrict d) => d.name_ ?? '');
  if (d?.id == null) return area;
  area = GhnArea(provinceId: area.provinceId, provinceName: area.provinceName, districtId: d!.id, districtName: d.name_ ?? '');
  final w = _bestMatch(await api.shipWards(districtId: d.id!) ?? const <ShipWard>[], ward, (ShipWard w) => w.name_ ?? '');
  if (w?.code == null) return area;
  return GhnArea(
    provinceId: area.provinceId,
    provinceName: area.provinceName,
    districtId: area.districtId,
    districtName: area.districtName,
    wardCode: w!.code,
    wardName: w.name_ ?? '',
  );
}

class GhnAreaPicker extends StatefulWidget {
  const GhnAreaPicker({super.key, required this.onChanged, this.required = true, this.initial = const GhnArea()});

  final ValueChanged<GhnArea> onChanged;
  final bool required;
  final GhnArea initial;

  @override
  State<GhnAreaPicker> createState() => _GhnAreaPickerState();
}

class _GhnAreaPickerState extends State<GhnAreaPicker> {
  late GhnArea _area = widget.initial;

  ShippingApi get _api => context.read<GatewayApi>().shipping;

  Future<void> _choose(String title, Future<List<_Option>> Function() load, void Function(_Option) onPick) async {
    final picked = await showModalBottomSheet<_Option>(
      context: context,
      isScrollControlled: true,
      showDragHandle: true,
      useSafeArea: true,
      builder: (context) => _OptionSheet(title: title, load: load),
    );
    if (picked != null && mounted) {
      setState(() => onPick(picked));
      widget.onChanged(_area);
    }
  }

  Widget _tile(String label, String value, bool enabled, VoidCallback onTap, FormFieldValidator<String> validator) => Padding(
        padding: const EdgeInsets.only(top: 8),
        child: FormField<String>(
          key: ValueKey('$label$value'),
          initialValue: value,
          validator: validator,
          builder: (state) => InkWell(
            borderRadius: BorderRadius.circular(16),
            onTap: enabled ? onTap : null,
            child: InputDecorator(
              isEmpty: value.isEmpty,
              decoration: InputDecoration(labelText: label, errorText: state.errorText, enabled: enabled, suffixIcon: const Icon(Icons.expand_more)),
              child: Text(value, maxLines: 1, overflow: TextOverflow.ellipsis),
            ),
          ),
        ),
      );

  @override
  Widget build(BuildContext context) {
    String? need(String? v) => widget.required && (v == null || v.isEmpty) ? 'Bắt buộc' : null;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        _tile(
          'Tỉnh/Thành phố',
          _area.provinceName,
          true,
          () => _choose(
            'Chọn tỉnh/thành phố',
            () async => [for (final p in await _api.shipProvinces() ?? const <ShipProvince>[]) if (p.id != null) _Option('${p.id}', p.name_ ?? '')],
            (o) => _area = GhnArea(provinceId: int.parse(o.key), provinceName: o.name),
          ),
          need,
        ),
        _tile(
          'Quận/Huyện',
          _area.districtName,
          _area.provinceId != null,
          () => _choose(
            'Chọn quận/huyện',
            () async => [for (final d in await _api.shipDistricts(provinceId: _area.provinceId!) ?? const <ShipDistrict>[]) if (d.id != null) _Option('${d.id}', d.name_ ?? '')],
            (o) => _area = GhnArea(provinceId: _area.provinceId, provinceName: _area.provinceName, districtId: int.parse(o.key), districtName: o.name),
          ),
          need,
        ),
        _tile(
          'Phường/Xã',
          _area.wardName,
          _area.districtId != null,
          () => _choose(
            'Chọn phường/xã',
            () async => [for (final w in await _api.shipWards(districtId: _area.districtId!) ?? const <ShipWard>[]) if (w.code != null) _Option(w.code!, w.name_ ?? '')],
            (o) => _area = GhnArea(
              provinceId: _area.provinceId,
              provinceName: _area.provinceName,
              districtId: _area.districtId,
              districtName: _area.districtName,
              wardCode: o.key,
              wardName: o.name,
            ),
          ),
          need,
        ),
      ],
    );
  }
}

class _OptionSheet extends StatefulWidget {
  const _OptionSheet({required this.title, required this.load});

  final String title;
  final Future<List<_Option>> Function() load;

  @override
  State<_OptionSheet> createState() => _OptionSheetState();
}

class _OptionSheetState extends State<_OptionSheet> {
  late final Future<List<_Option>> _future = widget.load();
  String _q = '';

  @override
  Widget build(BuildContext context) {
    return SizedBox(
      height: MediaQuery.sizeOf(context).height * 0.75,
      child: Column(
        children: [
          Padding(
            padding: const EdgeInsets.fromLTRB(16, 0, 16, 8),
            child: Text(widget.title, style: Theme.of(context).textTheme.titleMedium),
          ),
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: 16),
            child: TextField(
              autofocus: false,
              decoration: const InputDecoration(prefixIcon: Icon(Icons.search), hintText: 'Tìm kiếm'),
              onChanged: (v) => setState(() => _q = foldVi(v.trim())),
            ),
          ),
          Expanded(
            child: FutureBuilder<List<_Option>>(
              future: _future,
              builder: (context, snap) {
                if (snap.hasError) return Center(child: Text('Không tải được danh sách: ${snap.error}', textAlign: TextAlign.center));
                if (!snap.hasData) return const Center(child: CircularProgressIndicator());
                final items = snap.data!.where((o) => _q.isEmpty || foldVi(o.name).contains(_q)).toList();
                if (items.isEmpty) return const Center(child: Text('Không có kết quả'));
                return ListView.builder(
                  itemCount: items.length,
                  itemBuilder: (context, i) => ListTile(title: Text(items[i].name), onTap: () => Navigator.pop(context, items[i])),
                );
              },
            ),
          ),
        ],
      ),
    );
  }
}

int? ghnFee(Object? quote) {
  if (quote is num) return quote.toInt();
  if (quote is String) return int.tryParse(quote);
  if (quote is List) {
    final fees = quote.map(ghnFee).whereType<int>().toList()..sort();
    return fees.isEmpty ? null : fees.first;
  }
  if (quote is Map) {
    final recommended = quote['recommended'];
    if (recommended is Map) {
      for (final k in const ['CHEAPEST', 'BALANCED', 'FASTEST']) {
        final v = ghnFee(recommended[k]);
        if (v != null) return v;
      }
    }
    for (final k in const ['fee', 'total', 'total_fee', 'service_fee', 'amount']) {
      final v = ghnFee(quote[k]);
      if (v != null) return v;
    }
    for (final k in const ['options', 'data', 'quote', 'quotes', 'items', 'result']) {
      final v = ghnFee(quote[k]);
      if (v != null) return v;
    }
  }
  return null;
}

class GhnItem {
  const GhnItem(this.name, this.quantity, {this.price = 0, this.code});

  final String name;
  final int quantity;
  final int price;
  final String? code;
}

const ghnItemWeightGram = 500;

ShipParcelInput ghnParcel(List<GhnItem> items, {String? content}) {
  final units = items.fold<int>(0, (n, i) => n + i.quantity);
  final n = units < 1 ? 1 : units;
  return ShipParcelInput(
    weightGram: n * ghnItemWeightGram,
    lengthCm: 30,
    widthCm: 20,
    heightCm: 10 + 5 * (n > 6 ? 6 : n),
    content: content,
    items: [
      for (final i in items.isEmpty ? const [GhnItem('Hàng hoá', 1)] : items)
        ShipParcelItemInput(name_: i.name, code: i.code, quantity: i.quantity, weightGram: ghnItemWeightGram, price: i.price),
    ],
  );
}

Future<int?> quoteGhn(
  GatewayApi api,
  GhnArea area, {
  required String name,
  required String phone,
  required String street,
  required List<GhnItem> items,
  int cod = 0,
  int insurance = 0,
}) async {
  final r = await api.shipping.shipQuote(
    input: ShipQuoteInput(
      recipient: ShipAddressInput(name_: name, phone: phone, street: street, provinceId: area.provinceId, districtId: area.districtId, wardCode: area.wardCode),
      parcel: ghnParcel(items),
      codAmount: cod,
      insuranceValue: insurance > 5000000 ? 5000000 : insurance,
    ),
  );
  return ghnFee(r);
}
