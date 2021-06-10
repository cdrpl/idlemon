package main_test

import (
	"io/ioutil"
	"log"
	"os"
	"testing"

	. "github.com/cdrpl/idlemon"
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
