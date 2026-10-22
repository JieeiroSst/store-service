import 'package:equatable/equatable.dart';

class Wallet extends Equatable {
  const Wallet({required this.id, required this.currency, this.balance, this.status, this.tier, this.hasPin = false});

  final String id;
  final String currency;
  final String? balance;
  final String? status;
  final String? tier;
  final bool hasPin;

  factory Wallet.fromJson(Map<String, dynamic> json) => Wallet(
        id: json['id'] as String? ?? '',
        currency: json['currency'] as String? ?? '',
        balance: json['balance'] as String?,
        status: json['status'] as String?,
        tier: json['tier'] as String?,
        hasPin: json['has_pin'] as bool? ?? false,
      );

  @override
  List<Object?> get props => [id, currency, balance, status, tier, hasPin];
}

class WalletBalance extends Equatable {
  const WalletBalance({this.balance, this.held, this.available, this.currency});

  final String? balance;
  final String? held;
  final String? available;
  final String? currency;

  factory WalletBalance.fromJson(Map<String, dynamic> json) => WalletBalance(
        balance: json['balance'] as String?,
        held: json['held'] as String?,
        available: json['available'] as String?,
        currency: json['currency'] as String?,
      );

  @override
  List<Object?> get props => [balance, held, available, currency];
}

class WalletTransaction extends Equatable {
  const WalletTransaction({
    required this.id,
    this.type,
    this.status,
    this.amount,
    this.fee,
    this.currency,
    this.description,
    this.walletId,
    this.counterpartyWalletId,
    this.createdAt,
  });

  final String id;
  final String? type;
  final String? status;
  final String? amount;
  final String? fee;
  final String? currency;
  final String? description;
  final String? walletId;
  final String? counterpartyWalletId;
  final String? createdAt;

  bool isIncoming(String myWalletId) {
    final t = type?.toUpperCase() ?? '';
    if (t.contains('TOPUP') || t.contains('TOP_UP') || t.contains('CREDIT') || t.contains('REFUND')) return true;
    if (t.contains('TRANSFER')) return walletId != myWalletId;
    return false;
  }

  static const fields =
      'id type status amount fee currency description wallet_id counterparty_wallet_id created_at';

  factory WalletTransaction.fromJson(Map<String, dynamic> json) => WalletTransaction(
        id: json['id'] as String? ?? '',
        type: json['type'] as String?,
        status: json['status'] as String?,
        amount: json['amount'] as String?,
        fee: json['fee'] as String?,
        currency: json['currency'] as String?,
        description: json['description'] as String?,
        walletId: json['wallet_id'] as String?,
        counterpartyWalletId: json['counterparty_wallet_id'] as String?,
        createdAt: json['created_at'] as String?,
      );

  @override
  List<Object?> get props => [id, status, amount];
}

class WalletCard extends Equatable {
  const WalletCard({required this.id, this.cardNumber = '', this.holderName = '', this.network = '', this.status = ''});

  final String id;
  final String cardNumber;
  final String holderName;
  final String network;
  final String status;

  bool get isActive => status.toUpperCase() == 'ACTIVE';

  String get masked {
    final digits = cardNumber.replaceAll(RegExp(r'\D'), '');
    return digits.length >= 4 ? '•••• ${digits.substring(digits.length - 4)}' : cardNumber;
  }

  static const fields = 'id card_number holder_name network status';

  factory WalletCard.fromJson(Map<String, dynamic> json) => WalletCard(
        id: json['id'] as String? ?? '',
        cardNumber: json['card_number'] as String? ?? '',
        holderName: json['holder_name'] as String? ?? '',
        network: json['network'] as String? ?? '',
        status: json['status'] as String? ?? '',
      );

  @override
  List<Object?> get props => [id, status];
}

class BankAccount extends Equatable {
  const BankAccount({required this.id, this.bankId, this.accountNumber, this.accountName, this.isDefault = false});

  final String id;
  final String? bankId;
  final String? accountNumber;
  final String? accountName;
  final bool isDefault;

  String get label => '${bankId ?? ''} • ${accountNumber ?? ''}';

  factory BankAccount.fromJson(Map<String, dynamic> json) => BankAccount(
        id: json['id'] as String? ?? '',
        bankId: json['bank_id'] as String?,
        accountNumber: json['account_number'] as String?,
        accountName: json['account_name'] as String?,
        isDefault: json['is_default'] as bool? ?? false,
      );

  @override
  List<Object?> get props => [id, bankId, accountNumber, accountName, isDefault];
}

class PaymentRequest extends Equatable {
  const PaymentRequest({
    required this.id,
    this.requesterUserId,
    this.payerUserId,
    this.amount,
    this.currency,
    this.note,
    this.status,
    this.createdAt,
  });

  final String id;
  final String? requesterUserId;
  final String? payerUserId;
  final String? amount;
  final String? currency;
  final String? note;
  final String? status;
  final String? createdAt;

  bool get isPending => (status ?? '').toUpperCase() == 'PENDING';

  static const fields = 'id requester_user_id payer_user_id amount currency note status created_at';

  factory PaymentRequest.fromJson(Map<String, dynamic> json) => PaymentRequest(
        id: json['id'] as String? ?? '',
        requesterUserId: json['requester_user_id'] as String?,
        payerUserId: json['payer_user_id'] as String?,
        amount: json['amount'] as String?,
        currency: json['currency'] as String?,
        note: json['note'] as String?,
        status: json['status'] as String?,
        createdAt: json['created_at'] as String?,
      );

  @override
  List<Object?> get props => [id, status];
}

class PointAccount extends Equatable {
  const PointAccount({required this.balance, required this.totalEarned});

  final int balance;
  final int totalEarned;

  factory PointAccount.fromJson(Map<String, dynamic> json) => PointAccount(
        balance: (json['balance'] as num?)?.toInt() ?? 0,
        totalEarned: (json['total_earned'] as num?)?.toInt() ?? 0,
      );

  @override
  List<Object?> get props => [balance, totalEarned];
}

class QrCode extends Equatable {
  const QrCode({required this.id, this.title, this.content, this.imageBase64, this.imageUrl, this.shortCode});

  final String id;
  final String? title;
  final String? content;
  final String? imageBase64;
  final String? imageUrl;
  final String? shortCode;

  factory QrCode.fromJson(Map<String, dynamic> json) => QrCode(
        id: json['id'] as String? ?? '',
        title: json['title'] as String?,
        content: json['content'] as String?,
        imageBase64: json['image_base64'] as String?,
        imageUrl: json['image_url'] as String?,
        shortCode: json['short_code'] as String?,
      );

  @override
  List<Object?> get props => [id, content];
}
