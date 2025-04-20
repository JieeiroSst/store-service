use actix_web::{get, post, put, delete, web, HttpResponse, Responder};
use chrono::{DateTime, Utc};
use serde::Deserialize;
use sqlx::PgPool;
use std::sync::Arc;

use crate::adapters::repositories::postgres::{PostgresBookingRepository, PostgresRoomRepository, PostgresPassengerRepository};
use crate::application::services::BookingService;
use crate::domain::models::Booking;

#[derive(Deserialize)]
struct CreateBookingRequest {
    id: i32,
    from_date: DateTime<Utc>,
    to_date: DateTime<Utc>,
    room_no: i32,
    passenger_id: i32,
}

#[derive(Deserialize)]
struct UpdateBookingRequest {
    from_date: DateTime<Utc>,
    to_date: DateTime<Utc>,
    room_no: i32,
    passenger_id: i32,
}

#[get("/bookings")]
async fn get_bookings(pool: web::Data<PgPool>) -> impl Responder {
    let booking_repo = Arc::new(PostgresBookingRepository::new(pool.get_ref().clone()));
    let room_repo = Arc::new(PostgresRoomRepository::new(pool.get_ref().clone()));
    let passenger_repo = Arc::new(PostgresPassengerRepository::new(pool.get_ref().clone()));
    let booking_service = BookingService::new(booking_repo, room_repo, passenger_repo);

    match booking_service.get_all_bookings().await {
        Ok(bookings) => HttpResponse::Ok().json(bookings),
        Err(e) => {
            log::error!("Failed to get bookings: {}", e);
            HttpResponse::InternalServerError().body("Failed to get bookings")
        }
    }
}

#[get("/bookings/{id}")]
async fn get_booking(path: web::Path<i32>, pool: web::Data<PgPool>) -> impl Responder {
    let id = path.into_inner();
    let booking_repo = Arc::new(PostgresBookingRepository::new(pool.get_ref().clone()));
    let room_repo = Arc::new(PostgresRoomRepository::new(pool.get_ref().clone()));
    let passenger_repo = Arc::new(PostgresPassengerRepository::new(pool.get_ref().clone()));
    let booking_service = BookingService::new(booking_repo, room_repo, passenger_repo);

    match booking_service.get_booking(id).await {
        Ok(Some(booking)) => HttpResponse::Ok().json(booking),
        Ok(None) => HttpResponse::NotFound().body("Booking not found"),
        Err(e) => {
            log::error!("Failed to get booking {}: {}", id, e);
            HttpResponse::InternalServerError().body("Failed to get booking")
        }
    }
}

#[post("/bookings")]
async fn create_booking(req: web::Json<CreateBookingRequest>, pool: web::Data<PgPool>) -> impl Responder {
    let booking = Booking {
        id: req.id,
        from_date: req.from_date,
        to_date: req.to_date,
        room_no: req.room_no,
        passenger_id: req.passenger_id,
    };

    let booking_repo = Arc::new(PostgresBookingRepository::new(pool.get_ref().clone()));
    let room_repo = Arc::new(PostgresRoomRepository::new(pool.get_ref().clone()));
    let passenger_repo = Arc::new(PostgresPassengerRepository::new(pool.get_ref().clone()));
    let booking_service = BookingService::new(booking_repo, room_repo, passenger_repo);

    match booking_service.create_booking(booking).await {
        Ok(created) => HttpResponse::Created().json(created),
        Err(e) => {
            log::error!("Failed to create booking: {}", e);
            HttpResponse::BadRequest().body(format!("Failed to create booking: {}", e))
        }
    }
}

#[put("/bookings/{id}")]
async fn update_booking(
    path: web::Path<i32>,
    req: web::Json<UpdateBookingRequest>,
    pool: web::Data<PgPool>
) -> impl Responder {
    let id = path.into_inner();
    let booking = Booking {
        id,
        from_date: req.from_date,
        to_date: req.to_date,
        room_no: req.room_no,
        passenger_id: req.passenger_id,
    };

    let booking_repo = Arc::new(PostgresBookingRepository::new(pool.get_ref().clone()));
    let room_repo = Arc::new(PostgresRoomRepository::new(pool.get_ref().clone()));
    let passenger_repo = Arc::new(PostgresPassengerRepository::new(pool.get_ref().clone()));
    let booking_service = BookingService::new(booking_repo, room_repo, passenger_repo);

    match booking_service.update_booking(booking).await {
        Ok(updated) => HttpResponse::Ok().json(updated),
        Err(e) => {
            log::error!("Failed to update booking {}: {}", id, e);
            HttpResponse::BadRequest().body(format!("Failed to update booking: {}", e))
        }
    }
}

#[delete("/bookings/{id}")]
async fn delete_booking(path: web::Path<i32>, pool: web::Data<PgPool>) -> impl Responder {
    let id = path.into_inner();
    let booking_repo = Arc::new(PostgresBookingRepository::new(pool.get_ref().clone()));
    let room_repo = Arc::new(PostgresRoomRepository::new(pool.get_ref().clone()));
    let passenger_repo = Arc::new(PostgresPassengerRepository::new(pool.get_ref().clone()));
    let booking_service = BookingService::new(booking_repo, room_repo, passenger_repo);

    match booking_service.delete_booking(id).await {
        Ok(true) => HttpResponse::NoContent().finish(),
        Ok(false) => HttpResponse::NotFound().body("Booking not found"),
        Err(e) => {
            log::error!("Failed to delete booking {}: {}", id, e);
            HttpResponse::BadRequest().body(format!("Failed to delete booking: {}", e))
        }
    }
}

#[get("/passengers/{passenger_id}/bookings")]
async fn get_passenger_bookings(path: web::Path<i32>, pool: web::Data<PgPool>) -> impl Responder {
    let passenger_id = path.into_inner();
    let booking_repo = Arc::new(PostgresBookingRepository::new(pool.get_ref().clone()));
    let room_repo = Arc::new(PostgresRoomRepository::new(pool.get_ref().clone()));
    let passenger_repo = Arc::new(PostgresPassengerRepository::new(pool.get_ref().clone()));
    let booking_service = BookingService::new(booking_repo, room_repo, passenger_repo);

    match booking_service.get_bookings_by_passenger(passenger_id).await {
        Ok(bookings) => HttpResponse::Ok().json(bookings),
        Err(e) => {
            log::error!("Failed to get bookings for passenger {}: {}", passenger_id, e);
            HttpResponse::InternalServerError().body("Failed to get bookings for passenger")
        }
    }
}

pub fn configure() -> actix_web::Scope {
    web::scope("/api")
        .service(get_bookings)
        .service(get_booking)
        .service(create_booking)
        .service(update_booking)
        .service(delete_booking)
        .service(get_passenger_bookings)
}