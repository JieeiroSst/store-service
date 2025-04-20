use async_trait::async_trait;
use chrono::NaiveDate;
use std::error::Error;
use std::sync::Arc;

use crate::core::domain::{Booking, DomainError, Passenger, Room};
use crate::ports::repository::BookingRepository;
use crate::ports::service::BookingService;

pub struct BookingServiceImpl<R: BookingRepository> {
    repository: Arc<R>,
}

impl<R: BookingRepository> BookingServiceImpl<R> {
    pub fn new(repository: Arc<R>) -> Self {
        Self { repository }
    }
}

#[async_trait]
impl<R: BookingRepository + 'static> BookingService for BookingServiceImpl<R> {
    async fn check_availability(&self, from_date: NaiveDate, to_date: NaiveDate) -> Result<Vec<Room>, Box<dyn Error>> {
        if from_date >= to_date {
            return Err(Box::new(DomainError::InvalidDateRange));
        }

        self.repository.get_available_rooms(from_date, to_date).await
    }

    async fn make_booking(&self, passenger_id: i32, room_no: i32, from_date: NaiveDate, to_date: NaiveDate) -> Result<Booking, Box<dyn Error>> {
        if from_date >= to_date {
            return Err(Box::new(DomainError::InvalidDateRange));
        }

        // Check if passenger exists
        let passenger = self.repository.get_passenger(passenger_id).await?;
        if passenger.is_none() {
            return Err(Box::new(DomainError::PassengerNotFound(passenger_id)));
        }

        // Check if room exists
        let room = self.repository.get_room(room_no).await?;
        if room.is_none() {
            return Err(Box::new(DomainError::RoomNotFound(room_no)));
        }

        // Check if room is available for the requested dates
        let available_rooms = self.repository.get_available_rooms(from_date, to_date).await?;
        if !available_rooms.iter().any(|r| r.room_no == room_no) {
            return Err(Box::new(DomainError::RoomNotAvailable(room_no)));
        }

        // Create booking
        let booking = Booking {
            id: 0, // Will be set by the repository
            from_date,
            to_date,
            room_no,
            passenger_id,
            passenger: None,
            room: None,
        };

        self.repository.create_booking(booking).await
    }

    async fn cancel_booking(&self, booking_id: i32) -> Result<bool, Box<dyn Error>> {
        // Check if booking exists
        let booking = self.repository.get_booking(booking_id).await?;
        if booking.is_none() {
            return Err(Box::new(DomainError::BookingNotFound(booking_id)));
        }

        self.repository.delete_booking(booking_id).await
    }

    async fn register_passenger(&self, name: String) -> Result<Passenger, Box<dyn Error>> {
        let passenger = Passenger {
            id: 0, // Will be set by the repository
            name,
        };

        self.repository.create_passenger(passenger).await
    }
}
