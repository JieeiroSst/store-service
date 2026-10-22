// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

import '../gateway_executor.dart';
import '../models.g.dart';

class ReferralApi {
  const ReferralApi(this._executor);

  final GatewayExecutor _executor;

  Future<ReferralReferralLink?> referralLink({required String refCode, String? selectionSet}) async {
    final data = await _executor.public('query referralLink(\$ref_code: String!) { referralLink(ref_code: \$ref_code) ${selectionSet ?? '{ ref_code created_at owner_user_id channel status expires_at deep_link platform share_url utm_source utm_medium utm_campaign utm_term utm_content }'} }', variables: {
      'ref_code': refCode,
    }
    );
    return data['referralLink'] == null ? null : ReferralReferralLink.fromJson((data['referralLink'] as Map).cast<String, dynamic>());
  }

  Future<ReferralLinkList?> referralUserLinks({required String userId, int? limit, String? cursor, String? selectionSet}) async {
    final data = await _executor.query('query referralUserLinks(\$user_id: String!, \$limit: Int, \$cursor: String) { referralUserLinks(user_id: \$user_id, limit: \$limit, cursor: \$cursor) ${selectionSet ?? '{ data { ref_code created_at owner_user_id channel status expires_at deep_link platform share_url utm_source utm_medium utm_campaign utm_term utm_content } meta { count next_cursor } }'} }', variables: {
      'user_id': userId,
      'limit': ?limit,
      'cursor': ?cursor,
    }
    );
    return data['referralUserLinks'] == null ? null : ReferralLinkList.fromJson((data['referralUserLinks'] as Map).cast<String, dynamic>());
  }

  Future<ReferralReferralStatusResponse?> referralStatus({required String refCode, String? selectionSet}) async {
    final data = await _executor.query('query referralStatus(\$ref_code: String!) { referralStatus(ref_code: \$ref_code) ${selectionSet ?? '{ ref_code status invitation_status owner_user_id invited_at activated_at time_to_install_ms platform new_user_id reward_status reward_value utm_source utm_medium utm_campaign utm_term utm_content }'} }', variables: {
      'ref_code': refCode,
    }
    );
    return data['referralStatus'] == null ? null : ReferralReferralStatusResponse.fromJson((data['referralStatus'] as Map).cast<String, dynamic>());
  }

  Future<ReferralUserReferralStats?> referralUserStats({required String userId, String? selectionSet}) async {
    final data = await _executor.query('query referralUserStats(\$user_id: String!) { referralUserStats(user_id: \$user_id) ${selectionSet ?? '{ user_id total_invited total_installed total_rewarded total_reward_amt last_active_at }'} }', variables: {
      'user_id': userId,
    }
    );
    return data['referralUserStats'] == null ? null : ReferralUserReferralStats.fromJson((data['referralUserStats'] as Map).cast<String, dynamic>());
  }

  Future<List<ReferralShareTarget>?> referralShareTargets({required String refCode, String? selectionSet}) async {
    final data = await _executor.public('query referralShareTargets(\$ref_code: String!) { referralShareTargets(ref_code: \$ref_code) ${selectionSet ?? '{ channel label share_url qr_url utm_source utm_medium utm_campaign utm_term utm_content }'} }', variables: {
      'ref_code': refCode,
    }
    );
    return (data['referralShareTargets'] as List?)?.map((e) => (e == null ? null : ReferralShareTarget.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<ReferralRewardProgram?> referralActiveRewardProgram({String? selectionSet}) async {
    final data = await _executor.public('query referralActiveRewardProgram { referralActiveRewardProgram ${selectionSet ?? '{ program_id name status tiers { min_count max_count reward_value } created_at updated_at }'} }'
    );
    return data['referralActiveRewardProgram'] == null ? null : ReferralRewardProgram.fromJson((data['referralActiveRewardProgram'] as Map).cast<String, dynamic>());
  }

  Future<ReferralLinkList?> myReferralLinks({int? limit, String? cursor, String? selectionSet}) async {
    final data = await _executor.query('query myReferralLinks(\$limit: Int, \$cursor: String) { myReferralLinks(limit: \$limit, cursor: \$cursor) ${selectionSet ?? '{ data { ref_code created_at owner_user_id channel status expires_at deep_link platform share_url utm_source utm_medium utm_campaign utm_term utm_content } meta { count next_cursor } }'} }', variables: {
      'limit': ?limit,
      'cursor': ?cursor,
    }
    );
    return data['myReferralLinks'] == null ? null : ReferralLinkList.fromJson((data['myReferralLinks'] as Map).cast<String, dynamic>());
  }

  Future<ReferralUserReferralStats?> myReferralStats({String? selectionSet}) async {
    final data = await _executor.query('query myReferralStats { myReferralStats ${selectionSet ?? '{ user_id total_invited total_installed total_rewarded total_reward_amt last_active_at }'} }'
    );
    return data['myReferralStats'] == null ? null : ReferralUserReferralStats.fromJson((data['myReferralStats'] as Map).cast<String, dynamic>());
  }

  Future<ReferralAttribution?> referralUserAttribution({required String userId, String? selectionSet}) async {
    final data = await _executor.query('query referralUserAttribution(\$user_id: String!) { referralUserAttribution(user_id: \$user_id) ${selectionSet ?? '{ owner_user_id new_user_id ref_code platform device_id attributed_at firebase_instance_id utm_source utm_medium utm_campaign utm_term utm_content }'} }', variables: {
      'user_id': userId,
    }
    );
    return data['referralUserAttribution'] == null ? null : ReferralAttribution.fromJson((data['referralUserAttribution'] as Map).cast<String, dynamic>());
  }

  Future<ReferralAttribution?> myReferralAttribution({String? selectionSet}) async {
    final data = await _executor.query('query myReferralAttribution { myReferralAttribution ${selectionSet ?? '{ owner_user_id new_user_id ref_code platform device_id attributed_at firebase_instance_id utm_source utm_medium utm_campaign utm_term utm_content }'} }'
    );
    return data['myReferralAttribution'] == null ? null : ReferralAttribution.fromJson((data['myReferralAttribution'] as Map).cast<String, dynamic>());
  }

  Future<List<ReferralUTMPerformance>?> referralUTMReport({String? ownerUserId, String? utmSource, String? utmMedium, String? utmCampaign, int? from, int? to, String? selectionSet}) async {
    final data = await _executor.query('query referralUTMReport(\$owner_user_id: String, \$utm_source: String, \$utm_medium: String, \$utm_campaign: String, \$from: Int, \$to: Int) { referralUTMReport(owner_user_id: \$owner_user_id, utm_source: \$utm_source, utm_medium: \$utm_medium, utm_campaign: \$utm_campaign, from: \$from, to: \$to) ${selectionSet ?? '{ utm_source utm_medium utm_campaign clicks unique_clicks installs rewards conversion_rate }'} }', variables: {
      'owner_user_id': ?ownerUserId,
      'utm_source': ?utmSource,
      'utm_medium': ?utmMedium,
      'utm_campaign': ?utmCampaign,
      'from': ?from,
      'to': ?to,
    }
    );
    return (data['referralUTMReport'] as List?)?.map((e) => (e == null ? null : ReferralUTMPerformance.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<List<ReferralUTMPerformance>?> myReferralUTMReport({String? utmSource, String? utmMedium, String? utmCampaign, int? from, int? to, String? selectionSet}) async {
    final data = await _executor.query('query myReferralUTMReport(\$utm_source: String, \$utm_medium: String, \$utm_campaign: String, \$from: Int, \$to: Int) { myReferralUTMReport(utm_source: \$utm_source, utm_medium: \$utm_medium, utm_campaign: \$utm_campaign, from: \$from, to: \$to) ${selectionSet ?? '{ utm_source utm_medium utm_campaign clicks unique_clicks installs rewards conversion_rate }'} }', variables: {
      'utm_source': ?utmSource,
      'utm_medium': ?utmMedium,
      'utm_campaign': ?utmCampaign,
      'from': ?from,
      'to': ?to,
    }
    );
    return (data['myReferralUTMReport'] as List?)?.map((e) => (e == null ? null : ReferralUTMPerformance.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<ReferralReferralLink?> generateMyReferralLink({String? channel, String? platform, String? utmSource, String? utmMedium, String? utmCampaign, String? utmTerm, String? utmContent, String? selectionSet}) async {
    final data = await _executor.mutation('mutation generateMyReferralLink(\$channel: String, \$platform: String, \$utm_source: String, \$utm_medium: String, \$utm_campaign: String, \$utm_term: String, \$utm_content: String) { generateMyReferralLink(channel: \$channel, platform: \$platform, utm_source: \$utm_source, utm_medium: \$utm_medium, utm_campaign: \$utm_campaign, utm_term: \$utm_term, utm_content: \$utm_content) ${selectionSet ?? '{ ref_code created_at owner_user_id channel status expires_at deep_link platform share_url utm_source utm_medium utm_campaign utm_term utm_content }'} }', variables: {
      'channel': ?channel,
      'platform': ?platform,
      'utm_source': ?utmSource,
      'utm_medium': ?utmMedium,
      'utm_campaign': ?utmCampaign,
      'utm_term': ?utmTerm,
      'utm_content': ?utmContent,
    }
    );
    return data['generateMyReferralLink'] == null ? null : ReferralReferralLink.fromJson((data['generateMyReferralLink'] as Map).cast<String, dynamic>());
  }

  Future<Object?> referralTrackEvent({required String refCode, required String eventType, String? platform, String? channel, String? deviceId, String? firebaseInstanceId, String? utmSource, String? utmMedium, String? utmCampaign, String? utmTerm, String? utmContent}) async {
    final data = await _executor.public('mutation referralTrackEvent(\$ref_code: String!, \$event_type: String!, \$platform: String, \$channel: String, \$device_id: String, \$firebase_instance_id: String, \$utm_source: String, \$utm_medium: String, \$utm_campaign: String, \$utm_term: String, \$utm_content: String) { referralTrackEvent(ref_code: \$ref_code, event_type: \$event_type, platform: \$platform, channel: \$channel, device_id: \$device_id, firebase_instance_id: \$firebase_instance_id, utm_source: \$utm_source, utm_medium: \$utm_medium, utm_campaign: \$utm_campaign, utm_term: \$utm_term, utm_content: \$utm_content) }', variables: {
      'ref_code': refCode,
      'event_type': eventType,
      'platform': ?platform,
      'channel': ?channel,
      'device_id': ?deviceId,
      'firebase_instance_id': ?firebaseInstanceId,
      'utm_source': ?utmSource,
      'utm_medium': ?utmMedium,
      'utm_campaign': ?utmCampaign,
      'utm_term': ?utmTerm,
      'utm_content': ?utmContent,
    }
    );
    return data['referralTrackEvent'];
  }

  Future<Object?> referralConfirmInstall({required String refCode, String? clickId, String? platform, String? deviceId, String? firebaseInstanceId, String? utmSource, String? utmMedium, String? utmCampaign, String? utmTerm, String? utmContent}) async {
    final data = await _executor.mutation('mutation referralConfirmInstall(\$ref_code: String!, \$click_id: String, \$platform: String, \$device_id: String, \$firebase_instance_id: String, \$utm_source: String, \$utm_medium: String, \$utm_campaign: String, \$utm_term: String, \$utm_content: String) { referralConfirmInstall(ref_code: \$ref_code, click_id: \$click_id, platform: \$platform, device_id: \$device_id, firebase_instance_id: \$firebase_instance_id, utm_source: \$utm_source, utm_medium: \$utm_medium, utm_campaign: \$utm_campaign, utm_term: \$utm_term, utm_content: \$utm_content) }', variables: {
      'ref_code': refCode,
      'click_id': ?clickId,
      'platform': ?platform,
      'device_id': ?deviceId,
      'firebase_instance_id': ?firebaseInstanceId,
      'utm_source': ?utmSource,
      'utm_medium': ?utmMedium,
      'utm_campaign': ?utmCampaign,
      'utm_term': ?utmTerm,
      'utm_content': ?utmContent,
    }
    );
    return data['referralConfirmInstall'];
  }

  Future<Object?> referralActivate({required String refCode, String? clickId, String? platform, String? deviceId, String? firebaseInstanceId, String? utmSource, String? utmMedium, String? utmCampaign, String? utmTerm, String? utmContent}) async {
    final data = await _executor.mutation('mutation referralActivate(\$ref_code: String!, \$click_id: String, \$platform: String, \$device_id: String, \$firebase_instance_id: String, \$utm_source: String, \$utm_medium: String, \$utm_campaign: String, \$utm_term: String, \$utm_content: String) { referralActivate(ref_code: \$ref_code, click_id: \$click_id, platform: \$platform, device_id: \$device_id, firebase_instance_id: \$firebase_instance_id, utm_source: \$utm_source, utm_medium: \$utm_medium, utm_campaign: \$utm_campaign, utm_term: \$utm_term, utm_content: \$utm_content) }', variables: {
      'ref_code': refCode,
      'click_id': ?clickId,
      'platform': ?platform,
      'device_id': ?deviceId,
      'firebase_instance_id': ?firebaseInstanceId,
      'utm_source': ?utmSource,
      'utm_medium': ?utmMedium,
      'utm_campaign': ?utmCampaign,
      'utm_term': ?utmTerm,
      'utm_content': ?utmContent,
    }
    );
    return data['referralActivate'];
  }

  Future<ReferralRewardProgram?> referralCreateRewardProgram({required String name_, bool? activate, List<ReferralRewardTierInput>? tiers, String? selectionSet}) async {
    final data = await _executor.mutation('mutation referralCreateRewardProgram(\$name: String!, \$activate: Boolean, \$tiers: [ReferralRewardTierInput!]) { referralCreateRewardProgram(name: \$name, activate: \$activate, tiers: \$tiers) ${selectionSet ?? '{ program_id name status tiers { min_count max_count reward_value } created_at updated_at }'} }', variables: {
      'name': name_,
      'activate': ?activate,
      if (tiers != null) 'tiers': tiers.map((e) => e.toJson()).toList(),
    }
    );
    return data['referralCreateRewardProgram'] == null ? null : ReferralRewardProgram.fromJson((data['referralCreateRewardProgram'] as Map).cast<String, dynamic>());
  }

  Future<ReferralRewardProgram?> referralSetRewardProgramStatus({required String programId, required bool active, String? selectionSet}) async {
    final data = await _executor.mutation('mutation referralSetRewardProgramStatus(\$program_id: String!, \$active: Boolean!) { referralSetRewardProgramStatus(program_id: \$program_id, active: \$active) ${selectionSet ?? '{ program_id name status tiers { min_count max_count reward_value } created_at updated_at }'} }', variables: {
      'program_id': programId,
      'active': active,
    }
    );
    return data['referralSetRewardProgramStatus'] == null ? null : ReferralRewardProgram.fromJson((data['referralSetRewardProgramStatus'] as Map).cast<String, dynamic>());
  }

}
