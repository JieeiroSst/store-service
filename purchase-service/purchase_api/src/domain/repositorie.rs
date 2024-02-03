use crate::domain::*;
use mockall::predicate::*;
use mockall::*;

#[automock]
pub trait CustomerRepository {
    fn by_id(&self, id: &str) -> Result<Customer, String>;
    fn save(&self, Customer: customer);
    fn next_identity(&self) -> String;
    fn all(&self, limit: u64, page: u64) -> Vec<Customer>;
}

#[automock]
pub trait OrderItemRepository {
    fn by_id(&self, id: &str) -> Result<Ordertem, String>;
    fn save(&self, Ordertem: ordertem);
    fn next_identity(&self) -> String;
    fn all(&self, limit: u64, page: u64) -> Vec<Ordertem>;
}

#[automock]
pub trait OrderRepository {
    fn by_id(&self, id: &str) -> Result<Order, String>;
    fn save(&self, Order: order);
    fn next_identity(&self) -> String;
    fn all(&self, limit: u64, page: u64) -> Vec<Order>;
}

#[automock]
pub trait PosterRepository {
    fn by_id(&self, id: &str) -> Result<Poster, String>;
    fn save(&self, Poster: poster);
    fn next_identity(&self) -> String;
    fn all(&self, limit: u64, page: u64) -> Vec<Poster>;
}
