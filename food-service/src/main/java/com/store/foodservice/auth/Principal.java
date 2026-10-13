package com.store.foodservice.auth;

import java.util.List;
import java.util.Set;

public record Principal(String userId, String role, List<String> roles, boolean service) {
    private static final Set<String> ADMIN_ROLES = Set.of("admin", "super_admin");

    public static Principal internal() {
        return new Principal("", "", List.of(), true);
    }

    public boolean isAdmin() {
        return service || ADMIN_ROLES.contains(role) || roles.stream().anyMatch(ADMIN_ROLES::contains);
    }
}
