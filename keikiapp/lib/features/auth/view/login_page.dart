import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:go_router/go_router.dart';

import '../../../core/widgets/status_view.dart';
import '../../../core/widgets/ui.dart';
import '../bloc/auth_bloc.dart';
import 'mfa_code_form.dart';

class LoginPage extends StatefulWidget {
  const LoginPage({super.key});

  @override
  State<LoginPage> createState() => _LoginPageState();
}

class _LoginPageState extends State<LoginPage> {
  final _formKey = GlobalKey<FormState>();
  final _username = TextEditingController();
  final _password = TextEditingController();
  bool _obscure = true;

  @override
  void dispose() {
    _username.dispose();
    _password.dispose();
    super.dispose();
  }

  void _submit() {
    if (!_formKey.currentState!.validate()) return;
    context.read<AuthBloc>().add(AuthLoginRequested(_username.text.trim(), _password.text));
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Scaffold(
      body: BlocListener<AuthBloc, AuthState>(
        listenWhen: (prev, curr) => curr.error != null && prev.error != curr.error,
        listener: (context, state) => showSnack(context, state.error!),
        child: CustomScrollView(
          physics: const ClampingScrollPhysics(),
          slivers: [
            SliverToBoxAdapter(
              child: Stack(
                children: [
                  HeroBackdrop(
                    child: SafeArea(
                      bottom: false,
                      child: SizedBox(
                        width: double.infinity,
                        child: Column(
                          children: [
                            const SizedBox(height: 56),
                            const KeikiLogo(size: 84, onDark: true),
                            const SizedBox(height: 10),
                            Text(
                              'Mọi dịch vụ, một ứng dụng',
                              style: theme.textTheme.bodyLarge?.copyWith(color: Colors.white.withValues(alpha: 0.85)),
                            ),
                            const SizedBox(height: 18),
                            Wrap(
                              spacing: 8,
                              runSpacing: 8,
                              alignment: WrapAlignment.center,
                              children: [
                                for (final (icon, label) in const [
                                  (Icons.account_balance_wallet_rounded, 'Ví'),
                                  (Icons.restaurant_rounded, 'Đồ ăn'),
                                  (Icons.shopping_bag_rounded, 'Mua sắm'),
                                  (Icons.auto_awesome_rounded, 'AI'),
                                ])
                                  Container(
                                    padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
                                    decoration: BoxDecoration(
                                      color: Colors.white.withValues(alpha: 0.14),
                                      borderRadius: BorderRadius.circular(999),
                                      border: Border.all(color: Colors.white.withValues(alpha: 0.22)),
                                    ),
                                    child: Row(
                                      mainAxisSize: MainAxisSize.min,
                                      children: [
                                        Icon(icon, size: 15, color: Colors.white),
                                        const SizedBox(width: 6),
                                        Text(
                                          label,
                                          style: const TextStyle(color: Colors.white, fontWeight: FontWeight.w600, fontSize: 12),
                                        ),
                                      ],
                                    ),
                                  ),
                              ],
                            ),
                            const SizedBox(height: 36),
                            const SizedBox(height: 32),
                          ],
                        ),
                      ),
                    ),
                  ),
                  Positioned(
                    left: 0,
                    right: 0,
                    bottom: -1,
                    height: 33,
                    child: DecoratedBox(
                      decoration: BoxDecoration(
                        color: theme.scaffoldBackgroundColor,
                        borderRadius: const BorderRadius.vertical(top: Radius.circular(32)),
                      ),
                    ),
                  ),
                ],
              ),
            ),
            SliverToBoxAdapter(
              child: Container(
                padding: const EdgeInsets.fromLTRB(24, 4, 24, 24),
                decoration: BoxDecoration(
                  color: theme.scaffoldBackgroundColor,
                  borderRadius: const BorderRadius.vertical(top: Radius.circular(32)),
                ),
                child: Center(
                  child: ConstrainedBox(
                    constraints: const BoxConstraints(maxWidth: 420),
                    child: BlocBuilder<AuthBloc, AuthState>(
                      buildWhen: (p, c) => p.awaitingMfa != c.awaitingMfa,
                      builder: (context, state) => state.awaitingMfa
                          ? const MfaCodeForm()
                          : Form(
                          key: _formKey,
                          child: Column(
                            mainAxisSize: MainAxisSize.min,
                            crossAxisAlignment: CrossAxisAlignment.stretch,
                            children: [
                              Text('Đăng nhập', style: theme.textTheme.headlineSmall),
                              const SizedBox(height: 4),
                              Text(
                                'Chào mừng bạn quay lại 👋',
                                style: theme.textTheme.bodyMedium?.copyWith(color: theme.colorScheme.onSurfaceVariant),
                              ),
                              const SizedBox(height: 24),
                              TextFormField(
                                controller: _username,
                                textInputAction: TextInputAction.next,
                                autofillHints: const [AutofillHints.username],
                                decoration: const InputDecoration(labelText: 'Tên đăng nhập', prefixIcon: Icon(Icons.person_outline)),
                                validator: (v) => (v == null || v.trim().isEmpty) ? 'Vui lòng nhập tên đăng nhập' : null,
                              ),
                              const SizedBox(height: 16),
                              TextFormField(
                                controller: _password,
                                obscureText: _obscure,
                                autofillHints: const [AutofillHints.password],
                                onFieldSubmitted: (_) => _submit(),
                                decoration: InputDecoration(
                                  labelText: 'Mật khẩu',
                                  prefixIcon: const Icon(Icons.lock_outline),
                                  suffixIcon: IconButton(
                                    icon: Icon(_obscure ? Icons.visibility : Icons.visibility_off),
                                    onPressed: () => setState(() => _obscure = !_obscure),
                                  ),
                                ),
                                validator: (v) => (v == null || v.isEmpty) ? 'Vui lòng nhập mật khẩu' : null,
                              ),
                              const SizedBox(height: 28),
                              BlocBuilder<AuthBloc, AuthState>(
                                buildWhen: (p, c) => p.submitting != c.submitting,
                                builder: (context, state) => LoadingButton(label: 'Đăng nhập', loading: state.submitting, onPressed: _submit),
                              ),
                              const SizedBox(height: 12),
                              Wrap(
                                alignment: WrapAlignment.center,
                                crossAxisAlignment: WrapCrossAlignment.center,
                                children: [
                                  const Text('Chưa có tài khoản?'),
                                  TextButton(onPressed: () => context.push('/signup'), child: const Text('Đăng ký ngay')),
                                ],
                              ),
                            ],
                          ),
                        ),
                    ),
                  ),
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }
}
