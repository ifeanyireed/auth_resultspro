-- MySQL Schema for Central Auth Service

CREATE TABLE IF NOT EXISTS users (
    id VARCHAR(191) PRIMARY KEY,
    email VARCHAR(191) UNIQUE NOT NULL,
    password_hash TEXT,
    google_id VARCHAR(191) UNIQUE,
    microsoft_id VARCHAR(191) UNIQUE,
    auth_provider VARCHAR(191) NOT NULL,
    full_name VARCHAR(191),
    avatar_url TEXT,
    phone VARCHAR(191),
    sex VARCHAR(191),
    date_of_birth DATETIME(3),
    address TEXT,
    account_status VARCHAR(191) DEFAULT 'unverified',
    mfa_enabled BOOLEAN DEFAULT FALSE,
    mfa_secret TEXT,
    created_at DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3),
    updated_at DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS verification_tokens (
    id VARCHAR(191) PRIMARY KEY,
    user_id VARCHAR(191) NOT NULL,
    token_hash VARCHAR(191) NOT NULL,
    type VARCHAR(191) NOT NULL,
    expires_at DATETIME(3) NOT NULL,
    used BOOLEAN DEFAULT FALSE,
    CONSTRAINT fk_verification_token_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS apps (
    id VARCHAR(191) PRIMARY KEY,
    name VARCHAR(191) NOT NULL,
    secret_key VARCHAR(191) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS refresh_tokens (
    id VARCHAR(191) PRIMARY KEY,
    user_id VARCHAR(191) NOT NULL,
    token_hash VARCHAR(191) NOT NULL,
    device_info TEXT,
    expires_at DATETIME(3) NOT NULL,
    revoked BOOLEAN DEFAULT FALSE,
    created_at DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3),
    CONSTRAINT fk_refresh_token_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
