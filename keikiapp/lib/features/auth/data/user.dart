import 'package:equatable/equatable.dart';

class User extends Equatable {
  const User({
    required this.id,
    required this.username,
    required this.email,
    required this.name,
    this.phone = '',
    this.address = '',
    this.sex = '',
    this.roles = const [],
  });

  final int id;
  final String username;
  final String email;
  final String name;
  final String phone;
  final String address;
  final String sex;
  final List<String> roles;

  String get displayName => name.isNotEmpty ? name : username;

  factory User.fromJson(Map<String, dynamic> json) => User(
        id: (json['id'] as num).toInt(),
        username: json['username'] as String? ?? '',
        email: json['email'] as String? ?? '',
        name: json['name'] as String? ?? '',
        phone: json['phone'] as String? ?? '',
        address: json['address'] as String? ?? '',
        sex: json['sex'] as String? ?? '',
        roles: [for (final r in (json['roles'] as List? ?? const [])) (r as Map)['name'] as String],
      );

  @override
  List<Object?> get props => [id, username, email, name, phone, address, sex, roles];
}
