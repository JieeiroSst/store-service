import 'package:flutter_test/flutter_test.dart';
import 'package:keikiapp/features/address/view/ghn_area_picker.dart';
import 'package:keikiapp/gen/gateway_api.g.dart';

class _Ghn implements GatewayExecutor {
  @override
  Future<Map<String, dynamic>> public(String document, {Map<String, dynamic>? variables}) async {
    if (document.contains('shipProvinces')) {
      return {
        'shipProvinces': [
          {'id': 201, 'name': 'Hà Nội'},
          {'id': 202, 'name': 'Hồ Chí Minh'},
          {'id': 203, 'name': 'Đà Nẵng'},
        ],
      };
    }
    if (document.contains('shipDistricts')) {
      return {
        'shipDistricts': [
          {'id': 1442, 'name': 'Quận 1'},
          {'id': 1443, 'name': 'Quận 10'},
          {'id': 1444, 'name': 'Thành phố Thủ Đức'},
        ],
      };
    }
    return {
      'shipWards': [
        {'code': '20109', 'name': 'Phường Bến Nghé'},
        {'code': '20110', 'name': 'Phường Bến Thành'},
      ],
    };
  }

  @override
  Future<Map<String, dynamic>> query(String document, {Map<String, dynamic>? variables}) => public(document, variables: variables);

  @override
  Future<Map<String, dynamic>> mutation(String document, {Map<String, dynamic>? variables}) => public(document, variables: variables);
}

void main() {
  final api = GatewayApi(_Ghn()).shipping;

  test('area names match without accents and prefixes', () {
    expect(areaKey('Thành phố Hồ Chí Minh'), 'ho chi minh');
    expect(areaKey('TP. Hồ Chí Minh'), 'ho chi minh');
    expect(areaKey('Quận 01'), '1');
    expect(areaKey('P. Bến Nghé'), 'ben nghe');
    expect(foldVi('Đà Nẵng'), 'da nang');
  });

  test('a full old address resolves to GHN codes', () async {
    final a = await guessGhnArea(api, province: 'TP. Hồ Chí Minh', district: 'Quan 1', ward: 'Bến Nghé');
    expect(a.code, 'ghn:202:1442:20109');
    expect(a.wardName, 'Phường Bến Nghé');
  });

  test('ambiguous or unknown parts stop the guess where it is unsure', () async {
    final partial = await guessGhnArea(api, province: 'Hồ Chí Minh', district: 'Quận 1', ward: 'Bến');
    expect(partial.districtId, 1442);
    expect(partial.wardCode, isNull);
    expect((await guessGhnArea(api, province: 'Sài Gòn')).provinceId, isNull);
    expect((await guessGhnArea(api, province: '')).provinceId, isNull);
  });
}
