// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

// ignore_for_file: unused_import

import '../../../apps/kit/kit.dart';
import '../../../../gen/models.g.dart';
import '../admin_op.dart';

final pointAdminOps = <AdminOp>[
  AdminOp(
    'pointUserBalance',
    resource: 'user balance',
    verb: '',
    fields: [
      KField('user_id', 'User ID', required: true),
    ],
    run: (api, v) async => adminPlain(await api.point.pointUserBalance(userId: v.need(v.str('user_id'), 'user_id'))),
  ),
  AdminOp(
    'pointUserTransactions',
    resource: 'user transaction',
    verb: '',
    fields: [
      KField('user_id', 'User ID', required: true),
      KField('limit', 'Limit', kind: FieldKind.number),
      KField('offset', 'Offset', kind: FieldKind.number),
    ],
    run: (api, v) async => adminPlain(await api.point.pointUserTransactions(userId: v.need(v.str('user_id'), 'user_id'), limit: v.integer('limit'), offset: v.integer('offset'))),
  ),
  AdminOp(
    'pointRewardPoints',
    resource: 'reward point',
    verb: '',
    fields: [
      KField('per_page', 'Per page', kind: FieldKind.number),
      KField('sort_order', 'Sort order'),
      KField('cursor', 'Cursor'),
    ],
    run: (api, v) async => adminPlain(await api.point.pointRewardPoints(perPage: v.integer('per_page'), sortOrder: v.str('sort_order'), cursor: v.str('cursor'))),
  ),
  AdminOp(
    'pointRewardPoint',
    resource: 'reward point',
    verb: '',
    fields: [
      KField('id', 'ID', required: true),
    ],
    run: (api, v) async => adminPlain(await api.point.pointRewardPoint(id: v.need(v.str('id'), 'id'))),
  ),
  AdminOp(
    'pointRewardDiscounts',
    resource: 'reward discount',
    verb: '',
    fields: [
      KField('per_page', 'Per page', kind: FieldKind.number),
      KField('sort_order', 'Sort order'),
      KField('cursor', 'Cursor'),
    ],
    run: (api, v) async => adminPlain(await api.point.pointRewardDiscounts(perPage: v.integer('per_page'), sortOrder: v.str('sort_order'), cursor: v.str('cursor'))),
  ),
  AdminOp(
    'pointRewardDiscount',
    resource: 'reward discount',
    verb: '',
    fields: [
      KField('id', 'ID', required: true),
    ],
    run: (api, v) async => adminPlain(await api.point.pointRewardDiscount(id: v.need(v.str('id'), 'id'))),
  ),
  AdminOp(
    'pointConvertedRewardPoints',
    resource: 'converted reward point',
    verb: '',
    fields: [
      KField('per_page', 'Per page', kind: FieldKind.number),
      KField('sort_order', 'Sort order'),
      KField('cursor', 'Cursor'),
    ],
    run: (api, v) async => adminPlain(await api.point.pointConvertedRewardPoints(perPage: v.integer('per_page'), sortOrder: v.str('sort_order'), cursor: v.str('cursor'))),
  ),
  AdminOp(
    'pointConvertedRewardPoint',
    resource: 'converted reward point',
    verb: '',
    fields: [
      KField('id', 'ID', required: true),
    ],
    run: (api, v) async => adminPlain(await api.point.pointConvertedRewardPoint(id: v.need(v.str('id'), 'id'))),
  ),
  AdminOp(
    'pointCreateRewardPoint',
    resource: 'reward point',
    verb: 'create',
    mutation: true,
    fields: [
      KField('input.reward_points_id', 'Reward points ID'),
      KField('input.total_points', 'Total points', kind: FieldKind.number),
      KField('input.points_pending', 'Points pending', kind: FieldKind.number),
      KField('input.points_active', 'Points active', kind: FieldKind.number),
      KField('input.points_expired', 'Points expired', kind: FieldKind.number),
      KField('input.points_converted', 'Points converted', kind: FieldKind.number),
      KField('input.points_cancelled', 'Points cancelled', kind: FieldKind.number),
      KField('input.activate_date', 'Activate date', kind: FieldKind.number),
      KField('input.expire_date', 'Expire date', kind: FieldKind.number),
    ],
    run: (api, v) async => adminPlain(await api.point.pointCreateRewardPoint(input: PointRewardPointInput(rewardPointsId: v.str('input.reward_points_id'), totalPoints: v.integer('input.total_points'), pointsPending: v.integer('input.points_pending'), pointsActive: v.integer('input.points_active'), pointsExpired: v.integer('input.points_expired'), pointsConverted: v.integer('input.points_converted'), pointsCancelled: v.integer('input.points_cancelled'), activateDate: v.integer('input.activate_date'), expireDate: v.integer('input.expire_date')))),
  ),
  AdminOp(
    'pointUpdateRewardPoint',
    resource: 'reward point',
    verb: 'update',
    mutation: true,
    fields: [
      KField('input.reward_points_id', 'Reward points ID'),
      KField('input.total_points', 'Total points', kind: FieldKind.number),
      KField('input.points_pending', 'Points pending', kind: FieldKind.number),
      KField('input.points_active', 'Points active', kind: FieldKind.number),
      KField('input.points_expired', 'Points expired', kind: FieldKind.number),
      KField('input.points_converted', 'Points converted', kind: FieldKind.number),
      KField('input.points_cancelled', 'Points cancelled', kind: FieldKind.number),
      KField('input.activate_date', 'Activate date', kind: FieldKind.number),
      KField('input.expire_date', 'Expire date', kind: FieldKind.number),
    ],
    run: (api, v) async => adminPlain(await api.point.pointUpdateRewardPoint(input: PointRewardPointInput(rewardPointsId: v.str('input.reward_points_id'), totalPoints: v.integer('input.total_points'), pointsPending: v.integer('input.points_pending'), pointsActive: v.integer('input.points_active'), pointsExpired: v.integer('input.points_expired'), pointsConverted: v.integer('input.points_converted'), pointsCancelled: v.integer('input.points_cancelled'), activateDate: v.integer('input.activate_date'), expireDate: v.integer('input.expire_date')))),
  ),
  AdminOp(
    'pointCreateRewardDiscount',
    resource: 'reward discount',
    verb: 'create',
    mutation: true,
    fields: [
      KField('input.reward_discount_id', 'Reward discount ID'),
      KField('input.total_points', 'Total points', kind: FieldKind.number),
      KField('input.points_pending', 'Points pending', kind: FieldKind.number),
      KField('input.points_active', 'Points active', kind: FieldKind.number),
      KField('input.points_expired', 'Points expired', kind: FieldKind.number),
      KField('input.points_converted', 'Points converted', kind: FieldKind.number),
      KField('input.points_cancelled', 'Points cancelled', kind: FieldKind.number),
      KField('input.activate_date', 'Activate date', kind: FieldKind.number),
      KField('input.expire_date', 'Expire date', kind: FieldKind.number),
    ],
    run: (api, v) async => adminPlain(await api.point.pointCreateRewardDiscount(input: PointRewardDiscountInput(rewardDiscountId: v.str('input.reward_discount_id'), totalPoints: v.integer('input.total_points'), pointsPending: v.integer('input.points_pending'), pointsActive: v.integer('input.points_active'), pointsExpired: v.integer('input.points_expired'), pointsConverted: v.integer('input.points_converted'), pointsCancelled: v.integer('input.points_cancelled'), activateDate: v.integer('input.activate_date'), expireDate: v.integer('input.expire_date')))),
  ),
  AdminOp(
    'pointUpdateRewardDiscount',
    resource: 'reward discount',
    verb: 'update',
    mutation: true,
    fields: [
      KField('input.reward_discount_id', 'Reward discount ID'),
      KField('input.total_points', 'Total points', kind: FieldKind.number),
      KField('input.points_pending', 'Points pending', kind: FieldKind.number),
      KField('input.points_active', 'Points active', kind: FieldKind.number),
      KField('input.points_expired', 'Points expired', kind: FieldKind.number),
      KField('input.points_converted', 'Points converted', kind: FieldKind.number),
      KField('input.points_cancelled', 'Points cancelled', kind: FieldKind.number),
      KField('input.activate_date', 'Activate date', kind: FieldKind.number),
      KField('input.expire_date', 'Expire date', kind: FieldKind.number),
    ],
    run: (api, v) async => adminPlain(await api.point.pointUpdateRewardDiscount(input: PointRewardDiscountInput(rewardDiscountId: v.str('input.reward_discount_id'), totalPoints: v.integer('input.total_points'), pointsPending: v.integer('input.points_pending'), pointsActive: v.integer('input.points_active'), pointsExpired: v.integer('input.points_expired'), pointsConverted: v.integer('input.points_converted'), pointsCancelled: v.integer('input.points_cancelled'), activateDate: v.integer('input.activate_date'), expireDate: v.integer('input.expire_date')))),
  ),
  AdminOp(
    'pointCreateConvertedRewardPoint',
    resource: 'converted reward point',
    verb: 'create',
    mutation: true,
    fields: [
      KField('input.rew_convert_id', 'Rew convert ID'),
      KField('input.rew_convert_ord_detail_id', 'Rew convert ord detail ID'),
      KField('input.rew_convert_discount_id', 'Rew convert discount ID'),
      KField('input.rew_convert_points', 'Rew convert points', kind: FieldKind.number),
      KField('input.rew_convert_date', 'Rew convert date', kind: FieldKind.number),
    ],
    run: (api, v) async => adminPlain(await api.point.pointCreateConvertedRewardPoint(input: PointConvertedRewardPointInput(rewConvertId: v.str('input.rew_convert_id'), rewConvertOrdDetailId: v.str('input.rew_convert_ord_detail_id'), rewConvertDiscountId: v.str('input.rew_convert_discount_id'), rewConvertPoints: v.integer('input.rew_convert_points'), rewConvertDate: v.integer('input.rew_convert_date')))),
  ),
  AdminOp(
    'pointUpdateConvertedRewardPoint',
    resource: 'converted reward point',
    verb: 'update',
    mutation: true,
    fields: [
      KField('input.rew_convert_id', 'Rew convert ID'),
      KField('input.rew_convert_ord_detail_id', 'Rew convert ord detail ID'),
      KField('input.rew_convert_discount_id', 'Rew convert discount ID'),
      KField('input.rew_convert_points', 'Rew convert points', kind: FieldKind.number),
      KField('input.rew_convert_date', 'Rew convert date', kind: FieldKind.number),
    ],
    run: (api, v) async => adminPlain(await api.point.pointUpdateConvertedRewardPoint(input: PointConvertedRewardPointInput(rewConvertId: v.str('input.rew_convert_id'), rewConvertOrdDetailId: v.str('input.rew_convert_ord_detail_id'), rewConvertDiscountId: v.str('input.rew_convert_discount_id'), rewConvertPoints: v.integer('input.rew_convert_points'), rewConvertDate: v.integer('input.rew_convert_date')))),
  ),
];
