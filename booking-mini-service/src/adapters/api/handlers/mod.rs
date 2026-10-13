pub mod booking_handlers;
pub mod passenger_handlers;

use crate::adapters::api::auth::Principal;
use crate::core::domain::DomainError;

pub fn authorize_owner(
    principal: &Principal,
    owner: Option<Option<String>>,
    not_found: DomainError,
) -> Result<(), actix_web::Error> {
    match owner {
        None => Err(not_found.into()),
        Some(owner) => Ok(principal.require_owner(owner.as_deref())?),
    }
}
