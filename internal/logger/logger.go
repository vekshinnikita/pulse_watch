package logger

import (
	"context"
	"io"
	"log"
	"log/slog"
	"os"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/vekshinnikita/pulse_watch/internal/config"
)

type ctxKey string

const loggerAttrsKey ctxKey = "logger_attrs"

type LoggerHandler struct {
	slog.Handler
}

func (l *LoggerHandler) Handle(ctx context.Context, rec slog.Record) error {
	attrs, _ := ctx.Value(loggerAttrsKey).([]slog.Attr)
	if len(attrs) > 0 {
		rec.AddAttrs(attrs...)
	}

	return l.Handler.Handle(ctx, rec)
}

// Функция добавления атрибутов в контекст
func AddLogAttrs(ctx context.Context, attrs ...slog.Attr) context.Context {
	existing, _ := ctx.Value(loggerAttrsKey).([]slog.Attr)
	return context.WithValue(ctx, loggerAttrsKey, append(existing, attrs...))
}

func NewLoggerHandler(loggerHandler slog.Handler) *LoggerHandler {
	return &LoggerHandler{loggerHandler}
}

func NewFileWriter() io.Writer {
	loggerConfig := GetConfig()

	// Настройка ротации логов по дате
	writer, err := rotatelogs.New(
		loggerConfig.Filepath,
		rotatelogs.WithLinkName("logs/app.log"), // ссылка на последний лог
		rotatelogs.WithRotationTime(time.Duration(loggerConfig.RotationHours)*time.Hour),
		rotatelogs.WithMaxAge(time.Duration(loggerConfig.MaxAgeDays)*24*time.Hour),
	)
	if err != nil {
		log.Fatalf("error happened while initializing logger: %s", err.Error())
	}

	return writer
}

func InitLogger() error {
	var handler slog.Handler
	commonConfig := config.GetCommonConfig()

	var writer io.Writer

	level := slog.LevelDebug
	writer = os.Stdout
	if !commonConfig.Debug {
		writer = NewFileWriter()
		level = slog.LevelInfo
	}

	handler = NewLoggerHandler(
		slog.NewJSONHandler(
			writer,
			&slog.HandlerOptions{
				Level: level,
			},
		),
	)

	logger := slog.New(handler)
	slog.SetDefault(logger)

	return nil
}
