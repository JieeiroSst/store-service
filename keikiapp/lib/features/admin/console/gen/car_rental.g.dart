// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

// ignore_for_file: unused_import

import '../../../apps/kit/kit.dart';
import '../../../../gen/models.g.dart';
import '../admin_op.dart';

final carRentalAdminOps = <AdminOp>[
  AdminOp(
    'carRentalCreateVehicle',
    resource: 'vehicle',
    verb: 'create',
    mutation: true,
    fields: [
      KField('input.categoryId', 'Category ID'),
      KField('input.registrationNumber', 'Registration number'),
      KField('input.make', 'Make'),
      KField('input.model', 'Model'),
      KField('input.year', 'Year', kind: FieldKind.number),
      KField('input.color', 'Color'),
      KField('input.mileage', 'Mileage', kind: FieldKind.number),
      KField('input.status', 'Status', kind: FieldKind.choice, options: [('', 'Không chọn'), ('VEHICLE_STATUS_UNSPECIFIED', 'VEHICLE_STATUS_UNSPECIFIED'), ('VEHICLE_STATUS_AVAILABLE', 'VEHICLE_STATUS_AVAILABLE'), ('VEHICLE_STATUS_RENTED', 'VEHICLE_STATUS_RENTED'), ('VEHICLE_STATUS_MAINTENANCE', 'VEHICLE_STATUS_MAINTENANCE'), ('VEHICLE_STATUS_RETIRED', 'VEHICLE_STATUS_RETIRED')]),
      KField('input.hourlyRate', 'Hourly rate', kind: FieldKind.number),
      KField('input.dailyRate', 'Daily rate', kind: FieldKind.number),
      KField('input.locationId', 'Location ID'),
      KField('input.features', 'Features', kind: FieldKind.multiline, hint: 'JSON, ví dụ {"key": "value"}'),
      KField('input.carDetails', 'Car details', kind: FieldKind.multiline, hint: 'JSON, ví dụ {"fuelType":"","transmission":"","seatingCapacity":0,"trunkCapacity":0.0,"airConditioning":false}'),
      KField('input.motorcycleDetails', 'Motorcycle details', kind: FieldKind.multiline, hint: 'JSON, ví dụ {"engineCapacity":0,"motorcycleType":"","helmetIncluded":false}'),
      KField('input.bicycleDetails', 'Bicycle details', kind: FieldKind.multiline, hint: 'JSON, ví dụ {"bicycleType":"","frameSize":"","gearCount":0,"hasBasket":false}'),
    ],
    run: (api, v) async => adminPlain(await api.carRental.carRentalCreateVehicle(input: CarRentalCreateVehicleRequestInput(categoryId: v.str('input.categoryId'), registrationNumber: v.str('input.registrationNumber'), make: v.str('input.make'), model: v.str('input.model'), year: v.integer('input.year'), color: v.str('input.color'), mileage: v.number('input.mileage'), status: v.str('input.status') == null ? null : CarRentalVehicleStatus.fromJson(v.str('input.status').toString()), hourlyRate: v.number('input.hourlyRate'), dailyRate: v.number('input.dailyRate'), locationId: v.str('input.locationId'), features: v.jsonMap('input.features'), carDetails: v.json('input.carDetails') == null ? null : CarRentalCarDetailsInput.fromJson((v.json('input.carDetails') as Map).cast<String, dynamic>()), motorcycleDetails: v.json('input.motorcycleDetails') == null ? null : CarRentalMotorcycleDetailsInput.fromJson((v.json('input.motorcycleDetails') as Map).cast<String, dynamic>()), bicycleDetails: v.json('input.bicycleDetails') == null ? null : CarRentalBicycleDetailsInput.fromJson((v.json('input.bicycleDetails') as Map).cast<String, dynamic>())))),
  ),
  AdminOp(
    'carRentalUpdateVehicle',
    resource: 'vehicle',
    verb: 'update',
    mutation: true,
    fields: [
      KField('input.vehicleId', 'Vehicle ID'),
      KField('input.registrationNumber', 'Registration number'),
      KField('input.mileage', 'Mileage', kind: FieldKind.number),
      KField('input.status', 'Status', kind: FieldKind.choice, options: [('', 'Không chọn'), ('VEHICLE_STATUS_UNSPECIFIED', 'VEHICLE_STATUS_UNSPECIFIED'), ('VEHICLE_STATUS_AVAILABLE', 'VEHICLE_STATUS_AVAILABLE'), ('VEHICLE_STATUS_RENTED', 'VEHICLE_STATUS_RENTED'), ('VEHICLE_STATUS_MAINTENANCE', 'VEHICLE_STATUS_MAINTENANCE'), ('VEHICLE_STATUS_RETIRED', 'VEHICLE_STATUS_RETIRED')]),
      KField('input.hourlyRate', 'Hourly rate', kind: FieldKind.number),
      KField('input.dailyRate', 'Daily rate', kind: FieldKind.number),
      KField('input.locationId', 'Location ID'),
      KField('input.features', 'Features', kind: FieldKind.multiline, hint: 'JSON, ví dụ {"key": "value"}'),
    ],
    run: (api, v) async => adminPlain(await api.carRental.carRentalUpdateVehicle(input: CarRentalUpdateVehicleRequestInput(vehicleId: v.str('input.vehicleId'), registrationNumber: v.str('input.registrationNumber'), mileage: v.number('input.mileage'), status: v.str('input.status') == null ? null : CarRentalVehicleStatus.fromJson(v.str('input.status').toString()), hourlyRate: v.number('input.hourlyRate'), dailyRate: v.number('input.dailyRate'), locationId: v.str('input.locationId'), features: v.jsonMap('input.features')))),
  ),
];
