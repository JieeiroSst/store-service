use fake::{Dummy, Fake};

#[derive(Clone, Dummy, PartialEq, Eq)]
#[readonly::make]
pub struct Order {
    pub order_id: String,
    pub customer_id: String,
    pub order_date: String,
}

impl Order {
    pub fn new(order_id: &str, customer_id: &str, order_date: &str) -> Self {
        Self {
            order_id: order_id.to_string(),
            customer_id: customer_id.to_string(),
            order_date: order_date.to_string(),
        }
    }

    pub fn edit(&mut self, customer_id: &str, order_date: &str) {
        self.customer_id = customer_id.to_string();
        self.order_date = order_date.to_string();
    }
}
