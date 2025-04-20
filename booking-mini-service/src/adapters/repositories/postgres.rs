// src/adapters/repositories/postgres.rs
use async_trait::async_trait;
use chrono::{DateTime, Utc};
use sqlx::PgPool;

use crate::domain::models::*;
use crate::ports::repositories::*;

pub struct PostgresRoomTypeRepository {
    pool: PgPool,
}

impl PostgresRoomTypeRepository {
    pub fn new(pool: PgPool) -> Self {
        Self { pool }
    }
}

#[async_trait]
impl RoomTypeRepository for PostgresRoomTypeRepository {
    async fn find_by_id(&self, id: i32) -> Result<Option<RoomType>, anyhow::Error> {
        let room_type = sqlx::query_as!(
            RoomType,
            r#"SELECT id as "id!", description as "description!", capacity as "capacity!" FROM room_types WHERE id = $1"#,
            id
        )
            .fetch_optional(&self.pool)
            .await?;

        Ok(room_type)
    }

    async fn find_all(&self) -> Result<Vec<RoomType>, anyhow::Error> {
        let room_types = sqlx::query_as!(
            RoomType,
            r#"SELECT id as "id!", description as "description!", capacity as "capacity!" FROM room_types"#
        )
            .fetch_all(&self.pool)
            .await?;

        Ok(room_types)
    }

    async fn create(&self, room_type: RoomType) -> Result<RoomType, anyhow::Error> {
        let created = sqlx::query_as!(
            RoomType,
            r#"
            INSERT INTO room_types (id, description, capacity)
            VALUES ($1, $2, $3)
            RETURNING id as "id!", description as "description!", capacity as "capacity!"
            "#,
            room_type.id, room_type.description, room_type.capacity
        )
            .fetch_one(&self.pool)
            .await?;

        Ok(created)
    }

    async fn update(&self, room_type: RoomType) -> Result<RoomType, anyhow::Error> {
        let updated = sqlx::query_as!(
            RoomType,
            r#"
            UPDATE room_types
            SET description = $2, capacity = $3
            WHERE id = $1
            RETURNING id as "id!", description as "description!", capacity as "capacity!"
            "#,
            room_type.id, room_type.description, room_type.capacity
        )
            .fetch_one(&self.pool)
            .await?;

        Ok(updated)
    }

    async fn delete(&self, id: i32) -> Result<bool, anyhow::Error> {
        let result = sqlx::query!("DELETE FROM room_types WHERE id = $1", id)
            .execute(&self.pool)
            .await?;

        Ok(result.rows_affected() > 0)
    }
}

pub struct PostgresRoomRepository {
    pool: PgPool,
}

impl PostgresRoomRepository {
    pub fn new(pool: PgPool) -> Self {
        Self { pool }
    }
}

#[async_trait]
impl RoomRepository for PostgresRoomRepository {
    async fn find_by_id(&self, room_no: i32) -> Result<Option<Room>, anyhow::Error> {
        let room = sqlx::query_as!(
            Room,
            r#"SELECT room_no as "room_no!", room_type_id as "room_type_id!" FROM rooms WHERE room_no = $1"#,
            room_no
        )
            .fetch_optional(&self.pool)
            .await?;

        Ok(room)
    }

    async fn find_all(&self) -> Result<Vec<Room>, anyhow::Error> {
        let rooms = sqlx::query_as!(
            Room,
            r#"SELECT room_no as "room_no!", room_type_id as "room_type_id!" FROM rooms"#
        )
            .fetch_all(&self.pool)
            .await?;

        Ok(rooms)
    }

    async fn create(&self, room: Room) -> Result<Room, anyhow::Error> {
        let created = sqlx::query_as!(
            Room,
            r#"
            INSERT INTO rooms (room_no, room_type_id)
            VALUES ($1, $2)
            RETURNING room_no as "room_no!", room_type_id as "room_type_id!"
            "#,
            room.room_no, room.room_type_id
        )
            .fetch_one(&self.pool)
            .await?;

        Ok(created)
    }

    async fn update(&self, room: Room) -> Result<Room, anyhow::Error> {
        let updated = sqlx::query_as!(
            Room,
            r#"
            UPDATE rooms
            SET room_type_id = $2
            WHERE room_no = $1
            RETURNING room_no as "room_no!", room_type_id as "room_type_id!"
            "#,
            room.room_no, room.room_type_id
        )
            .fetch_one(&self.pool)
            .await?;

        Ok(updated)
    }

    async fn delete(&self, room_no: i32) -> Result<bool, anyhow::Error> {
        let result = sqlx::query!("DELETE FROM rooms WHERE room_no = $1", room_no)
            .execute(&self.pool)
            .await?;

        Ok(result.rows_affected() > 0)
    }

    async fn find_available(&self, from_date: DateTime<Utc>, to_date: DateTime<Utc>) -> Result<Vec<Room>, anyhow::Error> {
        let rooms = sqlx::query_as!(
            Room,
            r#"
            SELECT r.room_no as "room_no!", r.room_type_id as "room_type_id!"
            FROM rooms r
            WHERE r.room_no NOT IN (
                SELECT b.room_no
                FROM bookings b
                WHERE (b.from_date <= $2 AND b.to_date >= $1)
            )
            "#,
            from_date, to_date
        )
            .fetch_all(&self.pool)
            .await?;

        Ok(rooms)
    }
}

pub struct PostgresBookingRepository {
    pool: PgPool,
}

impl PostgresBookingRepository {
    pub fn new(pool: PgPool) -> Self {
        Self { pool }
    }
}

#[async_trait]
impl BookingRepository for PostgresBookingRepository {
    async fn find_by_id(&self, id: i32) -> Result<Option<Booking>, anyhow::Error> {
        let booking = sqlx::query_as!(
            Booking,
            r#"
            SELECT
                id as "id!",
                from_date as "from_date!",
                to_date as "to_date!",
                room_no as "room_no!",
                passenger_id as "passenger_id!"
            FROM bookings
            WHERE id = $1
            "#,
            id
        )
            .fetch_optional(&self.pool)
            .await?;

        Ok(booking)
    }

    async fn find_all(&self) -> Result<Vec<Booking>, anyhow::Error> {
        let bookings = sqlx::query_as!(
            Booking,
            r#"
            SELECT
                id as "id!",
                from_date as "from_date!",
                to_date as "to_date!",
                room_no as "room_no!",
                passenger_id as "passenger_id!"
            FROM bookings
            "#
        )
            .fetch_all(&self.pool)
            .await?;

        Ok(bookings)
    }

    async fn find_by_room(&self, room_no: i32) -> Result<Vec<Booking>, anyhow::Error> {
        let bookings = sqlx::query_as!(
            Booking,
            r#"
            SELECT
                id as "id!",
                from_date as "from_date!",
                to_date as "to_date!",
                room_no as "room_no!",
                passenger_id as "passenger_id!"
            FROM bookings
            WHERE room_no = $1
            "#,
            room_no
        )
            .fetch_all(&self.pool)
            .await?;

        Ok(bookings)
    }

    async fn find_by_passenger(&self, passenger_id: i32) -> Result<Vec<Booking>, anyhow::Error> {
        let bookings = sqlx::query_as!(
            Booking,
            r#"
            SELECT
                id as "id!",
                from_date as "from_date!",
                to_date as "to_date!",
                room_no as "room_no!",
                passenger_id as "passenger_id!"
            FROM bookings
            WHERE passenger_id = $1
            "#,
            passenger_id
        )
            .fetch_all(&self.pool)
            .await?;

        Ok(bookings)
    }

    async fn create(&self, booking: Booking) -> Result<Booking, anyhow::Error> {
        let created = sqlx::query_as!(
            Booking,
            r#"
            INSERT INTO bookings (id, from_date, to_date, room_no, passenger_id)
            VALUES ($1, $2, $3, $4, $5)
            RETURNING
                id as "id!",
                from_date as "from_date!",
                to_date as "to_date!",
                room_no as "room_no!",
                passenger_id as "passenger_id!"
            "#,
            booking.id, booking.from_date, booking.to_date, booking.room_no, booking.passenger_id
        )
            .fetch_one(&self.pool)
            .await?;

        Ok(created)
    }

    async fn update(&self, booking: Booking) -> Result<Booking, anyhow::Error> {
        let updated = sqlx::query_as!(
            Booking,
            r#"
            UPDATE bookings
            SET from_date = $2, to_date = $3, room_no = $4, passenger_id = $5
            WHERE id = $1
            RETURNING
                id as "id!",
                from_date as "from_date!",
                to_date as "to_date!",
                room_no as "room_no!",
                passenger_id as "passenger_id!"
            "#,
            booking.id, booking.from_date, booking.to_date, booking.room_no, booking.passenger_id
        )
            .fetch_one(&self.pool)
            .await?;

        Ok(updated)
    }

    async fn delete(&self, id: i32) -> Result<bool, anyhow::Error> {
        let result = sqlx::query!("DELETE FROM bookings WHERE id = $1", id)
            .execute(&self.pool)
            .await?;

        Ok(result.rows_affected() > 0)
    }
}

pub struct PostgresPassengerRepository {
    pool: PgPool,
}

impl PostgresPassengerRepository {
    pub fn new(pool: PgPool) -> Self {
        Self { pool }
    }
}

#[async_trait]
impl PassengerRepository for PostgresPassengerRepository {
    async fn find_by_id(&self, id: i32) -> Result<Option<Passenger>, anyhow::Error> {
        let passenger = sqlx::query_as!(
            Passenger,
            r#"SELECT id as "id!", name as "name!" FROM passengers WHERE id = $1"#,
            id
        )
            .fetch_optional(&self.pool)
            .await?;

        Ok(passenger)
    }

    async fn find_all(&self) -> Result<Vec<Passenger>, anyhow::Error> {
        let passengers = sqlx::query_as!(
            Passenger,
            r#"SELECT id as "id!", name as "name!" FROM passengers"#
        )
            .fetch_all(&self.pool)
            .await?;

        Ok(passengers)
    }

    async fn create(&self, passenger: Passenger) -> Result<Passenger, anyhow::Error> {
        let created = sqlx::query_as!(
            Passenger,
            r#"
            INSERT INTO passengers (id, name)
            VALUES ($1, $2)
            RETURNING id as "id!", name as "name!"
            "#,
            passenger.id, passenger.name
        )
            .fetch_one(&self.pool)
            .await?;

        Ok(created)
    }

    async fn update(&self, passenger: Passenger) -> Result<Passenger, anyhow::Error> {
        let updated = sqlx::query_as!(
            Passenger,
            r#"
            UPDATE passengers
            SET name = $2
            WHERE id = $1
            RETURNING id as "id!", name as "name!"
            "#,
            passenger.id, passenger.name
        )
            .fetch_one(&self.pool)
            .await?;

        Ok(updated)
    }

    async fn delete(&self, id: i32) -> Result<bool, anyhow::Error> {
        let result = sqlx::query!("DELETE FROM passengers WHERE id = $1", id)
            .execute(&self.pool)
            .await?;

        Ok(result.rows_affected() > 0)
    }
}