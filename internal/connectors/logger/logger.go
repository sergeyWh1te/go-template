package logger

import (
	"os"
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/evalphobia/logrus_sentry"

	"github.com/lidofinance/go-template/internal/env"
)

var (
	logger            *logrus.Logger
	onceDefaultClient sync.Once
)

func New(cfg *env.AppConfig) (*logrus.Logger, error) {
	var (
		err error
	)

	onceDefaultClient.Do(func() {
		logger = logrus.StandardLogger()

		logLevel, err := logrus.ParseLevel(cfg.LogLevel)
		if err != nil {
			return
		}

		logger.SetLevel(logLevel)
		logger.SetFormatter(&logrus.TextFormatter{})

		if cfg.LogFormat == "json" {
			logger.SetFormatter(&logrus.JSONFormatter{})
		}

		logger.SetOutput(os.Stdout)

		if cfg.SentryDsn != "" {
			hook, err := logrus_sentry.NewSentryHook(cfg.SentryDsn, []logrus.Level{
				logrus.PanicLevel,
				logrus.FatalLevel,
				logrus.ErrorLevel,
			})

			if err != nil {
				return 	
			}

			logger.Hooks.Add(hook)
		}
	})

	return logger, err
}
