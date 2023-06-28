package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"

	"github.com/sergeyWh1te/go-template/internal/connectors/metrics"
)

const (
	defaultReadTimeout  = 10 * time.Second
	defaultWriteTimeout = 10 * time.Second
	defaultIdleTimeout  = 60 * time.Second
)

type App struct {
	Logger  *logrus.Logger
	Metrics *metrics.Store
	usecase *usecase
	repo    *repository
}

func New(logger *logrus.Logger, metrics *metrics.Store, usecase *usecase, repo *repository) *App {
	return &App{
		Logger:  logger,
		Metrics: metrics,
		usecase: usecase,
		repo:    repo,
	}
}

func RunHTTPServer(ctx context.Context, appPort uint, router http.Handler) error {
	g, gCtx := errgroup.WithContext(ctx)

	server := &http.Server{
		Addr:           fmt.Sprintf(`:%d`, appPort),
		Handler:        router,
		ReadTimeout:    defaultReadTimeout,
		WriteTimeout:   defaultWriteTimeout,
		IdleTimeout:    defaultIdleTimeout,
		MaxHeaderBytes: http.DefaultMaxHeaderBytes,
	}

	g.Go(func() error {
		return server.ListenAndServe()
	})

	g.Go(func() error {
		<-gCtx.Done()
		return server.Shutdown(gCtx)
	})

	return g.Wait()
}
