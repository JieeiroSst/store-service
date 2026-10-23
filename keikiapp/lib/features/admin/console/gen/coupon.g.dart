// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

// ignore_for_file: unused_import

import '../../../apps/kit/kit.dart';
import '../../../../gen/models.g.dart';
import '../admin_op.dart';

final couponAdminOps = <AdminOp>[
  AdminOp(
    'couponCoupons',
    resource: 'coupon',
    verb: '',
    run: (api, v) async => adminPlain(await api.coupon.couponCoupons()),
  ),
  AdminOp(
    'couponCoupon',
    resource: 'coupon',
    verb: '',
    fields: [
      KField('id', 'ID', kind: FieldKind.number, required: true),
    ],
    run: (api, v) async => adminPlain(await api.coupon.couponCoupon(id: v.need(v.integer('id'), 'id'))),
  ),
  AdminOp(
    'couponRestrictions',
    resource: 'restriction',
    verb: '',
    fields: [
      KField('id', 'ID', kind: FieldKind.number, required: true),
    ],
    run: (api, v) async => adminPlain(await api.coupon.couponRestrictions(id: v.need(v.integer('id'), 'id'))),
  ),
  AdminOp(
    'couponUsages',
    resource: 'general',
    verb: '',
    fields: [
      KField('id', 'ID', kind: FieldKind.number, required: true),
    ],
    run: (api, v) async => adminPlain(await api.coupon.couponUsages(id: v.need(v.integer('id'), 'id'))),
  ),
  AdminOp(
    'couponUserCouponsByCoupon',
    resource: 'user coupon',
    verb: '',
    fields: [
      KField('id', 'ID', kind: FieldKind.number, required: true),
    ],
    run: (api, v) async => adminPlain(await api.coupon.couponUserCouponsByCoupon(id: v.need(v.integer('id'), 'id'))),
  ),
  AdminOp(
    'couponUserCoupons',
    resource: 'user coupon',
    verb: '',
    fields: [
      KField('user_id', 'User ID', kind: FieldKind.number, required: true),
    ],
    run: (api, v) async => adminPlain(await api.coupon.couponUserCoupons(userId: v.need(v.integer('user_id'), 'user_id'))),
  ),
  AdminOp(
    'couponUserUsages',
    resource: 'general',
    verb: '',
    fields: [
      KField('user_id', 'User ID', kind: FieldKind.number, required: true),
    ],
    run: (api, v) async => adminPlain(await api.coupon.couponUserUsages(userId: v.need(v.integer('user_id'), 'user_id'))),
  ),
  AdminOp(
    'couponCreateCoupon',
    resource: 'coupon',
    verb: 'create',
    mutation: true,
    fields: [
      KField('input.code', 'Code'),
      KField('input.type', 'Type'),
      KField('input.discount_value', 'Discount value', kind: FieldKind.number),
      KField('input.minimum_purchase', 'Minimum purchase', kind: FieldKind.number),
      KField('input.max_discount_amount', 'Max discount amount', kind: FieldKind.number),
      KField('input.description', 'Description', kind: FieldKind.multiline),
      KField('input.start_date', 'Start date', hint: 'ISO 8601, ví dụ 2026-10-22T09:00:00Z'),
      KField('input.end_date', 'End date', hint: 'ISO 8601, ví dụ 2026-10-22T09:00:00Z'),
      KField('input.is_active', 'Is active', kind: FieldKind.toggle),
      KField('input.max_uses', 'Max uses', kind: FieldKind.number),
    ],
    run: (api, v) async => adminPlain(await api.coupon.couponCreateCoupon(input: CouponCouponInput(code: v.str('input.code'), type_: v.str('input.type'), discountValue: v.number('input.discount_value'), minimumPurchase: v.number('input.minimum_purchase'), maxDiscountAmount: v.number('input.max_discount_amount'), description: v.str('input.description'), startDate: v.str('input.start_date'), endDate: v.str('input.end_date'), isActive: v.optFlag('input.is_active'), maxUses: v.integer('input.max_uses')))),
  ),
  AdminOp(
    'couponUpdateCoupon',
    resource: 'coupon',
    verb: 'update',
    mutation: true,
    fields: [
      KField('id', 'ID', kind: FieldKind.number, required: true),
      KField('input.code', 'Code'),
      KField('input.type', 'Type'),
      KField('input.discount_value', 'Discount value', kind: FieldKind.number),
      KField('input.minimum_purchase', 'Minimum purchase', kind: FieldKind.number),
      KField('input.max_discount_amount', 'Max discount amount', kind: FieldKind.number),
      KField('input.description', 'Description', kind: FieldKind.multiline),
      KField('input.start_date', 'Start date', hint: 'ISO 8601, ví dụ 2026-10-22T09:00:00Z'),
      KField('input.end_date', 'End date', hint: 'ISO 8601, ví dụ 2026-10-22T09:00:00Z'),
      KField('input.is_active', 'Is active', kind: FieldKind.toggle),
      KField('input.max_uses', 'Max uses', kind: FieldKind.number),
    ],
    run: (api, v) async => adminPlain(await api.coupon.couponUpdateCoupon(id: v.need(v.integer('id'), 'id'), input: CouponCouponInput(code: v.str('input.code'), type_: v.str('input.type'), discountValue: v.number('input.discount_value'), minimumPurchase: v.number('input.minimum_purchase'), maxDiscountAmount: v.number('input.max_discount_amount'), description: v.str('input.description'), startDate: v.str('input.start_date'), endDate: v.str('input.end_date'), isActive: v.optFlag('input.is_active'), maxUses: v.integer('input.max_uses')))),
  ),
  AdminOp(
    'couponDeleteCoupon',
    resource: 'coupon',
    verb: 'delete',
    mutation: true,
    destructive: true,
    fields: [
      KField('id', 'ID', kind: FieldKind.number, required: true),
    ],
    run: (api, v) async => adminPlain(await api.coupon.couponDeleteCoupon(id: v.need(v.integer('id'), 'id'))),
  ),
  AdminOp(
    'couponCreateRestriction',
    resource: 'restriction',
    verb: 'create',
    mutation: true,
    fields: [
      KField('id', 'ID', kind: FieldKind.number, required: true),
      KField('input.restriction_type', 'Restriction type'),
      KField('input.restricted_entity_id', 'Restricted entity ID', kind: FieldKind.number),
      KField('input.is_exclude', 'Is exclude', kind: FieldKind.toggle),
    ],
    run: (api, v) async => adminPlain(await api.coupon.couponCreateRestriction(id: v.need(v.integer('id'), 'id'), input: CouponRestrictionInput(restrictionType: v.str('input.restriction_type'), restrictedEntityId: v.integer('input.restricted_entity_id'), isExclude: v.optFlag('input.is_exclude')))),
  ),
  AdminOp(
    'couponUpdateRestriction',
    resource: 'restriction',
    verb: 'update',
    mutation: true,
    fields: [
      KField('id', 'ID', kind: FieldKind.number, required: true),
      KField('input.restriction_type', 'Restriction type'),
      KField('input.restricted_entity_id', 'Restricted entity ID', kind: FieldKind.number),
      KField('input.is_exclude', 'Is exclude', kind: FieldKind.toggle),
    ],
    run: (api, v) async => adminPlain(await api.coupon.couponUpdateRestriction(id: v.need(v.integer('id'), 'id'), input: CouponRestrictionInput(restrictionType: v.str('input.restriction_type'), restrictedEntityId: v.integer('input.restricted_entity_id'), isExclude: v.optFlag('input.is_exclude')))),
  ),
  AdminOp(
    'couponDeleteRestriction',
    resource: 'restriction',
    verb: 'delete',
    mutation: true,
    destructive: true,
    fields: [
      KField('id', 'ID', kind: FieldKind.number, required: true),
    ],
    run: (api, v) async => adminPlain(await api.coupon.couponDeleteRestriction(id: v.need(v.integer('id'), 'id'))),
  ),
  AdminOp(
    'couponAssign',
    resource: 'general',
    verb: 'assign',
    mutation: true,
    fields: [
      KField('input.user_id', 'User ID', kind: FieldKind.number),
      KField('input.coupon_id', 'Coupon ID', kind: FieldKind.number),
    ],
    run: (api, v) async => adminPlain(await api.coupon.couponAssign(input: CouponAssignInput(userId: v.integer('input.user_id'), couponId: v.integer('input.coupon_id')))),
  ),
  AdminOp(
    'couponUseUserCoupon',
    resource: 'general',
    verb: '',
    mutation: true,
    fields: [
      KField('id', 'ID', kind: FieldKind.number, required: true),
    ],
    run: (api, v) async => adminPlain(await api.coupon.couponUseUserCoupon(id: v.need(v.integer('id'), 'id'))),
  ),
  AdminOp(
    'couponUnuseUserCoupon',
    resource: 'general',
    verb: '',
    mutation: true,
    fields: [
      KField('id', 'ID', kind: FieldKind.number, required: true),
    ],
    run: (api, v) async => adminPlain(await api.coupon.couponUnuseUserCoupon(id: v.need(v.integer('id'), 'id'))),
  ),
  AdminOp(
    'couponDeleteUserCoupon',
    resource: 'user coupon',
    verb: 'delete',
    mutation: true,
    destructive: true,
    fields: [
      KField('id', 'ID', kind: FieldKind.number, required: true),
    ],
    run: (api, v) async => adminPlain(await api.coupon.couponDeleteUserCoupon(id: v.need(v.integer('id'), 'id'))),
  ),
];
