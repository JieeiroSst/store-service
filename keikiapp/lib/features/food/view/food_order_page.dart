import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

import '../../../core/utils/formatters.dart';
import '../../../core/widgets/status_view.dart';
import '../data/food_repository.dart';
import 'food_home_page.dart';

class FoodOrderPage extends StatefulWidget {
  const FoodOrderPage({super.key, required this.orderId});

  final int orderId;

  @override
  State<FoodOrderPage> createState() => _FoodOrderPageState();
}

class _FoodOrderPageState extends State<FoodOrderPage> {
  FoodOrder? _order;
  List<TrackingEvent> _tracking = const [];
  String? _error;
  Timer? _poll;

  @override
  void initState() {
    super.initState();
    _load();
    _poll = Timer.periodic(const Duration(seconds: 15), (_) {
      if (_order?.isActive ?? true) _load();
    });
  }

  @override
  void dispose() {
    _poll?.cancel();
    super.dispose();
  }

  Future<void> _load() async {
    final repo = context.read<FoodRepository>();
    try {
      final results = await Future.wait([repo.order(widget.orderId), repo.tracking(widget.orderId)]);
      if (!mounted) return;
      setState(() {
        _order = results[0] as FoodOrder?;
        _tracking = results[1] as List<TrackingEvent>;
        _error = _order == null ? 'Không tìm thấy đơn' : null;
      });
    } catch (e) {
      if (mounted) setState(() => _error = e.toString());
    }
  }

  Future<void> _cancel() async {
    final ok = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Huỷ đơn?'),
        actions: [
          TextButton(onPressed: () => Navigator.pop(context, false), child: const Text('Không')),
          FilledButton(onPressed: () => Navigator.pop(context, true), child: const Text('Huỷ đơn')),
        ],
      ),
    );
    if (ok != true || !mounted) return;
    try {
      await context.read<FoodRepository>().cancel(widget.orderId, 'Khách huỷ trên app');
      await _load();
    } catch (e) {
      if (mounted) showSnack(context, e.toString());
    }
  }

  @override
  Widget build(BuildContext context) {
    final order = _order;
    return Scaffold(
      appBar: AppBar(title: Text('Đơn #${widget.orderId}')),
      body: order == null
          ? (_error != null ? ErrorView(message: _error!, onRetry: _load) : const Center(child: CircularProgressIndicator()))
          : RefreshIndicator(
              onRefresh: _load,
              child: ListView(
                padding: const EdgeInsets.all(16),
                children: [
                  Row(
                    children: [
                      Expanded(child: Text('Trạng thái', style: Theme.of(context).textTheme.titleMedium)),
                      OrderStatusChip(status: order.status),
                    ],
                  ),
                  if (order.estimatedDelivery != null) Text('Dự kiến giao: ${formatDate(order.estimatedDelivery)}'),
                  const SizedBox(height: 16),
                  if (_tracking.isNotEmpty) ...[
                    Text('Hành trình', style: Theme.of(context).textTheme.titleMedium),
                    for (final (i, t) in _tracking.reversed.indexed)
                      ListTile(
                        dense: true,
                        leading: Icon(i == 0 ? Icons.radio_button_checked : Icons.radio_button_unchecked,
                            color: i == 0 ? Theme.of(context).colorScheme.primary : null),
                        title: Text(t.status),
                        subtitle: Text([formatDate(t.timestamp), if (t.notes.isNotEmpty) t.notes].join(' • ')),
                      ),
                    const Divider(),
                  ],
                  for (final item in order.items)
                    ListTile(
                      dense: true,
                      title: Text('${item.quantity} × ${item.name}'),
                      trailing: Text(formatMoney(item.unitPrice * item.quantity)),
                    ),
                  const Divider(),
                  ListTile(dense: true, title: const Text('Phí giao hàng'), trailing: Text(formatMoney(order.deliveryFee))),
                  ListTile(
                    title: const Text('Tổng cộng', style: TextStyle(fontWeight: FontWeight.bold)),
                    subtitle: Text('Thanh toán: ${order.paymentStatus}'),
                    trailing: Text(formatMoney(order.total), style: const TextStyle(fontWeight: FontWeight.bold)),
                  ),
                  if (order.isCancellable)
                    OutlinedButton.icon(onPressed: _cancel, icon: const Icon(Icons.cancel_outlined), label: const Text('Huỷ đơn')),
                ],
              ),
            ),
    );
  }
}
