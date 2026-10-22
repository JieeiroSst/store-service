// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

import '../gateway_executor.dart';
import '../models.g.dart';

class CalculateApi {
  const CalculateApi(this._executor);

  final GatewayExecutor _executor;

  Future<CalcForecast?> calcForecast({required double lat, required double lon, String? selectionSet}) async {
    final data = await _executor.public('query calcForecast(\$lat: Float!, \$lon: Float!) { calcForecast(lat: \$lat, lon: \$lon) ${selectionSet ?? '{ coordinate { lat lon } generated_at source hourly { time temperature_c precipitation_mm wind_speed_kmh wind_direction_deg surface_pressure_hpa pressure_msl_hpa } }'} }', variables: {
      'lat': lat,
      'lon': lon,
    }
    );
    return data['calcForecast'] == null ? null : CalcForecast.fromJson((data['calcForecast'] as Map).cast<String, dynamic>());
  }

  Future<CalcTidePrediction?> calcTide({required String station, String? selectionSet}) async {
    final data = await _executor.public('query calcTide(\$station: String!) { calcTide(station: \$station) ${selectionSet ?? '{ station_id source events { time type height_m } }'} }', variables: {
      'station': station,
    }
    );
    return data['calcTide'] == null ? null : CalcTidePrediction.fromJson((data['calcTide'] as Map).cast<String, dynamic>());
  }

  Future<CalcRainRadar?> calcRadar({String? selectionSet}) async {
    final data = await _executor.public('query calcRadar { calcRadar ${selectionSet ?? '{ generated_at host past { time tile_url } nowcast { time tile_url } source }'} }'
    );
    return data['calcRadar'] == null ? null : CalcRainRadar.fromJson((data['calcRadar'] as Map).cast<String, dynamic>());
  }

  Future<CalcWeatherSnapshot?> calcCurrentWeather({required String location, String? selectionSet}) async {
    final data = await _executor.public('query calcCurrentWeather(\$location: String!) { calcCurrentWeather(location: \$location) ${selectionSet ?? '{ location coordinate { lat lon } generated_at current { time temperature_c precipitation_mm wind_speed_kmh wind_direction_deg surface_pressure_hpa pressure_msl_hpa } next_tide { time type height_m } latest_radar { time tile_url } }'} }', variables: {
      'location': location,
    }
    );
    return data['calcCurrentWeather'] == null ? null : CalcWeatherSnapshot.fromJson((data['calcCurrentWeather'] as Map).cast<String, dynamic>());
  }

  Future<List<CalcTrackedLocation>?> calcLocations({String? selectionSet}) async {
    final data = await _executor.public('query calcLocations { calcLocations ${selectionSet ?? '{ name coordinate { lat lon } tide_station_id }'} }'
    );
    return (data['calcLocations'] as List?)?.map((e) => (e == null ? null : CalcTrackedLocation.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<CalcMarketSnapshot?> calcMarketSnapshot({String? selectionSet}) async {
    final data = await _executor.public('query calcMarketSnapshot { calcMarketSnapshot ${selectionSet ?? '{ vs_currency generated_at coins { id symbol name image current_price market_cap market_cap_rank fully_diluted_valuation total_volume high_24h low_24h price_change_24h price_change_percentage_24h market_cap_change_24h market_cap_change_percentage_24h circulating_supply total_supply max_supply ath ath_change_percentage ath_date atl atl_change_percentage atl_date roi { times currency percentage } last_updated } }'} }'
    );
    return data['calcMarketSnapshot'] == null ? null : CalcMarketSnapshot.fromJson((data['calcMarketSnapshot'] as Map).cast<String, dynamic>());
  }

}
