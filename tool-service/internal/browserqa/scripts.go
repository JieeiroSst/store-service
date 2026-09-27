package browserqa

const pageAudit = `(function(mobile){
  const out = {overflow: [], contrast: [], taps: [], leaks: [], text: '', lang: document.documentElement.lang || '', dir: getComputedStyle(document.documentElement).direction, clickNoKbd: [], posTab: 0, innerWidth: document.documentElement.clientWidth};
  const sel = el => { if (el.id) return el.tagName.toLowerCase()+'#'+el.id; let s = el.tagName.toLowerCase(); if (el.className && typeof el.className === 'string') s += '.'+el.className.trim().split(/\s+/).slice(0,2).join('.'); return s; };
  const visible = el => { const r = el.getBoundingClientRect(), cs = getComputedStyle(el); return r.width > 0 && r.height > 0 && cs.visibility !== 'hidden' && cs.display !== 'none' && parseFloat(cs.opacity) > 0.01; };

  const vw = document.documentElement.clientWidth;
  if (document.documentElement.scrollWidth > vw + 1) {
    for (const el of document.querySelectorAll('body *')) {
      const r = el.getBoundingClientRect();
      if (r.right > vw + 1 && r.width > 0 && visible(el)) { out.overflow.push(sel(el)+' extends to '+Math.round(r.right)+'px (viewport '+vw+'px)'); if (out.overflow.length >= 3) break; }
    }
    if (!out.overflow.length) out.overflow.push('page is wider than the viewport ('+document.documentElement.scrollWidth+'px > '+vw+'px)');
  }

  const parse = c => { const m = c.match(/^rgba?\(([^)]+)\)$/); if (!m) return null; const p = m[1].split(/[ ,\/]+/).filter(Boolean).map(Number); return {r:p[0], g:p[1], b:p[2], a: p.length > 3 ? p[3] : 1}; };
  const lin = v => { v /= 255; return v <= 0.03928 ? v/12.92 : Math.pow((v+0.055)/1.055, 2.4); };
  const lum = c => 0.2126*lin(c.r) + 0.7152*lin(c.g) + 0.0722*lin(c.b);
  const over = (fg, bg) => ({r: fg.r*fg.a + bg.r*(1-fg.a), g: fg.g*fg.a + bg.g*(1-fg.a), b: fg.b*fg.a + bg.b*(1-fg.a), a: 1});
  const bgOf = el => { let stack = []; for (let e = el; e; e = e.parentElement) { const cs = getComputedStyle(e); if (cs.backgroundImage !== 'none') return null; const c = parse(cs.backgroundColor); if (c && c.a > 0) { stack.push(c); if (c.a >= 0.99) break; } } let bg = {r:255,g:255,b:255,a:1}; for (let i = stack.length-1; i >= 0; i--) bg = over(stack[i], bg); return bg; };
  const seen = new Set(); let checked = 0;
  const walker = document.createTreeWalker(document.body, NodeFilter.SHOW_TEXT);
  while (walker.nextNode() && checked < 400) {
    const n = walker.currentNode, el = n.parentElement;
    if (!el || !n.textContent.trim() || ['SCRIPT','STYLE','NOSCRIPT'].includes(el.tagName) || !visible(el) || el.closest('[disabled],[aria-disabled=true]')) continue;
    checked++;
    const cs = getComputedStyle(el), fg = parse(cs.color), bg = bgOf(el);
    if (!fg || !bg) continue;
    const f = over(fg, bg), l1 = lum(f), l2 = lum(bg), ratio = (Math.max(l1,l2)+0.05)/(Math.min(l1,l2)+0.05);
    const size = parseFloat(cs.fontSize), bold = parseInt(cs.fontWeight) >= 700, large = size >= 24 || (size >= 18.66 && bold), need = large ? 3 : 4.5;
    if (ratio < need) { const k = sel(el)+cs.color+cs.backgroundColor; if (!seen.has(k) && out.contrast.length < 15) { seen.add(k); out.contrast.push(sel(el)+' "'+n.textContent.trim().slice(0,30)+'" ratio '+ratio.toFixed(2)+' (needs '+need+')'); } }
  }

  if (mobile) {
    for (const el of document.querySelectorAll('a[href],button,input:not([type=hidden]),select,textarea,[role=button],[onclick]')) {
      if (!visible(el)) continue; const r = el.getBoundingClientRect();
      if (Math.min(r.width, r.height) < 24 && out.taps.length < 10) out.taps.push(sel(el)+' is '+Math.round(r.width)+'x'+Math.round(r.height)+'px');
    }
  }

  const nativeFocusable = 'a[href],button,input,select,textarea,summary,[tabindex],[contenteditable=true],iframe,audio[controls],video[controls]';
  for (const el of document.querySelectorAll('body *')) {
    if (!visible(el)) continue;
    const clicky = el.hasAttribute('onclick') || el.getAttribute('role') === 'button' || getComputedStyle(el).cursor === 'pointer';
    if (clicky && !el.matches(nativeFocusable) && !el.closest(nativeFocusable) && !el.querySelector(nativeFocusable) && out.clickNoKbd.length < 10) out.clickNoKbd.push(sel(el)+' "'+(el.textContent||'').trim().slice(0,25)+'"');
    const ti = el.getAttribute('tabindex'); if (ti && parseInt(ti) > 0) out.posTab++;
  }

  const text = document.body.innerText || ''; out.text = text.slice(0, 200000);
  const leaks = [/\{\{[^}]{1,40}\}\}/, /\bundefined\b/, /\bNaN\b/, /\[object Object\]/, /translation[ _.-]?missing|missing[ _.-]translation/i, /\b__[A-Z][A-Z_]{2,}__\b/, /%[a-z]{3,}%/i, /\bi18n\.[a-z_.]+/i];
  for (const re of leaks) { const m = text.match(re); if (m) out.leaks.push(m[0]); }
  return JSON.stringify(out);
})`

const focusProbe = `(function(){
  const e = document.activeElement; if (!e || e === document.body || e === document.documentElement) return JSON.stringify({none:true});
  const cs = getComputedStyle(e), r = e.getBoundingClientRect();
  const path = []; for (let n = e; n && n.nodeType === 1 && path.length < 4; n = n.parentElement) { let s = n.tagName.toLowerCase(); if (n.id) s += '#'+n.id; path.unshift(s); }
  const indicator = (cs.outlineStyle !== 'none' && parseFloat(cs.outlineWidth) > 0) || cs.boxShadow !== 'none';
  return JSON.stringify({key: path.join('>')+'|'+Math.round(r.x)+','+Math.round(r.y), name: path.join('>'), indicator: indicator,
    visible: r.width > 0 && r.height > 0 && r.bottom > 0 && r.right > 0 && r.top < innerHeight && r.left < innerWidth});
})()`

const countFocusable = `document.querySelectorAll('a[href],button,input:not([type=hidden]),select,textarea,[tabindex]:not([tabindex="-1"])').length`
