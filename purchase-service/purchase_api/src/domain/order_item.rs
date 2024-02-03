use fake::{Dummy, Fake};

#[derive(Clone, Dummy, PartialEq, Eq)]
#[readonly::make]
pub struct Order_Item {
    pub order_item_id: String,
    pub order_id: String,
    pub poster_id: String,
    pub quantity: String,
}

impl Order_Item {
    pub fn new(order_item_id: &str, order_id: &str, poster_id: &str, quantity: &str) -> Self {
        Self {
            order_item_id: order_item_id.to_string(),
            order_id: order_id.to_string(),
            poster_id: poster_id.to_string(),
            quantity: quantity.to_string(),
        }
    }

    pub fn edit(&mut self, order_id: &str, poster_id: &str, quantity: &str) {
        self.order_id = order_id.to_string();
        self.poster_id = poster_id.to_string();
        self.quantity = quantity.to_string();
    }
}
