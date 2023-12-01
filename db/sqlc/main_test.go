package db

import (
	"database/sql"
	"github.com/fayca121/simplebank/util"
	_ "github.com/lib/pq"
	"log"
	"testing"
)

var testQueries *Queries
var testDb *sql.DB
var testDB *sql.DB

func TestMain(m *testing.M) {

	config, err := util.LoadConfig("../..")

	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	testDb, err := sql.Open(config.DBDriver, config.DBSource)
	testDB = testDb
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}
	testQueries = New(testDb)
	m.Run()
}
