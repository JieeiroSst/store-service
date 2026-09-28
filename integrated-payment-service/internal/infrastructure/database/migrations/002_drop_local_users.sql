-- Users live in user-service; payments.user_id now holds the user-service id
-- of the authenticated caller, so the local users table and its FK go away.
ALTER TABLE payments DROP FOREIGN KEY payments_ibfk_1;
DROP TABLE IF EXISTS users;
