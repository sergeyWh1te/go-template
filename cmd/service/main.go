package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sync/errgroup"

	"github.com/sergeyWh1te/go-template/internal/app/http_server"
	"github.com/sergeyWh1te/go-template/internal/connectors/logger"
	"github.com/sergeyWh1te/go-template/internal/connectors/metrics"
	"github.com/sergeyWh1te/go-template/internal/connectors/postgres"
	"github.com/sergeyWh1te/go-template/internal/env"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	cfg, envErr := env.Read(ctx)
	if envErr != nil {
		println("Read env error:", envErr.Error())

		os.Exit(1)
	}

	db, errDB := postgres.Connect(cfg.PgConfig)
	if errDB != nil {
		println("Connect db error:", errDB.Error())
		os.Exit(1)
	}

	log, logErr := logger.New(&cfg.AppConfig)
	if logErr != nil {
		println("Logger error:", logErr.Error())

		os.Exit(1)
	}

	log.Info(fmt.Sprintf(`started %s application`, cfg.AppConfig.Name))

	r := chi.NewRouter()
	metrics := metrics.New(prometheus.NewRegistry(), cfg.AppConfig.Name, cfg.AppConfig.Env)

	repo := server.Repository(db)
	usecase := server.Usecase(repo)

	app := server.New(log, metrics, usecase, repo)

	app.Metrics.BuildInfo.Inc()
	app.RegisterRoutes(r)

	if err := someDaemon(ctx); err != nil {
		fmt.Println("someDaemon error", err)
	}

	if err := server.RunHTTPServer(ctx, cfg.AppConfig.Port, r); err != nil {
		println(err.Error())
	}

	fmt.Println(`Main done`)
}

func someDaemon(ctx context.Context) error {
	g, gCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		for {
			select {
			case <-time.After(1 * time.Second):
				fmt.Println(2)
			case <-gCtx.Done():
				return nil
			}
		}
	})

	return g.Wait()
}
