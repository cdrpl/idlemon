package main_test

import (
	"context"
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
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/julienschmidt/httprouter"
	"golang.org/x/crypto/bcrypt"
)

var router *httprouter.Router
var dataCache DataCache
var db *pgxpool.Pool
var rdb *redis.Client

func TestMain(m *testing.M) {
	os.Setenv("ENV", "test")
	os.Setenv("DB_NAME", "test")
	LoadEnv(ENV_FILE, VERSION)

	dataCache = DataCache{}

	if err := dataCache.Load(); err != nil {
		log.Fatalf("fail to load data cache: %v\n", err)
	}

	var err error
	db, err = CreateDBConn(context.Background())
	if err != nil {
		log.Fatalf("fail to create DB connection: %v\n", err)
		os.Exit(1)
	}

	err = DropTables(context.Background(), db)
	if err != nil {
		log.Fatalf("fail to drop tables: %v\n", err)
		os.Exit(1)
	}

	log.SetOutput(ioutil.Discard)

	err = InitDatabase(context.Background(), db, dataCache)
	if err != nil {
		log.SetOutput(os.Stdout)
		log.Fatalf("fail to init database: %v\n", err)
		os.Exit(1)
	}

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

	email = email + "@fakemockemailfake.com"
	user := CreateUser(dataCache, name, email, pass)

	return user, nil
}

func InsertRandUser(db *pgxpool.Pool) (User, error) {
	user, err := RandUser()
	if err != nil {
		return User{}, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(user.Pass), BCRYPT_COST)
	if err != nil {
		return User{}, err
	}

	id, err := InsertUser(context.Background(), db, CreateUser(dataCache, user.Name, user.Email, string(hash)))
	if err != nil {
		return User{}, err
	}

	user.ID = id
	return user, nil
}

func AuthenticatedUser(db *pgxpool.Pool, rdb *redis.Client) (string, User, error) {
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
func InsertRandUnit(ctx context.Context, db *pgxpool.Pool, user *User) (Unit, error) {
	template := RandUnitTemplateID(dataCache)

	unit := CreateUnit(template)
	unit = AddUnitToUser(user, unit)
	err := UpdateUser(ctx, db, *user)

	return unit, err
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
