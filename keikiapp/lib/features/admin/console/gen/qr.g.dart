// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

// ignore_for_file: unused_import

import '../../../apps/kit/kit.dart';
import '../../../../gen/models.g.dart';
import '../admin_op.dart';

final qrAdminOps = <AdminOp>[
  AdminOp(
    'qrCodes',
    resource: 'qr code',
    verb: '',
    fields: [
      KField('page', 'Page', kind: FieldKind.number),
      KField('limit', 'Limit', kind: FieldKind.number),
      KField('status', 'Status'),
      KField('type', 'Type'),
      KField('search', 'Search'),
      KField('created_by', 'Created by'),
    ],
    run: (api, v) async => adminPlain(await api.qr.qrCodes(page: v.integer('page'), limit: v.integer('limit'), status: v.str('status'), type_: v.str('type'), search: v.str('search'), createdBy: v.str('created_by'))),
  ),
  AdminOp(
    'qrCode',
    resource: 'qr code',
    verb: '',
    fields: [
      KField('id', 'ID', required: true),
    ],
    run: (api, v) async => adminPlain(await api.qr.qrCode(id: v.need(v.str('id'), 'id'))),
  ),
  AdminOp(
    'qrScanHistory',
    resource: 'general',
    verb: '',
    fields: [
      KField('id', 'ID', required: true),
      KField('page', 'Page', kind: FieldKind.number),
      KField('limit', 'Limit', kind: FieldKind.number),
    ],
    run: (api, v) async => adminPlain(await api.qr.qrScanHistory(id: v.need(v.str('id'), 'id'), page: v.integer('page'), limit: v.integer('limit'))),
  ),
  AdminOp(
    'qrScanStats',
    resource: 'general',
    verb: '',
    fields: [
      KField('id', 'ID', required: true),
    ],
    run: (api, v) async => adminPlain(await api.qr.qrScanStats(id: v.need(v.str('id'), 'id'))),
  ),
  AdminOp(
    'qrGenerate',
    resource: 'general',
    verb: '',
    mutation: true,
    fields: [
      KField('input.title', 'Title'),
      KField('input.type', 'Type'),
      KField('input.content', 'Content', kind: FieldKind.multiline),
      KField('input.redirect_url', 'Redirect URL'),
      KField('input.size', 'Size', kind: FieldKind.number),
      KField('input.fore_color', 'Fore color'),
      KField('input.back_color', 'Back color'),
      KField('input.expires_at', 'Expires at', hint: 'ISO 8601, ví dụ 2026-10-22T09:00:00Z'),
    ],
    run: (api, v) async => adminPlain(await api.qr.qrGenerate(input: QrGenerateInput(title: v.str('input.title'), type_: v.str('input.type'), content: v.str('input.content'), redirectUrl: v.str('input.redirect_url'), size: v.integer('input.size'), foreColor: v.str('input.fore_color'), backColor: v.str('input.back_color'), expiresAt: v.str('input.expires_at')))),
  ),
  AdminOp(
    'qrUpdate',
    resource: 'general',
    verb: '',
    mutation: true,
    fields: [
      KField('id', 'ID', required: true),
      KField('input.title', 'Title'),
      KField('input.status', 'Status'),
      KField('input.fore_color', 'Fore color'),
      KField('input.back_color', 'Back color'),
      KField('input.expires_at', 'Expires at', hint: 'ISO 8601, ví dụ 2026-10-22T09:00:00Z'),
    ],
    run: (api, v) async => adminPlain(await api.qr.qrUpdate(id: v.need(v.str('id'), 'id'), input: QrUpdateInput(title: v.str('input.title'), status: v.str('input.status'), foreColor: v.str('input.fore_color'), backColor: v.str('input.back_color'), expiresAt: v.str('input.expires_at')))),
  ),
  AdminOp(
    'qrUpdateContent',
    resource: 'general',
    verb: '',
    mutation: true,
    fields: [
      KField('id', 'ID', required: true),
      KField('content', 'Content', kind: FieldKind.multiline),
      KField('redirect_url', 'Redirect URL'),
    ],
    run: (api, v) async => adminPlain(await api.qr.qrUpdateContent(id: v.need(v.str('id'), 'id'), content: v.str('content'), redirectUrl: v.str('redirect_url'))),
  ),
  AdminOp(
    'qrDelete',
    resource: 'general',
    verb: '',
    mutation: true,
    fields: [
      KField('id', 'ID', required: true),
    ],
    run: (api, v) async => adminPlain(await api.qr.qrDelete(id: v.need(v.str('id'), 'id'))),
  ),
  AdminOp(
    'qrRegenerate',
    resource: 'general',
    verb: '',
    mutation: true,
    fields: [
      KField('id', 'ID', required: true),
    ],
    run: (api, v) async => adminPlain(await api.qr.qrRegenerate(id: v.need(v.str('id'), 'id'))),
  ),
];
