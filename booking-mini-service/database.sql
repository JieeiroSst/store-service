-- Create RoomTypes table
CREATE TABLE RoomTypes (
    RoomTypeId INTEGER PRIMARY KEY,
    Description VARCHAR(50) NOT NULL,
    Capacity INTEGER NOT NULL
);

-- Create Rooms table (parent table)
CREATE TABLE Rooms (
    RoomNo INTEGER PRIMARY KEY,
    RoomTypeId INTEGER NOT NULL,
    FOREIGN KEY (RoomTypeId) REFERENCES RoomTypes(RoomTypeId)
);

-- Create Suites table (child of Rooms)
CREATE TABLE Suites (
    RoomNo INTEGER PRIMARY KEY,
    Amenities VARCHAR(100) NOT NULL,
    FOREIGN KEY (RoomNo) REFERENCES Rooms(RoomNo)
);

-- Create Doubles table (child of Rooms)
CREATE TABLE Doubles (
    RoomNo INTEGER PRIMARY KEY,
    BedCount INTEGER NOT NULL,
    FOREIGN KEY (RoomNo) REFERENCES Rooms(RoomNo)
);

-- Create Singles table (child of Rooms)
CREATE TABLE Singles (
    RoomNo INTEGER PRIMARY KEY,
    BedType VARCHAR(50) NOT NULL,
    FOREIGN KEY (RoomNo) REFERENCES Rooms(RoomNo)
);

-- Create Passengers table
CREATE TABLE Passengers (
    PassengerId INTEGER PRIMARY KEY,
    Name VARCHAR(100) NOT NULL
);

-- Create Bookings table with relationships
CREATE TABLE Bookings (
    BookingId INTEGER PRIMARY KEY,
    FromDate DATE NOT NULL,
    ToDate DATE NOT NULL,
    RoomNo INTEGER NOT NULL,
    PassengerId INTEGER NOT NULL,
    FOREIGN KEY (RoomNo) REFERENCES Rooms(RoomNo),
    FOREIGN KEY (PassengerId) REFERENCES Passengers(PassengerId)
);