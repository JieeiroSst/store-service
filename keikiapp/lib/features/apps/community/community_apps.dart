import 'package:flutter/material.dart';

import '../../../gen/models.g.dart';
import '../kit/kit.dart';
import '../extras/extras.dart';
import 'call_page.dart';

String _ago(String? iso) {
  final d = DateTime.tryParse(iso ?? '');
  if (d == null) return '';
  final diff = DateTime.now().difference(d);
  if (diff.inMinutes < 1) return 'vừa xong';
  if (diff.inHours < 1) return '${diff.inMinutes} phút';
  if (diff.inDays < 1) return '${diff.inHours} giờ';
  if (diff.inDays < 7) return '${diff.inDays} ngày';
  return shortDate(iso);
}

Widget _avatar(BuildContext context, String? name, String? url, {double radius = 20}) {
  final letter = (name ?? '?').isEmpty ? '?' : name!.characters.first.toUpperCase();
  return CircleAvatar(
    radius: radius,
    backgroundColor: context.appColor.withValues(alpha: 0.14),
    foregroundImage: url != null && url.startsWith('http') ? NetworkImage(url) : null,
    child: Text(letter, style: TextStyle(color: context.appColor, fontWeight: FontWeight.w800)),
  );
}

class ThreadsApp extends StatelessWidget {
  const ThreadsApp({super.key});

  @override
  Widget build(BuildContext context) {
    final api = context.api.threads;
    return ServiceApp(
      serviceId: 'threads',
      fab: Builder(
        builder: (context) => FloatingActionButton(
          backgroundColor: context.appColor,
          foregroundColor: Colors.white,
          tooltip: 'Viết bài',
          onPressed: () => showFormSheet(
            context,
            title: 'Bài viết mới',
            fields: const [
              KField('content', 'Bạn đang nghĩ gì?', kind: FieldKind.multiline, required: true),
              KField('tags', 'Thẻ (cách nhau bởi dấu phẩy)', hint: 'amthuc, dulich'),
            ],
            submitLabel: 'Đăng',
            success: 'Đã đăng bài, kéo xuống để làm mới',
            onSubmit: (v) => api.threadsCreatePost(
              content: v.str('content')!,
              tags: (v.str('tags') ?? '').split(',').map((t) => t.trim().replaceAll('#', '')).where((t) => t.isNotEmpty).toList(),
            ),
          ),
          child: const Icon(Icons.edit_rounded),
        ),
      ),
      tabs: [
        AppTab('Bảng tin', Icons.dynamic_feed_outlined, (context) => _Feed(load: () async {
              final home = await api.threadsHomeFeed(limit: 30).catchError((_) => null);
              if ((home?.rows ?? const []).isNotEmpty) return home!.rows;
              return (await api.threadsPosts(limit: 30))?.rows;
            })),
        AppTab('Xu hướng', Icons.tag, (context) => AsyncList<ThreadsTagCount>(
              load: () => api.threadsTrendingTags(limit: 30),
              emptyMessage: 'Chưa có chủ đề nổi bật',
              spacing: 6,
              itemBuilder: (context, t, _) => EntityTile(
                leading: CircleAvatar(backgroundColor: context.appColor.withValues(alpha: 0.12), child: Icon(Icons.tag, color: context.appColor)),
                title: '#${t.name_ ?? ''}',
                subtitle: '${t.count ?? 0} bài viết',
              ),
            )),
        AppTab('Đã lưu', Icons.bookmark_border, (context) => _Feed(load: () async => (await api.threadsBookmarks(limit: 30))?.rows, empty: 'Chưa lưu bài nào')),
      ],
    );
  }
}

class _Feed extends StatelessWidget {
  const _Feed({required this.load, this.empty = 'Chưa có bài viết nào'});

  final Future<List<ThreadsPost>?> Function() load;
  final String empty;

  @override
  Widget build(BuildContext context) {
    return AsyncList<ThreadsPost>(
      load: load,
      emptyMessage: empty,
      emptyIcon: Icons.forum_outlined,
      itemBuilder: (context, p, reload) => _PostCard(post: p, reload: reload),
    );
  }
}

class _PostCard extends StatefulWidget {
  const _PostCard({required this.post, required this.reload, this.detail = false});

  final ThreadsPost post;
  final Future<void> Function() reload;
  final bool detail;

  @override
  State<_PostCard> createState() => _PostCardState();
}

class _PostCardState extends State<_PostCard> {
  bool _liked = false;
  bool _saved = false;
  late int _likes = widget.post.likeCount ?? 0;

  @override
  Widget build(BuildContext context) {
    final api = context.api.threads;
    final p = widget.post.originalPost ?? widget.post;
    final repost = widget.post.originalPost != null;
    final theme = Theme.of(context);
    final mine = widget.post.userId == '${context.userId}';
    return KCard(
      onTap: widget.detail ? null : () => openPage(context, _ThreadPage(post: widget.post)),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          if (repost)
            Padding(
              padding: const EdgeInsets.only(bottom: 8),
              child: Text('${widget.post.author?.username ?? ''} đã đăng lại', style: theme.textTheme.labelMedium?.copyWith(color: theme.colorScheme.onSurfaceVariant)),
            ),
          Row(
            children: [
              GestureDetector(
                onTap: p.author?.id == null ? null : () => openPage(context, ThreadsProfilePage(author: p.author!)),
                child: _avatar(context, p.author?.username, p.author?.profileImageUrl),
              ),
              const SizedBox(width: 10),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(p.author?.username ?? 'Người dùng', style: const TextStyle(fontWeight: FontWeight.w800)),
                    Text(_ago(p.createdAt), style: theme.textTheme.bodySmall),
                  ],
                ),
              ),
              if (mine)
                IconButton(
                  tooltip: 'Xoá',
                  icon: const Icon(Icons.more_horiz),
                  onPressed: () async {
                    if (await runAction(context, () => api.threadsDeletePost(id: widget.post.id!), confirm: 'Xoá bài viết này?', destructive: true, success: 'Đã xoá')) {
                      await widget.reload();
                    }
                  },
                ),
            ],
          ),
          const SizedBox(height: 10),
          Text(p.content ?? '', style: theme.textTheme.bodyLarge?.copyWith(height: 1.45), maxLines: widget.detail ? null : 8, overflow: widget.detail ? null : TextOverflow.ellipsis),
          if ((p.tags ?? const []).isNotEmpty) ...[
            const SizedBox(height: 6),
            Wrap(spacing: 6, children: [for (final t in p.tags!) Text('#$t', style: TextStyle(color: context.appColor, fontWeight: FontWeight.w700))]),
          ],
          if ((p.mediaUrls ?? const []).isNotEmpty) ...[
            const SizedBox(height: 10),
            AspectRatio(
              aspectRatio: 16 / 10,
              child: LayoutBuilder(builder: (context, c) => KThumb(url: p.mediaUrls!.first, width: c.maxWidth, height: c.maxHeight, radius: 16)),
            ),
          ],
          const SizedBox(height: 6),
          Row(
            children: [
              _Action(
                icon: _liked ? Icons.favorite : Icons.favorite_border,
                color: _liked ? const Color(0xFFE11D48) : null,
                label: '$_likes',
                onTap: () async {
                  final like = !_liked;
                  setState(() {
                    _liked = like;
                    _likes += like ? 1 : -1;
                  });
                  final ok = await runAction(context, () => like ? api.threadsLikePost(id: p.id!) : api.threadsUnlikePost(id: p.id!));
                  if (!ok && mounted) {
                    setState(() {
                      _liked = !like;
                      _likes += like ? -1 : 1;
                    });
                  }
                },
              ),
              _Action(icon: Icons.mode_comment_outlined, label: '${p.commentCount ?? 0}', onTap: widget.detail ? null : () => openPage(context, _ThreadPage(post: widget.post))),
              _Action(
                icon: Icons.repeat_rounded,
                label: '${p.repostCount ?? 0}',
                onTap: () async {
                  if (repost && widget.post.userId == '${context.userId}') {
                    if (await runAction(context, () => api.threadsUnrepost(id: p.id!), confirm: 'Huỷ đăng lại bài này?', success: 'Đã huỷ đăng lại')) await widget.reload();
                  } else {
                    await runAction(context, () => api.threadsRepost(id: p.id!), success: 'Đã đăng lại');
                  }
                },
              ),
              const Spacer(),
              IconButton(
                tooltip: 'Lưu',
                icon: Icon(_saved ? Icons.bookmark : Icons.bookmark_border, color: _saved ? context.appColor : null),
                onPressed: () async {
                  final save = !_saved;
                  setState(() => _saved = save);
                  await runAction(context, () => save ? api.threadsBookmark(id: p.id!) : api.threadsUnbookmark(id: p.id!), success: save ? 'Đã lưu bài viết' : null);
                },
              ),
            ],
          ),
        ],
      ),
    );
  }
}

class _Action extends StatelessWidget {
  const _Action({required this.icon, required this.label, this.onTap, this.color});

  final IconData icon;
  final String label;
  final VoidCallback? onTap;
  final Color? color;

  @override
  Widget build(BuildContext context) {
    return InkWell(
      borderRadius: BorderRadius.circular(99),
      onTap: onTap,
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 8),
        child: Row(children: [Icon(icon, size: 20, color: color), const SizedBox(width: 4), Text(label)]),
      ),
    );
  }
}

class _ThreadPage extends StatefulWidget {
  const _ThreadPage({required this.post});

  final ThreadsPost post;

  @override
  State<_ThreadPage> createState() => _ThreadPageState();
}

class _ThreadPageState extends State<_ThreadPage> {
  final _input = TextEditingController();
  int _version = 0;
  bool _sending = false;

  @override
  void dispose() {
    _input.dispose();
    super.dispose();
  }

  Future<void> _send() async {
    final text = _input.text.trim();
    if (text.isEmpty) return;
    setState(() => _sending = true);
    final ok = await runAction(context, () => context.api.threads.threadsCreateComment(id: (widget.post.originalPost ?? widget.post).id!, content: text));
    if (!mounted) return;
    setState(() {
      _sending = false;
      if (ok) {
        _input.clear();
        _version++;
      }
    });
  }

  @override
  Widget build(BuildContext context) {
    final api = context.api.threads;
    final id = (widget.post.originalPost ?? widget.post).id!;
    return DetailPage(
      title: 'Bài viết',
      bottom: Row(
        children: [
          Expanded(
            child: TextField(
              controller: _input,
              minLines: 1,
              maxLines: 4,
              decoration: const InputDecoration(hintText: 'Viết bình luận…', isDense: true),
            ),
          ),
          const SizedBox(width: 8),
          IconButton.filled(onPressed: _sending ? null : _send, icon: const Icon(Icons.send_rounded)),
        ],
      ),
      children: [
        AsyncBuilder<ThreadsPost?>(
          key: ValueKey('p$_version'),
          load: () => api.threadsPost(id: (widget.post.originalPost ?? widget.post).id!).catchError((_) => null),
          loading: Padding(padding: const EdgeInsets.fromLTRB(16, 16, 16, 0), child: _PostCard(post: widget.post, reload: () async {}, detail: true)),
          builder: (context, fresh, _) => Padding(
            padding: const EdgeInsets.fromLTRB(16, 16, 16, 0),
            child: _PostCard(post: fresh ?? widget.post, reload: () async => setState(() => _version++), detail: true),
          ),
        ),
        const SectionHeader('Bình luận'),
        AsyncBuilder<ThreadsCommentPage?>(
          key: ValueKey(_version),
          load: () => api.threadsComments(id: id, limit: 50),
          loading: const Padding(padding: EdgeInsets.all(24), child: Center(child: CircularProgressIndicator())),
          builder: (context, page, reloadComments) => Column(
            children: [
              if ((page?.rows ?? const []).isEmpty) const Padding(padding: EdgeInsets.all(24), child: Text('Hãy là người bình luận đầu tiên')),
              for (final c in page?.rows ?? const <ThreadsComment>[])
                ListTile(
                  onLongPress: () => threadsCommentMenu(context, c, reloadComments),
                  leading: _avatar(context, c.author?.username, c.author?.profileImageUrl, radius: 18),
                  title: Text(c.author?.username ?? 'Người dùng', style: const TextStyle(fontWeight: FontWeight.w700, fontSize: 14)),
                  subtitle: Text(c.content ?? ''),
                  trailing: Text(_ago(c.createdAt), style: Theme.of(context).textTheme.bodySmall),
                ),
            ],
          ),
        ),
      ],
    );
  }
}

class PostApp extends StatelessWidget {
  const PostApp({super.key});

  @override
  Widget build(BuildContext context) {
    final api = context.api.post;
    return ServiceApp(
      serviceId: 'post',
      adminActions: [
        Builder(builder: (context) => IconButton(tooltip: 'Quản lý chuyên mục', icon: const Icon(Icons.folder_special_outlined), onPressed: () => openPage(context, const PostCategoryManager()))),
      ],
      fab: Builder(
        builder: (context) => FloatingActionButton.extended(
          backgroundColor: context.appColor,
          foregroundColor: Colors.white,
          icon: const Icon(Icons.edit_note),
          label: const Text('Viết bài'),
          onPressed: () async {
            final cats = (await api.postCategories(limit: 50).catchError((_) => null))?.rows ?? const <PostCategory>[];
            if (!context.mounted) return;
            await showFormSheet(
              context,
              title: 'Bài viết mới',
              fields: [
                const KField('name', 'Tiêu đề', required: true),
                const KField('description', 'Tóm tắt'),
                const KField('content', 'Nội dung', kind: FieldKind.multiline, required: true),
                if (cats.isNotEmpty) KField('category', 'Chuyên mục', kind: FieldKind.choice, options: [for (final c in cats) (c.id ?? '', c.name_ ?? '')]),
                const KField('file', 'Ảnh bìa', kind: FieldKind.file, icon: Icons.image_outlined),
              ],
              submitLabel: 'Đăng bài',
              success: 'Đã đăng bài viết',
              onSubmit: (v) => api.postCreatePost(name_: v.str('name')!, content: v.str('content')!, description: v.str('description'), categoryId: v.str('category'), file: v.file('file')),
            );
          },
        ),
      ),
      tabs: [
        AppTab('Mới nhất', Icons.article_outlined, (context) => AsyncList<PostPost>(
              load: () async => (await api.posts(limit: 40))?.rows,
              emptyMessage: 'Chưa có bài viết',
              emptyIcon: Icons.article_outlined,
              itemBuilder: (context, p, _) => _ArticleTile(post: p),
            )),
        AppTab('Chuyên mục', Icons.category_outlined, (context) => AsyncList<PostCategory>(
              load: () async => (await api.postCategories(limit: 50))?.rows,
              emptyMessage: 'Chưa có chuyên mục',
              itemBuilder: (context, c, _) => EntityTile(
                title: c.name_ ?? '',
                subtitle: c.description,
                icon: Icons.folder_open_rounded,
                caption: '${(c.posts ?? const []).length} bài viết',
                onTap: () => openPage(
                  context,
                  DetailPage(
                    title: c.name_ ?? 'Chuyên mục',
                    children: [
                      for (final p in c.posts ?? const <PostPost>[]) Padding(padding: const EdgeInsets.fromLTRB(16, 12, 16, 0), child: _ArticleTile(post: p)),
                      if ((c.posts ?? const []).isEmpty) const Padding(padding: EdgeInsets.all(24), child: Text('Chuyên mục chưa có bài viết')),
                    ],
                  ),
                ),
              ),
            )),
      ],
    );
  }
}

class _ArticleTile extends StatelessWidget {
  const _ArticleTile({required this.post});

  final PostPost post;

  @override
  Widget build(BuildContext context) {
    final p = post;
    return EntityTile(
      title: p.name_ ?? 'Bài viết',
      subtitle: p.description ?? p.content,
      imageUrl: p.media?.url,
      icon: Icons.article_rounded,
      caption: shortDate(p.createdAt),
      badges: [for (final c in (p.categories ?? const <PostCategory>[]).take(2)) Pill(c.name_ ?? '')],
      onTap: () => openPage(
        context,
        AsyncBuilder<PostPost?>(
          load: () => context.api.post.post(id: p.id!),
          loading: const Scaffold(body: SkeletonList()),
          builder: (context, full, reload) => DetailPage(
            title: full?.name_ ?? p.name_ ?? '',
            actions: [
              if (full?.authorId == '${context.userId}')
                IconButton(
                  tooltip: 'Sửa bài',
                  icon: const Icon(Icons.edit_outlined),
                  onPressed: () async {
                    if (await editPost(context, full!)) await reload();
                  },
                ),
            ],
            header: HeroHeader(title: full?.name_ ?? '', subtitle: full?.description, imageUrl: full?.media?.url, icon: Icons.article_rounded, badges: [Pill(formatDate(full?.createdAt))]),
            children: [Paragraph(full?.content)],
          ),
        ),
      ),
    );
  }
}

class RoomApp extends StatelessWidget {
  const RoomApp({super.key});

  @override
  Widget build(BuildContext context) {
    final api = context.api.room;
    return ServiceApp(
      serviceId: 'room',
      fab: Builder(
        builder: (context) => FloatingActionButton(
          backgroundColor: context.appColor,
          foregroundColor: Colors.white,
          tooltip: 'Tạo phòng',
          onPressed: () => showFormSheet(
            context,
            title: 'Tạo phòng chat',
            fields: const [KField('name', 'Tên phòng', required: true)],
            success: 'Đã tạo phòng, kéo xuống để làm mới',
            onSubmit: (v) => api.roomCreateRoom(name_: v.str('name')!),
          ),
          child: const Icon(Icons.add_comment_rounded),
        ),
      ),
      tabs: [
        AppTab('Phòng', Icons.forum_outlined, (context) => AsyncList<RoomRoom>(
              load: () => api.rooms(),
              header: AsyncBuilder<RoomUser?>(
                load: () => api.roomMe().catchError((_) => null),
                loading: const SizedBox.shrink(),
                builder: (context, me, _) => me == null
                    ? const SizedBox.shrink()
                    : Padding(
                        padding: const EdgeInsets.fromLTRB(20, 14, 20, 0),
                        child: Text('Bạn đang trò chuyện với tên @${me.username ?? ''}', style: Theme.of(context).textTheme.bodySmall),
                      ),
              ),
              emptyMessage: 'Bạn chưa tham gia phòng nào',
              emptyIcon: Icons.forum_outlined,
              itemBuilder: (context, r, _) => EntityTile(
                leading: _avatar(context, r.name_, null, radius: 26),
                title: r.name_ ?? 'Phòng',
                subtitle: '${r.memberCount ?? 0} thành viên • chủ phòng ${r.owner ?? ''}',
                onTap: () => openPage(context, _RoomPage(room: r)),
              ),
            )),
      ],
    );
  }
}

class _RoomPage extends StatelessWidget {
  const _RoomPage({required this.room});

  final RoomRoom room;

  @override
  Widget build(BuildContext context) {
    final api = context.api.room;
    final id = room.id!;
    return AsyncBuilder<(List<RoomMessage>, List<RoomMember>)>(
      load: () async {
        final r = await Future.wait<Object?>([api.roomMessages(id: id, limit: 50), api.roomMembers(id: id)]);
        return ((r[0] as List<RoomMessage>?) ?? const <RoomMessage>[], (r[1] as List<RoomMember>?) ?? const <RoomMember>[]);
      },
      loading: const Scaffold(body: SkeletonList()),
      builder: (context, d, reload) => DetailPage(
        title: room.name_ ?? 'Phòng',
        actions: [
          IconButton(
            tooltip: 'Thêm thành viên',
            icon: const Icon(Icons.person_add_alt_1),
            onPressed: () async {
              if (await showFormSheet(
                context,
                title: 'Thêm thành viên',
                fields: const [KField('username', 'Tên đăng nhập', required: true)],
                success: 'Đã thêm thành viên',
                onSubmit: (v) => api.roomAddMember(id: '$id', username: v.str('username')!),
              )) {
                await reload();
              }
            },
          ),
        ],
        children: [
          SizedBox(
            height: 88,
            child: ListView(
              scrollDirection: Axis.horizontal,
              padding: const EdgeInsets.fromLTRB(16, 12, 16, 0),
              children: [
                for (final m in d.$2)
                  Padding(
                    padding: const EdgeInsets.only(right: 14),
                    child: GestureDetector(
                      onLongPress: () async {
                        if (await runAction(context, () => api.roomRemoveMember(id: '$id', username: m.username!), confirm: 'Xoá ${m.username} khỏi phòng?', destructive: true, success: 'Đã xoá')) {
                          await reload();
                        }
                      },
                      child: Column(
                        children: [
                          _avatar(context, m.username, null, radius: 22),
                          const SizedBox(height: 4),
                          Text(m.username ?? '', style: const TextStyle(fontSize: 11)),
                        ],
                      ),
                    ),
                  ),
              ],
            ),
          ),
          const Divider(),
          if (d.$1.isEmpty) const Padding(padding: EdgeInsets.all(32), child: Center(child: Text('Chưa có tin nhắn'))),
          for (final m in d.$1.reversed)
            Align(
              alignment: m.username == context.userName ? Alignment.centerRight : Alignment.centerLeft,
              child: Container(
                margin: const EdgeInsets.fromLTRB(16, 4, 16, 4),
                padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 10),
                constraints: const BoxConstraints(maxWidth: 300),
                decoration: BoxDecoration(
                  color: context.appColor.withValues(alpha: 0.1),
                  borderRadius: BorderRadius.circular(18),
                ),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(m.username ?? '', style: TextStyle(fontWeight: FontWeight.w700, fontSize: 12, color: context.appColor)),
                    Text(m.content ?? ''),
                    Text(_ago(m.createdAt), style: Theme.of(context).textTheme.labelSmall),
                  ],
                ),
              ),
            ),
        ],
      ),
    );
  }
}

class ShareApp extends StatefulWidget {
  const ShareApp({super.key});

  @override
  State<ShareApp> createState() => _ShareAppState();
}

class _ShareAppState extends State<ShareApp> {
  final _input = TextEditingController();
  final _history = <(String, String?)>[];
  bool _busy = false;

  @override
  void dispose() {
    _input.dispose();
    super.dispose();
  }

  Future<void> _ask() async {
    final q = _input.text.trim();
    if (q.isEmpty) return;
    setState(() {
      _busy = true;
      _history.insert(0, (q, null));
      _input.clear();
    });
    try {
      final a = await context.api.share.shareAsk(content: q);
      setState(() => _history[0] = (q, a?.result ?? 'Không có câu trả lời'));
    } catch (e) {
      setState(() => _history[0] = (q, errorMessage(e)));
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return ServiceApp(
      serviceId: 'share',
      tabs: [
        AppTab('Hỏi đáp', Icons.question_answer_outlined, (context) => Column(
              children: [
                Expanded(
                  child: _history.isEmpty
                      ? const EmptyView(message: 'Đặt câu hỏi để nhận câu trả lời từ AI', icon: Icons.lightbulb_outline)
                      : ListView(
                          padding: const EdgeInsets.all(16),
                          children: [
                            for (final (q, a) in _history) ...[
                              Align(
                                alignment: Alignment.centerRight,
                                child: Container(
                                  padding: const EdgeInsets.all(12),
                                  constraints: const BoxConstraints(maxWidth: 300),
                                  decoration: BoxDecoration(color: context.appColor, borderRadius: BorderRadius.circular(18)),
                                  child: Text(q, style: const TextStyle(color: Colors.white)),
                                ),
                              ),
                              const SizedBox(height: 8),
                              KCard(child: a == null ? const LinearProgressIndicator() : SelectableText(a, style: const TextStyle(height: 1.5))),
                              const SizedBox(height: 18),
                            ],
                          ],
                        ),
                ),
                SafeArea(
                  top: false,
                  child: Padding(
                    padding: const EdgeInsets.fromLTRB(16, 8, 16, 12),
                    child: Row(
                      children: [
                        Expanded(
                          child: TextField(
                            controller: _input,
                            minLines: 1,
                            maxLines: 4,
                            onSubmitted: (_) => _ask(),
                            decoration: const InputDecoration(hintText: 'Hỏi bất cứ điều gì…', isDense: true),
                          ),
                        ),
                        const SizedBox(width: 8),
                        IconButton.filled(onPressed: _busy ? null : _ask, icon: const Icon(Icons.send_rounded)),
                      ],
                    ),
                  ),
                ),
              ],
            )),
      ],
    );
  }
}

class PhotoApp extends StatelessWidget {
  const PhotoApp({super.key});

  @override
  Widget build(BuildContext context) {
    final api = context.api.photo;
    return ServiceApp(
      serviceId: 'photo',
      tabs: [
        AppTab('Ghép ảnh', Icons.photo_library_outlined, (context) => ListView(
              padding: const EdgeInsets.only(bottom: 24),
              children: [
                const AppBanner(icon: Icons.dashboard_customize_rounded, title: 'Ghép ảnh thành collage', subtitle: 'Chọn tối đa 4 ảnh, hệ thống tự căn bố cục và xuất ảnh chất lượng cao'),
                Padding(
                  padding: const EdgeInsets.all(16),
                  child: FilledButton.icon(
                    icon: const Icon(Icons.add_photo_alternate_outlined),
                    label: const Text('Tạo ảnh ghép'),
                    onPressed: () async {
                      PhotoComposition? created;
                      await showFormSheet(
                        context,
                        title: 'Ảnh ghép mới',
                        fields: const [
                          KField('layout', 'Bố cục', kind: FieldKind.choice, options: [
                            ('{"type":"grid"}', 'Lưới'),
                            ('{"type":"grid","rows":1}', 'Hàng ngang'),
                            ('{"type":"grid","cols":1}', 'Hàng dọc'),
                            ('{"type":"mosaic"}', 'Mosaic'),
                          ]),
                          KField('a', 'Ảnh 1', kind: FieldKind.file, required: true, icon: Icons.image_outlined),
                          KField('b', 'Ảnh 2', kind: FieldKind.file, required: true, icon: Icons.image_outlined),
                          KField('c', 'Ảnh 3', kind: FieldKind.file, icon: Icons.image_outlined),
                          KField('d', 'Ảnh 4', kind: FieldKind.file, icon: Icons.image_outlined),
                        ],
                        submitLabel: 'Ghép ảnh',
                        onSubmit: (v) async {
                          created = await api.photoCreateComposition(images: [for (final k in ['a', 'b', 'c', 'd']) ?v.file(k)], layout: v.str('layout'));
                        },
                      );
                      if (created == null || !context.mounted) return;
                      await showResult(context, title: 'Đang xử lý ảnh ghép', message: 'Mã: ${created!.id}\nTrạng thái: ${created!.status ?? ''}');
                    },
                  ),
                ),
                const SectionHeader('Ảnh ghép gần đây'),
                const _RecentCompositions(),
              ],
            )),
      ],
    );
  }
}

class _RecentCompositions extends StatelessWidget {
  const _RecentCompositions();

  @override
  Widget build(BuildContext context) {
    final api = context.api.photo;
    return AsyncBuilder<List<PhotoComposition>>(
      load: () async => await api.myPhotoCompositions(limit: 24) ?? const <PhotoComposition>[],
      loading: const Padding(padding: EdgeInsets.all(24), child: Center(child: CircularProgressIndicator())),
      builder: (context, items, _) => items.isEmpty
          ? const Padding(padding: EdgeInsets.all(24), child: Center(child: Text('Chưa có ảnh ghép nào')))
          : Padding(
              padding: const EdgeInsets.symmetric(horizontal: 16),
              child: Wrap(
                spacing: 10,
                runSpacing: 10,
                children: [
                  for (final c in items)
                    SizedBox(
                      width: 160,
                      height: 200,
                      child: PosterCard(
                        title: '${c.width ?? '-'}×${c.height ?? '-'}',
                        subtitle: c.status,
                        imageUrl: c.url,
                        icon: Icons.dashboard_rounded,
                        onTap: () async {
                          final fresh = await api.photoComposition(id: c.id!).catchError((_) => null);
                          if (!context.mounted || fresh == null) return;
                          await openPage(
                            context,
                            DetailPage(
                              title: 'Ảnh ghép',
                              header: HeroHeader(title: '${fresh.width ?? '-'}×${fresh.height ?? '-'} • ${fresh.format ?? ''}', imageUrl: fresh.url, icon: Icons.dashboard_rounded, badges: [StatusPill(fresh.status)]),
                              children: [InfoSection(rows: [('Nguồn ảnh', '${(fresh.sources ?? const []).length}'), ('Lỗi', fresh.errorMessage), ('Tạo lúc', formatDate(fresh.createdAt))])],
                            ),
                          );
                        },
                      ),
                    ),
                ],
              ),
            ),
    );
  }
}

class ChatReportApp extends StatelessWidget {
  const ChatReportApp({super.key});

  @override
  Widget build(BuildContext context) {
    final api = context.api.chat;
    return ServiceApp(
      serviceId: 'chat',
      fab: Builder(
        builder: (context) => FloatingActionButton.extended(
          backgroundColor: context.appColor,
          foregroundColor: Colors.white,
          icon: const Icon(Icons.flag_outlined),
          label: const Text('Báo cáo'),
          onPressed: () => showFormSheet(
            context,
            title: 'Báo cáo vi phạm',
            fields: const [
              KField('type', 'Loại vi phạm', kind: FieldKind.choice, options: [('spam', 'Tin rác'), ('harassment', 'Quấy rối'), ('scam', 'Lừa đảo'), ('inappropriate', 'Nội dung không phù hợp'), ('other', 'Khác')]),
              KField('notes', 'Mô tả', kind: FieldKind.multiline, required: true),
            ],
            success: 'Cảm ơn bạn, chúng tôi sẽ xem xét sớm',
            onSubmit: (v) => api.createChatReport(input: ChatReportInput(reportType: v.str('type'), notes: v.str('notes'))),
          ),
        ),
      ),
      tabs: [
        AppTab('Báo cáo của tôi', Icons.flag_outlined, (context) => AsyncList<ChatReport>(
              load: () => api.myChatReports(),
              header: const AppBanner(icon: Icons.health_and_safety_rounded, title: 'An toàn khi trò chuyện', subtitle: 'Báo cáo tin nhắn rác, lừa đảo hoặc quấy rối để giữ cộng đồng lành mạnh'),
              emptyMessage: 'Bạn chưa gửi báo cáo nào',
              emptyIcon: Icons.verified_user_outlined,
              itemBuilder: (context, r, _) => EntityTile(
                title: switch (r.reportType) { 'spam' => 'Tin rác', 'harassment' => 'Quấy rối', 'scam' => 'Lừa đảo', 'inappropriate' => 'Nội dung không phù hợp', _ => r.reportType ?? 'Khác' },
                subtitle: r.notes,
                caption: formatDate(r.createdAt),
                icon: Icons.flag_rounded,
                badges: [StatusPill(r.status)],
              ),
            )),
      ],
    );
  }
}

class WebrtcApp extends StatelessWidget {
  const WebrtcApp({super.key});

  Future<void> _join(BuildContext context, String room) => Navigator.of(context, rootNavigator: true).push(
        MaterialPageRoute(builder: (_) => CallPage(roomId: room)),
      );

  @override
  Widget build(BuildContext context) {
    final api = context.api.webrtc;
    return ServiceApp(
      serviceId: 'webrtc',
      tabs: [
        AppTab('Gọi video', Icons.video_call_outlined, (context) => AsyncBuilder<WebrtcIceServers?>(
              load: () => api.webrtcIceServers(userId: '${context.userId}'),
              builder: (context, ice, _) => ListView(
                padding: const EdgeInsets.only(bottom: 24),
                children: [
                  AppBanner(
                    icon: Icons.video_call_rounded,
                    title: 'Phòng họp video',
                    subtitle: 'Kết nối ngang hàng, mã hoá đầu-cuối qua ${(ice?.iceServers ?? const []).length} máy chủ TURN/STUN',
                    children: [
                      Row(
                        children: [
                          Expanded(
                            child: FilledButton.icon(
                              style: FilledButton.styleFrom(backgroundColor: Colors.white, foregroundColor: context.appColor),
                              icon: const Icon(Icons.add_call),
                              label: const Text('Phòng mới'),
                              onPressed: () => _join(context, newIdempotencyKey().replaceAll('-', '').substring(0, 8)),
                            ),
                          ),
                          const SizedBox(width: 8),
                          Expanded(
                            child: FilledButton.tonalIcon(
                              icon: const Icon(Icons.meeting_room_outlined),
                              label: const Text('Vào phòng'),
                              onPressed: () async {
                                WebrtcRoomInfo? room;
                                String? code;
                                await showFormSheet(
                                  context,
                                  title: 'Mã phòng',
                                  fields: const [KField('room', 'Mã phòng', required: true)],
                                  submitLabel: 'Vào phòng',
                                  onSubmit: (v) async {
                                    code = v.str('room');
                                    room = await api.webrtcRoom(roomId: code!);
                                    final count = (room?.participants ?? const []).length;
                                    if (room?.capacity != null && count >= room!.capacity!) throw Exception('Phòng đã đủ ${room!.capacity} người');
                                  },
                                );
                                if (code == null || !context.mounted) return;
                                await _join(context, code!);
                              },
                            ),
                          ),
                        ],
                      ),
                    ],
                  ),
                  InfoSection(title: 'Máy chủ kết nối', rows: [
                    for (final (i, s) in (ice?.iceServers ?? const <WebrtcIceServer>[]).indexed) ('Máy chủ ${i + 1}', (s.urls ?? const []).join('\n')),
                  ]),
                ],
              ),
            )),
      ],
    );
  }
}

class RecruitmentApp extends StatelessWidget {
  const RecruitmentApp({super.key});

  @override
  Widget build(BuildContext context) {
    final api = context.api.recruitment;
    return ServiceApp(
      serviceId: 'recruitment',
      tabs: [
        AppTab('Việc làm', Icons.work_outline, (context) => ListView(
              padding: const EdgeInsets.only(bottom: 24),
              children: [
                const AppBanner(icon: Icons.work_rounded, title: 'Ứng tuyển nhanh', subtitle: 'Nộp hồ sơ bằng mã công việc, theo dõi vòng phỏng vấn và offer'),
                Padding(
                  padding: const EdgeInsets.fromLTRB(16, 16, 16, 0),
                  child: FilledButton.icon(
                    icon: const Icon(Icons.send_outlined),
                    label: const Text('Ứng tuyển'),
                    onPressed: () async {
                      final me = '${context.userId}';
                      RecruitApplication? app;
                      await showFormSheet(
                        context,
                        title: 'Ứng tuyển',
                        fields: const [
                          KField('job', 'Mã công việc', required: true),
                          KField('partner', 'Mã người giới thiệu'),
                          KField('cover', 'Thư giới thiệu', kind: FieldKind.multiline),
                        ],
                        submitLabel: 'Nộp hồ sơ',
                        onSubmit: (v) async => app = await api.recruitApply(
                          jobId: v.str('job')!,
                          candidateId: me,
                          referredByPartnerId: v.str('partner'),
                          coverLetter: v.str('cover'),
                        ),
                      );
                      if (app == null || !context.mounted) return;
                      await showResult(context, title: 'Đã nộp hồ sơ', message: 'Trạng thái: ${app!.status ?? ''}\nMã hồ sơ: ${app!.id ?? ''}');
                    },
                  ),
                ),
                const SectionHeader('Cộng tác viên tuyển dụng'),
                Padding(
                  padding: const EdgeInsets.symmetric(horizontal: 16),
                  child: KCard(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        const Text('Giới thiệu ứng viên, nhận hoa hồng khi họ được tuyển.', style: TextStyle(fontWeight: FontWeight.w600)),
                        const SizedBox(height: 10),
                        OutlinedButton.icon(
                          icon: const Icon(Icons.handshake_outlined),
                          label: const Text('Đăng ký cộng tác viên'),
                          onPressed: () async {
                            RecruitPartner? p;
                            await showFormSheet(
                              context,
                              title: 'Đăng ký cộng tác viên',
                              fields: [
                                KField('name', 'Họ tên', required: true, initial: context.userName),
                                const KField('email', 'Email', kind: FieldKind.email, required: true),
                                const KField('phone', 'Số điện thoại', kind: FieldKind.phone),
                                const KField('company', 'Công ty'),
                                const KField('ref', 'Mã giới thiệu'),
                              ],
                              submitLabel: 'Đăng ký',
                              onSubmit: (v) async => p = await api.recruitRegisterPartner(
                                input: RecruitPartnerInput(fullName: v.str('name'), email: v.str('email'), phone: v.str('phone'), company: v.str('company'), referredByToken: v.str('ref')),
                              ),
                            );
                            if (p == null || !context.mounted) return;
                            await showResult(context, title: 'Chào mừng cộng tác viên!', message: 'Hạng ${p!.tier ?? '-'} • Trạng thái ${p!.status ?? ''}');
                          },
                        ),
                      ],
                    ),
                  ),
                ),
              ],
            )),
      ],
    );
  }
}

class AdManagementApp extends StatefulWidget {
  const AdManagementApp({super.key});

  @override
  State<AdManagementApp> createState() => _AdManagementAppState();
}

class _AdManagementAppState extends State<AdManagementApp> {
  int _position = 1;
  final _session = newIdempotencyKey();

  @override
  Widget build(BuildContext context) {
    final api = context.api.admanagement;
    return ServiceApp(
      serviceId: 'admanagement',
      tabs: [
        AppTab('Quảng cáo', Icons.campaign_outlined, (context) => Column(
              children: [
                SizedBox(
                  height: 56,
                  child: ListView(
                    scrollDirection: Axis.horizontal,
                    padding: const EdgeInsets.fromLTRB(16, 12, 16, 0),
                    children: [
                      for (final (i, label) in const [(1, 'Trang chủ'), (2, 'Banner trên'), (3, 'Giữa trang'), (4, 'Chân trang')])
                        Padding(
                          padding: const EdgeInsets.only(right: 8),
                          child: ChoiceChip(label: Text(label), selected: _position == i, onSelected: (_) => setState(() => _position = i)),
                        ),
                    ],
                  ),
                ),
                Expanded(
                  child: AsyncBuilder<AdServedAd?>(
                    key: ValueKey(_position),
                    load: () => api.adServe(positionId: _position, sessionId: _session, country: 'VN', device: 'mobile'),
                    builder: (context, served, reload) {
                      final ad = served?.ad;
                      if (ad == null) return EmptyView(message: 'Chưa có quảng cáo cho vị trí này', icon: Icons.campaign_outlined, onRefresh: reload);
                      return ListView(
                        padding: const EdgeInsets.only(bottom: 24),
                        children: [
                          HeroHeader(title: ad.title ?? 'Quảng cáo', subtitle: ad.description, imageUrl: ad.contentUrl, icon: Icons.campaign_rounded, badges: [if (ad.adType != null) Pill(ad.adType!)]),
                          ActionButtons([
                            FilledButton.icon(
                              icon: const Icon(Icons.open_in_new),
                              label: const Text('Xem ưu đãi'),
                              onPressed: () async {
                                AdClickResult? r;
                                final ok = await runAction(context, () async => r = await api.adTrackClick(adId: ad.id!, impressionId: served?.impressionId, sessionId: _session));
                                if (ok && context.mounted) showSnack(context, 'Đích đến: ${r?.targetUrl ?? r?.location ?? ad.targetUrl ?? ''}');
                              },
                            ),
                            OutlinedButton.icon(icon: const Icon(Icons.refresh), label: const Text('Quảng cáo khác'), onPressed: reload),
                          ]),
                          InfoSection(rows: [('Chiến dịch', '${ad.campaignId ?? ''}'), ('Hiệu lực', '${shortDate(ad.startDate)} – ${shortDate(ad.endDate)}'), ('Đích đến', ad.targetUrl)]),
                        ],
                      );
                    },
                  ),
                ),
              ],
            )),
      ],
    );
  }
}

class ManageApp extends StatelessWidget {
  const ManageApp({super.key});

  @override
  Widget build(BuildContext context) {
    final api = context.api.manage;
    Future<void> call(BuildContext context, String title, List<KField> fields, Future<Object?> Function(FormData v) run) async {
      Object? out;
      await showFormSheet(context, title: title, fields: fields, onSubmit: (v) async => out = await run(v));
      if (out == null || !context.mounted) return;
      final token = pick(out, 'access_token');
      await showResult(
        context,
        title: title,
        message: token == null ? '$out' : 'Đã nhận access token (${'$token'.length} ký tự), hết hạn sau ${pick(out, 'expires_in') ?? '-'} giây',
      );
    }

    return ServiceApp(
      serviceId: 'manage',
      tabs: [
        AppTab('Đăng nhập SSO', Icons.key_outlined, (context) => ListView(
              padding: const EdgeInsets.only(bottom: 24),
              children: [
                const AppBanner(icon: Icons.admin_panel_settings_rounded, title: 'Tài khoản doanh nghiệp', subtitle: 'Đăng nhập một lần (Keycloak) cho cổng quản trị và đối tác'),
                for (final (icon, title, fields, run) in <(IconData, String, List<KField>, Future<Object?> Function(FormData))>[
                  (Icons.login, 'Đăng nhập', const [KField('username', 'Tên đăng nhập', required: true), KField('password', 'Mật khẩu', kind: FieldKind.password, required: true)], (v) => api.kcPostLogin(input: {'username': v.str('username'), 'password': v.str('password')})),
                  (Icons.sms_outlined, 'Đăng nhập bằng OTP', const [KField('username', 'Tên đăng nhập', required: true), KField('otp', 'Mã OTP', kind: FieldKind.number, required: true)], (v) => api.kcPostLoginOtp(input: {'username': v.str('username'), 'otp': v.str('otp')})),
                  (Icons.refresh, 'Làm mới phiên', const [KField('refresh_token', 'Refresh token', kind: FieldKind.multiline, required: true)], (v) => api.kcPostRefreshToken(input: {'refresh_token': v.str('refresh_token')})),
                  (Icons.vpn_key_outlined, 'Đăng nhập ứng dụng (client)', const [KField('client_id', 'Client ID', required: true), KField('client_secret', 'Client secret', kind: FieldKind.password, required: true)], (v) => api.kcPostLoginClient(input: {'client_id': v.str('client_id'), 'client_secret': v.str('client_secret')})),
                  (Icons.logout, 'Đăng xuất SSO', const [KField('refresh_token', 'Refresh token', kind: FieldKind.multiline, required: true)], (v) => api.kcPostLogout(input: {'refresh_token': v.str('refresh_token')})),
                ])
                  Padding(
                    padding: const EdgeInsets.fromLTRB(16, 12, 16, 0),
                    child: EntityTile(title: title, icon: icon, onTap: () => call(context, title, fields, run)),
                  ),
              ],
            )),
      ],
    );
  }
}
