package handlers

import (
	"encoding/json"
	"net/http"

	"auth.resultspro.ng/db"
	"auth.resultspro.ng/utils"
	"github.com/golang-jwt/jwt/v5"
)

func HandleIntrospect(w http.ResponseWriter, r *http.Request) {
	// Simple server-to-server auth check (App ID/Secret)
	appID := r.Header.Get("X-App-ID")
	appSecret := r.Header.Get("X-App-Secret")

	var secretKey string
	err := db.DB.QueryRow("SELECT secret_key FROM apps WHERE id = ?", appID).Scan(&secretKey)
	if err != nil || secretKey != appSecret {
		http.Error(w, "Unauthorized app", http.StatusUnauthorized)
		return
	}

	var input struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	token, err := utils.VerifyToken(input.Token)
	if err != nil || !token.Valid {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"active": false,
		})
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		http.Error(w, "Invalid claims", http.StatusInternalServerError)
		return
	}

	userID, err := claims.GetSubject()
	if err != nil {
		http.Error(w, "Subject not found in claims", http.StatusUnauthorized)
		return
	}

	var user struct {
		ID            string  `json:"id"`
		Email         string  `json:"email"`
		FullName      *string `json:"full_name"`
		AccountStatus string  `json:"account_status"`
	}
	err = db.DB.QueryRow("SELECT id, email, full_name, account_status FROM users WHERE id = ?", userID).Scan(&user.ID, &user.Email, &user.FullName, &user.AccountStatus)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	if user.AccountStatus == "suspended" {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"active": false,
			"reason": "account_suspended",
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"active": true,
		"user":   user,
	})
}
