import 'package:intl/intl.dart';
import 'package:uuid/uuid.dart';

String formatMoney(num? amount, [String? currency]) {
  final value = amount ?? 0;
  final code = (currency == null || currency.isEmpty) ? 'VND' : currency.toUpperCase();
  if (code == 'VND') return NumberFormat.currency(locale: 'vi_VN', symbol: '₫', decimalDigits: 0).format(value);
  return NumberFormat.currency(locale: 'vi_VN', name: code, symbol: '$code ', decimalDigits: 2).format(value);
}

String formatMoneyString(String? amount, [String? currency]) => formatMoney(num.tryParse(amount ?? ''), currency);

String formatDate(String? iso) {
  final parsed = DateTime.tryParse(iso ?? '');
  if (parsed == null) return iso ?? '';
  return DateFormat('dd/MM/yyyy HH:mm').format(parsed.toLocal());
}

String newIdempotencyKey() => const Uuid().v4();
