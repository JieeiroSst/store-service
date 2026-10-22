// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

import '../gateway_executor.dart';
import '../models.g.dart';

class FaceApi {
  const FaceApi(this._executor);

  final GatewayExecutor _executor;

  Future<FacePositions?> facePositions({String? selectionSet}) async {
    final data = await _executor.query('query facePositions { facePositions ${selectionSet ?? '{ faces { x y width height } colors { color x y width height } frame updatedAt }'} }'
    );
    return data['facePositions'] == null ? null : FacePositions.fromJson((data['facePositions'] as Map).cast<String, dynamic>());
  }

  Future<bool?> faceReportPositions({required FacePositionsInput input}) async {
    final data = await _executor.mutation('mutation faceReportPositions(\$input: FacePositionsInput!) { faceReportPositions(input: \$input) }', variables: {
      'input': input.toJson(),
    }
    );
    return data['faceReportPositions'] as bool?;
  }

  Future<String?> faceUploadFile({required GqlUpload avatar}) async {
    final data = await _executor.mutation('mutation faceUploadFile(\$avatar: Upload!) { faceUploadFile(avatar: \$avatar) }', variables: {
      'avatar': avatar,
    }
    );
    return data['faceUploadFile']?.toString();
  }

  Future<String?> faceUploadFiles({GqlUpload? avatar, List<GqlUpload>? boop}) async {
    final data = await _executor.mutation('mutation faceUploadFiles(\$avatar: Upload, \$boop: [Upload!]) { faceUploadFiles(avatar: \$avatar, boop: \$boop) }', variables: {
      'avatar': ?avatar,
      'boop': ?boop,
    }
    );
    return data['faceUploadFiles']?.toString();
  }

}
