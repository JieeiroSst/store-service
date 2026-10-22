import 'package:flutter/material.dart';
import 'package:flutter/services.dart';

import '../../../gen/models.g.dart';
import '../commerce/commerce_more.dart';
import '../extras/extras.dart';
import '../kit/kit.dart';

List<Map<String, dynamic>> _maps(Object? data, [String key = 'items']) {
  final raw = data is List ? data : (data is Map ? (data[key] ?? data['data'] ?? data['records']) : null);
  if (raw is! List) return const [];
  return [for (final e in raw) if (e is Map) e.cast<String, dynamic>()];
}

Future<void> _copy(BuildContext context, String text, [String label = 'Đã sao chép']) async {
  await Clipboard.setData(ClipboardData(text: text));
  if (context.mounted) showSnack(context, label);
}

class BillingApp extends StatelessWidget {
  const BillingApp({super.key});

  @override
  Widget build(BuildContext context) {
    return ServiceApp(
      serviceId: 'billing',
      tabs: [
        AppTab('Gói dịch vụ', Icons.workspace_premium_outlined, (context) => AsyncList<BillingPlan>(
              load: () => context.api.billing.billingPlans(),
              header: const AppBanner(icon: Icons.workspace_premium_rounded, title: 'Chọn gói phù hợp', subtitle: 'Thanh toán định kỳ, huỷ bất cứ lúc nào'),
              emptyMessage: 'Chưa có gói dịch vụ',
              itemBuilder: (context, p, _) => KCard(
                onTap: () async {
                  final full = await context.api.billing.billingPlan(id: p.planId!).catchError((_) => null);
                  if (context.mounted && full != null) {
                    await showResult(context, icon: Icons.workspace_premium, title: full.name_ ?? '', message: '${full.description ?? ''}\n${vnd(full.price)} / ${_cycle(full.billingCycle).toLowerCase()}');
                  }
                },
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Row(
                      children: [
                        Expanded(child: Text(p.name_ ?? 'Gói', style: const TextStyle(fontWeight: FontWeight.w800, fontSize: 17))),
                        Pill(_cycle(p.billingCycle)),
                      ],
                    ),
                    const SizedBox(height: 6),
                    if (p.description != null) Text(p.description!, style: Theme.of(context).textTheme.bodyMedium),
                    const SizedBox(height: 10),
                    Text.rich(TextSpan(children: [
                      TextSpan(text: vnd(p.price), style: TextStyle(fontSize: 22, fontWeight: FontWeight.w800, color: context.appColor)),
                      TextSpan(text: ' / ${_cycle(p.billingCycle).toLowerCase()}'),
                    ])),
                  ],
                ),
              ),
            )),
      ],
    );
  }
}

String _cycle(String? c) => switch ((c ?? '').toLowerCase()) {
      'monthly' || 'month' => 'Tháng',
      'yearly' || 'annual' || 'year' => 'Năm',
      'weekly' || 'week' => 'Tuần',
      'daily' || 'day' => 'Ngày',
      'quarterly' => 'Quý',
      _ => c ?? 'Kỳ',
    };

class BonuslinkApp extends StatelessWidget {
  const BonuslinkApp({super.key});

  @override
  Widget build(BuildContext context) {
    final api = context.api.bonuslink;
    return ServiceApp(
      serviceId: 'bonuslink',
      adminActions: [Builder(builder: (context) => IconButton(tooltip: 'Đối tác', icon: const Icon(Icons.storefront_outlined), onPressed: () => openPage(context, const BonusPartnerPage())))],
      tabs: [
        AppTab('Thưởng của tôi', Icons.redeem_outlined, (context) => AsyncList<BonusReward>(
              load: () => api.myBonusRewards(limit: 50),
              header: AsyncBuilder<List<BonusBalance>?>(
                load: () => api.myBonusBalances(),
                loading: const SizedBox(height: 120),
                builder: (context, balances, _) => AppBanner(
                  icon: Icons.redeem_rounded,
                  title: compact((balances ?? const <BonusBalance>[]).fold<double>(0, (a, b) => a + (b.totalValue ?? 0))),
                  subtitle: 'Tổng giá trị thưởng đã nhận',
                  children: [
                    Wrap(
                      spacing: 8,
                      runSpacing: 8,
                      children: [
                        for (final b in balances ?? const <BonusBalance>[])
                          Chip(
                            backgroundColor: Colors.white.withValues(alpha: 0.16),
                            side: BorderSide.none,
                            label: Text('${_rewardType(b.rewardType)}: ${compact(b.totalValue)} (${b.rewardCount ?? 0})', style: const TextStyle(color: Colors.white)),
                          ),
                      ],
                    ),
                  ],
                ),
              ),
              emptyMessage: 'Chưa có phần thưởng nào',
              emptyIcon: Icons.redeem_outlined,
              itemBuilder: (context, r, _) => EntityTile(
                title: _rewardType(r.rewardType),
                subtitle: '${r.source ?? ''} • ${formatDate(r.createdAt)}',
                icon: Icons.card_giftcard_rounded,
                trailing: Text('+${compact(r.rewardValue)}', style: const TextStyle(fontWeight: FontWeight.w800, color: Color(0xFF16A34A), fontSize: 16)),
              ),
            )),
      ],
    );
  }
}

String _rewardType(String? t) => switch ((t ?? '').toLowerCase()) {
      'points' || 'point' => 'Điểm',
      'cash' || 'cashback' => 'Hoàn tiền',
      'voucher' => 'Voucher',
      'referral' => 'Giới thiệu',
      _ => t ?? 'Thưởng',
    };

class CalculateApp extends StatelessWidget {
  const CalculateApp({super.key});

  @override
  Widget build(BuildContext context) {
    final api = context.api.calculate;
    return ServiceApp(
      serviceId: 'calculate',
      tabs: [
        AppTab('Thời tiết', Icons.wb_sunny_outlined, (context) => AsyncList<CalcTrackedLocation>(
              load: () => api.calcLocations(),
              emptyMessage: 'Chưa có địa điểm theo dõi',
              emptyIcon: Icons.location_off_outlined,
              itemBuilder: (context, l, _) => _WeatherCard(location: l),
            )),
        AppTab('Tiền số', Icons.currency_bitcoin, (context) => AsyncList<CalcCoin>(
              load: () async => (await api.calcMarketSnapshot())?.coins,
              emptyMessage: 'Chưa có dữ liệu thị trường',
              spacing: 6,
              itemBuilder: (context, c, _) {
                final up = (c.priceChangePercentage24h ?? 0) >= 0;
                return EntityTile(
                  leading: KThumb(url: c.image, icon: Icons.currency_bitcoin, size: 44, radius: 22),
                  title: '${c.name_ ?? ''} (${(c.symbol ?? '').toUpperCase()})',
                  subtitle: 'Vốn hoá ${compact(c.marketCap)} • #${c.marketCapRank ?? '-'}',
                  trailing: Column(
                    crossAxisAlignment: CrossAxisAlignment.end,
                    children: [
                      Text('\$${(c.currentPrice ?? 0).toStringAsFixed((c.currentPrice ?? 0) < 1 ? 4 : 2)}', style: const TextStyle(fontWeight: FontWeight.w800)),
                      Text(
                        '${up ? '+' : ''}${(c.priceChangePercentage24h ?? 0).toStringAsFixed(2)}%',
                        style: TextStyle(color: up ? const Color(0xFF16A34A) : const Color(0xFFDC2626), fontWeight: FontWeight.w700),
                      ),
                    ],
                  ),
                );
              },
            )),
      ],
    );
  }
}

class _WeatherCard extends StatelessWidget {
  const _WeatherCard({required this.location});

  final CalcTrackedLocation location;

  @override
  Widget build(BuildContext context) {
    final api = context.api.calculate;
    return AsyncBuilder<CalcWeatherSnapshot?>(
      load: () => api.calcCurrentWeather(location: location.name_ ?? ''),
      loading: const Skeleton(height: 120),
      builder: (context, w, _) {
        final c = w?.current;
        return KCard(
          onTap: location.coordinate == null ? null : () => openPage(context, _ForecastPage(location: location)),
          child: Row(
            children: [
              Icon((c?.precipitationMm ?? 0) > 0.2 ? Icons.grain_rounded : Icons.wb_sunny_rounded, size: 44, color: (c?.precipitationMm ?? 0) > 0.2 ? const Color(0xFF0284C7) : const Color(0xFFF59E0B)),
              const SizedBox(width: 14),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(location.name_ ?? '', style: const TextStyle(fontWeight: FontWeight.w800, fontSize: 16)),
                    Text('Gió ${c?.windSpeedKmh?.toStringAsFixed(0) ?? '-'} km/h • Mưa ${c?.precipitationMm?.toStringAsFixed(1) ?? '0'} mm', style: Theme.of(context).textTheme.bodySmall),
                    if (w?.nextTide != null)
                      Text('Thuỷ triều ${w!.nextTide!.type_ == 'high' ? 'lên' : 'xuống'} ${w.nextTide!.heightM?.toStringAsFixed(1) ?? ''} m lúc ${formatDate(w.nextTide!.time)}', style: Theme.of(context).textTheme.bodySmall),
                  ],
                ),
              ),
              Text('${c?.temperatureC?.round() ?? '--'}°', style: TextStyle(fontSize: 34, fontWeight: FontWeight.w800, color: context.appColor)),
              IconButton(tooltip: 'Thuỷ triều & radar', icon: const Icon(Icons.waves), onPressed: () => openPage(context, CalcMarinePage(location: location))),
            ],
          ),
        );
      },
    );
  }
}

class _ForecastPage extends StatelessWidget {
  const _ForecastPage({required this.location});

  final CalcTrackedLocation location;

  @override
  Widget build(BuildContext context) {
    final api = context.api.calculate;
    return AsyncBuilder<CalcForecast?>(
      load: () => api.calcForecast(lat: location.coordinate!.lat ?? 0, lon: location.coordinate!.lon ?? 0),
      loading: const Scaffold(body: SkeletonList()),
      builder: (context, f, _) {
        final hours = (f?.hourly ?? const <CalcForecastPoint>[]).take(48).toList();
        return DetailPage(
          title: 'Dự báo • ${location.name_}',
          children: [
            if (hours.isEmpty) const Padding(padding: EdgeInsets.all(24), child: Text('Chưa có dự báo')),
            for (final h in hours)
              ListTile(
                leading: Icon((h.precipitationMm ?? 0) > 0.2 ? Icons.grain_rounded : Icons.wb_sunny_outlined),
                title: Text(formatDate(h.time)),
                subtitle: Text('Gió ${h.windSpeedKmh?.toStringAsFixed(0) ?? '-'} km/h • Mưa ${h.precipitationMm?.toStringAsFixed(1) ?? '0'} mm'),
                trailing: Text('${h.temperatureC?.round() ?? '--'}°', style: const TextStyle(fontSize: 20, fontWeight: FontWeight.w800)),
              ),
          ],
        );
      },
    );
  }
}

class IntegratedPaymentApp extends StatelessWidget {
  const IntegratedPaymentApp({super.key});

  @override
  Widget build(BuildContext context) {
    final api = context.api.integratedPayment;
    return ServiceApp(
      serviceId: 'integrated_payment',
      fab: Builder(
        builder: (context) => FloatingActionButton.extended(
          backgroundColor: context.appColor,
          foregroundColor: Colors.white,
          icon: const Icon(Icons.add_card),
          label: const Text('Tạo thanh toán'),
          onPressed: () async {
            IntPayPaymentResult? result;
            await showFormSheet(
              context,
              title: 'Tạo thanh toán',
              fields: const [
                KField('amount', 'Số tiền (₫)', kind: FieldKind.number, required: true),
                KField('method', 'Phương thức', kind: FieldKind.choice, options: [('wallet', 'Ví Keiki'), ('card', 'Thẻ ngân hàng'), ('bank_transfer', 'Chuyển khoản'), ('momo', 'MoMo')]),
                KField('description', 'Nội dung'),
              ],
              submitLabel: 'Tạo',
              onSubmit: (v) async {
                result = await api.intPayCreatePayment(
                  input: IntPayCreatePaymentInput(amount: v.number('amount'), currency: 'VND', paymentMethod: v.str('method'), description: v.str('description')),
                );
              },
            );
            final r = result;
            if (r == null || !context.mounted) return;
            await showResult(context, title: 'Đã tạo thanh toán', message: 'Mã: ${r.id}\nTrạng thái: ${r.status ?? ''}${r.paymentUrl == null ? '' : '\nLiên kết: ${r.paymentUrl}'}${r.message == null ? '' : '\n${r.message}'}');
          },
        ),
      ),
      tabs: [
        AppTab('Giao dịch', Icons.receipt_long_outlined, (context) => AsyncList<IntPayPayment>(
              load: () => api.myIntPayPayments(),
              emptyMessage: 'Chưa có giao dịch nào',
              emptyIcon: Icons.receipt_long_outlined,
              itemBuilder: (context, p, reload) => EntityTile(
                onTap: () async {
                  final r = await Future.wait<Object?>([api.intPayPayment(id: p.id!), api.intPayPaymentStatus(id: p.id!).catchError((_) => null)]).catchError((_) => <Object?>[null, null]);
                  final full = r[0] as IntPayPayment?;
                  if (!context.mounted || full == null) return;
                  await showResult(
                    context,
                    icon: Icons.payments_rounded,
                    title: formatMoney(full.amount, full.currency),
                    message: 'Trạng thái: ${r[1] ?? full.status}\nMã đối tác: ${full.externalId ?? '-'}\n${full.processorResponse ?? ''}',
                  );
                  await reload();
                },
                title: p.description ?? 'Thanh toán ${p.id}',
                subtitle: '${p.paymentMethod ?? ''} • ${formatDate(p.createdAt)}',
                caption: formatMoney(p.amount, p.currency),
                icon: Icons.payments_rounded,
                badges: [StatusPill(p.status)],
                trailing: PopupMenuButton<String>(
                  onSelected: (v) async {
                    final ok = v == 'process'
                        ? await runAction(context, () => api.intPayProcessPayment(id: p.id!), success: 'Đã gửi xử lý')
                        : await showFormSheet(
                            context,
                            title: 'Hoàn tiền',
                            fields: [KField('amount', 'Số tiền hoàn (₫)', kind: FieldKind.number, required: true, initial: '${(p.amount ?? 0).round()}')],
                            success: 'Đã yêu cầu hoàn tiền',
                            onSubmit: (f) => api.intPayRefundPayment(id: p.id!, amount: f.number('amount')!),
                          );
                    if (ok) await reload();
                  },
                  itemBuilder: (_) => const [
                    PopupMenuItem(value: 'process', child: Text('Xử lý thanh toán')),
                    PopupMenuItem(value: 'refund', child: Text('Hoàn tiền')),
                  ],
                ),
              ),
            )),
      ],
    );
  }
}

class PaymentApp extends StatelessWidget {
  const PaymentApp({super.key});

  @override
  Widget build(BuildContext context) {
    final api = context.api.payment;
    return ServiceApp(
      serviceId: 'payment',
      tabs: [
        AppTab('Thanh toán', Icons.credit_card, (context) => ListView(
              padding: const EdgeInsets.only(bottom: 24),
              children: [
                const AppBanner(icon: Icons.credit_card_rounded, title: 'Cổng thanh toán', subtitle: 'Tạo yêu cầu thanh toán qua Stripe, PayPal hoặc VNPay'),
                Padding(
                  padding: const EdgeInsets.all(16),
                  child: FilledButton.icon(
                    icon: const Icon(Icons.add),
                    label: const Text('Tạo yêu cầu thanh toán'),
                    onPressed: () async {
                      PaymentPayment? created;
                      await showFormSheet(
                        context,
                        title: 'Yêu cầu thanh toán',
                        fields: const [
                          KField('provider', 'Cổng', kind: FieldKind.choice, options: [('stripe', 'Stripe'), ('paypal', 'PayPal'), ('vnpay', 'VNPay')]),
                          KField('amount', 'Số tiền (₫)', kind: FieldKind.number, required: true),
                          KField('email', 'Email người trả', kind: FieldKind.email),
                          KField('description', 'Nội dung'),
                        ],
                        submitLabel: 'Tạo',
                        onSubmit: (v) async {
                          created = await api.paymentCreatePayment(
                            input: PaymentCreateInput(
                              provider: v.str('provider'),
                              amount: v.integer('amount'),
                              currency: 'VND',
                              payerEmail: v.str('email'),
                              description: v.str('description'),
                              idempotencyKey: newIdempotencyKey(),
                            ),
                          );
                        },
                      );
                      final p = created;
                      if (p == null || !context.mounted) return;
                      await showResult(
                        context,
                        title: 'Đã tạo #${p.id}',
                        message: '${formatMoney(p.amount, p.currency)} qua ${p.provider}\nTrạng thái: ${p.status}${p.externalID == null ? '' : '\nMã đối tác: ${p.externalID}'}',
                      );
                    },
                  ),
                ),
              ],
            )),
      ],
    );
  }
}

class PaymentWalletApp extends StatelessWidget {
  const PaymentWalletApp({super.key});

  @override
  Widget build(BuildContext context) {
    final api = context.api.paymentWallet;
    return ServiceApp(
      serviceId: 'payment_wallet',
      tabs: [
        AppTab('Ví', Icons.account_balance_wallet_outlined, (context) => AsyncBuilder<WalletWallet?>(
              load: () => api.myWallet().catchError((_) => null),
              builder: (context, w, reload) {
                if (w == null) {
                  return Center(
                    child: Column(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        const SizedBox(height: 260, child: EmptyView(message: 'Bạn chưa có ví thanh toán', icon: Icons.account_balance_wallet_outlined)),
                        FilledButton.icon(
                          icon: const Icon(Icons.add),
                          label: const Text('Tạo ví ngay'),
                          onPressed: () async {
                            if (await runAction(context, () => api.createMyWallet(currency: 'VND'), success: 'Đã tạo ví')) await reload();
                          },
                        ),
                      ],
                    ),
                  );
                }
                return _WalletHome(wallet: w, reload: reload);
              },
            )),
        AppTab('Phương thức', Icons.credit_card_outlined, (context) => AsyncList<WalletPaymentMethod>(
              load: () => api.myPaymentMethods(),
              header: Padding(
                padding: const EdgeInsets.fromLTRB(16, 16, 16, 0),
                child: OutlinedButton.icon(
                  icon: const Icon(Icons.add_card),
                  label: const Text('Thêm thẻ / tài khoản'),
                  onPressed: () => showFormSheet(
                    context,
                    title: 'Thêm phương thức',
                    fields: const [
                      KField('type', 'Loại', kind: FieldKind.choice, options: [('card', 'Thẻ'), ('bank_account', 'Tài khoản ngân hàng')]),
                      KField('provider', 'Nhà cung cấp', required: true, hint: 'Visa, Vietcombank…'),
                      KField('number', 'Số thẻ / tài khoản', kind: FieldKind.number, required: true),
                    ],
                    success: 'Đã thêm phương thức thanh toán',
                    onSubmit: (v) => api.addMyPaymentMethod(type_: v.str('type')!, provider: v.str('provider')!, accountNumber: v.str('number')!),
                  ),
                ),
              ),
              emptyMessage: 'Chưa có phương thức thanh toán',
              emptyIcon: Icons.credit_card_off_outlined,
              itemBuilder: (context, m, _) => EntityTile(
                title: m.provider ?? 'Phương thức',
                subtitle: '${m.type_ == 'card' ? 'Thẻ' : 'Tài khoản'} •••• ${(m.accountNumber ?? '').length > 4 ? m.accountNumber!.substring(m.accountNumber!.length - 4) : m.accountNumber ?? ''}',
                icon: m.type_ == 'card' ? Icons.credit_card_rounded : Icons.account_balance_rounded,
                badges: [if (m.isDefault == true) const Pill('Mặc định'), if (m.isActive == false) const StatusPill('inactive')],
              ),
            )),
      ],
    );
  }
}

class _WalletHome extends StatelessWidget {
  const _WalletHome({required this.wallet, required this.reload});

  final WalletWallet wallet;
  final Future<void> Function() reload;

  @override
  Widget build(BuildContext context) {
    final api = context.api.paymentWallet;
    final id = wallet.walletId!;
    return RefreshIndicator(
      onRefresh: reload,
      child: ListView(
        padding: const EdgeInsets.only(bottom: 24),
        children: [
          AppBanner(
            icon: Icons.account_balance_wallet_rounded,
            title: formatMoney(wallet.balance, wallet.currency),
            subtitle: 'Mã ví: $id',
            trailing: IconButton(onPressed: () => _copy(context, id, 'Đã sao chép mã ví'), icon: const Icon(Icons.copy, color: Colors.white)),
            children: [
              StatRow([
                StatTile(label: 'Hạn mức/ngày', value: formatMoney(wallet.dailyLimit, wallet.currency), onDark: true),
                StatTile(label: 'Mỗi giao dịch', value: formatMoney(wallet.perTransactionLimit, wallet.currency), onDark: true),
              ]),
              const SizedBox(height: 12),
              Row(
                children: [
                  Expanded(
                    child: FilledButton.icon(
                      style: FilledButton.styleFrom(backgroundColor: Colors.white, foregroundColor: context.appColor),
                      icon: const Icon(Icons.send_rounded, size: 18),
                      label: const Text('Chuyển'),
                      onPressed: () async {
                        if (await showFormSheet(
                          context,
                          title: 'Chuyển tiền',
                          fields: const [
                            KField('to', 'Mã ví người nhận', required: true),
                            KField('amount', 'Số tiền (₫)', kind: FieldKind.number, required: true),
                            KField('note', 'Lời nhắn'),
                          ],
                          submitLabel: 'Chuyển',
                          success: 'Chuyển tiền thành công',
                          onSubmit: (v) => api.transferFromMyWallet(
                            senderWalletId: id,
                            receiverWalletId: v.str('to')!,
                            amount: v.integer('amount')!,
                            referenceId: newIdempotencyKey(),
                            description: v.str('note'),
                          ),
                        )) {
                          await reload();
                        }
                      },
                    ),
                  ),
                  const SizedBox(width: 8),
                  Expanded(
                    child: FilledButton.tonalIcon(
                      icon: const Icon(Icons.request_quote_outlined, size: 18),
                      label: const Text('Yêu cầu'),
                      onPressed: () async {
                        WalletPaymentRequest? req;
                        await showFormSheet(
                          context,
                          title: 'Yêu cầu thanh toán',
                          fields: const [
                            KField('amount', 'Số tiền (₫)', kind: FieldKind.number, required: true),
                            KField('payer', 'Mã ví người trả (tuỳ chọn)'),
                            KField('note', 'Nội dung'),
                          ],
                          submitLabel: 'Tạo yêu cầu',
                          onSubmit: (v) async {
                            req = await api.createMyPaymentRequest(requesterWalletId: id, payerWalletId: v.str('payer'), amount: v.integer('amount')!, description: v.str('note'), expiresInSeconds: 86400);
                          },
                        );
                        if (req == null || !context.mounted) return;
                        await showResult(context, title: 'Đã tạo yêu cầu', message: 'Gửi mã này cho người trả:\n${req!.paymentRequestId}');
                      },
                    ),
                  ),
                  const SizedBox(width: 8),
                  IconButton.filledTonal(
                    tooltip: 'Trả yêu cầu',
                    icon: const Icon(Icons.qr_code_scanner),
                    onPressed: () async {
                      if (await showFormSheet(
                        context,
                        title: 'Trả yêu cầu thanh toán',
                        fields: const [KField('id', 'Mã yêu cầu', required: true)],
                        submitLabel: 'Thanh toán',
                        success: 'Đã thanh toán',
                        onSubmit: (v) => api.payPaymentRequestFromMyWallet(id: v.str('id')!, payerWalletId: id),
                      )) {
                        await reload();
                      }
                    },
                  ),
                ],
              ),
            ],
          ),
          SectionHeader('Ngăn tiết kiệm', action: 'Thêm', onAction: () async {
            if (await showFormSheet(
              context,
              title: 'Ngăn mới',
              fields: const [KField('name', 'Tên ngăn', required: true, hint: 'Du lịch, Quỹ khẩn cấp…')],
              success: 'Đã tạo ngăn',
              onSubmit: (v) => api.createMyPocket(id: id, name_: v.str('name')!),
            )) {
              await reload();
            }
          }),
          AsyncBuilder<List<WalletPocket>?>(
            load: () => api.myWalletPockets(id: id),
            loading: const SizedBox(height: 90),
            builder: (context, pockets, _) => (pockets ?? const []).isEmpty
                ? const Padding(padding: EdgeInsets.symmetric(horizontal: 20), child: Text('Chưa có ngăn nào'))
                : Strip<WalletPocket>(
                    items: pockets!,
                    height: 92,
                    itemWidth: 160,
                    itemBuilder: (context, p) => KCard(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(p.name_ ?? '', maxLines: 1, overflow: TextOverflow.ellipsis, style: const TextStyle(fontWeight: FontWeight.w700)),
                          const Spacer(),
                          Text(formatMoney(p.balance, p.currency), style: TextStyle(fontWeight: FontWeight.w800, color: context.appColor)),
                        ],
                      ),
                    ),
                  ),
          ),
          const SectionHeader('Giao dịch gần đây'),
          AsyncBuilder<List<WalletTransaction>?>(
            load: () => api.myWalletTransactions(id: id, limit: 30),
            loading: const SizedBox(height: 120),
            builder: (context, txs, _) => Column(
              children: [
                if ((txs ?? const []).isEmpty) const Padding(padding: EdgeInsets.all(24), child: Text('Chưa có giao dịch')),
                for (final t in txs ?? const <WalletTransaction>[])
                  Padding(
                    padding: const EdgeInsets.fromLTRB(16, 0, 16, 8),
                    child: EntityTile(
                      title: t.description ?? t.type_ ?? 'Giao dịch',
                      subtitle: formatDate(t.createdAt),
                      icon: _incoming(t.type_) ? Icons.south_west_rounded : Icons.north_east_rounded,
                      color: _incoming(t.type_) ? const Color(0xFF16A34A) : const Color(0xFFDC2626),
                      trailing: Text(
                        '${_incoming(t.type_) ? '+' : '-'}${formatMoney(t.amount, t.currency)}',
                        style: TextStyle(fontWeight: FontWeight.w800, color: _incoming(t.type_) ? const Color(0xFF16A34A) : null),
                      ),
                    ),
                  ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}

bool _incoming(String? type) {
  final t = (type ?? '').toLowerCase();
  return t.contains('deposit') || t.contains('in') && !t.contains('withdraw') || t.contains('credit') || t.contains('refund');
}

class RecompenseApp extends StatelessWidget {
  const RecompenseApp({super.key});

  @override
  Widget build(BuildContext context) {
    final api = context.api.recompense;
    return ServiceApp(
      serviceId: 'recompense',
      tabs: [
        AppTab('Ưu đãi', Icons.loyalty_outlined, (context) => AsyncList<RecompRecompense>(
              load: () => api.recompenses(limit: 50),
              header: AsyncBuilder<RecompMemberStatus?>(
                load: () => api.myRecompensePoints().catchError((_) => null),
                loading: const SizedBox(height: 140),
                builder: (context, m, _) => AppBanner(
                  icon: Icons.military_tech_rounded,
                  title: '${m?.points ?? 0} điểm',
                  subtitle: 'Hạng ${m?.tierName ?? '-'}${(m?.pointsToNextTier ?? 0) > 0 ? ' • còn ${m!.pointsToNextTier} điểm lên hạng' : ''}',
                  children: [
                    SizedBox(
                      width: double.infinity,
                      child: FilledButton.icon(
                        style: FilledButton.styleFrom(backgroundColor: Colors.white, foregroundColor: context.appColor),
                        icon: const Icon(Icons.calculate_outlined),
                        label: const Text('Tìm ưu đãi tốt nhất'),
                        onPressed: () async {
                          RecompRecompense? best;
                          double amount = 0;
                          await showFormSheet(
                            context,
                            title: 'Giá trị đơn hàng',
                            fields: const [KField('amount', 'Số tiền (₫)', kind: FieldKind.number, required: true)],
                            submitLabel: 'Tìm',
                            onSubmit: (v) async {
                              amount = v.number('amount')!;
                              best = await api.myBestRecompense(amount: amount);
                            },
                          );
                          if (!context.mounted) return;
                          if (best == null) {
                            if (amount > 0) showSnack(context, 'Chưa có ưu đãi phù hợp');
                            return;
                          }
                          await _redeem(context, best!, amount);
                        },
                      ),
                    ),
                  ],
                ),
              ),
              emptyMessage: 'Chưa có ưu đãi',
              emptyIcon: Icons.loyalty_outlined,
              itemBuilder: (context, r, _) => EntityTile(
                title: r.reductionDescription ?? 'Mã ${r.codeDeReduction ?? ''}',
                subtitle: 'Cần ${r.minimumPoints ?? 0} điểm${r.expiryDate == null ? '' : ' • HSD ${shortDate(DateTime.fromMillisecondsSinceEpoch(r.expiryDate!.round() * (r.expiryDate! < 1e12 ? 1000 : 1)).toIso8601String())}'}',
                caption: r.reductionType == 'percentage' || r.reductionType == 'percent' ? 'Giảm ${r.reductionValue?.round()}%' : 'Giảm ${vnd(r.reductionValue)}',
                icon: Icons.local_offer_rounded,
                badges: [if (r.codeDeReduction != null) Pill(r.codeDeReduction!)],
                onTap: () => _redeem(context, r, null),
              ),
            )),
        AppTab('Của tôi', Icons.inventory_2_outlined, (context) => AsyncList<RecompRecompense>(
              load: () async {
                final raw = await api.myRecompenses(limit: 50);
                final list = raw is List ? raw : (pick(raw, 'items') ?? pick(raw, 'data'));
                return [for (final e in list is List ? list : const []) if (e is Map) RecompRecompense.fromJson(e.cast<String, dynamic>())];
              },
              emptyMessage: 'Bạn chưa có ưu đãi riêng',
              emptyIcon: Icons.inventory_2_outlined,
              itemBuilder: (context, r, _) => EntityTile(
                title: r.reductionDescription ?? 'Mã ${r.codeDeReduction ?? ''}',
                subtitle: 'Hạng ${r.tier ?? '-'} • cần ${r.minimumPoints ?? 0} điểm',
                caption: r.codeDeReduction,
                icon: Icons.card_giftcard_rounded,
                onTap: () => _redeem(context, r, null),
              ),
            )),
      ],
    );
  }

  Future<void> _redeem(BuildContext context, RecompRecompense r, double? amount) async {
    RecompRedemption? done;
    await showFormSheet(
      context,
      title: r.reductionDescription ?? 'Dùng ưu đãi',
      fields: [KField('amount', 'Giá trị đơn hàng (₫)', kind: FieldKind.number, required: true, initial: amount == null ? null : '${amount.round()}')],
      submitLabel: 'Đổi điểm',
      onSubmit: (v) async => done = await context.api.recompense.redeemMyRecompense(id: int.parse(r.id ?? '0'), amount: v.number('amount')!),
    );
    if (done == null || !context.mounted) return;
    await showResult(context, title: 'Đổi ưu đãi thành công', message: 'Giảm ${vnd(done!.discount)}\nCòn phải trả ${vnd(done!.total)}');
  }
}

class RewardApp extends StatelessWidget {
  const RewardApp({super.key});

  @override
  Widget build(BuildContext context) {
    return ServiceApp(
      serviceId: 'reward',
      tabs: [
        AppTab('Quà tặng', Icons.card_giftcard_outlined, (context) => SearchableList<RewardReward>(
              hint: 'Tìm quà tặng',
              load: (q, _) async => (await context.api.reward.rewards(name_: q.isEmpty ? null : q, pageSize: 50))?.records,
              gridExtent: 200,
              gridAspect: 0.82,
              emptyMessage: 'Chưa có quà tặng',
              emptyIcon: Icons.card_giftcard_outlined,
              itemBuilder: (context, r, _) => PosterCard(
                title: r.name_ ?? 'Quà tặng',
                subtitle: r.description,
                icon: Icons.redeem_rounded,
                badge: '${r.points ?? 0} điểm',
                onTap: () => openPage(
                  context,
                  DetailPage(
                    title: r.name_ ?? 'Quà tặng',
                    header: HeroHeader(title: r.name_ ?? '', icon: Icons.redeem_rounded, badges: [Pill('${r.points ?? 0} điểm', icon: Icons.stars_rounded)]),
                    children: [Paragraph(r.description, title: 'Mô tả')],
                  ),
                ),
              ),
            )),
      ],
    );
  }
}

class VoucherApp extends StatelessWidget {
  const VoucherApp({super.key});

  @override
  Widget build(BuildContext context) {
    final api = context.api.voucher;
    return ServiceApp(
      serviceId: 'voucher',
      adminActions: [
        Builder(builder: (context) => IconButton(tooltip: 'Doanh nghiệp', icon: const Icon(Icons.business_outlined), onPressed: () => openPage(context, const VoucherBusinessPage()))),
      ],
      tabs: [
        AppTab('Voucher của tôi', Icons.confirmation_number_outlined, (context) => AsyncList<VoucherVoucher>(
              load: () => api.vouchers(ownerId: '${context.userId}'),
              header: Column(
                children: [
                  AsyncBuilder<VoucherBalance?>(
                    load: () => api.voucherWalletBalance(ownerType: 'user', ownerId: '${context.userId}').catchError((_) => null),
                    loading: const SizedBox(height: 100),
                    builder: (context, b, _) => AppBanner(
                      icon: Icons.confirmation_number_rounded,
                      title: formatMoney(b?.amount, b?.currency),
                      subtitle: 'Số dư ví quà tặng',
                      children: [
                        Row(
                          children: [
                            Expanded(
                              child: FilledButton(
                                style: FilledButton.styleFrom(backgroundColor: Colors.white, foregroundColor: context.appColor),
                                onPressed: () => showFormSheet(
                                  context,
                                  title: 'Nhận voucher',
                                  fields: const [KField('token', 'Mã nhận quà', required: true)],
                                  success: 'Đã nhận voucher',
                                  onSubmit: (v) => api.voucherClaim(token: v.str('token')!),
                                ),
                                child: const Text('Nhận quà'),
                              ),
                            ),
                            const SizedBox(width: 8),
                            Expanded(
                              child: FilledButton.tonal(
                                onPressed: () => _check(context),
                                child: const Text('Kiểm tra mã'),
                              ),
                            ),
                          ],
                        ),
                      ],
                    ),
                  ),
                ],
              ),
              emptyMessage: 'Bạn chưa có voucher nào',
              emptyIcon: Icons.confirmation_number_outlined,
              itemBuilder: (context, v, reload) => EntityTile(
                title: v.code ?? v.id ?? 'Voucher',
                subtitle: 'Nhà cung cấp ${v.merchantId ?? '-'}',
                icon: Icons.local_activity_rounded,
                badges: [StatusPill(v.status)],
                trailing: PopupMenuButton<String>(
                  onSelected: (a) async {
                    var ok = false;
                    if (a == 'activate') {
                      ok = await runAction(context, () => api.voucherActivate(id: v.id!, ownerType: 'user', ownerId: '${context.userId}'), success: 'Đã kích hoạt');
                    } else {
                      ok = await showFormSheet(
                        context,
                        title: 'Dùng voucher',
                        fields: const [
                          KField('pin', 'Mã PIN', kind: FieldKind.password, required: true),
                          KField('amount', 'Số tiền sử dụng (₫)', kind: FieldKind.number, required: true),
                        ],
                        success: 'Đã sử dụng voucher',
                        onSubmit: (f) => api.voucherRedeem(id: v.id!, pin: f.str('pin')!, amount: f.integer('amount')!, currency: 'VND', idempotencyKey: newIdempotencyKey()),
                      );
                    }
                    if (ok) await reload();
                  },
                  itemBuilder: (_) => const [
                    PopupMenuItem(value: 'activate', child: Text('Kích hoạt')),
                    PopupMenuItem(value: 'redeem', child: Text('Sử dụng')),
                  ],
                ),
              ),
            )),
        AppTab('Đối tác', Icons.storefront_outlined, (context) => AsyncList<VoucherMerchant>(
              load: () => api.voucherMerchants(),
              emptyMessage: 'Chưa có đối tác',
              itemBuilder: (context, m, _) => EntityTile(title: m.name_ ?? '', subtitle: m.providerType, icon: Icons.storefront_rounded, badges: [StatusPill(m.status)]),
            )),
      ],
    );
  }

  Future<void> _check(BuildContext context) async {
    VoucherValidationResult? r;
    await showFormSheet(
      context,
      title: 'Kiểm tra voucher',
      fields: const [KField('id', 'Mã voucher', required: true), KField('pin', 'PIN (nếu có)', kind: FieldKind.password)],
      submitLabel: 'Kiểm tra',
      onSubmit: (v) async => r = await context.api.voucher.voucherValidate(id: v.str('id')!, pin: v.str('pin')),
    );
    if (r == null || !context.mounted) return;
    await showResult(
      context,
      icon: r!.valid == true ? Icons.verified_rounded : Icons.error_outline,
      title: r!.valid == true ? 'Voucher hợp lệ' : 'Voucher không hợp lệ',
      message: r!.valid == true ? 'Số dư ${formatMoney(r!.balance?.amount, r!.balance?.currency)}' : (r!.reason ?? ''),
    );
  }
}

class ReferralApp extends StatelessWidget {
  const ReferralApp({super.key});

  @override
  Widget build(BuildContext context) {
    final api = context.api.referral;
    return ServiceApp(
      serviceId: 'referral',
      actions: [Builder(builder: (context) => IconButton(tooltip: 'Công cụ', icon: const Icon(Icons.tune), onPressed: () => openPage(context, const ReferralToolsPage())))],
      tabs: [
        AppTab('Mời bạn', Icons.group_add_outlined, (context) => AsyncList<ReferralReferralLink>(
              load: () async => (await api.myReferralLinks(limit: 30))?.data,
              header: Column(
                children: [
                  AsyncBuilder<(ReferralUserReferralStats?, ReferralRewardProgram?)>(
                    load: () async {
                      final r = await Future.wait<Object?>([api.myReferralStats().catchError((_) => null), api.referralActiveRewardProgram().catchError((_) => null)]);
                      return (r[0] as ReferralUserReferralStats?, r[1] as ReferralRewardProgram?);
                    },
                    loading: const SizedBox(height: 160),
                    builder: (context, d, reload) => AppBanner(
                      icon: Icons.diversity_3_rounded,
                      title: 'Mời bạn, nhận thưởng',
                      subtitle: d.$2?.name_ ?? 'Mỗi người bạn cài app qua link của bạn đều được tính thưởng',
                      children: [
                        StatRow([
                          StatTile(label: 'Đã mời', value: '${d.$1?.totalInvited ?? 0}', onDark: true),
                          StatTile(label: 'Đã cài', value: '${d.$1?.totalInstalled ?? 0}', onDark: true),
                          StatTile(label: 'Thưởng', value: compact(d.$1?.totalRewardAmt), onDark: true),
                        ]),
                        const SizedBox(height: 12),
                        SizedBox(
                          width: double.infinity,
                          child: FilledButton.icon(
                            style: FilledButton.styleFrom(backgroundColor: Colors.white, foregroundColor: context.appColor),
                            icon: const Icon(Icons.link),
                            label: const Text('Tạo link mời mới'),
                            onPressed: () async {
                              try {
                                final l = await api.generateMyReferralLink(channel: 'app', platform: Theme.of(context).platform.name);
                                if (!context.mounted || l == null) return;
                                await _copy(context, l.shareUrl ?? l.deepLink ?? l.refCode ?? '', 'Đã tạo và sao chép link mời');
                              } catch (e) {
                                if (context.mounted) showSnack(context, errorMessage(e));
                              }
                            },
                          ),
                        ),
                      ],
                    ),
                  ),
                  const SectionHeader('Link của bạn'),
                ],
              ),
              emptyMessage: 'Chưa có link mời nào',
              emptyIcon: Icons.link_off,
              itemBuilder: (context, l, _) => EntityTile(
                title: l.refCode ?? '',
                subtitle: l.shareUrl ?? l.deepLink,
                icon: Icons.link_rounded,
                badges: [if (l.channel != null) Pill(l.channel!), StatusPill(l.status)],
                trailing: IconButton(icon: const Icon(Icons.copy), onPressed: () => _copy(context, l.shareUrl ?? l.deepLink ?? l.refCode ?? '')),
              ),
            )),
      ],
    );
  }
}

class SardAuthApp extends StatelessWidget {
  const SardAuthApp({super.key});

  @override
  Widget build(BuildContext context) {
    final api = context.api.sardAuth;
    return ServiceApp(
      serviceId: 'sard_auth',
      tabs: [
        AppTab('Xác thực giao dịch', Icons.verified_user_outlined, (context) => AsyncList<SardChallenge>(
              load: () async => _maps(await api.mySardPaymentAuthentications()).map(SardChallenge.fromJson).toList(),
              header: const AppBanner(icon: Icons.shield_rounded, title: 'Xác thực 3-D Secure', subtitle: 'Xác nhận giao dịch thẻ bằng mã OTP gửi về điện thoại của bạn'),
              emptyMessage: 'Không có giao dịch chờ xác thực',
              emptyIcon: Icons.verified_user_outlined,
              itemBuilder: (context, c, reload) {
                final pending = (c.status ?? '').toLowerCase().contains('pending') || (c.status ?? '').toLowerCase().contains('challenge');
                return KCard(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Row(
                        children: [
                          Expanded(child: Text(c.merchant ?? 'Giao dịch thẻ', style: const TextStyle(fontWeight: FontWeight.w800, fontSize: 16))),
                          StatusPill(c.status),
                        ],
                      ),
                      const SizedBox(height: 4),
                      Text('${formatMoney(c.amount, c.currency)} • Thẻ ${c.maskedPan ?? ''}', style: Theme.of(context).textTheme.bodyMedium),
                      Text('Hết hạn ${formatDate(c.otpExpiresAt ?? c.expiresAt)} • còn ${c.attemptsLeft ?? '-'} lần thử', style: Theme.of(context).textTheme.bodySmall),
                      if (pending) ...[
                        const SizedBox(height: 10),
                        Row(
                          children: [
                            Expanded(
                              child: FilledButton(
                                onPressed: () async {
                                  if (await showFormSheet(
                                    context,
                                    title: 'Nhập mã OTP',
                                    fields: const [KField('otp', 'Mã OTP', kind: FieldKind.number, required: true)],
                                    success: 'Đã xác thực giao dịch',
                                    onSubmit: (v) => api.sardVerifyPaymentAuthentication(id: c.id!, otp: v.str('otp')!),
                                  )) {
                                    await reload();
                                  }
                                },
                                child: const Text('Xác thực'),
                              ),
                            ),
                            const SizedBox(width: 8),
                            OutlinedButton(onPressed: () => runAction(context, () => api.sardResendPaymentAuthentication(id: c.id!), success: 'Đã gửi lại OTP'), child: const Text('Gửi lại')),
                            const SizedBox(width: 8),
                            IconButton(
                              tooltip: 'Từ chối',
                              icon: const Icon(Icons.block, color: Color(0xFFDC2626)),
                              onPressed: () async {
                                if (await runAction(context, () => api.sardDeclinePaymentAuthentication(id: c.id!), confirm: 'Từ chối giao dịch này?', destructive: true, success: 'Đã từ chối')) {
                                  await reload();
                                }
                              },
                            ),
                          ],
                        ),
                      ],
                    ],
                  ),
                );
              },
            )),
      ],
    );
  }
}
