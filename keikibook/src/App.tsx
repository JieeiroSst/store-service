import { Fragment, createContext, useContext, useEffect, useMemo, useRef, useState, type CSSProperties, type KeyboardEvent, type PointerEvent } from 'react'
import { ArrowRight, ArrowUp, BadgeCheck, Download, Menu, Moon, ReceiptText, Search, SearchX, Sparkles, Sun, X, Zap } from 'lucide-react'
import { consoles, features, gallery, groups, roles, security, services, stats, ui, type L, type Lang, type MiniApp } from './content'

const apkUrl = import.meta.env.VITE_APK_URL as string | undefined
const total = groups.reduce((n, g) => n + g.apps.length, 0)

const LangContext = createContext<{ lang: Lang; setLang: (l: Lang) => void }>({ lang: 'vi', setLang: () => {} })

function useI18n() {
  const { lang, setLang } = useContext(LangContext)
  return { lang, setLang, t: ui[lang], tr: (s: L) => s[lang] }
}

function initialLang(): Lang {
  const q = new URLSearchParams(window.location.search).get('lang')
  if (q === 'vi' || q === 'en') return q
  try {
    const saved = localStorage.getItem('lang')
    if (saved === 'vi' || saved === 'en') return saved
  } catch {}
  return 'vi'
}

type Theme = 'light' | 'dark'

function initialTheme(): Theme {
  try {
    const saved = localStorage.getItem('theme')
    if (saved === 'light' || saved === 'dark') return saved
  } catch {}
  return window.matchMedia?.('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
}

function ThemeSwitch() {
  const { t } = useI18n()
  const [theme, setTheme] = useState<Theme>(initialTheme)
  useEffect(() => {
    document.documentElement.dataset.theme = theme
    try {
      localStorage.setItem('theme', theme)
    } catch {}
  }, [theme])
  const dark = theme === 'dark'
  return (
    <button type="button" className="icon-btn" aria-label={dark ? t.lightMode : t.darkMode} title={dark ? t.lightMode : t.darkMode} onClick={() => setTheme(dark ? 'light' : 'dark')}>
      {dark ? <Sun size={18} /> : <Moon size={18} />}
    </button>
  )
}

function useActiveSection(ids: string[]) {
  const [active, setActive] = useState<string | null>(null)
  useEffect(() => {
    const els = ids.map((id) => document.getElementById(id)).filter((e): e is HTMLElement => !!e)
    const io = new IntersectionObserver(
      (entries) => {
        for (const e of entries) if (e.isIntersecting) setActive(e.target.id)
      },
      { rootMargin: '-45% 0px -50% 0px' },
    )
    els.forEach((e) => io.observe(e))
    return () => io.disconnect()
  }, [ids.join()])
  return active
}

function Logo({ size = 36 }: { size?: number }) {
  const { t } = useI18n()
  return (
    <a className="brand" href="#top" aria-label={t.toTop}>
      <img src="/logo.svg" width={size} height={size} alt="" />
      <span>Keikibook</span>
    </a>
  )
}

function LangSwitch() {
  const { lang, setLang, t } = useI18n()
  return (
    <div className="lang" role="group" aria-label={t.language}>
      {(['vi', 'en'] as const).map((l) => (
        <button key={l} type="button" lang={l} aria-pressed={lang === l} onClick={() => setLang(l)}>
          {l.toUpperCase()}
        </button>
      ))}
    </div>
  )
}

function Nav() {
  const { t } = useI18n()
  const [open, setOpen] = useState(false)
  const [scrolled, setScrolled] = useState(false)
  useEffect(() => {
    const onScroll = () => setScrolled(window.scrollY > 12)
    onScroll()
    window.addEventListener('scroll', onScroll, { passive: true })
    return () => window.removeEventListener('scroll', onScroll)
  }, [])
  const links = [
    ['dich-vu', t.links.services],
    ['mini-app', t.links.miniApps],
    ['vi', t.links.wallet],
    ['vai-tro', t.links.roles],
    ['bao-mat', t.links.security],
    ['tai-app', t.links.download],
  ]
  const current = useActiveSection(links.map(([id]) => id))
  return (
    <header className={`nav ${scrolled ? 'nav--solid' : ''}`}>
      <div className="container nav__inner">
        <Logo />
        <nav className={`nav__links ${open ? 'is-open' : ''}`} aria-label={t.mainNav}>
          {links.map(([id, label]) => (
            <a key={id} href={`#${id}`} aria-current={current === id ? 'location' : undefined} onClick={() => setOpen(false)}>
              {label}
            </a>
          ))}
          <a className="btn btn--light nav__cta" href="#tai-app" onClick={() => setOpen(false)}>
            <Download size={18} /> {t.getApp}
          </a>
        </nav>
        <div className="nav__end">
          <ThemeSwitch />
          <LangSwitch />
          <button className="nav__toggle" aria-label={open ? t.closeMenu : t.openMenu} aria-expanded={open} onClick={() => setOpen(!open)}>
            {open ? <X /> : <Menu />}
          </button>
        </div>
      </div>
    </header>
  )
}

function Phone({ src, alt, className = '' }: { src: string; alt: string; className?: string }) {
  return (
    <div className={`phone ${className}`}>
      <div className="phone__notch" />
      <img src={src} alt={alt} loading="lazy" width={390} height={844} />
    </div>
  )
}

const still = () => window.matchMedia?.('(prefers-reduced-motion: reduce)').matches

const idx = (i: number) => ({ '--i': i }) as CSSProperties

function Words({ text, from = 0 }: { text: string; from?: number }) {
  return (
    <>
      {text.split(' ').map((w, i) => (
        <Fragment key={i}>
          <span className="word" style={idx(from + i)}>
            {w}
          </span>{' '}
        </Fragment>
      ))}
    </>
  )
}

function CountUp({ value }: { value: string }) {
  const ref = useRef<HTMLElement>(null)
  const target = Number(value)
  const [n, setN] = useState(target)
  useEffect(() => {
    const el = ref.current
    if (!el || still() || !('IntersectionObserver' in window)) return
    setN(0)
    const io = new IntersectionObserver(([e]) => {
      if (!e.isIntersecting) return
      io.disconnect()
      const start = performance.now()
      const tick = (now: number) => {
        const p = Math.min(1, (now - start) / 1400)
        setN(Math.round(target * (1 - Math.pow(1 - p, 3))))
        if (p < 1) requestAnimationFrame(tick)
      }
      requestAnimationFrame(tick)
    })
    io.observe(el)
    return () => io.disconnect()
  }, [target])
  return <dt ref={ref}>{n}</dt>
}

function Sparkles_() {
  return (
    <div className="sparkles" aria-hidden="true">
      {Array.from({ length: 18 }, (_, i) => (
        <span key={i} style={{ ...idx(i), left: `${(i * 37) % 100}%`, top: `${(i * 53) % 100}%` }} />
      ))}
    </div>
  )
}

function tilt(e: PointerEvent<HTMLElement>) {
  if (e.pointerType !== 'mouse' || still()) return
  const r = e.currentTarget.getBoundingClientRect()
  e.currentTarget.style.setProperty('--mx', (((e.clientX - r.left) / r.width) * 2 - 1).toFixed(3))
  e.currentTarget.style.setProperty('--my', (((e.clientY - r.top) / r.height) * 2 - 1).toFixed(3))
}

function untilt(e: PointerEvent<HTMLElement>) {
  e.currentTarget.style.setProperty('--mx', '0')
  e.currentTarget.style.setProperty('--my', '0')
}

function Hero() {
  const { t, tr } = useI18n()
  return (
    <section className="hero" id="top" onPointerMove={tilt} onPointerLeave={untilt}>
      <div className="hero__glow hero__glow--a" />
      <div className="hero__glow hero__glow--b" />
      <Sparkles_ />
      <div className="container hero__inner">
        <div className="hero__copy">
          <span className="pill hero__in" style={idx(0)}>
            <Sparkles size={16} /> {t.heroPill}
          </span>
          <h1 key={t.heroTitle[0]} aria-label={t.heroTitle.join(' ')}>
            <span aria-hidden="true">
              <Words text={t.heroTitle[0]} />
              <br />
              <span className="hero__accent word" style={idx(t.heroTitle[0].split(' ').length)}>
                {t.heroTitle[1]}
              </span>
            </span>
          </h1>
          <p className="hero__lead hero__in" style={idx(5)}>
            {t.heroLead}
          </p>
          <div className="hero__actions hero__in" style={idx(6)}>
            <a className="btn btn--light" href="#tai-app">
              <Download size={20} /> {t.heroAndroid}
            </a>
            <a className="btn btn--ghost" href="#dich-vu">
              {t.heroExplore} <ArrowRight size={18} />
            </a>
          </div>
          <dl className="stats hero__in" style={idx(7)}>
            {stats.map((s) => (
              <div key={s.value}>
                <CountUp value={s.value} />
                <dd>{tr(s.label)}</dd>
              </div>
            ))}
          </dl>
        </div>
        <div className="hero__visual hero__in" style={idx(3)}>
          <Phone src="/screens/home.webp" alt={t.heroPhone} className="phone--main" />
          <div className="float-card float-card--pay">
            <span className="float-card__icon">
              <BadgeCheck size={18} />
            </span>
            <div>
              <strong>{t.payOk}</strong>
              <small>−125.000 ₫ • Keiki Food</small>
            </div>
          </div>
          <div className="float-card float-card--points">
            <span className="float-card__icon float-card__icon--amber">
              <Sparkles size={18} />
            </span>
            <div>
              <strong>{t.points}</strong>
              <small>{t.thanks}</small>
            </div>
          </div>
        </div>
      </div>
    </section>
  )
}

function Services() {
  const { t, tr } = useI18n()
  return (
    <section className="section" id="dich-vu">
      <div className="container">
        <div className="section__head reveal">
          <span className="kicker">{t.links.services}</span>
          <h2>{t.servicesTitle}</h2>
          <p>{t.servicesLead}</p>
        </div>
        <ul className="service-grid reveal stagger">
          {services.map(({ icon: Icon, name, note, tone, href }, i) => {
            const body = (
              <>
                <span className="service__icon">
                  <Icon size={26} strokeWidth={2.1} />
                </span>
                <strong>{tr(name)}</strong>
                <small>{tr(note)}</small>
              </>
            )
            return (
              <li key={name.vi} className={`service service--${tone} spot`} style={idx(i)}>
                {href ? (
                  <a href={href} className="service__link">
                    {body}
                  </a>
                ) : (
                  body
                )}
              </li>
            )
          })}
        </ul>
      </div>
    </section>
  )
}

function AppCard({ app, tone, onOpen }: { app: MiniApp; tone: string; onOpen: (app: MiniApp) => void }) {
  const { t, tr } = useI18n()
  const { icon: Icon, shot, name, note, desc } = app
  return (
    <li className={`app service--${tone} spot`}>
      <button type="button" className="app__shot" onClick={() => onOpen(app)} aria-label={`${t.viewShot}: ${tr(name)}`}>
        <img src={`/screens/apps/${shot}.webp`} alt="" loading="lazy" width={390} height={844} />
      </button>
      <div className="app__body">
        <span className="app__icon">
          <Icon size={20} strokeWidth={2.2} />
        </span>
        <span>
          <strong>{tr(name)}</strong>
          <small>{tr(note)}</small>
          <p>{tr(desc)}</p>
        </span>
      </div>
    </li>
  )
}

function ShotDialog({ app, onClose }: { app: MiniApp | null; onClose: () => void }) {
  const { t, tr } = useI18n()
  const ref = useRef<HTMLDialogElement>(null)
  useEffect(() => {
    const d = ref.current
    if (!d) return
    if (app && !d.open) d.showModal()
    if (!app && d.open) d.close()
  }, [app])
  return (
    <dialog ref={ref} className="shot" onClose={onClose} onClick={(e) => e.target === e.currentTarget && onClose()} aria-label={app ? tr(app.name) : undefined}>
      {app && (
        <div className="shot__inner">
          <button type="button" className="shot__close" onClick={onClose} aria-label={t.close}>
            <X />
          </button>
          <Phone src={`/screens/apps/${app.shot}.webp`} alt={tr(app.name)} />
          <div className="shot__copy">
            <strong>{tr(app.name)}</strong>
            <p>{tr(app.desc)}</p>
          </div>
        </div>
      )}
    </dialog>
  )
}

const fold = (s: string) => s.normalize('NFD').replace(/[\u0300-\u036f]/g, '').replace(/đ/g, 'd').replace(/Đ/g, 'D').toLowerCase()

const PREVIEW = 6

function MiniApps() {
  const { lang, t, tr } = useI18n()
  const [active, setActive] = useState('all')
  const [query, setQuery] = useState('')
  const [expanded, setExpanded] = useState<string[]>([])
  const [opened, setOpened] = useState<MiniApp | null>(null)
  const tabs = [{ id: 'all', label: t.all, tone: 'slate', count: total }, ...groups.map((g) => ({ id: g.id, label: tr(g.label), tone: g.tone, count: g.apps.length }))]
  const group = groups.find((g) => g.id === active)
  const q = fold(query.trim())
  const matches = useMemo(
    () =>
      groups
        .filter((g) => active === 'all' || g.id === active)
        .map((g) => ({ ...g, apps: q ? g.apps.filter((a) => [a.name, a.note, a.desc].some((x) => fold(x.vi).includes(q) || fold(x.en).includes(q))) : g.apps }))
        .filter((g) => g.apps.length > 0),
    [active, q],
  )
  const found = matches.reduce((n, g) => n + g.apps.length, 0)
  const onKey = (e: KeyboardEvent<HTMLDivElement>) => {
    if (e.key !== 'ArrowRight' && e.key !== 'ArrowLeft') return
    const i = tabs.findIndex((g) => g.id === active)
    const next = tabs[(i + (e.key === 'ArrowRight' ? 1 : tabs.length - 1)) % tabs.length]
    setActive(next.id)
    document.getElementById(`tab-${next.id}`)?.focus()
  }
  const panel = `panel-${active}`
  return (
    <section className="section section--tinted" id="mini-app">
      <div className="container">
        <div className="section__head reveal">
          <span className="kicker">{t.links.miniApps}</span>
          <h2>{t.miniTitle}</h2>
          <p>{t.miniLead}</p>
        </div>
        <div className="apps-toolbar">
          <label className="search">
            <Search size={18} />
            <input type="search" value={query} onChange={(e) => setQuery(e.target.value)} placeholder={t.searchApps} aria-label={t.searchApps} aria-controls={panel} />
            {query && (
              <button type="button" onClick={() => setQuery('')} aria-label={t.clearSearch}>
                <X size={16} />
              </button>
            )}
          </label>
          <div className="tabs" role="tablist" aria-label={t.miniGroups} onKeyDown={onKey}>
            {tabs.map((g) => (
              <button
                key={g.id}
                id={`tab-${g.id}`}
                role="tab"
                type="button"
                className={`tab service--${g.tone}`}
                aria-selected={g.id === active}
                aria-controls={panel}
                tabIndex={g.id === active ? 0 : -1}
                onClick={() => setActive(g.id)}
              >
                {g.label}
                <span className="tab__count">{g.count}</span>
              </button>
            ))}
          </div>
        </div>
        {q && (
          <p className="apps-count" role="status">
            {t.results(found)}
          </p>
        )}
        {group && !q ? (
          <div className="apps panel-in" key={`${active}-${lang}`} id={panel} role="tabpanel" aria-labelledby={`tab-${group.id}`}>
            <ul className="app-grid">
              {group.apps.map((app) => (
                <AppCard key={app.name.vi} app={app} tone={group.tone} onOpen={setOpened} />
              ))}
            </ul>
            <div className="apps__visual">
              <div className="feature__halo" />
              <Phone src={group.image} alt={`${t.groupPhone} ${tr(group.label)}`} />
            </div>
          </div>
        ) : (
          <div className="apps-all panel-in" key={`${active}-${q ? 'q' : ''}-${lang}`} id={panel} role="tabpanel" aria-labelledby={`tab-${active}`}>
            {matches.map((g) => {
              const all = !!q || expanded.includes(g.id)
              const shown = all ? g.apps : g.apps.slice(0, PREVIEW)
              return (
                <section key={g.id} className={`apps-group service--${g.tone}`} aria-labelledby={`group-${g.id}`}>
                  <h3 id={`group-${g.id}`}>
                    {tr(g.label)}
                    <span className="tab__count">{g.apps.length}</span>
                  </h3>
                  <ul className="app-grid app-grid--wide">
                    {shown.map((app) => (
                      <AppCard key={app.name.vi} app={app} tone={g.tone} onOpen={setOpened} />
                    ))}
                  </ul>
                  {shown.length < g.apps.length && (
                    <button type="button" className="more" onClick={() => setExpanded([...expanded, g.id])}>
                      {t.showMore(g.apps.length - shown.length)} <ArrowRight size={16} />
                    </button>
                  )}
                </section>
              )
            })}
            {found === 0 && (
              <div className="empty">
                <SearchX size={36} />
                <strong>{t.noResults}</strong>
                <p>{t.noResultsHint}</p>
                <button
                  type="button"
                  className="more"
                  onClick={() => {
                    setQuery('')
                    setActive('all')
                  }}
                >
                  {t.clearSearch}
                </button>
              </div>
            )}
          </div>
        )}
        <ShotDialog app={opened} onClose={() => setOpened(null)} />
        <ul className="gallery" aria-label={t.galleryLabel}>
          {gallery.map((g) => (
            <li key={g.src}>
              <Phone src={g.src} alt={tr(g.alt)} className="phone--small" />
              <span>{tr(g.alt)}</span>
            </li>
          ))}
        </ul>
      </div>
    </section>
  )
}

function Marquee() {
  const { tr } = useI18n()
  const apps = groups.flatMap((g) => g.apps.map((a) => ({ ...a, tone: g.tone })))
  const half = Math.ceil(apps.length / 2)
  const rows = [apps.slice(0, half), apps.slice(half)]
  return (
    <div className="marquee" aria-hidden="true">
      {rows.map((row, r) => (
        <div key={r} className={`marquee__row ${r ? 'marquee__row--reverse' : ''}`}>
          {[0, 1].map((copy) => (
            <ul key={copy}>
              {row.map(({ icon: Icon, name, tone }) => (
                <li key={name.vi} className={`service--${tone}`}>
                  <Icon size={18} strokeWidth={2.2} />
                  {tr(name)}
                </li>
              ))}
            </ul>
          ))}
        </div>
      ))}
    </div>
  )
}

function Features() {
  const { t, tr } = useI18n()
  return (
    <>
      {features.map((f, i) => (
        <section key={f.id} id={f.id} className={`section feature ${i % 2 ? 'feature--reverse' : ''}`}>
          <div className="container feature__inner reveal">
            <div className="feature__copy">
              <span className="kicker">{tr(f.kicker)}</span>
              <h2>{tr(f.title)}</h2>
              <p>{tr(f.text)}</p>
              <ul className="checks">
                {f.points.map((p) => (
                  <li key={p.vi}>
                    <BadgeCheck size={20} /> {tr(p)}
                  </li>
                ))}
              </ul>
            </div>
            <div className="feature__visual">
              <div className="feature__halo" />
              <Phone src={f.image} alt={`${t.screen} ${tr(f.kicker)}`} className="phone--drift" />
            </div>
          </div>
        </section>
      ))}
    </>
  )
}

function Roles() {
  const { t, tr } = useI18n()
  return (
    <section className="section" id="vai-tro">
      <div className="container">
        <div className="section__head reveal">
          <span className="kicker">{t.links.roles}</span>
          <h2>{t.rolesTitle}</h2>
          <p>{t.rolesLead}</p>
        </div>
        <div className="roles">
          {roles.map(({ id, icon: Icon, kicker, title, text, points, image }) => (
            <article key={id} className={`role role--${id} reveal`}>
              <div className="role__copy">
                <span className="role__badge">
                  <Icon size={18} /> {tr(kicker)}
                </span>
                <h3>{tr(title)}</h3>
                <p>{tr(text)}</p>
                <ul className="checks">
                  {points.map((p) => (
                    <li key={p.vi}>
                      <BadgeCheck size={20} /> {tr(p)}
                    </li>
                  ))}
                </ul>
              </div>
              <Phone src={image} alt={`${t.rolePhone} ${tr(kicker).toLowerCase()}`} className="phone--small role__phone" />
            </article>
          ))}
        </div>
        <ul className="gallery gallery--center" aria-label={t.consolesLabel}>
          {consoles.map((g) => (
            <li key={g.src}>
              <Phone src={g.src} alt={tr(g.alt)} className="phone--small" />
              <span>{tr(g.alt)}</span>
            </li>
          ))}
        </ul>
      </div>
    </section>
  )
}

function Security() {
  const { t, tr } = useI18n()
  return (
    <section className="section section--tinted" id="bao-mat">
      <div className="container">
        <div className="section__head reveal">
          <span className="kicker">{t.links.security}</span>
          <h2>{t.securityTitle}</h2>
          <p>{t.securityLead}</p>
        </div>
        <div className="security-grid reveal stagger">
          {security.map(({ icon: Icon, title, text }, i) => (
            <article key={title.vi} className="security spot" style={idx(i)}>
              <span className="security__icon">
                <Icon size={22} />
              </span>
              <h3>{tr(title)}</h3>
              <p>{tr(text)}</p>
            </article>
          ))}
        </div>
      </div>
    </section>
  )
}

function Download_() {
  const { t } = useI18n()
  return (
    <section className="section" id="tai-app">
      <div className="container">
        <div className="cta reveal" onPointerMove={tilt} onPointerLeave={untilt}>
          <div className="hero__glow hero__glow--a" />
          <Sparkles_ />
          <div className="cta__copy">
            <img src="/logo.svg" width={72} height={72} alt="" className="cta__logo" />
            <h2>{t.ctaTitle}</h2>
            <p>{t.ctaLead}</p>
            <div className="hero__actions">
              {apkUrl ? (
                <a className="btn btn--light" href={apkUrl}>
                  <Download size={20} /> {t.apk}
                </a>
              ) : (
                <span className="btn btn--light is-disabled" aria-disabled="true">
                  <Download size={20} /> {t.apkSoon}
                </span>
              )}
              <span className="btn btn--ghost is-disabled" aria-disabled="true">
                <Zap size={18} /> {t.iosSoon}
              </span>
            </div>
          </div>
          <Phone src="/screens/services.webp" alt={t.ctaPhone} className="cta__phone" />
        </div>
      </div>
    </section>
  )
}

function Footer() {
  const { t } = useI18n()
  return (
    <footer className="footer">
      <div className="container footer__inner">
        <Logo size={30} />
        <p>{t.tagline}</p>
        <nav aria-label={t.footerNav}>
          <a href="#dich-vu">{t.links.services}</a>
          <a href="#mini-app">{t.links.miniApps}</a>
          <a href="#vai-tro">{t.links.roles}</a>
          <a href="#bao-mat">{t.links.security}</a>
          <a href="#tai-app">
            <ReceiptText size={16} /> {t.links.download}
          </a>
        </nav>
        <small>© {new Date().getFullYear()} Keikibook</small>
      </div>
    </footer>
  )
}

function useReveal() {
  useEffect(() => {
    const els = document.querySelectorAll('.reveal')
    if (!('IntersectionObserver' in window)) return
    document.documentElement.classList.add('js-reveal')
    const io = new IntersectionObserver(
      (entries) => {
        for (const e of entries) {
          if (!e.isIntersecting) continue
          e.target.classList.add('is-visible')
          io.unobserve(e.target)
        }
      },
      { rootMargin: '0px 0px -8% 0px' },
    )
    els.forEach((e) => io.observe(e))
    return () => io.disconnect()
  }, [])
}

function ScrollProgress() {
  const ref = useRef<HTMLDivElement>(null)
  useEffect(() => {
    let frame = 0
    const update = () => {
      frame = 0
      const max = document.documentElement.scrollHeight - window.innerHeight
      ref.current?.style.setProperty('--p', String(max > 0 ? window.scrollY / max : 0))
    }
    const onScroll = () => {
      if (!frame) frame = requestAnimationFrame(update)
    }
    update()
    window.addEventListener('scroll', onScroll, { passive: true })
    window.addEventListener('resize', onScroll)
    return () => {
      window.removeEventListener('scroll', onScroll)
      window.removeEventListener('resize', onScroll)
      cancelAnimationFrame(frame)
    }
  }, [])
  return <div ref={ref} className="progress" aria-hidden="true" />
}

function useSpotlight() {
  useEffect(() => {
    const onMove = (e: globalThis.PointerEvent) => {
      if (e.pointerType !== 'mouse') return
      const el = (e.target as Element | null)?.closest?.('.spot') as HTMLElement | null
      if (!el) return
      const r = el.getBoundingClientRect()
      el.style.setProperty('--x', `${e.clientX - r.left}px`)
      el.style.setProperty('--y', `${e.clientY - r.top}px`)
    }
    document.addEventListener('pointermove', onMove, { passive: true })
    return () => document.removeEventListener('pointermove', onMove)
  }, [])
}

function BackToTop() {
  const { t } = useI18n()
  const [show, setShow] = useState(false)
  useEffect(() => {
    const onScroll = () => setShow(window.scrollY > 900)
    onScroll()
    window.addEventListener('scroll', onScroll, { passive: true })
    return () => window.removeEventListener('scroll', onScroll)
  }, [])
  return (
    <a href="#top" className={`to-top ${show ? 'is-shown' : ''}`} aria-label={t.backToTop} title={t.backToTop} tabIndex={show ? 0 : -1} aria-hidden={!show}>
      <ArrowUp size={20} />
    </a>
  )
}

export default function App() {
  const [lang, setLang] = useState<Lang>(initialLang)
  useReveal()
  useSpotlight()
  useEffect(() => {
    document.documentElement.lang = lang
    document.title = ui[lang].title
    try {
      localStorage.setItem('lang', lang)
    } catch {}
  }, [lang])
  return (
    <LangContext.Provider value={{ lang, setLang }}>
      <ScrollProgress />
      <Nav />
      <main>
        <Hero />
        <Services />
        <Marquee />
        <MiniApps />
        <Features />
        <Roles />
        <Security />
        <Download_ />
      </main>
      <Footer />
      <BackToTop />
    </LangContext.Provider>
  )
}
