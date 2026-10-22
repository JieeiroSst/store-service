// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

import '../gateway_executor.dart';
import '../models.g.dart';

class LotteryApi {
  const LotteryApi(this._executor);

  final GatewayExecutor _executor;

  Future<LotteryResult?> lotteryResults({required String date, required String province, String? selectionSet}) async {
    final data = await _executor.public('query lotteryResults(\$date: String!, \$province: String!) { lotteryResults(date: \$date, province: \$province) ${selectionSet ?? '{ date region province province_code prizes source }'} }', variables: {
      'date': date,
      'province': province,
    }
    );
    return data['lotteryResults'] == null ? null : LotteryResult.fromJson((data['lotteryResults'] as Map).cast<String, dynamic>());
  }

  Future<LotteryCheck?> lotteryCheck({required String date, required String province, required String ticketNumber, String? selectionSet}) async {
    final data = await _executor.public('query lotteryCheck(\$date: String!, \$province: String!, \$ticket_number: String!) { lotteryCheck(date: \$date, province: \$province, ticket_number: \$ticket_number) ${selectionSet ?? '{ won prizes_won detail { prize matched rule } result_found highest_prize }'} }', variables: {
      'date': date,
      'province': province,
      'ticket_number': ticketNumber,
    }
    );
    return data['lotteryCheck'] == null ? null : LotteryCheck.fromJson((data['lotteryCheck'] as Map).cast<String, dynamic>());
  }

  Future<LotteryScrape?> lotteryScrape({required String date, String? selectionSet}) async {
    final data = await _executor.mutation('mutation lotteryScrape(\$date: String!) { lotteryScrape(date: \$date) ${selectionSet ?? '{ date regions_ok regions_failed provinces_saved }'} }', variables: {
      'date': date,
    }
    );
    return data['lotteryScrape'] == null ? null : LotteryScrape.fromJson((data['lotteryScrape'] as Map).cast<String, dynamic>());
  }

  Future<LotteryBackfill?> lotteryBackfill({required String start, required String end, bool? force, String? selectionSet}) async {
    final data = await _executor.mutation('mutation lotteryBackfill(\$start: String!, \$end: String!, \$force: Boolean) { lotteryBackfill(start: \$start, end: \$end, force: \$force) ${selectionSet ?? '{ start end days_processed days_skipped errors }'} }', variables: {
      'start': start,
      'end': end,
      'force': ?force,
    }
    );
    return data['lotteryBackfill'] == null ? null : LotteryBackfill.fromJson((data['lotteryBackfill'] as Map).cast<String, dynamic>());
  }

}
