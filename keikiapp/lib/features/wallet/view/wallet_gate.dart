import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:go_router/go_router.dart';

import '../../../core/utils/load_status.dart';
import '../../../core/widgets/status_view.dart';
import '../bloc/wallet_cubit.dart';

class WalletGate extends StatefulWidget {
  const WalletGate({super.key, required this.title, required this.child});

  final String title;
  final Widget child;

  @override
  State<WalletGate> createState() => _WalletGateState();
}

class _WalletGateState extends State<WalletGate> {
  @override
  void initState() {
    super.initState();
    final cubit = context.read<WalletCubit>();
    if (cubit.state.status == LoadStatus.initial) cubit.load();
  }

  @override
  Widget build(BuildContext context) {
    final state = context.watch<WalletCubit>().state;
    if (state.hasWallet) return widget.child;
    final loading = state.status == LoadStatus.initial || state.status == LoadStatus.loading;
    return Scaffold(
      appBar: AppBar(title: Text(widget.title)),
      body: loading
          ? const Center(child: CircularProgressIndicator())
          : state.status == LoadStatus.failure
              ? ErrorView(message: state.error ?? 'Không tải được Ví Keiki', onRetry: context.read<WalletCubit>().load)
              : Center(
                  child: Padding(
                    padding: const EdgeInsets.all(32),
                    child: Column(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        Icon(Icons.account_balance_wallet_outlined, size: 64, color: Theme.of(context).colorScheme.outline),
                        const SizedBox(height: 16),
                        const Text('Bạn chưa kích hoạt Ví Keiki', textAlign: TextAlign.center),
                        const SizedBox(height: 16),
                        FilledButton(onPressed: () => context.go('/wallet'), child: const Text('Kích hoạt ví')),
                      ],
                    ),
                  ),
                ),
    );
  }
}
