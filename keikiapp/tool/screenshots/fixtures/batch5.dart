import 'fixtures.dart';

final _foods = [
  {'id': 1, 'name': 'Phở bò tái', 'price': 65000, 'status': 1, 'category': {'id': 1, 'name': 'Món nước'}},
  {'id': 2, 'name': 'Bún chả Hà Nội', 'price': 55000, 'status': 1, 'category': {'id': 1, 'name': 'Món nước'}},
  {'id': 3, 'name': 'Cơm tấm sườn bì', 'price': 59000, 'status': 1, 'category': {'id': 2, 'name': 'Cơm'}},
  {'id': 4, 'name': 'Trà đào cam sả', 'price': 35000, 'status': 0, 'category': {'id': 3, 'name': 'Đồ uống'}},
];

final batch5 = <String, Object?>{
  'myEkycStatus': {
    'identity': {'document_number': '079203001234', 'surname': 'LƯU', 'given_names': 'MẠNH QUÂN', 'nationality': 'VNM', 'date_of_birth': '2003-04-12', 'sex': 'M', 'date_of_expiry': '2043-04-12', 'nfc_verified': false},
    'face': null,
    'verification': null,
  },
  'ollamaPyTags': {'models': [{'name': 'llama3.1:8b'}, {'name': 'qwen2.5:7b'}, {'name': 'gemma2:9b'}]},
  'gymClasses': [
    {'id': 'c1', 'name': 'Yoga buổi sáng', 'day': 'Thứ 2, 4, 6', 'time': '06:30', 'coach': 'Mai', 'description': 'Giãn cơ, hít thở và cân bằng'},
    {'id': 'c2', 'name': 'HIIT đốt mỡ', 'day': 'Thứ 3, 5', 'time': '18:00', 'coach': 'Hùng', 'description': '45 phút cường độ cao'},
  ],
  'uploadFiles': {
    'files': [
      {'id': 'f1', 'file_name': 'hop-dong-thue-nha.pdf', 'content_type': 'application/pdf', 'size': 482133, 'created_at': '2026-10-20T00:00:00Z'},
      {'id': 'f2', 'file_name': 'anh-cccd-mat-truoc.jpg', 'content_type': 'image/jpeg', 'size': 1820133, 'created_at': '2026-10-18T00:00:00Z'},
    ],
  },
  'geoProvinces': [
    {'code': '01', 'name': 'Hà Nội', 'name_full': 'Thành phố Hà Nội', 'region': 'Đồng bằng sông Hồng', 'area_km2': 3359.8},
    {'code': '79', 'name': 'Hồ Chí Minh', 'name_full': 'Thành phố Hồ Chí Minh', 'region': 'Đông Nam Bộ', 'area_km2': 2095.4},
    {'code': '48', 'name': 'Đà Nẵng', 'name_full': 'Thành phố Đà Nẵng', 'region': 'Duyên hải Nam Trung Bộ', 'area_km2': 1284.9},
  ],
  'shopifyProducts': {
    'items': [
      {'id': 1, 'title': 'Áo hoodie Keiki', 'vendor': 'Keiki Store', 'variants': [{'price': '450000'}]},
      {'id': 2, 'title': 'Bình giữ nhiệt Keiki', 'vendor': 'Keiki Store', 'variants': [{'price': '290000'}]},
    ],
  },
  'rKitchenCategories': [{'id': 1, 'name': 'Món nước'}, {'id': 2, 'name': 'Cơm'}, {'id': 3, 'name': 'Đồ uống'}],
  'rKitchenFoods': {'rows': _foods},
  'rKitchenKitchens': {
    'rows': [
      {'id': 1, 'name': 'Bếp Việt – Quận 1', 'status': 'open', 'foods': _foods},
    ],
  },
  'vendMachines': {
    'items': [
      {'id': 'm1', 'location': 'Sảnh A – Toà Keiki Tower', 'model': 'Fuji VM-30', 'status': 'active'},
      {'id': 'm2', 'location': 'Ga Metro Bến Thành', 'model': 'Fuji VM-40', 'status': 'maintenance'},
    ],
  },
  'vendInventory': {
    'items': [
      for (final (i, n) in ['Coca-Cola', 'Pepsi', 'Aquafina', 'Trà xanh 0°', 'Red Bull', 'Sting'].indexed)
        {'slot': 'A${i + 1}', 'quantity': i == 3 ? 0 : 8 - i, 'product': {'name': n, 'price_cents': 1200000 + i * 100000}},
    ],
  },
};

const batch5Shots = <Shot>[
  Shot('ekyc', '/services/ekyc'),
  Shot('ollama_py', '/services/ollama_py'),
  Shot('identified', '/services/identified'),
  Shot('upload', '/services/upload'),
  Shot('geoservice', '/services/geoservice'),
  Shot('shopify', '/services/shopify'),
  Shot('restaurant_consumer_menu', '/services/restaurant_consumer', tap: 'Bếp Việt – Quận 1'),
  Shot('vending_machine', '/services/vending', tap: 'Sảnh A – Toà Keiki Tower'),
];
