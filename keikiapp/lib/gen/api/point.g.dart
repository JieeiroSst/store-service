// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

import '../gateway_executor.dart';
import '../models.g.dart';

class PointApi {
  const PointApi(this._executor);

  final GatewayExecutor _executor;

  Future<PointAccount?> pointMyBalance({String? selectionSet}) async {
    final data = await _executor.query('query pointMyBalance { pointMyBalance ${selectionSet ?? '{ user_id balance total_earned updated_at }'} }'
    );
    return data['pointMyBalance'] == null ? null : PointAccount.fromJson((data['pointMyBalance'] as Map).cast<String, dynamic>());
  }

  Future<PointTransactionPage?> pointMyTransactions({int? limit, int? offset, String? selectionSet}) async {
    final data = await _executor.query('query pointMyTransactions(\$limit: Int, \$offset: Int) { pointMyTransactions(limit: \$limit, offset: \$offset) ${selectionSet ?? '{ items { id event_id source title points balance_after occurred_at } total limit offset }'} }', variables: {
      'limit': ?limit,
      'offset': ?offset,
    }
    );
    return data['pointMyTransactions'] == null ? null : PointTransactionPage.fromJson((data['pointMyTransactions'] as Map).cast<String, dynamic>());
  }

  Future<PointAccount?> pointUserBalance({required String userId, String? selectionSet}) async {
    final data = await _executor.query('query pointUserBalance(\$user_id: String!) { pointUserBalance(user_id: \$user_id) ${selectionSet ?? '{ user_id balance total_earned updated_at }'} }', variables: {
      'user_id': userId,
    }
    );
    return data['pointUserBalance'] == null ? null : PointAccount.fromJson((data['pointUserBalance'] as Map).cast<String, dynamic>());
  }

  Future<PointTransactionPage?> pointUserTransactions({required String userId, int? limit, int? offset, String? selectionSet}) async {
    final data = await _executor.query('query pointUserTransactions(\$user_id: String!, \$limit: Int, \$offset: Int) { pointUserTransactions(user_id: \$user_id, limit: \$limit, offset: \$offset) ${selectionSet ?? '{ items { id event_id source title points balance_after occurred_at } total limit offset }'} }', variables: {
      'user_id': userId,
      'limit': ?limit,
      'offset': ?offset,
    }
    );
    return data['pointUserTransactions'] == null ? null : PointTransactionPage.fromJson((data['pointUserTransactions'] as Map).cast<String, dynamic>());
  }

  Future<PointRewardPointPage?> pointRewardPoints({int? perPage, String? sortOrder, String? cursor, String? selectionSet}) async {
    final data = await _executor.query('query pointRewardPoints(\$per_page: Int, \$sort_order: String, \$cursor: String) { pointRewardPoints(per_page: \$per_page, sort_order: \$sort_order, cursor: \$cursor) ${selectionSet ?? '{ success data { reward_points_id total_points points_pending points_active points_expired points_converted points_cancelled activate_date expire_date created_at updated_at } pagination { next_cursor prev_cursor } }'} }', variables: {
      'per_page': ?perPage,
      'sort_order': ?sortOrder,
      'cursor': ?cursor,
    }
    );
    return data['pointRewardPoints'] == null ? null : PointRewardPointPage.fromJson((data['pointRewardPoints'] as Map).cast<String, dynamic>());
  }

  Future<PointRewardPoint?> pointRewardPoint({required String id, String? selectionSet}) async {
    final data = await _executor.query('query pointRewardPoint(\$id: String!) { pointRewardPoint(id: \$id) ${selectionSet ?? '{ reward_points_id total_points points_pending points_active points_expired points_converted points_cancelled activate_date expire_date created_at updated_at }'} }', variables: {
      'id': id,
    }
    );
    return data['pointRewardPoint'] == null ? null : PointRewardPoint.fromJson((data['pointRewardPoint'] as Map).cast<String, dynamic>());
  }

  Future<PointRewardDiscountPage?> pointRewardDiscounts({int? perPage, String? sortOrder, String? cursor, String? selectionSet}) async {
    final data = await _executor.query('query pointRewardDiscounts(\$per_page: Int, \$sort_order: String, \$cursor: String) { pointRewardDiscounts(per_page: \$per_page, sort_order: \$sort_order, cursor: \$cursor) ${selectionSet ?? '{ success data { reward_discount_id total_points points_pending points_active points_expired points_converted points_cancelled activate_date expire_date created_at updated_at } pagination { next_cursor prev_cursor } }'} }', variables: {
      'per_page': ?perPage,
      'sort_order': ?sortOrder,
      'cursor': ?cursor,
    }
    );
    return data['pointRewardDiscounts'] == null ? null : PointRewardDiscountPage.fromJson((data['pointRewardDiscounts'] as Map).cast<String, dynamic>());
  }

  Future<PointRewardDiscount?> pointRewardDiscount({required String id, String? selectionSet}) async {
    final data = await _executor.query('query pointRewardDiscount(\$id: String!) { pointRewardDiscount(id: \$id) ${selectionSet ?? '{ reward_discount_id total_points points_pending points_active points_expired points_converted points_cancelled activate_date expire_date created_at updated_at }'} }', variables: {
      'id': id,
    }
    );
    return data['pointRewardDiscount'] == null ? null : PointRewardDiscount.fromJson((data['pointRewardDiscount'] as Map).cast<String, dynamic>());
  }

  Future<PointConvertedRewardPointPage?> pointConvertedRewardPoints({int? perPage, String? sortOrder, String? cursor, String? selectionSet}) async {
    final data = await _executor.query('query pointConvertedRewardPoints(\$per_page: Int, \$sort_order: String, \$cursor: String) { pointConvertedRewardPoints(per_page: \$per_page, sort_order: \$sort_order, cursor: \$cursor) ${selectionSet ?? '{ success data { rew_convert_id rew_convert_ord_detail_id rew_convert_discount_id rew_convert_points rew_convert_date created_at updated_at } pagination { next_cursor prev_cursor } }'} }', variables: {
      'per_page': ?perPage,
      'sort_order': ?sortOrder,
      'cursor': ?cursor,
    }
    );
    return data['pointConvertedRewardPoints'] == null ? null : PointConvertedRewardPointPage.fromJson((data['pointConvertedRewardPoints'] as Map).cast<String, dynamic>());
  }

  Future<PointConvertedRewardPoint?> pointConvertedRewardPoint({required String id, String? selectionSet}) async {
    final data = await _executor.query('query pointConvertedRewardPoint(\$id: String!) { pointConvertedRewardPoint(id: \$id) ${selectionSet ?? '{ rew_convert_id rew_convert_ord_detail_id rew_convert_discount_id rew_convert_points rew_convert_date created_at updated_at }'} }', variables: {
      'id': id,
    }
    );
    return data['pointConvertedRewardPoint'] == null ? null : PointConvertedRewardPoint.fromJson((data['pointConvertedRewardPoint'] as Map).cast<String, dynamic>());
  }

  Future<String?> pointCreateRewardPoint({required PointRewardPointInput input}) async {
    final data = await _executor.mutation('mutation pointCreateRewardPoint(\$input: PointRewardPointInput!) { pointCreateRewardPoint(input: \$input) }', variables: {
      'input': input.toJson(),
    }
    );
    return data['pointCreateRewardPoint']?.toString();
  }

  Future<String?> pointUpdateRewardPoint({required PointRewardPointInput input}) async {
    final data = await _executor.mutation('mutation pointUpdateRewardPoint(\$input: PointRewardPointInput!) { pointUpdateRewardPoint(input: \$input) }', variables: {
      'input': input.toJson(),
    }
    );
    return data['pointUpdateRewardPoint']?.toString();
  }

  Future<String?> pointCreateRewardDiscount({required PointRewardDiscountInput input}) async {
    final data = await _executor.mutation('mutation pointCreateRewardDiscount(\$input: PointRewardDiscountInput!) { pointCreateRewardDiscount(input: \$input) }', variables: {
      'input': input.toJson(),
    }
    );
    return data['pointCreateRewardDiscount']?.toString();
  }

  Future<String?> pointUpdateRewardDiscount({required PointRewardDiscountInput input}) async {
    final data = await _executor.mutation('mutation pointUpdateRewardDiscount(\$input: PointRewardDiscountInput!) { pointUpdateRewardDiscount(input: \$input) }', variables: {
      'input': input.toJson(),
    }
    );
    return data['pointUpdateRewardDiscount']?.toString();
  }

  Future<String?> pointCreateConvertedRewardPoint({required PointConvertedRewardPointInput input}) async {
    final data = await _executor.mutation('mutation pointCreateConvertedRewardPoint(\$input: PointConvertedRewardPointInput!) { pointCreateConvertedRewardPoint(input: \$input) }', variables: {
      'input': input.toJson(),
    }
    );
    return data['pointCreateConvertedRewardPoint']?.toString();
  }

  Future<String?> pointUpdateConvertedRewardPoint({required PointConvertedRewardPointInput input}) async {
    final data = await _executor.mutation('mutation pointUpdateConvertedRewardPoint(\$input: PointConvertedRewardPointInput!) { pointUpdateConvertedRewardPoint(input: \$input) }', variables: {
      'input': input.toJson(),
    }
    );
    return data['pointUpdateConvertedRewardPoint']?.toString();
  }

}
