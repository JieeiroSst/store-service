// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

import '../gateway_executor.dart';
import '../models.g.dart';

class BotApi {
  const BotApi(this._executor);

  final GatewayExecutor _executor;

  Future<BotPostList?> botPosts({int? limit, int? offset, String? status, String? campaign, String? selectionSet}) async {
    final data = await _executor.query('query botPosts(\$limit: Int, \$offset: Int, \$status: String, \$campaign: String) { botPosts(limit: \$limit, offset: \$offset, status: \$status, campaign: \$campaign) ${selectionSet ?? '{ posts { id title text hashtags media { url type thumbnail } media_kind channels scheduled_at timezone cron_expr max_runs_per_day max_runs_per_month runs_today runs_this_month last_run_date last_run_month last_run_at status reject_reason results { channel external_id published_url error published_at } campaign created_by approved_by status_changed_by created_at updated_at } }'} }', variables: {
      'limit': ?limit,
      'offset': ?offset,
      'status': ?status,
      'campaign': ?campaign,
    }
    );
    return data['botPosts'] == null ? null : BotPostList.fromJson((data['botPosts'] as Map).cast<String, dynamic>());
  }

  Future<BotPost?> botPost({required String id, String? selectionSet}) async {
    final data = await _executor.query('query botPost(\$id: String!) { botPost(id: \$id) ${selectionSet ?? '{ id title text hashtags media { url type thumbnail } media_kind channels scheduled_at timezone cron_expr max_runs_per_day max_runs_per_month runs_today runs_this_month last_run_date last_run_month last_run_at status reject_reason results { channel external_id published_url error published_at } campaign created_by approved_by status_changed_by created_at updated_at }'} }', variables: {
      'id': id,
    }
    );
    return data['botPost'] == null ? null : BotPost.fromJson((data['botPost'] as Map).cast<String, dynamic>());
  }

  Future<BotPost?> botCreatePost({required BotCreatePostInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation botCreatePost(\$input: BotCreatePostInput!) { botCreatePost(input: \$input) ${selectionSet ?? '{ id title text hashtags media { url type thumbnail } media_kind channels scheduled_at timezone cron_expr max_runs_per_day max_runs_per_month runs_today runs_this_month last_run_date last_run_month last_run_at status reject_reason results { channel external_id published_url error published_at } campaign created_by approved_by status_changed_by created_at updated_at }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['botCreatePost'] == null ? null : BotPost.fromJson((data['botCreatePost'] as Map).cast<String, dynamic>());
  }

  Future<BotPost?> botUpdatePost({required String id, required BotUpdatePostInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation botUpdatePost(\$id: String!, \$input: BotUpdatePostInput!) { botUpdatePost(id: \$id, input: \$input) ${selectionSet ?? '{ id title text hashtags media { url type thumbnail } media_kind channels scheduled_at timezone cron_expr max_runs_per_day max_runs_per_month runs_today runs_this_month last_run_date last_run_month last_run_at status reject_reason results { channel external_id published_url error published_at } campaign created_by approved_by status_changed_by created_at updated_at }'} }', variables: {
      'id': id,
      'input': input.toJson(),
    }
    );
    return data['botUpdatePost'] == null ? null : BotPost.fromJson((data['botUpdatePost'] as Map).cast<String, dynamic>());
  }

  Future<BotPost?> botSubmitPost({required String id, String? selectionSet}) async {
    final data = await _executor.mutation('mutation botSubmitPost(\$id: String!) { botSubmitPost(id: \$id) ${selectionSet ?? '{ id title text hashtags media { url type thumbnail } media_kind channels scheduled_at timezone cron_expr max_runs_per_day max_runs_per_month runs_today runs_this_month last_run_date last_run_month last_run_at status reject_reason results { channel external_id published_url error published_at } campaign created_by approved_by status_changed_by created_at updated_at }'} }', variables: {
      'id': id,
    }
    );
    return data['botSubmitPost'] == null ? null : BotPost.fromJson((data['botSubmitPost'] as Map).cast<String, dynamic>());
  }

  Future<BotPost?> botApprovePost({required String id, String? selectionSet}) async {
    final data = await _executor.mutation('mutation botApprovePost(\$id: String!) { botApprovePost(id: \$id) ${selectionSet ?? '{ id title text hashtags media { url type thumbnail } media_kind channels scheduled_at timezone cron_expr max_runs_per_day max_runs_per_month runs_today runs_this_month last_run_date last_run_month last_run_at status reject_reason results { channel external_id published_url error published_at } campaign created_by approved_by status_changed_by created_at updated_at }'} }', variables: {
      'id': id,
    }
    );
    return data['botApprovePost'] == null ? null : BotPost.fromJson((data['botApprovePost'] as Map).cast<String, dynamic>());
  }

  Future<BotPost?> botPublishPost({required String id, String? selectionSet}) async {
    final data = await _executor.mutation('mutation botPublishPost(\$id: String!) { botPublishPost(id: \$id) ${selectionSet ?? '{ id title text hashtags media { url type thumbnail } media_kind channels scheduled_at timezone cron_expr max_runs_per_day max_runs_per_month runs_today runs_this_month last_run_date last_run_month last_run_at status reject_reason results { channel external_id published_url error published_at } campaign created_by approved_by status_changed_by created_at updated_at }'} }', variables: {
      'id': id,
    }
    );
    return data['botPublishPost'] == null ? null : BotPost.fromJson((data['botPublishPost'] as Map).cast<String, dynamic>());
  }

  Future<BotPost?> botCancelPost({required String id, String? selectionSet}) async {
    final data = await _executor.mutation('mutation botCancelPost(\$id: String!) { botCancelPost(id: \$id) ${selectionSet ?? '{ id title text hashtags media { url type thumbnail } media_kind channels scheduled_at timezone cron_expr max_runs_per_day max_runs_per_month runs_today runs_this_month last_run_date last_run_month last_run_at status reject_reason results { channel external_id published_url error published_at } campaign created_by approved_by status_changed_by created_at updated_at }'} }', variables: {
      'id': id,
    }
    );
    return data['botCancelPost'] == null ? null : BotPost.fromJson((data['botCancelPost'] as Map).cast<String, dynamic>());
  }

  Future<BotPost?> botRejectPost({required String id, required String reason, String? selectionSet}) async {
    final data = await _executor.mutation('mutation botRejectPost(\$id: String!, \$reason: String!) { botRejectPost(id: \$id, reason: \$reason) ${selectionSet ?? '{ id title text hashtags media { url type thumbnail } media_kind channels scheduled_at timezone cron_expr max_runs_per_day max_runs_per_month runs_today runs_this_month last_run_date last_run_month last_run_at status reject_reason results { channel external_id published_url error published_at } campaign created_by approved_by status_changed_by created_at updated_at }'} }', variables: {
      'id': id,
      'reason': reason,
    }
    );
    return data['botRejectPost'] == null ? null : BotPost.fromJson((data['botRejectPost'] as Map).cast<String, dynamic>());
  }

}
