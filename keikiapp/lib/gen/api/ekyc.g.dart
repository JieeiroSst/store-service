// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

import '../gateway_executor.dart';
import '../models.g.dart';

class EkycApi {
  const EkycApi(this._executor);

  final GatewayExecutor _executor;

  Future<EkycStatus?> myEkycStatus({String? selectionSet}) async {
    final data = await _executor.query('query myEkycStatus { myEkycStatus ${selectionSet ?? '{ user_id identity { id user_id source document_number surname given_names nationality date_of_birth sex date_of_expiry mrz_line1 mrz_line2 mrz_line3 checksum_valid confidence nfc_verified created_at updated_at } face { id user_id quality_score created_at updated_at } verification { id user_id status match_score verified_at created_at updated_at } }'} }'
    );
    return data['myEkycStatus'] == null ? null : EkycStatus.fromJson((data['myEkycStatus'] as Map).cast<String, dynamic>());
  }

  Future<EkycCitizenIdentity?> ekycSubmitCitizenCard({required GqlUpload front, required GqlUpload back, String? selectionSet}) async {
    final data = await _executor.mutation('mutation ekycSubmitCitizenCard(\$front: Upload!, \$back: Upload!) { ekycSubmitCitizenCard(front: \$front, back: \$back) ${selectionSet ?? '{ id user_id source document_number surname given_names nationality date_of_birth sex date_of_expiry mrz_line1 mrz_line2 mrz_line3 checksum_valid confidence nfc_verified created_at updated_at }'} }', variables: {
      'front': front,
      'back': back,
    }
    );
    return data['ekycSubmitCitizenCard'] == null ? null : EkycCitizenIdentity.fromJson((data['ekycSubmitCitizenCard'] as Map).cast<String, dynamic>());
  }

  Future<EkycCitizenIdentity?> ekycSubmitNfcChip({GqlUpload? efCom, GqlUpload? dg1, GqlUpload? dg2, GqlUpload? efSod, String? selectionSet}) async {
    final data = await _executor.mutation('mutation ekycSubmitNfcChip(\$ef_com: Upload, \$dg1: Upload, \$dg2: Upload, \$ef_sod: Upload) { ekycSubmitNfcChip(ef_com: \$ef_com, dg1: \$dg1, dg2: \$dg2, ef_sod: \$ef_sod) ${selectionSet ?? '{ id user_id source document_number surname given_names nationality date_of_birth sex date_of_expiry mrz_line1 mrz_line2 mrz_line3 checksum_valid confidence nfc_verified created_at updated_at }'} }', variables: {
      'ef_com': ?efCom,
      'dg1': ?dg1,
      'dg2': ?dg2,
      'ef_sod': ?efSod,
    }
    );
    return data['ekycSubmitNfcChip'] == null ? null : EkycCitizenIdentity.fromJson((data['ekycSubmitNfcChip'] as Map).cast<String, dynamic>());
  }

  Future<EkycFaceBiometric?> ekycSubmitFaceScan({required List<GqlUpload> frames, String? selectionSet}) async {
    final data = await _executor.mutation('mutation ekycSubmitFaceScan(\$frames: [Upload!]!) { ekycSubmitFaceScan(frames: \$frames) ${selectionSet ?? '{ id user_id quality_score created_at updated_at }'} }', variables: {
      'frames': frames,
    }
    );
    return data['ekycSubmitFaceScan'] == null ? null : EkycFaceBiometric.fromJson((data['ekycSubmitFaceScan'] as Map).cast<String, dynamic>());
  }

  Future<EkycVerification?> ekycVerify({String? selectionSet}) async {
    final data = await _executor.mutation('mutation ekycVerify { ekycVerify ${selectionSet ?? '{ id user_id status match_score verified_at created_at updated_at }'} }'
    );
    return data['ekycVerify'] == null ? null : EkycVerification.fromJson((data['ekycVerify'] as Map).cast<String, dynamic>());
  }

}
