package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/vekshinnikita/pulse_watch/internal/constants"
	"github.com/vekshinnikita/pulse_watch/internal/logger"
)

type bodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *bodyWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func RequestLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		bw := &bodyWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = bw

		c.Next()

		ctx := c.Request.Context()
		duration := time.Since(start)
		status := c.Writer.Status()

		message := "request successfully handled"
		if status >= 400 && status < 600 {
			message = "error happen"

			// Если статус ошибка, то логгируем сообщение
			var data map[string]interface{}
			if err := json.Unmarshal(bw.body.Bytes(), &data); err == nil {
				if msg, ok := data["message"]; ok {
					message = msg.(string)
				}
			}
		}

		slog.InfoContext(ctx, message,
			slog.String("method", c.Request.Method),
			slog.String("path", c.FullPath()),
			slog.Int("status", status),
			slog.Int64("duration_ms", duration.Milliseconds()),
			slog.String("ip", c.ClientIP()),
		)
	}
}

func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		reqID := c.GetHeader("X-Request-ID")

		if reqID == "" {
			reqID = uuid.New().String()
		}

		ctx = context.WithValue(ctx, constants.RequestIDKey, reqID)
		// Добавляем request_id в контекст логгера
		ctx = logger.AddLogAttrs(ctx,
			slog.String("request_id", reqID),
		)

		c.Writer.Header().Set("X-Request-ID", reqID)

		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

func RecoveryWithLogging() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, err interface{}) {
		ctx := c.Request.Context()
		stack := string(debug.Stack())

		// Формируем лог
		slog.ErrorContext(ctx, fmt.Sprintf("runtime error: %v", err),
			slog.String("trace", stack),
		)

		// Отправляем клиенту 500
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error":      "Internal Server Error",
			"request_id": ctx.Value(constants.RequestIDKey),
		})
	})
}
