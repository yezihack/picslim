package logx

import "go.uber.org/zap"

func New(level string) *zap.Logger {
	cfg := zap.NewProductionConfig()
	if level == "debug" {
		cfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	}
	logger, err := cfg.Build()
	if err != nil {
		fallback, _ := zap.NewProduction()
		return fallback
	}
	return logger
}
