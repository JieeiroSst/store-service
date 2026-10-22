import 'dart:math' as math;

import 'package:flutter/material.dart';

import '../../../gen/models.g.dart';
import '../kit/kit.dart';

class LuckyApp extends StatelessWidget {
  const LuckyApp({super.key});

  @override
  Widget build(BuildContext context) {
    return ServiceApp(
      serviceId: 'lucky',
      tabs: [
        AppTab('Trò chơi', Icons.casino_outlined, (_) => const _Games()),
        AppTab('Điểm danh', Icons.event_available_outlined, (_) => const _Checkin()),
        AppTab('Lượt chơi', Icons.history, (context) => AsyncList<LuckyPlay>(
              load: () async => (await context.api.lucky.luckyMyPlays(limit: 50))?.items,
              emptyMessage: 'Bạn chưa chơi lượt nào',
              emptyIcon: Icons.casino_outlined,
              itemBuilder: (context, p, _) => EntityTile(
                title: p.prizeName ?? 'Chúc may mắn lần sau',
                subtitle: '${_gameName(p.gameType)} • ${formatDate(p.createdAt)}',
                icon: (p.points ?? 0) > 0 ? Icons.emoji_events_rounded : Icons.sentiment_neutral_rounded,
                caption: (p.points ?? 0) > 0 ? '+${p.points} điểm' : null,
              ),
            )),
      ],
    );
  }
}

String _gameName(String? type) => switch (type) {
      'wheel' || 'spin' => 'Vòng quay may mắn',
      'scratch' => 'Cào thẻ',
      'box' || 'gift_box' || 'pick' => 'Mở hộp quà',
      'slot' => 'Máy xèng',
      'dice' => 'Tung xúc xắc',
      'checkin' => 'Điểm danh',
      _ => type ?? 'Trò chơi',
    };

IconData _gameIcon(String? type) => switch (type) {
      'wheel' || 'spin' => Icons.album_rounded,
      'scratch' => Icons.style_rounded,
      'box' || 'gift_box' || 'pick' => Icons.card_giftcard_rounded,
      'slot' => Icons.casino_rounded,
      _ => Icons.celebration_rounded,
    };

class _Games extends StatelessWidget {
  const _Games();

  @override
  Widget build(BuildContext context) {
    return AsyncList<LuckyCampaign>(
      load: () async => (await context.api.lucky.luckyGames())?.where((g) => g.gameType != 'checkin').toList(),
      header: const AppBanner(
        icon: Icons.auto_awesome_rounded,
        title: 'Thử vận may hôm nay',
        subtitle: 'Mỗi ngày đều có lượt chơi miễn phí, điểm thưởng cộng thẳng vào tài khoản',
      ),
      emptyMessage: 'Chưa có trò chơi nào đang diễn ra',
      emptyIcon: Icons.casino_outlined,
      itemBuilder: (context, g, _) => EntityTile(
        title: g.name_ ?? _gameName(g.gameType),
        subtitle: g.description ?? _gameName(g.gameType),
        icon: _gameIcon(g.gameType),
        caption: '${g.dailySpins ?? 0} lượt miễn phí mỗi ngày',
        badges: [Pill('${(g.prizes ?? const []).length} giải thưởng', icon: Icons.emoji_events_outlined)],
        onTap: () => openPage(context, LuckyGamePage(game: g.gameType!)),
      ),
    );
  }
}

class LuckyGamePage extends StatefulWidget {
  const LuckyGamePage({super.key, required this.game});

  final String game;

  @override
  State<LuckyGamePage> createState() => _LuckyGamePageState();
}

class _LuckyGamePageState extends State<LuckyGamePage> with SingleTickerProviderStateMixin {
  late final _spin = AnimationController(vsync: this, duration: const Duration(milliseconds: 1600));
  LuckyPlayerStatus? _status;
  bool _busy = false;
  int? _choice;

  @override
  void dispose() {
    _spin.dispose();
    super.dispose();
  }

  Future<void> _play() async {
    setState(() => _busy = true);
    _spin.repeat();
    try {
      final r = await context.api.lucky.luckyPlay(game: widget.game, choice: _choice, idempotencyKey: newIdempotencyKey());
      await Future<void>.delayed(const Duration(milliseconds: 900));
      if (!mounted) return;
      setState(() => _status = r?.status ?? _status);
      final won = r?.won == true;
      await showResult(
        context,
        icon: won ? Icons.emoji_events_rounded : Icons.sentiment_satisfied_alt_rounded,
        title: won ? 'Chúc mừng!' : 'Tiếc quá',
        message: won ? 'Bạn nhận được ${r?.play?.prizeName ?? 'phần thưởng'}${(r?.play?.points ?? 0) > 0 ? ' (+${r!.play!.points} điểm)' : ''}' : 'Chúc bạn may mắn lần sau nhé!',
      );
    } catch (e) {
      if (mounted) showSnack(context, errorMessage(e));
    } finally {
      _spin.stop();
      if (mounted) setState(() => _busy = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final api = context.api.lucky;
    return AsyncBuilder<(LuckyCampaign?, LuckyPlayerStatus?)>(
      load: () async {
        final r = await Future.wait<Object?>([api.luckyGame(game: widget.game), api.luckyMyStatus(game: widget.game).catchError((_) => null)]);
        return (r[0] as LuckyCampaign?, r[1] as LuckyPlayerStatus?);
      },
      loading: const Scaffold(body: SkeletonList()),
      builder: (context, data, reload) {
        final g = data.$1;
        final status = _status ?? data.$2;
        final prizes = g?.prizes ?? const <LuckyPrize>[];
        final remaining = status?.remaining ?? 0;
        final pick = widget.game == 'box' || widget.game == 'gift_box' || widget.game == 'pick';
        return DetailPage(
          title: g?.name_ ?? _gameName(widget.game),
          header: AppBanner(
            icon: _gameIcon(widget.game),
            title: g?.name_ ?? _gameName(widget.game),
            subtitle: g?.description,
            children: [
              Center(
                child: RotationTransition(
                  turns: _spin,
                  child: Container(
                    width: 150,
                    height: 150,
                    decoration: BoxDecoration(
                      shape: BoxShape.circle,
                      gradient: SweepGradient(colors: [
                        for (var i = 0; i < math.max(prizes.length, 6); i++) HSLColor.fromAHSL(1, (i * 360 / math.max(prizes.length, 6)) % 360, 0.75, 0.62).toColor(),
                        HSLColor.fromAHSL(1, 0, 0.75, 0.62).toColor(),
                      ]),
                      border: Border.all(color: Colors.white, width: 5),
                    ),
                    child: const Icon(Icons.star_rounded, color: Colors.white, size: 54),
                  ),
                ),
              ),
              const SizedBox(height: 14),
              StatRow([
                StatTile(label: 'Còn lại', value: '$remaining', icon: Icons.bolt, onDark: true),
                StatTile(label: 'Mỗi ngày', value: '${status?.dailySpins ?? g?.dailySpins ?? 0}', icon: Icons.today, onDark: true),
                StatTile(label: 'Thưởng', value: '${status?.usage?.creditBalance ?? 0}', icon: Icons.card_giftcard, onDark: true),
              ]),
            ],
          ),
          bottom: LoadingButton(
            label: remaining > 0 ? 'Chơi ngay' : 'Hết lượt hôm nay',
            loading: _busy,
            onPressed: remaining > 0 && (!pick || _choice != null) ? _play : null,
          ),
          children: [
            if (pick) ...[
              const SectionHeader('Chọn một hộp quà'),
              Padding(
                padding: const EdgeInsets.symmetric(horizontal: 16),
                child: Wrap(
                  spacing: 10,
                  runSpacing: 10,
                  children: [
                    for (var i = 1; i <= 6; i++)
                      ChoiceChip(
                        avatar: const Icon(Icons.card_giftcard_rounded, size: 18),
                        label: Text('Hộp $i'),
                        selected: _choice == i,
                        onSelected: (_) => setState(() => _choice = i),
                      ),
                  ],
                ),
              ),
            ],
            const SectionHeader('Giải thưởng'),
            for (final p in prizes)
              Padding(
                padding: const EdgeInsets.fromLTRB(16, 0, 16, 8),
                child: EntityTile(
                  title: p.name_ ?? 'Giải thưởng',
                  imageUrl: p.imageUrl,
                  icon: p.type_ == 'points' ? Icons.stars_rounded : Icons.redeem_rounded,
                  caption: (p.points ?? 0) > 0 ? '${p.points} điểm' : null,
                  subtitle: p.quantity == null ? null : 'Còn ${(p.quantity ?? 0) - (p.awarded ?? 0)} / ${p.quantity}',
                ),
              ),
            const SectionHeader('Người trúng gần đây'),
            SizedBox(
              height: 220,
              child: AsyncBuilder<List<LuckyWinner>?>(
                load: () => api.luckyWinners(game: widget.game, limit: 10),
                loading: const Center(child: CircularProgressIndicator()),
                builder: (context, winners, _) => ListView(
                  padding: const EdgeInsets.symmetric(horizontal: 16),
                  children: [
                    if ((winners ?? const []).isEmpty) const Center(child: Padding(padding: EdgeInsets.all(24), child: Text('Chưa có ai trúng thưởng'))),
                    for (final w in winners ?? const <LuckyWinner>[])
                      ListTile(
                        contentPadding: EdgeInsets.zero,
                        leading: const CircleAvatar(child: Icon(Icons.person)),
                        title: Text(w.user ?? 'Người chơi'),
                        subtitle: Text(w.prizeName ?? ''),
                        trailing: Text(shortDate(w.at)),
                      ),
                  ],
                ),
              ),
            ),
          ],
        );
      },
    );
  }
}

class _Checkin extends StatefulWidget {
  const _Checkin();

  @override
  State<_Checkin> createState() => _CheckinState();
}

class _CheckinState extends State<_Checkin> {
  bool _busy = false;

  @override
  Widget build(BuildContext context) {
    final api = context.api.lucky;
    return AsyncBuilder<LuckyCheckinStatus?>(
      load: () => api.luckyCheckinStatus(),
      builder: (context, s, reload) {
        final rewards = s?.rewards ?? const <int>[];
        final day = s?.cycleDay ?? 0;
        return RefreshIndicator(
          onRefresh: reload,
          child: ListView(
            padding: const EdgeInsets.only(bottom: 24),
            children: [
              AppBanner(
                icon: Icons.local_fire_department_rounded,
                title: 'Chuỗi ${s?.streak ?? 0} ngày',
                subtitle: s?.checkedInToday == true ? 'Hôm nay bạn đã điểm danh. Hẹn gặp lại ngày mai!' : 'Điểm danh hôm nay để nhận ${s?.nextReward ?? 0} điểm',
                children: [
                  SizedBox(
                    width: double.infinity,
                    child: FilledButton(
                      style: FilledButton.styleFrom(backgroundColor: Colors.white, foregroundColor: context.appColor),
                      onPressed: s?.checkedInToday == true || _busy
                          ? null
                          : () async {
                              setState(() => _busy = true);
                              final ok = await runAction(context, () => api.luckyCheckIn(), success: 'Điểm danh thành công');
                              if (mounted) setState(() => _busy = false);
                              if (ok) await reload();
                            },
                      child: Text(s?.checkedInToday == true ? 'Đã điểm danh' : 'Điểm danh ngay'),
                    ),
                  ),
                ],
              ),
              if (rewards.isNotEmpty) ...[
                const SectionHeader('Phần thưởng theo ngày'),
                Padding(
                  padding: const EdgeInsets.symmetric(horizontal: 16),
                  child: GridView.count(
                    crossAxisCount: 4,
                    shrinkWrap: true,
                    physics: const NeverScrollableScrollPhysics(),
                    mainAxisSpacing: 8,
                    crossAxisSpacing: 8,
                    childAspectRatio: 0.95,
                    children: [
                      for (final (i, pts) in rewards.indexed)
                        Container(
                          decoration: BoxDecoration(
                            color: i < day ? context.appColor : context.appColor.withValues(alpha: 0.08),
                            borderRadius: BorderRadius.circular(16),
                          ),
                          child: Column(
                            mainAxisAlignment: MainAxisAlignment.center,
                            children: [
                              Text('Ngày ${i + 1}', style: TextStyle(fontSize: 11, color: i < day ? Colors.white70 : null)),
                              Icon(i < day ? Icons.check_circle_rounded : Icons.stars_rounded, color: i < day ? Colors.white : context.appColor),
                              Text('+$pts', style: TextStyle(fontWeight: FontWeight.w800, color: i < day ? Colors.white : null)),
                            ],
                          ),
                        ),
                    ],
                  ),
                ),
              ],
              const SectionHeader('Lịch sử điểm danh'),
              SizedBox(
                height: 300,
                child: AsyncBuilder<LuckyCheckinPage?>(
                  load: () => api.luckyCheckinHistory(limit: 30),
                  loading: const Center(child: CircularProgressIndicator()),
                  builder: (context, page, _) => Column(
                    children: [
                      if ((page?.items ?? const []).isEmpty) const Padding(padding: EdgeInsets.all(24), child: Text('Chưa có lần điểm danh nào')),
                      for (final c in (page?.items ?? const <LuckyCheckin>[]).take(5))
                        ListTile(
                          leading: const Icon(Icons.event_available_rounded),
                          title: Text(shortDate(c.day ?? c.createdAt)),
                          subtitle: Text('Chuỗi ${c.streak ?? 0} ngày'),
                          trailing: Text('+${c.points ?? 0}', style: const TextStyle(fontWeight: FontWeight.w800)),
                        ),
                    ],
                  ),
                ),
              ),
            ],
          ),
        );
      },
    );
  }
}
