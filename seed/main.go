package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load("../.env")
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	db, err := sql.Open("mysql", dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	fmt.Println("Seeding Central Auth database...")

	users := []struct {
		ID       string
		Email    string
		Pass     string
		Provider string
		Name     string
		Status   string
	}{
		{"bfb51c68-ccb0-401f-b58f-27fd41c6a856", "superadmin@resultspro.ng", "$2a$14$1zhGRoc.lxuxyO/9X27HpuUTq06m5p2pb69PgYa0UWksEJWT7kS8i", "local", "Super Admin", "active"},
		{"2db093ed-bdc9-47c4-b71c-66869f0f1ea7", "teacher@example.edu", "$2a$14$Jg0JSBXO09zmMOssPyzEj.VyO/iuXai.QCZQFicC4CTR.plVD9dMS", "local", "Mr. Adeniyi", "active"},
		{"111efa7d-e12d-4ed1-9902-d341c6826b50", "student@example.com", "$2a$14$OiOxIN4UiEuFHKIhwdmFHuNbtI2FoVpU95KVD8Dc3FxLhHM2.EMve", "local", "Jane Doe", "active"},
		{"dac38ffd-866f-47ab-8ac4-ecf6ea520ba8", "parent@example.com", "$2a$14$4nofWUGNaOyx9/2zF23ySuu5ehgcPa1kApyvp5dLAHszuA.NoLOWS", "local", "Mrs. Doe", "active"},
		{"8d3a7776-5d21-4f1e-9a6d-e4c1d63e9f01", "platform-admin@resultspro.ng", "$2a$14$1zhGRoc.lxuxyO/9X27HpuUTq06m5p2pb69PgYa0UWksEJWT7kS8i", "local", "Platform Admin", "active"},
		{"8d3a7776-5d21-4f1e-9a6d-e4c1d63e9f02", "school-admin@example.edu", "$2a$14$1zhGRoc.lxuxyO/9X27HpuUTq06m5p2pb69PgYa0UWksEJWT7kS8i", "local", "School Admin", "active"},
		{"8d3a7776-5d21-4f1e-9a6d-e4c1d63e9f03", "support-staff@resultspro.ng", "$2a$14$1zhGRoc.lxuxyO/9X27HpuUTq06m5p2pb69PgYa0UWksEJWT7kS8i", "local", "Support Staff", "active"},
	}

	for _, u := range users {
		_, err := db.Exec("INSERT IGNORE INTO users (id, email, password_hash, auth_provider, full_name, account_status) VALUES (?, ?, ?, ?, ?, ?)", u.ID, u.Email, u.Pass, u.Provider, u.Name, u.Status)
		if err != nil {
			log.Printf("Error seeding user %s: %v", u.Email, err)
		}
	}

	_, err = db.Exec("INSERT IGNORE INTO apps (id, name, secret_key) VALUES (?, ?, ?)", "classroompro-app-id", "ClassroomPRO", "your-app-secret-key-123")
	if err != nil {
		log.Printf("Error seeding ClassroomPRO app: %v", err)
	}
    
    // Also seed the Academics Service App
    _, err = db.Exec("INSERT IGNORE INTO apps (id, name, secret_key) VALUES (?, ?, ?)", "acad_service_001", "AcadService", "your_secret_here")
	if err != nil {
		log.Printf("Error seeding AcadService app: %v", err)
	}

	fmt.Println("Central Auth Seeding complete!")
}
