// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

import '../gateway_executor.dart';
import '../models.g.dart';

class RecompenseApi {
  const RecompenseApi(this._executor);

  final GatewayExecutor _executor;

  Future<List<RecompRecompense>?> recompenses({int? page, int? limit, String? selectionSet}) async {
    final data = await _executor.public('query recompenses(\$page: Int, \$limit: Int) { recompenses(page: \$page, limit: \$limit) ${selectionSet ?? '{ ID MemberID Source SourceID Tier PointsPerPurchase MinimumPoints CodeDeReduction ReductionDescription ReductionType ReductionValue MaxReduction ExpiryDate }'} }', variables: {
      'page': ?page,
      'limit': ?limit,
    }
    );
    return (data['recompenses'] as List?)?.map((e) => (e == null ? null : RecompRecompense.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<RecompRecompense?> recompense({required String id, String? selectionSet}) async {
    final data = await _executor.public('query recompense(\$id: String!) { recompense(id: \$id) ${selectionSet ?? '{ ID MemberID Source SourceID Tier PointsPerPurchase MinimumPoints CodeDeReduction ReductionDescription ReductionType ReductionValue MaxReduction ExpiryDate }'} }', variables: {
      'id': id,
    }
    );
    return data['recompense'] == null ? null : RecompRecompense.fromJson((data['recompense'] as Map).cast<String, dynamic>());
  }

  Future<List<RecompRecompense>?> recompMemberRecompenses({required int memberId, int? page, int? limit, String? selectionSet}) async {
    final data = await _executor.query('query recompMemberRecompenses(\$memberId: Int!, \$page: Int, \$limit: Int) { recompMemberRecompenses(memberId: \$memberId, page: \$page, limit: \$limit) ${selectionSet ?? '{ ID MemberID Source SourceID Tier PointsPerPurchase MinimumPoints CodeDeReduction ReductionDescription ReductionType ReductionValue MaxReduction ExpiryDate }'} }', variables: {
      'memberId': memberId,
      'page': ?page,
      'limit': ?limit,
    }
    );
    return (data['recompMemberRecompenses'] as List?)?.map((e) => (e == null ? null : RecompRecompense.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<RecompMemberStatus?> recompMemberPoints({required int memberId, String? selectionSet}) async {
    final data = await _executor.query('query recompMemberPoints(\$memberId: Int!) { recompMemberPoints(memberId: \$memberId) ${selectionSet ?? '{ MemberID Points Tier TierName NextTier PointsToNextTier }'} }', variables: {
      'memberId': memberId,
    }
    );
    return data['recompMemberPoints'] == null ? null : RecompMemberStatus.fromJson((data['recompMemberPoints'] as Map).cast<String, dynamic>());
  }

  Future<RecompRedemption?> recompBestRecompense({required int memberId, required double amount, String? selectionSet}) async {
    final data = await _executor.query('query recompBestRecompense(\$memberId: Int!, \$amount: Float!) { recompBestRecompense(memberId: \$memberId, amount: \$amount) ${selectionSet ?? '{ Recompense { ID MemberID Source SourceID Tier PointsPerPurchase MinimumPoints CodeDeReduction ReductionDescription ReductionType ReductionValue MaxReduction ExpiryDate } Amount Discount Total }'} }', variables: {
      'memberId': memberId,
      'amount': amount,
    }
    );
    return data['recompBestRecompense'] == null ? null : RecompRedemption.fromJson((data['recompBestRecompense'] as Map).cast<String, dynamic>());
  }

  Future<RecompMemberStatus?> myRecompensePoints({String? selectionSet}) async {
    final data = await _executor.query('query myRecompensePoints { myRecompensePoints ${selectionSet ?? '{ MemberID Points Tier TierName NextTier PointsToNextTier }'} }'
    );
    return data['myRecompensePoints'] == null ? null : RecompMemberStatus.fromJson((data['myRecompensePoints'] as Map).cast<String, dynamic>());
  }

  Future<RecompRecompense?> myBestRecompense({required double amount, String? selectionSet}) async {
    final data = await _executor.query('query myBestRecompense(\$amount: Float!) { myBestRecompense(amount: \$amount) ${selectionSet ?? '{ ID MemberID Source SourceID Tier PointsPerPurchase MinimumPoints CodeDeReduction ReductionDescription ReductionType ReductionValue MaxReduction ExpiryDate }'} }', variables: {
      'amount': amount,
    }
    );
    return data['myBestRecompense'] == null ? null : RecompRecompense.fromJson((data['myBestRecompense'] as Map).cast<String, dynamic>());
  }

  Future<Object?> myRecompenses({int? page, int? limit}) async {
    final data = await _executor.query('query myRecompenses(\$page: Int, \$limit: Int) { myRecompenses(page: \$page, limit: \$limit) }', variables: {
      'page': ?page,
      'limit': ?limit,
    }
    );
    return data['myRecompenses'];
  }

  Future<RecompRecompense?> recompCreateRecompense({required RecompRecompenseInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation recompCreateRecompense(\$input: RecompRecompenseInput!) { recompCreateRecompense(input: \$input) ${selectionSet ?? '{ ID MemberID Source SourceID Tier PointsPerPurchase MinimumPoints CodeDeReduction ReductionDescription ReductionType ReductionValue MaxReduction ExpiryDate }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['recompCreateRecompense'] == null ? null : RecompRecompense.fromJson((data['recompCreateRecompense'] as Map).cast<String, dynamic>());
  }

  Future<RecompRecompense?> recompUpdateRecompense({required int id, required RecompRecompenseInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation recompUpdateRecompense(\$id: Int!, \$input: RecompRecompenseInput!) { recompUpdateRecompense(id: \$id, input: \$input) ${selectionSet ?? '{ ID MemberID Source SourceID Tier PointsPerPurchase MinimumPoints CodeDeReduction ReductionDescription ReductionType ReductionValue MaxReduction ExpiryDate }'} }', variables: {
      'id': id,
      'input': input.toJson(),
    }
    );
    return data['recompUpdateRecompense'] == null ? null : RecompRecompense.fromJson((data['recompUpdateRecompense'] as Map).cast<String, dynamic>());
  }

  Future<bool?> recompDeleteRecompense({required int id}) async {
    final data = await _executor.mutation('mutation recompDeleteRecompense(\$id: Int!) { recompDeleteRecompense(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['recompDeleteRecompense'] as bool?;
  }

  Future<RecompMemberStatus?> recompEarnPoints({required int memberId, required double amount, required String idempotencyKey, String? selectionSet}) async {
    final data = await _executor.mutation('mutation recompEarnPoints(\$memberId: Int!, \$amount: Float!, \$idempotencyKey: String!) { recompEarnPoints(memberId: \$memberId, amount: \$amount, idempotencyKey: \$idempotencyKey) ${selectionSet ?? '{ MemberID Points Tier TierName NextTier PointsToNextTier }'} }', variables: {
      'memberId': memberId,
      'amount': amount,
      'idempotencyKey': idempotencyKey,
    }
    );
    return data['recompEarnPoints'] == null ? null : RecompMemberStatus.fromJson((data['recompEarnPoints'] as Map).cast<String, dynamic>());
  }

  Future<RecompRedemption?> recompRedeem({required int id, required int memberId, required double amount, String? selectionSet}) async {
    final data = await _executor.mutation('mutation recompRedeem(\$id: Int!, \$memberId: Int!, \$amount: Float!) { recompRedeem(id: \$id, memberId: \$memberId, amount: \$amount) ${selectionSet ?? '{ Recompense { ID MemberID Source SourceID Tier PointsPerPurchase MinimumPoints CodeDeReduction ReductionDescription ReductionType ReductionValue MaxReduction ExpiryDate } Amount Discount Total }'} }', variables: {
      'id': id,
      'memberId': memberId,
      'amount': amount,
    }
    );
    return data['recompRedeem'] == null ? null : RecompRedemption.fromJson((data['recompRedeem'] as Map).cast<String, dynamic>());
  }

  Future<RecompRedemption?> redeemMyRecompense({required int id, required double amount, String? selectionSet}) async {
    final data = await _executor.mutation('mutation redeemMyRecompense(\$id: Int!, \$amount: Float!) { redeemMyRecompense(id: \$id, amount: \$amount) ${selectionSet ?? '{ Recompense { ID MemberID Source SourceID Tier PointsPerPurchase MinimumPoints CodeDeReduction ReductionDescription ReductionType ReductionValue MaxReduction ExpiryDate } Amount Discount Total }'} }', variables: {
      'id': id,
      'amount': amount,
    }
    );
    return data['redeemMyRecompense'] == null ? null : RecompRedemption.fromJson((data['redeemMyRecompense'] as Map).cast<String, dynamic>());
  }

}
