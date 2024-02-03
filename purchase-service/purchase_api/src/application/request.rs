#[readonly::make]
pub struct CreateCustomerRequest {
    pub first_name: String,
    pub last_name: String,
    pub email: String,
}

impl CreateCustomerRequest {
    pub fn new(first_name: &str, last_name: &str, email: &str) -> Self {
        Self {
            first_name: first_name.to_string(),
            last_name: last_name.to_string(),
            email: email.to_string(),
        }
    }
}

#[readonly::make]
pub struct GetCustomerRequest {
    pub customer_id: String,
}

impl GetCustomerRequest {
    pub fn new(customer_id: &str) -> Self {
        Self {
            customer_id: customer_id.to_string(),
        }
    }
}

pub struct ListCustomerRequest {
    pub page: u64,
    pub limit: u64,
}

impl ListCustomerRequest {
    pub fn new(page: u64, limit: u64) -> Self {
        Self {
            page: page,
            limit: limit,
        }
    }
}

#[readonly::make]
pub struct EditCustomerRequest {
    pub customer_id: String,
    pub first_name: String,
    pub last_name: String,
    pub email: String,
}

impl EditCustomerRequest {
    pub fn new(customer_id: &str, first_name: &str, last_name: &str, email: &str) -> Self {
        Self {
            customer_id: customer_id.to_string(),
            first_name: first_name.to_string(),
            last_name: last_name.to_string(),
            email: email.to_string(),
        }
    }
}

#[readonly::make]
pub struct DeleteCustomerRequest {
    pub customer_id: String,
}

impl DeleteCustomerRequest {
    pub fn new(customer_id: &str) -> Self {
        Self {
            customer_id: customer_id.to_string(),
        }
    }
}

#[readonly::make]
pub struct CreateOrdertemRequest {
    pub order_id: String,
    pub poster_id: String,
    pub quantity: String,
}

impl CreateOrdertemRequest {
    pub fn new(order_id: &str, poster_id: &str, quantity: &str) -> Self {
        Self {
            order_id: order_id.to_string(),
            poster_id: poster_id.to_string(),
            quantity: quantity.to_string(),
        }
    }
}

#[readonly::make]
pub struct GetOrdertemRequest {
    pub order_item_id: String,
}

impl GetOrdertemRequest {
    pub fn new(order_item_id: &str) -> Self {
        Self {
            order_item_id: order_item_id.to_string(),
        }
    }
}

#[readonly::make]
pub struct EditOrdertemRequest {
    pub order_item_id: String,
    pub order_id: String,
    pub poster_id: String,
    pub quantity: String,
}

impl EditOrdertemRequest {
    pub fn new(order_item_id: &str, order_id: &str, poster_id: &str, quantity: &str) -> Self {
        Self {
            order_item_id: order_item_id.to_string(),
            order_id: order_id.to_string(),
            poster_id: poster_id.to_string(),
            quantity: quantity.to_string(),
        }
    }
}

#[readonly::make]
pub struct DeleteOrdertemRequest {
    pub order_item_id: String,
}

impl DeleteOrdertemRequest {
    pub fn new(order_item_id: &str) -> Self {
        Self {
            order_item_id: order_item_id.to_string(),
        }
    }
}

#[readonly::make]
pub struct ListOrdertemRequest {
    pub page: u64,
    pub limit: u64,
}

impl ListOrdertemRequest {
    pub fn new(page: u64, limit: u64) -> Self {
        Self {
            page: page,
            limit: limit,
        }
    }
}

#[readonly::make]
pub struct CreateOrderRequest {
    pub customer_id: String,
    pub order_date: String,
}

impl CreateOrderRequest {
    pub fn new() -> Self {
        Self {
            customer_id: customer_id.to_string(),
            order_date: order_date.to_string(),
        }
    }
}

#[readonly::make]
pub struct GetOrderRequest {
    pub order_id: String,
}

impl GetOrderRequest {
    pub fn new() -> Self {
        Self {
            order_id: order_id.to_string(),
        }
    }
}

#[readonly::make]
pub struct EditOrderRequest {
    pub order_id: String,
    pub customer_id: String,
    pub order_date: String,
}

impl EditOrderRequest {
    pub fn new() -> Self {
        Self {
            order_id: order_id.to_string(),
            customer_id: customer_id.to_string(),
            order_date: order_date.to_string(),
        }
    }
}

#[readonly::make]
pub struct DeleteOrderRequest {
    pub order_id: String,
}

impl DeleteOrderRequest {
    pub fn new() -> Self {
        Self {
            order_id: order_id.to_string(),
        }
    }
}

#[readonly::make]
pub struct ListOrderRequest {
    pub page: u64,
    pub limit: u64,
}

impl ListOrderRequest {
    pub fn new(page: u64, limit: u64) -> Self {
        Self {
            page: page,
            limit: limit,
        }
    }
}


#[readonly::make]
pub struct CreatePosterRequest {
    pub title: String,
    pub description: String,
    pub price: String,
    pub image_url: String,
}

impl CreatePosterReques {
    pub fn new(title: &str, description: &str, price: &str, image_url: &str) -> Self {
        Self {
            title: title.to_string(),
            description: description.to_string(),
            price: price.to_string(),
            image_url: image_url.to_string(),
        }
    }
}

#[readonly::make]
pub struct EditPosterRequest {
    pub poster_id: String,
    pub title: String,
    pub description: String,
    pub price: String,
    pub image_url: String,
}

impl EditPosterRequest {
    pub fn new(poster_id: &str,title: &str, description: &str, price: &str, image_url: &str) -> Self {
        Self {
            poster_id: poster_id.to_string(),
            title: title.to_string(),
            description: description.to_string(),
            price: price.to_string(),
            image_url: image_url.to_string(),
        }
    }
}

#[readonly::make]
pub struct DeletePosterRequest {
    pub poster_id: String,
}

impl DeletePosterRequest {
    pub fn new(poster_id: &str) -> Self {
        Self {
            poster_id: poster_id.to_string(),
        }
    }
}

#[readonly::make]
pub struct GetPosterRequest {
    pub poster_id: String,
}

impl GetPosterRequest {
    pub fn new(poster_id: &str) -> Self {
        Self {
            poster_id: poster_id.to_string(),
        }
    }
}

#[readonly::make]
pub struct ListPosterRequest {
    pub page: u64,
    pub limit: u64,
}

impl ListPosterRequest {
    pub fn new(page: u64, limit: u64) -> Self {
        Self {
            page: page,
            limit: limit,
        }
    }
}