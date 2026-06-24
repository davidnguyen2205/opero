import 'package:flutter/material.dart';

import '../api/api_client.dart';
import '../api/auth_store.dart';
import '../theme.dart';

/// REAL — POST /auth/login. Restyled to the prototype look. Unlike the
/// prototype mock (email + password only), Opero login requires a tenant slug,
/// so we keep that field (the API is multi-tenant).
class LoginScreen extends StatefulWidget {
  final ApiClient api;
  final AuthStore auth;
  final VoidCallback onAuthenticated;

  const LoginScreen({
    super.key,
    required this.api,
    required this.auth,
    required this.onAuthenticated,
  });

  @override
  State<LoginScreen> createState() => _LoginScreenState();
}

class _LoginScreenState extends State<LoginScreen> {
  final _tenant = TextEditingController();
  final _email = TextEditingController();
  final _password = TextEditingController();
  bool _busy = false;
  String? _error;

  @override
  void dispose() {
    _tenant.dispose();
    _email.dispose();
    _password.dispose();
    super.dispose();
  }

  Future<void> _submit() async {
    if (_busy) return;
    setState(() {
      _busy = true;
      _error = null;
    });
    try {
      final res = await widget.api.login(
        tenantSlug: _tenant.text.trim(),
        email: _email.text.trim(),
        password: _password.text,
      );
      await widget.auth.setSession(res);
      if (mounted) widget.onAuthenticated();
    } on ApiException catch (e) {
      setState(() => _error = e.message);
    } catch (_) {
      setState(() => _error = 'Could not reach the server. Check your connection.');
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: Colors.white,
      body: SafeArea(
        child: SingleChildScrollView(
          padding: const EdgeInsets.symmetric(horizontal: 24),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              const SizedBox(height: 56),
              const Center(child: OperoMark(size: 56)),
              const SizedBox(height: 14),
              const Center(
                child: Text('Welcome to Opero',
                    style: TextStyle(fontSize: 26, fontWeight: FontWeight.w700, color: AppColors.ink)),
              ),
              const SizedBox(height: 4),
              const Center(
                child: Text('Sign in to your work account',
                    style: TextStyle(fontSize: 14, color: AppColors.grey500)),
              ),
              const SizedBox(height: 36),
              _label('Company'),
              _field(_tenant, hint: 'tenant slug', autofill: false),
              const SizedBox(height: 16),
              _label('Work email'),
              _field(_email, keyboard: TextInputType.emailAddress),
              const SizedBox(height: 16),
              _label('Password'),
              _field(_password, obscure: true),
              if (_error != null) ...[
                const SizedBox(height: 14),
                Text(_error!, style: const TextStyle(color: AppColors.red, fontSize: 13)),
              ],
              const SizedBox(height: 24),
              PhoneButton(
                label: _busy ? 'Signing in…' : 'Sign In',
                onPressed: _busy ? null : _submit,
              ),
              const SizedBox(height: 24),
              const Center(
                child: Text('Need access? Ask your manager to invite you.',
                    style: TextStyle(fontSize: 12.5, color: AppColors.grey400)),
              ),
              const SizedBox(height: 24),
            ],
          ),
        ),
      ),
    );
  }

  Widget _label(String t) => Padding(
        padding: const EdgeInsets.only(bottom: 7),
        child: Text(t,
            style: const TextStyle(fontSize: 12.5, fontWeight: FontWeight.w600, color: AppColors.grey700)),
      );

  Widget _field(
    TextEditingController c, {
    bool obscure = false,
    bool autofill = true,
    String? hint,
    TextInputType? keyboard,
  }) {
    return TextField(
      controller: c,
      obscureText: obscure,
      autocorrect: false,
      keyboardType: keyboard,
      style: const TextStyle(fontSize: 15, color: AppColors.grey900),
      decoration: InputDecoration(
        hintText: hint,
        isDense: true,
        filled: true,
        fillColor: AppColors.grey50,
        contentPadding: const EdgeInsets.symmetric(horizontal: 14, vertical: 16),
        enabledBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(12),
          borderSide: const BorderSide(color: AppColors.grey200),
        ),
        focusedBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(12),
          borderSide: const BorderSide(color: AppColors.orange),
        ),
      ),
    );
  }
}
