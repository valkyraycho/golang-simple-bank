package main

import (
	"context"
	"database/sql"
	"log"
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
		log.Fatal("cannot load config: ", err)
	}

	conn, err := sql.Open(cfg.DBDriver, cfg.DBSource)
	if err != nil {
		log.Fatal("cannot connect to db: ", err)
	}

	runDBMigration(cfg.MigrationURL, cfg.DBSource)

	store := db.NewStore(conn)
	go runGatewayServer(cfg, store)
	runGrpcServer(cfg, store)
}

func runDBMigration(url, dbSource string) {
	migration, err := migrate.New(url, dbSource)
	if err != nil {
		log.Fatal("cannot create new migrate instance: ", err)
	}

	if err := migration.Up(); err != nil {
		if err == migrate.ErrNoChange {
			log.Println("database already migrated to latest version")
			return
		}
		log.Fatal("failed to migrate: ", err)
	}
	log.Println("database migrated successfully")

}

func runGrpcServer(cfg utils.Config, store db.Store) {
	server, err := gapi.NewServer(cfg, store)
	if err != nil {
		log.Fatal("cannot create server: ", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterBankServiceServer(grpcServer, server)
	reflection.Register(grpcServer)

	lis, err := net.Listen("tcp", cfg.GRPCServerAddress)
	if err != nil {
		log.Fatal("cannot create listener: ", err)
	}

	log.Printf("start gRPC server at %s", lis.Addr().String())
	err = grpcServer.Serve(lis)
	if err != nil {
		log.Fatal("cannot start gRPC server: ", err)
	}
}

func runGatewayServer(cfg utils.Config, store db.Store) {
	server, err := gapi.NewServer(cfg, store)
	if err != nil {
		log.Fatal("cannot create server: ", err)
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
		log.Fatal("cannot register handiler server: ", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/", grpcMux)

	statikFS, err := fs.New()
	if err != nil {
		log.Fatal("cannot create statik fs: ", err)
	}
	mux.Handle("/swagger/", http.StripPrefix("/swagger/", http.FileServer(statikFS)))

	lis, err := net.Listen("tcp", cfg.HTTPServerAddress)
	if err != nil {
		log.Fatal("cannot create listener", err)
	}

	log.Printf("start HTTP server at %s", lis.Addr().String())
	err = http.Serve(lis, mux)
	if err != nil {
		log.Fatal("cannot start HTTP gateway server: ", err)
	}
}

func runGinServer(cfg utils.Config, store db.Store) {
	server, err := api.NewServer(cfg, store)
	if err != nil {
		log.Fatal("cannot create server: ", err)
	}

	err = server.Start(cfg.HTTPServerAddress)
	if err != nil {
		log.Fatal("cannot start the server: ", err)
	}
}
