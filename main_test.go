package main_test

import (
	"context"
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	. "github.com/cdrpl/idlemon-server"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
	"github.com/julienschmidt/httprouter"
)

var router *httprouter.Router
var dataCache *DataCache
var db *sql.DB
var rdb *redis.Client

func TestMain(m *testing.M) {
	log.SetOutput(ioutil.Discard)

	os.Setenv("ENV", "test")
	os.Setenv("DB_NAME", "test")
	LoadEnv(ENV_FILE, VERSION)

	dataCache = &DataCache{}

	if err := dataCache.Load(); err != nil {
		log.SetOutput(os.Stdout)
		log.Fatalf("fail to load data cache: %v\n", err)
	}

	db = CreateDBConn()
	DropTables(db)
	InitDatabase(context.Background(), db, dataCache)

	rdb = CreateRedisClient()

	SeedRand()

	upgrader := websocket.Upgrader{
		ReadBufferSize:  WS_READ_BUFFER_SIZE,
		WriteBufferSize: WS_WRITE_BUFFER_SIZE,
	}
	wsHub := CreateWsHub(upgrader)
	go wsHub.Run()

	router = CreateRouter(db, rdb, dataCache, wsHub)

	os.Exit(m.Run())
}

/* Test Helpers */

func RandUser() (User, error) {
	user := User{}

	name, err := GenerateToken(16)
	if err != nil {
		return User{}, err
	}

	email, err := GenerateToken(16)
	if err != nil {
		return User{}, err
	}

	pass, err := GenerateToken(16)
	if err != nil {
		return User{}, err
	}

	user.Name = name
	user.Email = email + "@fakemockemailfake.com"
	user.Pass = pass

	return user, nil
}

func InsertRandUser(db *sql.DB) (User, error) {
	user, err := RandUser()
	if err != nil {
		return User{}, err
	}

	id, err := InsertUser(context.Background(), db, dataCache, user.Name, user.Email, user.Pass)
	if err != nil {
		return User{}, err
	}

	user.ID = id
	return user, nil
}

func AuthenticatedUser(db *sql.DB, rdb *redis.Client) (string, User, error) {
	user, err := InsertRandUser(db)
	if err != nil {
		return "", User{}, err
	}

	token, err := CreateApiToken(context.Background(), rdb, user.ID)
	if err != nil {
		return "", User{}, err
	}

	return token, user, err
}

// Create a random unit and insert it into the table.
func InsertRandUnit(db *sql.DB, userID int) (Unit, error) {
	template, err := RandUnitTemplateID(db)
	if err != nil {
		return Unit{}, err
	}

	return InsertUnit(context.Background(), db, userID, template)
}

// Will send an HTTP request without an Authorization header then call t.Fatalf if 401 not received.
func AuthTest(t *testing.T, router *httprouter.Router, method string, url string) {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(method, url, nil)

	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusUnauthorized {
		t.Errorf("unauthorized should be returned if no authorization header is given")
		t.Fatalf("expect status 401, received: %v, body: %v", status, rr.Body.String())
	}
}

func SetAuthorization(req *http.Request, userID int, token string) {
	req.Header.Add("Authorization", fmt.Sprintf("%d:%v", userID, token))
}
