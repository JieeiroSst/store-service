INSERT INTO roles (name, description) VALUES
    ('Owner', 'Full control over the project, including membership and deletion'),
    ('Member', 'Can create/update flags and strategies, and toggle them, but cannot delete the project or manage members'),
    ('Viewer', 'Read-only access to the project')
ON CONFLICT (name) DO NOTHING;

INSERT INTO permissions (name) VALUES
    ('CREATE_FEATURE'),
    ('UPDATE_FEATURE'),
    ('DELETE_FEATURE'),
    ('TOGGLE_FEATURE'),
    ('CREATE_STRATEGY'),
    ('MANAGE_PROJECT'),
    ('MANAGE_MEMBERS'),
    ('MANAGE_TOKENS'),
    ('VIEW')
ON CONFLICT (name) DO NOTHING;

-- Owner: all permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r CROSS JOIN permissions p WHERE r.name = 'Owner'
ON CONFLICT DO NOTHING;

-- Member: everyday flag/strategy management, no delete/manage-project/manage-members/manage-tokens
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r CROSS JOIN permissions p
WHERE r.name = 'Member'
  AND p.name IN ('CREATE_FEATURE', 'UPDATE_FEATURE', 'TOGGLE_FEATURE', 'CREATE_STRATEGY', 'VIEW')
ON CONFLICT DO NOTHING;

-- Viewer: read-only
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r CROSS JOIN permissions p
WHERE r.name = 'Viewer' AND p.name = 'VIEW'
ON CONFLICT DO NOTHING;
