import 'package:flutter/material.dart';

import '../../../gen/models.g.dart';
import '../extras/extras.dart';
import '../kit/kit.dart';
import '../kit/player.dart';

class VideoApp extends StatelessWidget {
  const VideoApp({super.key});

  @override
  Widget build(BuildContext context) {
    final api = context.api.video;
    Widget list(String sort) => SearchableList<VideoVideo>(
          hint: 'Tìm video',
          load: (q, _) async => (await api.videos(q: q.isEmpty ? null : q, sort: sort, pageSize: 40))?.items,
          gridExtent: 220,
          gridAspect: 0.95,
          emptyMessage: 'Chưa có video',
          emptyIcon: Icons.video_library_outlined,
          itemBuilder: (context, v, _) => _VideoCard(video: v),
        );
    return ServiceApp(
      serviceId: 'video',
      tabs: [
        AppTab('Mới nhất', Icons.fiber_new_outlined, (_) => list('newest')),
        AppTab('Xem nhiều', Icons.trending_up, (_) => list('views')),
      ],
    );
  }
}

class _VideoCard extends StatelessWidget {
  const _VideoCard({required this.video});

  final VideoVideo video;

  @override
  Widget build(BuildContext context) {
    return PosterCard(
      title: video.title ?? 'Video',
      imageUrl: video.id == null ? null : videoServiceUri(context, video.id!, 'thumbnail').toString(),
      subtitle: '${compact(video.views)} lượt xem • ${shortDate(video.createdAt)}',
      icon: Icons.play_circle_fill_rounded,
      badge: video.duration == null ? null : duration(video.duration),
      onTap: () => openPage(context, VideoDetailPage(id: video.id!)),
    );
  }
}

class VideoDetailPage extends StatefulWidget {
  const VideoDetailPage({super.key, required this.id});

  final String id;

  @override
  State<VideoDetailPage> createState() => _VideoDetailPageState();
}

class _VideoDetailPageState extends State<VideoDetailPage> {
  @override
  Widget build(BuildContext context) {
    final api = context.api.video;
    return AsyncBuilder<VideoVideo?>(
      load: () => api.video(id: widget.id),
      loading: const Scaffold(body: SkeletonList()),
      builder: (context, v, _) => DetailPage(
        title: v?.title ?? 'Video',
        actions: [ShareLinkButton(path: '/video/${widget.id}', title: v?.title ?? 'Video')],
        bottom: FilledButton.icon(
          icon: const Icon(Icons.play_arrow_rounded),
          label: const Text('Phát video'),
          onPressed: () => openPlayer(
            context,
            title: v?.title ?? 'Video',
            sources: videoSources(context, widget.id),
            onStarted: () => api.videoView(id: widget.id).catchError((_) => null),
          ),
        ),
        header: HeroHeader(
          title: v?.title ?? 'Video',
          imageUrl: videoServiceUri(context, widget.id, 'thumbnail').toString(),
          icon: Icons.play_circle_fill_rounded,
          subtitle: '${compact(v?.views)} lượt xem • ${formatDate(v?.createdAt)}',
          badges: [
            if (v?.duration != null) Pill(duration(v!.duration), icon: Icons.timer_outlined),
            if (v?.width != null) Pill('${v!.width}×${v.height}', icon: Icons.hd_outlined),
            StatusPill(v?.status),
          ],
        ),
        children: [
          Paragraph(v?.description, title: 'Mô tả'),
          const SectionHeader('Video liên quan'),
          AsyncBuilder<List<VideoVideo>?>(
            load: () => api.videoRelated(id: widget.id, limit: 10),
            loading: const Padding(padding: EdgeInsets.all(24), child: Center(child: CircularProgressIndicator())),
            builder: (context, items, _) => Strip<VideoVideo>(
              items: items ?? const [],
              height: 200,
              itemWidth: 180,
              itemBuilder: (context, r) => _VideoCard(video: r),
            ),
          ),
        ],
      ),
    );
  }
}

class MediaApp extends StatelessWidget {
  const MediaApp({super.key});

  @override
  Widget build(BuildContext context) {
    return ServiceApp(
      serviceId: 'media',
      tabs: [
        AppTab('Thư viện', Icons.video_collection_outlined, (context) => SearchableList<MediaVideo>(
              hint: 'Tìm theo mô tả hoặc thẻ',
              load: (q, _) async => (await context.api.media.mediaSearchVideos(query: q.isEmpty ? null : q, size: 40))?.videos,
              gridExtent: 220,
              gridAspect: 0.9,
              emptyMessage: 'Chưa có video',
              emptyIcon: Icons.video_collection_outlined,
              itemBuilder: (context, v, _) => PosterCard(
                title: v.description ?? 'Video #${v.videoId}',
                subtitle: [v.tag?.value, shortDate(v.uploadedOn)].whereType<String>().where((s) => s.isNotEmpty).join(' • '),
                imageUrl: v.thumbnailUrl,
                icon: Icons.play_circle_fill_rounded,
                onTap: () => openPage(
                  context,
                  DetailPage(
                    title: 'Video #${v.videoId}',
                    header: HeroHeader(title: v.description ?? 'Video', imageUrl: v.thumbnailUrl, icon: Icons.play_circle_fill_rounded, badges: [if (v.tag?.value != null) Pill('#${v.tag!.value}')]),
                    bottom: Builder(
                      builder: (context) => FilledButton.icon(
                        icon: const Icon(Icons.play_arrow_rounded),
                        label: const Text('Phát video'),
                        onPressed: (Uri.tryParse(v.streamUrl ?? '')?.hasScheme ?? false)
                            ? () => openPlayer(context, title: v.description ?? 'Video', sources: [Uri.parse(v.streamUrl!)])
                            : null,
                      ),
                    ),
                    children: [InfoSection(rows: [('Mã video', '${v.videoId}'), ('Tải lên', formatDate(v.uploadedOn)), ('Luồng phát', v.streamUrl)])],
                  ),
                ),
              ),
            )),
      ],
    );
  }
}

class LivestreamApp extends StatelessWidget {
  const LivestreamApp({super.key});

  @override
  Widget build(BuildContext context) {
    final api = context.api.livestream;
    Widget rooms(bool? live) => AsyncList<LiveRoom>(
          load: () => api.liveRooms(live: live),
          gridExtent: 220,
          gridAspect: 0.9,
          emptyMessage: live == true ? 'Chưa có ai đang phát trực tiếp' : 'Chưa có phòng live',
          emptyIcon: Icons.live_tv_outlined,
          itemBuilder: (context, r, _) => PosterCard(
            title: r.title ?? 'Phòng live',
            subtitle: r.description,
            icon: Icons.live_tv_rounded,
            badge: r.status == 'live' ? 'LIVE' : null,
            onTap: () => openPage(context, LiveRoomPage(id: r.id!)),
          ),
        );
    return ServiceApp(
      serviceId: 'livestream',
      fab: Builder(
        builder: (context) => FloatingActionButton.extended(
          backgroundColor: context.appColor,
          foregroundColor: Colors.white,
          icon: const Icon(Icons.videocam_rounded),
          label: const Text('Tạo phòng'),
          onPressed: () => showFormSheet(
            context,
            title: 'Tạo phòng live',
            fields: const [KField('title', 'Tiêu đề', required: true), KField('description', 'Mô tả', kind: FieldKind.multiline)],
            success: 'Đã tạo phòng. Xem stream key trong chi tiết phòng',
            onSubmit: (v) => api.liveCreateRoom(input: LiveCreateRoomInput(title: v.str('title'), description: v.str('description'))),
          ),
        ),
      ),
      tabs: [
        AppTab('Đang live', Icons.sensors, (_) => rooms(true)),
        AppTab('Tất cả phòng', Icons.grid_view_rounded, (_) => rooms(null)),
      ],
    );
  }
}

class LiveRoomPage extends StatelessWidget {
  const LiveRoomPage({super.key, required this.id});

  final String id;

  @override
  Widget build(BuildContext context) {
    final api = context.api.livestream;
    return AsyncBuilder<(LiveRoom?, LiveViewerCount?, LivePlaybackInfo?)>(
      load: () async {
        final r = await Future.wait<Object?>([
          api.liveRoom(id: id),
          api.liveViewerCount(id: id).catchError((_) => null),
          api.livePlayback(id: id).catchError((_) => null),
        ]);
        return (r[0] as LiveRoom?, r[1] as LiveViewerCount?, r[2] as LivePlaybackInfo?);
      },
      loading: const Scaffold(body: SkeletonList()),
      builder: (context, d, reload) {
        final room = d.$1;
        final mine = room?.ownerUserId == '${context.userId}';
        final playback = d.$3;
        final session = newIdempotencyKey();
        return DetailPage(
          title: room?.title ?? 'Phòng live',
          actions: [ShareLinkButton(path: '/live/$id', title: room?.title ?? 'Livestream', description: room?.description)],
          bottom: FilledButton.icon(
            icon: const Icon(Icons.live_tv_rounded),
            label: Text(playback?.isLive == true ? 'Xem trực tiếp' : 'Chưa phát sóng'),
            onPressed: playback?.isLive == true && (playback?.url ?? '').isNotEmpty
                ? () => openPlayer(
                      context,
                      title: room?.title ?? 'Live',
                      sources: [Uri.parse(playback!.url!)],
                      live: true,
                      heartbeat: () {
                        api.liveViewerHeartbeat(id: id, sessionId: session).catchError((_) => false);
                        reportLiveQuality(context, id);
                      },
                    )
                : null,
          ),
          header: AppBanner(
            icon: d.$3?.isLive == true ? Icons.sensors : Icons.live_tv_rounded,
            title: room?.title ?? 'Phòng live',
            subtitle: room?.description,
            children: [
              StatRow([
                StatTile(label: 'Đang xem', value: '${d.$2?.viewers ?? 0}', icon: Icons.visibility_outlined, onDark: true),
                StatTile(label: 'Trạng thái', value: d.$3?.isLive == true ? 'LIVE' : 'Offline', icon: Icons.circle, onDark: true),
              ]),
            ],
          ),
          children: [
            InfoSection(rows: [
              ('Đường phát', d.$3?.url),
              ('Hết hạn', d.$3?.expiresAt == null ? null : formatDate(d.$3!.expiresAt)),
              ('Tạo lúc', formatDate(room?.createdAt)),
              if (mine) ('Stream key', room?.streamKey),
            ]),
            if (mine) LiveOwnerTools(roomId: id),
            if (mine)
              ActionButtons([
                OutlinedButton.icon(
                  icon: const Icon(Icons.key_outlined),
                  label: const Text('Đổi stream key'),
                  onPressed: () async {
                    if (await runAction(context, () => api.liveRegenerateStreamKey(id: id), success: 'Đã tạo stream key mới')) await reload();
                  },
                ),
                OutlinedButton.icon(
                  icon: const Icon(Icons.stop_circle_outlined),
                  label: const Text('Kết thúc live'),
                  onPressed: () => runAction(context, () => api.liveEndStream(id: id), confirm: 'Kết thúc buổi phát?', success: 'Đã kết thúc'),
                ),
              ]),
            const SectionHeader('Bản ghi'),
            AsyncBuilder<List<LiveRecording>?>(
              load: () => api.liveRecordings(id: id),
              loading: const Padding(padding: EdgeInsets.all(24), child: Center(child: CircularProgressIndicator())),
              builder: (context, recs, _) => Column(
                children: [
                  if ((recs ?? const []).isEmpty) const Padding(padding: EdgeInsets.all(24), child: Text('Chưa có bản ghi')),
                  for (final r in recs ?? const <LiveRecording>[])
                    Padding(
                      padding: const EdgeInsets.fromLTRB(16, 0, 16, 8),
                      child: EntityTile(title: formatDate(r.createdAt), subtitle: '${duration(r.durationSeconds)} • ${compact((r.sizeBytes ?? 0) / 1048576)} MB', icon: Icons.video_file_outlined),
                    ),
                ],
              ),
            ),
          ],
        );
      },
    );
  }
}
