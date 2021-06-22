package main

import (
	"context"
	"errors"
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
	CreateIdlemonServer().Run()
}

type IdlemonServer struct {
	HttpServer *http.Server
	Db         *pgxpool.Pool
	Rdb        *redis.Client
	WsHub      *WsHub
	DataCache  *DataCache
}

func CreateIdlemonServer() *IdlemonServer {
	log.Println("starting server")

	envFile := ParseCliFlags()
	LoadEnv(envFile, VERSION)

	// startup context
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	log.Println("creating Postgres pool")
	db, err := CreateDBConn(ctx)
	if err != nil {
		log.Fatalf("failed to create postgres pool: %v\n", err)
	}

	if os.Getenv("DROP_TABLES") == "true" {
		log.Println("dropping database tables")

		if err := DropTables(ctx, db); err != nil {
			log.Fatalf("fail to drop tables: %v\n", err)
		}
	}

	// init data cache
	dataCache := &DataCache{}
	if err := dataCache.Load(); err != nil {
		log.Fatalf("failed to load the data cache: %v\n", err)
	}

	// create database tables
	if os.Getenv("CREATE_TABLES") == "true" {
		log.Println("initializing database")

		if err := CreateDatabaseTables(ctx, db); err != nil {
			log.Fatalf("fail to create database tables: %v\n", err)
		}
	}

	// insert admin user
	if os.Getenv("INSERT_ADMIN") == "true" {
		if err := InsertAdminUser(ctx, db, dataCache); err != nil {
			log.Fatalf("fail to insert admin user: %v\n", err)
		}
	}

	log.Println("connecting to redis")
	rdb := CreateRedisClient(ctx)

	SeedRand()

	// setup WebSocket
	upgrader := websocket.Upgrader{
		ReadBufferSize:  WS_READ_BUFFER_SIZE,
		WriteBufferSize: WS_WRITE_BUFFER_SIZE,
	}
	wsHub := CreateWsHub(upgrader)

	controller := CreateController(db, rdb, wsHub, dataCache)

	port := fmt.Sprintf(":%v", os.Getenv("PORT"))
	httpServer := &http.Server{
		Addr:    port,
		Handler: CreateRouter(controller),
	}

	return &IdlemonServer{
		HttpServer: httpServer,
		Db:         db,
		Rdb:        rdb,
		WsHub:      wsHub,
		DataCache:  dataCache,
	}
}

// Will run the Idlemon server.
func (s IdlemonServer) Run() {
	go s.RunHTTPServer()
	go s.WsHub.Run()
	s.ExitHandler()
}

// Will run the HTTP server. Is blocking.
func (s IdlemonServer) RunHTTPServer() {
	log.Printf("binding HTTP server to 0.0.0.0%v\n", s.HttpServer.Addr)

	err := s.HttpServer.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalln(err)
	}
}

// Graceful exit of Idlemon server on SIGINT or SIGTERM.
func (s IdlemonServer) ExitHandler() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	<-sigs
	log.Println("receive shutdown signal")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	log.Println("shutting down HTTP server")
	err := s.HttpServer.Shutdown(ctx)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("closing WebSocket connections")
	s.WsHub.shutdown <- true
	<-s.WsHub.shutdown // wait for shutdown complete

	log.Println("closing database connections")
	s.Db.Close()

	log.Println("closing Redis client")
	s.Rdb.Close()

	log.Println("shutdown complete")
	os.Exit(0)
}
