import 'fixtures.dart';

final _home = {
  'id': 1,
  'name': 'Mây Lang Thang Homestay',
  'description': 'Nhà gỗ giữa đồi thông, view săn mây buổi sáng, có bếp riêng và lò sưởi.',
  'type_name': 'Nhà nguyên căn',
  'address': '12 Đường Hoa Hồng, Phường 4, Đà Lạt',
  'guests': 6,
  'bedrooms': 3,
  'bathrooms': 2,
  'rating': 4.8,
  'review_count': 126,
  'accepts_wallet_payment': true,
  'rates': [
    {'model': 'nightly', 'price': '1200000', 'currency': 'VND', 'active': true},
    {'model': 'monthly', 'price': '18000000', 'currency': 'VND', 'active': true},
  ],
};

final _hotel = {
  'id': 1,
  'name': 'Keiki Riverside Hotel',
  'description': 'Khách sạn 4 sao bên sông Sài Gòn, hồ bơi vô cực và nhà hàng trên sân thượng.',
  'stars': 4,
  'city': 'TP. Hồ Chí Minh',
  'address': '8 Tôn Đức Thắng, Quận 1',
  'rating': 8.9,
  'review_count': 532,
  'free_cancel_hours': 24,
  'check_in_time': '14:00',
  'check_out_time': '12:00',
  'currency': 'VND',
  'amenities': ['Hồ bơi', 'Wifi', 'Gym', 'Spa'],
};

final batch2 = <String, Object?>{
  'rentHomestays': {
    'items': [
      _home,
      {..._home, 'id': 2, 'name': 'Hội An Ancient House', 'address': 'Phố cổ Hội An', 'rating': 4.6, 'review_count': 88},
    ],
  },
  'rentHomestay': _home,
  'rentReviews': [
    {'rating': 5, 'comment': 'View tuyệt đẹp, chủ nhà rất dễ thương!', 'created_at': '2026-10-02T00:00:00Z'},
  ],
  'resvHotels': {
    'items': [
      _hotel,
      {..._hotel, 'id': 2, 'name': 'Sea Breeze Nha Trang', 'city': 'Nha Trang', 'stars': 5, 'rating': 9.2},
      {..._hotel, 'id': 3, 'name': 'Old Quarter Boutique', 'city': 'Hà Nội', 'stars': 3, 'free_cancel_hours': 0},
    ],
  },
  'resvHotel': _hotel,
  'resvQuotes': [
    {'room_type': {'id': 1, 'name': 'Deluxe King', 'capacity': 2, 'bed': '1 giường King', 'size_m2': 32, 'view': 'Hướng sông'}, 'nights': 2, 'rooms_available': 4, 'total': '3800000', 'bookable': true},
    {'room_type': {'id': 2, 'name': 'Family Suite', 'capacity': 4, 'bed': '2 giường Queen', 'size_m2': 55, 'view': 'Hướng thành phố'}, 'nights': 2, 'rooms_available': 1, 'total': '6400000', 'bookable': true},
  ],
  'carRentalListVehicles': {
    'vehicles': [
      {'vehicleId': 'v1', 'make': 'VinFast', 'model': 'VF 8', 'year': 2025, 'color': 'Trắng', 'status': 'VEHICLE_STATUS_AVAILABLE', 'hourlyRate': 150000, 'dailyRate': 1300000, 'vehicleType': 'VEHICLE_TYPE_CAR', 'carDetails': {'fuelType': 'Điện', 'transmission': 'Tự động', 'seatingCapacity': 5}},
      {'vehicleId': 'v2', 'make': 'Toyota', 'model': 'Innova', 'year': 2023, 'color': 'Bạc', 'status': 'VEHICLE_STATUS_RENTED', 'hourlyRate': 120000, 'dailyRate': 950000, 'vehicleType': 'VEHICLE_TYPE_CAR', 'carDetails': {'fuelType': 'Xăng', 'transmission': 'Số sàn', 'seatingCapacity': 7}},
      {'vehicleId': 'v3', 'make': 'Honda', 'model': 'SH 160i', 'year': 2024, 'color': 'Đen', 'status': 'VEHICLE_STATUS_AVAILABLE', 'hourlyRate': 40000, 'dailyRate': 250000, 'vehicleType': 'VEHICLE_TYPE_MOTORCYCLE'},
    ],
  },
  'bookingMiniAvailableRooms': [
    {'room_no': 101, 'room_type': 'Standard', 'bed_type': 'đôi', 'bed_count': 1, 'amenities': 'Wifi, điều hoà, TV', 'room_type_details': {'description': 'Phòng tiêu chuẩn', 'capacity': 2}},
    {'room_no': 205, 'room_type': 'Family', 'bed_type': 'đơn', 'bed_count': 3, 'amenities': 'Wifi, bồn tắm, ban công', 'room_type_details': {'description': 'Phòng gia đình', 'capacity': 4}},
  ],
  'parkingAvailableSpots': [
    {'type': 'car', 'available': 42, 'total': 120},
    {'type': 'motorbike', 'available': 310, 'total': 400},
  ],
  'parkingRates': [
    {'VehicleType': 'car', 'HourlyRate': 30000, 'DailyMaxRate': 200000},
    {'VehicleType': 'motorbike', 'HourlyRate': 5000, 'DailyMaxRate': 30000},
  ],
};

const batch2Shots = <Shot>[
  Shot('rent_house', '/services/rent_house'),
  Shot('rent_house_detail', '/services/rent_house', tap: 'Mây Lang Thang Homestay'),
  Shot('reservation', '/services/reservation'),
  Shot('reservation_hotel', '/services/reservation', tap: 'Keiki Riverside Hotel'),
  Shot('car_rental', '/services/car_rental'),
  Shot('booking_mini', '/services/booking_mini'),
  Shot('parking_lot', '/services/parking_lot'),
];
