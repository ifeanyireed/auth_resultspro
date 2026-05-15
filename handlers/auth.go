package handlers

import (
        "database/sql"
        "encoding/json"
        "fmt"
        "log"
        "math/rand"
        "net/http"
        "time"

        "auth.resultspro.ng/db"
        "auth.resultspro.ng/models"
        "auth.resultspro.ng/utils"
        "github.com/google/uuid"
        "golang.org/x/crypto/bcrypt"
)

func generateOTP() string {
        r := rand.New(rand.NewSource(time.Now().UnixNano()))
        return fmt.Sprintf("%06d", r.Intn(1000000))
}

func HandleSignup(w http.ResponseWriter, r *http.Request) {
        var input struct {
                Email    string `json:"email"`
                Password string `json:"password"`
                FullName string `json:"full_name"`
        }
        if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
                http.Error(w, "Invalid request", http.StatusBadRequest)
                return
        }

        if input.Email == "" || input.Password == "" {
                http.Error(w, "Email and password are required", http.StatusBadRequest)
                return
        }

        hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
        if err != nil {
                http.Error(w, "Failed to hash password", http.StatusInternalServerError)
                return
        }

        user := models.User{
                ID:            uuid.New().String(),
                Email:         input.Email,
                PasswordHash:  sql.NullString{String: string(hashedPassword), Valid: true},
                AuthProvider:  "local",
                FullName:      sql.NullString{String: input.FullName, Valid: true},
                AccountStatus: "unverified",
                MFAEnabled:    false,
                CreatedAt:     time.Now(),
                UpdatedAt:     time.Now(),
        }

        query := `INSERT INTO users (id, email, password_hash, auth_provider, full_name, account_status, mfa_enabled, created_at, updated_at) 
              VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

        _, err = db.DB.Exec(query,
                user.ID, user.Email, user.PasswordHash.String, user.AuthProvider, user.FullName.String, user.AccountStatus, 0, user.CreatedAt, user.UpdatedAt)

        if err != nil {
                log.Printf("Signup DB Error: %v", err)
                http.Error(w, "User already exists or database error", http.StatusConflict)
                return
        }

        otp := generateOTP()
        expiresAt := time.Now().Add(time.Hour * 24)
        _, err = db.DB.Exec("INSERT INTO verification_tokens (id, user_id, token_hash, type, expires_at) VALUES (?, ?, ?, 'email_verify', ?)",
                uuid.New().String(), user.ID, otp, expiresAt)
        if err != nil {
                log.Printf("Failed to create verification token: %v", err)
        }

        log.Printf("OTP for %s: %s", user.Email, otp)

        go func() {
                if err := utils.SendVerificationEmail(user.Email, otp); err != nil {
                        log.Printf("Failed to send verification email: %v", err)
                }
        }()

        w.WriteHeader(http.StatusCreated)
        json.NewEncoder(w).Encode(map[string]string{"message": "User created. Please check your email for verification code."})
}

func HandleLogin(w http.ResponseWriter, r *http.Request) {
        var input struct {
                Email    string `json:"email"`
                Password string `json:"password"`
        }
        if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
                http.Error(w, "Invalid request", http.StatusBadRequest)
                return
        }

        var user models.User
        err := db.DB.QueryRow("SELECT id, email, password_hash, account_status FROM users WHERE email = ? AND (auth_provider = 'local' OR auth_provider = 'both')", input.Email).Scan(
                &user.ID, &user.Email, &user.PasswordHash, &user.AccountStatus)

        if err == sql.ErrNoRows {
                http.Error(w, "Invalid email or password", http.StatusUnauthorized)
                return
        } else if err != nil {
                http.Error(w, "Database error", http.StatusInternalServerError)
                return
        }

        if !user.PasswordHash.Valid {
                http.Error(w, "Invalid login method", http.StatusUnauthorized)
                return
        }

        err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash.String), []byte(input.Password))
        if err != nil {
                http.Error(w, "Invalid email or password", http.StatusUnauthorized)
                return
        }

        if user.AccountStatus == "suspended" {
                http.Error(w, "Account suspended", http.StatusForbidden)
                return
        }

        var mfaEnabled bool
        db.DB.QueryRow("SELECT mfa_enabled FROM users WHERE id = ?", user.ID).Scan(&mfaEnabled)
        if mfaEnabled {
                mfaToken, _ := utils.GenerateRandomString(32)
                w.WriteHeader(http.StatusAccepted)
                json.NewEncoder(w).Encode(map[string]string{
                        "mfa_required": "true",
                        "mfa_token":    mfaToken,
                        "user_id":      user.ID,
                })
                return
        }

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
