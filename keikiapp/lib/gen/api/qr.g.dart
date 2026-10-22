// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

import '../gateway_executor.dart';
import '../models.g.dart';

class QrApi {
  const QrApi(this._executor);

  final GatewayExecutor _executor;

  Future<QrQRListResponse?> qrCodes({int? page, int? limit, String? status, String? type_, String? search, String? createdBy, String? selectionSet}) async {
    final data = await _executor.query('query qrCodes(\$page: Int, \$limit: Int, \$status: String, \$type: String, \$search: String, \$created_by: String) { qrCodes(page: \$page, limit: \$limit, status: \$status, type: \$type, search: \$search, created_by: \$created_by) ${selectionSet ?? '{ data { id short_code title type content redirect_url image_base64 image_url status scan_count size fore_color back_color expires_at created_at updated_at created_by } total page limit total_pages }'} }', variables: {
      'page': ?page,
      'limit': ?limit,
      'status': ?status,
      'type': ?type_,
      'search': ?search,
      'created_by': ?createdBy,
    }
    );
    return data['qrCodes'] == null ? null : QrQRListResponse.fromJson((data['qrCodes'] as Map).cast<String, dynamic>());
  }

  Future<QrQRCode?> qrCode({required String id, String? selectionSet}) async {
    final data = await _executor.query('query qrCode(\$id: String!) { qrCode(id: \$id) ${selectionSet ?? '{ id short_code title type content redirect_url image_base64 image_url status scan_count size fore_color back_color expires_at created_at updated_at created_by }'} }', variables: {
      'id': id,
    }
    );
    return data['qrCode'] == null ? null : QrQRCode.fromJson((data['qrCode'] as Map).cast<String, dynamic>());
  }

  Future<QrScanHistoryResponse?> qrScanHistory({required String id, int? page, int? limit, String? selectionSet}) async {
    final data = await _executor.query('query qrScanHistory(\$id: String!, \$page: Int, \$limit: Int) { qrScanHistory(id: \$id, page: \$page, limit: \$limit) ${selectionSet ?? '{ data { id qr_code_id short_code ip_address user_agent referer country city device_type os browser scanned_at } total page limit total_pages }'} }', variables: {
      'id': id,
      'page': ?page,
      'limit': ?limit,
    }
    );
    return data['qrScanHistory'] == null ? null : QrScanHistoryResponse.fromJson((data['qrScanHistory'] as Map).cast<String, dynamic>());
  }

  Future<QrScanStats?> qrScanStats({required String id, String? selectionSet}) async {
    final data = await _executor.query('query qrScanStats(\$id: String!) { qrScanStats(id: \$id) ${selectionSet ?? '{ total_scans unique_ips device_breakdown os_breakdown daily_scans { date count } }'} }', variables: {
      'id': id,
    }
    );
    return data['qrScanStats'] == null ? null : QrScanStats.fromJson((data['qrScanStats'] as Map).cast<String, dynamic>());
  }

  Future<QrQRListResponse?> myQrCodes({int? page, int? limit, String? status, String? type_, String? search, String? selectionSet}) async {
    final data = await _executor.query('query myQrCodes(\$page: Int, \$limit: Int, \$status: String, \$type: String, \$search: String) { myQrCodes(page: \$page, limit: \$limit, status: \$status, type: \$type, search: \$search) ${selectionSet ?? '{ data { id short_code title type content redirect_url image_base64 image_url status scan_count size fore_color back_color expires_at created_at updated_at created_by } total page limit total_pages }'} }', variables: {
      'page': ?page,
      'limit': ?limit,
      'status': ?status,
      'type': ?type_,
      'search': ?search,
    }
    );
    return data['myQrCodes'] == null ? null : QrQRListResponse.fromJson((data['myQrCodes'] as Map).cast<String, dynamic>());
  }

  Future<QrQRCode?> myQrCode({required String id, String? selectionSet}) async {
    final data = await _executor.query('query myQrCode(\$id: String!) { myQrCode(id: \$id) ${selectionSet ?? '{ id short_code title type content redirect_url image_base64 image_url status scan_count size fore_color back_color expires_at created_at updated_at created_by }'} }', variables: {
      'id': id,
    }
    );
    return data['myQrCode'] == null ? null : QrQRCode.fromJson((data['myQrCode'] as Map).cast<String, dynamic>());
  }

  Future<QrScanStats?> myQrScanStats({required String id, String? selectionSet}) async {
    final data = await _executor.query('query myQrScanStats(\$id: String!) { myQrScanStats(id: \$id) ${selectionSet ?? '{ total_scans unique_ips device_breakdown os_breakdown daily_scans { date count } }'} }', variables: {
      'id': id,
    }
    );
    return data['myQrScanStats'] == null ? null : QrScanStats.fromJson((data['myQrScanStats'] as Map).cast<String, dynamic>());
  }

  Future<QrQRCode?> createMyQrCode({required QrGenerateInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation createMyQrCode(\$input: QrGenerateInput!) { createMyQrCode(input: \$input) ${selectionSet ?? '{ id short_code title type content redirect_url image_base64 image_url status scan_count size fore_color back_color expires_at created_at updated_at created_by }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['createMyQrCode'] == null ? null : QrQRCode.fromJson((data['createMyQrCode'] as Map).cast<String, dynamic>());
  }

  Future<QrQRCode?> updateMyQrCode({required String id, required QrUpdateInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation updateMyQrCode(\$id: String!, \$input: QrUpdateInput!) { updateMyQrCode(id: \$id, input: \$input) ${selectionSet ?? '{ id short_code title type content redirect_url image_base64 image_url status scan_count size fore_color back_color expires_at created_at updated_at created_by }'} }', variables: {
      'id': id,
      'input': input.toJson(),
    }
    );
    return data['updateMyQrCode'] == null ? null : QrQRCode.fromJson((data['updateMyQrCode'] as Map).cast<String, dynamic>());
  }

  Future<QrQRCode?> updateMyQrContent({required String id, String? content, String? redirectUrl, String? selectionSet}) async {
    final data = await _executor.mutation('mutation updateMyQrContent(\$id: String!, \$content: String, \$redirect_url: String) { updateMyQrContent(id: \$id, content: \$content, redirect_url: \$redirect_url) ${selectionSet ?? '{ id short_code title type content redirect_url image_base64 image_url status scan_count size fore_color back_color expires_at created_at updated_at created_by }'} }', variables: {
      'id': id,
      'content': ?content,
      'redirect_url': ?redirectUrl,
    }
    );
    return data['updateMyQrContent'] == null ? null : QrQRCode.fromJson((data['updateMyQrContent'] as Map).cast<String, dynamic>());
  }

  Future<bool?> deleteMyQrCode({required String id}) async {
    final data = await _executor.mutation('mutation deleteMyQrCode(\$id: String!) { deleteMyQrCode(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['deleteMyQrCode'] as bool?;
  }

  Future<QrQRCode?> regenerateMyQrCode({required String id, String? selectionSet}) async {
    final data = await _executor.mutation('mutation regenerateMyQrCode(\$id: String!) { regenerateMyQrCode(id: \$id) ${selectionSet ?? '{ id short_code title type content redirect_url image_base64 image_url status scan_count size fore_color back_color expires_at created_at updated_at created_by }'} }', variables: {
      'id': id,
    }
    );
    return data['regenerateMyQrCode'] == null ? null : QrQRCode.fromJson((data['regenerateMyQrCode'] as Map).cast<String, dynamic>());
  }

  Future<QrQRCode?> qrGenerate({required QrGenerateInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation qrGenerate(\$input: QrGenerateInput!) { qrGenerate(input: \$input) ${selectionSet ?? '{ id short_code title type content redirect_url image_base64 image_url status scan_count size fore_color back_color expires_at created_at updated_at created_by }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['qrGenerate'] == null ? null : QrQRCode.fromJson((data['qrGenerate'] as Map).cast<String, dynamic>());
  }

  Future<QrQRCode?> qrUpdate({required String id, required QrUpdateInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation qrUpdate(\$id: String!, \$input: QrUpdateInput!) { qrUpdate(id: \$id, input: \$input) ${selectionSet ?? '{ id short_code title type content redirect_url image_base64 image_url status scan_count size fore_color back_color expires_at created_at updated_at created_by }'} }', variables: {
      'id': id,
      'input': input.toJson(),
    }
    );
    return data['qrUpdate'] == null ? null : QrQRCode.fromJson((data['qrUpdate'] as Map).cast<String, dynamic>());
  }

  Future<QrQRCode?> qrUpdateContent({required String id, String? content, String? redirectUrl, String? selectionSet}) async {
    final data = await _executor.mutation('mutation qrUpdateContent(\$id: String!, \$content: String, \$redirect_url: String) { qrUpdateContent(id: \$id, content: \$content, redirect_url: \$redirect_url) ${selectionSet ?? '{ id short_code title type content redirect_url image_base64 image_url status scan_count size fore_color back_color expires_at created_at updated_at created_by }'} }', variables: {
      'id': id,
      'content': ?content,
      'redirect_url': ?redirectUrl,
    }
    );
    return data['qrUpdateContent'] == null ? null : QrQRCode.fromJson((data['qrUpdateContent'] as Map).cast<String, dynamic>());
  }

  Future<bool?> qrDelete({required String id}) async {
    final data = await _executor.mutation('mutation qrDelete(\$id: String!) { qrDelete(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['qrDelete'] as bool?;
  }

  Future<QrQRCode?> qrRegenerate({required String id, String? selectionSet}) async {
    final data = await _executor.mutation('mutation qrRegenerate(\$id: String!) { qrRegenerate(id: \$id) ${selectionSet ?? '{ id short_code title type content redirect_url image_base64 image_url status scan_count size fore_color back_color expires_at created_at updated_at created_by }'} }', variables: {
      'id': id,
    }
    );
    return data['qrRegenerate'] == null ? null : QrQRCode.fromJson((data['qrRegenerate'] as Map).cast<String, dynamic>());
  }

}
