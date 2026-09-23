-- Adds app-wide admin role and records how each account was created.
-- Apply once to databases created before these columns existed in schema.sql.
ALTER TABLE users
    ADD COLUMN role ENUM('user', 'admin') NOT NULL DEFAULT 'user' AFTER avatar_path,
    ADD COLUMN auth_provider ENUM('password', 'google') NOT NULL DEFAULT 'password' AFTER role;

-- Promote the first administrator (edit the email as needed):
-- UPDATE users SET role = 'admin' WHERE email = 'admin@example.com';
