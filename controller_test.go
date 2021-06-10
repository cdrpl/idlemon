package main_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/go-redis/redis/v8"
	"golang.org/x/crypto/bcrypt"

	. "github.com/cdrpl/idlemon"
)

/* App Routes */

func TestHealthCheck(t *testing.T) {
	router := CreateRouter(CreateDBConn(), CreateRedisClient())

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("expect status 200, received: %v", status)
	}

	m := map[string]int{"status": 0}
	bytes, _ := json.Marshal(m)

	expect := string(bytes) + "\n"
	if rr.Body.String() != expect {
		t.Errorf("expected body: %v, received: %v", expect, rr.Body.String())
	}
}

func TestVersion(t *testing.T) {
	router := CreateRouter(CreateDBConn(), CreateRedisClient())

	req, err := http.NewRequest("GET", "/version", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("expect status 200, received: %v", status)
	}

	m := map[string]string{
		"server": VERSION,
		"client": os.Getenv("CLIENT_VERSION"),
	}
	bytes, _ := json.Marshal(m)

	expect := string(bytes) + "\n"
	if rr.Body.String() != expect {
		t.Errorf("expected body: %v, received: %v", expect, rr.Body.String())
	}
}

func TestRobots(t *testing.T) {
	router := CreateRouter(CreateDBConn(), CreateRedisClient())

	req, err := http.NewRequest("GET", "/robots.txt", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("expect status 200, received: %v", status)
	}

	// open robots.txt to compare to the response body
	bytes, err := os.ReadFile("robots.txt")
	if err != nil {
		t.Fatalf("fail to open robots.txt: %v", err)
	}

	expect := string(bytes)
	if rr.Body.String() != expect {
		t.Errorf("expected body: %v, received: %v", expect, rr.Body.String())
	}
}

func TestNotFound(t *testing.T) {
	router := CreateRouter(CreateDBConn(), CreateRedisClient())

	req, err := http.NewRequest("GET", "/invalid-route", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("expect status 404, received: %v", status)
	}

	m := map[string]string{"message": "Not Found"}
	bytes, _ := json.Marshal(m)

	expect := string(bytes) + "\n"
	if rr.Body.String() != expect {
		t.Errorf("expected body: %v, received: %v", expect, rr.Body.String())
	}
}

/* Unit Routes */

func TestUnitLock(t *testing.T) {
	db := CreateDBConn()
	rdb := CreateRedisClient()
	router := CreateRouter(db, rdb)

	AuthTest(t, router, "PUT", "/unit/1/toggle-lock")

	token, user, err := AuthenticatedUser(db, rdb)
	if err != nil {
		t.Fatalf("fail to create test user: %v", err)
	}

	unit, err := InsertRandUnit(db, user.ID)
	if err != nil {
		t.Fatalf("insert rand user error: %v", err)
	}

	req := httptest.NewRequest("PUT", fmt.Sprintf("/unit/%v/toggle-lock", unit.ID), nil)
	req.Header.Add("Authorization", fmt.Sprintf("%d:%v", user.ID, token))

	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("expect status 200, received: %v, body: %v", status, rr.Body.String())
	}

	// lock should be updated in database
	var result bool
	err = db.QueryRow("SELECT is_locked FROM unit WHERE id = ?", unit.ID).Scan(&result)
	if err != nil {
		t.Fatalf("db query error: %v", err)
	}

	if result != true {
		t.Errorf("expected unit to be locked in databased, receive: %v", result)
	}

	// should reject unit ID of non existent unit
	req = httptest.NewRequest("PUT", fmt.Sprintf("/unit/%v/toggle-lock", 1234567), nil)
	req.Header.Add("Authorization", fmt.Sprintf("%d:%v", user.ID, token))
	rr = httptest.NewRecorder()

	router.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("expect status 400, received: %v, body: %v", status, rr.Body.String())
	}

	// should reject malformed unit ID
	req = httptest.NewRequest("PUT", fmt.Sprintf("/unit/%v/toggle-lock", "invalid-unit-id"), nil)
	req.Header.Add("Authorization", fmt.Sprintf("%d:%v", user.ID, token))
	rr = httptest.NewRecorder()

	router.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("expect status 400, received: %v, body: %v", status, rr.Body.String())
	}

	// should reject unit not owned by user
	user2, err := InsertRandUser(db)
	if err != nil {
		log.Fatalf("insert rand user error: %v", err)
	}

	unit2, err := InsertRandUnit(db, user2.ID)
	if err != nil {
		log.Fatalf("insert rand user error: %v", err)
	}

	req = httptest.NewRequest("PUT", fmt.Sprintf("/unit/%v/toggle-lock", unit2.ID), nil)
	req.Header.Add("Authorization", fmt.Sprintf("%d:%v", user.ID, token))
	rr = httptest.NewRecorder()

	router.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("expect status 400, received: %v, body: %v", status, rr.Body.String())
	}
}

/* User Routes */

func TestSignUp(t *testing.T) {
	db := CreateDBConn()
	router := CreateRouter(db, CreateRedisClient())

	userInsert := SignUpReq{Name: "name", Email: "name@name.com", Pass: "password"}
	js, _ := json.Marshal(userInsert)

	req := httptest.NewRequest("POST", "/user/sign-up", bytes.NewBuffer(js))
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("expect status 200, received: %v", status)
	}

	m := map[string]int{"status": 0}
	js, _ = json.Marshal(m)

	expect := string(js) + "\n"
	if rr.Body.String() != expect {
		t.Errorf("expected body: %v, received: %v", expect, rr.Body.String())
	}

	// user should exist
	user := User{}
	err := db.QueryRow("SELECT name, email, pass FROM user WHERE name = ?", userInsert.Name).Scan(&user.Name, &user.Email, &user.Pass)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			t.Fatal("user was not present in the database")
		} else {
			t.Fatalf("database query error: %v", err)
		}
	}

	// name should match
	if userInsert.Name != user.Name {
		t.Errorf("name not valid, expect: %v, receive: %v", userInsert.Name, user.Name)
	}

	// email should match
	if userInsert.Email != user.Email {
		t.Errorf("email not valid, expect: %v, receive: %v", userInsert.Email, user.Email)
	}

	// pass should be hashed
	err = bcrypt.CompareHashAndPassword([]byte(user.Pass), []byte(userInsert.Pass))
	if err != nil {
		t.Error("password was not correctly hashed")
	}

	// should return 400 with invalid json
	req = httptest.NewRequest("POST", "/user/sign-up", bytes.NewBuffer([]byte("invalid {json {{ 'f}")))
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("expect status 400, received: %v", status)
	}
}

func TestSignIn(t *testing.T) {
	db := CreateDBConn()
	rdb := CreateRedisClient()
	router := CreateRouter(db, rdb)

	// insert a test user
	user, err := InsertRandUser(db)
	if err != nil {
		t.Fatalf("fail to create test user: %v", err)
	}

	signInReq := SignInReq{Email: user.Email, Pass: user.Pass}
	js, _ := json.Marshal(signInReq)

	req := httptest.NewRequest("POST", "/user/sign-in", bytes.NewBuffer(js))
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("expect status 200, received: %v, body: %v", status, rr.Body.String())
	}

	// sign in response should be valid
	var signInRes SignInRes

	err = json.Unmarshal(rr.Body.Bytes(), &signInRes)
	if err != nil {
		t.Errorf("fail to unmarshal sign in response: %v", err)
	}

	// id should be valid
	if signInRes.User.ID != user.ID {
		t.Errorf("invalid id in response, expected: %v, received: %v", user.ID, signInRes.User.ID)
	}

	// email should be valid
	if signInRes.User.Email != user.Email {
		t.Errorf("invalid email in response, expected: %v, received: %v", user.Email, signInRes.User.Email)
	}

	// sign in response should have no password
	if signInRes.User.Pass != "" {
		t.Errorf("response should have no password, received: %v", signInRes.User.Pass)
	}

	// api token should exist
	idS := fmt.Sprintf("%d", signInRes.User.ID)
	result, err := rdb.Get(context.Background(), idS).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			t.Fatalf("api token was not present in redis")
		} else {
			t.Fatalf("redis error: %v", err)
		}
	}

	if result != signInRes.Token {
		t.Fatalf("api token does not match, expected: %v, receive: %v", signInRes.Token, result)
	}
}

func TestUserRename(t *testing.T) {
	method := "PUT"
	url := "/user/rename"

	db := CreateDBConn()
	rdb := CreateRedisClient()
	router := CreateRouter(db, rdb)

	AuthTest(t, router, method, url)

	token, user, err := AuthenticatedUser(db, rdb)
	if err != nil {
		t.Fatalf("fail to create test user: %v", err)
	}

	newName := "new name dog"
	renameReq := UserRenameReq{Name: newName}
	js, _ := json.Marshal(renameReq)

	req := httptest.NewRequest(method, url, bytes.NewBuffer(js))
	SetAuthorization(req, user.ID, token)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("expect status 200, received: %v, body: %v", status, rr.Body.String())
	}

	response := User{}
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("fail to unmarshal sign the response: %v", err)
	}

	// new name should be returned
	if response.Name != newName {
		t.Errorf("expected response name to equal %v, received: %v", newName, response.Name)
	}

	// name should be changed in database
	userQuery := User{}
	err = db.QueryRow("SELECT name FROM user WHERE id = ?", user.ID).Scan(&userQuery.Name)
	if err != nil {
		t.Errorf("expected name in database to equal %v, received: %v", newName, userQuery.Name)
	}
}
