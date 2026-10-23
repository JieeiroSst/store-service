import '../scan/scan.dart';

enum LinkTouchSource { branch, deepLink, scan, manual }

const utmKeys = ['utm_source', 'utm_medium', 'utm_campaign', 'utm_term', 'utm_content'];

class LinkTouch {
  const LinkTouch({
    required this.source,
    required this.receivedAt,
    this.url,
    this.refCode,
    this.clickId,
    this.channel,
    this.campaign,
    this.feature,
    this.route,
    this.utm = const {},
    this.clickedBranchLink = false,
    this.firstSession = false,
    this.raw = const {},
  });

  final LinkTouchSource source;
  final DateTime receivedAt;
  final String? url;
  final String? refCode;
  final String? clickId;
  final String? channel;
  final String? campaign;
  final String? feature;
  final String? route;
  final Map<String, String> utm;
  final bool clickedBranchLink;
  final bool firstSession;
  final Map<String, Object?> raw;

  bool get isReferral => refCode != null && refCode!.isNotEmpty;

  static String? _str(Object? v) {
    final s = v?.toString().trim();
    return s == null || s.isEmpty || s == 'null' ? null : s;
  }

  static (String?, String?) _splitRef(String? ref) {
    if (ref == null) return (null, null);
    final i = ref.indexOf(':');
    if (i <= 0) return (ref.toUpperCase(), null);
    return (ref.substring(0, i).toUpperCase(), _str(ref.substring(i + 1)));
  }

  static String? _routeFor(String? value) {
    final v = _str(value);
    if (v == null) return null;
    if (v.startsWith('/')) return v;
    final mapped = routeForCode(v);
    if (mapped != null) return mapped;
    if (!v.contains('://')) return '/$v';
    return null;
  }

  static LinkTouch? fromBranch(Map<dynamic, dynamic> params, {DateTime? now}) {
    final p = {for (final e in params.entries) '${e.key}': e.value};
    final clicked = p['+clicked_branch_link'] == true || p['+clicked_branch_link'] == 'true';
    final at = now ?? DateTime.now();
    if (!clicked) {
      final nonBranch = _str(p['+non_branch_link']);
      if (nonBranch == null) return null;
      final uri = Uri.tryParse(nonBranch);
      return uri == null ? null : fromUri(uri, source: LinkTouchSource.branch, now: at);
    }
    final (ref, refClick) = _splitRef(_str(p['ref_code']) ?? _str(p['ref']) ?? _str(p['referral_code']));
    final utm = <String, String>{
      for (final k in utmKeys)
        if (_str(p[k]) != null) k: _str(p[k])!,
    };
    utm.putIfAbsent('utm_source', () => _str(p['~channel']) ?? '');
    utm.putIfAbsent('utm_campaign', () => _str(p['~campaign']) ?? '');
    utm.removeWhere((_, v) => v.isEmpty);
    return LinkTouch(
      source: LinkTouchSource.branch,
      receivedAt: at,
      url: _str(p['~referring_link']) ?? _str(p['+url']),
      refCode: ref,
      clickId: _str(p['click_id']) ?? refClick,
      channel: _str(p['channel']) ?? _str(p['~channel']),
      campaign: _str(p['~campaign']) ?? utm['utm_campaign'],
      feature: _str(p['~feature']),
      route: _routeFor(_str(p[r'$deeplink_path']) ?? _str(p['route']) ?? _str(p[r'$canonical_url'])),
      utm: utm,
      clickedBranchLink: true,
      firstSession: p['+is_first_session'] == true || p['+is_first_session'] == 'true',
      raw: p,
    );
  }

  static LinkTouch? fromUri(Uri uri, {LinkTouchSource source = LinkTouchSource.deepLink, DateTime? now}) {
    final q = uri.queryParameters;
    var (ref, refClick) = _splitRef(_str(q['ref_code']) ?? _str(q['ref']) ?? _str(q['referral_code']));
    final segments = uri.pathSegments;
    if (ref == null && segments.length == 2 && segments.first == 'r') (ref, refClick) = _splitRef(segments.last);
    final utm = <String, String>{
      for (final k in utmKeys)
        if (_str(q[k]) != null) k: _str(q[k])!,
    };
    final route = uri.hasScheme ? routeForCode(uri.toString()) : _routeFor(uri.toString());
    if (ref == null && utm.isEmpty && route == null) return null;
    return LinkTouch(
      source: source,
      receivedAt: now ?? DateTime.now(),
      url: uri.toString(),
      refCode: ref,
      clickId: _str(q['click_id']) ?? refClick,
      channel: _str(q['channel']) ?? utm['utm_source'],
      campaign: utm['utm_campaign'],
      route: ref != null && (segments.isEmpty || segments.first == 'open' || segments.first == 'r' || uri.host == 'open') ? null : route,
      utm: utm,
      raw: {for (final e in q.entries) e.key: e.value},
    );
  }

  Map<String, Object?> toJson() => {
        'source': source.name,
        'received_at': receivedAt.toIso8601String(),
        'url': ?url,
        'ref_code': ?refCode,
        'click_id': ?clickId,
        'channel': ?channel,
        'campaign': ?campaign,
        'feature': ?feature,
        'route': ?route,
        'utm': utm,
        'clicked_branch_link': clickedBranchLink,
        'first_session': firstSession,
        'raw': {for (final e in raw.entries) e.key: e.value is num || e.value is bool || e.value == null ? e.value : '${e.value}'},
      };

  factory LinkTouch.fromJson(Map<String, dynamic> j) => LinkTouch(
        source: LinkTouchSource.values.asNameMap()[j['source']] ?? LinkTouchSource.deepLink,
        receivedAt: DateTime.tryParse('${j['received_at']}') ?? DateTime.now(),
        url: j['url'] as String?,
        refCode: j['ref_code'] as String?,
        clickId: j['click_id'] as String?,
        channel: j['channel'] as String?,
        campaign: j['campaign'] as String?,
        feature: j['feature'] as String?,
        route: j['route'] as String?,
        utm: ((j['utm'] as Map?) ?? const {}).map((k, v) => MapEntry('$k', '$v')),
        clickedBranchLink: j['clicked_branch_link'] == true,
        firstSession: j['first_session'] == true,
        raw: ((j['raw'] as Map?) ?? const {}).map((k, v) => MapEntry('$k', v)),
      );
}
