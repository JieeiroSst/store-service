import 'package:flutter/material.dart';
import 'package:flutter/services.dart';

import '../../../gen/models.g.dart';
import '../../../app/role.dart';
import '../kit/kit.dart';

const _green = Color(0xFF16A34A);
const _red = Color(0xFFDC2626);

String cents(num? price, [int share = 100]) => '${((price ?? 0) * 100 / (share == 0 ? 100 : share)).toStringAsFixed(0)}¢';

class _Sparkline extends CustomPainter {
  _Sparkline(this.points, this.color);

  final List<double> points;
  final Color color;

  @override
  void paint(Canvas canvas, Size size) {
    if (points.length < 2) return;
    final lo = points.reduce((a, b) => a < b ? a : b);
    final hi = points.reduce((a, b) => a > b ? a : b);
    final span = (hi - lo).abs() < 1e-9 ? 1.0 : hi - lo;
    final path = Path();
    for (final (i, p) in points.indexed) {
      final x = size.width * i / (points.length - 1);
      final y = size.height - (p - lo) / span * size.height;
      i == 0 ? path.moveTo(x, y) : path.lineTo(x, y);
    }
    final fill = Path.from(path)
      ..lineTo(size.width, size.height)
      ..lineTo(0, size.height)
      ..close();
    canvas.drawPath(fill, Paint()..shader = LinearGradient(begin: Alignment.topCenter, end: Alignment.bottomCenter, colors: [color.withValues(alpha: 0.25), color.withValues(alpha: 0)]).createShader(Offset.zero & size));
    canvas.drawPath(path, Paint()
      ..color = color
      ..style = PaintingStyle.stroke
      ..strokeWidth = 2.2);
  }

  @override
  bool shouldRepaint(_Sparkline old) => old.points != points || old.color != color;
}

class PolyMarketPage extends StatefulWidget {
  const PolyMarketPage({super.key, required this.id});

  final String id;

  @override
  State<PolyMarketPage> createState() => _PolyMarketPageState();
}

class _PolyMarketPageState extends State<PolyMarketPage> {
  String _range = '1w';
  String _outcome = 'yes';
  int _v = 0;

  Future<void> _order(PolyMarket m, String side) async {
    final api = context.api.polymarket;
    final share = m.shareValue ?? 100;
    if (await showFormSheet(
      context,
      title: '${side == 'buy' ? 'Mua' : 'Bán'} “${_outcome == 'yes' ? 'Có' : 'Không'}”',
      description: 'Lệnh giới hạn khớp khi có người đặt giá đối ứng. Giá tính theo đơn vị 0–$share.',
      fields: [
        KField('price', 'Giá mỗi cổ phần (0–$share)', kind: FieldKind.number, required: true, initial: '${((_outcome == 'yes' ? m.yesPrice : m.noPrice) ?? share / 2).round()}'),
        KField('size', 'Số cổ phần', kind: FieldKind.number, required: true, initial: '${m.minOrderSize ?? 10}'),
        const KField('tif', 'Hiệu lực', kind: FieldKind.choice, options: [('gtc', 'Tới khi huỷ'), ('ioc', 'Khớp ngay hoặc huỷ'), ('fok', 'Khớp hết hoặc huỷ')]),
      ],
      submitLabel: 'Đặt lệnh',
      success: 'Đã đặt lệnh',
      onSubmit: (v) => api.polyPlaceOrder(
        id: '${m.id}',
        input: PolyOrderInput(
          clientOrderId: newIdempotencyKey(),
          outcome: _outcome,
          side: side,
          type_: 'limit',
          timeInForce: v.str('tif'),
          price: v.integer('price'),
          size: v.integer('size'),
        ),
      ),
    )) {
      setState(() => _v++);
    }
  }

  Future<void> _resolution(PolyMarket m, String action) async {
    final api = context.api.polymarket;
    final id = '${m.id}';
    final Future<bool> pending = switch (action) {
      'propose' => showFormSheet(
          context,
          title: 'Đề xuất kết quả',
          description: 'Bạn đặt cọc ${vnd(m.disputeBond)} để đề xuất. Nếu không ai phản đối trong thời hạn, kết quả được chốt.',
          fields: const [KField('outcome', 'Kết quả', kind: FieldKind.choice, options: [('yes', 'Có'), ('no', 'Không')])],
          success: 'Đã gửi đề xuất',
          onSubmit: (v) => api.polyProposeResolution(id: id, outcome: v.str('outcome')!),
        ),
      'dispute' => showFormSheet(
          context,
          title: 'Phản đối kết quả',
          fields: const [KField('reason', 'Lý do và nguồn dẫn chứng', kind: FieldKind.multiline, required: true)],
          success: 'Đã gửi phản đối',
          onSubmit: (v) => api.polyDisputeResolution(id: id, reason: v.str('reason')!),
        ),
      'finalize' => runAction(context, () => api.polyFinalizeResolution(id: id), confirm: 'Chốt kết quả đã đề xuất?', success: 'Đã chốt'),
      'resolve' => showFormSheet(
          context,
          title: 'Phân xử (quản trị)',
          fields: const [KField('outcome', 'Kết quả', kind: FieldKind.choice, options: [('yes', 'Có'), ('no', 'Không')])],
          success: 'Đã phân xử',
          onSubmit: (v) => api.polyResolveMarket(id: id, outcome: v.str('outcome')!),
        ),
      'rewards' => showFormSheet(
          context,
          title: 'Thưởng tạo lập thị trường',
          fields: [
            KField('pool', 'Quỹ thưởng mỗi ngày', kind: FieldKind.number, required: true, initial: '${m.rewardPool ?? 0}'),
            KField('spread', 'Chênh lệch giá tối đa', kind: FieldKind.number, required: true, initial: '${m.rewardMaxSpread ?? 3}'),
            KField('min', 'Khối lượng lệnh tối thiểu', kind: FieldKind.number, required: true, initial: '${m.rewardMinSize ?? 10}'),
          ],
          success: 'Đã cập nhật thưởng',
          onSubmit: (v) => api.polySetMarketRewards(id: id, dailyPool: v.integer('pool')!, maxSpread: v.integer('spread')!, minSize: v.integer('min')!),
        ),
      _ => runAction(context, () => api.polyCancelAllOrders(id: id), confirm: 'Huỷ mọi lệnh đang chờ của bạn ở thị trường này?', success: 'Đã huỷ lệnh'),
    };
    final ok = await pending;
    if (ok && mounted) setState(() => _v++);
  }

  @override
  Widget build(BuildContext context) {
    final api = context.api.polymarket;
    final id = widget.id;
    return AsyncBuilder<(PolyMarket?, PolyOrderBook?, List<PolyCandle>, List<PolyTrade>, List<PolyHolderEntry>, PolyMarketRewards?)>(
      key: ValueKey('$_v$_range$_outcome'),
      load: () async {
        final r = await Future.wait<Object?>([
          api.polyMarket(id: id),
          api.polyOrderBook(id: id, outcome: _outcome).catchError((_) => null),
          api.polyPriceHistory(id: id, range: _range).catchError((_) => null),
          api.polyMarketTrades(id: id, limit: 20).catchError((_) => null),
          api.polyTopHolders(id: id, outcome: _outcome, limit: 10).catchError((_) => null),
          api.polyMarketRewards(id: id).catchError((_) => null),
        ]);
        return (
          r[0] as PolyMarket?,
          r[1] as PolyOrderBook?,
          (r[2] as List<PolyCandle>?) ?? const <PolyCandle>[],
          (r[3] as List<PolyTrade>?) ?? const <PolyTrade>[],
          (r[4] as List<PolyHolderEntry>?) ?? const <PolyHolderEntry>[],
          r[5] as PolyMarketRewards?,
        );
      },
      loading: const Scaffold(body: SkeletonList()),
      builder: (context, d, _) {
        final m = d.$1;
        if (m == null) return const Scaffold(body: EmptyView(message: 'Không tìm thấy thị trường'));
        final share = m.shareValue ?? 100;
        final book = d.$2;
        final open = m.status == 'open' || m.status == null;
        return DetailPage(
          title: m.groupItemTitle ?? m.question ?? 'Thị trường',
          actions: [
            PopupMenuButton<String>(
              onSelected: (a) => _resolution(m, a),
              itemBuilder: (_) => [
                const PopupMenuItem(value: 'cancel', child: Text('Huỷ mọi lệnh của tôi')),
                const PopupMenuItem(value: 'propose', child: Text('Đề xuất kết quả')),
                const PopupMenuItem(value: 'dispute', child: Text('Phản đối kết quả')),
                const PopupMenuItem(value: 'finalize', child: Text('Chốt kết quả')),
                if (context.isAdminNow) ...const [
                  PopupMenuItem(value: 'resolve', child: Text('Phân xử (quản trị)')),
                  PopupMenuItem(value: 'rewards', child: Text('Thưởng tạo lập')),
                ],
              ],
            ),
          ],
          bottom: open
              ? Row(
                  children: [
                    Expanded(child: FilledButton(style: FilledButton.styleFrom(backgroundColor: _green), onPressed: () => _order(m, 'buy'), child: const Text('Mua'))),
                    const SizedBox(width: 8),
                    Expanded(child: FilledButton(style: FilledButton.styleFrom(backgroundColor: _red), onPressed: () => _order(m, 'sell'), child: const Text('Bán'))),
                  ],
                )
              : null,
          header: Padding(
            padding: const EdgeInsets.fromLTRB(16, 16, 16, 0),
            child: KCard(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(m.question ?? '', style: const TextStyle(fontWeight: FontWeight.w800, fontSize: 17)),
                  const SizedBox(height: 8),
                  Row(
                    children: [
                      Text('Có ${cents(m.yesPrice, share)}', style: const TextStyle(color: _green, fontWeight: FontWeight.w800, fontSize: 20)),
                      const SizedBox(width: 16),
                      Text('Không ${cents(m.noPrice, share)}', style: const TextStyle(color: _red, fontWeight: FontWeight.w800, fontSize: 20)),
                      const Spacer(),
                      StatusPill(m.status),
                    ],
                  ),
                  const SizedBox(height: 12),
                  SizedBox(height: 120, width: double.infinity, child: CustomPaint(painter: _Sparkline([for (final c in d.$3) (c.close ?? 0).toDouble()], context.appColor))),
                  const SizedBox(height: 8),
                  Wrap(
                    spacing: 6,
                    children: [
                      for (final (r, label) in const [('1d', '1N'), ('1w', '1T'), ('1m', '1Th'), ('all', 'Tất cả')])
                        ChoiceChip(label: Text(label), selected: _range == r, onSelected: (_) => setState(() => _range = r)),
                    ],
                  ),
                ],
              ),
            ),
          ),
          children: [
            Padding(
              padding: const EdgeInsets.fromLTRB(16, 12, 16, 0),
              child: SegmentedButton<String>(
                showSelectedIcon: false,
                segments: const [ButtonSegment(value: 'yes', label: Text('Cửa Có')), ButtonSegment(value: 'no', label: Text('Cửa Không'))],
                selected: {_outcome},
                onSelectionChanged: (s) => setState(() => _outcome = s.first),
              ),
            ),
            const SectionHeader('Sổ lệnh'),
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: 16),
              child: Row(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  for (final (title, levels, color) in [('Mua', book?.bids ?? const <PolyBookLevel>[], _green), ('Bán', book?.asks ?? const <PolyBookLevel>[], _red)])
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.stretch,
                        children: [
                          Text(title, style: TextStyle(fontWeight: FontWeight.w800, color: color)),
                          if (levels.isEmpty) const Padding(padding: EdgeInsets.symmetric(vertical: 8), child: Text('Trống', style: TextStyle(fontSize: 12))),
                          for (final l in levels.take(6))
                            Padding(
                              padding: const EdgeInsets.symmetric(vertical: 2),
                              child: Row(children: [Text(cents(l.price, share), style: TextStyle(color: color, fontWeight: FontWeight.w700)), const Spacer(), Text('${l.size ?? 0}'), const SizedBox(width: 12)]),
                            ),
                        ],
                      ),
                    ),
                ],
              ),
            ),
            InfoSection(rows: [
              ('Chênh lệch', book?.spread == null ? null : cents(book!.spread, share)),
              ('Giá giữa', book?.midpoint == null ? null : cents(book!.midpoint, share)),
              ('Khớp gần nhất', book?.lastTradePrice == null ? null : cents(book!.lastTradePrice, share)),
              ('Khối lượng', vnd(m.volume)),
              ('Kết thúc', formatDate(m.endTime)),
              ('Đề xuất', m.proposedOutcome),
              ('Hạn phản đối', m.disputeDeadline == null ? null : formatDate(m.disputeDeadline)),
              ('Kết quả', m.resolvedOutcome),
            ]),
            if (d.$6 != null && (d.$6!.dailyPool ?? 0) > 0)
              InfoSection(title: 'Thưởng tạo lập thị trường', rows: [
                ('Quỹ mỗi ngày', vnd(d.$6!.dailyPool)),
                ('Quỹ kỳ này', vnd(d.$6!.epochPool)),
                ('Chênh lệch tối đa', cents(d.$6!.maxSpread, share)),
                ('Lệnh tối thiểu', '${d.$6!.minSize ?? 0} cổ phần'),
              ]),
            const SectionHeader('Giao dịch gần đây'),
            if (d.$4.isEmpty) const Padding(padding: EdgeInsets.symmetric(horizontal: 20), child: Text('Chưa có giao dịch')),
            for (final t in d.$4.take(10))
              ListTile(
                dense: true,
                leading: Icon(t.takerSide == 'buy' ? Icons.trending_up : Icons.trending_down, color: t.takerSide == 'buy' ? _green : _red),
                title: Text('${t.takerSide == 'buy' ? 'Mua' : 'Bán'} ${t.size ?? 0} “${t.takerOutcome == 'no' ? 'Không' : 'Có'}” @ ${cents(t.yesPrice, share)}'),
                subtitle: Text(formatDate(t.createdAt)),
              ),
            const SectionHeader('Người nắm giữ nhiều nhất'),
            for (final (i, h) in d.$5.indexed)
              ListTile(dense: true, leading: CircleAvatar(radius: 14, child: Text('${i + 1}', style: const TextStyle(fontSize: 12))), title: Text(h.username ?? h.userId ?? ''), trailing: Text('${h.shares ?? 0} cp')),
            Paragraph(m.description, title: 'Mô tả'),
          ],
        );
      },
    );
  }
}

class PolyMarketsSearch extends StatelessWidget {
  const PolyMarketsSearch({super.key});

  @override
  Widget build(BuildContext context) {
    return SearchableList<PolyMarket>(
      hint: 'Tìm câu hỏi dự đoán',
      filters: () async => const ['open', 'closed', 'resolved'],
      load: (q, status) async => (await context.api.polymarket.polyMarkets(q: q.isEmpty ? null : q, status: status, limit: 50))?.items,
      emptyMessage: 'Không có thị trường phù hợp',
      emptyIcon: Icons.query_stats,
      itemBuilder: (context, m, _) => EntityTile(
        title: m.question ?? m.groupItemTitle ?? '',
        subtitle: 'Khối lượng ${vnd(m.volume)} • kết thúc ${shortDate(m.endTime)}',
        caption: 'Có ${cents(m.yesPrice, m.shareValue ?? 100)} • Không ${cents(m.noPrice, m.shareValue ?? 100)}',
        icon: Icons.query_stats,
        badges: [StatusPill(m.status)],
        onTap: () => openPage(context, PolyMarketPage(id: '${m.id}')),
      ),
    );
  }
}

class PolyRelated extends StatelessWidget {
  const PolyRelated({super.key, required this.eventId, required this.open});

  final String eventId;
  final void Function(PolyEvent e) open;

  @override
  Widget build(BuildContext context) {
    return AsyncBuilder<List<PolyEvent>?>(
      load: () => context.api.polymarket.polyRelatedEvents(id: eventId, limit: 6).catchError((_) => null),
      loading: const SizedBox.shrink(),
      builder: (context, list, _) => (list ?? const []).isEmpty
          ? const SizedBox.shrink()
          : Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                const SectionHeader('Sự kiện liên quan'),
                Strip<PolyEvent>(
                  items: list!,
                  height: 96,
                  itemWidth: 220,
                  itemBuilder: (context, e) => KCard(
                    onTap: () => open(e),
                    child: Text(e.title ?? '', maxLines: 3, overflow: TextOverflow.ellipsis, style: const TextStyle(fontWeight: FontWeight.w700)),
                  ),
                ),
              ],
            ),
    );
  }
}

Future<bool> polyEventAdmin(BuildContext context, PolyEvent e, String action) {
  final api = context.api.polymarket;
  final id = '${e.id ?? e.slug}';
  final markets = [for (final m in e.markets ?? const <PolyMarket>[]) ('${m.id}', m.groupItemTitle ?? m.question ?? '#${m.id}')];
  return switch (action) {
    'convert' => showFormSheet(
        context,
        title: 'Chuyển đổi vị thế',
        description: 'Gộp cổ phần “Không” của các lựa chọn đã chọn thành “Có” của các lựa chọn còn lại (sự kiện nhiều lựa chọn loại trừ).',
        fields: [
          KField('markets', 'Mã các lựa chọn (cách nhau dấu phẩy)', required: true, hint: markets.map((m) => '${m.$1}=${m.$2}').join(', ')),
          const KField('amount', 'Số cổ phần', kind: FieldKind.number, required: true),
        ],
        success: 'Đã chuyển đổi',
        onSubmit: (v) => api.polyConvertPositions(id: id, marketIds: v.ints('markets'), amount: v.integer('amount')!),
      ),
    'propose' || 'resolve' => showFormSheet(
        context,
        title: action == 'propose' ? 'Đề xuất lựa chọn thắng' : 'Phân xử sự kiện (quản trị)',
        fields: [KField('winner', 'Lựa chọn thắng', kind: FieldKind.choice, options: markets)],
        success: 'Đã gửi',
        onSubmit: (v) => action == 'propose' ? api.polyProposeEvent(id: id, winnerMarketId: v.integer('winner')!) : api.polyResolveEvent(id: id, winnerMarketId: v.integer('winner')!),
      ),
    _ => runAction(context, () => api.polyUnwatchEvent(id: id), success: 'Đã bỏ theo dõi'),
  };
}

class PolyAccount extends StatefulWidget {
  const PolyAccount({super.key});

  @override
  State<PolyAccount> createState() => _PolyAccountState();
}

class _PolyAccountState extends State<PolyAccount> {
  int _tab = 0;
  int _v = 0;

  @override
  Widget build(BuildContext context) {
    final api = context.api.polymarket;
    final me = '${context.userId}';
    return Column(
      children: [
        SizedBox(
          height: 56,
          child: ListView(
            scrollDirection: Axis.horizontal,
            padding: const EdgeInsets.fromLTRB(16, 12, 16, 0),
            children: [
              for (final (i, label) in const [(0, 'Hồ sơ'), (1, 'Lệnh'), (2, 'Đã đóng'), (3, 'Hoạt động'), (4, 'Thưởng')])
                Padding(padding: const EdgeInsets.only(right: 8), child: ChoiceChip(label: Text(label), selected: _tab == i, onSelected: (_) => setState(() => _tab = i))),
            ],
          ),
        ),
        Expanded(
          child: switch (_tab) {
            0 => AsyncBuilder<(PolyProfile?, PolyBalance?, PolyWallet?, PolyReferralSummary?, Object?)>(
                key: ValueKey('p$_v'),
                load: () async {
                  final r = await Future.wait<Object?>([
                    api.polyProfile(userId: me).catchError((_) => null),
                    api.polyBalance(userId: me).catchError((_) => null),
                    api.polyWallet(userId: me).catchError((_) => null),
                    api.polyReferral(userId: me).catchError((_) => null),
                    api.polyMyReferral().catchError((_) => null),
                  ]);
                  return (r[0] as PolyProfile?, r[1] as PolyBalance?, r[2] as PolyWallet?, r[3] as PolyReferralSummary?, r[4]);
                },
                builder: (context, d, _) {
                  final ref = d.$4;
                  final code = ref?.refCode ?? pick(d.$5, 'ref_code')?.toString();
                  return ListView(
                    padding: const EdgeInsets.only(bottom: 96),
                    children: [
                      ListTile(
                        leading: CircleAvatar(radius: 26, foregroundImage: (d.$1?.avatarUrl ?? '').startsWith('http') ? NetworkImage(d.$1!.avatarUrl!) : null, child: const Icon(Icons.person)),
                        title: Text(d.$1?.username ?? 'Chưa đặt tên hiển thị', style: const TextStyle(fontWeight: FontWeight.w800)),
                        subtitle: Text(d.$1?.bio ?? 'Thêm giới thiệu ngắn về bạn'),
                        trailing: IconButton(
                          icon: const Icon(Icons.edit_outlined),
                          onPressed: () async {
                            if (await showFormSheet(
                              context,
                              title: 'Hồ sơ dự đoán',
                              fields: [
                                KField('username', 'Tên hiển thị', initial: d.$1?.username),
                                KField('bio', 'Giới thiệu', kind: FieldKind.multiline, initial: d.$1?.bio),
                                KField('avatar', 'Ảnh đại diện (URL)', initial: d.$1?.avatarUrl),
                              ],
                              success: 'Đã lưu hồ sơ',
                              onSubmit: (v) => api.polyUpdateMyProfile(username: v.str('username'), bio: v.str('bio'), avatarUrl: v.str('avatar')),
                            )) {
                              setState(() => _v++);
                            }
                          },
                        ),
                      ),
                      InfoSection(title: 'Tài khoản giao dịch', rows: [
                        ('Khả dụng', vnd(d.$2?.available)),
                        ('Đang khoá trong lệnh', vnd(d.$2?.locked)),
                        ('Ví liên kết', d.$3?.walletId),
                        ('Số dư ví', d.$3 == null ? null : formatMoney(d.$3!.balance, d.$3!.currency)),
                      ]),
                      InfoSection(title: 'Giới thiệu bạn bè', rows: [
                        ('Mã của bạn', code),
                        ('Đã mời', '${ref?.referees ?? 0}'),
                        ('Hoa hồng', ref?.commissionBps == null ? null : '${(ref!.commissionBps! / 100).toStringAsFixed(2)}% phí giao dịch'),
                        ('Đã nhận', vnd(ref?.earnings)),
                        ('Người giới thiệu bạn', ref?.referredBy),
                      ]),
                      ActionButtons([
                        if (code != null)
                          OutlinedButton.icon(
                            icon: const Icon(Icons.copy),
                            label: const Text('Sao chép link mời'),
                            onPressed: () async {
                              await Clipboard.setData(ClipboardData(text: ref?.deepLink ?? code));
                              if (context.mounted) showSnack(context, 'Đã sao chép');
                            },
                          ),
                        if (ref?.referredBy == null)
                          OutlinedButton.icon(
                            icon: const Icon(Icons.redeem),
                            label: const Text('Nhập mã giới thiệu'),
                            onPressed: () async {
                              if (await showFormSheet(
                                context,
                                title: 'Mã giới thiệu',
                                fields: const [KField('code', 'Mã', required: true)],
                                success: 'Đã áp dụng mã',
                                onSubmit: (v) => api.polyRedeemReferral(refCode: v.str('code')!),
                              )) {
                                setState(() => _v++);
                              }
                            },
                          ),
                      ]),
                    ],
                  );
                },
              ),
            1 => AsyncList<PolyOrder>(
                key: ValueKey('o$_v'),
                load: () async => (await api.polyUserOrders(userId: me, limit: 50))?.items,
                emptyMessage: 'Chưa có lệnh',
                itemBuilder: (context, o, reload) => EntityTile(
                  title: '${o.side == 'buy' ? 'Mua' : 'Bán'} “${o.outcome == 'no' ? 'Không' : 'Có'}” • thị trường #${o.marketId}',
                  subtitle: '${o.filled ?? 0}/${o.size ?? 0} đã khớp • giá ${o.price ?? '-'} • ${formatDate(o.createdAt)}',
                  icon: o.side == 'buy' ? Icons.trending_up : Icons.trending_down,
                  badges: [StatusPill(o.status)],
                  onTap: () async {
                    final full = await api.polyOrder(orderId: o.id!).catchError((_) => null);
                    if (!context.mounted || full == null) return;
                    await showResult(
                      context,
                      icon: Icons.receipt_long,
                      title: 'Lệnh #${full.id}',
                      message: 'Kiểu ${full.type_} • ${full.timeInForce ?? ''}\nKhớp ${full.filled ?? 0}/${full.size ?? 0} • tiền khớp ${vnd(full.filledCash)}\nPhí ${vnd(full.fee)}${full.expiresAt == null ? '' : '\nHết hạn ${formatDate(full.expiresAt)}'}',
                    );
                  },
                  trailing: o.status == 'open' || o.status == 'partial' || o.status == 'pending'
                      ? IconButton(
                          tooltip: 'Huỷ lệnh',
                          icon: const Icon(Icons.cancel_outlined),
                          onPressed: () async {
                            if (await runAction(context, () => api.polyCancelOrder(orderId: '${o.id}'), success: 'Đã huỷ lệnh')) await reload();
                          },
                        )
                      : null,
                ),
              ),
            2 => AsyncList<PolyPosition>(
                key: const ValueKey('c'),
                load: () => api.polyClosedPositions(userId: me),
                emptyMessage: 'Chưa có vị thế đã đóng',
                itemBuilder: (context, p, _) => EntityTile(
                  title: 'Thị trường #${p.marketId} • “${p.outcome == 'no' ? 'Không' : 'Có'}”',
                  subtitle: '${p.shares ?? 0} cổ phần • vốn ${vnd(p.costBasis)}',
                  icon: Icons.inventory_2_outlined,
                  trailing: Text('${(p.realizedPnl ?? 0) >= 0 ? '+' : ''}${vnd(p.realizedPnl)}', style: TextStyle(fontWeight: FontWeight.w800, color: (p.realizedPnl ?? 0) >= 0 ? _green : _red)),
                  onTap: () => openPage(context, PolyMarketPage(id: '${p.marketId}')),
                ),
              ),
            3 => AsyncList<PolyActivity>(
                key: const ValueKey('a'),
                load: () => api.polyActivity(userId: me, limit: 50),
                emptyMessage: 'Chưa có hoạt động',
                spacing: 6,
                itemBuilder: (context, a, _) => EntityTile(
                  title: '${a.type_ ?? ''} • thị trường #${a.marketId ?? '-'}',
                  subtitle: '${a.side ?? ''} ${a.size ?? ''} “${a.outcome ?? ''}” @ ${a.price ?? '-'} • ${formatDate(a.createdAt)}',
                  caption: vnd(a.amount),
                  icon: Icons.history,
                ),
              ),
            _ => AsyncList<PolyRewardPayout>(
                key: const ValueKey('r'),
                load: () async => (await api.polyUserRewards(userId: me, limit: 50))?.items,
                header: AsyncBuilder<Object?>(
                  load: () => api.polyMyRewards().catchError((_) => null),
                  loading: const SizedBox.shrink(),
                  builder: (context, r, _) => AppBanner(
                    icon: Icons.emoji_events_rounded,
                    title: vnd(pickInt(r, 'total_paid') ?? pickInt(r, 'total')),
                    subtitle: 'Thưởng tạo lập thị trường đã nhận',
                  ),
                ),
                emptyMessage: 'Chưa có thưởng',
                itemBuilder: (context, p, _) => EntityTile(
                  title: 'Thị trường #${p.marketId}',
                  subtitle: 'Kỳ ${formatDate(p.epochStart)} • điểm ${p.score?.toStringAsFixed(2) ?? '-'}',
                  caption: '+${vnd(p.amount)}',
                  icon: Icons.paid_outlined,
                ),
              ),
          },
        ),
      ],
    );
  }
}

class PolyCreatorPage extends StatelessWidget {
  const PolyCreatorPage({super.key});

  @override
  Widget build(BuildContext context) {
    final api = context.api.polymarket;
    return AsyncBuilder<PolyExchangeSummary?>(
      load: () => api.polyExchange().catchError((_) => null),
      loading: const Scaffold(body: SkeletonList()),
      builder: (context, ex, reload) => DetailPage(
        title: 'Tạo thị trường',
        header: AppBanner(
          icon: Icons.add_chart,
          title: 'Mở sự kiện dự đoán mới',
          subtitle: 'Mỗi dòng trong “Các lựa chọn” là một câu hỏi Có/Không. Nhiều lựa chọn loại trừ nhau thì bật “Chỉ một lựa chọn đúng”.',
        ),
        children: [
          ActionButtons([
            FilledButton.icon(
              icon: const Icon(Icons.add),
              label: const Text('Tạo sự kiện'),
              onPressed: () => showFormSheet(
                context,
                title: 'Sự kiện dự đoán',
                fields: const [
                  KField('title', 'Tiêu đề', required: true),
                  KField('category', 'Chủ đề', hint: 'Thể thao, Kinh tế…'),
                  KField('tags', 'Thẻ (phân cách dấu phẩy)'),
                  KField('end', 'Kết thúc', kind: FieldKind.datetime, required: true),
                  KField('image', 'Ảnh (URL)'),
                  KField('markets', 'Các lựa chọn (mỗi dòng một câu)', kind: FieldKind.multiline, required: true, initial: 'Có'),
                  KField('neg', 'Chỉ một lựa chọn đúng', kind: FieldKind.toggle, initial: 'false'),
                  KField('description', 'Luật phân xử', kind: FieldKind.multiline, required: true),
                ],
                submitLabel: 'Tạo',
                success: 'Đã tạo sự kiện',
                onSubmit: (v) => api.polyCreateEvent(
                  input: PolyEventInput(
                    title: v.str('title'),
                    category: v.str('category'),
                    tags: (v.str('tags') ?? '').split(',').map((t) => t.trim()).where((t) => t.isNotEmpty).toList(),
                    endDate: v.str('end'),
                    imageUrl: v.str('image'),
                    negRisk: v.flag('neg'),
                    description: v.str('description'),
                    markets: [
                      for (final q in (v.str('markets') ?? '').split('\n').map((s) => s.trim()).where((s) => s.isNotEmpty))
                        PolyMarketInput(question: q, groupItemTitle: q, endTime: v.str('end'), shareValue: 100),
                    ],
                  ),
                ),
              ),
            ),
          ]),
          InfoSection(title: 'Sàn (quản trị)', rows: [('Doanh thu phí', vnd(ex?.revenue)), ('Tiền cọc đang giữ', vnd(ex?.bondHeld))]),
          ActionButtons([
            OutlinedButton.icon(
              icon: const Icon(Icons.savings_outlined),
              label: const Text('Cấp vốn cho tài khoản'),
              onPressed: () => showFormSheet(
                context,
                title: 'Cấp vốn (quản trị)',
                fields: const [KField('user', 'Mã người dùng', required: true), KField('amount', 'Số tiền', kind: FieldKind.number, required: true)],
                success: 'Đã cấp vốn',
                onSubmit: (v) async {
                  await api.polyFundExchange(userId: v.str('user')!, amount: v.integer('amount')!);
                  await reload();
                },
              ),
            ),
          ]),
        ],
      ),
    );
  }
}
