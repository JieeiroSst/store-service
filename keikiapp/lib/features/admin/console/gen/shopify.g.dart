// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

// ignore_for_file: unused_import

import '../../../apps/kit/kit.dart';
import '../../../../gen/models.g.dart';
import '../admin_op.dart';

final shopifyAdminOps = <AdminOp>[
  AdminOp(
    'shopifyOrders',
    resource: 'order',
    verb: '',
    fields: [
      KField('limit', 'Limit', kind: FieldKind.number),
      KField('offset', 'Offset', kind: FieldKind.number),
    ],
    run: (api, v) async => adminPlain(await api.shopify.shopifyOrders(limit: v.integer('limit'), offset: v.integer('offset'))),
  ),
  AdminOp(
    'shopifyOrder',
    resource: 'order',
    verb: '',
    fields: [
      KField('id', 'ID', kind: FieldKind.number, required: true),
    ],
    run: (api, v) async => adminPlain(await api.shopify.shopifyOrder(id: v.need(v.integer('id'), 'id'))),
  ),
  AdminOp(
    'shopifyCreateProduct',
    resource: 'product',
    verb: 'create',
    mutation: true,
    fields: [
      KField('title', 'Title', required: true),
      KField('description_html', 'Description html', kind: FieldKind.multiline),
      KField('vendor', 'Vendor'),
      KField('product_type', 'Product type'),
      KField('tags', 'Tags', hint: 'Nhiều giá trị, cách nhau bởi dấu phẩy'),
      KField('status', 'Status'),
    ],
    run: (api, v) async => adminPlain(await api.shopify.shopifyCreateProduct(title: v.need(v.str('title'), 'title'), descriptionHtml: v.str('description_html'), vendor: v.str('vendor'), productType: v.str('product_type'), tags: v.strs('tags'), status: v.str('status'))),
  ),
  AdminOp(
    'shopifySyncProducts',
    resource: 'product',
    verb: 'sync',
    mutation: true,
    run: (api, v) async => adminPlain(await api.shopify.shopifySyncProducts()),
  ),
  AdminOp(
    'shopifySyncOrders',
    resource: 'order',
    verb: 'sync',
    mutation: true,
    run: (api, v) async => adminPlain(await api.shopify.shopifySyncOrders()),
  ),
];
