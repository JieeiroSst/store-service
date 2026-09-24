package auth

import pb "github.com/JIeeiroSst/lib-gateway/car-rental-servcie/gateway/car-rental-servcie"

// MethodLevels lists the access level of every RPC that is not simply "any logged-in user".
var MethodLevels = map[string]Level{
	pb.VehicleRentalService_RegisterUser_FullMethodName:            Public, // staff/admin creation is gated on an admin token in the handler
	pb.VehicleRentalService_ListLocations_FullMethodName:           Public,
	pb.VehicleRentalService_ListVehicles_FullMethodName:            Public,
	pb.VehicleRentalService_GetVehicle_FullMethodName:              Public,
	pb.VehicleRentalService_SearchAvailableVehicles_FullMethodName: Public,
	pb.VehicleRentalService_ListVehicleReviews_FullMethodName:      Public,
	pb.VehicleRentalService_CreateVehicle_FullMethodName:           Staff,
	pb.VehicleRentalService_UpdateVehicle_FullMethodName:           Staff,
	pb.VehicleRentalService_StartRental_FullMethodName:             Staff,
	pb.VehicleRentalService_CompleteRental_FullMethodName:          Staff,
}
