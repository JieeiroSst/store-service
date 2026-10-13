CREATE SEQUENCE IF NOT EXISTS passengers_passengerid_seq OWNED BY Passengers.PassengerId;
SELECT setval('passengers_passengerid_seq', COALESCE((SELECT MAX(PassengerId) FROM Passengers), 0) + 1, false);
ALTER TABLE Passengers ALTER COLUMN PassengerId SET DEFAULT nextval('passengers_passengerid_seq');

CREATE SEQUENCE IF NOT EXISTS bookings_bookingid_seq OWNED BY Bookings.BookingId;
SELECT setval('bookings_bookingid_seq', COALESCE((SELECT MAX(BookingId) FROM Bookings), 0) + 1, false);
ALTER TABLE Bookings ALTER COLUMN BookingId SET DEFAULT nextval('bookings_bookingid_seq');

ALTER TABLE Passengers ADD COLUMN IF NOT EXISTS UserId VARCHAR(64);
CREATE INDEX IF NOT EXISTS idx_passengers_userid ON Passengers (UserId);
