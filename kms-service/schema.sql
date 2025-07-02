-- schema.sql

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Bảng lưu trữ khóa mã hóa
CREATE TABLE IF NOT EXISTS keys (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    alias VARCHAR(255) NOT NULL,
    encrypted_key BYTEA NOT NULL,
    algorithm VARCHAR(50) DEFAULT 'AES-256',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP NOT NULL,
    status VARCHAR(20) DEFAULT 'active'  -- active | deprecated | deleted
);

-- Bảng lưu nhật ký kiểm toán các thao tác trên khóa
CREATE TABLE IF NOT EXISTS audit_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    actor_id UUID NOT NULL,
    action VARCHAR(255) NOT NULL,         -- e.g., create, delete, rotate
    key_id UUID NOT NULL REFERENCES keys(id) ON DELETE CASCADE,
    timestamp TIMESTAMP NOT NULL DEFAULT NOW(),
    metadata TEXT
);
