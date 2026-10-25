import 'dart:async';

import 'package:equatable/equatable.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

import '../../../core/network/graphql_exception.dart';
import '../data/auth_repository.dart';
import '../data/user.dart';

sealed class AuthEvent extends Equatable {
  const AuthEvent();

  @override
  List<Object?> get props => [];
}

final class AuthStarted extends AuthEvent {
  const AuthStarted();
}

final class AuthLoginRequested extends AuthEvent {
  const AuthLoginRequested(this.username, this.password);

  final String username;
  final String password;

  @override
  List<Object?> get props => [username];
}

final class AuthMfaSubmitted extends AuthEvent {
  const AuthMfaSubmitted(this.code);

  final String code;

  @override
  List<Object?> get props => [code];
}

final class AuthMfaCancelled extends AuthEvent {
  const AuthMfaCancelled();
}

final class AuthUserChanged extends AuthEvent {
  const AuthUserChanged(this.user);

  final User user;

  @override
  List<Object?> get props => [user];
}

final class AuthLogoutRequested extends AuthEvent {
  const AuthLogoutRequested();
}

final class _AuthSessionExpired extends AuthEvent {
  const _AuthSessionExpired();
}

enum AuthStatus { unknown, authenticated, unauthenticated }

class AuthState extends Equatable {
  const AuthState({this.status = AuthStatus.unknown, this.user, this.submitting = false, this.error, this.mfaToken});

  final AuthStatus status;
  final User? user;
  final bool submitting;
  final String? error;
  final String? mfaToken;

  bool get awaitingMfa => mfaToken != null;

  AuthState copyWith({AuthStatus? status, User? user, bool? submitting, String? error}) => AuthState(
        status: status ?? this.status,
        user: user ?? this.user,
        submitting: submitting ?? this.submitting,
        error: error,
        mfaToken: mfaToken,
      );

  @override
  List<Object?> get props => [status, user, submitting, error, mfaToken];
}

class AuthBloc extends Bloc<AuthEvent, AuthState> {
  AuthBloc(this._repository) : super(const AuthState()) {
    on<AuthStarted>(_onStarted);
    on<AuthLoginRequested>(_onLogin);
    on<AuthMfaSubmitted>(_onMfaSubmitted);
    on<AuthMfaCancelled>((_, emit) => emit(const AuthState(status: AuthStatus.unauthenticated)));
    on<AuthUserChanged>((event, emit) => emit(state.copyWith(user: event.user)));
    on<AuthLogoutRequested>(_onLogout);
    on<_AuthSessionExpired>((_, emit) => emit(const AuthState(status: AuthStatus.unauthenticated)));
    _expiredSub = _repository.sessionExpired.listen((_) => add(const _AuthSessionExpired()));
  }

  final AuthRepository _repository;
  late final StreamSubscription<void> _expiredSub;

  Future<void> _onStarted(AuthStarted event, Emitter<AuthState> emit) async {
    if (!await _repository.hasSession()) {
      emit(const AuthState(status: AuthStatus.unauthenticated));
      return;
    }
    try {
      final user = await _repository.me();
      emit(user == null
          ? const AuthState(status: AuthStatus.unauthenticated)
          : AuthState(status: AuthStatus.authenticated, user: user));
    } catch (_) {
      emit(const AuthState(status: AuthStatus.unauthenticated));
    }
  }

  Future<void> _onLogin(AuthLoginRequested event, Emitter<AuthState> emit) async {
    emit(state.copyWith(submitting: true));
    try {
      await _repository.login(event.username, event.password);
      final user = await _repository.me();
      emit(AuthState(status: AuthStatus.authenticated, user: user));
    } on MfaRequiredException catch (e) {
      emit(AuthState(status: AuthStatus.unauthenticated, mfaToken: e.mfaToken));
    } catch (e) {
      emit(AuthState(status: AuthStatus.unauthenticated, error: e.toString()));
    }
  }

  Future<void> _onMfaSubmitted(AuthMfaSubmitted event, Emitter<AuthState> emit) async {
    final token = state.mfaToken;
    if (token == null) return;
    emit(state.copyWith(submitting: true));
    try {
      await _repository.verifyTotpLogin(token, event.code);
      final user = await _repository.me();
      emit(AuthState(status: AuthStatus.authenticated, user: user));
    } on GraphQLException catch (e) {
      final expired = e.reason == 'AUTH_MFA_CHALLENGE_INVALID';
      emit(AuthState(
        status: AuthStatus.unauthenticated,
        mfaToken: expired ? null : token,
        error: expired ? 'Phiên xác thực đã hết hạn, vui lòng đăng nhập lại' : e.message,
      ));
    } catch (e) {
      emit(AuthState(status: AuthStatus.unauthenticated, mfaToken: token, error: e.toString()));
    }
  }

  Future<void> _onLogout(AuthLogoutRequested event, Emitter<AuthState> emit) async {
    try {
      await _repository.logout();
    } catch (_) {}
    emit(const AuthState(status: AuthStatus.unauthenticated));
  }

  @override
  Future<void> close() {
    _expiredSub.cancel();
    return super.close();
  }
}
