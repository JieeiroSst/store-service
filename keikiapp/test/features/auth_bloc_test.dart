import 'dart:async';

import 'package:bloc_test/bloc_test.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:keikiapp/core/network/graphql_exception.dart';
import 'package:keikiapp/features/auth/bloc/auth_bloc.dart';
import 'package:keikiapp/features/auth/data/auth_repository.dart';
import 'package:keikiapp/features/auth/data/user.dart';
import 'package:mocktail/mocktail.dart';

class _MockAuthRepository extends Mock implements AuthRepository {}

const _user = User(id: 7, username: 'quan', email: 'q@keiki.vn', name: 'Quan');

void main() {
  late _MockAuthRepository repo;
  late StreamController<void> expired;

  setUp(() {
    repo = _MockAuthRepository();
    expired = StreamController<void>.broadcast();
    when(() => repo.sessionExpired).thenAnswer((_) => expired.stream);
  });

  tearDown(() => expired.close());

  blocTest<AuthBloc, AuthState>(
    'starts unauthenticated without a stored session',
    build: () {
      when(() => repo.hasSession()).thenAnswer((_) async => false);
      return AuthBloc(repo);
    },
    act: (bloc) => bloc.add(const AuthStarted()),
    expect: () => [const AuthState(status: AuthStatus.unauthenticated)],
  );

  blocTest<AuthBloc, AuthState>(
    'restores the user from a stored session',
    build: () {
      when(() => repo.hasSession()).thenAnswer((_) async => true);
      when(() => repo.me()).thenAnswer((_) async => _user);
      return AuthBloc(repo);
    },
    act: (bloc) => bloc.add(const AuthStarted()),
    expect: () => [const AuthState(status: AuthStatus.authenticated, user: _user)],
  );

  blocTest<AuthBloc, AuthState>(
    'logs in and loads the profile',
    build: () {
      when(() => repo.login('quan', 'secret')).thenAnswer((_) async {});
      when(() => repo.me()).thenAnswer((_) async => _user);
      return AuthBloc(repo);
    },
    act: (bloc) => bloc.add(const AuthLoginRequested('quan', 'secret')),
    expect: () => [
      const AuthState(submitting: true),
      const AuthState(status: AuthStatus.authenticated, user: _user),
    ],
  );

  blocTest<AuthBloc, AuthState>(
    'surfaces login errors',
    build: () {
      when(() => repo.login(any(), any())).thenThrow(const GraphQLException('sai mật khẩu', code: 'UNAUTHENTICATED'));
      return AuthBloc(repo);
    },
    act: (bloc) => bloc.add(const AuthLoginRequested('quan', 'bad')),
    expect: () => [
      const AuthState(submitting: true),
      const AuthState(status: AuthStatus.unauthenticated, error: 'sai mật khẩu'),
    ],
  );

  blocTest<AuthBloc, AuthState>(
    'signs out when the client reports an expired session',
    build: () => AuthBloc(repo),
    seed: () => const AuthState(status: AuthStatus.authenticated, user: _user),
    act: (_) => expired.add(null),
    expect: () => [const AuthState(status: AuthStatus.unauthenticated)],
  );
}
