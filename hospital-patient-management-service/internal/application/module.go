package application

import (
	"github.com/JIeeroSst/hospital-patient-management-service/internal/domain/port"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(
		func() port.Clock { return systemClock{} },
		NewPatientNotifier,
		NewDepartmentService,
		NewStaffService,
		NewPatientService,
		NewMedicalRecordService,
		NewAppointmentService,
		NewPrescriptionService,
		NewLabResultService,
		NewBillingService,
		NewIdentityService,
		NewDocumentService,
	),
)
