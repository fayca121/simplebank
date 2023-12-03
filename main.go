package main

import (
	"database/sql"
	"github.com/fayca121/simplebank/api"
	db "github.com/fayca121/simplebank/db/sqlc"
	"github.com/fayca121/simplebank/util"
	_ "github.com/lib/pq"
	"log"
)

func main() {

	config, err := util.LoadConfig(".")

	if err != nil {
		log.Fatal("Cannot load config:", err)
	}

	conn, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}
	store := db.NewStore(conn)
	server, err := api.NewServer(config, store)
	if err != nil {
		log.Fatal("cannot create server:", err)
	}
	err = server.Start(config.ServerAddress)
	if err != nil {
		log.Fatal("cannot start server:", err)
	}
}
