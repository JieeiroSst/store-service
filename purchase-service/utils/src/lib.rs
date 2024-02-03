use sonyflake::Sonyflake;

pub fn generator_id() -> String {
    let sf = Sonyflake::new().unwrap();
    let next_id = sf.next_id().unwrap();

    next_id.to_string()
}
