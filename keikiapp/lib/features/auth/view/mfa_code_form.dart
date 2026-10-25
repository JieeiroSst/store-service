import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

import '../../../core/widgets/ui.dart';
import '../bloc/auth_bloc.dart';

class MfaCodeForm extends StatefulWidget {
  const MfaCodeForm({super.key});

  @override
  State<MfaCodeForm> createState() => _MfaCodeFormState();
}

class _MfaCodeFormState extends State<MfaCodeForm> {
  final _code = TextEditingController();
  bool _recovery = false;

  @override
  void dispose() {
    _code.dispose();
    super.dispose();
  }

  void _submit() {
    final code = _code.text.trim();
    if (_recovery ? code.length < 10 : code.length != 6) return;
    context.read<AuthBloc>().add(AuthMfaSubmitted(code));
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Column(
      mainAxisSize: MainAxisSize.min,
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        Text('Xác thực 2 lớp', style: theme.textTheme.headlineSmall),
        const SizedBox(height: 4),
        Text(
          _recovery
              ? 'Nhập một mã khôi phục bạn đã lưu khi bật xác thực 2 lớp. Mỗi mã chỉ dùng được một lần.'
              : 'Mở Google Authenticator và nhập mã 6 số của Keikibook.',
          style: theme.textTheme.bodyMedium?.copyWith(color: theme.colorScheme.onSurfaceVariant),
        ),
        const SizedBox(height: 24),
        TextField(
          key: ValueKey(_recovery),
          controller: _code,
          autofocus: true,
          textAlign: TextAlign.center,
          autofillHints: const [AutofillHints.oneTimeCode],
          keyboardType: _recovery ? TextInputType.text : TextInputType.number,
          maxLength: _recovery ? 11 : 6,
          inputFormatters: _recovery ? null : [FilteringTextInputFormatter.digitsOnly],
          style: theme.textTheme.headlineSmall?.copyWith(letterSpacing: _recovery ? 2 : 10),
          onChanged: (v) {
            if (!_recovery && v.length == 6) _submit();
          },
          onSubmitted: (_) => _submit(),
          decoration: InputDecoration(
            hintText: _recovery ? 'xxxxx-xxxxx' : '000000',
            counterText: '',
            prefixIcon: Icon(_recovery ? Icons.key_outlined : Icons.lock_clock_outlined),
          ),
        ),
        const SizedBox(height: 24),
        BlocBuilder<AuthBloc, AuthState>(
          buildWhen: (p, c) => p.submitting != c.submitting,
          builder: (context, state) => LoadingButton(label: 'Xác nhận', loading: state.submitting, onPressed: _submit),
        ),
        const SizedBox(height: 8),
        TextButton(
          onPressed: () => setState(() {
            _recovery = !_recovery;
            _code.clear();
          }),
          child: Text(_recovery ? 'Dùng mã từ ứng dụng Authenticator' : 'Không dùng được điện thoại? Nhập mã khôi phục'),
        ),
        TextButton(onPressed: () => context.read<AuthBloc>().add(const AuthMfaCancelled()), child: const Text('Quay lại đăng nhập')),
      ],
    );
  }
}
