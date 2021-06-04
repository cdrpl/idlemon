package main

import (
	"database/sql"
	"errors"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/go-redis/redis/v8"
	"github.com/go-sql-driver/mysql"
	"github.com/julienschmidt/httprouter"
	"golang.org/x/crypto/bcrypt"
)

func CreateController(db *sql.DB, rdb *redis.Client) Controller {
	return Controller{db: db, rdb: rdb}
}

type Controller struct {
	db  *sql.DB
	rdb *redis.Client
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

/* User Routes */

func (c Controller) SignUp(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	req := GetReqDTO(r).(*SignUpReq)

	exists, err := NameExists(c.db, req.Name)
	if err != nil {
		log.Printf("sign up error: %v\n", err)
		ErrResSanitize(w, http.StatusInternalServerError, err.Error())
		return
	} else if exists {
		ErrResCustom(w, http.StatusBadRequest, "name is already taken")
		return
	}

	exists, err = EmailExists(c.db, req.Email)
	if err != nil {
		log.Printf("sign up error: %v\n", err)
		ErrResSanitize(w, http.StatusInternalServerError, err.Error())
		return
	} else if exists {
		ErrResCustom(w, http.StatusBadRequest, "an account with this email already exists")
		return
	}

	_, err = Insert(c.db, req.Name, req.Email, req.Pass)
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

	var user User
	query := "SELECT id, name, pass, exp, campaign, created_at FROM user WHERE email = ?"
	err := c.db.QueryRow(query, signInReq.Email).Scan(&user.ID, &user.Name, &user.Pass, &user.Exp, &user.Campaign, &user.CreatedAt)
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

	token, err := CreateApiToken(c.rdb, user.ID)
	if err != nil {
		log.Printf("sign in error: %v\n", err)
		ErrResSanitize(w, http.StatusInternalServerError, err.Error())
		return
	}

	signInRes := SignInRes{Token: token, User: user}
	signInRes.User.Email = signInReq.Email
	signInRes.User.Pass = ""

	log.Printf("user sign in: %v\n", signInReq.Email)
	JsonRes(w, signInRes)
}

func (c Controller) UserRename(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id := GetUserID(r)
	req := GetReqDTO(r).(*UserRenameReq)

	_, err := c.db.Exec("UPDATE user SET name = ? WHERE id = ?", req.Name, id)
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
