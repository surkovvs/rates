package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewZap(initLvl int8) (*zap.Logger, zap.AtomicLevel) {
	lvl := zap.NewAtomicLevelAt(zapcore.Level(initLvl))
	logger := zap.New(
		zapcore.NewCore(
			zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
			os.Stdout,
			lvl))
	logger.Level()
	return logger, lvl
}
