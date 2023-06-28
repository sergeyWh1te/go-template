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
		fmt.Println("Read env error:", envErr.Error())
		return
	}

	log, logErr := logger.New(&cfg.AppConfig)
	if logErr != nil {
		fmt.Println("Logger error:", logErr.Error())
		return
	}

	db, errDB := postgres.Connect(cfg.PgConfig)
	if errDB != nil {
		log.Fatalf("Connect db error: %s", errDB.Error())
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
		log.Errorf("someDaemon error: %s", err.Error())
	}

	if err := server.RunHTTPServer(ctx, cfg.AppConfig.Port, r); err != nil {
		log.Errorf("RunHTTPServer error: %s", err.Error())
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
