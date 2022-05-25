package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	defaultReadTimeout  = 10 * time.Second
	defaultWriteTimeout = 10 * time.Second
	defaultIdleTimeout  = 60 * time.Second
)

type app struct {
	// add dependecies here
	// logger
	// metrics
}

func New() *app {
	return &app{}
}

func RunHTTPServer(ctx context.Context, appPort uint, router http.Handler) error {
	server := &http.Server{
		Addr:           fmt.Sprintf(`:%d`, appPort),
		Handler:        router,
		ReadTimeout:    defaultReadTimeout,
		WriteTimeout:   defaultWriteTimeout,
		IdleTimeout:    defaultIdleTimeout,
		MaxHeaderBytes: http.DefaultMaxHeaderBytes,
	}

	chErrors := make(chan error)
	chSignals := make(chan os.Signal, 1)

	signal.Notify(chSignals, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go listen(chErrors, server)

	var err error
	select {
	case err = <-chErrors:
		_ = shutdown(server)

	case <-chSignals:
		signal.Stop(chSignals)

		err = shutdown(server)
	case <-ctx.Done():

		if e := shutdown(server); e != nil {
			err = e
		}
	}

	close(chErrors)
	close(chSignals)

	return err
}

func shutdown(s *http.Server) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	s.SetKeepAlivesEnabled(false)

	if err := s.Shutdown(ctx); err != nil {
		return err
	}

	return nil
}

func listen(ch chan error, server *http.Server) {
	err := server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		ch <- err
	}
}
