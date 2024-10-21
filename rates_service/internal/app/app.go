package app

import (
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

func NewApp() *fx.App {
	app := fx.New(fx.WithLogger(func(logger *zap.Logger) fxevent.Logger {
		return &fxevent.ZapLogger{Logger: logger}
	}), fx.Provide())

	migr, err := migrate.New(fmt.Sprintf("file://%s", cfg.MigrationPath),
		fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
			cfg.DB.User, cfg.DB.Password, cfg.DB.Host, cfg.DB.Port, cfg.DB.Name))

	return app
}
