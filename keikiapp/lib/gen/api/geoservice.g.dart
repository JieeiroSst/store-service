// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

import '../gateway_executor.dart';
import '../models.g.dart';

class GeoserviceApi {
  const GeoserviceApi(this._executor);

  final GatewayExecutor _executor;

  Future<GeoLocationResult?> geoLocate({required double lng, required double lat, double? tolerance, String? selectionSet}) async {
    final data = await _executor.public('query geoLocate(\$lng: Float!, \$lat: Float!, \$tolerance: Float) { geoLocate(lng: \$lng, lat: \$lat, tolerance: \$tolerance) ${selectionSet ?? '{ found province { code name name_full name_en type region slug gid area_km2 } district { code province_code name name_full name_en type slug gid area_km2 } ward { code district_code province_code name name_full name_en type slug gid area_km2 } street { id province_code district_code name name_full name_en slug gid length_m distance_m } point { lng lat } distance_m }'} }', variables: {
      'lng': lng,
      'lat': lat,
      'tolerance': ?tolerance,
    }
    );
    return data['geoLocate'] == null ? null : GeoLocationResult.fromJson((data['geoLocate'] as Map).cast<String, dynamic>());
  }

  Future<List<GeoProvince>?> geoProvinces({String? selectionSet}) async {
    final data = await _executor.public('query geoProvinces { geoProvinces ${selectionSet ?? '{ code name name_full name_en type region slug gid area_km2 }'} }'
    );
    return (data['geoProvinces'] as List?)?.map((e) => (e == null ? null : GeoProvince.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<GeoProvince?> geoProvince({required String code, String? selectionSet}) async {
    final data = await _executor.public('query geoProvince(\$code: String!) { geoProvince(code: \$code) ${selectionSet ?? '{ code name name_full name_en type region slug gid area_km2 }'} }', variables: {
      'code': code,
    }
    );
    return data['geoProvince'] == null ? null : GeoProvince.fromJson((data['geoProvince'] as Map).cast<String, dynamic>());
  }

  Future<List<GeoDistrict>?> geoDistricts({required String code, String? selectionSet}) async {
    final data = await _executor.public('query geoDistricts(\$code: String!) { geoDistricts(code: \$code) ${selectionSet ?? '{ code province_code name name_full name_en type slug gid area_km2 }'} }', variables: {
      'code': code,
    }
    );
    return (data['geoDistricts'] as List?)?.map((e) => (e == null ? null : GeoDistrict.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<List<GeoWard>?> geoWards({required String code, String? selectionSet}) async {
    final data = await _executor.public('query geoWards(\$code: String!) { geoWards(code: \$code) ${selectionSet ?? '{ code district_code province_code name name_full name_en type slug gid area_km2 }'} }', variables: {
      'code': code,
    }
    );
    return (data['geoWards'] as List?)?.map((e) => (e == null ? null : GeoWard.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<List<GeoStreet>?> geoStreets({required String code, String? selectionSet}) async {
    final data = await _executor.public('query geoStreets(\$code: String!) { geoStreets(code: \$code) ${selectionSet ?? '{ id province_code district_code name name_full name_en slug gid length_m distance_m }'} }', variables: {
      'code': code,
    }
    );
    return (data['geoStreets'] as List?)?.map((e) => (e == null ? null : GeoStreet.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<GeoStats?> geoStats({String? selectionSet}) async {
    final data = await _executor.public('query geoStats { geoStats ${selectionSet ?? '{ provinces districts wards }'} }'
    );
    return data['geoStats'] == null ? null : GeoStats.fromJson((data['geoStats'] as Map).cast<String, dynamic>());
  }

  Future<GeoWhitelistResult?> geoWhitelistCheck({required double lng, required double lat, double? tolerance, String? selectionSet}) async {
    final data = await _executor.public('query geoWhitelistCheck(\$lng: Float!, \$lat: Float!, \$tolerance: Float) { geoWhitelistCheck(lng: \$lng, lat: \$lat, tolerance: \$tolerance) ${selectionSet ?? '{ whitelisted province { code name name_full name_en type region slug gid area_km2 } district { code province_code name name_full name_en type slug gid area_km2 } ward { code district_code province_code name name_full name_en type slug gid area_km2 } street { id province_code district_code name name_full name_en slug gid length_m distance_m } point { lng lat } }'} }', variables: {
      'lng': lng,
      'lat': lat,
      'tolerance': ?tolerance,
    }
    );
    return data['geoWhitelistCheck'] == null ? null : GeoWhitelistResult.fromJson((data['geoWhitelistCheck'] as Map).cast<String, dynamic>());
  }

  Future<List<GeoWhitelistCity>?> geoWhitelistCities({String? selectionSet}) async {
    final data = await _executor.public('query geoWhitelistCities { geoWhitelistCities ${selectionSet ?? '{ province_code note }'} }'
    );
    return (data['geoWhitelistCities'] as List?)?.map((e) => (e == null ? null : GeoWhitelistCity.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

}
