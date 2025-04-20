use actix_web::{get, post, put, delete, web, HttpResponse, Responder};
use serde::Deserialize;
use sqlx::PgPool;
use std::sync::Arc;

use crate::adapters::repositories::postgres::PostgresPassengerRepository;
use crate::domain::models::Passenger;
use crate::ports::repositories::PassengerRepository;

#[derive(Deserialize)]
struct CreatePassengerRequest {
    id: i32,
    name: String,
}

#[derive(Deserialize)]
struct UpdatePassengerRequest {
    name: String,
}

#[get("/passengers")]
async fn get_passengers(pool: web::Data<PgPool>) -> impl Responder {
    let repo = Arc::new(PostgresPassengerRepository::new(pool.get_ref().clone()));

    match repo.find_all().await {
        Ok(passengers) => HttpResponse::Ok().json(passengers),
        Err(e) => {
            log::error!("Failed to get passengers: {}", e);
            HttpResponse::InternalServerError().body("Failed to get passengers")
        }
    }
}

#[get("/passengers/{id}")]
async fn get_passenger(path: web::Path<i32>, pool: web::Data<PgPool>) -> impl Responder {
    let id = path.into_inner();
    let repo = Arc::new(PostgresPassengerRepository::new(pool.get_ref().clone()));

    match repo.find_by_id(id).await {
        Ok(Some(passenger)) => HttpResponse::Ok().json(passenger),
        Ok(None) => HttpResponse::NotFound().body("Passenger not found"),
        Err(e) => {
            log::error!("Failed to get passenger {}: {}", id, e);
            HttpResponse::InternalServerError().body("Failed to get passenger")
        }
    }
}

#[post("/passengers")]
async fn create_passenger(req: web::Json<CreatePassengerRequest>, pool: web::Data<PgPool>) -> impl Responder {
    let passenger = Passenger {
        id: req.id,
        name: req.name.clone(),
    };

    let repo = Arc::new(PostgresPassengerRepository::new(pool.get_ref().clone()));

    match repo.create(passenger).await {
        Ok(created) => HttpResponse::Created().json(created),
        Err(e) => {
            log::error!("Failed to create passenger: {}", e);
            HttpResponse::BadRequest().body(format!("Failed to create passenger: {}", e))
        }
    }
}

#[put("/passengers/{id}")]
async fn update_passenger(
    path: web::Path<i32>,
    req: web::Json<UpdatePassengerRequest>,
    pool: web::Data<PgPool>
) -> impl Responder {
    let id = path.into_inner();
    let passenger = Passenger {
        id,
        name: req.name.clone(),
    };

    let repo = Arc::new(PostgresPassengerRepository::new(pool.get_ref().clone()));

    match repo.update(passenger).await {
        Ok(updated) => HttpResponse::Ok().json(updated),
        Err(e) => {
            log::error!("Failed to update passenger {}: {}", id, e);
            HttpResponse::BadRequest().body(format!("Failed to update passenger: {}", e))
        }
    }
}

#[delete("/passengers/{id}")]
async fn delete_passenger(path: web::Path<i32>, pool: web::Data<PgPool>) -> impl Responder {
    let id = path.into_inner();
    let repo = Arc::new(PostgresPassengerRepository::new(pool.get_ref().clone()));

    match repo.delete(id).await {
        Ok(true) => HttpResponse::NoContent().finish(),
        Ok(false) => HttpResponse::NotFound().body("Passenger not found"),
        Err(e) => {
            log::error!("Failed to delete passenger {}: {}", id, e);
            HttpResponse::BadRequest().body(format!("Failed to delete passenger: {}", e))
        }
    }
}

pub fn configure() -> actix_web::Scope {
    web::scope("/api")
        .service(get_passengers)
        .service(get_passenger)
        .service(create_passenger)
        .service(update_passenger)
        .service(delete_passenger)
}