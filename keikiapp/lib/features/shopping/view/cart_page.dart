import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:go_router/go_router.dart';

import '../../../core/utils/formatters.dart';
import '../../../core/widgets/status_view.dart';
import '../../../core/widgets/ui.dart';
import '../../address/data/address_repository.dart';
import '../../address/view/address_picker.dart';
import '../../../core/network/graphql_exception.dart';
import '../../../gen/models.g.dart';
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
  BasketShippingQuote? _quote;
  String? _shipNote;
  bool _quoting = false;
  double? _quotedFor;

  String _message(Object e) => e is GraphQLException ? e.message : e.toString();

  Future<BasketShippingQuote?> _requote() async {
    final a = _address;
    if (a == null) return null;
    final cart = context.read<CartCubit>();
    setState(() {
      _quoting = true;
      _shipNote = null;
    });
    try {
      final q = await cart.shippingQuote(a.id);
      if (!mounted || _address != a) return null;
      setState(() {
        _quote = q;
        _quotedFor = cart.state.total;
      });
      return q;
    } catch (e) {
      if (mounted && _address == a) {
        setState(() {
          _quote = null;
          _shipNote = _message(e);
        });
      }
      return null;
    } finally {
      if (mounted) setState(() => _quoting = false);
    }
  }

  Future<void> _setAddress(UserAddress a) async {
    setState(() {
      _address = a;
      _quote = null;
    });
    await _requote();
  }

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
      await _setAddress(picked);
    }
    if (!mounted) return;
    final quote = _quote ?? await _requote();
    if (!mounted) return;
    if (quote == null) {
      showSnack(context, _shipNote ?? 'Chưa tính được phí giao hàng, hãy thử lại');
      return;
    }
    final pin = await askPin(context);
    if (pin == null || pin.isEmpty || !mounted) return;
    try {
      final done = await cart.checkout(
        walletId: wallet.id,
        pin: pin,
        addressId: _address?.id,
        shippingFee: quote.shippingCharged == true ? quote.shippingFee : null,
      );
      if (!mounted || done == null) return;
      context.read<WalletCubit>().load();
      final ship = done.shippingInclTax ?? 0;
      await showDialog<void>(
        context: context,
        builder: (context) => AlertDialog(
          icon: const Icon(Icons.check_circle, color: Colors.green, size: 48),
          title: const Text('Đặt hàng thành công'),
          content: Text(
            'Đã thanh toán ${formatMoney(done.totalInclTax, done.currency)} từ Ví Keiki'
            '${ship > 0 ? ', gồm phí giao hàng ${formatMoney(ship, done.currency)}' : ''}. Đơn sẽ được giao qua GHN.',
          ),
          actions: [FilledButton(onPressed: () => Navigator.pop(context), child: const Text('Xem đơn mua'))],
        ),
      );
      if (mounted) context.pushReplacement('/shop/purchases');
    } catch (e) {
      if (!mounted) return;
      if (e is GraphQLException && e.message.contains('phí giao hàng')) {
        final fresh = await _requote();
        if (!mounted) return;
        await showDialog<void>(
          context: context,
          builder: (context) => AlertDialog(
            icon: const Icon(Icons.local_shipping_outlined, size: 40),
            title: const Text('Phí giao hàng đã thay đổi'),
            content: Text(
              fresh == null
                  ? 'Không tính lại được phí giao hàng, hãy thử lại sau.'
                  : 'Phí giao hàng mới là ${formatMoney(fresh.shippingFee, fresh.currency)}, tổng ${formatMoney(fresh.total, fresh.currency)}. Bạn chưa bị trừ tiền, hãy kiểm tra rồi bấm Thanh toán lại.',
            ),
            actions: [FilledButton(onPressed: () => Navigator.pop(context), child: const Text('Đã hiểu'))],
          ),
        );
        return;
      }
      showSnack(context, _message(e));
    }
  }

  @override
  Widget build(BuildContext context) {
    return BlocConsumer<CartCubit, CartState>(
      listenWhen: (p, c) => p.total != c.total,
      listener: (context, state) {
        if (_address != null && _quotedFor != state.total && state.lines.isNotEmpty) _requote();
      },
      builder: (context, state) {
        final q = _quote;
        final charged = q?.shippingCharged == true;
        final total = q?.total ?? state.total;
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
                            Text(
                              charged ? 'Tổng cộng (gồm ship ${formatMoney(q!.shippingFee, q.currency)})' : 'Tổng cộng',
                              style: theme.textTheme.bodySmall,
                              maxLines: 1,
                              overflow: TextOverflow.ellipsis,
                            ),
                            Text(
                              formatMoney(total, state.currency),
                              style: theme.textTheme.titleLarge?.copyWith(color: theme.colorScheme.primary),
                            ),
                          ],
                        ),
                      ),
                      SizedBox(
                        width: 180,
                        child: LoadingButton(label: 'Thanh toán', loading: state.busy || _quoting, onPressed: _checkout),
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
                    subtitle: _address == null
                        ? null
                        : Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              Text(_address!.summary),
                              Padding(
                                padding: const EdgeInsets.only(top: 6),
                                child: Row(
                                  children: [
                                    Icon(Icons.local_shipping_outlined, size: 16, color: _shipNote != null ? theme.colorScheme.error : theme.colorScheme.primary),
                                    const SizedBox(width: 6),
                                    Expanded(
                                      child: Text(
                                        _quoting
                                            ? 'Đang tính phí giao hàng…'
                                            : _shipNote ??
                                                (q == null
                                                    ? 'Chưa tính phí giao hàng'
                                                    : charged
                                                        ? 'Phí giao hàng GHN ${formatMoney(q.shippingFee, q.currency)}${q.serviceName == null ? '' : ' • ${q.serviceName}'}'
                                                        : 'Miễn phí giao hàng'),
                                        style: theme.textTheme.bodySmall?.copyWith(color: _shipNote != null ? theme.colorScheme.error : theme.colorScheme.primary),
                                      ),
                                    ),
                                  ],
                                ),
                              ),
                            ],
                          ),
                    trailing: const Icon(Icons.chevron_right),
                    onTap: () async {
                      final picked = await pickAddress(context);
                      if (picked != null && context.mounted) await _setAddress(picked);
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
