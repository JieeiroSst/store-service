// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

import '../gateway_executor.dart';
import '../models.g.dart';

class MedicalApi {
  const MedicalApi(this._executor);

  final GatewayExecutor _executor;

  Future<MedicalMedicineList?> medicalMedicines({String? q, int? limit, int? offset, String? selectionSet}) async {
    final data = await _executor.query('query medicalMedicines(\$q: String, \$limit: Int, \$offset: Int) { medicalMedicines(q: \$q, limit: \$limit, offset: \$offset) ${selectionSet ?? '{ items { id code name ingredients allergen_groups dosage_form strength unit requires_prescription reorder_level created_at updated_at } total }'} }', variables: {
      'q': ?q,
      'limit': ?limit,
      'offset': ?offset,
    }
    );
    return data['medicalMedicines'] == null ? null : MedicalMedicineList.fromJson((data['medicalMedicines'] as Map).cast<String, dynamic>());
  }

  Future<MedicalMedicine?> medicalMedicine({required int id, String? selectionSet}) async {
    final data = await _executor.query('query medicalMedicine(\$id: Int!) { medicalMedicine(id: \$id) ${selectionSet ?? '{ id code name ingredients allergen_groups dosage_form strength unit requires_prescription reorder_level created_at updated_at }'} }', variables: {
      'id': id,
    }
    );
    return data['medicalMedicine'] == null ? null : MedicalMedicine.fromJson((data['medicalMedicine'] as Map).cast<String, dynamic>());
  }

  Future<MedicalStockSummary?> medicalStock({required int id, String? selectionSet}) async {
    final data = await _executor.query('query medicalStock(\$id: Int!) { medicalStock(id: \$id) ${selectionSet ?? '{ medicine_id available expired batches { id medicine_id batch_number expiry_date quantity received_quantity unit_cost supplier received_at created_at updated_at } }'} }', variables: {
      'id': id,
    }
    );
    return data['medicalStock'] == null ? null : MedicalStockSummary.fromJson((data['medicalStock'] as Map).cast<String, dynamic>());
  }

  Future<List<MedicalStockMovement>?> medicalMovements({required int id, int? limit, int? offset, String? selectionSet}) async {
    final data = await _executor.query('query medicalMovements(\$id: Int!, \$limit: Int, \$offset: Int) { medicalMovements(id: \$id, limit: \$limit, offset: \$offset) ${selectionSet ?? '{ id medicine_id batch_id type quantity reason reference created_by created_at }'} }', variables: {
      'id': id,
      'limit': ?limit,
      'offset': ?offset,
    }
    );
    return (data['medicalMovements'] as List?)?.map((e) => (e == null ? null : MedicalStockMovement.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<List<MedicalBatch>?> medicalExpiring({int? days, String? selectionSet}) async {
    final data = await _executor.query('query medicalExpiring(\$days: Int) { medicalExpiring(days: \$days) ${selectionSet ?? '{ id medicine_id batch_number expiry_date quantity received_quantity unit_cost supplier received_at created_at updated_at }'} }', variables: {
      'days': ?days,
    }
    );
    return (data['medicalExpiring'] as List?)?.map((e) => (e == null ? null : MedicalBatch.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<List<MedicalLowStockItem>?> medicalLowStock({String? selectionSet}) async {
    final data = await _executor.query('query medicalLowStock { medicalLowStock ${selectionSet ?? '{ medicine { id code name ingredients allergen_groups dosage_form strength unit requires_prescription reorder_level created_at updated_at } available }'} }'
    );
    return (data['medicalLowStock'] as List?)?.map((e) => (e == null ? null : MedicalLowStockItem.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<List<MedicalInteractionRule>?> medicalInteractions({int? limit, int? offset, String? selectionSet}) async {
    final data = await _executor.query('query medicalInteractions(\$limit: Int, \$offset: Int) { medicalInteractions(limit: \$limit, offset: \$offset) ${selectionSet ?? '{ id ingredient_a ingredient_b severity description source created_at updated_at }'} }', variables: {
      'limit': ?limit,
      'offset': ?offset,
    }
    );
    return (data['medicalInteractions'] as List?)?.map((e) => (e == null ? null : MedicalInteractionRule.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<MedicalTerminologyList?> medicalTerminology({String? selectionSet}) async {
    final data = await _executor.query('query medicalTerminology { medicalTerminology ${selectionSet ?? '{ aliases { id alias canonical source created_at } allergen_groups { id ingredient allergen_group source created_at } }'} }'
    );
    return data['medicalTerminology'] == null ? null : MedicalTerminologyList.fromJson((data['medicalTerminology'] as Map).cast<String, dynamic>());
  }

  Future<List<MedicalDispense>?> medicalDispenses({String? patientRef, int? limit, int? offset, String? selectionSet}) async {
    final data = await _executor.query('query medicalDispenses(\$patient_ref: String, \$limit: Int, \$offset: Int) { medicalDispenses(patient_ref: \$patient_ref, limit: \$limit, offset: \$offset) ${selectionSet ?? '{ id patient_ref prescription_ref dispensed_by override_reason warnings { type severity medicines ingredients description } items { id dispense_id medicine_id batch_id batch_number expiry_date quantity } created_at }'} }', variables: {
      'patient_ref': ?patientRef,
      'limit': ?limit,
      'offset': ?offset,
    }
    );
    return (data['medicalDispenses'] as List?)?.map((e) => (e == null ? null : MedicalDispense.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<MedicalDispense?> medicalDispense({required int id, String? selectionSet}) async {
    final data = await _executor.query('query medicalDispense(\$id: Int!) { medicalDispense(id: \$id) ${selectionSet ?? '{ id patient_ref prescription_ref dispensed_by override_reason warnings { type severity medicines ingredients description } items { id dispense_id medicine_id batch_id batch_number expiry_date quantity } created_at }'} }', variables: {
      'id': id,
    }
    );
    return data['medicalDispense'] == null ? null : MedicalDispense.fromJson((data['medicalDispense'] as Map).cast<String, dynamic>());
  }

  Future<Object?> medicalCheckSafety({required MedicalSafetyCheckInput input}) async {
    final data = await _executor.query('query medicalCheckSafety(\$input: MedicalSafetyCheckInput!) { medicalCheckSafety(input: \$input) }', variables: {
      'input': input.toJson(),
    }
    );
    return data['medicalCheckSafety'];
  }

  Future<MedicalMedicine?> medicalCreateMedicine({required MedicalMedicineInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation medicalCreateMedicine(\$input: MedicalMedicineInput!) { medicalCreateMedicine(input: \$input) ${selectionSet ?? '{ id code name ingredients allergen_groups dosage_form strength unit requires_prescription reorder_level created_at updated_at }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['medicalCreateMedicine'] == null ? null : MedicalMedicine.fromJson((data['medicalCreateMedicine'] as Map).cast<String, dynamic>());
  }

  Future<MedicalBatch?> medicalReceiveBatch({required int id, required MedicalBatchInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation medicalReceiveBatch(\$id: Int!, \$input: MedicalBatchInput!) { medicalReceiveBatch(id: \$id, input: \$input) ${selectionSet ?? '{ id medicine_id batch_number expiry_date quantity received_quantity unit_cost supplier received_at created_at updated_at }'} }', variables: {
      'id': id,
      'input': input.toJson(),
    }
    );
    return data['medicalReceiveBatch'] == null ? null : MedicalBatch.fromJson((data['medicalReceiveBatch'] as Map).cast<String, dynamic>());
  }

  Future<Object?> medicalAdjustBatch({required int id, required MedicalAdjustInput input}) async {
    final data = await _executor.mutation('mutation medicalAdjustBatch(\$id: Int!, \$input: MedicalAdjustInput!) { medicalAdjustBatch(id: \$id, input: \$input) }', variables: {
      'id': id,
      'input': input.toJson(),
    }
    );
    return data['medicalAdjustBatch'];
  }

  Future<Object?> medicalDisposeExpired() async {
    final data = await _executor.mutation('mutation medicalDisposeExpired { medicalDisposeExpired }'
    );
    return data['medicalDisposeExpired'];
  }

  Future<MedicalInteractionRule?> medicalCreateInteraction({required MedicalInteractionInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation medicalCreateInteraction(\$input: MedicalInteractionInput!) { medicalCreateInteraction(input: \$input) ${selectionSet ?? '{ id ingredient_a ingredient_b severity description source created_at updated_at }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['medicalCreateInteraction'] == null ? null : MedicalInteractionRule.fromJson((data['medicalCreateInteraction'] as Map).cast<String, dynamic>());
  }

  Future<MedicalIngredientAlias?> medicalAddAlias({required String alias, required String canonical, String? selectionSet}) async {
    final data = await _executor.mutation('mutation medicalAddAlias(\$alias: String!, \$canonical: String!) { medicalAddAlias(alias: \$alias, canonical: \$canonical) ${selectionSet ?? '{ id alias canonical source created_at }'} }', variables: {
      'alias': alias,
      'canonical': canonical,
    }
    );
    return data['medicalAddAlias'] == null ? null : MedicalIngredientAlias.fromJson((data['medicalAddAlias'] as Map).cast<String, dynamic>());
  }

  Future<MedicalIngredientAllergenGroup?> medicalAddAllergenGroup({required String ingredient, required String allergenGroup, String? selectionSet}) async {
    final data = await _executor.mutation('mutation medicalAddAllergenGroup(\$ingredient: String!, \$allergen_group: String!) { medicalAddAllergenGroup(ingredient: \$ingredient, allergen_group: \$allergen_group) ${selectionSet ?? '{ id ingredient allergen_group source created_at }'} }', variables: {
      'ingredient': ingredient,
      'allergen_group': allergenGroup,
    }
    );
    return data['medicalAddAllergenGroup'] == null ? null : MedicalIngredientAllergenGroup.fromJson((data['medicalAddAllergenGroup'] as Map).cast<String, dynamic>());
  }

  Future<MedicalDispense?> medicalCreateDispense({required MedicalDispenseInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation medicalCreateDispense(\$input: MedicalDispenseInput!) { medicalCreateDispense(input: \$input) ${selectionSet ?? '{ id patient_ref prescription_ref dispensed_by override_reason warnings { type severity medicines ingredients description } items { id dispense_id medicine_id batch_id batch_number expiry_date quantity } created_at }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['medicalCreateDispense'] == null ? null : MedicalDispense.fromJson((data['medicalCreateDispense'] as Map).cast<String, dynamic>());
  }

}
