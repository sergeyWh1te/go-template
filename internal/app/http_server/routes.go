package server

import (
	"net/http"

	chi "github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/sergeyWh1te/go-template/internal/http/handlers/health"
	userexample "github.com/sergeyWh1te/go-template/internal/http/handlers/user_example"
)

func (a *App) RegisterRoutes(r chi.Router) {
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/health", health.New().Handler)
	// Serve the app's own registry: promhttp.Handler() would expose the global
	// default registry, which is not where the metrics Store registers.
	r.Method(http.MethodGet, "/metrics", promhttp.HandlerFor(a.Metrics.Prometheus, promhttp.HandlerOpts{}))

	r.Get("/example", userexample.New(a.Logger, a.usecase.User).Handler)
}
