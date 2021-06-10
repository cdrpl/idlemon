package main_test

import (
	"database/sql"
	"io/ioutil"
	"log"
	"os"
	"testing"

	. "github.com/cdrpl/idlemon"
	"github.com/go-redis/redis/v8"
)

func TestMain(m *testing.M) {
	log.SetOutput(ioutil.Discard)

	os.Setenv("ENV", "test")
	os.Setenv("DB_NAME", "test")
	LoadEnv(ENV_FILE, VERSION)

	db := CreateDBConn()
	DropTables(db)
	InitDatabase(db)

	SeedRand()

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

	result, err := Insert(db, user.Name, user.Email, user.Pass)
	if err != nil {
		return User{}, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return User{}, err
	}

	user.ID = int(id)
	return user, nil
}

func AuthenticatedUser(db *sql.DB, rdb *redis.Client) (string, User, error) {
	user, err := InsertRandUser(db)
	if err != nil {
		return "", User{}, err
	}

	token, err := CreateApiToken(rdb, user.ID)
	if err != nil {
		return "", User{}, err
	}

	return token, user, err
}
