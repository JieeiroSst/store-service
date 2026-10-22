// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

import '../gateway_executor.dart';
import '../models.g.dart';

class NotificationApi {
  const NotificationApi(this._executor);

  final GatewayExecutor _executor;

  Future<List<NotifNotification>?> notifications({String? selectionSet}) async {
    final data = await _executor.query('query notifications { notifications ${selectionSet ?? '{ id user_id recipient email phone data raw_data title message type status priority created_at sent_at retry_count last_error requested_by template_type template_data }'} }'
    );
    return (data['notifications'] as List?)?.map((e) => (e == null ? null : NotifNotification.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<NotifNotification?> notification({required int id, String? selectionSet}) async {
    final data = await _executor.query('query notification(\$id: Int!) { notification(id: \$id) ${selectionSet ?? '{ id user_id recipient email phone data raw_data title message type status priority created_at sent_at retry_count last_error requested_by template_type template_data }'} }', variables: {
      'id': id,
    }
    );
    return data['notification'] == null ? null : NotifNotification.fromJson((data['notification'] as Map).cast<String, dynamic>());
  }

  Future<List<NotifCampaign>?> notifCampaigns({int? limit, int? offset, String? selectionSet}) async {
    final data = await _executor.query('query notifCampaigns(\$limit: Int, \$offset: Int) { notifCampaigns(limit: \$limit, offset: \$offset) ${selectionSet ?? '{ id channel audience topic title message data template_type template_data status recipients batches_planned last_error requested_by created_at updated_at started_at completed_at }'} }', variables: {
      'limit': ?limit,
      'offset': ?offset,
    }
    );
    return (data['notifCampaigns'] as List?)?.map((e) => (e == null ? null : NotifCampaign.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<NotifCampaignView?> notifCampaign({required int id, String? selectionSet}) async {
    final data = await _executor.query('query notifCampaign(\$id: Int!) { notifCampaign(id: \$id) ${selectionSet ?? '{ campaign { id channel audience topic title message data template_type template_data status recipients batches_planned last_error requested_by created_at updated_at started_at completed_at } progress { batches_queued batches_sending batches_sent batches_failed batches_cancelled targets sent failed invalid_tokens } }'} }', variables: {
      'id': id,
    }
    );
    return data['notifCampaign'] == null ? null : NotifCampaignView.fromJson((data['notifCampaign'] as Map).cast<String, dynamic>());
  }

  Future<NotifAuditPage?> notifAuditDeliveries({String? email, String? phone, String? channel, String? status, String? sourceType, String? requestedBy, int? userId, int? sourceId, int? beforeId, int? limit, String? from, String? to, String? selectionSet}) async {
    final data = await _executor.query('query notifAuditDeliveries(\$email: String, \$phone: String, \$channel: String, \$status: String, \$source_type: String, \$requested_by: String, \$user_id: Int, \$source_id: Int, \$before_id: Int, \$limit: Int, \$from: String, \$to: String) { notifAuditDeliveries(email: \$email, phone: \$phone, channel: \$channel, status: \$status, source_type: \$source_type, requested_by: \$requested_by, user_id: \$user_id, source_id: \$source_id, before_id: \$before_id, limit: \$limit, from: \$from, to: \$to) ${selectionSet ?? '{ items { id content_id source_type source_id batch_id attempt channel user_id recipient device_id device_type token_hash status message_id error requested_by created_at } contents next_before_id }'} }', variables: {
      'email': ?email,
      'phone': ?phone,
      'channel': ?channel,
      'status': ?status,
      'source_type': ?sourceType,
      'requested_by': ?requestedBy,
      'user_id': ?userId,
      'source_id': ?sourceId,
      'before_id': ?beforeId,
      'limit': ?limit,
      'from': ?from,
      'to': ?to,
    }
    );
    return data['notifAuditDeliveries'] == null ? null : NotifAuditPage.fromJson((data['notifAuditDeliveries'] as Map).cast<String, dynamic>());
  }

  Future<NotifAuditContent?> notifAuditContent({required int id, String? selectionSet}) async {
    final data = await _executor.query('query notifAuditContent(\$id: Int!) { notifAuditContent(id: \$id) ${selectionSet ?? '{ id channel title body data template_type created_at }'} }', variables: {
      'id': id,
    }
    );
    return data['notifAuditContent'] == null ? null : NotifAuditContent.fromJson((data['notifAuditContent'] as Map).cast<String, dynamic>());
  }

  Future<NotifUserContact?> notifUserContact({required int id, String? selectionSet}) async {
    final data = await _executor.query('query notifUserContact(\$id: Int!) { notifUserContact(id: \$id) ${selectionSet ?? '{ user_id email phone updated_at }'} }', variables: {
      'id': id,
    }
    );
    return data['notifUserContact'] == null ? null : NotifUserContact.fromJson((data['notifUserContact'] as Map).cast<String, dynamic>());
  }

  Future<List<NotifDevice>?> notifDevices({required int userId, String? selectionSet}) async {
    final data = await _executor.query('query notifDevices(\$user_id: Int!) { notifDevices(user_id: \$user_id) ${selectionSet ?? '{ id user_id device_id device_type is_active last_used_at created_at token_preview }'} }', variables: {
      'user_id': userId,
    }
    );
    return (data['notifDevices'] as List?)?.map((e) => (e == null ? null : NotifDevice.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<NotifDevice?> notifDevice({required int id, String? selectionSet}) async {
    final data = await _executor.query('query notifDevice(\$id: Int!) { notifDevice(id: \$id) ${selectionSet ?? '{ id user_id device_id device_type is_active last_used_at created_at token_preview }'} }', variables: {
      'id': id,
    }
    );
    return data['notifDevice'] == null ? null : NotifDevice.fromJson((data['notifDevice'] as Map).cast<String, dynamic>());
  }

  Future<List<NotifDevice>?> myDevices({String? selectionSet}) async {
    final data = await _executor.query('query myDevices { myDevices ${selectionSet ?? '{ id user_id device_id device_type is_active last_used_at created_at token_preview }'} }'
    );
    return (data['myDevices'] as List?)?.map((e) => (e == null ? null : NotifDevice.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<NotifUserContact?> myContact({String? selectionSet}) async {
    final data = await _executor.query('query myContact { myContact ${selectionSet ?? '{ user_id email phone updated_at }'} }'
    );
    return data['myContact'] == null ? null : NotifUserContact.fromJson((data['myContact'] as Map).cast<String, dynamic>());
  }

  Future<NotifNotification?> notifCreateNotification({required NotifNotificationInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation notifCreateNotification(\$input: NotifNotificationInput!) { notifCreateNotification(input: \$input) ${selectionSet ?? '{ id user_id recipient email phone data raw_data title message type status priority created_at sent_at retry_count last_error requested_by template_type template_data }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['notifCreateNotification'] == null ? null : NotifNotification.fromJson((data['notifCreateNotification'] as Map).cast<String, dynamic>());
  }

  Future<NotifNotification?> notifUpdateNotification({required int id, required NotifNotificationInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation notifUpdateNotification(\$id: Int!, \$input: NotifNotificationInput!) { notifUpdateNotification(id: \$id, input: \$input) ${selectionSet ?? '{ id user_id recipient email phone data raw_data title message type status priority created_at sent_at retry_count last_error requested_by template_type template_data }'} }', variables: {
      'id': id,
      'input': input.toJson(),
    }
    );
    return data['notifUpdateNotification'] == null ? null : NotifNotification.fromJson((data['notifUpdateNotification'] as Map).cast<String, dynamic>());
  }

  Future<bool?> notifDeleteNotification({required int id}) async {
    final data = await _executor.mutation('mutation notifDeleteNotification(\$id: Int!) { notifDeleteNotification(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['notifDeleteNotification'] as bool?;
  }

  Future<Object?> notifSendEmail({required NotifEmailInput input}) async {
    final data = await _executor.mutation('mutation notifSendEmail(\$input: NotifEmailInput!) { notifSendEmail(input: \$input) }', variables: {
      'input': input.toJson(),
    }
    );
    return data['notifSendEmail'];
  }

  Future<Object?> notifSendSlack({required NotifSlackInput input}) async {
    final data = await _executor.mutation('mutation notifSendSlack(\$input: NotifSlackInput!) { notifSendSlack(input: \$input) }', variables: {
      'input': input.toJson(),
    }
    );
    return data['notifSendSlack'];
  }

  Future<Object?> notifSendPush({required NotifPushInput input}) async {
    final data = await _executor.mutation('mutation notifSendPush(\$input: NotifPushInput!) { notifSendPush(input: \$input) }', variables: {
      'input': input.toJson(),
    }
    );
    return data['notifSendPush'];
  }

  Future<NotifCampaign?> notifCreateCampaign({required NotifCampaignInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation notifCreateCampaign(\$input: NotifCampaignInput!) { notifCreateCampaign(input: \$input) ${selectionSet ?? '{ id channel audience topic title message data template_type template_data status recipients batches_planned last_error requested_by created_at updated_at started_at completed_at }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['notifCreateCampaign'] == null ? null : NotifCampaign.fromJson((data['notifCreateCampaign'] as Map).cast<String, dynamic>());
  }

  Future<NotifCampaignView?> notifCancelCampaign({required int id, String? selectionSet}) async {
    final data = await _executor.mutation('mutation notifCancelCampaign(\$id: Int!) { notifCancelCampaign(id: \$id) ${selectionSet ?? '{ campaign { id channel audience topic title message data template_type template_data status recipients batches_planned last_error requested_by created_at updated_at started_at completed_at } progress { batches_queued batches_sending batches_sent batches_failed batches_cancelled targets sent failed invalid_tokens } }'} }', variables: {
      'id': id,
    }
    );
    return data['notifCancelCampaign'] == null ? null : NotifCampaignView.fromJson((data['notifCancelCampaign'] as Map).cast<String, dynamic>());
  }

  Future<NotifUserContact?> notifUpsertContact({required int id, String? email, String? phone, String? selectionSet}) async {
    final data = await _executor.mutation('mutation notifUpsertContact(\$id: Int!, \$email: String, \$phone: String) { notifUpsertContact(id: \$id, email: \$email, phone: \$phone) ${selectionSet ?? '{ user_id email phone updated_at }'} }', variables: {
      'id': id,
      'email': ?email,
      'phone': ?phone,
    }
    );
    return data['notifUpsertContact'] == null ? null : NotifUserContact.fromJson((data['notifUpsertContact'] as Map).cast<String, dynamic>());
  }

  Future<NotifDevice?> notifUpdateDevice({required int id, required NotifDeviceUpdateInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation notifUpdateDevice(\$id: Int!, \$input: NotifDeviceUpdateInput!) { notifUpdateDevice(id: \$id, input: \$input) ${selectionSet ?? '{ id user_id device_id device_type is_active last_used_at created_at token_preview }'} }', variables: {
      'id': id,
      'input': input.toJson(),
    }
    );
    return data['notifUpdateDevice'] == null ? null : NotifDevice.fromJson((data['notifUpdateDevice'] as Map).cast<String, dynamic>());
  }

  Future<bool?> notifDeleteDevice({required int id}) async {
    final data = await _executor.mutation('mutation notifDeleteDevice(\$id: Int!) { notifDeleteDevice(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['notifDeleteDevice'] as bool?;
  }

  Future<NotifDevice?> registerMyDevice({required NotifDeviceInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation registerMyDevice(\$input: NotifDeviceInput!) { registerMyDevice(input: \$input) ${selectionSet ?? '{ id user_id device_id device_type is_active last_used_at created_at token_preview }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['registerMyDevice'] == null ? null : NotifDevice.fromJson((data['registerMyDevice'] as Map).cast<String, dynamic>());
  }

  Future<Object?> unregisterMyDevice({required NotifUnregisterDeviceInput input}) async {
    final data = await _executor.mutation('mutation unregisterMyDevice(\$input: NotifUnregisterDeviceInput!) { unregisterMyDevice(input: \$input) }', variables: {
      'input': input.toJson(),
    }
    );
    return data['unregisterMyDevice'];
  }

  Future<NotifUserContact?> updateMyContact({String? email, String? phone, String? selectionSet}) async {
    final data = await _executor.mutation('mutation updateMyContact(\$email: String, \$phone: String) { updateMyContact(email: \$email, phone: \$phone) ${selectionSet ?? '{ user_id email phone updated_at }'} }', variables: {
      'email': ?email,
      'phone': ?phone,
    }
    );
    return data['updateMyContact'] == null ? null : NotifUserContact.fromJson((data['updateMyContact'] as Map).cast<String, dynamic>());
  }

}
