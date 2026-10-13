package com.store.foodservice.auth;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.nio.charset.StandardCharsets;
import java.security.MessageDigest;
import java.time.Duration;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.Base64;
import java.util.List;
import java.util.Map;
import java.util.Optional;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.atomic.AtomicBoolean;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

public class Authenticator {
    public static final String INTERNAL_TOKEN_HEADER = "X-Internal-Token";
    private static final long IDENTITY_TTL_MS = 30_000;
    private static final Logger log = LoggerFactory.getLogger(Authenticator.class);

    private record Cached(Principal principal, long expires) {
    }

    private final String userServiceUrl;
    private final List<byte[]> internalTokens;
    private final HttpClient http;
    private final ObjectMapper json = new ObjectMapper();
    private final Map<String, Cached> cache = new ConcurrentHashMap<>();
    private final AtomicBoolean warned = new AtomicBoolean();

    public Authenticator(String userServiceUrl, List<String> internalTokens) {
        this.userServiceUrl = userServiceUrl.replaceAll("/+$", "");
        this.internalTokens = internalTokens.stream()
                .map(String::trim)
                .filter(t -> !t.isEmpty())
                .map(t -> t.getBytes(StandardCharsets.UTF_8))
                .toList();
        this.http = HttpClient.newBuilder().connectTimeout(Duration.ofSeconds(3)).build();
    }

    public static Authenticator fromEnv() {
        String url = Optional.ofNullable(System.getenv("USER_SERVICE_URL")).filter(s -> !s.isBlank()).orElse("http://user-api-svc");
        String tokens = Optional.ofNullable(System.getenv("INTERNAL_API_TOKENS")).orElse("");
        return new Authenticator(url, Arrays.asList(tokens.split(",")));
    }

    public boolean enforcing() {
        return !internalTokens.isEmpty();
    }

    public Optional<Principal> authenticate(String authorization, String internalToken) {
        String internal = internalToken == null ? "" : internalToken.trim();
        if (!internal.isEmpty()) {
            byte[] got = internal.getBytes(StandardCharsets.UTF_8);
            boolean match = false;
            for (byte[] want : internalTokens) {
                match |= MessageDigest.isEqual(got, want);
            }
            if (match) {
                return Optional.of(Principal.internal());
            }
            throw AuthException.unauthenticated();
        }
        String auth = authorization == null ? "" : authorization.trim();
        if (auth.startsWith("Bearer ") && !auth.substring(7).isBlank()) {
            return Optional.of(user(auth.substring(7).trim()));
        }
        if (!enforcing()) {
            if (warned.compareAndSet(false, true)) {
                log.warn("INTERNAL_API_TOKENS is not set, requests without credentials are accepted as internal calls");
            }
            return Optional.of(Principal.internal());
        }
        return Optional.empty();
    }

    private Principal user(String token) {
        Cached hit = cache.get(token);
        long now = System.currentTimeMillis();
        if (hit != null && hit.expires() > now) {
            return hit.principal();
        }
        String userId = validate(token);
        JsonNode claims = claims(token);
        List<String> roles = new ArrayList<>();
        claims.path("roles").forEach(r -> roles.add(r.asText()));
        Principal p = new Principal(userId, claims.path("role").asText(""), roles, false);
        if (cache.size() > 10_000) {
            cache.clear();
        }
        cache.put(token, new Cached(p, now + IDENTITY_TTL_MS));
        return p;
    }

    private String validate(String token) {
        HttpResponse<String> res;
        try {
            String body = json.writeValueAsString(Map.of("session_token", token));
            HttpRequest req = HttpRequest.newBuilder(URI.create(userServiceUrl + "/api/v1/validate"))
                    .timeout(Duration.ofSeconds(3))
                    .header("Content-Type", "application/json")
                    .POST(HttpRequest.BodyPublishers.ofString(body))
                    .build();
            res = http.send(req, HttpResponse.BodyHandlers.ofString());
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
            throw AuthException.unavailable();
        } catch (Exception e) {
            log.error("validate session: {}", e.toString());
            throw AuthException.unavailable();
        }
        if (res.statusCode() == 401 || res.statusCode() == 403) {
            throw AuthException.unauthenticated();
        }
        if (res.statusCode() / 100 != 2) {
            log.error("validate session returned {}", res.statusCode());
            throw AuthException.unavailable();
        }
        try {
            JsonNode out = json.readTree(res.body());
            String userId = out.path("user_id").asText(out.path("userId").asText(""));
            if (!out.path("valid").asBoolean(false) || userId.isEmpty()) {
                throw AuthException.unauthenticated();
            }
            return userId;
        } catch (AuthException e) {
            throw e;
        } catch (Exception e) {
            log.error("decode validate response: {}", e.toString());
            throw AuthException.unavailable();
        }
    }

    private JsonNode claims(String token) {
        String[] parts = token.split("\\.");
        if (parts.length != 3) {
            return json.createObjectNode();
        }
        try {
            return json.readTree(Base64.getUrlDecoder().decode(parts[1]));
        } catch (Exception e) {
            return json.createObjectNode();
        }
    }
}
