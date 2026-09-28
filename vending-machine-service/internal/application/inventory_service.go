package application

import (
	"context"
	"errors"
	"time"

	"github.com/JIeeiroSst/vending-machine-service/internal/domain"
	"github.com/JIeeiroSst/vending-machine-service/internal/port"
	"github.com/google/uuid"
)

type inventoryService struct {
	tx        port.Transactor
	machines  port.MachineRepository
	products  port.ProductRepository
	inventory port.InventoryRepository
	events    port.EventRepository
	now       func() time.Time
}

func NewInventoryService(
	tx port.Transactor,
	machines port.MachineRepository,
	products port.ProductRepository,
	inventory port.InventoryRepository,
	events port.EventRepository,
) port.InventoryService {
	return &inventoryService{tx: tx, machines: machines, products: products, inventory: inventory, events: events, now: time.Now}
}

func (s *inventoryService) AssignSlot(ctx context.Context, machineID, slot string, in port.SlotInput) (*domain.Inventory, error) {
	inv := &domain.Inventory{
		MachineID:      machineID,
		ProductID:      in.ProductID,
		SlotIdentifier: slot,
		Quantity:       in.Quantity,
		MaxCapacity:    in.MaxCapacity,
		LowThreshold:   in.LowThreshold,
	}
	if err := inv.Validate(); err != nil {
		return nil, err
	}
	err := s.tx.WithinTx(ctx, func(ctx context.Context) error {
		if _, err := s.machines.Get(ctx, machineID); err != nil {
			return err
		}
		product, err := s.products.Get(ctx, in.ProductID)
		if err != nil {
			return err
		}
		inv.Product = product

		now := s.now().UTC()
		inv.UpdatedAt = now
		existing, err := s.inventory.GetBySlotForUpdate(ctx, machineID, slot)
		switch {
		case errors.Is(err, domain.ErrNotFound):
			inv.ID = uuid.NewString()
			inv.CreatedAt = now
			return s.inventory.Create(ctx, inv)
		case err != nil:
			return err
		}
		inv.ID = existing.ID
		inv.CreatedAt = existing.CreatedAt
		inv.LastRestocked = existing.LastRestocked
		return s.inventory.Update(ctx, inv)
	})
	if err != nil {
		return nil, err
	}
	return inv, nil
}

func (s *inventoryService) ListByMachine(ctx context.Context, machineID string, page domain.PageRequest) (domain.Page[domain.Inventory], error) {
	if _, err := s.machines.Get(ctx, machineID); err != nil {
		return domain.Page[domain.Inventory]{}, err
	}
	return s.inventory.ListByMachine(ctx, machineID, page)
}

func (s *inventoryService) ListLow(ctx context.Context, machineID string, page domain.PageRequest) (domain.Page[domain.Inventory], error) {
	return s.inventory.ListLow(ctx, machineID, page)
}

func (s *inventoryService) Restock(ctx context.Context, inventoryID string, units int) (*domain.Inventory, error) {
	var inv *domain.Inventory
	err := s.tx.WithinTx(ctx, func(ctx context.Context) error {
		var err error
		if inv, err = s.inventory.GetForUpdate(ctx, inventoryID); err != nil {
			return err
		}
		now := s.now().UTC()
		added, err := inv.Restock(units, now)
		if err != nil {
			return err
		}
		inv.UpdatedAt = now
		if err := s.inventory.Update(ctx, inv); err != nil {
			return err
		}
		return s.events.Append(ctx, newEvent(now, domain.EventInventoryRestocked, "inventory", inv.ID, inv.MachineID,
			map[string]any{"slot": inv.SlotIdentifier, "added": added, "quantity": inv.Quantity}))
	})
	if err != nil {
		return nil, err
	}
	return inv, nil
}
