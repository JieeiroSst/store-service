// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

import '../gateway_executor.dart';
import '../models.g.dart';

class PolymarketApi {
  const PolymarketApi(this._executor);

  final GatewayExecutor _executor;

  Future<List<PolyCategoryCount>?> polyCategories({String? selectionSet}) async {
    final data = await _executor.query('query polyCategories { polyCategories ${selectionSet ?? '{ category events }'} }'
    );
    return (data['polyCategories'] as List?)?.map((e) => (e == null ? null : PolyCategoryCount.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<List<PolyLeaderboardEntry>?> polyLeaderboard({String? metric, String? window, int? limit, String? selectionSet}) async {
    final data = await _executor.query('query polyLeaderboard(\$metric: String, \$window: String, \$limit: Int) { polyLeaderboard(metric: \$metric, window: \$window, limit: \$limit) ${selectionSet ?? '{ rank user_id username profit volume }'} }', variables: {
      'metric': ?metric,
      'window': ?window,
      'limit': ?limit,
    }
    );
    return (data['polyLeaderboard'] as List?)?.map((e) => (e == null ? null : PolyLeaderboardEntry.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<PolyEventPage?> polyEvents({String? status, String? category, String? tag, String? q, bool? featured, String? sort, int? limit, String? cursor, String? selectionSet}) async {
    final data = await _executor.query('query polyEvents(\$status: String, \$category: String, \$tag: String, \$q: String, \$featured: Boolean, \$sort: String, \$limit: Int, \$cursor: String) { polyEvents(status: \$status, category: \$category, tag: \$tag, q: \$q, featured: \$featured, sort: \$sort, limit: \$limit, cursor: \$cursor) ${selectionSet ?? '{ items { markets { yes_price no_price id event_id slug question group_item_title description share_value min_order_size end_time status proposed_outcome dispute_deadline disputed_by dispute_reason resolved_outcome reward_pool reward_max_spread reward_min_size dispute_bond best_bid best_ask last_price volume created_at updated_at } id slug title description category tags image_url featured neg_risk status end_date volume created_at updated_at } next_cursor is_last_page }'} }', variables: {
      'status': ?status,
      'category': ?category,
      'tag': ?tag,
      'q': ?q,
      'featured': ?featured,
      'sort': ?sort,
      'limit': ?limit,
      'cursor': ?cursor,
    }
    );
    return data['polyEvents'] == null ? null : PolyEventPage.fromJson((data['polyEvents'] as Map).cast<String, dynamic>());
  }

  Future<PolyEvent?> polyEvent({required String id, String? selectionSet}) async {
    final data = await _executor.query('query polyEvent(\$id: String!) { polyEvent(id: \$id) ${selectionSet ?? '{ markets { yes_price no_price id event_id slug question group_item_title description share_value min_order_size end_time status proposed_outcome dispute_deadline disputed_by dispute_reason resolved_outcome reward_pool reward_max_spread reward_min_size dispute_bond best_bid best_ask last_price volume created_at updated_at } id slug title description category tags image_url featured neg_risk status end_date volume created_at updated_at }'} }', variables: {
      'id': id,
    }
    );
    return data['polyEvent'] == null ? null : PolyEvent.fromJson((data['polyEvent'] as Map).cast<String, dynamic>());
  }

  Future<List<PolyEvent>?> polyRelatedEvents({required String id, int? limit, String? selectionSet}) async {
    final data = await _executor.query('query polyRelatedEvents(\$id: String!, \$limit: Int) { polyRelatedEvents(id: \$id, limit: \$limit) ${selectionSet ?? '{ markets { yes_price no_price id event_id slug question group_item_title description share_value min_order_size end_time status proposed_outcome dispute_deadline disputed_by dispute_reason resolved_outcome reward_pool reward_max_spread reward_min_size dispute_bond best_bid best_ask last_price volume created_at updated_at } id slug title description category tags image_url featured neg_risk status end_date volume created_at updated_at }'} }', variables: {
      'id': id,
      'limit': ?limit,
    }
    );
    return (data['polyRelatedEvents'] as List?)?.map((e) => (e == null ? null : PolyEvent.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<PolyCommentPage?> polyComments({required String id, String? sort, int? limit, String? cursor, String? selectionSet}) async {
    final data = await _executor.query('query polyComments(\$id: String!, \$sort: String, \$limit: Int, \$cursor: String) { polyComments(id: \$id, sort: \$sort, limit: \$limit, cursor: \$cursor) ${selectionSet ?? '{ items { id event_id parent_id user_id body created_at likes } next_cursor is_last_page }'} }', variables: {
      'id': id,
      'sort': ?sort,
      'limit': ?limit,
      'cursor': ?cursor,
    }
    );
    return data['polyComments'] == null ? null : PolyCommentPage.fromJson((data['polyComments'] as Map).cast<String, dynamic>());
  }

  Future<PolyMarketSearch?> polyMarkets({String? q, String? status, int? limit, int? offset, String? selectionSet}) async {
    final data = await _executor.query('query polyMarkets(\$q: String, \$status: String, \$limit: Int, \$offset: Int) { polyMarkets(q: \$q, status: \$status, limit: \$limit, offset: \$offset) ${selectionSet ?? '{ items { yes_price no_price id event_id slug question group_item_title description share_value min_order_size end_time status proposed_outcome dispute_deadline disputed_by dispute_reason resolved_outcome reward_pool reward_max_spread reward_min_size dispute_bond best_bid best_ask last_price volume created_at updated_at } total }'} }', variables: {
      'q': ?q,
      'status': ?status,
      'limit': ?limit,
      'offset': ?offset,
    }
    );
    return data['polyMarkets'] == null ? null : PolyMarketSearch.fromJson((data['polyMarkets'] as Map).cast<String, dynamic>());
  }

  Future<PolyMarket?> polyMarket({required String id, String? selectionSet}) async {
    final data = await _executor.query('query polyMarket(\$id: String!) { polyMarket(id: \$id) ${selectionSet ?? '{ yes_price no_price id event_id slug question group_item_title description share_value min_order_size end_time status proposed_outcome dispute_deadline disputed_by dispute_reason resolved_outcome reward_pool reward_max_spread reward_min_size dispute_bond best_bid best_ask last_price volume created_at updated_at }'} }', variables: {
      'id': id,
    }
    );
    return data['polyMarket'] == null ? null : PolyMarket.fromJson((data['polyMarket'] as Map).cast<String, dynamic>());
  }

  Future<PolyOrderBook?> polyOrderBook({required String id, String? outcome, String? selectionSet}) async {
    final data = await _executor.query('query polyOrderBook(\$id: String!, \$outcome: String) { polyOrderBook(id: \$id, outcome: \$outcome) ${selectionSet ?? '{ market_id outcome bids { price size } asks { price size } best_bid best_ask spread midpoint last_trade_price }'} }', variables: {
      'id': id,
      'outcome': ?outcome,
    }
    );
    return data['polyOrderBook'] == null ? null : PolyOrderBook.fromJson((data['polyOrderBook'] as Map).cast<String, dynamic>());
  }

  Future<PolyQuote?> polyQuote({required String id, String? side, String? outcome, int? amount, String? selectionSet}) async {
    final data = await _executor.query('query polyQuote(\$id: String!, \$side: String, \$outcome: String, \$amount: Int) { polyQuote(id: \$id, side: \$side, outcome: \$outcome, amount: \$amount) ${selectionSet ?? '{ shares cash fee avg_price worst_price fillable }'} }', variables: {
      'id': id,
      'side': ?side,
      'outcome': ?outcome,
      'amount': ?amount,
    }
    );
    return data['polyQuote'] == null ? null : PolyQuote.fromJson((data['polyQuote'] as Map).cast<String, dynamic>());
  }

  Future<List<PolyTrade>?> polyMarketTrades({required String id, int? limit, int? offset, String? selectionSet}) async {
    final data = await _executor.query('query polyMarketTrades(\$id: String!, \$limit: Int, \$offset: Int) { polyMarketTrades(id: \$id, limit: \$limit, offset: \$offset) ${selectionSet ?? '{ id market_id yes_price size kind maker_order_id taker_order_id maker_user_id taker_user_id maker_outcome maker_side taker_outcome taker_side taker_fee maker_rebate referral_fee created_at }'} }', variables: {
      'id': id,
      'limit': ?limit,
      'offset': ?offset,
    }
    );
    return (data['polyMarketTrades'] as List?)?.map((e) => (e == null ? null : PolyTrade.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<List<PolyCandle>?> polyPriceHistory({required String id, String? range, int? bucket, String? selectionSet}) async {
    final data = await _executor.query('query polyPriceHistory(\$id: String!, \$range: String, \$bucket: Int) { polyPriceHistory(id: \$id, range: \$range, bucket: \$bucket) ${selectionSet ?? '{ time open high low close volume }'} }', variables: {
      'id': id,
      'range': ?range,
      'bucket': ?bucket,
    }
    );
    return (data['polyPriceHistory'] as List?)?.map((e) => (e == null ? null : PolyCandle.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<List<PolyHolderEntry>?> polyTopHolders({required String id, String? outcome, int? limit, String? selectionSet}) async {
    final data = await _executor.query('query polyTopHolders(\$id: String!, \$outcome: String, \$limit: Int) { polyTopHolders(id: \$id, outcome: \$outcome, limit: \$limit) ${selectionSet ?? '{ user_id username shares }'} }', variables: {
      'id': id,
      'outcome': ?outcome,
      'limit': ?limit,
    }
    );
    return (data['polyTopHolders'] as List?)?.map((e) => (e == null ? null : PolyHolderEntry.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<PolyMarketRewards?> polyMarketRewards({required String id, String? selectionSet}) async {
    final data = await _executor.query('query polyMarketRewards(\$id: String!) { polyMarketRewards(id: \$id) ${selectionSet ?? '{ market_id daily_pool max_spread min_size epoch_seconds epoch_pool }'} }', variables: {
      'id': id,
    }
    );
    return data['polyMarketRewards'] == null ? null : PolyMarketRewards.fromJson((data['polyMarketRewards'] as Map).cast<String, dynamic>());
  }

  Future<PolyExchangeSummary?> polyExchange({String? selectionSet}) async {
    final data = await _executor.query('query polyExchange { polyExchange ${selectionSet ?? '{ revenue bond_held }'} }'
    );
    return data['polyExchange'] == null ? null : PolyExchangeSummary.fromJson((data['polyExchange'] as Map).cast<String, dynamic>());
  }

  Future<PolyOrder?> polyOrder({required int orderId, String? selectionSet}) async {
    final data = await _executor.query('query polyOrder(\$orderId: Int!) { polyOrder(orderId: \$orderId) ${selectionSet ?? '{ id market_id user_id client_order_id outcome side type time_in_force price size filled budget fee filled_cash status expires_at created_at updated_at }'} }', variables: {
      'orderId': orderId,
    }
    );
    return data['polyOrder'] == null ? null : PolyOrder.fromJson((data['polyOrder'] as Map).cast<String, dynamic>());
  }

  Future<PolyProfile?> polyProfile({required String userId, String? selectionSet}) async {
    final data = await _executor.query('query polyProfile(\$userId: String!) { polyProfile(userId: \$userId) ${selectionSet ?? '{ user_id username bio avatar_url created_at updated_at }'} }', variables: {
      'userId': userId,
    }
    );
    return data['polyProfile'] == null ? null : PolyProfile.fromJson((data['polyProfile'] as Map).cast<String, dynamic>());
  }

  Future<PolyBalance?> polyBalance({required String userId, String? selectionSet}) async {
    final data = await _executor.query('query polyBalance(\$userId: String!) { polyBalance(userId: \$userId) ${selectionSet ?? '{ user_id available locked updated_at }'} }', variables: {
      'userId': userId,
    }
    );
    return data['polyBalance'] == null ? null : PolyBalance.fromJson((data['polyBalance'] as Map).cast<String, dynamic>());
  }

  Future<PolyWallet?> polyWallet({required String userId, String? selectionSet}) async {
    final data = await _executor.query('query polyWallet(\$userId: String!) { polyWallet(userId: \$userId) ${selectionSet ?? '{ wallet_id balance currency status }'} }', variables: {
      'userId': userId,
    }
    );
    return data['polyWallet'] == null ? null : PolyWallet.fromJson((data['polyWallet'] as Map).cast<String, dynamic>());
  }

  Future<PolyPortfolio?> polyPortfolio({required String userId, String? selectionSet}) async {
    final data = await _executor.query('query polyPortfolio(\$userId: String!) { polyPortfolio(userId: \$userId) ${selectionSet ?? '{ balance { user_id available locked updated_at } positions { question event_id avg_price current_price value unrealized_pnl market_status id market_id user_id outcome shares locked_shares cost_basis realized_pnl settled created_at updated_at } positions_value unrealized_pnl realized_pnl total_value }'} }', variables: {
      'userId': userId,
    }
    );
    return data['polyPortfolio'] == null ? null : PolyPortfolio.fromJson((data['polyPortfolio'] as Map).cast<String, dynamic>());
  }

  Future<List<PolyPosition>?> polyClosedPositions({required String userId, String? selectionSet}) async {
    final data = await _executor.query('query polyClosedPositions(\$userId: String!) { polyClosedPositions(userId: \$userId) ${selectionSet ?? '{ id market_id user_id outcome shares locked_shares cost_basis realized_pnl settled created_at updated_at }'} }', variables: {
      'userId': userId,
    }
    );
    return (data['polyClosedPositions'] as List?)?.map((e) => (e == null ? null : PolyPosition.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<List<PolyActivity>?> polyActivity({required String userId, int? limit, String? selectionSet}) async {
    final data = await _executor.query('query polyActivity(\$userId: String!, \$limit: Int) { polyActivity(userId: \$userId, limit: \$limit) ${selectionSet ?? '{ type market_id outcome side price size amount created_at }'} }', variables: {
      'userId': userId,
      'limit': ?limit,
    }
    );
    return (data['polyActivity'] as List?)?.map((e) => (e == null ? null : PolyActivity.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<PolyOrderPage?> polyUserOrders({required String userId, int? marketId, String? status, int? limit, String? cursor, String? selectionSet}) async {
    final data = await _executor.query('query polyUserOrders(\$userId: String!, \$market_id: Int, \$status: String, \$limit: Int, \$cursor: String) { polyUserOrders(userId: \$userId, market_id: \$market_id, status: \$status, limit: \$limit, cursor: \$cursor) ${selectionSet ?? '{ items { id market_id user_id client_order_id outcome side type time_in_force price size filled budget fee filled_cash status expires_at created_at updated_at } next_cursor is_last_page }'} }', variables: {
      'userId': userId,
      'market_id': ?marketId,
      'status': ?status,
      'limit': ?limit,
      'cursor': ?cursor,
    }
    );
    return data['polyUserOrders'] == null ? null : PolyOrderPage.fromJson((data['polyUserOrders'] as Map).cast<String, dynamic>());
  }

  Future<List<PolyEvent>?> polyWatchlist({required String userId, String? selectionSet}) async {
    final data = await _executor.query('query polyWatchlist(\$userId: String!) { polyWatchlist(userId: \$userId) ${selectionSet ?? '{ markets { yes_price no_price id event_id slug question group_item_title description share_value min_order_size end_time status proposed_outcome dispute_deadline disputed_by dispute_reason resolved_outcome reward_pool reward_max_spread reward_min_size dispute_bond best_bid best_ask last_price volume created_at updated_at } id slug title description category tags image_url featured neg_risk status end_date volume created_at updated_at }'} }', variables: {
      'userId': userId,
    }
    );
    return (data['polyWatchlist'] as List?)?.map((e) => (e == null ? null : PolyEvent.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<PolyUserRewardsPage?> polyUserRewards({required String userId, int? limit, String? cursor, String? selectionSet}) async {
    final data = await _executor.query('query polyUserRewards(\$userId: String!, \$limit: Int, \$cursor: String) { polyUserRewards(userId: \$userId, limit: \$limit, cursor: \$cursor) ${selectionSet ?? '{ items { id epoch_start market_id user_id score amount created_at } next_cursor is_last_page total_paid }'} }', variables: {
      'userId': userId,
      'limit': ?limit,
      'cursor': ?cursor,
    }
    );
    return data['polyUserRewards'] == null ? null : PolyUserRewardsPage.fromJson((data['polyUserRewards'] as Map).cast<String, dynamic>());
  }

  Future<PolyReferralSummary?> polyReferral({required String userId, String? selectionSet}) async {
    final data = await _executor.query('query polyReferral(\$userId: String!) { polyReferral(userId: \$userId) ${selectionSet ?? '{ ref_code deep_link referred_by referees earnings commission_bps service_stats { total_invited total_installed total_rewarded total_reward_amt } }'} }', variables: {
      'userId': userId,
    }
    );
    return data['polyReferral'] == null ? null : PolyReferralSummary.fromJson((data['polyReferral'] as Map).cast<String, dynamic>());
  }

  Future<Object?> polyMyRewards() async {
    final data = await _executor.query('query polyMyRewards { polyMyRewards }'
    );
    return data['polyMyRewards'];
  }

  Future<Object?> polyMyReferral() async {
    final data = await _executor.query('query polyMyReferral { polyMyReferral }'
    );
    return data['polyMyReferral'];
  }

  Future<Object?> polyCreateEvent({required PolyEventInput input}) async {
    final data = await _executor.mutation('mutation polyCreateEvent(\$input: PolyEventInput!) { polyCreateEvent(input: \$input) }', variables: {
      'input': input.toJson(),
    }
    );
    return data['polyCreateEvent'];
  }

  Future<Object?> polyAddComment({required String id, required String body, int? parentId}) async {
    final data = await _executor.mutation('mutation polyAddComment(\$id: String!, \$body: String!, \$parent_id: Int) { polyAddComment(id: \$id, body: \$body, parent_id: \$parent_id) }', variables: {
      'id': id,
      'body': body,
      'parent_id': ?parentId,
    }
    );
    return data['polyAddComment'];
  }

  Future<Object?> polyConvertPositions({required String id, required List<int> marketIds, required int amount}) async {
    final data = await _executor.mutation('mutation polyConvertPositions(\$id: String!, \$market_ids: [Int!]!, \$amount: Int64!) { polyConvertPositions(id: \$id, market_ids: \$market_ids, amount: \$amount) }', variables: {
      'id': id,
      'market_ids': marketIds,
      'amount': amount,
    }
    );
    return data['polyConvertPositions'];
  }

  Future<Object?> polyProposeEvent({required String id, required int winnerMarketId}) async {
    final data = await _executor.mutation('mutation polyProposeEvent(\$id: String!, \$winner_market_id: Int!) { polyProposeEvent(id: \$id, winner_market_id: \$winner_market_id) }', variables: {
      'id': id,
      'winner_market_id': winnerMarketId,
    }
    );
    return data['polyProposeEvent'];
  }

  Future<Object?> polyResolveEvent({required String id, required int winnerMarketId}) async {
    final data = await _executor.mutation('mutation polyResolveEvent(\$id: String!, \$winner_market_id: Int!) { polyResolveEvent(id: \$id, winner_market_id: \$winner_market_id) }', variables: {
      'id': id,
      'winner_market_id': winnerMarketId,
    }
    );
    return data['polyResolveEvent'];
  }

  Future<Object?> polyWatchEvent({required String id}) async {
    final data = await _executor.mutation('mutation polyWatchEvent(\$id: String!) { polyWatchEvent(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['polyWatchEvent'];
  }

  Future<Object?> polyUnwatchEvent({required String id}) async {
    final data = await _executor.mutation('mutation polyUnwatchEvent(\$id: String!) { polyUnwatchEvent(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['polyUnwatchEvent'];
  }

  Future<Object?> polyDeleteComment({required String commentId}) async {
    final data = await _executor.mutation('mutation polyDeleteComment(\$commentId: String!) { polyDeleteComment(commentId: \$commentId) }', variables: {
      'commentId': commentId,
    }
    );
    return data['polyDeleteComment'];
  }

  Future<Object?> polyLikeComment({required String commentId}) async {
    final data = await _executor.mutation('mutation polyLikeComment(\$commentId: String!) { polyLikeComment(commentId: \$commentId) }', variables: {
      'commentId': commentId,
    }
    );
    return data['polyLikeComment'];
  }

  Future<Object?> polyUnlikeComment({required String commentId}) async {
    final data = await _executor.mutation('mutation polyUnlikeComment(\$commentId: String!) { polyUnlikeComment(commentId: \$commentId) }', variables: {
      'commentId': commentId,
    }
    );
    return data['polyUnlikeComment'];
  }

  Future<Object?> polySetMarketRewards({required String id, required int dailyPool, required int maxSpread, required int minSize}) async {
    final data = await _executor.mutation('mutation polySetMarketRewards(\$id: String!, \$daily_pool: Int64!, \$max_spread: Int64!, \$min_size: Int64!) { polySetMarketRewards(id: \$id, daily_pool: \$daily_pool, max_spread: \$max_spread, min_size: \$min_size) }', variables: {
      'id': id,
      'daily_pool': dailyPool,
      'max_spread': maxSpread,
      'min_size': minSize,
    }
    );
    return data['polySetMarketRewards'];
  }

  Future<Object?> polyPlaceOrder({required String id, required PolyOrderInput input}) async {
    final data = await _executor.mutation('mutation polyPlaceOrder(\$id: String!, \$input: PolyOrderInput!) { polyPlaceOrder(id: \$id, input: \$input) }', variables: {
      'id': id,
      'input': input.toJson(),
    }
    );
    return data['polyPlaceOrder'];
  }

  Future<Object?> polyCancelAllOrders({required String id}) async {
    final data = await _executor.mutation('mutation polyCancelAllOrders(\$id: String!) { polyCancelAllOrders(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['polyCancelAllOrders'];
  }

  Future<Object?> polyProposeResolution({required String id, required String outcome}) async {
    final data = await _executor.mutation('mutation polyProposeResolution(\$id: String!, \$outcome: String!) { polyProposeResolution(id: \$id, outcome: \$outcome) }', variables: {
      'id': id,
      'outcome': outcome,
    }
    );
    return data['polyProposeResolution'];
  }

  Future<Object?> polyDisputeResolution({required String id, required String reason}) async {
    final data = await _executor.mutation('mutation polyDisputeResolution(\$id: String!, \$reason: String!) { polyDisputeResolution(id: \$id, reason: \$reason) }', variables: {
      'id': id,
      'reason': reason,
    }
    );
    return data['polyDisputeResolution'];
  }

  Future<Object?> polyFinalizeResolution({required String id}) async {
    final data = await _executor.mutation('mutation polyFinalizeResolution(\$id: String!) { polyFinalizeResolution(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['polyFinalizeResolution'];
  }

  Future<Object?> polyResolveMarket({required String id, required String outcome}) async {
    final data = await _executor.mutation('mutation polyResolveMarket(\$id: String!, \$outcome: String!) { polyResolveMarket(id: \$id, outcome: \$outcome) }', variables: {
      'id': id,
      'outcome': outcome,
    }
    );
    return data['polyResolveMarket'];
  }

  Future<Object?> polyFundExchange({required String userId, required int amount}) async {
    final data = await _executor.mutation('mutation polyFundExchange(\$user_id: String!, \$amount: Int64!) { polyFundExchange(user_id: \$user_id, amount: \$amount) }', variables: {
      'user_id': userId,
      'amount': amount,
    }
    );
    return data['polyFundExchange'];
  }

  Future<Object?> polyCancelOrder({required String orderId}) async {
    final data = await _executor.mutation('mutation polyCancelOrder(\$orderId: String!) { polyCancelOrder(orderId: \$orderId) }', variables: {
      'orderId': orderId,
    }
    );
    return data['polyCancelOrder'];
  }

  Future<Object?> polyUpdateMyProfile({String? username, String? bio, String? avatarUrl}) async {
    final data = await _executor.mutation('mutation polyUpdateMyProfile(\$username: String, \$bio: String, \$avatar_url: String) { polyUpdateMyProfile(username: \$username, bio: \$bio, avatar_url: \$avatar_url) }', variables: {
      'username': ?username,
      'bio': ?bio,
      'avatar_url': ?avatarUrl,
    }
    );
    return data['polyUpdateMyProfile'];
  }

  Future<Object?> polyDeposit({required int amount}) async {
    final data = await _executor.mutation('mutation polyDeposit(\$amount: Int64!) { polyDeposit(amount: \$amount) }', variables: {
      'amount': amount,
    }
    );
    return data['polyDeposit'];
  }

  Future<Object?> polyWithdraw({required int amount}) async {
    final data = await _executor.mutation('mutation polyWithdraw(\$amount: Int64!) { polyWithdraw(amount: \$amount) }', variables: {
      'amount': amount,
    }
    );
    return data['polyWithdraw'];
  }

  Future<Object?> polyRedeemReferral({required String refCode}) async {
    final data = await _executor.mutation('mutation polyRedeemReferral(\$ref_code: String!) { polyRedeemReferral(ref_code: \$ref_code) }', variables: {
      'ref_code': refCode,
    }
    );
    return data['polyRedeemReferral'];
  }

}
