package main_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
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
	r := GetRequest(t, "/")
	body := ReadResponseBody(t, r)

	m := map[string]int{"status": 0}

	bytes, _ := json.Marshal(m)
	expect := string(bytes) + "\n"

	if body != expect {
		t.Errorf("expect body: %v, received: %v", expect, body)
	}
}

func TestVersionRoute(t *testing.T) {
	r := GetRequest(t, "/version")
	body := ReadResponseBody(t, r)

	m := map[string]string{
		"server": VERSION,
		"client": os.Getenv("CLIENT_VERSION"),
	}
	bytes, _ := json.Marshal(m)

	expect := string(bytes) + "\n"

	if body != expect {
		t.Errorf("expect body: %v, received: %v", expect, body)
	}
}

func TestRobotsRoute(t *testing.T) {
	r := GetRequest(t, "/robots.txt")
	body := ReadResponseBody(t, r)

	// open robots.txt to compare to the response body
	bytes, err := os.ReadFile("robots.txt")
	if err != nil {
		t.Fatalf("fail to open robots.txt: %v", err)
	}

	expect := string(bytes)

	if body != expect {
		t.Errorf("expect body: %v, received: %v", expect, body)
	}
}

func TestNotFoundRoute(t *testing.T) {
	r := GetRequest(t, "/invalid-route")
	body := ReadResponseBody(t, r)

	m := map[string]string{"message": "Not Found"}
	bytes, _ := json.Marshal(m)

	expect := string(bytes) + "\n"
	if body != expect {
		t.Errorf("expect body: %v, received: %v", expect, body)
	}
}

/* Campaign Routes */

func TestCampaignCollectRoute(t *testing.T) {
	method := "PUT"
	url := "/campaign/collect"

	token, user := AuthenticatedUser(t, idlemonServer.Db, idlemonServer.Rdb, idlemonServer.DataCache)

	response := SendRequest(t, method, url, user.Id, token, nil)
	body := ReadResponseBody(t, response)

	var res CampaignCollectRes

	if err := json.Unmarshal([]byte(body), &res); err != nil {
		t.Fatalf("fail to unmarshal response body: %v", err)
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

/* Chat Routs */

func TestChatMessageSendRoute(t *testing.T) {
	method := "POST"
	url := "/chat/message/send"

	// user used for sending chat message
	token, user := AuthenticatedUser(t, idlemonServer.Db, idlemonServer.Rdb, idlemonServer.DataCache)

	// secondary user used for listening for chat message from the WebSocket server
	token2, user2 := AuthenticatedUser(t, idlemonServer.Db, idlemonServer.Rdb, idlemonServer.DataCache)

	//create WebSocket connection, expect to receive message sent by user 1
	wsConn := CreateWsConn(t, user2.Id, token2)

	// channels for receiving data from WebSocket read
	errChan := make(chan error)
	wsMsg := make(chan ChatMessage)

	// read message from wsConn
	go func() {
		var webSocketMessage WebSocketMessage
		var wsChatMsg ChatMessage

		if err := wsConn.SetReadDeadline(time.Now().Add(time.Second * 2)); err != nil {
			errChan <- fmt.Errorf("fail to set WebSocket read deadline: %w", err)
			return
		}

		if err := wsConn.ReadJSON(&webSocketMessage); err != nil {
			errChan <- fmt.Errorf("fail to read WebSocket message: %w", err)
			return
		}

		if err := json.Unmarshal(webSocketMessage.Data, &wsChatMsg); err != nil {
			errChan <- fmt.Errorf("fail to unmarhsal WebSocketChatMessage: %w", err)
			return
		}

		wsMsg <- wsChatMsg
	}()

	// chat message send request
	request := &ChatMessageSendReq{Message: "Hello, World! :)"}
	response := SendRequest(t, method, url, user.Id, token, request)
	body := ReadResponseBody(t, response)

	if response.StatusCode != http.StatusOK {
		t.Fatalf("expect status code 200, receive: %v, body: %v", response.StatusCode, body)
	}

	// message should exist in database
	var id int
	var senderName string

	query := "SELECT id, sender_name FROM chat_messages WHERE (user_id = $1 AND message = $2)"
	err := idlemonServer.Db.QueryRow(context.Background(), query, user.Id, request.Message).Scan(&id, &senderName)
	if err != nil {
		t.Fatalf("fail to fetch chat message from databse: %v", err)
	}

	// sender name in database must be the same as the user name
	if senderName != user.Name {
		t.Fatalf("sender_name in database is invalid, expect: %v, receive: %v", user.Name, senderName)
	}

	// validate data received from WebSocket connection
	select {
	case err := <-errChan:
		t.Fatal(err)

	case wsMsg := <-wsMsg:
		if wsMsg.Id != id {
			t.Fatalf("id on WebSocket message was invalid, expect: %v, receive: %v", id, wsMsg.Id)
		}

		if wsMsg.UserId != user.Id {
			t.Fatalf("user ID on WebSocket message was invalid, expect: %v, receive: %v", user.Id, wsMsg.UserId)
		}

		if wsMsg.SenderName != user.Name {
			t.Fatalf("sender name on WebSocket message was invalid, expect: %v, receive: %v", user.Name, wsMsg.SenderName)
		}

		if wsMsg.Message != request.Message {
			t.Fatalf("message from WebSocket was invalid, expect: %v, receive: %v", request.Message, wsMsg.Message)
		}
	}
}

/* Daily Quest Routes */

func TestDailyQuestComplete(t *testing.T) {
	method := "PUT"
	url := fmt.Sprintf("/daily-quest/%v/complete", DAILY_QUEST_SIGN_IN)

	token, user := AuthenticatedUser(t, idlemonServer.Db, idlemonServer.Rdb, idlemonServer.DataCache)

	// send response expect requirements not met
	response := SendRequest(t, method, url, user.Id, token, nil)
	body := ReadResponseBody(t, response)

	// expect status 200
	if response.StatusCode != http.StatusOK {
		t.Fatalf("expect status 200, received: %v, body: %v", response.StatusCode, body)
	}

	var questCompleteRes DailyQuestCompleteRes

	// unmarshal requirements not met response
	if err := json.Unmarshal([]byte(body), &questCompleteRes); err != nil {
		t.Fatalf("fail to unmarshal response: %v", err)
	}

	// status should be 2 for requirements not met
	if questCompleteRes.Status != 2 {
		t.Fatalf("expect response status to be 1, receive: %v", response.Status)
	}

	// set quest progress to 100% to test successful completion
	query := "UPDATE daily_quest_progress SET count = 1 WHERE (user_id = $1 AND daily_quest_id = $2)"
	_, err := idlemonServer.Db.Exec(context.Background(), query, user.Id, DAILY_QUEST_SIGN_IN)
	if err != nil {
		t.Fatalf("fail to update daily_quest_progress table: %v", err)
	}

	// send response expect success
	response = SendRequest(t, method, url, user.Id, token, nil)
	body = ReadResponseBody(t, response)

	// expect status OK
	if response.StatusCode != http.StatusOK {
		t.Fatalf("expect status 200, received: %v, body: %v", response.StatusCode, body)
	}

	// unmarshal success response
	if err := json.Unmarshal([]byte(body), &questCompleteRes); err != nil {
		t.Fatalf("fail to unmarshal response: %v", err)
	}

	// status should be 0 for successful completion
	if questCompleteRes.Status != 0 {
		t.Fatalf("expect response status to equal 0, receive: %v", questCompleteRes.Status)
	}

	expect := idlemonServer.DataCache.DailyQuests[DAILY_QUEST_SIGN_IN].Transaction.Type
	if questCompleteRes.Transaction.Type != expect {
		t.Fatalf("unexpected transaction type, expect: %v, receive: %v", expect, questCompleteRes.Transaction.Type)
	}

	expect = idlemonServer.DataCache.DailyQuests[DAILY_QUEST_SIGN_IN].Transaction.Amount
	if questCompleteRes.Transaction.Amount != expect {
		t.Fatalf("unexpected transaction amount, expect: %v, receive: %v", expect, questCompleteRes.Transaction.Amount)
	}

	// check resources in database
	var amount int

	// select amount of gems from database
	query = "SELECT amount FROM resources WHERE (user_id = $1 AND type = $2)"
	err = idlemonServer.Db.QueryRow(context.Background(), query, user.Id, RESOURCE_GEMS).Scan(&amount)
	if err != nil {
		t.Fatalf("fail to query resources row: %v", err)
	}

	// expect gems in database to equal amount gained from completing the quest
	if amount != questCompleteRes.Transaction.Amount {
		t.Fatalf("unexpected amount in database, expect: %v, receive: %v", questCompleteRes.Transaction.Amount, amount)
	}

	// test already collected
	response = SendRequest(t, method, url, user.Id, token, nil)
	body = ReadResponseBody(t, response)

	if response.StatusCode != http.StatusOK {
		t.Fatalf("expect status 200, received: %v, body: %v", response.StatusCode, body)
	}

	// unmarshal already collected response
	if err := json.Unmarshal([]byte(body), &questCompleteRes); err != nil {
		t.Fatalf("fail to unmarshal response: %v", err)
	}

	// status should be 10 for already collected
	if questCompleteRes.Status != 1 {
		t.Fatalf("expect response status to equal 1, receive: %v", questCompleteRes.Status)
	}
}

/* Summon Routes */

func TestSummonUnit(t *testing.T) {
	method := "PUT"
	url := "/summon/unit"

	token, user := AuthenticatedUser(t, idlemonServer.Db, idlemonServer.Rdb, idlemonServer.DataCache)

	response := SendRequest(t, method, url, user.Id, token, nil)
	body := ReadResponseBody(t, response)

	// should receive 400 for not enough gems
	if response.StatusCode != http.StatusBadRequest {
		t.Fatalf("expect status 400, received: %v, body: %v", response.StatusCode, body)
	}

	// give user some gems
	query := "UPDATE resources SET amount = $1 WHERE (user_id = $2 AND type = $3)"
	_, err := idlemonServer.Db.Exec(context.Background(), query, UNIT_SUMMON_COST, user.Id, RESOURCE_GEMS)
	if err != nil {
		t.Fatalf("fail to update resources table: %v", err)
	}

	// send request which should succeed
	response = SendRequest(t, method, url, user.Id, token, nil)
	body = ReadResponseBody(t, response)

	// should receive 200
	if response.StatusCode != http.StatusOK {
		t.Fatalf("expect status 200, received: %v, body: %v", response.StatusCode, body)
	}

	// unmarshal success response
	var summonUnitRes SummonUnitRes

	if err := json.Unmarshal([]byte(body), &summonUnitRes); err != nil {
		t.Fatalf("fail to unmarshal successful response: %v", err)
	}

	// unit should exist in database
	var template int

	// fetch template of the summoned unit from the database
	query = "SELECT template FROM units WHERE (id = $1 AND user_id = $2)"
	err = idlemonServer.Db.QueryRow(context.Background(), query, summonUnitRes.Unit.Id, user.Id).Scan(&template)
	if err != nil {
		t.Errorf("fail to query units table: %v", err)
	}

	if template != summonUnitRes.Unit.Template {
		t.Errorf("unit template in database did not match response, expect: %v, receive: %v", summonUnitRes.Unit.Template, template)
	}

	// resource table should be updated
	var amount int

	// query gems amount from the database
	query = "SELECT amount FROM resources WHERE (user_id = $1 AND type = $2)"
	_, err = idlemonServer.Db.Exec(context.Background(), query, user.Id, RESOURCE_GEMS)
	if err != nil {
		t.Fatalf("fail to query resources table: %v", err)
	}

	// amount of gems remaining should be 0
	if amount != 0 {
		t.Fatalf("expect amount in database to equal 0, receive: %v", amount)
	}
}

/* Unit Routes */

func TestUnitLockRoute(t *testing.T) {
	token, user := AuthenticatedUser(t, idlemonServer.Db, idlemonServer.Rdb, idlemonServer.DataCache)
	unit := InsertRandUnit(t, idlemonServer.Db, idlemonServer.DataCache, user.Id)

	method := "PUT"
	url := "/unit/" + unit.Id.String() + "/toggle-lock"

	response := SendRequest(t, method, url, user.Id, token, nil)
	body := ReadResponseBody(t, response)

	if response.StatusCode != http.StatusOK {
		t.Fatalf("expect status 200, received: %v, body: %v", response.StatusCode, body)
	}

	// unit must be locked in database
	var isLocked bool

	// query units table
	query := "SELECT is_locked FROM units WHERE id = $1"
	err := idlemonServer.Db.QueryRow(context.Background(), query, unit.Id).Scan(&isLocked)
	if err != nil {
		t.Fatalf("fail to query units table: %v", err)
	}

	if !isLocked {
		t.Error("unit is not locked in the database")
	}
}

/* User Routes */

func TestUserSignUpRoute(t *testing.T) {
	url := "/user/sign-up"

	signUpReq := &SignUpReq{Name: "name", Email: "name@name.com", Pass: "password"}
	response := PostRequest(t, url, signUpReq)
	body := ReadResponseBody(t, response)

	if response.StatusCode != http.StatusOK {
		t.Errorf("expect status 200, received: %v, body: %v", response.StatusCode, body)
	}

	user := User{}

	// user should be inserted in the database
	query := "SELECT id, name, email, pass, exp, created_at FROM users WHERE name = $1"
	err := idlemonServer.Db.QueryRow(context.Background(), query, signUpReq.Name).Scan(&user.Id, &user.Name, &user.Email, &user.Pass, &user.Exp, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			t.Fatal("user was not present in the database")
		} else {
			t.Fatalf("test user sign up route query user error: %v", err)
		}
	}

	// name should match
	if signUpReq.Name != user.Name {
		t.Fatalf("name not valid, expect: %v, receive: %v", signUpReq.Name, user.Name)
	}

	// email should match
	if signUpReq.Email != user.Email {
		t.Fatalf("email not valid, expect: %v, receive: %v", signUpReq.Email, user.Email)
	}

	// pass should be hashed
	if err := bcrypt.CompareHashAndPassword([]byte(user.Pass), []byte(signUpReq.Pass)); err != nil {
		t.Fatal("password was not correctly hashed")
	}

	// exp should be 0
	if user.Exp != 0 {
		t.Fatalf("exp in database was not 0, actual: %v", user.Exp)
	}

	// created at should be accurate
	if time.Since(user.CreatedAt) > time.Second {
		t.Fatalf("time since user created_at is greater than a second: %v", time.Since(user.CreatedAt))
	}

	// should return 400 with invalid json
	response, err = http.Post(HOST+url, "application/json", bytes.NewReader([]byte("invalid {json {{ 'f}")))
	if err != nil {
		t.Fatalf("fail to send post request: %v", err)
	}

	body = ReadResponseBody(t, response)

	if response.StatusCode != http.StatusBadRequest {
		t.Fatalf("expect status 400, received: %v, body: %v", response.StatusCode, body)
	}
}

func TestUserSignInRoute(t *testing.T) {
	url := "/user/sign-in"

	user := InsertRandUser(t, idlemonServer.Db, idlemonServer.DataCache)

	signInReq := &SignInReq{Email: user.Email, Pass: user.Pass}
	response := PostRequest(t, url, signInReq)
	body := ReadResponseBody(t, response)

	// expect status OK
	if response.StatusCode != http.StatusOK {
		t.Fatalf("expect status 200, received: %v, body: %v", response.StatusCode, body)
	}

	// unmarshal sign in response
	var signInRes SignInRes
	if err := json.Unmarshal([]byte(body), &signInRes); err != nil {
		t.Fatalf("fail to unmarshal sign in response: %v", err)
	}

	// token should not be empty
	if signInRes.Token == "" {
		t.Fatal("token was empty")
	}

	// id should be valid
	if signInRes.User.Id != user.Id {
		t.Fatalf("invalid id in response, expected: %v, received: %v", user.Id, signInRes.User.Id)
	}

	// email should be valid
	if signInRes.User.Email != user.Email {
		t.Fatalf("invalid email in response, expected: %v, received: %v", user.Email, signInRes.User.Email)
	}

	// sign in response should have no password
	if signInRes.User.Pass != "" {
		t.Fatalf("response should have no password, received: %v", signInRes.User.Pass)
	}

	// api token should exist
	tokenResult, err := idlemonServer.Rdb.Get(context.Background(), user.Id.String()).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			t.Fatalf("api token was not present in redis")
		} else {
			t.Fatalf("redis error: %v", err)
		}
	}

	if tokenResult != signInRes.Token {
		t.Fatalf("api token does not match, expected: %v, receive: %v", signInRes.Token, tokenResult)
	}
}

func TestUserRenameRoute(t *testing.T) {
	method := "PUT"
	url := "/user/rename"

	token, user := AuthenticatedUser(t, idlemonServer.Db, idlemonServer.Rdb, idlemonServer.DataCache)

	// test request with invalid name
	renameReq := &UserRenameReq{Name: ""}
	response := SendRequest(t, method, url, user.Id, token, renameReq)
	body := ReadResponseBody(t, response)

	// expect bad request response
	if response.StatusCode != http.StatusBadRequest {
		t.Errorf("expect status 400, received: %v, body: %v", response.StatusCode, body)
	}

	// test request with valid name
	renameReq.Name = "ValidName"
	response = SendRequest(t, method, url, user.Id, token, renameReq)
	body = ReadResponseBody(t, response)

	// expect status OK
	if response.StatusCode != http.StatusOK {
		t.Errorf("expect status 200, received: %v, body: %v", response.StatusCode, body)
	}

	var name string

	// fetch user name from database
	err := idlemonServer.Db.QueryRow(context.Background(), "SELECT name FROM users WHERE id = $1", user.Id).Scan(&name)
	if err != nil {
		t.Fatalf("fail to query users table: %v", err)
	}

	// fetched name must be the same as the request name
	if name != renameReq.Name {
		t.Fatalf("invalid name in database, expect: %v, receive: %v", renameReq.Name, name)
	}
}
