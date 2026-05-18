package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"auth.resultspro.ng/db"
	"auth.resultspro.ng/utils"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func HandleVerifyEmail(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	var userID string
	var expiresAt time.Time
	var used bool
	err := db.DB.QueryRow("SELECT user_id, expires_at, used FROM verification_tokens WHERE token_hash = ? AND type = 'email_verify'", input.Token).Scan(&userID, &expiresAt, &used)
	if err != nil {
		http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
		return
	}

	if used || time.Now().After(expiresAt) {
		http.Error(w, "Token already used or expired", http.StatusUnauthorized)
		return
	}

	tx, err := db.DB.Begin()
	if err != nil {
		http.Error(w, "Transaction error", http.StatusInternalServerError)
		return
	}

	_, err = tx.Exec("UPDATE users SET account_status = 'active', updated_at = ? WHERE id = ?", time.Now().UTC().Format("2006-01-02T15:04:05.000Z"), userID)
	if err != nil {
		tx.Rollback()
		http.Error(w, "Failed to update user status", http.StatusInternalServerError)
		return
	}

	_, err = tx.Exec("UPDATE verification_tokens SET used = 1 WHERE token_hash = ?", input.Token)
	if err != nil {
		tx.Rollback()
		http.Error(w, "Failed to update token status", http.StatusInternalServerError)
		return
	}

	tx.Commit()
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Email verified successfully"})
}

func HandleForgotPassword(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		log.Printf("Forgot Password: Failed to decode request: %v", err)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	email := strings.ToLower(strings.TrimSpace(input.Email))
	log.Printf("Forgot Password request for: %s (Original: %s)", email, input.Email)

	var userID string
	err := db.DB.QueryRow("SELECT id FROM users WHERE email = ?", email).Scan(&userID)
	if err == sql.ErrNoRows {
		log.Printf("Forgot Password: No user found with email: %s", email)
		// Don't reveal if user exists
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "If an account exists, a reset link has been sent."})
		return
	} else if err != nil {
		log.Printf("Forgot Password: Database error: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	token, _ := utils.GenerateRandomString(32)
	expiresAt := time.Now().Add(time.Hour * 1)
	_, err = db.DB.Exec("INSERT INTO verification_tokens (id, user_id, token_hash, type, expires_at) VALUES (?, ?, ?, 'password_reset', ?)",
		uuid.New().String(), userID, token, expiresAt.UTC().Format("2006-01-02T15:04:05.000Z"))
	if err != nil {
		log.Printf("Forgot Password: Failed to insert token: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	log.Printf("Forgot Password: Token generated for %s. Sending email...", email)

	go func() {
		if err := utils.SendPasswordResetEmail(email, token); err != nil {
			log.Printf("CRITICAL: Failed to send password reset email to %s: %v", email, err)
		} else {
			log.Printf("SUCCESS: Password reset email sent to %s", email)
		}
	}()

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "If an account exists, a reset link has been sent."})
}

func HandleResetPassword(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Token       string `json:"token"`
		NewPassword string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	var userID string
	var expiresAt time.Time
	var used bool
	err := db.DB.QueryRow("SELECT user_id, expires_at, used FROM verification_tokens WHERE token_hash = ? AND type = 'password_reset'", input.Token).Scan(&userID, &expiresAt, &used)
	if err != nil {
		http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
		return
	}

	if used || time.Now().After(expiresAt) {
		http.Error(w, "Token already used or expired", http.StatusUnauthorized)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	tx, err := db.DB.Begin()
	if err != nil {
		http.Error(w, "Transaction error", http.StatusInternalServerError)
		return
	}

	_, err = tx.Exec("UPDATE users SET password_hash = ?, updated_at = ? WHERE id = ?", string(hashedPassword), time.Now().UTC().Format("2006-01-02T15:04:05.000Z"), userID)
	if err != nil {
		tx.Rollback()
		http.Error(w, "Failed to update password", http.StatusInternalServerError)
		return
	}

	_, err = tx.Exec("UPDATE verification_tokens SET used = 1 WHERE token_hash = ?", input.Token)
	if err != nil {
		tx.Rollback()
		http.Error(w, "Failed to update token status", http.StatusInternalServerError)
		return
	}

	tx.Commit()
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Password reset successfully"})
}

func HandleChangePassword(w http.ResponseWriter, r *http.Request) {
	// Authenticated endpoint
	userID, err := getUserIDFromToken(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var input struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	var currentHash *string
	err = db.DB.QueryRow("SELECT password_hash FROM users WHERE id = ?", userID).Scan(&currentHash)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	if currentHash != nil {
		err = bcrypt.CompareHashAndPassword([]byte(*currentHash), []byte(input.OldPassword))
		if err != nil {
			http.Error(w, "Invalid old password", http.StatusUnauthorized)
			return
		}
	}

	newHash, _ := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
	_, err = db.DB.Exec("UPDATE users SET password_hash = ?, auth_provider = CASE WHEN auth_provider = 'google' THEN 'both' ELSE auth_provider END, updated_at = ? WHERE id = ?", string(newHash), time.Now().UTC().Format("2006-01-02T15:04:05.000Z"), userID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Password changed successfully"})
}

func HandleChangeEmail(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserIDFromToken(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var input struct {
		NewEmail string `json:"new_email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	newEmail := strings.ToLower(strings.TrimSpace(input.NewEmail))
	_, err = db.DB.Exec("UPDATE users SET email = ?, account_status = 'unverified', updated_at = ? WHERE id = ?", newEmail, time.Now().UTC().Format("2006-01-02T15:04:05.000Z"), userID)
	if err != nil {
		http.Error(w, "Email already in use or database error", http.StatusConflict)
		return
	}

	// Trigger new verification email (Phase 3)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Email updated. Please verify your new email."})
}

func getUserIDFromToken(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if len(authHeader) < 8 {
		return "", sql.ErrNoRows
	}
	tokenString := authHeader[7:]
	token, err := utils.VerifyToken(tokenString)
	if err != nil || !token.Valid {
		return "", err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", err
	}
	return claims.GetSubject()
}
