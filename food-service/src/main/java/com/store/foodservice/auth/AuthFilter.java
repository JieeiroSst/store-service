package com.store.foodservice.auth;

import jakarta.servlet.FilterChain;
import jakarta.servlet.ServletException;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;
import java.io.IOException;
import java.util.Optional;
import java.util.Set;
import org.springframework.web.filter.OncePerRequestFilter;

public class AuthFilter extends OncePerRequestFilter {
    public static final String PRINCIPAL_ATTRIBUTE = "authn.principal";
    private static final Set<String> READ_METHODS = Set.of("GET", "HEAD", "OPTIONS");

    private final Authenticator authenticator;

    public AuthFilter(Authenticator authenticator) {
        this.authenticator = authenticator;
    }

    @Override
    protected void doFilterInternal(HttpServletRequest request, HttpServletResponse response, FilterChain chain)
            throws ServletException, IOException {
        try {
            Optional<Principal> principal = authenticator.authenticate(
                    request.getHeader("Authorization"),
                    request.getHeader(Authenticator.INTERNAL_TOKEN_HEADER));
            principal.ifPresent(p -> request.setAttribute(PRINCIPAL_ATTRIBUTE, p));
            if (!READ_METHODS.contains(request.getMethod())) {
                Principal p = principal.orElseThrow(AuthException::unauthenticated);
                if (!p.isAdmin()) {
                    throw AuthException.forbidden();
                }
            }
        } catch (AuthException e) {
            response.setStatus(e.status());
            response.setContentType("application/json");
            response.getWriter().write("{\"error\":\"" + e.getMessage() + "\"}");
            return;
        }
        chain.doFilter(request, response);
    }
}
