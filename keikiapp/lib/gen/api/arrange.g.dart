// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

import '../gateway_executor.dart';
import '../models.g.dart';

class ArrangeApi {
  const ArrangeApi(this._executor);

  final GatewayExecutor _executor;

  Future<List<ArrangeDispatchAssignment>?> arrangeDispatchAssignments({required int driverId, String? selectionSet}) async {
    final data = await _executor.query('query arrangeDispatchAssignments(\$driverId: Int!) { arrangeDispatchAssignments(driverId: \$driverId) ${selectionSet ?? '{ id orderId driverId sequence status }'} }', variables: {
      'driverId': driverId,
    }
    );
    return (data['arrangeDispatchAssignments'] as List?)?.map((e) => (e == null ? null : ArrangeDispatchAssignment.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<List<ArrangeShift>?> arrangeShifts({required int storeId, String? selectionSet}) async {
    final data = await _executor.query('query arrangeShifts(\$storeId: Int!) { arrangeShifts(storeId: \$storeId) ${selectionSet ?? '{ id employeeId storeId role startTime endTime status }'} }', variables: {
      'storeId': storeId,
    }
    );
    return (data['arrangeShifts'] as List?)?.map((e) => (e == null ? null : ArrangeShift.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<List<ArrangeSpace>?> arrangeSpaces({required int storeId, int? minCapacity, String? selectionSet}) async {
    final data = await _executor.query('query arrangeSpaces(\$storeId: Int!, \$minCapacity: Int) { arrangeSpaces(storeId: \$storeId, minCapacity: \$minCapacity) ${selectionSet ?? '{ id storeId name type capacity }'} }', variables: {
      'storeId': storeId,
      'minCapacity': ?minCapacity,
    }
    );
    return (data['arrangeSpaces'] as List?)?.map((e) => (e == null ? null : ArrangeSpace.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<List<ArrangeSpaceAssignment>?> arrangeSpaceAssignments({required int storeId, String? selectionSet}) async {
    final data = await _executor.query('query arrangeSpaceAssignments(\$storeId: Int!) { arrangeSpaceAssignments(storeId: \$storeId) ${selectionSet ?? '{ id spaceId referenceId partySize arrangedFrom arrangedTo status }'} }', variables: {
      'storeId': storeId,
    }
    );
    return (data['arrangeSpaceAssignments'] as List?)?.map((e) => (e == null ? null : ArrangeSpaceAssignment.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<ArrangeDispatchAssignment?> arrangeDispatch({required ArrangeDispatchInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation arrangeDispatch(\$input: ArrangeDispatchInput!) { arrangeDispatch(input: \$input) ${selectionSet ?? '{ id orderId driverId sequence status }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['arrangeDispatch'] == null ? null : ArrangeDispatchAssignment.fromJson((data['arrangeDispatch'] as Map).cast<String, dynamic>());
  }

  Future<bool?> arrangeMarkDelivered({required int id}) async {
    final data = await _executor.mutation('mutation arrangeMarkDelivered(\$id: Int!) { arrangeMarkDelivered(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['arrangeMarkDelivered'] as bool?;
  }

  Future<ArrangeShift?> arrangeShift({required ArrangeShiftInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation arrangeShift(\$input: ArrangeShiftInput!) { arrangeShift(input: \$input) ${selectionSet ?? '{ id employeeId storeId role startTime endTime status }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['arrangeShift'] == null ? null : ArrangeShift.fromJson((data['arrangeShift'] as Map).cast<String, dynamic>());
  }

  Future<bool?> arrangeCancelShift({required int id}) async {
    final data = await _executor.mutation('mutation arrangeCancelShift(\$id: Int!) { arrangeCancelShift(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['arrangeCancelShift'] as bool?;
  }

  Future<ArrangeSpaceAssignment?> arrangeSpace({required ArrangeSpaceInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation arrangeSpace(\$input: ArrangeSpaceInput!) { arrangeSpace(input: \$input) ${selectionSet ?? '{ id spaceId referenceId partySize arrangedFrom arrangedTo status }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['arrangeSpace'] == null ? null : ArrangeSpaceAssignment.fromJson((data['arrangeSpace'] as Map).cast<String, dynamic>());
  }

  Future<bool?> arrangeReleaseSpace({required int id}) async {
    final data = await _executor.mutation('mutation arrangeReleaseSpace(\$id: Int!) { arrangeReleaseSpace(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['arrangeReleaseSpace'] as bool?;
  }

}
