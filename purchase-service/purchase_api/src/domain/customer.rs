use fake::{Dummy, Fake};

#[derive(Clone, Dummy, PartialEq, Eq)]
#[readonly::make]
pub struct Customer {
    pub customer_id: String,
    pub first_name: String,
    pub last_name: String,
    pub email: String,
}

impl Customer {
    pub fn new(customer_id: &str, first_name: &str, last_name: &str, email: &str) -> Self {
        Self {
            customer_id: customer_id.to_string(),
            first_name: first_name.to_string(),
            last_name: last_name.to_string(),
            email: email.to_string(),
        }
    }

    pub fn edit(&mut self, first_name: &str, last_name: &str, email: &str) {
        self.first_name = first_name.to_string();
        self.last_name = last_name.to_string();
        self.email = email.to_string();
    }
}
