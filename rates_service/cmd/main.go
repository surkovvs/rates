package main

import (
	"fmt"
	"os"

	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	var lvl = zap.AtomicLevel{}
	logger, _ := zap.New(
		zapcore.NewCore(
			zapcore.NewConsoleEncoder(),
			os.Stdout,
			lvl))
	fx.New(
		func(logger *zap.Logger) fxevent.Logger {
			return &fxevent.ZapLogger{Logger: logger}
		},
		fx.Provide(SomeShit()))
}

func SomeShit() {
	for {
		fmt.Println("shit")
	}
}
