-- seed_central_auth.sql
-- Run this on your droplet: sqlite3 /var/lib/auth_resultspro/data/auth.db < seed_central_auth.sql

-- 1. Create/Update Test Users
-- These IDs match the ones seeded in ClassroomPRO for perfect SSO sync.

INSERT OR REPLACE INTO users (id, email, password_hash, auth_provider, full_name, account_status)
VALUES 
('bfb51c68-ccb0-401f-b58f-27fd41c6a856', 'platform-admin@resultspro.ng', '$2a$14$1zhGRoc.lxuxyO/9X27HpuUTq06m5p2pb69PgYa0UWksEJWT7kS8i', 'local', 'Platform Admin', 'active'),
('2db093ed-bdc9-47c4-b71c-66869f0f1ea7', 'school-admin@example.edu', '$2a$14$1zhGRoc.lxuxyO/9X27HpuUTq06m5p2pb69PgYa0UWksEJWT7kS8i', 'local', 'School Admin', 'active'),
('111efa7d-e12d-4ed1-9902-d341c6826b50', 'support-staff@resultspro.ng', '$2a$14$1zhGRoc.lxuxyO/9X27HpuUTq06m5p2pb69PgYa0UWksEJWT7kS8i', 'local', 'Support Staff', 'active');

-- 2. Provision ClassroomPRO App
-- This allows ClassroomPRO to use the /auth/introspect endpoint securely.
INSERT OR REPLACE INTO apps (id, name, secret_key)
VALUES ('classroompro-app-id', 'ClassroomPRO', 'your-app-secret-key-123');

SELECT 'Central Auth Seeding Complete' as status;
