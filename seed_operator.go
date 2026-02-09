package main

import (
	"database/sql"
	"fmt"
	"log"

	"golang.org/x/crypto/bcrypt"
)

// CreateDefaultOperator creates a default operator account for testing
func CreateDefaultOperator(db *sql.DB) {
	// Check if operator already exists
	var count int
	db.QueryRow("SELECT COUNT(*) FROM operators WHERE username = 'operator'").Scan(&count)
	
	if count > 0 {
		log.Println("Default operator already exists")
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("operator123"), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal("Failed to hash password:", err)
	}

	// Insert default operator
	_, err = db.Exec(`
		INSERT INTO operators (username, password, name, phone, is_active)
		VALUES (?, ?, ?, ?, ?)`,
		"operator", string(hashedPassword), "Operator Default", "081234567890", true)

	if err != nil {
		log.Fatal("Failed to create default operator:", err)
	}

	fmt.Println("✓ Default operator created:")
	fmt.Println("  Username: operator")
	fmt.Println("  Password: operator123")
}
