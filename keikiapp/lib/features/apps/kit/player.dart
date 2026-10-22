import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:video_player/video_player.dart';

import '../../../core/network/graphql_client.dart';
import 'kit.dart';

Uri videoServiceUri(BuildContext context, String videoId, String path) =>
    context.read<GraphQLClient>().mediaUri('video-service', 'api/videos/${Uri.encodeComponent(videoId)}/$path');

List<Uri> videoSources(BuildContext context, String videoId) => [
      videoServiceUri(context, videoId, 'hls/master.m3u8'),
      videoServiceUri(context, videoId, 'stream'),
    ];

Future<void> openPlayer(BuildContext context, {required String title, required List<Uri> sources, bool live = false, VoidCallback? onStarted, VoidCallback? heartbeat}) {
  return Navigator.of(context, rootNavigator: true).push(
    MaterialPageRoute(fullscreenDialog: true, builder: (_) => PlayerPage(title: title, sources: sources, live: live, onStarted: onStarted, heartbeat: heartbeat)),
  );
}

class PlayerPage extends StatefulWidget {
  const PlayerPage({super.key, required this.title, required this.sources, this.live = false, this.onStarted, this.heartbeat});

  final String title;
  final List<Uri> sources;
  final bool live;
  final VoidCallback? onStarted;
  final VoidCallback? heartbeat;

  @override
  State<PlayerPage> createState() => _PlayerPageState();
}

class _PlayerPageState extends State<PlayerPage> {
  VideoPlayerController? _controller;
  String? _error;
  bool _controls = true;
  Timer? _hide;
  Timer? _beat;

  @override
  void initState() {
    super.initState();
    SystemChrome.setPreferredOrientations(DeviceOrientation.values);
    _open(0);
  }

  Future<void> _open(int index) async {
    if (index >= widget.sources.length) {
      setState(() => _error = 'Không phát được video này');
      return;
    }
    final c = VideoPlayerController.networkUrl(
      widget.sources[index],
      formatHint: widget.sources[index].path.endsWith('.m3u8') ? VideoFormat.hls : null,
    );
    try {
      await c.initialize();
    } catch (_) {
      await c.dispose();
      if (mounted) await _open(index + 1);
      return;
    }
    if (!mounted) {
      await c.dispose();
      return;
    }
    c.addListener(_tick);
    setState(() => _controller = c);
    await c.play();
    widget.onStarted?.call();
    if (widget.heartbeat != null) {
      widget.heartbeat!();
      _beat = Timer.periodic(const Duration(seconds: 20), (_) {
        if (_controller?.value.isPlaying ?? false) widget.heartbeat!();
      });
    }
    _scheduleHide();
  }

  void _tick() {
    if (mounted) setState(() {});
  }

  void _scheduleHide() {
    _hide?.cancel();
    _hide = Timer(const Duration(seconds: 3), () {
      if (mounted && (_controller?.value.isPlaying ?? false)) setState(() => _controls = false);
    });
  }

  void _toggleControls() {
    setState(() => _controls = !_controls);
    if (_controls) _scheduleHide();
  }

  @override
  void dispose() {
    _hide?.cancel();
    _beat?.cancel();
    _controller?.removeListener(_tick);
    _controller?.dispose();
    SystemChrome.setPreferredOrientations([DeviceOrientation.portraitUp]);
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final c = _controller;
    final v = c?.value;
    return Scaffold(
      backgroundColor: Colors.black,
      body: SafeArea(
        child: Stack(
          children: [
            Center(
              child: _error != null
                  ? Column(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        const Icon(Icons.error_outline, color: Colors.white70, size: 48),
                        const SizedBox(height: 12),
                        Text(_error!, style: const TextStyle(color: Colors.white)),
                      ],
                    )
                  : c == null || !(v?.isInitialized ?? false)
                      ? const CircularProgressIndicator(color: Colors.white)
                      : GestureDetector(
                          onTap: _toggleControls,
                          child: AspectRatio(aspectRatio: v!.aspectRatio == 0 ? 16 / 9 : v.aspectRatio, child: VideoPlayer(c)),
                        ),
            ),
            if (v?.isBuffering ?? false) const Center(child: CircularProgressIndicator(color: Colors.white)),
            AnimatedOpacity(
              opacity: _controls || c == null ? 1 : 0,
              duration: const Duration(milliseconds: 200),
              child: IgnorePointer(
                ignoring: !_controls && c != null,
                child: Column(
                  children: [
                    Row(
                      children: [
                        IconButton(color: Colors.white, icon: const Icon(Icons.close), onPressed: () => Navigator.pop(context)),
                        Expanded(
                          child: Text(widget.title, maxLines: 1, overflow: TextOverflow.ellipsis, style: const TextStyle(color: Colors.white, fontWeight: FontWeight.w700, fontSize: 16)),
                        ),
                        if (widget.live)
                          Container(
                            margin: const EdgeInsets.only(right: 12),
                            padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
                            decoration: BoxDecoration(color: const Color(0xFFDC2626), borderRadius: BorderRadius.circular(6)),
                            child: const Text('LIVE', style: TextStyle(color: Colors.white, fontWeight: FontWeight.w800, fontSize: 12)),
                          ),
                      ],
                    ),
                    const Spacer(),
                    if (c != null && (v?.isInitialized ?? false))
                      Padding(
                        padding: const EdgeInsets.fromLTRB(8, 0, 8, 8),
                        child: Row(
                          children: [
                            IconButton(
                              color: Colors.white,
                              iconSize: 34,
                              icon: Icon(v!.isPlaying ? Icons.pause_rounded : Icons.play_arrow_rounded),
                              onPressed: () {
                                v.isPlaying ? c.pause() : c.play();
                                _scheduleHide();
                              },
                            ),
                            if (!widget.live) ...[
                              Text(duration(v.position.inSeconds), style: const TextStyle(color: Colors.white, fontSize: 12)),
                              Expanded(
                                child: Padding(
                                  padding: const EdgeInsets.symmetric(horizontal: 10),
                                  child: VideoProgressIndicator(
                                    c,
                                    allowScrubbing: true,
                                    colors: VideoProgressColors(playedColor: context.appColor, bufferedColor: Colors.white38, backgroundColor: Colors.white12),
                                  ),
                                ),
                              ),
                              Text(duration(v.duration.inSeconds), style: const TextStyle(color: Colors.white, fontSize: 12)),
                            ] else
                              const Spacer(),
                            IconButton(
                              color: Colors.white,
                              icon: Icon(v.volume == 0 ? Icons.volume_off_rounded : Icons.volume_up_rounded),
                              onPressed: () => c.setVolume(v.volume == 0 ? 1 : 0),
                            ),
                          ],
                        ),
                      ),
                  ],
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }
}
