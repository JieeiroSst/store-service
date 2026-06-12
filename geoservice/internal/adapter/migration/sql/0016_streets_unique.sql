-- 0016_streets_unique.sql : thêm unique constraint cho bảng streets.
-- Cần chạy sau 0014_streets_table.sql và trước khi seed data.
-- Nếu đã có data trùng, xóa trước khi chạy: TRUNCATE streets;

ALTER TABLE streets
    ADD CONSTRAINT uq_streets_province_name UNIQUE (province_code, name);
