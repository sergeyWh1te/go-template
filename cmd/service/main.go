package main

import (
	"context"
	"fmt"
	"os"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"

	"github.com/lidofinance/go-template/internal/app/server"
	"github.com/lidofinance/go-template/internal/connectors/logger"
	"github.com/lidofinance/go-template/internal/env"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg, envErr := env.Read(ctx)
	if envErr != nil {
		println("Read env error:", envErr.Error())

		os.Exit(1)
	}

	log, logErr := logger.New(&cfg.AppConfig)
	if logErr != nil {
		println("Logger error:", logErr.Error())

		os.Exit(1)
	}

	log.Info(fmt.Sprintf(`started %s application`, cfg.AppConfig.Name))

	r := mux.NewRouter()
	app := server.New()

	app.RegisterAuthRoutes(r)

	if err := server.RunHTTPServer(ctx, cfg.AppConfig.Port, handlers.RecoveryHandler()(r)); err != nil {
		println(err)
	}
}
