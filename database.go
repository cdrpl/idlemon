package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"time"

	"github.com/go-sql-driver/mysql"
)

func CreateDBConn() *sql.DB {
	config := mysql.Config{
		User:                 os.Getenv("DB_USER"),
		Passwd:               os.Getenv("DB_PASS"),
		DBName:               os.Getenv("DB_NAME"),
		Addr:                 os.Getenv("DB_HOST"),
		Net:                  "tcp",
		AllowNativePasswords: true,
		ParseTime:            true,
		MultiStatements:      true,
	}

	db, err := sql.Open("mysql", config.FormatDSN())
	if err != nil {
		log.Fatalln("database connection error:", err)
	}

	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	return db
}

func InitDatabase(db *sql.DB) {
	CreateDatabaseTables(db)
	InsertUnitTypes(db)
	InsertUnitTemplates(db)
	InsertResources(db)
	InsertAdminUser(context.Background(), db)
}

func CreateDatabaseTables(db *sql.DB) {
	_, err := db.Exec(upSql)
	if err != nil {
		log.Fatalln("create tables error:", err)
	}
}

// Be careful since this function will delete all generated tables.
func DropTables(db *sql.DB) {
	_, err := db.Exec(downSql)
	if err != nil {
		log.Fatalln("drop tables error:", err)
	}
}

// Test connection and exit if it fails, attempt multiple times before exit since database may still be starting up
func DbConnectionTest(db *sql.DB) {
	for i := 0; i < DB_CONN_RETRIES; i++ {
		_, err := db.Exec("SELECT 1")
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
