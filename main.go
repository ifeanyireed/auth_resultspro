package main

import (
	"log"
	"net/http"
	"os"

	"auth.resultspro.ng/config"
	"auth.resultspro.ng/db"
	"auth.resultspro.ng/handlers"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "auth.db"
	}

	db.InitDB(dbPath)
	config.InitConfig()

	http.HandleFunc("/auth/signup", handlers.HandleSignup)
	http.HandleFunc("/auth/login", handlers.HandleLogin)
	http.HandleFunc("/auth/google", handlers.HandleGoogleLogin)
	http.HandleFunc("/callback", handlers.HandleGoogleCallback)
	http.HandleFunc("/auth/microsoft", handlers.HandleMicrosoftLogin)
	http.HandleFunc("/callback/microsoft", handlers.HandleMicrosoftCallback)
	http.HandleFunc("/auth/refresh", handlers.HandleTokenRefresh)
	http.HandleFunc("/auth/logout", handlers.HandleLogout)
	http.HandleFunc("/auth/logout-all", handlers.HandleLogoutAll)
	http.HandleFunc("/auth/introspect", handlers.HandleIntrospect)

	// Account Management
	http.HandleFunc("/auth/verify-email", handlers.HandleVerifyEmail)
	http.HandleFunc("/auth/forgot-password", handlers.HandleForgotPassword)
	http.HandleFunc("/auth/reset-password", handlers.HandleResetPassword)
	http.HandleFunc("/auth/change-password", handlers.HandleChangePassword)
	http.HandleFunc("/auth/change-email", handlers.HandleChangeEmail)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s...", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
