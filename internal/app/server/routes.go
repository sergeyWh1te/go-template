package server

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/lidofinance/go-template/internal/http/handlers/health"
)

func (app *app) RegisterAuthRoutes(router *mux.Router) {
	router.HandleFunc("/health", health.New().Handler).Methods(http.MethodGet)
}
