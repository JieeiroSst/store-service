-- migrations/20250420000001_create_schema.sql
CREATE TABLE room_types (
    id INTEGER PRIMARY KEY,
    description VARCHAR(50) NOT NULL,
    capacity INTEGER NOT NULL
);

CREATE TABLE rooms (
    room_no INTEGER PRIMARY KEY,
    room_type_id INTEGER NOT NULL,
    FOREIGN KEY (room_type_id) REFERENCES room_types(id)
);

CREATE TABLE suites (
    room_no INTEGER PRIMARY KEY,
    amenities VARCHAR(100) NOT NULL,
    FOREIGN KEY (room_no) REFERENCES rooms(room_no) ON DELETE CASCADE
);

CREATE TABLE doubles (
    room_no INTEGER PRIMARY KEY,
    bed_count INTEGER NOT NULL,
    FOREIGN KEY (room_no) REFERENCES rooms(room_no) ON DELETE CASCADE
);

CREATE TABLE singles (
    room_no INTEGER PRIMARY KEY,
    bed_type VARCHAR(50) NOT NULL,
    FOREIGN KEY (room_no) REFERENCES rooms(room_no) ON DELETE CASCADE
);

CREATE TABLE passengers (
    id INTEGER PRIMARY KEY,
    name VARCHAR(100) NOT NULL
);

CREATE TABLE bookings (
    id INTEGER PRIMARY KEY,
    from_date TIMESTAMP WITH TIME ZONE NOT NULL,
    to_date TIMESTAMP WITH TIME ZONE NOT NULL,
    room_no INTEGER NOT NULL,
    passenger_id INTEGER NOT NULL,
    FOREIGN KEY (room_no) REFERENCES rooms(room_no),
    FOREIGN KEY (passenger_id) REFERENCES passengers(id)
);

CREATE INDEX idx_bookings_room_no ON bookings(room_no);
CREATE INDEX idx_bookings_passenger_id ON bookings(passenger_id);
CREATE INDEX idx_bookings_dates ON bookings(from_date, to_date);