package main

import (
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/julienschmidt/httprouter"
	"golang.org/x/crypto/bcrypt"
)

func CreateController(db *pgxpool.Pool, rdb *redis.Client, wsHub *WsHub, dataCache DataCache) Controller {
	return Controller{
		db:        db,
		rdb:       rdb,
		wsHub:     wsHub,
		dataCache: dataCache,
	}
}

type Controller struct {
	db        *pgxpool.Pool
	rdb       *redis.Client
	wsHub     *WsHub
	dataCache DataCache
}

/* App Routes */

func (c Controller) HealthCheck(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	JsonSuccess(w)
}

// Route will return the current server version and expected client version.
func (c Controller) Version(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	m := map[string]string{
		"server": os.Getenv("SERVER_VERSION"),
		"client": os.Getenv("CLIENT_VERSION"),
	}
	JsonRes(w, m)
}

func (c Controller) Robots(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	_, err := io.WriteString(w, robots)
	if err != nil {
		log.Fatalf("fail to write robots.txt to http response writer: %v\n", err)
	}
}

func (c Controller) MethodNotAllowed(w http.ResponseWriter, r *http.Request) {
	ErrRes(w, http.StatusMethodNotAllowed)
}

func (c Controller) NotFound(w http.ResponseWriter, r *http.Request) {
	ErrRes(w, http.StatusNotFound)
}

/* Campaign Routes */

func (c Controller) CampaignCollect(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	userId := GetUserId(r)

	tx, err := c.db.Begin(r.Context())
	if err != nil {
		log.Printf("campaign collect error: %v\n", err)
		ErrResSanitize(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer tx.Rollback(r.Context())

	campaign := Campaign{UserId: userId}

	query := "SELECT id, level, last_collected_at FROM campaign WHERE user_id = $1"
	err = tx.QueryRow(r.Context(), query, userId).Scan(&campaign.Id, &campaign.Level, &campaign.LastCollectedAt)
	if err != nil {
		log.Printf("fail to find campaign row: %v\n", err)
		ErrResSanitize(w, http.StatusInternalServerError, err.Error())
		return
	}

	transactions, err := campaign.Collect(r.Context(), tx)
	if err != nil {
		log.Printf("fail to collect campaign resources: %v\n", err)
		ErrResSanitize(w, http.StatusInternalServerError, err.Error())
		return
	}

	if err := tx.Commit(r.Context()); err != nil {
		log.Printf("campaign collect error: %v\n", err)
		ErrResSanitize(w, http.StatusInternalServerError, err.Error())
		return
	}

	res := CampaignCollectRes{
		Transactions:    transactions,
		LastCollectedAt: campaign.LastCollectedAt,
	}

	log.Printf("user %v has collected resources: %+v\n", userId, transactions)
	JsonRes(w, res)
}

/* Daily Quest Routes */

func (c Controller) DailyQuestComplete(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	userId := GetUserId(r)

	questId, err := strconv.Atoi(p.ByName("id"))
	if err != nil {
		log.Printf("daily quest complete, quest ID could not be converted to int: %v\n", err)
		ErrResSanitize(w, http.StatusInternalServerError, err.Error())
		return
	}

	tx, err := c.db.Begin(r.Context())
	if err != nil {
		log.Printf("daily quest complete, fail to begin transaction: %v\n", err)
		ErrResSanitize(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer tx.Rollback(r.Context())

	questProgress, err := FindDailyQuestProgress(r.Context(), tx, userId, questId)
	if err != nil {
		log.Printf("fail to find daily quest progress: %v\n", err)
		ErrResSanitize(w, http.StatusInternalServerError, err.Error())
		return
	}

	dailyQuest := c.dataCache.DailyQuests[questProgress.DailyQuestId]

	if questProgress.IsCompleted() {
		JsonRes(w, DailyQuestCompleteRes{Status: 1, Message: "already completed"})
		return
	}

	if questProgress.Count < dailyQuest.Required {
		JsonRes(w, DailyQuestCompleteRes{Status: 2, Message: "requirements not met"})
		return
	}

	if err := questProgress.Complete(r.Context(), tx); err != nil {
		log.Printf("fail to complete daily quest: %v\n", err)
		ErrResSanitize(w, http.StatusInternalServerError, err.Error())
		return
	}

	if err := dailyQuest.Transaction.Apply(r.Context(), tx, userId); err != nil {
		log.Printf("fail to apply daily quest reward: %v\n", err)
		ErrResSanitize(w, http.StatusInternalServerError, err.Error())
		return
	}

	if err := tx.Commit(r.Context()); err != nil {
		log.Printf("daily quest complete, failed to commit transaction: %v\n", err)
		ErrResSanitize(w, http.StatusInternalServerError, err.Error())
		return
	}

	log.Printf("user %v completed daily quest %v", userId, questId)
	res := DailyQuestCompleteRes{Status: 0, Transaction: dailyQuest.Transaction}
	JsonRes(w, res)
}

/* Summon Routes */

func (c Controller) SummonUnit(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	userId := GetUserId(r)

	tx, err := c.db.Begin(r.Context())
	if err != nil {
		log.Printf("fail to begin transaction: %v\n", err)
		ErrResSanitize(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer tx.Rollback(r.Context())

	resource, err := FindResourceLock(r.Context(), tx, userId, RESOURCE_GEMS)
	if err != nil {
		log.Printf("fail to find gems resource: %v\n", err)
		ErrResSanitize(w, http.StatusInternalServerError, err.Error())
		return
	}

	// must have enough resources to summon
	if resource.Amount < UNIT_SUMMON_COST {
		ErrResCustom(w, http.StatusBadRequest, "not enough gems")
		return
	}

	resource.Amount -= UNIT_SUMMON_COST

	unit := RandUnit(c.dataCache)

	if err := InsertUnit(r.Context(), tx, userId, unit); err != nil {
		log.Printf("fail to insert unit: %v\n", err)
		ErrResSanitize(w, http.StatusInternalServerError, err.Error())
		return
	}

	if err := IncResource(r.Context(), tx, userId, RESOURCE_GEMS, -UNIT_SUMMON_COST); err != nil {
		log.Printf("fail to increase gems resource: %v\n", err)
		ErrResSanitize(w, http.StatusInternalServerError, err.Error())
		return
	}

	if err := tx.Commit(r.Context()); err != nil {
		log.Printf("fail to commit transaction: %v\n", err)
		ErrResSanitize(w, http.StatusInternalServerError, err.Error())
		return
	}

	log.Printf("user %v summoned a unit {templateId:%v}\n", userId, unit.Template)
	JsonRes(w, SummonUnitRes{
		Unit: unit,
		Transaction: Transaction{
			Type:   TRANSACTION_GEMS,
			Amount: -UNIT_SUMMON_COST,
		},
	})
}

/* Unit Routes */

// Toggle a unit's lock. Only works on units owned by the user.
func (c Controller) UnitToggleLock(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	userId := GetUserId(r)

	unitId, err := uuid.Parse(p.ByName("id"))
	if err != nil {
		log.Printf("invalid unit ID: %v\n", err)
		ErrResSanitize(w, http.StatusInternalServerError, err.Error())
		return
	}

	tx, err := c.db.Begin(r.Context())
	if err != nil {
		log.Printf("fail to begin transaction: %v\n", err)
		ErrResSanitize(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer tx.Rollback(r.Context())

	query := "UPDATE units SET is_locked = NOT is_locked WHERE id = $1"
	cmdTag, err := c.db.Exec(r.Context(), query, unitId)
	if err != nil {
		log.Printf("fail to update units table: %v\n", err)
		ErrResSanitize(w, http.StatusInternalServerError, err.Error())
		return
	}

	if cmdTag.RowsAffected() > 0 {
		if err := tx.Commit(r.Context()); err != nil {
			log.Printf("fail to commit transaction: %v\n", err)
			ErrResSanitize(w, http.StatusInternalServerError, err.Error())
			return
		}
	} else {
		ErrRes(w, http.StatusNotFound)
		return
	}

	log.Printf("user %v toggled lock for unit %v", userId, unitId)
	JsonSuccess(w)
}

/* User Routes */

func (c Controller) SignUp(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	req := GetReqDto(r).(*SignUpReq)

	exists, err := NameExists(r.Context(), c.db, req.Name)
	if err != nil {
		log.Printf("sign up error: %v\n", err)
		ErrResSanitize(w, http.StatusInternalServerError, err.Error())
		return
	} else if exists {
		ErrResCustom(w, http.StatusBadRequest, "name is already taken")
		return
	}

	exists, err = EmailExists(r.Context(), c.db, req.Email)
	if err != nil {
		log.Printf("sign up error: %v\n", err)
		ErrResSanitize(w, http.StatusInternalServerError, err.Error())
		return
	} else if exists {
		ErrResCustom(w, http.StatusBadRequest, "an account with this email already exists")
		return
	}

	tx, err := c.db.Begin(r.Context())
	if err != nil {
		log.Printf("fail to begin transaction: %v\n", err)
		ErrResSanitize(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer tx.Rollback(r.Context())

	err = InsertUser(r.Context(), tx, c.dataCache, CreateUser(c.dataCache, req.Name, req.Email, req.Pass))
	if err != nil {
		log.Printf("sign up error: %v\n", err)
		ErrResSanitize(w, http.StatusInternalServerError, err.Error())
		return
	}

	if err := tx.Commit(r.Context()); err != nil {
		log.Printf("fail to commit transaction: %v\n", err)
		ErrResSanitize(w, http.StatusInternalServerError, err.Error())
		return
	}

	log.Printf("new user registration: {name:%v email:%v}\n", req.Name, req.Email)
	JsonSuccess(w)
}

func (c Controller) SignIn(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	signInReq := GetReqDto(r).(*SignInReq)

	var user User

	query := "SELECT id, name, email, pass, exp, created_at FROM users WHERE email = $1"
	err := c.db.QueryRow(r.Context(), query, signInReq.Email).Scan(&user.Id, &user.Name, &user.Email, &user.Pass, &user.Exp, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			ErrRes(w, http.StatusUnauthorized)
		} else {
			log.Printf("fail to select user row: %v\n", err)
			ErrResSanitize(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	passBytes := []byte(signInReq.Pass)
	hashBytes := []byte(user.Pass)

	err = bcrypt.CompareHashAndPassword(hashBytes, passBytes)
	if err != nil {
		ErrRes(w, http.StatusUnauthorized)
		return
	}

	// check if hash cost must be updated
	cost, err := bcrypt.Cost(hashBytes)
	if err != nil {
		log.Printf("fail to check password's hash cost: %v\n", err)
		ErrResSanitize(w, http.StatusInternalServerError, err.Error())
		return
	}

	// if cost is invalid, generate new hash and update database
	if cost != BCRYPT_COST {
		bytes, err := bcrypt.GenerateFromPassword(passBytes, BCRYPT_COST)
		if err != nil {
			log.Printf("fail to create new password hash: %v\n", err)
			ErrResSanitize(w, http.StatusInternalServerError, err.Error())
			return
		}

		query := "UPDATE users SET pass = $1 WHERE id = $2"
		_, err = c.db.Exec(r.Context(), query, bytes, user.Id)
		if err != nil {
			log.Printf("fail to update user password: %v\n", err)
			ErrResSanitize(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	token, err := CreateApiToken(r.Context(), c.rdb, user.Id)
	if err != nil {
		log.Printf("fail to create API token: %v\n", err)
		ErrResSanitize(w, http.StatusInternalServerError, err.Error())
		return
	}

	tx, err := c.db.Begin(r.Context())
	if err != nil {
		log.Printf("fail to begin transaction: %v\n", err)
		ErrResSanitize(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer tx.Rollback(r.Context())

	// daily sign in quest progress
	questProgress := DailyQuestProgress{DailyQuestId: DAILY_QUEST_SIGN_IN, UserId: user.Id}

	query = "SELECT id, last_completed_at FROM daily_quest_progress WHERE (user_id = $1 AND daily_quest_id = $2) FOR UPDATE"
	err = tx.QueryRow(r.Context(), query, user.Id, questProgress.DailyQuestId).Scan(&questProgress.Id, &questProgress.LastCompletedAt)
	if err != nil {
		log.Printf("fail to query daily_quest_progress table: %v\n", err)
		ErrResSanitize(w, http.StatusInternalServerError, err.Error())
		return
	}

	// increase sign in quest count if not completed
	if !questProgress.IsCompleted() {
		if err := questProgress.IncreaseCount(r.Context(), tx); err != nil {
			log.Printf("fail to increase daily quest progress count: %v\n", err)
			ErrResSanitize(w, http.StatusInternalServerError, err.Error())
			return
		}

		if err := tx.Commit(r.Context()); err != nil {
			log.Printf("fail to commit transaction: %v\n", err)
			ErrResSanitize(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	campaign, err := FindCampaign(r.Context(), c.db, user.Id)
	if err != nil {
		log.Printf("faily to find campaign: %v\n", err)
		ErrResSanitize(w, http.StatusInternalServerError, err.Error())
		return
	}

	dailyQuestProgress, err := FindAllDailyQuestProgress(r.Context(), c.db, user.Id)
	if err != nil {
		log.Printf("fail to find daily quest progress: %v\n", err)
		ErrResSanitize(w, http.StatusInternalServerError, err.Error())
		return
	}

	resources, err := FindResources(r.Context(), c.db, user.Id)
	if err != nil {
		log.Printf("fail to find resources: %v\n", err)
		ErrResSanitize(w, http.StatusInternalServerError, err.Error())
		return
	}

	units, err := FindUnits(r.Context(), c.db, user.Id)
	if err != nil {
		log.Printf("fail to find units: %v\n", err)
		ErrResSanitize(w, http.StatusInternalServerError, err.Error())
		return
	}

	signInRes := SignInRes{
		Token:              token,
		User:               user,
		Campaign:           campaign,
		DailyQuestProgress: dailyQuestProgress,
		Resources:          resources,
		Units:              units,
		UnitTemplates:      c.dataCache.UnitTemplates,
	}

	log.Printf("user sign in: {id:%v name:%v email:%v}\n", user.Id, user.Name, user.Email)
	JsonRes(w, signInRes)
}

func (c Controller) UserRename(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id := GetUserId(r)
	req := GetReqDto(r).(*UserRenameReq)

	_, err := c.db.Exec(r.Context(), "UPDATE users SET name = $1 WHERE id = $2", req.Name, id)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			ErrResCustom(w, http.StatusBadRequest, "the name is already taken")
		} else {
			log.Printf("user rename error: %v\n", err)
			ErrResSanitize(w, http.StatusInternalServerError, err.Error())
		}

		return
	}

	log.Printf("user %v change name to %v\n", id, req.Name)
	JsonSuccess(w)
}

// WebSocket connection handler
func (c Controller) WebSocketConnectionHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	userId := GetUserId(r)

	conn, err := c.wsHub.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("fail to upgrade WebSocket connection: %v\n", err)
		ErrResSanitize(w, http.StatusInternalServerError, err.Error())
		return
	}

	client := &WsClient{
		wsHub:  c.wsHub,
		userId: userId,
		conn:   conn,
		send:   make(chan WebSocketMessage, 256),
	}

	c.wsHub.registerClient <- client

	go client.writePump()
	go client.readPump()
}
