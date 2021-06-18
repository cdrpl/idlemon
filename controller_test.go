package main_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v4"
	"golang.org/x/crypto/bcrypt"

	. "github.com/cdrpl/idlemon-server"
)

/* App Routes */

func TestHealthCheckRoute(t *testing.T) {
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

func TestVersionRoute(t *testing.T) {
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

func TestRobotsRoute(t *testing.T) {
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

func TestNotFoundRoute(t *testing.T) {
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

/* Campaign Routes */

func TestCampaignCollectRoute(t *testing.T) {
	AuthTest(t, router, "PUT", "/campaign/collect")

	token, user, err := AuthenticatedUser(db, rdb)
	if err != nil {
		t.Fatalf("fail to create test user: %v", err)
	}

	req := httptest.NewRequest("PUT", "/campaign/collect", nil)
	req.Header.Add("Authorization", fmt.Sprintf("%d:%v", user.ID, token))
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("expect status 200, received: %v, body: %v", status, rr.Body.String())
	}

	var res CampaignCollectRes
	err = json.Unmarshal(rr.Body.Bytes(), &res)
	if err != nil {
		t.Fatalf("fail to unmarshall response body: %v", err)
	}

	if res.Exp != 0 || res.Gold != 0 || res.ExpStones != 0 {
		t.Fatalf("collected resources should be 0: %v", res)
	}
}

/* Unit Routes */

func TestUnitLockRoute(t *testing.T) {
	AuthTest(t, router, "PUT", "/unit/1/toggle-lock")

	token, user, err := AuthenticatedUser(db, rdb)
	if err != nil {
		t.Fatalf("fail to create test user: %v", err)
	}

	unit, err := InsertRandUnit(context.Background(), db, &user)
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
}

/* User Routes */

func TestUserSignUpRoute(t *testing.T) {
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
	query := "SELECT id, name, email, pass, created_at FROM users WHERE name = $1"
	err := db.QueryRow(context.Background(), query, userInsert.Name).Scan(&user.ID, &user.Name, &user.Email, &user.Pass, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			t.Fatal("user was not present in the database")
		} else {
			t.Fatalf("test user sign up route query user error: %v", err)
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

	// created at should be accurate
	if time.Since(user.CreatedAt) > time.Second {
		t.Errorf("more than a second has passed since user created at: %v", user.CreatedAt)
	}

	// should return 400 with invalid json
	req = httptest.NewRequest("POST", "/user/sign-up", bytes.NewBuffer([]byte("invalid {json {{ 'f}")))
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("expect status 400, received: %v", status)
	}
}

func TestUserSignInRoute(t *testing.T) {
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
		t.Fatalf("expect status 200, received: %v, body: %v", status, rr.Body.String())
	}

	// sign in response should be valid
	var signInRes SignInRes

	err = json.Unmarshal(rr.Body.Bytes(), &signInRes)
	if err != nil {
		t.Errorf("fail to unmarshal sign in response: %v", err)
	}

	// token should not be empty
	if signInRes.Token == "" {
		t.Error("token was empty")
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

	if len(signInRes.Resources) == 0 {
		t.Error("resources should not be empty")
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

	// validate response json keys
	var m map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &m)
	if err != nil {
		t.Errorf("fail to unmarshal sign in response into map: %v", err)
	}

	if _, ok := m["token"]; !ok {
		t.Error("sign in response didn't have a token property")
	}

	if _, ok := m["user"]; !ok {
		t.Error("sign in response didn't have a user property")
	}

	if _, ok := m["resources"]; !ok {
		t.Error("sign in response didn't have a resources property")
	}
}

func TestUserRenameRoute(t *testing.T) {
	method := "PUT"
	url := "/user/rename"

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
	err = db.QueryRow(context.Background(), "SELECT name FROM users WHERE id = $1", user.ID).Scan(&userQuery.Name)
	if err != nil {
		t.Errorf("expected name in database to equal %v, received: %v", newName, userQuery.Name)
	}
}
