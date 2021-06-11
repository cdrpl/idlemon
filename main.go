package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-redis/redis/v8"
)

func main() {
	log.Println("starting server")

	envFile, dropTables := parseFlags()
	LoadEnv(envFile, VERSION)

	log.Println("creating database connection")
	db := CreateDBConn()
	log.Println("database connection created")

	log.Println("testing database connection")
	DbConnectionTest(db)

	if dropTables {
		log.Println("dropping database tables")
		DropTables(db)
	}

	log.Println("initializing database")
	InitDatabase(db)

	log.Println("connecting to redis")
	rdb := CreateRedisClient()

	SeedRand()

	port := fmt.Sprintf(":%v", os.Getenv("PORT"))
	server := &http.Server{
		Addr:    port,
		Handler: CreateRouter(db, rdb),
	}
	go RunHTTPServer(server)

	ExitHandler(server, db, rdb)
}

func parseFlags() (envFile string, dropTables bool) {
	flag.StringVar(&envFile, "e", ENV_FILE, "path to the .env file. use -e nil to prevent .env file from being loaded")
	flag.BoolVar(&dropTables, "d", false, "this will cause all tables to be dropped then recreated during startup")
	flag.Parse()
	return
}

func RunHTTPServer(server *http.Server) {
	log.Printf("binding HTTP server to 0.0.0.0%v\n", server.Addr)

	err := server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Fatalln(err)
	}
}

// Graceful exit
func ExitHandler(server *http.Server, db *sql.DB, rdb *redis.Client) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	<-sigs
	log.Println("receive shutdown signal")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	log.Println("shutting down HTTP server")
	err := server.Shutdown(ctx)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("closing database connections")
	db.Close()

	log.Println("closing Redis client")
	rdb.Close()

	log.Println("shutdown complete")
	os.Exit(0)
}
