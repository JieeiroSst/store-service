package domain

import (
	"fmt"
	"strings"
	"time"
)

type MachineStatus string

const (
	MachineActive       MachineStatus = "active"
	MachineInactive     MachineStatus = "inactive"
	MachineMaintenance  MachineStatus = "maintenance"
	MachineOutOfService MachineStatus = "out_of_service"
)

func ParseMachineStatus(s string) (MachineStatus, error) {
	switch st := MachineStatus(s); st {
	case MachineActive, MachineInactive, MachineMaintenance, MachineOutOfService:
		return st, nil
	}
	return "", fmt.Errorf("%w: unknown machine status %q", ErrInvalidInput, s)
}

type Machine struct {
	ID              string
	Location        string
	Model           string
	Status          MachineStatus
	LastMaintenance *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func (m *Machine) Validate() error {
	if strings.TrimSpace(m.Location) == "" || strings.TrimSpace(m.Model) == "" {
		return fmt.Errorf("%w: location and model are required", ErrInvalidInput)
	}
	return nil
}

func (m *Machine) CanSell() bool { return m.Status == MachineActive }

type MaintenanceLog struct {
	ID              string
	MachineID       string
	TechnicianID    string
	MaintenanceType string
	Notes           string
	PerformedAt     time.Time
}

func (l *MaintenanceLog) Validate() error {
	if strings.TrimSpace(l.MaintenanceType) == "" {
		return fmt.Errorf("%w: maintenance_type is required", ErrInvalidInput)
	}
	return nil
}
