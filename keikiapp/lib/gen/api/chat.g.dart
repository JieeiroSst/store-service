// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

import '../gateway_executor.dart';
import '../models.g.dart';

class ChatApi {
  const ChatApi(this._executor);

  final GatewayExecutor _executor;

  Future<ChatMessage?> chatMessage({required int id, String? selectionSet}) async {
    final data = await _executor.query('query chatMessage(\$id: Int!) { chatMessage(id: \$id) ${selectionSet ?? '{ _id sender_id message_type created_at deleted_at }'} }', variables: {
      'id': id,
    }
    );
    return data['chatMessage'] == null ? null : ChatMessage.fromJson((data['chatMessage'] as Map).cast<String, dynamic>());
  }

  Future<List<ChatReport>?> myChatReports({String? selectionSet}) async {
    final data = await _executor.query('query myChatReports { myChatReports ${selectionSet ?? '{ _id users_id report_type notes status created_at }'} }'
    );
    return (data['myChatReports'] as List?)?.map((e) => (e == null ? null : ChatReport.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<List<ChatReport>?> chatUserReports({required int userID, String? selectionSet}) async {
    final data = await _executor.query('query chatUserReports(\$userID: Int!) { chatUserReports(userID: \$userID) ${selectionSet ?? '{ _id users_id report_type notes status created_at }'} }', variables: {
      'userID': userID,
    }
    );
    return (data['chatUserReports'] as List?)?.map((e) => (e == null ? null : ChatReport.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<ChatCreatedReport?> createChatReport({required ChatReportInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation createChatReport(\$input: ChatReportInput!) { createChatReport(input: \$input) ${selectionSet ?? '{ ID UserId ReportType Notes Status CreatedAt }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['createChatReport'] == null ? null : ChatCreatedReport.fromJson((data['createChatReport'] as Map).cast<String, dynamic>());
  }

  Future<bool?> deleteMyChatMessage({required int messageID}) async {
    final data = await _executor.mutation('mutation deleteMyChatMessage(\$messageID: Int!) { deleteMyChatMessage(messageID: \$messageID) }', variables: {
      'messageID': messageID,
    }
    );
    return data['deleteMyChatMessage'] as bool?;
  }

}
