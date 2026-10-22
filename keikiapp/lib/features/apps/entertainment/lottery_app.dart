import 'package:flutter/material.dart';

import '../../../gen/models.g.dart';
import '../kit/kit.dart';

const _provinces = [
  'Miền Bắc', 'TP. Hồ Chí Minh', 'Đồng Nai', 'Cần Thơ', 'Sóc Trăng', 'Vũng Tàu', 'Bến Tre', 'Bạc Liêu', 'Tây Ninh', 'An Giang',
  'Bình Thuận', 'Vĩnh Long', 'Bình Dương', 'Trà Vinh', 'Long An', 'Bình Phước', 'Hậu Giang', 'Tiền Giang', 'Kiên Giang', 'Đà Lạt',
  'Đà Nẵng', 'Khánh Hòa', 'Thừa Thiên Huế', 'Phú Yên', 'Quảng Nam', 'Đắk Lắk', 'Bình Định', 'Quảng Trị', 'Quảng Bình', 'Gia Lai', 'Ninh Thuận',
];

const _prizeOrder = ['db', 'g1', 'g2', 'g3', 'g4', 'g5', 'g6', 'g7', 'g8'];

String _prizeLabel(String key) => switch (key.toLowerCase()) {
      'db' || 'dac_biet' || 'special' => 'Giải đặc biệt',
      'g1' => 'Giải nhất',
      'g2' => 'Giải nhì',
      'g3' => 'Giải ba',
      'g4' => 'Giải tư',
      'g5' => 'Giải năm',
      'g6' => 'Giải sáu',
      'g7' => 'Giải bảy',
      'g8' => 'Giải tám',
      _ => key,
    };

String _iso(DateTime d) => d.toIso8601String().substring(0, 10);

class LotteryApp extends StatelessWidget {
  const LotteryApp({super.key});

  @override
  Widget build(BuildContext context) {
    return ServiceApp(
      serviceId: 'lottery',
      tabs: [
        AppTab('Kết quả', Icons.format_list_numbered, (_) => const _Results()),
        AppTab('Dò vé', Icons.manage_search, (_) => const _Check()),
      ],
    );
  }
}

class _Picker extends StatelessWidget {
  const _Picker({required this.date, required this.province, required this.onDate, required this.onProvince});

  final DateTime date;
  final String province;
  final ValueChanged<DateTime> onDate;
  final ValueChanged<String> onProvince;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.fromLTRB(16, 14, 16, 0),
      child: Row(
        children: [
          Expanded(
            child: OutlinedButton.icon(
              icon: const Icon(Icons.event),
              label: Text(shortDate(date.toIso8601String())),
              onPressed: () async {
                final now = DateTime.now();
                final d = await showDatePicker(context: context, initialDate: date, firstDate: DateTime(2010), lastDate: now);
                if (d != null) onDate(d);
              },
            ),
          ),
          const SizedBox(width: 8),
          Expanded(
            child: DropdownButtonFormField<String>(
              initialValue: province,
              isExpanded: true,
              isDense: true,
              decoration: const InputDecoration(prefixIcon: Icon(Icons.place_outlined), isDense: true),
              items: [for (final p in _provinces) DropdownMenuItem(value: p, child: Text(p, overflow: TextOverflow.ellipsis))],
              onChanged: (v) => onProvince(v ?? province),
            ),
          ),
        ],
      ),
    );
  }
}

class _Results extends StatefulWidget {
  const _Results();

  @override
  State<_Results> createState() => _ResultsState();
}

class _ResultsState extends State<_Results> {
  DateTime _date = DateTime.now().subtract(const Duration(days: 1));
  String _province = _provinces.first;

  @override
  Widget build(BuildContext context) {
    return Column(
      children: [
        _Picker(
          date: _date,
          province: _province,
          onDate: (d) => setState(() => _date = d),
          onProvince: (p) => setState(() => _province = p),
        ),
        Expanded(
          child: AsyncBuilder<LotteryResult?>(
            key: ValueKey('$_date$_province'),
            load: () => context.api.lottery.lotteryResults(date: _iso(_date), province: _province),
            builder: (context, r, reload) {
              final prizes = r?.prizes ?? const <String, dynamic>{};
              if (prizes.isEmpty) return const EmptyView(message: 'Chưa có kết quả cho ngày này', icon: Icons.hourglass_empty);
              final keys = prizes.keys.toList()
                ..sort((a, b) {
                  final ia = _prizeOrder.indexOf(a.toLowerCase());
                  final ib = _prizeOrder.indexOf(b.toLowerCase());
                  return (ia < 0 ? 99 : ia).compareTo(ib < 0 ? 99 : ib);
                });
              return ListView(
                padding: const EdgeInsets.fromLTRB(16, 12, 16, 24),
                children: [
                  Text('${r?.province ?? _province} • ${shortDate(r?.date)}', style: Theme.of(context).textTheme.titleMedium?.copyWith(fontWeight: FontWeight.w800)),
                  const SizedBox(height: 10),
                  KCard(
                    padding: EdgeInsets.zero,
                    child: Column(
                      children: [
                        for (final (i, k) in keys.indexed)
                          Container(
                            color: i.isEven ? context.appColor.withValues(alpha: 0.04) : null,
                            padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 12),
                            child: Row(
                              crossAxisAlignment: CrossAxisAlignment.start,
                              children: [
                                SizedBox(width: 104, child: Text(_prizeLabel(k), style: const TextStyle(fontWeight: FontWeight.w600))),
                                Expanded(
                                  child: Wrap(
                                    alignment: WrapAlignment.end,
                                    spacing: 12,
                                    runSpacing: 4,
                                    children: [
                                      for (final n in (prizes[k] is List ? prizes[k] as List : [prizes[k]]))
                                        Text(
                                          '$n',
                                          style: TextStyle(
                                            fontSize: _prizeOrder.indexOf(k.toLowerCase()) == 0 ? 24 : 17,
                                            fontWeight: FontWeight.w800,
                                            letterSpacing: 1,
                                            color: _prizeOrder.indexOf(k.toLowerCase()) == 0 ? const Color(0xFFDC2626) : null,
                                          ),
                                        ),
                                    ],
                                  ),
                                ),
                              ],
                            ),
                          ),
                      ],
                    ),
                  ),
                  if (r?.source != null) Padding(padding: const EdgeInsets.only(top: 8), child: Text('Nguồn: ${r!.source}', style: Theme.of(context).textTheme.bodySmall)),
                ],
              );
            },
          ),
        ),
      ],
    );
  }
}

class _Check extends StatefulWidget {
  const _Check();

  @override
  State<_Check> createState() => _CheckState();
}

class _CheckState extends State<_Check> {
  DateTime _date = DateTime.now().subtract(const Duration(days: 1));
  String _province = _provinces.first;
  final _number = TextEditingController();
  LotteryCheck? _result;
  bool _busy = false;

  @override
  void dispose() {
    _number.dispose();
    super.dispose();
  }

  Future<void> _check() async {
    if (_number.text.trim().isEmpty) return;
    setState(() => _busy = true);
    try {
      final r = await context.api.lottery.lotteryCheck(date: _iso(_date), province: _province, ticketNumber: _number.text.trim());
      setState(() => _result = r);
    } catch (e) {
      if (mounted) showSnack(context, errorMessage(e));
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final r = _result;
    return ListView(
      padding: const EdgeInsets.only(bottom: 24),
      children: [
        _Picker(date: _date, province: _province, onDate: (d) => setState(() => _date = d), onProvince: (p) => setState(() => _province = p)),
        Padding(
          padding: const EdgeInsets.fromLTRB(16, 12, 16, 0),
          child: TextField(
            controller: _number,
            keyboardType: TextInputType.number,
            maxLength: 6,
            textAlign: TextAlign.center,
            style: const TextStyle(fontSize: 28, fontWeight: FontWeight.w800, letterSpacing: 6),
            decoration: const InputDecoration(hintText: '123456', counterText: ''),
            onSubmitted: (_) => _check(),
          ),
        ),
        Padding(padding: const EdgeInsets.fromLTRB(16, 12, 16, 0), child: LoadingButton(label: 'Dò vé', icon: Icons.search, loading: _busy, onPressed: _check)),
        if (r != null)
          AppBanner(
            icon: r.won == true ? Icons.emoji_events_rounded : (r.resultFound == false ? Icons.hourglass_empty : Icons.sentiment_neutral_rounded),
            title: r.resultFound == false ? 'Chưa có kết quả' : (r.won == true ? 'Chúc mừng, vé trúng thưởng!' : 'Vé chưa trúng thưởng'),
            subtitle: r.won == true ? 'Giải cao nhất: ${_prizeLabel(r.highestPrize ?? '')}' : 'Chúc bạn may mắn lần sau',
            children: [
              for (final d in r.detail ?? const <LotteryPrizeDetail>[])
                Padding(
                  padding: const EdgeInsets.only(bottom: 6),
                  child: Text('• ${_prizeLabel(d.prize ?? '')}: ${d.matched ?? ''} ${d.rule == null ? '' : '(${d.rule})'}', style: const TextStyle(color: Colors.white)),
                ),
            ],
          ),
      ],
    );
  }
}
