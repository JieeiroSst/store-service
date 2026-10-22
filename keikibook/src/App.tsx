import { useEffect, useState, type KeyboardEvent } from 'react'
import {
  Armchair,
  ArrowRight,
  AtSign,
  Award,
  BadgeCheck,
  BedDouble,
  BookOpen,
  Bot,
  BrainCircuit,
  Briefcase,
  Building2,
  Car,
  ChefHat,
  Clapperboard,
  Cloud,
  ConciergeBell,
  CreditCard,
  CupSoda,
  Dices,
  Download,
  Dumbbell,
  Fingerprint,
  Flag,
  Gift,
  Globe,
  History,
  House,
  IdCard,
  Images,
  KeyRound,
  LayoutGrid,
  Library,
  Link,
  Lock,
  Map as MapIcon,
  Megaphone,
  Menu,
  MessageCircle,
  MessageCircleQuestion,
  MessagesSquare,
  MonitorPlay,
  Newspaper,
  PiggyBank,
  Radio,
  ReceiptText,
  ScanFace,
  Search,
  ShieldCheck,
  ShieldHalf,
  ShoppingBag,
  Sparkles,
  SquareParking,
  Store,
  Sun,
  Ticket,
  TrendingUp,
  Trophy,
  Truck,
  Upload,
  UserRound,
  UsersRound,
  Utensils,
  Video,
  Wallet,
  X,
  Youtube,
  Zap,
  type LucideIcon,
} from 'lucide-react'

type Service = { icon: LucideIcon; name: string; note: string; tone: string; href?: string }

const services: Service[] = [
  { icon: Wallet, name: 'Ví Keiki', note: 'Nạp, chuyển, nhận tiền', tone: 'indigo' },
  { icon: Utensils, name: 'Food', note: 'Đặt món từ nhà hàng', tone: 'orange' },
  { icon: ShoppingBag, name: 'Mua sắm', note: 'Hàng chính hãng, giao nhanh', tone: 'pink' },
  { icon: MessageCircle, name: 'Chat', note: 'Bạn bè và nhóm', tone: 'teal' },
  { icon: Clapperboard, name: 'Phim', note: 'Gợi ý theo sở thích', tone: 'violet' },
  { icon: Ticket, name: 'Vé sự kiện', note: 'Concert, hội thảo', tone: 'blue' },
  { icon: Building2, name: 'Khách sạn', note: 'Đặt phòng giá tốt', tone: 'green' },
  { icon: Car, name: 'Thuê xe', note: 'Tự lái hoặc có tài xế', tone: 'amber' },
  { icon: Bot, name: 'Trợ lý AI', note: 'Hỏi gì cũng được', tone: 'violet' },
  { icon: LayoutGrid, name: '54 mini-app', note: 'Xem tất cả bên dưới', tone: 'slate', href: '#mini-app' },
]

type App = { icon: LucideIcon; name: string; note: string }
type Group = { id: string; label: string; tone: string; image: string; apps: App[] }

const groups: Group[] = [
  {
    id: 'giai-tri',
    label: 'Giải trí',
    tone: 'violet',
    image: '/screens/ticket-event.webp',
    apps: [
      { icon: Clapperboard, name: 'Phim', note: 'Gợi ý theo sở thích, xem ngay' },
      { icon: Ticket, name: 'Vé sự kiện', note: 'Chọn vé, trả bằng ví, vé QR' },
      { icon: Dices, name: 'Vòng quay may mắn', note: 'Quay mỗi ngày, điểm danh nhận quà' },
      { icon: Trophy, name: 'Xổ số', note: 'Kết quả và dò vé' },
      { icon: TrendingUp, name: 'Thị trường dự đoán', note: 'Đặt cửa Có/Không cho sự kiện' },
      { icon: MonitorPlay, name: 'Video', note: 'Xem video HLS chất lượng cao' },
      { icon: Youtube, name: 'Video YouTube', note: 'Thư viện video theo chủ đề' },
      { icon: Radio, name: 'Livestream', note: 'Xem trực tiếp, mở phòng live' },
      { icon: BookOpen, name: 'Đọc sách', note: 'Đọc theo chương, chỉnh cỡ chữ' },
      { icon: Library, name: 'Nhà sách', note: 'Mua sách, tác giả nổi bật' },
    ],
  },
  {
    id: 'du-lich',
    label: 'Du lịch & đi lại',
    tone: 'green',
    image: '/screens/reservation-hotel.webp',
    apps: [
      { icon: House, name: 'Homestay', note: 'Đặt theo đêm hoặc theo tháng' },
      { icon: Building2, name: 'Khách sạn', note: 'Báo giá theo ngày, huỷ miễn phí' },
      { icon: Car, name: 'Thuê xe', note: 'Ô tô, xe máy, xe đạp' },
      { icon: BedDouble, name: 'Đặt phòng nhanh', note: 'Phòng trống trong vài chạm' },
      { icon: SquareParking, name: 'Bãi đỗ xe', note: 'Chỗ trống, gửi và lấy xe' },
    ],
  },
  {
    id: 'tai-chinh',
    label: 'Tài chính & ưu đãi',
    tone: 'indigo',
    image: '/screens/payment-wallet.webp',
    apps: [
      { icon: PiggyBank, name: 'Ví túi tiền', note: 'Ngăn tiết kiệm, chuyển, yêu cầu tiền' },
      { icon: Gift, name: 'Voucher', note: 'Nhận quà, kiểm tra và dùng mã' },
      { icon: UsersRound, name: 'Giới thiệu bạn bè', note: 'Link mời và thưởng theo bậc' },
      { icon: Award, name: 'Đổi thưởng', note: 'Điểm hạng thành viên, ưu đãi tốt nhất' },
      { icon: Sparkles, name: 'Phần thưởng', note: 'Quà tặng đổi bằng điểm' },
      { icon: BadgeCheck, name: 'Bonuslink', note: 'Điểm thưởng liên kết' },
      { icon: ShieldCheck, name: 'Xác thực thanh toán', note: 'OTP 3-D Secure cho thẻ' },
      { icon: CreditCard, name: 'Thanh toán thẻ', note: 'Stripe, PayPal, VNPay' },
      { icon: Wallet, name: 'Cổng thanh toán', note: 'MoMo, chuyển khoản, hoàn tiền' },
      { icon: ReceiptText, name: 'Gói dịch vụ', note: 'Gói tháng, gói năm' },
      { icon: Sun, name: 'Thời tiết & thị trường', note: 'Dự báo, thuỷ triều, giá tiền số' },
    ],
  },
  {
    id: 'cong-dong',
    label: 'Cộng đồng',
    tone: 'pink',
    image: '/screens/threads.webp',
    apps: [
      { icon: AtSign, name: 'Threads', note: 'Bảng tin, thích, bình luận, lưu' },
      { icon: Newspaper, name: 'Bài viết', note: 'Chuyên mục và bài có ảnh bìa' },
      { icon: MessagesSquare, name: 'Phòng chat', note: 'Phòng nhiều người, quản lý thành viên' },
      { icon: Video, name: 'Gọi video', note: 'Họp nhóm tới 8 người' },
      { icon: MessageCircleQuestion, name: 'Hỏi & chia sẻ', note: 'Hỏi đáp cùng AI' },
      { icon: Images, name: 'Ghép ảnh nghệ thuật', note: 'Lưới, mosaic từ ảnh của bạn' },
      { icon: Flag, name: 'Báo cáo chat', note: 'Báo tin rác, lừa đảo' },
      { icon: Briefcase, name: 'Tuyển dụng', note: 'Ứng tuyển, làm cộng tác viên' },
      { icon: Megaphone, name: 'Quảng cáo', note: 'Ưu đãi theo vị trí hiển thị' },
      { icon: KeyRound, name: 'SSO doanh nghiệp', note: 'Đăng nhập cổng đối tác' },
    ],
  },
  {
    id: 'tien-ich',
    label: 'Tiện ích & ăn uống',
    tone: 'orange',
    image: '/screens/ekyc.webp',
    apps: [
      { icon: IdCard, name: 'Xác thực danh tính', note: 'eKYC 3 bước bằng CCCD' },
      { icon: ScanFace, name: 'Nhận diện khuôn mặt', note: 'Tải ảnh, phát hiện khuôn mặt' },
      { icon: Dumbbell, name: 'Phòng gym', note: 'Lịch lớp, mua và kiểm tra thẻ' },
      { icon: ChefHat, name: 'Bếp & món ăn', note: 'Thực đơn theo danh mục' },
      { icon: ConciergeBell, name: 'Gọi món', note: 'Chọn món, gửi thẳng xuống bếp' },
      { icon: Armchair, name: 'Đặt bàn', note: 'Giữ bàn trước khi đến' },
      { icon: CupSoda, name: 'Máy bán hàng', note: 'Chọn khe, trả bằng ví' },
      { icon: Store, name: 'Shopify', note: 'Sản phẩm và phiên bản' },
      { icon: Truck, name: 'Vận chuyển', note: 'Báo giá giao hàng' },
      { icon: MapIcon, name: 'Bản đồ & địa giới', note: 'Tỉnh, quận, phường, toạ độ' },
      { icon: Search, name: 'Tìm kiếm', note: 'Tìm trong toàn bộ Keikibook' },
      { icon: Globe, name: 'Tìm kiếm web', note: 'Web, tin tức, hình ảnh' },
      { icon: Bot, name: 'Chatbot', note: 'Lịch sử hội thoại AI' },
      { icon: BrainCircuit, name: 'Mô hình AI', note: 'Trò chuyện với Llama, Qwen…' },
      { icon: Images, name: 'Ghép ảnh', note: 'Collage nhiều cột' },
      { icon: Upload, name: 'Tải tệp', note: 'Lưu và chia sẻ tệp' },
      { icon: Cloud, name: 'Tệp tin', note: 'Lưu trữ CDN' },
      { icon: Link, name: 'Link rút gọn', note: 'Giải mã deep link' },
    ],
  },
]

const gallery = [
  { src: '/screens/lucky-game.webp', alt: 'Vòng quay may mắn' },
  { src: '/screens/reservation-hotel.webp', alt: 'Phòng trống của khách sạn' },
  { src: '/screens/payment-wallet.webp', alt: 'Ví túi tiền' },
  { src: '/screens/voucher.webp', alt: 'Voucher của tôi' },
  { src: '/screens/ekyc.webp', alt: 'Xác minh danh tính eKYC' },
  { src: '/screens/restaurant-consumer-menu.webp', alt: 'Gọi món tại nhà hàng' },
]

const apkUrl = import.meta.env.VITE_APK_URL as string | undefined

const stats = [
  { value: '95', label: 'dịch vụ' },
  { value: '54', label: 'mini-app' },
  { value: '1', label: 'ví cho tất cả' },
]

const features = [
  {
    id: 'vi',
    kicker: 'Ví Keiki',
    title: 'Tiền của bạn, trong tầm tay',
    text: 'Nạp từ ngân hàng, chuyển cho bạn bè bằng mã ví, nhận tiền qua QR và thanh toán mọi dịch vụ trong app. Mỗi giao dịch đều có mã PIN và khoá chống trừ tiền hai lần.',
    points: ['Thẻ ảo miễn phí cho thanh toán', 'Lịch sử giao dịch rõ ràng', 'Điểm thưởng sau mỗi lần dùng'],
    image: '/screens/wallet.webp',
  },
  {
    id: 'food',
    kicker: 'Keiki Food',
    title: 'Món ngon gần bạn',
    text: 'Chọn nhà hàng, thêm món vào giỏ và đặt chỉ trong vài chạm. Tiền chỉ bị trừ khi đơn được giao; huỷ đơn là hoàn ngay.',
    points: ['Theo dõi đơn theo thời gian thực', 'Thanh toán bằng thẻ Ví Keiki', 'Tip cho tài xế ngay trong app'],
    image: '/screens/food.webp',
  },
  {
    id: 'shop',
    kicker: 'Keiki Shop',
    title: 'Mua sắm không lo hết hàng',
    text: 'Giá và tồn kho luôn được kiểm tra trên máy chủ. Khi thanh toán, hàng được giữ cho bạn trước khi trừ tiền.',
    points: ['Giỏ hàng đồng bộ mọi thiết bị', 'Mã giảm giá và danh sách yêu thích', 'Thông báo khi đơn được xử lý'],
    image: '/screens/shop.webp',
  },
  {
    id: 'giai-tri',
    kicker: 'Giải trí',
    title: 'Xem phim, săn vé, thử vận may',
    text: 'Phim được gợi ý theo những gì bạn đã xem và phát ngay trong app. Đặt vé concert, chọn loại vé rồi trả bằng Ví Keiki; vé QR nằm sẵn trong máy, tặng hay bán lại đều được.',
    points: ['Phát video và livestream ngay trong app', 'Vòng quay, điểm danh mỗi ngày nhận điểm', 'Đọc sách theo chương, dò vé số'],
    image: '/screens/movie.webp',
  },
  {
    id: 'du-lich',
    kicker: 'Du lịch',
    title: 'Chỗ ở và xe cho mọi chuyến đi',
    text: 'Tìm homestay, khách sạn theo ngày đi và số khách, xem giá thật cho từng loại phòng. Giữ chỗ trước, trả bằng ví; huỷ trong hạn là được hoàn theo chính sách của nơi ở.',
    points: ['Homestay theo đêm hoặc theo tháng', 'Thuê ô tô, xe máy bằng tài khoản Keikibook', 'Gửi xe và lấy xe ở bãi đỗ'],
    image: '/screens/rent-house.webp',
  },
  {
    id: 'ket-noi',
    kicker: 'Kết nối',
    title: 'Tin nhắn đến ngay, gọi video một chạm',
    text: 'Chat với bạn bè và nhóm nhận tin theo thời gian thực, tự nối lại khi mạng chập chờn. Tạo phòng họp video cho tới 8 người, hoặc chia sẻ khoảnh khắc trên Threads.',
    points: ['Chat realtime, không sót tin khi mất mạng', 'Gọi video nhóm, đổi camera, tắt mic', 'Threads, bài viết và phòng chat'],
    image: '/screens/threads.webp',
  },
]

const roles = [
  {
    id: 'khach-hang',
    icon: UserRound,
    kicker: 'Khách hàng',
    title: 'Dùng dịch vụ, không thấy gì thừa',
    text: 'Mọi tài khoản mới là khách hàng. Ứng dụng chỉ hiện những gì bạn cần: mua, đặt, xem, trò chuyện và quản lý ví của chính mình.',
    points: ['Ba tab gọn: Trang chủ, Dịch vụ, Tài khoản', '54 mini-app cho nhu cầu hằng ngày', 'Lịch sử, vé, đơn hàng gắn với tài khoản'],
    image: '/screens/home.webp',
  },
  {
    id: 'quan-tri',
    icon: ShieldHalf,
    kicker: 'Quản trị',
    title: 'Một tab để vận hành cả nền tảng',
    text: 'Tài khoản được cấp quyền admin có thêm tab Quản trị: hàng chờ duyệt, khu quản lý cho ban tổ chức, khách sạn, homestay, voucher doanh nghiệp, quầy giao xe và mọi chức năng quản trị của từng dịch vụ.',
    points: ['Duyệt sự kiện, khách sạn, homestay mới', 'Tạo sự kiện, loại vé, sơ đồ ghế, khuyến mãi, soát vé', 'Nút "Xem như khách" để kiểm tra giao diện khách hàng'],
    image: '/screens/admin-home.webp',
  },
]

const consoles = [
  { src: '/screens/ticket-organizer.webp', alt: 'Ban tổ chức: sự kiện của tôi' },
  { src: '/screens/ticket-manage-event.webp', alt: 'Quản lý sự kiện: doanh thu, loại vé' },
  { src: '/screens/poly-market.webp', alt: 'Thị trường dự đoán: sổ lệnh, biểu đồ giá' },
]

const security = [
  { icon: Lock, title: 'Mã PIN cho mọi giao dịch', text: 'Chuyển, rút và thanh toán đều cần PIN của riêng bạn.' },
  { icon: KeyRound, title: 'Phiên đăng nhập an toàn', text: 'Token lưu trong bộ nhớ bảo mật của máy và tự làm mới.' },
  { icon: ShieldCheck, title: 'Quyền theo vai trò', text: 'Khách hàng và admin thấy giao diện khác nhau; quyền được cấp ở authorize-service và kiểm tra lại trên máy chủ.' },
  { icon: Fingerprint, title: 'Xác thực danh tính', text: 'eKYC bằng CCCD và khuôn mặt khi bạn cần hạn mức cao hơn.' },
  { icon: MessageCircle, title: 'Chat và gọi video đúng người', text: 'Máy chủ tự gắn danh tính từ phiên đăng nhập, không ai mạo danh được bạn.' },
  { icon: History, title: 'Dữ liệu theo tài khoản', text: 'Đặt phòng, giao dịch, ảnh ghép lưu trên máy chủ; đổi điện thoại vẫn còn nguyên.' },
]

function Logo({ size = 36 }: { size?: number }) {
  return (
    <a className="brand" href="#top" aria-label="Keikibook — về đầu trang">
      <img src="/logo.svg" width={size} height={size} alt="" />
      <span>Keikibook</span>
    </a>
  )
}

function Nav() {
  const [open, setOpen] = useState(false)
  const [scrolled, setScrolled] = useState(false)
  useEffect(() => {
    const onScroll = () => setScrolled(window.scrollY > 12)
    onScroll()
    window.addEventListener('scroll', onScroll, { passive: true })
    return () => window.removeEventListener('scroll', onScroll)
  }, [])
  const links = [
    ['#dich-vu', 'Dịch vụ'],
    ['#mini-app', 'Mini-app'],
    ['#vi', 'Ví Keiki'],
    ['#vai-tro', 'Vai trò'],
    ['#bao-mat', 'Bảo mật'],
    ['#tai-app', 'Tải app'],
  ]
  return (
    <header className={`nav ${scrolled ? 'nav--solid' : ''}`}>
      <div className="container nav__inner">
        <Logo />
        <nav className={`nav__links ${open ? 'is-open' : ''}`} aria-label="Điều hướng chính">
          {links.map(([href, label]) => (
            <a key={href} href={href} onClick={() => setOpen(false)}>
              {label}
            </a>
          ))}
          <a className="btn btn--light nav__cta" href="#tai-app" onClick={() => setOpen(false)}>
            <Download size={18} /> Tải Keikibook
          </a>
        </nav>
        <button className="nav__toggle" aria-label={open ? 'Đóng menu' : 'Mở menu'} aria-expanded={open} onClick={() => setOpen(!open)}>
          {open ? <X /> : <Menu />}
        </button>
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

function Hero() {
  return (
    <section className="hero" id="top">
      <div className="hero__glow hero__glow--a" />
      <div className="hero__glow hero__glow--b" />
      <div className="container hero__inner">
        <div className="hero__copy">
          <span className="pill">
            <Sparkles size={16} /> Super app cho người Việt
          </span>
          <h1>
            Mọi dịch vụ,
            <br />
            <span className="hero__accent">một ứng dụng.</span>
          </h1>
          <p className="hero__lead">
            Ví điện tử, đồ ăn, mua sắm, phim, vé, khách sạn, chat và gọi video — 54 mini-app chung một tài khoản, chung một ví.
          </p>
          <div className="hero__actions">
            <a className="btn btn--light" href="#tai-app">
              <Download size={20} /> Tải cho Android
            </a>
            <a className="btn btn--ghost" href="#dich-vu">
              Khám phá dịch vụ <ArrowRight size={18} />
            </a>
          </div>
          <dl className="stats">
            {stats.map((s) => (
              <div key={s.label}>
                <dt>{s.value}</dt>
                <dd>{s.label}</dd>
              </div>
            ))}
          </dl>
        </div>
        <div className="hero__visual">
          <Phone src="/screens/home.webp" alt="Màn hình trang chủ Keikibook" className="phone--main" />
          <div className="float-card float-card--pay">
            <span className="float-card__icon">
              <BadgeCheck size={18} />
            </span>
            <div>
              <strong>Thanh toán thành công</strong>
              <small>−125.000 ₫ • Keiki Food</small>
            </div>
          </div>
          <div className="float-card float-card--points">
            <span className="float-card__icon float-card__icon--amber">
              <Sparkles size={18} />
            </span>
            <div>
              <strong>+120 điểm thưởng</strong>
              <small>Cảm ơn bạn đã dùng Keikibook</small>
            </div>
          </div>
        </div>
      </div>
    </section>
  )
}

function Services() {
  return (
    <section className="section" id="dich-vu">
      <div className="container">
        <div className="section__head">
          <span className="kicker">Dịch vụ</span>
          <h2>Một tài khoản, mở ra cả thế giới</h2>
          <p>Mỗi dịch vụ là một mini-app nhẹ, mở ngay trong Keikibook — không phải cài thêm gì.</p>
        </div>
        <ul className="service-grid">
          {services.map(({ icon: Icon, name, note, tone, href }) => {
            const body = (
              <>
                <span className="service__icon">
                  <Icon size={26} strokeWidth={2.1} />
                </span>
                <strong>{name}</strong>
                <small>{note}</small>
              </>
            )
            return (
              <li key={name} className={`service service--${tone}`}>
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

function MiniApps() {
  const [active, setActive] = useState(groups[0].id)
  const group = groups.find((g) => g.id === active) ?? groups[0]
  const onKey = (e: KeyboardEvent<HTMLDivElement>) => {
    if (e.key !== 'ArrowRight' && e.key !== 'ArrowLeft') return
    const i = groups.findIndex((g) => g.id === active)
    const next = groups[(i + (e.key === 'ArrowRight' ? 1 : groups.length - 1)) % groups.length]
    setActive(next.id)
    document.getElementById(`tab-${next.id}`)?.focus()
  }
  return (
    <section className="section section--tinted" id="mini-app">
      <div className="container">
        <div className="section__head">
          <span className="kicker">Mini-app</span>
          <h2>54 mini-app, mở là dùng</h2>
          <p>Mỗi dịch vụ có giao diện riêng, màu riêng, nhưng dùng chung tài khoản, ví và điểm thưởng.</p>
        </div>
        <div className="tabs" role="tablist" aria-label="Nhóm mini-app" onKeyDown={onKey}>
          {groups.map((g) => (
            <button
              key={g.id}
              id={`tab-${g.id}`}
              role="tab"
              type="button"
              className={`tab service--${g.tone}`}
              aria-selected={g.id === active}
              aria-controls={`panel-${g.id}`}
              tabIndex={g.id === active ? 0 : -1}
              onClick={() => setActive(g.id)}
            >
              {g.label}
              <span className="tab__count">{g.apps.length}</span>
            </button>
          ))}
        </div>
        <div className="apps" id={`panel-${group.id}`} role="tabpanel" aria-labelledby={`tab-${group.id}`}>
          <ul className="app-grid">
            {group.apps.map(({ icon: Icon, name, note }) => (
              <li key={name} className={`app service--${group.tone}`}>
                <span className="app__icon">
                  <Icon size={20} strokeWidth={2.2} />
                </span>
                <span>
                  <strong>{name}</strong>
                  <small>{note}</small>
                </span>
              </li>
            ))}
          </ul>
          <div className="apps__visual">
            <div className="feature__halo" />
            <Phone src={group.image} alt={`Mini-app nhóm ${group.label}`} />
          </div>
        </div>
        <ul className="gallery" aria-label="Ảnh chụp các mini-app">
          {gallery.map((g) => (
            <li key={g.src}>
              <Phone src={g.src} alt={g.alt} className="phone--small" />
              <span>{g.alt}</span>
            </li>
          ))}
        </ul>
      </div>
    </section>
  )
}

function Features() {
  return (
    <>
      {features.map((f, i) => (
        <section key={f.id} id={f.id} className={`section feature ${i % 2 ? 'feature--reverse' : ''}`}>
          <div className="container feature__inner">
            <div className="feature__copy">
              <span className="kicker">{f.kicker}</span>
              <h2>{f.title}</h2>
              <p>{f.text}</p>
              <ul className="checks">
                {f.points.map((p) => (
                  <li key={p}>
                    <BadgeCheck size={20} /> {p}
                  </li>
                ))}
              </ul>
            </div>
            <div className="feature__visual">
              <div className="feature__halo" />
              <Phone src={f.image} alt={`Màn hình ${f.kicker}`} />
            </div>
          </div>
        </section>
      ))}
    </>
  )
}

function Roles() {
  return (
    <section className="section" id="vai-tro">
      <div className="container">
        <div className="section__head">
          <span className="kicker">Vai trò</span>
          <h2>Một ứng dụng, hai vai trò</h2>
          <p>Cùng một Keikibook, giao diện tự đổi theo quyền của tài khoản. Máy chủ vẫn kiểm tra quyền ở từng thao tác.</p>
        </div>
        <div className="roles">
          {roles.map(({ id, icon: Icon, kicker, title, text, points, image }) => (
            <article key={id} className={`role role--${id}`}>
              <div className="role__copy">
                <span className="role__badge">
                  <Icon size={18} /> {kicker}
                </span>
                <h3>{title}</h3>
                <p>{text}</p>
                <ul className="checks">
                  {points.map((p) => (
                    <li key={p}>
                      <BadgeCheck size={20} /> {p}
                    </li>
                  ))}
                </ul>
              </div>
              <Phone src={image} alt={`Giao diện ${kicker.toLowerCase()}`} className="phone--small role__phone" />
            </article>
          ))}
        </div>
        <ul className="gallery gallery--center" aria-label="Các khu quản lý">
          {consoles.map((g) => (
            <li key={g.src}>
              <Phone src={g.src} alt={g.alt} className="phone--small" />
              <span>{g.alt}</span>
            </li>
          ))}
        </ul>
      </div>
    </section>
  )
}

function Security() {
  return (
    <section className="section section--tinted" id="bao-mat">
      <div className="container">
        <div className="section__head">
          <span className="kicker">Bảo mật</span>
          <h2>An toàn từ thiết kế</h2>
          <p>Mọi giao dịch đi qua một cổng duy nhất, được xác thực và kiểm tra quyền trước khi tới từng dịch vụ.</p>
        </div>
        <div className="security-grid">
          {security.map(({ icon: Icon, title, text }) => (
            <article key={title} className="security">
              <span className="security__icon">
                <Icon size={22} />
              </span>
              <h3>{title}</h3>
              <p>{text}</p>
            </article>
          ))}
        </div>
      </div>
    </section>
  )
}

function Download_() {
  return (
    <section className="section" id="tai-app">
      <div className="container">
        <div className="cta">
          <div className="hero__glow hero__glow--a" />
          <div className="cta__copy">
            <img src="/logo.svg" width={72} height={72} alt="" className="cta__logo" />
            <h2>Sẵn sàng thử Keikibook?</h2>
            <p>Tải bản Android ngay hôm nay. Bản iOS đang được hoàn thiện.</p>
            <div className="hero__actions">
              {apkUrl ? (
                <a className="btn btn--light" href={apkUrl}>
                  <Download size={20} /> Tải APK cho Android
                </a>
              ) : (
                <span className="btn btn--light is-disabled" aria-disabled="true">
                  <Download size={20} /> Android — sắp có link tải
                </span>
              )}
              <span className="btn btn--ghost is-disabled" aria-disabled="true">
                <Zap size={18} /> iOS — sắp ra mắt
              </span>
            </div>
          </div>
          <Phone src="/screens/services.webp" alt="Màn hình tất cả dịch vụ" className="cta__phone" />
        </div>
      </div>
    </section>
  )
}

function Footer() {
  return (
    <footer className="footer">
      <div className="container footer__inner">
        <Logo size={30} />
        <p>Mọi dịch vụ, một ứng dụng.</p>
        <nav aria-label="Liên kết cuối trang">
          <a href="#dich-vu">Dịch vụ</a>
          <a href="#mini-app">Mini-app</a>
          <a href="#vai-tro">Vai trò</a>
          <a href="#bao-mat">Bảo mật</a>
          <a href="#tai-app">
            <ReceiptText size={16} /> Tải app
          </a>
        </nav>
        <small>© {new Date().getFullYear()} Keikibook</small>
      </div>
    </footer>
  )
}

export default function App() {
  return (
    <>
      <Nav />
      <main>
        <Hero />
        <Services />
        <MiniApps />
        <Features />
        <Roles />
        <Security />
        <Download_ />
      </main>
      <Footer />
    </>
  )
}
