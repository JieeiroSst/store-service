// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

import '../gateway_executor.dart';
import '../models.g.dart';

class IdentifiedApi {
  const IdentifiedApi(this._executor);

  final GatewayExecutor _executor;

  Future<List<GymClass>?> gymClasses({String? selectionSet}) async {
    final data = await _executor.public('query gymClasses { gymClasses ${selectionSet ?? '{ id name day time coach description }'} }'
    );
    return (data['gymClasses'] as List?)?.map((e) => (e == null ? null : GymClass.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<GymClass?> gymClass({required String classId, String? selectionSet}) async {
    final data = await _executor.public('query gymClass(\$class_id: String!) { gymClass(class_id: \$class_id) ${selectionSet ?? '{ id name day time coach description }'} }', variables: {
      'class_id': classId,
    }
    );
    return data['gymClass'] == null ? null : GymClass.fromJson((data['gymClass'] as Map).cast<String, dynamic>());
  }

  Future<GymPass?> gymPassVerification({required String gymPassId, String? selectionSet}) async {
    final data = await _executor.query('query gymPassVerification(\$gym_pass_id: String!) { gymPassVerification(gym_pass_id: \$gym_pass_id) ${selectionSet ?? '{ id is_active }'} }', variables: {
      'gym_pass_id': gymPassId,
    }
    );
    return data['gymPassVerification'] == null ? null : GymPass.fromJson((data['gymPassVerification'] as Map).cast<String, dynamic>());
  }

  Future<GymClient?> gymCreateClient({required GymClientInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation gymCreateClient(\$input: GymClientInput!) { gymCreateClient(input: \$input) ${selectionSet ?? '{ id first_name last_name email status }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['gymCreateClient'] == null ? null : GymClient.fromJson((data['gymCreateClient'] as Map).cast<String, dynamic>());
  }

  Future<GymClient?> gymUpdateClient({required String clientId, required GymClientInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation gymUpdateClient(\$client_id: String!, \$input: GymClientInput!) { gymUpdateClient(client_id: \$client_id, input: \$input) ${selectionSet ?? '{ id first_name last_name email status }'} }', variables: {
      'client_id': clientId,
      'input': input.toJson(),
    }
    );
    return data['gymUpdateClient'] == null ? null : GymClient.fromJson((data['gymUpdateClient'] as Map).cast<String, dynamic>());
  }

  Future<bool?> gymArchiveClient({required String clientId}) async {
    final data = await _executor.mutation('mutation gymArchiveClient(\$client_id: String!) { gymArchiveClient(client_id: \$client_id) }', variables: {
      'client_id': clientId,
    }
    );
    return data['gymArchiveClient'] as bool?;
  }

  Future<Object?> gymExportClients({required String format}) async {
    final data = await _executor.mutation('mutation gymExportClients(\$format: String!) { gymExportClients(format: \$format) }', variables: {
      'format': format,
    }
    );
    return data['gymExportClients'];
  }

  Future<GymPass?> gymCreatePass({required String validity, String? selectionSet}) async {
    final data = await _executor.mutation('mutation gymCreatePass(\$validity: String!) { gymCreatePass(validity: \$validity) ${selectionSet ?? '{ id is_active }'} }', variables: {
      'validity': validity,
    }
    );
    return data['gymCreatePass'] == null ? null : GymPass.fromJson((data['gymCreatePass'] as Map).cast<String, dynamic>());
  }

  Future<GymPass?> gymPausePass({required String gymPassId, required int days, String? selectionSet}) async {
    final data = await _executor.mutation('mutation gymPausePass(\$gym_pass_id: String!, \$days: Int!) { gymPausePass(gym_pass_id: \$gym_pass_id, days: \$days) ${selectionSet ?? '{ id is_active }'} }', variables: {
      'gym_pass_id': gymPassId,
      'days': days,
    }
    );
    return data['gymPausePass'] == null ? null : GymPass.fromJson((data['gymPausePass'] as Map).cast<String, dynamic>());
  }

  Future<GymPass?> gymRenewPass({required String gymPassId, String? selectionSet}) async {
    final data = await _executor.mutation('mutation gymRenewPass(\$gym_pass_id: String!) { gymRenewPass(gym_pass_id: \$gym_pass_id) ${selectionSet ?? '{ id is_active }'} }', variables: {
      'gym_pass_id': gymPassId,
    }
    );
    return data['gymRenewPass'] == null ? null : GymPass.fromJson((data['gymRenewPass'] as Map).cast<String, dynamic>());
  }

  Future<GymClass?> gymCreateClass({required GymClassInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation gymCreateClass(\$input: GymClassInput!) { gymCreateClass(input: \$input) ${selectionSet ?? '{ id name day time coach description }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['gymCreateClass'] == null ? null : GymClass.fromJson((data['gymCreateClass'] as Map).cast<String, dynamic>());
  }

  Future<GymClass?> gymUpdateClass({required String classId, required GymClassInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation gymUpdateClass(\$class_id: String!, \$input: GymClassInput!) { gymUpdateClass(class_id: \$class_id, input: \$input) ${selectionSet ?? '{ id name day time coach description }'} }', variables: {
      'class_id': classId,
      'input': input.toJson(),
    }
    );
    return data['gymUpdateClass'] == null ? null : GymClass.fromJson((data['gymUpdateClass'] as Map).cast<String, dynamic>());
  }

  Future<bool?> gymDeleteClass({required String classId}) async {
    final data = await _executor.mutation('mutation gymDeleteClass(\$class_id: String!) { gymDeleteClass(class_id: \$class_id) }', variables: {
      'class_id': classId,
    }
    );
    return data['gymDeleteClass'] as bool?;
  }

}
