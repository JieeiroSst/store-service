import 'package:equatable/equatable.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

import '../data/wallet_models.dart';

enum ActionStatus { idle, submitting, done, failure }

class WalletActionState extends Equatable {
  const WalletActionState({this.status = ActionStatus.idle, this.result, this.error});

  final ActionStatus status;
  final WalletTransaction? result;
  final String? error;

  @override
  List<Object?> get props => [status, result, error];
}

class WalletActionCubit extends Cubit<WalletActionState> {
  WalletActionCubit() : super(const WalletActionState());

  Future<void> run(Future<WalletTransaction?> Function() action) async {
    emit(const WalletActionState(status: ActionStatus.submitting));
    try {
      emit(WalletActionState(status: ActionStatus.done, result: await action()));
    } catch (e) {
      emit(WalletActionState(status: ActionStatus.failure, error: e.toString()));
    }
  }
}
