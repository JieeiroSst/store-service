// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

import '../gateway_executor.dart';
import '../models.g.dart';

class LuckyApi {
  const LuckyApi(this._executor);

  final GatewayExecutor _executor;

  Future<List<LuckyCampaign>?> luckyGames({String? selectionSet}) async {
    final data = await _executor.public('query luckyGames { luckyGames ${selectionSet ?? '{ id game_type name description start_at end_at timezone status win_rate daily_spins max_spins_per_day max_points_per_user_day budget_points points_awarded review_points spin_rules { event spins max_per_day } extra_play_events checkin_rewards settings prizes { id campaign_id name type points weight quantity awarded position image_url } created_at updated_at }'} }'
    );
    return (data['luckyGames'] as List?)?.map((e) => (e == null ? null : LuckyCampaign.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<LuckyCampaign?> luckyGame({required String game, String? selectionSet}) async {
    final data = await _executor.public('query luckyGame(\$game: String!) { luckyGame(game: \$game) ${selectionSet ?? '{ id game_type name description start_at end_at timezone status win_rate daily_spins max_spins_per_day max_points_per_user_day budget_points points_awarded review_points spin_rules { event spins max_per_day } extra_play_events checkin_rewards settings prizes { id campaign_id name type points weight quantity awarded position image_url } created_at updated_at }'} }', variables: {
      'game': game,
    }
    );
    return data['luckyGame'] == null ? null : LuckyCampaign.fromJson((data['luckyGame'] as Map).cast<String, dynamic>());
  }

  Future<List<LuckyWinner>?> luckyWinners({required String game, int? limit, String? selectionSet}) async {
    final data = await _executor.public('query luckyWinners(\$game: String!, \$limit: Int) { luckyWinners(game: \$game, limit: \$limit) ${selectionSet ?? '{ campaign_id game_type user prize_name points at }'} }', variables: {
      'game': game,
      'limit': ?limit,
    }
    );
    return (data['luckyWinners'] as List?)?.map((e) => (e == null ? null : LuckyWinner.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<LuckyPlayerStatus?> luckyMyStatus({required String game, String? selectionSet}) async {
    final data = await _executor.query('query luckyMyStatus(\$game: String!) { luckyMyStatus(game: \$game) ${selectionSet ?? '{ campaign_id game_type day daily_spins remaining usage { daily_today credit_today credit_balance } ends_at }'} }', variables: {
      'game': game,
    }
    );
    return data['luckyMyStatus'] == null ? null : LuckyPlayerStatus.fromJson((data['luckyMyStatus'] as Map).cast<String, dynamic>());
  }

  Future<LuckyPlayPage?> luckyMyPlays({String? gameType, int? campaignId, int? limit, int? offset, String? selectionSet}) async {
    final data = await _executor.query('query luckyMyPlays(\$game_type: String, \$campaign_id: Int64, \$limit: Int, \$offset: Int) { luckyMyPlays(game_type: \$game_type, campaign_id: \$campaign_id, limit: \$limit, offset: \$offset) ${selectionSet ?? '{ items { id campaign_id game_type user_id prize_id prize_name prize_type points position choice source day client { ip device_id user_agent } created_at } total limit offset }'} }', variables: {
      'game_type': ?gameType,
      'campaign_id': ?campaignId,
      'limit': ?limit,
      'offset': ?offset,
    }
    );
    return data['luckyMyPlays'] == null ? null : LuckyPlayPage.fromJson((data['luckyMyPlays'] as Map).cast<String, dynamic>());
  }

  Future<LuckyCheckinStatus?> luckyCheckinStatus({String? selectionSet}) async {
    final data = await _executor.query('query luckyCheckinStatus { luckyCheckinStatus ${selectionSet ?? '{ campaign_id day checked_in_today streak cycle_day next_reward rewards ends_at }'} }'
    );
    return data['luckyCheckinStatus'] == null ? null : LuckyCheckinStatus.fromJson((data['luckyCheckinStatus'] as Map).cast<String, dynamic>());
  }

  Future<LuckyCheckinPage?> luckyCheckinHistory({int? limit, int? offset, String? selectionSet}) async {
    final data = await _executor.query('query luckyCheckinHistory(\$limit: Int, \$offset: Int) { luckyCheckinHistory(limit: \$limit, offset: \$offset) ${selectionSet ?? '{ items { id campaign_id user_id day streak points created_at } total limit offset }'} }', variables: {
      'limit': ?limit,
      'offset': ?offset,
    }
    );
    return data['luckyCheckinHistory'] == null ? null : LuckyCheckinPage.fromJson((data['luckyCheckinHistory'] as Map).cast<String, dynamic>());
  }

  Future<LuckyCampaignPage?> luckyCampaigns({String? gameType, String? status, int? limit, int? offset, String? selectionSet}) async {
    final data = await _executor.query('query luckyCampaigns(\$game_type: String, \$status: String, \$limit: Int, \$offset: Int) { luckyCampaigns(game_type: \$game_type, status: \$status, limit: \$limit, offset: \$offset) ${selectionSet ?? '{ items { id game_type name description start_at end_at timezone status win_rate daily_spins max_spins_per_day max_points_per_user_day budget_points points_awarded review_points spin_rules { event spins max_per_day } extra_play_events checkin_rewards settings prizes { id campaign_id name type points weight quantity awarded position image_url } created_at updated_at } total limit offset }'} }', variables: {
      'game_type': ?gameType,
      'status': ?status,
      'limit': ?limit,
      'offset': ?offset,
    }
    );
    return data['luckyCampaigns'] == null ? null : LuckyCampaignPage.fromJson((data['luckyCampaigns'] as Map).cast<String, dynamic>());
  }

  Future<LuckyCampaign?> luckyCampaign({required int id, String? selectionSet}) async {
    final data = await _executor.query('query luckyCampaign(\$id: Int64!) { luckyCampaign(id: \$id) ${selectionSet ?? '{ id game_type name description start_at end_at timezone status win_rate daily_spins max_spins_per_day max_points_per_user_day budget_points points_awarded review_points spin_rules { event spins max_per_day } extra_play_events checkin_rewards settings prizes { id campaign_id name type points weight quantity awarded position image_url } created_at updated_at }'} }', variables: {
      'id': id,
    }
    );
    return data['luckyCampaign'] == null ? null : LuckyCampaign.fromJson((data['luckyCampaign'] as Map).cast<String, dynamic>());
  }

  Future<LuckyBlockPage?> luckyFraudBlocks({int? limit, int? offset, String? selectionSet}) async {
    final data = await _executor.query('query luckyFraudBlocks(\$limit: Int, \$offset: Int) { luckyFraudBlocks(limit: \$limit, offset: \$offset) ${selectionSet ?? '{ items { user_id reason blocked_by created_at } total limit offset }'} }', variables: {
      'limit': ?limit,
      'offset': ?offset,
    }
    );
    return data['luckyFraudBlocks'] == null ? null : LuckyBlockPage.fromJson((data['luckyFraudBlocks'] as Map).cast<String, dynamic>());
  }

  Future<LuckyReviewRewardPage?> luckyFraudRewards({String? status, String? userId, int? limit, int? offset, String? selectionSet}) async {
    final data = await _executor.query('query luckyFraudRewards(\$status: String, \$user_id: String, \$limit: Int, \$offset: Int) { luckyFraudRewards(status: \$status, user_id: \$user_id, limit: \$limit, offset: \$offset) ${selectionSet ?? '{ items { id event_id source game_type user_id campaign_id reference_id title points status hold_reason reviewed_by occurred_at published_at } total limit offset }'} }', variables: {
      'status': ?status,
      'user_id': ?userId,
      'limit': ?limit,
      'offset': ?offset,
    }
    );
    return data['luckyFraudRewards'] == null ? null : LuckyReviewRewardPage.fromJson((data['luckyFraudRewards'] as Map).cast<String, dynamic>());
  }

  Future<LuckyPlayPage?> luckyFraudPlays({String? userId, String? deviceId, String? ip, int? limit, int? offset, String? selectionSet}) async {
    final data = await _executor.query('query luckyFraudPlays(\$user_id: String, \$device_id: String, \$ip: String, \$limit: Int, \$offset: Int) { luckyFraudPlays(user_id: \$user_id, device_id: \$device_id, ip: \$ip, limit: \$limit, offset: \$offset) ${selectionSet ?? '{ items { id campaign_id game_type user_id prize_id prize_name prize_type points position choice source day client { ip device_id user_agent } created_at } total limit offset }'} }', variables: {
      'user_id': ?userId,
      'device_id': ?deviceId,
      'ip': ?ip,
      'limit': ?limit,
      'offset': ?offset,
    }
    );
    return data['luckyFraudPlays'] == null ? null : LuckyPlayPage.fromJson((data['luckyFraudPlays'] as Map).cast<String, dynamic>());
  }

  Future<LuckyPlayPage?> luckyCampaignPlays({required int id, String? userId, int? limit, int? offset, String? selectionSet}) async {
    final data = await _executor.query('query luckyCampaignPlays(\$id: Int64!, \$user_id: String, \$limit: Int, \$offset: Int) { luckyCampaignPlays(id: \$id, user_id: \$user_id, limit: \$limit, offset: \$offset) ${selectionSet ?? '{ items { id campaign_id game_type user_id prize_id prize_name prize_type points position choice source day client { ip device_id user_agent } created_at } total limit offset }'} }', variables: {
      'id': id,
      'user_id': ?userId,
      'limit': ?limit,
      'offset': ?offset,
    }
    );
    return data['luckyCampaignPlays'] == null ? null : LuckyPlayPage.fromJson((data['luckyCampaignPlays'] as Map).cast<String, dynamic>());
  }

  Future<LuckyPlayResult?> luckyPlay({required String game, int? choice, String? idempotencyKey, String? selectionSet}) async {
    final data = await _executor.mutation('mutation luckyPlay(\$game: String!, \$choice: Int, \$idempotency_key: String) { luckyPlay(game: \$game, choice: \$choice, idempotency_key: \$idempotency_key) ${selectionSet ?? '{ play { id campaign_id game_type user_id prize_id prize_name prize_type points position choice source day client { ip device_id user_agent } created_at } won replayed status { campaign_id game_type day daily_spins remaining usage { daily_today credit_today credit_balance } ends_at } }'} }', variables: {
      'game': game,
      'choice': ?choice,
      'idempotency_key': ?idempotencyKey,
    }
    );
    return data['luckyPlay'] == null ? null : LuckyPlayResult.fromJson((data['luckyPlay'] as Map).cast<String, dynamic>());
  }

  Future<LuckyCheckinResult?> luckyCheckIn({String? selectionSet}) async {
    final data = await _executor.mutation('mutation luckyCheckIn { luckyCheckIn ${selectionSet ?? '{ checkin { id campaign_id user_id day streak points created_at } status { campaign_id day checked_in_today streak cycle_day next_reward rewards ends_at } }'} }'
    );
    return data['luckyCheckIn'] == null ? null : LuckyCheckinResult.fromJson((data['luckyCheckIn'] as Map).cast<String, dynamic>());
  }

  Future<LuckyGrantResult?> luckyGrantPlays({required String userId, required String event, required String referenceId, String? selectionSet}) async {
    final data = await _executor.mutation('mutation luckyGrantPlays(\$user_id: String!, \$event: String!, \$reference_id: String!) { luckyGrantPlays(user_id: \$user_id, event: \$event, reference_id: \$reference_id) ${selectionSet ?? '{ grants { id campaign_id game_type user_id event reference_id spins day created_at } skipped { campaign_id game_type reason } }'} }', variables: {
      'user_id': userId,
      'event': event,
      'reference_id': referenceId,
    }
    );
    return data['luckyGrantPlays'] == null ? null : LuckyGrantResult.fromJson((data['luckyGrantPlays'] as Map).cast<String, dynamic>());
  }

  Future<LuckyCampaign?> luckyCampaignCreate({required LuckyCampaignInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation luckyCampaignCreate(\$input: LuckyCampaignInput!) { luckyCampaignCreate(input: \$input) ${selectionSet ?? '{ id game_type name description start_at end_at timezone status win_rate daily_spins max_spins_per_day max_points_per_user_day budget_points points_awarded review_points spin_rules { event spins max_per_day } extra_play_events checkin_rewards settings prizes { id campaign_id name type points weight quantity awarded position image_url } created_at updated_at }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['luckyCampaignCreate'] == null ? null : LuckyCampaign.fromJson((data['luckyCampaignCreate'] as Map).cast<String, dynamic>());
  }

  Future<LuckyCampaign?> luckyCampaignUpdate({required int id, required LuckyCampaignInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation luckyCampaignUpdate(\$id: Int64!, \$input: LuckyCampaignInput!) { luckyCampaignUpdate(id: \$id, input: \$input) ${selectionSet ?? '{ id game_type name description start_at end_at timezone status win_rate daily_spins max_spins_per_day max_points_per_user_day budget_points points_awarded review_points spin_rules { event spins max_per_day } extra_play_events checkin_rewards settings prizes { id campaign_id name type points weight quantity awarded position image_url } created_at updated_at }'} }', variables: {
      'id': id,
      'input': input.toJson(),
    }
    );
    return data['luckyCampaignUpdate'] == null ? null : LuckyCampaign.fromJson((data['luckyCampaignUpdate'] as Map).cast<String, dynamic>());
  }

  Future<bool?> luckyCampaignDelete({required int id}) async {
    final data = await _executor.mutation('mutation luckyCampaignDelete(\$id: Int64!) { luckyCampaignDelete(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['luckyCampaignDelete'] as bool?;
  }

  Future<LuckyCampaign?> luckyCampaignActivate({required int id, String? selectionSet}) async {
    final data = await _executor.mutation('mutation luckyCampaignActivate(\$id: Int64!) { luckyCampaignActivate(id: \$id) ${selectionSet ?? '{ id game_type name description start_at end_at timezone status win_rate daily_spins max_spins_per_day max_points_per_user_day budget_points points_awarded review_points spin_rules { event spins max_per_day } extra_play_events checkin_rewards settings prizes { id campaign_id name type points weight quantity awarded position image_url } created_at updated_at }'} }', variables: {
      'id': id,
    }
    );
    return data['luckyCampaignActivate'] == null ? null : LuckyCampaign.fromJson((data['luckyCampaignActivate'] as Map).cast<String, dynamic>());
  }

  Future<LuckyBlockResult?> luckyFraudBlock({required String userId, required String reason, String? selectionSet}) async {
    final data = await _executor.mutation('mutation luckyFraudBlock(\$user_id: String!, \$reason: String!) { luckyFraudBlock(user_id: \$user_id, reason: \$reason) ${selectionSet ?? '{ block { user_id reason blocked_by created_at } held_rewards }'} }', variables: {
      'user_id': userId,
      'reason': reason,
    }
    );
    return data['luckyFraudBlock'] == null ? null : LuckyBlockResult.fromJson((data['luckyFraudBlock'] as Map).cast<String, dynamic>());
  }

  Future<bool?> luckyFraudUnblock({required String userId}) async {
    final data = await _executor.mutation('mutation luckyFraudUnblock(\$user_id: String!) { luckyFraudUnblock(user_id: \$user_id) }', variables: {
      'user_id': userId,
    }
    );
    return data['luckyFraudUnblock'] as bool?;
  }

  Future<LuckyReviewReward?> luckyFraudApproveReward({required int id, String? selectionSet}) async {
    final data = await _executor.mutation('mutation luckyFraudApproveReward(\$id: Int64!) { luckyFraudApproveReward(id: \$id) ${selectionSet ?? '{ id event_id source game_type user_id campaign_id reference_id title points status hold_reason reviewed_by occurred_at published_at }'} }', variables: {
      'id': id,
    }
    );
    return data['luckyFraudApproveReward'] == null ? null : LuckyReviewReward.fromJson((data['luckyFraudApproveReward'] as Map).cast<String, dynamic>());
  }

  Future<LuckyReviewReward?> luckyFraudRejectReward({required int id, String? selectionSet}) async {
    final data = await _executor.mutation('mutation luckyFraudRejectReward(\$id: Int64!) { luckyFraudRejectReward(id: \$id) ${selectionSet ?? '{ id event_id source game_type user_id campaign_id reference_id title points status hold_reason reviewed_by occurred_at published_at }'} }', variables: {
      'id': id,
    }
    );
    return data['luckyFraudRejectReward'] == null ? null : LuckyReviewReward.fromJson((data['luckyFraudRejectReward'] as Map).cast<String, dynamic>());
  }

  Future<LuckyCampaign?> luckyCampaignPause({required int id, String? selectionSet}) async {
    final data = await _executor.mutation('mutation luckyCampaignPause(\$id: Int64!) { luckyCampaignPause(id: \$id) ${selectionSet ?? '{ id game_type name description start_at end_at timezone status win_rate daily_spins max_spins_per_day max_points_per_user_day budget_points points_awarded review_points spin_rules { event spins max_per_day } extra_play_events checkin_rewards settings prizes { id campaign_id name type points weight quantity awarded position image_url } created_at updated_at }'} }', variables: {
      'id': id,
    }
    );
    return data['luckyCampaignPause'] == null ? null : LuckyCampaign.fromJson((data['luckyCampaignPause'] as Map).cast<String, dynamic>());
  }

}
