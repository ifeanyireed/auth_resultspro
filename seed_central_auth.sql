-- seed_central_auth.sql
-- Run this on your droplet: sqlite3 /var/lib/auth_resultspro/data/auth.db < seed_central_auth.sql

-- 1. Create/Update Test Users
-- These IDs match the ones seeded in ClassroomPRO for perfect SSO sync.

INSERT OR REPLACE INTO users (id, email, password_hash, auth_provider, full_name, phone, sex, date_of_birth, address, account_status)
VALUES 
-- Original Accounts
('bfb51c68-ccb0-401f-b58f-27fd41c6a856', 'superadmin@esultspro.ng', '$2a$14$1zhGRoc.lxuxyO/9X27HpuUTq06m5p2pb69PgYa0UWksEJWT7kS8i', 'local', 'Super Admin', NULL, NULL, NULL, NULL, 'active'),
('2db093ed-bdc9-47c4-b71c-66869f0f1ea7', 'teacher@example.edu', '$2a$14$Jg0JSBXO09zmMOssPyzEj.VyO/iuXai.QCZQFicC4CTR.plVD9dMS', 'local', 'Mr. Adeniyi', NULL, NULL, NULL, NULL, 'active'),
('111efa7d-e12d-4ed1-9902-d341c6826b50', 'student@example.com', '$2a$14$OiOxIN4UiEuFHKIhwdmFHuNbtI2FoVpU95KVD8Dc3FxLhHM2.EMve', 'local', 'Jane Doe', NULL, NULL, NULL, NULL, 'active'),
('dac38ffd-866f-47ab-8ac4-ecf6ea520ba8', 'parent@example.com', '$2a$14$4nofWUGNaOyx9/2zF23ySuu5ehgcPa1kApyvp5dLAHszuA.NoLOWS', 'local', 'Mrs. Doe', NULL, NULL, NULL, NULL, 'active'),
-- Requested Admin/Support Accounts
('8d3a7776-5d21-4f1e-9a6d-e4c1d63e9f01', 'platform-admin@resultspro.ng', '$2a$14$1zhGRoc.lxuxyO/9X27HpuUTq06m5p2pb69PgYa0UWksEJWT7kS8i', 'local', 'Platform Admin', NULL, NULL, NULL, NULL, 'active'),
('8d3a7776-5d21-4f1e-9a6d-e4c1d63e9f02', 'school-admin@example.edu', '$2a$14$1zhGRoc.lxuxyO/9X27HpuUTq06m5p2pb69PgYa0UWksEJWT7kS8i', 'local', 'School Admin', NULL, NULL, NULL, NULL, 'active'),
('8d3a7776-5d21-4f1e-9a6d-e4c1d63e9f03', 'support-staff@resultspro.ng', '$2a$14$1zhGRoc.lxuxyO/9X27HpuUTq06m5p2pb69PgYa0UWksEJWT7kS8i', 'local', 'Support Staff', NULL, NULL, NULL, NULL, 'active');

-- 2. Provision ClassroomPRO App
-- This allows ClassroomPRO to use the /auth/introspect endpoint securely.
INSERT OR REPLACE INTO apps (id, name, secret_key)
VALUES ('classroompro-app-id', 'ClassroomPRO', 'your-app-secret-key-123');

SELECT 'Central Auth Seeding Complete' as status;
