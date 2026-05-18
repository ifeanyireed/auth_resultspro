-- Run this script on your production database using:
-- sqlite3 /var/www/auth_resultspro/auth.db < fix_db.sql

-- Fix mfa_enabled boolean values (modernc.org/sqlite might have stored them as strings or Go bools if not careful)
UPDATE users SET mfa_enabled = 0 WHERE mfa_enabled = 'false' OR mfa_enabled IS NULL;
UPDATE users SET mfa_enabled = 1 WHERE mfa_enabled = 'true';

-- Fix Date formats by replacing the space with a 'T' and adding 'Z' if needed.
-- Prisma strictly requires RFC3339 formatted dates (e.g., 2024-05-16T00:00:00.000Z)

-- Users table
UPDATE users SET created_at = replace(created_at, ' ', 'T') WHERE created_at LIKE '% %' AND created_at NOT LIKE '%T%';
UPDATE users SET updated_at = replace(updated_at, ' ', 'T') WHERE updated_at LIKE '% %' AND updated_at NOT LIKE '%T%';
UPDATE users SET date_of_birth = replace(date_of_birth, ' ', 'T') WHERE date_of_birth LIKE '% %' AND date_of_birth NOT LIKE '%T%';

UPDATE users SET created_at = replace(created_at, '+00:00', 'Z') WHERE created_at LIKE '%+00:00';
UPDATE users SET updated_at = replace(updated_at, '+00:00', 'Z') WHERE updated_at LIKE '%+00:00';
UPDATE users SET date_of_birth = replace(date_of_birth, '+00:00', 'Z') WHERE date_of_birth LIKE '%+00:00';

-- Verification tokens table
UPDATE verification_tokens SET expires_at = replace(expires_at, ' ', 'T') WHERE expires_at LIKE '% %' AND expires_at NOT LIKE '%T%';
UPDATE verification_tokens SET expires_at = replace(expires_at, '+00:00', 'Z') WHERE expires_at LIKE '%+00:00';
UPDATE verification_tokens SET used = 0 WHERE used = 'false' OR used IS NULL;
UPDATE verification_tokens SET used = 1 WHERE used = 'true';

-- Refresh tokens table
UPDATE refresh_tokens SET expires_at = replace(expires_at, ' ', 'T') WHERE expires_at LIKE '% %' AND expires_at NOT LIKE '%T%';
UPDATE refresh_tokens SET created_at = replace(created_at, ' ', 'T') WHERE created_at LIKE '% %' AND created_at NOT LIKE '%T%';
UPDATE refresh_tokens SET expires_at = replace(expires_at, '+00:00', 'Z') WHERE expires_at LIKE '%+00:00';
UPDATE refresh_tokens SET created_at = replace(created_at, '+00:00', 'Z') WHERE created_at LIKE '%+00:00';
UPDATE refresh_tokens SET revoked = 0 WHERE revoked = 'false' OR revoked IS NULL;
UPDATE refresh_tokens SET revoked = 1 WHERE revoked = 'true';
