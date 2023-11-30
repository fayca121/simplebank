package db

import (
	"database/sql"
	_ "github.com/lib/pq"
	"log"
	"testing"
)

const (
	dbDriver = "postgres"
	dbSource = "postgresql://postgres:rocks@localhost:5432/simple_bank?sslmode=disable"
)

var testQueries *Queries
var testDb *sql.DB
var testDB *sql.DB

func TestMain(m *testing.M) {
	testDb, err := sql.Open(dbDriver, dbSource)
	testDB = testDb
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}
	testQueries = New(testDb)
	m.Run()
}
