package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

// Model of the user table
type User struct {
	Id        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Pass      string    `json:"-"`
	Exp       int       `json:"exp"`
	CreatedAt time.Time `json:"createdAt"`
}

// Create a user struct with CreatedAt set to now.
func CreateUser(dc DataCache, name string, email string, pass string) User {
	now := time.Now()

	return User{
		Id:        uuid.New(),
		Name:      name,
		Email:     email,
		Pass:      pass,
		CreatedAt: now,
	}
}

// Will hash the user's password then insert it into the database. Returns the user's ID.
func InsertUser(ctx context.Context, tx pgx.Tx, dc DataCache, user User) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(user.Pass), BCRYPT_COST)
	if err != nil {
		return fmt.Errorf("fail to hash password: %w", err)
	}

	query := "INSERT INTO users (id, name, email, pass, created_at) VALUES ($1, $2, $3, $4, $5)"
	_, err = tx.Exec(ctx, query, user.Id, user.Name, user.Email, hash, user.CreatedAt)
	if err != nil {
		return err
	}

	if err := InsertResources(ctx, tx, dc, user.Id); err != nil {
		return err
	}

	if err := InsertDailyQuestProgress(ctx, tx, dc, user.Id); err != nil {
		return err
	}

	if err := InsertCampaign(ctx, tx, user.Id); err != nil {
		return err
	}

	return nil
}

// Will insert the admin user if it doesn't exist.
func InsertAdminUser(ctx context.Context, db *pgxpool.Pool, dc DataCache) error {
	user := CreateUser(dc, os.Getenv("ADMIN_NAME"), os.Getenv("ADMIN_EMAIL"), os.Getenv("ADMIN_PASS"))

	tx, err := db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("fail to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := InsertUser(ctx, tx, dc, user); err != nil {
		var pgErr *pgconn.PgError

		// don't consider as error if admin user already exists
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return nil
		} else {
			return fmt.Errorf("fail to insert user: %w", err)
		}
	}

	// give admin user some resources to make testing easier
	for _, resource := range dc.Resources {
		if err := IncResource(ctx, tx, user.Id, resource.Type, 100000); err != nil {
			return fmt.Errorf("fail to increase user resources: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("fail to commit transaction: %w", err)
	}

	log.Printf("insert admin user {ID:%v Name:%v Email:%v}\n", user.Id, user.Name, user.Email)

	return nil
}

func IncUserExp(ctx context.Context, tx pgx.Tx, userId uuid.UUID, amount int) error {
	query := "UPDATE users SET exp = exp + $1 WHERE id = $2"

	_, err := tx.Exec(ctx, query, amount, userId)
	if err != nil {
		return fmt.Errorf("fail to update users row: %w", err)
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
			return false, fmt.Errorf("fail to query user row: %w", err)
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
			return false, fmt.Errorf("fail to query user row: %w", err)
		}
	}

	return true, nil
}
