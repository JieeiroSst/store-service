import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:go_router/go_router.dart';

import '../../../core/utils/formatters.dart';
import '../../../core/widgets/status_view.dart';
import '../../../core/widgets/ui.dart';
import '../../address/data/address_repository.dart';
import '../../address/view/address_picker.dart';
import '../../wallet/bloc/wallet_cubit.dart';
import '../bloc/cart_cubit.dart';
import 'shop_home_page.dart';

class CartPage extends StatefulWidget {
  const CartPage({super.key});

  @override
  State<CartPage> createState() => _CartPageState();
}

class _CartPageState extends State<CartPage> {
  UserAddress? _address;

  Future<void> _checkout() async {
    final cart = context.read<CartCubit>();
    final wallet = context.read<WalletCubit>().state.wallet;
    if (wallet == null) {
      showSnack(context, 'Hãy kích hoạt Ví Keiki trước khi thanh toán');
      context.push('/wallet');
      return;
    }
    if (_address == null) {
      final picked = await pickAddress(context);
      if (picked == null || !mounted) return;
      setState(() => _address = picked);
    }
    if (!mounted) return;
    final pin = await askPin(context);
    if (pin == null || pin.isEmpty || !mounted) return;
    try {
      final done = await cart.checkout(walletId: wallet.id, pin: pin, addressId: _address?.id);
      if (!mounted || done == null) return;
      context.read<WalletCubit>().load();
      await showDialog<void>(
        context: context,
        builder: (context) => AlertDialog(
          icon: const Icon(Icons.check_circle, color: Colors.green, size: 48),
          title: const Text('Đặt hàng thành công'),
          content: Text('Đã thanh toán ${formatMoney(done.totalInclTax, done.currency)} từ Ví Keiki.'),
          actions: [FilledButton(onPressed: () => Navigator.pop(context), child: const Text('Xem đơn mua'))],
        ),
      );
      if (mounted) context.pushReplacement('/shop/purchases');
    } catch (e) {
      if (mounted) showSnack(context, e.toString());
    }
  }

  @override
  Widget build(BuildContext context) {
    return BlocBuilder<CartCubit, CartState>(
      builder: (context, state) {
        final cubit = context.read<CartCubit>();
        final theme = Theme.of(context);
        return Scaffold(
          appBar: AppBar(title: Text('Giỏ hàng (${state.count})')),
          bottomNavigationBar: state.lines.isEmpty
              ? null
              : BottomActionBar(
                  child: Row(
                    children: [
                      Expanded(
                        child: Column(
                          mainAxisSize: MainAxisSize.min,
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text('Tổng cộng', style: theme.textTheme.bodySmall),
                            Text(
                              formatMoney(state.total, state.currency),
                              style: theme.textTheme.titleLarge?.copyWith(color: theme.colorScheme.primary),
                            ),
                          ],
                        ),
                      ),
                      SizedBox(
                        width: 180,
                        child: LoadingButton(label: 'Thanh toán', loading: state.busy, onPressed: _checkout),
                      ),
                    ],
                  ),
                ),
          body: StatusView(
            status: state.status,
            error: state.error,
            onRetry: cubit.load,
            isEmpty: state.lines.isEmpty,
            emptyMessage: 'Giỏ hàng đang trống',
            builder: (context) => ListView(
              padding: const EdgeInsets.all(16),
              children: [
                Card(
                  child: ListTile(
                    leading: const Icon(Icons.location_on_outlined),
                    title: Text(_address == null ? 'Chọn địa chỉ giao hàng' : _address!.fullName),
                    subtitle: _address == null ? null : Text(_address!.summary),
                    trailing: const Icon(Icons.chevron_right),
                    onTap: () async {
                      final picked = await pickAddress(context);
                      if (picked != null) setState(() => _address = picked);
                    },
                  ),
                ),
                const SizedBox(height: 12),
                for (final line in state.lines)
                  Padding(
                    padding: const EdgeInsets.only(bottom: 10),
                    child: Card(
                      child: Padding(
                        padding: const EdgeInsets.all(12),
                        child: Row(
                          children: [
                            ClipRRect(
                              borderRadius: BorderRadius.circular(12),
                              child: SizedBox(width: 64, height: 64, child: ProductImage(url: line.product?.thumbnail)),
                            ),
                            const SizedBox(width: 12),
                            Expanded(
                              child: Column(
                                crossAxisAlignment: CrossAxisAlignment.start,
                                children: [
                                  Text(line.title, maxLines: 2, overflow: TextOverflow.ellipsis, style: theme.textTheme.titleSmall),
                                  const SizedBox(height: 4),
                                  Text(
                                    formatMoney(line.unitPrice, line.line.priceCurrency),
                                    style: theme.textTheme.bodyMedium?.copyWith(color: theme.colorScheme.primary),
                                  ),
                                ],
                              ),
                            ),
                            Row(
                              mainAxisSize: MainAxisSize.min,
                              children: [
                                IconButton(
                                  icon: Icon(line.quantity == 1 ? Icons.delete_outline : Icons.remove),
                                  onPressed: state.busy
                                      ? null
                                      : () => cubit.setQuantity(line, line.quantity - 1).catchError((Object e) {
                                          if (context.mounted) showSnack(context, e.toString());
                                        }),
                                ),
                                Text('${line.quantity}', style: theme.textTheme.titleSmall),
                                IconButton(
                                  icon: const Icon(Icons.add),
                                  onPressed: state.busy
                                      ? null
                                      : () => cubit.setQuantity(line, line.quantity + 1).catchError((Object e) {
                                          if (context.mounted) showSnack(context, e.toString());
                                        }),
                                ),
                              ],
                            ),
                          ],
                        ),
                      ),
                    ),
                  ),
              ],
            ),
          ),
        );
      },
    );
  }
}
