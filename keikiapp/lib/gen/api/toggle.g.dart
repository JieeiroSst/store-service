// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

import '../gateway_executor.dart';
import '../models.g.dart';

class ToggleApi {
  const ToggleApi(this._executor);

  final GatewayExecutor _executor;

  Future<List<ToggleRole>?> toggleRoles({String? selectionSet}) async {
    final data = await _executor.query('query toggleRoles { toggleRoles ${selectionSet ?? '{ id name description }'} }'
    );
    return (data['toggleRoles'] as List?)?.map((e) => (e == null ? null : ToggleRole.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<List<ToggleEnvironment>?> toggleEnvironments({String? selectionSet}) async {
    final data = await _executor.query('query toggleEnvironments { toggleEnvironments ${selectionSet ?? '{ id name type sortOrder enabled createdAt }'} }'
    );
    return (data['toggleEnvironments'] as List?)?.map((e) => (e == null ? null : ToggleEnvironment.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<List<ToggleAPIToken>?> toggleTokens({String? selectionSet}) async {
    final data = await _executor.query('query toggleTokens { toggleTokens ${selectionSet ?? '{ id name tokenPrefix type projectId environmentId expiresAt createdBy createdAt }'} }'
    );
    return (data['toggleTokens'] as List?)?.map((e) => (e == null ? null : ToggleAPIToken.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<List<ToggleProject>?> toggleProjects({String? selectionSet}) async {
    final data = await _executor.query('query toggleProjects { toggleProjects ${selectionSet ?? '{ id name description createdBy createdAt updatedAt }'} }'
    );
    return (data['toggleProjects'] as List?)?.map((e) => (e == null ? null : ToggleProject.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<ToggleProject?> toggleProject({required String projectId, String? selectionSet}) async {
    final data = await _executor.query('query toggleProject(\$projectId: String!) { toggleProject(projectId: \$projectId) ${selectionSet ?? '{ id name description createdBy createdAt updatedAt }'} }', variables: {
      'projectId': projectId,
    }
    );
    return data['toggleProject'] == null ? null : ToggleProject.fromJson((data['toggleProject'] as Map).cast<String, dynamic>());
  }

  Future<List<ToggleProjectMembership>?> toggleProjectMembers({required String projectId, String? selectionSet}) async {
    final data = await _executor.query('query toggleProjectMembers(\$projectId: String!) { toggleProjectMembers(projectId: \$projectId) ${selectionSet ?? '{ id projectId userId roleId createdAt }'} }', variables: {
      'projectId': projectId,
    }
    );
    return (data['toggleProjectMembers'] as List?)?.map((e) => (e == null ? null : ToggleProjectMembership.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<List<ToggleAuditEvent>?> toggleAudit({required String projectId, String? entityType, String? since, String? until, String? selectionSet}) async {
    final data = await _executor.query('query toggleAudit(\$projectId: String!, \$entityType: String, \$since: String, \$until: String) { toggleAudit(projectId: \$projectId, entityType: \$entityType, since: \$since, until: \$until) ${selectionSet ?? '{ id entityType entityId action projectId environmentId userId before after createdAt }'} }', variables: {
      'projectId': projectId,
      'entityType': ?entityType,
      'since': ?since,
      'until': ?until,
    }
    );
    return (data['toggleAudit'] as List?)?.map((e) => (e == null ? null : ToggleAuditEvent.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<List<ToggleFeatureFlag>?> toggleFlags({required String projectId, String? selectionSet}) async {
    final data = await _executor.query('query toggleFlags(\$projectId: String!) { toggleFlags(projectId: \$projectId) ${selectionSet ?? '{ id projectId key name description type archived createdBy createdAt updatedAt }'} }', variables: {
      'projectId': projectId,
    }
    );
    return (data['toggleFlags'] as List?)?.map((e) => (e == null ? null : ToggleFeatureFlag.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<ToggleFeatureFlag?> toggleFlag({required String projectId, required String key, String? selectionSet}) async {
    final data = await _executor.query('query toggleFlag(\$projectId: String!, \$key: String!) { toggleFlag(projectId: \$projectId, key: \$key) ${selectionSet ?? '{ id projectId key name description type archived createdBy createdAt updatedAt }'} }', variables: {
      'projectId': projectId,
      'key': key,
    }
    );
    return data['toggleFlag'] == null ? null : ToggleFeatureFlag.fromJson((data['toggleFlag'] as Map).cast<String, dynamic>());
  }

  Future<List<ToggleActivationStrategy>?> toggleStrategies({required String projectId, required String key, required String envName, String? selectionSet}) async {
    final data = await _executor.query('query toggleStrategies(\$projectId: String!, \$key: String!, \$envName: String!) { toggleStrategies(projectId: \$projectId, key: \$key, envName: \$envName) ${selectionSet ?? '{ id featureFlagEnvironmentId strategyType parameters sortOrder constraints { id strategyId contextField operator values caseInsensitive } createdAt updatedAt }'} }', variables: {
      'projectId': projectId,
      'key': key,
      'envName': envName,
    }
    );
    return (data['toggleStrategies'] as List?)?.map((e) => (e == null ? null : ToggleActivationStrategy.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<ToggleEnvironment?> toggleCreateEnvironment({required String name_, String? type_, int? sortOrder, String? selectionSet}) async {
    final data = await _executor.mutation('mutation toggleCreateEnvironment(\$name: String!, \$type: String, \$sortOrder: Int) { toggleCreateEnvironment(name: \$name, type: \$type, sortOrder: \$sortOrder) ${selectionSet ?? '{ id name type sortOrder enabled createdAt }'} }', variables: {
      'name': name_,
      'type': ?type_,
      'sortOrder': ?sortOrder,
    }
    );
    return data['toggleCreateEnvironment'] == null ? null : ToggleEnvironment.fromJson((data['toggleCreateEnvironment'] as Map).cast<String, dynamic>());
  }

  Future<ToggleEnvironment?> toggleUpdateEnvironment({required String environmentId, String? name_, String? type_, bool? enabled, String? selectionSet}) async {
    final data = await _executor.mutation('mutation toggleUpdateEnvironment(\$environmentId: String!, \$name: String, \$type: String, \$enabled: Boolean) { toggleUpdateEnvironment(environmentId: \$environmentId, name: \$name, type: \$type, enabled: \$enabled) ${selectionSet ?? '{ id name type sortOrder enabled createdAt }'} }', variables: {
      'environmentId': environmentId,
      'name': ?name_,
      'type': ?type_,
      'enabled': ?enabled,
    }
    );
    return data['toggleUpdateEnvironment'] == null ? null : ToggleEnvironment.fromJson((data['toggleUpdateEnvironment'] as Map).cast<String, dynamic>());
  }

  Future<bool?> toggleDeleteEnvironment({required String environmentId}) async {
    final data = await _executor.mutation('mutation toggleDeleteEnvironment(\$environmentId: String!) { toggleDeleteEnvironment(environmentId: \$environmentId) }', variables: {
      'environmentId': environmentId,
    }
    );
    return data['toggleDeleteEnvironment'] as bool?;
  }

  Future<Object?> toggleCreateToken({required String name_, required String type_, String? projectId, String? environmentId, String? expiresAt}) async {
    final data = await _executor.mutation('mutation toggleCreateToken(\$name: String!, \$type: String!, \$projectId: String, \$environmentId: String, \$expiresAt: String) { toggleCreateToken(name: \$name, type: \$type, projectId: \$projectId, environmentId: \$environmentId, expiresAt: \$expiresAt) }', variables: {
      'name': name_,
      'type': type_,
      'projectId': ?projectId,
      'environmentId': ?environmentId,
      'expiresAt': ?expiresAt,
    }
    );
    return data['toggleCreateToken'];
  }

  Future<bool?> toggleRevokeToken({required String tokenId}) async {
    final data = await _executor.mutation('mutation toggleRevokeToken(\$tokenId: String!) { toggleRevokeToken(tokenId: \$tokenId) }', variables: {
      'tokenId': tokenId,
    }
    );
    return data['toggleRevokeToken'] as bool?;
  }

  Future<Object?> toggleUpdateMember({required String membershipId, required String roleId}) async {
    final data = await _executor.mutation('mutation toggleUpdateMember(\$membershipId: String!, \$roleId: String!) { toggleUpdateMember(membershipId: \$membershipId, roleId: \$roleId) }', variables: {
      'membershipId': membershipId,
      'roleId': roleId,
    }
    );
    return data['toggleUpdateMember'];
  }

  Future<bool?> toggleRemoveMember({required String membershipId}) async {
    final data = await _executor.mutation('mutation toggleRemoveMember(\$membershipId: String!) { toggleRemoveMember(membershipId: \$membershipId) }', variables: {
      'membershipId': membershipId,
    }
    );
    return data['toggleRemoveMember'] as bool?;
  }

  Future<ToggleProject?> toggleCreateProject({required String name_, String? description, String? selectionSet}) async {
    final data = await _executor.mutation('mutation toggleCreateProject(\$name: String!, \$description: String) { toggleCreateProject(name: \$name, description: \$description) ${selectionSet ?? '{ id name description createdBy createdAt updatedAt }'} }', variables: {
      'name': name_,
      'description': ?description,
    }
    );
    return data['toggleCreateProject'] == null ? null : ToggleProject.fromJson((data['toggleCreateProject'] as Map).cast<String, dynamic>());
  }

  Future<ToggleProject?> toggleUpdateProject({required String projectId, String? name_, String? description, String? selectionSet}) async {
    final data = await _executor.mutation('mutation toggleUpdateProject(\$projectId: String!, \$name: String, \$description: String) { toggleUpdateProject(projectId: \$projectId, name: \$name, description: \$description) ${selectionSet ?? '{ id name description createdBy createdAt updatedAt }'} }', variables: {
      'projectId': projectId,
      'name': ?name_,
      'description': ?description,
    }
    );
    return data['toggleUpdateProject'] == null ? null : ToggleProject.fromJson((data['toggleUpdateProject'] as Map).cast<String, dynamic>());
  }

  Future<bool?> toggleDeleteProject({required String projectId}) async {
    final data = await _executor.mutation('mutation toggleDeleteProject(\$projectId: String!) { toggleDeleteProject(projectId: \$projectId) }', variables: {
      'projectId': projectId,
    }
    );
    return data['toggleDeleteProject'] as bool?;
  }

  Future<Object?> toggleAddMember({required String projectId, required String userId, required String roleId}) async {
    final data = await _executor.mutation('mutation toggleAddMember(\$projectId: String!, \$userId: String!, \$roleId: String!) { toggleAddMember(projectId: \$projectId, userId: \$userId, roleId: \$roleId) }', variables: {
      'projectId': projectId,
      'userId': userId,
      'roleId': roleId,
    }
    );
    return data['toggleAddMember'];
  }

  Future<ToggleFeatureFlag?> toggleCreateFlag({required String projectId, required String key, required String name_, String? description, String? type_, String? selectionSet}) async {
    final data = await _executor.mutation('mutation toggleCreateFlag(\$projectId: String!, \$key: String!, \$name: String!, \$description: String, \$type: String) { toggleCreateFlag(projectId: \$projectId, key: \$key, name: \$name, description: \$description, type: \$type) ${selectionSet ?? '{ id projectId key name description type archived createdBy createdAt updatedAt }'} }', variables: {
      'projectId': projectId,
      'key': key,
      'name': name_,
      'description': ?description,
      'type': ?type_,
    }
    );
    return data['toggleCreateFlag'] == null ? null : ToggleFeatureFlag.fromJson((data['toggleCreateFlag'] as Map).cast<String, dynamic>());
  }

  Future<ToggleFeatureFlag?> toggleUpdateFlag({required String projectId, required String key, String? name_, String? description, String? type_, String? selectionSet}) async {
    final data = await _executor.mutation('mutation toggleUpdateFlag(\$projectId: String!, \$key: String!, \$name: String, \$description: String, \$type: String) { toggleUpdateFlag(projectId: \$projectId, key: \$key, name: \$name, description: \$description, type: \$type) ${selectionSet ?? '{ id projectId key name description type archived createdBy createdAt updatedAt }'} }', variables: {
      'projectId': projectId,
      'key': key,
      'name': ?name_,
      'description': ?description,
      'type': ?type_,
    }
    );
    return data['toggleUpdateFlag'] == null ? null : ToggleFeatureFlag.fromJson((data['toggleUpdateFlag'] as Map).cast<String, dynamic>());
  }

  Future<bool?> toggleArchiveFlag({required String projectId, required String key}) async {
    final data = await _executor.mutation('mutation toggleArchiveFlag(\$projectId: String!, \$key: String!) { toggleArchiveFlag(projectId: \$projectId, key: \$key) }', variables: {
      'projectId': projectId,
      'key': key,
    }
    );
    return data['toggleArchiveFlag'] as bool?;
  }

  Future<Object?> toggleFlagOn({required String projectId, required String key, required String envName}) async {
    final data = await _executor.mutation('mutation toggleFlagOn(\$projectId: String!, \$key: String!, \$envName: String!) { toggleFlagOn(projectId: \$projectId, key: \$key, envName: \$envName) }', variables: {
      'projectId': projectId,
      'key': key,
      'envName': envName,
    }
    );
    return data['toggleFlagOn'];
  }

  Future<Object?> toggleFlagOff({required String projectId, required String key, required String envName}) async {
    final data = await _executor.mutation('mutation toggleFlagOff(\$projectId: String!, \$key: String!, \$envName: String!) { toggleFlagOff(projectId: \$projectId, key: \$key, envName: \$envName) }', variables: {
      'projectId': projectId,
      'key': key,
      'envName': envName,
    }
    );
    return data['toggleFlagOff'];
  }

  Future<ToggleActivationStrategy?> toggleCreateStrategy({required String projectId, required String key, required String envName, required ToggleStrategyInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation toggleCreateStrategy(\$projectId: String!, \$key: String!, \$envName: String!, \$input: ToggleStrategyInput!) { toggleCreateStrategy(projectId: \$projectId, key: \$key, envName: \$envName, input: \$input) ${selectionSet ?? '{ id featureFlagEnvironmentId strategyType parameters sortOrder constraints { id strategyId contextField operator values caseInsensitive } createdAt updatedAt }'} }', variables: {
      'projectId': projectId,
      'key': key,
      'envName': envName,
      'input': input.toJson(),
    }
    );
    return data['toggleCreateStrategy'] == null ? null : ToggleActivationStrategy.fromJson((data['toggleCreateStrategy'] as Map).cast<String, dynamic>());
  }

  Future<ToggleActivationStrategy?> toggleUpdateStrategy({required String projectId, required String key, required String envName, required String strategyId, required ToggleStrategyInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation toggleUpdateStrategy(\$projectId: String!, \$key: String!, \$envName: String!, \$strategyId: String!, \$input: ToggleStrategyInput!) { toggleUpdateStrategy(projectId: \$projectId, key: \$key, envName: \$envName, strategyId: \$strategyId, input: \$input) ${selectionSet ?? '{ id featureFlagEnvironmentId strategyType parameters sortOrder constraints { id strategyId contextField operator values caseInsensitive } createdAt updatedAt }'} }', variables: {
      'projectId': projectId,
      'key': key,
      'envName': envName,
      'strategyId': strategyId,
      'input': input.toJson(),
    }
    );
    return data['toggleUpdateStrategy'] == null ? null : ToggleActivationStrategy.fromJson((data['toggleUpdateStrategy'] as Map).cast<String, dynamic>());
  }

  Future<bool?> toggleDeleteStrategy({required String projectId, required String key, required String envName, required String strategyId}) async {
    final data = await _executor.mutation('mutation toggleDeleteStrategy(\$projectId: String!, \$key: String!, \$envName: String!, \$strategyId: String!) { toggleDeleteStrategy(projectId: \$projectId, key: \$key, envName: \$envName, strategyId: \$strategyId) }', variables: {
      'projectId': projectId,
      'key': key,
      'envName': envName,
      'strategyId': strategyId,
    }
    );
    return data['toggleDeleteStrategy'] as bool?;
  }

}
