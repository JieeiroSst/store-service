import 'package:equatable/equatable.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

import '../../../core/config/env.dart';
import '../../../core/utils/load_status.dart';
import '../data/wallet_models.dart';
import '../data/wallet_repository.dart';

class WalletState extends Equatable {
  const WalletState({
    this.status = LoadStatus.initial,
    this.wallet,
    this.balance,
    this.transactions = const [],
    this.bankAccounts = const [],
    this.incomingRequests = const [],
    this.points,
    this.error,
  });

  final LoadStatus status;
  final Wallet? wallet;
  final WalletBalance? balance;
  final List<WalletTransaction> transactions;
  final List<BankAccount> bankAccounts;
  final List<PaymentRequest> incomingRequests;
  final PointAccount? points;
  final String? error;

  bool get hasWallet => wallet != null;

  WalletState copyWith({
    LoadStatus? status,
    Wallet? wallet,
    WalletBalance? balance,
    List<WalletTransaction>? transactions,
    List<BankAccount>? bankAccounts,
    List<PaymentRequest>? incomingRequests,
    PointAccount? points,
    String? error,
  }) =>
      WalletState(
        status: status ?? this.status,
        wallet: wallet ?? this.wallet,
        balance: balance ?? this.balance,
        transactions: transactions ?? this.transactions,
        bankAccounts: bankAccounts ?? this.bankAccounts,
        incomingRequests: incomingRequests ?? this.incomingRequests,
        points: points ?? this.points,
        error: error,
      );

  @override
  List<Object?> get props => [status, wallet, balance, transactions, bankAccounts, incomingRequests, points, error];
}

class WalletCubit extends Cubit<WalletState> {
  WalletCubit(this._repository) : super(const WalletState());

  final WalletRepository _repository;

  WalletRepository get repository => _repository;

  void reset() => emit(const WalletState());

  Future<void> load() async {
    emit(state.copyWith(status: LoadStatus.loading));
    try {
      final wallets = await _repository.myWallets();
      final wallet = wallets.where((w) => w.currency == Env.defaultCurrency).firstOrNull ?? wallets.firstOrNull;
      if (wallet == null) {
        emit(const WalletState(status: LoadStatus.success));
        return;
      }
      final results = await Future.wait([
        _repository.balance(wallet.id).then<WalletBalance?>((b) => b).catchError((_) => null),
        _repository.transactions(wallet.id, pageSize: 10).catchError((_) => <WalletTransaction>[]),
        _repository.bankAccounts().catchError((_) => <BankAccount>[]),
        _repository.paymentRequests(role: 'incoming', status: 'PENDING').catchError((_) => <PaymentRequest>[]),
        _repository.points().catchError((_) => null),
      ]);
      emit(WalletState(
        status: LoadStatus.success,
        wallet: wallet,
        balance: results[0] as WalletBalance?,
        transactions: results[1] as List<WalletTransaction>,
        bankAccounts: results[2] as List<BankAccount>,
        incomingRequests: results[3] as List<PaymentRequest>,
        points: results[4] as PointAccount?,
      ));
    } catch (e) {
      emit(state.copyWith(status: LoadStatus.failure, error: e.toString()));
    }
  }

  Future<void> createWallet({String? pin}) async {
    emit(state.copyWith(status: LoadStatus.loading));
    try {
      final wallet = await _repository.createWallet(Env.defaultCurrency);
      if (pin != null && pin.isNotEmpty) await _repository.setPin(wallet.id, pin);
      await load();
    } catch (e) {
      emit(state.copyWith(status: LoadStatus.failure, error: e.toString()));
    }
  }

  Future<void> payRequest(PaymentRequest request, String pin) async {
    await _repository.payRequest(request.id, state.wallet!.id, pin);
    await load();
  }

  Future<void> declineRequest(PaymentRequest request) async {
    await _repository.declineRequest(request.id);
    await load();
  }
}
