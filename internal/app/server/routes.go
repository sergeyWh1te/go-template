package server

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"

	"github.com/sergeyWh1te/go-template/internal/http/handlers/health"
	userexample "github.com/sergeyWh1te/go-template/internal/http/handlers/user_example"
)

func (app *App) RegisterRoutes(r chi.Router) {
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/health", health.New().Handler)
	r.Method(http.MethodGet, "/metrics", promhttp.HandlerFor(app.Metrics.Prometheus, promhttp.HandlerOpts{}))

	r.Get("/example", userexample.New(app.Logger, app.usecase.User).Handler)
}
