package main

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

// Model of the user table
type User struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Pass      string    `json:"pass"`
	CreatedAt time.Time `json:"createdAt"`
	Data      UserData  `json:"data"`
}

func CreateUser(dc DataCache, name string, email string, pass string) User {
	now := time.Now().UTC().Round(time.Second)

	user := User{
		Name:      name,
		Email:     email,
		Pass:      pass,
		CreatedAt: now,
		Data:      CreateUserData(dc, now),
	}

	return user
}

type UserData struct {
	Exp         int              `json:"exp"`
	Campaign    Campaign         `json:"campaign"`
	DailyQuests []UserDailyQuest `json:"dailyQuests"`
	Resources   []UserResource   `json:"resources"`
	Units       []Unit           `json:"units"`
}

func CreateUserData(dc DataCache, time time.Time) UserData {
	data := UserData{
		Campaign:    Campaign{Level: 1, LastCollectedAt: time},
		DailyQuests: make([]UserDailyQuest, 0),
		Resources:   make([]UserResource, 0),
		Units:       make([]Unit, 0),
	}

	for range dc.Resources {
		data.Resources = append(data.Resources, UserResource{})
	}

	for range dc.DailyQuests {
		data.DailyQuests = append(data.DailyQuests, UserDailyQuest{})
	}

	return data
}

// Implement the sql driver.Valuer interface.
func (u *UserData) Value() (driver.Value, error) {
	return json.Marshal(u)
}

// Implement the sql.Scanner interface.
func (u *UserData) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &u)
}

// Find a user by ID from the database.
func FindUser(ctx context.Context, db *pgxpool.Pool, id int) (User, error) {
	user := User{ID: id}

	query := "SELECT name, email, pass, created_at, data FROM users WHERE id = $1"
	err := db.QueryRow(ctx, query, id).Scan(&user.Name, &user.Email, &user.Pass, &user.CreatedAt, &user.Data)

	return user, err
}

// Find a user by ID from the database for update.
func FindUserLock(ctx context.Context, tx pgx.Tx, id int) (User, error) {
	user := User{ID: id}

	query := "SELECT name, email, pass, created_at, data FROM users WHERE id = $1 FOR UPDATE"
	err := tx.QueryRow(ctx, query, id).Scan(&user.Name, &user.Email, &user.Pass, &user.CreatedAt, &user.Data)

	return user, err
}

// Will update the user's data column in the database.
func UpdateUser(ctx context.Context, db *pgxpool.Pool, user User) error {
	query := "UPDATE users SET data = $1 WHERE id = $2"
	_, err := db.Exec(ctx, query, user.Data, user.ID)

	return err
}

// Will update the user's data column in the database using an existing transaction.
func UpdateUserLock(ctx context.Context, tx pgx.Tx, user User) error {
	query := "UPDATE users SET data = $1 WHERE id = $2"
	_, err := tx.Exec(ctx, query, user.Data, user.ID)

	return err
}

// Will insert the user into the database. Returns the user's ID.
func InsertUser(ctx context.Context, db *pgxpool.Pool, user User) (int, error) {
	var userID int

	query := "INSERT INTO users (name, email, pass, created_at, data) VALUES ($1, $2, $3, $4, $5) RETURNING id"
	err := db.QueryRow(ctx, query, user.Name, user.Email, user.Pass, user.CreatedAt, user.Data).Scan(&userID)

	return userID, err
}

// Will insert the admin user if it doesn't exist.
func InsertAdminUser(ctx context.Context, db *pgxpool.Pool, dc DataCache) error {
	var id int

	// only insert admin user if no admin user exists
	err := db.QueryRow(ctx, "SELECT id FROM users WHERE id = $1", ADMIN_ID).Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			pass := os.Getenv("ADMIN_PASS")

			hash, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
			if err != nil {
				return err
			}

			_, err = InsertUser(ctx, db, CreateUser(dc, ADMIN_NAME, ADMIN_EMAIL, string(hash)))
			if err != nil {
				return err
			}

			log.Printf("insert admin user {ID:%v Name:%v Email:%v}\n", ADMIN_ID, ADMIN_NAME, ADMIN_EMAIL)
		} else {
			return fmt.Errorf("fail to query admin user: %v", err)
		}
	}

	return nil
}

// Returns true if the user name is already taken.
func NameExists(ctx context.Context, db *pgxpool.Pool, name string) (bool, error) {
	err := db.QueryRow(ctx, "SELECT name FROM users WHERE name = $1", name).Scan(&name)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		} else {
			return false, err
		}
	}

	return true, nil
}

// Returns true if the user email is already taken.
func EmailExists(ctx context.Context, db *pgxpool.Pool, email string) (bool, error) {
	err := db.QueryRow(ctx, "SELECT email FROM users WHERE email = $1", email).Scan(&email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		} else {
			return false, err
		}
	}

	return true, nil
}
