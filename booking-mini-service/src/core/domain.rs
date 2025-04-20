use chrono::NaiveDate;
use serde::{Deserialize, Serialize};

#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum RoomType {
    Single,
    Double,
    Suite,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RoomTypeDetails {
    pub id: i32,
    pub description: String,
    pub capacity: i32,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Room {
    pub room_no: i32,
    pub room_type: RoomType,
    pub room_type_details: RoomTypeDetails,
    // Specific fields based on room type
    pub bed_type: Option<String>,       // For Singles
    pub bed_count: Option<i32>,         // For Doubles
    pub amenities: Option<String>,      // For Suites
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Passenger {
    pub id: i32,
    pub name: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Booking {
    pub id: i32,
    pub from_date: NaiveDate,
    pub to_date: NaiveDate,
    pub room_no: i32,
    pub passenger_id: i32,
    pub passenger: Option<Passenger>,
    pub room: Option<Room>,
}

#[derive(Debug, thiserror::Error)]
pub enum DomainError {
    #[error("Room {0} is not available for the requested dates")]
    RoomNotAvailable(i32),

    #[error("Room {0} does not exist")]
    RoomNotFound(i32),

    #[error("Passenger {0} does not exist")]
    PassengerNotFound(i32),

    #[error("Booking {0} does not exist")]
    BookingNotFound(i32),

    #[error("Invalid date range: from date must be before to date")]
    InvalidDateRange,
}
