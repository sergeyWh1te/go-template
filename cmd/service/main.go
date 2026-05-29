package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sync/errgroup"

	server "github.com/sergeyWh1te/go-template/internal/app/http_server"
	"github.com/sergeyWh1te/go-template/internal/connectors/logger"
	"github.com/sergeyWh1te/go-template/internal/connectors/metrics"
	"github.com/sergeyWh1te/go-template/internal/connectors/postgres"
	"github.com/sergeyWh1te/go-template/internal/env"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	cfg, envErr := env.Read()
	if envErr != nil {
		fmt.Println("Read env error:", envErr.Error())
		return
	}

	log, logErr := logger.New(&cfg.AppConfig)
	if logErr != nil {
		fmt.Println("Logger error:", logErr.Error())
		return
	}

	db, errDB := postgres.Connect(&cfg.PgConfig)
	if errDB != nil {
		fmt.Println("Connect db error:", errDB.Error())
		return
	}
	defer func(db *sqlx.DB) {
		if err := db.Close(); err != nil {
			log.Error("could not close db connection", "err", err)
		}
	}(db)

	log.Info("started application", "name", cfg.AppConfig.Name)

	r := chi.NewRouter()
	metricsStore := metrics.New(prometheus.NewRegistry(), cfg.AppConfig.Name, cfg.AppConfig.Env)

	repo := server.Repository(db)
	usecase := server.Usecase(repo)

	app := server.New(log, metricsStore, usecase)

	app.Metrics.BuildInfo.Inc()
	app.RegisterRoutes(r)

	g, gCtx := errgroup.WithContext(ctx)

	app.RunHTTPServer(gCtx, g, cfg.AppConfig.Port, r)

	if err := g.Wait(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Error("server stopped", "err", err)
	}

	log.Info("shutdown complete")
}
