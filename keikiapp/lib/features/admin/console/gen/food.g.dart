// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

// ignore_for_file: unused_import

import '../../../apps/kit/kit.dart';
import '../../../../gen/models.g.dart';
import '../admin_op.dart';

final foodAdminOps = <AdminOp>[
  AdminOp(
    'foodCreateFood',
    resource: 'food',
    verb: 'create',
    mutation: true,
    fields: [
      KField('input.name', 'Name'),
      KField('input.description', 'Description', kind: FieldKind.multiline),
      KField('input.price', 'Price', kind: FieldKind.number),
      KField('input.category', 'Category'),
      KField('input.restaurant_id', 'Restaurant ID', kind: FieldKind.number),
      KField('input.is_active', 'Is active', kind: FieldKind.toggle),
    ],
    run: (api, v) async => adminPlain(await api.food.foodCreateFood(input: FoodFoodInput(name_: v.str('input.name'), description: v.str('input.description'), price: v.integer('input.price'), category: v.str('input.category'), restaurantId: v.integer('input.restaurant_id'), isActive: v.optFlag('input.is_active')))),
  ),
  AdminOp(
    'foodUpdateFood',
    resource: 'food',
    verb: 'update',
    mutation: true,
    fields: [
      KField('id', 'ID', kind: FieldKind.number, required: true),
      KField('input.name', 'Name'),
      KField('input.description', 'Description', kind: FieldKind.multiline),
      KField('input.price', 'Price', kind: FieldKind.number),
      KField('input.category', 'Category'),
      KField('input.restaurant_id', 'Restaurant ID', kind: FieldKind.number),
      KField('input.is_active', 'Is active', kind: FieldKind.toggle),
    ],
    run: (api, v) async => adminPlain(await api.food.foodUpdateFood(id: v.need(v.integer('id'), 'id'), input: FoodFoodInput(name_: v.str('input.name'), description: v.str('input.description'), price: v.integer('input.price'), category: v.str('input.category'), restaurantId: v.integer('input.restaurant_id'), isActive: v.optFlag('input.is_active')))),
  ),
  AdminOp(
    'foodDeleteFood',
    resource: 'food',
    verb: 'delete',
    mutation: true,
    destructive: true,
    fields: [
      KField('id', 'ID', kind: FieldKind.number, required: true),
    ],
    run: (api, v) async => adminPlain(await api.food.foodDeleteFood(id: v.need(v.integer('id'), 'id'))),
  ),
  AdminOp(
    'foodDeleteAllFoods',
    resource: 'all food',
    verb: 'delete',
    mutation: true,
    destructive: true,
    run: (api, v) async => adminPlain(await api.food.foodDeleteAllFoods()),
  ),
  AdminOp(
    'foodCreateAllergen',
    resource: 'allergen',
    verb: 'create',
    mutation: true,
    fields: [
      KField('name', 'Name', required: true),
    ],
    run: (api, v) async => adminPlain(await api.food.foodCreateAllergen(name_: v.need(v.str('name'), 'name'))),
  ),
  AdminOp(
    'foodCreateIngredient',
    resource: 'ingredient',
    verb: 'create',
    mutation: true,
    fields: [
      KField('name', 'Name', required: true),
      KField('unit', 'Unit'),
    ],
    run: (api, v) async => adminPlain(await api.food.foodCreateIngredient(name_: v.need(v.str('name'), 'name'), unit: v.str('unit'))),
  ),
  AdminOp(
    'foodCreateRestaurant',
    resource: 'restaurant',
    verb: 'create',
    mutation: true,
    fields: [
      KField('input.name', 'Name'),
      KField('input.description', 'Description', kind: FieldKind.multiline),
      KField('input.address', 'Address'),
      KField('input.image_url', 'Image URL'),
      KField('input.is_active', 'Is active', kind: FieldKind.toggle),
      KField('input.commission_rate', 'Commission rate', kind: FieldKind.number),
    ],
    run: (api, v) async => adminPlain(await api.food.foodCreateRestaurant(input: FoodRestaurantInput(name_: v.str('input.name'), description: v.str('input.description'), address: v.str('input.address'), imageUrl: v.str('input.image_url'), isActive: v.optFlag('input.is_active'), commissionRate: v.number('input.commission_rate')))),
  ),
  AdminOp(
    'foodUpdateRestaurant',
    resource: 'restaurant',
    verb: 'update',
    mutation: true,
    fields: [
      KField('id', 'ID', kind: FieldKind.number, required: true),
      KField('input.name', 'Name'),
      KField('input.description', 'Description', kind: FieldKind.multiline),
      KField('input.address', 'Address'),
      KField('input.image_url', 'Image URL'),
      KField('input.is_active', 'Is active', kind: FieldKind.toggle),
      KField('input.commission_rate', 'Commission rate', kind: FieldKind.number),
    ],
    run: (api, v) async => adminPlain(await api.food.foodUpdateRestaurant(id: v.need(v.integer('id'), 'id'), input: FoodRestaurantInput(name_: v.str('input.name'), description: v.str('input.description'), address: v.str('input.address'), imageUrl: v.str('input.image_url'), isActive: v.optFlag('input.is_active'), commissionRate: v.number('input.commission_rate')))),
  ),
];
