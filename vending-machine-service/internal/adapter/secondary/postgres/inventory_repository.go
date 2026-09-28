package postgres

import (
	"context"

	"github.com/JIeeiroSst/vending-machine-service/internal/domain"
	"github.com/JIeeiroSst/vending-machine-service/internal/port"
	"github.com/jackc/pgx/v5"
)

const inventorySelect = `SELECT i.inventory_id, i.machine_id, i.product_id, i.slot_identifier, i.quantity,
	i.max_capacity, i.low_threshold, i.last_restocked_at, i.created_at, i.updated_at, ` + productColumns + `
	FROM inventory i JOIN products p ON p.product_id = i.product_id`

type inventoryRepository struct{ db *DB }

func NewInventoryRepository(db *DB) port.InventoryRepository { return &inventoryRepository{db: db} }

func scanInventory(row pgx.Row) (domain.Inventory, error) {
	i := domain.Inventory{Product: &domain.Product{}}
	dest := append([]any{&i.ID, &i.MachineID, &i.ProductID, &i.SlotIdentifier, &i.Quantity,
		&i.MaxCapacity, &i.LowThreshold, &i.LastRestocked, &i.CreatedAt, &i.UpdatedAt}, productDest(i.Product)...)
	err := row.Scan(dest...)
	return i, err
}

func (r *inventoryRepository) one(ctx context.Context, sql string, args ...any) (*domain.Inventory, error) {
	i, err := scanInventory(r.db.q(ctx).QueryRow(ctx, sql, args...))
	if err != nil {
		return nil, mapErr(err)
	}
	return &i, nil
}

func (r *inventoryRepository) Create(ctx context.Context, i *domain.Inventory) error {
	_, err := r.db.q(ctx).Exec(ctx, `INSERT INTO inventory
		(inventory_id, machine_id, product_id, slot_identifier, quantity, max_capacity, low_threshold,
		 last_restocked_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		i.ID, i.MachineID, i.ProductID, i.SlotIdentifier, i.Quantity, i.MaxCapacity, i.LowThreshold,
		i.LastRestocked, i.CreatedAt, i.UpdatedAt)
	return mapErr(err)
}

func (r *inventoryRepository) Update(ctx context.Context, i *domain.Inventory) error {
	tag, err := r.db.q(ctx).Exec(ctx, `UPDATE inventory
		SET product_id = $2, quantity = $3, max_capacity = $4, low_threshold = $5, last_restocked_at = $6, updated_at = $7
		WHERE inventory_id = $1`,
		i.ID, i.ProductID, i.Quantity, i.MaxCapacity, i.LowThreshold, i.LastRestocked, i.UpdatedAt)
	if err == nil && tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return mapErr(err)
}

func (r *inventoryRepository) Get(ctx context.Context, id string) (*domain.Inventory, error) {
	return r.one(ctx, inventorySelect+` WHERE i.inventory_id = $1`, id)
}

func (r *inventoryRepository) GetForUpdate(ctx context.Context, id string) (*domain.Inventory, error) {
	return r.one(ctx, inventorySelect+` WHERE i.inventory_id = $1 FOR UPDATE OF i`, id)
}

func (r *inventoryRepository) GetBySlotForUpdate(ctx context.Context, machineID, slot string) (*domain.Inventory, error) {
	return r.one(ctx, inventorySelect+` WHERE i.machine_id = $1 AND i.slot_identifier = $2 FOR UPDATE OF i`, machineID, slot)
}

func (r *inventoryRepository) list(ctx context.Context, size int, sql string, args ...any) (domain.Page[domain.Inventory], error) {
	rows, err := r.db.q(ctx).Query(ctx, sql, append(args, size+1)...)
	if err != nil {
		return domain.Page[domain.Inventory]{}, mapErr(err)
	}
	items, err := collect(rows, scanInventory)
	if err != nil {
		return domain.Page[domain.Inventory]{}, err
	}
	return paginate(items, size, func(i *domain.Inventory) []string {
		return []string{i.MachineID, i.SlotIdentifier}
	}), nil
}

func (r *inventoryRepository) ListByMachine(ctx context.Context, machineID string, page domain.PageRequest) (domain.Page[domain.Inventory], error) {
	after, err := cursorArgs(page.Cursor, 2)
	if err != nil {
		return domain.Page[domain.Inventory]{}, err
	}
	return r.list(ctx, page.Size(), inventorySelect+`
		WHERE i.machine_id = $1 AND ($2::text IS NULL OR i.slot_identifier > $3::text)
		ORDER BY i.slot_identifier LIMIT $4`, machineID, after[0], after[1])
}

func (r *inventoryRepository) ListLow(ctx context.Context, machineID string, page domain.PageRequest) (domain.Page[domain.Inventory], error) {
	after, err := cursorArgs(page.Cursor, 2)
	if err != nil {
		return domain.Page[domain.Inventory]{}, err
	}
	return r.list(ctx, page.Size(), inventorySelect+`
		WHERE i.quantity <= i.low_threshold AND ($1 = '' OR i.machine_id = $1)
		  AND ($2::text IS NULL OR (i.machine_id, i.slot_identifier) > ($2::text, $3::text))
		ORDER BY i.machine_id, i.slot_identifier LIMIT $4`, machineID, after[0], after[1])
}
