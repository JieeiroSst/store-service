import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:go_router/go_router.dart';

import '../../../core/utils/formatters.dart';
import '../../../core/utils/load_status.dart';
import '../../../core/widgets/status_view.dart';
import '../bloc/wallet_action_cubit.dart';
import '../bloc/wallet_cubit.dart';
import '../data/wallet_models.dart';
import 'wallet_page.dart';

class _ActionScaffold extends StatelessWidget {
  const _ActionScaffold({required this.title, required this.child});

  final String title;
  final Widget child;

  @override
  Widget build(BuildContext context) {
    return BlocProvider(
      create: (_) => WalletActionCubit(),
      child: BlocListener<WalletActionCubit, WalletActionState>(
        listener: (context, state) {
          if (state.status == ActionStatus.failure) showSnack(context, state.error ?? 'Thất bại');
          if (state.status == ActionStatus.done) {
            final tx = state.result;
            showSnack(context, tx == null ? 'Thành công' : 'Thành công • ${tx.status ?? ''}');
            context.pop(true);
          }
        },
        child: Scaffold(appBar: AppBar(title: Text(title)), body: child),
      ),
    );
  }
}

class _SubmitButton extends StatelessWidget {
  const _SubmitButton({required this.label, required this.onPressed});

  final String label;
  final VoidCallback onPressed;

  @override
  Widget build(BuildContext context) {
    return BlocBuilder<WalletActionCubit, WalletActionState>(
      builder: (context, state) => FilledButton(
        onPressed: state.status == ActionStatus.submitting ? null : onPressed,
        child: state.status == ActionStatus.submitting
            ? const SizedBox(height: 20, width: 20, child: CircularProgressIndicator(strokeWidth: 2))
            : Text(label),
      ),
    );
  }
}

class _AmountField extends StatelessWidget {
  const _AmountField({required this.controller, required this.currency});

  final TextEditingController controller;
  final String currency;

  @override
  Widget build(BuildContext context) {
    return TextFormField(
      controller: controller,
      keyboardType: const TextInputType.numberWithOptions(decimal: true),
      inputFormatters: [FilteringTextInputFormatter.allow(RegExp(r'[0-9.]'))],
      decoration: InputDecoration(labelText: 'Số tiền', suffixText: currency),
      validator: (v) => (num.tryParse(v ?? '') ?? 0) <= 0 ? 'Số tiền không hợp lệ' : null,
    );
  }
}

class _BankDropdown extends StatelessWidget {
  const _BankDropdown({required this.accounts, required this.value, required this.onChanged});

  final List<BankAccount> accounts;
  final String? value;
  final ValueChanged<String?> onChanged;

  @override
  Widget build(BuildContext context) {
    if (accounts.isEmpty) {
      return ListTile(
        contentPadding: EdgeInsets.zero,
        leading: const Icon(Icons.info_outline),
        title: const Text('Chưa liên kết tài khoản ngân hàng'),
        trailing: TextButton(
          onPressed: () => context.push('/wallet/banks'),
          child: const Text('Liên kết'),
        ),
      );
    }
    return DropdownButtonFormField<String>(
      initialValue: value,
      decoration: const InputDecoration(labelText: 'Tài khoản ngân hàng'),
      items: [for (final a in accounts) DropdownMenuItem(value: a.id, child: Text(a.label))],
      onChanged: onChanged,
      validator: (v) => v == null ? 'Chọn tài khoản' : null,
    );
  }
}

class TopUpPage extends StatefulWidget {
  const TopUpPage({super.key});

  @override
  State<TopUpPage> createState() => _TopUpPageState();
}

class _TopUpPageState extends State<TopUpPage> {
  final _form = GlobalKey<FormState>();
  final _amount = TextEditingController();
  String? _bank;

  @override
  void dispose() {
    _amount.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final wallet = context.watch<WalletCubit>().state;
    _bank ??= wallet.bankAccounts.where((a) => a.isDefault).firstOrNull?.id ?? wallet.bankAccounts.firstOrNull?.id;
    return _ActionScaffold(
      title: 'Nạp tiền',
      child: Builder(
        builder: (context) => Form(
          key: _form,
          child: ListView(
            padding: const EdgeInsets.all(16),
            children: [
              _BankDropdown(accounts: wallet.bankAccounts, value: _bank, onChanged: (v) => setState(() => _bank = v)),
              const SizedBox(height: 16),
              _AmountField(controller: _amount, currency: wallet.wallet!.currency),
              const SizedBox(height: 24),
              _SubmitButton(
                label: 'Nạp tiền',
                onPressed: () {
                  if (!_form.currentState!.validate() || _bank == null) return;
                  final repo = context.read<WalletCubit>().repository;
                  context.read<WalletActionCubit>().run(() => repo.topUp(
                        walletId: wallet.wallet!.id,
                        bankAccountId: _bank!,
                        amount: _amount.text,
                        description: 'Nạp tiền vào ví',
                      ));
                },
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class WithdrawPage extends StatefulWidget {
  const WithdrawPage({super.key});

  @override
  State<WithdrawPage> createState() => _WithdrawPageState();
}

class _WithdrawPageState extends State<WithdrawPage> {
  final _form = GlobalKey<FormState>();
  final _amount = TextEditingController();
  String? _bank;

  @override
  void dispose() {
    _amount.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final wallet = context.watch<WalletCubit>().state;
    _bank ??= wallet.bankAccounts.where((a) => a.isDefault).firstOrNull?.id ?? wallet.bankAccounts.firstOrNull?.id;
    return _ActionScaffold(
      title: 'Rút tiền',
      child: Builder(
        builder: (context) => Form(
          key: _form,
          child: ListView(
            padding: const EdgeInsets.all(16),
            children: [
              Text('Khả dụng: ${formatMoneyString(wallet.balance?.available, wallet.wallet!.currency)}'),
              const SizedBox(height: 16),
              _BankDropdown(accounts: wallet.bankAccounts, value: _bank, onChanged: (v) => setState(() => _bank = v)),
              const SizedBox(height: 16),
              _AmountField(controller: _amount, currency: wallet.wallet!.currency),
              const SizedBox(height: 24),
              _SubmitButton(
                label: 'Rút tiền',
                onPressed: () async {
                  if (!_form.currentState!.validate() || _bank == null) return;
                  final action = context.read<WalletActionCubit>();
                  final repo = context.read<WalletCubit>().repository;
                  final pin = await askPin(context);
                  if (pin == null || pin.isEmpty) return;
                  action.run(() => repo.withdraw(
                        walletId: wallet.wallet!.id,
                        bankAccountId: _bank!,
                        amount: _amount.text,
                        pin: pin,
                        description: 'Rút tiền về ngân hàng',
                      ));
                },
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class TransferPage extends StatefulWidget {
  const TransferPage({super.key, this.initialWalletId, this.initialAmount});

  final String? initialWalletId;
  final String? initialAmount;

  @override
  State<TransferPage> createState() => _TransferPageState();
}

class _TransferPageState extends State<TransferPage> {
  final _form = GlobalKey<FormState>();
  late final _recipient = TextEditingController(text: widget.initialWalletId);
  late final _amount = TextEditingController(text: widget.initialAmount);
  final _note = TextEditingController();
  bool _byWallet = true;

  @override
  void dispose() {
    _recipient.dispose();
    _amount.dispose();
    _note.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final wallet = context.watch<WalletCubit>().state;
    return _ActionScaffold(
      title: 'Chuyển tiền',
      child: Builder(
        builder: (context) => Form(
          key: _form,
          child: ListView(
            padding: const EdgeInsets.all(16),
            children: [
              SegmentedButton<bool>(
                segments: const [
                  ButtonSegment(value: true, label: Text('Mã ví'), icon: Icon(Icons.wallet)),
                  ButtonSegment(value: false, label: Text('Mã người dùng'), icon: Icon(Icons.person)),
                ],
                selected: {_byWallet},
                onSelectionChanged: (s) => setState(() => _byWallet = s.first),
              ),
              const SizedBox(height: 16),
              TextFormField(
                controller: _recipient,
                decoration: InputDecoration(labelText: _byWallet ? 'Mã ví người nhận' : 'ID người nhận'),
                validator: (v) => (v == null || v.trim().isEmpty) ? 'Bắt buộc' : null,
              ),
              const SizedBox(height: 16),
              _AmountField(controller: _amount, currency: wallet.wallet!.currency),
              const SizedBox(height: 16),
              TextFormField(controller: _note, decoration: const InputDecoration(labelText: 'Lời nhắn')),
              const SizedBox(height: 24),
              _SubmitButton(
                label: 'Chuyển tiền',
                onPressed: () async {
                  if (!_form.currentState!.validate()) return;
                  final action = context.read<WalletActionCubit>();
                  final repo = context.read<WalletCubit>().repository;
                  final pin = await askPin(context);
                  if (pin == null || pin.isEmpty) return;
                  final recipient = _recipient.text.trim();
                  action.run(() => repo.transfer(
                        fromWalletId: wallet.wallet!.id,
                        toWalletId: _byWallet ? recipient : null,
                        toUserId: _byWallet ? null : recipient,
                        amount: _amount.text,
                        pin: pin,
                        description: _note.text,
                      ));
                },
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class SetPinPage extends StatefulWidget {
  const SetPinPage({super.key});

  @override
  State<SetPinPage> createState() => _SetPinPageState();
}

class _SetPinPageState extends State<SetPinPage> {
  final _form = GlobalKey<FormState>();
  final _current = TextEditingController();
  final _pin = TextEditingController();
  final _confirm = TextEditingController();

  @override
  void dispose() {
    _current.dispose();
    _pin.dispose();
    _confirm.dispose();
    super.dispose();
  }

  InputDecoration _decoration(String label) => InputDecoration(labelText: label, counterText: '');

  @override
  Widget build(BuildContext context) {
    final wallet = context.watch<WalletCubit>().state.wallet!;
    return _ActionScaffold(
      title: wallet.hasPin ? 'Đổi mã PIN' : 'Đặt mã PIN',
      child: Builder(
        builder: (context) => Form(
          key: _form,
          child: ListView(
            padding: const EdgeInsets.all(16),
            children: [
              if (wallet.hasPin)
                TextFormField(controller: _current, obscureText: true, maxLength: 6, decoration: _decoration('PIN hiện tại')),
              TextFormField(
                controller: _pin,
                obscureText: true,
                maxLength: 6,
                keyboardType: TextInputType.number,
                decoration: _decoration('PIN mới'),
                validator: (v) => RegExp(r'^\d{4,6}$').hasMatch(v ?? '') ? null : 'PIN gồm 4–6 chữ số',
              ),
              TextFormField(
                controller: _confirm,
                obscureText: true,
                maxLength: 6,
                keyboardType: TextInputType.number,
                decoration: _decoration('Nhập lại PIN'),
                validator: (v) => v == _pin.text ? null : 'PIN không khớp',
              ),
              const SizedBox(height: 24),
              _SubmitButton(
                label: 'Lưu',
                onPressed: () {
                  if (!_form.currentState!.validate()) return;
                  final repo = context.read<WalletCubit>().repository;
                  context.read<WalletActionCubit>().run(() async {
                    await repo.setPin(wallet.id, _pin.text, currentPin: wallet.hasPin ? _current.text : null);
                    return null;
                  });
                },
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class BankAccountsPage extends StatelessWidget {
  const BankAccountsPage({super.key});

  Future<void> _link(BuildContext context) async {
    final bank = TextEditingController();
    final number = TextEditingController();
    final name = TextEditingController();
    final ok = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Liên kết ngân hàng'),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            TextField(controller: bank, decoration: const InputDecoration(labelText: 'Mã ngân hàng (VCB, TCB...)')),
            TextField(controller: number, keyboardType: TextInputType.number, decoration: const InputDecoration(labelText: 'Số tài khoản')),
            TextField(controller: name, decoration: const InputDecoration(labelText: 'Tên chủ tài khoản')),
          ],
        ),
        actions: [
          TextButton(onPressed: () => Navigator.pop(context, false), child: const Text('Huỷ')),
          FilledButton(onPressed: () => Navigator.pop(context, true), child: const Text('Liên kết')),
        ],
      ),
    );
    if (ok != true || !context.mounted) return;
    final cubit = context.read<WalletCubit>();
    try {
      await cubit.repository.linkBankAccount(
        bankId: bank.text.trim(),
        accountNumber: number.text.trim(),
        accountName: name.text.trim(),
      );
      await cubit.load();
    } catch (e) {
      if (context.mounted) showSnack(context, e.toString());
    }
  }

  @override
  Widget build(BuildContext context) {
    final accounts = context.watch<WalletCubit>().state.bankAccounts;
    return Scaffold(
      appBar: AppBar(title: const Text('Tài khoản ngân hàng')),
      floatingActionButton: FloatingActionButton.extended(
        onPressed: () => _link(context),
        icon: const Icon(Icons.add),
        label: const Text('Liên kết'),
      ),
      body: accounts.isEmpty
          ? const EmptyView(message: 'Chưa có tài khoản ngân hàng')
          : ListView(
              children: [
                for (final a in accounts)
                  ListTile(
                    leading: const Icon(Icons.account_balance),
                    title: Text(a.accountName ?? ''),
                    subtitle: Text(a.label),
                    trailing: a.isDefault ? const Chip(label: Text('Mặc định')) : null,
                  ),
              ],
            ),
    );
  }
}

class TransactionsPage extends StatefulWidget {
  const TransactionsPage({super.key});

  @override
  State<TransactionsPage> createState() => _TransactionsPageState();
}

class _TransactionsPageState extends State<TransactionsPage> {
  final _items = <WalletTransaction>[];
  var _status = LoadStatus.initial;
  var _page = 1;
  var _hasMore = true;
  String? _error;

  @override
  void initState() {
    super.initState();
    _loadMore();
  }

  Future<void> _loadMore() async {
    if (_status.isLoading || !_hasMore) return;
    setState(() => _status = LoadStatus.loading);
    final cubit = context.read<WalletCubit>();
    try {
      final page = await cubit.repository.transactions(cubit.state.wallet!.id, page: _page);
      setState(() {
        _items.addAll(page);
        _hasMore = page.length == 20;
        _page++;
        _status = LoadStatus.success;
      });
    } catch (e) {
      setState(() {
        _status = LoadStatus.failure;
        _error = e.toString();
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    final walletId = context.read<WalletCubit>().state.wallet!.id;
    return Scaffold(
      appBar: AppBar(title: const Text('Lịch sử giao dịch')),
      body: _items.isEmpty
          ? StatusView(status: _status, error: _error, onRetry: _loadMore, isEmpty: true, builder: (_) => const SizedBox())
          : NotificationListener<ScrollNotification>(
              onNotification: (n) {
                if (n.metrics.extentAfter < 200) _loadMore();
                return false;
              },
              child: ListView.builder(
                padding: const EdgeInsets.symmetric(horizontal: 16),
                itemCount: _items.length + (_hasMore ? 1 : 0),
                itemBuilder: (context, i) => i == _items.length
                    ? const Padding(padding: EdgeInsets.all(16), child: Center(child: CircularProgressIndicator()))
                    : TransactionTile(tx: _items[i], walletId: walletId),
              ),
            ),
    );
  }
}

class ReceivePage extends StatefulWidget {
  const ReceivePage({super.key});

  @override
  State<ReceivePage> createState() => _ReceivePageState();
}

class _ReceivePageState extends State<ReceivePage> {
  QrCode? _qr;
  String? _error;

  @override
  void initState() {
    super.initState();
    _generate();
  }

  Future<void> _generate() async {
    final cubit = context.read<WalletCubit>();
    try {
      final qr = await cubit.repository.createReceiveQr(walletId: cubit.state.wallet!.id);
      if (mounted) setState(() => _qr = qr);
    } catch (e) {
      if (mounted) setState(() => _error = e.toString());
    }
  }

  @override
  Widget build(BuildContext context) {
    final wallet = context.read<WalletCubit>().state.wallet!;
    Widget image;
    if (_error != null) {
      image = ErrorView(message: _error!, onRetry: _generate);
    } else if (_qr == null) {
      image = const CircularProgressIndicator();
    } else if (_qr!.imageBase64?.isNotEmpty == true) {
      final raw = _qr!.imageBase64!;
      image = Image.memory(base64Decode(raw.contains(',') ? raw.split(',').last : raw), width: 260);
    } else if (_qr!.imageUrl?.isNotEmpty == true) {
      image = Image.network(_qr!.imageUrl!, width: 260);
    } else {
      image = const Icon(Icons.qr_code_2, size: 200);
    }
    return Scaffold(
      appBar: AppBar(title: const Text('Nhận tiền')),
      body: Center(
        child: Padding(
          padding: const EdgeInsets.all(24),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              image,
              const SizedBox(height: 24),
              const Text('Mã ví của bạn'),
              SelectableText(wallet.id, style: Theme.of(context).textTheme.titleMedium),
              TextButton.icon(
                icon: const Icon(Icons.copy),
                label: const Text('Sao chép'),
                onPressed: () {
                  Clipboard.setData(ClipboardData(text: wallet.id));
                  showSnack(context, 'Đã sao chép mã ví');
                },
              ),
            ],
          ),
        ),
      ),
    );
  }
}
