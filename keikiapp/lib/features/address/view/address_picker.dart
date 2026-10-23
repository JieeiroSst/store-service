import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

import '../../../core/widgets/status_view.dart';
import '../data/address_repository.dart';
import '../../../gen/gateway_api.g.dart';
import 'ghn_area_picker.dart';

Future<UserAddress?> pickAddress(BuildContext context) {
  return showModalBottomSheet<UserAddress>(
    context: context,
    isScrollControlled: true,
    showDragHandle: true,
    builder: (_) => RepositoryProvider.value(
      value: context.read<AddressRepository>(),
      child: const _AddressSheet(),
    ),
  );
}

class _AddressSheet extends StatefulWidget {
  const _AddressSheet();

  @override
  State<_AddressSheet> createState() => _AddressSheetState();
}

class _AddressSheetState extends State<_AddressSheet> {
  late Future<List<UserAddress>> _future = context.read<AddressRepository>().myAddresses();

  Future<void> _add() async {
    final created = await showDialog<UserAddress>(
      context: context,
      builder: (_) => RepositoryProvider.value(value: context.read<AddressRepository>(), child: const _AddressDialog()),
    );
    if (created != null && mounted) Navigator.pop(context, created);
  }

  @override
  Widget build(BuildContext context) {
    return SafeArea(
      child: SizedBox(
        height: MediaQuery.sizeOf(context).height * 0.6,
        child: FutureBuilder<List<UserAddress>>(
          future: _future,
          builder: (context, snap) {
            if (snap.hasError) {
              return ErrorView(
                message: snap.error.toString(),
                onRetry: () => setState(() => _future = context.read<AddressRepository>().myAddresses()),
              );
            }
            if (!snap.hasData) return const Center(child: CircularProgressIndicator());
            return ListView(
              children: [
                const ListTile(title: Text('Chọn địa chỉ giao hàng', style: TextStyle(fontWeight: FontWeight.bold))),
                for (final a in snap.data!)
                  ListTile(
                    leading: Icon(a.isDefaultShipping ? Icons.home : Icons.location_on_outlined),
                    title: Text(a.fullName.isEmpty ? a.line1 : '${a.fullName} • ${a.phone}'),
                    subtitle: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(a.summary),
                        if (!a.hasGhnArea)
                          Text(
                            'Cần bổ sung Tỉnh/Quận/Phường để giao hàng',
                            style: TextStyle(color: Theme.of(context).colorScheme.error, fontSize: 12),
                          ),
                      ],
                    ),
                    onTap: () async {
                      if (a.hasGhnArea) {
                        Navigator.pop(context, a);
                        return;
                      }
                      final fixed = await fixAddressArea(context, a);
                      if (fixed != null && context.mounted) Navigator.pop(context, fixed);
                    },
                  ),
                ListTile(leading: const Icon(Icons.add_location_alt_outlined), title: const Text('Thêm địa chỉ mới'), onTap: _add),
              ],
            );
          },
        ),
      ),
    );
  }
}

class _AddressDialog extends StatefulWidget {
  const _AddressDialog();

  @override
  State<_AddressDialog> createState() => _AddressDialogState();
}

class _AddressDialogState extends State<_AddressDialog> {
  final _form = GlobalKey<FormState>();
  final _name = TextEditingController();
  final _phone = TextEditingController();
  final _line1 = TextEditingController();
  GhnArea _area = const GhnArea();
  bool _saving = false;

  @override
  void dispose() {
    for (final c in [_name, _phone, _line1]) {
      c.dispose();
    }
    super.dispose();
  }

  Future<void> _save() async {
    if (!_form.currentState!.validate()) return;
    setState(() => _saving = true);
    final parts = _name.text.trim().split(RegExp(r'\s+'));
    try {
      final created = await context.read<AddressRepository>().create(
            firstName: parts.length > 1 ? parts.sublist(0, parts.length - 1).join(' ') : parts.first,
            lastName: parts.length > 1 ? parts.last : '',
            line1: _line1.text.trim(),
            line2: [_area.wardName, _area.districtName].where((s) => s.isNotEmpty).join(', '),
            state: _area.provinceName,
            postcode: _area.complete ? _area.code : null,
            phone: _phone.text.trim(),
            isDefault: true,
          );
      if (mounted) Navigator.pop(context, created);
    } catch (e) {
      if (mounted) {
        setState(() => _saving = false);
        showSnack(context, e.toString());
      }
    }
  }

  String? _required(String? v) => (v == null || v.trim().isEmpty) ? 'Bắt buộc' : null;

  @override
  Widget build(BuildContext context) {
    return AlertDialog(
      title: const Text('Địa chỉ mới'),
      content: Form(
        key: _form,
        child: SingleChildScrollView(
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              TextFormField(controller: _name, decoration: const InputDecoration(labelText: 'Người nhận'), validator: _required),
              TextFormField(controller: _phone, keyboardType: TextInputType.phone, decoration: const InputDecoration(labelText: 'Số điện thoại'), validator: _required),
              TextFormField(controller: _line1, decoration: const InputDecoration(labelText: 'Số nhà, đường'), validator: _required),
              GhnAreaPicker(onChanged: (a) => _area = a),
            ],
          ),
        ),
      ),
      actions: [
        TextButton(onPressed: () => Navigator.pop(context), child: const Text('Huỷ')),
        FilledButton(onPressed: _saving ? null : _save, child: const Text('Lưu')),
      ],
    );
  }
}

Future<UserAddress?> fixAddressArea(BuildContext context, UserAddress address) => showDialog<UserAddress>(
      context: context,
      builder: (_) => RepositoryProvider.value(value: context.read<AddressRepository>(), child: _FixAreaDialog(address: address)),
    );

class _FixAreaDialog extends StatefulWidget {
  const _FixAreaDialog({required this.address});

  final UserAddress address;

  @override
  State<_FixAreaDialog> createState() => _FixAreaDialogState();
}

class _FixAreaDialogState extends State<_FixAreaDialog> {
  final _form = GlobalKey<FormState>();
  GhnArea? _guess;
  GhnArea _area = const GhnArea();
  bool _saving = false;

  @override
  void initState() {
    super.initState();
    final a = widget.address;
    guessGhnArea(context.read<GatewayApi>().shipping, province: a.state, district: a.districtText, ward: a.wardText).then((g) {
      if (mounted) setState(() => _guess = _area = g);
    }).catchError((_) {
      if (mounted) setState(() => _guess = const GhnArea());
    });
  }

  Future<void> _save() async {
    if (!_form.currentState!.validate() || !_area.complete) return;
    setState(() => _saving = true);
    try {
      final updated = await context.read<AddressRepository>().setArea(
            widget.address,
            line2: '${_area.wardName}, ${_area.districtName}',
            state: _area.provinceName,
            postcode: _area.code,
          );
      if (mounted) Navigator.pop(context, updated);
    } catch (e) {
      if (mounted) {
        setState(() => _saving = false);
        showSnack(context, e.toString());
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final guess = _guess;
    return AlertDialog(
      title: const Text('Bổ sung khu vực giao hàng'),
      content: Form(
        key: _form,
        child: SingleChildScrollView(
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              Text('${widget.address.fullName}\n${widget.address.summary}', style: Theme.of(context).textTheme.bodySmall),
              const SizedBox(height: 8),
              if (guess == null)
                const Padding(padding: EdgeInsets.all(16), child: Center(child: CircularProgressIndicator()))
              else ...[
                if (guess.complete)
                  const Text('Đã tự tìm khu vực theo địa chỉ cũ, kiểm tra lại rồi bấm Lưu.')
                else
                  const Text('Chọn Tỉnh/Quận/Phường theo danh mục của GHN.'),
                GhnAreaPicker(initial: guess, onChanged: (a) => _area = a),
              ],
            ],
          ),
        ),
      ),
      actions: [
        TextButton(onPressed: () => Navigator.pop(context), child: const Text('Huỷ')),
        FilledButton(onPressed: _saving || guess == null ? null : _save, child: const Text('Lưu')),
      ],
    );
  }
}
