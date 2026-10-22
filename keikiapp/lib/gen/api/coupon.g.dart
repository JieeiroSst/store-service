// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

import '../gateway_executor.dart';
import '../models.g.dart';

class CouponApi {
  const CouponApi(this._executor);

  final GatewayExecutor _executor;

  Future<List<CouponCoupon>?> couponCoupons({String? selectionSet}) async {
    final data = await _executor.query('query couponCoupons { couponCoupons ${selectionSet ?? '{ id code type discount_value minimum_purchase max_discount_amount description start_date end_date is_active max_uses current_uses created_at updated_at }'} }'
    );
    return (data['couponCoupons'] as List?)?.map((e) => (e == null ? null : CouponCoupon.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<CouponCoupon?> couponCoupon({required int id, String? selectionSet}) async {
    final data = await _executor.query('query couponCoupon(\$id: Int!) { couponCoupon(id: \$id) ${selectionSet ?? '{ id code type discount_value minimum_purchase max_discount_amount description start_date end_date is_active max_uses current_uses created_at updated_at }'} }', variables: {
      'id': id,
    }
    );
    return data['couponCoupon'] == null ? null : CouponCoupon.fromJson((data['couponCoupon'] as Map).cast<String, dynamic>());
  }

  Future<CouponCoupon?> couponByCode({required String code, String? selectionSet}) async {
    final data = await _executor.query('query couponByCode(\$code: String!) { couponByCode(code: \$code) ${selectionSet ?? '{ id code type discount_value minimum_purchase max_discount_amount description start_date end_date is_active max_uses current_uses created_at updated_at }'} }', variables: {
      'code': code,
    }
    );
    return data['couponByCode'] == null ? null : CouponCoupon.fromJson((data['couponByCode'] as Map).cast<String, dynamic>());
  }

  Future<List<CouponRestriction>?> couponRestrictions({required int id, String? selectionSet}) async {
    final data = await _executor.query('query couponRestrictions(\$id: Int!) { couponRestrictions(id: \$id) ${selectionSet ?? '{ id coupon_id restriction_type restricted_entity_id is_exclude created_at }'} }', variables: {
      'id': id,
    }
    );
    return (data['couponRestrictions'] as List?)?.map((e) => (e == null ? null : CouponRestriction.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<List<CouponUsage>?> couponUsages({required int id, String? selectionSet}) async {
    final data = await _executor.query('query couponUsages(\$id: Int!) { couponUsages(id: \$id) ${selectionSet ?? '{ id coupon_id user_id order_id discount_amount used_at }'} }', variables: {
      'id': id,
    }
    );
    return (data['couponUsages'] as List?)?.map((e) => (e == null ? null : CouponUsage.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<List<CouponUserCoupon>?> couponUserCouponsByCoupon({required int id, String? selectionSet}) async {
    final data = await _executor.query('query couponUserCouponsByCoupon(\$id: Int!) { couponUserCouponsByCoupon(id: \$id) ${selectionSet ?? '{ id user_id coupon_id is_used assigned_at used_at coupon { id code type discount_value minimum_purchase max_discount_amount description start_date end_date is_active max_uses current_uses created_at updated_at } }'} }', variables: {
      'id': id,
    }
    );
    return (data['couponUserCouponsByCoupon'] as List?)?.map((e) => (e == null ? null : CouponUserCoupon.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<CouponValidation?> validateCoupon({required CouponValidateInput input, String? selectionSet}) async {
    final data = await _executor.query('query validateCoupon(\$input: CouponValidateInput!) { validateCoupon(input: \$input) ${selectionSet ?? '{ coupon { id code type discount_value minimum_purchase max_discount_amount description start_date end_date is_active max_uses current_uses created_at updated_at } discount }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['validateCoupon'] == null ? null : CouponValidation.fromJson((data['validateCoupon'] as Map).cast<String, dynamic>());
  }

  Future<List<CouponUserCoupon>?> couponUserCoupons({required int userId, String? selectionSet}) async {
    final data = await _executor.query('query couponUserCoupons(\$user_id: Int!) { couponUserCoupons(user_id: \$user_id) ${selectionSet ?? '{ id user_id coupon_id is_used assigned_at used_at coupon { id code type discount_value minimum_purchase max_discount_amount description start_date end_date is_active max_uses current_uses created_at updated_at } }'} }', variables: {
      'user_id': userId,
    }
    );
    return (data['couponUserCoupons'] as List?)?.map((e) => (e == null ? null : CouponUserCoupon.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<List<CouponUserCoupon>?> myCoupons({String? selectionSet}) async {
    final data = await _executor.query('query myCoupons { myCoupons ${selectionSet ?? '{ id user_id coupon_id is_used assigned_at used_at coupon { id code type discount_value minimum_purchase max_discount_amount description start_date end_date is_active max_uses current_uses created_at updated_at } }'} }'
    );
    return (data['myCoupons'] as List?)?.map((e) => (e == null ? null : CouponUserCoupon.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<List<CouponUsage>?> couponUserUsages({required int userId, String? selectionSet}) async {
    final data = await _executor.query('query couponUserUsages(\$user_id: Int!) { couponUserUsages(user_id: \$user_id) ${selectionSet ?? '{ id coupon_id user_id order_id discount_amount used_at }'} }', variables: {
      'user_id': userId,
    }
    );
    return (data['couponUserUsages'] as List?)?.map((e) => (e == null ? null : CouponUsage.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<List<CouponUsage>?> myCouponUsages({String? selectionSet}) async {
    final data = await _executor.query('query myCouponUsages { myCouponUsages ${selectionSet ?? '{ id coupon_id user_id order_id discount_amount used_at }'} }'
    );
    return (data['myCouponUsages'] as List?)?.map((e) => (e == null ? null : CouponUsage.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<CouponCoupon?> couponCreateCoupon({required CouponCouponInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation couponCreateCoupon(\$input: CouponCouponInput!) { couponCreateCoupon(input: \$input) ${selectionSet ?? '{ id code type discount_value minimum_purchase max_discount_amount description start_date end_date is_active max_uses current_uses created_at updated_at }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['couponCreateCoupon'] == null ? null : CouponCoupon.fromJson((data['couponCreateCoupon'] as Map).cast<String, dynamic>());
  }

  Future<CouponCoupon?> couponUpdateCoupon({required int id, required CouponCouponInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation couponUpdateCoupon(\$id: Int!, \$input: CouponCouponInput!) { couponUpdateCoupon(id: \$id, input: \$input) ${selectionSet ?? '{ id code type discount_value minimum_purchase max_discount_amount description start_date end_date is_active max_uses current_uses created_at updated_at }'} }', variables: {
      'id': id,
      'input': input.toJson(),
    }
    );
    return data['couponUpdateCoupon'] == null ? null : CouponCoupon.fromJson((data['couponUpdateCoupon'] as Map).cast<String, dynamic>());
  }

  Future<bool?> couponDeleteCoupon({required int id}) async {
    final data = await _executor.mutation('mutation couponDeleteCoupon(\$id: Int!) { couponDeleteCoupon(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['couponDeleteCoupon'] as bool?;
  }

  Future<CouponRestriction?> couponCreateRestriction({required int id, required CouponRestrictionInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation couponCreateRestriction(\$id: Int!, \$input: CouponRestrictionInput!) { couponCreateRestriction(id: \$id, input: \$input) ${selectionSet ?? '{ id coupon_id restriction_type restricted_entity_id is_exclude created_at }'} }', variables: {
      'id': id,
      'input': input.toJson(),
    }
    );
    return data['couponCreateRestriction'] == null ? null : CouponRestriction.fromJson((data['couponCreateRestriction'] as Map).cast<String, dynamic>());
  }

  Future<CouponRestriction?> couponUpdateRestriction({required int id, required CouponRestrictionInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation couponUpdateRestriction(\$id: Int!, \$input: CouponRestrictionInput!) { couponUpdateRestriction(id: \$id, input: \$input) ${selectionSet ?? '{ id coupon_id restriction_type restricted_entity_id is_exclude created_at }'} }', variables: {
      'id': id,
      'input': input.toJson(),
    }
    );
    return data['couponUpdateRestriction'] == null ? null : CouponRestriction.fromJson((data['couponUpdateRestriction'] as Map).cast<String, dynamic>());
  }

  Future<bool?> couponDeleteRestriction({required int id}) async {
    final data = await _executor.mutation('mutation couponDeleteRestriction(\$id: Int!) { couponDeleteRestriction(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['couponDeleteRestriction'] as bool?;
  }

  Future<CouponUsage?> applyCoupon({required CouponApplyInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation applyCoupon(\$input: CouponApplyInput!) { applyCoupon(input: \$input) ${selectionSet ?? '{ id coupon_id user_id order_id discount_amount used_at }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['applyCoupon'] == null ? null : CouponUsage.fromJson((data['applyCoupon'] as Map).cast<String, dynamic>());
  }

  Future<CouponUserCoupon?> couponAssign({required CouponAssignInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation couponAssign(\$input: CouponAssignInput!) { couponAssign(input: \$input) ${selectionSet ?? '{ id user_id coupon_id is_used assigned_at used_at coupon { id code type discount_value minimum_purchase max_discount_amount description start_date end_date is_active max_uses current_uses created_at updated_at } }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['couponAssign'] == null ? null : CouponUserCoupon.fromJson((data['couponAssign'] as Map).cast<String, dynamic>());
  }

  Future<CouponUserCoupon?> couponUseUserCoupon({required int id, String? selectionSet}) async {
    final data = await _executor.mutation('mutation couponUseUserCoupon(\$id: Int!) { couponUseUserCoupon(id: \$id) ${selectionSet ?? '{ id user_id coupon_id is_used assigned_at used_at coupon { id code type discount_value minimum_purchase max_discount_amount description start_date end_date is_active max_uses current_uses created_at updated_at } }'} }', variables: {
      'id': id,
    }
    );
    return data['couponUseUserCoupon'] == null ? null : CouponUserCoupon.fromJson((data['couponUseUserCoupon'] as Map).cast<String, dynamic>());
  }

  Future<CouponUserCoupon?> couponUnuseUserCoupon({required int id, String? selectionSet}) async {
    final data = await _executor.mutation('mutation couponUnuseUserCoupon(\$id: Int!) { couponUnuseUserCoupon(id: \$id) ${selectionSet ?? '{ id user_id coupon_id is_used assigned_at used_at coupon { id code type discount_value minimum_purchase max_discount_amount description start_date end_date is_active max_uses current_uses created_at updated_at } }'} }', variables: {
      'id': id,
    }
    );
    return data['couponUnuseUserCoupon'] == null ? null : CouponUserCoupon.fromJson((data['couponUnuseUserCoupon'] as Map).cast<String, dynamic>());
  }

  Future<bool?> couponDeleteUserCoupon({required int id}) async {
    final data = await _executor.mutation('mutation couponDeleteUserCoupon(\$id: Int!) { couponDeleteUserCoupon(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['couponDeleteUserCoupon'] as bool?;
  }

}
