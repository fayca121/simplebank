package db

import (
	"context"
	"github.com/fayca121/simplebank/util"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
	"log"
	"os"
	"testing"
)

var testStore Store

func TestMain(m *testing.M) {

	config, err := util.LoadConfig("../..")

	if err != nil {
		log.Fatal("cannot connect to connPool:", err)
	}

	testPool, err := pgxpool.New(context.Background(), config.DBSource)
	if err != nil {
		log.Fatal("cannot connect to connPool:", err)
	}
	testStore = NewStore(testPool)
	os.Exit(m.Run())
}
