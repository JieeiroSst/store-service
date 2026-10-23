import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:flutter/services.dart';

import '../../../gen/models.g.dart';
import '../kit/kit.dart';
import '../../../core/tracking/attribution_cubit.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

Future<void> _copy(BuildContext context, String text) async {
  await Clipboard.setData(ClipboardData(text: text));
  if (context.mounted) showSnack(context, 'Đã sao chép');
}

String _json(Object? data) => const JsonEncoder.withIndent('  ').convert(data);

class JsonPage extends StatelessWidget {
  const JsonPage({super.key, required this.title, required this.data});

  final String title;
  final Object? data;

  @override
  Widget build(BuildContext context) {
    return DetailPage(
      title: title,
      actions: [IconButton(tooltip: 'Sao chép', icon: const Icon(Icons.copy), onPressed: () => _copy(context, _json(data)))],
      children: [
        Padding(
          padding: const EdgeInsets.all(16),
          child: KCard(child: SelectableText(_json(data), style: const TextStyle(fontFamily: 'monospace', fontSize: 12))),
        ),
      ],
    );
  }
}

class ThreadsProfilePage extends StatefulWidget {
  const ThreadsProfilePage({super.key, required this.author});

  final ThreadsAuthor author;

  @override
  State<ThreadsProfilePage> createState() => _ThreadsProfilePageState();
}

class _ThreadsProfilePageState extends State<ThreadsProfilePage> {
  bool? _following;

  @override
  Widget build(BuildContext context) {
    final api = context.api.threads;
    final a = widget.author;
    final id = a.id!;
    return AsyncBuilder<(ThreadsAuthorPage?, ThreadsAuthorPage?, ThreadsPostPage?)>(
      load: () async {
        final r = await Future.wait<Object?>([
          api.threadsFollowers(id: id, limit: 50).catchError((_) => null),
          api.threadsFollowing(id: id, limit: 50).catchError((_) => null),
          api.threadsPosts(authorId: id, limit: 30).catchError((_) => null),
        ]);
        return (r[0] as ThreadsAuthorPage?, r[1] as ThreadsAuthorPage?, r[2] as ThreadsPostPage?);
      },
      loading: const Scaffold(body: SkeletonList()),
      builder: (context, d, reload) {
        final followers = d.$1?.rows ?? const <ThreadsAuthor>[];
        final following = _following ?? followers.any((f) => f.id == '${context.userId}');
        final mine = id == '${context.userId}';
        return DetailPage(
          title: '@${a.username ?? ''}',
          header: AppBanner(
            icon: Icons.person_rounded,
            title: a.username ?? '',
            subtitle: '${followers.length} người theo dõi • đang theo dõi ${(d.$2?.rows ?? const []).length}',
            children: [
              if (!mine)
                SizedBox(
                  width: double.infinity,
                  child: FilledButton(
                    style: FilledButton.styleFrom(backgroundColor: Colors.white, foregroundColor: context.appColor),
                    onPressed: () async {
                      final follow = !following;
                      if (await runAction(context, () => follow ? api.threadsFollowUser(id: id) : api.threadsUnfollowUser(id: id), success: follow ? 'Đã theo dõi' : 'Đã bỏ theo dõi')) {
                        setState(() => _following = follow);
                      }
                    },
                    child: Text(following ? 'Đang theo dõi' : 'Theo dõi'),
                  ),
                ),
            ],
          ),
          children: [
            const SectionHeader('Bài viết'),
            for (final p in d.$3?.rows ?? const <ThreadsPost>[])
              Padding(
                padding: const EdgeInsets.fromLTRB(16, 0, 16, 8),
                child: EntityTile(title: p.content ?? '', subtitle: '${p.likeCount ?? 0} thích • ${p.commentCount ?? 0} bình luận', icon: Icons.notes_rounded),
              ),
            const SectionHeader('Người theo dõi'),
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: 16),
              child: Wrap(spacing: 8, runSpacing: 8, children: [for (final f in followers) Chip(avatar: const Icon(Icons.person, size: 16), label: Text(f.username ?? ''))]),
            ),
          ],
        );
      },
    );
  }
}

Future<void> threadsCommentMenu(BuildContext context, ThreadsComment c, Future<void> Function() reload) async {
  final api = context.api.threads;
  final mine = c.userId == '${context.userId}';
  final action = await showModalBottomSheet<String>(
    context: context,
    builder: (_) => SafeArea(
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          ListTile(leading: const Icon(Icons.favorite_border), title: const Text('Thích'), onTap: () => Navigator.pop(context, 'like')),
          ListTile(leading: const Icon(Icons.heart_broken_outlined), title: const Text('Bỏ thích'), onTap: () => Navigator.pop(context, 'unlike')),
          if (mine) ListTile(leading: const Icon(Icons.delete_outline), title: const Text('Xoá bình luận'), onTap: () => Navigator.pop(context, 'delete')),
        ],
      ),
    ),
  );
  if (action == null || !context.mounted) return;
  final Future<bool> pending = switch (action) {
    'like' => runAction(context, () => api.threadsLikeComment(id: c.id!), success: 'Đã thích'),
    'unlike' => runAction(context, () => api.threadsUnlikeComment(id: c.id!)),
    _ => runAction(context, () => api.threadsDeleteComment(id: c.id!), confirm: 'Xoá bình luận?', destructive: true, success: 'Đã xoá'),
  };
  if (await pending) await reload();
}

class BookStoreBrowse extends StatelessWidget {
  const BookStoreBrowse({super.key});

  @override
  Widget build(BuildContext context) {
    final api = context.api.bookStore;
    return AsyncBuilder<(List<BookStorePublisher>, List<BookStoreCategory>, BookStoreBook?, BookStoreAuthorStat?)>(
      load: () async {
        final r = await Future.wait<Object?>([
          api.bookStorePublishers().catchError((_) => null),
          api.bookStoreCategories().catchError((_) => null),
          api.bookStoreHighestPriceBook().catchError((_) => null),
          api.bookStoreMostPurchasedAuthor().catchError((_) => null),
        ]);
        return ((r[0] as List<BookStorePublisher>?) ?? const <BookStorePublisher>[], (r[1] as List<BookStoreCategory>?) ?? const <BookStoreCategory>[], r[2] as BookStoreBook?, r[3] as BookStoreAuthorStat?);
      },
      builder: (context, d, reload) => RefreshIndicator(
        onRefresh: reload,
        child: ListView(
          padding: const EdgeInsets.only(bottom: 32),
          children: [
            InfoSection(title: 'Kỷ lục', rows: [
              ('Sách đắt nhất', d.$3 == null ? null : '${d.$3!.name_} • ${vnd(d.$3!.price)}'),
              ('Tác giả bán chạy', d.$4 == null ? null : '${d.$4!.authorName} • ${d.$4!.totalPurchases ?? 0} lượt mua'),
            ]),
            const SectionHeader('Nhà xuất bản'),
            for (final p in d.$1)
              ListTile(
                contentPadding: const EdgeInsets.symmetric(horizontal: 20),
                leading: const Icon(Icons.business_outlined),
                title: Text(p.name_ ?? ''),
                subtitle: Text(p.address ?? ''),
                onTap: () async {
                  final full = await api.bookStorePublisher(id: p.id!).catchError((_) => null);
                  if (context.mounted && full != null) await showResult(context, icon: Icons.business, title: full.name_ ?? '', message: full.address ?? '');
                },
              ),
            const SectionHeader('Thể loại'),
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: 16),
              child: Wrap(
                spacing: 8,
                runSpacing: 8,
                children: [
                  for (final c in d.$2)
                    ActionChip(
                      label: Text(c.name_ ?? ''),
                      onPressed: () async {
                        final full = await api.bookStoreCategory(id: c.id!).catchError((_) => null);
                        final books = (await api.bookStoreBooks().catchError((_) => null) ?? const <BookStoreBook>[]).where((b) => b.categoryId == c.id).toList();
                        if (!context.mounted) return;
                        await showResult(context, icon: Icons.category, title: full?.name_ ?? c.name_ ?? '', message: books.isEmpty ? 'Chưa có sách' : books.map((b) => '• ${b.name_} — ${vnd(b.price)}').join('\n'));
                      },
                    ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }
}

Future<void> bookStoreAuthorInfo(BuildContext context, int authorId) async {
  final a = await context.api.bookStore.bookStoreAuthor(id: authorId).catchError((_) => null);
  if (context.mounted && a != null) await showResult(context, icon: Icons.person, title: a.name_ ?? '', message: a.bio ?? '');
}

class CarStaffPage extends StatelessWidget {
  const CarStaffPage({super.key});

  @override
  Widget build(BuildContext context) {
    final api = context.api.carRental;
    final staff = '${context.userId}';
    return DetailPage(
      title: 'Quầy giao xe',
      header: const AppBanner(icon: Icons.badge_rounded, title: 'Dành cho nhân viên', subtitle: 'Giao xe theo mã đặt xe, nhận lại xe và tính phụ phí'),
      children: [
        ActionButtons([
          FilledButton.icon(
            icon: const Icon(Icons.key),
            label: const Text('Giao xe'),
            onPressed: () async {
              CarRentalRentalResponse? r;
              await showFormSheet(
                context,
                title: 'Giao xe cho khách',
                fields: const [KField('resv', 'Mã đặt xe', required: true, scannable: true), KField('km', 'Số km lúc giao', kind: FieldKind.number, required: true)],
                submitLabel: 'Giao xe',
                onSubmit: (v) async {
                  final resv = await api.carRentalGetReservation(input: CarRentalGetReservationRequestInput(reservationId: v.str('resv')));
                  r = await api.carRentalStartRental(
                    input: CarRentalStartRentalRequestInput(reservationId: v.str('resv'), pickupMileage: v.number('km'), staffId: staff, pickupLocationId: resv?.pickupLocationId),
                  );
                },
              );
              if (r != null && context.mounted) await showResult(context, title: 'Đã giao xe', message: 'Mã chuyến ${r!.rentalId}\n${r!.vehicle?.make ?? ''} ${r!.vehicle?.model ?? ''}');
            },
          ),
          OutlinedButton.icon(
            icon: const Icon(Icons.assignment_return),
            label: const Text('Nhận lại xe'),
            onPressed: () async {
              CarRentalRentalResponse? r;
              await showFormSheet(
                context,
                title: 'Nhận lại xe',
                fields: const [
                  KField('rental', 'Mã chuyến thuê', required: true, scannable: true),
                  KField('km', 'Số km lúc trả', kind: FieldKind.number, required: true),
                  KField('fuel', 'Phụ phí nhiên liệu (₫)', kind: FieldKind.number),
                  KField('damage', 'Phụ phí hư hỏng (₫)', kind: FieldKind.number),
                  KField('notes', 'Ghi chú', kind: FieldKind.multiline),
                ],
                submitLabel: 'Hoàn tất',
                onSubmit: (v) async {
                  final cur = await api.carRentalGetRental(input: CarRentalGetRentalRequestInput(rentalId: v.str('rental')));
                  r = await api.carRentalCompleteRental(
                    input: CarRentalCompleteRentalRequestInput(
                      rentalId: v.str('rental'),
                      returnMileage: v.number('km'),
                      staffId: staff,
                      returnLocationId: cur?.returnLocationId,
                      additionalFees: {if (v.number('fuel') != null) 'fuel': v.number('fuel'), if (v.number('damage') != null) 'damage': v.number('damage')},
                      notes: v.str('notes'),
                    ),
                  );
                },
              );
              if (r != null && context.mounted) await showResult(context, title: 'Đã nhận xe', message: 'Tổng ${vnd((r!.baseFee ?? 0) + (r!.additionalFees ?? 0))}');
            },
          ),
        ]),
      ],
    );
  }
}

Future<void> carReservationDetail(BuildContext context, String reservationId, Future<void> Function() reload) async {
  final api = context.api.carRental;
  final r = await api.carRentalGetReservation(input: CarRentalGetReservationRequestInput(reservationId: reservationId)).catchError((_) => null);
  if (!context.mounted || r == null) return;
  final change = await showDialog<bool>(
    context: context,
    builder: (context) => AlertDialog(
      title: Text('${r.vehicle?.make ?? ''} ${r.vehicle?.model ?? ''}'),
      content: Text('Nhận ${formatDate(r.startTime)}\nTrả ${formatDate(r.endTime)}\nTrạng thái ${r.status?.value ?? ''}'),
      actions: [
        TextButton(onPressed: () => Navigator.pop(context, true), child: const Text('Đổi ngày')),
        FilledButton(onPressed: () => Navigator.pop(context, false), child: const Text('Đóng')),
      ],
    ),
  );
  if (change != true || !context.mounted) return;
  if (await showFormSheet(
    context,
    title: 'Đổi lịch thuê xe',
    fields: [
      KField('start', 'Nhận xe', kind: FieldKind.datetime, required: true, initial: r.startTime),
      KField('end', 'Trả xe', kind: FieldKind.datetime, required: true, initial: r.endTime),
    ],
    success: 'Đã đổi lịch',
    onSubmit: (v) => api.carRentalUpdateReservation(
      input: CarRentalUpdateReservationRequestInput(reservationId: reservationId, startTime: v.str('start'), endTime: v.str('end'), pickupLocationId: r.pickupLocationId, returnLocationId: r.returnLocationId),
    ),
  )) {
    await reload();
  }
}

Future<void> carRentalDetail(BuildContext context, String rentalId) async {
  final r = await context.api.carRental.carRentalGetRental(input: CarRentalGetRentalRequestInput(rentalId: rentalId)).catchError((_) => null);
  if (!context.mounted || r == null) return;
  await showResult(
    context,
    icon: Icons.route,
    title: 'Chuyến ${r.rentalId}',
    message: 'Km lúc nhận ${r.pickupMileage ?? '-'} • lúc trả ${r.returnMileage ?? '-'}\nPhí thuê ${vnd(r.baseFee)} • phụ phí ${vnd(r.additionalFees)}',
  );
}

Future<CarRentalVehicleResponse?> carVehicle(BuildContext context, String id) => context.api.carRental.carRentalGetVehicle(input: CarRentalGetVehicleRequestInput(vehicleId: id)).catchError((_) => null);

Future<List<CarRentalVehicleResponse>?> carAvailable(BuildContext context, DateTime start, DateTime end) async =>
    (await context.api.carRental.carRentalSearchAvailableVehicles(input: CarRentalSearchVehiclesRequestInput(startTime: start.toUtc().toIso8601String(), endTime: end.toUtc().toIso8601String(), pageSize: 50)))?.vehicles;

class ReferralToolsPage extends StatelessWidget {
  const ReferralToolsPage({super.key});

  @override
  Widget build(BuildContext context) {
    final api = context.api.referral;
    return AsyncBuilder<(ReferralAttribution?, List<ReferralUTMPerformance>)>(
      load: () async {
        final r = await Future.wait<Object?>([api.myReferralAttribution().catchError((_) => null), api.myReferralUTMReport().catchError((_) => null)]);
        return (r[0] as ReferralAttribution?, (r[1] as List<ReferralUTMPerformance>?) ?? const <ReferralUTMPerformance>[]);
      },
      loading: const Scaffold(body: SkeletonList()),
      builder: (context, d, reload) => DetailPage(
        title: 'Công cụ giới thiệu',
        children: [
          InfoSection(title: 'Bạn được mời bởi', rows: [
            ('Mã mời', d.$1?.refCode),
            ('Người mời', d.$1?.ownerUserId),
            ('Nền tảng', d.$1?.platform),
            ('Chiến dịch', d.$1?.utmCampaign),
          ]),
          if (d.$1?.refCode == null)
            ActionButtons([
              FilledButton.icon(
                icon: const Icon(Icons.redeem),
                label: const Text('Nhập mã mời'),
                onPressed: () async {
                  if (await showFormSheet(
                    context,
                    title: 'Mã mời của bạn bè',
                    fields: const [KField('code', 'Mã', required: true)],
                    success: 'Đã nhận lời mời, phần thưởng sẽ được cộng',
                    onSubmit: (v) async {
                      final r = await context.read<AttributionCubit>().applyManual(v.str('code')!);
                      if (r != null && r.ok && !r.attributed) throw Exception('Tài khoản đã được ghi nhận nguồn mời trước đó');
                    },
                  )) {
                    await reload();
                  }
                },
              ),
            ]),
          ActionButtons([
            OutlinedButton.icon(
              icon: const Icon(Icons.search),
              label: const Text('Tra cứu mã mời'),
              onPressed: () async {
                ReferralReferralLink? link;
                ReferralReferralStatusResponse? status;
                List<ReferralShareTarget>? targets;
                await showFormSheet(
                  context,
                  title: 'Mã mời',
                  fields: const [KField('code', 'Mã', required: true)],
                  submitLabel: 'Tra cứu',
                  onSubmit: (v) async {
                    final code = v.str('code')!.toUpperCase();
                    link = await api.referralLink(refCode: code);
                    status = await api.referralStatus(refCode: code).catchError((_) => null);
                    targets = await api.referralShareTargets(refCode: code).catchError((_) => null);
                  },
                );
                if (link == null || !context.mounted) return;
                await openPage(
                  context,
                  DetailPage(
                    title: 'Mã ${link!.refCode}',
                    children: [
                      InfoSection(rows: [
                        ('Link', link!.shareUrl ?? link!.deepLink),
                        ('Kênh', link!.channel),
                        ('Trạng thái', status?.status ?? link!.status),
                        ('Lời mời', status?.invitationStatus),
                        ('Thưởng', status?.rewardStatus == null ? null : '${status!.rewardStatus} • ${vnd(status!.rewardValue)}'),
                      ]),
                      const SectionHeader('Chia sẻ qua'),
                      for (final t in targets ?? const <ReferralShareTarget>[])
                        ListTile(
                          leading: const Icon(Icons.share),
                          title: Text(t.label ?? t.channel ?? ''),
                          subtitle: Text(t.shareUrl ?? '', maxLines: 1, overflow: TextOverflow.ellipsis),
                          onTap: () async {
                            await api.referralTrackEvent(refCode: link!.refCode!, eventType: 'link_copied', channel: t.channel).catchError((_) => null);
                            if (context.mounted) await _copy(context, t.shareUrl ?? '');
                          },
                        ),
                    ],
                  ),
                );
              },
            ),
          ]),
          const SectionHeader('Hiệu quả theo chiến dịch'),
          if (d.$2.isEmpty) const Padding(padding: EdgeInsets.symmetric(horizontal: 20), child: Text('Chưa có dữ liệu')),
          for (final u in d.$2)
            ListTile(
              contentPadding: const EdgeInsets.symmetric(horizontal: 20),
              title: Text([u.utmSource, u.utmMedium, u.utmCampaign].whereType<String>().join(' / ')),
              subtitle: Text('${u.clicks ?? 0} lượt bấm • ${u.installs ?? 0} lượt cài • ${u.rewards ?? 0} thưởng'),
              trailing: Text('${((u.conversionRate ?? 0) * 100).toStringAsFixed(1)}%', style: const TextStyle(fontWeight: FontWeight.w800)),
            ),
        ],
      ),
    );
  }
}

class SerpToolsPage extends StatelessWidget {
  const SerpToolsPage({super.key});

  @override
  Widget build(BuildContext context) {
    final api = context.api.serpapi;
    return AsyncBuilder<Object?>(
      load: () => api.serpEngines().catchError((_) => null),
      loading: const Scaffold(body: SkeletonList()),
      builder: (context, engines, _) {
        final list = engines is List ? engines.map((e) => e is Map ? '${e['name'] ?? e['engine'] ?? e}' : '$e').toList() : (engines is Map ? engines.keys.map((k) => '$k').toList() : const <String>[]);
        return DetailPage(
          title: 'Công cụ tìm kiếm',
          children: [
            ActionButtons([
              FilledButton.icon(
                icon: const Icon(Icons.image_search),
                label: const Text('Tìm bằng hình ảnh'),
                onPressed: () async {
                  Object? r;
                  await showFormSheet(context, title: 'Tìm bằng hình', fields: const [KField('img', 'Ảnh', kind: FieldKind.file, required: true)], submitLabel: 'Tìm', onSubmit: (v) async => r = await api.serpUploadImage(image: v.file('img')!));
                  if (r != null && context.mounted) await openPage(context, JsonPage(title: 'Kết quả hình ảnh', data: r));
                },
              ),
              OutlinedButton.icon(
                icon: const Icon(Icons.compare_arrows),
                label: const Text('So sánh nhiều từ khoá'),
                onPressed: () async {
                  SerpBatch? r;
                  await showFormSheet(
                    context,
                    title: 'Tìm cùng lúc',
                    fields: const [KField('qs', 'Từ khoá (mỗi dòng một từ)', kind: FieldKind.multiline, required: true)],
                    submitLabel: 'Tìm',
                    onSubmit: (v) async => r = await api.serpSearchBatch(searches: [
                      for (final q in (v.str('qs') ?? '').split('\n').map((s) => s.trim()).where((s) => s.isNotEmpty)) SerpSearchInput(engine: 'google', params: {'q': q, 'hl': 'vi', 'gl': 'vn'}),
                    ]),
                  );
                  if (r == null || !context.mounted) return;
                  await openPage(
                    context,
                    DetailPage(
                      title: 'Kết quả so sánh',
                      children: [
                        for (final it in r!.results ?? const <SerpBatchItem>[])
                          ListTile(
                            title: Text(it.searchId ?? 'Lượt tìm'),
                            subtitle: Text(it.error ?? '${(pick(it.data, 'organic_results') as List?)?.length ?? 0} kết quả${it.cached == true ? ' • từ bộ nhớ đệm' : ''}'),
                            onTap: it.searchId == null
                                ? null
                                : () async {
                                    final a = await api.serpArchive(id: it.searchId!).catchError((_) => null);
                                    if (context.mounted) await openPage(context, JsonPage(title: 'Lưu trữ ${it.searchId}', data: a));
                                  },
                          ),
                      ],
                    ),
                  );
                },
              ),
              OutlinedButton.icon(
                icon: const Icon(Icons.travel_explore),
                label: const Text('Tìm nâng cao'),
                onPressed: () async {
                  Object? r;
                  final locations = await api.serpLocations(q: 'Vietnam', limit: 20).catchError((_) => null);
                  final locs = locations is List ? [for (final l in locations) if (l is Map) ('${l['canonical_name'] ?? l['name']}', '${l['name']}')] : const <(String, String)>[];
                  if (!context.mounted) return;
                  await showFormSheet(
                    context,
                    title: 'Tìm nâng cao',
                    fields: [
                      KField('engine', 'Công cụ', kind: FieldKind.choice, options: [for (final e in (list.isEmpty ? const ['google', 'bing', 'youtube'] : list)) (e, e)]),
                      const KField('q', 'Từ khoá', required: true),
                      if (locs.isNotEmpty) KField('loc', 'Vị trí', kind: FieldKind.choice, options: locs),
                      const KField('nocache', 'Bỏ qua bộ nhớ đệm', kind: FieldKind.toggle, initial: 'false'),
                    ],
                    submitLabel: 'Tìm',
                    onSubmit: (v) async => r = await api.serpSearchBody(input: SerpSearchInput(engine: v.str('engine'), params: {'q': v.str('q'), 'location': ?v.str('loc')}, noCache: v.flag('nocache'))),
                  );
                  if (r != null && context.mounted) await openPage(context, JsonPage(title: 'Kết quả', data: r));
                },
              ),
            ]),
            const SectionHeader('Công cụ hỗ trợ'),
            for (final e in list)
              ListTile(
                contentPadding: const EdgeInsets.symmetric(horizontal: 20),
                leading: const Icon(Icons.manage_search),
                title: Text(e),
                onTap: () async {
                  final info = await api.serpEngine(engine: e).catchError((_) => null);
                  if (context.mounted) await openPage(context, JsonPage(title: e, data: info));
                },
              ),
          ],
        );
      },
    );
  }
}

class ShortlinkToolsPage extends StatelessWidget {
  const ShortlinkToolsPage({super.key});

  @override
  Widget build(BuildContext context) {
    final api = context.api.shortlink;
    return DetailPage(
      title: 'Deep link & đo lường',
      children: [
        ListTile(
          leading: const Icon(Icons.view_quilt_outlined),
          title: const Text('Giải mã theo mẫu'),
          subtitle: const Text('Link dạng /<mẫu>/<mã>'),
          onTap: () async {
            ShortlinkResolve? r;
            await showFormSheet(
              context,
              title: 'Link theo mẫu',
              fields: const [KField('tpl', 'Mẫu', required: true), KField('code', 'Mã', required: true)],
              submitLabel: 'Giải mã',
              onSubmit: (v) async => r = await api.shortlinkResolveWithTemplate(templateSlug: v.str('tpl')!, shortCode: v.str('code')!),
            );
            if (r != null && context.mounted) await showResult(context, title: r!.shortCode ?? '', message: [r!.webUrl, r!.androidUrl, r!.iosUrl, r!.deepLinkPath].whereType<String>().join('\n'));
          },
        ),
        ListTile(
          leading: const Icon(Icons.fingerprint),
          title: const Text('Ghi nhận cài đặt (thử SDK)'),
          subtitle: const Text('Gửi thông tin thiết bị để khớp với lượt bấm link gần nhất'),
          onTap: () async {
            ShortlinkAttribution? a;
            final size = MediaQuery.sizeOf(context);
            final platform = Theme.of(context).platform.name;
            final ok = await runAction(context, () async {
              a = await api.shortlinkSdkInstall(
                input: ShortlinkInstallInput(
                  platform: platform,
                  screenWidth: size.width.round(),
                  screenHeight: size.height.round(),
                  timezone: DateTime.now().timeZoneName,
                  language: 'vi',
                  sdkName: 'keikibook',
                  sdkVersion: '1.0.0',
                  attributionWindowHours: 24,
                ),
              );
              return a;
            });
            if (!ok || a == null || !context.mounted) return;
            final full = a!.fingerprint == null ? a : await api.shortlinkAttribution(fingerprint: a!.fingerprint!).catchError((_) => a);
            await api.shortlinkSdkEvent(input: ShortlinkEventInput(installId: full?.installEvent?.id, eventName: 'app_open', timestamp: DateTime.now().toUtc().toIso8601String(), attributedLinkId: full?.linkData?.shortCode, sdkName: 'keikibook', sdkVersion: '1.0.0')).catchError((_) => null);
            if (!context.mounted) return;
            await showResult(
              context,
              icon: full?.attributed == true ? Icons.link : Icons.link_off,
              title: full?.attributed == true ? 'Đã khớp link' : 'Chưa khớp link nào',
              message: full?.attributed == true ? '${full!.linkData?.originalUrl ?? ''}\nĐộ tin cậy ${((full.installEvent?.confidenceScore ?? 0) * 100).round()}%' : 'Mã thiết bị ${full?.fingerprint ?? ''}',
            );
          },
        ),
        ListTile(
          leading: const Icon(Icons.apple),
          title: const Text('apple-app-site-association'),
          onTap: () async {
            final r = await api.shortlinkAppleAppSiteAssociation().catchError((_) => null);
            if (context.mounted) await openPage(context, JsonPage(title: 'AASA', data: r));
          },
        ),
        ListTile(
          leading: const Icon(Icons.android),
          title: const Text('assetlinks.json'),
          onTap: () async {
            final r = await api.shortlinkAssetLinks().catchError((_) => null);
            if (context.mounted) await openPage(context, JsonPage(title: 'Asset links', data: r));
          },
        ),
      ],
    );
  }
}

class VendCatalogPage extends StatelessWidget {
  const VendCatalogPage({super.key});

  @override
  Widget build(BuildContext context) {
    final api = context.api.vending;
    return Scaffold(
      appBar: AppBar(title: const Text('Sản phẩm máy bán hàng')),
      body: SearchableList<VendProduct>(
        hint: 'Tìm sản phẩm',
        filters: () async => (await api.vendCategories(limit: 50))?.items?.map((c) => '${c.id}|${c.name_}').toList(),
        load: (q, cat) async {
          final items = (await api.vendProducts(categoryId: cat?.split('|').first, limit: 100))?.items ?? const <VendProduct>[];
          return q.isEmpty ? items : items.where((p) => (p.name_ ?? '').toLowerCase().contains(q.toLowerCase())).toList();
        },
        emptyMessage: 'Không có sản phẩm',
        itemBuilder: (context, p, _) => EntityTile(
          title: p.name_ ?? '',
          subtitle: p.description,
          caption: vnd((p.priceCents ?? 0) / 100),
          imageUrl: p.imageUrl,
          icon: Icons.local_drink_rounded,
          onTap: () async {
            final full = await api.vendProduct(id: p.id!).catchError((_) => null);
            if (context.mounted && full != null) await showResult(context, icon: Icons.local_drink, title: full.name_ ?? '', message: '${full.description ?? ''}\nMã vạch ${full.barcode ?? '-'}');
          },
        ),
      ),
    );
  }
}

Future<String?> vendMachineInfo(BuildContext context, String id) async {
  final m = await context.api.vending.vendMachine(id: id).catchError((_) => null);
  return m == null ? null : '${m.model ?? ''} • bảo trì ${shortDate(m.lastMaintenanceAt)}';
}

class LiveOwnerTools extends StatelessWidget {
  const LiveOwnerTools({super.key, required this.roomId});

  final String roomId;

  @override
  Widget build(BuildContext context) {
    final api = context.api.livestream;
    return AsyncBuilder<LiveStream?>(
      load: () => api.liveActiveStream(id: roomId).catchError((_) => null),
      loading: const SizedBox.shrink(),
      builder: (context, s, reload) => Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          InfoSection(title: 'Buổi phát hiện tại', rows: [
            ('Trạng thái', s?.status ?? 'Chưa phát'),
            ('Bắt đầu', s?.startedAt == null ? null : formatDate(s!.startedAt)),
            ('Người xem cao nhất', s == null ? null : '${s.peakViewer ?? 0}'),
            ('Máy chủ', s?.nodeId),
          ]),
          ActionButtons([
            FilledButton.icon(
              icon: const Icon(Icons.podcasts),
              label: const Text('Cấp máy chủ phát'),
              onPressed: () async {
                Object? r;
                final ok = await runAction(context, () async => r = await api.liveRequestIngest(id: roomId));
                if (ok && context.mounted) await showResult(context, title: 'Địa chỉ phát (RTMP)', message: '${pick(r, 'url') ?? pick(r, 'ingest_url') ?? r}');
                await reload();
              },
            ),
            OutlinedButton.icon(
              icon: const Icon(Icons.block),
              label: const Text('Cấm chat'),
              onPressed: () => showFormSheet(
                context,
                title: 'Cấm chat',
                fields: const [KField('user', 'Mã người dùng', required: true), KField('min', 'Thời gian (phút, bỏ trống = vĩnh viễn)', kind: FieldKind.number)],
                success: 'Đã cấm chat',
                onSubmit: (v) => api.liveBanChat(id: roomId, targetUserId: v.str('user')!, durationSeconds: v.integer('min') == null ? null : v.integer('min')! * 60),
              ),
            ),
            OutlinedButton.icon(
              icon: const Icon(Icons.lock_open),
              label: const Text('Bỏ cấm'),
              onPressed: () => showFormSheet(
                context,
                title: 'Bỏ cấm chat',
                fields: const [KField('user', 'Mã người dùng', required: true)],
                success: 'Đã bỏ cấm',
                onSubmit: (v) => api.liveUnbanChat(id: roomId, targetUserId: v.str('user')!),
              ),
            ),
          ]),
        ],
      ),
    );
  }
}

void reportLiveQuality(BuildContext context, String roomId, {int buffering = 0}) {
  context.api.livestream.liveReportQoE(id: roomId, bufferingEvents: buffering).catchError((_) => false);
}

class ChatGroupMembersPage extends StatefulWidget {
  const ChatGroupMembersPage({super.key, required this.groupId, required this.title});

  final int groupId;
  final String title;

  @override
  State<ChatGroupMembersPage> createState() => _ChatGroupMembersPageState();
}

class _ChatGroupMembersPageState extends State<ChatGroupMembersPage> {
  int _v = 0;

  @override
  Widget build(BuildContext context) {
    final api = context.api.ollamaChat;
    final gid = widget.groupId;
    return AsyncList<OllamaChatGroupMember>(
      key: ValueKey(_v),
      load: () async => (await api.ollamaChatGroupMembers(groupId: gid))?.members,
      header: Column(
        children: [
          AppBar(title: Text('Thành viên • ${widget.title}')),
          ActionButtons([
            FilledButton.icon(
              icon: const Icon(Icons.person_add_alt),
              label: const Text('Mời thành viên'),
              onPressed: () async {
                if (await showFormSheet(
                  context,
                  title: 'Mời vào nhóm',
                  fields: const [KField('user', 'Mã người dùng', kind: FieldKind.number, required: true), KField('role', 'Vai trò', kind: FieldKind.choice, options: [('member', 'Thành viên'), ('admin', 'Quản trị')])],
                  success: 'Đã mời',
                  onSubmit: (v) => api.ollamaChatInviteToGroup(groupId: gid, userId: v.integer('user')!, role: v.str('role')!),
                )) {
                  setState(() => _v++);
                }
              },
            ),
          ]),
        ],
      ),
      emptyMessage: 'Nhóm chưa có thành viên',
      itemBuilder: (context, m, reload) => EntityTile(
        title: m.username ?? '#${m.id}',
        subtitle: '${m.email ?? ''} • vào ${shortDate(m.joinedAt)}',
        icon: m.role == 'admin' || m.role == 'owner' ? Icons.shield_rounded : Icons.person_rounded,
        badges: [Pill(m.role ?? 'member')],
        trailing: PopupMenuButton<String>(
          onSelected: (a) async {
            final ok = a == 'promote'
                ? await runAction(context, () => api.ollamaChatPromoteMember(groupId: gid, userId: m.id!), success: 'Đã nâng quyền')
                : await runAction(context, () => api.ollamaChatRemoveMember(groupId: gid, userId: m.id!), confirm: 'Mời ${m.username} rời nhóm?', destructive: true, success: 'Đã xoá khỏi nhóm');
            if (ok) await reload();
          },
          itemBuilder: (_) => const [PopupMenuItem(value: 'promote', child: Text('Nâng làm quản trị')), PopupMenuItem(value: 'remove', child: Text('Xoá khỏi nhóm'))],
        ),
      ),
    );
  }
}

class ChatGroupSearchPage extends StatelessWidget {
  const ChatGroupSearchPage({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Tìm nhóm')),
      body: SearchableList<OllamaChatGroupSearchHit>(
        hint: 'Tên nhóm',
        load: (q, _) async => q.isEmpty ? const <OllamaChatGroupSearchHit>[] : (await context.api.ollamaChat.ollamaChatSearchGroups(q: q))?.groups,
        emptyMessage: 'Nhập tên để tìm nhóm',
        itemBuilder: (context, g, _) => EntityTile(
          title: g.name_ ?? '',
          subtitle: g.description,
          icon: Icons.groups_rounded,
          badges: [if (g.isMember == true) const Pill('Đã tham gia')],
          onTap: g.isMember == true ? () => openPage(context, ChatGroupMembersPage(groupId: g.id!, title: g.name_ ?? '')) : null,
        ),
      ),
    );
  }
}

class PostCategoryManager extends StatefulWidget {
  const PostCategoryManager({super.key});

  @override
  State<PostCategoryManager> createState() => _PostCategoryManagerState();
}

class _PostCategoryManagerState extends State<PostCategoryManager> {
  int _v = 0;

  @override
  Widget build(BuildContext context) {
    final api = context.api.post;
    Future<bool> form([PostCategory? c]) => showFormSheet(
          context,
          title: c == null ? 'Chuyên mục mới' : 'Sửa ${c.name_}',
          fields: [KField('name', 'Tên', required: true, initial: c?.name_), KField('desc', 'Mô tả', kind: FieldKind.multiline, initial: c?.description)],
          success: 'Đã lưu',
          onSubmit: (v) => c == null ? api.postCreateCategory(name_: v.str('name')!, description: v.str('desc')) : api.postUpdateCategory(id: c.id!, name_: v.str('name'), description: v.str('desc')),
        );
    return Scaffold(
      appBar: AppBar(title: const Text('Quản lý chuyên mục')),
      floatingActionButton: FloatingActionButton(
        onPressed: () async {
          if (await form()) setState(() => _v++);
        },
        child: const Icon(Icons.add),
      ),
      body: AsyncList<PostCategory>(
        key: ValueKey(_v),
        load: () async => (await api.postCategories(limit: 100))?.rows,
        emptyMessage: 'Chưa có chuyên mục',
        itemBuilder: (context, c, reload) => EntityTile(
          title: c.name_ ?? '',
          subtitle: c.description,
          icon: Icons.folder_rounded,
          onTap: () async {
            final full = await api.postCategory(id: c.id!).catchError((_) => null);
            if (!context.mounted) return;
            if (await form(full ?? c)) await reload();
          },
          trailing: IconButton(
            icon: const Icon(Icons.delete_outline),
            onPressed: () async {
              if (await runAction(context, () => api.postDeleteCategory(id: c.id!), confirm: 'Xoá chuyên mục ${c.name_}?', destructive: true, success: 'Đã xoá')) await reload();
            },
          ),
        ),
      ),
    );
  }
}

Future<bool> editPost(BuildContext context, PostPost p) => showFormSheet(
      context,
      title: 'Sửa bài viết',
      fields: [
        KField('name', 'Tiêu đề', required: true, initial: p.name_),
        KField('description', 'Tóm tắt', initial: p.description),
        KField('content', 'Nội dung', kind: FieldKind.multiline, required: true, initial: p.content),
      ],
      success: 'Đã lưu bài viết',
      onSubmit: (v) => context.api.post.postUpdatePost(id: p.id!, name_: v.str('name'), content: v.str('content'), description: v.str('description')),
    );

class SearchDocPage extends StatelessWidget {
  const SearchDocPage({super.key, required this.index, required this.id});

  final String index;
  final String id;

  @override
  Widget build(BuildContext context) {
    final api = context.api.search;
    return AsyncBuilder<(SearchDocument?, Object?, List<SearchFieldInfo>)>(
      load: () async {
        final r = await Future.wait<Object?>([
          api.searchDocument(index_: index, id: id),
          api.searchSimilar(index_: index, id: id, size: 5).catchError((_) => null),
          api.searchFields(index_: index).catchError((_) => null),
        ]);
        return (r[0] as SearchDocument?, r[1], (r[2] as List<SearchFieldInfo>?) ?? const <SearchFieldInfo>[]);
      },
      loading: const Scaffold(body: SkeletonList()),
      builder: (context, d, _) {
        final similar = pick(d.$2, 'items') ?? d.$2;
        return DetailPage(
          title: '${d.$1?.source?['title'] ?? d.$1?.source?['name'] ?? id}',
          children: [
            InfoSection(rows: [
              for (final f in d.$3)
                if (d.$1?.source?[f.name_] != null && d.$1!.source![f.name_] is! Map && d.$1!.source![f.name_] is! List) (f.name_ ?? '', '${d.$1!.source![f.name_]}'),
            ]),
            const SectionHeader('Tương tự'),
            for (final s in similar is List ? similar : const [])
              ListTile(
                contentPadding: const EdgeInsets.symmetric(horizontal: 20),
                leading: const Icon(Icons.description_outlined),
                title: Text('${pick(s, 'title') ?? pick(s, 'name') ?? pick(s, 'id') ?? ''}'),
                onTap: pick(s, 'id') == null ? null : () => openPage(context, SearchDocPage(index: '${pick(s, 'index') ?? index}', id: '${pick(s, 'id')}')),
              ),
          ],
        );
      },
    );
  }
}

Future<List<String>> searchSuggest(BuildContext context, String q) async {
  final r = await context.api.search.searchAutocomplete(q: q, size: 6).catchError((_) => null);
  final list = r is List ? r : (pick(r, 'suggestions') ?? pick(r, 'items'));
  return [for (final s in list is List ? list : const []) s is Map ? '${s['text'] ?? s['title'] ?? s['value'] ?? ''}' : '$s'].where((s) => s.isNotEmpty).toList();
}

Future<List<SearchDocument>?> searchDocs(BuildContext context, String q, String? index) async =>
    (await context.api.search.searchDocuments(q: q, index_: index == null ? null : [index], size: 30))?.items;

class GeoExtrasPage extends StatelessWidget {
  const GeoExtrasPage({super.key});

  @override
  Widget build(BuildContext context) {
    final api = context.api.geoservice;
    return AsyncBuilder<List<GeoWhitelistCity>?>(
      load: () => api.geoWhitelistCities().catchError((_) => null),
      loading: const Scaffold(body: SkeletonList()),
      builder: (context, cities, _) => DetailPage(
        title: 'Vùng phục vụ',
        children: [
          ActionButtons([
            FilledButton.icon(
              icon: const Icon(Icons.where_to_vote_outlined),
              label: const Text('Kiểm tra vị trí có được phục vụ'),
              onPressed: () async {
                GeoWhitelistResult? r;
                await showFormSheet(
                  context,
                  title: 'Toạ độ',
                  fields: const [KField('lat', 'Vĩ độ', kind: FieldKind.number, required: true, initial: '21.0285'), KField('lng', 'Kinh độ', kind: FieldKind.number, required: true, initial: '105.8542')],
                  submitLabel: 'Kiểm tra',
                  onSubmit: (v) async => r = await api.geoWhitelistCheck(lat: v.number('lat')!, lng: v.number('lng')!),
                );
                if (r == null || !context.mounted) return;
                await showResult(
                  context,
                  icon: r!.whitelisted == true ? Icons.check_circle : Icons.cancel,
                  title: r!.whitelisted == true ? 'Trong vùng phục vụ' : 'Ngoài vùng phục vụ',
                  message: [r!.ward?.nameFull, r!.district?.nameFull, r!.province?.nameFull].whereType<String>().join(', '),
                );
              },
            ),
          ]),
          const SectionHeader('Tỉnh/thành đang phục vụ'),
          for (final c in cities ?? const <GeoWhitelistCity>[])
            ListTile(
              contentPadding: const EdgeInsets.symmetric(horizontal: 20),
              leading: const Icon(Icons.location_city),
              title: Text('Mã ${c.provinceCode}'),
              subtitle: Text(c.note ?? ''),
              onTap: () async {
                final p = await api.geoProvince(code: c.provinceCode!).catchError((_) => null);
                if (context.mounted && p != null) showSnack(context, '${p.nameFull} • ${p.region ?? ''}');
              },
            ),
        ],
      ),
    );
  }
}

Future<List<(String, String, String?)>> geoStreetsOf(BuildContext context, String wardOrDistrict) async => [
      for (final s in await context.api.geoservice.geoStreets(code: wardOrDistrict).catchError((_) => null) ?? const <GeoStreet>[])
        ('${s.id}', s.nameFull ?? s.name_ ?? '', s.lengthM == null ? null : '${(s.lengthM! / 1000).toStringAsFixed(1)} km'),
    ];

class AiStudioPage extends StatelessWidget {
  const AiStudioPage({super.key});

  @override
  Widget build(BuildContext context) {
    final api = context.api.aiAgent;
    return AsyncBuilder<AiAgentTags?>(
      load: () => api.aiAgentModels().catchError((_) => null),
      loading: const Scaffold(body: SkeletonList()),
      builder: (context, tags, _) {
        final models = tags?.models ?? const <AiAgentModel>[];
        return DetailPage(
          title: 'AI Studio',
          color: const Color(0xFF6D28D9),
          header: const AppBanner(icon: Icons.auto_awesome, title: 'Mô hình của Keiki AI', subtitle: 'Xem thông tin mô hình và sinh văn bản một lần, không lưu vào hội thoại'),
          children: [
            ActionButtons([
              FilledButton.icon(
                icon: const Icon(Icons.edit_note),
                label: const Text('Sinh văn bản'),
                onPressed: () async {
                  AiAgentGenerateResponse? r;
                  await showFormSheet(
                    context,
                    title: 'Sinh văn bản',
                    fields: [
                      KField('model', 'Mô hình', kind: FieldKind.choice, options: [for (final m in models) (m.name_ ?? '', m.name_ ?? '')]),
                      const KField('prompt', 'Yêu cầu', kind: FieldKind.multiline, required: true, hint: 'Viết caption cho ảnh món phở…'),
                    ],
                    submitLabel: 'Sinh',
                    onSubmit: (v) async => r = await api.aiAgentGenerate(input: AiAgentGenerateInput(model: v.str('model'), prompt: v.str('prompt'))),
                  );
                  if (r == null || !context.mounted) return;
                  await openPage(
                    context,
                    DetailPage(
                      title: r!.model ?? 'Kết quả',
                      actions: [IconButton(icon: const Icon(Icons.copy), onPressed: () => _copy(context, r!.response ?? ''))],
                      children: [
                        Paragraph(r!.response),
                        InfoSection(rows: [('Token sinh ra', '${r!.evalCount ?? 0}'), ('Thời gian', '${((r!.totalDuration ?? 0) / 1e9).toStringAsFixed(1)} giây')]),
                      ],
                    ),
                  );
                },
              ),
            ]),
            const SectionHeader('Mô hình'),
            for (final m in models)
              ListTile(
                contentPadding: const EdgeInsets.symmetric(horizontal: 20),
                leading: const Icon(Icons.memory),
                title: Text(m.name_ ?? ''),
                subtitle: Text('${m.details?.parameterSize ?? ''} • ${m.details?.quantizationLevel ?? ''} • ${((m.size ?? 0) / 1e9).toStringAsFixed(1)} GB'),
                onTap: () async {
                  final s = await api.aiAgentShowModel(input: AiAgentShowInput(model: m.name_)).catchError((_) => null);
                  if (context.mounted && s != null) {
                    await openPage(
                      context,
                      DetailPage(
                        title: m.name_ ?? '',
                        children: [
                          InfoSection(rows: [('Họ mô hình', s.details?.family), ('Khả năng', (s.capabilities ?? const []).join(', ')), ('Tham số', s.parameters)]),
                          Paragraph(s.template, title: 'Mẫu prompt'),
                        ],
                      ),
                    );
                  }
                },
              ),
          ],
        );
      },
    );
  }
}

class BonusPartnerPage extends StatelessWidget {
  const BonusPartnerPage({super.key});

  @override
  Widget build(BuildContext context) {
    final api = context.api.bonuslink;
    return DetailPage(
      title: 'Đối tác Bonuslink',
      children: [
        ActionButtons([
          FilledButton.icon(
            icon: const Icon(Icons.add_task),
            label: const Text('Ghi nhận thưởng'),
            onPressed: () async {
              BonusRecordResult? r;
              await showFormSheet(
                context,
                title: 'Ghi nhận thưởng cho khách',
                fields: const [
                  KField('user', 'Mã khách hàng', required: true),
                  KField('type', 'Loại thưởng', kind: FieldKind.choice, options: [('points', 'Điểm'), ('cashback', 'Hoàn tiền'), ('voucher', 'Voucher')]),
                  KField('value', 'Giá trị', kind: FieldKind.number, required: true),
                  KField('source', 'Nguồn', initial: 'store'),
                  KField('ref', 'Mã giới thiệu'),
                ],
                submitLabel: 'Ghi nhận',
                onSubmit: (v) async => r = await api.bonusRecordReward(
                  input: BonusRecordRewardInput(eventId: newIdempotencyKey(), userId: v.str('user'), rewardType: v.str('type'), rewardValue: v.number('value'), source: v.str('source'), refCode: v.str('ref')),
                ),
              );
              if (r != null && context.mounted) await showResult(context, title: r!.duplicate == true ? 'Đã ghi nhận trước đó' : 'Đã ghi nhận', message: 'Mã thưởng ${r!.rewardId}');
            },
          ),
          OutlinedButton.icon(
            icon: const Icon(Icons.person_search),
            label: const Text('Tra cứu khách'),
            onPressed: () async {
              String? user;
              List<BonusBalance>? balances;
              List<BonusReward>? rewards;
              await showFormSheet(
                context,
                title: 'Tra cứu thưởng của khách',
                fields: const [KField('user', 'Mã khách hàng', required: true)],
                submitLabel: 'Xem',
                onSubmit: (v) async {
                  user = v.str('user');
                  balances = await api.bonusUserBalances(userId: user!);
                  rewards = await api.bonusUserRewards(userId: user!, limit: 20);
                },
              );
              if (user == null || !context.mounted) return;
              await showResult(
                context,
                title: 'Khách $user',
                message: [
                  for (final b in balances ?? const <BonusBalance>[]) '${b.rewardType}: ${compact(b.totalValue)} (${b.rewardCount} lần)',
                  '',
                  for (final r in (rewards ?? const <BonusReward>[]).take(8)) '• ${r.rewardType} +${compact(r.rewardValue)} ${shortDate(r.createdAt)}',
                ].join('\n'),
              );
            },
          ),
        ]),
      ],
    );
  }
}

class CalcMarinePage extends StatelessWidget {
  const CalcMarinePage({super.key, required this.location});

  final CalcTrackedLocation location;

  @override
  Widget build(BuildContext context) {
    final api = context.api.calculate;
    return AsyncBuilder<(CalcTidePrediction?, CalcRainRadar?)>(
      load: () async {
        final r = await Future.wait<Object?>([
          location.tideStationId == null ? Future.value(null) : api.calcTide(station: location.tideStationId!).catchError((_) => null),
          api.calcRadar().catchError((_) => null),
        ]);
        return (r[0] as CalcTidePrediction?, r[1] as CalcRainRadar?);
      },
      loading: const Scaffold(body: SkeletonList()),
      builder: (context, d, _) {
        final frames = [...?d.$2?.past, ...?d.$2?.nowcast];
        final latest = frames.isEmpty ? null : frames.last;
        return DetailPage(
          title: 'Thuỷ triều & radar • ${location.name_}',
          children: [
            if (latest?.tileUrl != null)
              Padding(
                padding: const EdgeInsets.all(16),
                child: AspectRatio(
                  aspectRatio: 1,
                  child: LayoutBuilder(builder: (context, c) => KThumb(url: latest!.tileUrl!.replaceAll('{z}', '6').replaceAll('{x}', '51').replaceAll('{y}', '28'), icon: Icons.radar, width: c.maxWidth, height: c.maxHeight)),
                ),
              ),
            InfoSection(title: 'Radar mưa', rows: [('Cập nhật', formatDate(d.$2?.generatedAt)), ('Số khung hình', '${frames.length}'), ('Nguồn', d.$2?.source)]),
            InfoSection(title: 'Thuỷ triều', rows: [
              for (final e in (d.$1?.events ?? const <CalcTideEvent>[]).take(12)) (formatDate(e.time), '${e.type_ == 'high' ? 'Lên' : 'Xuống'} ${e.heightM?.toStringAsFixed(2) ?? '-'} m'),
            ]),
          ],
        );
      },
    );
  }
}

Future<void> faceReportFrom(BuildContext context, FacePositions? p) async {
  final api = context.api.face;
  await showFormSheet(
    context,
    title: 'Gửi vị trí khuôn mặt',
    description: 'Dùng cho máy ảnh/thiết bị gửi kết quả nhận diện lên máy chủ. Mỗi dòng: x y rộng cao.',
    fields: [
      KField('faces', 'Khung mặt', kind: FieldKind.multiline, required: true, initial: (p?.faces ?? const <FaceRect>[]).map((f) => '${f.x} ${f.y} ${f.width} ${f.height}').join('\n')),
      const KField('more', 'Ảnh phụ (tuỳ chọn)', kind: FieldKind.file),
    ],
    success: 'Đã gửi',
    onSubmit: (v) async {
      final faces = [
        for (final line in (v.str('faces') ?? '').split('\n'))
          if (line.trim().split(RegExp(r'\s+')).length == 4)
            FaceRectInput(
              x: double.tryParse(line.trim().split(RegExp(r'\s+'))[0]),
              y: double.tryParse(line.trim().split(RegExp(r'\s+'))[1]),
              width: double.tryParse(line.trim().split(RegExp(r'\s+'))[2]),
              height: double.tryParse(line.trim().split(RegExp(r'\s+'))[3]),
            ),
      ];
      await api.faceReportPositions(input: FacePositionsInput(faces: faces));
      if (v.file('more') != null) await api.faceUploadFiles(boop: [v.file('more')!]);
    },
  );
}
