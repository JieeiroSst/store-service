// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

import '../gateway_executor.dart';
import '../models.g.dart';

class AddressApi {
  const AddressApi(this._executor);

  final GatewayExecutor _executor;

  Future<List<AddressCountry>?> addressCountries({String? selectionSet}) async {
    final data = await _executor.public('query addressCountries { addressCountries ${selectionSet ?? '{ iso31661a2 prinatable_name name display_order is_shipping_country }'} }'
    );
    return (data['addressCountries'] as List?)?.map((e) => (e == null ? null : AddressCountry.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<AddressCountry?> addressCountry({required int id, String? selectionSet}) async {
    final data = await _executor.public('query addressCountry(\$id: Int!) { addressCountry(id: \$id) ${selectionSet ?? '{ iso31661a2 prinatable_name name display_order is_shipping_country }'} }', variables: {
      'id': id,
    }
    );
    return data['addressCountry'] == null ? null : AddressCountry.fromJson((data['addressCountry'] as Map).cast<String, dynamic>());
  }

  Future<List<AddressOrderBillingAddress>?> addressOrderBillingAddresses({String? selectionSet}) async {
    final data = await _executor.query('query addressOrderBillingAddresses { addressOrderBillingAddresses ${selectionSet ?? '{ id title first_name last_name line1 line2 line3 line4 state postcode search_text country_id address_country { iso31661a2 prinatable_name name display_order is_shipping_country } }'} }'
    );
    return (data['addressOrderBillingAddresses'] as List?)?.map((e) => (e == null ? null : AddressOrderBillingAddress.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<AddressOrderBillingAddress?> addressOrderBillingAddress({required int id, String? selectionSet}) async {
    final data = await _executor.query('query addressOrderBillingAddress(\$id: Int!) { addressOrderBillingAddress(id: \$id) ${selectionSet ?? '{ id title first_name last_name line1 line2 line3 line4 state postcode search_text country_id address_country { iso31661a2 prinatable_name name display_order is_shipping_country } }'} }', variables: {
      'id': id,
    }
    );
    return data['addressOrderBillingAddress'] == null ? null : AddressOrderBillingAddress.fromJson((data['addressOrderBillingAddress'] as Map).cast<String, dynamic>());
  }

  Future<List<AddressOrderShippingAddress>?> addressOrderShippingAddresses({String? selectionSet}) async {
    final data = await _executor.query('query addressOrderShippingAddresses { addressOrderShippingAddresses ${selectionSet ?? '{ id title first_name last_name line1 line2 line3 line4 state postcode search_text country_id phone_number notes address_country { iso31661a2 prinatable_name name display_order is_shipping_country } }'} }'
    );
    return (data['addressOrderShippingAddresses'] as List?)?.map((e) => (e == null ? null : AddressOrderShippingAddress.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<AddressOrderShippingAddress?> addressOrderShippingAddress({required int id, String? selectionSet}) async {
    final data = await _executor.query('query addressOrderShippingAddress(\$id: Int!) { addressOrderShippingAddress(id: \$id) ${selectionSet ?? '{ id title first_name last_name line1 line2 line3 line4 state postcode search_text country_id phone_number notes address_country { iso31661a2 prinatable_name name display_order is_shipping_country } }'} }', variables: {
      'id': id,
    }
    );
    return data['addressOrderShippingAddress'] == null ? null : AddressOrderShippingAddress.fromJson((data['addressOrderShippingAddress'] as Map).cast<String, dynamic>());
  }

  Future<List<AddressPartnerAddress>?> addressPartnerAddresses({String? selectionSet}) async {
    final data = await _executor.query('query addressPartnerAddresses { addressPartnerAddresses ${selectionSet ?? '{ id title first_name last_name line1 line2 line3 line4 state postcode search_text country_id partner_id address_country { iso31661a2 prinatable_name name display_order is_shipping_country } }'} }'
    );
    return (data['addressPartnerAddresses'] as List?)?.map((e) => (e == null ? null : AddressPartnerAddress.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<AddressPartnerAddress?> addressPartnerAddress({required int id, String? selectionSet}) async {
    final data = await _executor.query('query addressPartnerAddress(\$id: Int!) { addressPartnerAddress(id: \$id) ${selectionSet ?? '{ id title first_name last_name line1 line2 line3 line4 state postcode search_text country_id partner_id address_country { iso31661a2 prinatable_name name display_order is_shipping_country } }'} }', variables: {
      'id': id,
    }
    );
    return data['addressPartnerAddress'] == null ? null : AddressPartnerAddress.fromJson((data['addressPartnerAddress'] as Map).cast<String, dynamic>());
  }

  Future<List<AddressShippingOrderAndItemChangesCountry>?> addressShippingOrderAndItemChangesCountries({String? selectionSet}) async {
    final data = await _executor.query('query addressShippingOrderAndItemChangesCountries { addressShippingOrderAndItemChangesCountries ${selectionSet ?? '{ id orderanditemchanges_id country_id address_country { iso31661a2 prinatable_name name display_order is_shipping_country } }'} }'
    );
    return (data['addressShippingOrderAndItemChangesCountries'] as List?)?.map((e) => (e == null ? null : AddressShippingOrderAndItemChangesCountry.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<AddressShippingOrderAndItemChangesCountry?> addressShippingOrderAndItemChangesCountry({required int id, String? selectionSet}) async {
    final data = await _executor.query('query addressShippingOrderAndItemChangesCountry(\$id: Int!) { addressShippingOrderAndItemChangesCountry(id: \$id) ${selectionSet ?? '{ id orderanditemchanges_id country_id address_country { iso31661a2 prinatable_name name display_order is_shipping_country } }'} }', variables: {
      'id': id,
    }
    );
    return data['addressShippingOrderAndItemChangesCountry'] == null ? null : AddressShippingOrderAndItemChangesCountry.fromJson((data['addressShippingOrderAndItemChangesCountry'] as Map).cast<String, dynamic>());
  }

  Future<List<AddressShippingWeightBasedCountry>?> addressShippingWeightBasedCountries({String? selectionSet}) async {
    final data = await _executor.query('query addressShippingWeightBasedCountries { addressShippingWeightBasedCountries ${selectionSet ?? '{ id weightbased_id country_id address_country { iso31661a2 prinatable_name name display_order is_shipping_country } }'} }'
    );
    return (data['addressShippingWeightBasedCountries'] as List?)?.map((e) => (e == null ? null : AddressShippingWeightBasedCountry.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<AddressShippingWeightBasedCountry?> addressShippingWeightBasedCountry({required int id, String? selectionSet}) async {
    final data = await _executor.query('query addressShippingWeightBasedCountry(\$id: Int!) { addressShippingWeightBasedCountry(id: \$id) ${selectionSet ?? '{ id weightbased_id country_id address_country { iso31661a2 prinatable_name name display_order is_shipping_country } }'} }', variables: {
      'id': id,
    }
    );
    return data['addressShippingWeightBasedCountry'] == null ? null : AddressShippingWeightBasedCountry.fromJson((data['addressShippingWeightBasedCountry'] as Map).cast<String, dynamic>());
  }

  Future<List<AddressUserAddress>?> addressUserAddresses({String? selectionSet}) async {
    final data = await _executor.query('query addressUserAddresses { addressUserAddresses ${selectionSet ?? '{ id title first_name last_name line1 line2 line3 line4 state postcode search_text country_id phone_number notes is_default_for_shipping is_default_for_billing number_orders_as_shipping_address number_orders_as_billing_address hash date_created user_id address_country { iso31661a2 prinatable_name name display_order is_shipping_country } }'} }'
    );
    return (data['addressUserAddresses'] as List?)?.map((e) => (e == null ? null : AddressUserAddress.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<AddressUserAddress?> addressUserAddress({required int id, String? selectionSet}) async {
    final data = await _executor.query('query addressUserAddress(\$id: Int!) { addressUserAddress(id: \$id) ${selectionSet ?? '{ id title first_name last_name line1 line2 line3 line4 state postcode search_text country_id phone_number notes is_default_for_shipping is_default_for_billing number_orders_as_shipping_address number_orders_as_billing_address hash date_created user_id address_country { iso31661a2 prinatable_name name display_order is_shipping_country } }'} }', variables: {
      'id': id,
    }
    );
    return data['addressUserAddress'] == null ? null : AddressUserAddress.fromJson((data['addressUserAddress'] as Map).cast<String, dynamic>());
  }

  Future<List<AddressUserAddress>?> myUserAddresses({String? selectionSet}) async {
    final data = await _executor.query('query myUserAddresses { myUserAddresses ${selectionSet ?? '{ id title first_name last_name line1 line2 line3 line4 state postcode search_text country_id phone_number notes is_default_for_shipping is_default_for_billing number_orders_as_shipping_address number_orders_as_billing_address hash date_created user_id address_country { iso31661a2 prinatable_name name display_order is_shipping_country } }'} }'
    );
    return (data['myUserAddresses'] as List?)?.map((e) => (e == null ? null : AddressUserAddress.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<AddressCountry?> addressCreateCountry({required AddressCountryInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation addressCreateCountry(\$input: AddressCountryInput!) { addressCreateCountry(input: \$input) ${selectionSet ?? '{ iso31661a2 prinatable_name name display_order is_shipping_country }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['addressCreateCountry'] == null ? null : AddressCountry.fromJson((data['addressCreateCountry'] as Map).cast<String, dynamic>());
  }

  Future<AddressCountry?> addressUpdateCountry({required int id, required AddressCountryInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation addressUpdateCountry(\$id: Int!, \$input: AddressCountryInput!) { addressUpdateCountry(id: \$id, input: \$input) ${selectionSet ?? '{ iso31661a2 prinatable_name name display_order is_shipping_country }'} }', variables: {
      'id': id,
      'input': input.toJson(),
    }
    );
    return data['addressUpdateCountry'] == null ? null : AddressCountry.fromJson((data['addressUpdateCountry'] as Map).cast<String, dynamic>());
  }

  Future<bool?> addressDeleteCountry({required int id}) async {
    final data = await _executor.mutation('mutation addressDeleteCountry(\$id: Int!) { addressDeleteCountry(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['addressDeleteCountry'] as bool?;
  }

  Future<AddressOrderBillingAddress?> addressCreateOrderBillingAddress({required AddressOrderBillingAddressInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation addressCreateOrderBillingAddress(\$input: AddressOrderBillingAddressInput!) { addressCreateOrderBillingAddress(input: \$input) ${selectionSet ?? '{ id title first_name last_name line1 line2 line3 line4 state postcode search_text country_id address_country { iso31661a2 prinatable_name name display_order is_shipping_country } }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['addressCreateOrderBillingAddress'] == null ? null : AddressOrderBillingAddress.fromJson((data['addressCreateOrderBillingAddress'] as Map).cast<String, dynamic>());
  }

  Future<AddressOrderBillingAddress?> addressUpdateOrderBillingAddress({required int id, required AddressOrderBillingAddressInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation addressUpdateOrderBillingAddress(\$id: Int!, \$input: AddressOrderBillingAddressInput!) { addressUpdateOrderBillingAddress(id: \$id, input: \$input) ${selectionSet ?? '{ id title first_name last_name line1 line2 line3 line4 state postcode search_text country_id address_country { iso31661a2 prinatable_name name display_order is_shipping_country } }'} }', variables: {
      'id': id,
      'input': input.toJson(),
    }
    );
    return data['addressUpdateOrderBillingAddress'] == null ? null : AddressOrderBillingAddress.fromJson((data['addressUpdateOrderBillingAddress'] as Map).cast<String, dynamic>());
  }

  Future<bool?> addressDeleteOrderBillingAddress({required int id}) async {
    final data = await _executor.mutation('mutation addressDeleteOrderBillingAddress(\$id: Int!) { addressDeleteOrderBillingAddress(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['addressDeleteOrderBillingAddress'] as bool?;
  }

  Future<AddressOrderShippingAddress?> addressCreateOrderShippingAddress({required AddressOrderShippingAddressInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation addressCreateOrderShippingAddress(\$input: AddressOrderShippingAddressInput!) { addressCreateOrderShippingAddress(input: \$input) ${selectionSet ?? '{ id title first_name last_name line1 line2 line3 line4 state postcode search_text country_id phone_number notes address_country { iso31661a2 prinatable_name name display_order is_shipping_country } }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['addressCreateOrderShippingAddress'] == null ? null : AddressOrderShippingAddress.fromJson((data['addressCreateOrderShippingAddress'] as Map).cast<String, dynamic>());
  }

  Future<AddressOrderShippingAddress?> addressUpdateOrderShippingAddress({required int id, required AddressOrderShippingAddressInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation addressUpdateOrderShippingAddress(\$id: Int!, \$input: AddressOrderShippingAddressInput!) { addressUpdateOrderShippingAddress(id: \$id, input: \$input) ${selectionSet ?? '{ id title first_name last_name line1 line2 line3 line4 state postcode search_text country_id phone_number notes address_country { iso31661a2 prinatable_name name display_order is_shipping_country } }'} }', variables: {
      'id': id,
      'input': input.toJson(),
    }
    );
    return data['addressUpdateOrderShippingAddress'] == null ? null : AddressOrderShippingAddress.fromJson((data['addressUpdateOrderShippingAddress'] as Map).cast<String, dynamic>());
  }

  Future<bool?> addressDeleteOrderShippingAddress({required int id}) async {
    final data = await _executor.mutation('mutation addressDeleteOrderShippingAddress(\$id: Int!) { addressDeleteOrderShippingAddress(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['addressDeleteOrderShippingAddress'] as bool?;
  }

  Future<AddressPartnerAddress?> addressCreatePartnerAddress({required AddressPartnerAddressInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation addressCreatePartnerAddress(\$input: AddressPartnerAddressInput!) { addressCreatePartnerAddress(input: \$input) ${selectionSet ?? '{ id title first_name last_name line1 line2 line3 line4 state postcode search_text country_id partner_id address_country { iso31661a2 prinatable_name name display_order is_shipping_country } }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['addressCreatePartnerAddress'] == null ? null : AddressPartnerAddress.fromJson((data['addressCreatePartnerAddress'] as Map).cast<String, dynamic>());
  }

  Future<AddressPartnerAddress?> addressUpdatePartnerAddress({required int id, required AddressPartnerAddressInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation addressUpdatePartnerAddress(\$id: Int!, \$input: AddressPartnerAddressInput!) { addressUpdatePartnerAddress(id: \$id, input: \$input) ${selectionSet ?? '{ id title first_name last_name line1 line2 line3 line4 state postcode search_text country_id partner_id address_country { iso31661a2 prinatable_name name display_order is_shipping_country } }'} }', variables: {
      'id': id,
      'input': input.toJson(),
    }
    );
    return data['addressUpdatePartnerAddress'] == null ? null : AddressPartnerAddress.fromJson((data['addressUpdatePartnerAddress'] as Map).cast<String, dynamic>());
  }

  Future<bool?> addressDeletePartnerAddress({required int id}) async {
    final data = await _executor.mutation('mutation addressDeletePartnerAddress(\$id: Int!) { addressDeletePartnerAddress(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['addressDeletePartnerAddress'] as bool?;
  }

  Future<AddressShippingOrderAndItemChangesCountry?> addressCreateShippingOrderAndItemChangesCountry({required AddressShippingOrderAndItemChangesCountryInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation addressCreateShippingOrderAndItemChangesCountry(\$input: AddressShippingOrderAndItemChangesCountryInput!) { addressCreateShippingOrderAndItemChangesCountry(input: \$input) ${selectionSet ?? '{ id orderanditemchanges_id country_id address_country { iso31661a2 prinatable_name name display_order is_shipping_country } }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['addressCreateShippingOrderAndItemChangesCountry'] == null ? null : AddressShippingOrderAndItemChangesCountry.fromJson((data['addressCreateShippingOrderAndItemChangesCountry'] as Map).cast<String, dynamic>());
  }

  Future<AddressShippingOrderAndItemChangesCountry?> addressUpdateShippingOrderAndItemChangesCountry({required int id, required AddressShippingOrderAndItemChangesCountryInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation addressUpdateShippingOrderAndItemChangesCountry(\$id: Int!, \$input: AddressShippingOrderAndItemChangesCountryInput!) { addressUpdateShippingOrderAndItemChangesCountry(id: \$id, input: \$input) ${selectionSet ?? '{ id orderanditemchanges_id country_id address_country { iso31661a2 prinatable_name name display_order is_shipping_country } }'} }', variables: {
      'id': id,
      'input': input.toJson(),
    }
    );
    return data['addressUpdateShippingOrderAndItemChangesCountry'] == null ? null : AddressShippingOrderAndItemChangesCountry.fromJson((data['addressUpdateShippingOrderAndItemChangesCountry'] as Map).cast<String, dynamic>());
  }

  Future<bool?> addressDeleteShippingOrderAndItemChangesCountry({required int id}) async {
    final data = await _executor.mutation('mutation addressDeleteShippingOrderAndItemChangesCountry(\$id: Int!) { addressDeleteShippingOrderAndItemChangesCountry(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['addressDeleteShippingOrderAndItemChangesCountry'] as bool?;
  }

  Future<AddressShippingWeightBasedCountry?> addressCreateShippingWeightBasedCountry({required AddressShippingWeightBasedCountryInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation addressCreateShippingWeightBasedCountry(\$input: AddressShippingWeightBasedCountryInput!) { addressCreateShippingWeightBasedCountry(input: \$input) ${selectionSet ?? '{ id weightbased_id country_id address_country { iso31661a2 prinatable_name name display_order is_shipping_country } }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['addressCreateShippingWeightBasedCountry'] == null ? null : AddressShippingWeightBasedCountry.fromJson((data['addressCreateShippingWeightBasedCountry'] as Map).cast<String, dynamic>());
  }

  Future<AddressShippingWeightBasedCountry?> addressUpdateShippingWeightBasedCountry({required int id, required AddressShippingWeightBasedCountryInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation addressUpdateShippingWeightBasedCountry(\$id: Int!, \$input: AddressShippingWeightBasedCountryInput!) { addressUpdateShippingWeightBasedCountry(id: \$id, input: \$input) ${selectionSet ?? '{ id weightbased_id country_id address_country { iso31661a2 prinatable_name name display_order is_shipping_country } }'} }', variables: {
      'id': id,
      'input': input.toJson(),
    }
    );
    return data['addressUpdateShippingWeightBasedCountry'] == null ? null : AddressShippingWeightBasedCountry.fromJson((data['addressUpdateShippingWeightBasedCountry'] as Map).cast<String, dynamic>());
  }

  Future<bool?> addressDeleteShippingWeightBasedCountry({required int id}) async {
    final data = await _executor.mutation('mutation addressDeleteShippingWeightBasedCountry(\$id: Int!) { addressDeleteShippingWeightBasedCountry(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['addressDeleteShippingWeightBasedCountry'] as bool?;
  }

  Future<AddressUserAddress?> addressCreateUserAddress({required AddressAdminUserAddressInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation addressCreateUserAddress(\$input: AddressAdminUserAddressInput!) { addressCreateUserAddress(input: \$input) ${selectionSet ?? '{ id title first_name last_name line1 line2 line3 line4 state postcode search_text country_id phone_number notes is_default_for_shipping is_default_for_billing number_orders_as_shipping_address number_orders_as_billing_address hash date_created user_id address_country { iso31661a2 prinatable_name name display_order is_shipping_country } }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['addressCreateUserAddress'] == null ? null : AddressUserAddress.fromJson((data['addressCreateUserAddress'] as Map).cast<String, dynamic>());
  }

  Future<AddressUserAddress?> addressUpdateUserAddress({required int id, required AddressAdminUserAddressInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation addressUpdateUserAddress(\$id: Int!, \$input: AddressAdminUserAddressInput!) { addressUpdateUserAddress(id: \$id, input: \$input) ${selectionSet ?? '{ id title first_name last_name line1 line2 line3 line4 state postcode search_text country_id phone_number notes is_default_for_shipping is_default_for_billing number_orders_as_shipping_address number_orders_as_billing_address hash date_created user_id address_country { iso31661a2 prinatable_name name display_order is_shipping_country } }'} }', variables: {
      'id': id,
      'input': input.toJson(),
    }
    );
    return data['addressUpdateUserAddress'] == null ? null : AddressUserAddress.fromJson((data['addressUpdateUserAddress'] as Map).cast<String, dynamic>());
  }

  Future<bool?> addressDeleteUserAddress({required int id}) async {
    final data = await _executor.mutation('mutation addressDeleteUserAddress(\$id: Int!) { addressDeleteUserAddress(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['addressDeleteUserAddress'] as bool?;
  }

  Future<AddressUserAddress?> createMyUserAddress({required AddressUserAddressInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation createMyUserAddress(\$input: AddressUserAddressInput!) { createMyUserAddress(input: \$input) ${selectionSet ?? '{ id title first_name last_name line1 line2 line3 line4 state postcode search_text country_id phone_number notes is_default_for_shipping is_default_for_billing number_orders_as_shipping_address number_orders_as_billing_address hash date_created user_id address_country { iso31661a2 prinatable_name name display_order is_shipping_country } }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['createMyUserAddress'] == null ? null : AddressUserAddress.fromJson((data['createMyUserAddress'] as Map).cast<String, dynamic>());
  }

  Future<AddressUserAddress?> updateMyUserAddress({required int id, required AddressUserAddressInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation updateMyUserAddress(\$id: Int!, \$input: AddressUserAddressInput!) { updateMyUserAddress(id: \$id, input: \$input) ${selectionSet ?? '{ id title first_name last_name line1 line2 line3 line4 state postcode search_text country_id phone_number notes is_default_for_shipping is_default_for_billing number_orders_as_shipping_address number_orders_as_billing_address hash date_created user_id address_country { iso31661a2 prinatable_name name display_order is_shipping_country } }'} }', variables: {
      'id': id,
      'input': input.toJson(),
    }
    );
    return data['updateMyUserAddress'] == null ? null : AddressUserAddress.fromJson((data['updateMyUserAddress'] as Map).cast<String, dynamic>());
  }

  Future<bool?> deleteMyUserAddress({required int id}) async {
    final data = await _executor.mutation('mutation deleteMyUserAddress(\$id: Int!) { deleteMyUserAddress(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['deleteMyUserAddress'] as bool?;
  }

}
