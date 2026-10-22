import 'package:equatable/equatable.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

import '../../../core/utils/load_status.dart';
import '../../../gen/gateway_executor.dart';
import '../data/schema_catalog.dart';

class OperationState extends Equatable {
  const OperationState({this.status = LoadStatus.initial, this.result, this.error, this.elapsed, this.runs = 0});

  final LoadStatus status;
  final Object? result;
  final String? error;
  final Duration? elapsed;
  final int runs;

  @override
  List<Object?> get props => [status, result, error, elapsed, runs];
}

class OperationCubit extends Cubit<OperationState> {
  OperationCubit(this._executor, this.operation) : super(const OperationState());

  final GatewayExecutor _executor;
  final GqlOperation operation;

  Future<void> run(Map<String, dynamic> variables) async {
    emit(OperationState(status: LoadStatus.loading, result: state.result, runs: state.runs));
    final watch = Stopwatch()..start();
    final doc = operation.document();
    final vars = {
      for (final e in variables.entries)
        if (e.value != null) e.key: e.value,
    };
    try {
      final data = switch (operation.access) {
        OperationAccess.public => await _executor.public(doc, variables: vars),
        _ when operation.isMutation => await _executor.mutation(doc, variables: vars),
        _ => await _executor.query(doc, variables: vars),
      };
      emit(OperationState(
        status: LoadStatus.success,
        result: data[operation.name],
        elapsed: watch.elapsed,
        runs: state.runs + 1,
      ));
    } catch (e) {
      emit(OperationState(status: LoadStatus.failure, error: e.toString(), elapsed: watch.elapsed, runs: state.runs + 1));
    }
  }
}
