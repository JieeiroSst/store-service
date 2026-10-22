import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

import '../../../core/utils/formatters.dart';
import '../../../core/widgets/status_view.dart';
import '../bloc/notification_settings_cubit.dart';

class NotificationSettingsPage extends StatelessWidget {
  const NotificationSettingsPage({super.key});

  Future<void> _editContact(BuildContext context, NotificationSettingsState state) async {
    final email = TextEditingController(text: state.contact.email);
    final phone = TextEditingController(text: state.contact.phone);
    final ok = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Kênh nhận thông báo'),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            TextField(controller: email, keyboardType: TextInputType.emailAddress, decoration: const InputDecoration(labelText: 'Email')),
            TextField(controller: phone, keyboardType: TextInputType.phone, decoration: const InputDecoration(labelText: 'Số điện thoại')),
          ],
        ),
        actions: [
          TextButton(onPressed: () => Navigator.pop(context, false), child: const Text('Huỷ')),
          FilledButton(onPressed: () => Navigator.pop(context, true), child: const Text('Lưu')),
        ],
      ),
    );
    if (ok != true || !context.mounted) return;
    try {
      await context.read<NotificationSettingsCubit>().updateContact(email.text.trim(), phone.text.trim());
    } catch (e) {
      if (context.mounted) showSnack(context, e.toString());
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Thông báo')),
      body: BlocBuilder<NotificationSettingsCubit, NotificationSettingsState>(
        builder: (context, state) => StatusView(
          status: state.status,
          error: state.error,
          onRetry: context.read<NotificationSettingsCubit>().load,
          builder: (context) => RefreshIndicator(
            onRefresh: context.read<NotificationSettingsCubit>().load,
            child: ListView(
              children: [
                ListTile(
                  leading: const Icon(Icons.alternate_email),
                  title: const Text('Email nhận thông báo'),
                  subtitle: Text(state.contact.email.isEmpty ? 'Chưa thiết lập' : state.contact.email),
                  trailing: const Icon(Icons.edit_outlined),
                  onTap: () => _editContact(context, state),
                ),
                ListTile(
                  leading: const Icon(Icons.sms_outlined),
                  title: const Text('SMS'),
                  subtitle: Text(state.contact.phone.isEmpty ? 'Chưa thiết lập' : state.contact.phone),
                  trailing: const Icon(Icons.edit_outlined),
                  onTap: () => _editContact(context, state),
                ),
                const Divider(),
                const ListTile(title: Text('Thiết bị nhận push', style: TextStyle(fontWeight: FontWeight.bold))),
                if (state.devices.isEmpty)
                  const ListTile(subtitle: Text('Chưa có thiết bị nào đăng ký nhận push')),
                for (final d in state.devices)
                  ListTile(
                    leading: Icon(d.deviceType.toLowerCase() == 'ios' ? Icons.phone_iphone : Icons.phone_android),
                    title: Text(d.deviceId.isEmpty ? d.tokenPreview : d.deviceId),
                    subtitle: Text('Dùng lần cuối ${formatDate(d.lastUsedAt)}'),
                    trailing: Icon(d.isActive ? Icons.notifications_active : Icons.notifications_off),
                  ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}
