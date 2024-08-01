use serde::{Deserialize, Serialize};

#[derive(Serialize, Deserialize, Debug)]
pub struct  Room {
    room_no: String
}
