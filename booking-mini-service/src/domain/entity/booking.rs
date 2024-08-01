use serde::{Deserialize, Serialize};

#[derive(Serialize, Deserialize, Debug)]
pub struct Booking {
    booking_id: i64,
    from_date: i64,
    to_date: i64
}