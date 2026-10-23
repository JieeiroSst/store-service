// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

// ignore_for_file: unused_import

import '../../../apps/kit/kit.dart';
import '../../../../gen/models.g.dart';
import '../admin_op.dart';

final luckyAdminOps = <AdminOp>[
  AdminOp(
    'luckyCampaigns',
    resource: 'campaign',
    verb: '',
    fields: [
      KField('game_type', 'Game type'),
      KField('status', 'Status'),
      KField('limit', 'Limit', kind: FieldKind.number),
      KField('offset', 'Offset', kind: FieldKind.number),
    ],
    run: (api, v) async => adminPlain(await api.lucky.luckyCampaigns(gameType: v.str('game_type'), status: v.str('status'), limit: v.integer('limit'), offset: v.integer('offset'))),
  ),
  AdminOp(
    'luckyCampaign',
    resource: 'campaign',
    verb: '',
    fields: [
      KField('id', 'ID', kind: FieldKind.number, required: true),
    ],
    run: (api, v) async => adminPlain(await api.lucky.luckyCampaign(id: v.need(v.integer('id'), 'id'))),
  ),
  AdminOp(
    'luckyFraudBlocks',
    resource: 'fraud block',
    verb: '',
    fields: [
      KField('limit', 'Limit', kind: FieldKind.number),
      KField('offset', 'Offset', kind: FieldKind.number),
    ],
    run: (api, v) async => adminPlain(await api.lucky.luckyFraudBlocks(limit: v.integer('limit'), offset: v.integer('offset'))),
  ),
  AdminOp(
    'luckyFraudRewards',
    resource: 'general',
    verb: '',
    fields: [
      KField('status', 'Status'),
      KField('user_id', 'User ID'),
      KField('limit', 'Limit', kind: FieldKind.number),
      KField('offset', 'Offset', kind: FieldKind.number),
    ],
    run: (api, v) async => adminPlain(await api.lucky.luckyFraudRewards(status: v.str('status'), userId: v.str('user_id'), limit: v.integer('limit'), offset: v.integer('offset'))),
  ),
  AdminOp(
    'luckyFraudPlays',
    resource: 'general',
    verb: '',
    fields: [
      KField('user_id', 'User ID'),
      KField('device_id', 'Device ID'),
      KField('ip', 'IP'),
      KField('limit', 'Limit', kind: FieldKind.number),
      KField('offset', 'Offset', kind: FieldKind.number),
    ],
    run: (api, v) async => adminPlain(await api.lucky.luckyFraudPlays(userId: v.str('user_id'), deviceId: v.str('device_id'), ip: v.str('ip'), limit: v.integer('limit'), offset: v.integer('offset'))),
  ),
  AdminOp(
    'luckyCampaignPlays',
    resource: 'campaign',
    verb: '',
    fields: [
      KField('id', 'ID', kind: FieldKind.number, required: true),
      KField('user_id', 'User ID'),
      KField('limit', 'Limit', kind: FieldKind.number),
      KField('offset', 'Offset', kind: FieldKind.number),
    ],
    run: (api, v) async => adminPlain(await api.lucky.luckyCampaignPlays(id: v.need(v.integer('id'), 'id'), userId: v.str('user_id'), limit: v.integer('limit'), offset: v.integer('offset'))),
  ),
  AdminOp(
    'luckyGrantPlays',
    resource: 'general',
    verb: 'grant',
    mutation: true,
    fields: [
      KField('user_id', 'User ID', required: true),
      KField('event', 'Event', required: true),
      KField('reference_id', 'Reference ID', required: true),
    ],
    run: (api, v) async => adminPlain(await api.lucky.luckyGrantPlays(userId: v.need(v.str('user_id'), 'user_id'), event: v.need(v.str('event'), 'event'), referenceId: v.need(v.str('reference_id'), 'reference_id'))),
  ),
  AdminOp(
    'luckyCampaignCreate',
    resource: 'campaign',
    verb: '',
    mutation: true,
    fields: [
      KField('input.game_type', 'Game type'),
      KField('input.name', 'Name', required: true),
      KField('input.description', 'Description', kind: FieldKind.multiline),
      KField('input.start_at', 'Start at', required: true, hint: 'ISO 8601, ví dụ 2026-10-22T09:00:00Z'),
      KField('input.end_at', 'End at', required: true, hint: 'ISO 8601, ví dụ 2026-10-22T09:00:00Z'),
      KField('input.timezone', 'Timezone'),
      KField('input.win_rate', 'Win rate', kind: FieldKind.number),
      KField('input.daily_spins', 'Daily spins', kind: FieldKind.number),
      KField('input.max_spins_per_day', 'Max spins per day', kind: FieldKind.number),
      KField('input.max_points_per_user_day', 'Max points per user day', kind: FieldKind.number),
      KField('input.budget_points', 'Budget points', kind: FieldKind.number),
      KField('input.review_points', 'Review points', kind: FieldKind.number),
      KField('input.spin_rules', 'Spin rules', kind: FieldKind.multiline, hint: 'JSON, ví dụ [{"event":"","spins":0,"max_per_day":0}]'),
      KField('input.checkin_rewards', 'Checkin rewards', hint: 'Nhiều giá trị, cách nhau bởi dấu phẩy'),
      KField('input.settings', 'Settings', kind: FieldKind.multiline, hint: 'JSON, ví dụ {"key": "value"}'),
      KField('input.prizes', 'Prizes', kind: FieldKind.multiline, hint: 'JSON, ví dụ [{"id":0,"name":"","type":"","points":0,"weight":0,"quantity":0,"position":0,"image_url":""}]'),
    ],
    run: (api, v) async => adminPlain(await api.lucky.luckyCampaignCreate(input: LuckyCampaignInput(gameType: v.str('input.game_type'), name_: v.need(v.str('input.name'), 'input.name'), description: v.str('input.description'), startAt: v.need(v.str('input.start_at'), 'input.start_at'), endAt: v.need(v.str('input.end_at'), 'input.end_at'), timezone: v.str('input.timezone'), winRate: v.number('input.win_rate'), dailySpins: v.integer('input.daily_spins'), maxSpinsPerDay: v.integer('input.max_spins_per_day'), maxPointsPerUserDay: v.integer('input.max_points_per_user_day'), budgetPoints: v.integer('input.budget_points'), reviewPoints: v.integer('input.review_points'), spinRules: (v.json('input.spin_rules') as List?)?.map((e) => (e == null ? null : LuckySpinRuleInput.fromJson((e as Map).cast<String, dynamic>()))!).toList(), checkinRewards: v.intList('input.checkin_rewards'), settings: v.jsonMap('input.settings'), prizes: (v.json('input.prizes') as List?)?.map((e) => (e == null ? null : LuckyPrizeInput.fromJson((e as Map).cast<String, dynamic>()))!).toList()))),
  ),
  AdminOp(
    'luckyCampaignUpdate',
    resource: 'campaign',
    verb: '',
    mutation: true,
    fields: [
      KField('id', 'ID', kind: FieldKind.number, required: true),
      KField('input.game_type', 'Game type'),
      KField('input.name', 'Name', required: true),
      KField('input.description', 'Description', kind: FieldKind.multiline),
      KField('input.start_at', 'Start at', required: true, hint: 'ISO 8601, ví dụ 2026-10-22T09:00:00Z'),
      KField('input.end_at', 'End at', required: true, hint: 'ISO 8601, ví dụ 2026-10-22T09:00:00Z'),
      KField('input.timezone', 'Timezone'),
      KField('input.win_rate', 'Win rate', kind: FieldKind.number),
      KField('input.daily_spins', 'Daily spins', kind: FieldKind.number),
      KField('input.max_spins_per_day', 'Max spins per day', kind: FieldKind.number),
      KField('input.max_points_per_user_day', 'Max points per user day', kind: FieldKind.number),
      KField('input.budget_points', 'Budget points', kind: FieldKind.number),
      KField('input.review_points', 'Review points', kind: FieldKind.number),
      KField('input.spin_rules', 'Spin rules', kind: FieldKind.multiline, hint: 'JSON, ví dụ [{"event":"","spins":0,"max_per_day":0}]'),
      KField('input.checkin_rewards', 'Checkin rewards', hint: 'Nhiều giá trị, cách nhau bởi dấu phẩy'),
      KField('input.settings', 'Settings', kind: FieldKind.multiline, hint: 'JSON, ví dụ {"key": "value"}'),
      KField('input.prizes', 'Prizes', kind: FieldKind.multiline, hint: 'JSON, ví dụ [{"id":0,"name":"","type":"","points":0,"weight":0,"quantity":0,"position":0,"image_url":""}]'),
    ],
    run: (api, v) async => adminPlain(await api.lucky.luckyCampaignUpdate(id: v.need(v.integer('id'), 'id'), input: LuckyCampaignInput(gameType: v.str('input.game_type'), name_: v.need(v.str('input.name'), 'input.name'), description: v.str('input.description'), startAt: v.need(v.str('input.start_at'), 'input.start_at'), endAt: v.need(v.str('input.end_at'), 'input.end_at'), timezone: v.str('input.timezone'), winRate: v.number('input.win_rate'), dailySpins: v.integer('input.daily_spins'), maxSpinsPerDay: v.integer('input.max_spins_per_day'), maxPointsPerUserDay: v.integer('input.max_points_per_user_day'), budgetPoints: v.integer('input.budget_points'), reviewPoints: v.integer('input.review_points'), spinRules: (v.json('input.spin_rules') as List?)?.map((e) => (e == null ? null : LuckySpinRuleInput.fromJson((e as Map).cast<String, dynamic>()))!).toList(), checkinRewards: v.intList('input.checkin_rewards'), settings: v.jsonMap('input.settings'), prizes: (v.json('input.prizes') as List?)?.map((e) => (e == null ? null : LuckyPrizeInput.fromJson((e as Map).cast<String, dynamic>()))!).toList()))),
  ),
  AdminOp(
    'luckyCampaignDelete',
    resource: 'campaign',
    verb: '',
    mutation: true,
    fields: [
      KField('id', 'ID', kind: FieldKind.number, required: true),
    ],
    run: (api, v) async => adminPlain(await api.lucky.luckyCampaignDelete(id: v.need(v.integer('id'), 'id'))),
  ),
  AdminOp(
    'luckyCampaignActivate',
    resource: 'campaign',
    verb: '',
    mutation: true,
    fields: [
      KField('id', 'ID', kind: FieldKind.number, required: true),
    ],
    run: (api, v) async => adminPlain(await api.lucky.luckyCampaignActivate(id: v.need(v.integer('id'), 'id'))),
  ),
  AdminOp(
    'luckyFraudBlock',
    resource: 'fraud block',
    verb: '',
    mutation: true,
    fields: [
      KField('user_id', 'User ID', required: true),
      KField('reason', 'Reason', kind: FieldKind.multiline, required: true),
    ],
    run: (api, v) async => adminPlain(await api.lucky.luckyFraudBlock(userId: v.need(v.str('user_id'), 'user_id'), reason: v.need(v.str('reason'), 'reason'))),
  ),
  AdminOp(
    'luckyFraudUnblock',
    resource: 'general',
    verb: '',
    mutation: true,
    fields: [
      KField('user_id', 'User ID', required: true),
    ],
    run: (api, v) async => adminPlain(await api.lucky.luckyFraudUnblock(userId: v.need(v.str('user_id'), 'user_id'))),
  ),
  AdminOp(
    'luckyFraudApproveReward',
    resource: 'general',
    verb: '',
    mutation: true,
    fields: [
      KField('id', 'ID', kind: FieldKind.number, required: true),
    ],
    run: (api, v) async => adminPlain(await api.lucky.luckyFraudApproveReward(id: v.need(v.integer('id'), 'id'))),
  ),
  AdminOp(
    'luckyFraudRejectReward',
    resource: 'general',
    verb: '',
    mutation: true,
    fields: [
      KField('id', 'ID', kind: FieldKind.number, required: true),
    ],
    run: (api, v) async => adminPlain(await api.lucky.luckyFraudRejectReward(id: v.need(v.integer('id'), 'id'))),
  ),
  AdminOp(
    'luckyCampaignPause',
    resource: 'campaign',
    verb: '',
    mutation: true,
    fields: [
      KField('id', 'ID', kind: FieldKind.number, required: true),
    ],
    run: (api, v) async => adminPlain(await api.lucky.luckyCampaignPause(id: v.need(v.integer('id'), 'id'))),
  ),
];
