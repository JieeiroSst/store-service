```
CREATE TABLE Customers (
    MaKH INT PRIMARY KEY,
    TenKH VARCHAR(50),
    SDT VARCHAR(20),
    Email VARCHAR(50),
    DiaChi TEXT
);

CREATE TABLE Tickets (
    MaVe INT PRIMARY KEY,
    TenPhim VARCHAR(100),
    NgayChieu DATE,
    GioChieu TIME,
    PhongChieu VARCHAR(20),
    GiaVe DECIMAL(10,2),
    SoLuong INT,
    TrangThai VARCHAR(20)
);

CREATE TABLE Invoices (
    MaHD INT PRIMARY KEY AUTO_INCREMENT,  -- Mã hóa đơn, tự tăng
    MaKH INT,
    NgayLapHD DATE,
    TongTien DECIMAL(10,2),
    GhiChu TEXT,
    FOREIGN KEY (MaKH) REFERENCES Customers(MaKH)
);

CREATE TABLE InvoiceDetails (
    MaCTHD INT PRIMARY KEY AUTO_INCREMENT,  -- Mã chi tiết hóa đơn, tự tăng
    MaHD INT,
    MaVe INT,
    SoLuong INT,
    FOREIGN KEY (MaHD) REFERENCES Invoices(MaHD),
    FOREIGN KEY (MaVe) REFERENCES Tickets(MaVe)
);

-- Tạo khóa ngoại
ALTER TABLE Invoices ADD FOREIGN KEY (MaKH) REFERENCES Customers(MaKH);
ALTER TABLE InvoiceDetails ADD FOREIGN KEY (MaHD) REFERENCES Invoices(MaHD);
ALTER TABLE InvoiceDetails ADD FOREIGN KEY (MaVe) REFERENCES Tickets(MaVe);

```