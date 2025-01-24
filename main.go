package main

import (
	"context"
	"database/sql"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"net"
	"net/http"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	_ "github.com/lib/pq"
	"github.com/rakyll/statik/fs"
	"github.com/valkyraycho/bank/api"
	db "github.com/valkyraycho/bank/db/sqlc"
	_ "github.com/valkyraycho/bank/docs/statik"
	"github.com/valkyraycho/bank/gapi"
	"github.com/valkyraycho/bank/pb"
	"github.com/valkyraycho/bank/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/encoding/protojson"
)

func main() {

	cfg, err := utils.LoadConfig(".")
	if err != nil {
		log.Fatal().Err(err).Msg("cannot load config: ")
	}

	if cfg.Environment == "dev" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	conn, err := sql.Open(cfg.DBDriver, cfg.DBSource)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot connect to db: ")
	}

	runDBMigration(cfg.MigrationURL, cfg.DBSource)

	store := db.NewStore(conn)
	go runGatewayServer(cfg, store)
	runGrpcServer(cfg, store)
}

func runDBMigration(url, dbSource string) {
	migration, err := migrate.New(url, dbSource)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create new migrate instance: ")
	}

	if err := migration.Up(); err != nil {
		if err == migrate.ErrNoChange {
			log.Info().Msg("database already migrated to latest version")
			return
		}
		log.Fatal().Err(err).Msg("failed to migrate: ")
	}
	log.Info().Msg("database migrated successfully")

}

func runGrpcServer(cfg utils.Config, store db.Store) {
	server, err := gapi.NewServer(cfg, store)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create server: ")
	}

	grpcLogger := grpc.UnaryInterceptor(gapi.GRPCLogger)

	grpcServer := grpc.NewServer(grpcLogger)
	pb.RegisterBankServiceServer(grpcServer, server)
	reflection.Register(grpcServer)

	lis, err := net.Listen("tcp", cfg.GRPCServerAddress)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create listener: ")
	}

	log.Info().Msgf("start gRPC server at %s", lis.Addr().String())
	err = grpcServer.Serve(lis)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot start gRPC server: ")
	}
}

func runGatewayServer(cfg utils.Config, store db.Store) {
	server, err := gapi.NewServer(cfg, store)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create server: ")
	}

	grpcMux := runtime.NewServeMux(
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
			MarshalOptions: protojson.MarshalOptions{
				UseProtoNames: true,
			},
			UnmarshalOptions: protojson.UnmarshalOptions{
				DiscardUnknown: true,
			},
		}),
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = pb.RegisterBankServiceHandlerServer(ctx, grpcMux, server)

	if err != nil {
		log.Fatal().Err(err).Msg("cannot register handiler server: ")
	}

	mux := http.NewServeMux()
	mux.Handle("/", grpcMux)

	statikFS, err := fs.New()
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create statik fs: ")
	}
	mux.Handle("/swagger/", http.StripPrefix("/swagger/", http.FileServer(statikFS)))

	lis, err := net.Listen("tcp", cfg.HTTPServerAddress)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create listener")
	}

	log.Info().Msgf("start HTTP server at %s", lis.Addr().String())

	err = http.Serve(lis, gapi.HTTPLogger(mux))
	if err != nil {
		log.Fatal().Err(err).Msg("cannot start HTTP gateway server: ")
	}
}

func runGinServer(cfg utils.Config, store db.Store) {
	server, err := api.NewServer(cfg, store)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create server: ")
	}

	err = server.Start(cfg.HTTPServerAddress)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot start the server: ")
	}
}
