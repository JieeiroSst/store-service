package application

import (
	"context"
	"time"

	"github.com/JIeeroSst/hospital-patient-management-service/internal/domain/port"
	"github.com/sirupsen/logrus"
)

const notifyTimeout = 5 * time.Second

// PatientNotifier tells a patient about something that happened. Delivery is
// best-effort: a failure is logged and never fails the business operation
// that triggered it. The zero (nil) notifier does nothing.
type PatientNotifier struct {
	patients port.PatientRepository
	out      port.Notifier
}

func NewPatientNotifier(patients port.PatientRepository, out port.Notifier) *PatientNotifier {
	return &PatientNotifier{patients: patients, out: out}
}

func (n *PatientNotifier) Notify(ctx context.Context, patientID int32, title, message string) {
	if n == nil || n.out == nil {
		return
	}
	// The request may finish (and cancel its context) before we do.
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), notifyTimeout)
	defer cancel()

	log := logrus.WithField("patient_id", patientID)
	p, err := n.patients.Get(ctx, patientID)
	if err != nil {
		log.WithError(err).Warn("notification skipped: patient lookup failed")
		return
	}
	err = n.out.Notify(ctx, port.Notification{PatientID: p.ID, Email: p.Email, Title: title, Message: message})
	if err != nil {
		log.WithError(err).Warn("notification failed")
	}
}
