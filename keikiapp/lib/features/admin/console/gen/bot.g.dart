// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

// ignore_for_file: unused_import

import '../../../apps/kit/kit.dart';
import '../../../../gen/models.g.dart';
import '../admin_op.dart';

final botAdminOps = <AdminOp>[
  AdminOp(
    'botPosts',
    resource: 'post',
    verb: '',
    fields: [
      KField('limit', 'Limit', kind: FieldKind.number),
      KField('offset', 'Offset', kind: FieldKind.number),
      KField('status', 'Status'),
      KField('campaign', 'Campaign'),
    ],
    run: (api, v) async => adminPlain(await api.bot.botPosts(limit: v.integer('limit'), offset: v.integer('offset'), status: v.str('status'), campaign: v.str('campaign'))),
  ),
  AdminOp(
    'botPost',
    resource: 'post',
    verb: '',
    fields: [
      KField('id', 'ID', required: true),
    ],
    run: (api, v) async => adminPlain(await api.bot.botPost(id: v.need(v.str('id'), 'id'))),
  ),
  AdminOp(
    'botCreatePost',
    resource: 'post',
    verb: 'create',
    mutation: true,
    fields: [
      KField('input.save_as_draft', 'Save as draft', kind: FieldKind.toggle),
      KField('input.title', 'Title'),
      KField('input.text', 'Text', kind: FieldKind.multiline),
      KField('input.hashtags', 'Hashtags', hint: 'Nhiều giá trị, cách nhau bởi dấu phẩy'),
      KField('input.media', 'Media', kind: FieldKind.multiline, hint: 'JSON, ví dụ [{"url":"","type":"","thumbnail":""}]'),
      KField('input.media_kind', 'Media kind'),
      KField('input.channels', 'Channels', hint: 'Nhiều giá trị, cách nhau bởi dấu phẩy'),
      KField('input.scheduled_at', 'Scheduled at', hint: 'ISO 8601, ví dụ 2026-10-22T09:00:00Z'),
      KField('input.timezone', 'Timezone'),
      KField('input.cron_expr', 'Cron expr'),
      KField('input.max_runs_per_day', 'Max runs per day', kind: FieldKind.number),
      KField('input.max_runs_per_month', 'Max runs per month', kind: FieldKind.number),
      KField('input.campaign', 'Campaign'),
    ],
    run: (api, v) async => adminPlain(await api.bot.botCreatePost(input: BotCreatePostInput(saveAsDraft: v.optFlag('input.save_as_draft'), title: v.str('input.title'), text: v.str('input.text'), hashtags: v.strs('input.hashtags'), media: (v.json('input.media') as List?)?.map((e) => (e == null ? null : BotMediaInput.fromJson((e as Map).cast<String, dynamic>()))!).toList(), mediaKind: v.str('input.media_kind'), channels: v.strs('input.channels'), scheduledAt: v.str('input.scheduled_at'), timezone: v.str('input.timezone'), cronExpr: v.str('input.cron_expr'), maxRunsPerDay: v.integer('input.max_runs_per_day'), maxRunsPerMonth: v.integer('input.max_runs_per_month'), campaign: v.str('input.campaign')))),
  ),
  AdminOp(
    'botUpdatePost',
    resource: 'post',
    verb: 'update',
    mutation: true,
    fields: [
      KField('id', 'ID', required: true),
      KField('input.title', 'Title'),
      KField('input.text', 'Text', kind: FieldKind.multiline),
      KField('input.hashtags', 'Hashtags', hint: 'Nhiều giá trị, cách nhau bởi dấu phẩy'),
      KField('input.media', 'Media', kind: FieldKind.multiline, hint: 'JSON, ví dụ [{"url":"","type":"","thumbnail":""}]'),
      KField('input.media_kind', 'Media kind'),
      KField('input.channels', 'Channels', hint: 'Nhiều giá trị, cách nhau bởi dấu phẩy'),
      KField('input.scheduled_at', 'Scheduled at', hint: 'ISO 8601, ví dụ 2026-10-22T09:00:00Z'),
      KField('input.timezone', 'Timezone'),
      KField('input.cron_expr', 'Cron expr'),
      KField('input.max_runs_per_day', 'Max runs per day', kind: FieldKind.number),
      KField('input.max_runs_per_month', 'Max runs per month', kind: FieldKind.number),
      KField('input.campaign', 'Campaign'),
    ],
    run: (api, v) async => adminPlain(await api.bot.botUpdatePost(id: v.need(v.str('id'), 'id'), input: BotUpdatePostInput(title: v.str('input.title'), text: v.str('input.text'), hashtags: v.strs('input.hashtags'), media: (v.json('input.media') as List?)?.map((e) => (e == null ? null : BotMediaInput.fromJson((e as Map).cast<String, dynamic>()))!).toList(), mediaKind: v.str('input.media_kind'), channels: v.strs('input.channels'), scheduledAt: v.str('input.scheduled_at'), timezone: v.str('input.timezone'), cronExpr: v.str('input.cron_expr'), maxRunsPerDay: v.integer('input.max_runs_per_day'), maxRunsPerMonth: v.integer('input.max_runs_per_month'), campaign: v.str('input.campaign')))),
  ),
  AdminOp(
    'botSubmitPost',
    resource: 'post',
    verb: 'submit',
    mutation: true,
    fields: [
      KField('id', 'ID', required: true),
    ],
    run: (api, v) async => adminPlain(await api.bot.botSubmitPost(id: v.need(v.str('id'), 'id'))),
  ),
  AdminOp(
    'botApprovePost',
    resource: 'post',
    verb: 'approve',
    mutation: true,
    fields: [
      KField('id', 'ID', required: true),
    ],
    run: (api, v) async => adminPlain(await api.bot.botApprovePost(id: v.need(v.str('id'), 'id'))),
  ),
  AdminOp(
    'botPublishPost',
    resource: 'post',
    verb: 'publish',
    mutation: true,
    fields: [
      KField('id', 'ID', required: true),
    ],
    run: (api, v) async => adminPlain(await api.bot.botPublishPost(id: v.need(v.str('id'), 'id'))),
  ),
  AdminOp(
    'botCancelPost',
    resource: 'post',
    verb: 'cancel',
    mutation: true,
    destructive: true,
    fields: [
      KField('id', 'ID', required: true),
    ],
    run: (api, v) async => adminPlain(await api.bot.botCancelPost(id: v.need(v.str('id'), 'id'))),
  ),
  AdminOp(
    'botRejectPost',
    resource: 'post',
    verb: 'reject',
    mutation: true,
    fields: [
      KField('id', 'ID', required: true),
      KField('reason', 'Reason', kind: FieldKind.multiline, required: true),
    ],
    run: (api, v) async => adminPlain(await api.bot.botRejectPost(id: v.need(v.str('id'), 'id'), reason: v.need(v.str('reason'), 'reason'))),
  ),
];
