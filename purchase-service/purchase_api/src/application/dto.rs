use crate::domain::*;

#[readonly::make]
#[derive(Debug, PartialEq, Eq)]
pub struct CustomerDto {
    pub customer_id: String,
    pub first_name: String,
    pub last_name: String,
    pub email: String,
}

impl CustomerDto {
    pub fn from_entity(customer: &Customer) -> Self {
        Self {
            customer_id: customer.customer_id.clone(),
            first_name: customer.first_name.clone(),
            last_name: customer.last_name.clone(),
            email: customer.email.clone(),
        }
    }
}

#[readonly::make]
#[derive(Debug, PartialEq, Eq)]
pub struct OrdertemDto {
    pub order_item_id: String,
    pub order_id: String,
    pub poster_id: String,
    pub quantity: String,
}

impl OrdertemDto {
    pub fn from_entity(ordertem: &Ordertem) -> Self {
        Self {
            order_item_id: ordertem.order_item_id.clone(),
            order_id: ordertem.order_id.clone(),
            poster_id: ordertem.poster_id.clone(),
            quantity: ordertem.quantity.clone(),
        }
    }
}

#[readonly::make]
#[derive(Debug, PartialEq, Eq)]
pub struct OrderDto {
    pub order_id: String,
    pub customer_id: String,
    pub order_date: String,
}

impl OrderDto {
    pub fn from_entity(order: &Order) -> Self {
        Self {
            order_id: order.order_id.clone(),
            customer_id: order.customer_id.clone(),
            order_date: order.order_date.clone(),
        }
    }
}

#[readonly::make]
#[derive(Debug, PartialEq, Eq)]
pub struct PosterDto {
    pub poster_id: String,
    pub title: String,
    pub description: String,
    pub price: u64,
    pub image_url: String,
}

impl PosterDto {
    pub fn from_entity(poster: &Poster) -> Self {
        Self {
            poster_id: poster.poster_id.clone(),
            title: poster.title.clone(),
            description: poster.description.clone(),
            price: poster.price.clone(),
            image_url: poster.image_url.clone(),
        }
    }
}

#[derive(Debug)]
pub struct DtoList<T>(pub Vec<T>);
