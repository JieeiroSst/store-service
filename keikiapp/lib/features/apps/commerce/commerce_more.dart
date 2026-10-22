import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:flutter/services.dart';

import '../../../gen/models.g.dart';
import '../kit/kit.dart';

class WalletMorePage extends StatefulWidget {
  const WalletMorePage({super.key});

  @override
  State<WalletMorePage> createState() => _WalletMorePageState();
}

class _WalletMorePageState extends State<WalletMorePage> {
  int _v = 0;

  @override
  Widget build(BuildContext context) {
    final api = context.api.walletService;
    return AsyncBuilder<(WalletSvcWallet?, WalletSvcLimits?, WalletSvcStatement?, List<WalletSvcBankAccount>, List<WalletSvcPaymentRequest>)>(
      key: ValueKey(_v),
      load: () async {
        final wallets = await api.walletSvcMyWallets() ?? const <WalletSvcWallet>[];
        final currency = wallets.isEmpty ? 'VND' : wallets.first.currency ?? 'VND';
        final w = await api.walletSvcMyWallet(currency: currency).catchError((_) => null);
        if (w?.id == null) return (w, null, null, const <WalletSvcBankAccount>[], const <WalletSvcPaymentRequest>[]);
        final now = DateTime.now();
        final r = await Future.wait<Object?>([
          api.walletSvcMyLimits(walletId: w!.id!).catchError((_) => null),
          api.walletSvcMyStatement(walletId: w.id!, from: DateTime(now.year, now.month).toUtc().toIso8601String(), to: now.toUtc().toIso8601String()).catchError((_) => null),
          api.walletSvcMyBankAccounts().catchError((_) => null),
          api.walletSvcMyPaymentRequests(role: 'requester', status: 'pending', pageSize: 20).then((p) => p?.items).catchError((_) => null),
        ]);
        return (w, r[0] as WalletSvcLimits?, r[1] as WalletSvcStatement?, (r[2] as List<WalletSvcBankAccount>?) ?? const <WalletSvcBankAccount>[], (r[3] as List<WalletSvcPaymentRequest>?) ?? const <WalletSvcPaymentRequest>[]);
      },
      loading: const Scaffold(body: SkeletonList()),
      builder: (context, d, _) {
        final w = d.$1;
        final lim = d.$2;
        final st = d.$3;
        void bump() => setState(() => _v++);
        return DetailPage(
          title: 'Quản lý ví',
          color: const Color(0xFF4338CA),
          header: AppBanner(
            icon: Icons.account_balance_wallet_rounded,
            title: formatMoneyString(w?.balance, w?.currency),
            subtitle: 'Hạng ${w?.tier ?? '-'} • đang giữ ${formatMoneyString(w?.frozenAmount, w?.currency)}',
            children: [
              StatRow([
                StatTile(label: 'Còn dùng hôm nay', value: formatMoneyString(lim?.remainingToday, lim?.currency), onDark: true),
                StatTile(label: 'Mỗi giao dịch', value: formatMoneyString(lim?.limits?.perTransaction, lim?.currency), onDark: true),
              ]),
            ],
          ),
          children: [
            InfoSection(title: 'Sao kê tháng này', rows: [
              ('Số dư đầu kỳ', formatMoneyString(st?.openingBalance, st?.currency)),
              ('Tiền vào', '${formatMoneyString(st?.totalCredit, st?.currency)} (${st?.creditCount ?? 0} lần)'),
              ('Tiền ra', '${formatMoneyString(st?.totalDebit, st?.currency)} (${st?.debitCount ?? 0} lần)'),
              ('Số dư cuối kỳ', formatMoneyString(st?.closingBalance, st?.currency)),
              ('Hạn mức ngày', formatMoneyString(lim?.limits?.daily, lim?.currency)),
              ('Đặt lại lúc', lim?.resetsAt == null ? null : formatDate(lim!.resetsAt)),
            ]),
            ActionButtons([
              if (w?.id != null)
                OutlinedButton.icon(icon: const Icon(Icons.list_alt), label: const Text('Sổ cái'), onPressed: () => openPage(context, _Ledger(walletId: w!.id!, currency: w.currency))),
            ]),
            const SectionHeader('Tài khoản ngân hàng'),
            if (d.$4.isEmpty) const Padding(padding: EdgeInsets.symmetric(horizontal: 20), child: Text('Chưa liên kết tài khoản')),
            for (final b in d.$4)
              Padding(
                padding: const EdgeInsets.fromLTRB(16, 0, 16, 8),
                child: EntityTile(
                  title: b.accountName ?? '',
                  subtitle: '•••• ${(b.accountNumber ?? '').length > 4 ? b.accountNumber!.substring(b.accountNumber!.length - 4) : b.accountNumber}',
                  icon: Icons.account_balance_rounded,
                  badges: [if (b.isDefault == true) const Pill('Mặc định'), StatusPill(b.status)],
                  trailing: PopupMenuButton<String>(
                    onSelected: (a) async {
                      final ok = a == 'default'
                          ? await runAction(context, () => api.walletSvcMySetDefaultBankAccount(accountId: b.id!), success: 'Đã đặt làm mặc định')
                          : await runAction(context, () => api.walletSvcMyRemoveBankAccount(accountId: b.id!), confirm: 'Gỡ tài khoản ${b.accountName}?', destructive: true, success: 'Đã gỡ');
                      if (ok) bump();
                    },
                    itemBuilder: (_) => [
                      if (b.isDefault != true) const PopupMenuItem(value: 'default', child: Text('Đặt mặc định')),
                      const PopupMenuItem(value: 'remove', child: Text('Gỡ liên kết')),
                    ],
                  ),
                ),
              ),
            const SectionHeader('Yêu cầu thanh toán đang chờ'),
            if (d.$5.isEmpty) const Padding(padding: EdgeInsets.symmetric(horizontal: 20), child: Text('Không có yêu cầu đang chờ')),
            for (final p in d.$5)
              Padding(
                padding: const EdgeInsets.fromLTRB(16, 0, 16, 8),
                child: EntityTile(
                  title: p.note ?? 'Yêu cầu ${p.id}',
                  subtitle: 'Hết hạn ${formatDate(p.expiresAt)}',
                  caption: formatMoneyString(p.amount, p.currency),
                  icon: Icons.request_quote_rounded,
                  trailing: TextButton(
                    onPressed: () async {
                      if (await runAction(context, () => api.walletSvcMyCancelPaymentRequest(id: p.id!), success: 'Đã huỷ yêu cầu')) bump();
                    },
                    child: const Text('Huỷ'),
                  ),
                ),
              ),
            const SectionHeader('Lịch sử điểm thưởng'),
            const _PointHistory(),
            if (w?.id != null)
              ActionButtons([
                TextButton.icon(
                  style: TextButton.styleFrom(foregroundColor: const Color(0xFFDC2626)),
                  icon: const Icon(Icons.no_accounts_outlined),
                  label: const Text('Đóng ví'),
                  onPressed: () async {
                    if (await runAction(
                      context,
                      () => api.walletSvcMyCloseWallet(walletId: w!.id!),
                      confirm: 'Đóng ví vĩnh viễn? Số dư phải bằng 0 và không thể mở lại ví này.',
                      destructive: true,
                      success: 'Đã đóng ví',
                    )) {
                      bump();
                    }
                  },
                ),
              ]),
          ],
        );
      },
    );
  }
}

class _PointHistory extends StatelessWidget {
  const _PointHistory();

  @override
  Widget build(BuildContext context) {
    return AsyncBuilder<PointTransactionPage?>(
      load: () => context.api.point.pointMyTransactions(limit: 20).catchError((_) => null),
      loading: const SizedBox(height: 40),
      builder: (context, page, _) => Column(
        children: [
          if ((page?.items ?? const []).isEmpty) const Padding(padding: EdgeInsets.all(16), child: Text('Chưa có điểm thưởng')),
          for (final t in page?.items ?? const <PointTransaction>[])
            ListTile(
              dense: true,
              leading: Icon(Icons.stars_rounded, color: (t.points ?? 0) >= 0 ? const Color(0xFFD97706) : Colors.grey),
              title: Text(t.title ?? t.source ?? 'Điểm thưởng'),
              subtitle: Text(formatDate(t.occurredAt)),
              trailing: Text('${(t.points ?? 0) >= 0 ? '+' : ''}${t.points ?? 0} • còn ${t.balanceAfter ?? 0}', style: const TextStyle(fontWeight: FontWeight.w700)),
            ),
        ],
      ),
    );
  }
}

class _Ledger extends StatelessWidget {
  const _Ledger({required this.walletId, this.currency});

  final String walletId;
  final String? currency;

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Sổ cái ví')),
      body: AsyncList<WalletSvcLedgerEntry>(
        load: () async => (await context.api.walletService.walletSvcMyLedger(walletId: walletId, pageSize: 100))?.items,
        emptyMessage: 'Chưa có bút toán',
        spacing: 4,
        itemBuilder: (context, e, _) {
          final credit = (e.direction ?? '').toLowerCase().startsWith('c');
          return ListTile(
            leading: Icon(credit ? Icons.south_west_rounded : Icons.north_east_rounded, color: credit ? const Color(0xFF16A34A) : const Color(0xFFDC2626)),
            title: Text('${credit ? '+' : '-'}${formatMoneyString(e.amount, currency)}', style: const TextStyle(fontWeight: FontWeight.w800)),
            subtitle: Text('${formatDate(e.createdAt)} • GD ${e.transactionId ?? ''}'),
            trailing: Text('Số dư\n${formatMoneyString(e.balanceAfter, currency)}', textAlign: TextAlign.right, style: const TextStyle(fontSize: 12)),
          );
        },
      ),
    );
  }
}

class ShopMorePage extends StatelessWidget {
  const ShopMorePage({super.key});

  @override
  Widget build(BuildContext context) {
    const color = Color(0xFFDB2777);
    return AppColor(
      color: color,
      child: DefaultTabController(
        length: 4,
        child: Scaffold(
          appBar: AppBar(
            foregroundColor: Colors.white,
            backgroundColor: color,
            flexibleSpace: DecoratedBox(decoration: BoxDecoration(gradient: appGradient(color))),
            title: const Text('Mua sắm nâng cao', style: TextStyle(color: Colors.white, fontWeight: FontWeight.w700)),
            bottom: const TabBar(
              isScrollable: true,
              tabAlignment: TabAlignment.start,
              labelColor: Colors.white,
              unselectedLabelColor: Colors.white70,
              indicatorColor: Colors.white,
              dividerColor: Colors.transparent,
              tabs: [Tab(text: 'Danh sách mong muốn'), Tab(text: 'Giỏ & đơn'), Tab(text: 'Mã giảm giá'), Tab(text: 'Danh mục')],
            ),
          ),
          body: const TabBarView(children: [_Wishlists(), _Baskets(), _Coupons(), _Catalogue()]),
        ),
      ),
    );
  }
}

class _Wishlists extends StatefulWidget {
  const _Wishlists();

  @override
  State<_Wishlists> createState() => _WishlistsState();
}

class _WishlistsState extends State<_Wishlists> {
  int _v = 0;

  @override
  Widget build(BuildContext context) {
    final api = context.api.wishlists;
    return AsyncList<Wishlist>(
      key: ValueKey(_v),
      load: () async => (await api.wishlistsOfUser(userId: '${context.userId}', limit: 50))?.items,
      header: ActionButtons([
        FilledButton.icon(
          icon: const Icon(Icons.playlist_add),
          label: const Text('Danh sách mới'),
          onPressed: () async {
            if (await showFormSheet(
              context,
              title: 'Danh sách mong muốn mới',
              fields: const [
                KField('name', 'Tên', required: true, hint: 'Quà sinh nhật, Đồ cho bé…'),
                KField('visibility', 'Chia sẻ', kind: FieldKind.choice, options: [('Private', 'Chỉ mình tôi'), ('Shared', 'Ai có link'), ('Public', 'Công khai')]),
              ],
              success: 'Đã tạo',
              onSubmit: (v) => api.wishlistCreate(input: WishlistCreateInput(name_: v.str('name'), visibility: v.str('visibility'))),
            )) {
              setState(() => _v++);
            }
          },
        ),
        OutlinedButton.icon(
          icon: const Icon(Icons.download_outlined),
          label: const Text('Mở link chia sẻ'),
          onPressed: () async {
            Wishlist? found;
            String? key;
            await showFormSheet(
              context,
              title: 'Danh sách được chia sẻ',
              fields: const [KField('key', 'Mã chia sẻ', required: true)],
              submitLabel: 'Xem',
              onSubmit: (v) async {
                key = v.str('key');
                found = await api.wishlistShared(key: key!);
              },
            );
            if (found == null || !context.mounted) return;
            final copy = await showDialog<bool>(
              context: context,
              builder: (context) => AlertDialog(
                title: Text(found!.name_ ?? 'Danh sách'),
                content: Text([for (final l in found!.lines ?? const <WishlistLine>[]) '• ${l.title} × ${l.quantity ?? 1}'].join('\n')),
                actions: [
                  TextButton(onPressed: () => Navigator.pop(context, false), child: const Text('Đóng')),
                  FilledButton(onPressed: () => Navigator.pop(context, true), child: const Text('Lưu bản sao')),
                ],
              ),
            );
            if (copy == true && context.mounted && await runAction(context, () => api.wishlistCopyShared(key: key!), success: 'Đã lưu bản sao')) setState(() => _v++);
          },
        ),
      ]),
      emptyMessage: 'Chưa có danh sách nào',
      emptyIcon: Icons.favorite_border,
      itemBuilder: (context, w, reload) => EntityTile(
        title: w.name_ ?? 'Danh sách',
        subtitle: '${(w.lines ?? const []).length} sản phẩm • ${w.visibility ?? ''}',
        icon: Icons.favorite_rounded,
        onTap: () async {
          await openPage(context, _WishlistPage(id: w.id!));
          await reload();
        },
      ),
    );
  }
}

class _WishlistPage extends StatefulWidget {
  const _WishlistPage({required this.id});

  final int id;

  @override
  State<_WishlistPage> createState() => _WishlistPageState();
}

class _WishlistPageState extends State<_WishlistPage> {
  int _v = 0;

  @override
  Widget build(BuildContext context) {
    final api = context.api.wishlists;
    final id = widget.id;
    void bump() => setState(() => _v++);
    return AsyncBuilder<(Wishlist?, List<Wishlist>)>(
      key: ValueKey(_v),
      load: () async {
        final r = await Future.wait<Object?>([api.wishlist(id: id), api.wishlistsOfUser(userId: '${context.userId}', limit: 50).then((p) => p?.items).catchError((_) => null)]);
        return (r[0] as Wishlist?, (r[1] as List<Wishlist>?) ?? const <Wishlist>[]);
      },
      loading: const Scaffold(body: SkeletonList()),
      builder: (context, d, _) {
        final w = d.$1;
        final others = d.$2.where((o) => o.id != id).toList();
        return DetailPage(
          title: w?.name_ ?? 'Danh sách',
          actions: [
            PopupMenuButton<String>(
              onSelected: (a) async {
                var ok = false;
                if (a == 'edit') {
                  ok = await showFormSheet(
                    context,
                    title: 'Sửa danh sách',
                    fields: [
                      KField('name', 'Tên', required: true, initial: w?.name_),
                      KField('visibility', 'Chia sẻ', kind: FieldKind.choice, initial: w?.visibility, options: const [('Private', 'Chỉ mình tôi'), ('Shared', 'Ai có link'), ('Public', 'Công khai')]),
                      const KField('regen', 'Tạo link chia sẻ mới', kind: FieldKind.toggle, initial: 'false'),
                    ],
                    success: 'Đã lưu',
                    onSubmit: (v) => api.wishlistUpdate(id: id, input: WishlistUpdateInput(name_: v.str('name'), visibility: v.str('visibility'), regenerateKey: v.flag('regen'))),
                  );
                } else if (a == 'share') {
                  await Clipboard.setData(ClipboardData(text: w?.key ?? ''));
                  if (context.mounted) showSnack(context, 'Đã sao chép mã chia sẻ');
                } else if (a == 'clear') {
                  ok = await runAction(context, () => api.wishlistClearLines(id: id), confirm: 'Xoá hết sản phẩm khỏi danh sách?', destructive: true, success: 'Đã xoá hết');
                } else if (a == 'delete') {
                  if (await runAction(context, () => api.wishlistDelete(id: id), confirm: 'Xoá danh sách này?', destructive: true, success: 'Đã xoá danh sách') && context.mounted) {
                    Navigator.pop(context);
                    return;
                  }
                }
                if (ok) bump();
              },
              itemBuilder: (_) => const [
                PopupMenuItem(value: 'edit', child: Text('Sửa')),
                PopupMenuItem(value: 'share', child: Text('Sao chép mã chia sẻ')),
                PopupMenuItem(value: 'clear', child: Text('Xoá hết sản phẩm')),
                PopupMenuItem(value: 'delete', child: Text('Xoá danh sách')),
              ],
            ),
          ],
          bottom: FilledButton.icon(
            icon: const Icon(Icons.add),
            label: const Text('Thêm sản phẩm'),
            onPressed: () async {
              if (await showFormSheet(
                context,
                title: 'Thêm sản phẩm',
                fields: const [
                  KField('product', 'Mã sản phẩm', kind: FieldKind.number, required: true),
                  KField('title', 'Tên sản phẩm', required: true),
                  KField('qty', 'Số lượng', kind: FieldKind.number, initial: '1'),
                ],
                success: 'Đã thêm',
                onSubmit: (v) => api.wishlistAddLine(id: id, input: WishlistLineInput(productId: v.integer('product')!, title: v.str('title')!, quantity: v.integer('qty'))),
              )) {
                bump();
              }
            },
          ),
          children: [
            if ((w?.lines ?? const []).isEmpty) const Padding(padding: EdgeInsets.all(32), child: Center(child: Text('Danh sách đang trống'))),
            for (final l in w?.lines ?? const <WishlistLine>[])
              Padding(
                padding: const EdgeInsets.fromLTRB(16, 12, 16, 0),
                child: EntityTile(
                  title: l.title ?? 'Sản phẩm #${l.productId}',
                  subtitle: 'Mã ${l.productId}',
                  icon: Icons.shopping_bag_rounded,
                  trailing: Row(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      IconButton(
                        visualDensity: VisualDensity.compact,
                        icon: const Icon(Icons.remove_circle_outline),
                        onPressed: (l.quantity ?? 1) <= 1
                            ? null
                            : () async {
                                if (await runAction(context, () => api.wishlistUpdateLine(id: id, lineId: l.id!, quantity: (l.quantity ?? 1) - 1))) bump();
                              },
                      ),
                      Text('${l.quantity ?? 1}', style: const TextStyle(fontWeight: FontWeight.w800)),
                      IconButton(
                        visualDensity: VisualDensity.compact,
                        icon: const Icon(Icons.add_circle_outline),
                        onPressed: () async {
                          if (await runAction(context, () => api.wishlistUpdateLine(id: id, lineId: l.id!, quantity: (l.quantity ?? 1) + 1))) bump();
                        },
                      ),
                      if (others.isNotEmpty)
                        PopupMenuButton<int>(
                          tooltip: 'Chuyển sang danh sách khác',
                          icon: const Icon(Icons.drive_file_move_outline),
                          onSelected: (target) async {
                            if (await runAction(context, () => api.wishlistMoveLine(id: id, lineId: l.id!, targetWishlistId: target), success: 'Đã chuyển')) bump();
                          },
                          itemBuilder: (_) => [for (final o in others) PopupMenuItem(value: o.id!, child: Text(o.name_ ?? ''))],
                        ),
                    ],
                  ),
                ),
              ),
          ],
        );
      },
    );
  }
}

class _Baskets extends StatelessWidget {
  const _Baskets();

  @override
  Widget build(BuildContext context) {
    final api = context.api.basket;
    return AsyncBuilder<(List<BasketOrder>, List<BasketBasket>)>(
      load: () async {
        final r = await Future.wait<Object?>([api.basketOrders().catchError((_) => null), api.basketBaskets().catchError((_) => null)]);
        return ((r[0] as List<BasketOrder>?) ?? const <BasketOrder>[], (r[1] as List<BasketBasket>?) ?? const <BasketBasket>[]);
      },
      builder: (context, d, reload) => RefreshIndicator(
        onRefresh: reload,
        child: ListView(
          padding: const EdgeInsets.only(bottom: 32),
          children: [
            const SectionHeader('Đơn hàng'),
            if (d.$1.isEmpty) const Padding(padding: EdgeInsets.symmetric(horizontal: 20), child: Text('Chưa có đơn')),
            for (final o in d.$1)
              Padding(
                padding: const EdgeInsets.fromLTRB(16, 0, 16, 8),
                child: EntityTile(
                  title: 'Đơn ${o.number ?? o.id}',
                  subtitle: o.shippingMethod,
                  caption: formatMoney(o.totalInclTax, o.currency),
                  icon: Icons.inventory_2_rounded,
                  badges: [StatusPill(o.status)],
                  onTap: () async {
                    final full = await api.basketOrder(id: o.id!).catchError((_) => null);
                    if (!context.mounted || full == null) return;
                    await showResult(
                      context,
                      icon: Icons.inventory_2_rounded,
                      title: 'Đơn ${full.number ?? full.id}',
                      message: 'Tổng ${formatMoney(full.totalInclTax, full.currency)} (chưa thuế ${formatMoney(full.totalExclTax, full.currency)})\nPhí giao ${formatMoney(full.shippingInclTax, full.currency)} • ${full.shippingMethod ?? ''}\nTrạng thái ${full.status ?? ''}',
                    );
                  },
                ),
              ),
            const SectionHeader('Giỏ hàng'),
            for (final b in d.$2)
              Padding(
                padding: const EdgeInsets.fromLTRB(16, 0, 16, 8),
                child: EntityTile(
                  title: 'Giỏ #${b.id}',
                  subtitle: 'Tạo ${formatDate(b.dateCreated)}${b.dateSubmitted == null ? '' : ' • gửi ${formatDate(b.dateSubmitted)}'}',
                  caption: b.totalInclTax == null ? null : formatMoney(b.totalInclTax, b.currency),
                  icon: Icons.shopping_basket_rounded,
                  badges: [StatusPill(b.status)],
                  onTap: () async {
                    await openPage(context, _BasketPage(id: b.id!));
                    await reload();
                  },
                ),
              ),
          ],
        ),
      ),
    );
  }
}

class _BasketPage extends StatefulWidget {
  const _BasketPage({required this.id});

  final int id;

  @override
  State<_BasketPage> createState() => _BasketPageState();
}

class _BasketPageState extends State<_BasketPage> {
  int _v = 0;

  @override
  Widget build(BuildContext context) {
    final api = context.api.basket;
    final id = widget.id;
    void bump() => setState(() => _v++);
    return AsyncBuilder<(BasketBasket?, List<BasketLine>, List<BasketLineAttribute>, BasketUser?)>(
      key: ValueKey(_v),
      load: () async {
        final b = await api.basketBasket(id: id);
        final r = await Future.wait<Object?>([
          api.basketLines().catchError((_) => null),
          api.basketLineAttributes().catchError((_) => null),
          b?.userId == null ? Future.value(null) : api.basketUser(id: b!.userId!).catchError((_) => null),
        ]);
        final lines = ((r[0] as List<BasketLine>?) ?? const <BasketLine>[]).where((l) => l.basketId == id).toList();
        final attrs = ((r[1] as List<BasketLineAttribute>?) ?? const <BasketLineAttribute>[]).where((a) => a.basketId == id).toList();
        return (b, lines, attrs, r[2] as BasketUser?);
      },
      loading: const Scaffold(body: SkeletonList()),
      builder: (context, d, _) {
        final b = d.$1;
        return DetailPage(
          title: 'Giỏ #$id',
          actions: [
            PopupMenuButton<String>(
              onSelected: (a) async {
                if (a == 'status') {
                  if (await showFormSheet(
                    context,
                    title: 'Cập nhật giỏ',
                    fields: [
                      KField('status', 'Trạng thái', kind: FieldKind.choice, initial: b?.status ?? 'Open', options: const [('Open', 'Đang mở'), ('Saved', 'Lưu để mua sau'), ('Frozen', 'Tạm khoá')]),
                      KField('ship', 'Mã địa chỉ giao hàng', kind: FieldKind.number, initial: b?.shippingAddressId?.toString()),
                      KField('bill', 'Mã địa chỉ thanh toán', kind: FieldKind.number, initial: b?.billingAddressId?.toString()),
                    ],
                    success: 'Đã cập nhật',
                    onSubmit: (v) => api.basketUpdateBasket(id: id, input: BasketBasketInput(status: v.str('status'), shippingAddressId: v.integer('ship'), billingAddressId: v.integer('bill'))),
                  )) {
                    bump();
                  }
                } else if (await runAction(context, () => api.basketDeleteBasket(id: id), confirm: 'Xoá giỏ hàng này?', destructive: true, success: 'Đã xoá giỏ') && context.mounted) {
                  Navigator.pop(context);
                }
              },
              itemBuilder: (_) => const [PopupMenuItem(value: 'status', child: Text('Cập nhật')), PopupMenuItem(value: 'delete', child: Text('Xoá giỏ'))],
            ),
          ],
          children: [
            InfoSection(rows: [
              ('Trạng thái', b?.status),
              ('Chủ giỏ', d.$4 == null ? null : '${d.$4!.name_ ?? d.$4!.username} (${d.$4!.email ?? ''})'),
              ('Tạo lúc', formatDate(b?.dateCreated)),
              ('Thanh toán', b?.paymentReference),
              ('Mã giao vận', b?.fulfillmentId),
            ]),
            const SectionHeader('Dòng sản phẩm'),
            for (final l in d.$2)
              Padding(
                padding: const EdgeInsets.fromLTRB(16, 0, 16, 8),
                child: EntityTile(
                  title: 'Sản phẩm #${l.productId} × ${l.quantity ?? 1}',
                  subtitle: 'Kho #${l.stockrecordId ?? '-'} • ${l.lineReference ?? ''}',
                  caption: formatMoney(l.priceInclTax, l.priceCurrency),
                  icon: Icons.shopping_bag_rounded,
                  onTap: () async {
                    final line = await api.basketLine(id: l.id!).catchError((_) => null);
                    if (!context.mounted) return;
                    if (await showFormSheet(
                      context,
                      title: 'Tuỳ chọn cho dòng #${line?.id ?? l.id}',
                      description: 'Gắn giá trị tuỳ chọn (màu, size, lời nhắn…) theo mã tuỳ chọn của danh mục.',
                      fields: const [KField('option', 'Mã tuỳ chọn', kind: FieldKind.number, required: true), KField('value', 'Giá trị', required: true)],
                      success: 'Đã thêm tuỳ chọn',
                      onSubmit: (v) => api.basketCreateLineAttribute(input: BasketLineAttributeInput(basketId: id, lineId: l.id, optionId: v.integer('option'), value: v.str('value'))),
                    )) {
                      bump();
                    }
                  },
                ),
              ),
            if (d.$3.isNotEmpty) ...[
              const SectionHeader('Tuỳ chọn đã gắn'),
              for (final a in d.$3)
                ListTile(
                  contentPadding: const EdgeInsets.symmetric(horizontal: 20),
                  title: Text('Dòng #${a.lineId} • tuỳ chọn #${a.optionId}'),
                  subtitle: Text(a.value ?? ''),
                  onTap: () async {
                    final cur = await api.basketLineAttribute(id: a.id!).catchError((_) => null);
                    if (!context.mounted) return;
                    if (await showFormSheet(
                      context,
                      title: 'Sửa tuỳ chọn',
                      fields: [KField('value', 'Giá trị', required: true, initial: cur?.value ?? a.value)],
                      success: 'Đã lưu',
                      onSubmit: (v) => api.basketUpdateLineAttribute(id: a.id!, input: BasketLineAttributeInput(basketId: id, lineId: a.lineId, optionId: a.optionId, value: v.str('value'))),
                    )) {
                      bump();
                    }
                  },
                  trailing: IconButton(
                    icon: const Icon(Icons.delete_outline),
                    onPressed: () async {
                      if (await runAction(context, () => api.basketDeleteLineAttribute(id: a.id!), success: 'Đã xoá')) bump();
                    },
                  ),
                ),
            ],
          ],
        );
      },
    );
  }
}

class _Coupons extends StatelessWidget {
  const _Coupons();

  @override
  Widget build(BuildContext context) {
    final api = context.api.coupon;
    CouponValidateInput input(FormData v) => CouponValidateInput(code: v.str('code')?.toUpperCase(), purchaseAmount: v.number('amount'), productIds: v.ints('products'));
    return AsyncList<CouponUsage>(
      load: () => api.myCouponUsages(),
      header: ActionButtons([
        FilledButton.icon(
          icon: const Icon(Icons.fact_check_outlined),
          label: const Text('Kiểm tra mã'),
          onPressed: () async {
            CouponValidation? r;
            await showFormSheet(
              context,
              title: 'Kiểm tra mã giảm giá',
              fields: const [
                KField('code', 'Mã', required: true),
                KField('amount', 'Giá trị đơn (₫)', kind: FieldKind.number, required: true),
                KField('products', 'Mã sản phẩm (tuỳ chọn, phân cách dấu phẩy)'),
              ],
              submitLabel: 'Kiểm tra',
              onSubmit: (v) async => r = await api.validateCoupon(input: input(v)),
            );
            if (r == null || !context.mounted) return;
            await showResult(context, title: 'Mã ${r!.coupon?.code ?? ''} hợp lệ', message: 'Giảm ${vnd(r!.discount)}\n${r!.coupon?.description ?? ''}');
          },
        ),
        OutlinedButton.icon(
          icon: const Icon(Icons.check_circle_outline),
          label: const Text('Áp mã cho đơn'),
          onPressed: () => showFormSheet(
            context,
            title: 'Áp mã giảm giá',
            fields: const [
              KField('code', 'Mã', required: true),
              KField('amount', 'Giá trị đơn (₫)', kind: FieldKind.number, required: true),
              KField('order', 'Mã đơn hàng', kind: FieldKind.number, required: true),
            ],
            success: 'Đã áp mã, kéo xuống để làm mới',
            onSubmit: (v) => api.applyCoupon(input: CouponApplyInput(code: v.str('code')?.toUpperCase(), purchaseAmount: v.number('amount'), orderId: v.integer('order'))),
          ),
        ),
      ]),
      emptyMessage: 'Bạn chưa dùng mã nào',
      emptyIcon: Icons.local_offer_outlined,
      itemBuilder: (context, u, _) => EntityTile(
        title: 'Mã #${u.couponId} • đơn #${u.orderId ?? '-'}',
        subtitle: formatDate(u.usedAt),
        caption: '-${vnd(u.discountAmount)}',
        icon: Icons.local_offer_rounded,
      ),
    );
  }
}

class _Catalogue extends StatelessWidget {
  const _Catalogue();

  @override
  Widget build(BuildContext context) {
    final cat = context.api.catalogues;
    final wl = context.api.wishlists;
    return AsyncBuilder<(List<CatalogueCategory>, List<CatalogueProductClass>, List<CatalogueOption>, List<WishlistProductStat>)>(
      load: () async {
        final r = await Future.wait<Object?>([
          cat.catalogueCategories().catchError((_) => null),
          cat.catalogueProductClasses().catchError((_) => null),
          cat.catalogueOptions().catchError((_) => null),
          wl.wishlistPopularProducts(limit: 10).catchError((_) => null),
        ]);
        return (
          (r[0] as List<CatalogueCategory>?) ?? const <CatalogueCategory>[],
          (r[1] as List<CatalogueProductClass>?) ?? const <CatalogueProductClass>[],
          (r[2] as List<CatalogueOption>?) ?? const <CatalogueOption>[],
          (r[3] as List<WishlistProductStat>?) ?? const <WishlistProductStat>[],
        );
      },
      builder: (context, d, reload) => RefreshIndicator(
        onRefresh: reload,
        child: ListView(
          padding: const EdgeInsets.only(bottom: 32),
          children: [
            ActionButtons([
              FilledButton.icon(
                icon: const Icon(Icons.inventory_outlined),
                label: const Text('Tồn kho sản phẩm'),
                onPressed: () async {
                  List<CatalogueStockRecord>? stock;
                  WishlistProductLookup? look;
                  await showFormSheet(
                    context,
                    title: 'Tra cứu sản phẩm',
                    fields: const [KField('id', 'Mã sản phẩm', kind: FieldKind.number, required: true)],
                    submitLabel: 'Tra cứu',
                    onSubmit: (v) async {
                      stock = await cat.catalogueStockRecords(productId: v.integer('id')!);
                      look = await wl.wishlistProductLookup(productId: v.integer('id')!).catchError((_) => null);
                    },
                  );
                  if (stock == null || !context.mounted) return;
                  final first = stock!.isEmpty ? null : await cat.catalogueStockRecord(id: stock!.first.id!).catchError((_) => null);
                  if (!context.mounted) return;
                  await showResult(
                    context,
                    icon: Icons.inventory_outlined,
                    title: '${stock!.length} kho hàng',
                    message: [
                      for (final s in stock!) '• ${s.partnerSku ?? 'SKU'}: ${formatMoney(s.price, s.priceCurrency)} • còn ${s.numAvailable ?? s.numInStock ?? 0}',
                      if (first != null) 'Cập nhật ${formatDate(first.dateUpdated)}',
                      if (look?.inWishlist == true) 'Đã có trong ${look!.wishlists?.map((w) => w.wishlistName).join(', ')}',
                    ].join('\n'),
                  );
                },
              ),
            ]),
            if (d.$4.isNotEmpty) ...[
              const SectionHeader('Được yêu thích nhất'),
              for (final p in d.$4)
                ListTile(
                  contentPadding: const EdgeInsets.symmetric(horizontal: 20),
                  leading: const Icon(Icons.favorite, color: Color(0xFFDB2777)),
                  title: Text(p.title ?? 'Sản phẩm #${p.productId}'),
                  trailing: Text('${p.wishlistCount ?? 0} danh sách'),
                ),
            ],
            const SectionHeader('Danh mục'),
            for (final c in d.$1)
              ListTile(
                contentPadding: const EdgeInsets.symmetric(horizontal: 20),
                leading: const Icon(Icons.category_outlined),
                title: Text(c.name_ ?? ''),
                trailing: const Icon(Icons.chevron_right),
                onTap: () async {
                  final full = await cat.catalogueCategory(id: c.id!).catchError((_) => null);
                  if (!context.mounted || full == null) return;
                  await showResult(context, icon: Icons.category, title: full.name_ ?? '', message: [full.description, full.metaDescription, 'Đường dẫn: ${full.slug ?? ''}'].whereType<String>().where((s) => s.isNotEmpty).join('\n'));
                },
              ),
            const SectionHeader('Loại sản phẩm'),
            for (final p in d.$2)
              ListTile(
                contentPadding: const EdgeInsets.symmetric(horizontal: 20),
                leading: const Icon(Icons.class_outlined),
                title: Text(p.name_ ?? ''),
                subtitle: Text('${p.requiresShipping == true ? 'Cần giao hàng' : 'Không cần giao'} • ${p.trackStock == true ? 'Theo dõi tồn kho' : 'Không theo dõi kho'}'),
                onTap: () async {
                  final full = await cat.catalogueProductClass(id: p.id!).catchError((_) => null);
                  final opts = d.$3.where((o) => (full?.optionIds ?? const []).contains(o.id)).map((o) => o.name_).join(', ');
                  if (context.mounted) showSnack(context, opts.isEmpty ? 'Không có tuỳ chọn' : 'Tuỳ chọn: $opts');
                },
              ),
            const SectionHeader('Tuỳ chọn sản phẩm'),
            for (final o in d.$3)
              ListTile(
                contentPadding: const EdgeInsets.symmetric(horizontal: 20),
                leading: const Icon(Icons.tune),
                title: Text('${o.name_} (#${o.id})'),
                subtitle: Text('${o.type_ ?? ''}${o.required_ == true ? ' • bắt buộc' : ''}'),
                onTap: () async {
                  final full = await cat.catalogueOption(id: o.id!).catchError((_) => null);
                  if (context.mounted) showSnack(context, 'Mã ${full?.code ?? o.code}');
                },
              ),
          ],
        ),
      ),
    );
  }
}

class FoodSearchPage extends StatelessWidget {
  const FoodSearchPage({super.key});

  @override
  Widget build(BuildContext context) {
    final api = context.api.food;
    return AppColor(
      color: const Color(0xFFEA580C),
      child: Scaffold(
        appBar: AppBar(title: const Text('Tìm món')),
        body: SearchableList<FoodFood>(
          hint: 'Phở, cơm tấm, trà sữa…',
          load: (q, _) async => q.isEmpty ? const <FoodFood>[] : await api.foodSearchFoods(name_: q),
          emptyMessage: 'Nhập tên món để tìm',
          emptyIcon: Icons.ramen_dining_outlined,
          itemBuilder: (context, f, _) => EntityTile(
            title: f.name_ ?? '',
            subtitle: f.description,
            caption: vnd(f.price),
            icon: Icons.ramen_dining_rounded,
            badges: [if (f.category != null) Pill(f.category!), if (f.isActive == false) const StatusPill('out_of_stock')],
            onTap: f.restaurantId == null
                ? null
                : () async {
                    final r = await api.foodRestaurant(id: f.restaurantId!).catchError((_) => null);
                    if (!context.mounted || r == null) return;
                    await showResult(context, icon: Icons.storefront, title: r.name_ ?? 'Nhà hàng', message: [r.address, r.description].whereType<String>().join('\n'));
                  },
          ),
        ),
      ),
    );
  }
}

class DriverJobsPage extends StatelessWidget {
  const DriverJobsPage({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Chuyến giao của tôi')),
      body: AsyncList<DoordashAssignment>(
        load: () => context.api.doordash.myDriverAssignments(),
        emptyMessage: 'Bạn chưa được giao chuyến nào',
        emptyIcon: Icons.delivery_dining_outlined,
        itemBuilder: (context, a, _) => EntityTile(
          title: 'Đơn #${a.orderId}',
          subtitle: 'Giao lúc ${formatDate(a.assignedAt)}${a.completedAt == null ? '' : ' • xong ${formatDate(a.completedAt)}'}',
          caption: a.rejectionReason,
          icon: Icons.delivery_dining_rounded,
          badges: [StatusPill(a.status)],
        ),
      ),
    );
  }
}

class QrCodesPage extends StatelessWidget {
  const QrCodesPage({super.key});

  @override
  Widget build(BuildContext context) {
    final api = context.api.qr;
    return AppColor(
      color: const Color(0xFF0F766E),
      child: Scaffold(
        appBar: AppBar(title: const Text('Mã QR của tôi')),
        body: SearchableList<QrQRCode>(
          hint: 'Tìm mã QR',
          filters: () async => const ['active', 'inactive', 'expired'],
          load: (q, status) async => (await api.myQrCodes(search: q.isEmpty ? null : q, status: status, limit: 50))?.data,
          emptyMessage: 'Chưa có mã QR',
          emptyIcon: Icons.qr_code_2,
          itemBuilder: (context, c, reload) => EntityTile(
            title: c.title ?? c.shortCode ?? '',
            subtitle: c.redirectUrl ?? c.content,
            caption: '${c.scanCount ?? 0} lượt quét',
            leading: c.imageBase64 == null
                ? null
                : ClipRRect(borderRadius: BorderRadius.circular(8), child: Image.memory(base64Decode(c.imageBase64!.split(',').last), width: 56, height: 56, errorBuilder: (_, _, _) => const Icon(Icons.qr_code_2))),
            icon: Icons.qr_code_2,
            badges: [StatusPill(c.status)],
            onTap: () async {
              await openPage(context, _QrPage(id: c.id!));
              await reload();
            },
          ),
        ),
      ),
    );
  }
}

class _QrPage extends StatefulWidget {
  const _QrPage({required this.id});

  final String id;

  @override
  State<_QrPage> createState() => _QrPageState();
}

class _QrPageState extends State<_QrPage> {
  int _v = 0;

  @override
  Widget build(BuildContext context) {
    final api = context.api.qr;
    final id = widget.id;
    void bump() => setState(() => _v++);
    return AsyncBuilder<(QrQRCode?, QrScanStats?)>(
      key: ValueKey(_v),
      load: () async {
        final r = await Future.wait<Object?>([api.myQrCode(id: id), api.myQrScanStats(id: id).catchError((_) => null)]);
        return (r[0] as QrQRCode?, r[1] as QrScanStats?);
      },
      loading: const Scaffold(body: SkeletonList()),
      builder: (context, d, _) {
        final c = d.$1;
        final s = d.$2;
        return DetailPage(
          title: c?.title ?? 'Mã QR',
          actions: [
            PopupMenuButton<String>(
              onSelected: (a) async {
                var ok = false;
                switch (a) {
                  case 'edit':
                    ok = await showFormSheet(
                      context,
                      title: 'Sửa mã QR',
                      fields: [
                        KField('title', 'Tên', initial: c?.title),
                        KField('status', 'Trạng thái', kind: FieldKind.choice, initial: c?.status ?? 'active', options: const [('active', 'Đang dùng'), ('inactive', 'Tạm tắt')]),
                        KField('fore', 'Màu mã', initial: c?.foreColor ?? '#000000'),
                        KField('back', 'Màu nền', initial: c?.backColor ?? '#FFFFFF'),
                        KField('expires', 'Hết hạn', kind: FieldKind.datetime, initial: c?.expiresAt),
                      ],
                      success: 'Đã lưu',
                      onSubmit: (v) => api.updateMyQrCode(id: id, input: QrUpdateInput(title: v.str('title'), status: v.str('status'), foreColor: v.str('fore'), backColor: v.str('back'), expiresAt: v.str('expires'))),
                    );
                  case 'content':
                    ok = await showFormSheet(
                      context,
                      title: 'Đổi nội dung',
                      description: 'Mã in sẵn vẫn dùng được, chỉ đích đến thay đổi.',
                      fields: [KField('content', 'Nội dung', initial: c?.content), KField('url', 'Link chuyển hướng', initial: c?.redirectUrl)],
                      success: 'Đã đổi nội dung',
                      onSubmit: (v) => api.updateMyQrContent(id: id, content: v.str('content'), redirectUrl: v.str('url')),
                    );
                  case 'regen':
                    ok = await runAction(context, () => api.regenerateMyQrCode(id: id), success: 'Đã tạo lại ảnh mã');
                  case 'delete':
                    if (await runAction(context, () => api.deleteMyQrCode(id: id), confirm: 'Xoá mã QR này?', destructive: true, success: 'Đã xoá') && context.mounted) {
                      Navigator.pop(context);
                    }
                }
                if (ok) bump();
              },
              itemBuilder: (_) => const [
                PopupMenuItem(value: 'edit', child: Text('Sửa')),
                PopupMenuItem(value: 'content', child: Text('Đổi nội dung')),
                PopupMenuItem(value: 'regen', child: Text('Tạo lại ảnh')),
                PopupMenuItem(value: 'delete', child: Text('Xoá')),
              ],
            ),
          ],
          header: c?.imageBase64 == null
              ? null
              : Padding(
                  padding: const EdgeInsets.all(24),
                  child: Center(child: Image.memory(base64Decode(c!.imageBase64!.split(',').last), width: 200, height: 200, errorBuilder: (_, _, _) => const Icon(Icons.qr_code_2, size: 160))),
                ),
          children: [
            InfoSection(rows: [
              ('Mã ngắn', c?.shortCode),
              ('Loại', c?.type_),
              ('Đích đến', c?.redirectUrl ?? c?.content),
              ('Trạng thái', c?.status),
              ('Hết hạn', c?.expiresAt == null ? null : formatDate(c!.expiresAt)),
            ]),
            InfoSection(title: 'Thống kê quét', rows: [
              ('Tổng lượt', '${s?.totalScans ?? c?.scanCount ?? 0}'),
              ('Người quét khác nhau', '${s?.uniqueIps ?? 0}'),
              for (final e in (s?.deviceBreakdown ?? const <String, dynamic>{}).entries) ('Thiết bị ${e.key}', '${e.value}'),
              for (final e in (s?.osBreakdown ?? const <String, dynamic>{}).entries) ('Hệ điều hành ${e.key}', '${e.value}'),
            ]),
            if ((s?.dailyScans ?? const []).isNotEmpty)
              InfoSection(title: 'Theo ngày', rows: [for (final x in s!.dailyScans!.take(14)) (shortDate(x.date), '${x.count ?? 0}')]),
          ],
        );
      },
    );
  }
}

class SecurityPage extends StatelessWidget {
  const SecurityPage({super.key});

  @override
  Widget build(BuildContext context) {
    final auth = context.api.authorize;
    final addr = context.api.address;
    return Scaffold(
      appBar: AppBar(title: const Text('Bảo mật & địa chỉ')),
      body: ListView(
        padding: const EdgeInsets.only(bottom: 32),
        children: [
          const SectionHeader('Bảo mật'),
          ListTile(
            leading: const Icon(Icons.password),
            title: const Text('Xác minh mã OTP'),
            subtitle: const Text('Nhập mã được gửi qua SMS hoặc email'),
            onTap: () async {
              OtpVerification? r;
              await showFormSheet(
                context,
                title: 'Mã OTP',
                fields: const [KField('otp', 'Mã', kind: FieldKind.number, required: true)],
                submitLabel: 'Xác minh',
                onSubmit: (v) async => r = await auth.verifyOtp(otp: v.str('otp')!),
              );
              if (r == null || !context.mounted) return;
              await showResult(context, icon: r!.valid == true ? Icons.verified_user : Icons.gpp_bad, title: r!.valid == true ? 'Mã hợp lệ' : 'Mã không hợp lệ', message: r!.message ?? '');
            },
          ),
          ListTile(
            leading: const Icon(Icons.policy_outlined),
            title: const Text('Kiểm tra quyền truy cập'),
            subtitle: const Text('Xem tài khoản có được dùng một chức năng không'),
            onTap: () async {
              bool? allowed;
              await showFormSheet(
                context,
                title: 'Kiểm tra quyền',
                fields: const [KField('object', 'Đối tượng', required: true, initial: '/graphql/ticket-service'), KField('action', 'Hành động', required: true, initial: 'QUERY')],
                submitLabel: 'Kiểm tra',
                onSubmit: (v) async => allowed = await auth.checkMyPermission(object: v.str('object')!, action: v.str('action')!),
              );
              if (allowed == null || !context.mounted) return;
              showSnack(context, allowed! ? 'Bạn có quyền này' : 'Bạn chưa được cấp quyền này');
            },
          ),
          const SectionHeader('Sổ địa chỉ'),
          AsyncBuilder<(List<AddressUserAddress>, List<AddressCountry>)>(
            load: () async {
              final r = await Future.wait<Object?>([addr.myUserAddresses().catchError((_) => null), addr.addressCountries().catchError((_) => null)]);
              return ((r[0] as List<AddressUserAddress>?) ?? const <AddressUserAddress>[], (r[1] as List<AddressCountry>?) ?? const <AddressCountry>[]);
            },
            loading: const SizedBox(height: 80),
            builder: (context, d, reload) => Column(
              children: [
                if (d.$1.isEmpty) const Padding(padding: EdgeInsets.all(16), child: Text('Chưa có địa chỉ')),
                for (final a in d.$1)
                  ListTile(
                    leading: const Icon(Icons.place_outlined),
                    title: Text(a.title ?? '${a.firstName ?? ''} ${a.lastName ?? ''}'),
                    subtitle: Text([a.line1, a.line4, a.state, a.countryId].whereType<String>().where((s) => s.isNotEmpty).join(', ')),
                    trailing: const Icon(Icons.edit_outlined),
                    onTap: () async {
                      if (await showFormSheet(
                        context,
                        title: 'Sửa địa chỉ',
                        fields: [
                          KField('title', 'Tên gợi nhớ', initial: a.title),
                          KField('first', 'Tên', initial: a.firstName),
                          KField('last', 'Họ', initial: a.lastName),
                          KField('phone', 'Điện thoại', kind: FieldKind.phone, initial: a.phoneNumber),
                          KField('line1', 'Số nhà, đường', required: true, initial: a.line1),
                          KField('line4', 'Quận/huyện, thành phố', initial: a.line4),
                          KField('state', 'Tỉnh', initial: a.state),
                          KField('country', 'Quốc gia', kind: FieldKind.choice, initial: a.countryId ?? 'VN', options: [
                            for (final c in d.$2.isEmpty ? const [AddressCountry(iso31661a2: 'VN', name_: 'Việt Nam')] : d.$2) (c.iso31661a2 ?? '', c.prinatableName ?? c.name_ ?? ''),
                          ]),
                          KField('ship', 'Mặc định khi giao hàng', kind: FieldKind.toggle, initial: '${a.isDefaultForShipping ?? false}'),
                          KField('notes', 'Ghi chú cho người giao', kind: FieldKind.multiline, initial: a.notes),
                        ],
                        success: 'Đã lưu địa chỉ',
                        onSubmit: (v) => addr.updateMyUserAddress(
                          id: a.id!,
                          input: AddressUserAddressInput(
                            title: v.str('title'),
                            firstName: v.str('first'),
                            lastName: v.str('last'),
                            phoneNumber: v.str('phone'),
                            line1: v.str('line1'),
                            line4: v.str('line4'),
                            state: v.str('state'),
                            countryId: v.str('country'),
                            isDefaultForShipping: v.flag('ship'),
                            notes: v.str('notes'),
                          ),
                        ),
                      )) {
                        await reload();
                      }
                    },
                  ),
                if (d.$2.isNotEmpty)
                  ListTile(
                    leading: const Icon(Icons.public),
                    title: Text('${d.$2.where((c) => c.isShippingCountry == true).length} quốc gia có giao hàng'),
                    subtitle: const Text('Chạm để xem chi tiết Việt Nam'),
                    onTap: () async {
                      final vn = d.$2.indexWhere((c) => c.iso31661a2 == 'VN');
                      final c = await addr.addressCountry(id: vn < 0 ? 1 : vn + 1).catchError((_) => null);
                      if (context.mounted && c != null) showSnack(context, '${c.prinatableName ?? c.name_} • ${c.isShippingCountry == true ? 'có giao hàng' : 'chưa giao hàng'}');
                    },
                  ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}

class VoucherBusinessPage extends StatelessWidget {
  const VoucherBusinessPage({super.key});

  @override
  Widget build(BuildContext context) {
    final api = context.api.voucher;
    final me = '${context.userId}';
    Future<void> show(Object? r, String title) async {
      if (r == null || !context.mounted) return;
      await showResult(context, title: title, message: '${pick(r, 'id') ?? pick(r, 'job_id') ?? ''} ${pick(r, 'status') ?? ''}'.trim());
    }

    return AsyncBuilder<List<VoucherRedemptionReport>?>(
      load: () => api.voucherRedemptionReports().catchError((_) => null),
      loading: const Scaffold(body: SkeletonList()),
      builder: (context, reports, reload) => DetailPage(
        title: 'Voucher cho doanh nghiệp',
        header: const AppBanner(icon: Icons.business_rounded, title: 'Tặng quà cho nhân viên & khách hàng', subtitle: 'Mua voucher số lượng lớn, phân phối tự động, theo dõi tỷ lệ sử dụng'),
        children: [
          const SectionHeader('Mua voucher'),
          ActionButtons([
            FilledButton.icon(
              icon: const Icon(Icons.shopping_cart_checkout),
              label: const Text('Đặt mua'),
              onPressed: () async {
                final merchants = await api.voucherMerchants().catchError((_) => null) ?? const <VoucherMerchant>[];
                if (!context.mounted) return;
                VoucherOrder? order;
                await showFormSheet(
                  context,
                  title: 'Đơn mua voucher',
                  fields: [
                    KField('merchant', 'Đối tác', kind: FieldKind.choice, options: [for (final m in merchants) (m.id ?? '', m.name_ ?? '')]),
                    const KField('sku', 'Mã sản phẩm', required: true),
                    const KField('price', 'Mệnh giá (₫)', kind: FieldKind.number, required: true),
                    const KField('qty', 'Số lượng', kind: FieldKind.number, required: true, initial: '10'),
                    const KField('method', 'Thanh toán', kind: FieldKind.choice, options: [('wallet', 'Ví Keiki'), ('gateway', 'Cổng thanh toán')]),
                  ],
                  submitLabel: 'Tạo đơn & thanh toán',
                  onSubmit: (v) async {
                    order = await api.voucherCreateOrder(
                      buyerType: 'user',
                      buyerId: me,
                      currency: 'VND',
                      items: [VoucherOrderItemInput(merchantId: v.str('merchant'), productSku: v.str('sku'), unitPrice: v.integer('price'), quantity: v.integer('qty'))],
                      idempotencyKey: newIdempotencyKey(),
                    );
                    if (order?.id == null) return;
                    if (v.str('method') == 'gateway') {
                      await api.voucherInitPayment(orderId: order!.id!, amount: order!.totalAmount ?? 0, currency: 'VND', provider: 'vnpay', idempotencyKey: newIdempotencyKey());
                    } else {
                      await api.voucherCheckoutOrder(id: order!.id!, paymentMethod: 'wallet', idempotencyKey: newIdempotencyKey());
                    }
                  },
                );
                if (order == null || !context.mounted) return;
                final fresh = await api.voucherOrder(id: order!.id!).catchError((_) => order);
                if (!context.mounted) return;
                final cancel = await showDialog<bool>(
                  context: context,
                  builder: (context) => AlertDialog(
                    title: Text('Đơn ${fresh?.id}'),
                    content: Text('${formatMoney(fresh?.totalAmount, fresh?.currency)} • ${fresh?.status ?? ''}'),
                    actions: [
                      TextButton(onPressed: () => Navigator.pop(context, true), child: const Text('Huỷ đơn')),
                      FilledButton(onPressed: () => Navigator.pop(context, false), child: const Text('Đóng')),
                    ],
                  ),
                );
                if (cancel == true && context.mounted) await runAction(context, () => api.voucherCancelOrder(id: fresh!.id!), success: 'Đã huỷ đơn');
              },
            ),
          ]),
          const SectionHeader('Doanh nghiệp'),
          ActionButtons([
            OutlinedButton.icon(
              icon: const Icon(Icons.apartment),
              label: const Text('Đăng ký doanh nghiệp'),
              onPressed: () async {
                Object? r;
                await showFormSheet(
                  context,
                  title: 'Đăng ký doanh nghiệp',
                  fields: const [KField('name', 'Tên công ty', required: true), KField('tax', 'Mã số thuế'), KField('email', 'Email liên hệ', kind: FieldKind.email)],
                  onSubmit: (v) async => r = await api.voucherRegisterCorporate(name_: v.str('name')!, taxCode: v.str('tax'), contactEmail: v.str('email')),
                );
                await show(r, 'Đã đăng ký doanh nghiệp');
              },
            ),
            OutlinedButton.icon(
              icon: const Icon(Icons.savings_outlined),
              label: const Text('Nạp ngân sách'),
              onPressed: () => showFormSheet(
                context,
                title: 'Ngân sách quà tặng',
                fields: const [KField('corp', 'Mã doanh nghiệp', required: true), KField('amount', 'Số tiền (₫)', kind: FieldKind.number, required: true)],
                success: 'Đã cập nhật ngân sách',
                onSubmit: (v) => api.voucherSetCorporateBudget(id: v.str('corp')!, amount: v.integer('amount')!, currency: 'VND'),
              ),
            ),
            OutlinedButton.icon(
              icon: const Icon(Icons.send_outlined),
              label: const Text('Phát voucher hàng loạt'),
              onPressed: () async {
                Object? r;
                await showFormSheet(
                  context,
                  title: 'Phát voucher',
                  fields: const [
                    KField('corp', 'Mã doanh nghiệp', required: true),
                    KField('merchant', 'Mã đối tác', required: true),
                    KField('sku', 'Mã sản phẩm', required: true),
                    KField('price', 'Mệnh giá (₫)', kind: FieldKind.number, required: true),
                    KField('to', 'Người nhận (email/SĐT, mỗi dòng một người)', kind: FieldKind.multiline, required: true),
                  ],
                  onSubmit: (v) async => r = await api.voucherCreateDistributionJob(
                    corporateId: v.str('corp')!,
                    merchantId: v.str('merchant')!,
                    productSku: v.str('sku')!,
                    denominationAmount: v.integer('price')!,
                    currency: 'VND',
                    recipients: (v.str('to') ?? '').split(RegExp(r'[\n,]')).map((s) => s.trim()).where((s) => s.isNotEmpty).toList(),
                  ),
                );
                await show(r, 'Đã tạo đợt phát voucher');
              },
            ),
            OutlinedButton.icon(
              icon: const Icon(Icons.query_stats),
              label: const Text('Báo cáo chi tiêu'),
              onPressed: () async {
                VoucherCorporateSpendReport? r;
                await showFormSheet(
                  context,
                  title: 'Chi tiêu doanh nghiệp',
                  fields: const [KField('corp', 'Mã doanh nghiệp', required: true)],
                  submitLabel: 'Xem',
                  onSubmit: (v) async => r = await api.voucherCorporateSpend(id: v.str('corp')!),
                );
                if (r == null || !context.mounted) return;
                await showResult(context, title: 'Chi tiêu', message: '${formatMoney(r!.totalSpent, r!.currency)} • ${r!.vouchersSent ?? 0} voucher đã gửi');
              },
            ),
          ]),
          const SectionHeader('Đối tác phát hành'),
          ActionButtons([
            OutlinedButton.icon(
              icon: const Icon(Icons.storefront),
              label: const Text('Đăng ký đối tác'),
              onPressed: () => showFormSheet(
                context,
                title: 'Đối tác voucher',
                fields: const [KField('name', 'Tên', required: true), KField('type', 'Kiểu tích hợp', kind: FieldKind.choice, options: [('internal', 'Nội bộ'), ('api', 'Qua API'), ('stock', 'Nhập mã sẵn')])],
                success: 'Đã đăng ký đối tác',
                onSubmit: (v) => api.voucherRegisterMerchant(name_: v.str('name')!, providerType: v.str('type')!),
              ),
            ),
            OutlinedButton.icon(
              icon: const Icon(Icons.info_outline),
              label: const Text('Tra cứu đối tác'),
              onPressed: () async {
                VoucherMerchant? m;
                await showFormSheet(context, title: 'Mã đối tác', fields: const [KField('id', 'Mã', required: true)], submitLabel: 'Xem', onSubmit: (v) async => m = await api.voucherMerchant(id: v.str('id')!));
                if (m != null && context.mounted) await showResult(context, title: m!.name_ ?? '', message: '${m!.providerType ?? ''} • ${m!.status ?? ''}');
              },
            ),
            OutlinedButton.icon(
              icon: const Icon(Icons.add_card),
              label: const Text('Phát hành mã'),
              onPressed: () => showFormSheet(
                context,
                title: 'Phát hành voucher',
                fields: const [
                  KField('merchant', 'Mã đối tác', required: true),
                  KField('sku', 'Mã sản phẩm', required: true),
                  KField('price', 'Mệnh giá (₫)', kind: FieldKind.number, required: true),
                  KField('qty', 'Số lượng', kind: FieldKind.number, required: true),
                  KField('days', 'Hiệu lực (ngày)', kind: FieldKind.number, initial: '90'),
                ],
                success: 'Đã phát hành',
                onSubmit: (v) => api.voucherIssue(merchantId: v.str('merchant')!, productSku: v.str('sku')!, denominationAmount: v.integer('price')!, currency: 'VND', quantity: v.integer('qty')!, expiresInDays: v.integer('days'), idempotencyKey: newIdempotencyKey()),
              ),
            ),
            OutlinedButton.icon(
              icon: const Icon(Icons.upload_file),
              label: const Text('Nhập kho mã (CSV)'),
              onPressed: () => showFormSheet(
                context,
                title: 'Nhập kho mã',
                fields: const [KField('merchant', 'Mã đối tác', required: true), KField('file', 'Tệp CSV', kind: FieldKind.file, required: true)],
                success: 'Đã nhập kho',
                onSubmit: (v) => api.voucherImportStock(id: v.str('merchant')!, file: v.file('file')!),
              ),
            ),
            OutlinedButton.icon(
              icon: const Icon(Icons.block),
              label: const Text('Thu hồi voucher'),
              onPressed: () => showFormSheet(
                context,
                title: 'Thu hồi',
                fields: const [KField('id', 'Mã voucher', required: true)],
                success: 'Đã thu hồi',
                onSubmit: (v) => api.voucherRevoke(id: v.str('id')!),
              ),
            ),
            OutlinedButton.icon(
              icon: const Icon(Icons.account_balance_wallet_outlined),
              label: const Text('Cộng ví quà tặng'),
              onPressed: () => showFormSheet(
                context,
                title: 'Cộng tiền ví quà tặng',
                fields: const [
                  KField('type', 'Loại chủ ví', kind: FieldKind.choice, options: [('user', 'Người dùng'), ('corporate', 'Doanh nghiệp')]),
                  KField('owner', 'Mã chủ ví', required: true),
                  KField('amount', 'Số tiền (₫)', kind: FieldKind.number, required: true),
                  KField('ref', 'Mã tham chiếu'),
                ],
                success: 'Đã cộng tiền',
                onSubmit: (v) => api.voucherCreditWallet(
                  ownerType: v.str('type')!,
                  ownerId: v.str('owner')!,
                  amount: v.integer('amount')!,
                  currency: 'VND',
                  refType: 'manual',
                  refId: v.str('ref'),
                  idempotencyKey: newIdempotencyKey(),
                ),
              ),
            ),
          ]),
          InfoSection(title: 'Tỷ lệ sử dụng theo đối tác', rows: [
            for (final r in reports ?? const <VoucherRedemptionReport>[]) ('Đối tác ${r.merchantId}', '${r.totalRedeemed ?? 0}/${r.totalIssued ?? 0} • ${((r.redemptionRate ?? 0) * 100).toStringAsFixed(0)}%'),
          ]),
        ],
      ),
    );
  }
}
