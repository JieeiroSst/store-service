use actix_web::{web, HttpResponse};
use serde::Deserialize;
use std::sync::Arc;

use crate::adapters::api::auth::Principal;
use crate::adapters::db::PostgresRepository;
use crate::core::domain::DomainError;
use crate::ports::repository::BookingRepository;

#[derive(Deserialize)]
pub struct PassengerRequest {
    name: String,
    user_id: Option<String>,
}

pub async fn register_passenger(
    principal: Principal,
    passenger_req: web::Json<PassengerRequest>,
    repo: web::Data<Arc<PostgresRepository>>,
) -> Result<HttpResponse, actix_web::Error> {
    let owner = passenger_owner_for(&principal, passenger_req.user_id.clone());

    let passenger = repo.create_owned_passenger(passenger_req.name.clone(), owner)
        .await
        .map_err(|e| {
            log::error!("Error registering passenger: {:?}", e);
            actix_web::error::ErrorInternalServerError(e.to_string())
        })?;

    Ok(HttpResponse::Created().json(passenger))
}

pub async fn get_passenger(
    principal: Principal,
    id: web::Path<i32>,
    repo: web::Data<Arc<PostgresRepository>>,
) -> Result<HttpResponse, actix_web::Error> {
    let owner = repo.passenger_owner(*id)
        .await
        .map_err(|e| {
            log::error!("Error fetching passenger owner: {:?}", e);
            actix_web::error::ErrorInternalServerError(e.to_string())
        })?;
    super::authorize_owner(&principal, owner, DomainError::PassengerNotFound(*id))?;

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
pub fn passenger_owner_for(principal: &Principal, requested: Option<String>) -> Option<String> {
    let requested = requested.map(|u| u.trim().to_string()).filter(|u| !u.is_empty());
    match (principal.is_admin(), requested) {
        (true, Some(user_id)) => Some(user_id),
        _ => principal.user_id.clone(),
    }
}

pub async fn my_passengers(
    principal: Principal,
    repo: web::Data<Arc<PostgresRepository>>,
) -> Result<HttpResponse, actix_web::Error> {
    let user_id = super::require_user(&principal)?;
    let passengers = repo.passengers_owned_by(user_id)
        .await
        .map_err(|e| {
            log::error!("Error listing passengers: {:?}", e);
            actix_web::error::ErrorInternalServerError(e.to_string())
        })?;
    Ok(HttpResponse::Ok().json(passengers))
}
