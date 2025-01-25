package worker

import (
	"fmt"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type AsynqLogger struct {
}

func NewAsynqLogger() *AsynqLogger {
	return &AsynqLogger{}
}

func (logger *AsynqLogger) Print(level zerolog.Level, args ...interface{}) {
	log.WithLevel(level).Msg(fmt.Sprint(args...))
}

func (logger *AsynqLogger) Debug(args ...interface{}) {
	logger.Print(zerolog.DebugLevel, args...)
}

func (logger *AsynqLogger) Info(args ...interface{}) {
	logger.Print(zerolog.InfoLevel, args...)
}

func (logger *AsynqLogger) Warn(args ...interface{}) {
	logger.Print(zerolog.WarnLevel, args...)
}

func (logger *AsynqLogger) Error(args ...interface{}) {
	logger.Print(zerolog.ErrorLevel, args...)
}

func (logger *AsynqLogger) Fatal(args ...interface{}) {
	logger.Print(zerolog.FatalLevel, args...)
}
