// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

// ignore_for_file: unused_import

import '../../../apps/kit/kit.dart';
import '../../../../gen/models.g.dart';
import '../admin_op.dart';

final bookStoreAdminOps = <AdminOp>[
  AdminOp(
    'bookStoreCreateBook',
    resource: 'book',
    verb: 'create',
    mutation: true,
    fields: [
      KField('input.name', 'Name'),
      KField('input.author_id', 'Author ID', kind: FieldKind.number),
      KField('input.publisher_id', 'Publisher ID', kind: FieldKind.number),
      KField('input.category_id', 'Category ID', kind: FieldKind.number),
      KField('input.price', 'Price', kind: FieldKind.number),
      KField('input.quantity', 'Quantity', kind: FieldKind.number),
      KField('input.unit_of_measurement', 'Unit of measurement'),
    ],
    run: (api, v) async => adminPlain(await api.bookStore.bookStoreCreateBook(input: BookStoreBookInput(name_: v.str('input.name'), authorId: v.integer('input.author_id'), publisherId: v.integer('input.publisher_id'), categoryId: v.integer('input.category_id'), price: v.number('input.price'), quantity: v.integer('input.quantity'), unitOfMeasurement: v.str('input.unit_of_measurement')))),
  ),
  AdminOp(
    'bookStoreUpdateBook',
    resource: 'book',
    verb: 'update',
    mutation: true,
    fields: [
      KField('id', 'ID', kind: FieldKind.number, required: true),
      KField('input.name', 'Name'),
      KField('input.author_id', 'Author ID', kind: FieldKind.number),
      KField('input.publisher_id', 'Publisher ID', kind: FieldKind.number),
      KField('input.category_id', 'Category ID', kind: FieldKind.number),
      KField('input.price', 'Price', kind: FieldKind.number),
      KField('input.quantity', 'Quantity', kind: FieldKind.number),
      KField('input.unit_of_measurement', 'Unit of measurement'),
    ],
    run: (api, v) async => adminPlain(await api.bookStore.bookStoreUpdateBook(id: v.need(v.integer('id'), 'id'), input: BookStoreBookInput(name_: v.str('input.name'), authorId: v.integer('input.author_id'), publisherId: v.integer('input.publisher_id'), categoryId: v.integer('input.category_id'), price: v.number('input.price'), quantity: v.integer('input.quantity'), unitOfMeasurement: v.str('input.unit_of_measurement')))),
  ),
  AdminOp(
    'bookStoreDeleteBook',
    resource: 'book',
    verb: 'delete',
    mutation: true,
    destructive: true,
    fields: [
      KField('id', 'ID', kind: FieldKind.number, required: true),
    ],
    run: (api, v) async => adminPlain(await api.bookStore.bookStoreDeleteBook(id: v.need(v.integer('id'), 'id'))),
  ),
  AdminOp(
    'bookStoreCreateAuthor',
    resource: 'author',
    verb: 'create',
    mutation: true,
    fields: [
      KField('input.name', 'Name'),
      KField('input.bio', 'Bio', kind: FieldKind.multiline),
    ],
    run: (api, v) async => adminPlain(await api.bookStore.bookStoreCreateAuthor(input: BookStoreAuthorInput(name_: v.str('input.name'), bio: v.str('input.bio')))),
  ),
  AdminOp(
    'bookStoreUpdateAuthor',
    resource: 'author',
    verb: 'update',
    mutation: true,
    fields: [
      KField('id', 'ID', kind: FieldKind.number, required: true),
      KField('input.name', 'Name'),
      KField('input.bio', 'Bio', kind: FieldKind.multiline),
    ],
    run: (api, v) async => adminPlain(await api.bookStore.bookStoreUpdateAuthor(id: v.need(v.integer('id'), 'id'), input: BookStoreAuthorInput(name_: v.str('input.name'), bio: v.str('input.bio')))),
  ),
  AdminOp(
    'bookStoreDeleteAuthor',
    resource: 'author',
    verb: 'delete',
    mutation: true,
    destructive: true,
    fields: [
      KField('id', 'ID', kind: FieldKind.number, required: true),
    ],
    run: (api, v) async => adminPlain(await api.bookStore.bookStoreDeleteAuthor(id: v.need(v.integer('id'), 'id'))),
  ),
  AdminOp(
    'bookStoreCreatePublisher',
    resource: 'publisher',
    verb: 'create',
    mutation: true,
    fields: [
      KField('input.name', 'Name'),
      KField('input.address', 'Address'),
    ],
    run: (api, v) async => adminPlain(await api.bookStore.bookStoreCreatePublisher(input: BookStorePublisherInput(name_: v.str('input.name'), address: v.str('input.address')))),
  ),
  AdminOp(
    'bookStoreUpdatePublisher',
    resource: 'publisher',
    verb: 'update',
    mutation: true,
    fields: [
      KField('id', 'ID', kind: FieldKind.number, required: true),
      KField('input.name', 'Name'),
      KField('input.address', 'Address'),
    ],
    run: (api, v) async => adminPlain(await api.bookStore.bookStoreUpdatePublisher(id: v.need(v.integer('id'), 'id'), input: BookStorePublisherInput(name_: v.str('input.name'), address: v.str('input.address')))),
  ),
  AdminOp(
    'bookStoreDeletePublisher',
    resource: 'publisher',
    verb: 'delete',
    mutation: true,
    destructive: true,
    fields: [
      KField('id', 'ID', kind: FieldKind.number, required: true),
    ],
    run: (api, v) async => adminPlain(await api.bookStore.bookStoreDeletePublisher(id: v.need(v.integer('id'), 'id'))),
  ),
  AdminOp(
    'bookStoreCreateCategory',
    resource: 'category',
    verb: 'create',
    mutation: true,
    fields: [
      KField('input.name', 'Name'),
    ],
    run: (api, v) async => adminPlain(await api.bookStore.bookStoreCreateCategory(input: BookStoreCategoryInput(name_: v.str('input.name')))),
  ),
  AdminOp(
    'bookStoreUpdateCategory',
    resource: 'category',
    verb: 'update',
    mutation: true,
    fields: [
      KField('id', 'ID', kind: FieldKind.number, required: true),
      KField('input.name', 'Name'),
    ],
    run: (api, v) async => adminPlain(await api.bookStore.bookStoreUpdateCategory(id: v.need(v.integer('id'), 'id'), input: BookStoreCategoryInput(name_: v.str('input.name')))),
  ),
  AdminOp(
    'bookStoreDeleteCategory',
    resource: 'category',
    verb: 'delete',
    mutation: true,
    destructive: true,
    fields: [
      KField('id', 'ID', kind: FieldKind.number, required: true),
    ],
    run: (api, v) async => adminPlain(await api.bookStore.bookStoreDeleteCategory(id: v.need(v.integer('id'), 'id'))),
  ),
];
