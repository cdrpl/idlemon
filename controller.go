package main

import (
	"database/sql"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/go-redis/redis/v8"
	"github.com/go-sql-driver/mysql"
	"github.com/julienschmidt/httprouter"
	"golang.org/x/crypto/bcrypt"
)

func CreateController(db *sql.DB, rdb *redis.Client, dc *DataCache) Controller {
	return Controller{db: db, rdb: rdb, dc: dc}
}

type Controller struct {
	db  *sql.DB
	rdb *redis.Client
	dc  *DataCache
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
	userID := GetUserID(r)

	tx, err := c.db.BeginTx(r.Context(), nil)
	if err != nil {
		log.Printf("campaign collect error: %v\n", err)
		ErrResSanitize(w, http.StatusInternalServerError, err.Error())
	}
	defer tx.Rollback()

	campaign, err := GetCampaignLock(r.Context(), tx, userID)
	if err != nil {
		log.Printf("campaign collect error: %v\n", err)
		ErrResSanitize(w, http.StatusInternalServerError, err.Error())
		return
	}

	exp, gold, expStones := campaign.Collect()

	if exp > 0 || gold > 0 || expStones > 0 {
		_, err = tx.ExecContext(r.Context(), "UPDATE campaign SET last_collected_at = ? WHERE id = ?", campaign.LastCollectedAt, campaign.ID)
		if err != nil {
			log.Printf("campaign collect error: %v\n", err)
			ErrResSanitize(w, http.StatusInternalServerError, err.Error())
			return
		}

		// update exp
		query := "UPDATE user SET exp = exp + ? WHERE id = ?"
		_, err = tx.ExecContext(r.Context(), query, exp, userID)
		if err != nil {
			log.Printf("campaign collect error: %v\n", err)
			ErrResSanitize(w, http.StatusInternalServerError, err.Error())
			return
		}

		// gold
		query = "UPDATE user_resource SET amount = amount + ? WHERE (resource_id = ? AND user_id = ?)"
		_, err = tx.ExecContext(r.Context(), query, gold, RESOURCE_GOLD, userID)
		if err != nil {
			log.Printf("campaign collect error: %v\n", err)
			ErrResSanitize(w, http.StatusInternalServerError, err.Error())
			return
		}

		// exp stone
		query = "UPDATE user_resource SET amount = amount + ? WHERE (resource_id = ? AND user_id = ?)"
		_, err = tx.ExecContext(r.Context(), query, expStones, RESOURCE_EXP_STONE, userID)
		if err != nil {
			log.Printf("campaign collect error: %v\n", err)
			ErrResSanitize(w, http.StatusInternalServerError, err.Error())
			return
		}

		// commit
		if err := tx.Commit(); err != nil {
			log.Printf("campaign collect error: %v\n", err)
			ErrResSanitize(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	res := CampaignCollectRes{
		Exp:             exp,
		Gold:            gold,
		ExpStones:       expStones,
		LastCollectedAt: campaign.LastCollectedAt,
	}

	log.Printf("user %v has collected resources: %v\n", userID, res)
	JsonRes(w, res)
}

/* Unit Routes */

// Toggle a unit's lock. Only works on units owned by the user.
func (c Controller) UnitToggleLock(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	userID := GetUserID(r)

	unitID, err := strconv.Atoi(p.ByName("id"))
	if err != nil {
		ErrResCustom(w, http.StatusBadRequest, "invalid unit ID")
		return
	}

	// query will only run if unit is owned by the user
	result, err := c.db.ExecContext(r.Context(), "UPDATE unit SET is_locked = !is_locked WHERE (id = ? AND user_id = ?)", unitID, userID)
	if err != nil {
		ErrResSanitize(w, http.StatusInternalServerError, err.Error())
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		ErrResSanitize(w, http.StatusInternalServerError, err.Error())
	}

	if rowsAffected > 0 {
		log.Printf("user %v toggled lock on unit %v", userID, unitID)
		JsonSuccess(w)
	} else {
		log.Printf("user %v attempt toggle lock on invalid unit ID %v", userID, unitID)
		ErrResCustom(w, http.StatusBadRequest, "invalid unit ID")
	}
}

/* User Routes */

func (c Controller) SignUp(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	req := GetReqDTO(r).(*SignUpReq)

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

	_, err = InsertUser(r.Context(), c.db, c.dc, req.Name, req.Email, req.Pass)
	if err != nil {
		log.Printf("sign up error: %v\n", err)
		ErrResSanitize(w, http.StatusInternalServerError, err.Error())
		return
	}

	log.Printf("new user registration: %v\n", req.Email)
	JsonSuccess(w)
}

func (c Controller) SignIn(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	signInReq := GetReqDTO(r).(*SignInReq)

	user := User{}
	query := "SELECT id, name, pass, exp, created_at FROM user WHERE email = ?"
	err := c.db.QueryRowContext(r.Context(), query, signInReq.Email).Scan(&user.ID, &user.Name, &user.Pass, &user.Exp, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			ErrRes(w, http.StatusUnauthorized)
		} else {
			log.Printf("sign in error: %v\n", err)
			ErrResSanitize(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Pass), []byte(signInReq.Pass))
	if err != nil {
		ErrRes(w, http.StatusUnauthorized)
		return
	}

	token, err := CreateApiToken(r.Context(), c.rdb, user.ID)
	if err != nil {
		log.Printf("sign in error: %v\n", err)
		ErrResSanitize(w, http.StatusInternalServerError, err.Error())
		return
	}

	// get user units
	units, err := Units(r.Context(), c.db, user.ID)
	if err != nil {
		log.Printf("fetch user units error: %v\n", err)
		ErrResSanitize(w, http.StatusInternalServerError, err.Error())
		return
	}

	// get user resources
	userResources, err := GetUserResources(r.Context(), c.db, user.ID)
	if err != nil {
		log.Printf("fetch user resources error: %v\n", err)
		ErrResSanitize(w, http.StatusInternalServerError, err.Error())
		return
	}

	// get campaign
	campaign, err := GetCampaign(r.Context(), c.db, user.ID)
	if err != nil {
		log.Printf("fetch campaign error: %v\n", err)
		ErrResSanitize(w, http.StatusInternalServerError, err.Error())
		return
	}

	// update daily sign in quest
	if err := SignInDailyQuest(r.Context(), c.db, user.ID); err != nil {
		log.Printf("user sign in daily quest error: %v\n", err)
		ErrResSanitize(w, http.StatusInternalServerError, err.Error())
		return
	}

	user.Email = signInReq.Email
	user.Pass = ""

	signInRes := SignInRes{
		Token:         token,
		User:          user,
		Units:         units,
		UserResources: userResources,
		Campaign:      campaign,
		Resources:     c.dc.Resources,
	}

	log.Printf("user sign in: %v\n", signInReq.Email)
	JsonRes(w, signInRes)
}

func (c Controller) UserRename(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id := GetUserID(r)
	req := GetReqDTO(r).(*UserRenameReq)

	_, err := c.db.ExecContext(r.Context(), "UPDATE user SET name = ? WHERE id = ?", req.Name, id)
	if err != nil {
		if sqlErr, ok := err.(*mysql.MySQLError); ok {
			if sqlErr.Number == ER_DUP_ENTRY {
				ErrResCustom(w, http.StatusBadRequest, "the name is already taken")
				return
			}
		}

		log.Printf("user rename error: %v\n", err)
		ErrResSanitize(w, http.StatusInternalServerError, err.Error())
		return
	}

	log.Printf("user ID %v change name to %v\n", id, req.Name)

	res := map[string]string{"name": req.Name}
	JsonRes(w, res)
}
