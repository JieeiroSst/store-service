import 'batch1.dart';
import 'batch2.dart';
import 'batch3.dart';
import 'batch4.dart';
import 'batch5.dart';
import 'batch6.dart';
import 'batch7.dart';

export 'batch6.dart' show pageShots;

class Shot {
  const Shot(this.name, this.route, {this.tap, this.tab});

  final String name;
  final String route;
  final String? tap;
  final String? tab;
}

final fixtures = <String, Object?>{...batch1, ...batch2, ...batch3, ...batch4, ...batch5, ...batch6, ...batch7};

final shots = <Shot>[...batch1Shots, ...batch2Shots, ...batch3Shots, ...batch4Shots, ...batch5Shots, ...batch7Shots];
