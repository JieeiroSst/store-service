// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

import '../gateway_executor.dart';
import '../models.g.dart';

class BonuslinkApi {
  const BonuslinkApi(this._executor);

  final GatewayExecutor _executor;

  Future<List<BonusReward>?> bonusUserRewards({required String userId, int? limit, int? offset, String? selectionSet}) async {
    final data = await _executor.query('query bonusUserRewards(\$user_id: String!, \$limit: Int, \$offset: Int) { bonusUserRewards(user_id: \$user_id, limit: \$limit, offset: \$offset) ${selectionSet ?? '{ reward_id event_id user_id ref_code source reward_type reward_value created_at }'} }', variables: {
      'user_id': userId,
      'limit': ?limit,
      'offset': ?offset,
    }
    );
    return (data['bonusUserRewards'] as List?)?.map((e) => (e == null ? null : BonusReward.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<List<BonusBalance>?> bonusUserBalances({required String userId, String? selectionSet}) async {
    final data = await _executor.query('query bonusUserBalances(\$user_id: String!) { bonusUserBalances(user_id: \$user_id) ${selectionSet ?? '{ user_id reward_type total_value reward_count }'} }', variables: {
      'user_id': userId,
    }
    );
    return (data['bonusUserBalances'] as List?)?.map((e) => (e == null ? null : BonusBalance.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<List<BonusReward>?> myBonusRewards({int? limit, int? offset, String? selectionSet}) async {
    final data = await _executor.query('query myBonusRewards(\$limit: Int, \$offset: Int) { myBonusRewards(limit: \$limit, offset: \$offset) ${selectionSet ?? '{ reward_id event_id user_id ref_code source reward_type reward_value created_at }'} }', variables: {
      'limit': ?limit,
      'offset': ?offset,
    }
    );
    return (data['myBonusRewards'] as List?)?.map((e) => (e == null ? null : BonusReward.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<List<BonusBalance>?> myBonusBalances({String? selectionSet}) async {
    final data = await _executor.query('query myBonusBalances { myBonusBalances ${selectionSet ?? '{ user_id reward_type total_value reward_count }'} }'
    );
    return (data['myBonusBalances'] as List?)?.map((e) => (e == null ? null : BonusBalance.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<BonusRecordResult?> bonusRecordReward({required BonusRecordRewardInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation bonusRecordReward(\$input: BonusRecordRewardInput!) { bonusRecordReward(input: \$input) ${selectionSet ?? '{ reward_id event_id user_id ref_code source reward_type reward_value created_at duplicate }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['bonusRecordReward'] == null ? null : BonusRecordResult.fromJson((data['bonusRecordReward'] as Map).cast<String, dynamic>());
  }

}
