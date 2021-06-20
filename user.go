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
	Pass      string    `json:"-"`
	CreatedAt time.Time `json:"createdAt"`
	Data      UserData  `json:"data"`
}

// Create a user struct with CreatedAt set to now.
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
	Exp                int                  `json:"exp"`
	UnitSerial         int                  `json:"unitSerial"` // used for auto incrementing unit ID
	Campaign           Campaign             `json:"campaign"`
	DailyQuestProgress []DailyQuestProgress `json:"dailyQuestProgress"`
	Resources          []Resource           `json:"resources"`
	Units              map[int]Unit         `json:"units"` // map uses unit ID as key
}

func CreateUserData(dc DataCache, time time.Time) UserData {
	data := UserData{
		Campaign:           Campaign{Level: 1, LastCollectedAt: time},
		DailyQuestProgress: make([]DailyQuestProgress, 0),
		Resources:          make([]Resource, 0),
		Units:              make(map[int]Unit),
	}

	data.Resources = dc.Resources

	for i := range dc.DailyQuests {
		data.DailyQuestProgress = append(data.DailyQuestProgress, CreateDailyQuestProgress(i))
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
func InsertUser(ctx context.Context, db *pgxpool.Pool, dc DataCache, user User) (int, error) {
	tx, err := db.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	var userId int
	query := "INSERT INTO users (name, email, pass, created_at, data) VALUES ($1, $2, $3, $4, $5) RETURNING id"
	err = tx.QueryRow(ctx, query, user.Name, user.Email, user.Pass, user.CreatedAt, user.Data).Scan(&userId)
	if err != nil {
		return userId, err
	}

	// insert resource rows
	for i := range dc.Resources {
		query := "INSERT INTO resources (user_id, type) VALUES ($1, $2)"

		_, err := tx.Exec(ctx, query, userId, i)
		if err != nil {
			return userId, err
		}
	}

	err = tx.Commit(ctx)

	return userId, err
}

// Will insert the admin user if it doesn't exist.
func InsertAdminUser(ctx context.Context, db *pgxpool.Pool, dc DataCache) error {
	var id int

	// only insert admin user if no admin user exists
	err := db.QueryRow(ctx, "SELECT id FROM users WHERE id = $1", ADMIN_ID).Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			pass := os.Getenv("ADMIN_PASS")

			hash, err := bcrypt.GenerateFromPassword([]byte(pass), BCRYPT_COST)
			if err != nil {
				return err
			}

			user := CreateUser(dc, ADMIN_NAME, ADMIN_EMAIL, string(hash))

			// give admin user a lot of resources for easy testing
			for _, resource := range dc.Resources {
				user.Data.Resources[resource.Id].Amount = 2000000000
			}

			_, err = InsertUser(ctx, db, dc, user)
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
