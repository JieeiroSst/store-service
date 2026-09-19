-- parking-lot-service owns the parking domain: lots, floors, gates,
-- spots, vehicles, rates, tickets (parking_history) and payments. See
-- Readme.md for the system design this schema implements.
--
-- This file is applied automatically on startup by
-- internal/infrastructure/database.applySchema, which execs it statement
-- by statement (split on ';'). It's plain idempotent DDL (IF NOT EXISTS
-- everywhere) plus ON CONFLICT DO NOTHING seed inserts - no PL/pgSQL
-- blocks, since their internal semicolons would break that naive split -
-- so re-running it on every deploy is safe.
--
-- Lot/floor/gate management has no CRUD API yet (see Readme.md's
-- "Hướng mở rộng" / future extensions); a single default lot is seeded
-- below so the service is usable out of the box.

CREATE TABLE IF NOT EXISTS parking_lots (
    lot_id UUID PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    address VARCHAR(255)
);

CREATE TABLE IF NOT EXISTS floors (
    floor_id UUID PRIMARY KEY,
    lot_id UUID REFERENCES parking_lots(lot_id),
    floor_number INT NOT NULL
);

CREATE TABLE IF NOT EXISTS gates (
    gate_id UUID PRIMARY KEY,
    lot_id UUID REFERENCES parking_lots(lot_id),
    name VARCHAR(50),
    gate_type VARCHAR(10) CHECK (gate_type IN ('ENTRY', 'EXIT'))
);

CREATE TABLE IF NOT EXISTS parking_spots (
    spot_id UUID PRIMARY KEY,
    floor_id UUID REFERENCES floors(floor_id),
    type VARCHAR(50) CHECK (type IN (
        'MOTORCYCLE_SPOT', 'COMPACT_SPOT', 'LARGE_SPOT',
        'HANDICAP_SPOT', 'EV_CHARGING_SPOT'
    )),
    is_available BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE TABLE IF NOT EXISTS vehicles (
    license_plate VARCHAR(20) PRIMARY KEY,
    type VARCHAR(50) CHECK (type IN (
        'MOTORCYCLE', 'CAR', 'TRUCK', 'BUS', 'BICYCLE', 'ELECTRIC_VEHICLE'
    ))
);

CREATE TABLE IF NOT EXISTS rates (
    rate_id UUID PRIMARY KEY,
    vehicle_type VARCHAR(50) NOT NULL,
    hourly_rate DECIMAL(10,2) NOT NULL,
    daily_max_rate DECIMAL(10,2),
    effective_from TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS parking_history (
    history_id UUID PRIMARY KEY,
    vehicle_plate VARCHAR(20) REFERENCES vehicles(license_plate),
    spot_id UUID REFERENCES parking_spots(spot_id),
    entry_gate_id UUID REFERENCES gates(gate_id),
    exit_gate_id UUID REFERENCES gates(gate_id),
    parked_time TIMESTAMP NOT NULL,
    leave_time TIMESTAMP,
    status VARCHAR(20) CHECK (status IN ('ACTIVE', 'COMPLETED')) NOT NULL DEFAULT 'ACTIVE',
    CONSTRAINT check_leave_time CHECK (leave_time IS NULL OR leave_time > parked_time)
);

CREATE TABLE IF NOT EXISTS payments (
    payment_id UUID PRIMARY KEY,
    history_id UUID REFERENCES parking_history(history_id),
    amount DECIMAL(10,2) NOT NULL,
    method VARCHAR(30) CHECK (method IN ('CASH', 'CARD', 'E_WALLET')),
    paid_at TIMESTAMP,
    status VARCHAR(20) CHECK (status IN ('PENDING', 'PAID', 'FAILED')) NOT NULL DEFAULT 'PENDING'
);

CREATE INDEX IF NOT EXISTS idx_parking_history_vehicle ON parking_history(vehicle_plate);
CREATE INDEX IF NOT EXISTS idx_parking_history_spot ON parking_history(spot_id);
CREATE INDEX IF NOT EXISTS idx_parking_history_times ON parking_history(parked_time, leave_time);
CREATE INDEX IF NOT EXISTS idx_parking_spot_type_availability ON parking_spots(type, is_available);
CREATE INDEX IF NOT EXISTS idx_payments_history ON payments(history_id);

INSERT INTO parking_lots (lot_id, name, address) VALUES
    ('00000000-0000-0000-0000-000000000001', 'Main Lot', '1 Main St')
ON CONFLICT (lot_id) DO NOTHING;

INSERT INTO floors (floor_id, lot_id, floor_number) VALUES
    ('00000000-0000-0000-0000-000000000101', '00000000-0000-0000-0000-000000000001', 1)
ON CONFLICT (floor_id) DO NOTHING;

INSERT INTO gates (gate_id, lot_id, name, gate_type) VALUES
    ('00000000-0000-0000-0000-000000000201', '00000000-0000-0000-0000-000000000001', 'Gate A', 'ENTRY'),
    ('00000000-0000-0000-0000-000000000202', '00000000-0000-0000-0000-000000000001', 'Gate B', 'EXIT')
ON CONFLICT (gate_id) DO NOTHING;

INSERT INTO rates (rate_id, vehicle_type, hourly_rate, daily_max_rate, effective_from) VALUES
    ('00000000-0000-0000-0000-000000000301', 'MOTORCYCLE', 5000, 40000, '2026-01-01 00:00:00'),
    ('00000000-0000-0000-0000-000000000302', 'BICYCLE', 3000, 20000, '2026-01-01 00:00:00'),
    ('00000000-0000-0000-0000-000000000303', 'CAR', 20000, 150000, '2026-01-01 00:00:00'),
    ('00000000-0000-0000-0000-000000000304', 'TRUCK', 30000, 220000, '2026-01-01 00:00:00'),
    ('00000000-0000-0000-0000-000000000305', 'BUS', 30000, 220000, '2026-01-01 00:00:00'),
    ('00000000-0000-0000-0000-000000000306', 'ELECTRIC_VEHICLE', 20000, 150000, '2026-01-01 00:00:00')
ON CONFLICT (rate_id) DO NOTHING;

INSERT INTO parking_spots (spot_id, floor_id, type, is_available) VALUES
    ('00000000-0000-0000-0000-000000000401', '00000000-0000-0000-0000-000000000101', 'COMPACT_SPOT', true),
    ('00000000-0000-0000-0000-000000000402', '00000000-0000-0000-0000-000000000101', 'COMPACT_SPOT', true),
    ('00000000-0000-0000-0000-000000000403', '00000000-0000-0000-0000-000000000101', 'COMPACT_SPOT', true),
    ('00000000-0000-0000-0000-000000000404', '00000000-0000-0000-0000-000000000101', 'COMPACT_SPOT', true),
    ('00000000-0000-0000-0000-000000000405', '00000000-0000-0000-0000-000000000101', 'COMPACT_SPOT', true),
    ('00000000-0000-0000-0000-000000000501', '00000000-0000-0000-0000-000000000101', 'MOTORCYCLE_SPOT', true),
    ('00000000-0000-0000-0000-000000000502', '00000000-0000-0000-0000-000000000101', 'MOTORCYCLE_SPOT', true),
    ('00000000-0000-0000-0000-000000000503', '00000000-0000-0000-0000-000000000101', 'MOTORCYCLE_SPOT', true),
    ('00000000-0000-0000-0000-000000000601', '00000000-0000-0000-0000-000000000101', 'LARGE_SPOT', true),
    ('00000000-0000-0000-0000-000000000602', '00000000-0000-0000-0000-000000000101', 'LARGE_SPOT', true),
    ('00000000-0000-0000-0000-000000000701', '00000000-0000-0000-0000-000000000101', 'HANDICAP_SPOT', true),
    ('00000000-0000-0000-0000-000000000801', '00000000-0000-0000-0000-000000000101', 'EV_CHARGING_SPOT', true)
ON CONFLICT (spot_id) DO NOTHING;
