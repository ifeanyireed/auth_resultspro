package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"auth.resultspro.ng/db"
	"auth.resultspro.ng/utils"
	"github.com/golang-jwt/jwt/v5"
)

func HandleTokenRefresh(w http.ResponseWriter, r *http.Request) {
	var input struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	var userID string
	var expiresAt time.Time
	var revoked bool
	err := db.DB.QueryRow("SELECT user_id, expires_at, revoked FROM refresh_tokens WHERE token_hash = ?", input.RefreshToken).Scan(&userID, &expiresAt, &revoked)
	if err != nil {
		http.Error(w, "Invalid refresh token", http.StatusUnauthorized)
		return
	}

	if revoked || time.Now().After(expiresAt) {
		http.Error(w, "Token expired or revoked", http.StatusUnauthorized)
		return
	}

	// Check if account is still active
	var status string
	err = db.DB.QueryRow("SELECT account_status FROM users WHERE id = ?", userID).Scan(&status)
	if err != nil || status == "suspended" {
		http.Error(w, "Account suspended or not found", http.StatusForbidden)
		return
	}

	accessToken, err := utils.GenerateAccessToken(userID)
	if err != nil {
		http.Error(w, "Failed to generate access token", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"access_token": accessToken,
	})
}

func HandleLogout(w http.ResponseWriter, r *http.Request) {
	var input struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	_, err := db.DB.Exec("UPDATE refresh_tokens SET revoked = 1 WHERE token_hash = ?", input.RefreshToken)
	if err != nil {
		http.Error(w, "Failed to revoke token", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Logged out"))
}

func HandleLogoutAll(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if len(authHeader) < 8 {
		http.Error(w, "Invalid authorization header", http.StatusUnauthorized)
		return
	}

	tokenString := authHeader[7:]
	token, err := utils.VerifyToken(tokenString)
	if err != nil || !token.Valid {
		http.Error(w, "Invalid access token", http.StatusUnauthorized)
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		http.Error(w, "Invalid claims", http.StatusInternalServerError)
		return
	}

	userID, err := claims.GetSubject()
	if err != nil {
		http.Error(w, "Subject not found", http.StatusUnauthorized)
		return
	}

	_, err = db.DB.Exec("UPDATE refresh_tokens SET revoked = 1 WHERE user_id = ?", userID)
	if err != nil {
		http.Error(w, "Failed to revoke tokens", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Logged out from all devices"))
}
