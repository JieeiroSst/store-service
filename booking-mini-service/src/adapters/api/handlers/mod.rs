pub mod booking_handlers;
pub mod passenger_handlers;

use crate::adapters::api::auth::{AuthError, Principal};
use crate::core::domain::DomainError;

pub fn require_user(principal: &Principal) -> Result<&str, actix_web::Error> {
    principal
        .user_id
        .as_deref()
        .filter(|id| !id.is_empty())
        .ok_or_else(|| AuthError::Unauthenticated.into())
}

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
