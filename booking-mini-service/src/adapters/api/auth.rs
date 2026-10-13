use actix_web::dev::Payload;
use actix_web::http::StatusCode;
use actix_web::{web, FromRequest, HttpRequest, HttpResponse, ResponseError};
use base64::Engine;
use serde::Deserialize;
use std::collections::HashMap;
use std::fmt;
use std::future::Future;
use std::pin::Pin;
use std::sync::{Mutex, Once};
use std::time::{Duration, Instant};

pub const INTERNAL_TOKEN_HEADER: &str = "X-Internal-Token";
const IDENTITY_TTL: Duration = Duration::from_secs(30);
const DEFAULT_USER_SERVICE: &str = "http://user-api-svc";
const MAX_CACHE: usize = 10_000;

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum AuthError {
    Unauthenticated,
    Forbidden,
    Unavailable,
}

impl fmt::Display for AuthError {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        f.write_str(match self {
            AuthError::Unauthenticated => "authentication required",
            AuthError::Forbidden => "forbidden",
            AuthError::Unavailable => "authentication service unavailable",
        })
    }
}

impl std::error::Error for AuthError {}

impl ResponseError for AuthError {
    fn status_code(&self) -> StatusCode {
        match self {
            AuthError::Unauthenticated => StatusCode::UNAUTHORIZED,
            AuthError::Forbidden => StatusCode::FORBIDDEN,
            AuthError::Unavailable => StatusCode::SERVICE_UNAVAILABLE,
        }
    }

    fn error_response(&self) -> HttpResponse {
        HttpResponse::build(self.status_code()).json(serde_json::json!({ "error": self.to_string() }))
    }
}

#[derive(Debug, Clone, Default, PartialEq, Eq)]
pub struct Principal {
    pub user_id: Option<String>,
    pub role: String,
    pub roles: Vec<String>,
    pub service: bool,
}

impl Principal {
    pub fn service() -> Self {
        Principal { service: true, ..Default::default() }
    }

    pub fn is_admin(&self) -> bool {
        self.service
            || std::iter::once(&self.role)
                .chain(self.roles.iter())
                .any(|r| r == "admin" || r == "super_admin")
    }

    pub fn owns(&self, owner: Option<&str>) -> bool {
        if self.is_admin() {
            return true;
        }
        matches!((self.user_id.as_deref(), owner), (Some(me), Some(owner)) if !me.is_empty() && me == owner)
    }

    pub fn require_admin(&self) -> Result<(), AuthError> {
        if self.is_admin() { Ok(()) } else { Err(AuthError::Forbidden) }
    }

    pub fn require_owner(&self, owner: Option<&str>) -> Result<(), AuthError> {
        if self.owns(owner) { Ok(()) } else { Err(AuthError::Forbidden) }
    }
}

pub struct Authenticator {
    base_url: String,
    client: reqwest::Client,
    internal: Vec<String>,
    cache: Mutex<HashMap<String, (Principal, Instant)>>,
    warn: Once,
}

impl Authenticator {
    pub fn new(user_service_url: &str, internal_tokens: &[String], timeout: Duration) -> Self {
        let client = reqwest::Client::builder()
            .timeout(timeout)
            .build()
            .unwrap_or_else(|_| reqwest::Client::new());
        Authenticator {
            base_url: user_service_url.trim_end_matches('/').to_string(),
            client,
            internal: internal_tokens
                .iter()
                .map(|t| t.trim().to_string())
                .filter(|t| !t.is_empty())
                .collect(),
            cache: Mutex::new(HashMap::new()),
            warn: Once::new(),
        }
    }

    pub fn from_env() -> Self {
        let url = std::env::var("USER_SERVICE_URL")
            .ok()
            .filter(|u| !u.trim().is_empty())
            .unwrap_or_else(|| DEFAULT_USER_SERVICE.to_string());
        let tokens: Vec<String> = std::env::var("INTERNAL_API_TOKENS")
            .unwrap_or_default()
            .split(',')
            .map(str::to_string)
            .collect();
        Self::new(&url, &tokens, Duration::from_secs(3))
    }

    pub fn enforcing(&self) -> bool {
        !self.internal.is_empty()
    }

    pub async fn authenticate(&self, authorization: Option<&str>, internal_token: Option<&str>) -> Result<Principal, AuthError> {
        if let Some(token) = internal_token.map(str::trim).filter(|t| !t.is_empty()) {
            return if self.valid_internal(token) {
                Ok(Principal::service())
            } else {
                Err(AuthError::Unauthenticated)
            };
        }
        if let Some(token) = authorization
            .map(str::trim)
            .and_then(|h| h.strip_prefix("Bearer "))
            .map(str::trim)
            .filter(|t| !t.is_empty())
        {
            return self.user(token).await;
        }
        if !self.enforcing() {
            self.warn.call_once(|| {
                log::warn!("INTERNAL_API_TOKENS is not set, requests without credentials are accepted as internal calls")
            });
            return Ok(Principal::service());
        }
        Err(AuthError::Unauthenticated)
    }

    fn valid_internal(&self, token: &str) -> bool {
        self.internal
            .iter()
            .fold(false, |found, want| constant_time_eq(token.as_bytes(), want.as_bytes()) | found)
    }

    async fn user(&self, token: &str) -> Result<Principal, AuthError> {
        if let Some((p, expires)) = self.cache.lock().unwrap().get(token) {
            if Instant::now() < *expires {
                return Ok(p.clone());
            }
        }
        let user_id = self.validate(token).await?;
        let claims = read_claims(token);
        let principal = Principal {
            user_id: Some(user_id),
            role: claims.role.unwrap_or_default(),
            roles: claims.roles.unwrap_or_default(),
            service: false,
        };
        let mut cache = self.cache.lock().unwrap();
        if cache.len() > MAX_CACHE {
            cache.clear();
        }
        cache.insert(token.to_string(), (principal.clone(), Instant::now() + IDENTITY_TTL));
        Ok(principal)
    }

    async fn validate(&self, token: &str) -> Result<String, AuthError> {
        let resp = self
            .client
            .post(format!("{}/api/v1/validate", self.base_url))
            .json(&serde_json::json!({ "session_token": token }))
            .send()
            .await
            .map_err(|e| {
                log::error!("validate session: {e}");
                AuthError::Unavailable
            })?;
        match resp.status().as_u16() {
            200 => {}
            401 | 403 => return Err(AuthError::Unauthenticated),
            status => {
                log::error!("validate session returned {status}");
                return Err(AuthError::Unavailable);
            }
        }
        let body: serde_json::Value = resp.json().await.map_err(|e| {
            log::error!("decode validate response: {e}");
            AuthError::Unavailable
        })?;
        let valid = body.get("valid").and_then(|v| v.as_bool()).unwrap_or(false);
        let user_id = ["user_id", "userId"]
            .iter()
            .filter_map(|k| body.get(*k))
            .find_map(|v| match v {
                serde_json::Value::String(s) if !s.is_empty() => Some(s.clone()),
                serde_json::Value::Number(n) => Some(n.to_string()),
                _ => None,
            });
        match (valid, user_id) {
            (true, Some(id)) => Ok(id),
            _ => Err(AuthError::Unauthenticated),
        }
    }
}

impl FromRequest for Principal {
    type Error = actix_web::Error;
    type Future = Pin<Box<dyn Future<Output = Result<Self, Self::Error>>>>;

    fn from_request(req: &HttpRequest, _: &mut Payload) -> Self::Future {
        let auth = req.app_data::<web::Data<Authenticator>>().cloned();
        let header = |name: &str| {
            req.headers()
                .get(name)
                .and_then(|v| v.to_str().ok())
                .map(str::to_string)
        };
        let authorization = header("Authorization");
        let internal = header(INTERNAL_TOKEN_HEADER);
        Box::pin(async move {
            let auth = auth.ok_or_else(|| {
                log::error!("authenticator is not configured");
                AuthError::Unavailable
            })?;
            Ok(auth.authenticate(authorization.as_deref(), internal.as_deref()).await?)
        })
    }
}

#[derive(Default, Deserialize)]
struct Claims {
    role: Option<String>,
    roles: Option<Vec<String>>,
}

fn read_claims(token: &str) -> Claims {
    let parts: Vec<&str> = token.split('.').collect();
    if parts.len() != 3 {
        return Claims::default();
    }
    base64::engine::general_purpose::URL_SAFE_NO_PAD
        .decode(parts[1].trim_end_matches('='))
        .ok()
        .and_then(|payload| serde_json::from_slice(&payload).ok())
        .unwrap_or_default()
}

fn constant_time_eq(a: &[u8], b: &[u8]) -> bool {
    if a.len() != b.len() {
        return false;
    }
    a.iter().zip(b).fold(0u8, |acc, (x, y)| acc | (x ^ y)) == 0
}
