// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

// ignore_for_file: unused_import

import '../../../apps/kit/kit.dart';
import '../../../../gen/models.g.dart';
import '../admin_op.dart';

final recompenseAdminOps = <AdminOp>[
  AdminOp(
    'recompMemberRecompenses',
    resource: 'general',
    verb: '',
    fields: [
      KField('memberId', 'Member ID', kind: FieldKind.number, required: true),
      KField('page', 'Page', kind: FieldKind.number),
      KField('limit', 'Limit', kind: FieldKind.number),
    ],
    run: (api, v) async => adminPlain(await api.recompense.recompMemberRecompenses(memberId: v.need(v.integer('memberId'), 'memberId'), page: v.integer('page'), limit: v.integer('limit'))),
  ),
  AdminOp(
    'recompMemberPoints',
    resource: 'general',
    verb: '',
    fields: [
      KField('memberId', 'Member ID', kind: FieldKind.number, required: true),
    ],
    run: (api, v) async => adminPlain(await api.recompense.recompMemberPoints(memberId: v.need(v.integer('memberId'), 'memberId'))),
  ),
  AdminOp(
    'recompBestRecompense',
    resource: 'general',
    verb: '',
    fields: [
      KField('memberId', 'Member ID', kind: FieldKind.number, required: true),
      KField('amount', 'Amount', kind: FieldKind.number, required: true),
    ],
    run: (api, v) async => adminPlain(await api.recompense.recompBestRecompense(memberId: v.need(v.integer('memberId'), 'memberId'), amount: v.need(v.number('amount'), 'amount'))),
  ),
  AdminOp(
    'recompCreateRecompense',
    resource: 'general',
    verb: '',
    mutation: true,
    fields: [
      KField('input.ID', 'ID'),
      KField('input.MemberID', 'Member ID', kind: FieldKind.number),
      KField('input.Source', 'Source'),
      KField('input.SourceID', 'Source ID'),
      KField('input.Tier', 'Tier', kind: FieldKind.number),
      KField('input.PointsPerPurchase', 'Points per purchase', kind: FieldKind.number),
      KField('input.MinimumPoints', 'Minimum points', kind: FieldKind.number),
      KField('input.CodeDeReduction', 'Code de reduction'),
      KField('input.ReductionDescription', 'Reduction description', kind: FieldKind.multiline),
      KField('input.ReductionType', 'Reduction type'),
      KField('input.ReductionValue', 'Reduction value', kind: FieldKind.number),
      KField('input.MaxReduction', 'Max reduction', kind: FieldKind.number),
      KField('input.ExpiryDate', 'Expiry date', kind: FieldKind.number),
    ],
    run: (api, v) async => adminPlain(await api.recompense.recompCreateRecompense(input: RecompRecompenseInput(id: v.str('input.ID'), memberID: v.integer('input.MemberID'), source: v.str('input.Source'), sourceID: v.str('input.SourceID'), tier: v.integer('input.Tier'), pointsPerPurchase: v.integer('input.PointsPerPurchase'), minimumPoints: v.integer('input.MinimumPoints'), codeDeReduction: v.str('input.CodeDeReduction'), reductionDescription: v.str('input.ReductionDescription'), reductionType: v.str('input.ReductionType'), reductionValue: v.number('input.ReductionValue'), maxReduction: v.number('input.MaxReduction'), expiryDate: v.number('input.ExpiryDate')))),
  ),
  AdminOp(
    'recompUpdateRecompense',
    resource: 'general',
    verb: '',
    mutation: true,
    fields: [
      KField('id', 'ID', kind: FieldKind.number, required: true),
      KField('input.ID', 'ID'),
      KField('input.MemberID', 'Member ID', kind: FieldKind.number),
      KField('input.Source', 'Source'),
      KField('input.SourceID', 'Source ID'),
      KField('input.Tier', 'Tier', kind: FieldKind.number),
      KField('input.PointsPerPurchase', 'Points per purchase', kind: FieldKind.number),
      KField('input.MinimumPoints', 'Minimum points', kind: FieldKind.number),
      KField('input.CodeDeReduction', 'Code de reduction'),
      KField('input.ReductionDescription', 'Reduction description', kind: FieldKind.multiline),
      KField('input.ReductionType', 'Reduction type'),
      KField('input.ReductionValue', 'Reduction value', kind: FieldKind.number),
      KField('input.MaxReduction', 'Max reduction', kind: FieldKind.number),
      KField('input.ExpiryDate', 'Expiry date', kind: FieldKind.number),
    ],
    run: (api, v) async => adminPlain(await api.recompense.recompUpdateRecompense(id: v.need(v.integer('id'), 'id'), input: RecompRecompenseInput(id: v.str('input.ID'), memberID: v.integer('input.MemberID'), source: v.str('input.Source'), sourceID: v.str('input.SourceID'), tier: v.integer('input.Tier'), pointsPerPurchase: v.integer('input.PointsPerPurchase'), minimumPoints: v.integer('input.MinimumPoints'), codeDeReduction: v.str('input.CodeDeReduction'), reductionDescription: v.str('input.ReductionDescription'), reductionType: v.str('input.ReductionType'), reductionValue: v.number('input.ReductionValue'), maxReduction: v.number('input.MaxReduction'), expiryDate: v.number('input.ExpiryDate')))),
  ),
  AdminOp(
    'recompDeleteRecompense',
    resource: 'general',
    verb: '',
    mutation: true,
    fields: [
      KField('id', 'ID', kind: FieldKind.number, required: true),
    ],
    run: (api, v) async => adminPlain(await api.recompense.recompDeleteRecompense(id: v.need(v.integer('id'), 'id'))),
  ),
  AdminOp(
    'recompEarnPoints',
    resource: 'general',
    verb: '',
    mutation: true,
    fields: [
      KField('memberId', 'Member ID', kind: FieldKind.number, required: true),
      KField('amount', 'Amount', kind: FieldKind.number, required: true),
      KField('idempotencyKey', 'Idempotency key', required: true),
    ],
    run: (api, v) async => adminPlain(await api.recompense.recompEarnPoints(memberId: v.need(v.integer('memberId'), 'memberId'), amount: v.need(v.number('amount'), 'amount'), idempotencyKey: v.need(v.str('idempotencyKey'), 'idempotencyKey'))),
  ),
  AdminOp(
    'recompRedeem',
    resource: 'general',
    verb: '',
    mutation: true,
    fields: [
      KField('id', 'ID', kind: FieldKind.number, required: true),
      KField('memberId', 'Member ID', kind: FieldKind.number, required: true),
      KField('amount', 'Amount', kind: FieldKind.number, required: true),
    ],
    run: (api, v) async => adminPlain(await api.recompense.recompRedeem(id: v.need(v.integer('id'), 'id'), memberId: v.need(v.integer('memberId'), 'memberId'), amount: v.need(v.number('amount'), 'amount'))),
  ),
];
