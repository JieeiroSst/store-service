import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:go_router/go_router.dart';

import '../../../core/utils/formatters.dart';
import '../../../core/widgets/status_view.dart';
import '../../address/data/address_repository.dart';
import '../../address/view/address_picker.dart';
import '../../auth/bloc/auth_bloc.dart';
import '../../wallet/bloc/wallet_cubit.dart';
import '../../wallet/data/wallet_models.dart';
import '../bloc/food_cubit.dart';
import '../data/food_repository.dart';

class FoodCheckoutPage extends StatefulWidget {
  const FoodCheckoutPage({super.key});

  @override
  State<FoodCheckoutPage> createState() => _FoodCheckoutPageState();
}

class _FoodCheckoutPageState extends State<FoodCheckoutPage> {
  static const _deliveryFee = 15000.0;
  static const _tips = [0.0, 5000.0, 10000.0, 20000.0];

  final _note = TextEditingController();
  UserAddress? _address;
  double _tip = 0;
  bool _placing = false;
  List<WalletCard> _cards = const [];
  WalletCard? _card;
  bool _loadingCards = true;
  String? _cardError;

  @override
  void initState() {
    super.initState();
    _loadCards();
    context.read<AddressRepository>().myAddresses().then((list) {
      if (!mounted || list.isEmpty || _address != null) return;
      setState(() => _address = list.where((a) => a.isDefaultShipping).firstOrNull ?? list.first);
    }).catchError((_) {});
  }

  @override
  void dispose() {
    _note.dispose();
    super.dispose();
  }

  Future<void> _loadCards() async {
    final wallet = context.read<WalletCubit>();
    final walletId = wallet.state.wallet?.id;
    if (walletId == null) {
      setState(() {
        _loadingCards = false;
        _cardError = 'Hãy kích hoạt Ví Keiki để thanh toán';
      });
      return;
    }
    try {
      final cards = (await wallet.repository.cards(walletId)).where((c) => c.isActive).toList();
      if (!mounted) return;
      setState(() {
        _cards = cards;
        _card = cards.firstOrNull;
        _loadingCards = false;
        _cardError = null;
      });
    } catch (e) {
      if (mounted) {
        setState(() {
          _loadingCards = false;
          _cardError = e.toString();
        });
      }
    }
  }

  Future<void> _issueCard() async {
    final wallet = context.read<WalletCubit>();
    final holder = context.read<AuthBloc>().state.user?.displayName ?? 'KEIKI USER';
    setState(() => _loadingCards = true);
    try {
      final card = await wallet.repository.issueVirtualCard(wallet.state.wallet!.id, holder.toUpperCase());
      if (!mounted) return;
      setState(() {
        _cards = [..._cards, card];
        _card = card;
        _loadingCards = false;
      });
    } catch (e) {
      if (mounted) {
        setState(() => _loadingCards = false);
        showSnack(context, e.toString());
      }
    }
  }

  Future<void> _place() async {
    final menu = context.read<FoodMenuCubit>();
    if (_address == null) {
      showSnack(context, 'Vui lòng chọn địa chỉ giao hàng');
      return;
    }
    if (_card == null || menu.state.restaurant == null) {
      showSnack(context, 'Vui lòng chọn thẻ thanh toán');
      return;
    }
    setState(() => _placing = true);
    try {
      final order = await context.read<FoodRepository>().placeOrder(
            restaurantId: '${menu.state.restaurant!.id}',
            deliveryAddressId: '${_address!.id}',
            paymentMethodId: _card!.id,
            items: menu.state.cart,
            deliveryFee: _deliveryFee,
            tip: _tip,
            note: _note.text,
          );
      menu.clearCart();
      if (mounted) context.pop(order.id);
    } catch (e) {
      if (mounted) {
        setState(() => _placing = false);
        showSnack(context, e.toString());
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final state = context.watch<FoodMenuCubit>().state;
    final total = state.cartTotal + _deliveryFee + _tip;
    return Scaffold(
      appBar: AppBar(title: const Text('Xác nhận đơn')),
      body: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          Card(
            child: ListTile(
              leading: const Icon(Icons.location_on),
              title: Text(_address == null ? 'Chọn địa chỉ giao hàng' : _address!.fullName),
              subtitle: _address == null ? null : Text(_address!.summary),
              trailing: const Icon(Icons.chevron_right),
              onTap: () async {
                final picked = await pickAddress(context);
                if (picked != null) setState(() => _address = picked);
              },
            ),
          ),
          const SizedBox(height: 8),
          for (final e in state.cart.entries)
            ListTile(
              dense: true,
              title: Text('${e.value} × ${e.key.name}'),
              trailing: Text(formatMoney(e.key.price * e.value)),
            ),
          const Divider(),
          TextField(controller: _note, decoration: const InputDecoration(labelText: 'Ghi chú cho quán', prefixIcon: Icon(Icons.edit_note))),
          const SizedBox(height: 16),
          Text('Tip cho tài xế', style: Theme.of(context).textTheme.titleSmall),
          Wrap(
            spacing: 8,
            children: [
              for (final t in _tips)
                ChoiceChip(label: Text(t == 0 ? 'Không' : formatMoney(t)), selected: _tip == t, onSelected: (_) => setState(() => _tip = t)),
            ],
          ),
          const SizedBox(height: 16),
          _Row('Tạm tính', formatMoney(state.cartTotal)),
          _Row('Phí giao hàng', formatMoney(_deliveryFee)),
          if (_tip > 0) _Row('Tip', formatMoney(_tip)),
          const Divider(),
          _Row('Tổng cộng', formatMoney(total), bold: true),
          const SizedBox(height: 8),
          const SizedBox(height: 8),
          Text('Thanh toán bằng Ví Keiki', style: Theme.of(context).textTheme.titleSmall),
          const SizedBox(height: 8),
          if (_loadingCards)
            const LinearProgressIndicator()
          else if (_cardError != null)
            Card(child: ListTile(leading: const Icon(Icons.info_outline), title: Text(_cardError!), trailing: TextButton(onPressed: _loadCards, child: const Text('Thử lại'))))
          else if (_cards.isEmpty)
            Card(
              child: ListTile(
                leading: const Icon(Icons.credit_card_off_outlined),
                title: const Text('Bạn chưa có thẻ Keiki'),
                subtitle: const Text('Tạo thẻ ảo miễn phí, tiền trừ trực tiếp từ ví'),
                trailing: FilledButton.tonal(onPressed: _issueCard, child: const Text('Tạo thẻ')),
              ),
            )
          else
            RadioGroup<String>(
              groupValue: _card?.id,
              onChanged: (id) => setState(() => _card = _cards.firstWhere((c) => c.id == id)),
              child: Column(
                children: [
                  for (final c in _cards)
                    RadioListTile<String>(
                      value: c.id,
                      secondary: const Icon(Icons.credit_card),
                      title: Text('${c.network} ${c.masked}'),
                      subtitle: Text(c.holderName),
                    ),
                ],
              ),
            ),
          if (state.restaurant != null)
            Padding(
              padding: const EdgeInsets.only(top: 8),
              child: Text('Nhà hàng: ${state.restaurant!.name}', style: Theme.of(context).textTheme.bodySmall),
            ),
        ],
      ),
      bottomNavigationBar: SafeArea(
        child: Padding(
          padding: const EdgeInsets.all(16),
          child: FilledButton(
            onPressed: _placing || state.cart.isEmpty || _card == null ? null : _place,
            child: _placing
                ? const SizedBox(height: 20, width: 20, child: CircularProgressIndicator(strokeWidth: 2))
                : Text('Đặt đơn • ${formatMoney(total)}'),
          ),
        ),
      ),
    );
  }
}

class _Row extends StatelessWidget {
  const _Row(this.label, this.value, {this.bold = false});

  final String label;
  final String value;
  final bool bold;

  @override
  Widget build(BuildContext context) {
    final style = bold ? Theme.of(context).textTheme.titleMedium : null;
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 4),
      child: Row(children: [Expanded(child: Text(label, style: style)), Text(value, style: style)]),
    );
  }
}
