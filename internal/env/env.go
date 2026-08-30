package env

import (
	"errors"
	"io/fs"
	"sync"

	"github.com/spf13/viper"
)

type Config struct {
	AppConfig AppConfig
	PgConfig  PgConfig
}

type AppConfig struct {
	Name      string
	Env       string
	URL       string
	Port      uint
	LogFormat string
	LogLevel  string
	SentryDsn string
}

type PgConfig struct {
	Port     uint
	Host     string
	Username string
	Password string
	Database string
	Schema   string
	SslMode  string
}

var (
	cfg Config

	onceDefaultClient sync.Once
)

func Read() (*Config, error) {
	var err error

	onceDefaultClient.Do(func() {
		viper.SetConfigType("env")
		viper.AddConfigPath(".")
		viper.SetConfigFile(".env")

		viper.AutomaticEnv()

		// READ_ENV_FROM_SHELL lets the process run with no .env at all — the
		// shell environment is the config. That is how the service runs in
		// Docker and in CI, where there is no file to read.
		if !viper.GetBool("READ_ENV_FROM_SHELL") {
			if viperErr := viper.ReadInConfig(); viperErr != nil {
				// SetConfigFile makes viper report a missing file as *fs.PathError,
				// not ConfigFileNotFoundError, so both have to be tolerated.
				_, notFound := errors.AsType[viper.ConfigFileNotFoundError](viperErr)
				if !notFound && !errors.Is(viperErr, fs.ErrNotExist) {
					err = viperErr
					return
				}
			}
		}

		cfg = Config{
			AppConfig: AppConfig{
				Name:      viper.GetString("APP_NAME"),
				Env:       viper.GetString("ENV"),
				URL:       viper.GetString("app.url"),
				Port:      viper.GetUint("PORT"),
				LogFormat: viper.GetString("LOG_FORMAT"),
				LogLevel:  viper.GetString("LOG_LEVEL"),
				SentryDsn: viper.GetString("SENTRY_DSN"),
			},
			PgConfig: PgConfig{
				Port:     viper.GetUint("PG_PORT"),
				Host:     viper.GetString("PG_HOST"),
				Username: viper.GetString("PG_USERNAME"),
				Password: viper.GetString("PG_PASSWORD"),
				Database: viper.GetString("PG_DATABASE"),
				Schema:   viper.GetString("PG_SCHEMA"),
				SslMode:  viper.GetString("PG_SSLMODE"),
			},
		}
	})

	return &cfg, err
}
