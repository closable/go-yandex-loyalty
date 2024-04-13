package handlers

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewLogger() zap.Logger {
	// add color level for convience
	zapConfig := zap.NewDevelopmentConfig()
	zapConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	logger, err := zapConfig.Build()

	if err != nil {
		panic(err)
	}
	defer logger.Sync()
	return *logger
}
