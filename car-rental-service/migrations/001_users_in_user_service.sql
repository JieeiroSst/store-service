-- Users moved to user-service. Converts every user reference from the old
-- local UUIDs to user-service ids (BIGINT), keeps driving licenses in
-- customer_profiles, and drops the local users table (password hashes).
--
-- Before running, fill legacy_user_map with each old user's user-service id,
-- e.g. by matching emails between an export of users and user-service:
--
--   CREATE TABLE legacy_user_map (legacy_id UUID PRIMARY KEY, user_id BIGINT NOT NULL);
--
-- The migration aborts if any referenced user has no mapping, so no rental
-- history is silently orphaned.
BEGIN;

DO $$
DECLARE missing INT;
BEGIN
    SELECT count(*) INTO missing FROM users u
    WHERE NOT EXISTS (SELECT 1 FROM legacy_user_map m WHERE m.legacy_id = u.user_id);
    IF missing > 0 THEN
        RAISE EXCEPTION '% users have no user-service id in legacy_user_map', missing;
    END IF;
END $$;

CREATE TABLE IF NOT EXISTS customer_profiles (
    user_id BIGINT PRIMARY KEY,
    driving_license VARCHAR(50),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
INSERT INTO customer_profiles (user_id, driving_license, created_at, updated_at)
SELECT m.user_id, u.driving_license, u.created_at, u.updated_at
FROM users u JOIN legacy_user_map m ON m.legacy_id = u.user_id
ON CONFLICT (user_id) DO NOTHING;

-- Re-point user_id on every table that referenced users(user_id).
DO $$
DECLARE t TEXT;
BEGIN
    FOREACH t IN ARRAY ARRAY['reservations', 'rentals', 'payments', 'reviews', 'user_documents'] LOOP
        EXECUTE format('ALTER TABLE %I DROP CONSTRAINT IF EXISTS %I', t, t || '_user_id_fkey');
        EXECUTE format('ALTER TABLE %I ADD COLUMN user_id_new BIGINT', t);
        EXECUTE format('UPDATE %I x SET user_id_new = m.user_id FROM legacy_user_map m WHERE m.legacy_id = x.user_id', t);
        EXECUTE format('ALTER TABLE %I DROP COLUMN user_id', t);
        EXECUTE format('ALTER TABLE %I RENAME COLUMN user_id_new TO user_id', t);
    END LOOP;
END $$;

ALTER TABLE user_documents DROP CONSTRAINT IF EXISTS user_documents_verified_by_fkey;
ALTER TABLE user_documents ADD COLUMN verified_by_new BIGINT;
UPDATE user_documents d SET verified_by_new = m.user_id
FROM legacy_user_map m WHERE m.legacy_id = d.verified_by;
ALTER TABLE user_documents DROP COLUMN verified_by;
ALTER TABLE user_documents RENAME COLUMN verified_by_new TO verified_by;

DROP TABLE users;
DROP TABLE legacy_user_map;

COMMIT;
