import 'package:equatable/equatable.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

import '../../../core/utils/load_status.dart';
import '../data/notification_repository.dart';

class NotificationSettingsState extends Equatable {
  const NotificationSettingsState({
    this.status = LoadStatus.initial,
    this.devices = const [],
    this.contact = const NotifContact(),
    this.error,
  });

  final LoadStatus status;
  final List<NotifDevice> devices;
  final NotifContact contact;
  final String? error;

  @override
  List<Object?> get props => [status, devices, contact, error];
}

class NotificationSettingsCubit extends Cubit<NotificationSettingsState> {
  NotificationSettingsCubit(this._repository) : super(const NotificationSettingsState());

  final NotificationRepository _repository;

  Future<void> load() async {
    emit(NotificationSettingsState(status: LoadStatus.loading, devices: state.devices, contact: state.contact));
    try {
      final results = await Future.wait([_repository.myDevices(), _repository.myContact()]);
      emit(NotificationSettingsState(
        status: LoadStatus.success,
        devices: results[0] as List<NotifDevice>,
        contact: results[1] as NotifContact,
      ));
    } catch (e) {
      emit(NotificationSettingsState(status: LoadStatus.failure, error: e.toString()));
    }
  }

  Future<void> updateContact(String email, String phone) async {
    await _repository.updateContact(email: email, phone: phone);
    await load();
  }
}
