import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

import '../../../core/utils/formatters.dart';
import '../../../core/widgets/status_view.dart';
import '../data/shopping_repository.dart';

class CouponsPage extends StatefulWidget {
  const CouponsPage({super.key});

  @override
  State<CouponsPage> createState() => _CouponsPageState();
}

class _CouponsPageState extends State<CouponsPage> {
  final _code = TextEditingController();
  late Future<List<UserCoupon>> _mine = context.read<ShoppingRepository>().myCoupons();
  Coupon? _found;
  bool _checking = false;

  @override
  void dispose() {
    _code.dispose();
    super.dispose();
  }

  Future<void> _check() async {
    final code = _code.text.trim();
    if (code.isEmpty) return;
    setState(() => _checking = true);
    try {
      final coupon = await context.read<ShoppingRepository>().couponByCode(code);
      if (!mounted) return;
      setState(() => _found = coupon);
      if (coupon == null) showSnack(context, 'Mã không tồn tại');
    } catch (e) {
      if (mounted) showSnack(context, e.toString());
    } finally {
      if (mounted) setState(() => _checking = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Mã giảm giá')),
      body: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          Row(
            children: [
              Expanded(
                child: TextField(
                  controller: _code,
                  textCapitalization: TextCapitalization.characters,
                  decoration: const InputDecoration(labelText: 'Nhập mã', prefixIcon: Icon(Icons.local_offer_outlined)),
                  onSubmitted: (_) => _check(),
                ),
              ),
              const SizedBox(width: 8),
              FilledButton(onPressed: _checking ? null : _check, child: const Text('Kiểm tra')),
            ],
          ),
          if (_found != null) _CouponCard(coupon: _found!),
          const SizedBox(height: 24),
          Text('Coupon của tôi', style: Theme.of(context).textTheme.titleMedium),
          FutureBuilder<List<UserCoupon>>(
            future: _mine,
            builder: (context, snap) {
              if (snap.hasError) {
                return ErrorView(
                  message: snap.error.toString(),
                  onRetry: () => setState(() => _mine = context.read<ShoppingRepository>().myCoupons()),
                );
              }
              if (!snap.hasData) {
                return const Padding(
                  padding: EdgeInsets.all(24),
                  child: Center(child: CircularProgressIndicator()),
                );
              }
              if (snap.data!.isEmpty) return const Padding(padding: EdgeInsets.all(16), child: Text('Bạn chưa được tặng coupon nào'));
              return Column(
                children: [
                  for (final c in snap.data!)
                    if (c.coupon != null)
                      Opacity(
                        opacity: c.isUsed ? 0.5 : 1,
                        child: _CouponCard(coupon: c.coupon!, used: c.isUsed),
                      )
                    else
                      ListTile(
                        leading: const Icon(Icons.confirmation_number_outlined),
                        title: Text('Coupon #${c.couponId}'),
                        subtitle: Text('Nhận lúc ${formatDate(c.assignedAt)}'),
                        trailing: Text(c.isUsed ? 'Đã dùng' : 'Còn hiệu lực'),
                      ),
                ],
              );
            },
          ),
        ],
      ),
    );
  }
}

class _CouponCard extends StatelessWidget {
  const _CouponCard({required this.coupon, this.used = false});

  final Coupon coupon;
  final bool used;

  @override
  Widget build(BuildContext context) {
    final value = coupon.isPercentage ? '${coupon.value.toStringAsFixed(0)}%' : formatMoney(coupon.value);
    return Card(
      margin: const EdgeInsets.only(top: 16),
      color: Theme.of(context).colorScheme.secondaryContainer,
      child: ListTile(
        leading: const Icon(Icons.discount),
        title: Text(used ? 'Giảm $value • Đã dùng' : 'Giảm $value'),
        subtitle: Text(
          [
            coupon.code,
            if (coupon.description.isNotEmpty) coupon.description,
            if (coupon.minimumPurchase > 0) 'Đơn tối thiểu ${formatMoney(coupon.minimumPurchase)}',
            if (coupon.maxDiscount != null) 'Giảm tối đa ${formatMoney(coupon.maxDiscount)}',
            if (coupon.endDate != null) 'HSD ${formatDate(coupon.endDate)}',
          ].join('\n'),
        ),
        trailing: IconButton(
          icon: const Icon(Icons.copy),
          onPressed: () {
            Clipboard.setData(ClipboardData(text: coupon.code));
            showSnack(context, 'Đã sao chép ${coupon.code}');
          },
        ),
      ),
    );
  }
}
