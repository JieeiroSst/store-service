use async_trait::async_trait;
use chrono::NaiveDate;
use std::error::Error;

use crate::core::domain::{Booking, Passenger, Room};

#[async_trait]
pub trait BookingRepository: Send + Sync {
    async fn get_room(&self, room_no: i32) -> Result<Option<Room>, Box<dyn Error>>;
    async fn get_available_rooms(&self, from_date: NaiveDate, to_date: NaiveDate) -> Result<Vec<Room>, Box<dyn Error>>;
    async fn create_booking(&self, booking: Booking) -> Result<Booking, Box<dyn Error>>;
    async fn get_booking(&self, booking_id: i32) -> Result<Option<Booking>, Box<dyn Error>>;
    async fn update_booking(&self, booking: Booking) -> Result<Booking, Box<dyn Error>>;
    async fn delete_booking(&self, booking_id: i32) -> Result<bool, Box<dyn Error>>;
    async fn get_passenger(&self, passenger_id: i32) -> Result<Option<Passenger>, Box<dyn Error>>;
    async fn create_passenger(&self, passenger: Passenger) -> Result<Passenger, Box<dyn Error>>;
}