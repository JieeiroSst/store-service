import 'package:flutter/widgets.dart';

import 'community/community_apps.dart';
import 'entertainment/books_apps.dart';
import 'finance/finance_apps.dart';
import 'entertainment/lottery_app.dart';
import 'entertainment/lucky_app.dart';
import 'entertainment/media_apps.dart';
import 'entertainment/movie_app.dart';
import 'entertainment/polymarket_app.dart';
import 'entertainment/ticket_app.dart';
import 'travel/mobility_apps.dart';
import 'travel/stays_apps.dart';
import 'utility/restaurant_apps.dart';
import 'utility/utility_apps.dart';

final serviceApps = <String, Widget Function()>{
  'movie': () => const MovieApp(),
  'ticket': () => const TicketApp(),
  'lucky': () => const LuckyApp(),
  'lottery': () => const LotteryApp(),
  'polymarket': () => const PolymarketApp(),
  'video': () => const VideoApp(),
  'media': () => const MediaApp(),
  'livestream': () => const LivestreamApp(),
  'book': () => const BookApp(),
  'book_store': () => const BookStoreApp(),
  'rent_house': () => const RentHouseApp(),
  'reservation': () => const ReservationApp(),
  'car_rental': () => const CarRentalApp(),
  'booking_mini': () => const BookingMiniApp(),
  'parking_lot': () => const ParkingApp(),
  'billing': () => const BillingApp(),
  'bonuslink': () => const BonuslinkApp(),
  'calculate': () => const CalculateApp(),
  'integrated_payment': () => const IntegratedPaymentApp(),
  'payment': () => const PaymentApp(),
  'payment_wallet': () => const PaymentWalletApp(),
  'recompense': () => const RecompenseApp(),
  'reward': () => const RewardApp(),
  'voucher': () => const VoucherApp(),
  'referral': () => const ReferralApp(),
  'sard_auth': () => const SardAuthApp(),
  'threads': () => const ThreadsApp(),
  'post': () => const PostApp(),
  'room': () => const RoomApp(),
  'share': () => const ShareApp(),
  'photo': () => const PhotoApp(),
  'chat': () => const ChatReportApp(),
  'webrtc': () => const WebrtcApp(),
  'recruitment': () => const RecruitmentApp(),
  'admanagement': () => const AdManagementApp(),
  'manage': () => const ManageApp(),
  'chatbot': () => const ChatbotApp(),
  'ollama_py': () => const OllamaPyApp(),
  'draw_image': () => const DrawImageApp(),
  'ekyc': () => const EkycApp(),
  'face': () => const FaceApp(),
  'identified': () => const GymApp(),
  'search': () => const SearchApp(),
  'serpapi': () => const SerpApp(),
  'shortlink': () => const ShortlinkApp(),
  'upload': () => const UploadApp(),
  'cdn': () => const CdnApp(),
  'geoservice': () => const GeoApp(),
  'shipping': () => const ShippingApp(),
  'shopify': () => const ShopifyApp(),
  'restaurant_kitchen': () => const RestaurantKitchenApp(),
  'restaurant_consumer': () => const RestaurantConsumerApp(),
  'restaurant_order': () => const RestaurantOrderApp(),
  'vending': () => const VendingApp(),
};
