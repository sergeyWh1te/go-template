package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi/v5"
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

	log := logger.New(&cfg.AppConfig)

	db, errDB := postgres.Connect(&cfg.PgConfig)
	if errDB != nil {
		log.Error("could not connect to db", "err", errDB)
		return
	}
	// Closed explicitly after the HTTP server has drained (see the end of main),
	// with this defer as the safety net for the early-return paths below.
	defer func() {
		if err := postgres.Close(); err != nil {
			log.Error("could not close db connection", "err", err)
		}
	}()

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

	// RunHTTPServer already treats ErrServerClosed as success, so anything left
	// here is a real failure. Filtering it out again would hide the shutdown
	// error, which errgroup would otherwise drop as the second one reported.
	err := g.Wait()

	// Stop trapping signals now that the server is down: from here on a second
	// Ctrl-C should kill the process outright rather than be swallowed while the
	// database connection is closing.
	stop()

	// Close the database only now: the HTTP server has drained, so no handler is
	// still holding a connection. database/sql waits for in-flight queries, and
	// the deferred Close above turns into a no-op once this one has run.
	if closeErr := postgres.Close(); closeErr != nil {
		log.Error("could not close db connection", "err", closeErr)
	}

	if err != nil {
		log.Error("server stopped", "err", err)

		return
	}

	log.Info("shutdown complete")
}
