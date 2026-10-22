// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

import '../gateway_executor.dart';
import '../models.g.dart';

class PartnerApi {
  const PartnerApi(this._executor);

  final GatewayExecutor _executor;

  Future<PartnerPartnerPage?> partners({int? limit, int? page, String? sort, String? selectionSet}) async {
    final data = await _executor.query('query partners(\$limit: Int, \$page: Int, \$sort: String) { partners(limit: \$limit, page: \$page, sort: \$sort) ${selectionSet ?? '{ limit page sort total_rows total_pages rows { id type name email phone address status score user_id created_at updated_at deleted_at } }'} }', variables: {
      'limit': ?limit,
      'page': ?page,
      'sort': ?sort,
    }
    );
    return data['partners'] == null ? null : PartnerPartnerPage.fromJson((data['partners'] as Map).cast<String, dynamic>());
  }

  Future<PartnerPartnerPage?> partnerSearchPartners({int? limit, int? page, String? sort, String? type_, String? status, String? name_, String? userId, String? selectionSet}) async {
    final data = await _executor.query('query partnerSearchPartners(\$limit: Int, \$page: Int, \$sort: String, \$type: String, \$status: String, \$name: String, \$user_id: String) { partnerSearchPartners(limit: \$limit, page: \$page, sort: \$sort, type: \$type, status: \$status, name: \$name, user_id: \$user_id) ${selectionSet ?? '{ limit page sort total_rows total_pages rows { id type name email phone address status score user_id created_at updated_at deleted_at } }'} }', variables: {
      'limit': ?limit,
      'page': ?page,
      'sort': ?sort,
      'type': ?type_,
      'status': ?status,
      'name': ?name_,
      'user_id': ?userId,
    }
    );
    return data['partnerSearchPartners'] == null ? null : PartnerPartnerPage.fromJson((data['partnerSearchPartners'] as Map).cast<String, dynamic>());
  }

  Future<PartnerPartnerStatistics?> partnerStatistics({String? selectionSet}) async {
    final data = await _executor.query('query partnerStatistics { partnerStatistics ${selectionSet ?? '{ total by_status by_type }'} }'
    );
    return data['partnerStatistics'] == null ? null : PartnerPartnerStatistics.fromJson((data['partnerStatistics'] as Map).cast<String, dynamic>());
  }

  Future<PartnerPartnerActivity?> partnerActivity({required String id, String? selectionSet}) async {
    final data = await _executor.query('query partnerActivity(\$id: String!) { partnerActivity(id: \$id) ${selectionSet ?? '{ partner_id status score has_open_partnership last_activity_at eligible_for_closure is_active }'} }', variables: {
      'id': id,
    }
    );
    return data['partnerActivity'] == null ? null : PartnerPartnerActivity.fromJson((data['partnerActivity'] as Map).cast<String, dynamic>());
  }

  Future<PartnerPartner?> partner({required String id, String? selectionSet}) async {
    final data = await _executor.query('query partner(\$id: String!) { partner(id: \$id) ${selectionSet ?? '{ id type name email phone address status score user_id created_at updated_at deleted_at }'} }', variables: {
      'id': id,
    }
    );
    return data['partner'] == null ? null : PartnerPartner.fromJson((data['partner'] as Map).cast<String, dynamic>());
  }

  Future<PartnerPartnership?> partnership({required String id, String? selectionSet}) async {
    final data = await _executor.query('query partnership(\$id: String!) { partnership(id: \$id) ${selectionSet ?? '{ id project_id description started_on expires_on projects { id name user_id created_at } user_id created_at }'} }', variables: {
      'id': id,
    }
    );
    return data['partnership'] == null ? null : PartnerPartnership.fromJson((data['partnership'] as Map).cast<String, dynamic>());
  }

  Future<PartnerPartnershipPage?> partnerships({int? limit, int? page, String? sort, String? selectionSet}) async {
    final data = await _executor.query('query partnerships(\$limit: Int, \$page: Int, \$sort: String) { partnerships(limit: \$limit, page: \$page, sort: \$sort) ${selectionSet ?? '{ limit page sort total_rows total_pages rows { id project_id description started_on expires_on projects { id name user_id created_at } user_id created_at } }'} }', variables: {
      'limit': ?limit,
      'page': ?page,
      'sort': ?sort,
    }
    );
    return data['partnerships'] == null ? null : PartnerPartnershipPage.fromJson((data['partnerships'] as Map).cast<String, dynamic>());
  }

  Future<PartnerPartnershipsPartner?> partnershipsPartner({required String id, String? selectionSet}) async {
    final data = await _executor.query('query partnershipsPartner(\$id: String!) { partnershipsPartner(id: \$id) ${selectionSet ?? '{ id partner_id partnership_id content joined_on left_on partners { id type name email phone address status score user_id created_at updated_at deleted_at } partnerships { id project_id description started_on expires_on projects { id name user_id created_at } user_id created_at } user_id created_at }'} }', variables: {
      'id': id,
    }
    );
    return data['partnershipsPartner'] == null ? null : PartnerPartnershipsPartner.fromJson((data['partnershipsPartner'] as Map).cast<String, dynamic>());
  }

  Future<PartnerPartnershipsPartnerPage?> partnershipsPartners({int? limit, int? page, String? sort, String? selectionSet}) async {
    final data = await _executor.query('query partnershipsPartners(\$limit: Int, \$page: Int, \$sort: String) { partnershipsPartners(limit: \$limit, page: \$page, sort: \$sort) ${selectionSet ?? '{ limit page sort total_rows total_pages rows { id partner_id partnership_id content joined_on left_on partners { id type name email phone address status score user_id created_at updated_at deleted_at } partnerships { id project_id description started_on expires_on user_id created_at } user_id created_at } }'} }', variables: {
      'limit': ?limit,
      'page': ?page,
      'sort': ?sort,
    }
    );
    return data['partnershipsPartners'] == null ? null : PartnerPartnershipsPartnerPage.fromJson((data['partnershipsPartners'] as Map).cast<String, dynamic>());
  }

  Future<PartnerProjectPage?> partnerProjects({int? limit, int? page, String? sort, String? selectionSet}) async {
    final data = await _executor.query('query partnerProjects(\$limit: Int, \$page: Int, \$sort: String) { partnerProjects(limit: \$limit, page: \$page, sort: \$sort) ${selectionSet ?? '{ limit page sort total_rows total_pages rows { id name user_id created_at } }'} }', variables: {
      'limit': ?limit,
      'page': ?page,
      'sort': ?sort,
    }
    );
    return data['partnerProjects'] == null ? null : PartnerProjectPage.fromJson((data['partnerProjects'] as Map).cast<String, dynamic>());
  }

  Future<PartnerProject?> partnerProject({required String id, String? selectionSet}) async {
    final data = await _executor.query('query partnerProject(\$id: String!) { partnerProject(id: \$id) ${selectionSet ?? '{ id name user_id created_at }'} }', variables: {
      'id': id,
    }
    );
    return data['partnerProject'] == null ? null : PartnerProject.fromJson((data['partnerProject'] as Map).cast<String, dynamic>());
  }

  Future<Object?> partnerCreatePartner({required PartnerPartnerInput input}) async {
    final data = await _executor.mutation('mutation partnerCreatePartner(\$input: PartnerPartnerInput!) { partnerCreatePartner(input: \$input) }', variables: {
      'input': input.toJson(),
    }
    );
    return data['partnerCreatePartner'];
  }

  Future<Object?> partnerUpdatePartner({required String id, required PartnerPartnerInput input}) async {
    final data = await _executor.mutation('mutation partnerUpdatePartner(\$id: String!, \$input: PartnerPartnerInput!) { partnerUpdatePartner(id: \$id, input: \$input) }', variables: {
      'id': id,
      'input': input.toJson(),
    }
    );
    return data['partnerUpdatePartner'];
  }

  Future<Object?> partnerDeletePartner({required String id}) async {
    final data = await _executor.mutation('mutation partnerDeletePartner(\$id: String!) { partnerDeletePartner(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['partnerDeletePartner'];
  }

  Future<Object?> partnerCreatePartnership({required PartnerPartnershipInput input}) async {
    final data = await _executor.mutation('mutation partnerCreatePartnership(\$input: PartnerPartnershipInput!) { partnerCreatePartnership(input: \$input) }', variables: {
      'input': input.toJson(),
    }
    );
    return data['partnerCreatePartnership'];
  }

  Future<Object?> partnerUpdatePartnership({required String id, required PartnerPartnershipInput input}) async {
    final data = await _executor.mutation('mutation partnerUpdatePartnership(\$id: String!, \$input: PartnerPartnershipInput!) { partnerUpdatePartnership(id: \$id, input: \$input) }', variables: {
      'id': id,
      'input': input.toJson(),
    }
    );
    return data['partnerUpdatePartnership'];
  }

  Future<Object?> partnerDeletePartnership({required String id}) async {
    final data = await _executor.mutation('mutation partnerDeletePartnership(\$id: String!) { partnerDeletePartnership(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['partnerDeletePartnership'];
  }

  Future<Object?> partnerCreatePartnershipsPartner({required PartnerPartnershipsPartnerInput input}) async {
    final data = await _executor.mutation('mutation partnerCreatePartnershipsPartner(\$input: PartnerPartnershipsPartnerInput!) { partnerCreatePartnershipsPartner(input: \$input) }', variables: {
      'input': input.toJson(),
    }
    );
    return data['partnerCreatePartnershipsPartner'];
  }

  Future<Object?> partnerUpdatePartnershipsPartner({required String id, required PartnerPartnershipsPartnerInput input}) async {
    final data = await _executor.mutation('mutation partnerUpdatePartnershipsPartner(\$id: String!, \$input: PartnerPartnershipsPartnerInput!) { partnerUpdatePartnershipsPartner(id: \$id, input: \$input) }', variables: {
      'id': id,
      'input': input.toJson(),
    }
    );
    return data['partnerUpdatePartnershipsPartner'];
  }

  Future<Object?> partnerDeletePartnershipsPartner({required String id}) async {
    final data = await _executor.mutation('mutation partnerDeletePartnershipsPartner(\$id: String!) { partnerDeletePartnershipsPartner(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['partnerDeletePartnershipsPartner'];
  }

  Future<Object?> partnerCreateProject({required PartnerProjectInput input}) async {
    final data = await _executor.mutation('mutation partnerCreateProject(\$input: PartnerProjectInput!) { partnerCreateProject(input: \$input) }', variables: {
      'input': input.toJson(),
    }
    );
    return data['partnerCreateProject'];
  }

  Future<Object?> partnerUpdateProject({required String id, required PartnerProjectInput input}) async {
    final data = await _executor.mutation('mutation partnerUpdateProject(\$id: String!, \$input: PartnerProjectInput!) { partnerUpdateProject(id: \$id, input: \$input) }', variables: {
      'id': id,
      'input': input.toJson(),
    }
    );
    return data['partnerUpdateProject'];
  }

  Future<Object?> partnerDeleteProject({required String id}) async {
    final data = await _executor.mutation('mutation partnerDeleteProject(\$id: String!) { partnerDeleteProject(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['partnerDeleteProject'];
  }

  Future<Object?> partnerUpdateStatus({required String id, required String status}) async {
    final data = await _executor.mutation('mutation partnerUpdateStatus(\$id: String!, \$status: String!) { partnerUpdateStatus(id: \$id, status: \$status) }', variables: {
      'id': id,
      'status': status,
    }
    );
    return data['partnerUpdateStatus'];
  }

  Future<Object?> partnerRestore({required String id}) async {
    final data = await _executor.mutation('mutation partnerRestore(\$id: String!) { partnerRestore(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['partnerRestore'];
  }

  Future<Object?> partnerAdjustScore({required String id, required int points}) async {
    final data = await _executor.mutation('mutation partnerAdjustScore(\$id: String!, \$points: Int!) { partnerAdjustScore(id: \$id, points: \$points) }', variables: {
      'id': id,
      'points': points,
    }
    );
    return data['partnerAdjustScore'];
  }

  Future<Object?> partnerCloseInactive() async {
    final data = await _executor.mutation('mutation partnerCloseInactive { partnerCloseInactive }'
    );
    return data['partnerCloseInactive'];
  }

}
