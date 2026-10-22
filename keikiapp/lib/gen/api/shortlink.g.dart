// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

import '../gateway_executor.dart';
import '../models.g.dart';

class ShortlinkApi {
  const ShortlinkApi(this._executor);

  final GatewayExecutor _executor;

  Future<List<ShortlinkLink>?> shortlinkLinks({String? userId, String? selectionSet}) async {
    final data = await _executor.query('query shortlinkLinks(\$userId: String) { shortlinkLinks(userId: \$userId) ${selectionSet ?? '{ id user_id template_id short_code original_url title description ios_app_store_url android_app_store_url web_fallback_url app_scheme ios_universal_link android_app_link deep_link_path deep_link_parameters utm_parameters { source medium campaign term content } targeting_rules { countries devices languages } og_title og_description og_image_url og_type attribution_window_hours is_active expires_at append_click_id created_at updated_at organization_id warn_at disabled_at disabled_reason click_count template_slug }'} }', variables: {
      'userId': ?userId,
    }
    );
    return (data['shortlinkLinks'] as List?)?.map((e) => (e == null ? null : ShortlinkLink.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<ShortlinkLink?> shortlinkLink({required String id, String? userId, String? selectionSet}) async {
    final data = await _executor.query('query shortlinkLink(\$id: String!, \$userId: String) { shortlinkLink(id: \$id, userId: \$userId) ${selectionSet ?? '{ id user_id template_id short_code original_url title description ios_app_store_url android_app_store_url web_fallback_url app_scheme ios_universal_link android_app_link deep_link_path deep_link_parameters utm_parameters { source medium campaign term content } targeting_rules { countries devices languages } og_title og_description og_image_url og_type attribution_window_hours is_active expires_at append_click_id created_at updated_at organization_id warn_at disabled_at disabled_reason click_count template_slug }'} }', variables: {
      'id': id,
      'userId': ?userId,
    }
    );
    return data['shortlinkLink'] == null ? null : ShortlinkLink.fromJson((data['shortlinkLink'] as Map).cast<String, dynamic>());
  }

  Future<ShortlinkAnalytics?> shortlinkAnalyticsOverview({String? userId, int? days, String? selectionSet}) async {
    final data = await _executor.query('query shortlinkAnalyticsOverview(\$userId: String, \$days: Int) { shortlinkAnalyticsOverview(userId: \$userId, days: \$days) ${selectionSet ?? '{ totalClicks uniqueClicks clicksByDate { date clicks } clicksByCountry { countryCode country clicks } clicksByDevice { device clicks } clicksByPlatform { platform clicks } topLinks { id shortCode title originalUrl totalClicks uniqueClicks } }'} }', variables: {
      'userId': ?userId,
      'days': ?days,
    }
    );
    return data['shortlinkAnalyticsOverview'] == null ? null : ShortlinkAnalytics.fromJson((data['shortlinkAnalyticsOverview'] as Map).cast<String, dynamic>());
  }

  Future<ShortlinkAnalytics?> shortlinkLinkAnalytics({required String linkId, String? userId, int? days, String? selectionSet}) async {
    final data = await _executor.query('query shortlinkLinkAnalytics(\$linkId: String!, \$userId: String, \$days: Int) { shortlinkLinkAnalytics(linkId: \$linkId, userId: \$userId, days: \$days) ${selectionSet ?? '{ totalClicks uniqueClicks clicksByDate { date clicks } clicksByCountry { countryCode country clicks } clicksByDevice { device clicks } clicksByPlatform { platform clicks } topLinks { id shortCode title originalUrl totalClicks uniqueClicks } }'} }', variables: {
      'linkId': linkId,
      'userId': ?userId,
      'days': ?days,
    }
    );
    return data['shortlinkLinkAnalytics'] == null ? null : ShortlinkAnalytics.fromJson((data['shortlinkLinkAnalytics'] as Map).cast<String, dynamic>());
  }

  Future<ShortlinkResolve?> shortlinkResolve({required String shortCode, String? selectionSet}) async {
    final data = await _executor.public('query shortlinkResolve(\$shortCode: String!) { shortlinkResolve(shortCode: \$shortCode) ${selectionSet ?? '{ shortCode linkId deepLinkPath appScheme iosUrl androidUrl webUrl utmParameters { source medium campaign term content } customParameters clickedAt }'} }', variables: {
      'shortCode': shortCode,
    }
    );
    return data['shortlinkResolve'] == null ? null : ShortlinkResolve.fromJson((data['shortlinkResolve'] as Map).cast<String, dynamic>());
  }

  Future<ShortlinkResolve?> shortlinkResolveWithTemplate({required String templateSlug, required String shortCode, String? selectionSet}) async {
    final data = await _executor.public('query shortlinkResolveWithTemplate(\$templateSlug: String!, \$shortCode: String!) { shortlinkResolveWithTemplate(templateSlug: \$templateSlug, shortCode: \$shortCode) ${selectionSet ?? '{ shortCode linkId deepLinkPath appScheme iosUrl androidUrl webUrl utmParameters { source medium campaign term content } customParameters clickedAt }'} }', variables: {
      'templateSlug': templateSlug,
      'shortCode': shortCode,
    }
    );
    return data['shortlinkResolveWithTemplate'] == null ? null : ShortlinkResolve.fromJson((data['shortlinkResolveWithTemplate'] as Map).cast<String, dynamic>());
  }

  Future<ShortlinkAttribution?> shortlinkAttribution({required String fingerprint, String? selectionSet}) async {
    final data = await _executor.public('query shortlinkAttribution(\$fingerprint: String!) { shortlinkAttribution(fingerprint: \$fingerprint) ${selectionSet ?? '{ fingerprint attributed installEvent { id installedAt firstOpenAt confidenceScore deepLinkRetrieved } clickEvent { id clickedAt deviceType platform countryCode city } linkData { shortCode originalUrl iosUrl androidUrl webFallbackUrl utmParameters { source medium campaign term content } deepLinkParameters } }'} }', variables: {
      'fingerprint': fingerprint,
    }
    );
    return data['shortlinkAttribution'] == null ? null : ShortlinkAttribution.fromJson((data['shortlinkAttribution'] as Map).cast<String, dynamic>());
  }

  Future<ShortlinkSdkHealth?> shortlinkSdkHealth({String? selectionSet}) async {
    final data = await _executor.public('query shortlinkSdkHealth { shortlinkSdkHealth ${selectionSet ?? '{ status version timestamp }'} }'
    );
    return data['shortlinkSdkHealth'] == null ? null : ShortlinkSdkHealth.fromJson((data['shortlinkSdkHealth'] as Map).cast<String, dynamic>());
  }

  Future<List<ShortlinkWebhook>?> shortlinkWebhooks({String? userId, String? selectionSet}) async {
    final data = await _executor.query('query shortlinkWebhooks(\$userId: String) { shortlinkWebhooks(userId: \$userId) ${selectionSet ?? '{ id user_id name url events is_active retry_count timeout_ms headers created_at updated_at }'} }', variables: {
      'userId': ?userId,
    }
    );
    return (data['shortlinkWebhooks'] as List?)?.map((e) => (e == null ? null : ShortlinkWebhook.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<ShortlinkWebhook?> shortlinkWebhook({required String id, String? userId, String? selectionSet}) async {
    final data = await _executor.query('query shortlinkWebhook(\$id: String!, \$userId: String) { shortlinkWebhook(id: \$id, userId: \$userId) ${selectionSet ?? '{ id user_id name url events is_active retry_count timeout_ms headers created_at updated_at }'} }', variables: {
      'id': id,
      'userId': ?userId,
    }
    );
    return data['shortlinkWebhook'] == null ? null : ShortlinkWebhook.fromJson((data['shortlinkWebhook'] as Map).cast<String, dynamic>());
  }

  Future<List<ShortlinkTemplate>?> shortlinkTemplates({String? userId, String? selectionSet}) async {
    final data = await _executor.query('query shortlinkTemplates(\$userId: String) { shortlinkTemplates(userId: \$userId) ${selectionSet ?? '{ id user_id name slug description settings is_default created_at updated_at }'} }', variables: {
      'userId': ?userId,
    }
    );
    return (data['shortlinkTemplates'] as List?)?.map((e) => (e == null ? null : ShortlinkTemplate.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<ShortlinkTemplate?> shortlinkTemplate({required String id, String? userId, String? selectionSet}) async {
    final data = await _executor.query('query shortlinkTemplate(\$id: String!, \$userId: String) { shortlinkTemplate(id: \$id, userId: \$userId) ${selectionSet ?? '{ id user_id name slug description settings is_default created_at updated_at }'} }', variables: {
      'id': id,
      'userId': ?userId,
    }
    );
    return data['shortlinkTemplate'] == null ? null : ShortlinkTemplate.fromJson((data['shortlinkTemplate'] as Map).cast<String, dynamic>());
  }

  Future<Map<String, dynamic>?> shortlinkAppleAppSiteAssociation() async {
    final data = await _executor.public('query shortlinkAppleAppSiteAssociation { shortlinkAppleAppSiteAssociation }'
    );
    return (data['shortlinkAppleAppSiteAssociation'] as Map?)?.cast<String, dynamic>();
  }

  Future<List<Map<String, dynamic>>?> shortlinkAssetLinks() async {
    final data = await _executor.public('query shortlinkAssetLinks { shortlinkAssetLinks }'
    );
    return (data['shortlinkAssetLinks'] as List?)?.map((e) => ((e as Map?)?.cast<String, dynamic>())!).toList();
  }

  Future<ShortlinkLink?> shortlinkCreateLink({required ShortlinkCreateLinkInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation shortlinkCreateLink(\$input: ShortlinkCreateLinkInput!) { shortlinkCreateLink(input: \$input) ${selectionSet ?? '{ id user_id template_id short_code original_url title description ios_app_store_url android_app_store_url web_fallback_url app_scheme ios_universal_link android_app_link deep_link_path deep_link_parameters utm_parameters { source medium campaign term content } targeting_rules { countries devices languages } og_title og_description og_image_url og_type attribution_window_hours is_active expires_at append_click_id created_at updated_at organization_id warn_at disabled_at disabled_reason click_count template_slug }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['shortlinkCreateLink'] == null ? null : ShortlinkLink.fromJson((data['shortlinkCreateLink'] as Map).cast<String, dynamic>());
  }

  Future<ShortlinkLink?> shortlinkUpdateLink({required String id, required ShortlinkUpdateLinkInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation shortlinkUpdateLink(\$id: String!, \$input: ShortlinkUpdateLinkInput!) { shortlinkUpdateLink(id: \$id, input: \$input) ${selectionSet ?? '{ id user_id template_id short_code original_url title description ios_app_store_url android_app_store_url web_fallback_url app_scheme ios_universal_link android_app_link deep_link_path deep_link_parameters utm_parameters { source medium campaign term content } targeting_rules { countries devices languages } og_title og_description og_image_url og_type attribution_window_hours is_active expires_at append_click_id created_at updated_at organization_id warn_at disabled_at disabled_reason click_count template_slug }'} }', variables: {
      'id': id,
      'input': input.toJson(),
    }
    );
    return data['shortlinkUpdateLink'] == null ? null : ShortlinkLink.fromJson((data['shortlinkUpdateLink'] as Map).cast<String, dynamic>());
  }

  Future<bool?> shortlinkDeleteLink({required String id}) async {
    final data = await _executor.mutation('mutation shortlinkDeleteLink(\$id: String!) { shortlinkDeleteLink(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['shortlinkDeleteLink'] as bool?;
  }

  Future<ShortlinkLink?> shortlinkDuplicateLink({required String id, String? selectionSet}) async {
    final data = await _executor.mutation('mutation shortlinkDuplicateLink(\$id: String!) { shortlinkDuplicateLink(id: \$id) ${selectionSet ?? '{ id user_id template_id short_code original_url title description ios_app_store_url android_app_store_url web_fallback_url app_scheme ios_universal_link android_app_link deep_link_path deep_link_parameters utm_parameters { source medium campaign term content } targeting_rules { countries devices languages } og_title og_description og_image_url og_type attribution_window_hours is_active expires_at append_click_id created_at updated_at organization_id warn_at disabled_at disabled_reason click_count template_slug }'} }', variables: {
      'id': id,
    }
    );
    return data['shortlinkDuplicateLink'] == null ? null : ShortlinkLink.fromJson((data['shortlinkDuplicateLink'] as Map).cast<String, dynamic>());
  }

  Future<ShortlinkAttribution?> shortlinkSdkInstall({required ShortlinkInstallInput input, String? selectionSet}) async {
    final data = await _executor.public('mutation shortlinkSdkInstall(\$input: ShortlinkInstallInput!) { shortlinkSdkInstall(input: \$input) ${selectionSet ?? '{ fingerprint attributed installEvent { id installedAt firstOpenAt confidenceScore deepLinkRetrieved } clickEvent { id clickedAt deviceType platform countryCode city } linkData { shortCode originalUrl iosUrl androidUrl webFallbackUrl utmParameters { source medium campaign term content } deepLinkParameters } }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['shortlinkSdkInstall'] == null ? null : ShortlinkAttribution.fromJson((data['shortlinkSdkInstall'] as Map).cast<String, dynamic>());
  }

  Future<Object?> shortlinkSdkEvent({required ShortlinkEventInput input}) async {
    final data = await _executor.public('mutation shortlinkSdkEvent(\$input: ShortlinkEventInput!) { shortlinkSdkEvent(input: \$input) }', variables: {
      'input': input.toJson(),
    }
    );
    return data['shortlinkSdkEvent'];
  }

  Future<ShortlinkWebhook?> shortlinkCreateWebhook({required ShortlinkWebhookInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation shortlinkCreateWebhook(\$input: ShortlinkWebhookInput!) { shortlinkCreateWebhook(input: \$input) ${selectionSet ?? '{ id user_id name url events is_active retry_count timeout_ms headers created_at updated_at }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['shortlinkCreateWebhook'] == null ? null : ShortlinkWebhook.fromJson((data['shortlinkCreateWebhook'] as Map).cast<String, dynamic>());
  }

  Future<ShortlinkWebhook?> shortlinkUpdateWebhook({required String id, required ShortlinkWebhookInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation shortlinkUpdateWebhook(\$id: String!, \$input: ShortlinkWebhookInput!) { shortlinkUpdateWebhook(id: \$id, input: \$input) ${selectionSet ?? '{ id user_id name url events is_active retry_count timeout_ms headers created_at updated_at }'} }', variables: {
      'id': id,
      'input': input.toJson(),
    }
    );
    return data['shortlinkUpdateWebhook'] == null ? null : ShortlinkWebhook.fromJson((data['shortlinkUpdateWebhook'] as Map).cast<String, dynamic>());
  }

  Future<bool?> shortlinkDeleteWebhook({required String id}) async {
    final data = await _executor.mutation('mutation shortlinkDeleteWebhook(\$id: String!) { shortlinkDeleteWebhook(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['shortlinkDeleteWebhook'] as bool?;
  }

  Future<Object?> shortlinkTestWebhook({required String id}) async {
    final data = await _executor.mutation('mutation shortlinkTestWebhook(\$id: String!) { shortlinkTestWebhook(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['shortlinkTestWebhook'];
  }

  Future<ShortlinkTemplate?> shortlinkCreateTemplate({required ShortlinkTemplateInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation shortlinkCreateTemplate(\$input: ShortlinkTemplateInput!) { shortlinkCreateTemplate(input: \$input) ${selectionSet ?? '{ id user_id name slug description settings is_default created_at updated_at }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['shortlinkCreateTemplate'] == null ? null : ShortlinkTemplate.fromJson((data['shortlinkCreateTemplate'] as Map).cast<String, dynamic>());
  }

  Future<ShortlinkTemplate?> shortlinkUpdateTemplate({required String id, required ShortlinkTemplateInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation shortlinkUpdateTemplate(\$id: String!, \$input: ShortlinkTemplateInput!) { shortlinkUpdateTemplate(id: \$id, input: \$input) ${selectionSet ?? '{ id user_id name slug description settings is_default created_at updated_at }'} }', variables: {
      'id': id,
      'input': input.toJson(),
    }
    );
    return data['shortlinkUpdateTemplate'] == null ? null : ShortlinkTemplate.fromJson((data['shortlinkUpdateTemplate'] as Map).cast<String, dynamic>());
  }

  Future<bool?> shortlinkDeleteTemplate({required String id}) async {
    final data = await _executor.mutation('mutation shortlinkDeleteTemplate(\$id: String!) { shortlinkDeleteTemplate(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['shortlinkDeleteTemplate'] as bool?;
  }

  Future<Object?> shortlinkSetDefaultTemplate({required String id}) async {
    final data = await _executor.mutation('mutation shortlinkSetDefaultTemplate(\$id: String!) { shortlinkSetDefaultTemplate(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['shortlinkSetDefaultTemplate'];
  }

}
