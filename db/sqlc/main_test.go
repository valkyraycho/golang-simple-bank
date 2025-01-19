package db

import (
	"database/sql"
	"log"
	"os"
	"testing"

	_ "github.com/lib/pq"
	"github.com/valkyraycho/bank/utils"
)

var testDB *sql.DB
var testQueries *Queries

func TestMain(m *testing.M) {
	cfg, err := utils.LoadConfig("../..")
	if err != nil {
		log.Fatal("cannot load config: ", err)
	}

	testDB, err = sql.Open(cfg.DBDriver, cfg.DBSource)
	if err != nil {
		log.Fatal("unable to connect to db: ", err)
	}
	testQueries = New(testDB)
	os.Exit(m.Run())
}
