package main

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"os"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// Model of the user table
type User struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Pass      string    `json:"pass"`
	Exp       int       `json:"exp"`
	CreatedAt time.Time `json:"createdAt"`
}

// Will hash the user password then insert into the database.
func InsertUser(ctx context.Context, db *sql.DB, name string, email string, pass string) (sql.Result, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	result, err := db.ExecContext(ctx, "INSERT INTO user (name, email, pass) VALUES (?, ?, ?)", name, email, string(hash))
	if err != nil {
		return nil, err
	}

	return result, nil
}

// Will insert the admin user if it doesn't exist.
func InsertAdminUser(db *sql.DB) {
	var id int

	// only insert admin user if no admin user exists
	err := db.QueryRow("SELECT id FROM user WHERE id = ?", ADMIN_ID).Scan(&id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			pass := os.Getenv("ADMIN_PASS")

			hash, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
			if err != nil {
				log.Fatalln("insert admin user error:", err)
			}

			_, err = db.Exec("INSERT INTO user (id, name, email, pass) VALUES (?, ?, ?, ?)", ADMIN_ID, ADMIN_NAME, ADMIN_EMAIL, string(hash))
			if err != nil {
				log.Fatalln("insert admin user error:", err)
			}

			log.Printf("insert admin user {ID:%v Name:%v Email:%v}\n", ADMIN_ID, ADMIN_NAME, ADMIN_EMAIL)
		} else {
			log.Fatalln("insert admin user error:", err)
		}
	}
}

// Returns true if the user name is already taken.
func NameExists(ctx context.Context, db *sql.DB, name string) (bool, error) {
	err := db.QueryRowContext(ctx, "SELECT name FROM user WHERE name = ?", name).Scan(&name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		} else {
			return false, err
		}
	}

	return true, nil
}

// Returns true if the user email is already taken.
func EmailExists(ctx context.Context, db *sql.DB, email string) (bool, error) {
	err := db.QueryRowContext(ctx, "SELECT email FROM user WHERE email = ?", email).Scan(&email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		} else {
			return false, err
		}
	}

	return true, nil
}
