import 'package:flutter/material.dart';

import '../api/api_client.dart';
import '../api/auth_store.dart';

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
      await widget.auth.setToken(res.token);
      if (mounted) widget.onAuthenticated();
    } on ApiException catch (e) {
      setState(() => _error = e.message);
    } catch (e) {
      setState(() => _error = 'Could not reach the server. Check your connection.');
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Opero — Sign in')),
      body: Padding(
        padding: const EdgeInsets.all(16),
        child: ListView(
          children: [
            TextField(
              controller: _tenant,
              decoration: const InputDecoration(labelText: 'Company (tenant slug)'),
              autocorrect: false,
            ),
            TextField(
              controller: _email,
              decoration: const InputDecoration(labelText: 'Email'),
              keyboardType: TextInputType.emailAddress,
              autocorrect: false,
            ),
            TextField(
              controller: _password,
              decoration: const InputDecoration(labelText: 'Password'),
              obscureText: true,
            ),
            const SizedBox(height: 16),
            if (_error != null)
              Padding(
                padding: const EdgeInsets.only(bottom: 12),
                child: Text(_error!, style: const TextStyle(color: Colors.red)),
              ),
            FilledButton(
              onPressed: _busy ? null : _submit,
              child: _busy
                  ? const SizedBox(height: 18, width: 18, child: CircularProgressIndicator(strokeWidth: 2))
                  : const Text('Sign in'),
            ),
          ],
        ),
      ),
    );
  }
}
