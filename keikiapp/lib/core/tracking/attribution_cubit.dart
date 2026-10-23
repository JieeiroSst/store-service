import 'dart:async';
import 'dart:convert';

import 'package:equatable/equatable.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';

import '../../gen/gateway_api.g.dart';
import '../network/graphql_exception.dart';
import 'link_source.dart';
import 'link_touch.dart';

abstract interface class AttributionStore {
  Future<String?> read();

  Future<void> write(String value);
}

class SecureAttributionStore implements AttributionStore {
  const SecureAttributionStore();

  static const _key = 'tracking.attribution';
  static const _storage = FlutterSecureStorage();

  @override
  Future<String?> read() async {
    try {
      return await _storage.read(key: _key);
    } catch (_) {
      return null;
    }
  }

  @override
  Future<void> write(String value) async {
    try {
      await _storage.write(key: _key, value: value);
    } catch (_) {}
  }
}

class MemoryAttributionStore implements AttributionStore {
  String? value;

  @override
  Future<String?> read() async => value;

  @override
  Future<void> write(String v) async => value = v;
}

class AttributionResult extends Equatable {
  const AttributionResult({required this.at, required this.refCode, this.attributed = false, this.ownerUserId, this.rewardType, this.error});

  final DateTime at;
  final String refCode;
  final bool attributed;
  final String? ownerUserId;
  final String? rewardType;
  final String? error;

  bool get ok => error == null;

  Map<String, Object?> toJson() => {
        'at': at.toIso8601String(),
        'ref_code': refCode,
        'attributed': attributed,
        'owner_user_id': ?ownerUserId,
        'reward_type': ?rewardType,
        'error': ?error,
      };

  factory AttributionResult.fromJson(Map<String, dynamic> j) => AttributionResult(
        at: DateTime.tryParse('${j['at']}') ?? DateTime.now(),
        refCode: '${j['ref_code'] ?? ''}',
        attributed: j['attributed'] == true,
        ownerUserId: j['owner_user_id'] as String?,
        rewardType: j['reward_type'] as String?,
        error: j['error'] as String?,
      );

  @override
  List<Object?> get props => [at, refCode, attributed, ownerUserId, rewardType, error];
}

class AttributionState extends Equatable {
  const AttributionState({
    this.sourceLabel = '',
    this.enabled = false,
    this.ready = false,
    this.initError,
    this.pending,
    this.history = const [],
    this.lastResult,
    this.activating = false,
  });

  final String sourceLabel;
  final bool enabled;
  final bool ready;
  final String? initError;
  final LinkTouch? pending;
  final List<LinkTouch> history;
  final AttributionResult? lastResult;
  final bool activating;

  AttributionState copyWith({
    bool? ready,
    String? initError,
    LinkTouch? pending,
    bool clearPending = false,
    List<LinkTouch>? history,
    AttributionResult? lastResult,
    bool? activating,
  }) =>
      AttributionState(
        sourceLabel: sourceLabel,
        enabled: enabled,
        ready: ready ?? this.ready,
        initError: initError ?? this.initError,
        pending: clearPending ? null : (pending ?? this.pending),
        history: history ?? this.history,
        lastResult: lastResult ?? this.lastResult,
        activating: activating ?? this.activating,
      );

  @override
  List<Object?> get props => [sourceLabel, enabled, ready, initError, pending?.toJson().toString(), history.length, history.firstOrNull?.receivedAt, lastResult, activating];
}

const _permanentErrors = ['self', 'already', 'not found', 'not_found', 'expired', 'used', 'inactive', 'invalid', 'không tồn tại', 'hết hạn'];

class AttributionCubit extends Cubit<AttributionState> {
  AttributionCubit({required this.source, required this.api, required this.deviceId, this.store = const SecureAttributionStore()})
      : super(AttributionState(sourceLabel: source.label, enabled: source.enabled));

  final LinkSource source;
  final GatewayApi api;
  final Future<String> Function() deviceId;
  final AttributionStore store;
  StreamSubscription<Map<dynamic, dynamic>>? _sub;
  String? _userId;
  void Function(String route)? onNavigate;

  Future<void> start() async {
    await _restore();
    if (!source.enabled) {
      emit(state.copyWith(ready: true));
      return;
    }
    try {
      await source.init();
      _sub = source.sessions().listen(_onParams, onError: (Object e) => emit(state.copyWith(initError: '$e')));
      emit(state.copyWith(ready: true));
    } catch (e) {
      emit(state.copyWith(ready: true, initError: '$e'));
    }
  }

  Future<void> _restore() async {
    final raw = await store.read();
    if (raw == null) return;
    try {
      final j = jsonDecode(raw) as Map<String, dynamic>;
      emit(state.copyWith(
        pending: j['pending'] == null ? null : LinkTouch.fromJson((j['pending'] as Map).cast<String, dynamic>()),
        history: [for (final h in (j['history'] as List? ?? const [])) LinkTouch.fromJson((h as Map).cast<String, dynamic>())],
        lastResult: j['last_result'] == null ? null : AttributionResult.fromJson((j['last_result'] as Map).cast<String, dynamic>()),
      ));
    } catch (_) {}
  }

  Future<void> _persist() => store.write(jsonEncode({
        'pending': state.pending?.toJson(),
        'history': [for (final h in state.history.take(20)) h.toJson()],
        'last_result': state.lastResult?.toJson(),
      }));

  void _onParams(Map<dynamic, dynamic> params) {
    final touch = LinkTouch.fromBranch(params);
    if (touch != null) unawaited(record(touch, navigate: touch.clickedBranchLink));
  }

  void observeUri(Uri uri, {LinkTouchSource source = LinkTouchSource.deepLink}) {
    final touch = LinkTouch.fromUri(uri, source: source);
    if (touch != null && (touch.isReferral || touch.utm.isNotEmpty)) unawaited(record(touch, navigate: false));
  }

  bool _duplicate(LinkTouch t) {
    final last = state.history.firstOrNull;
    if (last == null) return false;
    final close = t.receivedAt.difference(last.receivedAt).inSeconds.abs() < 10;
    return close && last.url == t.url && last.clickId == t.clickId && last.refCode == t.refCode;
  }

  Future<void> record(LinkTouch touch, {bool navigate = true}) async {
    if (_duplicate(touch)) return;
    emit(state.copyWith(history: [touch, ...state.history].take(20).toList(), pending: touch.isReferral ? touch : null));
    await _persist();
    final route = touch.route;
    if (navigate && route != null) onNavigate?.call(route);
    await _activate();
  }

  Future<AttributionResult?> applyManual(String code) async {
    final touch = LinkTouch(source: LinkTouchSource.manual, receivedAt: DateTime.now(), refCode: code.trim().toUpperCase(), channel: 'manual');
    emit(state.copyWith(history: [touch, ...state.history].take(20).toList(), pending: touch));
    await _persist();
    await _activate(rethrowErrors: true);
    return state.lastResult;
  }

  Future<void> onAuthenticated(String userId) async {
    _userId = userId;
    if (source.enabled) {
      try {
        source.identify(userId);
      } catch (_) {}
    }
    await _activate();
  }

  void onLoggedOut() {
    _userId = null;
    if (source.enabled) {
      try {
        source.logout();
      } catch (_) {}
    }
  }

  Future<void> retry() => _activate();

  Future<void> _activate({bool rethrowErrors = false}) async {
    final touch = state.pending;
    if (touch == null || _userId == null || state.activating || !touch.isReferral) return;
    emit(state.copyWith(activating: true));
    try {
      final platform = defaultTargetPlatform == TargetPlatform.iOS ? 'ios' : defaultTargetPlatform.name.toLowerCase();
      final device = await deviceId();
      final r = touch.firstSession
          ? await api.referral.referralConfirmInstall(
              refCode: touch.refCode!,
              clickId: touch.clickId,
              platform: platform,
              deviceId: device,
              utmSource: touch.utm['utm_source'],
              utmMedium: touch.utm['utm_medium'],
              utmCampaign: touch.utm['utm_campaign'],
              utmTerm: touch.utm['utm_term'],
              utmContent: touch.utm['utm_content'],
            )
          : await api.referral.referralActivate(
              refCode: touch.refCode!,
              clickId: touch.clickId,
              platform: platform,
              deviceId: device,
              utmSource: touch.utm['utm_source'],
              utmMedium: touch.utm['utm_medium'],
              utmCampaign: touch.utm['utm_campaign'],
              utmTerm: touch.utm['utm_term'],
              utmContent: touch.utm['utm_content'],
            );
      final m = r is Map ? r : const {};
      emit(state.copyWith(
        activating: false,
        clearPending: true,
        lastResult: AttributionResult(
          at: DateTime.now(),
          refCode: touch.refCode!,
          attributed: m['attributed'] == true,
          ownerUserId: m['owner_user_id']?.toString(),
          rewardType: m['reward_type']?.toString(),
        ),
      ));
      await _persist();
    } catch (e) {
      final message = e is GraphQLException ? e.message : '$e';
      final permanent = e is GraphQLException && _permanentErrors.any(message.toLowerCase().contains);
      emit(state.copyWith(
        activating: false,
        clearPending: permanent,
        lastResult: AttributionResult(at: DateTime.now(), refCode: touch.refCode!, error: message),
      ));
      await _persist();
      if (rethrowErrors) rethrow;
    }
  }

  @override
  Future<void> close() async {
    await _sub?.cancel();
    return super.close();
  }
}
