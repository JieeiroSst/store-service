import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_branch_sdk/flutter_branch_sdk.dart';
import 'package:share_plus/share_plus.dart';

import '../config/env.dart';
import '../tracking/link_source.dart';
import '../widgets/status_view.dart';

typedef ShortLinkBuilder = Future<String?> Function(String path, String webUrl, {String? title, String? description, String? imageUrl, String channel});

String webLinkFor(String path) {
  final u = Uri.parse(path.startsWith('/') ? path : '/$path');
  return Uri(scheme: 'https', host: Env.webHost, path: u.path, queryParameters: u.queryParameters.isEmpty ? null : u.queryParameters).toString();
}

String appLinkFor(String path) {
  final u = Uri.parse(path.startsWith('/') ? path : '/$path');
  return Uri(scheme: 'keikibook', host: 'app', path: u.path, queryParameters: u.queryParameters.isEmpty ? null : u.queryParameters).toString();
}

Future<String?> _branchShortLink(String path, String webUrl, {String? title, String? description, String? imageUrl, String channel = 'app'}) async {
  if (!BranchLinkSource.supported || Env.branchKey.isEmpty) return null;
  final buo = BranchUniversalObject(
    canonicalIdentifier: 'keikibook$path',
    canonicalUrl: webUrl,
    title: title ?? 'Keikibook',
    contentDescription: description ?? '',
    imageUrl: imageUrl ?? '',
  );
  final props = BranchLinkProperties(channel: channel, feature: 'share')
    ..addControlParam(r'$deeplink_path', path)
    ..addControlParam(r'$fallback_url', webUrl)
    ..addControlParam(r'$desktop_url', webUrl);
  final r = await FlutterBranchSdk.getShortUrl(buo: buo, linkProperties: props);
  return r.success ? r.result as String? : null;
}

ShortLinkBuilder shortLinkBuilder = _branchShortLink;

Future<String> buildShareLink(String path, {String? title, String? description, String? imageUrl, String channel = 'app'}) async {
  final web = webLinkFor(path);
  try {
    return await shortLinkBuilder(path, web, title: title, description: description, imageUrl: imageUrl, channel: channel) ?? web;
  } catch (_) {
    return web;
  }
}

Future<void> shareDeepLink(BuildContext context, {required String path, required String title, String? description, String? imageUrl}) async {
  final link = await buildShareLink(path, title: title, description: description, imageUrl: imageUrl);
  if (!context.mounted) return;
  await showLinkSheet(context, link: link, path: path, title: title);
}

Future<void> showLinkSheet(BuildContext context, {required String link, required String path, required String title}) {
  return showModalBottomSheet<void>(
    context: context,
    showDragHandle: true,
    builder: (context) => SafeArea(
      child: Padding(
        padding: const EdgeInsets.fromLTRB(20, 0, 20, 20),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Text('Chia sẻ "$title"', style: Theme.of(context).textTheme.titleMedium, maxLines: 2, overflow: TextOverflow.ellipsis),
            const SizedBox(height: 12),
            Container(
              padding: const EdgeInsets.all(12),
              decoration: BoxDecoration(color: Theme.of(context).colorScheme.surfaceContainerHighest, borderRadius: BorderRadius.circular(12)),
              child: SelectableText(link),
            ),
            const SizedBox(height: 6),
            Text('Mở trong app: ${appLinkFor(path)}', style: Theme.of(context).textTheme.bodySmall),
            const SizedBox(height: 16),
            Row(
              children: [
                Expanded(
                  child: OutlinedButton.icon(
                    icon: const Icon(Icons.copy),
                    label: const Text('Sao chép'),
                    onPressed: () {
                      Clipboard.setData(ClipboardData(text: link));
                      Navigator.pop(context);
                      showSnack(context, 'Đã sao chép link');
                    },
                  ),
                ),
                const SizedBox(width: 12),
                Expanded(
                  child: FilledButton.icon(
                    icon: const Icon(Icons.ios_share),
                    label: const Text('Gửi'),
                    onPressed: () async {
                      Navigator.pop(context);
                      await SharePlus.instance.share(ShareParams(text: '$title\n$link', subject: title));
                    },
                  ),
                ),
              ],
            ),
          ],
        ),
      ),
    ),
  );
}

class ShareLinkButton extends StatefulWidget {
  const ShareLinkButton({super.key, required this.path, required this.title, this.description, this.imageUrl, this.color});

  final String path;
  final String title;
  final String? description;
  final String? imageUrl;
  final Color? color;

  @override
  State<ShareLinkButton> createState() => _ShareLinkButtonState();
}

class _ShareLinkButtonState extends State<ShareLinkButton> {
  bool _busy = false;

  @override
  Widget build(BuildContext context) => IconButton(
        tooltip: 'Chia sẻ',
        color: widget.color,
        icon: _busy ? const SizedBox(width: 20, height: 20, child: CircularProgressIndicator(strokeWidth: 2)) : const Icon(Icons.share_outlined),
        onPressed: _busy
            ? null
            : () async {
                setState(() => _busy = true);
                final link = await buildShareLink(widget.path, title: widget.title, description: widget.description, imageUrl: widget.imageUrl);
                if (!context.mounted) return;
                setState(() => _busy = false);
                await showLinkSheet(context, link: link, path: widget.path, title: widget.title);
              },
      );
}
