use sonyflake::Sonyflake;
use base64::{encode, decode};

pub fn generator_id() -> String {
    let sf = Sonyflake::new().unwrap();
    let next_id = sf.next_id().unwrap();

    next_id.to_string()
}

pub fn encoded_data(key: String) -> String{
    encode(key)
}

pub fn decoded_data(encoded_data: String) -> String {
    let decoded_data = decode(&encoded_data).unwrap();
    let decoded_string = String::from_utf8(decoded_data).unwrap();

    decoded_string
}