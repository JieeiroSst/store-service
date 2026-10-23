// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

// ignore_for_file: unused_import

import '../../../apps/kit/kit.dart';
import '../../../../gen/models.g.dart';
import '../admin_op.dart';

final referralAdminOps = <AdminOp>[
  AdminOp(
    'referralUserLinks',
    resource: 'user link',
    verb: '',
    fields: [
      KField('user_id', 'User ID', required: true),
      KField('limit', 'Limit', kind: FieldKind.number),
      KField('cursor', 'Cursor'),
    ],
    run: (api, v) async => adminPlain(await api.referral.referralUserLinks(userId: v.need(v.str('user_id'), 'user_id'), limit: v.integer('limit'), cursor: v.str('cursor'))),
  ),
  AdminOp(
    'referralUserStats',
    resource: 'user stat',
    verb: '',
    fields: [
      KField('user_id', 'User ID', required: true),
    ],
    run: (api, v) async => adminPlain(await api.referral.referralUserStats(userId: v.need(v.str('user_id'), 'user_id'))),
  ),
  AdminOp(
    'referralUserAttribution',
    resource: 'user attribution',
    verb: '',
    fields: [
      KField('user_id', 'User ID', required: true),
    ],
    run: (api, v) async => adminPlain(await api.referral.referralUserAttribution(userId: v.need(v.str('user_id'), 'user_id'))),
  ),
  AdminOp(
    'referralUTMReport',
    resource: 'utm report',
    verb: '',
    fields: [
      KField('owner_user_id', 'Owner user ID'),
      KField('utm_source', 'UTM source'),
      KField('utm_medium', 'UTM medium'),
      KField('utm_campaign', 'UTM campaign'),
      KField('from', 'From', kind: FieldKind.number),
      KField('to', 'To', kind: FieldKind.number),
    ],
    run: (api, v) async => adminPlain(await api.referral.referralUTMReport(ownerUserId: v.str('owner_user_id'), utmSource: v.str('utm_source'), utmMedium: v.str('utm_medium'), utmCampaign: v.str('utm_campaign'), from: v.integer('from'), to: v.integer('to'))),
  ),
  AdminOp(
    'referralCreateRewardProgram',
    resource: 'reward program',
    verb: 'create',
    mutation: true,
    fields: [
      KField('name', 'Name', required: true),
      KField('activate', 'Activate', kind: FieldKind.toggle),
      KField('tiers', 'Tiers', kind: FieldKind.multiline, hint: 'JSON, ví dụ [{"min_count":0,"max_count":0,"reward_value":0.0}]'),
    ],
    run: (api, v) async => adminPlain(await api.referral.referralCreateRewardProgram(name_: v.need(v.str('name'), 'name'), activate: v.optFlag('activate'), tiers: (v.json('tiers') as List?)?.map((e) => (e == null ? null : ReferralRewardTierInput.fromJson((e as Map).cast<String, dynamic>()))!).toList())),
  ),
  AdminOp(
    'referralSetRewardProgramStatus',
    resource: 'reward program status',
    verb: 'set',
    mutation: true,
    fields: [
      KField('program_id', 'Program ID', required: true),
      KField('active', 'Active', kind: FieldKind.toggle),
    ],
    run: (api, v) async => adminPlain(await api.referral.referralSetRewardProgramStatus(programId: v.need(v.str('program_id'), 'program_id'), active: v.flag('active'))),
  ),
];
