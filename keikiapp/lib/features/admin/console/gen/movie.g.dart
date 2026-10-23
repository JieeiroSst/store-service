// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

// ignore_for_file: unused_import

import '../../../apps/kit/kit.dart';
import '../../../../gen/models.g.dart';
import '../admin_op.dart';

final movieAdminOps = <AdminOp>[
  AdminOp(
    'movieForUser',
    resource: 'general',
    verb: '',
    fields: [
      KField('user_id', 'User ID', required: true),
      KField('page', 'Page', kind: FieldKind.number),
      KField('page_size', 'Page size', kind: FieldKind.number),
      KField('snapshot', 'Snapshot'),
    ],
    run: (api, v) async => adminPlain(await api.movie.movieForUser(userId: v.need(v.str('user_id'), 'user_id'), page: v.integer('page'), pageSize: v.integer('page_size'), snapshot: v.str('snapshot'))),
  ),
  AdminOp(
    'movieHome',
    resource: 'home',
    verb: '',
    fields: [
      KField('user_id', 'User ID', required: true),
    ],
    run: (api, v) async => adminPlain(await api.movie.movieHome(userId: v.need(v.str('user_id'), 'user_id'))),
  ),
  AdminOp(
    'movieContinueWatching',
    resource: 'continue watching',
    verb: '',
    fields: [
      KField('user_id', 'User ID', required: true),
      KField('page', 'Page', kind: FieldKind.number),
      KField('page_size', 'Page size', kind: FieldKind.number),
      KField('snapshot', 'Snapshot'),
    ],
    run: (api, v) async => adminPlain(await api.movie.movieContinueWatching(userId: v.need(v.str('user_id'), 'user_id'), page: v.integer('page'), pageSize: v.integer('page_size'), snapshot: v.str('snapshot'))),
  ),
  AdminOp(
    'movieHistory',
    resource: 'history',
    verb: '',
    fields: [
      KField('user_id', 'User ID', required: true),
      KField('page', 'Page', kind: FieldKind.number),
      KField('page_size', 'Page size', kind: FieldKind.number),
      KField('snapshot', 'Snapshot'),
    ],
    run: (api, v) async => adminPlain(await api.movie.movieHistory(userId: v.need(v.str('user_id'), 'user_id'), page: v.integer('page'), pageSize: v.integer('page_size'), snapshot: v.str('snapshot'))),
  ),
  AdminOp(
    'movieUpsertVideo',
    resource: 'video',
    verb: 'upsert',
    mutation: true,
    fields: [
      KField('id', 'ID', required: true),
      KField('input.title', 'Title'),
      KField('input.description', 'Description', kind: FieldKind.multiline),
      KField('input.tags', 'Tags', hint: 'Nhiều giá trị, cách nhau bởi dấu phẩy'),
      KField('input.status', 'Status'),
      KField('input.duration', 'Duration', kind: FieldKind.number),
    ],
    run: (api, v) async => adminPlain(await api.movie.movieUpsertVideo(id: v.need(v.str('id'), 'id'), input: MovieVideoInput(title: v.str('input.title'), description: v.str('input.description'), tags: v.strs('input.tags'), status: v.str('input.status'), duration: v.number('input.duration')))),
  ),
  AdminOp(
    'movieDeleteVideo',
    resource: 'video',
    verb: 'delete',
    mutation: true,
    destructive: true,
    fields: [
      KField('id', 'ID', required: true),
    ],
    run: (api, v) async => adminPlain(await api.movie.movieDeleteVideo(id: v.need(v.str('id'), 'id'))),
  ),
];
