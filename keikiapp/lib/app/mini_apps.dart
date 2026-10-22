import '../features/chat/chat_mini_app.dart';
import '../features/food/food_mini_app.dart';
import '../features/services/services_mini_app.dart';
import '../features/shopping/shopping_mini_app.dart';
import '../features/wallet/wallet_mini_app.dart';
import 'mini_app.dart';

final miniApps = <MiniApp>[walletMiniApp, foodMiniApp, shoppingMiniApp, chatMiniApp];

final featuredServices = <MiniApp>[
  ...miniApps,
  MiniApp.service('movie'),
  MiniApp.service('ticket'),
  MiniApp.service('reservation'),
  MiniApp.service('rent_house'),
  MiniApp.service('car_rental'),
  MiniApp.service('lucky'),
  MiniApp.service('threads'),
];

final miniAppRoutes = [for (final app in miniApps) ...app.routes, ...servicesRoutes];
