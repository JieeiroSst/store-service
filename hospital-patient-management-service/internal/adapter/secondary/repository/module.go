package repository

import "go.uber.org/fx"

var Module = fx.Options(
	fx.Provide(
		NewDepartmentRepo,
		NewStaffRepo,
		NewPatientRepo,
		NewMedicalRecordRepo,
		NewPrescriptionRepo,
		NewLabResultRepo,
		NewAppointmentRepo,
		NewBillingRepo,
		NewBillingAccountRepo,
	),
)
