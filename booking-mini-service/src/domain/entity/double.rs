use serde::{Deserialize, Serialize};

#[derive(Serialize, Deserialize, Debug)]
pub struct Double {
    room_id: String,
    bed_count: i64
}