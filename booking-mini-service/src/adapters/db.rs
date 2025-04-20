use async_trait::async_trait;
use chrono::NaiveDate;
use log::{error, info};
use sqlx::{postgres::PgPoolOptions, PgPool};
use std::error::Error;

use crate::core::domain::{Booking, Passenger, Room, RoomType, RoomTypeDetails};
use crate::ports::repository::BookingRepository;

pub struct PostgresRepository {
    pool: PgPool,
}

impl PostgresRepository {
    pub async fn new(database_url: &str) -> Result<Self, Box<dyn Error>> {
        let pool = PgPoolOptions::new()
            .max_connections(5)
            .connect(database_url)
            .await?;

        info!("Connected to database");
        Ok(Self { pool })
    }
}

#[async_trait]
impl BookingRepository for PostgresRepository {

    async fn get_room(&self, room_no: i32) -> Result<Option<Room>, Box<dyn Error>> {
        // SQL query that joins the Rooms table with RoomTypes and all specific room type tables
        let room_query = sqlx::query!(
        r#"
        SELECT r.RoomNo, rt.RoomTypeId, rt.Description, rt.Capacity,
               s.BedType as "bed_type?",
               d.BedCount as "bed_count?",
               su.Amenities as "amenities?"
        FROM Rooms r
        JOIN RoomTypes rt ON r.RoomTypeId = rt.RoomTypeId
        LEFT JOIN Singles s ON r.RoomNo = s.RoomNo
        LEFT JOIN Doubles d ON r.RoomNo = d.RoomNo
        LEFT JOIN Suites su ON r.RoomNo = su.RoomNo
        WHERE r.RoomNo = $1
        "#,
        room_no
    )
            .fetch_optional(&self.pool)
            .await?;

        match room_query {
            Some(row) => {
                // Determine room type based on which specialized table has data
                let room_type = if row.bed_type.is_some() {
                    RoomType::Single
                } else if row.bed_count.is_some() {
                    RoomType::Double
                } else {
                    RoomType::Suite
                };

                let room = Room {
                    room_no: row.roomno,
                    room_type,
                    room_type_details: RoomTypeDetails {
                        id: row.roomtypeid,
                        description: row.description,
                        capacity: row.capacity,
                    },
                    bed_type: row.bed_type,
                    bed_count: row.bed_count,
                    amenities: row.amenities,
                };

                Ok(Some(room))
            },
            None => Ok(None),
        }
    }

    async fn get_available_rooms(&self, from_date: NaiveDate, to_date: NaiveDate) -> Result<Vec<Room>, Box<dyn Error>> {
        // SQL query that finds rooms not in any booking that overlaps with the requested date range
        let rooms = sqlx::query!(
        r#"
        SELECT r.RoomNo, rt.RoomTypeId, rt.Description, rt.Capacity,
               s.BedType as "bed_type?",
               d.BedCount as "bed_count?",
               su.Amenities as "amenities?"
        FROM Rooms r
        JOIN RoomTypes rt ON r.RoomTypeId = rt.RoomTypeId
        LEFT JOIN Singles s ON r.RoomNo = s.RoomNo
        LEFT JOIN Doubles d ON r.RoomNo = d.RoomNo
        LEFT JOIN Suites su ON r.RoomNo = su.RoomNo
        WHERE r.RoomNo NOT IN (
            SELECT RoomNo
            FROM Bookings
            WHERE (FromDate <= $1 AND ToDate > $1)  -- Booking starts before and ends after from_date
               OR (FromDate < $2 AND ToDate >= $2)  -- Booking starts before and ends after to_date
               OR (FromDate >= $1 AND ToDate <= $2) -- Booking is completely within requested range
        )
        "#,
        from_date,
        to_date
    )
            .fetch_all(&self.pool)
            .await?;

        let mut available_rooms = Vec::new();
        for row in rooms {
            // Determine room type based on which specialized table has data
            let room_type = if row.bed_type.is_some() {
                RoomType::Single
            } else if row.bed_count.is_some() {
                RoomType::Double
            } else {
                RoomType::Suite
            };

            let room = Room {
                room_no: row.roomno.unwrap(),
                room_type,
                room_type_details: RoomTypeDetails {
                    id: row.roomtypeid.unwrap(),
                    description: row.description.unwrap(),
                    capacity: row.capacity.unwrap(),
                },
                bed_type: row.bed_type,
                bed_count: row.bed_count,
                amenities: row.amenities,
            };

            available_rooms.push(room);
        }

        Ok(available_rooms)
    }
    async fn create_booking(&self, booking: Booking) -> Result<Booking, Box<dyn Error>> {
        let result = sqlx::query!(
            r#"
            INSERT INTO Bookings (FromDate, ToDate, RoomNo, PassengerId)
            VALUES ($1, $2, $3, $4)
            RETURNING BookingId
            "#,
            booking.from_date,
            booking.to_date,
            booking.room_no,
            booking.passenger_id
        )
            .fetch_one(&self.pool)
            .await?;

        let booking_id = result.bookingid;

        // Get the complete booking with room and passenger details
        self.get_booking(booking_id).await.map(|opt| opt.unwrap())
    }

    async fn get_booking(&self, booking_id: i32) -> Result<Option<Booking>, Box<dyn Error>> {
        let booking = sqlx::query!(
            r#"
            SELECT b.BookingId, b.FromDate, b.ToDate, b.RoomNo, b.PassengerId,
                   p.Name as PassengerName
            FROM Bookings b
            JOIN Passengers p ON b.PassengerId = p.PassengerId
            WHERE b.BookingId = $1
            "#,
            booking_id
        )
            .fetch_optional(&self.pool)
            .await?;

        match booking {
            Some(row) => {
                let room = self.get_room(row.roomno).await?;

                let booking = Booking {
                    id: row.bookingid,
                    from_date: row.fromdate,
                    to_date: row.todate,
                    room_no: row.roomno,
                    passenger_id: row.passengerid,
                    passenger: Some(Passenger {
                        id: row.passengerid,
                        name: row.passengername,
                    }),
                    room,
                };

                Ok(Some(booking))
            },
            None => Ok(None),
        }
    }

    async fn update_booking(&self, booking: Booking) -> Result<Booking, Box<dyn Error>> {
        sqlx::query!(
            r#"
            UPDATE Bookings
            SET FromDate = $1, ToDate = $2, RoomNo = $3, PassengerId = $4
            WHERE BookingId = $5
            "#,
            booking.from_date,
            booking.to_date,
            booking.room_no,
            booking.passenger_id,
            booking.id
        )
            .execute(&self.pool)
            .await?;

        self.get_booking(booking.id).await.map(|opt| opt.unwrap())
    }

    async fn delete_booking(&self, booking_id: i32) -> Result<bool, Box<dyn Error>> {
        let result = sqlx::query!(
            "DELETE FROM Bookings WHERE BookingId = $1",
            booking_id
        )
            .execute(&self.pool)
            .await?;

        Ok(result.rows_affected() > 0)
    }

    async fn get_passenger(&self, passenger_id: i32) -> Result<Option<Passenger>, Box<dyn Error>> {
        let passenger = sqlx::query!(
            "SELECT PassengerId, Name FROM Passengers WHERE PassengerId = $1",
            passenger_id
        )
            .fetch_optional(&self.pool)
            .await?;

        match passenger {
            Some(row) => {
                Ok(Some(Passenger {
                    id: row.passengerid,
                    name: row.name,
                }))
            },
            None => Ok(None),
        }
    }

    async fn create_passenger(&self, passenger: Passenger) -> Result<Passenger, Box<dyn Error>> {
        let result = sqlx::query!(
            "INSERT INTO Passengers (Name) VALUES ($1) RETURNING PassengerId",
            passenger.name
        )
            .fetch_one(&self.pool)
            .await?;

        Ok(Passenger {
            id: result.passengerid,
            name: passenger.name,
        })
    }
}