import 'package:flutter/material.dart';

import '../../../gen/models.g.dart';
import '../extras/extras.dart';
import '../kit/kit.dart';

class BookApp extends StatelessWidget {
  const BookApp({super.key});

  @override
  Widget build(BuildContext context) {
    final api = context.api.book;
    return ServiceApp(
      serviceId: 'book',
      tabs: [
        AppTab('Thư viện', Icons.menu_book_outlined, (context) => SearchableList<BookBook>(
              hint: 'Tên sách, tác giả…',
              filters: () async => (await api.bookCategories())?.map((c) => c.slug ?? '').where((s) => s.isNotEmpty).toList(),
              load: (q, cat) async => cat == null
                  ? (await api.bookBooks(q: q.isEmpty ? null : q, limit: 40))?.data
                  : (await api.bookCategoryBooks(slug: cat, q: q.isEmpty ? null : q, limit: 40))?.data,
              gridExtent: 180,
              gridAspect: 0.56,
              emptyMessage: 'Không có sách phù hợp',
              emptyIcon: Icons.menu_book_outlined,
              itemBuilder: (context, b, _) => PosterCard(
                title: b.title ?? 'Sách',
                subtitle: b.author,
                imageUrl: b.coverUrl,
                icon: Icons.auto_stories_rounded,
                onTap: () => openPage(context, BookDetailPage(id: b.id!)),
              ),
            )),
      ],
    );
  }
}

class BookDetailPage extends StatelessWidget {
  const BookDetailPage({super.key, required this.id});

  final int id;

  @override
  Widget build(BuildContext context) {
    final api = context.api.book;
    return AsyncBuilder<(BookBook?, List<BookChapter>)>(
      load: () async {
        final r = await Future.wait<Object?>([api.bookBook(id: id), api.bookChapters(id: id, limit: 200, selectionSet: '{ data { bookId number title sizeBytes } pagination { limit nextCursor hasMore } }')]);
        return (r[0] as BookBook?, (r[1] as BookChapterPage?)?.data ?? const <BookChapter>[]);
      },
      loading: const Scaffold(body: SkeletonList()),
      builder: (context, d, _) {
        final b = d.$1;
        final chapters = d.$2;
        return DetailPage(
          title: b?.title ?? 'Sách',
          actions: [ShareLinkButton(path: '/book/$id', title: b?.title ?? 'Sách')],
          header: Padding(
            padding: const EdgeInsets.fromLTRB(16, 16, 16, 0),
            child: Row(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                KThumb(url: b?.coverUrl, icon: Icons.auto_stories_rounded, width: 110, height: 160, radius: 14),
                const SizedBox(width: 16),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(b?.title ?? '', style: Theme.of(context).textTheme.titleLarge?.copyWith(fontWeight: FontWeight.w800)),
                      const SizedBox(height: 4),
                      Text(b?.author ?? '', style: Theme.of(context).textTheme.bodyLarge),
                      const SizedBox(height: 8),
                      Wrap(spacing: 6, runSpacing: 6, children: [
                        for (final c in b?.categories ?? const <BookCategory>[])
                          GestureDetector(
                            onTap: c.slug == null
                                ? null
                                : () async {
                                    final cat = await context.api.book.bookCategory(slug: c.slug!).catchError((_) => null);
                                    if (context.mounted && cat != null) showSnack(context, 'Thể loại ${cat.name_} (${cat.slug})');
                                  },
                            child: Pill(c.name_ ?? c.slug ?? ''),
                          ),
                        if (b?.publishedYear != null) Pill('${b!.publishedYear}'),
                        Pill('${b?.chapterCount ?? chapters.length} chương', icon: Icons.list),
                      ]),
                    ],
                  ),
                ),
              ],
            ),
          ),
          bottom: FilledButton.icon(
            icon: const Icon(Icons.chrome_reader_mode_outlined),
            label: const Text('Đọc từ đầu'),
            onPressed: chapters.isEmpty ? null : () => openPage(context, ChapterPage(bookId: id, number: chapters.first.number ?? 1, total: chapters.length)),
          ),
          children: [
            Paragraph(b?.description, title: 'Giới thiệu'),
            InfoSection(rows: [('ISBN', b?.isbn)]),
            const SectionHeader('Mục lục'),
            for (final c in chapters)
              ListTile(
                contentPadding: const EdgeInsets.symmetric(horizontal: 20),
                leading: CircleAvatar(radius: 16, backgroundColor: context.appColor.withValues(alpha: 0.12), child: Text('${c.number}', style: const TextStyle(fontSize: 12))),
                title: Text(c.title ?? 'Chương ${c.number}'),
                trailing: const Icon(Icons.chevron_right),
                onTap: () => openPage(context, ChapterPage(bookId: id, number: c.number ?? 1, total: chapters.length)),
              ),
          ],
        );
      },
    );
  }
}

class ChapterPage extends StatefulWidget {
  const ChapterPage({super.key, required this.bookId, required this.number, required this.total});

  final int bookId;
  final int number;
  final int total;

  @override
  State<ChapterPage> createState() => _ChapterPageState();
}

class _ChapterPageState extends State<ChapterPage> {
  late int _n = widget.number;
  double _size = 17;

  @override
  Widget build(BuildContext context) {
    return AsyncBuilder<BookChapter?>(
      key: ValueKey(_n),
      load: () => context.api.book.bookChapter(id: widget.bookId, number: _n),
      loading: const Scaffold(body: SkeletonList()),
      builder: (context, c, _) => DetailPage(
        title: c?.title ?? 'Chương $_n',
        actions: [
          IconButton(tooltip: 'Chữ nhỏ', onPressed: () => setState(() => _size = (_size - 1).clamp(13, 26)), icon: const Icon(Icons.text_decrease)),
          IconButton(tooltip: 'Chữ lớn', onPressed: () => setState(() => _size = (_size + 1).clamp(13, 26)), icon: const Icon(Icons.text_increase)),
        ],
        bottom: Row(
          children: [
            Expanded(child: OutlinedButton(onPressed: _n > 1 ? () => setState(() => _n--) : null, child: const Text('Chương trước'))),
            const SizedBox(width: 8),
            Expanded(child: FilledButton(onPressed: _n < widget.total ? () => setState(() => _n++) : null, child: const Text('Chương sau'))),
          ],
        ),
        children: [
          Padding(
            padding: const EdgeInsets.fromLTRB(20, 20, 20, 0),
            child: SelectableText(c?.content ?? 'Chương này chưa có nội dung', style: TextStyle(fontSize: _size, height: 1.7)),
          ),
        ],
      ),
    );
  }
}

class BookStoreApp extends StatelessWidget {
  const BookStoreApp({super.key});

  @override
  Widget build(BuildContext context) {
    final api = context.api.bookStore;
    return ServiceApp(
      serviceId: 'book_store',
      tabs: [
        AppTab('Cửa hàng', Icons.storefront_outlined, (context) => AsyncList<BookStoreBook>(
              load: () => api.bookStoreBooks(),
              header: AsyncBuilder<(BookStoreBook?, BookStoreAuthorStat?)>(
                load: () async {
                  final r = await Future.wait<Object?>([
                    api.bookStoreMostPurchasedBook().catchError((_) => null),
                    api.bookStoreMostReadAuthor().catchError((_) => null),
                  ]);
                  return (r[0] as BookStoreBook?, r[1] as BookStoreAuthorStat?);
                },
                loading: const SizedBox(height: 8),
                builder: (context, d, _) => d.$1 == null
                    ? const SizedBox.shrink()
                    : AppBanner(
                        icon: Icons.local_fire_department_rounded,
                        title: d.$1!.name_ ?? '',
                        subtitle: 'Bán chạy nhất • ${d.$1!.purchaseCount ?? 0} lượt mua${d.$2 == null ? '' : '\nTác giả được đọc nhiều: ${d.$2!.authorName}'}',
                      ),
              ),
              emptyMessage: 'Cửa hàng chưa có sách',
              emptyIcon: Icons.menu_book_outlined,
              itemBuilder: (context, b, reload) => EntityTile(
                title: b.name_ ?? 'Sách',
                subtitle: [b.author?.name_, b.category?.name_].whereType<String>().join(' • '),
                icon: Icons.menu_book_rounded,
                caption: vnd(b.price),
                badges: [if ((b.quantity ?? 0) <= 0) const StatusPill('out_of_stock') else Pill('Còn ${b.quantity}')],
                onTap: () async {
                  final fresh = await api.bookStoreBook(id: b.id!).catchError((_) => b);
                  if (!context.mounted) return;
                  await openPage(context, _StoreBookPage(book: fresh ?? b));
                  await reload();
                },
              ),
            )),
        AppTab('Khám phá', Icons.explore_outlined, (_) => const BookStoreBrowse()),
        AppTab('Tác giả', Icons.person_outline, (context) => AsyncList<BookStoreAuthor>(
              load: () => api.bookStoreAuthors(),
              emptyMessage: 'Chưa có tác giả',
              itemBuilder: (context, a, _) => EntityTile(title: a.name_ ?? '', subtitle: a.bio, icon: Icons.person_rounded, onTap: () => bookStoreAuthorInfo(context, a.id!)),
            )),
        AppTab('Thể loại', Icons.category_outlined, (context) => AsyncList<BookStoreCategoryPriceStat>(
              load: () => api.bookStoreCategoryPriceStats(),
              emptyMessage: 'Chưa có thể loại',
              itemBuilder: (context, c, _) => EntityTile(
                title: c.categoryName ?? '',
                subtitle: '${c.bookCount ?? 0} đầu sách • ${vnd(c.minPrice)} – ${vnd(c.maxPrice)}',
                icon: Icons.category_rounded,
                caption: 'Giá TB ${vnd(c.avgPrice)}',
              ),
            )),
      ],
    );
  }
}

class _StoreBookPage extends StatelessWidget {
  const _StoreBookPage({required this.book});

  final BookStoreBook book;

  @override
  Widget build(BuildContext context) {
    final api = context.api.bookStore;
    final b = book;
    return DetailPage(
      title: b.name_ ?? 'Sách',
      actions: [if (b.id != null) ShareLinkButton(path: '/store-book/${b.id}', title: b.name_ ?? 'Sách')],
      header: HeroHeader(title: b.name_ ?? 'Sách', subtitle: b.author?.name_, icon: Icons.menu_book_rounded, badges: [if (b.category?.name_ != null) Pill(b.category!.name_!)]),
      bottom: Row(
        children: [
          Expanded(
            child: OutlinedButton.icon(
              icon: const Icon(Icons.chrome_reader_mode_outlined),
              label: const Text('Đọc thử'),
              onPressed: () => runAction(context, () => api.bookStoreReadBook(id: b.id!), success: 'Đã ghi nhận lượt đọc'),
            ),
          ),
          const SizedBox(width: 8),
          Expanded(
            child: FilledButton.icon(
              icon: const Icon(Icons.shopping_bag_outlined),
              label: Text(vnd(b.price)),
              onPressed: (b.quantity ?? 0) <= 0
                  ? null
                  : () => showFormSheet(
                        context,
                        title: 'Mua “${b.name_}”',
                        fields: const [KField('qty', 'Số lượng', kind: FieldKind.number, required: true, initial: '1')],
                        submitLabel: 'Mua ngay',
                        success: 'Mua sách thành công',
                        onSubmit: (v) => api.bookStorePurchaseBook(id: b.id!, quantity: v.integer('qty') ?? 1),
                      ),
            ),
          ),
        ],
      ),
      children: [
        InfoSection(rows: [
          ('Tác giả', b.author?.name_),
          ('Nhà xuất bản', b.publisher?.name_),
          ('Thể loại', b.category?.name_),
          ('Giá', vnd(b.price)),
          ('Tồn kho', '${b.quantity ?? 0} ${b.unitOfMeasurement ?? ''}'),
          ('Lượt mua', '${b.purchaseCount ?? 0}'),
          ('Lượt đọc', '${b.readCount ?? 0}'),
        ]),
        Paragraph(b.author?.bio, title: 'Về tác giả'),
      ],
    );
  }
}

class StoreBookLinkPage extends StatelessWidget {
  const StoreBookLinkPage({super.key, required this.id});

  final int id;

  @override
  Widget build(BuildContext context) => AsyncBuilder<BookStoreBook?>(
        load: () => context.api.bookStore.bookStoreBook(id: id),
        loading: const Scaffold(body: SkeletonList()),
        builder: (context, b, _) => b == null ? const DetailPage(title: 'Sách', children: [EmptyView(message: 'Không tìm thấy sách')]) : _StoreBookPage(book: b),
      );
}
