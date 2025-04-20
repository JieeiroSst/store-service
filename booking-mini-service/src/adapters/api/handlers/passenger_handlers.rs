use actix_web::{web, HttpResponse};
use serde::Deserialize;
use std::sync::Arc;

use crate::adapters::db::PostgresRepository;
use crate::core::services::BookingServiceImpl;
use crate::ports::repository::BookingRepository;
use crate::ports::service::BookingService;

#[derive(Deserialize)]
pub struct PassengerRequest {
    name: String,
}

pub async fn register_passenger(
    passenger_req: web::Json<PassengerRequest>,
    repo: web::Data<Arc<PostgresRepository>>,
) -> Result<HttpResponse, actix_web::Error> {
    let service = BookingServiceImpl::new(repo.get_ref().clone());

    let passenger = service.register_passenger(passenger_req.name.clone())
        .await
        .map_err(|e| {
            log::error!("Error registering passenger: {:?}", e);
            actix_web::error::ErrorInternalServerError(e.to_string())
        })?;

    Ok(HttpResponse::Created().json(passenger))
}

pub async fn get_passenger(
    id: web::Path<i32>,
    repo: web::Data<Arc<PostgresRepository>>,
) -> Result<HttpResponse, actix_web::Error> {
    let passenger = repo.get_passenger(*id)
        .await
        .map_err(|e| {
            log::error!("Error fetching passenger: {:?}", e);
            actix_web::error::ErrorInternalServerError(e.to_string())
        })?;

    match passenger {
        Some(p) => Ok(HttpResponse::Ok().json(p)),
        None => Ok(HttpResponse::NotFound().finish()),
    }
}