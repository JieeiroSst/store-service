use serde::{Deserialize, Serialize};

#[derive(Serialize, Deserialize, Debug)]
pub struct RoomType {
    room_id: String,
    description: String,
    capacity: i64
}