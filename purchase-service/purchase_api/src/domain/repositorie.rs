use crate::domain::*;
use mockall::predicate::*;
use mockall::*;

#[automock]
pub trait CustomerRepository {
    fn by_id(&self, id: &str) -> Result<Customer, String>;
    fn save(&self, Customer: Customer);
    fn next_identity(&self) -> String;
    fn all(&self) -> Vec<Customer>;
}
