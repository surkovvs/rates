package main

import (
	"os"
	"os/signal"
	"rates_service/config"
	"rates_service/infrastructure/logger"
	"rates_service/internal/app"
	"syscall"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	zapLogger, zapLvl := logger.NewZap(2)

	cfg, err := config.NewAppConfig()
	if err != nil {
		zapLogger.Fatal("config init failed", zap.Error(err))
	}
	zapLvl.SetLevel(zapcore.Level(cfg.LogLvl))
	zapLogger.Info("configurations setted succsessfuly", zap.Reflect("config struct", cfg))

	App, err := app.NewApp(cfg, zapLogger)
	if err != nil {
		zapLogger.Error("app init failed", zap.Error(err))
	}
	zapLogger.Info("app init succsessfuly")

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
	if err := App.Run(sig); err != nil {
		zapLogger.Fatal("rates service running error", zap.Error(err))
	}
	zapLogger.Warn("rates service finished", zap.Int("PID", os.Getpid()))
}
