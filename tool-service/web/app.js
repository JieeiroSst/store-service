'use strict';
const $ = id => document.getElementById(id);
const API = '/api/v1';
const esc = s => String(s ?? '').replace(/[&<>"']/g, c => ({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[c]));
const fmt = (n, d = 1) => (n === undefined || n === null || isNaN(n)) ? '–' : Number(n).toFixed(d);
const store = { get(k){ try { return sessionStorage.getItem(k) || '' } catch { return '' } }, set(k,v){ try { sessionStorage.setItem(k,v) } catch {} } };

(function initTheme(){
  let t = ''; try { t = localStorage.getItem('theme') || '' } catch {}
  if (!t) t = matchMedia('(prefers-color-scheme:dark)').matches ? 'dark' : 'light';
  document.documentElement.dataset.theme = t;
  $('themeBtn').onclick = () => {
    const n = document.documentElement.dataset.theme === 'dark' ? 'light' : 'dark';
    document.documentElement.dataset.theme = n; try { localStorage.setItem('theme', n) } catch {}
  };
})();
const TABS = {
  task:  ['Task từ file', 'Upload PDF/Word — AI xác định cần test gì, viết test case, chạy và benchmark.'],
  scan:  ['Quét sản phẩm', 'Quét bảo mật thụ động và duyệt qua website: link hỏng, lỗi HTML, khả năng truy cập cơ bản.'],
  adv:   ['Kiểm thử nâng cao', 'Trình duyệt thật, cơ sở dữ liệu, luồng bất đồng bộ, hàng đợi, bảo mật chủ động, stress và soak.'],
  suite: ['Golden suite', 'Bộ test đã được người duyệt: chạy xác định, không dùng AI — dùng làm cổng CI chặn build.'],
  qc:    ['QC từ mô tả', 'Mô tả API hoặc dán code handler, AI làm trọn việc của một QC.'],
  bench: ['Benchmark', 'Đo throughput và độ trễ p50 / p95 / p99 của một endpoint.'],
  code:  ['Review code', 'AI review logic, độ phức tạp, điểm nóng, sinh unit test và benchmark.'],
  learn: ['Học & thống kê', 'Theo dõi chất lượng agent và duyệt các ca AI chưa chắc chắn.'],
};
function showTab(t){
  document.querySelectorAll('#nav button').forEach(b => b.classList.toggle('on', b.dataset.t === t));
  document.querySelectorAll('.tab').forEach(s => s.hidden = s.id !== t);
  $('title').textContent = TABS[t][0]; $('subtitle').textContent = TABS[t][1];
  $('out').innerHTML = ''; store.set('tab', t);
}
document.querySelectorAll('#nav button').forEach(b => b.onclick = () => showTab(b.dataset.t));
$('setBtn').onclick = () => { $('token').value = store.get('token'); $('dlg').showModal(); };
$('dlgok').onclick = () => { store.set('token', $('token').value.trim()); $('dlg').close(); toast('Đã lưu token'); };

let toastT;
function toast(m){ const t = $('toast'); t.textContent = m; t.classList.add('on'); clearTimeout(toastT); toastT = setTimeout(() => t.classList.remove('on'), 3200); }

function headers(json){ const h = {}; if (json) h['Content-Type'] = 'application/json'; const tk = store.get('token'); if (tk) h.Authorization = 'Bearer ' + tk; return h; }
async function api(path, { method = 'GET', body, form, title = 'Đang chạy…', sub = 'Model có thể mất vài phút.', btn } = {}) {
  $('out').innerHTML = ''; $('loadtitle').textContent = title; $('loadsub').textContent = sub; $('loader').classList.add('on');
  if (btn) btn.disabled = true;
  try {
    const r = await fetch(API + path, { method, headers: headers(!!body), body: form || (body ? JSON.stringify(body) : undefined) });
    let j = null; const ct = r.headers.get('content-type') || '';
    if (ct.includes('json')) j = await r.json(); else j = { raw: await r.text() };
    if (r.status === 401) { toast('Cần API token — mở Cài đặt'); $('setBtn').click(); }
    if (!r.ok) { $('out').innerHTML = alertBox('bad', `Lỗi ${r.status}`, j.error || j.raw || ''); return r.status === 422 ? j : null; }
    return j;
  } catch (e) { $('out').innerHTML = alertBox('bad', 'Không kết nối được', e.message); return null; }
  finally { $('loader').classList.remove('on'); if (btn) btn.disabled = false; }
}
const alertBox = (k, t, m) => `<div class="banner ${k === 'bad' ? 'fail' : 'info'}"><div><div class="big" style="font-size:16px">${esc(t)}</div><div style="color:var(--fg);font-weight:500">${esc(m)}</div></div></div>`;

const pill = (ok, a = 'PASS', b = 'FAIL') => `<span class="pill ${ok ? 'ok' : 'bad'}">${ok ? a : b}</span>`;
const tile = (n, l, cls = '') => `<div class="tile"><div class="n ${cls}">${esc(n)}</div><div class="l">${esc(l)}</div></div>`;
const methodTag = m => `<span class="method">${esc((m || 'GET').toUpperCase())}</span>`;
const rawBlock = (t, o) => `<details><summary>${esc(t)}</summary><div class="in"><pre>${esc(JSON.stringify(o, null, 2))}</pre></div></details>`;
const list = a => (a && a.length) ? `<ul class="list">${a.map(x => `<li>${esc(typeof x === 'string' ? x : JSON.stringify(x))}</li>`).join('')}</ul>` : '<span class="muted">–</span>';
function codeBlock(text){ const id = 'cb' + Math.random().toString(36).slice(2, 8); return `<div class="codewrap"><button class="btn sm" data-copy="${id}">Copy</button><pre id="${id}">${esc(text)}</pre></div>`; }
document.addEventListener('click', e => {
  const b = e.target.closest('[data-copy]'); if (!b) return;
  navigator.clipboard?.writeText($(b.dataset.copy).textContent).then(() => toast('Đã copy'));
});

function gateBanner(g){
  if (!g) return `<div class="banner info"><div>Chưa nhập URL đích: mới sinh kế hoạch và test case, <b>chưa chạy</b>.</div></div>`;
  const st = g.status || (g.passed ? 'passed' : 'failed');
  const cls = st === 'passed' ? 'pass' : st === 'needs_review' ? 'warn' : 'fail';
  const title = { passed: '✓ QUALITY GATE PASSED', failed: '✗ QUALITY GATE FAILED', needs_review: '⚠ CẦN NGƯỜI DUYỆT' }[st] || st;
  const sub = (g.total_cases || (st === 'needs_review' && g.passed)) ? `<div style="font-size:12px;font-weight:500;color:var(--mut)">${g.total_cases ? `${g.trusted_cases ?? 0}/${g.total_cases} case đáng tin` : ''}${st === 'needs_review' && g.passed ? ' · chế độ chỉ cảnh báo' : ''}</div>` : '';
  return `<div class="banner ${cls}"><div><div class="big">${title}</div>${sub}${(g.reasons || []).length ? `<ul>${g.reasons.map(r => `<li>${esc(r)}</li>`).join('')}</ul>` : ''}</div></div>`;
}

function trustBadge(t){
  if (!t) return '';
  const money = t.critical ? ' <span class="pill warn" title="Liên quan tiền">💰</span>' : '';
  if (t.approved) return `<span class="pill ok">Đã duyệt</span>${money}`;
  if (t.trusted) return `<span class="pill ok">Đáng tin</span>${money}`;
  return `<span class="pill warn" title="${esc((t.reasons || []).join(' | '))}">Cần duyệt</span>${money}`;
}
function casesTable(cases, report, trust, selectable){
  const res = {}; (report?.results || []).forEach((r, i) => res[i] = r);
  if (!(cases || []).length) return '<div class="empty">Không có test case.</div>';
  const rows = cases.map((c, i) => {
    const r = res[i], t = (trust || [])[i];
    return `<tr>${selectable ? `<td><input type="checkbox" class="pick" value="${esc(c.name)}" ${t && t.approved ? 'checked disabled' : ''} aria-label="Chọn ${esc(c.name)}"></td>` : ''}
      <td>${methodTag(c.method)} <span class="mono">${esc(c.path)}</span><div>${esc(c.name)}</div>
      ${c.spec_quote ? `<div class="muted" style="font-size:12px">“${esc(c.spec_quote)}”</div>` : ''}
      ${c.expect_json ? `<div class="mono muted">${esc(JSON.stringify(c.expect_json))}</div>` : ''}
      ${t && t.reasons && t.reasons.length ? `<ul class="fails" style="color:var(--warn)">${t.reasons.map(x => `<li>${esc(x)}</li>`).join('')}</ul>` : ''}
      ${r && r.failures ? `<ul class="fails">${r.failures.map(f => `<li>${esc(f)}</li>`).join('')}</ul>` : ''}</td>
      ${trust ? `<td class="trustcell">${trustBadge(t)}${t ? `<div class="muted" style="font-size:11px">${t.votes} lần đồng thuận</div>` : ''}</td>` : ''}
      <td>${c.expect_status || 200}</td><td>${r ? r.status || '–' : '–'}</td><td>${r ? r.latency_ms + ' ms' : '–'}</td>
      <td>${r ? pill(r.passed) : '<span class="pill mut">chưa chạy</span>'}</td></tr>`;
  }).join('');
  return `<div class="tw"><table><thead><tr>${selectable ? '<th></th>' : ''}<th>Test case</th>${trust ? '<th>Độ tin cậy</th>' : ''}<th>Mong đợi</th><th>Thực tế</th><th>Latency</th><th>Kết quả</th></tr></thead><tbody>${rows}</tbody></table></div>`;
}

function statsTiles(s){
  const errRate = s.total_requests ? (100 * s.failed / s.total_requests) : 0;
  return `<div class="grid g4">${tile(fmt(s.rps, 0), 'requests / giây')}${tile(fmt(s.p50_ms, 1) + ' ms', 'p50')}${tile(fmt(s.p95_ms, 1) + ' ms', 'p95')}${tile(fmt(s.p99_ms, 1) + ' ms', 'p99')}
    ${tile(s.total_requests, 'tổng request')}${tile(fmt(errRate, 1) + '%', 'tỉ lệ lỗi', errRate > 1 ? 'bad-t' : '')}${tile(fmt(s.mean_ms, 1) + ' ms', 'trung bình')}${tile(fmt(s.max_ms, 1) + ' ms', 'max')}</div>`;
}
function statusCodes(s){
  const e = Object.entries(s.status_codes || {}); if (!e.length) return '';
  return `<div style="margin-top:12px">${e.map(([c, n]) => `<span class="pill ${c < 400 ? 'ok' : 'bad'}" style="margin-right:6px">${esc(c)} × ${n}</span>`).join('')}</div>` +
    (s.errors ? `<ul class="fails">${Object.entries(s.errors).map(([m, n]) => `<li>${n}× ${esc(m)}</li>`).join('')}</ul>` : '');
}
function perfTable(perf){
  if (!(perf || []).length) return '';
  const rows = perf.map(p => {
    const t = p.target || {}, s = p.stats;
    if (p.skipped) return `<tr><td>${methodTag(t.method)} <span class="mono">${esc(t.path)}</span></td><td colspan="5"><span class="pill mut">bỏ qua</span> <span class="muted">${esc(p.skipped)}</span></td></tr>`;
    const pct = t.p95_ms ? Math.min(100, 100 * s.p95_ms / t.p95_ms) : 0;
    return `<tr><td>${methodTag(t.method)} <span class="mono">${esc(t.path)}</span>${p.failures ? `<ul class="fails">${p.failures.map(f => `<li>${esc(f)}</li>`).join('')}</ul>` : ''}</td>
      <td>${fmt(s.p95_ms)} ms${t.p95_ms ? ` <span class="muted">/ ${t.p95_ms}</span>` : ''}${t.p95_ms ? `<div class="bar"><i class="${s.p95_ms > t.p95_ms ? 'over' : ''}" style="width:${pct}%"></i></div>` : ''}</td>
      <td>${fmt(s.p99_ms)} ms</td><td>${fmt(s.rps, 0)}${t.min_rps ? ` <span class="muted">/ ${t.min_rps}</span>` : ''}</td><td>${s.failed}/${s.total_requests}</td><td>${pill(p.passed, 'OK', 'FAIL')}</td></tr>`;
  }).join('');
  return `<div class="tw"><table><thead><tr><th>Endpoint</th><th>p95 / mục tiêu</th><th>p99</th><th>RPS / mục tiêu</th><th>Lỗi</th><th></th></tr></thead><tbody>${rows}</tbody></table></div>`;
}
function planCard(p){
  p = p || {};
  const eps = (p.endpoints || []).map(e => `<tr><td>${methodTag(e.method)} <span class="mono">${esc(e.path)}</span></td><td>${esc(e.description)}</td><td>${esc(e.auth || '–')}</td></tr>`).join('');
  const comps = (p.code_components || []).map(c => `<div class="issue"><b>${esc(c.name)}</b> — ${esc(c.responsibilities)}${(c.risks || []).length ? `<div class="muted">Rủi ro: ${esc(c.risks.join('; '))}</div>` : ''}</div>`).join('');
  return `<div class="card"><h2>AI xác định cần test</h2><p style="margin-top:0">${esc(p.summary || '')}</p>
    ${(p.acceptance_criteria || []).length ? `<h3>Tiêu chí nghiệm thu</h3>${list(p.acceptance_criteria)}` : ''}
    ${eps ? `<h3 style="margin-top:14px">Endpoint</h3><div class="tw"><table><thead><tr><th>Endpoint</th><th>Mô tả</th><th>Auth</th></tr></thead><tbody>${eps}</tbody></table></div>` : ''}
    ${comps ? `<h3 style="margin-top:14px">Thành phần code</h3>${comps}` : ''}</div>`;
}
const SEV = { high: ['bad', 'Cao'], medium: ['warn', 'Trung bình'], low: ['mut', 'Thấp'], info: ['mut', 'Thông tin'], error: ['bad', 'Lỗi'], warn: ['warn', 'Cảnh báo'] };
const sevPill = s => `<span class="pill ${(SEV[s] || ['mut'])[0]}">${esc((SEV[s] || [0, s])[1])}</span>`;
function coverageCard(c, ct){
  if (!c || !c.total_requirements) return '';
  const bar = (v, cls) => `<div class="bar" style="height:10px"><i class="${cls}" style="width:${v}%"></i></div>`;
  return `<div class="card"><h2>Độ phủ yêu cầu</h2>
    <div class="grid g2"><div><b>${c.percent}%</b> yêu cầu có test <span class="muted">(${c.covered}/${c.total_requirements})</span>${bar(c.percent, '')}</div>
    <div><b>${ct ? ct.percent : 0}%</b> có test <u>đáng tin</u> <span class="muted">(${ct ? ct.covered : 0}/${c.total_requirements})</span>${bar(ct ? ct.percent : 0, '')}</div></div>
    ${(c.uncovered || []).length ? `<details style="margin-top:12px"><summary>${c.uncovered.length} yêu cầu chưa có test nào</summary><div class="in">${list(c.uncovered)}</div></details>` : '<p class="muted" style="margin-bottom:0">Mọi yêu cầu trong tài liệu đều có test.</p>'}
    <div class="hint">Yêu cầu được trích tự động từ các câu có “must / shall / phải / không được / tối đa…” — chỉ là ước lượng, không thay thế việc rà soát tài liệu.</div></div>`;
}
function securityCard(list_){
  if (!list_) return '';
  const rows = list_.map(f => `<tr><td>${sevPill(f.severity)}</td><td class="mono">${esc(f.id)}</td><td><b>${esc(f.title)}</b>${f.detail ? `<div class="muted">${esc(f.detail)}</div>` : ''}${f.fix ? `<div style="font-size:12px">↳ ${esc(f.fix)}</div>` : ''}</td></tr>`).join('');
  return `<div class="card"><h2>Bảo mật (quét thụ động)</h2>${rows ? `<div class="tw"><table><thead><tr><th>Mức</th><th>Mã</th><th>Phát hiện</th></tr></thead><tbody>${rows}</tbody></table></div>` : '<div class="empty">Không phát hiện vấn đề.</div>'}</div>`;
}
function webCard(w){
  if (!w) return '';
  const broken = (w.broken_links || []).map(b => `<tr><td class="mono">${esc(b.url)}</td><td>${b.status || esc(b.error || '')}</td><td class="mono muted">${esc(b.from)}</td></tr>`).join('');
  const pages = (w.pages || []).map(p => `<details><summary><span><span class="mono">${esc(p.url)}</span> <span class="muted">${p.status} · ${p.latency_ms} ms · ${p.size_kb.toFixed(1)} KB</span></span><span>${(p.issues || []).length ? `<span class="pill warn">${p.issues.length} vấn đề</span>` : '<span class="pill ok">OK</span>'}</span></summary>
    <div class="in">${(p.issues || []).length ? `<div class="tw"><table><tbody>${p.issues.map(i => `<tr><td>${sevPill(i.severity)}</td><td class="mono">${esc(i.rule)}</td><td>${esc(i.detail)}</td></tr>`).join('')}</tbody></table></div>` : '<span class="muted">Không có vấn đề.</span>'}</div></details>`).join('');
  return `<div class="card"><h2>Trang web</h2><div class="grid g4" style="margin-bottom:12px">${tile((w.pages || []).length, 'trang đã duyệt')}${tile((w.broken_links || []).length, 'link hỏng', (w.broken_links || []).length ? 'bad-t' : '')}${tile(w.error_count, 'lỗi', w.error_count ? 'bad-t' : '')}${tile(w.warning_count, 'cảnh báo')}</div>
    ${broken ? `<h3>Link hỏng</h3><div class="tw"><table><thead><tr><th>URL</th><th>Mã</th><th>Từ trang</th></tr></thead><tbody>${broken}</tbody></table></div>` : ''}<div style="margin-top:12px">${pages}</div></div>`;
}
function e2eCard(e){
  if (!e) return '';
  return `<div class="card"><h2>Kịch bản E2E (đã sinh, chưa chạy)</h2><p class="muted" style="margin-top:0">${esc(e.note || '')}</p>
    ${(e.scenarios || []).map(s => `<div class="issue"><b>${esc(s.name)}</b><ol class="list">${(s.steps || []).map(x => `<li>${esc(x)}</li>`).join('')}</ol><div class="muted">Kỳ vọng: ${esc(s.expected || '')}</div></div>`).join('')}
    ${e.playwright_script ? `<details><summary>Script Playwright</summary><div class="in">${codeBlock(e.playwright_script)}</div></details>` : ''}</div>`;
}
function scopeCard(sc){
  if (!(sc || []).length) return '';
  const st = { automated: ['ok', 'Tự động'], generated: ['warn', 'Chỉ sinh script'], manual: ['bad', 'Chưa bao phủ'] };
  return `<details><summary>Phạm vi kiểm thử: những gì công cụ này KHÔNG thay thế được</summary><div class="in"><div class="tw"><table><thead><tr><th>Hạng mục</th><th>Trạng thái</th><th>Ghi chú</th></tr></thead><tbody>
    ${sc.map(x => `<tr><td>${esc(x.area)}</td><td><span class="pill ${st[x.status][0]}">${st[x.status][1]}</span></td><td class="muted">${esc(x.note)}</td></tr>`).join('')}</tbody></table></div></div></details>`;
}
function approveBar(j){
  if (!j.experience_id) return '';
  return `<div class="fb"><b>Duyệt case vào golden suite</b>
    <div class="muted" style="font-size:12px">Tick những case bạn đã kiểm tra kỹ (đặc biệt số tiền, kỳ vọng). Case đã duyệt chạy xác định, không cần AI, và được phép chặn build. Case không tick sẽ tính là bị từ chối.</div>
    <input id="ap-suite" aria-label="Tên suite" style="margin-top:10px" placeholder="Tên suite, vd: payments" value="${esc(j.suite || '')}">
    <div class="row"><button class="btn primary" id="ap-btn" data-exp="${esc(j.experience_id)}">Duyệt các case đã chọn</button></div>
    ${j.suite_precision ? precisionLine(j.suite_precision) : ''}</div>`;
}
function precisionLine(p){
  if (!p || p.accepted + p.rejected === 0) return '<div class="hint">Chưa có dữ liệu độ chính xác của AI cho suite này.</div>';
  return `<div class="hint">Độ chính xác của AI theo người duyệt: <b>${(p.rate * 100).toFixed(1)}%</b> (${p.accepted} nhận / ${p.rejected} loại) · cận dưới 95%: <b>${(p.lower_bound_95 * 100).toFixed(1)}%</b> ${p.meets_99_percent ? '<span class="pill ok">đạt 99%</span>' : '<span class="pill warn">chưa chứng minh được 99%</span>'}</div>`;
}
function codeCard(a){
  if (!a) return '';
  if (a.error) return `<div class="card"><h2>Phân tích code</h2>${alertBox('bad', 'Lỗi', a.error)}</div>`;
  const li = (a.logic_issues || []).map(i => `<div class="issue bug"><b>${esc(i.line || '')}</b> ${esc(i.issue || '')}${i.fix ? `<div class="muted">Sửa: ${esc(i.fix)}</div>` : ''}</div>`).join('');
  const pf = a.performance || {};
  return `<div class="card"><h2>Phân tích code</h2><p style="margin-top:0">${esc(a.summary || '')}</p>
    ${li ? `<h3>Vấn đề logic</h3>${li}` : '<p class="muted">Không phát hiện vấn đề logic.</p>'}
    <h3 style="margin-top:14px">Hiệu năng</h3><dl class="kv"><dt>Time</dt><dd>${esc(pf.time_complexity || '–')}</dd><dt>Space</dt><dd>${esc(pf.space_complexity || '–')}</dd>
      <dt>Điểm nóng</dt><dd>${list(pf.hotspots)}</dd><dt>Tối ưu</dt><dd>${list(pf.optimizations)}</dd></dl>
    ${a.unit_tests ? `<details style="margin-top:12px"><summary>Unit test đề xuất</summary><div class="in">${codeBlock(a.unit_tests)}</div></details>` : ''}
    ${a.benchmark_code ? `<details><summary>Benchmark code đề xuất</summary><div class="in">${codeBlock(a.benchmark_code)}</div></details>` : ''}</div>`;
}
function feedbackBox(id){
  if (!id) return '';
  return `<div class="fb"><b>Kết quả này có tốt không?</b><div class="muted" style="font-size:12px">Đánh giá của bạn giúp AI học và luôn ghi đè đánh giá của AI judge.</div>
    <input id="fb-note" aria-label="Ghi chú đánh giá" style="margin-top:10px" placeholder="Ghi chú: case sai ở đâu, bug thật hay AI viết sai…">
    <div class="row"><button class="btn ok" data-fb="good" data-id="${esc(id)}">👍 Tốt</button><button class="btn bd" data-fb="bad" data-id="${esc(id)}">👎 Chưa tốt</button></div></div>`;
}
document.addEventListener('click', async e => {
  const b = e.target.closest('[data-fb]'); if (!b) return;
  const card = b.closest('.fb'), note = card.querySelector('input')?.value || '';
  b.disabled = true;
  const r = await fetch(API + '/learning/feedback', { method: 'POST', headers: headers(true), body: JSON.stringify({ id: b.dataset.id, verdict: b.dataset.fb, note }) });
  if (r.ok) { card.innerHTML = '<b>Cảm ơn! ✓</b> <span class="muted">Đã ghi nhận đánh giá.</span>'; toast('Đã gửi đánh giá'); } else { b.disabled = false; toast('Lỗi ' + r.status); }
});

let pickedFile = null;
const drop = $('drop'), fileIn = $('t-file');
function setFile(f){
  pickedFile = f; $('chip').classList.toggle('on', !!f); $('chipname').textContent = f ? `${f.name} · ${(f.size / 1024).toFixed(0)} KB` : '';
}
drop.onclick = () => fileIn.click();
drop.onkeydown = e => { if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); fileIn.click(); } };
fileIn.onchange = () => setFile(fileIn.files[0] || null);
$('chipx').onclick = () => { fileIn.value = ''; setFile(null); };
['dragenter', 'dragover'].forEach(ev => drop.addEventListener(ev, e => { e.preventDefault(); drop.classList.add('over'); }));
['dragleave', 'drop'].forEach(ev => drop.addEventListener(ev, e => { e.preventDefault(); drop.classList.remove('over'); }));
drop.addEventListener('drop', e => { const f = e.dataTransfer.files[0]; if (f) setFile(f); });

$('t-run').onclick = async () => {
  if (!pickedFile) { toast('Hãy chọn file task trước'); return; }
  const fd = new FormData();
  fd.append('file', pickedFile); fd.append('base_url', $('t-url').value.trim()); fd.append('code', $('t-code').value);
  fd.append('language', $('t-lang').value); fd.append('min_pass_rate', $('t-min').value); fd.append('benchmark', $('t-bench').value);
  fd.append('suite', $('t-suite').value.trim()); fd.append('e2e', $('t-e2e').value); fd.append('consistency_runs', $('t-runs').value);
  const j = await api('/tasks/analyze', { method: 'POST', form: fd, btn: $('t-run'), title: 'Đang phân tích task…', sub: 'Đọc file → lập kế hoạch → sinh test nhiều lần độc lập → kiểm chứng → chạy → benchmark.' });
  if (!j || j.error) return;
  const r = j.report, pr = r && r.total ? Math.round(100 * r.passed / r.total) : null;
  const trustedN = (j.trust || []).filter(t => t.trusted).length;
  const perfOk = (j.performance || []).filter(p => !p.skipped && p.passed).length, perfN = (j.performance || []).filter(p => !p.skipped).length;
  $('out').innerHTML = gateBanner(j.gate)
    + `<div class="grid g4" style="margin-bottom:16px">${tile((j.cases || []).length, 'test case')}${tile(`${trustedN}/${(j.cases || []).length}`, 'case đáng tin')}${tile(pr === null ? '–' : pr + '%', 'test pass')}${tile(perfN ? `${perfOk}/${perfN}` : '–', 'endpoint đạt hiệu năng')}${tile(j.model, 'model')}</div>`
    + planCard(j.plan)
    + `<div class="card"><h2>Test case</h2>${casesTable(j.cases, r, j.trust, !!j.experience_id)}${approveBar(j)}</div>`
    + (perfN || (j.performance || []).length ? `<div class="card"><h2>Hiệu năng</h2>${perfTable(j.performance)}</div>` : '')
    + coverageCard(j.requirement_coverage, j.requirement_coverage_trusted) + securityCard(j.security) + webCard(j.web)
    + e2eCard(j.e2e) + codeCard(j.code_analysis) + scopeCard(j.scope) + rawBlock('JSON thô', j) + feedbackBox(j.experience_id);
};

$('qc-run').onclick = async () => {
  const j = await api('/qc/run', { method: 'POST', btn: $('qc-run'), title: 'Đang QC…', body: { base_url: $('qc-url').value.trim(), spec: $('qc-spec').value, count: +$('qc-count').value, benchmark: $('qc-bench').value === 'true' } });
  if (!j || j.error) return;
  const v = j.verdict || {}, ok = String(v.verdict).toLowerCase() === 'pass', r = j.evidence?.report, b = j.evidence?.benchmark;
  const bugs = (v.bugs || []).map(x => `<div class="issue bug"><b>${esc(x.title)}</b> <span class="pill warn">${esc(x.severity || '')}</span><div class="muted">${esc(x.evidence || '')}</div></div>`).join('');
  $('out').innerHTML = `<div class="banner ${ok ? 'pass' : 'fail'}"><div><div class="big">${ok ? '✓ PASS' : '✗ FAIL'}</div><div style="color:var(--fg);font-weight:500">Độ tin cậy: ${v.confidence !== undefined ? Math.round(v.confidence * 100) + '%' : '–'}</div></div></div>`
    + `<div class="card"><h2>Phán quyết của AI</h2>${bugs || '<p class="muted">Không phát hiện bug.</p>'}
       ${(v.likely_test_errors || []).length ? `<h3 style="margin-top:12px">Có thể do test viết sai</h3>${list(v.likely_test_errors)}` : ''}
       ${v.performance ? `<h3 style="margin-top:12px">Hiệu năng</h3><p>${esc(v.performance)}</p>` : ''}
       ${(v.recommendations || []).length ? `<h3 style="margin-top:12px">Khuyến nghị</h3>${list(v.recommendations)}` : ''}</div>`
    + `<div class="card"><h2>Test case</h2>${casesTable(j.cases, r)}</div>`
    + (b ? `<div class="card"><h2>Benchmark</h2>${statsTiles(b)}${statusCodes(b)}</div>` : '') + rawBlock('JSON thô', j) + feedbackBox(j.experience_id);
};

$('b-run').onclick = async () => {
  let hd; try { hd = $('b-headers').value ? JSON.parse($('b-headers').value) : undefined; } catch { toast('Headers không phải JSON hợp lệ'); return; }
  const j = await api('/benchmark', { method: 'POST', btn: $('b-run'), title: 'Đang benchmark…', sub: 'Đang bắn tải vào endpoint.', body: { url: $('b-url').value.trim(), method: $('b-method').value, headers: hd, body: $('b-body').value,
    concurrency: +$('b-conc').value, requests: +$('b-req').value, duration_sec: +$('b-dur').value, analyze: $('b-an').checked } });
  if (!j || !j.stats) return;
  const s = j.stats, an = j.analysis;
  $('out').innerHTML = `<div class="card"><h2>Kết quả</h2>${statsTiles(s)}${statusCodes(s)}</div>`
    + (an ? `<div class="card"><h2>Phân tích của AI</h2><div style="white-space:pre-wrap">${esc(typeof an === 'string' ? an : JSON.stringify(an, null, 2))}</div></div>` : '') + rawBlock('JSON thô', j);
};

$('c-run').onclick = async () => {
  const j = await api('/analyze/code', { method: 'POST', btn: $('c-run'), title: 'Đang phân tích code…', body: { code: $('c-code').value, language: $('c-lang').value, focus: $('c-focus').value } });
  if (!j || !j.analysis) return;
  $('out').innerHTML = codeCard(j.analysis) + rawBlock('JSON thô', j) + feedbackBox(j.experience_id);
};

document.addEventListener('click', async e => {
  const b = e.target.closest('#ap-btn'); if (!b) return;
  const name = $('ap-suite').value.trim(); if (!name) { toast('Nhập tên suite'); return; }
  const picks = [...document.querySelectorAll('.pick:not(:disabled)')], approve = picks.filter(x => x.checked).map(x => x.value), reject = picks.filter(x => !x.checked).map(x => x.value);
  if (!approve.length) { toast('Chưa tick case nào'); return; }
  b.disabled = true;
  const r = await fetch(API + '/suites/' + encodeURIComponent(name) + '/approve', { method: 'POST', headers: headers(true), body: JSON.stringify({ experience_id: b.dataset.exp, approve, reject }) });
  const j = await r.json(); b.disabled = false;
  if (!r.ok) { toast(j.error || 'Lỗi ' + r.status); return; }
  toast(`Đã duyệt ${approve.length} case vào "${name}"`);
  b.closest('.fb').innerHTML = `<b>Đã duyệt ✓</b> Suite "${esc(name)}" có ${j.suite.cases.length} case. ${precisionLine(j.precision)}`;
});
function suiteName(){ const n = $('s-name').value.trim(); if (!n) toast('Nhập tên suite'); return n; }
$('s-run').onclick = async () => {
  const n = suiteName(); if (!n) return;
  const j = await api('/suites/' + encodeURIComponent(n) + '/run', { method: 'POST', btn: $('s-run'), title: 'Đang chạy suite…', sub: 'Chạy xác định, không gọi AI.', body: { base_url: $('s-url').value.trim() } });
  if (!j || !j.gate) return;
  $('out').innerHTML = gateBanner(j.gate) + `<div class="card"><h2>Kết quả (${(j.cases || []).length} case đã duyệt)</h2>${casesTable(j.cases, j.report, j.trust)}</div>` + rawBlock('JSON thô', j);
};
$('s-show').onclick = async () => {
  const n = suiteName(); if (!n) return;
  const j = await api('/suites/' + encodeURIComponent(n), { btn: $('s-show'), title: 'Đang tải suite…', sub: '' }); if (!j || !j.suite) return;
  const cs = j.suite.cases || [];
  $('out').innerHTML = `<div class="card"><h2>${esc(j.suite.name)} — ${cs.length} case đã duyệt</h2>${precisionLine(j.precision)}
    ${cs.length ? `<div class="tw"><table><thead><tr><th>Case</th><th>Nguồn</th><th>Duyệt lúc</th><th></th></tr></thead><tbody>${cs.map(a => `<tr><td>${methodTag(a.case.method)} <span class="mono">${esc(a.case.path)}</span><div>${esc(a.case.name)}</div></td><td>${a.from_ai ? 'AI + người duyệt' : 'Người viết'}</td><td>${esc(new Date(a.approved_at).toLocaleString())}</td><td><button class="btn sm bd" data-rm="${esc(a.hash)}" data-suite="${esc(j.suite.name)}">Xoá</button></td></tr>`).join('')}</tbody></table></div>` : '<div class="empty">Suite trống.</div>'}</div>`;
};
document.addEventListener('click', async e => {
  const b = e.target.closest('[data-rm]'); if (!b || !confirm('Xoá case này khỏi suite?')) return;
  const r = await fetch(API + '/suites/' + encodeURIComponent(b.dataset.suite) + '/cases/' + b.dataset.rm, { method: 'DELETE', headers: headers(false) });
  if (r.ok) { toast('Đã xoá'); $('s-show').click(); } else toast('Lỗi ' + r.status);
});

$('sc-run').onclick = async () => {
  const paths = $('sc-paths').value.split('\n').map(x => x.trim()).filter(Boolean);
  const j = await api('/qa/scan', { method: 'POST', btn: $('sc-run'), title: 'Đang quét sản phẩm…', sub: 'Kiểm tra bảo mật và duyệt qua các trang.',
    body: { base_url: $('sc-url').value.trim(), protected_paths: paths, max_pages: +$('sc-pages').value, unreviewed: $('sc-un').value } });
  if (!j || !j.gate) return;
  $('out').innerHTML = gateBanner(j.gate) + securityCard(j.security) + webCard(j.web) + scopeCard(j.scope) + rawBlock('JSON thô', j);
};

const jsonField = (id, def) => { const v = $(id).value.trim(); if (!v) return def; try { return JSON.parse(v); } catch { throw new Error('JSON không hợp lệ ở ô: ' + id); } };
$('adv-kind').onchange = () => {
  document.querySelectorAll('.adv').forEach(d => d.hidden = d.dataset.k !== $('adv-kind').value);
  const k = $('adv-kind').value;
  $('adv-actions').hidden = k === 'jobs' || k === 'mdevices'; $('out').innerHTML = '';
  if (k === 'jobs') showJobs();
  if (k === 'mdevices') showDevices();
};

async function pollJob(id, title){
  $('loadtitle').textContent = title; $('loader').classList.add('on');
  for (;;) {
    let j; try { const r = await fetch(API + '/jobs/' + id, { headers: headers(false) }); j = await r.json(); if (!r.ok) throw new Error(j.error); } catch (e) { $('loader').classList.remove('on'); return { status: 'failed', error: e.message }; }
    if (j.status === 'running') { $('loadsub').textContent = j.progress || 'Đang chạy…'; await new Promise(r => setTimeout(r, 2500)); continue; }
    $('loader').classList.remove('on'); return j;
  }
}
async function startJob(path, body, btn, title){
  const j = await api(path, { method: 'POST', ...(body instanceof FormData ? { form: body } : { body }), btn, title, sub: 'Đang tạo job…' });
  if (!j) return null;
  if (!j.job_id) return j;
  $('out').innerHTML = `<div class="banner info"><div>Job <span class="mono">${esc(j.job_id)}</span> đang chạy nền. Bạn có thể đóng trang; xem lại ở mục “Các job đang chạy”.</div></div>`;
  const done = await pollJob(j.job_id, title);
  if (done.status !== 'done') { $('out').innerHTML = alertBox('bad', 'Job ' + done.status, done.error || ''); return null; }
  return done.result;
}
const findingsTable = (rows, cols) => rows.length ? `<div class="tw"><table><thead><tr>${cols.map(c => `<th>${c[0]}</th>`).join('')}</tr></thead><tbody>${rows.map(r => `<tr>${cols.map(c => `<td>${c[1](r)}</td>`).join('')}</tr>`).join('')}</tbody></table></div>` : '<div class="empty">Không phát hiện vấn đề.</div>';
const sevCols = [['Mức', f => sevPill(f.severity)], ['Mã', f => `<span class="mono">${esc(f.id)}</span>`], ['Phát hiện', f => `<b>${esc(f.title)}</b>${f.endpoint ? `<div class="mono muted">${esc(f.endpoint)}</div>` : ''}${f.detail ? `<div class="muted">${esc(f.detail)}</div>` : ''}`]];

async function shotImg(name, label){
  try { const r = await fetch(API + '/artifacts/' + encodeURIComponent(name), { headers: headers(false) }); if (!r.ok) return ''; const u = URL.createObjectURL(await r.blob());
    return `<figure style="margin:0"><img src="${u}" alt="${esc(label)}" style="max-width:100%;border:1px solid var(--bd);border-radius:8px"><figcaption class="muted" style="font-size:12px">${esc(label)}</figcaption></figure>`; } catch { return ''; }
}
async function browserCard(rep){
  const rows = [];
  for (const r of rep.results || []) {
    const shot = r.screenshot ? await shotImg(r.screenshot, `${r.viewport}${r.locale ? ' ' + r.locale : ''}`) : '';
    const diff = r.diff_image ? await shotImg(r.diff_image, 'khác biệt so với ảnh chuẩn (đỏ)') : '';
    rows.push(`<details><summary><span><span class="mono">${esc(r.page)}</span> <span class="pill mut">${esc(r.viewport)}${r.locale ? ' · ' + esc(r.locale) : ''}</span>${r.baseline ? ` <span class="pill mut">${esc(r.baseline)}</span>` : ''}${r.visual_diff_pct != null ? ` <span class="pill ${r.visual_diff_pct > 0.5 ? 'bad' : 'ok'}">${r.visual_diff_pct.toFixed(2)}% khác</span>` : ''}</span><span>${(r.issues || []).length ? `<span class="pill warn">${r.issues.length} vấn đề</span>` : '<span class="pill ok">OK</span>'}</span></summary><div class="in">
      ${(r.issues || []).length ? `<div class="tw"><table><tbody>${r.issues.map(i => `<tr><td>${sevPill(i.severity)}</td><td class="mono">${esc(i.rule)}</td><td>${esc(i.detail)}</td></tr>`).join('')}</tbody></table></div>` : '<span class="muted">Không có vấn đề.</span>'}
      ${r.tab_stops ? `<p class="muted">Bàn phím: Tab đi qua ${r.tab_stops} phần tử.</p>` : ''}
      <div class="grid g2" style="margin-top:10px">${shot}${diff}</div></div></details>`);
  }
  return `<div class="card"><h2>Kết quả trình duyệt</h2><div class="grid g4" style="margin-bottom:12px">${tile((rep.results || []).length, 'lượt kiểm tra')}${tile(rep.error_count, 'lỗi', rep.error_count ? 'bad-t' : '')}${tile(rep.warning_count, 'cảnh báo')}</div>${rows.join('')}</div>`;
}
const dbCard = r => `<div class="card"><h2>Cơ sở dữ liệu <span class="pill mut">${esc(r.dialect)}</span> <span class="muted" style="font-weight:400">${r.latency_ms} ms</span></h2>${findingsTable(r.findings || [], sevCols)}
  ${(r.assertions || []).length ? `<h3 style="margin-top:14px">Quy tắc nghiệp vụ</h3>${findingsTable(r.assertions, [['', a => pill(a.passed)], ['Quy tắc', a => esc(a.name)], ['Chi tiết', a => `<span class="muted">${esc(a.detail || '')}</span>`]])}` : ''}</div>`;
const asyncCard = r => `<div class="card"><h2>Luồng bất đồng bộ</h2><div class="grid g4">${tile(r.settled ? (r.settle_ms / 1000).toFixed(2) + ' s' : '–', 'thời gian tới khi nhất quán')}${tile(r.polls, 'lần kiểm tra')}${tile(r.trigger_status, 'mã kích hoạt')}${tile((r.replay_statuses || []).join(', ') || '–', 'mã khi gửi lặp')}</div>${(r.failures || []).length ? `<ul class="fails">${r.failures.map(f => `<li>${esc(f)}</li>`).join('')}</ul>` : ''}</div>`;
function stressCard(r){
  const max = Math.max(1, ...r.stages.map(s => s.stats.p95_ms));
  return `<div class="card"><h2>Stress</h2><div class="grid g4" style="margin-bottom:12px">${tile(r.max_sustainable_concurrency, 'mức tải chịu được')}${tile(r.breaking_concurrency || '–', 'điểm gãy', r.breaking_concurrency ? 'bad-t' : '')}${tile(fmt(r.peak_rps, 0), 'RPS cao nhất')}</div>
  <div class="tw"><table><thead><tr><th>Concurrency</th><th>RPS</th><th>p95</th><th></th><th>Lỗi</th><th>Vi phạm</th></tr></thead><tbody>${r.stages.map(s => `<tr><td>${s.concurrency}</td><td>${fmt(s.stats.rps, 0)}</td><td>${fmt(s.stats.p95_ms)} ms</td><td style="min-width:120px"><div class="bar"><i class="${s.violations ? 'over' : ''}" style="width:${100 * s.stats.p95_ms / max}%"></i></div></td><td>${s.stats.failed}/${s.stats.total_requests}</td><td>${(s.violations || []).map(esc).join('; ') || pill(true, 'OK')}</td></tr>`).join('')}</tbody></table></div></div>`;
}
function soakCard(r){
  const w = r.windows || []; if (w.length < 2) return `<div class="card"><h2>Soak</h2><div class="empty">Chưa đủ dữ liệu.</div></div>`;
  const W = 820, H = 200, P = 34, mx = Math.max(1, ...w.map(x => x.p95_ms)), x = i => P + i * (W - 2 * P) / (w.length - 1), y = v => H - P - v / mx * (H - 2 * P);
  return `<div class="card"><h2>Soak</h2><p class="${r.degraded ? '' : 'muted'}" style="margin-top:0;${r.degraded ? 'color:var(--bad);font-weight:600' : ''}">${r.degraded ? '⚠ ' + esc(r.reason) : 'Không phát hiện suy giảm.'}</p>
   <svg class="chart" viewBox="0 0 ${W} ${H}" width="100%" role="img" aria-label="p95 theo thời gian">${[0, .5, 1].map(v => `<line x1="${P}" x2="${W - P}" y1="${y(v * mx)}" y2="${y(v * mx)}" stroke="var(--bd)"/><text x="2" y="${y(v * mx) + 3}">${(v * mx).toFixed(0)}ms</text>`).join('')}
   <polyline fill="none" stroke="var(--ac)" stroke-width="2.4" points="${w.map((p, i) => x(i) + ',' + y(p.p95_ms)).join(' ')}"/>${w.map((p, i) => `<circle cx="${x(i)}" cy="${y(p.p95_ms)}" r="3" fill="var(--ac)"><title>${p.at_sec}s: p95 ${p.p95_ms}ms, ${p.rps} rps, lỗi ${p.error_pct}%</title></circle>`).join('')}</svg>
   <div class="grid g4" style="margin-top:8px">${tile(fmt(r.drift.first_quarter_p95_ms) + ' ms', 'p95 đầu')}${tile(fmt(r.drift.last_quarter_p95_ms) + ' ms', 'p95 cuối')}${tile((r.drift.p95_change_pct > 0 ? '+' : '') + fmt(r.drift.p95_change_pct, 0) + '%', 'thay đổi', r.degraded ? 'bad-t' : '')}</div></div>`;
}
async function showJobs(){
  const j = await api('/jobs', { title: 'Đang tải…', sub: '' }); if (!j) return;
  $('out').innerHTML = `<div class="card"><h2>Job (giữ trong bộ nhớ, mất khi service khởi động lại)</h2>${(j.jobs || []).length ? `<div class="tw"><table><thead><tr><th>ID</th><th>Loại</th><th>Trạng thái</th><th>Tiến độ</th><th>Bắt đầu</th><th></th></tr></thead><tbody>${j.jobs.map(x => `<tr><td class="mono">${esc(x.id)}</td><td>${esc(x.kind)}</td><td>${pill(x.status === 'done', 'xong', esc(x.status))}</td><td class="muted">${esc(x.progress || '')}</td><td>${esc(new Date(x.started_at).toLocaleString())}</td><td>${x.status === 'running' ? `<button class="btn sm bd" data-cancel="${esc(x.id)}">Huỷ</button>` : `<button class="btn sm" data-view="${esc(x.id)}">Xem</button>`}</td></tr>`).join('')}</tbody></table></div>` : '<div class="empty">Chưa có job nào.</div>'}</div>`;
}
document.addEventListener('click', async e => {
  const c = e.target.closest('[data-cancel]'), v = e.target.closest('[data-view]');
  if (c) { await fetch(API + '/jobs/' + c.dataset.cancel + '/cancel', { method: 'POST', headers: headers(false) }); toast('Đã huỷ'); showJobs(); }
  if (v) { const r = await fetch(API + '/jobs/' + v.dataset.view, { headers: headers(false) }); const j = await r.json(); $('out').innerHTML = j.result ? await renderAdv(j.kind, j.result) + rawBlock('JSON thô', j) : alertBox('bad', j.status, j.error || ''); }
});
function appInfoTiles(rep){
  if (rep.android) { const a = rep.android; return `<div class="grid g4">${tile(a.package, 'package')}${tile(a.version_name + ' (' + a.version_code + ')', 'phiên bản')}${tile(a.min_sdk + ' → ' + a.target_sdk, 'min → target SDK')}${tile((a.permissions || []).length, 'quyền')}${tile((a.components || []).length, 'component')}${tile(a.size_mb + ' MB', 'kích thước')}${tile(a.signed ? (a.signed_v2_or_later ? 'v2+' : 'v1') : 'không', 'chữ ký', a.signed ? '' : 'bad-t')}${tile((a.native_abis || []).join(', ') || '–', 'ABI native')}</div>`; }
  const i = rep.ios; return `<div class="grid g4">${tile(i.bundle_id, 'bundle id')}${tile(i.version + ' (' + i.build + ')', 'phiên bản')}${tile(i.minimum_os || '–', 'iOS tối thiểu')}${tile((i.usage_descriptions || []).length, 'mô tả quyền')}${tile((i.detected_sensitive_apis || []).length, 'API nhạy cảm dùng')}${tile(i.provisioning || '–', 'provisioning')}${tile(i.size_mb + ' MB', 'kích thước')}${tile((i.embedded_frameworks || []).length, 'framework')}</div>`;
}
async function scenarioTable(list){
  const rows = [];
  for (const s of list || []) {
    const shot = s.screenshot ? await shotImg(s.screenshot, s.name) : '';
    rows.push(`<tr><td>${esc(s.name)}</td><td>${pill(s.passed)}</td><td>${(s.problems || []).length ? `<ul class="fails">${s.problems.map(p => `<li>${esc(p)}</li>`).join('')}</ul>` : ''}</td><td style="width:140px">${shot}</td></tr>`);
  }
  return rows.length ? `<h3 style="margin-top:14px">Kịch bản</h3><div class="tw"><table><thead><tr><th>Kịch bản</th><th>Kết quả</th><th>Vấn đề</th><th>Ảnh</th></tr></thead><tbody>${rows.join('')}</tbody></table></div>` : '';
}
async function androidCard(r){
  const main = r.screenshot ? await shotImg(r.screenshot, 'màn hình chính') : '';
  const dev = r.device || {};
  return `<div class="card"><h2>Android <span class="muted" style="font-weight:400">${esc(dev.model || '')} · Android ${esc(dev.android_version || '')} (SDK ${dev.sdk || '?'})</span></h2>
    <div class="grid g4">${tile((r.startup.median_ms || 0) + ' ms', 'khởi động nguội (trung vị)')}${tile((r.memory.pss_mb || 0) + ' MB', 'bộ nhớ PSS')}${tile(r.frames ? fmt(r.frames.janky_pct, 1) + '%' : '–', 'khung hình giật')}${tile(r.monkey ? r.monkey.events_injected : '–', 'sự kiện monkey', r.monkey && r.monkey.crashed ? 'bad-t' : '')}${tile((r.crashes || []).length, 'crash / ANR', (r.crashes || []).length ? 'bad-t' : '')}${tile(r.ui_clickable_controls, 'điều khiển bấm được')}</div>
    ${(r.crashes || []).length ? `<h3 style="margin-top:14px">Crash</h3>${(r.crashes).map(c => `<div class="issue bug"><b>${esc(c.kind)}</b> ${esc(c.summary)}${(c.stack || []).length ? `<pre style="margin-top:6px">${esc(c.stack.join('\n'))}</pre>` : ''}</div>`).join('')}` : ''}
    ${(r.ui_issues || []).length ? `<h3 style="margin-top:14px">Truy cập được &amp; vùng chạm</h3>${findingsTable(r.ui_issues, [['Luật', u => `<span class="mono">${esc(u.rule)}</span>`], ['Chi tiết', u => esc(u.detail)]])}` : ''}
    ${await scenarioTable(r.scenarios)}${main ? `<h3 style="margin-top:14px">Màn hình chính</h3><div style="max-width:260px">${main}</div>` : ''}</div>`;
}
async function iosCard(r){
  const main = r.screenshot ? await shotImg(r.screenshot, 'sau khi khởi chạy') : '';
  return `<div class="card"><h2>iOS <span class="muted" style="font-weight:400">${esc(r.device.name)} · ${esc(r.device.runtime)}</span></h2>
    <div class="grid g4">${tile(r.pid || '–', 'PID')}${tile((r.crashes || []).length, 'crash', (r.crashes || []).length ? 'bad-t' : '')}${tile(r.log_faults, 'log mức fault')}${tile((r.scenarios || []).length, 'kịch bản')}</div>
    ${(r.crashes || []).length ? `<h3 style="margin-top:14px">Crash</h3>${r.crashes.map(c => `<div class="issue bug"><b>${esc(c.file)}</b> ${esc(c.summary)}</div>`).join('')}` : ''}
    ${await scenarioTable(r.scenarios)}${main ? `<h3 style="margin-top:14px">Sau khi khởi chạy</h3><div style="max-width:260px">${main}</div>` : ''}</div>`;
}
async function showDevices(){
  const j = await api('/qa/mobile/devices', { title: 'Đang tìm thiết bị…', sub: '' }); if (!j) return;
  const a = j.android || {}, i = j.ios || {};
  $('out').innerHTML = `<div class="card"><h2>Android (adb)</h2>${a.available ? findingsTable(a.devices || [], [['Serial', d => `<span class="mono">${esc(d.serial)}</span>`], ['Trạng thái', d => pill(d.state === 'device', 'sẵn sàng', esc(d.state))], ['Model', d => esc(d.model || '')]]) : `<div class="banner info"><div>${esc(a.error || 'Không khả dụng')}</div></div>`}</div>
    <div class="card"><h2>iOS (simulator)</h2>${i.available ? findingsTable(i.simulators || [], [['Tên', d => esc(d.name)], ['Runtime', d => esc(d.runtime)], ['Trạng thái', d => pill(d.state === 'Booted', 'đang chạy', esc(d.state))], ['UDID', d => `<span class="mono">${esc(d.udid)}</span>`]]) : `<div class="banner info"><div>${esc(i.error || 'Không khả dụng')}</div></div>`}</div>`;
}
async function renderAdv(kind, res){
  const g = res.gate ? gateBanner(res.gate) : '';
  switch (kind) {
    case 'browser': return g + await browserCard(res.report);
    case 'database': return g + dbCard(res.report);
    case 'async-flow': return g + asyncCard(res.result);
    case 'queue': return g + `<div class="card"><h2>Hàng đợi</h2>${findingsTable(res.findings || [], sevCols)}</div>`;
    case 'active-security': return g + `<div class="card"><h2>Bảo mật chủ động <span class="muted" style="font-weight:400">${res.result.requests_sent} request</span></h2>${findingsTable(res.result.findings || [], sevCols)}${(res.result.skipped || []).length ? `<h3 style="margin-top:14px">Đã bỏ qua</h3>${list(res.result.skipped)}` : ''}</div>`;
    case 'mobile-static': return g + `<div class="card"><h2>${res.report.platform === 'android' ? 'Android APK' : 'iOS IPA'} (phân tích tĩnh)</h2>${appInfoTiles(res.report)}</div><div class="card"><h2>Phát hiện</h2>${findingsTable(res.report.findings || [], sevCols)}</div>`;
    case 'android': return g + await androidCard(res.report) + (res.static ? `<div class="card"><h2>Phân tích tĩnh APK</h2>${appInfoTiles(res.static)}${findingsTable(res.static.findings || [], sevCols)}</div>` : '') + `<div class="card"><h2>Phát hiện khi chạy</h2>${findingsTable(res.report.findings || [], sevCols)}</div>`;
    case 'ios': return g + await iosCard(res.report) + `<div class="card"><h2>Phát hiện</h2>${findingsTable(res.report.findings || [], sevCols)}</div>`;
    case 'stress': return g + stressCard(res.report);
    case 'soak': return g + soakCard(res.report);
  }
  return rawBlock('JSON thô', res);
}
$('adv-run').onclick = async () => {
  const kind = $('adv-kind').value, btn = $('adv-run'); let res, name;
  try {
    if (kind === 'browser') {
      name = 'browser'; const vps = [...document.querySelectorAll('.br-vp:checked')].map(x => x.value);
      res = await startJob('/qa/browser?async=true', { base_url: $('br-url').value.trim(), max_pages: +$('br-pages').value, viewports: vps, locales: $('br-loc').value.split(',').map(x => x.trim()).filter(Boolean), name: $('br-name').value, update_baseline: $('br-upd').checked }, btn, 'Đang kiểm thử trình duyệt…');
    } else if (kind === 'db') {
      name = 'database'; res = await api('/qa/database', { method: 'POST', btn, title: 'Đang kiểm tra cơ sở dữ liệu…', body: { connection: $('db-conn').value.trim(), assertions: jsonField('db-asserts', []) } });
    } else if (kind === 'async') {
      name = 'async-flow'; const [m, ...p] = $('as-tm').value.trim().split(/\s+/), rep = +$('as-rep').value;
      res = await api('/qa/async', { method: 'POST', btn, title: 'Đang chạy luồng bất đồng bộ…', body: { base_url: $('as-url').value.trim(), flow: { trigger: { method: m, path: p.join(' '), body: jsonField('as-body', null) }, poll: { path: $('as-poll').value.trim(), expect_json: jsonField('as-exp', {}) }, timeout_sec: +$('as-to').value, idempotency: rep ? { repeat: rep } : undefined } } });
    } else if (kind === 'queue') {
      name = 'queue'; res = await api('/qa/queue', { method: 'POST', btn, title: 'Đang kiểm tra hàng đợi…', body: { broker: $('q-broker').value.trim(), max_ready: +$('q-max').value } });
    } else if (kind === 'active') {
      if (!$('ac-consent').checked) { toast('Cần xác nhận quyền sở hữu hệ thống'); return; }
      name = 'active-security';
      res = await startJob('/qa/active?async=true', { base_url: $('ac-url').value.trim(), active: true, endpoints: jsonField('ac-eps', []), identities: jsonField('ac-ids', []), resources: jsonField('ac-res', []), include_writes: $('ac-writes').checked, explore_forms: $('ac-forms').checked }, btn, 'Đang kiểm thử bảo mật chủ động…');
    } else if (kind === 'mstatic') {
      const f = $('ms-file').files[0]; if (!f) { toast('Chọn file .apk hoặc .ipa'); return; }
      name = 'mobile-static'; const fd = new FormData(); fd.append('file', f);
      res = await api('/qa/mobile/static', { method: 'POST', form: fd, btn, title: 'Đang phân tích ứng dụng…', sub: 'Đọc manifest, quét bí mật…' });
    } else if (kind === 'android') {
      name = 'android'; const fd = new FormData(); const f = $('an-file').files[0]; if (f) fd.append('file', f);
      fd.append('package', $('an-pkg').value.trim()); fd.append('serial', $('an-serial').value.trim()); fd.append('monkey_events', $('an-monkey').value); fd.append('seed', $('an-seed').value); fd.append('uninstall', $('an-un').checked);
      res = await startJob('/qa/mobile/android?async=true', fd, btn, 'Đang kiểm thử Android…');
    } else if (kind === 'ios') {
      name = 'ios'; const fd = new FormData(); const f = $('io-file').files[0]; if (f) fd.append('file', f);
      fd.append('bundle_id', $('io-bundle').value.trim()); fd.append('device', $('io-dev').value.trim()); fd.append('deep_links', $('io-links').value);
      res = await startJob('/qa/mobile/ios?async=true', fd, btn, 'Đang kiểm thử iOS…');
    } else if (kind === 'load') {
      const stress = $('ld-mode').value === 'stress'; name = stress ? 'stress' : 'soak';
      const body = { url: $('ld-url').value.trim(), max_p95_ms: +$('ld-p95').value, concurrency: +$('ld-conc').value, duration_sec: +$('ld-dur').value, stages: $('ld-stages').value.split(',').map(x => +x.trim()).filter(Boolean) };
      res = await startJob(stress ? '/jobs/stress' : '/jobs/soak', body, btn, stress ? 'Đang stress test…' : 'Đang soak test…');
    }
  } catch (e) { toast(e.message); return; }
  if (res && name) $('out').innerHTML = await renderAdv(name, res) + rawBlock('JSON thô', res);
};

async function loadModelChip(){
  try { const r = await fetch('/health'); const j = await r.json(); $('modelchip').innerHTML = 'Model: <b>' + esc(j.model || '?') + '</b>'; } catch {}
}
$('l-stats').onclick = async () => {
  const j = await api('/learning/stats', { btn: $('l-stats'), title: 'Đang tải thống kê…', sub: '' }); if (!j) return;
  const models = Object.entries(j.by_model || {}), human = j.by_model_human_only || {};
  $('out').innerHTML = `<div class="grid g4" style="margin-bottom:16px">${tile(Math.round((j.approval_rate || 0) * 100) + '%', 'approval rate')}${tile(j.total, 'tổng kết quả')}${tile(j.good, 'tốt')}${tile(j.bad, 'chưa tốt')}${tile(j.pending, 'chờ duyệt')}${tile(j.feedback_since_reflect, 'feedback mới')}</div>`
    + `<div class="card"><h2>Theo model</h2>${models.length ? `<div class="tw"><table><thead><tr><th>Model</th><th>Tốt</th><th>Chưa tốt</th><th>Tỉ lệ</th><th>Chỉ người duyệt</th></tr></thead><tbody>${models.map(([m, x]) => `<tr><td class="mono">${esc(m || '(chưa ghi)')} ${m === j.active_model ? '<span class="pill ok">đang dùng</span>' : ''}</td><td>${x.good}</td><td>${x.bad}</td><td>${Math.round(x.rate * 100)}%</td><td>${human[m] ? Math.round(human[m].rate * 100) + '% (' + (human[m].good + human[m].bad) + ')' : '–'}</td></tr>`).join('')}</tbody></table></div>` : '<div class="empty">Chưa có dữ liệu.</div>'}</div>`
    + `<div class="card"><h2>Playbook hiện tại</h2>${j.playbook ? `<div style="white-space:pre-wrap">${esc(j.playbook)}</div>` : '<span class="muted">Chưa có — bấm “Cập nhật playbook” sau khi có vài đánh giá.</span>'}</div>`;
};
$('l-hist').onclick = async () => {
  const j = await api('/learning/history', { btn: $('l-hist'), title: 'Đang tải lịch sử…', sub: '' }); if (!j) return;
  const h = j.history || [];
  if (h.length < 2) { $('out').innerHTML = '<div class="card"><div class="empty">Chưa đủ dữ liệu (cần ít nhất 2 mốc; mỗi mốc lưu theo chu kỳ tự học hoặc khi có feedback).</div></div>'; return; }
  const W = 820, H = 240, P = 34, x = i => P + i * (W - 2 * P) / (h.length - 1), y = v => H - P - v * (H - 2 * P);
  const line = (k, c) => `<polyline fill="none" stroke="${c}" stroke-width="2.4" stroke-linejoin="round" stroke-linecap="round" points="${h.map((p, i) => x(i) + ',' + y(p[k] || 0)).join(' ')}"/>` + h.map((p, i) => `<circle cx="${x(i)}" cy="${y(p[k] || 0)}" r="3" fill="${c}"><title>${esc(new Date(p.at).toLocaleString())}: ${Math.round((p[k] || 0) * 100)}% (${esc(p.model)})</title></circle>`).join('');
  const last = h[h.length - 1];
  $('out').innerHTML = `<div class="card"><h2>Approval rate theo thời gian</h2><svg class="chart" viewBox="0 0 ${W} ${H}" width="100%" role="img" aria-label="Approval rate theo thời gian">`
    + [0, .25, .5, .75, 1].map(v => `<line x1="${P}" x2="${W - P}" y1="${y(v)}" y2="${y(v)}" stroke="var(--bd)"/><text x="4" y="${y(v) + 3}">${Math.round(v * 100)}%</text>`).join('')
    + line('approval_rate', 'var(--ac)') + line('human_rate', 'var(--ok)') + '</svg>'
    + `<div class="legend"><span><i style="background:var(--ac)"></i>Tổng</span><span><i style="background:var(--ok)"></i>Người duyệt</span></div>
       <p class="muted" style="margin-bottom:0">Mốc cuối: <b>${esc(last.model || 'base')}</b> — ${(last.approval_rate * 100).toFixed(1)}% trên ${last.reviews} đánh giá.</p></div>`;
};
$('l-pend').onclick = async () => {
  const j = await api('/learning/pending', { btn: $('l-pend'), title: 'Đang tải hàng chờ…', sub: '' }); if (!j) return;
  const it = j.items || [];
  $('out').innerHTML = it.length ? it.map(e => `<div class="card"><div style="display:flex;justify-content:space-between;gap:8px;flex-wrap:wrap"><b>${esc(e.kind)}</b><span class="pill warn">judge ${e.score !== undefined ? Math.round(e.score * 100) + '%' : '?'}</span></div>
      ${e.note ? `<p class="muted">${esc(e.note)}</p>` : ''}<h3>Đầu vào</h3><pre>${esc((e.input || '').slice(0, 500))}</pre><h3 style="margin-top:10px">Kết quả AI</h3><pre>${esc(String(typeof e.output === 'string' ? e.output : JSON.stringify(e.output)).slice(0, 700))}</pre>${feedbackBox(e.id)}</div>`).join('')
    : '<div class="card"><div class="empty">🎉 Không có mục nào cần duyệt.</div></div>';
};
$('l-ref').onclick = async () => { const j = await api('/learning/reflect', { method: 'POST', btn: $('l-ref'), title: 'Đang cập nhật playbook…' }); if (j?.playbook !== undefined) $('out').innerHTML = `<div class="card"><h2>Playbook mới</h2><div style="white-space:pre-wrap">${esc(j.playbook)}</div></div>`; };
$('l-evo').onclick = async () => { const j = await api('/learning/evolve', { method: 'POST', btn: $('l-evo'), title: 'Đang tạo model mới…' }); if (j?.model) { toast('Đã chuyển sang ' + j.model); $('out').innerHTML = `<div class="card"><h2>Model mới</h2><dl class="kv"><dt>Model</dt><dd class="mono">${esc(j.model)}</dd><dt>Dựa trên</dt><dd class="mono">${esc(j.base)}</dd><dt>Ví dụ nhúng</dt><dd>${j.examples}</dd></dl></div>`; loadModelChip(); } };
$('l-exp').onclick = async () => {
  const r = await fetch(API + '/learning/export', { headers: headers(false) });
  if (!r.ok) { toast('Lỗi ' + r.status); return; }
  const a = document.createElement('a'); a.href = URL.createObjectURL(await r.blob()); a.download = 'qc-finetune.jsonl'; a.click();
};

showTab(TABS[store.get('tab')] ? store.get('tab') : 'task');
loadModelChip();
