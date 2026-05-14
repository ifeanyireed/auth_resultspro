package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"auth.resultspro.ng/db"
	"auth.resultspro.ng/utils"
	"github.com/google/uuid"
	"github.com/pquerna/otp/totp"
)

func HandleMFASetup(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserIDFromToken(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var email string
	err = db.DB.QueryRow("SELECT email FROM users WHERE id = ?", userID).Scan(&email)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "ResultsPro",
		AccountName: email,
	})
	if err != nil {
		http.Error(w, "Failed to generate MFA key", http.StatusInternalServerError)
		return
	}

	// Store secret temporarily or expect user to verify it before finalizing
	// For simplicity, we'll store it in the users table but keep mfa_enabled = false
	_, err = db.DB.Exec("UPDATE users SET mfa_secret = ? WHERE id = ?", key.Secret(), userID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"secret": key.Secret(),
		"url":    key.URL(),
	})
}

func HandleMFAVerify(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserIDFromToken(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var input struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	var secret string
	err = db.DB.QueryRow("SELECT mfa_secret FROM users WHERE id = ?", userID).Scan(&secret)
	if err != nil || secret == "" {
		http.Error(w, "MFA not set up", http.StatusBadRequest)
		return
	}

	valid := totp.Validate(input.Code, secret)
	if !valid {
		http.Error(w, "Invalid MFA code", http.StatusUnauthorized)
		return
	}

	_, err = db.DB.Exec("UPDATE users SET mfa_enabled = TRUE WHERE id = ?", userID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "MFA enabled successfully"})
}

func HandleMFADisable(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserIDFromToken(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var input struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	var secret string
	err = db.DB.QueryRow("SELECT mfa_secret FROM users WHERE id = ?", userID).Scan(&secret)
	if err != nil {
		http.Error(w, "MFA not active", http.StatusBadRequest)
		return
	}

	if !totp.Validate(input.Code, secret) {
		http.Error(w, "Invalid MFA code", http.StatusUnauthorized)
		return
	}

	_, err = db.DB.Exec("UPDATE users SET mfa_enabled = FALSE, mfa_secret = NULL WHERE id = ?", userID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "MFA disabled successfully"})
}

func HandleMFAChallenge(w http.ResponseWriter, r *http.Request) {
	var input struct {
		UserID   string `json:"user_id"`
		Code     string `json:"code"`
		MFAToken string `json:"mfa_token"` // In a real app, verify this token's validity/expiry
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	var secret string
	err := db.DB.QueryRow("SELECT mfa_secret FROM users WHERE id = ?", input.UserID).Scan(&secret)
	if err != nil {
		http.Error(w, "User not found or MFA not active", http.StatusUnauthorized)
		return
	}

	if !totp.Validate(input.Code, secret) {
		http.Error(w, "Invalid MFA code", http.StatusUnauthorized)
		return
	}

	// MFA verified, generate final tokens
	accessToken, err := utils.GenerateAccessToken(input.UserID)
	if err != nil {
		http.Error(w, "Failed to generate access token", http.StatusInternalServerError)
		return
	}

	refreshToken, err := utils.GenerateRefreshToken()
	if err != nil {
		http.Error(w, "Failed to generate refresh token", http.StatusInternalServerError)
		return
	}

	// Save refresh token
	refreshTokenID := uuid.New().String()
	expiresAt := time.Now().Add(time.Hour * 24 * 7)
	deviceInfo := r.UserAgent()
	_, err = db.DB.Exec("INSERT INTO refresh_tokens (id, user_id, token_hash, device_info, expires_at) VALUES (?, ?, ?, ?, ?)",
		refreshTokenID, input.UserID, refreshToken, deviceInfo, expiresAt)
	if err != nil {
		http.Error(w, "Failed to save refresh token", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}
