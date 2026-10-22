import 'package:flutter/material.dart';

import '../../../gen/models.g.dart';
import '../kit/kit.dart';
import '../extras/extras.dart';

class RestaurantKitchenApp extends StatelessWidget {
  const RestaurantKitchenApp({super.key});

  @override
  Widget build(BuildContext context) {
    final api = context.api.restaurantKitchen;
    return ServiceApp(
      serviceId: 'restaurant_kitchen',
      tabs: [
        AppTab('Thực đơn', Icons.restaurant_menu, (context) => SearchableList<RKitchenFood>(
              hint: 'Tìm món',
              filters: () async => (await api.rKitchenCategories())?.map((c) => c.name_ ?? '').where((s) => s.isNotEmpty).toList(),
              load: (q, cat) async {
                final foods = (await api.rKitchenFoods(limit: 200))?.rows ?? const <RKitchenFood>[];
                final n = q.toLowerCase();
                return foods.where((f) => (cat == null || f.category?.name_ == cat) && (n.isEmpty || (f.name_ ?? '').toLowerCase().contains(n))).toList();
              },
              emptyMessage: 'Không có món phù hợp',
              emptyIcon: Icons.no_food_outlined,
              itemBuilder: (context, f, _) => EntityTile(
                title: f.name_ ?? 'Món',
                subtitle: f.category?.name_,
                caption: vnd(f.price),
                icon: Icons.ramen_dining_rounded,
                badges: [if (f.status == 0) const StatusPill('out_of_stock')],
              ),
            )),
        AppTab('Bếp', Icons.soup_kitchen_outlined, (context) => AsyncList<RKitchenKitchen>(
              load: () async => (await api.rKitchenKitchens(limit: 50))?.rows,
              emptyMessage: 'Chưa có bếp nào',
              itemBuilder: (context, k, _) => EntityTile(
                title: k.name_ ?? 'Bếp',
                subtitle: '${(k.foods ?? const []).length} món',
                icon: Icons.soup_kitchen_rounded,
                badges: [StatusPill(k.status)],
              ),
            )),
      ],
    );
  }
}

class RestaurantConsumerApp extends StatelessWidget {
  const RestaurantConsumerApp({super.key});

  @override
  Widget build(BuildContext context) {
    final kitchen = context.api.restaurantKitchen;
    return ServiceApp(
      serviceId: 'restaurant_consumer',
      tabs: [
        AppTab('Gọi món', Icons.room_service_outlined, (context) => AsyncList<RKitchenKitchen>(
              load: () async => (await kitchen.rKitchenKitchens(limit: 50))?.rows,
              header: const AppBanner(icon: Icons.room_service_rounded, title: 'Gọi món tại nhà hàng', subtitle: 'Chọn bếp, thêm món và gửi đơn thẳng xuống bếp'),
              emptyMessage: 'Chưa có nhà hàng mở cửa',
              emptyIcon: Icons.storefront_outlined,
              itemBuilder: (context, k, _) => EntityTile(
                title: k.name_ ?? 'Bếp',
                subtitle: '${(k.foods ?? const []).length} món đang phục vụ',
                icon: Icons.storefront_rounded,
                badges: [StatusPill(k.status)],
                onTap: () => openPage(context, _MenuPage(kitchen: k)),
              ),
            )),
      ],
    );
  }
}

class _MenuPage extends StatefulWidget {
  const _MenuPage({required this.kitchen});

  final RKitchenKitchen kitchen;

  @override
  State<_MenuPage> createState() => _MenuPageState();
}

class _MenuPageState extends State<_MenuPage> {
  final _qty = <int, int>{};

  @override
  Widget build(BuildContext context) {
    final foods = widget.kitchen.foods ?? const <RKitchenFood>[];
    final total = foods.fold<double>(0, (a, f) => a + (f.price ?? 0) * (_qty[f.id] ?? 0));
    final count = _qty.values.fold<int>(0, (a, b) => a + b);
    return DetailPage(
      title: widget.kitchen.name_ ?? 'Thực đơn',
      bottom: Row(
        children: [
          Expanded(
            child: Column(
              mainAxisSize: MainAxisSize.min,
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text('$count món', style: Theme.of(context).textTheme.bodySmall),
                Text(vnd(total), style: const TextStyle(fontWeight: FontWeight.w800, fontSize: 18)),
              ],
            ),
          ),
          FilledButton(
            onPressed: count == 0
                ? null
                : () async {
                    final ok = await showFormSheet(
                      context,
                      title: 'Xác nhận đơn',
                      description: '$count món • ${vnd(total)}',
                      fields: [
                        KField('name', 'Tên người nhận', required: true, initial: context.userName),
                        const KField('address', 'Bàn số / địa chỉ giao', required: true),
                        const KField('payment', 'Thanh toán', kind: FieldKind.choice, options: [('wallet', 'Ví Keiki'), ('cash', 'Tiền mặt'), ('card', 'Thẻ')]),
                      ],
                      submitLabel: 'Đặt món',
                      success: 'Đã gửi đơn xuống bếp',
                      onSubmit: (v) async {
                        final consumer = context.api.restaurantConsumer;
                        final ids = [for (final f in foods) if ((_qty[f.id] ?? 0) > 0) f.id];
                        final current = await context.api.restaurantKitchen.rKitchenFoodsByIDs(ids: ids.join(',')).catchError((_) => null);
                        final changed = (current ?? const <RKitchenFood>[]).where((c) => foods.any((f) => f.id == c.id && f.price != c.price)).toList();
                        if (changed.isNotEmpty) throw Exception('Giá đã thay đổi: ${changed.map((c) => '${c.name_} ${vnd(c.price)}').join(', ')}. Mở lại thực đơn để cập nhật.');
                        if ((current ?? const <RKitchenFood>[]).any((c) => c.status == 0)) throw Exception('Có món vừa hết, hãy chọn lại');
                        await consumer.rConsumerPlaceOrder(input: {
                        'name': v.str('name'),
                        'kitchen_id': widget.kitchen.id,
                        'delivery_name': v.str('name'),
                        'delivery_address': v.str('address'),
                        'payment_method': v.str('payment'),
                        'menu': [
                          for (final f in foods)
                            if ((_qty[f.id] ?? 0) > 0)
                              {
                                'id': f.id,
                                'name': f.name_,
                                'money': f.price,
                                'quantity': _qty[f.id],
                                'category': {'id': f.category?.id ?? f.categoryId, 'name': f.category?.name_},
                              },
                        ],
                      });
                      },
                    );
                    if (ok && mounted) setState(_qty.clear);
                  },
            child: const Text('Đặt món'),
          ),
        ],
      ),
      children: [
        if (foods.isEmpty) const Padding(padding: EdgeInsets.all(32), child: Center(child: Text('Bếp chưa có món'))),
        for (final f in foods)
          Padding(
            padding: const EdgeInsets.fromLTRB(16, 12, 16, 0),
            child: EntityTile(
              title: f.name_ ?? '',
              subtitle: f.category?.name_,
              caption: vnd(f.price),
              icon: Icons.ramen_dining_rounded,
              trailing: Row(
                mainAxisSize: MainAxisSize.min,
                children: [
                  if ((_qty[f.id] ?? 0) > 0) ...[
                    IconButton(visualDensity: VisualDensity.compact, onPressed: () => setState(() => _qty[f.id!] = (_qty[f.id] ?? 0) - 1), icon: const Icon(Icons.remove_circle_outline)),
                    Text('${_qty[f.id]}', style: const TextStyle(fontWeight: FontWeight.w800)),
                  ],
                  IconButton(visualDensity: VisualDensity.compact, onPressed: () => setState(() => _qty[f.id!] = (_qty[f.id] ?? 0) + 1), icon: Icon(Icons.add_circle, color: context.appColor)),
                ],
              ),
            ),
          ),
      ],
    );
  }
}

class RestaurantOrderApp extends StatefulWidget {
  const RestaurantOrderApp({super.key});

  @override
  State<RestaurantOrderApp> createState() => _RestaurantOrderAppState();
}

class _RestaurantOrderAppState extends State<RestaurantOrderApp> {
  int _version = 0;

  Future<void> _book(BuildContext context) async {
    final api = context.api.restaurantOrder;
    Object? res;
    DateTime? at;
    await showFormSheet(
      context,
      title: 'Đặt bàn',
      fields: [
        KField('name', 'Tên khách', required: true, initial: context.userName),
        const KField('size', 'Số khách', kind: FieldKind.number, required: true, initial: '2'),
        KField('date', 'Ngày', kind: FieldKind.date, required: true, initial: DateTime.now().toIso8601String().substring(0, 10)),
        const KField('time', 'Giờ (HH:mm)', required: true, initial: '19:00'),
        const KField('table', 'Khu vực / bàn mong muốn', kind: FieldKind.choice, options: [('Sảnh chính', 'Sảnh chính'), ('Ban công', 'Ban công'), ('Phòng riêng', 'Phòng riêng')]),
      ],
      submitLabel: 'Giữ bàn',
      onSubmit: (v) async {
        at = DateTime.parse('${v.str('date')}T${v.str('time')}:00');
        res = await api.rOrderCreateReservation(input: {
          'customer_name': v.str('name'),
          'party_size': v.integer('size'),
          'reserved_at': at!.toUtc().toIso8601String(),
          'table_name': v.str('table'),
        });
      },
    );
    if (res == null || !context.mounted) return;
    setState(() => _version++);
    await showResult(context, title: 'Đã giữ bàn', message: 'Hẹn gặp bạn lúc ${formatDate(at?.toIso8601String())}');
  }

  @override
  Widget build(BuildContext context) {
    final api = context.api.restaurantOrder;
    return ServiceApp(
      serviceId: 'restaurant_order',
      tabs: [
        AppTab('Đặt bàn', Icons.table_restaurant_outlined, (context) => AsyncList<ROrderReservation>(
              key: ValueKey(_version),
              load: () => api.myROrderReservations(limit: 50),
              header: AppBanner(
                icon: Icons.table_restaurant_rounded,
                title: 'Giữ bàn trước khi đến',
                subtitle: 'Chọn giờ và số khách, nhà hàng sẽ xếp bàn phù hợp',
                children: [
                  SizedBox(
                    width: double.infinity,
                    child: Builder(
                      builder: (context) => FilledButton.icon(
                        style: FilledButton.styleFrom(backgroundColor: Colors.white, foregroundColor: context.appColor),
                        icon: const Icon(Icons.event_seat_outlined),
                        label: const Text('Đặt bàn mới'),
                        onPressed: () => _book(context),
                      ),
                    ),
                  ),
                ],
              ),
              emptyMessage: 'Bạn chưa đặt bàn nào',
              emptyIcon: Icons.event_seat_outlined,
              itemBuilder: (context, r, reload) => EntityTile(
                title: '${r.tableName ?? 'Bàn'} • ${r.partySize ?? '-'} khách',
                subtitle: formatDate(r.reservedAt),
                caption: r.customerName,
                icon: Icons.table_restaurant_rounded,
                badges: [StatusPill(r.status)],
                trailing: r.status != 'booked'
                    ? null
                    : IconButton(
                        tooltip: 'Huỷ',
                        icon: const Icon(Icons.cancel_outlined),
                        onPressed: () async {
                          if (await runAction(context, () => api.rOrderCancelReservation(id: '${r.id}'), confirm: 'Huỷ đặt bàn lúc ${formatDate(r.reservedAt)}?', destructive: true, success: 'Đã huỷ đặt bàn')) {
                            await reload();
                          }
                        },
                      ),
              ),
            )),
      ],
    );
  }
}

class VendingApp extends StatelessWidget {
  const VendingApp({super.key});

  @override
  Widget build(BuildContext context) {
    final api = context.api.vending;
    return ServiceApp(
      serviceId: 'vending',
      actions: [Builder(builder: (context) => IconButton(tooltip: 'Sản phẩm', icon: const Icon(Icons.category_outlined), onPressed: () => openPage(context, const VendCatalogPage())))],
      tabs: [
        AppTab('Máy bán hàng', Icons.local_drink_outlined, (context) => AsyncList<VendMachine>(
              load: () async => (await api.vendMachines(limit: 50))?.items,
              emptyMessage: 'Chưa có máy nào gần bạn',
              emptyIcon: Icons.local_drink_outlined,
              itemBuilder: (context, m, _) => EntityTile(
                title: m.location ?? 'Máy ${m.id}',
                subtitle: m.model,
                icon: Icons.kitchen_rounded,
                badges: [StatusPill(m.status)],
                onTap: () => openPage(context, _MachinePage(machine: m)),
              ),
            )),
      ],
    );
  }
}

class _MachinePage extends StatelessWidget {
  const _MachinePage({required this.machine});

  final VendMachine machine;

  Future<void> _buy(BuildContext context, VendInventory slot) async {
    final api = context.api.vending;
    final p = slot.product;
    final ok = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: Text(p?.name_ ?? 'Sản phẩm'),
        content: Text('Mua 1 sản phẩm ở khe ${slot.slot} với giá ${vnd((p?.priceCents ?? 0) / 100)}?'),
        actions: [
          TextButton(onPressed: () => Navigator.pop(context, false), child: const Text('Huỷ')),
          FilledButton(onPressed: () => Navigator.pop(context, true), child: const Text('Thanh toán bằng Ví')),
        ],
      ),
    );
    if (ok != true || !context.mounted) return;
    VendOrder? order;
    final done = await runAction(context, () async {
      final s = await api.vendStartSession(machineId: machine.id!);
      if (s?.id == null) throw Exception('Máy đang bận, thử lại sau');
      try {
        await api.vendReserve(id: s!.id!, slot: slot.slot!);
        try {
          order = await api.vendCheckout(id: s.id!, paymentMethod: 'wallet');
        } catch (_) {
          await api.vendCancelReservation(id: s.id!).catchError((_) => null);
          rethrow;
        }
      } finally {
        await api.vendEndSession(id: s!.id!).catchError((_) => null);
      }
      return order;
    });
    if (!done || !context.mounted) return;
    await showResult(context, icon: Icons.local_drink_rounded, title: 'Mời lấy hàng!', message: 'Đơn #${order?.orderNo ?? ''} • ${order?.status ?? ''}\nSản phẩm đang được nhả ở khe ${slot.slot}.');
  }

  @override
  Widget build(BuildContext context) {
    final api = context.api.vending;
    return AsyncBuilder<VendInventoryPage?>(
      load: () => api.vendInventory(id: machine.id!, limit: 100),
      loading: const Scaffold(body: SkeletonList()),
      builder: (context, inv, reload) {
        final items = inv?.items ?? const <VendInventory>[];
        return DetailPage(
          title: machine.location ?? 'Máy bán hàng',
          header: AsyncBuilder<String?>(
            load: () => vendMachineInfo(context, machine.id!),
            loading: AppBanner(icon: Icons.kitchen_rounded, title: machine.location ?? '', subtitle: '${machine.model ?? ''} • ${items.length} khe hàng'),
            builder: (context, info, _) => AppBanner(icon: Icons.kitchen_rounded, title: machine.location ?? '', subtitle: '${info ?? machine.model ?? ''} • ${items.length} khe hàng'),
          ),
          children: [
            Padding(
              padding: const EdgeInsets.all(16),
              child: GridView.count(
                crossAxisCount: 3,
                shrinkWrap: true,
                physics: const NeverScrollableScrollPhysics(),
                mainAxisSpacing: 10,
                crossAxisSpacing: 10,
                childAspectRatio: 0.66,
                children: [
                  for (final s in items)
                    KCard(
                      padding: const EdgeInsets.all(8),
                      onTap: (s.quantity ?? 0) > 0 ? () => _buy(context, s) : null,
                      child: Column(
                        children: [
                          Expanded(child: LayoutBuilder(builder: (context, c) => KThumb(url: s.product?.imageUrl, icon: Icons.local_drink_rounded, width: c.maxWidth, height: c.maxHeight, radius: 12))),
                          const SizedBox(height: 6),
                          Text(s.product?.name_ ?? '', maxLines: 1, overflow: TextOverflow.ellipsis, style: const TextStyle(fontWeight: FontWeight.w700, fontSize: 12)),
                          Text(vnd((s.product?.priceCents ?? 0) / 100), style: TextStyle(color: context.appColor, fontWeight: FontWeight.w800, fontSize: 12)),
                          Text((s.quantity ?? 0) > 0 ? 'Khe ${s.slot} • còn ${s.quantity}' : 'Hết hàng', style: const TextStyle(fontSize: 10)),
                        ],
                      ),
                    ),
                ],
              ),
            ),
          ],
        );
      },
    );
  }
}
