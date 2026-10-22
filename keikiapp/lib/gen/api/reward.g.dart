// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

import '../gateway_executor.dart';
import '../models.g.dart';

class RewardApi {
  const RewardApi(this._executor);

  final GatewayExecutor _executor;

  Future<RewardPage?> rewards({String? name_, int? page, int? pageSize, String? selectionSet}) async {
    final data = await _executor.public('query rewards(\$name: String, \$page: Int, \$page_size: Int) { rewards(name: \$name, page: \$page, page_size: \$page_size) ${selectionSet ?? '{ meta { count page pages } records { id name description points created_at updated_at } }'} }', variables: {
      'name': ?name_,
      'page': ?page,
      'page_size': ?pageSize,
    }
    );
    return data['rewards'] == null ? null : RewardPage.fromJson((data['rewards'] as Map).cast<String, dynamic>());
  }

  Future<RewardPage?> reward({required String rewardId, String? selectionSet}) async {
    final data = await _executor.public('query reward(\$reward_id: String!) { reward(reward_id: \$reward_id) ${selectionSet ?? '{ meta { count page pages } records { id name description points created_at updated_at } }'} }', variables: {
      'reward_id': rewardId,
    }
    );
    return data['reward'] == null ? null : RewardPage.fromJson((data['reward'] as Map).cast<String, dynamic>());
  }

  Future<RewardReward?> rewardCreateReward({required RewardRewardInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation rewardCreateReward(\$input: RewardRewardInput!) { rewardCreateReward(input: \$input) ${selectionSet ?? '{ id name description points created_at updated_at }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['rewardCreateReward'] == null ? null : RewardReward.fromJson((data['rewardCreateReward'] as Map).cast<String, dynamic>());
  }

  Future<RewardReward?> rewardUpdateReward({required String rewardId, required RewardRewardInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation rewardUpdateReward(\$reward_id: String!, \$input: RewardRewardInput!) { rewardUpdateReward(reward_id: \$reward_id, input: \$input) ${selectionSet ?? '{ id name description points created_at updated_at }'} }', variables: {
      'reward_id': rewardId,
      'input': input.toJson(),
    }
    );
    return data['rewardUpdateReward'] == null ? null : RewardReward.fromJson((data['rewardUpdateReward'] as Map).cast<String, dynamic>());
  }

  Future<bool?> rewardDeleteReward({required String rewardId}) async {
    final data = await _executor.mutation('mutation rewardDeleteReward(\$reward_id: String!) { rewardDeleteReward(reward_id: \$reward_id) }', variables: {
      'reward_id': rewardId,
    }
    );
    return data['rewardDeleteReward'] as bool?;
  }

}
