# Centralized Authentication System Blueprint

## Objective
Design and implement a centralized authentication service at `auth.resultspro.ng` to manage single sign-on (SSO), comprehensive user identities (OAuth and local credentials), and multi-device session management for a suite of applications (`*.resultspro.ng`).

## Core Technologies
*   **Backend:** Go (Golang) for high performance and concurrency.
*   **Database:** SQLite for lightweight, file-based relational storage.
*   **Authentication:** Google OAuth 2.0, Microsoft OAuth 2.0, and Local Email/Password (with bcrypt hashing).
*   **MFA:** TOTP-based Multi-Factor Authentication.
*   **Session Management:** JWT (JSON Web Tokens) with a short-lived access token and long-lived refresh token strategy.

## Architecture & Security Model

### 1. Identity & Authentication
*   **Multiple Providers:** Support for Google, Microsoft, and Local Email/Password authentication.
*   **Account Status:** Users can be active, suspended, or unverified.
*   **Security:** Passwords hashed securely using bcrypt.
*   **Verification:** Email verification via secure, time-limited tokens.
*   **MFA (TOTP)**: Secure secondary verification layer using standard authenticator apps.

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
    password_hash TEXT,           -- NULL if OAuth-only
    google_id TEXT UNIQUE,        -- NULL if not linked
    microsoft_id TEXT UNIQUE,     -- NULL if not linked
    auth_provider TEXT NOT NULL,  -- 'google', 'microsoft', 'local', or 'mixed'
    full_name TEXT,
    avatar_url TEXT,
    account_status TEXT DEFAULT 'unverified', -- 'unverified', 'active', 'suspended'
    mfa_enabled BOOLEAN DEFAULT FALSE,
    mfa_secret TEXT,              -- TOTP Secret
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
    device_info TEXT,             -- Store User-Agent
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
*   `GET /auth/microsoft` & `/callback/microsoft` - Microsoft OAuth flow.
*   `POST /auth/verify-email` - Validate email address using token.
*   `POST /auth/forgot-password` - Request password reset email.
*   `POST /auth/reset-password` - Set new password using token.
*   `POST /auth/change-password` - Authenticated endpoint to change password.
*   `POST /auth/change-email` - Authenticated endpoint to change email.

### MFA Management
*   `POST /auth/mfa/setup` - Generate TOTP secret.
*   `POST /auth/mfa/verify` - Verify and enable MFA.
*   `POST /auth/mfa/challenge` - Complete login with TOTP code.
*   `POST /auth/mfa/disable` - Disable MFA with code.

### Session Management
*   `POST /auth/refresh` - Exchange valid refresh token for new access token.
*   `POST /auth/logout` - Revoke current refresh token.
*   `POST /auth/logout-all` - Revoke all refresh tokens for the authenticated user.
*   `POST /auth/introspect` - Sub-app endpoint to validate access token and check account status.

---

## Deployment Process (DigitalOcean Droplet with Docker)

### 1. Host Preparation
```bash
# Create persistent DB directory
sudo mkdir -p /var/lib/auth_resultspro/data
sudo chown -R $USER:$USER /var/lib/auth_resultspro/data

# Setup app directory (optional if using git clone directly)
# sudo mkdir -p /var/www/auth_resultspro
```

### 2. Launch
```bash
# Assuming code is in /var/www/auth_resultspro
cd /var/www/auth_resultspro
docker compose up -d --build
```

### 3. Nginx Reverse Proxy
```nginx
server {
    server_name auth.resultspro.ng;
    location / {
        proxy_pass http://localhost:7000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```
