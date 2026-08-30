package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"log/slog"

	"golang.org/x/sync/errgroup"

	"github.com/sergeyWh1te/go-template/internal/connectors/metrics"
)

const (
	defaultReadTimeout  = 10 * time.Second
	defaultWriteTimeout = 10 * time.Second
	defaultIdleTimeout  = 60 * time.Second

	// How long in-flight requests get to finish once a signal arrives. It has to
	// be its own budget: the context that triggers the shutdown is already
	// canceled, and Shutdown returns immediately on a canceled context.
	defaultShutdownTimeout = 15 * time.Second
)

type App struct {
	Logger  *slog.Logger
	Metrics *metrics.Store
	usecase *usecase
}

func New(logger *slog.Logger, metricsStore *metrics.Store, usecase *usecase) *App {
	return &App{
		Logger:  logger,
		Metrics: metricsStore,
		usecase: usecase,
	}
}

func (a *App) RunHTTPServer(ctx context.Context, g *errgroup.Group, appPort uint, router http.Handler) {
	server := &http.Server{
		Addr:           fmt.Sprintf(`:%d`, appPort),
		Handler:        router,
		ReadTimeout:    defaultReadTimeout,
		WriteTimeout:   defaultWriteTimeout,
		IdleTimeout:    defaultIdleTimeout,
		MaxHeaderBytes: http.DefaultMaxHeaderBytes,
	}

	g.Go(func() error {
		// ErrServerClosed is the expected result of a graceful Shutdown, not a
		// failure — swallow it here so g.Wait() reports only real errors, and
		// so it cannot mask the shutdown error by arriving first.
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("http server: %w", err)
		}

		return nil
	})

	g.Go(func() error {
		<-ctx.Done()

		// A fresh context: ctx is already canceled, and Shutdown given a
		// canceled context returns instantly without draining anything.
		shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), defaultShutdownTimeout)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("http server shutdown: %w", err)
		}

		return nil
	})
}
