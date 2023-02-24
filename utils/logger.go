package utils

import (
	"log"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Logger *zap.SugaredLogger

func InitializeLogger(filename string) {

	config := zap.NewProductionConfig()
	config.OutputPaths = []string{filename}
	config.EncoderConfig.CallerKey = "caller"
	config.EncoderConfig.LevelKey = "level"
	config.EncoderConfig.MessageKey = "msg"
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder
	config.EncoderConfig.StacktraceKey = "" // to hide stacktrace info

	logger, err := config.Build()
	if err != nil {
		log.Fatal(err)
	}

	Logger = logger.Sugar()
	Logger.Sync()
}
