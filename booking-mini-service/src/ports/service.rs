use async_trait::async_trait;
use chrono::NaiveDate;
use std::error::Error;

use crate::core::domain::{Booking, Passenger, Room};

#[async_trait]
pub trait BookingService: Send + Sync {
    async fn check_availability(&self, from_date: NaiveDate, to_date: NaiveDate) -> Result<Vec<Room>, Box<dyn Error>>;
    async fn make_booking(&self, passenger_id: i32, room_no: i32, from_date: NaiveDate, to_date: NaiveDate) -> Result<Booking, Box<dyn Error>>;
    async fn cancel_booking(&self, booking_id: i32) -> Result<bool, Box<dyn Error>>;
    async fn register_passenger(&self, name: String) -> Result<Passenger, Box<dyn Error>>;
}