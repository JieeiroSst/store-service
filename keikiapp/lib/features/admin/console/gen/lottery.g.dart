// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

// ignore_for_file: unused_import

import '../../../apps/kit/kit.dart';
import '../../../../gen/models.g.dart';
import '../admin_op.dart';

final lotteryAdminOps = <AdminOp>[
  AdminOp(
    'lotteryScrape',
    resource: 'scrape',
    verb: '',
    mutation: true,
    fields: [
      KField('date', 'Date', required: true, hint: 'ISO 8601, ví dụ 2026-10-22T09:00:00Z'),
    ],
    run: (api, v) async => adminPlain(await api.lottery.lotteryScrape(date: v.need(v.str('date'), 'date'))),
  ),
  AdminOp(
    'lotteryBackfill',
    resource: 'backfill',
    verb: '',
    mutation: true,
    fields: [
      KField('start', 'Start', required: true),
      KField('end', 'End', required: true),
      KField('force', 'Force', kind: FieldKind.toggle),
    ],
    run: (api, v) async => adminPlain(await api.lottery.lotteryBackfill(start: v.need(v.str('start'), 'start'), end: v.need(v.str('end'), 'end'), force: v.optFlag('force'))),
  ),
];
