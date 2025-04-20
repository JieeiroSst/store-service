use async_trait::async_trait;
use crate::domain::models::*;
use chrono::{DateTime, Utc};

#[async_trait]
pub trait RoomTypeRepository {
    async fn find_by_id(&self, id: i32) -> Result<Option<RoomType>, anyhow::Error>;
    async fn find_all(&self) -> Result<Vec<RoomType>, anyhow::Error>;
    async fn create(&self, room_type: RoomType) -> Result<RoomType, anyhow::Error>;
    async fn update(&self, room_type: RoomType) -> Result<RoomType, anyhow::Error>;
    async fn delete(&self, id: i32) -> Result<bool, anyhow::Error>;
}

#[async_trait]
pub trait RoomRepository {
    async fn find_by_id(&self, room_no: i32) -> Result<Option<Room>, anyhow::Error>;
    async fn find_all(&self) -> Result<Vec<Room>, anyhow::Error>;
    async fn create(&self, room: Room) -> Result<Room, anyhow::Error>;
    async fn update(&self, room: Room) -> Result<Room, anyhow::Error>;
    async fn delete(&self, room_no: i32) -> Result<bool, anyhow::Error>;
    async fn find_available(&self, from_date: DateTime<Utc>, to_date: DateTime<Utc>) -> Result<Vec<Room>, anyhow::Error>;
}

#[async_trait]
pub trait SuiteRepository {
    async fn find_by_room_no(&self, room_no: i32) -> Result<Option<Suite>, anyhow::Error>;
    async fn create(&self, suite: Suite) -> Result<Suite, anyhow::Error>;
    async fn update(&self, suite: Suite) -> Result<Suite, anyhow::Error>;
}

#[async_trait]
pub trait DoubleRepository {
    async fn find_by_room_no(&self, room_no: i32) -> Result<Option<Double>, anyhow::Error>;
    async fn create(&self, double: Double) -> Result<Double, anyhow::Error>;
    async fn update(&self, double: Double) -> Result<Double, anyhow::Error>;
}

#[async_trait]
pub trait SingleRepository {
    async fn find_by_room_no(&self, room_no: i32) -> Result<Option<Single>, anyhow::Error>;
    async fn create(&self, single: Single) -> Result<Single, anyhow::Error>;
    async fn update(&self, single: Single) -> Result<Single, anyhow::Error>;
}

#[async_trait]
pub trait PassengerRepository {
    async fn find_by_id(&self, id: i32) -> Result<Option<Passenger>, anyhow::Error>;
    async fn find_all(&self) -> Result<Vec<Passenger>, anyhow::Error>;
    async fn create(&self, passenger: Passenger) -> Result<Passenger, anyhow::Error>;
    async fn update(&self, passenger: Passenger) -> Result<Passenger, anyhow::Error>;
    async fn delete(&self, id: i32) -> Result<bool, anyhow::Error>;
}

#[async_trait]
pub trait BookingRepository {
    async fn find_by_id(&self, id: i32) -> Result<Option<Booking>, anyhow::Error>;
    async fn find_all(&self) -> Result<Vec<Booking>, anyhow::Error>;
    async fn find_by_room(&self, room_no: i32) -> Result<Vec<Booking>, anyhow::Error>;
    async fn find_by_passenger(&self, passenger_id: i32) -> Result<Vec<Booking>, anyhow::Error>;
    async fn create(&self, booking: Booking) -> Result<Booking, anyhow::Error>;
    async fn update(&self, booking: Booking) -> Result<Booking, anyhow::Error>;
    async fn delete(&self, id: i32) -> Result<bool, anyhow::Error>;
}
