package main_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"testing"

	. "github.com/cdrpl/idlemon-server"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v4/pgxpool"
)

const (
	PORT = "3100"
	HOST = "http://localhost:" + PORT
)

var idlemonServer *IdlemonServer

func TestMain(m *testing.M) {
	os.Setenv("ENV", "test")
	os.Setenv("PORT", PORT)
	os.Setenv("DB_NAME", "test")
	os.Setenv("CREATE_TABLES", "true")
	os.Setenv("DROP_TABLES", "true")
	os.Setenv("INSERT_ADMIN", "false")

	idlemonServer = CreateIdlemonServer()
	go idlemonServer.Run()

	log.SetOutput(ioutil.Discard)

	os.Exit(m.Run())
}

/* Test Helpers */

func RandUser(t *testing.T, dataCache *DataCache) User {
	name, err := GenerateToken(16)
	if err != nil {
		t.Fatalf("fail to generate rand name: %v", err)
	}

	email, err := GenerateToken(16)
	if err != nil {
		t.Fatalf("fail to generate rand email: %v", err)
	}

	pass, err := GenerateToken(16)
	if err != nil {
		t.Fatalf("fail to generate rand pass: %v", err)
	}

	email = email + "@fakemockemailfake.com"
	return CreateUser(dataCache, name, email, pass)
}

func InsertRandUser(t *testing.T, db *pgxpool.Pool, dataCache *DataCache) User {
	user := RandUser(t, dataCache)
	user = CreateUser(dataCache, user.Name, user.Email, user.Pass)

	tx, err := db.Begin(context.Background())
	if err != nil {
		t.Fatalf("fail to begin transaction: %v", err)
	}
	defer tx.Rollback(context.Background())

	if err := InsertUser(context.Background(), tx, dataCache, user); err != nil {
		t.Fatalf("fail to insert used: %v", err)
	}

	if err := tx.Commit(context.Background()); err != nil {
		t.Fatalf("fail to commit transaction: %v", err)
	}

	return user
}

func AuthenticatedUser(t *testing.T, db *pgxpool.Pool, rdb *redis.Client, dataCache *DataCache) (string, User) {
	user := InsertRandUser(t, db, dataCache)

	token, err := CreateApiToken(context.Background(), rdb, user.Id)
	if err != nil {
		t.Fatalf("fail to create API token: %v", err)
	}

	return token, user
}

// Create a random unit and insert it into the table.
func InsertRandUnit(t *testing.T, db *pgxpool.Pool, dataCache *DataCache, userId uuid.UUID) Unit {
	template := RandUnitTemplateID(dataCache)

	tx, err := db.Begin(context.Background())
	if err != nil {
		t.Fatalf("fail to begin transaction: %v", err)
	}
	defer tx.Rollback(context.Background())

	unit := CreateUnit(template)

	if err = InsertUnit(context.Background(), tx, userId, unit); err != nil {
		t.Fatalf("fail to insert unit: %v", err)
	}

	if err := tx.Commit(context.Background()); err != nil {
		t.Fatalf("fail to commit transaction: %v", err)
	}

	return unit
}

func CreateWsConn(t *testing.T, userId uuid.UUID, token string) *websocket.Conn {
	authorization := fmt.Sprintf("%v:%v", userId.String(), token)
	host := fmt.Sprintf("127.0.0.1:%v", PORT)

	headers := http.Header{}
	headers.Add("authorization", authorization)

	u := url.URL{Scheme: "ws", Host: host, Path: "/ws"}

	c, _, err := websocket.DefaultDialer.Dial(u.String(), headers)
	if err != nil {
		t.Fatalf("fail to connect to WebSocket server: %v", err)
	}

	return c
}

// Send GET request without an authorization header.
func GetRequest(t *testing.T, url string) *http.Response {
	response, err := http.Get(HOST + url)
	if err != nil {
		t.Fatalf("fail to send get request: %v", url)
	}

	return response
}

// Send POST request without an authorization header.
func PostRequest(t *testing.T, url string, body RequestDTO) *http.Response {
	bodyB, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("fail to marshall request body: %v", err)
	}

	response, err := http.Post(HOST+url, "application/json", bytes.NewReader(bodyB))
	if err != nil {
		t.Fatalf("fail to send get request: %v", url)
	}

	return response
}

// Send an HTTP request. First checks that auth middleware catches missing authorization header. Then sends the real request with the header included.
func SendRequest(t *testing.T, method string, url string, userId uuid.UUID, token string, body RequestDTO) *http.Response {
	bodyB := make([]byte, 0)

	if body != nil {
		var err error
		if bodyB, err = json.Marshal(body); err != nil {
			t.Fatalf("fail to marshall request body: %v", err)
		}
	}

	httpClient := http.Client{}

	// create HTTP request
	req, err := http.NewRequest(method, HOST+url, bytes.NewReader(bodyB))
	if err != nil {
		t.Fatalf("fail to create new request: %v", err)
	}

	// test auth middleware by sending request without authorization header
	authResponse, err := httpClient.Do(req)
	if err != nil {
		t.Fatalf("fail to send request: %v", err)
	}

	// 401 unauthorized should be returned
	if authResponse.StatusCode != http.StatusUnauthorized {
		body := ReadResponseBody(t, authResponse)
		t.Errorf("unauthorized should be returned if no authorization header is given")
		t.Fatalf("expect status 401, received: %v, body: %v", authResponse.StatusCode, body)
	}

	// recreate request since body will have already been read
	req, err = http.NewRequest(method, HOST+url, bytes.NewReader(bodyB))
	if err != nil {
		t.Fatalf("fail to create new request: %v", err)
	}

	// add headers required for request
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("%v:%v", userId.String(), token))

	// send the real HTTP request
	response, err := httpClient.Do(req)
	if err != nil {
		t.Fatalf("fail to send request: %v", err)
	}

	return response
}

// Will read the response body, close it, then return it in string format.
func ReadResponseBody(t *testing.T, r *http.Response) string {
	bodyB, err := io.ReadAll(r.Body)
	if err != nil {
		t.Fatalf("fail to read response body: %v", err)
	}
	defer r.Body.Close()

	return string(bodyB)
}
