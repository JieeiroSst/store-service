import 'package:equatable/equatable.dart';

import '../../../core/network/graphql_client.dart';

class UserAddress extends Equatable {
  const UserAddress({
    required this.id,
    required this.fullName,
    required this.line1,
    this.line2 = '',
    this.state = '',
    this.phone = '',
    this.notes = '',
    this.isDefaultShipping = false,
  });

  final int id;
  final String fullName;
  final String line1;
  final String line2;
  final String state;
  final String phone;
  final String notes;
  final bool isDefaultShipping;

  String get summary => [line1, line2, state].where((s) => s.isNotEmpty).join(', ');

  static const fields =
      'id first_name last_name line1 line2 state phone_number notes is_default_for_shipping';

  factory UserAddress.fromJson(Map<String, dynamic> json) => UserAddress(
        id: (json['id'] as num).toInt(),
        fullName: '${json['first_name'] ?? ''} ${json['last_name'] ?? ''}'.trim(),
        line1: json['line1'] as String? ?? '',
        line2: json['line2'] as String? ?? '',
        state: json['state'] as String? ?? '',
        phone: json['phone_number'] as String? ?? '',
        notes: json['notes'] as String? ?? '',
        isDefaultShipping: json['is_default_for_shipping'] as bool? ?? false,
      );

  @override
  List<Object?> get props => [id, fullName, line1, line2, state, phone, isDefaultShipping];
}

class AddressRepository {
  AddressRepository(this._client);

  final GraphQLClient _client;

  Future<List<UserAddress>> myAddresses() async {
    final data = await _client.query('query { myUserAddresses { ${UserAddress.fields} } }');
    return [for (final a in data['myUserAddresses'] as List) UserAddress.fromJson(a as Map<String, dynamic>)];
  }

  Future<UserAddress> create({
    required String firstName,
    required String lastName,
    required String line1,
    String? line2,
    String? state,
    required String phone,
    bool isDefault = false,
  }) async {
    final data = await _client.mutation('''
      mutation CreateAddress(\$input: AddressUserAddressInput!) { createMyUserAddress(input: \$input) { ${UserAddress.fields} } }
    ''', variables: {
      'input': {
        'first_name': firstName,
        'last_name': lastName,
        'line1': line1,
        'line2': line2 ?? '',
        'state': state ?? '',
        'phone_number': phone,
        'country_id': 'VN',
        'is_default_for_shipping': isDefault,
      },
    });
    return UserAddress.fromJson(data['createMyUserAddress'] as Map<String, dynamic>);
  }

  Future<void> delete(int id) async {
    await _client.mutation(r'mutation Delete($id: Int!) { deleteMyUserAddress(id: $id) }', variables: {'id': id});
  }
}
