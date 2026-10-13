package com.store.foodservice.auth;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertTrue;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.sun.net.httpserver.HttpServer;
import java.net.InetSocketAddress;
import java.nio.charset.StandardCharsets;
import java.util.Base64;
import java.util.List;
import java.util.Map;
import java.util.concurrent.atomic.AtomicBoolean;
import org.junit.jupiter.api.AfterAll;
import org.junit.jupiter.api.BeforeAll;
import org.junit.jupiter.api.Test;
import org.springframework.mock.web.MockHttpServletRequest;
import org.springframework.mock.web.MockHttpServletResponse;

class AuthFilterTest {
    private static HttpServer userService;
    private static String url;
    private static final ObjectMapper JSON = new ObjectMapper();

    static String token(String sub, String role) {
        String payload = Base64.getUrlEncoder().withoutPadding()
                .encodeToString(("{\"sub\":\"" + sub + "\",\"role\":\"" + role + "\"}").getBytes(StandardCharsets.UTF_8));
        return "h." + payload + ".s";
    }

    @BeforeAll
    static void start() throws Exception {
        userService = HttpServer.create(new InetSocketAddress("127.0.0.1", 0), 0);
        userService.createContext("/api/v1/validate", ex -> {
            Map<?, ?> in = JSON.readValue(ex.getRequestBody(), Map.class);
            String[] parts = String.valueOf(in.get("session_token")).split("\\.");
            String sub = "";
            if (parts.length == 3) {
                sub = String.valueOf(JSON.readValue(Base64.getUrlDecoder().decode(parts[1]), Map.class).get("sub"));
            }
            byte[] out = JSON.writeValueAsBytes(Map.of("valid", !sub.isEmpty(), "user_id", sub));
            ex.sendResponseHeaders(200, out.length);
            ex.getResponseBody().write(out);
            ex.close();
        });
        userService.start();
        url = "http://127.0.0.1:" + userService.getAddress().getPort();
    }

    @AfterAll
    static void stop() {
        userService.stop(0);
    }

    int run(AuthFilter filter, String method, String bearer, String internal, AtomicBoolean reached) throws Exception {
        MockHttpServletRequest req = new MockHttpServletRequest(method, "/api/food");
        if (bearer != null) {
            req.addHeader("Authorization", "Bearer " + bearer);
        }
        if (internal != null) {
            req.addHeader(Authenticator.INTERNAL_TOKEN_HEADER, internal);
        }
        MockHttpServletResponse res = new MockHttpServletResponse();
        reached.set(false);
        filter.doFilter(req, res, (rq, rs) -> reached.set(true));
        return res.getStatus();
    }

    @Test
    void enforcesRules() throws Exception {
        AuthFilter filter = new AuthFilter(new Authenticator(url, List.of("s3cret")));
        AtomicBoolean reached = new AtomicBoolean();

        assertEquals(200, run(filter, "GET", null, null, reached));
        assertTrue(reached.get());
        assertEquals(401, run(filter, "POST", null, null, reached));
        assertFalse(reached.get());
        assertEquals(403, run(filter, "POST", token("7", "user"), null, reached));
        assertFalse(reached.get());
        assertEquals(200, run(filter, "DELETE", token("1", "admin"), null, reached));
        assertTrue(reached.get());
        assertEquals(200, run(filter, "PUT", null, "s3cret", reached));
        assertTrue(reached.get());
        assertEquals(401, run(filter, "PUT", null, "wrong", reached));
        assertEquals(401, run(filter, "GET", "nope", null, reached));
    }

    @Test
    void rolloutModeAcceptsAnonymousInternalCalls() throws Exception {
        AuthFilter filter = new AuthFilter(new Authenticator(url, List.of()));
        AtomicBoolean reached = new AtomicBoolean();
        assertEquals(200, run(filter, "POST", null, null, reached));
        assertTrue(reached.get());
        assertEquals(403, run(filter, "POST", token("7", "user"), null, reached));
    }
}
