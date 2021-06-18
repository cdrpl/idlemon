package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
)

func CreateDBConn(ctx context.Context) (*pgxpool.Pool, error) {
	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASS")
	db := os.Getenv("DB_NAME")
	host := os.Getenv("DB_HOST")
	maxConns := 10

	connString := fmt.Sprintf("user=%v password=%v dbname=%v host=%v pool_max_conns=%v", user, pass, db, host, maxConns)

	return pgxpool.Connect(ctx, connString)
}

func InitDatabase(ctx context.Context, db *pgxpool.Pool, dc DataCache) error {
	err := CreateDatabaseTables(ctx, db)
	if err != nil {
		return fmt.Errorf("fail to create database tables: %v", err)
	}

	err = InsertAdminUser(ctx, db, dc)
	if err != nil {
		return fmt.Errorf("fail to insert admin user: %v", err)
	}

	return nil
}

func CreateDatabaseTables(ctx context.Context, db *pgxpool.Pool) error {
	_, err := db.Exec(ctx, upSql)

	return err
}

// Be careful since this function will delete all generated tables.
func DropTables(ctx context.Context, db *pgxpool.Pool) error {
	_, err := db.Exec(ctx, downSql)

	return err
}

// Test connection and exit if it fails, attempt multiple times before exit since database may still be starting up
func DbConnectionTest(ctx context.Context, db *pgxpool.Pool) {
	for i := 0; i < DB_CONN_RETRIES; i++ {
		_, err := db.Exec(ctx, "SELECT 1")
		if err != nil {
			if i == DB_CONN_RETRIES-1 {
				log.Fatalln("database connection error:", err)
			} else {
				log.Println("database test", i+1, "failed")
				time.Sleep(time.Second * 1)
				continue
			}
		} else {
			log.Println("database connection successful")
			break
		}
	}
}
