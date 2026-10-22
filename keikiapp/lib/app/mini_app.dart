import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';

import '../features/services/data/service_meta.dart';

class MiniApp {
  const MiniApp({
    required this.id,
    required this.title,
    required this.icon,
    required this.color,
    required this.path,
    this.routes = const [],
  });

  factory MiniApp.service(String serviceId, {String? title}) {
    final meta = metaFor(serviceId);
    return MiniApp(
      id: serviceId,
      title: title ?? meta.title,
      icon: meta.icon,
      color: meta.category.color,
      path: '/services/$serviceId',
    );
  }

  final String id;
  final String title;
  final IconData icon;
  final Color color;
  final String path;
  final List<RouteBase> routes;
}
