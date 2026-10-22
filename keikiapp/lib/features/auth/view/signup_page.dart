import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:go_router/go_router.dart';

import '../../../core/widgets/status_view.dart';
import '../data/auth_repository.dart';

class SignUpPage extends StatefulWidget {
  const SignUpPage({super.key});

  @override
  State<SignUpPage> createState() => _SignUpPageState();
}

class _SignUpPageState extends State<SignUpPage> {
  final _formKey = GlobalKey<FormState>();
  final _username = TextEditingController();
  final _password = TextEditingController();
  final _email = TextEditingController();
  final _name = TextEditingController();
  final _phone = TextEditingController();
  bool _submitting = false;

  @override
  void dispose() {
    for (final c in [_username, _password, _email, _name, _phone]) {
      c.dispose();
    }
    super.dispose();
  }

  Future<void> _submit() async {
    if (!_formKey.currentState!.validate()) return;
    setState(() => _submitting = true);
    try {
      await context.read<AuthRepository>().signUp(
            username: _username.text.trim(),
            password: _password.text,
            email: _email.text.trim(),
            name: _name.text.trim(),
            phone: _phone.text.trim(),
          );
      if (!mounted) return;
      showSnack(context, 'Đăng ký thành công, hãy đăng nhập');
      context.pop();
    } catch (e) {
      if (mounted) showSnack(context, e.toString());
    } finally {
      if (mounted) setState(() => _submitting = false);
    }
  }

  String? _required(String? v) => (v == null || v.trim().isEmpty) ? 'Bắt buộc' : null;

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Tạo tài khoản')),
      body: Form(
        key: _formKey,
        child: ListView(
          padding: const EdgeInsets.all(24),
          children: [
            TextFormField(controller: _name, decoration: const InputDecoration(labelText: 'Họ tên'), validator: _required),
            const SizedBox(height: 16),
            TextFormField(controller: _username, decoration: const InputDecoration(labelText: 'Tên đăng nhập'), validator: _required),
            const SizedBox(height: 16),
            TextFormField(
              controller: _email,
              keyboardType: TextInputType.emailAddress,
              decoration: const InputDecoration(labelText: 'Email'),
              validator: (v) => (v == null || !v.contains('@')) ? 'Email không hợp lệ' : null,
            ),
            const SizedBox(height: 16),
            TextFormField(controller: _phone, keyboardType: TextInputType.phone, decoration: const InputDecoration(labelText: 'Số điện thoại')),
            const SizedBox(height: 16),
            TextFormField(
              controller: _password,
              obscureText: true,
              decoration: const InputDecoration(labelText: 'Mật khẩu'),
              validator: (v) => (v == null || v.length < 6) ? 'Tối thiểu 6 ký tự' : null,
            ),
            const SizedBox(height: 24),
            FilledButton(onPressed: _submitting ? null : _submit, child: const Text('Đăng ký')),
          ],
        ),
      ),
    );
  }
}
