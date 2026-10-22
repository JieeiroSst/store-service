// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

import '../gateway_executor.dart';
import '../models.g.dart';

class CarRentalApi {
  const CarRentalApi(this._executor);

  final GatewayExecutor _executor;

  Future<CarRentalUserResponse?> carRentalGetUser({required CarRentalGetUserRequestInput input, String? selectionSet}) async {
    final data = await _executor.query('query carRentalGetUser(\$input: CarRentalGetUserRequestInput!) { carRentalGetUser(input: \$input) ${selectionSet ?? '{ userId email firstName lastName phoneNumber address userType drivingLicense createdAt updatedAt }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['carRentalGetUser'] == null ? null : CarRentalUserResponse.fromJson((data['carRentalGetUser'] as Map).cast<String, dynamic>());
  }

  Future<CarRentalVehicleResponse?> carRentalGetVehicle({required CarRentalGetVehicleRequestInput input, String? selectionSet}) async {
    final data = await _executor.public('query carRentalGetVehicle(\$input: CarRentalGetVehicleRequestInput!) { carRentalGetVehicle(input: \$input) ${selectionSet ?? '{ vehicleId categoryId registrationNumber make model year color mileage status hourlyRate dailyRate locationId vehicleType features carDetails { fuelType transmission seatingCapacity trunkCapacity airConditioning } motorcycleDetails { engineCapacity motorcycleType helmetIncluded } bicycleDetails { bicycleType frameSize gearCount hasBasket } createdAt updatedAt }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['carRentalGetVehicle'] == null ? null : CarRentalVehicleResponse.fromJson((data['carRentalGetVehicle'] as Map).cast<String, dynamic>());
  }

  Future<CarRentalListVehiclesResponse?> carRentalListVehicles({required CarRentalListVehiclesRequestInput input, String? selectionSet}) async {
    final data = await _executor.public('query carRentalListVehicles(\$input: CarRentalListVehiclesRequestInput!) { carRentalListVehicles(input: \$input) ${selectionSet ?? '{ vehicles { vehicleId categoryId registrationNumber make model year color mileage status hourlyRate dailyRate locationId vehicleType features carDetails { fuelType transmission seatingCapacity trunkCapacity airConditioning } motorcycleDetails { engineCapacity motorcycleType helmetIncluded } bicycleDetails { bicycleType frameSize gearCount hasBasket } createdAt updatedAt } nextPageToken totalCount }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['carRentalListVehicles'] == null ? null : CarRentalListVehiclesResponse.fromJson((data['carRentalListVehicles'] as Map).cast<String, dynamic>());
  }

  Future<CarRentalListVehiclesResponse?> carRentalSearchAvailableVehicles({required CarRentalSearchVehiclesRequestInput input, String? selectionSet}) async {
    final data = await _executor.public('query carRentalSearchAvailableVehicles(\$input: CarRentalSearchVehiclesRequestInput!) { carRentalSearchAvailableVehicles(input: \$input) ${selectionSet ?? '{ vehicles { vehicleId categoryId registrationNumber make model year color mileage status hourlyRate dailyRate locationId vehicleType features carDetails { fuelType transmission seatingCapacity trunkCapacity airConditioning } motorcycleDetails { engineCapacity motorcycleType helmetIncluded } bicycleDetails { bicycleType frameSize gearCount hasBasket } createdAt updatedAt } nextPageToken totalCount }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['carRentalSearchAvailableVehicles'] == null ? null : CarRentalListVehiclesResponse.fromJson((data['carRentalSearchAvailableVehicles'] as Map).cast<String, dynamic>());
  }

  Future<CarRentalReservationResponse?> carRentalGetReservation({required CarRentalGetReservationRequestInput input, String? selectionSet}) async {
    final data = await _executor.query('query carRentalGetReservation(\$input: CarRentalGetReservationRequestInput!) { carRentalGetReservation(input: \$input) ${selectionSet ?? '{ reservationId userId vehicleId pickupLocationId returnLocationId startTime endTime status createdAt updatedAt vehicle { vehicleId categoryId registrationNumber make model year color mileage status hourlyRate dailyRate locationId vehicleType features carDetails { fuelType transmission seatingCapacity trunkCapacity airConditioning } motorcycleDetails { engineCapacity motorcycleType helmetIncluded } bicycleDetails { bicycleType frameSize gearCount hasBasket } createdAt updatedAt } }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['carRentalGetReservation'] == null ? null : CarRentalReservationResponse.fromJson((data['carRentalGetReservation'] as Map).cast<String, dynamic>());
  }

  Future<CarRentalListReservationsResponse?> carRentalListUserReservations({required CarRentalListUserReservationsRequestInput input, String? selectionSet}) async {
    final data = await _executor.query('query carRentalListUserReservations(\$input: CarRentalListUserReservationsRequestInput!) { carRentalListUserReservations(input: \$input) ${selectionSet ?? '{ reservations { reservationId userId vehicleId pickupLocationId returnLocationId startTime endTime status createdAt updatedAt vehicle { vehicleId categoryId registrationNumber make model year color mileage status hourlyRate dailyRate locationId vehicleType features createdAt updatedAt } } nextPageToken totalCount }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['carRentalListUserReservations'] == null ? null : CarRentalListReservationsResponse.fromJson((data['carRentalListUserReservations'] as Map).cast<String, dynamic>());
  }

  Future<CarRentalRentalResponse?> carRentalGetRental({required CarRentalGetRentalRequestInput input, String? selectionSet}) async {
    final data = await _executor.query('query carRentalGetRental(\$input: CarRentalGetRentalRequestInput!) { carRentalGetRental(input: \$input) ${selectionSet ?? '{ rentalId reservationId vehicleId userId pickupTime actualReturnTime pickupLocationId returnLocationId pickupMileage returnMileage status baseFee additionalFees paymentStatus vehicle { vehicleId categoryId registrationNumber make model year color mileage status hourlyRate dailyRate locationId vehicleType features carDetails { fuelType transmission seatingCapacity trunkCapacity airConditioning } motorcycleDetails { engineCapacity motorcycleType helmetIncluded } bicycleDetails { bicycleType frameSize gearCount hasBasket } createdAt updatedAt } createdAt }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['carRentalGetRental'] == null ? null : CarRentalRentalResponse.fromJson((data['carRentalGetRental'] as Map).cast<String, dynamic>());
  }

  Future<CarRentalListRentalsResponse?> carRentalListUserRentals({required CarRentalListUserRentalsRequestInput input, String? selectionSet}) async {
    final data = await _executor.query('query carRentalListUserRentals(\$input: CarRentalListUserRentalsRequestInput!) { carRentalListUserRentals(input: \$input) ${selectionSet ?? '{ rentals { rentalId reservationId vehicleId userId pickupTime actualReturnTime pickupLocationId returnLocationId pickupMileage returnMileage status baseFee additionalFees paymentStatus vehicle { vehicleId categoryId registrationNumber make model year color mileage status hourlyRate dailyRate locationId vehicleType features createdAt updatedAt } createdAt } nextPageToken totalCount }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['carRentalListUserRentals'] == null ? null : CarRentalListRentalsResponse.fromJson((data['carRentalListUserRentals'] as Map).cast<String, dynamic>());
  }

  Future<CarRentalListReviewsResponse?> carRentalListVehicleReviews({required CarRentalListVehicleReviewsRequestInput input, String? selectionSet}) async {
    final data = await _executor.public('query carRentalListVehicleReviews(\$input: CarRentalListVehicleReviewsRequestInput!) { carRentalListVehicleReviews(input: \$input) ${selectionSet ?? '{ reviews { reviewId rentalId userId vehicleId rating comment createdAt user { userId email firstName lastName phoneNumber address userType drivingLicense createdAt updatedAt } } nextPageToken totalCount }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['carRentalListVehicleReviews'] == null ? null : CarRentalListReviewsResponse.fromJson((data['carRentalListVehicleReviews'] as Map).cast<String, dynamic>());
  }

  Future<CarRentalListLocationsResponse?> carRentalListLocations({required CarRentalListLocationsRequestInput input, String? selectionSet}) async {
    final data = await _executor.public('query carRentalListLocations(\$input: CarRentalListLocationsRequestInput!) { carRentalListLocations(input: \$input) ${selectionSet ?? '{ locations { locationId name address city state country zipCode latitude longitude contactPhone openingHours } nextPageToken totalCount }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['carRentalListLocations'] == null ? null : CarRentalListLocationsResponse.fromJson((data['carRentalListLocations'] as Map).cast<String, dynamic>());
  }

  Future<CarRentalUserResponse?> carRentalRegisterUser({required CarRentalRegisterUserRequestInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation carRentalRegisterUser(\$input: CarRentalRegisterUserRequestInput!) { carRentalRegisterUser(input: \$input) ${selectionSet ?? '{ userId email firstName lastName phoneNumber address userType drivingLicense createdAt updatedAt }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['carRentalRegisterUser'] == null ? null : CarRentalUserResponse.fromJson((data['carRentalRegisterUser'] as Map).cast<String, dynamic>());
  }

  Future<CarRentalUserResponse?> carRentalUpdateUser({required CarRentalUpdateUserRequestInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation carRentalUpdateUser(\$input: CarRentalUpdateUserRequestInput!) { carRentalUpdateUser(input: \$input) ${selectionSet ?? '{ userId email firstName lastName phoneNumber address userType drivingLicense createdAt updatedAt }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['carRentalUpdateUser'] == null ? null : CarRentalUserResponse.fromJson((data['carRentalUpdateUser'] as Map).cast<String, dynamic>());
  }

  Future<CarRentalVehicleResponse?> carRentalCreateVehicle({required CarRentalCreateVehicleRequestInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation carRentalCreateVehicle(\$input: CarRentalCreateVehicleRequestInput!) { carRentalCreateVehicle(input: \$input) ${selectionSet ?? '{ vehicleId categoryId registrationNumber make model year color mileage status hourlyRate dailyRate locationId vehicleType features carDetails { fuelType transmission seatingCapacity trunkCapacity airConditioning } motorcycleDetails { engineCapacity motorcycleType helmetIncluded } bicycleDetails { bicycleType frameSize gearCount hasBasket } createdAt updatedAt }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['carRentalCreateVehicle'] == null ? null : CarRentalVehicleResponse.fromJson((data['carRentalCreateVehicle'] as Map).cast<String, dynamic>());
  }

  Future<CarRentalVehicleResponse?> carRentalUpdateVehicle({required CarRentalUpdateVehicleRequestInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation carRentalUpdateVehicle(\$input: CarRentalUpdateVehicleRequestInput!) { carRentalUpdateVehicle(input: \$input) ${selectionSet ?? '{ vehicleId categoryId registrationNumber make model year color mileage status hourlyRate dailyRate locationId vehicleType features carDetails { fuelType transmission seatingCapacity trunkCapacity airConditioning } motorcycleDetails { engineCapacity motorcycleType helmetIncluded } bicycleDetails { bicycleType frameSize gearCount hasBasket } createdAt updatedAt }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['carRentalUpdateVehicle'] == null ? null : CarRentalVehicleResponse.fromJson((data['carRentalUpdateVehicle'] as Map).cast<String, dynamic>());
  }

  Future<CarRentalReservationResponse?> carRentalCreateReservation({required CarRentalCreateReservationRequestInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation carRentalCreateReservation(\$input: CarRentalCreateReservationRequestInput!) { carRentalCreateReservation(input: \$input) ${selectionSet ?? '{ reservationId userId vehicleId pickupLocationId returnLocationId startTime endTime status createdAt updatedAt vehicle { vehicleId categoryId registrationNumber make model year color mileage status hourlyRate dailyRate locationId vehicleType features carDetails { fuelType transmission seatingCapacity trunkCapacity airConditioning } motorcycleDetails { engineCapacity motorcycleType helmetIncluded } bicycleDetails { bicycleType frameSize gearCount hasBasket } createdAt updatedAt } }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['carRentalCreateReservation'] == null ? null : CarRentalReservationResponse.fromJson((data['carRentalCreateReservation'] as Map).cast<String, dynamic>());
  }

  Future<CarRentalReservationResponse?> carRentalUpdateReservation({required CarRentalUpdateReservationRequestInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation carRentalUpdateReservation(\$input: CarRentalUpdateReservationRequestInput!) { carRentalUpdateReservation(input: \$input) ${selectionSet ?? '{ reservationId userId vehicleId pickupLocationId returnLocationId startTime endTime status createdAt updatedAt vehicle { vehicleId categoryId registrationNumber make model year color mileage status hourlyRate dailyRate locationId vehicleType features carDetails { fuelType transmission seatingCapacity trunkCapacity airConditioning } motorcycleDetails { engineCapacity motorcycleType helmetIncluded } bicycleDetails { bicycleType frameSize gearCount hasBasket } createdAt updatedAt } }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['carRentalUpdateReservation'] == null ? null : CarRentalReservationResponse.fromJson((data['carRentalUpdateReservation'] as Map).cast<String, dynamic>());
  }

  Future<CarRentalReservationResponse?> carRentalCancelReservation({required CarRentalCancelReservationRequestInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation carRentalCancelReservation(\$input: CarRentalCancelReservationRequestInput!) { carRentalCancelReservation(input: \$input) ${selectionSet ?? '{ reservationId userId vehicleId pickupLocationId returnLocationId startTime endTime status createdAt updatedAt vehicle { vehicleId categoryId registrationNumber make model year color mileage status hourlyRate dailyRate locationId vehicleType features carDetails { fuelType transmission seatingCapacity trunkCapacity airConditioning } motorcycleDetails { engineCapacity motorcycleType helmetIncluded } bicycleDetails { bicycleType frameSize gearCount hasBasket } createdAt updatedAt } }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['carRentalCancelReservation'] == null ? null : CarRentalReservationResponse.fromJson((data['carRentalCancelReservation'] as Map).cast<String, dynamic>());
  }

  Future<CarRentalRentalResponse?> carRentalStartRental({required CarRentalStartRentalRequestInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation carRentalStartRental(\$input: CarRentalStartRentalRequestInput!) { carRentalStartRental(input: \$input) ${selectionSet ?? '{ rentalId reservationId vehicleId userId pickupTime actualReturnTime pickupLocationId returnLocationId pickupMileage returnMileage status baseFee additionalFees paymentStatus vehicle { vehicleId categoryId registrationNumber make model year color mileage status hourlyRate dailyRate locationId vehicleType features carDetails { fuelType transmission seatingCapacity trunkCapacity airConditioning } motorcycleDetails { engineCapacity motorcycleType helmetIncluded } bicycleDetails { bicycleType frameSize gearCount hasBasket } createdAt updatedAt } createdAt }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['carRentalStartRental'] == null ? null : CarRentalRentalResponse.fromJson((data['carRentalStartRental'] as Map).cast<String, dynamic>());
  }

  Future<CarRentalRentalResponse?> carRentalCompleteRental({required CarRentalCompleteRentalRequestInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation carRentalCompleteRental(\$input: CarRentalCompleteRentalRequestInput!) { carRentalCompleteRental(input: \$input) ${selectionSet ?? '{ rentalId reservationId vehicleId userId pickupTime actualReturnTime pickupLocationId returnLocationId pickupMileage returnMileage status baseFee additionalFees paymentStatus vehicle { vehicleId categoryId registrationNumber make model year color mileage status hourlyRate dailyRate locationId vehicleType features carDetails { fuelType transmission seatingCapacity trunkCapacity airConditioning } motorcycleDetails { engineCapacity motorcycleType helmetIncluded } bicycleDetails { bicycleType frameSize gearCount hasBasket } createdAt updatedAt } createdAt }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['carRentalCompleteRental'] == null ? null : CarRentalRentalResponse.fromJson((data['carRentalCompleteRental'] as Map).cast<String, dynamic>());
  }

  Future<CarRentalPaymentResponse?> carRentalProcessPayment({required CarRentalProcessPaymentRequestInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation carRentalProcessPayment(\$input: CarRentalProcessPaymentRequestInput!) { carRentalProcessPayment(input: \$input) ${selectionSet ?? '{ paymentId rentalId userId amount paymentMethod transactionId paymentStatus paymentDate }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['carRentalProcessPayment'] == null ? null : CarRentalPaymentResponse.fromJson((data['carRentalProcessPayment'] as Map).cast<String, dynamic>());
  }

  Future<CarRentalReviewResponse?> carRentalSubmitReview({required CarRentalSubmitReviewRequestInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation carRentalSubmitReview(\$input: CarRentalSubmitReviewRequestInput!) { carRentalSubmitReview(input: \$input) ${selectionSet ?? '{ reviewId rentalId userId vehicleId rating comment createdAt user { userId email firstName lastName phoneNumber address userType drivingLicense createdAt updatedAt } }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['carRentalSubmitReview'] == null ? null : CarRentalReviewResponse.fromJson((data['carRentalSubmitReview'] as Map).cast<String, dynamic>());
  }

}
