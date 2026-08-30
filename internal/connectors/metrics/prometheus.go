package metrics

import (
	"regexp"
	"runtime"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var re = regexp.MustCompile(`[-\s]+`)

// Commit is stamped at build time:
//
//	go build -ldflags="-X github.com/sergeyWh1te/go-template/internal/connectors/metrics.Commit=$(git rev-parse HEAD)"
var Commit string

type Store struct {
	Prometheus *prometheus.Registry
	BuildInfo  prometheus.Counter
}

func New(promRegistry *prometheus.Registry, appName, env string) *Store {
	// The default registry is never exposed — /metrics serves promRegistry — so
	// the Go and process collectors have to be registered here to show up at all.
	promRegistry.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)

	prefix := re.ReplaceAllString(appName, `_`)

	return &Store{
		Prometheus: promRegistry,
		// promauto.With(promRegistry) registers into the registry that /metrics
		// serves; bare promauto.NewCounter would publish to the global default
		// registry instead, where nothing ever reads it.
		BuildInfo: promauto.With(promRegistry).NewCounter(prometheus.CounterOpts{
			Name: prefix + "_metric_build_info",
			Help: "Build information",
			ConstLabels: prometheus.Labels{
				"name":    appName,
				"env":     env,
				"commit":  Commit,
				"version": runtime.Version(),
			},
		}),
	}
}
