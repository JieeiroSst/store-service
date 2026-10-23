import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:go_router/go_router.dart';

import '../../../app/theme.dart';
import '../../../core/scan/scan.dart';
import '../../../core/utils/formatters.dart';
import '../../../core/widgets/status_view.dart';
import '../../../core/widgets/ui.dart';
import '../bloc/wallet_cubit.dart';
import '../data/wallet_models.dart';
import '../../apps/commerce/commerce_more.dart';

class WalletPage extends StatelessWidget {
  const WalletPage({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Ví Keiki'),
        actions: [
          IconButton(tooltip: 'Quét mã thanh toán', icon: const Icon(Icons.qr_code_scanner), onPressed: () => scanAndOpen(context)),
          IconButton(
            tooltip: 'Quản lý ví',
            icon: const Icon(Icons.manage_accounts_outlined),
            onPressed: () => Navigator.of(context).push(MaterialPageRoute<void>(builder: (_) => const WalletMorePage())),
          ),
        ],
      ),
      body: BlocBuilder<WalletCubit, WalletState>(
        builder: (context, state) => StatusView(
          status: state.status,
          error: state.error,
          onRetry: context.read<WalletCubit>().load,
          builder: (context) => state.hasWallet ? _WalletBody(state: state) : const _CreateWallet(),
        ),
      ),
    );
  }
}

class _WalletBody extends StatelessWidget {
  const _WalletBody({required this.state});

  final WalletState state;

  Future<void> _open(BuildContext context, String path) async {
    final changed = await context.push<bool>(path);
    if (changed == true && context.mounted) context.read<WalletCubit>().load();
  }

  @override
  Widget build(BuildContext context) {
    final wallet = state.wallet!;
    final theme = Theme.of(context);
    final held = num.tryParse(state.balance?.held ?? '') ?? 0;
    return RefreshIndicator(
      onRefresh: context.read<WalletCubit>().load,
      child: ListView(
        padding: const EdgeInsets.fromLTRB(16, 8, 16, 32),
        children: [
          GradientCard(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  children: [
                    const Icon(Icons.account_balance_wallet, color: Colors.white),
                    const SizedBox(width: 8),
                    Text('Ví Keiki • ${wallet.tier ?? 'BASIC'}', style: const TextStyle(color: Colors.white70)),
                    const Spacer(),
                    if (state.points != null)
                      Row(children: [
                        const Icon(Icons.stars_rounded, color: Colors.amberAccent, size: 18),
                        const SizedBox(width: 4),
                        Text('${state.points!.balance} điểm', style: const TextStyle(color: Colors.white, fontWeight: FontWeight.w600)),
                      ]),
                  ],
                ),
                const SizedBox(height: 20),
                const Text('Số dư khả dụng', style: TextStyle(color: Colors.white70)),
                Text(
                  formatMoneyString(state.balance?.available ?? wallet.balance, wallet.currency),
                  style: theme.textTheme.headlineMedium?.copyWith(color: Colors.white),
                ),
                if (held > 0) Text('Đang tạm giữ ${formatMoneyString(state.balance?.held, wallet.currency)}', style: const TextStyle(color: Colors.white70)),
              ],
            ),
          ),
          const SizedBox(height: 16),
          Card(
            child: Padding(
              padding: const EdgeInsets.symmetric(vertical: 8),
              child: Row(
                children: [
                  Expanded(child: QuickAction(icon: Icons.add_rounded, label: 'Nạp tiền', onTap: () => _open(context, '/wallet/topup'))),
                  Expanded(child: QuickAction(icon: Icons.north_east_rounded, label: 'Chuyển', color: Colors.teal, onTap: () => _open(context, '/wallet/transfer'))),
                  Expanded(child: QuickAction(icon: Icons.account_balance_outlined, label: 'Rút tiền', color: Colors.orange, onTap: () => _open(context, '/wallet/withdraw'))),
                  Expanded(child: QuickAction(icon: Icons.qr_code_rounded, label: 'Nhận tiền', color: Colors.purple, onTap: () => _open(context, '/wallet/receive'))),
                ],
              ),
            ),
          ),
          if (!wallet.hasPin) ...[
            const SizedBox(height: 12),
            Card(
              color: theme.colorScheme.tertiaryContainer,
              child: ListTile(
                leading: const Icon(Icons.pin_outlined),
                title: const Text('Đặt mã PIN để bảo vệ ví'),
                subtitle: const Text('Cần PIN để chuyển, rút tiền và thanh toán'),
                trailing: const Icon(Icons.chevron_right),
                onTap: () => _open(context, '/wallet/pin'),
              ),
            ),
          ],
          if (state.incomingRequests.isNotEmpty) ...[
            const SectionHeader('Yêu cầu thanh toán', padding: EdgeInsets.fromLTRB(4, 24, 4, 8)),
            for (final r in state.incomingRequests) Padding(padding: const EdgeInsets.only(bottom: 8), child: _PaymentRequestTile(request: r, currency: wallet.currency)),
          ],
          SectionHeader('Giao dịch gần đây', action: 'Xem tất cả', onAction: () => _open(context, '/wallet/transactions'), padding: const EdgeInsets.fromLTRB(4, 24, 0, 8)),
          Card(
            child: state.transactions.isEmpty
                ? const Padding(padding: EdgeInsets.all(24), child: Center(child: Text('Chưa có giao dịch nào')))
                : Padding(
                    padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 4),
                    child: Column(children: [for (final t in state.transactions) TransactionTile(tx: t, walletId: wallet.id)]),
                  ),
          ),
          const SectionHeader('Quản lý ví', padding: EdgeInsets.fromLTRB(4, 24, 4, 8)),
          Card(
            child: Column(
              children: [
                ListTile(
                  leading: const IconBadge(icon: Icons.account_balance_outlined, color: Colors.indigo, size: 40),
                  title: const Text('Tài khoản ngân hàng'),
                  subtitle: Text('${state.bankAccounts.length} tài khoản đã liên kết'),
                  trailing: const Icon(Icons.chevron_right),
                  onTap: () => _open(context, '/wallet/banks'),
                ),
                ListTile(
                  leading: const IconBadge(icon: Icons.pin_outlined, color: Colors.teal, size: 40),
                  title: Text(wallet.hasPin ? 'Đổi mã PIN' : 'Đặt mã PIN'),
                  trailing: const Icon(Icons.chevron_right),
                  onTap: () => _open(context, '/wallet/pin'),
                ),
                ListTile(
                  leading: const IconBadge(icon: Icons.more_horiz, color: Colors.blueGrey, size: 40),
                  title: const Text('Hạn mức, sao kê, thẻ…'),
                  subtitle: const Text('Mọi chức năng của Ví Keiki'),
                  trailing: const Icon(Icons.chevron_right),
                  onTap: () => context.push('/services/wallet_service'),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}

class _PaymentRequestTile extends StatelessWidget {
  const _PaymentRequestTile({required this.request, required this.currency});

  final PaymentRequest request;
  final String currency;

  Future<void> _pay(BuildContext context) async {
    final pin = await askPin(context);
    if (pin == null || pin.isEmpty || !context.mounted) return;
    try {
      await context.read<WalletCubit>().payRequest(request, pin);
      if (context.mounted) showSnack(context, 'Đã thanh toán');
    } catch (e) {
      if (context.mounted) showSnack(context, e.toString());
    }
  }

  @override
  Widget build(BuildContext context) {
    return Card(
      child: ListTile(
        leading: const Icon(Icons.request_quote_outlined),
        title: Text(formatMoneyString(request.amount, request.currency ?? currency)),
        subtitle: Text(request.note?.isNotEmpty == true ? request.note! : 'Từ người dùng ${request.requesterUserId}'),
        trailing: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            IconButton(
              tooltip: 'Từ chối',
              icon: const Icon(Icons.close),
              onPressed: () => context.read<WalletCubit>().declineRequest(request),
            ),
            IconButton.filled(tooltip: 'Thanh toán', icon: const Icon(Icons.check), onPressed: () => _pay(context)),
          ],
        ),
      ),
    );
  }
}

class TransactionTile extends StatelessWidget {
  const TransactionTile({super.key, required this.tx, required this.walletId});

  final WalletTransaction tx;
  final String walletId;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final incoming = tx.isIncoming(walletId);
    final color = incoming ? KeikiColors.success : theme.colorScheme.onSurface;
    final status = (tx.status ?? '').toUpperCase();
    final pending = status.isNotEmpty && status != 'COMPLETED' && status != 'SETTLED' && status != 'CAPTURED';
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 10),
      child: Row(
        children: [
          IconBadge(
            icon: incoming ? Icons.south_west_rounded : Icons.north_east_rounded,
            color: incoming ? KeikiColors.success : KeikiColors.brand,
            size: 42,
          ),
          const SizedBox(width: 12),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  tx.description?.isNotEmpty == true ? tx.description! : (tx.type ?? 'Giao dịch'),
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                  style: theme.textTheme.titleSmall,
                ),
                const SizedBox(height: 2),
                Row(
                  children: [
                    Text(formatDate(tx.createdAt), style: theme.textTheme.bodySmall?.copyWith(color: theme.colorScheme.onSurfaceVariant)),
                    if (pending) ...[const SizedBox(width: 6), Pill(status == 'FAILED' || status == 'DECLINED' ? 'Thất bại' : 'Đang xử lý', color: Colors.orange)],
                  ],
                ),
              ],
            ),
          ),
          const SizedBox(width: 8),
          Text(
            '${incoming ? '+' : '−'}${formatMoneyString(tx.amount, tx.currency)}',
            style: theme.textTheme.titleSmall?.copyWith(color: color, fontWeight: FontWeight.w700),
          ),
        ],
      ),
    );
  }
}

class _CreateWallet extends StatefulWidget {
  const _CreateWallet();

  @override
  State<_CreateWallet> createState() => _CreateWalletState();
}

class _CreateWalletState extends State<_CreateWallet> {
  final _pin = TextEditingController();

  @override
  void dispose() {
    _pin.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.all(24),
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          const Icon(Icons.account_balance_wallet_outlined, size: 72),
          const SizedBox(height: 16),
          const Text('Kích hoạt ví Keiki để thanh toán, chuyển tiền và nhận tiền', textAlign: TextAlign.center),
          const SizedBox(height: 24),
          TextField(
            controller: _pin,
            obscureText: true,
            maxLength: 6,
            keyboardType: TextInputType.number,
            decoration: const InputDecoration(labelText: 'Đặt mã PIN (6 số)'),
          ),
          const SizedBox(height: 8),
          FilledButton(
            onPressed: () => context.read<WalletCubit>().createWallet(pin: _pin.text),
            child: const Text('Kích hoạt ví'),
          ),
        ],
      ),
    );
  }
}
