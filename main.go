package main

import (
	"database/sql"
	"github.com/fayca121/simplebank/api"
	db "github.com/fayca121/simplebank/db/sqlc"
	"github.com/fayca121/simplebank/gapi"
	"github.com/fayca121/simplebank/pb"
	"github.com/fayca121/simplebank/util"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
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
	//runGinServer(config, store)
	runGrpcServer(config, store)
}

func runGinServer(config util.Config, store db.Store) {
	server, err := api.NewServer(config, store)
	if err != nil {
		log.Fatal("cannot create server:", err)
	}
	err = server.Start(config.HTTPServerAddress)
	if err != nil {
		log.Fatal("cannot start server:", err)
	}
}

func runGrpcServer(config util.Config, store db.Store) {

	server, err := gapi.NewServer(config, store)
	if err != nil {
		log.Fatal("cannot create server:", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterSimpleBankServer(grpcServer, server)
	reflection.Register(grpcServer) //optional

	listener, err := net.Listen("tcp", config.GRPCServerAddress)

	if err != nil {
		log.Fatal("cannot create listener:", err)
	}

	log.Printf("start grpc server at %s", listener.Addr().String())

	err = grpcServer.Serve(listener)
	if err != nil {
		log.Fatal("cannot start grpc server:")
	}
}
