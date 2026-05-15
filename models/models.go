package models

import (
    "database/sql"
    "time"
)

type User struct {
        ID            string         `json:"id"`
        Email         string         `json:"email"`
        PasswordHash  sql.NullString `json:"-"`
        GoogleID      sql.NullString `json:"google_id,omitempty"`
        MicrosoftID   sql.NullString `json:"microsoft_id,omitempty"`
        AuthProvider  string         `json:"auth_provider"`
        FullName      sql.NullString `json:"full_name"`
        AvatarURL     sql.NullString `json:"avatar_url"`
        AccountStatus string         `json:"account_status"`
        MFAEnabled    bool           `json:"mfa_enabled"`
        MFASecret     sql.NullString `json:"-"`
        CreatedAt     time.Time      `json:"created_at"`
        UpdatedAt     time.Time      `json:"updated_at"`
}

type VerificationToken struct {
        ID        string    `json:"id"`
        UserID    string    `json:"user_id"`
        TokenHash string    `json:"-"`
        Type      string    `json:"type"`
        ExpiresAt time.Time `json:"expires_at"`
        Used      bool      `json:"used"`
}

type App struct {
        ID        string `json:"id"`
        Name      string `json:"name"`
        SecretKey string `json:"secret_key"`
}

type RefreshToken struct {
        ID         string         `json:"id"`
        UserID     string         `json:"user_id"`
        TokenHash  string         `json:"-"`
        DeviceInfo sql.NullString `json:"device_info"`
        ExpiresAt  time.Time      `json:"expires_at"`
        Revoked    bool           `json:"revoked"`
        CreatedAt  time.Time      `json:"created_at"`
}
