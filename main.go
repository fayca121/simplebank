package main

import (
	"context"
	"database/sql"
	"errors"
	"github.com/fayca121/simplebank/api"
	db "github.com/fayca121/simplebank/db/sqlc"
	_ "github.com/fayca121/simplebank/doc/statik"
	"github.com/fayca121/simplebank/gapi"
	"github.com/fayca121/simplebank/pb"
	"github.com/fayca121/simplebank/util"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	_ "github.com/lib/pq"
	"github.com/rakyll/statik/fs"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/encoding/protojson"
	"net"
	"net/http"
	"os"
)

func main() {

	config, err := util.LoadConfig(".")

	if err != nil {
		log.Fatal().Msgf("Cannot load config: %s", err)
	}

	if config.Environment == "Dev" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	conn, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal().Msgf("cannot connect to db: %s", err)
	}
	// run db migration
	runDBMigration(config.MigrationUrl, config.DBSource)

	store := db.NewStore(conn)
	//runGinServer(config, store)
	go runGatewayServer(config, store)
	runGrpcServer(config, store)
}

func runGinServer(config util.Config, store db.Store) {
	server, err := api.NewServer(config, store)
	if err != nil {
		log.Fatal().Msgf("cannot create server: %s", err)
	}
	err = server.Start(config.HTTPServerAddress)
	if err != nil {
		log.Fatal().Msgf("cannot start server: %s", err)
	}
}

func runGrpcServer(config util.Config, store db.Store) {

	server, err := gapi.NewServer(config, store)
	if err != nil {
		log.Fatal().Msgf("cannot create server: %s", err)
	}

	grpcLogger := grpc.UnaryInterceptor(gapi.GrpcLogger)
	grpcServer := grpc.NewServer(grpcLogger)
	pb.RegisterSimpleBankServer(grpcServer, server)
	reflection.Register(grpcServer) //optional

	listener, err := net.Listen("tcp", config.GRPCServerAddress)

	if err != nil {
		log.Fatal().Msgf("cannot create listener: %s", err)
	}

	log.Info().Msgf("start grpc server at %s", listener.Addr().String())

	err = grpcServer.Serve(listener)
	if err != nil {
		log.Fatal().Msgf("cannot start grpc server: %s", err)
	}
}

func runGatewayServer(config util.Config, store db.Store) {

	server, err := gapi.NewServer(config, store)
	if err != nil {
		log.Fatal().Msgf("cannot create server: %s", err)
	}

	jsonOption := runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
		MarshalOptions: protojson.MarshalOptions{
			UseProtoNames: true,
		},
		UnmarshalOptions: protojson.UnmarshalOptions{
			DiscardUnknown: true,
		},
	})

	grpcMux := runtime.NewServeMux(jsonOption)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err = pb.RegisterSimpleBankHandlerServer(ctx, grpcMux, server)
	if err != nil {
		log.Fatal().Msgf("cannot register handler server %s", err)
	}
	mux := http.NewServeMux()
	mux.Handle("/", grpcMux)

	//fs := http.FileServer(http.Dir("./doc/swagger"))
	statikFs, err := fs.New()

	if err != nil {
		log.Fatal().Msgf("cannot create statik fs: %s", err)
	}
	swaggerHandler := http.StripPrefix("/swagger/", http.FileServer(statikFs))

	mux.Handle("/swagger/", swaggerHandler)

	listener, err := net.Listen("tcp", config.HTTPServerAddress)

	if err != nil {
		log.Fatal().Msgf("cannot create listener: %s", err)
	}

	log.Info().Msgf("start http gateway server at %s", listener.Addr().String())

	handler := gapi.HttpLogger(mux)

	err = http.Serve(listener, handler)
	if err != nil {
		log.Fatal().Msgf("cannot start http gateway server: %s", err)
	}
}

func runDBMigration(migrationURL string, dbSource string) {
	migration, err := migrate.New(migrationURL, dbSource)

	if err != nil {
		log.Fatal().Msg("cannot create new migrate instance")
	}
	if err = migration.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Fatal().Msgf("failed to run migrate up: %s", err)
	}
	log.Info().Msg("DB migrated successfully")
}
