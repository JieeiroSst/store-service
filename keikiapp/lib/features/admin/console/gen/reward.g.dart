// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

// ignore_for_file: unused_import

import '../../../apps/kit/kit.dart';
import '../../../../gen/models.g.dart';
import '../admin_op.dart';

final rewardAdminOps = <AdminOp>[
  AdminOp(
    'rewardCreateReward',
    resource: 'reward',
    verb: 'create',
    mutation: true,
    fields: [
      KField('input.name', 'Name'),
      KField('input.description', 'Description', kind: FieldKind.multiline),
      KField('input.points', 'Points', kind: FieldKind.number),
    ],
    run: (api, v) async => adminPlain(await api.reward.rewardCreateReward(input: RewardRewardInput(name_: v.str('input.name'), description: v.str('input.description'), points: v.integer('input.points')))),
  ),
  AdminOp(
    'rewardUpdateReward',
    resource: 'reward',
    verb: 'update',
    mutation: true,
    fields: [
      KField('reward_id', 'Reward ID', required: true),
      KField('input.name', 'Name'),
      KField('input.description', 'Description', kind: FieldKind.multiline),
      KField('input.points', 'Points', kind: FieldKind.number),
    ],
    run: (api, v) async => adminPlain(await api.reward.rewardUpdateReward(rewardId: v.need(v.str('reward_id'), 'reward_id'), input: RewardRewardInput(name_: v.str('input.name'), description: v.str('input.description'), points: v.integer('input.points')))),
  ),
  AdminOp(
    'rewardDeleteReward',
    resource: 'reward',
    verb: 'delete',
    mutation: true,
    destructive: true,
    fields: [
      KField('reward_id', 'Reward ID', required: true),
    ],
    run: (api, v) async => adminPlain(await api.reward.rewardDeleteReward(rewardId: v.need(v.str('reward_id'), 'reward_id'))),
  ),
];
