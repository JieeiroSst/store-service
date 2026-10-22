// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

import '../gateway_executor.dart';
import '../models.g.dart';

class AdmanagementApi {
  const AdmanagementApi(this._executor);

  final GatewayExecutor _executor;

  Future<List<AdCampaign>?> adCampaigns({String? selectionSet}) async {
    final data = await _executor.query('query adCampaigns { adCampaigns ${selectionSet ?? '{ id user_id name description budget start_date end_date status created_at updated_at }'} }'
    );
    return (data['adCampaigns'] as List?)?.map((e) => (e == null ? null : AdCampaign.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<AdCampaign?> adCampaign({required int id, String? selectionSet}) async {
    final data = await _executor.query('query adCampaign(\$id: Int!) { adCampaign(id: \$id) ${selectionSet ?? '{ id user_id name description budget start_date end_date status created_at updated_at }'} }', variables: {
      'id': id,
    }
    );
    return data['adCampaign'] == null ? null : AdCampaign.fromJson((data['adCampaign'] as Map).cast<String, dynamic>());
  }

  Future<List<AdAd>?> adCampaignAds({required int id, String? selectionSet}) async {
    final data = await _executor.query('query adCampaignAds(\$id: Int!) { adCampaignAds(id: \$id) ${selectionSet ?? '{ id campaign_id category_id title description ad_type content_url target_url file_path file_size mime_type duration priority is_active start_date end_date created_at updated_at }'} }', variables: {
      'id': id,
    }
    );
    return (data['adCampaignAds'] as List?)?.map((e) => (e == null ? null : AdAd.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<AdCampaignPerformance?> adCampaignPerformance({required int id, String? from, String? to, String? selectionSet}) async {
    final data = await _executor.query('query adCampaignPerformance(\$id: Int!, \$from: String, \$to: String) { adCampaignPerformance(id: \$id, from: \$from, to: \$to) ${selectionSet ?? '{ campaign_id impressions clicks ctr cost revenue }'} }', variables: {
      'id': id,
      'from': ?from,
      'to': ?to,
    }
    );
    return data['adCampaignPerformance'] == null ? null : AdCampaignPerformance.fromJson((data['adCampaignPerformance'] as Map).cast<String, dynamic>());
  }

  Future<List<AdCategory>?> adCategories({String? selectionSet}) async {
    final data = await _executor.query('query adCategories { adCategories ${selectionSet ?? '{ id name description parent_id is_active created_at }'} }'
    );
    return (data['adCategories'] as List?)?.map((e) => (e == null ? null : AdCategory.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<AdCategory?> adCategory({required int id, String? selectionSet}) async {
    final data = await _executor.query('query adCategory(\$id: Int!) { adCategory(id: \$id) ${selectionSet ?? '{ id name description parent_id is_active created_at }'} }', variables: {
      'id': id,
    }
    );
    return data['adCategory'] == null ? null : AdCategory.fromJson((data['adCategory'] as Map).cast<String, dynamic>());
  }

  Future<List<AdPosition>?> adPositions({String? selectionSet}) async {
    final data = await _executor.query('query adPositions { adPositions ${selectionSet ?? '{ id name description width height max_file_size allowed_formats is_active }'} }'
    );
    return (data['adPositions'] as List?)?.map((e) => (e == null ? null : AdPosition.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<AdPosition?> adPosition({required int id, String? selectionSet}) async {
    final data = await _executor.query('query adPosition(\$id: Int!) { adPosition(id: \$id) ${selectionSet ?? '{ id name description width height max_file_size allowed_formats is_active }'} }', variables: {
      'id': id,
    }
    );
    return data['adPosition'] == null ? null : AdPosition.fromJson((data['adPosition'] as Map).cast<String, dynamic>());
  }

  Future<List<AdAd>?> adAds({String? selectionSet}) async {
    final data = await _executor.query('query adAds { adAds ${selectionSet ?? '{ id campaign_id category_id title description ad_type content_url target_url file_path file_size mime_type duration priority is_active start_date end_date created_at updated_at }'} }'
    );
    return (data['adAds'] as List?)?.map((e) => (e == null ? null : AdAd.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<AdAd?> adAd({required int id, String? selectionSet}) async {
    final data = await _executor.query('query adAd(\$id: Int!) { adAd(id: \$id) ${selectionSet ?? '{ id campaign_id category_id title description ad_type content_url target_url file_path file_size mime_type duration priority is_active start_date end_date created_at updated_at }'} }', variables: {
      'id': id,
    }
    );
    return data['adAd'] == null ? null : AdAd.fromJson((data['adAd'] as Map).cast<String, dynamic>());
  }

  Future<List<AdImpression>?> adAdImpressions({required int id, String? selectionSet}) async {
    final data = await _executor.query('query adAdImpressions(\$id: Int!) { adAdImpressions(id: \$id) ${selectionSet ?? '{ id ad_id user_id session_id ip_address user_agent referrer_url page_url position_id created_at }'} }', variables: {
      'id': id,
    }
    );
    return (data['adAdImpressions'] as List?)?.map((e) => (e == null ? null : AdImpression.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<List<AdClick>?> adAdClicks({required int id, String? selectionSet}) async {
    final data = await _executor.query('query adAdClicks(\$id: Int!) { adAdClicks(id: \$id) ${selectionSet ?? '{ id ad_id impression_id user_id session_id ip_address user_agent referrer_url target_url created_at }'} }', variables: {
      'id': id,
    }
    );
    return (data['adAdClicks'] as List?)?.map((e) => (e == null ? null : AdClick.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<List<AdPositionMapping>?> adPositionMappings({String? selectionSet}) async {
    final data = await _executor.query('query adPositionMappings { adPositionMappings ${selectionSet ?? '{ id ad_id position_id weight is_active created_at }'} }'
    );
    return (data['adPositionMappings'] as List?)?.map((e) => (e == null ? null : AdPositionMapping.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<AdPositionMapping?> adPositionMapping({required int id, String? selectionSet}) async {
    final data = await _executor.query('query adPositionMapping(\$id: Int!) { adPositionMapping(id: \$id) ${selectionSet ?? '{ id ad_id position_id weight is_active created_at }'} }', variables: {
      'id': id,
    }
    );
    return data['adPositionMapping'] == null ? null : AdPositionMapping.fromJson((data['adPositionMapping'] as Map).cast<String, dynamic>());
  }

  Future<List<AdImpression>?> adImpressions({String? selectionSet}) async {
    final data = await _executor.query('query adImpressions { adImpressions ${selectionSet ?? '{ id ad_id user_id session_id ip_address user_agent referrer_url page_url position_id created_at }'} }'
    );
    return (data['adImpressions'] as List?)?.map((e) => (e == null ? null : AdImpression.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<AdImpression?> adImpression({required int id, String? selectionSet}) async {
    final data = await _executor.query('query adImpression(\$id: Int!) { adImpression(id: \$id) ${selectionSet ?? '{ id ad_id user_id session_id ip_address user_agent referrer_url page_url position_id created_at }'} }', variables: {
      'id': id,
    }
    );
    return data['adImpression'] == null ? null : AdImpression.fromJson((data['adImpression'] as Map).cast<String, dynamic>());
  }

  Future<List<AdClick>?> adClicks({String? selectionSet}) async {
    final data = await _executor.query('query adClicks { adClicks ${selectionSet ?? '{ id ad_id impression_id user_id session_id ip_address user_agent referrer_url target_url created_at }'} }'
    );
    return (data['adClicks'] as List?)?.map((e) => (e == null ? null : AdClick.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<AdClick?> adClick({required int id, String? selectionSet}) async {
    final data = await _executor.query('query adClick(\$id: Int!) { adClick(id: \$id) ${selectionSet ?? '{ id ad_id impression_id user_id session_id ip_address user_agent referrer_url target_url created_at }'} }', variables: {
      'id': id,
    }
    );
    return data['adClick'] == null ? null : AdClick.fromJson((data['adClick'] as Map).cast<String, dynamic>());
  }

  Future<List<AdTargetingRule>?> adTargetingRules({String? selectionSet}) async {
    final data = await _executor.query('query adTargetingRules { adTargetingRules ${selectionSet ?? '{ id ad_id rule_type rule_operator rule_value is_active created_at }'} }'
    );
    return (data['adTargetingRules'] as List?)?.map((e) => (e == null ? null : AdTargetingRule.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<AdTargetingRule?> adTargetingRule({required int id, String? selectionSet}) async {
    final data = await _executor.query('query adTargetingRule(\$id: Int!) { adTargetingRule(id: \$id) ${selectionSet ?? '{ id ad_id rule_type rule_operator rule_value is_active created_at }'} }', variables: {
      'id': id,
    }
    );
    return data['adTargetingRule'] == null ? null : AdTargetingRule.fromJson((data['adTargetingRule'] as Map).cast<String, dynamic>());
  }

  Future<List<AdPerformanceSummary>?> adPerformanceSummaries({String? selectionSet}) async {
    final data = await _executor.query('query adPerformanceSummaries { adPerformanceSummaries ${selectionSet ?? '{ id ad_id date impressions clicks ctr cost revenue created_at updated_at }'} }'
    );
    return (data['adPerformanceSummaries'] as List?)?.map((e) => (e == null ? null : AdPerformanceSummary.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<AdPerformanceSummary?> adPerformanceSummary({required int id, String? selectionSet}) async {
    final data = await _executor.query('query adPerformanceSummary(\$id: Int!) { adPerformanceSummary(id: \$id) ${selectionSet ?? '{ id ad_id date impressions clicks ctr cost revenue created_at updated_at }'} }', variables: {
      'id': id,
    }
    );
    return data['adPerformanceSummary'] == null ? null : AdPerformanceSummary.fromJson((data['adPerformanceSummary'] as Map).cast<String, dynamic>());
  }

  Future<AdServedAd?> adServe({required int positionId, String? sessionId, String? country, String? device, String? gender, int? age, String? referrerUrl, String? pageUrl, String? selectionSet}) async {
    final data = await _executor.public('query adServe(\$positionId: Int!, \$session_id: String, \$country: String, \$device: String, \$gender: String, \$age: Int, \$referrer_url: String, \$page_url: String) { adServe(positionId: \$positionId, session_id: \$session_id, country: \$country, device: \$device, gender: \$gender, age: \$age, referrer_url: \$referrer_url, page_url: \$page_url) ${selectionSet ?? '{ ad { id campaign_id category_id title description ad_type content_url target_url file_path file_size mime_type duration priority is_active start_date end_date created_at updated_at } impression_id }'} }', variables: {
      'positionId': positionId,
      'session_id': ?sessionId,
      'country': ?country,
      'device': ?device,
      'gender': ?gender,
      'age': ?age,
      'referrer_url': ?referrerUrl,
      'page_url': ?pageUrl,
    }
    );
    return data['adServe'] == null ? null : AdServedAd.fromJson((data['adServe'] as Map).cast<String, dynamic>());
  }

  Future<AdClickResult?> adTrackClick({required int adId, int? impressionId, String? sessionId, String? referrerUrl, String? selectionSet}) async {
    final data = await _executor.public('query adTrackClick(\$ad_id: Int!, \$impression_id: Int, \$session_id: String, \$referrer_url: String) { adTrackClick(ad_id: \$ad_id, impression_id: \$impression_id, session_id: \$session_id, referrer_url: \$referrer_url) ${selectionSet ?? '{ target_url location }'} }', variables: {
      'ad_id': adId,
      'impression_id': ?impressionId,
      'session_id': ?sessionId,
      'referrer_url': ?referrerUrl,
    }
    );
    return data['adTrackClick'] == null ? null : AdClickResult.fromJson((data['adTrackClick'] as Map).cast<String, dynamic>());
  }

  Future<AdCampaign?> adCreateCampaign({required AdCampaignInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation adCreateCampaign(\$input: AdCampaignInput!) { adCreateCampaign(input: \$input) ${selectionSet ?? '{ id user_id name description budget start_date end_date status created_at updated_at }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['adCreateCampaign'] == null ? null : AdCampaign.fromJson((data['adCreateCampaign'] as Map).cast<String, dynamic>());
  }

  Future<AdCampaign?> adUpdateCampaign({required int id, required AdCampaignInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation adUpdateCampaign(\$id: Int!, \$input: AdCampaignInput!) { adUpdateCampaign(id: \$id, input: \$input) ${selectionSet ?? '{ id user_id name description budget start_date end_date status created_at updated_at }'} }', variables: {
      'id': id,
      'input': input.toJson(),
    }
    );
    return data['adUpdateCampaign'] == null ? null : AdCampaign.fromJson((data['adUpdateCampaign'] as Map).cast<String, dynamic>());
  }

  Future<bool?> adDeleteCampaign({required int id}) async {
    final data = await _executor.mutation('mutation adDeleteCampaign(\$id: Int!) { adDeleteCampaign(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['adDeleteCampaign'] as bool?;
  }

  Future<AdCampaign?> adChangeCampaignStatus({required int id, required AdCampaignStatusInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation adChangeCampaignStatus(\$id: Int!, \$input: AdCampaignStatusInput!) { adChangeCampaignStatus(id: \$id, input: \$input) ${selectionSet ?? '{ id user_id name description budget start_date end_date status created_at updated_at }'} }', variables: {
      'id': id,
      'input': input.toJson(),
    }
    );
    return data['adChangeCampaignStatus'] == null ? null : AdCampaign.fromJson((data['adChangeCampaignStatus'] as Map).cast<String, dynamic>());
  }

  Future<AdCategory?> adCreateCategory({required AdCategoryInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation adCreateCategory(\$input: AdCategoryInput!) { adCreateCategory(input: \$input) ${selectionSet ?? '{ id name description parent_id is_active created_at }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['adCreateCategory'] == null ? null : AdCategory.fromJson((data['adCreateCategory'] as Map).cast<String, dynamic>());
  }

  Future<AdCategory?> adUpdateCategory({required int id, required AdCategoryInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation adUpdateCategory(\$id: Int!, \$input: AdCategoryInput!) { adUpdateCategory(id: \$id, input: \$input) ${selectionSet ?? '{ id name description parent_id is_active created_at }'} }', variables: {
      'id': id,
      'input': input.toJson(),
    }
    );
    return data['adUpdateCategory'] == null ? null : AdCategory.fromJson((data['adUpdateCategory'] as Map).cast<String, dynamic>());
  }

  Future<bool?> adDeleteCategory({required int id}) async {
    final data = await _executor.mutation('mutation adDeleteCategory(\$id: Int!) { adDeleteCategory(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['adDeleteCategory'] as bool?;
  }

  Future<AdPosition?> adCreatePosition({required AdPositionInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation adCreatePosition(\$input: AdPositionInput!) { adCreatePosition(input: \$input) ${selectionSet ?? '{ id name description width height max_file_size allowed_formats is_active }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['adCreatePosition'] == null ? null : AdPosition.fromJson((data['adCreatePosition'] as Map).cast<String, dynamic>());
  }

  Future<AdPosition?> adUpdatePosition({required int id, required AdPositionInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation adUpdatePosition(\$id: Int!, \$input: AdPositionInput!) { adUpdatePosition(id: \$id, input: \$input) ${selectionSet ?? '{ id name description width height max_file_size allowed_formats is_active }'} }', variables: {
      'id': id,
      'input': input.toJson(),
    }
    );
    return data['adUpdatePosition'] == null ? null : AdPosition.fromJson((data['adUpdatePosition'] as Map).cast<String, dynamic>());
  }

  Future<bool?> adDeletePosition({required int id}) async {
    final data = await _executor.mutation('mutation adDeletePosition(\$id: Int!) { adDeletePosition(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['adDeletePosition'] as bool?;
  }

  Future<AdAd?> adCreateAd({required AdAdInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation adCreateAd(\$input: AdAdInput!) { adCreateAd(input: \$input) ${selectionSet ?? '{ id campaign_id category_id title description ad_type content_url target_url file_path file_size mime_type duration priority is_active start_date end_date created_at updated_at }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['adCreateAd'] == null ? null : AdAd.fromJson((data['adCreateAd'] as Map).cast<String, dynamic>());
  }

  Future<AdAd?> adUpdateAd({required int id, required AdAdInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation adUpdateAd(\$id: Int!, \$input: AdAdInput!) { adUpdateAd(id: \$id, input: \$input) ${selectionSet ?? '{ id campaign_id category_id title description ad_type content_url target_url file_path file_size mime_type duration priority is_active start_date end_date created_at updated_at }'} }', variables: {
      'id': id,
      'input': input.toJson(),
    }
    );
    return data['adUpdateAd'] == null ? null : AdAd.fromJson((data['adUpdateAd'] as Map).cast<String, dynamic>());
  }

  Future<bool?> adDeleteAd({required int id}) async {
    final data = await _executor.mutation('mutation adDeleteAd(\$id: Int!) { adDeleteAd(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['adDeleteAd'] as bool?;
  }

  Future<AdPerformanceSummary?> adRecomputePerformance({required int id, String? date, String? selectionSet}) async {
    final data = await _executor.mutation('mutation adRecomputePerformance(\$id: Int!, \$date: String) { adRecomputePerformance(id: \$id, date: \$date) ${selectionSet ?? '{ id ad_id date impressions clicks ctr cost revenue created_at updated_at }'} }', variables: {
      'id': id,
      'date': ?date,
    }
    );
    return data['adRecomputePerformance'] == null ? null : AdPerformanceSummary.fromJson((data['adRecomputePerformance'] as Map).cast<String, dynamic>());
  }

  Future<AdPositionMapping?> adCreatePositionMapping({required AdPositionMappingInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation adCreatePositionMapping(\$input: AdPositionMappingInput!) { adCreatePositionMapping(input: \$input) ${selectionSet ?? '{ id ad_id position_id weight is_active created_at }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['adCreatePositionMapping'] == null ? null : AdPositionMapping.fromJson((data['adCreatePositionMapping'] as Map).cast<String, dynamic>());
  }

  Future<AdPositionMapping?> adUpdatePositionMapping({required int id, required AdPositionMappingInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation adUpdatePositionMapping(\$id: Int!, \$input: AdPositionMappingInput!) { adUpdatePositionMapping(id: \$id, input: \$input) ${selectionSet ?? '{ id ad_id position_id weight is_active created_at }'} }', variables: {
      'id': id,
      'input': input.toJson(),
    }
    );
    return data['adUpdatePositionMapping'] == null ? null : AdPositionMapping.fromJson((data['adUpdatePositionMapping'] as Map).cast<String, dynamic>());
  }

  Future<bool?> adDeletePositionMapping({required int id}) async {
    final data = await _executor.mutation('mutation adDeletePositionMapping(\$id: Int!) { adDeletePositionMapping(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['adDeletePositionMapping'] as bool?;
  }

  Future<AdTargetingRule?> adCreateTargetingRule({required AdTargetingRuleInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation adCreateTargetingRule(\$input: AdTargetingRuleInput!) { adCreateTargetingRule(input: \$input) ${selectionSet ?? '{ id ad_id rule_type rule_operator rule_value is_active created_at }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['adCreateTargetingRule'] == null ? null : AdTargetingRule.fromJson((data['adCreateTargetingRule'] as Map).cast<String, dynamic>());
  }

  Future<AdTargetingRule?> adUpdateTargetingRule({required int id, required AdTargetingRuleInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation adUpdateTargetingRule(\$id: Int!, \$input: AdTargetingRuleInput!) { adUpdateTargetingRule(id: \$id, input: \$input) ${selectionSet ?? '{ id ad_id rule_type rule_operator rule_value is_active created_at }'} }', variables: {
      'id': id,
      'input': input.toJson(),
    }
    );
    return data['adUpdateTargetingRule'] == null ? null : AdTargetingRule.fromJson((data['adUpdateTargetingRule'] as Map).cast<String, dynamic>());
  }

  Future<bool?> adDeleteTargetingRule({required int id}) async {
    final data = await _executor.mutation('mutation adDeleteTargetingRule(\$id: Int!) { adDeleteTargetingRule(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['adDeleteTargetingRule'] as bool?;
  }

}
