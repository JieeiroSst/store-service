-- Identity (users) is owned by user_service, not toggle-service — see
-- port.UserDirectory. Roles/permissions here are toggle-service's own
-- project-scoped RBAC, keyed by user_service's opaque user ID wherever a
-- user needs referencing (project_memberships.user_id, audit_events.user_id,
-- api_tokens.created_by), never by a local users table.

CREATE TABLE roles (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(100) NOT NULL UNIQUE,
    description TEXT NOT NULL DEFAULT ''
);

CREATE TABLE permissions (
    id   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL UNIQUE
);

CREATE TABLE role_permissions (
    role_id       UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id)
);
