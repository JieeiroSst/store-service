use serde::{Deserialize, Serialize};

#[derive(Serialize, Deserialize, Debug)]
pub struct Single {
    room_id: String,
    bed_type: String
}