import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:intl/intl.dart';

import '../../../core/widgets/status_view.dart';
import '../../../core/widgets/ui.dart';
import '../../../gen/models.g.dart' show MfaStatus, TotpSetup;
import '../../auth/data/auth_repository.dart';

class TwoFactorPage extends StatefulWidget {
  const TwoFactorPage({super.key});

  @override
  State<TwoFactorPage> createState() => _TwoFactorPageState();
}

class _TwoFactorPageState extends State<TwoFactorPage> {
  late final AuthRepository _repo = context.read<AuthRepository>();
  final _code = TextEditingController();
  MfaStatus? _status;
  TotpSetup? _setup;
  Object? _error;
  bool _busy = false;

  @override
  void initState() {
    super.initState();
    _load();
  }

  @override
  void dispose() {
    _code.dispose();
    super.dispose();
  }

  Future<void> _load() async {
    setState(() => _error = null);
    try {
      final status = await _repo.mfaStatus();
      if (mounted) setState(() => _status = status);
    } catch (e) {
      if (mounted) setState(() => _error = e);
    }
  }

  Future<T?> _run<T>(Future<T> Function() action) async {
    setState(() => _busy = true);
    try {
      return await action();
    } catch (e) {
      if (mounted) showSnack(context, e.toString());
      return null;
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  Future<void> _startSetup() async {
    final setup = await _run(_repo.setupTotp);
    if (setup != null && mounted) setState(() => _setup = setup);
  }

  Future<void> _enable() async {
    final code = _code.text.trim();
    if (code.length != 6) {
      showSnack(context, 'Nhập đủ 6 số từ ứng dụng Authenticator');
      return;
    }
    final codes = await _run(() => _repo.enableTotp(code));
    if (codes == null || !mounted) return;
    _code.clear();
    setState(() => _setup = null);
    await _showRecoveryCodes(codes);
    await _load();
  }

  Future<void> _regenerate() async {
    final code = await _askCode(title: 'Tạo lại mã khôi phục', allowRecovery: false);
    if (code == null) return;
    final codes = await _run(() => _repo.regenerateRecoveryCodes(code));
    if (codes == null || !mounted) return;
    await _showRecoveryCodes(codes);
    await _load();
  }

  Future<void> _disable() async {
    final code = await _askCode(title: 'Tắt xác thực 2 lớp', allowRecovery: true);
    if (code == null) return;
    final ok = await _run(() async {
      await _repo.disableTotp(code);
      return true;
    });
    if (ok != true || !mounted) return;
    showSnack(context, 'Đã tắt xác thực 2 lớp');
    await _load();
  }

  Future<String?> _askCode({required String title, required bool allowRecovery}) {
    final controller = TextEditingController();
    return showDialog<String>(
      context: context,
      builder: (context) => AlertDialog(
        icon: const Icon(Icons.lock_clock_outlined),
        title: Text(title),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Text(allowRecovery ? 'Nhập mã 6 số từ Google Authenticator hoặc một mã khôi phục.' : 'Nhập mã 6 số từ Google Authenticator.'),
            const SizedBox(height: 12),
            TextField(
              controller: controller,
              autofocus: true,
              textAlign: TextAlign.center,
              keyboardType: allowRecovery ? TextInputType.text : TextInputType.number,
              maxLength: allowRecovery ? 11 : 6,
              inputFormatters: allowRecovery ? null : [FilteringTextInputFormatter.digitsOnly],
              decoration: InputDecoration(hintText: allowRecovery ? '000000 / xxxxx-xxxxx' : '000000', counterText: ''),
            ),
          ],
        ),
        actions: [
          TextButton(onPressed: () => Navigator.pop(context), child: const Text('Huỷ')),
          FilledButton(onPressed: () => Navigator.pop(context, controller.text.trim()), child: const Text('Xác nhận')),
        ],
      ),
    );
  }

  Future<void> _showRecoveryCodes(List<String> codes) {
    return showDialog<void>(
      context: context,
      barrierDismissible: false,
      builder: (context) => AlertDialog(
        icon: const Icon(Icons.key_outlined),
        title: const Text('Lưu mã khôi phục'),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            const Text('Dùng một mã khi mất điện thoại. Mỗi mã chỉ dùng được một lần và sẽ không hiển thị lại.'),
            const SizedBox(height: 16),
            Container(
              padding: const EdgeInsets.all(12),
              decoration: BoxDecoration(
                color: Theme.of(context).colorScheme.surfaceContainerHighest,
                borderRadius: BorderRadius.circular(12),
              ),
              child: Wrap(
                spacing: 16,
                runSpacing: 8,
                children: [
                  for (final c in codes)
                    SizedBox(
                      width: 110,
                      child: Text(c, style: const TextStyle(fontWeight: FontWeight.w600, letterSpacing: 1, fontFeatures: [FontFeature.tabularFigures()])),
                    ),
                ],
              ),
            ),
          ],
        ),
        actions: [
          TextButton.icon(
            onPressed: () {
              Clipboard.setData(ClipboardData(text: codes.join('\n')));
              showSnack(context, 'Đã sao chép mã khôi phục');
            },
            icon: const Icon(Icons.copy),
            label: const Text('Sao chép'),
          ),
          FilledButton(onPressed: () => Navigator.pop(context), child: const Text('Tôi đã lưu')),
        ],
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final status = _status;
    final Widget body;
    if (_error != null) {
      body = ErrorView(message: _error.toString(), onRetry: _load);
    } else if (status == null) {
      body = const Center(child: CircularProgressIndicator());
    } else if (status.enabled == true) {
      body = _EnabledView(status: status, busy: _busy, onRegenerate: _regenerate, onDisable: _disable);
    } else if (_setup != null) {
      body = _SetupView(setup: _setup!, code: _code, busy: _busy, onEnable: _enable, onCancel: () => setState(() => _setup = null));
    } else {
      body = _IntroView(busy: _busy, onStart: _startSetup);
    }
    return Scaffold(
      appBar: AppBar(title: const Text('Xác thực 2 lớp')),
      body: body,
    );
  }
}

class _IntroView extends StatelessWidget {
  const _IntroView({required this.busy, required this.onStart});

  final bool busy;
  final VoidCallback onStart;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return ListView(
      padding: const EdgeInsets.all(20),
      children: [
        const Center(child: IconBadge(icon: Icons.phonelink_lock_outlined, color: Colors.deepPurple, size: 72)),
        const SizedBox(height: 20),
        Text('Bảo vệ tài khoản bằng Google Authenticator', textAlign: TextAlign.center, style: theme.textTheme.titleLarge),
        const SizedBox(height: 8),
        Text(
          'Mỗi lần đăng nhập, ngoài mật khẩu bạn cần nhập mã 6 số thay đổi mỗi 30 giây trên điện thoại. Kẻ gian có mật khẩu cũng không vào được.',
          textAlign: TextAlign.center,
          style: theme.textTheme.bodyMedium?.copyWith(color: theme.colorScheme.onSurfaceVariant),
        ),
        const SizedBox(height: 28),
        LoadingButton(label: 'Bật xác thực 2 lớp', loading: busy, onPressed: onStart),
      ],
    );
  }
}

class _SetupView extends StatelessWidget {
  const _SetupView({required this.setup, required this.code, required this.busy, required this.onEnable, required this.onCancel});

  final TotpSetup setup;
  final TextEditingController code;
  final bool busy;
  final VoidCallback onEnable;
  final VoidCallback onCancel;

  Uint8List? get _qr {
    final uri = setup.qrCode ?? '';
    final i = uri.indexOf('base64,');
    if (i < 0) return null;
    try {
      return base64Decode(uri.substring(i + 7));
    } on FormatException {
      return null;
    }
  }

  String get _groupedSecret {
    final s = setup.secret ?? '';
    return [for (var i = 0; i < s.length; i += 4) s.substring(i, i + 4 > s.length ? s.length : i + 4)].join(' ');
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final qr = _qr;
    Widget step(int n, String title, Widget child) => Padding(
          padding: const EdgeInsets.only(bottom: 20),
          child: Row(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              CircleAvatar(radius: 14, child: Text('$n', style: const TextStyle(fontSize: 13, fontWeight: FontWeight.w700))),
              const SizedBox(width: 12),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [Text(title, style: theme.textTheme.titleSmall), const SizedBox(height: 8), child],
                ),
              ),
            ],
          ),
        );
    return ListView(
      padding: const EdgeInsets.all(20),
      children: [
        step(
          1,
          'Cài Google Authenticator',
          Text('Tải từ App Store hoặc Google Play, rồi chọn "Thêm mã" (+).', style: theme.textTheme.bodyMedium),
        ),
        step(
          2,
          'Quét mã QR',
          Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              if (qr != null)
                Container(
                  padding: const EdgeInsets.all(8),
                  decoration: BoxDecoration(color: Colors.white, borderRadius: BorderRadius.circular(12)),
                  child: Image.memory(qr, width: 200, height: 200, filterQuality: FilterQuality.none),
                ),
              const SizedBox(height: 12),
              Text('Hoặc nhập khoá thủ công (${setup.account ?? ''}):', style: theme.textTheme.bodySmall),
              const SizedBox(height: 4),
              Row(
                children: [
                  Expanded(
                    child: SelectableText(_groupedSecret, style: const TextStyle(fontWeight: FontWeight.w600, letterSpacing: 1.5, fontFeatures: [FontFeature.tabularFigures()])),
                  ),
                  IconButton(
                    tooltip: 'Sao chép khoá',
                    icon: const Icon(Icons.copy),
                    onPressed: () {
                      Clipboard.setData(ClipboardData(text: setup.secret ?? ''));
                      showSnack(context, 'Đã sao chép khoá');
                    },
                  ),
                ],
              ),
            ],
          ),
        ),
        step(
          3,
          'Nhập mã 6 số hiển thị trong ứng dụng',
          TextField(
            controller: code,
            keyboardType: TextInputType.number,
            textAlign: TextAlign.center,
            maxLength: 6,
            autofillHints: const [AutofillHints.oneTimeCode],
            inputFormatters: [FilteringTextInputFormatter.digitsOnly],
            style: theme.textTheme.headlineSmall?.copyWith(letterSpacing: 10),
            onSubmitted: (_) => onEnable(),
            decoration: const InputDecoration(hintText: '000000', counterText: ''),
          ),
        ),
        LoadingButton(label: 'Xác nhận và bật', loading: busy, onPressed: onEnable),
        const SizedBox(height: 8),
        TextButton(onPressed: busy ? null : onCancel, child: const Text('Huỷ')),
      ],
    );
  }
}

class _EnabledView extends StatelessWidget {
  const _EnabledView({required this.status, required this.busy, required this.onRegenerate, required this.onDisable});

  final MfaStatus status;
  final bool busy;
  final VoidCallback onRegenerate;
  final VoidCallback onDisable;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final since = DateTime.tryParse(status.enabledAt ?? '');
    final remaining = status.recoveryCodesRemaining ?? 0;
    return ListView(
      padding: const EdgeInsets.all(16),
      children: [
        Card(
          child: ListTile(
            leading: const IconBadge(icon: Icons.verified_user_outlined, color: Colors.green, size: 44),
            title: const Text('Đang bật'),
            subtitle: Text(since == null ? 'Google Authenticator' : 'Google Authenticator · từ ${DateFormat('dd/MM/yyyy').format(since.toLocal())}'),
          ),
        ),
        const SizedBox(height: 12),
        Card(
          child: Column(
            children: [
              ListTile(
                leading: const Icon(Icons.key_outlined),
                title: const Text('Mã khôi phục'),
                subtitle: Text(remaining <= 2 ? 'Còn $remaining mã, nên tạo lại' : 'Còn $remaining mã'),
                trailing: const Icon(Icons.chevron_right),
                enabled: !busy,
                onTap: onRegenerate,
              ),
              ListTile(
                leading: Icon(Icons.lock_open_outlined, color: theme.colorScheme.error),
                title: Text('Tắt xác thực 2 lớp', style: TextStyle(color: theme.colorScheme.error)),
                enabled: !busy,
                onTap: onDisable,
              ),
            ],
          ),
        ),
      ],
    );
  }
}
