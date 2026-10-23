import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:go_router/go_router.dart';
import 'package:image_picker/image_picker.dart';
import 'package:mobile_scanner/mobile_scanner.dart';

import '../tracking/attribution_cubit.dart';
import '../tracking/link_touch.dart';
import '../widgets/status_view.dart';

final _compact = FilledButton.styleFrom(minimumSize: const Size(0, 48));

const keikiHosts = {'keikibook.vn', 'www.keikibook.vn', 'app.keikibook.vn'};

bool get cameraSupported => !kIsWeb && (defaultTargetPlatform == TargetPlatform.android || defaultTargetPlatform == TargetPlatform.iOS);

Future<String?> scanCode(BuildContext context, {String title = 'Quét mã'}) =>
    Navigator.of(context, rootNavigator: true).push<String>(MaterialPageRoute(fullscreenDialog: true, builder: (_) => ScanPage(title: title)));

String? routeForCode(String code) {
  final uri = Uri.tryParse(code.trim());
  if (uri == null || !uri.hasScheme) return null;
  String withQuery(String path) => Uri(path: path.isEmpty ? '/' : path, queryParameters: uri.queryParameters.isEmpty ? null : uri.queryParameters).toString();
  if (uri.scheme == 'keiki' && uri.host == 'pay') return withQuery('/wallet/transfer');
  if (uri.scheme == 'keikibook' && (uri.host == 'open' || uri.path == '/open')) return '/';
  if (uri.scheme == 'keikibook') return withQuery(uri.host.isEmpty || uri.host == 'app' ? uri.path : '/${uri.host}${uri.path}');
  if ((uri.scheme == 'https' || uri.scheme == 'http') && keikiHosts.contains(uri.host) && uri.path.length > 1) return withQuery(uri.path);
  return null;
}

String codeValue(String scanned) {
  final uri = Uri.tryParse(scanned.trim());
  if (uri == null || !uri.hasScheme) return scanned.trim();
  final q = uri.queryParameters;
  for (final k in const ['wallet_id', 'code', 'id', 'ref', 'token']) {
    final v = q[k];
    if (v != null && v.isNotEmpty) return v;
  }
  if (uri.pathSegments.isNotEmpty && (uri.scheme == 'keikibook' || keikiHosts.contains(uri.host))) return uri.pathSegments.last;
  return scanned.trim();
}

Future<void> scanAndOpen(BuildContext context) async {
  final code = await scanCode(context);
  if (code == null || !context.mounted) return;
  await openScanned(context, code);
}

Future<void> openScanned(BuildContext context, String code) async {
  final uri = Uri.tryParse(code.trim());
  if (uri != null && uri.hasScheme) {
    try {
      context.read<AttributionCubit>().observeUri(uri, source: LinkTouchSource.scan);
    } catch (_) {}
  }
  if (!context.mounted) return;
  final route = routeForCode(code);
  if (route != null) {
    context.push(route);
    return;
  }
  await showDialog<void>(
    context: context,
    builder: (context) => AlertDialog(
      icon: const Icon(Icons.qr_code_2, size: 40),
      title: const Text('Nội dung mã'),
      content: SelectableText(code),
      actions: [
        TextButton(
          onPressed: () {
            Clipboard.setData(ClipboardData(text: code));
            Navigator.pop(context);
            showSnack(context, 'Đã sao chép');
          },
          child: const Text('Sao chép'),
        ),
        FilledButton(onPressed: () => Navigator.pop(context), child: const Text('Đóng')),
      ],
    ),
  );
}

class ScanPage extends StatefulWidget {
  const ScanPage({super.key, this.title = 'Quét mã', this.onCode});

  final String title;
  final void Function(BuildContext context, String code)? onCode;

  @override
  State<ScanPage> createState() => _ScanPageState();
}

class _ScanPageState extends State<ScanPage> {
  MobileScannerController? _controller = cameraSupported ? MobileScannerController(detectionSpeed: DetectionSpeed.noDuplicates) : null;
  bool _done = false;

  @override
  void dispose() {
    _controller?.dispose();
    _controller = null;
    super.dispose();
  }

  void _finish(String? code) {
    final v = code?.trim();
    if (_done || v == null || v.isEmpty) return;
    _done = true;
    HapticFeedback.mediumImpact();
    final onCode = widget.onCode;
    if (onCode == null) {
      Navigator.pop(context, v);
      return;
    }
    final nav = Navigator.of(context);
    final host = nav.context;
    if (nav.canPop()) {
      nav.pop();
    } else {
      GoRouter.of(context).go('/');
    }
    onCode(host, v);
  }

  Future<void> _fromGallery() async {
    final c = _controller;
    if (c == null) return;
    final image = await ImagePicker().pickImage(source: ImageSource.gallery);
    if (image == null) return;
    final result = await c.analyzeImage(image.path);
    final code = result?.barcodes.map((b) => b.rawValue).whereType<String>().firstOrNull;
    if (code != null) {
      _finish(code);
    } else if (mounted) {
      showSnack(context, 'Không tìm thấy mã trong ảnh');
    }
  }

  Future<void> _manual() async {
    final v = await showDialog<String>(context: context, builder: (_) => const _ManualEntry());
    _finish(v);
  }

  @override
  Widget build(BuildContext context) {
    final c = _controller;
    return Scaffold(
      backgroundColor: Colors.black,
      appBar: AppBar(
        backgroundColor: Colors.black,
        foregroundColor: Colors.white,
        title: Text(widget.title),
        actions: [
          if (c != null)
            ValueListenableBuilder<MobileScannerState>(
              valueListenable: c,
              builder: (context, s, _) => IconButton(
                tooltip: 'Đèn pin',
                icon: Icon(s.torchState == TorchState.on ? Icons.flash_on : Icons.flash_off),
                onPressed: s.torchState == TorchState.unavailable ? null : c.toggleTorch,
              ),
            ),
        ],
      ),
      body: Stack(
        fit: StackFit.expand,
        children: [
          if (c != null)
            MobileScanner(
              controller: c,
              onDetect: (capture) => _finish(capture.barcodes.map((b) => b.rawValue).whereType<String>().firstOrNull),
              errorBuilder: (context, error) => _Message(
                icon: Icons.no_photography_outlined,
                text: error.errorCode == MobileScannerErrorCode.permissionDenied ? 'Hãy cho phép Keikibook dùng camera trong Cài đặt để quét mã' : 'Không mở được camera',
              ),
            )
          else
            const _Message(icon: Icons.qr_code_scanner, text: 'Thiết bị này không quét bằng camera được, hãy nhập mã'),
          if (c != null) const IgnorePointer(child: _Frame()),
          Positioned(
            left: 16,
            right: 16,
            bottom: 24 + MediaQuery.paddingOf(context).bottom,
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                const Text('Đưa mã QR vào giữa khung', style: TextStyle(color: Colors.white70)),
                const SizedBox(height: 16),
                Row(
                  mainAxisAlignment: MainAxisAlignment.center,
                  children: [
                    if (c != null) ...[
                      FilledButton.tonalIcon(style: _compact, onPressed: _fromGallery, icon: const Icon(Icons.photo_library_outlined), label: const Text('Ảnh có sẵn')),
                      const SizedBox(width: 12),
                    ],
                    FilledButton.tonalIcon(style: _compact, onPressed: _manual, icon: const Icon(Icons.keyboard_outlined), label: const Text('Nhập mã')),
                  ],
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}

class _ManualEntry extends StatefulWidget {
  const _ManualEntry();

  @override
  State<_ManualEntry> createState() => _ManualEntryState();
}

class _ManualEntryState extends State<_ManualEntry> {
  final _c = TextEditingController();

  @override
  void dispose() {
    _c.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) => AlertDialog(
        title: const Text('Nhập mã'),
        content: TextField(controller: _c, autofocus: true, decoration: const InputDecoration(hintText: 'Mã hoặc đường dẫn'), onSubmitted: (v) => Navigator.pop(context, v)),
        actions: [
          TextButton(onPressed: () => Navigator.pop(context), child: const Text('Huỷ')),
          FilledButton(onPressed: () => Navigator.pop(context, _c.text), child: const Text('Xong')),
        ],
      );
}

class _Message extends StatelessWidget {
  const _Message({required this.icon, required this.text});

  final IconData icon;
  final String text;

  @override
  Widget build(BuildContext context) => Center(
        child: Padding(
          padding: const EdgeInsets.all(32),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Icon(icon, color: Colors.white54, size: 64),
              const SizedBox(height: 16),
              Text(text, textAlign: TextAlign.center, style: const TextStyle(color: Colors.white)),
            ],
          ),
        ),
      );
}

class _Frame extends StatelessWidget {
  const _Frame();

  @override
  Widget build(BuildContext context) => Center(
        child: Container(
          width: 260,
          height: 260,
          decoration: BoxDecoration(
            border: Border.all(color: Colors.white, width: 3),
            borderRadius: BorderRadius.circular(28),
            boxShadow: const [BoxShadow(color: Color(0x88000000), spreadRadius: 2000)],
          ),
        ),
      );
}
