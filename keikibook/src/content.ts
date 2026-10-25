import {
  Armchair,
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
  Smartphone,
  Sparkles,
  SquareParking,
  Store,
  Sun,
  Ticket,
  TrendingUp,
  Trophy,
  Truck,
  Upload,
  UserPlus,
  UserRound,
  UsersRound,
  Utensils,
  Video,
  Wallet,
  Youtube,
  type LucideIcon,
} from 'lucide-react'

export type Lang = 'vi' | 'en'
export type L = Record<Lang, string>

export const ui = {
  vi: {
    title: 'Keikibook — Mọi dịch vụ, một ứng dụng',
    toTop: 'Keikibook — về đầu trang',
    mainNav: 'Điều hướng chính',
    footerNav: 'Liên kết cuối trang',
    openMenu: 'Mở menu',
    closeMenu: 'Đóng menu',
    language: 'Ngôn ngữ',
    viewShot: 'Xem ảnh màn hình',
    darkMode: 'Chế độ tối',
    lightMode: 'Chế độ sáng',
    backToTop: 'Lên đầu trang',
    searchApps: 'Tìm mini-app: ví, phim, đặt bàn…',
    results: (n: number) => `${n} mini-app phù hợp`,
    noResults: 'Không tìm thấy mini-app nào',
    noResultsHint: 'Thử từ khoá khác hoặc xem tất cả.',
    clearSearch: 'Xoá tìm kiếm',
    showMore: (n: number) => `Xem thêm ${n} app`,
    close: 'Đóng',
    links: { services: 'Dịch vụ', miniApps: 'Mini-app', wallet: 'Ví Keiki', roles: 'Vai trò', security: 'Bảo mật', start: 'Bắt đầu', faq: 'Hỏi đáp', download: 'Tải app' },
    getApp: 'Tải Keikibook',
    heroPill: 'Super app cho người Việt',
    heroTitle: ['Mọi dịch vụ,', 'một ứng dụng.'],
    heroLead: 'Ví điện tử, đồ ăn, mua sắm, phim, vé, khách sạn, chat và gọi video — 54 mini-app chung một tài khoản, chung một ví.',
    heroAndroid: 'Tải cho Android',
    heroExplore: 'Khám phá dịch vụ',
    heroPhone: 'Màn hình trang chủ Keikibook',
    payOk: 'Thanh toán thành công',
    points: '+120 điểm thưởng',
    thanks: 'Cảm ơn bạn đã dùng Keikibook',
    servicesTitle: 'Một tài khoản, mở ra cả thế giới',
    servicesLead: 'Mỗi dịch vụ là một mini-app nhẹ, mở ngay trong Keikibook — không phải cài thêm gì.',
    miniTitle: '54 mini-app, mở là dùng',
    miniLead: 'Mỗi dịch vụ có giao diện riêng, màu riêng, nhưng dùng chung tài khoản, ví và điểm thưởng.',
    miniGroups: 'Nhóm mini-app',
    all: 'Tất cả',
    groupPhone: 'Mini-app nhóm',
    galleryLabel: 'Ảnh chụp các mini-app',
    screen: 'Màn hình',
    rolesTitle: 'Một ứng dụng, hai vai trò',
    rolesLead: 'Cùng một Keikibook, giao diện tự đổi theo quyền của tài khoản. Máy chủ vẫn kiểm tra quyền ở từng thao tác.',
    rolePhone: 'Giao diện',
    consolesLabel: 'Các khu quản lý',
    securityTitle: 'An toàn từ thiết kế',
    securityLead: 'Mọi giao dịch đi qua một cổng duy nhất, được xác thực và kiểm tra quyền trước khi tới từng dịch vụ.',
    ctaTitle: 'Sẵn sàng thử Keikibook?',
    ctaLead: 'Tải bản Android ngay hôm nay. Bản iOS đang được hoàn thiện.',
    apk: 'Tải APK cho Android',
    apkSoon: 'Android — sắp có link tải',
    iosSoon: 'iOS — sắp ra mắt',
    ctaPhone: 'Màn hình tất cả dịch vụ',
    tagline: 'Mọi dịch vụ, một ứng dụng.',
    ad: 'Quảng cáo',
    quickSearch: 'Tìm nhanh',
    paletteLabel: 'Tìm nhanh mini-app và mục trên trang',
    palettePlaceholder: 'Tìm mini-app hoặc mục trên trang…',
    paletteSections: 'Mục trên trang',
    paletteApps: 'Mini-app',
    paletteEmpty: 'Không có kết quả',
    paletteHint: 'để chọn',
    paletteNav: 'để di chuyển',
    paletteTip: 'Mẹo: nhấn ⌘K hoặc Ctrl K ở bất cứ đâu để tìm nhanh.',
    stepsTitle: 'Bắt đầu trong 3 bước',
    stepsLead: 'Từ lúc cài app tới lần thanh toán đầu tiên chỉ mất vài phút.',
    step: 'Bước',
    faqTitle: 'Câu hỏi thường gặp',
    faqLead: 'Những điều mọi người hay hỏi trước khi dùng Keikibook.',
    faqMore: 'Chưa thấy câu trả lời? Tải app và thử ngay — mọi mini-app đều mở được mà không cần cài thêm.',
    footerExplore: 'Khám phá',
    footerHelp: 'Tìm hiểu',
    footerGroups: 'Nhóm mini-app',
  },
  en: {
    title: 'Keikibook — Every service, one app',
    toTop: 'Keikibook — back to top',
    mainNav: 'Main navigation',
    footerNav: 'Footer links',
    openMenu: 'Open menu',
    closeMenu: 'Close menu',
    language: 'Language',
    viewShot: 'View screenshot',
    darkMode: 'Dark mode',
    lightMode: 'Light mode',
    backToTop: 'Back to top',
    searchApps: 'Search mini-apps: wallet, movies, tables…',
    results: (n: number) => `${n} matching mini-app${n === 1 ? '' : 's'}`,
    noResults: 'No mini-apps found',
    noResultsHint: 'Try another keyword or browse them all.',
    clearSearch: 'Clear search',
    showMore: (n: number) => `Show ${n} more`,
    close: 'Close',
    links: { services: 'Services', miniApps: 'Mini-apps', wallet: 'Keiki Wallet', roles: 'Roles', security: 'Security', start: 'Get started', faq: 'FAQ', download: 'Get the app' },
    getApp: 'Get Keikibook',
    heroPill: 'The super app for Vietnam',
    heroTitle: ['Every service,', 'one app.'],
    heroLead: 'E-wallet, food, shopping, movies, tickets, hotels, chat and video calls — 54 mini-apps sharing one account and one wallet.',
    heroAndroid: 'Get it for Android',
    heroExplore: 'Explore services',
    heroPhone: 'Keikibook home screen',
    payOk: 'Payment successful',
    points: '+120 reward points',
    thanks: 'Thanks for using Keikibook',
    servicesTitle: 'One account, a whole world',
    servicesLead: 'Every service is a lightweight mini-app that opens right inside Keikibook — nothing extra to install.',
    miniTitle: '54 mini-apps, ready to go',
    miniLead: 'Each service has its own look and colour, but they all share your account, wallet and reward points.',
    miniGroups: 'Mini-app groups',
    all: 'All',
    groupPhone: 'Mini-apps:',
    galleryLabel: 'Mini-app screenshots',
    screen: 'Screen:',
    rolesTitle: 'One app, two roles',
    rolesLead: 'The same Keikibook adapts its interface to your account’s permissions. The server still checks them on every action.',
    rolePhone: 'Interface:',
    consolesLabel: 'Management consoles',
    securityTitle: 'Secure by design',
    securityLead: 'Every transaction goes through a single gateway and is authenticated and authorised before it reaches a service.',
    ctaTitle: 'Ready to try Keikibook?',
    ctaLead: 'Get the Android build today. The iOS version is on its way.',
    apk: 'Download the Android APK',
    apkSoon: 'Android — download link coming soon',
    iosSoon: 'iOS — coming soon',
    ctaPhone: 'All services screen',
    tagline: 'Every service, one app.',
    ad: 'Advertisement',
    quickSearch: 'Quick search',
    paletteLabel: 'Quick search for mini-apps and page sections',
    palettePlaceholder: 'Search mini-apps or page sections…',
    paletteSections: 'On this page',
    paletteApps: 'Mini-apps',
    paletteEmpty: 'No results',
    paletteHint: 'to select',
    paletteNav: 'to move',
    paletteTip: 'Tip: press ⌘K or Ctrl K anywhere to search.',
    stepsTitle: 'Get started in 3 steps',
    stepsLead: 'From installing the app to your first payment takes just a few minutes.',
    step: 'Step',
    faqTitle: 'Frequently asked questions',
    faqLead: 'What people usually ask before trying Keikibook.',
    faqMore: 'Didn’t find your answer? Get the app and try it — every mini-app opens without installing anything else.',
    footerExplore: 'Explore',
    footerHelp: 'Learn more',
    footerGroups: 'Mini-app groups',
  },
}

type Service = { icon: LucideIcon; name: L; note: L; tone: string; href?: string }

export const services: Service[] = [
  { icon: Wallet, name: { vi: 'Ví Keiki', en: 'Keiki Wallet' }, note: { vi: 'Nạp, chuyển, nhận tiền', en: 'Top up, send, receive' }, tone: 'indigo' },
  { icon: Utensils, name: { vi: 'Food', en: 'Food' }, note: { vi: 'Đặt món từ nhà hàng', en: 'Order from restaurants' }, tone: 'orange' },
  { icon: ShoppingBag, name: { vi: 'Mua sắm', en: 'Shopping' }, note: { vi: 'Hàng chính hãng, giao nhanh', en: 'Genuine goods, fast delivery' }, tone: 'pink' },
  { icon: MessageCircle, name: { vi: 'Chat', en: 'Chat' }, note: { vi: 'Bạn bè và nhóm', en: 'Friends and groups' }, tone: 'teal' },
  { icon: Clapperboard, name: { vi: 'Phim', en: 'Movies' }, note: { vi: 'Gợi ý theo sở thích', en: 'Picked for your taste' }, tone: 'violet' },
  { icon: Ticket, name: { vi: 'Vé sự kiện', en: 'Event tickets' }, note: { vi: 'Concert, hội thảo', en: 'Concerts, conferences' }, tone: 'blue' },
  { icon: Building2, name: { vi: 'Khách sạn', en: 'Hotels' }, note: { vi: 'Đặt phòng giá tốt', en: 'Great room rates' }, tone: 'green' },
  { icon: Car, name: { vi: 'Thuê xe', en: 'Car rental' }, note: { vi: 'Tự lái hoặc có tài xế', en: 'Self-drive or chauffeured' }, tone: 'amber' },
  { icon: Bot, name: { vi: 'Trợ lý AI', en: 'AI assistant' }, note: { vi: 'Hỏi gì cũng được', en: 'Ask it anything' }, tone: 'violet' },
  { icon: LayoutGrid, name: { vi: '54 mini-app', en: '54 mini-apps' }, note: { vi: 'Xem tất cả bên dưới', en: 'See them all below' }, tone: 'slate', href: '#mini-app' },
]

export type MiniApp = { icon: LucideIcon; shot: string; name: L; note: L; desc: L }
export type Group = { id: string; label: L; tone: string; image: string; apps: MiniApp[] }

export const groups: Group[] = [
  {
    id: 'giai-tri',
    label: { vi: 'Giải trí', en: 'Entertainment' },
    tone: 'violet',
    image: '/screens/ticket-event.webp',
    apps: [
      {
        icon: Clapperboard,
        shot: 'movie',
        name: { vi: 'Phim', en: 'Movies' },
        note: { vi: 'Gợi ý theo sở thích, xem ngay', en: 'Personal picks, watch now' },
        desc: {
          vi: 'Kho phim được gợi ý dựa trên lịch sử xem của bạn. Xem trailer, đọc thông tin phim và phát ngay trong app, tiếp tục từ chỗ đang dừng.',
          en: 'A movie library recommended from your watch history. Watch trailers, read details and stream right in the app, resuming where you left off.',
        },
      },
      {
        icon: Ticket,
        shot: 'ticket',
        name: { vi: 'Vé sự kiện', en: 'Event tickets' },
        note: { vi: 'Chọn vé, trả bằng ví, vé QR', en: 'Pick, pay by wallet, QR ticket' },
        desc: {
          vi: 'Tìm concert, hội thảo, chọn loại vé hoặc ghế trên sơ đồ rồi thanh toán bằng Ví Keiki. Vé QR lưu sẵn trong máy, có thể tặng hoặc bán lại cho người khác.',
          en: 'Find concerts and conferences, choose a ticket type or a seat on the map and pay with Keiki Wallet. QR tickets live on your phone and can be gifted or resold.',
        },
      },
      {
        icon: Dices,
        shot: 'lucky',
        name: { vi: 'Vòng quay may mắn', en: 'Lucky wheel' },
        note: { vi: 'Quay mỗi ngày, điểm danh nhận quà', en: 'Daily spins and check-in gifts' },
        desc: {
          vi: 'Mỗi ngày có lượt quay miễn phí để trúng điểm thưởng, voucher hoặc quà. Điểm danh liên tục nhiều ngày để mở phần thưởng lớn hơn.',
          en: 'A free daily spin for reward points, vouchers or gifts. Check in on consecutive days to unlock bigger rewards.',
        },
      },
      {
        icon: Trophy,
        shot: 'lottery',
        name: { vi: 'Xổ số', en: 'Lottery' },
        note: { vi: 'Kết quả và dò vé', en: 'Results and ticket check' },
        desc: {
          vi: 'Xem kết quả xổ số các miền theo ngày và dò vé của bạn chỉ bằng cách nhập dãy số.',
          en: 'See daily lottery results for every region and check your ticket by entering its number.',
        },
      },
      {
        icon: TrendingUp,
        shot: 'polymarket',
        name: { vi: 'Thị trường dự đoán', en: 'Prediction market' },
        note: { vi: 'Đặt cửa Có/Không cho sự kiện', en: 'Trade Yes/No on events' },
        desc: {
          vi: 'Mua cửa Có hoặc Không cho các sự kiện thể thao, kinh tế, công nghệ. Theo dõi sổ lệnh, biểu đồ giá và nhận thưởng khi sự kiện có kết quả.',
          en: 'Buy Yes or No shares on sports, economic and tech events. Follow the order book and price chart, and get paid when the event resolves.',
        },
      },
      {
        icon: MonitorPlay,
        shot: 'video',
        name: { vi: 'Video', en: 'Video' },
        note: { vi: 'Xem video HLS chất lượng cao', en: 'High-quality HLS streaming' },
        desc: {
          vi: 'Trình phát video HLS tự điều chỉnh chất lượng theo tốc độ mạng, xem mượt cả khi đang di chuyển.',
          en: 'An HLS player that adapts quality to your connection, so playback stays smooth even on the move.',
        },
      },
      {
        icon: Youtube,
        shot: 'media',
        name: { vi: 'Video YouTube', en: 'YouTube videos' },
        note: { vi: 'Thư viện video theo chủ đề', en: 'Videos sorted by topic' },
        desc: {
          vi: 'Thư viện video tuyển chọn được chia theo chủ đề, xem ngay trong Keikibook mà không cần mở ứng dụng khác.',
          en: 'A curated video library organised by topic, played inside Keikibook without switching apps.',
        },
      },
      {
        icon: Radio,
        shot: 'livestream',
        name: { vi: 'Livestream', en: 'Livestream' },
        note: { vi: 'Xem trực tiếp, mở phòng live', en: 'Watch or go live' },
        desc: {
          vi: 'Xem các buổi phát trực tiếp đang diễn ra hoặc tự mở phòng live của riêng bạn để chia sẻ với mọi người.',
          en: 'Watch live broadcasts as they happen or open your own live room to stream to others.',
        },
      },
      {
        icon: BookOpen,
        shot: 'book',
        name: { vi: 'Đọc sách', en: 'Reader' },
        note: { vi: 'Đọc theo chương, chỉnh cỡ chữ', en: 'Read by chapter, adjust font' },
        desc: {
          vi: 'Đọc sách và truyện theo từng chương, chỉnh cỡ chữ cho dễ nhìn và tự lưu vị trí đang đọc.',
          en: 'Read books and stories chapter by chapter, resize the text for comfort and keep your place automatically.',
        },
      },
      {
        icon: Library,
        shot: 'book_store',
        name: { vi: 'Nhà sách', en: 'Bookstore' },
        note: { vi: 'Mua sách, tác giả nổi bật', en: 'Buy books, featured authors' },
        desc: {
          vi: 'Duyệt sách theo thể loại, khám phá tác giả nổi bật và mua sách in hoặc sách điện tử bằng Ví Keiki.',
          en: 'Browse books by genre, discover featured authors and buy print or e-books with Keiki Wallet.',
        },
      },
    ],
  },
  {
    id: 'du-lich',
    label: { vi: 'Du lịch & đi lại', en: 'Travel & transport' },
    tone: 'green',
    image: '/screens/reservation-hotel.webp',
    apps: [
      {
        icon: House,
        shot: 'rent_house',
        name: { vi: 'Homestay', en: 'Homestay' },
        note: { vi: 'Đặt theo đêm hoặc theo tháng', en: 'Nightly or monthly stays' },
        desc: {
          vi: 'Tìm homestay theo khu vực, ngày đi và số khách. Thuê theo đêm cho chuyến đi ngắn hoặc theo tháng cho kỳ ở dài.',
          en: 'Find homestays by area, dates and number of guests. Book by the night for short trips or by the month for longer stays.',
        },
      },
      {
        icon: Building2,
        shot: 'reservation',
        name: { vi: 'Khách sạn', en: 'Hotels' },
        note: { vi: 'Báo giá theo ngày, huỷ miễn phí', en: 'Daily quotes, free cancellation' },
        desc: {
          vi: 'So sánh giá thật của từng loại phòng theo ngày, giữ phòng trước và trả bằng ví. Huỷ trong hạn được hoàn tiền theo chính sách của khách sạn.',
          en: 'Compare real prices for each room type by date, hold a room and pay by wallet. Cancel in time for a refund under the hotel’s policy.',
        },
      },
      {
        icon: Car,
        shot: 'car_rental',
        name: { vi: 'Thuê xe', en: 'Vehicle rental' },
        note: { vi: 'Ô tô, xe máy, xe đạp', en: 'Cars, motorbikes, bicycles' },
        desc: {
          vi: 'Thuê ô tô tự lái hoặc có tài xế, xe máy và xe đạp bằng tài khoản Keikibook. Nhận và trả xe tại quầy giao xe.',
          en: 'Rent self-drive or chauffeured cars, motorbikes and bicycles with your Keikibook account. Pick up and return at the rental desk.',
        },
      },
      {
        icon: BedDouble,
        shot: 'booking_mini',
        name: { vi: 'Đặt phòng nhanh', en: 'Quick booking' },
        note: { vi: 'Phòng trống trong vài chạm', en: 'Free rooms in a few taps' },
        desc: {
          vi: 'Xem ngay phòng còn trống gần bạn và đặt chỉ trong vài chạm, phù hợp khi cần chỗ nghỉ gấp.',
          en: 'See available rooms near you and book in a few taps — ideal when you need a place to stay right now.',
        },
      },
      {
        icon: SquareParking,
        shot: 'parking_lot',
        name: { vi: 'Bãi đỗ xe', en: 'Parking' },
        note: { vi: 'Chỗ trống, gửi và lấy xe', en: 'Free spots, park and collect' },
        desc: {
          vi: 'Xem số chỗ trống ở các bãi đỗ, gửi xe và lấy xe bằng mã trong app, phí gửi trừ thẳng vào ví.',
          en: 'Check free spaces at car parks, park and collect with an in-app code, and pay the fee straight from your wallet.',
        },
      },
    ],
  },
  {
    id: 'tai-chinh',
    label: { vi: 'Tài chính & ưu đãi', en: 'Finance & rewards' },
    tone: 'indigo',
    image: '/screens/payment-wallet.webp',
    apps: [
      {
        icon: PiggyBank,
        shot: 'payment_wallet',
        name: { vi: 'Ví túi tiền', en: 'Pocket wallet' },
        note: { vi: 'Ngăn tiết kiệm, chuyển, yêu cầu tiền', en: 'Savings pockets, send, request' },
        desc: {
          vi: 'Chia tiền vào các ngăn tiết kiệm theo mục tiêu, chuyển tiền cho bạn bè hoặc gửi yêu cầu thanh toán cho người khác.',
          en: 'Split money into savings pockets for each goal, send money to friends or request payment from others.',
        },
      },
      {
        icon: Gift,
        shot: 'voucher',
        name: { vi: 'Voucher', en: 'Vouchers' },
        note: { vi: 'Nhận quà, kiểm tra và dùng mã', en: 'Claim, check and redeem codes' },
        desc: {
          vi: 'Nhận voucher từ Keikibook và đối tác, kiểm tra điều kiện áp dụng và dùng mã khi thanh toán ở mọi dịch vụ.',
          en: 'Collect vouchers from Keikibook and partners, check their conditions and apply codes at checkout in any service.',
        },
      },
      {
        icon: UsersRound,
        shot: 'referral',
        name: { vi: 'Giới thiệu bạn bè', en: 'Refer a friend' },
        note: { vi: 'Link mời và thưởng theo bậc', en: 'Invite links, tiered rewards' },
        desc: {
          vi: 'Chia sẻ link mời riêng; mỗi người bạn đăng ký và dùng dịch vụ giúp bạn lên bậc và nhận thưởng lớn hơn.',
          en: 'Share your personal invite link; every friend who signs up and uses a service moves you up a tier for bigger rewards.',
        },
      },
      {
        icon: Award,
        shot: 'recompense',
        name: { vi: 'Đổi thưởng', en: 'Redeem' },
        note: { vi: 'Điểm hạng thành viên, ưu đãi tốt nhất', en: 'Member tiers, best deals' },
        desc: {
          vi: 'Theo dõi điểm và hạng thành viên, đổi điểm lấy ưu đãi; hạng càng cao, ưu đãi càng tốt.',
          en: 'Track your points and membership tier and swap points for perks — the higher your tier, the better the deals.',
        },
      },
      {
        icon: Sparkles,
        shot: 'reward',
        name: { vi: 'Phần thưởng', en: 'Rewards' },
        note: { vi: 'Quà tặng đổi bằng điểm', en: 'Gifts paid with points' },
        desc: {
          vi: 'Danh mục quà tặng, thẻ quà và sản phẩm có thể đổi bằng điểm tích được sau mỗi giao dịch.',
          en: 'A catalogue of gifts, gift cards and products you can claim with points earned on every transaction.',
        },
      },
      {
        icon: BadgeCheck,
        shot: 'bonuslink',
        name: { vi: 'Bonuslink', en: 'Bonuslink' },
        note: { vi: 'Điểm thưởng liên kết', en: 'Linked loyalty points' },
        desc: {
          vi: 'Liên kết thẻ Bonuslink để tích và dùng điểm thưởng của chương trình đối tác ngay trong Keikibook.',
          en: 'Link your Bonuslink card to earn and spend partner loyalty points directly in Keikibook.',
        },
      },
      {
        icon: ShieldCheck,
        shot: 'sard_auth',
        name: { vi: 'Xác thực thanh toán', en: 'Payment authentication' },
        note: { vi: 'OTP 3-D Secure cho thẻ', en: '3-D Secure OTP for cards' },
        desc: {
          vi: 'Xác nhận thanh toán thẻ bằng mã OTP 3-D Secure, đảm bảo chỉ chủ thẻ mới hoàn tất được giao dịch.',
          en: 'Confirm card payments with a 3-D Secure OTP so only the cardholder can complete the transaction.',
        },
      },
      {
        icon: CreditCard,
        shot: 'payment',
        name: { vi: 'Thanh toán thẻ', en: 'Card payments' },
        note: { vi: 'Stripe, PayPal, VNPay', en: 'Stripe, PayPal, VNPay' },
        desc: {
          vi: 'Thanh toán bằng thẻ quốc tế qua Stripe, tài khoản PayPal hoặc thẻ nội địa qua VNPay.',
          en: 'Pay with international cards via Stripe, a PayPal account, or domestic cards through VNPay.',
        },
      },
      {
        icon: Wallet,
        shot: 'integrated_payment',
        name: { vi: 'Cổng thanh toán', en: 'Payment gateway' },
        note: { vi: 'MoMo, chuyển khoản, hoàn tiền', en: 'MoMo, bank transfer, refunds' },
        desc: {
          vi: 'Một cổng thanh toán chung hỗ trợ MoMo và chuyển khoản ngân hàng, theo dõi trạng thái đơn và xử lý hoàn tiền.',
          en: 'One shared checkout supporting MoMo and bank transfer, with order status tracking and refund handling.',
        },
      },
      {
        icon: ReceiptText,
        shot: 'billing',
        name: { vi: 'Gói dịch vụ', en: 'Subscriptions' },
        note: { vi: 'Gói tháng, gói năm', en: 'Monthly and yearly plans' },
        desc: {
          vi: 'Đăng ký gói dịch vụ theo tháng hoặc theo năm, xem hoá đơn và gia hạn tự động bằng ví.',
          en: 'Subscribe to monthly or yearly plans, view invoices and renew automatically from your wallet.',
        },
      },
      {
        icon: Sun,
        shot: 'calculate',
        name: { vi: 'Thời tiết & thị trường', en: 'Weather & markets' },
        note: { vi: 'Dự báo, thuỷ triều, giá tiền số', en: 'Forecasts, tides, crypto prices' },
        desc: {
          vi: 'Xem dự báo thời tiết, lịch thuỷ triều và giá tiền số cập nhật liên tục trong một màn hình.',
          en: 'Weather forecasts, tide tables and live crypto prices together on one screen.',
        },
      },
    ],
  },
  {
    id: 'cong-dong',
    label: { vi: 'Cộng đồng', en: 'Community' },
    tone: 'pink',
    image: '/screens/threads.webp',
    apps: [
      {
        icon: AtSign,
        shot: 'threads',
        name: { vi: 'Threads', en: 'Threads' },
        note: { vi: 'Bảng tin, thích, bình luận, lưu', en: 'Feed, like, comment, save' },
        desc: {
          vi: 'Bảng tin ngắn để chia sẻ suy nghĩ và khoảnh khắc; thích, bình luận và lưu những bài bạn muốn đọc lại.',
          en: 'A short-form feed for sharing thoughts and moments; like, comment on and save posts you want to revisit.',
        },
      },
      {
        icon: Newspaper,
        shot: 'post',
        name: { vi: 'Bài viết', en: 'Articles' },
        note: { vi: 'Chuyên mục và bài có ảnh bìa', en: 'Sections with cover images' },
        desc: {
          vi: 'Đọc bài viết dài được chia theo chuyên mục, có ảnh bìa và trình bày dễ đọc trên điện thoại.',
          en: 'Long-form articles grouped into sections, with cover images and a layout built for phones.',
        },
      },
      {
        icon: MessagesSquare,
        shot: 'room',
        name: { vi: 'Phòng chat', en: 'Chat rooms' },
        note: { vi: 'Phòng nhiều người, quản lý thành viên', en: 'Group rooms, member management' },
        desc: {
          vi: 'Tạo phòng trò chuyện nhiều người theo chủ đề, mời, phân quyền và quản lý thành viên.',
          en: 'Create topic-based group rooms, invite people and manage members and their permissions.',
        },
      },
      {
        icon: Video,
        shot: 'webrtc',
        name: { vi: 'Gọi video', en: 'Video calls' },
        note: { vi: 'Họp nhóm tới 8 người', en: 'Group calls up to 8 people' },
        desc: {
          vi: 'Gọi video nhóm tới 8 người qua WebRTC; đổi camera trước/sau, tắt mic hoặc camera chỉ với một chạm.',
          en: 'WebRTC group video calls for up to 8 people; switch cameras or mute your mic and camera in one tap.',
        },
      },
      {
        icon: MessageCircleQuestion,
        shot: 'share',
        name: { vi: 'Hỏi & chia sẻ', en: 'Ask & share' },
        note: { vi: 'Hỏi đáp cùng AI', en: 'Q&A with AI' },
        desc: {
          vi: 'Đặt câu hỏi cho cộng đồng và AI, nhận câu trả lời gợi ý rồi chia sẻ lại cho mọi người.',
          en: 'Ask the community and AI, get suggested answers and share them with others.',
        },
      },
      {
        icon: Images,
        shot: 'photo',
        name: { vi: 'Ghép ảnh nghệ thuật', en: 'Photo art' },
        note: { vi: 'Lưới, mosaic từ ảnh của bạn', en: 'Grids and mosaics from your photos' },
        desc: {
          vi: 'Biến bộ ảnh của bạn thành ảnh lưới hoặc mosaic nghệ thuật; ảnh lưu trên máy chủ, đổi máy vẫn còn.',
          en: 'Turn your photos into grid or mosaic artworks; results are stored on the server and survive a phone change.',
        },
      },
      {
        icon: Flag,
        shot: 'chat_report',
        name: { vi: 'Báo cáo chat', en: 'Report chat' },
        note: { vi: 'Báo tin rác, lừa đảo', en: 'Report spam and scams' },
        desc: {
          vi: 'Báo cáo tin nhắn rác, lừa đảo hay quấy rối để đội kiểm duyệt xử lý và giữ cộng đồng an toàn.',
          en: 'Report spam, scam or abusive messages so moderators can act and keep the community safe.',
        },
      },
      {
        icon: Briefcase,
        shot: 'recruitment',
        name: { vi: 'Tuyển dụng', en: 'Jobs' },
        note: { vi: 'Ứng tuyển, làm cộng tác viên', en: 'Apply, become a partner' },
        desc: {
          vi: 'Xem tin tuyển dụng, nộp hồ sơ ứng tuyển hoặc đăng ký làm cộng tác viên cho các dịch vụ Keikibook.',
          en: 'Browse job listings, submit applications or sign up as a partner for Keikibook services.',
        },
      },
      {
        icon: Megaphone,
        shot: 'admanagement',
        name: { vi: 'Quảng cáo', en: 'Ads' },
        note: { vi: 'Ưu đãi theo vị trí hiển thị', en: 'Offers by placement' },
        desc: {
          vi: 'Khám phá ưu đãi từ nhãn hàng theo từng vị trí hiển thị; doanh nghiệp có thể tạo và quản lý chiến dịch.',
          en: 'Discover brand offers by placement; businesses can create and manage their own campaigns.',
        },
      },
      {
        icon: KeyRound,
        shot: 'manage',
        name: { vi: 'SSO doanh nghiệp', en: 'Enterprise SSO' },
        note: { vi: 'Đăng nhập cổng đối tác', en: 'Sign in to partner portals' },
        desc: {
          vi: 'Dùng tài khoản Keikibook để đăng nhập một lần vào cổng của doanh nghiệp và đối tác.',
          en: 'Use your Keikibook account for single sign-on to business and partner portals.',
        },
      },
    ],
  },
  {
    id: 'tien-ich',
    label: { vi: 'Tiện ích & ăn uống', en: 'Utilities & dining' },
    tone: 'orange',
    image: '/screens/ekyc.webp',
    apps: [
      {
        icon: IdCard,
        shot: 'ekyc',
        name: { vi: 'Xác thực danh tính', en: 'Identity check' },
        note: { vi: 'eKYC 3 bước bằng CCCD', en: '3-step eKYC with national ID' },
        desc: {
          vi: 'Xác minh danh tính trong 3 bước: chụp hai mặt CCCD và quay khuôn mặt để mở hạn mức giao dịch cao hơn.',
          en: 'Verify your identity in 3 steps — photograph both sides of your ID and scan your face — to unlock higher limits.',
        },
      },
      {
        icon: ScanFace,
        shot: 'face',
        name: { vi: 'Nhận diện khuôn mặt', en: 'Face detection' },
        note: { vi: 'Tải ảnh, phát hiện khuôn mặt', en: 'Upload, detect faces' },
        desc: {
          vi: 'Tải ảnh lên để phát hiện và đánh dấu các khuôn mặt có trong ảnh.',
          en: 'Upload a photo to detect and highlight the faces in it.',
        },
      },
      {
        icon: Dumbbell,
        shot: 'identified',
        name: { vi: 'Phòng gym', en: 'Gym' },
        note: { vi: 'Lịch lớp, mua và kiểm tra thẻ', en: 'Classes, buy and check passes' },
        desc: {
          vi: 'Xem lịch lớp tập, mua thẻ thành viên và kiểm tra hạn thẻ ngay khi đến phòng gym.',
          en: 'See class schedules, buy a membership and check its validity when you arrive at the gym.',
        },
      },
      {
        icon: ChefHat,
        shot: 'restaurant_kitchen',
        name: { vi: 'Bếp & món ăn', en: 'Kitchen & menu' },
        note: { vi: 'Thực đơn theo danh mục', en: 'Menu by category' },
        desc: {
          vi: 'Màn hình cho bếp: thực đơn theo danh mục và các món khách vừa gọi để chế biến theo thứ tự.',
          en: 'The kitchen view: the menu by category and incoming orders lined up for preparation.',
        },
      },
      {
        icon: ConciergeBell,
        shot: 'restaurant_consumer_menu',
        name: { vi: 'Gọi món', en: 'Table ordering' },
        note: { vi: 'Chọn món, gửi thẳng xuống bếp', en: 'Order straight to the kitchen' },
        desc: {
          vi: 'Ngồi tại bàn, quét mã để xem thực đơn, chọn món và gửi thẳng xuống bếp, không cần chờ phục vụ.',
          en: 'Scan the code at your table to see the menu, pick dishes and send them straight to the kitchen — no waiting for staff.',
        },
      },
      {
        icon: Armchair,
        shot: 'restaurant_order',
        name: { vi: 'Đặt bàn', en: 'Table booking' },
        note: { vi: 'Giữ bàn trước khi đến', en: 'Reserve before you arrive' },
        desc: {
          vi: 'Chọn nhà hàng, giờ đến và số người để giữ bàn trước, nhận xác nhận ngay trong app.',
          en: 'Choose a restaurant, time and party size to reserve a table, with instant confirmation in the app.',
        },
      },
      {
        icon: CupSoda,
        shot: 'vending_machine',
        name: { vi: 'Máy bán hàng', en: 'Vending' },
        note: { vi: 'Chọn khe, trả bằng ví', en: 'Pick a slot, pay by wallet' },
        desc: {
          vi: 'Chọn khe hàng trên máy bán hàng tự động và trả bằng Ví Keiki, không cần tiền mặt.',
          en: 'Pick a slot on a vending machine and pay with Keiki Wallet — no cash needed.',
        },
      },
      {
        icon: Store,
        shot: 'shopify',
        name: { vi: 'Shopify', en: 'Shopify' },
        note: { vi: 'Sản phẩm và phiên bản', en: 'Products and variants' },
        desc: {
          vi: 'Duyệt sản phẩm từ các cửa hàng Shopify liên kết, xem từng phiên bản màu sắc, kích cỡ và giá.',
          en: 'Browse products from linked Shopify stores, with every colour, size and price variant.',
        },
      },
      {
        icon: Truck,
        shot: 'shipping',
        name: { vi: 'Vận chuyển', en: 'Shipping' },
        note: { vi: 'Báo giá giao hàng', en: 'Delivery quotes' },
        desc: {
          vi: 'Nhập điểm gửi, điểm nhận và kích thước gói hàng để nhận báo giá giao hàng từ các hãng vận chuyển như GHN.',
          en: 'Enter pickup, drop-off and parcel size to get delivery quotes from carriers such as GHN.',
        },
      },
      {
        icon: MapIcon,
        shot: 'geoservice',
        name: { vi: 'Bản đồ & địa giới', en: 'Maps & regions' },
        note: { vi: 'Tỉnh, quận, phường, toạ độ', en: 'Provinces, districts, wards, coordinates' },
        desc: {
          vi: 'Tra cứu danh sách tỉnh, quận, phường của Việt Nam và chuyển đổi giữa địa chỉ và toạ độ.',
          en: 'Look up Vietnam’s provinces, districts and wards and convert between addresses and coordinates.',
        },
      },
      {
        icon: Search,
        shot: 'search',
        name: { vi: 'Tìm kiếm', en: 'Search' },
        note: { vi: 'Tìm trong toàn bộ Keikibook', en: 'Search all of Keikibook' },
        desc: {
          vi: 'Một ô tìm kiếm cho mọi thứ trong Keikibook: dịch vụ, sản phẩm, phim, bài viết và người dùng.',
          en: 'One search box for everything in Keikibook: services, products, movies, articles and people.',
        },
      },
      {
        icon: Globe,
        shot: 'serpapi',
        name: { vi: 'Tìm kiếm web', en: 'Web search' },
        note: { vi: 'Web, tin tức, hình ảnh', en: 'Web, news, images' },
        desc: {
          vi: 'Tìm kiếm trên Internet ngay trong app với kết quả web, tin tức và hình ảnh.',
          en: 'Search the Internet from inside the app, with web, news and image results.',
        },
      },
      {
        icon: Bot,
        shot: 'chatbot',
        name: { vi: 'Chatbot', en: 'Chatbot' },
        note: { vi: 'Lịch sử hội thoại AI', en: 'AI conversation history' },
        desc: {
          vi: 'Trò chuyện với trợ lý AI để hỏi đáp, gợi ý và hỗ trợ; lịch sử hội thoại được lưu để xem lại.',
          en: 'Chat with an AI assistant for answers, ideas and help; your conversation history is saved for later.',
        },
      },
      {
        icon: BrainCircuit,
        shot: 'ollama_py',
        name: { vi: 'Mô hình AI', en: 'AI models' },
        note: { vi: 'Trò chuyện với Llama, Qwen…', en: 'Chat with Llama, Qwen…' },
        desc: {
          vi: 'Chọn và trò chuyện trực tiếp với các mô hình ngôn ngữ mở như Llama, Qwen để so sánh câu trả lời.',
          en: 'Pick and chat directly with open language models such as Llama and Qwen to compare their answers.',
        },
      },
      {
        icon: Images,
        shot: 'draw_image',
        name: { vi: 'Ghép ảnh', en: 'Collage' },
        note: { vi: 'Collage nhiều cột', en: 'Multi-column collages' },
        desc: {
          vi: 'Ghép nhiều ảnh thành một collage nhiều cột, chọn bố cục rồi tải về hoặc chia sẻ.',
          en: 'Combine several photos into a multi-column collage, choose a layout, then download or share it.',
        },
      },
      {
        icon: Upload,
        shot: 'upload',
        name: { vi: 'Tải tệp', en: 'File upload' },
        note: { vi: 'Lưu và chia sẻ tệp', en: 'Store and share files' },
        desc: {
          vi: 'Tải tệp lên kho lưu trữ của bạn và tạo liên kết để chia sẻ với người khác.',
          en: 'Upload files to your storage and create links to share them with others.',
        },
      },
      {
        icon: Cloud,
        shot: 'cdn',
        name: { vi: 'Tệp tin', en: 'Files' },
        note: { vi: 'Lưu trữ CDN', en: 'CDN storage' },
        desc: {
          vi: 'Quản lý tệp được lưu trên CDN, tải về nhanh từ máy chủ gần bạn nhất.',
          en: 'Manage files stored on a CDN and download them quickly from the server closest to you.',
        },
      },
      {
        icon: Link,
        shot: 'shortlink',
        name: { vi: 'Link rút gọn', en: 'Short links' },
        note: { vi: 'Giải mã deep link', en: 'Resolve deep links' },
        desc: {
          vi: 'Tạo link rút gọn và mở deep link để đi thẳng tới đúng màn hình trong Keikibook.',
          en: 'Create short links and resolve deep links that take you straight to the right screen in Keikibook.',
        },
      },
    ],
  },
]

export const gallery: { src: string; alt: L }[] = [
  { src: '/screens/lucky-game.webp', alt: { vi: 'Vòng quay may mắn', en: 'Lucky wheel' } },
  { src: '/screens/reservation-hotel.webp', alt: { vi: 'Phòng trống của khách sạn', en: 'Available hotel rooms' } },
  { src: '/screens/payment-wallet.webp', alt: { vi: 'Ví túi tiền', en: 'Pocket wallet' } },
  { src: '/screens/voucher.webp', alt: { vi: 'Voucher của tôi', en: 'My vouchers' } },
  { src: '/screens/ekyc.webp', alt: { vi: 'Xác minh danh tính eKYC', en: 'eKYC identity check' } },
  { src: '/screens/restaurant-consumer-menu.webp', alt: { vi: 'Gọi món tại nhà hàng', en: 'Ordering at a restaurant' } },
]

export const stats: { value: string; label: L }[] = [
  { value: '95', label: { vi: 'dịch vụ', en: 'services' } },
  { value: '54', label: { vi: 'mini-app', en: 'mini-apps' } },
  { value: '1', label: { vi: 'ví cho tất cả', en: 'wallet for all' } },
]

type Feature = { id: string; kicker: L; title: L; text: L; points: L[]; image: string }

export const features: Feature[] = [
  {
    id: 'vi',
    kicker: { vi: 'Ví Keiki', en: 'Keiki Wallet' },
    title: { vi: 'Tiền của bạn, trong tầm tay', en: 'Your money, at your fingertips' },
    text: {
      vi: 'Nạp từ ngân hàng, chuyển cho bạn bè bằng mã ví, nhận tiền qua QR và thanh toán mọi dịch vụ trong app. Mỗi giao dịch đều có mã PIN và khoá chống trừ tiền hai lần.',
      en: 'Top up from your bank, send to friends by wallet ID, receive by QR and pay for every service in the app. Every transaction is PIN-protected and guarded against double charges.',
    },
    points: [
      { vi: 'Thẻ ảo miễn phí cho thanh toán', en: 'Free virtual card for payments' },
      { vi: 'Lịch sử giao dịch rõ ràng', en: 'Clear transaction history' },
      { vi: 'Điểm thưởng sau mỗi lần dùng', en: 'Reward points every time you pay' },
    ],
    image: '/screens/wallet.webp',
  },
  {
    id: 'food',
    kicker: { vi: 'Keiki Food', en: 'Keiki Food' },
    title: { vi: 'Món ngon gần bạn', en: 'Great food near you' },
    text: {
      vi: 'Chọn nhà hàng, thêm món vào giỏ và đặt chỉ trong vài chạm. Tiền chỉ bị trừ khi đơn được giao; huỷ đơn là hoàn ngay.',
      en: 'Pick a restaurant, add dishes to your cart and order in a few taps. You are only charged when the order is delivered; cancel and you are refunded instantly.',
    },
    points: [
      { vi: 'Theo dõi đơn theo thời gian thực', en: 'Real-time order tracking' },
      { vi: 'Thanh toán bằng thẻ Ví Keiki', en: 'Pay with your Keiki Wallet card' },
      { vi: 'Tip cho tài xế ngay trong app', en: 'Tip your driver in the app' },
    ],
    image: '/screens/food.webp',
  },
  {
    id: 'shop',
    kicker: { vi: 'Keiki Shop', en: 'Keiki Shop' },
    title: { vi: 'Mua sắm không lo hết hàng', en: 'Shopping without stock surprises' },
    text: {
      vi: 'Giá và tồn kho luôn được kiểm tra trên máy chủ. Khi thanh toán, hàng được giữ cho bạn trước khi trừ tiền.',
      en: 'Prices and stock are always checked on the server. At checkout your items are reserved for you before any money is taken.',
    },
    points: [
      { vi: 'Giỏ hàng đồng bộ mọi thiết bị', en: 'Cart synced across devices' },
      { vi: 'Mã giảm giá và danh sách yêu thích', en: 'Discount codes and wishlists' },
      { vi: 'Thông báo khi đơn được xử lý', en: 'Notifications as your order progresses' },
    ],
    image: '/screens/shop.webp',
  },
  {
    id: 'giai-tri',
    kicker: { vi: 'Giải trí', en: 'Entertainment' },
    title: { vi: 'Xem phim, săn vé, thử vận may', en: 'Watch, get tickets, try your luck' },
    text: {
      vi: 'Phim được gợi ý theo những gì bạn đã xem và phát ngay trong app. Đặt vé concert, chọn loại vé rồi trả bằng Ví Keiki; vé QR nằm sẵn trong máy, tặng hay bán lại đều được.',
      en: 'Movies are recommended from what you have watched and play right in the app. Book concert tickets, pick a type and pay with Keiki Wallet; QR tickets stay on your phone, ready to gift or resell.',
    },
    points: [
      { vi: 'Phát video và livestream ngay trong app', en: 'Videos and livestreams in the app' },
      { vi: 'Vòng quay, điểm danh mỗi ngày nhận điểm', en: 'Daily spins and check-ins earn points' },
      { vi: 'Đọc sách theo chương, dò vé số', en: 'Read by chapter, check lottery tickets' },
    ],
    image: '/screens/movie.webp',
  },
  {
    id: 'du-lich',
    kicker: { vi: 'Du lịch', en: 'Travel' },
    title: { vi: 'Chỗ ở và xe cho mọi chuyến đi', en: 'Stays and rides for every trip' },
    text: {
      vi: 'Tìm homestay, khách sạn theo ngày đi và số khách, xem giá thật cho từng loại phòng. Giữ chỗ trước, trả bằng ví; huỷ trong hạn là được hoàn theo chính sách của nơi ở.',
      en: 'Find homestays and hotels by dates and guests, with real prices for every room type. Reserve ahead and pay by wallet; cancel in time and get refunded under the property’s policy.',
    },
    points: [
      { vi: 'Homestay theo đêm hoặc theo tháng', en: 'Homestays by the night or month' },
      { vi: 'Thuê ô tô, xe máy bằng tài khoản Keikibook', en: 'Rent cars and motorbikes with your Keikibook account' },
      { vi: 'Gửi xe và lấy xe ở bãi đỗ', en: 'Park and collect at car parks' },
    ],
    image: '/screens/rent-house.webp',
  },
  {
    id: 'ket-noi',
    kicker: { vi: 'Kết nối', en: 'Connect' },
    title: { vi: 'Tin nhắn đến ngay, gọi video một chạm', en: 'Instant messages, one-tap video calls' },
    text: {
      vi: 'Chat với bạn bè và nhóm nhận tin theo thời gian thực, tự nối lại khi mạng chập chờn. Tạo phòng họp video cho tới 8 người, hoặc chia sẻ khoảnh khắc trên Threads.',
      en: 'Chat with friends and groups in real time, reconnecting automatically on a shaky network. Start a video room for up to 8 people, or share moments on Threads.',
    },
    points: [
      { vi: 'Chat realtime, không sót tin khi mất mạng', en: 'Real-time chat that never drops messages' },
      { vi: 'Gọi video nhóm, đổi camera, tắt mic', en: 'Group video, switch camera, mute' },
      { vi: 'Threads, bài viết và phòng chat', en: 'Threads, articles and chat rooms' },
    ],
    image: '/screens/threads.webp',
  },
]

type Role = { id: string; icon: LucideIcon; kicker: L; title: L; text: L; points: L[]; image: string }

export const roles: Role[] = [
  {
    id: 'khach-hang',
    icon: UserRound,
    kicker: { vi: 'Khách hàng', en: 'Customer' },
    title: { vi: 'Dùng dịch vụ, không thấy gì thừa', en: 'Just the services, nothing extra' },
    text: {
      vi: 'Mọi tài khoản mới là khách hàng. Ứng dụng chỉ hiện những gì bạn cần: mua, đặt, xem, trò chuyện và quản lý ví của chính mình.',
      en: 'Every new account is a customer. The app shows only what you need: buy, book, watch, chat and manage your own wallet.',
    },
    points: [
      { vi: 'Ba tab gọn: Trang chủ, Dịch vụ, Tài khoản', en: 'Three tidy tabs: Home, Services, Account' },
      { vi: '54 mini-app cho nhu cầu hằng ngày', en: '54 mini-apps for everyday needs' },
      { vi: 'Lịch sử, vé, đơn hàng gắn với tài khoản', en: 'History, tickets and orders tied to your account' },
    ],
    image: '/screens/home.webp',
  },
  {
    id: 'quan-tri',
    icon: ShieldHalf,
    kicker: { vi: 'Quản trị', en: 'Admin' },
    title: { vi: 'Một tab để vận hành cả nền tảng', en: 'One tab to run the whole platform' },
    text: {
      vi: 'Tài khoản được cấp quyền admin có thêm tab Quản trị: hàng chờ duyệt, khu quản lý cho ban tổ chức, khách sạn, homestay, voucher doanh nghiệp, quầy giao xe và mọi chức năng quản trị của từng dịch vụ.',
      en: 'Accounts granted admin rights get an extra Admin tab: approval queues, consoles for organisers, hotels, homestays, business vouchers, the rental desk and every service’s admin tools.',
    },
    points: [
      { vi: 'Duyệt sự kiện, khách sạn, homestay mới', en: 'Approve new events, hotels and homestays' },
      { vi: 'Tạo sự kiện, loại vé, sơ đồ ghế, khuyến mãi, soát vé', en: 'Create events, ticket types, seat maps, promos; scan tickets' },
      { vi: 'Nút "Xem như khách" để kiểm tra giao diện khách hàng', en: '"View as customer" to check the customer experience' },
    ],
    image: '/screens/admin-home.webp',
  },
]

export const consoles: { src: string; alt: L }[] = [
  { src: '/screens/ticket-organizer.webp', alt: { vi: 'Ban tổ chức: sự kiện của tôi', en: 'Organiser: my events' } },
  { src: '/screens/ticket-manage-event.webp', alt: { vi: 'Quản lý sự kiện: doanh thu, loại vé', en: 'Event console: revenue, ticket types' } },
  { src: '/screens/poly-market.webp', alt: { vi: 'Thị trường dự đoán: sổ lệnh, biểu đồ giá', en: 'Prediction market: order book, price chart' } },
]

export const security: { icon: LucideIcon; title: L; text: L }[] = [
  {
    icon: Lock,
    title: { vi: 'Mã PIN cho mọi giao dịch', en: 'A PIN for every transaction' },
    text: { vi: 'Chuyển, rút và thanh toán đều cần PIN của riêng bạn.', en: 'Transfers, withdrawals and payments all require your own PIN.' },
  },
  {
    icon: KeyRound,
    title: { vi: 'Phiên đăng nhập an toàn', en: 'Secure sessions' },
    text: { vi: 'Token lưu trong bộ nhớ bảo mật của máy và tự làm mới.', en: 'Tokens live in the device’s secure storage and refresh automatically.' },
  },
  {
    icon: ShieldCheck,
    title: { vi: 'Quyền theo vai trò', en: 'Role-based access' },
    text: {
      vi: 'Khách hàng và admin thấy giao diện khác nhau; quyền được cấp ở authorize-service và kiểm tra lại trên máy chủ.',
      en: 'Customers and admins see different interfaces; permissions are granted by authorize-service and re-checked on the server.',
    },
  },
  {
    icon: Fingerprint,
    title: { vi: 'Xác thực danh tính', en: 'Identity verification' },
    text: { vi: 'eKYC bằng CCCD và khuôn mặt khi bạn cần hạn mức cao hơn.', en: 'eKYC with your national ID and face when you need higher limits.' },
  },
  {
    icon: MessageCircle,
    title: { vi: 'Chat và gọi video đúng người', en: 'Chat and calls with the right people' },
    text: {
      vi: 'Máy chủ tự gắn danh tính từ phiên đăng nhập, không ai mạo danh được bạn.',
      en: 'The server attaches your identity from your session, so nobody can impersonate you.',
    },
  },
  {
    icon: History,
    title: { vi: 'Dữ liệu theo tài khoản', en: 'Data tied to your account' },
    text: {
      vi: 'Đặt phòng, giao dịch, ảnh ghép lưu trên máy chủ; đổi điện thoại vẫn còn nguyên.',
      en: 'Bookings, transactions and collages are stored on the server; switch phones and everything is still there.',
    },
  },
]

export const steps: { icon: LucideIcon; title: L; text: L }[] = [
  {
    icon: Smartphone,
    title: { vi: 'Tải Keikibook', en: 'Get Keikibook' },
    text: { vi: 'Cài bản Android ngay từ trang này. Bản iOS đang được hoàn thiện.', en: 'Install the Android build right from this page. The iOS version is on its way.' },
  },
  {
    icon: UserPlus,
    title: { vi: 'Tạo tài khoản', en: 'Create an account' },
    text: {
      vi: 'Đăng ký một lần, dùng cho cả 54 mini-app. Xác minh eKYC khi bạn cần hạn mức cao hơn.',
      en: 'Sign up once for all 54 mini-apps. Verify with eKYC when you need higher limits.',
    },
  },
  {
    icon: Wallet,
    title: { vi: 'Nạp ví và dùng', en: 'Top up and go' },
    text: {
      vi: 'Nạp từ ngân hàng vào Ví Keiki, rồi đặt món, mua sắm, đặt vé chỉ với mã PIN của bạn.',
      en: 'Top up Keiki Wallet from your bank, then order food, shop and book tickets with just your PIN.',
    },
  },
]

export const faq: { q: L; a: L }[] = [
  {
    q: { vi: 'Có cần cài riêng từng mini-app không?', en: 'Do I need to install each mini-app?' },
    a: {
      vi: 'Không. Cả 54 mini-app mở ngay trong Keikibook và dùng chung một tài khoản, một ví, một kho điểm thưởng.',
      en: 'No. All 54 mini-apps open right inside Keikibook and share one account, one wallet and one pool of reward points.',
    },
  },
  {
    q: { vi: 'Keikibook có bản iOS chưa?', en: 'Is there an iOS version?' },
    a: { vi: 'Hiện có bản Android. Bản iOS đang được hoàn thiện và sẽ ra mắt sớm.', en: 'Android is available now. The iOS version is being finished and is coming soon.' },
  },
  {
    q: { vi: 'Nạp tiền vào Ví Keiki thế nào?', en: 'How do I top up Keiki Wallet?' },
    a: {
      vi: 'Nạp từ tài khoản ngân hàng ngay trong app. Bạn cũng có thể nhận tiền từ bạn bè qua mã ví hoặc mã QR.',
      en: 'Top up from your bank account inside the app. You can also receive money from friends by wallet ID or QR code.',
    },
  },
  {
    q: { vi: 'Thanh toán có an toàn không?', en: 'Are payments safe?' },
    a: {
      vi: 'Mọi giao dịch cần mã PIN của riêng bạn và có khoá chống trừ tiền hai lần. Yêu cầu đi qua một cổng duy nhất, được xác thực và kiểm tra quyền trước khi tới từng dịch vụ.',
      en: 'Every transaction needs your own PIN and is guarded against double charges. Requests pass through a single gateway and are authenticated and authorised before reaching any service.',
    },
  },
  {
    q: { vi: 'Đặt đồ ăn rồi huỷ thì sao?', en: 'What if I cancel a food order?' },
    a: {
      vi: 'Tiền chỉ bị trừ khi đơn được giao. Huỷ đơn trước đó là được hoàn ngay vào ví.',
      en: 'You are only charged when the order is delivered. Cancel before that and the money goes straight back to your wallet.',
    },
  },
  {
    q: { vi: 'Đổi điện thoại có mất dữ liệu không?', en: 'Will I lose my data if I switch phones?' },
    a: {
      vi: 'Không. Đặt phòng, giao dịch, vé và đơn hàng đều lưu trên máy chủ theo tài khoản; đăng nhập trên máy mới là có lại đủ.',
      en: 'No. Bookings, transactions, tickets and orders are stored on the server with your account; sign in on the new phone and it is all there.',
    },
  },
]
