package domain

import (
	"fmt"
	"strings"
	"time"
)

type Inventory struct {
	ID             string
	MachineID      string
	ProductID      string
	SlotIdentifier string
	Quantity       int
	MaxCapacity    int
	LowThreshold   int
	LastRestocked  *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time

	Product *Product
}

func (i *Inventory) Validate() error {
	if strings.TrimSpace(i.SlotIdentifier) == "" || i.ProductID == "" {
		return fmt.Errorf("%w: slot and product_id are required", ErrInvalidInput)
	}
	if i.MaxCapacity <= 0 {
		return fmt.Errorf("%w: max_capacity must be positive", ErrInvalidInput)
	}
	if i.Quantity < 0 || i.Quantity > i.MaxCapacity {
		return fmt.Errorf("%w: quantity must be between 0 and max_capacity", ErrInvalidInput)
	}
	if i.LowThreshold < 0 || i.LowThreshold > i.MaxCapacity {
		return fmt.Errorf("%w: low_threshold must be between 0 and max_capacity", ErrInvalidInput)
	}
	return nil
}

func (i *Inventory) IsLow() bool { return i.Quantity <= i.LowThreshold }

func (i *Inventory) Take() error {
	if i.Quantity <= 0 {
		return ErrOutOfStock
	}
	i.Quantity--
	return nil
}

func (i *Inventory) Return() {
	if i.Quantity < i.MaxCapacity {
		i.Quantity++
	}
}

func (i *Inventory) Restock(units int, at time.Time) (int, error) {
	if units <= 0 {
		return 0, fmt.Errorf("%w: quantity must be positive", ErrInvalidInput)
	}
	added := min(units, i.MaxCapacity-i.Quantity)
	i.Quantity += added
	i.LastRestocked = &at
	return added, nil
}
