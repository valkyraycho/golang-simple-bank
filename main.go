package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"

	"github.com/hibiken/asynq"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"

	"net"
	"net/http"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	_ "github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rakyll/statik/fs"
	"github.com/valkyraycho/bank/api"
	db "github.com/valkyraycho/bank/db/sqlc"
	_ "github.com/valkyraycho/bank/docs/statik"
	"github.com/valkyraycho/bank/gapi"
	"github.com/valkyraycho/bank/pb"
	"github.com/valkyraycho/bank/utils"
	"github.com/valkyraycho/bank/worker"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/encoding/protojson"
)

var interruptSignals = []os.Signal{
	os.Interrupt,
	syscall.SIGTERM,
	syscall.SIGINT,
}

func main() {

	cfg, err := utils.LoadConfig(".")
	if err != nil {
		log.Fatal().Err(err).Msg("cannot load config: ")
	}

	if cfg.Environment == "dev" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	ctx, stop := signal.NotifyContext(context.Background(), interruptSignals...)
	defer stop()

	connPool, err := pgxpool.New(ctx, cfg.DBSource)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot connect to db: ")
	}

	runDBMigration(cfg.MigrationURL, cfg.DBSource)

	redisOpt := asynq.RedisClientOpt{Addr: cfg.RedisAddress}
	taskDistributor := worker.NewRedisTaskDistributor(redisOpt)

	store := db.NewStore(connPool)

	waitGroup, ctx := errgroup.WithContext(ctx)

	runTaskHandler(ctx, waitGroup, redisOpt, store)
	runGatewayServer(ctx, waitGroup, cfg, store, taskDistributor)
	runGrpcServer(ctx, waitGroup, cfg, store, taskDistributor)

	err = waitGroup.Wait()
	if err != nil {
		log.Fatal().Err(err).Msg("error from wait group")
	}
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

func runTaskHandler(
	ctx context.Context,
	waitGroup *errgroup.Group,
	redisOpt asynq.RedisClientOpt,
	store db.Store,
) {
	taskHandler := worker.NewRedisTaskHandler(&redisOpt, store)
	log.Info().Msg("start task processor")
	err := taskHandler.Start()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to start task handler")
	}

	waitGroup.Go(func() error {
		<-ctx.Done()
		log.Info().Msg("gracefully shutdowning task handler")

		taskHandler.Shutdown()
		log.Info().Msg("task handler is gracefully shutdown")
		return nil
	})
}

func runGrpcServer(
	ctx context.Context,
	waitGroup *errgroup.Group,
	cfg utils.Config,
	store db.Store,
	taskDistributor worker.TaskDistributor,
) {
	server, err := gapi.NewServer(cfg, store, taskDistributor)
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

	waitGroup.Go(func() error {
		log.Info().Msgf("start gRPC server at %s", lis.Addr().String())
		err = grpcServer.Serve(lis)
		if err != nil {
			if errors.Is(err, grpc.ErrServerStopped) {
				return nil
			}
			log.Fatal().Err(err).Msg("gRPC server failed to serve")
			return err
		}
		return nil
	})

	waitGroup.Go(func() error {
		<-ctx.Done()
		log.Info().Msg("gracefully shutdowning gRPC server")

		grpcServer.GracefulStop()
		log.Info().Msg("gRPC server is gracefully shutdown")
		return nil
	})

}

func runGatewayServer(
	ctx context.Context,
	waitGroup *errgroup.Group,
	cfg utils.Config,
	store db.Store,
	taskDistributor worker.TaskDistributor,
) {
	server, err := gapi.NewServer(cfg, store, taskDistributor)
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

	err = pb.RegisterBankServiceHandlerServer(ctx, grpcMux, server)

	if err != nil {
		log.Fatal().Err(err).Msg("cannot register handler server: ")
	}

	mux := http.NewServeMux()
	mux.Handle("/", grpcMux)

	statikFS, err := fs.New()
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create statik fs: ")
	}
	mux.Handle("/swagger/", http.StripPrefix("/swagger/", http.FileServer(statikFS)))

	httpServer := &http.Server{
		Handler: gapi.HTTPLogger(mux),
		Addr:    cfg.HTTPServerAddress,
	}

	waitGroup.Go(func() error {
		log.Info().Msgf("start HTTP server at %s", httpServer.Addr)
		err := httpServer.ListenAndServe()
		if err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				return nil
			}
			log.Fatal().Err(err).Msg("cannot start HTTP gateway server")
			return err
		}
		return nil
	})

	waitGroup.Go(func() error {
		<-ctx.Done()
		log.Info().Msg("gracefully shutdowning HTTP gateway server")

		err := httpServer.Shutdown(context.Background())
		if err != nil {
			log.Error().Err(err).Msg("failed to shutdown HTTP gateway server")
			return err
		}
		log.Info().Msg("HTTP server is gracefully shutdown")
		return nil
	})

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
