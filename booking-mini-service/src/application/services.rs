use anyhow::Result;
use chrono::{DateTime, Utc};
use std::sync::Arc;

use crate::domain::models::*;
use crate::ports::repositories::*;

pub struct RoomService {
    room_repository: Arc<dyn RoomRepository + Send + Sync>,
    room_type_repository: Arc<dyn RoomTypeRepository + Send + Sync>,
    booking_repository: Arc<dyn BookingRepository + Send + Sync>,
}

impl RoomService {
    pub fn new(
        room_repository: Arc<dyn RoomRepository + Send + Sync>,
        room_type_repository: Arc<dyn RoomTypeRepository + Send + Sync>,
        booking_repository: Arc<dyn BookingRepository + Send + Sync>,
    ) -> Self {
        Self {
            room_repository,
            room_type_repository,
            booking_repository,
        }
    }

    pub async fn get_room(&self, room_no: i32) -> Result<Option<Room>> {
        self.room_repository.find_by_id(room_no).await
    }

    pub async fn get_all_rooms(&self) -> Result<Vec<Room>> {
        self.room_repository.find_all().await
    }

    pub async fn create_room(&self, room: Room) -> Result<Room> {
        // Validate that room type exists
        let room_type = self.room_type_repository.find_by_id(room.room_type_id).await?;
        if room_type.is_none() {
            return Err(anyhow::anyhow!("Room type not found"));
        }

        self.room_repository.create(room).await
    }

    pub async fn update_room(&self, room: Room) -> Result<Room> {
        // Validate that room type exists
        let room_type = self.room_type_repository.find_by_id(room.room_type_id).await?;
        if room_type.is_none() {
            return Err(anyhow::anyhow!("Room type not found"));
        }

        self.room_repository.update(room).await
    }

    pub async fn delete_room(&self, room_no: i32) -> Result<bool> {
        // Check if room has bookings
        let bookings = self.booking_repository.find_by_room(room_no).await?;
        if !bookings.is_empty() {
            return Err(anyhow::anyhow!("Cannot delete room with existing bookings"));
        }

        self.room_repository.delete(room_no).await
    }

    pub async fn find_available_rooms(&self, from_date: DateTime<Utc>, to_date: DateTime<Utc>) -> Result<Vec<Room>> {
        if from_date >= to_date {
            return Err(anyhow::anyhow!("FromDate must be before ToDate"));
        }

        self.room_repository.find_available(from_date, to_date).await
    }
}

pub struct BookingService {
    booking_repository: Arc<dyn BookingRepository + Send + Sync>,
    room_repository: Arc<dyn RoomRepository + Send + Sync>,
    passenger_repository: Arc<dyn PassengerRepository + Send + Sync>,
}

impl BookingService {
    pub fn new(
        booking_repository: Arc<dyn BookingRepository + Send + Sync>,
        room_repository: Arc<dyn RoomRepository + Send + Sync>,
        passenger_repository: Arc<dyn PassengerRepository + Send + Sync>,
    ) -> Self {
        Self {
            booking_repository,
            room_repository,
            passenger_repository,
        }
    }

    pub async fn get_booking(&self, id: i32) -> Result<Option<Booking>> {
        self.booking_repository.find_by_id(id).await
    }

    pub async fn get_all_bookings(&self) -> Result<Vec<Booking>> {
        self.booking_repository.find_all().await
    }

    pub async fn create_booking(&self, booking: Booking) -> Result<Booking> {
        // Validate dates
        if booking.from_date >= booking.to_date {
            return Err(anyhow::anyhow!("FromDate must be before ToDate"));
        }

        // Validate room exists
        let room = self.room_repository.find_by_id(booking.room_no).await?;
        if room.is_none() {
            return Err(anyhow::anyhow!("Room not found"));
        }

        // Validate passenger exists
        let passenger = self.passenger_repository.find_by_id(booking.passenger_id).await?;
        if passenger.is_none() {
            return Err(anyhow::anyhow!("Passenger not found"));
        }

        // Check room availability
        let room_bookings = self.booking_repository.find_by_room(booking.room_no).await?;
        let room = room.unwrap();
        if !room.is_available(booking.from_date, booking.to_date, &room_bookings) {
            return Err(anyhow::anyhow!("Room is not available for the requested dates"));
        }

        self.booking_repository.create(booking).await
    }

    pub async fn update_booking(&self, booking: Booking) -> Result<Booking> {
        // Validate dates
        if booking.from_date >= booking.to_date {
            return Err(anyhow::anyhow!("FromDate must be before ToDate"));
        }

        // Validate room exists
        let room = self.room_repository.find_by_id(booking.room_no).await?;
        if room.is_none() {
            return Err(anyhow::anyhow!("Room not found"));
        }

        // Validate passenger exists
        let passenger = self.passenger_repository.find_by_id(booking.passenger_id).await?;
        if passenger.is_none() {
            return Err(anyhow::anyhow!("Passenger not found"));
        }

        // Check room availability (excluding the current booking)
        let current_booking = self.booking_repository.find_by_id(booking.id).await?;
        if current_booking.is_none() {
            return Err(anyhow::anyhow!("Booking not found"));
        }

        let room_bookings = self.booking_repository.find_by_room(booking.room_no).await?;
        let relevant_bookings: Vec<Booking> = room_bookings
            .into_iter()
            .filter(|b| b.id != booking.id)
            .collect();

        let room = room.unwrap();
        if !room.is_available(booking.from_date, booking.to_date, &relevant_bookings) {
            return Err(anyhow::anyhow!("Room is not available for the requested dates"));
        }

        self.booking_repository.update(booking).await
    }

    pub async fn delete_booking(&self, id: i32) -> Result<bool> {
        self.booking_repository.delete(id).await
    }

    pub async fn get_bookings_by_passenger(&self, passenger_id: i32) -> Result<Vec<Booking>> {
        self.booking_repository.find_by_passenger(passenger_id).await
    }
}