# Centralized Authentication System Blueprint

## Objective
Design and implement a centralized authentication service at `auth.resultspro.ng` to manage single sign-on (SSO), comprehensive user identities (OAuth and local credentials), and multi-device session management for a suite of applications (`*.resultspro.ng`).

## Core Technologies
*   **Backend:** Go (Golang) for high performance and concurrency.
*   **Database:** SQLite for lightweight, file-based relational storage.
*   **Authentication:** Google OAuth 2.0 and Local Email/Password (with bcrypt hashing).
*   **Session Management:** JWT (JSON Web Tokens) with a short-lived access token and long-lived refresh token strategy.

## Architecture & Security Model

### 1. Identity & Authentication
*   **Multiple Providers:** Support for both Google OAuth and Local Email/Password authentication.
*   **Account Status:** Users can be active, suspended, or unverified.
*   **Security:** Passwords hashed securely using bcrypt. Foundation laid for future MFA support.
*   **Verification:** Email verification via secure, time-limited tokens.

### 2. Token Strategy & Sessions
*   **Access Token (JWT):** Short-lived (e.g., 15 minutes). Contains minimal, non-sensitive claims (User ID). Used by sub-apps to authorize requests.
*   **Refresh Token:** Long-lived (e.g., 7 days) and stored securely. Used exclusively to obtain new Access Tokens.
*   **Multi-Device Sessions:** Users can log in from multiple devices. Each login issues a unique Refresh Token.

### 3. Token Revocation & Multi-Device Logout
*   **Stateful Revocation:** Immediate revocation of tokens when an account is compromised, suspended, or when a user logs out.
*   **Logout Scope:** 
    *   *Current Device:* Revokes the specific Refresh Token used for the request.
    *   *All Devices:* Revokes all Refresh Tokens associated with the User ID.
*   **Introspection Endpoint:** Sub-apps query this to check if an Access Token is valid and the user's account is still active.

---

## Database Schema (SQLite)

```sql
-- Users Table
CREATE TABLE users (
    id TEXT PRIMARY KEY,          -- Global User ID (UUID)
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT,           -- NULL if Google-only
    google_id TEXT UNIQUE,        -- NULL if local-only
    auth_provider TEXT NOT NULL,  -- 'google', 'local', or 'both'
    full_name TEXT,
    avatar_url TEXT,
    account_status TEXT DEFAULT 'unverified', -- 'unverified', 'active', 'suspended'
    mfa_enabled BOOLEAN DEFAULT FALSE,
    mfa_secret TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Verification & Password Reset Tokens
CREATE TABLE verification_tokens (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    token_hash TEXT NOT NULL,
    type TEXT NOT NULL,           -- 'email_verify', 'password_reset'
    expires_at DATETIME NOT NULL,
    used BOOLEAN DEFAULT FALSE,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

-- Apps Table
CREATE TABLE apps (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    secret_key TEXT NOT NULL
);

-- Refresh Tokens / Sessions (Multi-device support)
CREATE TABLE refresh_tokens (
    id TEXT PRIMARY KEY,          -- Session ID
    user_id TEXT NOT NULL,
    token_hash TEXT NOT NULL,
    device_info TEXT,             -- Optional: Store User-Agent or IP for session management UI
    expires_at DATETIME NOT NULL,
    revoked BOOLEAN DEFAULT FALSE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);
```

---

## API Endpoints

### Identity Management
*   `POST /auth/signup` - Email/password registration.
*   `POST /auth/login` - Email/password authentication.
*   `GET /auth/google` & `/callback` - Google OAuth flow.
*   `POST /auth/verify-email` - Validate email address using token.
*   `POST /auth/forgot-password` - Request password reset email.
*   `POST /auth/reset-password` - Set new password using token.
*   `POST /auth/change-password` - Authenticated endpoint to change password.
*   `POST /auth/change-email` - Authenticated endpoint to change email (requires re-verification).

### Session Management
*   `POST /auth/refresh` - Exchange valid refresh token for new access token.
*   `POST /auth/logout` - Revoke current refresh token.
*   `POST /auth/logout-all` - Revoke all refresh tokens for the authenticated user.
*   `POST /auth/introspect` - Sub-app endpoint to validate access token and check account status.

---

## Implementation Steps

### Phase 1: Local Auth & Data Model Updates
1.  Update SQLite schema to include new tables and columns.
2.  Implement `bcrypt` for password hashing.
3.  Implement signup and login handlers.
4.  Update Google callback to handle merging or distinguishing local/Google accounts.

### Phase 2: Session Enhancements
1.  Update refresh token logic to support multi-device tracking.
2.  Implement logout-all functionality.
3.  Update introspection to check account status (reject if suspended/unverified).

### Phase 3: Account Recovery & Management
1.  Implement secure token generation for verification/resets.
2.  Implement email verification handlers.
3.  Implement password reset and change handlers.
4.  (Optional) Add dummy email sending function to be replaced with real SMTP/API later.

### Phase 4: MFA Foundation (Optional later)
1.  Add MFA fields to schema (already included).
2.  Prepare endpoints for MFA setup/verification.

---

## Deployment Process (DigitalOcean Droplet with Docker)

Since you are running multiple apps on a single Droplet, we use **Docker Compose** to isolate the auth service while mounting the SQLite database to a persistent external directory.

### 1. Host Preparation
On your Droplet, create the persistent directory for the database:
```bash
sudo mkdir -p /var/lib/resultspro/auth
sudo chown -R $USER:$USER /var/lib/resultspro/auth
```

### 2. Environment Configuration
Ensure your `.env` file on the server has the `DB_PATH` set to the *internal* container path:
```bash
DB_PATH=/app/data/auth.db
```

### 3. Docker Compose Setup
The `docker-compose.yml` file handles the build and volume mapping:
```yaml
version: '3.8'
services:
  auth-service:
    build: .
    container_name: auth-resultspro
    restart: always
    ports:
      - "8080:8080"
    env_file: .env
    volumes:
      - /var/lib/resultspro/auth:/app/data
```

### 4. Launch
```bash
docker compose up -d --build
```

### 5. Nginx Reverse Proxy (Optional)
To point `auth.resultspro.ng` to this service, add an Nginx config:
```nginx
server {
    server_name auth.resultspro.ng;
    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```
Use `certbot` to enable HTTPS.
