package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v4/pgxpool"
)

func CreateDBConn(ctx context.Context) (*pgxpool.Pool, error) {
	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASS")
	db := os.Getenv("DB_NAME")
	host := os.Getenv("DB_HOST")
	maxConns := MAX_PG_CONN

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
