package asynq_server

import (
	"fmt"
	"log/slog"
	"os"
)

type SlogAdapter struct {
	logger *slog.Logger
}

func NewLogger() *SlogAdapter {
	return &SlogAdapter{
		logger: slog.Default(),
	}
}

func (l *SlogAdapter) Debug(args ...interface{}) {
	l.logger.Debug(fmt.Sprint(args...))
}

func (l *SlogAdapter) Info(args ...interface{}) {
	l.logger.Info(fmt.Sprint(args...))
}

func (l *SlogAdapter) Warn(args ...interface{}) {
	l.logger.Warn(fmt.Sprint(args...))
}

func (l *SlogAdapter) Error(args ...interface{}) {
	l.logger.Error(fmt.Sprint(args...))
}

func (l *SlogAdapter) Fatal(args ...interface{}) {
	l.logger.Error(fmt.Sprint(args...))
	os.Exit(1)
}
