mod handlers;

use actix_web::web;
use handlers::{booking_handlers, passenger_handlers};

pub fn config_routes(cfg: &mut web::ServiceConfig) {
    cfg.service(
        web::scope("/api")
            .service(
                web::scope("/rooms")
                    .route("", web::get().to(booking_handlers::get_available_rooms))
            )
            .service(
                web::scope("/bookings")
                    .route("", web::post().to(booking_handlers::create_booking))
                    .route("/{id}", web::get().to(booking_handlers::get_booking))
                    .route("/{id}", web::delete().to(booking_handlers::cancel_booking))
            )
            .service(
                web::scope("/passengers")
                    .route("", web::post().to(passenger_handlers::register_passenger))
                    .route("/{id}", web::get().to(passenger_handlers::get_passenger))
            )
    );
}