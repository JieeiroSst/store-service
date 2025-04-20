use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};

#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct RoomType {
    pub id: i32,
    pub description: String,
    pub capacity: i32,
}

#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct Room {
    pub room_no: i32,
    pub room_type_id: i32,
}

#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct Suite {
    pub room_no: i32,
    pub amenities: String,
}

#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct Double {
    pub room_no: i32,
    pub bed_count: i32,
}

#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct Single {
    pub room_no: i32,
    pub bed_type: String,
}

#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct Passenger {
    pub id: i32,
    pub name: String,
}

#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct Booking {
    pub id: i32,
    pub from_date: DateTime<Utc>,
    pub to_date: DateTime<Utc>,
    pub room_no: i32,
    pub passenger_id: i32,
}

impl Room {
    pub fn is_available(&self, from_date: DateTime<Utc>, to_date: DateTime<Utc>, booking: &[Booking]) -> bool {
        !booking.iter().any(|booking|{
            booking.room_no == self.room_no && !(booking.to_date <= from_date || booking.from_date >= to_date)
        })
    }
}