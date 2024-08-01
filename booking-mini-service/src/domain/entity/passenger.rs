use serde::{Deserialize, Serialize};

#[derive(Serialize, Deserialize, Debug)]
pub struct Passenger {
    passenger_id: i64,
    name: String
}