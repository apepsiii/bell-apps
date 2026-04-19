package repository

import (
	"database/sql"
	"fmt"
	"log"

	"golang.org/x/crypto/bcrypt"
)

func CreateDefaultOperator(db *sql.DB) {
	var count int
	db.QueryRow("SELECT COUNT(*) FROM operators WHERE username = 'operator'").Scan(&count)

	if count > 0 {
		log.Println("Default operator already exists")
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("operator123"), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal("Failed to hash password:", err)
	}

	_, err = db.Exec(`
		INSERT INTO operators (username, password, name, phone, is_active)
		VALUES (?, ?, ?, ?, ?)`,
		"operator", string(hashedPassword), "Operator Default", "081234567890", true)

	if err != nil {
		log.Fatal("Failed to create default operator:", err)
	}

	fmt.Println("Default operator created:")
	fmt.Println("  Username: operator")
	fmt.Println("  Password: operator123")
}
