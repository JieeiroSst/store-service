use serde::{Deserialize, Serialize};

#[derive(Serialize, Deserialize, Debug)]
pub struct Suite {
    room_id: String,
    amenities: String
}
