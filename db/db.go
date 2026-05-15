package db

import (
	"database/sql"
	"log"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

func InitDB(dataSourceName string) {
	var err error
	DB, err = sql.Open("sqlite", dataSourceName)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	if err = DB.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	createTables()
}

func createTables() {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			email TEXT UNIQUE NOT NULL,
			password_hash TEXT,
                        role TEXT DEFAULT "STUDENT",
			google_id TEXT UNIQUE,
			microsoft_id TEXT UNIQUE,
			auth_provider TEXT NOT NULL,
			full_name TEXT,
			avatar_url TEXT,
			account_status TEXT DEFAULT 'unverified',
			mfa_enabled BOOLEAN DEFAULT FALSE,
			mfa_secret TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE IF NOT EXISTS verification_tokens (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			token_hash TEXT NOT NULL,
			type TEXT NOT NULL,
			expires_at DATETIME NOT NULL,
			used BOOLEAN DEFAULT FALSE,
			FOREIGN KEY (user_id) REFERENCES users(id)
		);`,
		`CREATE TABLE IF NOT EXISTS apps (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			secret_key TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS refresh_tokens (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			token_hash TEXT NOT NULL,
			device_info TEXT,
			expires_at DATETIME NOT NULL,
			revoked BOOLEAN DEFAULT FALSE,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id)
		);`,
	}

	for _, query := range queries {
		if _, err := DB.Exec(query); err != nil {
			log.Fatalf("Failed to create table: %v", err)
		}
	}
}
