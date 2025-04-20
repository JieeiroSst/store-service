use actix_web::{web, HttpResponse, ResponseError};
use chrono::NaiveDate;
use serde::{Deserialize, Serialize};
use std::sync::Arc;

use crate::adapters::db::PostgresRepository;
use crate::core::domain::DomainError;
use crate::core::services::BookingServiceImpl;
use crate::ports::repository::BookingRepository;
use crate::ports::service::BookingService;

#[derive(Debug, Deserialize)]
pub struct AvailabilityQuery {
    from_date: NaiveDate,
    to_date: NaiveDate,
}

#[derive(Debug, Deserialize)]
pub struct BookingRequest {
    passenger_id: i32,
    room_no: i32,
    from_date: NaiveDate,
    to_date: NaiveDate,
}

impl ResponseError for DomainError {
    fn status_code(&self) -> actix_web::http::StatusCode {
        use actix_web::http::StatusCode;
        match self {
            DomainError::RoomNotAvailable(_) => StatusCode::CONFLICT,
            DomainError::RoomNotFound(_) => StatusCode::NOT_FOUND,
            DomainError::PassengerNotFound(_) => StatusCode::NOT_FOUND,
            DomainError::BookingNotFound(_) => StatusCode::NOT_FOUND,
            DomainError::InvalidDateRange => StatusCode::BAD_REQUEST,
        }
    }
}

pub async fn get_available_rooms(
    query: web::Query<AvailabilityQuery>,
    repo: web::Data<Arc<PostgresRepository>>,
) -> Result<HttpResponse, actix_web::Error> {
    let service = BookingServiceImpl::new(repo.get_ref().clone());

    let rooms = service.check_availability(query.from_date, query.to_date)
        .await
        .map_err(|e| {
            log::error!("Error checking availability: {:?}", e);
            actix_web::error::ErrorInternalServerError(e.to_string())
        })?;

    Ok(HttpResponse::Ok().json(rooms))
}

pub async fn create_booking(
    booking_req: web::Json<BookingRequest>,
    repo: web::Data<Arc<PostgresRepository>>,
) -> Result<HttpResponse, actix_web::Error> {
    let service = BookingServiceImpl::new(repo.get_ref().clone());

    let booking = service.make_booking(
        booking_req.passenger_id,
        booking_req.room_no,
        booking_req.from_date,
        booking_req.to_date,
    )
        .await
        .map_err(|e| {
            match e.downcast::<DomainError>() {
                Ok(domain_err) => {
                    // This unwraps the Box<DomainError> to DomainError
                    actix_web::error::Error::from(*domain_err)
                }
                Err(e) => {
                    log::error!("Error creating booking: {:?}", e);
                    actix_web::error::ErrorInternalServerError(e.to_string())
                }
            }
        })?;

    Ok(HttpResponse::Created().json(booking))
}

pub async fn get_booking(
    id: web::Path<i32>,
    repo: web::Data<Arc<PostgresRepository>>,
) -> Result<HttpResponse, actix_web::Error> {
    let booking = repo.get_booking(*id)
        .await
        .map_err(|e| {
            log::error!("Error fetching booking: {:?}", e);
            actix_web::error::ErrorInternalServerError(e.to_string())
        })?;

    match booking {
        Some(b) => Ok(HttpResponse::Ok().json(b)),
        None => Ok(HttpResponse::NotFound().finish()),
    }
}

pub async fn cancel_booking(
    id: web::Path<i32>,
    repo: web::Data<Arc<PostgresRepository>>,
) -> Result<HttpResponse, actix_web::Error> {
    let service = BookingServiceImpl::new(repo.get_ref().clone());

    let result = service.cancel_booking(*id)
        .await
        .map_err(|e| {
            log::error!("Error fetching booking: {:?}", e);
            actix_web::error::ErrorInternalServerError(e.to_string())
        })?;

    if result {
        Ok(HttpResponse::NoContent().finish())
    } else {
        Ok(HttpResponse::NotFound().finish())
    }
}