// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

import '../gateway_executor.dart';
import '../models.g.dart';

class SolanaApi {
  const SolanaApi(this._executor);

  final GatewayExecutor _executor;

  Future<SolAccountInfo?> solAccount({required String address, String? selectionSet}) async {
    final data = await _executor.query('query solAccount(\$address: String!) { solAccount(address: \$address) ${selectionSet ?? '{ Address Balance Owner Executable RentEpoch }'} }', variables: {
      'address': address,
    }
    );
    return data['solAccount'] == null ? null : SolAccountInfo.fromJson((data['solAccount'] as Map).cast<String, dynamic>());
  }

  Future<SolBalance?> solBalance({required String address, String? selectionSet}) async {
    final data = await _executor.query('query solBalance(\$address: String!) { solBalance(address: \$address) ${selectionSet ?? '{ balance }'} }', variables: {
      'address': address,
    }
    );
    return data['solBalance'] == null ? null : SolBalance.fromJson((data['solBalance'] as Map).cast<String, dynamic>());
  }

  Future<SolTransaction?> solTransaction({required String signature, String? selectionSet}) async {
    final data = await _executor.query('query solTransaction(\$signature: String!) { solTransaction(signature: \$signature) ${selectionSet ?? '{ Signature BlockTime Slot Fee Status Instructions { ProgramID Accounts Data } }'} }', variables: {
      'signature': signature,
    }
    );
    return data['solTransaction'] == null ? null : SolTransaction.fromJson((data['solTransaction'] as Map).cast<String, dynamic>());
  }

  Future<SolProgram?> solProgram({required String id, String? selectionSet}) async {
    final data = await _executor.query('query solProgram(\$id: String!) { solProgram(id: \$id) ${selectionSet ?? '{ ID Name Executable Owner }'} }', variables: {
      'id': id,
    }
    );
    return data['solProgram'] == null ? null : SolProgram.fromJson((data['solProgram'] as Map).cast<String, dynamic>());
  }

  Future<SolCircleWalletSet?> solCircleWalletSet({required String id, String? selectionSet}) async {
    final data = await _executor.query('query solCircleWalletSet(\$id: String!) { solCircleWalletSet(id: \$id) ${selectionSet ?? '{ id name custodyType createDate updateDate }'} }', variables: {
      'id': id,
    }
    );
    return data['solCircleWalletSet'] == null ? null : SolCircleWalletSet.fromJson((data['solCircleWalletSet'] as Map).cast<String, dynamic>());
  }

  Future<List<SolCircleWalletSet>?> solCircleWalletSets({String? selectionSet}) async {
    final data = await _executor.query('query solCircleWalletSets { solCircleWalletSets ${selectionSet ?? '{ id name custodyType createDate updateDate }'} }'
    );
    return (data['solCircleWalletSets'] as List?)?.map((e) => (e == null ? null : SolCircleWalletSet.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<SolCircleWallet?> solCircleWallet({required String id, String? selectionSet}) async {
    final data = await _executor.query('query solCircleWallet(\$id: String!) { solCircleWallet(id: \$id) ${selectionSet ?? '{ id walletSetId custodyType address blockchain state accountType createDate updateDate }'} }', variables: {
      'id': id,
    }
    );
    return data['solCircleWallet'] == null ? null : SolCircleWallet.fromJson((data['solCircleWallet'] as Map).cast<String, dynamic>());
  }

  Future<List<SolCircleWallet>?> solCircleWallets({String? walletSetId, String? blockchain, String? selectionSet}) async {
    final data = await _executor.query('query solCircleWallets(\$walletSetId: String, \$blockchain: String) { solCircleWallets(walletSetId: \$walletSetId, blockchain: \$blockchain) ${selectionSet ?? '{ id walletSetId custodyType address blockchain state accountType createDate updateDate }'} }', variables: {
      'walletSetId': ?walletSetId,
      'blockchain': ?blockchain,
    }
    );
    return (data['solCircleWallets'] as List?)?.map((e) => (e == null ? null : SolCircleWallet.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<List<SolCircleBalance>?> solCircleWalletBalance({required String id, String? selectionSet}) async {
    final data = await _executor.query('query solCircleWalletBalance(\$id: String!) { solCircleWalletBalance(id: \$id) ${selectionSet ?? '{ tokenId token { id blockchain name symbol decimals isNative tokenAddress standard createDate updateDate } amount updateDate }'} }', variables: {
      'id': id,
    }
    );
    return (data['solCircleWalletBalance'] as List?)?.map((e) => (e == null ? null : SolCircleBalance.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<List<SolNFTBalance>?> solCircleWalletNFTs({required String id, String? selectionSet}) async {
    final data = await _executor.query('query solCircleWalletNFTs(\$id: String!) { solCircleWalletNFTs(id: \$id) ${selectionSet ?? '{ tokenId name standard blockchain tokenCount updateDate }'} }', variables: {
      'id': id,
    }
    );
    return (data['solCircleWalletNFTs'] as List?)?.map((e) => (e == null ? null : SolNFTBalance.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<SolCircleTransaction?> solCircleTransaction({required String id, String? selectionSet}) async {
    final data = await _executor.query('query solCircleTransaction(\$id: String!) { solCircleTransaction(id: \$id) ${selectionSet ?? '{ id blockchain tokenId walletId sourceAddress destinationAddress transactionType custodyType state amounts nftTokenIds txHash blockHash blockHeight networkFee firstConfirmDate operation abiParameters { name type value } createDate updateDate userId errorReason }'} }', variables: {
      'id': id,
    }
    );
    return data['solCircleTransaction'] == null ? null : SolCircleTransaction.fromJson((data['solCircleTransaction'] as Map).cast<String, dynamic>());
  }

  Future<SolFeeEstimate?> solCircleEstimateFee({required String walletId, required String destinationAddress, required String tokenId, required String amount, String? selectionSet}) async {
    final data = await _executor.query('query solCircleEstimateFee(\$walletId: String!, \$destinationAddress: String!, \$tokenId: String!, \$amount: String!) { solCircleEstimateFee(walletId: \$walletId, destinationAddress: \$destinationAddress, tokenId: \$tokenId, amount: \$amount) ${selectionSet ?? '{ low { maxFee gasLimit baseFee priorityFee } medium { maxFee gasLimit baseFee priorityFee } high { maxFee gasLimit baseFee priorityFee } }'} }', variables: {
      'walletId': walletId,
      'destinationAddress': destinationAddress,
      'tokenId': tokenId,
      'amount': amount,
    }
    );
    return data['solCircleEstimateFee'] == null ? null : SolFeeEstimate.fromJson((data['solCircleEstimateFee'] as Map).cast<String, dynamic>());
  }

  Future<SolToken?> solCircleToken({required String id, String? selectionSet}) async {
    final data = await _executor.query('query solCircleToken(\$id: String!) { solCircleToken(id: \$id) ${selectionSet ?? '{ id blockchain name symbol decimals isNative tokenAddress standard createDate updateDate }'} }', variables: {
      'id': id,
    }
    );
    return data['solCircleToken'] == null ? null : SolToken.fromJson((data['solCircleToken'] as Map).cast<String, dynamic>());
  }

  Future<List<SolToken>?> solCircleTokens({String? blockchain, String? selectionSet}) async {
    final data = await _executor.query('query solCircleTokens(\$blockchain: String) { solCircleTokens(blockchain: \$blockchain) ${selectionSet ?? '{ id blockchain name symbol decimals isNative tokenAddress standard createDate updateDate }'} }', variables: {
      'blockchain': ?blockchain,
    }
    );
    return (data['solCircleTokens'] as List?)?.map((e) => (e == null ? null : SolToken.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<SolBridgeTransfer?> solBridgeTransfer({required String id, String? selectionSet}) async {
    final data = await _executor.query('query solBridgeTransfer(\$id: String!) { solBridgeTransfer(id: \$id) ${selectionSet ?? '{ id solanaAddress circleWalletId direction amount tokenId status solanaTxSignature circleTxId createDate completeDate errorMessage }'} }', variables: {
      'id': id,
    }
    );
    return data['solBridgeTransfer'] == null ? null : SolBridgeTransfer.fromJson((data['solBridgeTransfer'] as Map).cast<String, dynamic>());
  }

  Future<Object?> solEstimateTransferFee({required String from, required String to, required int amount}) async {
    final data = await _executor.mutation('mutation solEstimateTransferFee(\$from: String!, \$to: String!, \$amount: Int64!) { solEstimateTransferFee(from: \$from, to: \$to, amount: \$amount) }', variables: {
      'from': from,
      'to': to,
      'amount': amount,
    }
    );
    return data['solEstimateTransferFee'];
  }

  Future<Object?> solFindPda({required String programId, required String userPubkey, required String identifier}) async {
    final data = await _executor.mutation('mutation solFindPda(\$program_id: String!, \$user_pubkey: String!, \$identifier: String!) { solFindPda(program_id: \$program_id, user_pubkey: \$user_pubkey, identifier: \$identifier) }', variables: {
      'program_id': programId,
      'user_pubkey': userPubkey,
      'identifier': identifier,
    }
    );
    return data['solFindPda'];
  }

  Future<SolCircleWalletSet?> circleCreateWalletSet({required String name_, String? custodyType, String? selectionSet}) async {
    final data = await _executor.mutation('mutation circleCreateWalletSet(\$name: String!, \$custodyType: String) { circleCreateWalletSet(name: \$name, custodyType: \$custodyType) ${selectionSet ?? '{ id name custodyType createDate updateDate }'} }', variables: {
      'name': name_,
      'custodyType': ?custodyType,
    }
    );
    return data['circleCreateWalletSet'] == null ? null : SolCircleWalletSet.fromJson((data['circleCreateWalletSet'] as Map).cast<String, dynamic>());
  }

  Future<Object?> circleCreateWallet({required String walletSetId, required String blockchain, String? accountType}) async {
    final data = await _executor.mutation('mutation circleCreateWallet(\$walletSetId: String!, \$blockchain: String!, \$accountType: String) { circleCreateWallet(walletSetId: \$walletSetId, blockchain: \$blockchain, accountType: \$accountType) }', variables: {
      'walletSetId': walletSetId,
      'blockchain': blockchain,
      'accountType': ?accountType,
    }
    );
    return data['circleCreateWallet'];
  }

  Future<Object?> circleFreezeWallet({required String id}) async {
    final data = await _executor.mutation('mutation circleFreezeWallet(\$id: String!) { circleFreezeWallet(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['circleFreezeWallet'];
  }

  Future<Object?> circleUnfreezeWallet({required String id}) async {
    final data = await _executor.mutation('mutation circleUnfreezeWallet(\$id: String!) { circleUnfreezeWallet(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['circleUnfreezeWallet'];
  }

  Future<SolCircleTransaction?> circleCreateTransfer({required String walletId, required String destinationAddress, required String tokenId, required String amount, SolFeeInput? fee, String? idempotencyKey, String? selectionSet}) async {
    final data = await _executor.mutation('mutation circleCreateTransfer(\$wallet_id: String!, \$destination_address: String!, \$token_id: String!, \$amount: String!, \$fee: SolFeeInput, \$idempotency_key: String) { circleCreateTransfer(wallet_id: \$wallet_id, destination_address: \$destination_address, token_id: \$token_id, amount: \$amount, fee: \$fee, idempotency_key: \$idempotency_key) ${selectionSet ?? '{ id blockchain tokenId walletId sourceAddress destinationAddress transactionType custodyType state amounts nftTokenIds txHash blockHash blockHeight networkFee firstConfirmDate operation abiParameters { name type value } createDate updateDate userId errorReason }'} }', variables: {
      'wallet_id': walletId,
      'destination_address': destinationAddress,
      'token_id': tokenId,
      'amount': amount,
      if (fee != null) 'fee': fee.toJson(),
      'idempotency_key': ?idempotencyKey,
    }
    );
    return data['circleCreateTransfer'] == null ? null : SolCircleTransaction.fromJson((data['circleCreateTransfer'] as Map).cast<String, dynamic>());
  }

  Future<SolCircleTransaction?> circleCreateNftTransfer({required Map<String, dynamic> input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation circleCreateNftTransfer(\$input: Map!) { circleCreateNftTransfer(input: \$input) ${selectionSet ?? '{ id blockchain tokenId walletId sourceAddress destinationAddress transactionType custodyType state amounts nftTokenIds txHash blockHash blockHeight networkFee firstConfirmDate operation abiParameters { name type value } createDate updateDate userId errorReason }'} }', variables: {
      'input': input,
    }
    );
    return data['circleCreateNftTransfer'] == null ? null : SolCircleTransaction.fromJson((data['circleCreateNftTransfer'] as Map).cast<String, dynamic>());
  }

  Future<SolCircleTransaction?> circleExecuteContract({required Map<String, dynamic> input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation circleExecuteContract(\$input: Map!) { circleExecuteContract(input: \$input) ${selectionSet ?? '{ id blockchain tokenId walletId sourceAddress destinationAddress transactionType custodyType state amounts nftTokenIds txHash blockHash blockHeight networkFee firstConfirmDate operation abiParameters { name type value } createDate updateDate userId errorReason }'} }', variables: {
      'input': input,
    }
    );
    return data['circleExecuteContract'] == null ? null : SolCircleTransaction.fromJson((data['circleExecuteContract'] as Map).cast<String, dynamic>());
  }

  Future<SolToken?> circleImportToken({required Map<String, dynamic> input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation circleImportToken(\$input: Map!) { circleImportToken(input: \$input) ${selectionSet ?? '{ id blockchain name symbol decimals isNative tokenAddress standard createDate updateDate }'} }', variables: {
      'input': input,
    }
    );
    return data['circleImportToken'] == null ? null : SolToken.fromJson((data['circleImportToken'] as Map).cast<String, dynamic>());
  }

  Future<SolBridgeTransfer?> bridgeSolanaToCircle({required String fromSolanaAddress, required String toCircleWalletId, required int amount, required String privateKey, String? selectionSet}) async {
    final data = await _executor.mutation('mutation bridgeSolanaToCircle(\$fromSolanaAddress: String!, \$toCircleWalletId: String!, \$amount: Int64!, \$privateKey: String!) { bridgeSolanaToCircle(fromSolanaAddress: \$fromSolanaAddress, toCircleWalletId: \$toCircleWalletId, amount: \$amount, privateKey: \$privateKey) ${selectionSet ?? '{ id solanaAddress circleWalletId direction amount tokenId status solanaTxSignature circleTxId createDate completeDate errorMessage }'} }', variables: {
      'fromSolanaAddress': fromSolanaAddress,
      'toCircleWalletId': toCircleWalletId,
      'amount': amount,
      'privateKey': privateKey,
    }
    );
    return data['bridgeSolanaToCircle'] == null ? null : SolBridgeTransfer.fromJson((data['bridgeSolanaToCircle'] as Map).cast<String, dynamic>());
  }

  Future<SolBridgeTransfer?> bridgeCircleToSolana({required String fromCircleWalletId, required String toSolanaAddress, required String amount, String? selectionSet}) async {
    final data = await _executor.mutation('mutation bridgeCircleToSolana(\$fromCircleWalletId: String!, \$toSolanaAddress: String!, \$amount: String!) { bridgeCircleToSolana(fromCircleWalletId: \$fromCircleWalletId, toSolanaAddress: \$toSolanaAddress, amount: \$amount) ${selectionSet ?? '{ id solanaAddress circleWalletId direction amount tokenId status solanaTxSignature circleTxId createDate completeDate errorMessage }'} }', variables: {
      'fromCircleWalletId': fromCircleWalletId,
      'toSolanaAddress': toSolanaAddress,
      'amount': amount,
    }
    );
    return data['bridgeCircleToSolana'] == null ? null : SolBridgeTransfer.fromJson((data['bridgeCircleToSolana'] as Map).cast<String, dynamic>());
  }

}
