package models

import (
	"time"
)

type User struct {
	ID            string     `json:"id"`
	Email         string     `json:"email"`
	PasswordHash  *string    `json:"-"`
	GoogleID      *string    `json:"google_id,omitempty"`
	MicrosoftID   *string    `json:"microsoft_id,omitempty"`
	AuthProvider  string     `json:"auth_provider"`
	FullName      *string    `json:"full_name"`
	AvatarURL     *string    `json:"avatar_url"`
	Phone         *string    `json:"phone,omitempty"`
	Sex           *string    `json:"sex,omitempty"`
	DateOfBirth   *time.Time `json:"date_of_birth,omitempty"`
	Address       *string    `json:"address,omitempty"`
	AccountStatus string     `json:"account_status"`
	MFAEnabled    bool       `json:"mfa_enabled"`
	MFASecret     *string    `json:"-"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
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
	ID         string    `json:"id"`
	UserID     string    `json:"user_id"`
	TokenHash  string    `json:"-"`
	DeviceInfo *string   `json:"device_info"`
	ExpiresAt  time.Time `json:"expires_at"`
	Revoked    bool      `json:"revoked"`
	CreatedAt  time.Time `json:"created_at"`
}
