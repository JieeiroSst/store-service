import 'dart:async';
import 'dart:convert';

import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:flutter_webrtc/flutter_webrtc.dart';
import 'package:web_socket_channel/io.dart';

import '../../../core/network/graphql_client.dart';
import '../../../gen/models.g.dart';
import '../kit/kit.dart';

const _fatal = {'room_full', 'replaced', 'invalid_id'};

class _Peer {
  _Peer(this.pc, this.renderer);

  final RTCPeerConnection pc;
  final RTCVideoRenderer renderer;
  final pending = <RTCIceCandidate>[];
  bool remoteSet = false;
  bool audio = true;
  bool video = true;
}

class CallPage extends StatefulWidget {
  const CallPage({super.key, required this.roomId});

  final String roomId;

  @override
  State<CallPage> createState() => _CallPageState();
}

class _CallPageState extends State<CallPage> {
  final _local = RTCVideoRenderer();
  final _peers = <String, _Peer>{};
  MediaStream? _stream;
  IOWebSocketChannel? _ws;
  StreamSubscription<dynamic>? _sub;
  List<Map<String, dynamic>> _ice = const [];
  bool _mic = true;
  bool _cam = true;
  bool _connecting = true;
  String? _error;

  @override
  void initState() {
    super.initState();
    _start();
  }

  Future<void> _start() async {
    if (kIsWeb) {
      setState(() => _error = 'Gọi video chưa hỗ trợ trên bản web, hãy dùng app Android/iOS');
      return;
    }
    final client = context.read<GraphQLClient>();
    final api = context.api.webrtc;
    try {
      await _local.initialize();
      final servers = await api.webrtcIceServers();
      _ice = [
        for (final s in servers?.iceServers ?? const <WebrtcIceServer>[])
          {'urls': s.urls ?? const <String>[], 'username': ?s.username, 'credential': ?s.credential},
      ];
      _stream = await navigator.mediaDevices.getUserMedia({
        'audio': true,
        'video': {'facingMode': 'user'},
      });
      _local.srcObject = _stream;
      final token = await client.accessToken();
      final ws = IOWebSocketChannel.connect(client.signalingUri(widget.roomId), headers: {'Authorization': 'Bearer $token'});
      _ws = ws;
      await ws.ready;
      _sub = ws.stream.listen(
        (raw) => _onMessage(jsonDecode(raw as String) as Map<String, dynamic>),
        onDone: () {
          if (mounted && _error == null) setState(() => _error = 'Mất kết nối tới máy chủ gọi');
        },
      );
      if (mounted) setState(() => _connecting = false);
    } catch (e) {
      if (mounted) setState(() => _error = 'Không bắt đầu được cuộc gọi: ${errorMessage(e)}');
    }
  }

  void _send(Map<String, dynamic> msg) => _ws?.sink.add(jsonEncode(msg));

  Future<_Peer> _peer(String id) async {
    final existing = _peers[id];
    if (existing != null) return existing;
    final pc = await createPeerConnection({'iceServers': _ice, 'sdpSemantics': 'unified-plan'});
    final renderer = RTCVideoRenderer();
    await renderer.initialize();
    final peer = _Peer(pc, renderer);
    for (final track in _stream?.getTracks() ?? const <MediaStreamTrack>[]) {
      await pc.addTrack(track, _stream!);
    }
    pc.onIceCandidate = (c) {
      if (c.candidate != null) _send({'type': 'ice_candidate', 'to_user_id': id, 'data': c.toMap()});
    };
    pc.onTrack = (e) {
      if (e.streams.isNotEmpty && mounted) setState(() => renderer.srcObject = e.streams.first);
    };
    pc.onConnectionState = (s) {
      if (s == RTCPeerConnectionState.RTCPeerConnectionStateFailed) pc.restartIce();
    };
    setState(() => _peers[id] = peer);
    return peer;
  }

  Future<void> _flush(_Peer p) async {
    p.remoteSet = true;
    for (final c in p.pending) {
      await p.pc.addCandidate(c);
    }
    p.pending.clear();
  }

  Future<void> _onMessage(Map<String, dynamic> m) async {
    final from = m['user_id'] as String?;
    final data = m['data'];
    switch (m['type']) {
      case 'room_users':
        for (final id in (data as List? ?? const []).cast<String>()) {
          final p = await _peer(id);
          final offer = await p.pc.createOffer();
          await p.pc.setLocalDescription(offer);
          _send({'type': 'offer', 'to_user_id': id, 'data': offer.toMap()});
        }
        _sendMediaState();
      case 'offer':
        if (from == null) return;
        final p = await _peer(from);
        await p.pc.setRemoteDescription(RTCSessionDescription(data['sdp'] as String?, data['type'] as String?));
        await _flush(p);
        final answer = await p.pc.createAnswer();
        await p.pc.setLocalDescription(answer);
        _send({'type': 'answer', 'to_user_id': from, 'data': answer.toMap()});
      case 'answer':
        final p = _peers[from];
        if (p == null) return;
        await p.pc.setRemoteDescription(RTCSessionDescription(data['sdp'] as String?, data['type'] as String?));
        await _flush(p);
      case 'ice_candidate':
        final p = _peers[from];
        if (p == null || data is! Map) return;
        final c = RTCIceCandidate(data['candidate'] as String?, data['sdpMid'] as String?, (data['sdpMLineIndex'] as num?)?.toInt());
        if (p.remoteSet) {
          await p.pc.addCandidate(c);
        } else {
          p.pending.add(c);
        }
      case 'media_state':
        final p = _peers[from];
        if (p != null && data is Map && mounted) {
          setState(() {
            p.audio = data['audio'] != false;
            p.video = data['video'] != false;
          });
        }
      case 'user_left':
        await _drop(from);
      case 'error':
        final code = data is Map ? data['code'] as String? : null;
        final message = switch (code) {
          'room_full' => 'Phòng đã đủ người',
          'replaced' => 'Bạn đã vào phòng này trên thiết bị khác',
          _ => data is Map ? (data['message'] as String? ?? 'Lỗi cuộc gọi') : 'Lỗi cuộc gọi',
        };
        if (_fatal.contains(code)) {
          await _sub?.cancel();
          if (mounted) setState(() => _error = message);
        } else if (mounted) {
          showSnack(context, message);
        }
    }
  }

  Future<void> _drop(String? id) async {
    final p = _peers.remove(id);
    if (p == null) return;
    await p.pc.close();
    p.renderer.srcObject = null;
    await p.renderer.dispose();
    if (mounted) setState(() {});
  }

  void _sendMediaState() => _send({
        'type': 'media_state',
        'data': {'audio': _mic, 'video': _cam, 'screen': false},
      });

  void _toggleMic() {
    setState(() => _mic = !_mic);
    for (final t in _stream?.getAudioTracks() ?? const <MediaStreamTrack>[]) {
      t.enabled = _mic;
    }
    _sendMediaState();
  }

  void _toggleCam() {
    setState(() => _cam = !_cam);
    for (final t in _stream?.getVideoTracks() ?? const <MediaStreamTrack>[]) {
      t.enabled = _cam;
    }
    _sendMediaState();
  }

  Future<void> _switchCam() async {
    final tracks = _stream?.getVideoTracks() ?? const <MediaStreamTrack>[];
    if (tracks.isNotEmpty) await Helper.switchCamera(tracks.first);
  }

  @override
  void dispose() {
    _sub?.cancel();
    _ws?.sink.close();
    for (final p in _peers.values) {
      p.pc.close();
      p.renderer.dispose();
    }
    for (final t in _stream?.getTracks() ?? const <MediaStreamTrack>[]) {
      t.stop();
    }
    _stream?.dispose();
    _local.dispose();
    super.dispose();
  }

  Widget _tile(RTCVideoRenderer r, String label, {bool mirror = false, bool video = true, bool audio = true}) {
    return ClipRRect(
      borderRadius: BorderRadius.circular(18),
      child: Container(
        color: const Color(0xFF1F2937),
        child: Stack(
          fit: StackFit.expand,
          children: [
            if (video && r.srcObject != null)
              RTCVideoView(r, mirror: mirror, objectFit: RTCVideoViewObjectFit.RTCVideoViewObjectFitCover)
            else
              Center(
                child: CircleAvatar(
                  radius: 32,
                  backgroundColor: Colors.white12,
                  child: Text(label.isEmpty ? '?' : label.characters.first.toUpperCase(), style: const TextStyle(color: Colors.white, fontSize: 24)),
                ),
              ),
            Positioned(
              left: 8,
              bottom: 8,
              child: Container(
                padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
                decoration: BoxDecoration(color: Colors.black54, borderRadius: BorderRadius.circular(8)),
                child: Row(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    if (!audio) const Padding(padding: EdgeInsets.only(right: 4), child: Icon(Icons.mic_off, size: 14, color: Colors.white)),
                    Text(label, style: const TextStyle(color: Colors.white, fontSize: 12)),
                  ],
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final remotes = _peers.entries.toList();
    return Scaffold(
      backgroundColor: const Color(0xFF111827),
      appBar: AppBar(
        backgroundColor: const Color(0xFF111827),
        foregroundColor: Colors.white,
        title: Text('Phòng ${widget.roomId}', style: const TextStyle(color: Colors.white)),
        actions: [
          Padding(
            padding: const EdgeInsets.only(right: 16),
            child: Center(child: Text('${remotes.length + 1} người', style: const TextStyle(color: Colors.white70))),
          ),
        ],
      ),
      body: _error != null
          ? Center(
              child: Padding(
                padding: const EdgeInsets.all(32),
                child: Text(_error!, textAlign: TextAlign.center, style: const TextStyle(color: Colors.white, fontSize: 16)),
              ),
            )
          : Column(
              children: [
                Expanded(
                  child: Padding(
                    padding: const EdgeInsets.all(12),
                    child: remotes.isEmpty
                        ? Stack(
                            children: [
                              Positioned.fill(child: _tile(_local, 'Bạn', mirror: true, video: _cam, audio: _mic)),
                              Center(
                                child: Container(
                                  padding: const EdgeInsets.all(16),
                                  decoration: BoxDecoration(color: Colors.black54, borderRadius: BorderRadius.circular(16)),
                                  child: Text(
                                    _connecting ? 'Đang kết nối…' : 'Đang chờ người khác vào phòng\nMã phòng: ${widget.roomId}',
                                    textAlign: TextAlign.center,
                                    style: const TextStyle(color: Colors.white),
                                  ),
                                ),
                              ),
                            ],
                          )
                        : Stack(
                            children: [
                              GridView.count(
                                crossAxisCount: remotes.length == 1 ? 1 : 2,
                                mainAxisSpacing: 8,
                                crossAxisSpacing: 8,
                                childAspectRatio: remotes.length == 1 ? 0.62 : 0.75,
                                children: [for (final e in remotes) _tile(e.value.renderer, e.key, video: e.value.video, audio: e.value.audio)],
                              ),
                              Positioned(
                                right: 12,
                                bottom: 12,
                                width: 110,
                                height: 150,
                                child: _tile(_local, 'Bạn', mirror: true, video: _cam, audio: _mic),
                              ),
                            ],
                          ),
                  ),
                ),
                SafeArea(
                  top: false,
                  child: Padding(
                    padding: const EdgeInsets.only(bottom: 16, top: 4),
                    child: Row(
                      mainAxisAlignment: MainAxisAlignment.spaceEvenly,
                      children: [
                        _Round(icon: _mic ? Icons.mic : Icons.mic_off, active: _mic, onTap: _toggleMic),
                        _Round(icon: _cam ? Icons.videocam : Icons.videocam_off, active: _cam, onTap: _toggleCam),
                        _Round(icon: Icons.cameraswitch, active: true, onTap: _switchCam),
                        _Round(icon: Icons.call_end, active: false, color: const Color(0xFFDC2626), onTap: () => Navigator.pop(context)),
                      ],
                    ),
                  ),
                ),
              ],
            ),
    );
  }
}

class _Round extends StatelessWidget {
  const _Round({required this.icon, required this.active, required this.onTap, this.color});

  final IconData icon;
  final bool active;
  final VoidCallback onTap;
  final Color? color;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: color ?? (active ? Colors.white24 : Colors.white),
      shape: const CircleBorder(),
      child: InkWell(
        customBorder: const CircleBorder(),
        onTap: onTap,
        child: SizedBox(width: 60, height: 60, child: Icon(icon, color: color != null ? Colors.white : (active ? Colors.white : Colors.black87))),
      ),
    );
  }
}
