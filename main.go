package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v4/pgxpool"

	_ "embed"
)

//go:embed robots.txt
var robots string

//go:embed database_up.sql
var upSql string

//go:embed database_down.sql
var downSql string

//go:embed unit_templates.json
var unitTemplatesJson string

func main() {
	log.Println("starting server")

	envFile, dropTables := parseFlags()
	LoadEnv(envFile, VERSION)

	log.Println("creating Postgres pool")
	db, err := CreateDBConn(context.Background())
	if err != nil {
		log.Fatalf("failed to create postgres pool: %v\n", err)
	}

	if dropTables {
		log.Println("dropping database tables")

		if err := DropTables(context.Background(), db); err != nil {
			log.Fatalf("fail to drop tables: %v\n", err)
		}
	}

	// init data cache
	dc := DataCache{}
	if err := dc.Load(); err != nil {
		log.Fatalf("failed to load the data cache: %v\n", err)
	}

	// init database
	if os.Getenv("INIT_DATABASE") == "true" {
		log.Println("initializing database")

		if err := InitDatabase(context.Background(), db, dc); err != nil {
			log.Fatalf("fail to init database: %v", err)
		}
	}

	log.Println("connecting to redis")
	rdb := CreateRedisClient()

	SeedRand()

	// setup WebSocket
	upgrader := websocket.Upgrader{
		ReadBufferSize:  WS_READ_BUFFER_SIZE,
		WriteBufferSize: WS_WRITE_BUFFER_SIZE,
	}
	wsHub := CreateWsHub(upgrader)
	go wsHub.Run()

	port := fmt.Sprintf(":%v", os.Getenv("PORT"))
	server := &http.Server{
		Addr:    port,
		Handler: CreateRouter(db, rdb, dc, wsHub),
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
func ExitHandler(server *http.Server, db *pgxpool.Pool, rdb *redis.Client) {
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
