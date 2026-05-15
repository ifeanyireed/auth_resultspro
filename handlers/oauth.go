package handlers

import (
        "context"
        "crypto/rand"
        "database/sql"
        "encoding/base64"
        "encoding/json"
        "net/http"
        "time"

        "auth.resultspro.ng/config"
        "auth.resultspro.ng/db"
        "auth.resultspro.ng/models"
        "auth.resultspro.ng/utils"
        "github.com/google/uuid"
)

func HandleGoogleLogin(w http.ResponseWriter, r *http.Request) {
        state := generateStateOauthCookie(w)
        url := config.GoogleOAuthConfig.AuthCodeURL(state)
        http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func HandleGoogleCallback(w http.ResponseWriter, r *http.Request) {
        oauthState, _ := r.Cookie("oauthstate")

        if oauthState == nil || r.FormValue("state") != oauthState.Value {
                http.Error(w, "Invalid state", http.StatusBadRequest)
                return
        }

        token, err := config.GoogleOAuthConfig.Exchange(context.Background(), r.FormValue("code"))
        if err != nil {
                http.Error(w, "Code exchange failed", http.StatusInternalServerError)
                return
        }

        response, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
        if err != nil {
                http.Error(w, "Failed to get user info", http.StatusInternalServerError)
                return
        }
        defer response.Body.Close()

        var googleUser struct {
                ID      string `json:"id"`
                Email   string `json:"email"`
                Name    string `json:"name"`
                Picture string `json:"picture"`
        }

        if err := json.NewDecoder(response.Body).Decode(&googleUser); err != nil {
                http.Error(w, "Failed to decode user info", http.StatusInternalServerError)
                return
        }

        processOAuthUser(w, r, googleUser.ID, "", googleUser.Email, googleUser.Name, googleUser.Picture, "google")
}

func HandleMicrosoftLogin(w http.ResponseWriter, r *http.Request) {
        state := generateStateOauthCookie(w)
        url := config.MicrosoftOAuthConfig.AuthCodeURL(state)
        http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func HandleMicrosoftCallback(w http.ResponseWriter, r *http.Request) {
        oauthState, _ := r.Cookie("oauthstate")

        if oauthState == nil || r.FormValue("state") != oauthState.Value {
                http.Error(w, "Invalid state", http.StatusBadRequest)
                return
        }

        token, err := config.MicrosoftOAuthConfig.Exchange(context.Background(), r.FormValue("code"))
        if err != nil {
                http.Error(w, "Code exchange failed", http.StatusInternalServerError)
                return
        }

        client := config.MicrosoftOAuthConfig.Client(context.Background(), token)
        response, err := client.Get("https://graph.microsoft.com/v1.0/me")
        if err != nil {
                http.Error(w, "Failed to get user info", http.StatusInternalServerError)
                return
        }
        defer response.Body.Close()

        var microsoftUser struct {
                ID                string `json:"id"`
                UserPrincipalName string `json:"userPrincipalName"`
                DisplayName       string `json:"displayName"`
        }

        if err := json.NewDecoder(response.Body).Decode(&microsoftUser); err != nil {
                http.Error(w, "Failed to decode user info", http.StatusInternalServerError)
                return
        }

        processOAuthUser(w, r, "", microsoftUser.ID, microsoftUser.UserPrincipalName, microsoftUser.DisplayName, "", "microsoft")
}

func processOAuthUser(w http.ResponseWriter, r *http.Request, googleID, microsoftID, email, name, avatar, provider string) {
        var user models.User
        query := "SELECT id, email, google_id, microsoft_id, auth_provider, full_name, avatar_url, account_status FROM users WHERE email = ?"
        params := []interface{}{email}

        if googleID != "" {
                query += " OR google_id = ?"
                params = append(params, googleID)
        }
        if microsoftID != "" {
                query += " OR microsoft_id = ?"
                params = append(params, microsoftID)
        }

        err := db.DB.QueryRow(query, params...).Scan(
                &user.ID, &user.Email, &user.GoogleID, &user.MicrosoftID, &user.AuthProvider, &user.FullName, &user.AvatarURL, &user.AccountStatus)

        if err == sql.ErrNoRows {
                user = models.User{
                        ID:            uuid.New().String(),
                        Email:         email,
                        GoogleID:      sql.NullString{String: googleID, Valid: googleID != ""},
                        MicrosoftID:   sql.NullString{String: microsoftID, Valid: microsoftID != ""},
                        AuthProvider:  provider,
                        FullName:      sql.NullString{String: name, Valid: name != ""},
                        AvatarURL:     sql.NullString{String: avatar, Valid: avatar != ""},
                        AccountStatus: "active",
                        CreatedAt:     time.Now(),
                        UpdatedAt:     time.Now(),
                }
                _, err = db.DB.Exec("INSERT INTO users (id, email, google_id, microsoft_id, auth_provider, full_name, avatar_url, account_status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
                        user.ID, user.Email, user.GoogleID, user.MicrosoftID, user.AuthProvider, user.FullName, user.AvatarURL, user.AccountStatus, user.CreatedAt, user.UpdatedAt)
                if err != nil {
                        http.Error(w, "Failed to save user", http.StatusInternalServerError)
                        return
                }
        } else if err != nil {
                http.Error(w, "Database error", http.StatusInternalServerError)
                return
        } else {
                updated := false
                if googleID != "" && !user.GoogleID.Valid {
                        user.GoogleID = sql.NullString{String: googleID, Valid: true}
                        updated = true
                }
                if microsoftID != "" && !user.MicrosoftID.Valid {
                        user.MicrosoftID = sql.NullString{String: microsoftID, Valid: true}
                        updated = true
                }

                if updated {
                        if user.AuthProvider != provider && user.AuthProvider != "both" && user.AuthProvider != "mixed" {
                                user.AuthProvider = "mixed"
                        }
                        user.AccountStatus = "active"
                        _, err = db.DB.Exec("UPDATE users SET google_id = ?, microsoft_id = ?, auth_provider = ?, account_status = ?, updated_at = ? WHERE id = ?",
                                user.GoogleID, user.MicrosoftID, user.AuthProvider, user.AccountStatus, time.Now(), user.ID)
                        if err != nil {
                                http.Error(w, "Failed to update user", http.StatusInternalServerError)
                                return
                        }
                }
        }

        if user.AccountStatus == "suspended" {
                http.Error(w, "Account suspended", http.StatusForbidden)
                return
        }

        // Generate tokens
        accessToken, err := utils.GenerateAccessToken(user.ID)
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
                refreshTokenID, user.ID, refreshToken, deviceInfo, expiresAt)
        if err != nil {
                http.Error(w, "Failed to save refresh token", http.StatusInternalServerError)
                return
        }

        json.NewEncoder(w).Encode(map[string]string{
                "access_token":  accessToken,
                "refresh_token": refreshToken,
        })
}

func generateStateOauthCookie(w http.ResponseWriter) string {
        var b [16]byte
        rand.Read(b[:])
        state := base64.URLEncoding.EncodeToString(b[:])
        cookie := http.Cookie{Name: "oauthstate", Value: state, Expires: time.Now().Add(365 * 24 * time.Hour), HttpOnly: true}
        http.SetCookie(w, &cookie)
        return state
}
