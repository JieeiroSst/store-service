import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:go_router/go_router.dart';

import '../../../app/theme.dart';
import '../../../core/push/push_service.dart';
import '../../../core/widgets/status_view.dart';
import '../../../core/widgets/ui.dart';
import '../../auth/bloc/auth_bloc.dart';
import '../../auth/data/auth_repository.dart';
import '../../auth/data/user.dart';
import '../../apps/commerce/commerce_more.dart';
import '../../../app/role.dart';

class ProfilePage extends StatelessWidget {
  const ProfilePage({super.key});

  Future<void> _logout(BuildContext context) async {
    final ok = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        icon: const Icon(Icons.logout),
        title: const Text('Đăng xuất?'),
        content: const Text('Bạn sẽ cần đăng nhập lại để dùng các dịch vụ.'),
        actions: [
          TextButton(onPressed: () => Navigator.pop(context, false), child: const Text('Huỷ')),
          FilledButton(onPressed: () => Navigator.pop(context, true), child: const Text('Đăng xuất')),
        ],
      ),
    );
    if (ok != true || !context.mounted) return;
    final auth = context.read<AuthBloc>();
    await context.read<PushService>().onLogout();
    auth.add(const AuthLogoutRequested());
  }

  @override
  Widget build(BuildContext context) {
    final user = context.select<AuthBloc, User?>((b) => b.state.user);
    final theme = Theme.of(context);
    final initial = (user?.displayName.isNotEmpty ?? false) ? user!.displayName[0].toUpperCase() : '?';
    final top = MediaQuery.paddingOf(context).top;
    Widget group(String title, List<Widget> tiles) => Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            SectionHeader(title, padding: const EdgeInsets.fromLTRB(20, 20, 20, 8)),
            Padding(padding: const EdgeInsets.symmetric(horizontal: 16), child: Card(child: Column(children: tiles))),
          ],
        );
    Widget tile(IconData icon, Color color, String title, String? subtitle, VoidCallback onTap) => ListTile(
          leading: IconBadge(icon: icon, color: color, size: 40),
          title: Text(title),
          subtitle: subtitle == null ? null : Text(subtitle),
          trailing: const Icon(Icons.chevron_right),
          onTap: onTap,
        );
    return Scaffold(
      body: ListView(
        padding: EdgeInsets.zero,
        children: [
          Container(
            padding: EdgeInsets.fromLTRB(20, top + 24, 20, 28),
            decoration: const BoxDecoration(
              gradient: KeikiColors.heroGradient,
              borderRadius: BorderRadius.vertical(bottom: Radius.circular(32)),
            ),
            child: Row(
              children: [
                CircleAvatar(
                  radius: 36,
                  backgroundColor: Colors.white,
                  child: Text(initial, style: theme.textTheme.headlineMedium?.copyWith(color: KeikiColors.brand)),
                ),
                const SizedBox(width: 16),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(user?.displayName ?? '', style: theme.textTheme.titleLarge?.copyWith(color: Colors.white)),
                      Text('@${user?.username ?? ''}', style: const TextStyle(color: Colors.white70)),
                      const SizedBox(height: 8),
                      Wrap(spacing: 6, children: [
                        _HeaderChip(icon: Icons.badge_outlined, text: 'ID ${user?.id ?? ''}'),
                        for (final r in user?.roles ?? const <String>[]) _HeaderChip(icon: Icons.shield_outlined, text: r),
                      ]),
                    ],
                  ),
                ),
                IconButton(
                  tooltip: 'Chỉnh sửa',
                  icon: const Icon(Icons.edit_outlined, color: Colors.white),
                  onPressed: () => context.push('/profile/edit'),
                ),
              ],
            ),
          ),
          group('Thông tin cá nhân', [
            ListTile(leading: const Icon(Icons.email_outlined), title: const Text('Email'), subtitle: Text(user?.email.isNotEmpty == true ? user!.email : 'Chưa cập nhật')),
            ListTile(leading: const Icon(Icons.phone_outlined), title: const Text('Điện thoại'), subtitle: Text(user?.phone.isNotEmpty == true ? user!.phone : 'Chưa cập nhật')),
            ListTile(leading: const Icon(Icons.home_outlined), title: const Text('Địa chỉ'), subtitle: Text(user?.address.isNotEmpty == true ? user!.address : 'Chưa cập nhật')),
          ]),
          group('Tài khoản & bảo mật', [
            tile(Icons.location_on_outlined, Colors.teal, 'Sổ địa chỉ', 'Địa chỉ giao hàng', () => context.push('/services/address/myUserAddresses')),
            tile(Icons.badge_outlined, Colors.indigo, 'Xác thực danh tính', 'eKYC bằng CCCD và khuôn mặt', () => context.push('/services/ekyc')),
            tile(Icons.admin_panel_settings_outlined, Colors.blueGrey, 'Quyền của tôi', 'Vai trò và quyền truy cập', () => context.push('/services/authorize/myPermissions')),
            tile(Icons.phonelink_lock_outlined, Colors.green, 'Xác thực 2 lớp', 'Google Authenticator', () => context.push('/profile/2fa')),
            tile(Icons.shield_outlined, Colors.deepPurple, 'Bảo mật & địa chỉ', 'Mã OTP, kiểm tra quyền, sửa địa chỉ', () => Navigator.of(context).push(MaterialPageRoute<void>(builder: (_) => const SecurityPage()))),
            tile(Icons.qr_code_2, Colors.teal, 'Mã QR của tôi', 'Sửa, đổi nội dung, thống kê quét', () => Navigator.of(context).push(MaterialPageRoute<void>(builder: (_) => const QrCodesPage()))),
            if (context.isAdmin)
              tile(Icons.delivery_dining_outlined, Colors.deepOrange, 'Chuyến giao (tài xế)', 'Đơn được giao cho bạn', () => Navigator.of(context).push(MaterialPageRoute<void>(builder: (_) => const DriverJobsPage()))),
            if (context.select<RoleCubit, bool>((c) => c.state.serverAdmin && c.state.previewCustomer))
              tile(Icons.admin_panel_settings, Colors.indigo, 'Quay lại chế độ quản trị', 'Bạn đang xem giao diện khách hàng', () => context.read<RoleCubit>().previewAsCustomer(false)),
          ]),
          group('Cài đặt', [
            tile(Icons.notifications_outlined, Colors.orange, 'Thông báo', 'Email, SMS, thiết bị', () => context.push('/notifications')),
            tile(Icons.apps_outlined, Colors.purple, 'Tất cả dịch vụ', null, () => context.go('/services')),
          ]),
          Padding(
            padding: const EdgeInsets.fromLTRB(16, 24, 16, 32),
            child: OutlinedButton.icon(
              style: OutlinedButton.styleFrom(foregroundColor: theme.colorScheme.error, side: BorderSide(color: theme.colorScheme.error.withValues(alpha: 0.5))),
              onPressed: () => _logout(context),
              icon: const Icon(Icons.logout),
              label: const Text('Đăng xuất'),
            ),
          ),
        ],
      ),
    );
  }
}

class _HeaderChip extends StatelessWidget {
  const _HeaderChip({required this.icon, required this.text});

  final IconData icon;
  final String text;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
      decoration: BoxDecoration(color: Colors.white.withValues(alpha: 0.2), borderRadius: BorderRadius.circular(999)),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [Icon(icon, size: 14, color: Colors.white), const SizedBox(width: 4), Text(text, style: const TextStyle(color: Colors.white, fontSize: 12))],
      ),
    );
  }
}

class EditProfilePage extends StatefulWidget {
  const EditProfilePage({super.key});

  @override
  State<EditProfilePage> createState() => _EditProfilePageState();
}

class _EditProfilePageState extends State<EditProfilePage> {
  late final _user = context.read<AuthBloc>().state.user!;
  late final _name = TextEditingController(text: _user.name);
  late final _email = TextEditingController(text: _user.email);
  late final _phone = TextEditingController(text: _user.phone);
  late final _address = TextEditingController(text: _user.address);
  bool _saving = false;

  @override
  void dispose() {
    for (final c in [_name, _email, _phone, _address]) {
      c.dispose();
    }
    super.dispose();
  }

  Future<void> _save() async {
    setState(() => _saving = true);
    final auth = context.read<AuthBloc>();
    final repo = context.read<AuthRepository>();
    try {
      final updated = await repo.updateProfile(
        name: _name.text.trim(),
        email: _email.text.trim(),
        phone: _phone.text.trim(),
        address: _address.text.trim(),
      );
      final fresh = updated ?? await repo.me();
      if (fresh != null) auth.add(AuthUserChanged(fresh));
      if (mounted) context.pop();
    } catch (e) {
      if (mounted) {
        setState(() => _saving = false);
        showSnack(context, e.toString());
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Chỉnh sửa hồ sơ')),
      body: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          TextField(controller: _name, decoration: const InputDecoration(labelText: 'Họ tên')),
          const SizedBox(height: 12),
          TextField(controller: _email, keyboardType: TextInputType.emailAddress, decoration: const InputDecoration(labelText: 'Email')),
          const SizedBox(height: 12),
          TextField(controller: _phone, keyboardType: TextInputType.phone, decoration: const InputDecoration(labelText: 'Điện thoại')),
          const SizedBox(height: 12),
          TextField(controller: _address, decoration: const InputDecoration(labelText: 'Địa chỉ')),
          const SizedBox(height: 24),
          FilledButton(onPressed: _saving ? null : _save, child: const Text('Lưu')),
        ],
      ),
    );
  }
}
