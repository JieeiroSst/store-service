use actix_web::{get, post, put, delete, web, HttpResponse, Responder};
use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};
use sqlx::PgPool;
use std::sync::Arc;

use crate::adapters::repositories::postgres::{PostgresRoomRepository, PostgresRoomTypeRepository, PostgresBookingRepository};
use crate::application::services::RoomService;
use crate::domain::models::Room;

#[derive(Deserialize)]
struct CreateRoomRequest {
    room_no: i32,
    room_type_id: i32,
}

#[derive(Deserialize)]
struct UpdateRoomRequest {
    room_type_id: i32,
}

#[derive(Deserialize)]
struct AvailabilityQuery {
    from_date: DateTime<Utc>,
    to_date: DateTime<Utc>,
}

#[get("/rooms")]
async fn get_rooms(pool: web::Data<PgPool>) -> impl Responder {
    let room_repo = Arc::new(PostgresRoomRepository::new(pool.get_ref().clone()));
    let room_type_repo = Arc::new(PostgresRoomTypeRepository::new(pool.get_ref().clone()));
    let booking_repo = Arc::new(PostgresBookingRepository::new(pool.get_ref().clone()));
    let room_service = RoomService::new(room_repo, room_type_repo, booking_repo);

    match room_service.get_all_rooms().await {
        Ok(rooms) => HttpResponse::Ok().json(rooms),
        Err(e) => {
            log::error!("Failed to get rooms: {}", e);
            HttpResponse::InternalServerError().body("Failed to get rooms")
        }
    }
}

#[get("/rooms/{room_no}")]
async fn get_room(path: web::Path<i32>, pool: web::Data<PgPool>) -> impl Responder {
    let room_no = path.into_inner();
    let room_repo = Arc::new(PostgresRoomRepository::new(pool.get_ref().clone()));
    let room_type_repo = Arc::new(PostgresRoomTypeRepository::new(pool.get_ref().clone()));
    let booking_repo = Arc::new(PostgresBookingRepository::new(pool.get_ref().clone()));
    let room_service = RoomService::new(room_repo, room_type_repo, booking_repo);

    match room_service.get_room(room_no).await {
        Ok(Some(room)) => HttpResponse::Ok().json(room),
        Ok(None) => HttpResponse::NotFound().body("Room not found"),
        Err(e) => {
            log::error!("Failed to get room {}: {}", room_no, e);
            HttpResponse::InternalServerError().body("Failed to get room")
        }
    }
}

#[post("/rooms")]
async fn create_room(req: web::Json<CreateRoomRequest>, pool: web::Data<PgPool>) -> impl Responder {
    let room = Room {
        room_no: req.room_no,
        room_type_id: req.room_type_id,
    };

    let room_repo = Arc::new(PostgresRoomRepository::new(pool.get_ref().clone()));
    let room_type_repo = Arc::new(PostgresRoomTypeRepository::new(pool.get_ref().clone()));
    let booking_repo = Arc::new(PostgresBookingRepository::new(pool.get_ref().clone()));
    let room_service = RoomService::new(room_repo, room_type_repo, booking_repo);

    match room_service.create_room(room).await {
        Ok(created) => HttpResponse::Created().json(created),
        Err(e) => {
            log::error!("Failed to create room: {}", e);
            HttpResponse::BadRequest().body(format!("Failed to create room: {}", e))
        }
    }
}

#[put("/rooms/{room_no}")]
async fn update_room(
    path: web::Path<i32>,
    req: web::Json<UpdateRoomRequest>,
    pool: web::Data<PgPool>
) -> impl Responder {
    let room_no = path.into_inner();
    let room = Room {
        room_no,
        room_type_id: req.room_type_id,
    };

    let room_repo = Arc::new(PostgresRoomRepository::new(pool.get_ref().clone()));
    let room_type_repo = Arc::new(PostgresRoomTypeRepository::new(pool.get_ref().clone()));
    let booking_repo = Arc::new(PostgresBookingRepository::new(pool.get_ref().clone()));
    let room_service = RoomService::new(room_repo, room_type_repo, booking_repo);

    match room_service.update_room(room).await {
        Ok(updated) => HttpResponse::Ok().json(updated),
        Err(e) => {
            log::error!("Failed to update room {}: {}", room_no, e);
            HttpResponse::BadRequest().body(format!("Failed to update room: {}", e))
        }
    }
}

#[delete("/rooms/{room_no}")]
async fn delete_room(path: web::Path<i32>, pool: web::Data<PgPool>) -> impl Responder {
    let room_no = path.into_inner();
    let room_repo = Arc::new(PostgresRoomRepository::new(pool.get_ref().clone()));
    let room_type_repo = Arc::new(PostgresRoomTypeRepository::new(pool.get_ref().clone()));
    let booking_repo = Arc::new(PostgresBookingRepository::new(pool.get_ref().clone()));
    let room_service = RoomService::new(room_repo, room_type_repo, booking_repo);

    match room_service.delete_room(room_no).await {
        Ok(true) => HttpResponse::NoContent().finish(),
        Ok(false) => HttpResponse::NotFound().body("Room not found"),
        Err(e) => {
            log::error!("Failed to delete room {}: {}", room_no, e);
            HttpResponse::BadRequest().body(format!("Failed to delete room: {}", e))
        }
    }
}

#[get("/rooms/available")]
async fn get_available_rooms(query: web::Query<AvailabilityQuery>, pool: web::Data<PgPool>) -> impl Responder {
    let room_repo = Arc::new(PostgresRoomRepository::new(pool.get_ref().clone()));
    let room_type_repo = Arc::new(PostgresRoomTypeRepository::new(pool.get_ref().clone()));
    let booking_repo = Arc::new(PostgresBookingRepository::new(pool.get_ref().clone()));
    let room_service = RoomService::new(room_repo, room_type_repo, booking_repo);

    match room_service.find_available_rooms(query.from_date, query.to_date).await {
        Ok(rooms) => HttpResponse::Ok().json(rooms),
        Err(e) => {
            log::error!("Failed to get available rooms: {}", e);
            HttpResponse::BadRequest().body(format!("Failed to get available rooms: {}", e))
        }
    }
}

pub fn configure() -> actix_web::Scope {
    web::scope("/api")
        .service(get_rooms)
        .service(get_room)
        .service(create_room)
        .service(update_room)
        .service(delete_room)
        .service(get_available_rooms)
}