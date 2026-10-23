import 'package:flutter/material.dart';

import '../../../gen/models.g.dart';
import '../../../app/role.dart';
import '../kit/kit.dart';
import 'polymarket_more.dart';

class PolymarketApp extends StatelessWidget {
  const PolymarketApp({super.key});

  @override
  Widget build(BuildContext context) {
    return ServiceApp(
      serviceId: 'polymarket',
      adminActions: [
        Builder(builder: (context) => IconButton(tooltip: 'Tạo thị trường', icon: const Icon(Icons.add_chart), onPressed: () => openPage(context, const PolyCreatorPage()))),
      ],
      tabs: [
        AppTab('Thị trường', Icons.insights_outlined, (_) => const _Events()),
        AppTab('Kèo', Icons.query_stats, (_) => const PolyMarketsSearch()),
        AppTab('Danh mục', Icons.account_balance_wallet_outlined, (_) => const _Portfolio()),
        AppTab('Tài khoản', Icons.person_outline, (_) => const PolyAccount()),
        AppTab('Theo dõi', Icons.bookmark_border, (context) => AsyncList<PolyEvent>(
              load: () => context.api.polymarket.polyWatchlist(userId: '${context.userId}'),
              emptyMessage: 'Bạn chưa theo dõi sự kiện nào',
              emptyIcon: Icons.bookmark_border,
              itemBuilder: (context, e, reload) => _EventCard(event: e),
            )),
        AppTab('Xếp hạng', Icons.leaderboard_outlined, (context) => AsyncList<PolyLeaderboardEntry>(
              load: () => context.api.polymarket.polyLeaderboard(metric: 'profit', limit: 50),
              emptyMessage: 'Chưa có bảng xếp hạng',
              spacing: 6,
              itemBuilder: (context, e, _) => EntityTile(
                leading: CircleAvatar(
                  backgroundColor: (e.rank ?? 99) <= 3 ? const Color(0xFFF59E0B) : context.appColor.withValues(alpha: 0.12),
                  child: Text('${e.rank ?? '-'}', style: TextStyle(fontWeight: FontWeight.w800, color: (e.rank ?? 99) <= 3 ? Colors.white : null)),
                ),
                title: e.username ?? e.userId ?? 'Người chơi',
                subtitle: 'Khối lượng ${vnd(e.volume)}',
                trailing: Text(
                  '${(e.profit ?? 0) >= 0 ? '+' : ''}${vnd(e.profit)}',
                  style: TextStyle(fontWeight: FontWeight.w800, color: (e.profit ?? 0) >= 0 ? const Color(0xFF16A34A) : const Color(0xFFDC2626)),
                ),
              ),
            )),
      ],
    );
  }
}

double _prob(PolyMarket m) {
  final share = (m.shareValue ?? 100).toDouble();
  final p = (m.yesPrice ?? 0) / (share == 0 ? 100 : share);
  return p.clamp(0, 1).toDouble();
}

class _Events extends StatelessWidget {
  const _Events();

  @override
  Widget build(BuildContext context) {
    final api = context.api.polymarket;
    return SearchableList<PolyEvent>(
      hint: 'Tìm sự kiện để dự đoán',
      filters: () async => (await api.polyCategories())?.map((c) => c.category ?? '').where((c) => c.isNotEmpty).toList(),
      load: (q, cat) async => (await api.polyEvents(q: q.isEmpty ? null : q, category: cat, status: 'open', sort: 'volume', limit: 40))?.items,
      emptyMessage: 'Chưa có sự kiện dự đoán',
      emptyIcon: Icons.insights_outlined,
      itemBuilder: (context, e, _) => _EventCard(event: e),
    );
  }
}

class _EventCard extends StatelessWidget {
  const _EventCard({required this.event});

  final PolyEvent event;

  @override
  Widget build(BuildContext context) {
    final markets = event.markets ?? const <PolyMarket>[];
    final theme = Theme.of(context);
    return KCard(
      onTap: () => openPage(context, PolyEventPage(id: '${event.id ?? event.slug}')),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              KThumb(url: event.imageUrl, icon: Icons.insights_rounded, size: 44, radius: 12),
              const SizedBox(width: 12),
              Expanded(child: Text(event.title ?? 'Sự kiện', maxLines: 2, overflow: TextOverflow.ellipsis, style: theme.textTheme.titleSmall?.copyWith(fontWeight: FontWeight.w700))),
            ],
          ),
          const SizedBox(height: 10),
          for (final m in markets.take(3))
            Padding(
              padding: const EdgeInsets.only(bottom: 6),
              child: Row(
                children: [
                  Expanded(child: Text(m.groupItemTitle ?? m.question ?? 'Có', maxLines: 1, overflow: TextOverflow.ellipsis)),
                  Text('${(_prob(m) * 100).round()}%', style: const TextStyle(fontWeight: FontWeight.w800)),
                ],
              ),
            ),
          Row(
            children: [
              Icon(Icons.bar_chart_rounded, size: 16, color: theme.colorScheme.onSurfaceVariant),
              Text(' ${vnd(event.volume)}', style: theme.textTheme.bodySmall),
              const Spacer(),
              if (event.endDate != null) Text('Kết thúc ${shortDate(event.endDate)}', style: theme.textTheme.bodySmall),
            ],
          ),
        ],
      ),
    );
  }
}

class PolyEventPage extends StatelessWidget {
  const PolyEventPage({super.key, required this.id});

  final String id;

  @override
  Widget build(BuildContext context) {
    final api = context.api.polymarket;
    return AsyncBuilder<PolyEvent?>(
      load: () => api.polyEvent(id: id),
      loading: const Scaffold(body: SkeletonList()),
      builder: (context, e, reload) {
        if (e == null) return const Scaffold(body: EmptyView(message: 'Không tìm thấy sự kiện'));
        return DetailPage(
          title: e.title ?? 'Sự kiện',
          actions: [
            ShareLinkButton(path: '/prediction/${e.slug ?? id}', title: e.title ?? 'Sự kiện dự đoán', imageUrl: e.imageUrl),
            IconButton(
              tooltip: 'Theo dõi',
              icon: const Icon(Icons.bookmark_add_outlined),
              onPressed: () => runAction(context, () => api.polyWatchEvent(id: id), success: 'Đã theo dõi sự kiện'),
            ),
            PopupMenuButton<String>(
              onSelected: (a) async {
                if (await polyEventAdmin(context, e, a)) await reload();
              },
              itemBuilder: (_) => [
                const PopupMenuItem(value: 'unwatch', child: Text('Bỏ theo dõi')),
                if (e.negRisk == true) const PopupMenuItem(value: 'convert', child: Text('Chuyển đổi vị thế')),
                const PopupMenuItem(value: 'propose', child: Text('Đề xuất lựa chọn thắng')),
                if (context.isAdminNow) const PopupMenuItem(value: 'resolve', child: Text('Phân xử (quản trị)')),
              ],
            ),
          ],
          header: HeroHeader(
            title: e.title ?? 'Sự kiện',
            imageUrl: e.imageUrl,
            icon: Icons.insights_rounded,
            subtitle: 'Khối lượng ${vnd(e.volume)}${e.endDate == null ? '' : ' • Kết thúc ${shortDate(e.endDate)}'}',
            badges: [if (e.category != null) Pill(e.category!), StatusPill(e.status)],
          ),
          children: [
            const SectionHeader('Kết quả có thể xảy ra'),
            for (final m in e.markets ?? const <PolyMarket>[])
              Padding(
                padding: const EdgeInsets.fromLTRB(16, 0, 16, 10),
                child: KCard(
                  onTap: () => openPage(context, PolyMarketPage(id: '${m.id}')),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Row(
                        children: [
                          Expanded(child: Text(m.groupItemTitle ?? m.question ?? '', style: const TextStyle(fontWeight: FontWeight.w700))),
                          Text('${(_prob(m) * 100).round()}%', style: TextStyle(fontWeight: FontWeight.w800, fontSize: 18, color: context.appColor)),
                        ],
                      ),
                      const SizedBox(height: 8),
                      ClipRRect(
                        borderRadius: BorderRadius.circular(99),
                        child: LinearProgressIndicator(value: _prob(m), minHeight: 6, color: const Color(0xFF16A34A), backgroundColor: const Color(0xFFDC2626).withValues(alpha: 0.25)),
                      ),
                      const SizedBox(height: 10),
                      Row(
                        children: [
                          Expanded(
                            child: FilledButton(
                              style: FilledButton.styleFrom(backgroundColor: const Color(0xFF16A34A)),
                              onPressed: m.status == 'open' || m.status == null ? () => _trade(context, m, 'yes') : null,
                              child: Text('Có ${(_prob(m) * 100).round()}¢'),
                            ),
                          ),
                          const SizedBox(width: 8),
                          Expanded(
                            child: FilledButton(
                              style: FilledButton.styleFrom(backgroundColor: const Color(0xFFDC2626)),
                              onPressed: m.status == 'open' || m.status == null ? () => _trade(context, m, 'no') : null,
                              child: Text('Không ${((1 - _prob(m)) * 100).round()}¢'),
                            ),
                          ),
                        ],
                      ),
                      if (m.resolvedOutcome != null && m.resolvedOutcome!.isNotEmpty)
                        Padding(padding: const EdgeInsets.only(top: 8), child: Pill('Kết quả: ${m.resolvedOutcome == 'yes' ? 'Có' : 'Không'}')),
                    ],
                  ),
                ),
              ),
            Paragraph(e.description, title: 'Luật phân xử'),
            PolyRelated(eventId: id, open: (r) => openPage(context, PolyEventPage(id: '${r.id ?? r.slug}'))),
            const SectionHeader('Bình luận'),
            _Comments(eventId: id),
          ],
        );
      },
    );
  }

  Future<void> _trade(BuildContext context, PolyMarket m, String outcome) async {
    final api = context.api.polymarket;
    await showFormSheet(
      context,
      title: '${outcome == 'yes' ? 'Mua “Có”' : 'Mua “Không”'} • ${m.groupItemTitle ?? m.question ?? ''}',
      description: 'Nếu kết quả đúng, mỗi cổ phần được trả ${vnd(m.shareValue ?? 100)}.',
      fields: const [KField('amount', 'Số tiền (₫)', kind: FieldKind.number, required: true, icon: Icons.payments_outlined)],
      submitLabel: 'Đặt lệnh',
      success: 'Đã đặt lệnh',
      onSubmit: (v) async {
        final amount = v.integer('amount')!;
        final q = await api.polyQuote(id: '${m.id}', side: 'buy', outcome: outcome, amount: amount);
        if (q != null && q.fillable == false) throw Exception('Không đủ thanh khoản cho số tiền này');
        await api.polyPlaceOrder(
          id: '${m.id}',
          input: PolyOrderInput(clientOrderId: newIdempotencyKey(), outcome: outcome, side: 'buy', type_: 'market', amount: amount),
        );
      },
    );
  }
}

class _Comments extends StatelessWidget {
  const _Comments({required this.eventId});

  final String eventId;

  @override
  Widget build(BuildContext context) {
    final api = context.api.polymarket;
    return AsyncBuilder<PolyCommentPage?>(
      load: () => api.polyComments(id: eventId, limit: 20),
      loading: const Padding(padding: EdgeInsets.all(24), child: Center(child: CircularProgressIndicator())),
      builder: (context, page, reload) => Padding(
        padding: const EdgeInsets.symmetric(horizontal: 16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            OutlinedButton.icon(
              icon: const Icon(Icons.add_comment_outlined),
              label: const Text('Viết bình luận'),
              onPressed: () async {
                if (await showFormSheet(
                  context,
                  title: 'Bình luận',
                  fields: const [KField('body', 'Nội dung', kind: FieldKind.multiline, required: true)],
                  onSubmit: (v) => api.polyAddComment(id: eventId, body: v.str('body')!),
                )) {
                  await reload();
                }
              },
            ),
            for (final c in page?.items ?? const <PolyComment>[])
              ListTile(
                contentPadding: EdgeInsets.zero,
                leading: const CircleAvatar(child: Icon(Icons.person_outline)),
                title: Text(c.body ?? ''),
                subtitle: Text('${c.userId ?? ''} • ${formatDate(c.createdAt)}'),
                onLongPress: () async {
                  final mine = c.userId == '${context.userId}';
                  final action = await showModalBottomSheet<String>(
                    context: context,
                    builder: (_) => SafeArea(
                      child: Column(
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          ListTile(leading: const Icon(Icons.thumb_down_alt_outlined), title: const Text('Bỏ thích'), onTap: () => Navigator.pop(context, 'unlike')),
                          if (mine) ListTile(leading: const Icon(Icons.delete_outline), title: const Text('Xoá bình luận'), onTap: () => Navigator.pop(context, 'delete')),
                        ],
                      ),
                    ),
                  );
                  if (action == null || !context.mounted) return;
                  final ok = action == 'delete'
                      ? await runAction(context, () => api.polyDeleteComment(commentId: '${c.id}'), confirm: 'Xoá bình luận này?', destructive: true, success: 'Đã xoá')
                      : await runAction(context, () => api.polyUnlikeComment(commentId: '${c.id}'));
                  if (ok) await reload();
                },
                trailing: TextButton.icon(
                  onPressed: () => runAction(context, () => api.polyLikeComment(commentId: '${c.id}')),
                  icon: const Icon(Icons.thumb_up_alt_outlined, size: 16),
                  label: Text('${c.likes ?? 0}'),
                ),
              ),
          ],
        ),
      ),
    );
  }
}

class _Portfolio extends StatelessWidget {
  const _Portfolio();

  @override
  Widget build(BuildContext context) {
    final api = context.api.polymarket;
    return AsyncBuilder<PolyPortfolio?>(
      load: () => api.polyPortfolio(userId: '${context.userId}'),
      builder: (context, p, reload) {
        final positions = p?.positions ?? const <PolyPositionView>[];
        return RefreshIndicator(
          onRefresh: reload,
          child: ListView(
            padding: const EdgeInsets.only(bottom: 24),
            children: [
              AppBanner(
                icon: Icons.account_balance_wallet_rounded,
                title: vnd(p?.totalValue),
                subtitle: 'Tổng giá trị danh mục',
                children: [
                  StatRow([
                    StatTile(label: 'Khả dụng', value: vnd(p?.balance?.available), onDark: true),
                    StatTile(label: 'Lãi/lỗ tạm tính', value: vnd(p?.unrealizedPnl), onDark: true),
                  ]),
                  const SizedBox(height: 12),
                  Row(
                    children: [
                      Expanded(
                        child: FilledButton.tonal(
                          onPressed: () async {
                            if (await _money(context, 'Nạp tiền từ Ví Keiki', (a) => api.polyDeposit(amount: a))) await reload();
                          },
                          child: const Text('Nạp'),
                        ),
                      ),
                      const SizedBox(width: 8),
                      Expanded(
                        child: FilledButton.tonal(
                          onPressed: () async {
                            if (await _money(context, 'Rút về Ví Keiki', (a) => api.polyWithdraw(amount: a))) await reload();
                          },
                          child: const Text('Rút'),
                        ),
                      ),
                    ],
                  ),
                ],
              ),
              const SectionHeader('Vị thế đang nắm'),
              if (positions.isEmpty) const Padding(padding: EdgeInsets.all(24), child: Center(child: Text('Chưa có vị thế nào'))),
              for (final pos in positions)
                Padding(
                  padding: const EdgeInsets.fromLTRB(16, 0, 16, 8),
                  child: EntityTile(
                    title: pos.question ?? 'Thị trường #${pos.marketId}',
                    subtitle: '${pos.shares ?? 0} cổ phần “${pos.outcome == 'yes' ? 'Có' : 'Không'}” • giá TB ${pos.avgPrice?.toStringAsFixed(1) ?? '-'}',
                    icon: Icons.pie_chart_outline,
                    caption: 'Giá trị ${vnd(pos.value)}',
                    trailing: Text(
                      vnd(pos.unrealizedPnl),
                      style: TextStyle(fontWeight: FontWeight.w800, color: (pos.unrealizedPnl ?? 0) >= 0 ? const Color(0xFF16A34A) : const Color(0xFFDC2626)),
                    ),
                    onTap: pos.eventId == null ? null : () => openPage(context, PolyEventPage(id: '${pos.eventId}')),
                  ),
                ),
            ],
          ),
        );
      },
    );
  }

  Future<bool> _money(BuildContext context, String title, Future<Object?> Function(int) call) => showFormSheet(
        context,
        title: title,
        fields: const [KField('amount', 'Số tiền (₫)', kind: FieldKind.number, required: true)],
        success: 'Giao dịch thành công',
        onSubmit: (v) => call(v.integer('amount')!),
      );
}
