use std::io::Write;
use chrono::prelude::*;

// write_log("Đây là một thông báo log", "INFO").unwrap();
// write_log("Đây là một lỗi", "ERROR").unwrap();
fn write_log(message: &str, level: &str) -> Result<(), std::io::Error> {
    let mut log_file = std::fs::OpenOptions::new()
        .append(true)
        .create(true)
        .open("log.txt")?;

    let timestamp = chrono::Utc::now().format("%Y-%m-%d %H:%M:%S");
    let log_entry = format!("{} {} {}\n", timestamp, level, message);

    log_file.write_all(log_entry.as_bytes())?;

    Ok(())
}