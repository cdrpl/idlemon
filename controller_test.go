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

	token, user, err := AuthenticatedUser(context.TODO(), db, rdb)
	if err != nil {
		t.Fatalf("fail to create test user: %v", err)
	}

	req := httptest.NewRequest("PUT", "/campaign/collect", nil)
	SetAuthorization(req, user.Id, token)
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

	if res.Transactions[0].Amount != 0 {
		t.Fatalf("collected resources should be 0: %v", res)
	}

	if res.Transactions[1].Amount != 0 {
		t.Fatalf("collected resources should be 0: %v", res)
	}

	if res.Transactions[2].Amount != 0 {
		t.Fatalf("collected resources should be 0: %v", res)
	}
}

/* Daily Quest Routes */

func TestCompleteDailyQuest(t *testing.T) {
	url := fmt.Sprintf("/daily-quest/%v/complete", DAILY_QUEST_SIGN_IN)
	AuthTest(t, router, "PUT", url)

	token, user, err := AuthenticatedUser(context.TODO(), db, rdb)
	if err != nil {
		t.Fatalf("fail to create test user: %v", err)
	}

	// test requirements not met
	req := httptest.NewRequest("PUT", url, nil)
	SetAuthorization(req, user.Id, token)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Fatalf("expect status 200, received: %v, body: %v", status, rr.Body.String())
	}

	var response DailyQuestCompleteRes

	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("fail to unmarshal response: %v", err)
	}

	// status should be 2 for requirements not met
	if response.Status != 2 {
		t.Fatalf("expect response status to be 1, receive: %v", response.Status)
	}

	// test successful collect
	query := "UPDATE daily_quest_progress SET count = 1 WHERE (user_id = $1 AND daily_quest_id = $2)"
	_, err = db.Exec(context.TODO(), query, user.Id, DAILY_QUEST_SIGN_IN)
	if err != nil {
		t.Fatalf("fail to update daily_quest_progress table: %v", err)
	}

	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Fatalf("expect status 200, received: %v, body: %v", status, rr.Body.String())
	}

	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("fail to unmarshal response: %v", err)
	}

	// status should be 0 for successful completion
	if response.Status != 0 {
		t.Fatalf("expect response status to equal 0, receive: %v", response.Status)
	}

	expect := dataCache.DailyQuests[DAILY_QUEST_SIGN_IN].Transaction.Type
	if response.Transaction.Type != expect {
		t.Fatalf("unexpected transaction type, expect: %v, receive: %v", expect, response.Transaction.Type)
	}

	expect = dataCache.DailyQuests[DAILY_QUEST_SIGN_IN].Transaction.Amount
	if response.Transaction.Amount != expect {
		t.Fatalf("unexpected transaction amount, expect: %v, receive: %v", expect, response.Transaction.Amount)
	}

	// database should be updated
	var amount int
	query = "SELECT amount FROM resources WHERE (user_id = $1 AND type = $2)"
	err = db.QueryRow(context.TODO(), query, user.Id, RESOURCE_GEMS).Scan(&amount)
	if err != nil {
		t.Fatalf("fail to query resources row: %v", err)
	}

	if amount != response.Transaction.Amount {
		t.Fatalf("unexpected amount in database, expect: %v, receive: %v", response.Transaction.Amount, amount)
	}

	// test already collected
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Fatalf("expect status 200, received: %v, body: %v", status, rr.Body.String())
	}

	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("fail to unmarshal response: %v", err)
	}

	// status should be 10 for already collected
	if response.Status != 1 {
		t.Fatalf("expect response status to equal 1, receive: %v", response.Status)
	}
}

/* Unit Routes */

func TestUnitLockRoute(t *testing.T) {
	AuthTest(t, router, "PUT", "/unit/1/toggle-lock")

	token, user, err := AuthenticatedUser(context.TODO(), db, rdb)
	if err != nil {
		t.Fatalf("fail to create test user: %v", err)
	}

	unit, err := InsertRandUnit(context.TODO(), db, user.Id)
	if err != nil {
		t.Fatalf("insert rand unit error: %v", err)
	}

	req := httptest.NewRequest("PUT", fmt.Sprintf("/unit/%v/toggle-lock", unit.Id), nil)
	SetAuthorization(req, user.Id, token)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Fatalf("expect status 200, received: %v, body: %v", status, rr.Body.String())
	}

	// unit must be locked in database
	var isLocked bool

	query := "SELECT is_locked FROM units WHERE id = $1"
	err = db.QueryRow(context.TODO(), query, unit.Id).Scan(&isLocked)
	if err != nil {
		t.Fatalf("fail to query units table: %v", err)
	}

	if !isLocked {
		t.Error("unit is not locked in the database")
	}
}

/* User Routes */

func TestUserSignUpRoute(t *testing.T) {
	userInsert := SignUpReq{Name: "name", Email: "name@name.com", Pass: "password"}
	js, _ := json.Marshal(userInsert)

	req := httptest.NewRequest("POST", "/user/sign-up", bytes.NewBuffer(js))
	req.Header.Set("Content-Type", "application/json")
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
	err := db.QueryRow(context.TODO(), query, userInsert.Name).Scan(&user.Id, &user.Name, &user.Email, &user.Pass, &user.CreatedAt)
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
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("expect status 400, received: %v, %v", status, rr.Body.String())
	}
}

func TestUserSignInRoute(t *testing.T) {
	user, err := InsertRandUser(context.TODO(), db)
	if err != nil {
		t.Fatalf("fail to create test user: %v", err)
	}

	signInReq := SignInReq{Email: user.Email, Pass: user.Pass}
	js, _ := json.Marshal(signInReq)

	req := httptest.NewRequest("POST", "/user/sign-in", bytes.NewBuffer(js))
	req.Header.Set("Content-Type", "application/json")
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
	if signInRes.User.Id != user.Id {
		t.Errorf("invalid id in response, expected: %v, received: %v", user.Id, signInRes.User.Id)
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
	result, err := rdb.Get(context.TODO(), user.Id.String()).Result()
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
}

func TestUserRenameRoute(t *testing.T) {
	method := "PUT"
	url := "/user/rename"

	AuthTest(t, router, method, url)

	token, user, err := AuthenticatedUser(context.TODO(), db, rdb)
	if err != nil {
		t.Fatalf("fail to create test user: %v", err)
	}

	newName := "new name dog"
	renameReq := UserRenameReq{Name: newName}
	js, _ := json.Marshal(renameReq)

	req := httptest.NewRequest(method, url, bytes.NewBuffer(js))
	req.Header.Set("Content-Type", "application/json")
	SetAuthorization(req, user.Id, token)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("expect status 200, received: %v, body: %v", status, rr.Body.String())
	}

	// name should be changed in database
	userQuery := User{}
	err = db.QueryRow(context.TODO(), "SELECT name FROM users WHERE id = $1", user.Id).Scan(&userQuery.Name)
	if err != nil {
		t.Errorf("expected name in database to equal %v, received: %v", newName, userQuery.Name)
	}
}
