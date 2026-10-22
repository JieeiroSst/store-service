import 'package:equatable/equatable.dart';

import '../../../core/network/graphql_client.dart';

class NotifDevice extends Equatable {
  const NotifDevice({required this.id, this.deviceId = '', this.deviceType = '', this.isActive = false, this.lastUsedAt, this.tokenPreview = ''});

  final int id;
  final String deviceId;
  final String deviceType;
  final bool isActive;
  final String? lastUsedAt;
  final String tokenPreview;

  factory NotifDevice.fromJson(Map<String, dynamic> json) => NotifDevice(
        id: (json['id'] as num?)?.toInt() ?? 0,
        deviceId: json['device_id'] as String? ?? '',
        deviceType: json['device_type'] as String? ?? '',
        isActive: json['is_active'] as bool? ?? false,
        lastUsedAt: json['last_used_at'] as String?,
        tokenPreview: json['token_preview'] as String? ?? '',
      );

  @override
  List<Object?> get props => [id, deviceId, isActive];
}

class NotifContact extends Equatable {
  const NotifContact({this.email = '', this.phone = ''});

  final String email;
  final String phone;

  @override
  List<Object?> get props => [email, phone];
}

class NotificationRepository {
  NotificationRepository(this._client);

  final GraphQLClient _client;

  Future<List<NotifDevice>> myDevices() async {
    final data = await _client.query('query { myDevices { id device_id device_type is_active last_used_at token_preview } }');
    return [for (final d in (data['myDevices'] as List? ?? const [])) NotifDevice.fromJson(d as Map<String, dynamic>)];
  }

  Future<NotifContact> myContact() async {
    final data = await _client.query('query { myContact { email phone } }');
    final c = data['myContact'] as Map<String, dynamic>?;
    return NotifContact(email: c?['email'] as String? ?? '', phone: c?['phone'] as String? ?? '');
  }

  Future<void> updateContact({String? email, String? phone}) async {
    await _client.mutation(r'''
      mutation Contact($email: String, $phone: String) { updateMyContact(email: $email, phone: $phone) { email phone } }
    ''', variables: {'email': email, 'phone': phone});
  }

  Future<void> registerDevice({required String pushToken, required String deviceType}) async {
    await _client.mutation(r'''
      mutation Register($input: NotifDeviceInput!) { registerMyDevice(input: $input) { id } }
    ''', variables: {
      'input': {'device_token': pushToken, 'device_id': await _client.storage.deviceId(), 'device_type': deviceType},
    });
  }

  Future<void> unregisterThisDevice() async {
    await _client.mutation(r'''
      mutation Unregister($input: NotifUnregisterDeviceInput!) { unregisterMyDevice(input: $input) }
    ''', variables: {
      'input': {'device_id': await _client.storage.deviceId()},
    });
  }
}
