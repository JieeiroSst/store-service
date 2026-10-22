import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';

import '../../app/mini_app.dart';
import 'view/wallet_forms.dart';
import 'view/wallet_page.dart';

final walletMiniApp = MiniApp(
  id: 'wallet',
  title: 'Ví Keiki',
  icon: Icons.account_balance_wallet,
  color: Colors.indigo,
  path: '/wallet',
  routes: [
    GoRoute(
      path: '/wallet',
      builder: (_, _) => const WalletPage(),
      routes: [
        GoRoute(path: 'topup', builder: (_, _) => const TopUpPage()),
        GoRoute(path: 'withdraw', builder: (_, _) => const WithdrawPage()),
        GoRoute(
          path: 'transfer',
          builder: (_, state) => TransferPage(
            initialWalletId: state.uri.queryParameters['wallet_id'],
            initialAmount: state.uri.queryParameters['amount'],
          ),
        ),
        GoRoute(path: 'receive', builder: (_, _) => const ReceivePage()),
        GoRoute(path: 'pin', builder: (_, _) => const SetPinPage()),
        GoRoute(path: 'banks', builder: (_, _) => const BankAccountsPage()),
        GoRoute(path: 'transactions', builder: (_, _) => const TransactionsPage()),
      ],
    ),
  ],
);
