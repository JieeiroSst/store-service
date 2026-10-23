// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

// ignore_for_file: unused_import

import '../../../apps/kit/kit.dart';
import '../../../../gen/models.g.dart';
import '../admin_op.dart';

final mediaAdminOps = <AdminOp>[
  AdminOp(
    'mediaVideo',
    resource: 'video',
    verb: '',
    fields: [
      KField('video_id', 'Video ID', kind: FieldKind.number, required: true),
      KField('user_id', 'User ID', kind: FieldKind.number, required: true),
    ],
    run: (api, v) async => adminPlain(await api.media.mediaVideo(videoId: v.need(v.integer('video_id'), 'video_id'), userId: v.need(v.integer('user_id'), 'user_id'))),
  ),
  AdminOp(
    'mediaUploadVideo',
    resource: 'upload video',
    verb: '',
    mutation: true,
    fields: [
      KField('video', 'Video', kind: FieldKind.file, required: true),
    ],
    run: (api, v) async => adminPlain(await api.media.mediaUploadVideo(video: v.need(v.file('video'), 'video'))),
  ),
  AdminOp(
    'mediaSaveVideo',
    resource: 'save video',
    verb: '',
    mutation: true,
    fields: [
      KField('video', 'Video', kind: FieldKind.multiline, required: true, hint: 'JSON, ví dụ {"key": "value"}'),
      KField('tag', 'Tag', kind: FieldKind.multiline, hint: 'JSON, ví dụ {"key": "value"}'),
    ],
    run: (api, v) async => adminPlain(await api.media.mediaSaveVideo(video: v.need(v.jsonMap('video'), 'video'), tag: v.jsonMap('tag'))),
  ),
];
