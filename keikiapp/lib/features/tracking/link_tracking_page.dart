import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

import 'package:go_router/go_router.dart';

import '../../app/deep_links.dart';
import '../../core/config/env.dart';
import '../../core/scan/scan.dart';
import '../../core/tracking/attribution_cubit.dart';
import '../../core/tracking/link_touch.dart';
import '../../core/network/graphql_client.dart';
import '../../gen/models.g.dart';
import '../apps/kit/kit.dart';

enum LinkKind { branch, appScheme, payment, website, referralPage, other }

class LinkInspection {
  const LinkInspection({required this.uri, required this.kind, this.touch, this.route});

  final Uri uri;
  final LinkKind kind;
  final LinkTouch? touch;
  final String? route;

  String get kindLabel => switch (kind) {
        LinkKind.branch => 'Link Branch (${uri.host})',
        LinkKind.appScheme => 'Deep link keikibook://',
        LinkKind.payment => 'QR thanh toán keiki://pay',
        LinkKind.website => 'Link website keikibook.vn',
        LinkKind.referralPage => 'Link mời qua referral-service (/r/…)',
        LinkKind.other => 'Link ngoài, app không xử lý',
      };

  static LinkInspection? of(String input) {
    final uri = Uri.tryParse(input.trim());
    if (uri == null || !uri.hasScheme) return null;
    final kind = Env.isBranchHost(uri.host)
        ? LinkKind.branch
        : uri.scheme == 'keikibook'
            ? LinkKind.appScheme
            : uri.scheme == 'keiki' && uri.host == 'pay'
                ? LinkKind.payment
                : uri.pathSegments.length == 2 && uri.pathSegments.first == 'r'
                    ? LinkKind.referralPage
                    : keikiHosts.contains(uri.host)
                        ? LinkKind.website
                        : LinkKind.other;
    return LinkInspection(uri: uri, kind: kind, touch: LinkTouch.fromUri(uri), route: kind == LinkKind.branch ? null : routeForCode(uri.toString()));
  }
}

class LinkTrackingPage extends StatefulWidget {
  const LinkTrackingPage({super.key, this.initialLink});

  final String? initialLink;

  @override
  State<LinkTrackingPage> createState() => _LinkTrackingPageState();
}

class _LinkTrackingPageState extends State<LinkTrackingPage> {
  late final _link = TextEditingController(text: widget.initialLink);
  LinkInspection? _inspection;
  ReferralReferralLink? _refLink;
  ReferralReferralStatusResponse? _refStatus;
  String? _refError;
  bool _checking = false;
  String? _deviceId;

  @override
  void initState() {
    super.initState();
    context.read<GraphQLClient>().storage.deviceId().then((id) {
      if (mounted) setState(() => _deviceId = id);
    }).catchError((_) {});
    if (widget.initialLink != null) WidgetsBinding.instance.addPostFrameCallback((_) => _inspect());
  }

  @override
  void dispose() {
    _link.dispose();
    super.dispose();
  }

  Future<void> _inspect() async {
    final i = LinkInspection.of(_link.text);
    setState(() {
      _inspection = i;
      _refLink = null;
      _refStatus = null;
      _refError = null;
    });
    if (i == null) {
      showSnack(context, 'Không đọc được link, hãy dán đầy đủ cả https:// hoặc keikibook://');
      return;
    }
    final ref = i.touch?.refCode;
    if (ref == null) return;
    final api = context.api.referral;
    setState(() => _checking = true);
    try {
      final r = await Future.wait<Object?>([api.referralLink(refCode: ref), api.referralStatus(refCode: ref).catchError((_) => null)]);
      if (!mounted) return;
      setState(() {
        _refLink = r[0] as ReferralReferralLink?;
        _refStatus = r[1] as ReferralReferralStatusResponse?;
      });
    } catch (e) {
      if (mounted) setState(() => _refError = errorMessage(e));
    } finally {
      if (mounted) setState(() => _checking = false);
    }
  }

  void _simulate() {
    final i = _inspection;
    if (i == null) return;
    final cubit = context.read<AttributionCubit>();
    if (i.kind == LinkKind.branch) {
      if (!cubit.state.enabled) {
        showSnack(context, 'Branch chưa bật trong bản build này nên không giải được link Branch');
        return;
      }
      cubit.source.handle(i.uri.toString());
      showSnack(context, 'Đã gửi link cho Branch SDK, kết quả sẽ hiện ở Lịch sử');
      return;
    }
    cubit.observeUri(i.uri, source: LinkTouchSource.manual);
    final route = i.route;
    if (route != null) {
      Navigator.of(context).popUntil((r) => r.isFirst);
      openScanned(context, i.uri.toString());
    } else {
      showSnack(context, 'Đã ghi nhận vào luồng tracking');
    }
  }

  Future<void> _showJson(String title, Object? value) => openPage(
        context,
        DetailPage(
          title: title,
          actions: [
            IconButton(
              tooltip: 'Sao chép',
              icon: const Icon(Icons.copy_all_outlined),
              onPressed: () {
                Clipboard.setData(ClipboardData(text: const JsonEncoder.withIndent('  ').convert(value)));
                showSnack(context, 'Đã sao chép');
              },
            ),
          ],
          children: [
            Padding(
              padding: const EdgeInsets.all(16),
              child: SelectableText(
                const JsonEncoder.withIndent('  ').convert(value),
                style: const TextStyle(fontFamily: 'monospace', fontSize: 12),
              ),
            ),
          ],
        ),
      );

  Future<void> _params(bool first) async {
    final cubit = context.read<AttributionCubit>();
    try {
      final p = first ? await cubit.source.firstParams() : await cubit.source.latestParams();
      if (mounted) await _showJson(first ? 'Tham số lần cài đầu' : 'Tham số gần nhất', {for (final e in p.entries) '${e.key}': e.value is num || e.value is bool || e.value == null ? e.value : '${e.value}'});
    } catch (e) {
      if (mounted) showSnack(context, errorMessage(e));
    }
  }

  @override
  Widget build(BuildContext context) {
    final state = context.watch<AttributionCubit>().state;
    final theme = Theme.of(context);
    final i = _inspection;
    final t = i?.touch;
    final me = '${context.userId}';
    return DetailPage(
      title: 'Kiểm tra link & tracking',
      header: AppBanner(
        icon: state.enabled ? Icons.link_rounded : Icons.link_off_rounded,
        title: state.enabled ? (state.initError == null ? 'Branch đang hoạt động' : 'Branch lỗi khởi tạo') : 'Branch đang tắt',
        subtitle: state.initError ?? state.sourceLabel,
      ),
      children: [
        InfoSection(title: 'Thiết bị', rows: [
          ('Nguồn link', state.sourceLabel),
          ('Domain Branch', '${Env.branchLinkDomain} • ${Env.branchAlternateDomain}'),
          ('Device ID gửi server', _deviceId),
          ('Người dùng', me == '0' ? 'Chưa đăng nhập' : me),
        ]),
        const SectionHeader('Kiểm tra một link'),
        Padding(
          padding: const EdgeInsets.symmetric(horizontal: 16),
          child: TextField(
            controller: _link,
            minLines: 1,
            maxLines: 3,
            keyboardType: TextInputType.url,
            decoration: InputDecoration(
              hintText: 'https://${Env.branchLinkDomain}/… hoặc keikibook://…',
              prefixIcon: const Icon(Icons.link),
              suffixIcon: Row(
                mainAxisSize: MainAxisSize.min,
                children: [
                  IconButton(
                    tooltip: 'Dán',
                    icon: const Icon(Icons.content_paste),
                    onPressed: () async {
                      final d = await Clipboard.getData(Clipboard.kTextPlain);
                      if (d?.text != null) _link.text = d!.text!.trim();
                    },
                  ),
                  IconButton(
                    tooltip: 'Quét QR',
                    icon: const Icon(Icons.qr_code_scanner),
                    onPressed: () async {
                      final code = await scanCode(context, title: 'Quét link cần kiểm tra');
                      if (code == null) return;
                      _link.text = code;
                      await _inspect();
                    },
                  ),
                ],
              ),
            ),
            onSubmitted: (_) => _inspect(),
          ),
        ),
        ActionButtons([
          FilledButton.icon(onPressed: _checking ? null : _inspect, icon: const Icon(Icons.fact_check_outlined), label: const Text('Kiểm tra')),
          if (i != null) OutlinedButton.icon(onPressed: _simulate, icon: const Icon(Icons.play_arrow_rounded), label: const Text('Mở như người dùng bấm')),
        ]),
        if (i != null) ...[
          InfoSection(title: 'App đọc được', rows: [
            ('Loại link', i.kindLabel),
            ('Mở màn', i.kind == LinkKind.branch ? 'Do Branch trả về (\$deeplink_path)' : (i.route ?? 'Không có')),
            ('Mã mời', t?.refCode),
            ('click_id', t?.clickId),
            for (final k in utmKeys) (k, t?.utm[k]),
          ]),
          if (_checking) const Padding(padding: EdgeInsets.all(16), child: LinearProgressIndicator()),
          if (_refError != null)
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: 16),
              child: KCard(color: theme.colorScheme.errorContainer, child: Text('Server: $_refError', style: TextStyle(color: theme.colorScheme.onErrorContainer))),
            ),
          if (_refLink != null)
            InfoSection(title: 'Server (referral-service)', rows: [
              ('Trạng thái link', _refStatus?.status ?? _refLink!.status),
              ('Lời mời', _refStatus?.invitationStatus),
              ('Người mời', _refLink!.ownerUserId == me ? '${_refLink!.ownerUserId} (chính bạn, sẽ bị từ chối)' : _refLink!.ownerUserId),
              ('Kênh', _refLink!.channel),
              ('Hết hạn', _refLink!.expiresAt == null ? null : formatDate(DateTime.fromMillisecondsSinceEpoch(_refLink!.expiresAt! < 100000000000 ? _refLink!.expiresAt! * 1000 : _refLink!.expiresAt!).toIso8601String())),
              ('Thưởng', _refStatus?.rewardStatus),
            ]),
        ],
        SectionHeader('Ghi nhận nguồn cài đặt', action: state.pending != null ? 'Gửi lại' : null, onAction: state.pending != null ? context.read<AttributionCubit>().retry : null),
        InfoSection(rows: [
          ('Đang chờ gửi', state.pending == null ? 'Không' : '${state.pending!.refCode} • ${state.pending!.source.name}${state.activating ? ' • đang gửi' : ''}'),
          ('Lần gửi gần nhất', state.lastResult == null ? null : '${state.lastResult!.refCode} • ${formatDate(state.lastResult!.at.toIso8601String())}'),
          (
            'Kết quả',
            state.lastResult == null
                ? null
                : state.lastResult!.ok
                    ? (state.lastResult!.attributed ? 'Đã ghi nhận, người mời ${state.lastResult!.ownerUserId ?? '-'}${state.lastResult!.rewardType == null ? '' : ' • thưởng ${state.lastResult!.rewardType}'}' : 'Server không ghi nhận (đã có nguồn trước đó hoặc không khớp)')
                    : 'Lỗi: ${state.lastResult!.error}'
          ),
        ]),
        if (state.enabled)
          ActionButtons([
            OutlinedButton.icon(onPressed: () => _params(true), icon: const Icon(Icons.install_mobile), label: const Text('Tham số lần cài đầu')),
            OutlinedButton.icon(onPressed: () => _params(false), icon: const Icon(Icons.history), label: const Text('Tham số gần nhất')),
            OutlinedButton.icon(
              onPressed: () {
                context.read<AttributionCubit>().source.validate();
                showSnack(context, 'Đã chạy kiểm tra tích hợp, xem kết quả trong Logcat/Console');
              },
              icon: const Icon(Icons.verified_outlined),
              label: const Text('Kiểm tra tích hợp SDK'),
            ),
          ]),
        ActionButtons([
          OutlinedButton.icon(
            onPressed: () async {
              final a = await context.api.referral.myReferralAttribution().catchError((_) => null);
              if (context.mounted) await _showJson('Nguồn cài đặt trên server', a == null ? null : adminPlainJson(a.toJson()));
            },
            icon: const Icon(Icons.cloud_done_outlined),
            label: const Text('Nguồn cài đặt trên server'),
          ),
        ]),
        SectionHeader('Danh mục deep link (${deepLinkCatalog.length})'),
        Padding(
          padding: const EdgeInsets.fromLTRB(20, 0, 20, 8),
          child: Text('Mỗi đường dẫn mở được bằng keikibook://app<đường dẫn>, https://${Env.webHost}<đường dẫn> hoặc link Branch có \$deeplink_path.', style: theme.textTheme.bodySmall),
        ),
        for (final group in {for (final d in deepLinkCatalog) d.group})
          ExpansionTile(
            title: Text(group),
            subtitle: Text('${deepLinkCatalog.where((d) => d.group == group).length} link'),
            children: [
              for (final d in deepLinkCatalog.where((d) => d.group == group))
                ListTile(
                  dense: true,
                  title: Text(d.title),
                  subtitle: Text(d.pattern, style: const TextStyle(fontFamily: 'monospace')),
                  onTap: () => context.push(d.example),
                  trailing: PopupMenuButton<String>(
                    tooltip: 'Sao chép',
                    icon: const Icon(Icons.more_vert),
                    onSelected: (v) {
                      if (v == 'check') {
                        _link.text = d.appUrl;
                        _inspect();
                        return;
                      }
                      Clipboard.setData(ClipboardData(text: v == 'web' ? d.webUrl : d.appUrl));
                      showSnack(context, 'Đã sao chép');
                    },
                    itemBuilder: (_) => [
                      PopupMenuItem(value: 'web', child: Text('Sao chép ${d.webUrl}', maxLines: 1, overflow: TextOverflow.ellipsis)),
                      PopupMenuItem(value: 'app', child: Text('Sao chép ${d.appUrl}', maxLines: 1, overflow: TextOverflow.ellipsis)),
                      const PopupMenuItem(value: 'check', child: Text('Kiểm tra link mẫu')),
                    ],
                  ),
                ),
            ],
          ),
        SectionHeader('Lịch sử link (${state.history.length})'),
        if (state.history.isEmpty)
          const Padding(padding: EdgeInsets.symmetric(horizontal: 20), child: Text('Chưa nhận link nào trên thiết bị này'))
        else
          for (final h in state.history)
            ListTile(
              leading: Icon(switch (h.source) {
                LinkTouchSource.branch => Icons.alt_route,
                LinkTouchSource.deepLink => Icons.link,
                LinkTouchSource.scan => Icons.qr_code_scanner,
                LinkTouchSource.manual => Icons.keyboard_outlined,
              }),
              title: Text(h.refCode != null ? 'Mã mời ${h.refCode}' : (h.route ?? h.url ?? h.source.name), maxLines: 1, overflow: TextOverflow.ellipsis),
              subtitle: Text(
                [formatDate(h.receivedAt.toIso8601String()), h.source.name, if (h.utm['utm_source'] != null) h.utm['utm_source']!, if (h.firstSession) 'lần mở đầu'].join(' • '),
                maxLines: 1,
                overflow: TextOverflow.ellipsis,
              ),
              trailing: const Icon(Icons.chevron_right),
              onTap: () => _showJson('Chi tiết link', h.toJson()),
            ),
        const SizedBox(height: 24),
      ],
    );
  }
}

Object? adminPlainJson(Object? v) => jsonDecode(jsonEncode(v));
