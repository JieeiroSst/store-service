// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

// ignore_for_file: unused_import

import '../../../apps/kit/kit.dart';
import '../../../../gen/models.g.dart';
import '../admin_op.dart';

final bookAdminOps = <AdminOp>[
  AdminOp(
    'bookCreateBook',
    resource: 'book',
    verb: 'create',
    mutation: true,
    fields: [
      KField('input.title', 'Title'),
      KField('input.author', 'Author'),
      KField('input.isbn', 'Isbn'),
      KField('input.description', 'Description', kind: FieldKind.multiline),
      KField('input.publishedYear', 'Published year', kind: FieldKind.number),
      KField('input.categories', 'Categories', hint: 'Nhiều giá trị, cách nhau bởi dấu phẩy'),
    ],
    run: (api, v) async => adminPlain(await api.book.bookCreateBook(input: BookBookInput(title: v.str('input.title'), author: v.str('input.author'), isbn: v.str('input.isbn'), description: v.str('input.description'), publishedYear: v.integer('input.publishedYear'), categories: v.strs('input.categories')))),
  ),
  AdminOp(
    'bookUpdateBook',
    resource: 'book',
    verb: 'update',
    mutation: true,
    fields: [
      KField('id', 'ID', kind: FieldKind.number, required: true),
      KField('input.title', 'Title'),
      KField('input.author', 'Author'),
      KField('input.isbn', 'Isbn'),
      KField('input.description', 'Description', kind: FieldKind.multiline),
      KField('input.publishedYear', 'Published year', kind: FieldKind.number),
      KField('input.categories', 'Categories', hint: 'Nhiều giá trị, cách nhau bởi dấu phẩy'),
    ],
    run: (api, v) async => adminPlain(await api.book.bookUpdateBook(id: v.need(v.integer('id'), 'id'), input: BookBookInput(title: v.str('input.title'), author: v.str('input.author'), isbn: v.str('input.isbn'), description: v.str('input.description'), publishedYear: v.integer('input.publishedYear'), categories: v.strs('input.categories')))),
  ),
  AdminOp(
    'bookDeleteBook',
    resource: 'book',
    verb: 'delete',
    mutation: true,
    destructive: true,
    fields: [
      KField('id', 'ID', kind: FieldKind.number, required: true),
    ],
    run: (api, v) async => adminPlain(await api.book.bookDeleteBook(id: v.need(v.integer('id'), 'id'))),
  ),
  AdminOp(
    'bookPutCover',
    resource: 'cover',
    verb: 'update',
    mutation: true,
    fields: [
      KField('id', 'ID', kind: FieldKind.number, required: true),
      KField('file', 'File', kind: FieldKind.file, required: true),
    ],
    run: (api, v) async => adminPlain(await api.book.bookPutCover(id: v.need(v.integer('id'), 'id'), file: v.need(v.file('file'), 'file'))),
  ),
  AdminOp(
    'bookDeleteCover',
    resource: 'cover',
    verb: 'delete',
    mutation: true,
    destructive: true,
    fields: [
      KField('id', 'ID', kind: FieldKind.number, required: true),
    ],
    run: (api, v) async => adminPlain(await api.book.bookDeleteCover(id: v.need(v.integer('id'), 'id'))),
  ),
  AdminOp(
    'bookPutChapter',
    resource: 'chapter',
    verb: 'update',
    mutation: true,
    fields: [
      KField('id', 'ID', kind: FieldKind.number, required: true),
      KField('number', 'Number', kind: FieldKind.number, required: true),
      KField('input.title', 'Title'),
      KField('input.content', 'Content', kind: FieldKind.multiline),
    ],
    run: (api, v) async => adminPlain(await api.book.bookPutChapter(id: v.need(v.integer('id'), 'id'), number: v.need(v.integer('number'), 'number'), input: BookChapterInput(title: v.str('input.title'), content: v.str('input.content')))),
  ),
  AdminOp(
    'bookDeleteChapter',
    resource: 'chapter',
    verb: 'delete',
    mutation: true,
    destructive: true,
    fields: [
      KField('id', 'ID', kind: FieldKind.number, required: true),
      KField('number', 'Number', kind: FieldKind.number, required: true),
    ],
    run: (api, v) async => adminPlain(await api.book.bookDeleteChapter(id: v.need(v.integer('id'), 'id'), number: v.need(v.integer('number'), 'number'))),
  ),
  AdminOp(
    'bookCreateCategory',
    resource: 'category',
    verb: 'create',
    mutation: true,
    fields: [
      KField('input.slug', 'Slug'),
      KField('input.name', 'Name'),
    ],
    run: (api, v) async => adminPlain(await api.book.bookCreateCategory(input: BookCategoryInput(slug: v.str('input.slug'), name_: v.str('input.name')))),
  ),
  AdminOp(
    'bookRenameCategory',
    resource: 'category',
    verb: 'rename',
    mutation: true,
    fields: [
      KField('slug', 'Slug', required: true),
      KField('name', 'Name', required: true),
    ],
    run: (api, v) async => adminPlain(await api.book.bookRenameCategory(slug: v.need(v.str('slug'), 'slug'), name_: v.need(v.str('name'), 'name'))),
  ),
  AdminOp(
    'bookDeleteCategory',
    resource: 'category',
    verb: 'delete',
    mutation: true,
    destructive: true,
    fields: [
      KField('slug', 'Slug', required: true),
    ],
    run: (api, v) async => adminPlain(await api.book.bookDeleteCategory(slug: v.need(v.str('slug'), 'slug'))),
  ),
];
