use actix_web::{test as atest, web, App, HttpResponse, HttpServer};
use base64::Engine;
use booking_mini_service::adapters::api::auth::{AuthError, Authenticator, Principal, INTERNAL_TOKEN_HEADER};
use booking_mini_service::adapters::api::handlers::authorize_owner;
use booking_mini_service::adapters::api::handlers::passenger_handlers::passenger_owner_for;
use booking_mini_service::core::domain::DomainError;
use std::time::Duration;

const INTERNAL: &str = "internal-secret";

async fn validate(body: web::Json<serde_json::Value>) -> HttpResponse {
    let token = body["session_token"].as_str().unwrap_or_default();
    let payload = token.split('.').nth(1).unwrap_or_default();
    let claims: serde_json::Value = base64::engine::general_purpose::URL_SAFE_NO_PAD
        .decode(payload)
        .ok()
        .and_then(|b| serde_json::from_slice(&b).ok())
        .unwrap_or_default();
    match claims["sub"].as_str() {
        Some(sub) => HttpResponse::Ok().json(serde_json::json!({ "valid": true, "user_id": sub })),
        None => HttpResponse::Unauthorized().finish(),
    }
}

fn fake_user_service() -> String {
    let server = HttpServer::new(|| App::new().route("/api/v1/validate", web::post().to(validate)))
        .workers(1)
        .bind(("127.0.0.1", 0))
        .unwrap();
    let addr = server.addrs()[0];
    actix_web::rt::spawn(server.run());
    format!("http://{addr}")
}

fn token(user_id: &str, role: &str) -> String {
    let enc = |v: serde_json::Value| base64::engine::general_purpose::URL_SAFE_NO_PAD.encode(v.to_string());
    format!(
        "{}.{}.sig",
        enc(serde_json::json!({ "alg": "HS256" })),
        enc(serde_json::json!({ "sub": user_id, "role": role }))
    )
}

fn authenticator(url: &str, tokens: &[&str]) -> Authenticator {
    let tokens: Vec<String> = tokens.iter().map(|t| t.to_string()).collect();
    Authenticator::new(url, &tokens, Duration::from_secs(2))
}

async fn whoami(p: Principal) -> HttpResponse {
    HttpResponse::Ok().json(serde_json::json!({ "user_id": p.user_id, "admin": p.is_admin() }))
}

#[actix_web::test]
async fn bearer_token_resolves_user_and_role() {
    let auth = authenticator(&fake_user_service(), &[INTERNAL]);

    let user = auth.authenticate(Some(&format!("Bearer {}", token("7", "user"))), None).await.unwrap();
    assert_eq!(user.user_id.as_deref(), Some("7"));
    assert!(!user.is_admin());

    let admin = auth.authenticate(Some(&format!("Bearer {}", token("1", "admin"))), None).await.unwrap();
    assert!(admin.is_admin());
}

#[actix_web::test]
async fn rejected_session_is_unauthenticated() {
    let auth = authenticator(&fake_user_service(), &[INTERNAL]);
    let err = auth.authenticate(Some("Bearer not-a-jwt"), None).await.unwrap_err();
    assert_eq!(err, AuthError::Unauthenticated);
}

#[actix_web::test]
async fn internal_token_is_checked() {
    let auth = authenticator("http://127.0.0.1:1", &[INTERNAL]);
    assert!(auth.authenticate(None, Some(INTERNAL)).await.unwrap().service);
    assert_eq!(auth.authenticate(None, Some("wrong")).await.unwrap_err(), AuthError::Unauthenticated);
}

#[actix_web::test]
async fn anonymous_depends_on_rollout_mode() {
    assert!(authenticator("http://127.0.0.1:1", &[]).authenticate(None, None).await.unwrap().service);
    assert_eq!(
        authenticator("http://127.0.0.1:1", &[INTERNAL]).authenticate(None, None).await.unwrap_err(),
        AuthError::Unauthenticated
    );
}

#[actix_web::test]
async fn unreachable_user_service_is_unavailable() {
    let auth = authenticator("http://127.0.0.1:1", &[INTERNAL]);
    let err = auth.authenticate(Some(&format!("Bearer {}", token("7", "user"))), None).await.unwrap_err();
    assert_eq!(err, AuthError::Unavailable);
}

#[actix_web::test]
async fn extractor_maps_errors_to_status_codes() {
    let auth = web::Data::new(authenticator(&fake_user_service(), &[INTERNAL]));
    let app = atest::init_service(App::new().app_data(auth.clone()).route("/me", web::get().to(whoami))).await;

    let res = atest::call_service(&app, atest::TestRequest::get().uri("/me").to_request()).await;
    assert_eq!(res.status(), 401);

    let req = atest::TestRequest::get().uri("/me").insert_header((INTERNAL_TOKEN_HEADER, "wrong")).to_request();
    assert_eq!(atest::call_service(&app, req).await.status(), 401);

    let req = atest::TestRequest::get().uri("/me").insert_header((INTERNAL_TOKEN_HEADER, INTERNAL)).to_request();
    assert_eq!(atest::call_service(&app, req).await.status(), 200);

    let req = atest::TestRequest::get()
        .uri("/me")
        .insert_header(("Authorization", format!("Bearer {}", token("7", "user"))))
        .to_request();
    let body: serde_json::Value = atest::call_and_read_body_json(&app, req).await;
    assert_eq!(body["user_id"], "7");
    assert_eq!(body["admin"], false);
}

#[test]
fn ownership_rules() {
    let user = Principal { user_id: Some("7".into()), ..Default::default() };
    let admin = Principal { user_id: Some("1".into()), role: "admin".into(), ..Default::default() };

    assert!(authorize_owner(&user, Some(Some("7".into())), DomainError::BookingNotFound(1)).is_ok());
    let other = authorize_owner(&user, Some(Some("8".into())), DomainError::BookingNotFound(1)).unwrap_err();
    assert_eq!(other.as_response_error().status_code(), 403);
    let legacy = authorize_owner(&user, Some(None), DomainError::BookingNotFound(1)).unwrap_err();
    assert_eq!(legacy.as_response_error().status_code(), 403);
    let missing = authorize_owner(&user, None, DomainError::BookingNotFound(1)).unwrap_err();
    assert_eq!(missing.as_response_error().status_code(), 404);

    assert!(authorize_owner(&admin, Some(Some("8".into())), DomainError::BookingNotFound(1)).is_ok());
    assert!(authorize_owner(&Principal::service(), Some(None), DomainError::BookingNotFound(1)).is_ok());
}

#[test]
fn passenger_owner_cannot_be_spoofed() {
    let user = Principal { user_id: Some("7".into()), ..Default::default() };
    let admin = Principal { user_id: Some("1".into()), role: "admin".into(), ..Default::default() };

    assert_eq!(passenger_owner_for(&user, Some("8".into())).as_deref(), Some("7"));
    assert_eq!(passenger_owner_for(&admin, Some("8".into())).as_deref(), Some("8"));
    assert_eq!(passenger_owner_for(&admin, None).as_deref(), Some("1"));
    assert_eq!(passenger_owner_for(&Principal::service(), Some("9".into())).as_deref(), Some("9"));
}
