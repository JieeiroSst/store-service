// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

// ignore_for_file: unused_import

import '../../../apps/kit/kit.dart';
import '../../../../gen/models.g.dart';
import '../admin_op.dart';

final medicalAdminOps = <AdminOp>[
  AdminOp(
    'medicalMedicines',
    resource: 'medicine',
    verb: '',
    fields: [
      KField('q', 'Q'),
      KField('limit', 'Limit', kind: FieldKind.number),
      KField('offset', 'Offset', kind: FieldKind.number),
    ],
    run: (api, v) async => adminPlain(await api.medical.medicalMedicines(q: v.str('q'), limit: v.integer('limit'), offset: v.integer('offset'))),
  ),
  AdminOp(
    'medicalMedicine',
    resource: 'medicine',
    verb: '',
    fields: [
      KField('id', 'ID', kind: FieldKind.number, required: true),
    ],
    run: (api, v) async => adminPlain(await api.medical.medicalMedicine(id: v.need(v.integer('id'), 'id'))),
  ),
  AdminOp(
    'medicalStock',
    resource: 'general',
    verb: '',
    fields: [
      KField('id', 'ID', kind: FieldKind.number, required: true),
    ],
    run: (api, v) async => adminPlain(await api.medical.medicalStock(id: v.need(v.integer('id'), 'id'))),
  ),
  AdminOp(
    'medicalMovements',
    resource: 'general',
    verb: '',
    fields: [
      KField('id', 'ID', kind: FieldKind.number, required: true),
      KField('limit', 'Limit', kind: FieldKind.number),
      KField('offset', 'Offset', kind: FieldKind.number),
    ],
    run: (api, v) async => adminPlain(await api.medical.medicalMovements(id: v.need(v.integer('id'), 'id'), limit: v.integer('limit'), offset: v.integer('offset'))),
  ),
  AdminOp(
    'medicalExpiring',
    resource: 'general',
    verb: '',
    fields: [
      KField('days', 'Days', kind: FieldKind.number),
    ],
    run: (api, v) async => adminPlain(await api.medical.medicalExpiring(days: v.integer('days'))),
  ),
  AdminOp(
    'medicalLowStock',
    resource: 'general',
    verb: '',
    run: (api, v) async => adminPlain(await api.medical.medicalLowStock()),
  ),
  AdminOp(
    'medicalInteractions',
    resource: 'interaction',
    verb: '',
    fields: [
      KField('limit', 'Limit', kind: FieldKind.number),
      KField('offset', 'Offset', kind: FieldKind.number),
    ],
    run: (api, v) async => adminPlain(await api.medical.medicalInteractions(limit: v.integer('limit'), offset: v.integer('offset'))),
  ),
  AdminOp(
    'medicalTerminology',
    resource: 'general',
    verb: '',
    run: (api, v) async => adminPlain(await api.medical.medicalTerminology()),
  ),
  AdminOp(
    'medicalDispenses',
    resource: 'dispense',
    verb: '',
    fields: [
      KField('patient_ref', 'Patient ref'),
      KField('limit', 'Limit', kind: FieldKind.number),
      KField('offset', 'Offset', kind: FieldKind.number),
    ],
    run: (api, v) async => adminPlain(await api.medical.medicalDispenses(patientRef: v.str('patient_ref'), limit: v.integer('limit'), offset: v.integer('offset'))),
  ),
  AdminOp(
    'medicalDispense',
    resource: 'dispense',
    verb: '',
    fields: [
      KField('id', 'ID', kind: FieldKind.number, required: true),
    ],
    run: (api, v) async => adminPlain(await api.medical.medicalDispense(id: v.need(v.integer('id'), 'id'))),
  ),
  AdminOp(
    'medicalCheckSafety',
    resource: 'general',
    verb: 'check',
    fields: [
      KField('input.medicine_ids', 'Medicine ids', hint: 'Nhiều giá trị, cách nhau bởi dấu phẩy'),
      KField('input.current_medicine_ids', 'Current medicine ids', hint: 'Nhiều giá trị, cách nhau bởi dấu phẩy'),
      KField('input.allergies', 'Allergies', hint: 'Nhiều giá trị, cách nhau bởi dấu phẩy'),
    ],
    run: (api, v) async => adminPlain(await api.medical.medicalCheckSafety(input: MedicalSafetyCheckInput(medicineIds: v.intList('input.medicine_ids'), currentMedicineIds: v.intList('input.current_medicine_ids'), allergies: v.strs('input.allergies')))),
  ),
  AdminOp(
    'medicalCreateMedicine',
    resource: 'medicine',
    verb: 'create',
    mutation: true,
    fields: [
      KField('input.code', 'Code'),
      KField('input.name', 'Name'),
      KField('input.ingredients', 'Ingredients', hint: 'Nhiều giá trị, cách nhau bởi dấu phẩy'),
      KField('input.allergen_groups', 'Allergen groups', hint: 'Nhiều giá trị, cách nhau bởi dấu phẩy'),
      KField('input.dosage_form', 'Dosage form'),
      KField('input.strength', 'Strength'),
      KField('input.unit', 'Unit'),
      KField('input.requires_prescription', 'Requires prescription', kind: FieldKind.toggle),
      KField('input.reorder_level', 'Reorder level', kind: FieldKind.number),
    ],
    run: (api, v) async => adminPlain(await api.medical.medicalCreateMedicine(input: MedicalMedicineInput(code: v.str('input.code'), name_: v.str('input.name'), ingredients: v.strs('input.ingredients'), allergenGroups: v.strs('input.allergen_groups'), dosageForm: v.str('input.dosage_form'), strength: v.str('input.strength'), unit: v.str('input.unit'), requiresPrescription: v.optFlag('input.requires_prescription'), reorderLevel: v.integer('input.reorder_level')))),
  ),
  AdminOp(
    'medicalReceiveBatch',
    resource: 'general',
    verb: '',
    mutation: true,
    fields: [
      KField('id', 'ID', kind: FieldKind.number, required: true),
      KField('input.batch_number', 'Batch number'),
      KField('input.expiry_date', 'Expiry date', hint: 'ISO 8601, ví dụ 2026-10-22T09:00:00Z'),
      KField('input.quantity', 'Quantity', kind: FieldKind.number),
      KField('input.unit_cost', 'Unit cost', kind: FieldKind.number),
      KField('input.supplier', 'Supplier'),
    ],
    run: (api, v) async => adminPlain(await api.medical.medicalReceiveBatch(id: v.need(v.integer('id'), 'id'), input: MedicalBatchInput(batchNumber: v.str('input.batch_number'), expiryDate: v.str('input.expiry_date'), quantity: v.integer('input.quantity'), unitCost: v.integer('input.unit_cost'), supplier: v.str('input.supplier')))),
  ),
  AdminOp(
    'medicalAdjustBatch',
    resource: 'general',
    verb: '',
    mutation: true,
    fields: [
      KField('id', 'ID', kind: FieldKind.number, required: true),
      KField('input.quantity', 'Quantity', kind: FieldKind.number),
      KField('input.reason', 'Reason', kind: FieldKind.multiline),
    ],
    run: (api, v) async => adminPlain(await api.medical.medicalAdjustBatch(id: v.need(v.integer('id'), 'id'), input: MedicalAdjustInput(quantity: v.integer('input.quantity'), reason: v.str('input.reason')))),
  ),
  AdminOp(
    'medicalDisposeExpired',
    resource: 'general',
    verb: '',
    mutation: true,
    run: (api, v) async => adminPlain(await api.medical.medicalDisposeExpired()),
  ),
  AdminOp(
    'medicalCreateInteraction',
    resource: 'interaction',
    verb: 'create',
    mutation: true,
    fields: [
      KField('input.ingredient_a', 'Ingredient a'),
      KField('input.ingredient_b', 'Ingredient b'),
      KField('input.severity', 'Severity'),
      KField('input.description', 'Description', kind: FieldKind.multiline),
    ],
    run: (api, v) async => adminPlain(await api.medical.medicalCreateInteraction(input: MedicalInteractionInput(ingredientA: v.str('input.ingredient_a'), ingredientB: v.str('input.ingredient_b'), severity: v.str('input.severity'), description: v.str('input.description')))),
  ),
  AdminOp(
    'medicalAddAlias',
    resource: 'general',
    verb: 'add',
    mutation: true,
    fields: [
      KField('alias', 'Alias', required: true),
      KField('canonical', 'Canonical', required: true),
    ],
    run: (api, v) async => adminPlain(await api.medical.medicalAddAlias(alias: v.need(v.str('alias'), 'alias'), canonical: v.need(v.str('canonical'), 'canonical'))),
  ),
  AdminOp(
    'medicalAddAllergenGroup',
    resource: 'general',
    verb: 'add',
    mutation: true,
    fields: [
      KField('ingredient', 'Ingredient', required: true),
      KField('allergen_group', 'Allergen group', required: true),
    ],
    run: (api, v) async => adminPlain(await api.medical.medicalAddAllergenGroup(ingredient: v.need(v.str('ingredient'), 'ingredient'), allergenGroup: v.need(v.str('allergen_group'), 'allergen_group'))),
  ),
  AdminOp(
    'medicalCreateDispense',
    resource: 'dispense',
    verb: 'create',
    mutation: true,
    fields: [
      KField('input.patient_ref', 'Patient ref'),
      KField('input.prescription_ref', 'Prescription ref'),
      KField('input.items', 'Items', kind: FieldKind.multiline, hint: 'JSON, ví dụ [{"medicine_id":0,"quantity":0}]'),
      KField('input.current_medicine_ids', 'Current medicine ids', hint: 'Nhiều giá trị, cách nhau bởi dấu phẩy'),
      KField('input.allergies', 'Allergies', hint: 'Nhiều giá trị, cách nhau bởi dấu phẩy'),
      KField('input.override_reason', 'Override reason', kind: FieldKind.multiline),
    ],
    run: (api, v) async => adminPlain(await api.medical.medicalCreateDispense(input: MedicalDispenseInput(patientRef: v.str('input.patient_ref'), prescriptionRef: v.str('input.prescription_ref'), items: (v.json('input.items') as List?)?.map((e) => (e == null ? null : MedicalDispenseItemInput.fromJson((e as Map).cast<String, dynamic>()))!).toList(), currentMedicineIds: v.intList('input.current_medicine_ids'), allergies: v.strs('input.allergies'), overrideReason: v.str('input.override_reason')))),
  ),
];
