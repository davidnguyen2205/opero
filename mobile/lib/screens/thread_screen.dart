import 'package:flutter/material.dart';

import '../mock/field_mock.dart';
import '../theme.dart';

/// MOCK — a chat thread. No messaging API exists in v1; messages you "send" are
/// appended to local state only and are never persisted or delivered.
class ThreadScreen extends StatefulWidget {
  final MockThread thread;
  const ThreadScreen({super.key, required this.thread});

  @override
  State<ThreadScreen> createState() => _ThreadScreenState();
}

class _ThreadScreenState extends State<ThreadScreen> {
  late final List<MockMessage> _messages = [...widget.thread.messages];
  final _controller = TextEditingController();

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  void _send() {
    final text = _controller.text.trim();
    if (text.isEmpty) return;
    setState(() {
      _messages.add(MockMessage(fromMe: true, text: text, time: 'now'));
      _controller.clear();
    });
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        leading: IconButton(
          icon: const Icon(Icons.chevron_left, color: AppColors.grey700),
          onPressed: () => Navigator.of(context).maybePop(),
        ),
        title: Row(
          children: [
            Text(widget.thread.name, style: kCardTitle),
            const SizedBox(width: 8),
            const DemoBadge(),
          ],
        ),
      ),
      body: Column(
        children: [
          Expanded(
            child: Container(
              color: AppColors.grey50,
              child: ListView.builder(
                padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 16),
                itemCount: _messages.length,
                itemBuilder: (context, i) => _bubble(_messages[i]),
              ),
            ),
          ),
          _composer(),
        ],
      ),
    );
  }

  Widget _bubble(MockMessage m) {
    final me = m.fromMe;
    return Align(
      alignment: me ? Alignment.centerRight : Alignment.centerLeft,
      child: Container(
        constraints: BoxConstraints(maxWidth: MediaQuery.of(context).size.width * 0.78),
        margin: const EdgeInsets.only(bottom: 10),
        child: Column(
          crossAxisAlignment: me ? CrossAxisAlignment.end : CrossAxisAlignment.start,
          children: [
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 13, vertical: 9),
              decoration: BoxDecoration(
                color: me ? AppColors.orange : Colors.white,
                borderRadius: BorderRadius.only(
                  topLeft: const Radius.circular(16),
                  topRight: const Radius.circular(16),
                  bottomLeft: Radius.circular(me ? 16 : 4),
                  bottomRight: Radius.circular(me ? 4 : 16),
                ),
                border: me ? null : Border.all(color: AppColors.grey200),
              ),
              child: Text(m.text,
                  style: TextStyle(fontSize: 14, height: 1.4, color: me ? Colors.white : AppColors.grey900)),
            ),
            const SizedBox(height: 3),
            Text(m.time, style: const TextStyle(fontSize: 10.5, color: AppColors.grey400)),
          ],
        ),
      ),
    );
  }

  Widget _composer() {
    return SafeArea(
      top: false,
      child: Container(
        padding: const EdgeInsets.fromLTRB(12, 10, 12, 10),
        decoration: const BoxDecoration(
          color: Colors.white,
          border: Border(top: BorderSide(color: AppColors.grey200)),
        ),
        child: Row(
          children: [
            Expanded(
              child: TextField(
                controller: _controller,
                textInputAction: TextInputAction.send,
                onSubmitted: (_) => _send(),
                style: const TextStyle(fontSize: 14, color: AppColors.grey900),
                decoration: InputDecoration(
                  hintText: 'Message…',
                  isDense: true,
                  filled: true,
                  fillColor: AppColors.grey50,
                  contentPadding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
                  border: OutlineInputBorder(
                    borderRadius: BorderRadius.circular(21),
                    borderSide: const BorderSide(color: AppColors.grey200),
                  ),
                  enabledBorder: OutlineInputBorder(
                    borderRadius: BorderRadius.circular(21),
                    borderSide: const BorderSide(color: AppColors.grey200),
                  ),
                  focusedBorder: OutlineInputBorder(
                    borderRadius: BorderRadius.circular(21),
                    borderSide: const BorderSide(color: AppColors.orange),
                  ),
                ),
              ),
            ),
            const SizedBox(width: 8),
            Material(
              color: AppColors.orange,
              shape: const CircleBorder(),
              child: InkWell(
                customBorder: const CircleBorder(),
                onTap: _send,
                child: const SizedBox(
                  width: 42,
                  height: 42,
                  child: Icon(Icons.send, size: 18, color: Colors.white),
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }
}
