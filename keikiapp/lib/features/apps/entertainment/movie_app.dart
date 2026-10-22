import 'package:flutter/material.dart';

import '../../../gen/models.g.dart';
import '../kit/kit.dart';
import '../kit/player.dart';

const _videoFields = '{ id title description tags status duration views created_at }';
const _pageSel = '{ items { video $_videoFields score reason because $_videoFields progress at } total page page_size snapshot }';
const _homeSel = '{ sections { id title items { video $_videoFields score reason because $_videoFields progress at } } }';

class MovieApp extends StatelessWidget {
  const MovieApp({super.key});

  @override
  Widget build(BuildContext context) {
    return ServiceApp(
      serviceId: 'movie',
      tabs: [
        AppTab('Khám phá', Icons.explore_outlined, (_) => const _Discover()),
        AppTab('Đang xem', Icons.play_circle_outline, (context) => _RecList(
              load: () async => (await context.api.movie.myMovieContinueWatching(pageSize: 30, selectionSet: _pageSel))?.items,
              empty: 'Bạn chưa xem dở phim nào',
            )),
        AppTab('Lịch sử', Icons.history, (context) => _RecList(
              load: () async => (await context.api.movie.myMovieHistory(pageSize: 50, selectionSet: _pageSel))?.items,
              empty: 'Chưa có lịch sử xem',
              removable: true,
            )),
      ],
    );
  }
}

class _Discover extends StatelessWidget {
  const _Discover();

  @override
  Widget build(BuildContext context) {
    final api = context.api.movie;
    return AsyncBuilder<List<(String, List<MovieRecommendation>)>>(
      load: () async {
        final results = await Future.wait([
          api.movieTrending(pageSize: 12, selectionSet: _pageSel).then((p) => p?.items ?? const <MovieRecommendation>[]),
          api.movieNewReleases(pageSize: 12, selectionSet: _pageSel).then((p) => p?.items ?? const <MovieRecommendation>[]),
          api.myMovieForUser(pageSize: 12, selectionSet: _pageSel).then((p) => p?.items ?? const <MovieRecommendation>[]).catchError((_) => <MovieRecommendation>[]),
        ]);
        final home = await api.myMovieHome(selectionSet: _homeSel).catchError((_) => null);
        return [
          ('Dành cho bạn', results[2]),
          ('Đang thịnh hành', results[0]),
          ('Mới ra mắt', results[1]),
          for (final s in home?.sections ?? const <MovieSection>[]) (s.title ?? '', s.items ?? const <MovieRecommendation>[]),
        ].where((s) => s.$2.isNotEmpty && s.$1.isNotEmpty).toList();
      },
      builder: (context, sections, reload) => RefreshIndicator(
        onRefresh: reload,
        child: ListView(
          padding: const EdgeInsets.only(bottom: 24),
          children: [
            const AppBanner(
              icon: Icons.movie_filter_rounded,
              title: 'Tối nay xem gì?',
              subtitle: 'Gợi ý được cá nhân hoá theo những gì bạn đã xem',
            ),
            if (sections.isEmpty) const SizedBox(height: 320, child: EmptyView(message: 'Chưa có phim nào', icon: Icons.movie_outlined)),
            for (final (title, items) in sections) ...[
              SectionHeader(title),
              Strip<MovieRecommendation>(
                items: items,
                height: 236,
                itemBuilder: (context, r) => PosterCard(
                  title: r.video?.title ?? 'Phim',
                  imageUrl: _poster(context, r.video),
                  subtitle: [if (r.video?.duration != null) duration(r.video!.duration), if (r.video?.views != null) '${compact(r.video!.views)} lượt xem'].join(' • '),
                  icon: Icons.movie_rounded,
                  badge: r.reason == null || r.reason!.isEmpty ? null : _reason(r.reason!),
                  onTap: r.video == null ? null : () => openPage(context, MovieDetailPage(video: r.video!)),
                ),
              ),
            ],
          ],
        ),
      ),
    );
  }
}

String? _poster(BuildContext context, MovieVideo? v) => v?.id == null ? null : videoServiceUri(context, v!.id!, 'thumbnail').toString();

String _reason(String r) => switch (r) {
      'trending' => 'Hot',
      'new' || 'new_release' => 'Mới',
      'similar' || 'because_you_watched' => 'Tương tự',
      'continue' || 'continue_watching' => 'Xem tiếp',
      _ => r.length > 12 ? r.substring(0, 12) : r,
    };

class _RecList extends StatelessWidget {
  const _RecList({required this.load, required this.empty, this.removable = false});

  final Future<List<MovieRecommendation>?> Function() load;
  final String empty;
  final bool removable;

  @override
  Widget build(BuildContext context) {
    return AsyncList<MovieRecommendation>(
      load: load,
      emptyMessage: empty,
      emptyIcon: Icons.movie_outlined,
      itemBuilder: (context, r, reload) {
        final v = r.video;
        return EntityTile(
          title: v?.title ?? 'Phim',
          subtitle: v?.description,
          icon: Icons.movie_rounded,
          imageUrl: _poster(context, v),
          caption: r.progress == null ? null : 'Đã xem ${(r.progress! * (r.progress! <= 1 ? 100 : 1)).round()}%',
          onTap: v == null ? null : () => openPage(context, MovieDetailPage(video: v)),
          trailing: removable && v?.id != null
              ? IconButton(
                  tooltip: 'Xoá khỏi lịch sử',
                  icon: const Icon(Icons.delete_outline),
                  onPressed: () async {
                    if (await runAction(context, () => context.api.movie.movieRemoveFromMyHistory(videoId: v!.id!), success: 'Đã xoá khỏi lịch sử')) {
                      await reload();
                    }
                  },
                )
              : null,
        );
      },
    );
  }
}

class MovieDetailPage extends StatelessWidget {
  const MovieDetailPage({super.key, required this.video});

  final MovieVideo video;

  @override
  Widget build(BuildContext context) {
    final api = context.api.movie;
    final id = video.id ?? '';
    return DetailPage(
      title: video.title ?? 'Phim',
      header: HeroHeader(
        title: video.title ?? 'Phim',
        icon: Icons.movie_rounded,
        imageUrl: _poster(context, video),
        subtitle: [if (video.duration != null) duration(video.duration), if (video.views != null) '${compact(video.views)} lượt xem', shortDate(video.createdAt)]
            .where((s) => s.isNotEmpty)
            .join(' • '),
        badges: [for (final t in video.tags ?? const <String>[]) Pill(t)],
      ),
      bottom: Row(
        children: [
          IconButton.filledTonal(
            tooltip: 'Thích',
            onPressed: () => runAction(context, () => api.movieRecordEvent(videoId: id, type_: 'like'), success: 'Đã thêm vào phim yêu thích'),
            icon: const Icon(Icons.favorite_border),
          ),
          const SizedBox(width: 12),
          Expanded(
            child: FilledButton.icon(
              onPressed: id.isEmpty
                  ? null
                  : () => openPlayer(
                        context,
                        title: video.title ?? 'Phim',
                        sources: videoSources(context, id),
                        onStarted: () => api.movieRecordEvent(videoId: id, type_: 'play', value: 0).catchError((_) => null),
                      ),
              icon: const Icon(Icons.play_arrow_rounded),
              label: const Text('Xem ngay'),
            ),
          ),
        ],
      ),
      children: [
        Paragraph(video.description, title: 'Nội dung'),
        const SectionHeader('Phim tương tự'),
        SizedBox(
          height: 236,
          child: AsyncBuilder<List<MovieRecommendation>>(
            load: () async => (await api.movieSimilar(id: id, pageSize: 12, selectionSet: _pageSel))?.items ?? const [],
            loading: const Center(child: CircularProgressIndicator()),
            builder: (context, items, _) => items.isEmpty
                ? const Center(child: Text('Chưa có gợi ý'))
                : Strip<MovieRecommendation>(
                    items: items,
                    height: 236,
                    itemBuilder: (context, r) => PosterCard(
                      title: r.video?.title ?? 'Phim',
                      imageUrl: _poster(context, r.video),
                      subtitle: r.video?.duration == null ? null : duration(r.video!.duration),
                      icon: Icons.movie_rounded,
                      onTap: r.video == null ? null : () => openPage(context, MovieDetailPage(video: r.video!)),
                    ),
                  ),
          ),
        ),
      ],
    );
  }
}
