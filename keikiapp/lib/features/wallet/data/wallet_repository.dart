import '../../../core/config/env.dart';
import '../../../core/network/graphql_client.dart';
import '../../../core/utils/formatters.dart';
import 'wallet_models.dart';

class WalletRepository {
  WalletRepository(this._client);

  final GraphQLClient _client;

  static const _walletFields = 'id currency balance status tier has_pin';

  Future<List<Wallet>> myWallets() async {
    final data = await _client.query('query { walletSvcMyWallets { $_walletFields } }');
    return [for (final w in data['walletSvcMyWallets'] as List) Wallet.fromJson(w as Map<String, dynamic>)];
  }

  Future<Wallet> createWallet(String currency) async {
    final data = await _client.mutation('''
      mutation CreateWallet(\$currency: String!) { walletSvcMyCreateWallet(currency: \$currency) { $_walletFields } }
    ''', variables: {'currency': currency});
    return Wallet.fromJson(data['walletSvcMyCreateWallet'] as Map<String, dynamic>);
  }

  Future<WalletBalance?> balance(String walletId) async {
    final data = await _client.query(r'''
      query Balance($id: String!) { walletSvcMyBalance(wallet_id: $id) { balance held available currency } }
    ''', variables: {'id': walletId});
    final b = data['walletSvcMyBalance'];
    return b == null ? null : WalletBalance.fromJson(b as Map<String, dynamic>);
  }

  Future<List<WalletTransaction>> transactions(String walletId, {int page = 1, int pageSize = 20}) async {
    final data = await _client.query('''
      query Txs(\$id: String!, \$page: Int, \$size: Int) {
        walletSvcMyTransactions(wallet_id: \$id, page: \$page, page_size: \$size) { items { ${WalletTransaction.fields} } total }
      }
    ''', variables: {'id': walletId, 'page': page, 'size': pageSize});
    final items = (data['walletSvcMyTransactions'] as Map<String, dynamic>?)?['items'] as List? ?? const [];
    return [for (final t in items) WalletTransaction.fromJson(t as Map<String, dynamic>)];
  }

  Future<List<WalletCard>> cards(String walletId) async {
    final data = await _client.query('''
      query Cards(\$id: String!) { walletSvcMyCards(wallet_id: \$id) { ${WalletCard.fields} } }
    ''', variables: {'id': walletId});
    return [for (final c in data['walletSvcMyCards'] as List) WalletCard.fromJson(c as Map<String, dynamic>)];
  }

  Future<WalletCard> issueVirtualCard(String walletId, String holderName) async {
    final data = await _client.mutation('''
      mutation Issue(\$id: String!, \$bank: String!, \$holder: String!) {
        walletSvcMyIssueCard(wallet_id: \$id, issuer_bank_id: \$bank, network: "VISA", holder_name: \$holder) { ${WalletCard.fields} }
      }
    ''', variables: {'id': walletId, 'bank': Env.walletIssuerBankId, 'holder': holderName});
    return WalletCard.fromJson(data['walletSvcMyIssueCard'] as Map<String, dynamic>);
  }

  Future<List<BankAccount>> bankAccounts() async {
    final data = await _client.query(
        'query { walletSvcMyBankAccounts { id bank_id account_number account_name is_default } }');
    return [for (final a in data['walletSvcMyBankAccounts'] as List) BankAccount.fromJson(a as Map<String, dynamic>)];
  }

  Future<BankAccount> linkBankAccount({required String bankId, required String accountNumber, required String accountName}) async {
    final data = await _client.mutation(r'''
      mutation Link($bank: String!, $number: String!, $name: String!) {
        walletSvcMyLinkBankAccount(bank_id: $bank, account_number: $number, account_name: $name) {
          id bank_id account_number account_name is_default
        }
      }
    ''', variables: {'bank': bankId, 'number': accountNumber, 'name': accountName});
    return BankAccount.fromJson(data['walletSvcMyLinkBankAccount'] as Map<String, dynamic>);
  }

  Future<void> setPin(String walletId, String newPin, {String? currentPin}) async {
    await _client.mutation(r'''
      mutation SetPin($id: String!, $current: String, $pin: String!) {
        walletSvcMySetPin(wallet_id: $id, current_pin: $current, new_pin: $pin)
      }
    ''', variables: {'id': walletId, 'current': currentPin, 'pin': newPin});
  }

  Future<WalletTransaction> topUp({
    required String walletId,
    required String bankAccountId,
    required String amount,
    String? description,
  }) async {
    final data = await _client.mutation('''
      mutation TopUp(\$id: String!, \$bank: String!, \$amount: String!, \$desc: String, \$key: String!) {
        walletSvcMyTopUp(wallet_id: \$id, bank_account_id: \$bank, amount: \$amount, description: \$desc, idempotency_key: \$key) {
          ${WalletTransaction.fields}
        }
      }
    ''', variables: {'id': walletId, 'bank': bankAccountId, 'amount': amount, 'desc': description, 'key': newIdempotencyKey()});
    return WalletTransaction.fromJson(data['walletSvcMyTopUp'] as Map<String, dynamic>);
  }

  Future<WalletTransaction> withdraw({
    required String walletId,
    required String bankAccountId,
    required String amount,
    required String pin,
    String? description,
  }) async {
    final data = await _client.mutation('''
      mutation Withdraw(\$id: String!, \$bank: String!, \$amount: String!, \$pin: String!, \$desc: String, \$key: String!) {
        walletSvcMyWithdraw(wallet_id: \$id, bank_account_id: \$bank, amount: \$amount, pin: \$pin, description: \$desc, idempotency_key: \$key) {
          ${WalletTransaction.fields}
        }
      }
    ''', variables: {
      'id': walletId,
      'bank': bankAccountId,
      'amount': amount,
      'pin': pin,
      'desc': description,
      'key': newIdempotencyKey(),
    });
    return WalletTransaction.fromJson(data['walletSvcMyWithdraw'] as Map<String, dynamic>);
  }

  Future<WalletTransaction> transfer({
    required String fromWalletId,
    String? toWalletId,
    String? toUserId,
    required String amount,
    required String pin,
    String? description,
  }) async {
    final data = await _client.mutation('''
      mutation Transfer(\$from: String!, \$toWallet: String, \$toUser: String, \$amount: String!, \$pin: String!, \$desc: String, \$key: String!) {
        walletSvcMyTransfer(from_wallet_id: \$from, to_wallet_id: \$toWallet, to_user_id: \$toUser, amount: \$amount, pin: \$pin, description: \$desc, idempotency_key: \$key) {
          ${WalletTransaction.fields}
        }
      }
    ''', variables: {
      'from': fromWalletId,
      'toWallet': toWalletId,
      'toUser': toUserId,
      'amount': amount,
      'pin': pin,
      'desc': description,
      'key': newIdempotencyKey(),
    });
    return WalletTransaction.fromJson(data['walletSvcMyTransfer'] as Map<String, dynamic>);
  }

  Future<WalletTransaction> pay({
    required String walletId,
    required String merchantId,
    required String amount,
    required String pin,
    String? description,
    Map<String, dynamic>? metadata,
    String? idempotencyKey,
  }) async {
    final data = await _client.mutation('''
      mutation Pay(\$id: String!, \$merchant: String!, \$amount: String!, \$pin: String!, \$desc: String, \$meta: Map, \$key: String!) {
        walletSvcMyPay(wallet_id: \$id, merchant_id: \$merchant, amount: \$amount, pin: \$pin, description: \$desc, metadata: \$meta, idempotency_key: \$key) {
          ${WalletTransaction.fields}
        }
      }
    ''', variables: {
      'id': walletId,
      'merchant': merchantId,
      'amount': amount,
      'pin': pin,
      'desc': description,
      'meta': metadata,
      'key': idempotencyKey ?? newIdempotencyKey(),
    });
    return WalletTransaction.fromJson(data['walletSvcMyPay'] as Map<String, dynamic>);
  }

  Future<List<PaymentRequest>> paymentRequests({String? role, String? status}) async {
    final data = await _client.query('''
      query Requests(\$role: String, \$status: String) {
        walletSvcMyPaymentRequests(role: \$role, status: \$status, page: 1, page_size: 50) { items { ${PaymentRequest.fields} } }
      }
    ''', variables: {'role': role, 'status': status});
    final items = (data['walletSvcMyPaymentRequests'] as Map<String, dynamic>?)?['items'] as List? ?? const [];
    return [for (final r in items) PaymentRequest.fromJson(r as Map<String, dynamic>)];
  }

  Future<PaymentRequest> requestPayment({
    required String walletId,
    required String payerUserId,
    required String amount,
    String? note,
  }) async {
    final data = await _client.mutation('''
      mutation Request(\$id: String!, \$payer: String!, \$amount: String!, \$note: String) {
        walletSvcMyRequestPayment(wallet_id: \$id, payer_user_id: \$payer, amount: \$amount, note: \$note) { ${PaymentRequest.fields} }
      }
    ''', variables: {'id': walletId, 'payer': payerUserId, 'amount': amount, 'note': note});
    return PaymentRequest.fromJson(data['walletSvcMyRequestPayment'] as Map<String, dynamic>);
  }

  Future<void> payRequest(String requestId, String walletId, String pin) async {
    await _client.mutation(r'''
      mutation PayRequest($id: String!, $wallet: String!, $pin: String!) {
        walletSvcMyPayPaymentRequest(id: $id, wallet_id: $wallet, pin: $pin) { id status }
      }
    ''', variables: {'id': requestId, 'wallet': walletId, 'pin': pin});
  }

  Future<void> declineRequest(String requestId) async {
    await _client.mutation(r'''
      mutation Decline($id: String!) { walletSvcMyDeclinePaymentRequest(id: $id) { id status } }
    ''', variables: {'id': requestId});
  }

  Future<PointAccount?> points() async {
    final data = await _client.query('query { pointMyBalance { balance total_earned } }');
    final p = data['pointMyBalance'];
    return p == null ? null : PointAccount.fromJson(p as Map<String, dynamic>);
  }

  Future<QrCode> createReceiveQr({required String walletId, String? amount}) async {
    final content = Uri(scheme: 'keiki', host: 'pay', queryParameters: {
      'wallet_id': walletId,
      if (amount != null && amount.isNotEmpty) 'amount': amount,
    }).toString();
    final data = await _client.mutation(r'''
      mutation Qr($input: QrGenerateInput!) {
        createMyQrCode(input: $input) { id title content image_base64 image_url short_code }
      }
    ''', variables: {
      'input': {'title': 'Nhận tiền Keiki', 'type': 'text', 'content': content, 'size': 512},
    });
    return QrCode.fromJson(data['createMyQrCode'] as Map<String, dynamic>);
  }
}
