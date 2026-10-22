import 'package:flutter/material.dart';
import 'package:flutter/services.dart';

import '../../../gen/models.g.dart';
import '../extras/extras.dart';
import '../kit/kit.dart';

String _size(num? bytes) {
  final b = (bytes ?? 0).toDouble();
  if (b >= 1 << 30) return '${(b / (1 << 30)).toStringAsFixed(1)} GB';
  if (b >= 1 << 20) return '${(b / (1 << 20)).toStringAsFixed(1)} MB';
  if (b >= 1 << 10) return '${(b / (1 << 10)).toStringAsFixed(0)} KB';
  return '${b.toStringAsFixed(0)} B';
}

IconData _fileIcon(String? type) {
  final t = type ?? '';
  if (t.startsWith('image/')) return Icons.image_rounded;
  if (t.startsWith('video/')) return Icons.movie_rounded;
  if (t.startsWith('audio/')) return Icons.audiotrack_rounded;
  if (t.contains('pdf')) return Icons.picture_as_pdf_rounded;
  return Icons.insert_drive_file_rounded;
}

class _ChatBubble extends StatelessWidget {
  const _ChatBubble({required this.text, required this.mine, this.meta});

  final String text;
  final bool mine;
  final String? meta;

  @override
  Widget build(BuildContext context) {
    final color = context.appColor;
    return Align(
      alignment: mine ? Alignment.centerRight : Alignment.centerLeft,
      child: Container(
        margin: const EdgeInsets.symmetric(vertical: 4),
        padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 10),
        constraints: const BoxConstraints(maxWidth: 310),
        decoration: BoxDecoration(
          color: mine ? color : Theme.of(context).colorScheme.surfaceContainerHighest,
          borderRadius: BorderRadius.only(
            topLeft: const Radius.circular(18),
            topRight: const Radius.circular(18),
            bottomLeft: Radius.circular(mine ? 18 : 4),
            bottomRight: Radius.circular(mine ? 4 : 18),
          ),
        ),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            SelectableText(text, style: TextStyle(color: mine ? Colors.white : null, height: 1.4)),
            if (meta != null) Text(meta!, style: TextStyle(fontSize: 11, color: mine ? Colors.white70 : Theme.of(context).colorScheme.onSurfaceVariant)),
          ],
        ),
      ),
    );
  }
}

class ChatbotApp extends StatelessWidget {
  const ChatbotApp({super.key});

  @override
  Widget build(BuildContext context) {
    final api = context.api.chatbot;
    return ServiceApp(
      serviceId: 'chatbot',
      tabs: [
        AppTab('Hội thoại', Icons.smart_toy_outlined, (context) => AsyncList<ChatbotConversation>(
              load: () async => (await api.chatbotConversations())?.conversations,
              emptyMessage: 'Chưa có hội thoại nào',
              emptyIcon: Icons.smart_toy_outlined,
              itemBuilder: (context, c, _) => EntityTile(
                title: c.isAiChat == true ? 'Trợ lý ${c.activeAiModel ?? 'AI'}' : 'Hội thoại #${c.id}',
                subtitle: 'Cập nhật ${formatDate(c.updatedAt ?? c.createdAt)}',
                icon: c.isAiChat == true ? Icons.smart_toy_rounded : Icons.forum_rounded,
                onTap: () => openPage(
                  context,
                  AsyncBuilder<ChatbotHistory?>(
                    load: () => api.chatbotConversationHistory(conversationId: c.id!, limit: 100),
                    loading: const Scaffold(body: SkeletonList()),
                    builder: (context, h, _) => DetailPage(
                      title: c.isAiChat == true ? 'Trợ lý AI' : 'Hội thoại #${c.id}',
                      children: [
                        Padding(
                          padding: const EdgeInsets.all(16),
                          child: Column(
                            children: [
                              if ((h?.messages ?? const []).isEmpty) const Padding(padding: EdgeInsets.all(24), child: Text('Chưa có tin nhắn')),
                              for (final m in h?.messages ?? const <ChatbotMessage>[])
                                _ChatBubble(text: m.content ?? '', mine: m.senderId == context.userId, meta: formatDate(m.createdAt)),
                            ],
                          ),
                        ),
                      ],
                    ),
                  ),
                ),
              ),
            )),
      ],
    );
  }
}

class OllamaPyApp extends StatefulWidget {
  const OllamaPyApp({super.key});

  @override
  State<OllamaPyApp> createState() => _OllamaPyAppState();
}

class _OllamaPyAppState extends State<OllamaPyApp> {
  final _input = TextEditingController();
  final _messages = <OllamaPyMessageInput>[];
  String? _model;
  bool _busy = false;

  @override
  void dispose() {
    _input.dispose();
    super.dispose();
  }

  Future<void> _send() async {
    final text = _input.text.trim();
    if (text.isEmpty || _model == null) return;
    setState(() {
      _messages.add(OllamaPyMessageInput(role: 'user', content: text));
      _input.clear();
      _busy = true;
    });
    try {
      final r = await context.api.ollamaPy.ollamaPyChat(input: OllamaPyChatInput(model: _model, messages: List.of(_messages)));
      setState(() => _messages.add(OllamaPyMessageInput(role: 'assistant', content: r?.message?.content ?? '')));
    } catch (e) {
      if (mounted) showSnack(context, errorMessage(e));
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return ServiceApp(
      serviceId: 'ollama_py',
      actions: [
        Builder(
          builder: (context) => IconButton(
            tooltip: 'Sinh một lần',
            icon: const Icon(Icons.bolt),
            onPressed: () async {
              OllamaPyGenerate? r;
              await showFormSheet(
                context,
                title: 'Sinh văn bản một lần',
                fields: [KField('model', 'Mô hình', required: true, initial: _model), const KField('prompt', 'Yêu cầu', kind: FieldKind.multiline, required: true)],
                submitLabel: 'Sinh',
                onSubmit: (v) async => r = await context.api.ollamaPy.ollamaPyGenerate(input: OllamaPyGenerateInput(model: v.str('model'), prompt: v.str('prompt'))),
              );
              if (r != null && context.mounted) await openPage(context, DetailPage(title: r!.model ?? 'Kết quả', children: [Paragraph(r!.response)]));
            },
          ),
        ),
      ],
      tabs: [
        AppTab('Trò chuyện', Icons.psychology_outlined, (context) => AsyncBuilder<OllamaPyTags?>(
              load: () => context.api.ollamaPy.ollamaPyTags(),
              builder: (context, tags, _) {
                final models = tags?.models ?? const <OllamaPyTagsModel>[];
                _model ??= models.isEmpty ? null : models.first.name_;
                return Column(
                  children: [
                    if (models.isNotEmpty)
                      SizedBox(
                        height: 54,
                        child: ListView(
                          scrollDirection: Axis.horizontal,
                          padding: const EdgeInsets.fromLTRB(16, 10, 16, 0),
                          children: [
                            for (final m in models)
                              Padding(
                                padding: const EdgeInsets.only(right: 8),
                                child: GestureDetector(
                                  onLongPress: () async {
                                    final info = await context.api.ollamaPy.ollamaPyShow(input: OllamaPyShowInput(model: m.name_)).catchError((_) => null);
                                    if (context.mounted) await openPage(context, JsonPage(title: m.name_ ?? '', data: info));
                                  },
                                  child: ChoiceChip(label: Text(m.name_ ?? ''), selected: _model == m.name_, onSelected: (_) => setState(() => _model = m.name_)),
                                ),
                              ),
                          ],
                        ),
                      ),
                    Expanded(
                      child: _messages.isEmpty
                          ? EmptyView(message: models.isEmpty ? 'Máy chủ chưa có mô hình nào' : 'Bắt đầu trò chuyện với ${_model ?? 'AI'}', icon: Icons.psychology_outlined)
                          : ListView(
                              padding: const EdgeInsets.all(16),
                              children: [
                                for (final m in _messages) _ChatBubble(text: m.content ?? '', mine: m.role == 'user'),
                                if (_busy) const Padding(padding: EdgeInsets.all(8), child: LinearProgressIndicator()),
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
                                decoration: const InputDecoration(hintText: 'Nhập tin nhắn…', isDense: true),
                              ),
                            ),
                            const SizedBox(width: 8),
                            IconButton.filled(onPressed: _busy || _model == null ? null : _send, icon: const Icon(Icons.send_rounded)),
                          ],
                        ),
                      ),
                    ),
                  ],
                );
              },
            )),
      ],
    );
  }
}

class DrawImageApp extends StatelessWidget {
  const DrawImageApp({super.key});

  @override
  Widget build(BuildContext context) {
    final api = context.api.drawImage;
    return ServiceApp(
      serviceId: 'draw_image',
      tabs: [
        AppTab('Ghép ảnh', Icons.grid_view_outlined, (context) => ListView(
              padding: const EdgeInsets.only(bottom: 24),
              children: [
                const AppBanner(icon: Icons.auto_awesome_mosaic_rounded, title: 'Tạo collage', subtitle: 'Chọn ảnh, số cột và ảnh trung tâm (tuỳ chọn) để tạo một bức ảnh ghép'),
                Padding(
                  padding: const EdgeInsets.fromLTRB(16, 16, 16, 0),
                  child: OutlinedButton.icon(
                    icon: const Icon(Icons.search),
                    label: const Text('Mở collage theo mã'),
                    onPressed: () async {
                      DrawImageCollage? c;
                      await showFormSheet(context, title: 'Mã collage', fields: const [KField('id', 'Mã', required: true)], submitLabel: 'Mở', onSubmit: (v) async => c = await api.drawImageCollage(id: v.str('id')!));
                      if (c == null || !context.mounted) return;
                      await openPage(context, DetailPage(title: 'Collage ${c!.id}', header: HeroHeader(title: '${c!.imageCount ?? 0} ảnh', imageUrl: c!.url, icon: Icons.auto_awesome_mosaic_rounded), children: [InfoSection(rows: [('Kích thước', '${c!.width}×${c!.height}'), ('Tạo lúc', formatDate(c!.createdAt))])]));
                    },
                  ),
                ),
                Padding(
                  padding: const EdgeInsets.all(16),
                  child: FilledButton.icon(
                    icon: const Icon(Icons.add_photo_alternate_outlined),
                    label: const Text('Tạo collage mới'),
                    onPressed: () async {
                      DrawImageCollage? c;
                      await showFormSheet(
                        context,
                        title: 'Collage mới',
                        fields: const [
                          KField('columns', 'Số cột', kind: FieldKind.choice, options: [('2', '2 cột'), ('3', '3 cột'), ('4', '4 cột')]),
                          KField('a', 'Ảnh 1', kind: FieldKind.file, required: true, icon: Icons.image_outlined),
                          KField('b', 'Ảnh 2', kind: FieldKind.file, required: true, icon: Icons.image_outlined),
                          KField('c', 'Ảnh 3', kind: FieldKind.file, icon: Icons.image_outlined),
                          KField('d', 'Ảnh 4', kind: FieldKind.file, icon: Icons.image_outlined),
                          KField('center', 'Ảnh trung tâm', kind: FieldKind.file, icon: Icons.center_focus_strong_outlined),
                        ],
                        submitLabel: 'Tạo',
                        onSubmit: (v) async => c = await api.drawImageCreateCollage(
                          files: [for (final k in ['a', 'b', 'c', 'd']) ?v.file(k)],
                          center: v.file('center'),
                          columns: v.integer('columns'),
                          layout: 'grid',
                        ),
                      );
                      if (c == null || !context.mounted) return;
                      await openPage(
                        context,
                        DetailPage(
                          title: 'Collage ${c!.id}',
                          header: HeroHeader(title: '${c!.imageCount ?? 0} ảnh • ${c!.columns ?? '-'} cột', imageUrl: c!.url, icon: Icons.auto_awesome_mosaic_rounded),
                          children: [InfoSection(rows: [('Kích thước', '${c!.width}×${c!.height}'), ('Dung lượng', _size(c!.size)), ('Đường dẫn', c!.url)])],
                        ),
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

class EkycApp extends StatelessWidget {
  const EkycApp({super.key});

  @override
  Widget build(BuildContext context) {
    final api = context.api.ekyc;
    return ServiceApp(
      serviceId: 'ekyc',
      tabs: [
        AppTab('Xác minh danh tính', Icons.badge_outlined, (context) => AsyncBuilder<EkycStatus?>(
              load: () => api.myEkycStatus().catchError((_) => null),
              builder: (context, s, reload) {
                final id = s?.identity;
                final face = s?.face;
                final verified = (s?.verification?.status ?? '').toLowerCase().contains('verif') || (s?.verification?.status ?? '').toLowerCase() == 'approved';
                Widget step(int n, String title, String subtitle, bool done, VoidCallback? onTap) => Padding(
                      padding: const EdgeInsets.fromLTRB(16, 0, 16, 10),
                      child: EntityTile(
                        leading: CircleAvatar(
                          backgroundColor: done ? const Color(0xFF16A34A) : context.appColor.withValues(alpha: 0.12),
                          child: done ? const Icon(Icons.check, color: Colors.white) : Text('$n', style: TextStyle(color: context.appColor, fontWeight: FontWeight.w800)),
                        ),
                        title: title,
                        subtitle: subtitle,
                        onTap: onTap,
                      ),
                    );
                return RefreshIndicator(
                  onRefresh: reload,
                  child: ListView(
                    padding: const EdgeInsets.only(bottom: 24),
                    children: [
                      AppBanner(
                        icon: verified ? Icons.verified_rounded : Icons.badge_rounded,
                        title: verified ? 'Đã xác minh danh tính' : 'Xác minh danh tính (eKYC)',
                        subtitle: verified
                            ? 'Độ khớp khuôn mặt ${((s?.verification?.matchScore ?? 0) * ((s?.verification?.matchScore ?? 0) <= 1 ? 100 : 1)).round()}%'
                            : 'Hoàn tất 3 bước để tăng hạn mức Ví Keiki và mở khoá các dịch vụ tài chính',
                      ),
                      const SizedBox(height: 12),
                      step(
                        1,
                        'Chụp CCCD hai mặt',
                        id == null ? 'Ảnh rõ nét, không loá sáng' : '${id.surname ?? ''} ${id.givenNames ?? ''} • số ${id.documentNumber ?? ''}'.trim(),
                        id != null,
                        () async {
                          if (await showFormSheet(
                            context,
                            title: 'Căn cước công dân',
                            fields: const [
                              KField('front', 'Mặt trước', kind: FieldKind.file, required: true, icon: Icons.credit_card),
                              KField('back', 'Mặt sau', kind: FieldKind.file, required: true, icon: Icons.credit_card),
                            ],
                            submitLabel: 'Gửi',
                            success: 'Đã đọc thông tin CCCD',
                            onSubmit: (v) => api.ekycSubmitCitizenCard(front: v.file('front')!, back: v.file('back')!),
                          )) {
                            await reload();
                          }
                        },
                      ),
                      step(
                        2,
                        'Quét khuôn mặt',
                        face == null ? 'Chọn 1–3 ảnh chân dung, nhìn thẳng' : 'Chất lượng ${((face.qualityScore ?? 0) * ((face.qualityScore ?? 0) <= 1 ? 100 : 1)).round()}%',
                        face != null,
                        () async {
                          if (await showFormSheet(
                            context,
                            title: 'Ảnh khuôn mặt',
                            fields: const [
                              KField('f1', 'Ảnh 1', kind: FieldKind.file, required: true, icon: Icons.face),
                              KField('f2', 'Ảnh 2', kind: FieldKind.file, icon: Icons.face),
                              KField('f3', 'Ảnh 3', kind: FieldKind.file, icon: Icons.face),
                            ],
                            submitLabel: 'Gửi',
                            success: 'Đã ghi nhận khuôn mặt',
                            onSubmit: (v) => api.ekycSubmitFaceScan(frames: [for (final k in ['f1', 'f2', 'f3']) ?v.file(k)]),
                          )) {
                            await reload();
                          }
                        },
                      ),
                      step(
                        0,
                        'Đọc chip NFC (tuỳ chọn)',
                        id?.nfcVerified == true ? 'Đã xác thực chip' : 'Tải các tệp EF.COM, DG1, DG2, EF.SOD đọc từ chip CCCD',
                        id?.nfcVerified == true,
                        () async {
                          if (await showFormSheet(
                            context,
                            title: 'Dữ liệu chip CCCD',
                            fields: const [
                              KField('com', 'EF.COM', kind: FieldKind.file, required: true),
                              KField('dg1', 'DG1 (thông tin)', kind: FieldKind.file, required: true),
                              KField('dg2', 'DG2 (ảnh)', kind: FieldKind.file),
                              KField('sod', 'EF.SOD (chữ ký)', kind: FieldKind.file, required: true),
                            ],
                            submitLabel: 'Gửi',
                            success: 'Đã xác thực chip',
                            onSubmit: (v) => api.ekycSubmitNfcChip(efCom: v.file('com'), dg1: v.file('dg1'), dg2: v.file('dg2'), efSod: v.file('sod')),
                          )) {
                            await reload();
                          }
                        },
                      ),
                      step(3, 'Đối chiếu & xác minh', verified ? 'Hoàn tất ${formatDate(s?.verification?.verifiedAt)}' : 'So khớp ảnh CCCD với khuôn mặt', verified, id == null || face == null
                          ? null
                          : () async {
                              if (await runAction(context, () => api.ekycVerify(), success: 'Đã gửi xác minh')) await reload();
                            }),
                      if (id != null)
                        InfoSection(title: 'Thông tin trên giấy tờ', rows: [
                          ('Họ tên', '${id.surname ?? ''} ${id.givenNames ?? ''}'.trim()),
                          ('Số giấy tờ', id.documentNumber),
                          ('Ngày sinh', id.dateOfBirth),
                          ('Giới tính', id.sex),
                          ('Quốc tịch', id.nationality),
                          ('Hết hạn', id.dateOfExpiry),
                          ('Chip NFC', id.nfcVerified == true ? 'Đã xác thực' : 'Chưa'),
                        ]),
                    ],
                  ),
                );
              },
            )),
      ],
    );
  }
}

class FaceApp extends StatelessWidget {
  const FaceApp({super.key});

  @override
  Widget build(BuildContext context) {
    final api = context.api.face;
    return ServiceApp(
      serviceId: 'face',
      tabs: [
        AppTab('Nhận diện', Icons.face_retouching_natural, (context) => AsyncBuilder<FacePositions?>(
              load: () => api.facePositions().catchError((_) => null),
              builder: (context, p, reload) => RefreshIndicator(
                onRefresh: reload,
                child: ListView(
                  padding: const EdgeInsets.only(bottom: 24),
                  children: [
                    AppBanner(
                      icon: Icons.face_retouching_natural,
                      title: '${(p?.faces ?? const []).length} khuôn mặt',
                      subtitle: p?.updatedAt == null ? 'Tải ảnh lên để nhận diện' : 'Cập nhật ${formatDate(p!.updatedAt)}',
                      children: [
                        SizedBox(
                          width: double.infinity,
                          child: FilledButton.icon(
                            style: FilledButton.styleFrom(backgroundColor: Colors.white, foregroundColor: context.appColor),
                            icon: const Icon(Icons.upload),
                            label: const Text('Tải ảnh chân dung'),
                            onPressed: () async {
                              if (await showFormSheet(
                                context,
                                title: 'Ảnh chân dung',
                                fields: const [KField('avatar', 'Ảnh', kind: FieldKind.file, required: true, icon: Icons.face)],
                                submitLabel: 'Tải lên',
                                success: 'Đã tải ảnh lên',
                                onSubmit: (v) => api.faceUploadFile(avatar: v.file('avatar')!),
                              )) {
                                await reload();
                              }
                            },
                          ),
                        ),
                        const SizedBox(height: 8),
                        SizedBox(
                          width: double.infinity,
                          child: FilledButton.tonalIcon(
                            icon: const Icon(Icons.crop_free),
                            label: const Text('Gửi vị trí khuôn mặt'),
                            onPressed: () async {
                              await faceReportFrom(context, p);
                              await reload();
                            },
                          ),
                        ),
                      ],
                    ),
                    InfoSection(title: 'Vị trí khuôn mặt', rows: [
                      for (final (i, f) in (p?.faces ?? const <FaceRect>[]).indexed) ('Mặt ${i + 1}', 'x ${f.x} • y ${f.y} • ${f.width}×${f.height}'),
                    ]),
                  ],
                ),
              ),
            )),
      ],
    );
  }
}

class GymApp extends StatelessWidget {
  const GymApp({super.key});

  @override
  Widget build(BuildContext context) {
    final api = context.api.identified;
    return ServiceApp(
      serviceId: 'identified',
      tabs: [
        AppTab('Lớp học', Icons.fitness_center, (context) => AsyncList<GymClass>(
              load: () => api.gymClasses(),
              header: AppBanner(
                icon: Icons.fitness_center_rounded,
                title: 'Thẻ tập Keiki Gym',
                subtitle: 'Mua thẻ, kiểm tra hiệu lực và xem lịch lớp',
                children: [
                  Row(
                    children: [
                      Expanded(
                        child: FilledButton(
                          style: FilledButton.styleFrom(backgroundColor: Colors.white, foregroundColor: context.appColor),
                          onPressed: () async {
                            GymPass? pass;
                            await showFormSheet(
                              context,
                              title: 'Mua thẻ tập',
                              fields: const [
                                KField('validity', 'Thời hạn', kind: FieldKind.choice, options: [('1m', '1 tháng'), ('3m', '3 tháng'), ('6m', '6 tháng'), ('12m', '12 tháng')]),
                              ],
                              submitLabel: 'Mua thẻ',
                              onSubmit: (v) async => pass = await api.gymCreatePass(validity: v.str('validity')!),
                            );
                            if (pass == null || !context.mounted) return;
                            await showResult(context, title: 'Thẻ của bạn', message: 'Mã thẻ: ${pass!.id}\n${pass!.isActive == true ? 'Đang hiệu lực' : 'Chưa kích hoạt'}', icon: Icons.card_membership_rounded);
                          },
                          child: const Text('Mua thẻ'),
                        ),
                      ),
                      const SizedBox(width: 8),
                      Expanded(
                        child: FilledButton.tonal(
                          onPressed: () async {
                            GymPass? pass;
                            await showFormSheet(
                              context,
                              title: 'Kiểm tra thẻ',
                              fields: const [KField('id', 'Mã thẻ', required: true)],
                              submitLabel: 'Kiểm tra',
                              onSubmit: (v) async => pass = await api.gymPassVerification(gymPassId: v.str('id')!),
                            );
                            if (pass == null || !context.mounted) return;
                            await showResult(
                              context,
                              icon: pass!.isActive == true ? Icons.verified_rounded : Icons.block,
                              title: pass!.isActive == true ? 'Thẻ hợp lệ' : 'Thẻ hết hiệu lực',
                              message: 'Mã thẻ ${pass!.id}',
                            );
                          },
                          child: const Text('Kiểm tra thẻ'),
                        ),
                      ),
                    ],
                  ),
                ],
              ),
              emptyMessage: 'Chưa có lớp học',
              itemBuilder: (context, c, _) => EntityTile(
                onTap: () async {
                  final full = await api.gymClass(classId: c.id!).catchError((_) => null);
                  if (context.mounted && full != null) await showResult(context, icon: Icons.sports_gymnastics, title: full.name_ ?? '', message: '${full.day ?? ''} • ${full.time ?? ''}\nHLV ${full.coach ?? '-'}\n${full.description ?? ''}');
                },
                title: c.name_ ?? 'Lớp',
                subtitle: c.description,
                caption: '${c.day ?? ''} • ${c.time ?? ''}',
                icon: Icons.sports_gymnastics_rounded,
                badges: [if (c.coach != null) Pill('HLV ${c.coach}', icon: Icons.person_outline)],
              ),
            )),
      ],
    );
  }
}

String _docTitle(SearchDocument d) {
  final s = d.source ?? const {};
  for (final k in ['title', 'name', 'headline', 'question', 'subject']) {
    if (s[k] is String && (s[k] as String).isNotEmpty) return s[k] as String;
  }
  return d.id ?? '';
}

String? _docText(SearchDocument d) {
  final s = d.source ?? const {};
  for (final k in ['description', 'content', 'summary', 'body', 'text']) {
    if (s[k] is String && (s[k] as String).isNotEmpty) return s[k] as String;
  }
  return null;
}

class SearchApp extends StatelessWidget {
  const SearchApp({super.key});

  @override
  Widget build(BuildContext context) {
    final api = context.api.search;
    return ServiceApp(
      serviceId: 'search',
      tabs: [
        AppTab('Tìm kiếm', Icons.search, (context) => SearchableList<SearchDocument>(
              hint: 'Tìm trong toàn bộ Keikibook',
              filters: () async => (await api.searchIndices())?.map((i) => i.name_ ?? i.index_ ?? '').where((s) => s.isNotEmpty).toList(),
              load: (q, index) async => q.isEmpty && index == null ? const [] : (await api.search(q: q.isEmpty ? null : q, index_: index == null ? null : [index], size: 30))?.items,
              emptyMessage: 'Nhập từ khoá để tìm kiếm',
              emptyIcon: Icons.manage_search,
              itemBuilder: (context, d, _) => EntityTile(
                title: _docTitle(d),
                subtitle: _docText(d),
                caption: d.index_,
                icon: Icons.description_outlined,
                onTap: () => openPage(context, SearchDocPage(index: d.index_ ?? '', id: d.id ?? '')),
              ),
            )),
        AppTab('Gợi ý', Icons.auto_awesome, (_) => const _SuggestSearch()),
      ],
    );
  }
}

class SerpApp extends StatelessWidget {
  const SerpApp({super.key});

  @override
  Widget build(BuildContext context) {
    return ServiceApp(
      serviceId: 'serpapi',
      actions: [Builder(builder: (context) => IconButton(tooltip: 'Công cụ', icon: const Icon(Icons.build_outlined), onPressed: () => openPage(context, const SerpToolsPage())))],
      tabs: [
        AppTab('Web', Icons.travel_explore, (context) => _SerpList(engine: 'google', key_: 'organic_results')),
        AppTab('Tin tức', Icons.newspaper, (context) => _SerpList(engine: 'google_news', key_: 'news_results')),
        AppTab('Hình ảnh', Icons.image_search, (context) => _SerpList(engine: 'google_images', key_: 'images_results', grid: true)),
      ],
    );
  }
}

class _SerpList extends StatelessWidget {
  const _SerpList({required this.engine, required this.key_, this.grid = false});

  final String engine;
  final String key_;
  final bool grid;

  @override
  Widget build(BuildContext context) {
    final serp = context.api.serpapi;
    return SearchableList<Map<String, dynamic>>(
      hint: 'Tìm trên Google',
      load: (q, _) async {
        if (q.isEmpty) return const [];
        final r = await serp.serpSearch(engine: engine, params: {'q': q, 'hl': 'vi', 'gl': 'vn'});
        final raw = r is Map ? (r[key_] ?? (r['data'] is Map ? r['data'][key_] : null)) : null;
        return [for (final e in raw is List ? raw : const []) if (e is Map) e.cast<String, dynamic>()];
      },
      gridExtent: grid ? 180 : null,
      gridAspect: 0.85,
      emptyMessage: 'Nhập từ khoá để tìm',
      emptyIcon: Icons.travel_explore,
      itemBuilder: (context, m, _) => grid
          ? PosterCard(title: '${m['title'] ?? ''}', subtitle: '${m['source'] ?? ''}', imageUrl: '${m['thumbnail'] ?? m['original'] ?? ''}', icon: Icons.image_outlined)
          : EntityTile(
              title: '${m['title'] ?? ''}',
              subtitle: '${m['snippet'] ?? (m['source'] is Map ? (m['source'] as Map)['name'] : m['source']) ?? ''}',
              caption: '${m['displayed_link'] ?? m['link'] ?? ''}',
              imageUrl: m['thumbnail'] is String ? m['thumbnail'] as String : null,
              icon: Icons.public,
              onTap: () async {
                await Clipboard.setData(ClipboardData(text: '${m['link'] ?? ''}'));
                if (context.mounted) showSnack(context, 'Đã sao chép liên kết');
              },
            ),
    );
  }
}

class ShortlinkApp extends StatelessWidget {
  const ShortlinkApp({super.key});

  @override
  Widget build(BuildContext context) {
    final api = context.api.shortlink;
    return ServiceApp(
      serviceId: 'shortlink',
      adminActions: [Builder(builder: (context) => IconButton(tooltip: 'Đo lường', icon: const Icon(Icons.analytics_outlined), onPressed: () => openPage(context, const ShortlinkToolsPage())))],
      tabs: [
        AppTab('Tra cứu link', Icons.link, (context) => ListView(
              padding: const EdgeInsets.only(bottom: 24),
              children: [
                AsyncBuilder<ShortlinkSdkHealth?>(
                  load: () => api.shortlinkSdkHealth().catchError((_) => null),
                  loading: const SizedBox(height: 110),
                  builder: (context, h, _) => AppBanner(
                    icon: Icons.link_rounded,
                    title: 'Link rút gọn & deep link',
                    subtitle: h == null ? 'Kiểm tra đích đến của một link rút gọn' : 'Dịch vụ ${h.status ?? ''} • phiên bản ${h.version ?? '-'}',
                  ),
                ),
                Padding(
                  padding: const EdgeInsets.all(16),
                  child: FilledButton.icon(
                    icon: const Icon(Icons.search),
                    label: const Text('Giải mã link rút gọn'),
                    onPressed: () async {
                      ShortlinkResolve? r;
                      await showFormSheet(
                        context,
                        title: 'Mã link rút gọn',
                        fields: const [KField('code', 'Mã (phần sau dấu /)', required: true, hint: 'abc123')],
                        submitLabel: 'Tra cứu',
                        onSubmit: (v) async => r = await api.shortlinkResolve(shortCode: v.str('code')!.split('/').last),
                      );
                      if (r == null || !context.mounted) return;
                      await openPage(
                        context,
                        DetailPage(
                          title: 'Link ${r!.shortCode}',
                          children: [
                            InfoSection(rows: [
                              ('Web', r!.webUrl),
                              ('Android', r!.androidUrl),
                              ('iOS', r!.iosUrl),
                              ('Deep link', r!.deepLinkPath),
                              ('Scheme', r!.appScheme),
                              ('UTM source', r!.utmParameters?.source),
                              ('UTM campaign', r!.utmParameters?.campaign),
                            ]),
                          ],
                        ),
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

class UploadApp extends StatelessWidget {
  const UploadApp({super.key});

  @override
  Widget build(BuildContext context) {
    final api = context.api.upload;
    return ServiceApp(
      serviceId: 'upload',
      tabs: [
        AppTab('Tệp của tôi', Icons.folder_outlined, (context) => AsyncList<UploadFile>(
              load: () async => (await api.uploadFiles(limit: 100))?.files,
              header: Padding(
                padding: const EdgeInsets.fromLTRB(16, 16, 16, 0),
                child: Builder(
                  builder: (context) => FilledButton.icon(
                    icon: const Icon(Icons.cloud_upload_outlined),
                    label: const Text('Tải tệp lên'),
                    onPressed: () => showFormSheet(
                      context,
                      title: 'Tải tệp lên',
                      fields: const [
                        KField('file', 'Tệp', kind: FieldKind.file, required: true),
                        KField('receiver', 'Gửi cho (mã người nhận, tuỳ chọn)'),
                      ],
                      submitLabel: 'Tải lên',
                      success: 'Đã tải lên, kéo xuống để làm mới',
                      onSubmit: (v) => context.api.upload.uploadFileMutation(file: v.file('file')!, receiverId: v.str('receiver')),
                    ),
                  ),
                ),
              ),
              emptyMessage: 'Chưa có tệp nào',
              emptyIcon: Icons.folder_open_outlined,
              itemBuilder: (context, f, reload) => EntityTile(
                onTap: () async {
                  final full = await api.uploadFile(id: f.id!).catchError((_) => null);
                  if (!context.mounted || full == null) return;
                  final replace = await showDialog<bool>(
                    context: context,
                    builder: (context) => AlertDialog(
                      title: Text(full.fileName ?? ''),
                      content: Text('${_size(full.size)} • ${full.contentType ?? ''}\nSHA-256 ${(full.sha256 ?? '').length > 16 ? full.sha256!.substring(0, 16) : full.sha256 ?? ''}…\nCập nhật ${formatDate(full.updatedAt)}'),
                      actions: [
                        TextButton(onPressed: () => Navigator.pop(context, true), child: const Text('Thay tệp')),
                        FilledButton(onPressed: () => Navigator.pop(context, false), child: const Text('Đóng')),
                      ],
                    ),
                  );
                  if (replace != true || !context.mounted) return;
                  if (await showFormSheet(
                    context,
                    title: 'Thay tệp',
                    fields: const [KField('file', 'Tệp mới', kind: FieldKind.file, required: true)],
                    success: 'Đã thay tệp',
                    onSubmit: (v) => api.uploadReplaceFile(id: f.id!, file: v.file('file')!),
                  )) {
                    await reload();
                  }
                },
                title: f.fileName ?? f.id ?? '',
                subtitle: '${_size(f.size)} • ${formatDate(f.createdAt)}',
                icon: _fileIcon(f.contentType),
                imageUrl: (f.contentType ?? '').startsWith('image/') ? f.url : null,
                trailing: PopupMenuButton<String>(
                  onSelected: (a) async {
                    if (a == 'copy') {
                      await Clipboard.setData(ClipboardData(text: f.url ?? ''));
                      if (context.mounted) showSnack(context, 'Đã sao chép liên kết');
                    } else if (await runAction(context, () => api.uploadDeleteFile(id: f.id!), confirm: 'Xoá ${f.fileName}?', destructive: true, success: 'Đã xoá')) {
                      await reload();
                    }
                  },
                  itemBuilder: (_) => const [
                    PopupMenuItem(value: 'copy', child: Text('Sao chép liên kết')),
                    PopupMenuItem(value: 'delete', child: Text('Xoá')),
                  ],
                ),
              ),
            )),
      ],
    );
  }
}

class CdnApp extends StatelessWidget {
  const CdnApp({super.key});

  @override
  Widget build(BuildContext context) {
    final api = context.api.cdn;
    return ServiceApp(
      serviceId: 'cdn',
      tabs: [
        AppTab('Tệp CDN', Icons.cloud_outlined, (context) => AsyncList<CdnFile>(
              load: () async => (await api.cdnFiles(limit: 100))?.files,
              header: Padding(
                padding: const EdgeInsets.fromLTRB(16, 16, 16, 0),
                child: Builder(
                  builder: (context) => OutlinedButton.icon(
                    icon: const Icon(Icons.vpn_key_outlined),
                    label: const Text('Tạo link tải lên (presign)'),
                    onPressed: () async {
                      CdnPresign? p;
                      await showFormSheet(
                        context,
                        title: 'Tải tệp lên CDN',
                        fields: const [KField('file', 'Tệp', kind: FieldKind.file, required: true)],
                        submitLabel: 'Tạo link',
                        onSubmit: (v) async {
                          final f = v.file('file')!;
                          p = await api.cdnPresign(input: CdnPresignInput(fileName: f.filename, contentType: f.contentType ?? 'application/octet-stream', sizeBytes: f.bytes.length));
                        },
                      );
                      if (p == null || !context.mounted) return;
                      await showResult(context, title: 'Link tải lên', message: 'Mã tệp: ${p!.fileId}\nHết hạn: ${formatDate(p!.expiresAt)}\nPUT tệp lên URL đã cấp rồi xác nhận trong danh sách.');
                    },
                  ),
                ),
              ),
              emptyMessage: 'Chưa có tệp trên CDN',
              emptyIcon: Icons.cloud_off_outlined,
              itemBuilder: (context, f, reload) => EntityTile(
                onTap: () async {
                  final full = await api.cdnFile(id: f.id!).catchError((_) => null);
                  if (context.mounted && full != null) await showResult(context, icon: _fileIcon(full.contentType), title: full.fileName ?? '', message: '${_size(full.sizeBytes)} • ${full.contentType ?? ''}\nTrạng thái ${full.status ?? ''}\nCập nhật ${formatDate(full.updatedAt)}');
                },
                title: f.fileName ?? f.id ?? '',
                subtitle: '${_size(f.sizeBytes)} • ${formatDate(f.createdAt)}',
                icon: _fileIcon(f.contentType),
                badges: [StatusPill(f.status)],
                trailing: PopupMenuButton<String>(
                  onSelected: (a) async {
                    if (a == 'confirm') {
                      if (await runAction(context, () => api.cdnConfirm(id: f.id!), success: 'Đã xác nhận tệp')) await reload();
                    } else if (a == 'link') {
                      try {
                        final d = await api.cdnDownloadUrl(id: f.id!);
                        await Clipboard.setData(ClipboardData(text: d?.location ?? ''));
                        if (context.mounted) showSnack(context, 'Đã sao chép link tải');
                      } catch (e) {
                        if (context.mounted) showSnack(context, errorMessage(e));
                      }
                    } else if (await runAction(context, () => api.cdnDeleteFile(id: f.id!), confirm: 'Xoá tệp?', destructive: true, success: 'Đã xoá')) {
                      await reload();
                    }
                  },
                  itemBuilder: (_) => const [
                    PopupMenuItem(value: 'link', child: Text('Sao chép link tải')),
                    PopupMenuItem(value: 'confirm', child: Text('Xác nhận đã tải lên')),
                    PopupMenuItem(value: 'delete', child: Text('Xoá')),
                  ],
                ),
              ),
            )),
      ],
    );
  }
}

class GeoApp extends StatelessWidget {
  const GeoApp({super.key});

  @override
  Widget build(BuildContext context) {
    final api = context.api.geoservice;
    return ServiceApp(
      serviceId: 'geoservice',
      actions: [Builder(builder: (context) => IconButton(tooltip: 'Vùng phục vụ', icon: const Icon(Icons.verified_outlined), onPressed: () => openPage(context, const GeoExtrasPage())))],
      tabs: [
        AppTab('Tỉnh thành', Icons.map_outlined, (context) => SearchableList<GeoProvince>(
              hint: 'Tìm tỉnh, thành phố',
              load: (q, _) async {
                final all = await api.geoProvinces() ?? const <GeoProvince>[];
                final n = q.toLowerCase();
                return n.isEmpty ? all : all.where((p) => '${p.name_} ${p.nameEn}'.toLowerCase().contains(n)).toList();
              },
              emptyMessage: 'Không tìm thấy',
              itemBuilder: (context, p, _) => EntityTile(
                title: p.nameFull ?? p.name_ ?? '',
                subtitle: [p.region, if (p.areaKm2 != null) '${compact(p.areaKm2)} km²'].whereType<String>().join(' • '),
                icon: Icons.location_city_rounded,
                onTap: () => openPage(context, _GeoChildren(title: p.nameFull ?? p.name_ ?? '', load: () async => [for (final d in await api.geoDistricts(code: p.code!) ?? const <GeoDistrict>[]) (d.code!, d.nameFull ?? d.name_ ?? '', d.type_)], next: (code) => api.geoWards(code: code).then((w) => [for (final x in w ?? const <GeoWard>[]) (x.code ?? '', x.nameFull ?? x.name_ ?? '', x.type_)]), leaf: (code) => geoStreetsOf(context, code))),
              ),
            )),
        AppTab('Vị trí', Icons.my_location, (context) => ListView(
              padding: const EdgeInsets.only(bottom: 24),
              children: [
                AsyncBuilder<GeoStats?>(
                  load: () => api.geoStats().catchError((_) => null),
                  loading: const SizedBox(height: 100),
                  builder: (context, s, _) => AppBanner(
                    icon: Icons.public,
                    title: 'Địa giới hành chính Việt Nam',
                    children: [
                      StatRow([
                        StatTile(label: 'Tỉnh/TP', value: '${s?.provinces ?? '-'}', onDark: true),
                        StatTile(label: 'Quận/huyện', value: '${s?.districts ?? '-'}', onDark: true),
                        StatTile(label: 'Phường/xã', value: '${s?.wards ?? '-'}', onDark: true),
                      ]),
                    ],
                  ),
                ),
                Padding(
                  padding: const EdgeInsets.all(16),
                  child: FilledButton.icon(
                    icon: const Icon(Icons.pin_drop_outlined),
                    label: const Text('Tra địa chỉ từ toạ độ'),
                    onPressed: () async {
                      GeoLocationResult? r;
                      await showFormSheet(
                        context,
                        title: 'Toạ độ',
                        fields: const [
                          KField('lat', 'Vĩ độ', kind: FieldKind.number, required: true, initial: '10.7769'),
                          KField('lng', 'Kinh độ', kind: FieldKind.number, required: true, initial: '106.7009'),
                        ],
                        submitLabel: 'Tra cứu',
                        onSubmit: (v) async => r = await api.geoLocate(lat: v.number('lat')!, lng: v.number('lng')!),
                      );
                      if (r == null || !context.mounted) return;
                      await showResult(
                        context,
                        icon: Icons.place_rounded,
                        title: r!.found == true ? 'Đã tìm thấy' : 'Ngoài lãnh thổ',
                        message: [r!.street?.nameFull ?? r!.street?.name_, r!.ward?.nameFull, r!.district?.nameFull, r!.province?.nameFull].whereType<String>().join(', '),
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

class _GeoChildren extends StatelessWidget {
  const _GeoChildren({required this.title, required this.load, this.next, this.leaf});

  final String title;
  final Future<List<(String, String, String?)>> Function() load;
  final Future<List<(String, String, String?)>> Function(String code)? next;
  final Future<List<(String, String, String?)>> Function(String code)? leaf;

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: Text(title)),
      body: AsyncList<(String, String, String?)>(
        load: load,
        spacing: 6,
        emptyMessage: 'Không có dữ liệu',
        itemBuilder: (context, x, _) => EntityTile(
          title: x.$2,
          subtitle: x.$3,
          icon: Icons.place_outlined,
          onTap: next != null
              ? () => openPage(context, _GeoChildren(title: x.$2, load: () => next!(x.$1), next: leaf))
              : null,
        ),
      ),
    );
  }
}

class ShippingApp extends StatelessWidget {
  const ShippingApp({super.key});

  @override
  Widget build(BuildContext context) {
    final api = context.api.shipping;
    return ServiceApp(
      serviceId: 'shipping',
      tabs: [
        AppTab('Tính phí ship', Icons.local_shipping_outlined, (context) => AsyncBuilder<List<ShipProvince>?>(
              load: () => api.shipProvinces(),
              builder: (context, provinces, _) => ListView(
                padding: const EdgeInsets.only(bottom: 24),
                children: [
                  const AppBanner(icon: Icons.local_shipping_rounded, title: 'Báo giá vận chuyển', subtitle: 'So sánh phí giao hàng giữa các hãng theo địa chỉ và kích thước gói'),
                  Padding(
                    padding: const EdgeInsets.all(16),
                    child: FilledButton.icon(
                      icon: const Icon(Icons.calculate_outlined),
                      label: const Text('Tính phí'),
                      onPressed: () => _quote(context, provinces ?? const []),
                    ),
                  ),
                  InfoSection(title: 'Khu vực hỗ trợ', rows: [('Tỉnh/thành', '${(provinces ?? const []).length}')]),
                ],
              ),
            )),
      ],
    );
  }

  Future<void> _quote(BuildContext context, List<ShipProvince> provinces) async {
    final api = context.api.shipping;
    Object? result;
    await showFormSheet(
      context,
      title: 'Thông tin gói hàng',
      fields: [
        KField('province', 'Tỉnh/thành nhận', kind: FieldKind.choice, options: [for (final p in provinces) ('${p.id}', p.name_ ?? '')]),
        const KField('address', 'Địa chỉ chi tiết', required: true),
        const KField('weight', 'Khối lượng (gram)', kind: FieldKind.number, required: true, initial: '500'),
        const KField('cod', 'Thu hộ COD (₫)', kind: FieldKind.number, initial: '0'),
      ],
      submitLabel: 'Báo giá',
      onSubmit: (v) async {
        final provinceId = v.integer('province');
        final districts = provinceId == null ? const <ShipDistrict>[] : (await api.shipDistricts(provinceId: provinceId) ?? const <ShipDistrict>[]);
        final wards = districts.isEmpty ? const <ShipWard>[] : (await api.shipWards(districtId: districts.first.id!).catchError((_) => null) ?? const <ShipWard>[]);
        result = await api.shipQuote(
          input: ShipQuoteInput(
            recipient: ShipAddressInput(provinceId: provinceId, districtId: districts.isEmpty ? null : districts.first.id, wardCode: wards.isEmpty ? null : wards.first.code, address: v.str('address'), name_: 'Khách Keiki'),
            parcel: ShipParcelInput(weight: v.integer('weight'), length: 20, width: 15, height: 10),
            codAmount: v.integer('cod'),
          ),
        );
      },
    );
    if (result == null || !context.mounted) return;
    final quotes = result is List ? result as List : (pick(result, 'quotes') ?? pick(result, 'items') ?? const []) as List;
    await showResult(
      context,
      icon: Icons.local_shipping_rounded,
      title: 'Báo giá',
      message: quotes.isEmpty
          ? 'Phí: ${vnd(pickInt(result, 'total') ?? pickInt(result, 'fee'))}'
          : quotes.map((q) => '${pick(q, 'carrier') ?? pick(q, 'provider') ?? ''}: ${vnd(pickInt(q, 'total') ?? pickInt(q, 'fee'))}').join('\n'),
    );
  }
}

class ShopifyApp extends StatelessWidget {
  const ShopifyApp({super.key});

  @override
  Widget build(BuildContext context) {
    final api = context.api.shopify;
    return ServiceApp(
      serviceId: 'shopify',
      tabs: [
        AppTab('Sản phẩm', Icons.storefront_outlined, (context) => AsyncList<ShopifyProduct>(
              load: () async => (await api.shopifyProducts(limit: 50))?.items,
              gridExtent: 200,
              gridAspect: 0.72,
              emptyMessage: 'Cửa hàng chưa có sản phẩm',
              emptyIcon: Icons.storefront_outlined,
              itemBuilder: (context, p, _) {
                final v = (p.variants ?? const <ShopifyProductVariant>[]).firstOrNull;
                return PosterCard(
                  title: p.title ?? 'Sản phẩm',
                  subtitle: p.vendor,
                  icon: Icons.shopping_bag_rounded,
                  badge: v?.price == null ? null : formatMoneyString(v!.price),
                  onTap: () => openPage(
                    context,
                    AsyncBuilder<ShopifyProduct?>(
                      load: () => api.shopifyProduct(id: p.id!).catchError((_) => p),
                      loading: const Scaffold(body: SkeletonList()),
                      builder: (context, fresh, _) => _ShopifyDetail(p: fresh ?? p),
                    ),
                  ),
                );
              },
            )),
      ],
    );
  }
}

class _ShopifyDetail extends StatelessWidget {
  const _ShopifyDetail({required this.p});

  final ShopifyProduct p;

  @override
  Widget build(BuildContext context) {
    return DetailPage(
                      title: p.title ?? 'Sản phẩm',
                      header: HeroHeader(title: p.title ?? '', subtitle: p.vendor, icon: Icons.shopping_bag_rounded, badges: [if (p.productType != null) Pill(p.productType!), StatusPill(p.status)]),
                      children: [
                        const SectionHeader('Phiên bản'),
                        for (final x in p.variants ?? const <ShopifyProductVariant>[])
                          Padding(
                            padding: const EdgeInsets.fromLTRB(16, 0, 16, 8),
                            child: EntityTile(title: x.title ?? '', subtitle: 'SKU ${x.sku ?? '-'} • còn ${x.inventoryQuantity ?? 0}', caption: formatMoneyString(x.price), icon: Icons.style_outlined),
                          ),
                        InfoSection(rows: [('Thẻ', p.tags), ('Handle', p.handle)]),
                      ],
    );
  }
}

class _SuggestSearch extends StatefulWidget {
  const _SuggestSearch();

  @override
  State<_SuggestSearch> createState() => _SuggestSearchState();
}

class _SuggestSearchState extends State<_SuggestSearch> {
  List<String> _suggest = const [];
  String _q = '';

  @override
  Widget build(BuildContext context) {
    return Column(
      children: [
        Padding(
          padding: const EdgeInsets.fromLTRB(16, 14, 16, 0),
          child: TextField(
            decoration: const InputDecoration(hintText: 'Gõ vài chữ để xem gợi ý', prefixIcon: Icon(Icons.auto_awesome), isDense: true),
            onChanged: (v) async {
              final list = v.trim().length < 2 ? const <String>[] : await searchSuggest(context, v.trim());
              if (mounted) setState(() => _suggest = list);
            },
            onSubmitted: (v) => setState(() => _q = v.trim()),
          ),
        ),
        if (_suggest.isNotEmpty)
          SizedBox(
            height: 52,
            child: ListView(
              scrollDirection: Axis.horizontal,
              padding: const EdgeInsets.fromLTRB(16, 10, 16, 0),
              children: [for (final s in _suggest) Padding(padding: const EdgeInsets.only(right: 8), child: ActionChip(label: Text(s), onPressed: () => setState(() => _q = s)))],
            ),
          ),
        Expanded(
          child: _q.isEmpty
              ? const EmptyView(message: 'Chọn một gợi ý để tìm', icon: Icons.manage_search)
              : AsyncList<SearchDocument>(
                  key: ValueKey(_q),
                  load: () => searchDocs(context, _q, null),
                  emptyMessage: 'Không có kết quả',
                  itemBuilder: (context, d, _) => EntityTile(
                    title: _docTitle(d),
                    subtitle: _docText(d),
                    caption: d.index_,
                    icon: Icons.description_outlined,
                    onTap: () => openPage(context, SearchDocPage(index: d.index_ ?? '', id: d.id ?? '')),
                  ),
                ),
        ),
      ],
    );
  }
}
